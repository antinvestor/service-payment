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
	"context"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect/v2"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/service/handlers"
	"github.com/pitabwire/frame/v3/frametests/definition"
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

		orgResp, err := ledgerServer.CreateBook(ctx, &ledgerv1.CreateBookRequest{
			Name:     "Test Organization",
			Type:     "platform",
			Currency: "UGX",
		})
		require.NoError(t, err)
		require.NotNil(t, orgResp)
		assert.Equal(t, "Test Organization", orgResp.GetData().GetName())
		assert.Equal(t, "platform", orgResp.GetData().GetType())
		orgID := orgResp.GetData().GetId()
		require.NotEmpty(t, orgID)

		groupResp, err := ledgerServer.CreateBook(ctx, &ledgerv1.CreateBookRequest{
			Name:     "Group A",
			Type:     "group",
			ParentId: orgID,
			Currency: "UGX",
		})
		require.NoError(t, err)
		assert.Equal(t, orgID, groupResp.GetData().GetParentId())

		got, err := ledgerServer.GetBook(ctx, &ledgerv1.GetBookRequest{Id: groupResp.GetData().GetId()})
		require.NoError(t, err)
		assert.Equal(t, "Group A", got.GetData().GetName())

		listed, err := ledgerServer.ListBooksByType(ctx, &ledgerv1.ListBooksByTypeRequest{Type: "group"})
		require.NoError(t, err)
		require.Len(t, listed.GetData(), 1)
		assert.Equal(t, "Group A", listed.GetData()[0].GetName())
	})
}

// TestCreateBookWithCallerIDHandler proves the handler keeps a caller-
// supplied book id, answers a repeated create with the same book, maps a
// conflicting reuse to AlreadyExists and an unknown book to NotFound, and
// scopes book ids to the caller's tenancy.
func (s *LedgerHandlersTestSuite) TestCreateBookWithCallerIDHandler() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		tenantCtx := s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(tenantCtx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
			resources.ReportBusiness,
			resources.BookBusiness,
		)

		const roomID = "room-0123456789abcdef"
		create := func(c context.Context, name string) (*ledgerv1.CreateBookResponse, error) {
			return ledgerServer.CreateBook(c, &ledgerv1.CreateBookRequest{Id: roomID, Name: name, Type: "group", Currency: "UGX"})
		}

		first, err := create(tenantCtx, "Room")
		require.NoError(t, err)
		assert.Equal(t, roomID, first.GetData().GetId())

		again, err := create(tenantCtx, "Room")
		require.NoError(t, err)
		assert.Equal(t, roomID, again.GetData().GetId())

		_, err = create(tenantCtx, "Renamed")
		assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))

		_, err = ledgerServer.GetBook(tenantCtx, &ledgerv1.GetBookRequest{Id: "missing-book"})
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))

		_, err = ledgerServer.CreateBook(tenantCtx, &ledgerv1.CreateBookRequest{Id: "Not Valid", Name: "Room", Type: "group"})
		assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))

		// Book ids are scoped to the tenancy: another tenant neither sees,
		// matches nor is blocked by this tenant's id.
		otherTenant := util.IDString()
		otherPartition := util.IDString()
		otherProfile := util.IDString()
		otherCtx := s.WithAuthClaims(ctx, otherTenant, otherPartition, otherProfile)
		s.SeedTenantRole(otherCtx, svc, otherTenant, otherPartition, otherProfile, authz.RoleOwner)

		_, err = ledgerServer.GetBook(otherCtx, &ledgerv1.GetBookRequest{Id: roomID})
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err),
			"another tenant must get the same answer as for an id nobody holds")

		theirs, err := create(otherCtx, "Their room")
		require.NoError(t, err, "another tenant must not be blocked by this tenant's id")
		assert.Equal(t, roomID, theirs.GetData().GetId())
		assert.Equal(t, "Their room", theirs.GetData().GetName())

		_, err = create(otherCtx, "Room")
		assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err),
			"a create is matched against the caller's own book only")

		ours, err := ledgerServer.GetBook(tenantCtx, &ledgerv1.GetBookRequest{Id: roomID})
		require.NoError(t, err)
		assert.Equal(t, "Room", ours.GetData().GetName())

		for c, want := range map[context.Context]string{tenantCtx: "Room", otherCtx: "Their room"} {
			listed, lErr := ledgerServer.ListBooksByType(c, &ledgerv1.ListBooksByTypeRequest{Type: "group"})
			require.NoError(t, lErr)
			require.Len(t, listed.GetData(), 1)
			assert.Equal(t, want, listed.GetData()[0].GetName())
		}
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
		_, err := ledgerServer.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{Id: "tb-asset", Type: ledgerv1.LedgerType_ASSET})
		require.NoError(t, err)
		_, err = ledgerServer.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{Id: "tb-income", Type: ledgerv1.LedgerType_INCOME})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{Id: "tb-cash", LedgerId: "tb-asset", Currency: "UGX"})
		require.NoError(t, err)
		_, err = ledgerServer.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{Id: "tb-rev", LedgerId: "tb-income", Currency: "UGX"})
		require.NoError(t, err)

		_, err = ledgerServer.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
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
		})
		require.NoError(t, err)

		tb, err := ledgerServer.GetTrialBalance(ctx, &ledgerv1.GetTrialBalanceRequest{Currency: "UGX"})
		require.NoError(t, err)
		require.Len(t, tb.GetTotals(), 1)
		total := tb.GetTotals()[0]
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

		_, err := ledgerServer.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{Id: "vd-asset", Type: ledgerv1.LedgerType_ASSET})
		require.NoError(t, err)
		_, err = ledgerServer.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{Id: "vd-income", Type: ledgerv1.LedgerType_INCOME})
		require.NoError(t, err)
		_, err = ledgerServer.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{Id: "vd-cash", LedgerId: "vd-asset", Currency: "UGX"})
		require.NoError(t, err)
		_, err = ledgerServer.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{Id: "vd-rev", LedgerId: "vd-income", Currency: "UGX"})
		require.NoError(t, err)

		// Pending transaction (cleared=false).
		_, err = ledgerServer.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "vd-pending",
			Currency: "UGX",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "vd-cash", Credit: false, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 60}},
				{AccountId: "vd-rev", Credit: true, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 60}},
			},
		})
		require.NoError(t, err)

		voided, err := ledgerServer.VoidTransaction(ctx, &ledgerv1.VoidTransactionRequest{Id: "vd-pending"})
		require.NoError(t, err)
		assert.Equal(t, ledgerv1.TransactionStatus_VOIDED, voided.GetData().GetStatus())
		assert.NotEmpty(t, voided.GetData().GetVoidedAt())

		// Voiding a posted transaction must fail.
		_, err = ledgerServer.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "vd-posted",
			Currency: "UGX",
			Type:     ledgerv1.TransactionType_NORMAL,
			Cleared:  true,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "vd-cash", Credit: false, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 30}},
				{AccountId: "vd-rev", Credit: true, Amount: &commonv1.Money{CurrencyCode: "UGX", Units: 30}},
			},
		})
		require.NoError(t, err)

		_, err = ledgerServer.VoidTransaction(ctx, &ledgerv1.VoidTransactionRequest{Id: "vd-posted"})
		require.Error(t, err, "voiding a posted transaction must be rejected")
	})
}
