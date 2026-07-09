// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers_test

import (
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/service/handlers"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateAndGetBookHandler proves the Book CRUD handler chain end to
// end: a CreateBook payload becomes persisted with id + parent_id + type,
// and a subsequent GetBook returns the same row.
func (s *LedgerHandlersTestSuite) TestCreateAndGetBookHandler() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
			resources.ReportBusiness,
			resources.BookBusiness,
		)

		orgResp, err := ledgerServer.CreateBook(ctx, &connect.Request[ledgerv1.CreateBookRequest]{
			Msg: &ledgerv1.CreateBookRequest{
				Name:     "Test Organization",
				Type:     "platform",
				Currency: "UGX",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, orgResp)
		assert.Equal(t, "Test Organization", orgResp.Msg.GetData().GetName())
		assert.Equal(t, "platform", orgResp.Msg.GetData().GetType())
		orgID := orgResp.Msg.GetData().GetId()
		require.NotEmpty(t, orgID)

		groupResp, err := ledgerServer.CreateBook(ctx, &connect.Request[ledgerv1.CreateBookRequest]{
			Msg: &ledgerv1.CreateBookRequest{
				Name:     "Group A",
				Type:     "group",
				ParentId: orgID,
				Currency: "UGX",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, orgID, groupResp.Msg.GetData().GetParentId())

		got, err := ledgerServer.GetBook(ctx, &connect.Request[ledgerv1.GetBookRequest]{
			Msg: &ledgerv1.GetBookRequest{Id: groupResp.Msg.GetData().GetId()},
		})
		require.NoError(t, err)
		assert.Equal(t, "Group A", got.Msg.GetData().GetName())

		listed, err := ledgerServer.ListBooksByType(ctx, &connect.Request[ledgerv1.ListBooksByTypeRequest]{
			Msg: &ledgerv1.ListBooksByTypeRequest{Type: "group"},
		})
		require.NoError(t, err)
		require.Len(t, listed.Msg.GetData(), 1)
		assert.Equal(t, "Group A", listed.Msg.GetData()[0].GetName())
	})
}

// TestGetTrialBalanceHandler exercises the handler end to end against a
// posted transaction and confirms the integrity check rolls up correctly
// per currency.
func (s *LedgerHandlersTestSuite) TestGetTrialBalanceHandler() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
			resources.ReportBusiness,
			resources.BookBusiness,
		)

		// Minimal fixtures: asset + income ledgers, one account each.
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "tb-asset", Type: ledgerv1.LedgerType_ASSET},
		})
		require.NoError(t, err)
		_, err = ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "tb-income", Type: ledgerv1.LedgerType_INCOME},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{Id: "tb-cash", LedgerId: "tb-asset", Currency: "UGX"},
		})
		require.NoError(t, err)
		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{Id: "tb-rev", LedgerId: "tb-income", Currency: "UGX"},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{
				Id:       "tb-txn",
				Currency: "UGX",
				Type:     ledgerv1.TransactionType_NORMAL,
				Cleared:  true,
				Entries: []*ledgerv1.TransactionEntry{
					{
						AccountId: "tb-cash", Credit: false,
						Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 150},
					},
					{
						AccountId: "tb-rev", Credit: true,
						Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 150},
					},
				},
			},
		})
		require.NoError(t, err)

		tb, err := ledgerServer.GetTrialBalance(ctx, &connect.Request[ledgerv1.GetTrialBalanceRequest]{
			Msg: &ledgerv1.GetTrialBalanceRequest{Currency: "UGX"},
		})
		require.NoError(t, err)
		require.Len(t, tb.Msg.GetTotals(), 1)
		total := tb.Msg.GetTotals()[0]
		assert.Equal(t, "UGX", total.GetCurrency())
		assert.True(t, total.GetIsBalanced())
		assert.Equal(t, int64(150), total.GetTotalDebits().GetUnits())
		assert.Equal(t, int64(150), total.GetTotalCredits().GetUnits())
	})
}

// TestVoidPendingTransactionHandler verifies the handler enforces the
// state-machine rule: only draft/pending transactions can be voided.
func (s *LedgerHandlersTestSuite) TestVoidPendingTransactionHandler() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
			resources.ReportBusiness,
			resources.BookBusiness,
		)

		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "vd-asset", Type: ledgerv1.LedgerType_ASSET},
		})
		require.NoError(t, err)
		_, err = ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "vd-income", Type: ledgerv1.LedgerType_INCOME},
		})
		require.NoError(t, err)
		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{Id: "vd-cash", LedgerId: "vd-asset", Currency: "UGX"},
		})
		require.NoError(t, err)
		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{Id: "vd-rev", LedgerId: "vd-income", Currency: "UGX"},
		})
		require.NoError(t, err)

		// Pending transaction (cleared=false).
		_, err = ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{
				Id:       "vd-pending",
				Currency: "UGX",
				Type:     ledgerv1.TransactionType_NORMAL,
				Entries: []*ledgerv1.TransactionEntry{
					{AccountId: "vd-cash", Credit: false, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 60}},
					{AccountId: "vd-rev", Credit: true, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 60}},
				},
			},
		})
		require.NoError(t, err)

		voided, err := ledgerServer.VoidTransaction(ctx, &connect.Request[ledgerv1.VoidTransactionRequest]{
			Msg: &ledgerv1.VoidTransactionRequest{Id: "vd-pending"},
		})
		require.NoError(t, err)
		assert.Equal(t, ledgerv1.TransactionStatus_VOIDED, voided.Msg.GetData().GetStatus())
		assert.NotEmpty(t, voided.Msg.GetData().GetVoidedAt())

		// Voiding a posted transaction must fail.
		_, err = ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{
				Id:       "vd-posted",
				Currency: "UGX",
				Type:     ledgerv1.TransactionType_NORMAL,
				Cleared:  true,
				Entries: []*ledgerv1.TransactionEntry{
					{AccountId: "vd-cash", Credit: false, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 30}},
					{AccountId: "vd-rev", Credit: true, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 30}},
				},
			},
		})
		require.NoError(t, err)

		_, err = ledgerServer.VoidTransaction(ctx, &connect.Request[ledgerv1.VoidTransactionRequest]{
			Msg: &ledgerv1.VoidTransactionRequest{Id: "vd-posted"},
		})
		require.Error(t, err, "voiding a posted transaction must be rejected")
	})
}
