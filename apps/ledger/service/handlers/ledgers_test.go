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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v1/ledgerv1connect"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/service/handlers"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"
)

// LedgerHandlersTestSuite extends BaseTestSuite for handler tests.
type LedgerHandlersTestSuite struct {
	tests.BaseTestSuite
}

// --- ToConnectError tests ---

func TestToConnectError_Nil(t *testing.T) {
	assert.NoError(t, handlers.ToConnectError(nil))
}

func TestToConnectError_NonApplicationError(t *testing.T) {
	plainErr := errors.New("plain error")
	result := handlers.ToConnectError(plainErr)
	assert.Equal(t, plainErr, result)
}

func TestToConnectError_InvalidArgument(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrUnspecifiedID)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

func TestToConnectError_NotFound(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrLedgerNotFound)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
}

func TestToConnectError_AlreadyExists(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrTransactionAlreadyExists)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
}

func TestToConnectError_FailedPrecondition(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrTransactionTypeNotReversible)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeFailedPrecondition, connectErr.Code())
}

func TestToConnectError_CurrencyUnknown(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrAccountsCurrencyUnknown)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

func TestToConnectError_Default(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrSystemFailure)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInternal, connectErr.Code())
}

func TestToConnectError_AccountNotFound(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrAccountNotFound)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
}

func TestToConnectError_TransactionNotFound(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrTransactionNotFound)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
}

func TestToConnectError_BadDataSupplied(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrBadDataSupplied)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

func TestToConnectError_TransactionConflicting(t *testing.T) {
	err := handlers.ToConnectError(apperrors.ErrTransactionIsConflicting)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
}

func TestToConnectError_SearchErrors(t *testing.T) {
	searchErrors := []struct {
		err  apperrors.ApplicationError
		code connect.Code
	}{
		{apperrors.ErrSearchNamespaceUnknown, connect.CodeNotFound},
		{apperrors.ErrSearchQueryHasInvalidFormat, connect.CodeInvalidArgument},
		{apperrors.ErrSearchQueryHasInvalidKeys, connect.CodeInvalidArgument},
	}

	for _, tc := range searchErrors {
		err := handlers.ToConnectError(tc.err)
		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, tc.code, connectErr.Code())
	}
}

func TestToConnectError_ExtendedError(t *testing.T) {
	extended := apperrors.ErrUnspecifiedID.Extend("extra detail")
	err := handlers.ToConnectError(extended)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

// --- Integration handler tests ---

func (s *LedgerHandlersTestSuite) TestCreateLedger() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		req := &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{
				Id:       "test-ledger",
				Type:     ledgerv1.LedgerType_ASSET,
				ParentId: "",
				Data:     nil,
			},
		}

		resp, err := ledgerServer.CreateLedger(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "test-ledger", resp.Msg.GetData().GetId())
		assert.Equal(t, ledgerv1.LedgerType_ASSET, resp.Msg.GetData().GetType())
	})
}

func (s *LedgerHandlersTestSuite) TestUpdateLedger() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// Create ledger first
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{
				Id:   "update-handler-ledger",
				Type: ledgerv1.LedgerType_ASSET,
			},
		})
		require.NoError(t, err)

		// Update it
		strct, _ := structpb.NewStruct(map[string]any{"name": "updated"})
		resp, err := ledgerServer.UpdateLedger(ctx, &connect.Request[ledgerv1.UpdateLedgerRequest]{
			Msg: &ledgerv1.UpdateLedgerRequest{
				Id:   "update-handler-ledger",
				Data: strct,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "updated", resp.Msg.GetData().GetData().GetFields()["name"].GetStringValue())
	})
}

func (s *LedgerHandlersTestSuite) TestCreateAccount() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// Create ledger first
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{
				Id:   "handler-acc-ledger",
				Type: ledgerv1.LedgerType_ASSET,
			},
		})
		require.NoError(t, err)

		// Create account
		resp, err := ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-account",
				LedgerId: "handler-acc-ledger",
				Currency: "USD",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "handler-account", resp.Msg.GetData().GetId())
	})
}

func (s *LedgerHandlersTestSuite) TestUpdateAccount() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// Create ledger and account
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{
				Id:   "handler-upd-ledger",
				Type: ledgerv1.LedgerType_ASSET,
			},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-upd-account",
				LedgerId: "handler-upd-ledger",
				Currency: "USD",
			},
		})
		require.NoError(t, err)

		// Update account
		strct, _ := structpb.NewStruct(map[string]any{"desc": "updated"})
		resp, err := ledgerServer.UpdateAccount(ctx, &connect.Request[ledgerv1.UpdateAccountRequest]{
			Msg: &ledgerv1.UpdateAccountRequest{
				Id:   "handler-upd-account",
				Data: strct,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func (s *LedgerHandlersTestSuite) TestCreateTransaction() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// Setup ledgers and accounts
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "handler-txn-asset", Type: ledgerv1.LedgerType_ASSET},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "handler-txn-income", Type: ledgerv1.LedgerType_INCOME},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-asset-acc",
				LedgerId: "handler-txn-asset",
				Currency: "USD",
			},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-income-acc",
				LedgerId: "handler-txn-income",
				Currency: "USD",
			},
		})
		require.NoError(t, err)

		// Create transaction
		resp, err := ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{
				Id:       "handler-txn-1",
				Currency: "USD",
				Type:     ledgerv1.TransactionType_NORMAL,
				Entries: []*ledgerv1.TransactionEntry{
					{
						AccountId: "handler-asset-acc",
						Credit:    false,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
					{
						AccountId: "handler-income-acc",
						Credit:    true,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "handler-txn-1", resp.Msg.GetData().GetId())
		assert.Len(t, resp.Msg.GetData().GetEntries(), 2)
	})
}

func (s *LedgerHandlersTestSuite) TestReverseTransaction() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// Setup
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "handler-rev-asset", Type: ledgerv1.LedgerType_ASSET},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "handler-rev-income", Type: ledgerv1.LedgerType_INCOME},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-rev-asset-acc",
				LedgerId: "handler-rev-asset",
				Currency: "USD",
			},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-rev-income-acc",
				LedgerId: "handler-rev-income",
				Currency: "USD",
			},
		})
		require.NoError(t, err)

		// Create transaction. Reversal is only valid on posted transactions,
		// so cleared=true posts it immediately.
		_, err = ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{
				Id:       "handler-rev-txn",
				Currency: "USD",
				Type:     ledgerv1.TransactionType_NORMAL,
				Cleared:  true,
				Entries: []*ledgerv1.TransactionEntry{
					{
						AccountId: "handler-rev-asset-acc",
						Credit:    false,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
					{
						AccountId: "handler-rev-income-acc",
						Credit:    true,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
				},
			},
		})
		require.NoError(t, err)

		// Reverse it
		resp, err := ledgerServer.ReverseTransaction(ctx, &connect.Request[ledgerv1.ReverseTransactionRequest]{
			Msg: &ledgerv1.ReverseTransactionRequest{Id: "handler-rev-txn"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, ledgerv1.TransactionType_REVERSAL, resp.Msg.GetData().GetType())
	})
}

func (s *LedgerHandlersTestSuite) TestUpdateTransaction() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// Setup
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "handler-upd-txn-asset", Type: ledgerv1.LedgerType_ASSET},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "handler-upd-txn-income", Type: ledgerv1.LedgerType_INCOME},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-upd-asset-acc",
				LedgerId: "handler-upd-txn-asset",
				Currency: "USD",
			},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-upd-income-acc",
				LedgerId: "handler-upd-txn-income",
				Currency: "USD",
			},
		})
		require.NoError(t, err)

		// Create transaction
		_, err = ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{
				Id:       "handler-upd-txn",
				Currency: "USD",
				Type:     ledgerv1.TransactionType_NORMAL,
				Entries: []*ledgerv1.TransactionEntry{
					{
						AccountId: "handler-upd-asset-acc",
						Credit:    false,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
					{
						AccountId: "handler-upd-income-acc",
						Credit:    true,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
				},
			},
		})
		require.NoError(t, err)

		// Update
		strct, _ := structpb.NewStruct(map[string]any{"status": "settled"})
		resp, err := ledgerServer.UpdateTransaction(ctx, &connect.Request[ledgerv1.UpdateTransactionRequest]{
			Msg: &ledgerv1.UpdateTransactionRequest{
				Id:   "handler-upd-txn",
				Data: strct,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func (s *LedgerHandlersTestSuite) TestCreateLedgerMissingID() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Type: ledgerv1.LedgerType_ASSET},
		})
		require.Error(t, err)
	})
}

func (s *LedgerHandlersTestSuite) TestCreateAccountInvalidCurrency() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "handler-bad-curr-ledger", Type: ledgerv1.LedgerType_ASSET},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{
				Id:       "handler-bad-curr-acc",
				LedgerId: "handler-bad-curr-ledger",
				Currency: "INVALID",
			},
		})
		require.Error(t, err)
	})
}

func (s *LedgerHandlersTestSuite) TestUpdateLedgerMissingID() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		_, err := ledgerServer.UpdateLedger(ctx, &connect.Request[ledgerv1.UpdateLedgerRequest]{
			Msg: &ledgerv1.UpdateLedgerRequest{},
		})
		require.Error(t, err)
	})
}

func (s *LedgerHandlersTestSuite) TestErrorPaths() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// UpdateAccount error path — non-existent account
		_, err := ledgerServer.UpdateAccount(ctx, &connect.Request[ledgerv1.UpdateAccountRequest]{
			Msg: &ledgerv1.UpdateAccountRequest{Id: "non-existent"},
		})
		require.Error(t, err)

		// CreateTransaction error path — missing data
		_, err = ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{},
		})
		require.Error(t, err)

		// ReverseTransaction error path — non-existent
		_, err = ledgerServer.ReverseTransaction(ctx, &connect.Request[ledgerv1.ReverseTransactionRequest]{
			Msg: &ledgerv1.ReverseTransactionRequest{Id: "non-existent"},
		})
		require.Error(t, err)

		// UpdateTransaction error path — non-existent
		_, err = ledgerServer.UpdateTransaction(ctx, &connect.Request[ledgerv1.UpdateTransactionRequest]{
			Msg: &ledgerv1.UpdateTransactionRequest{Id: "non-existent"},
		})
		require.Error(t, err)

		// CreateAccount error path — invalid currency
		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{Id: "bad", Currency: "INVALID"},
		})
		require.Error(t, err)
	})
}

func (s *LedgerHandlersTestSuite) TestStreamingSearchEndpoints() {
	s.WithTestDependencies(s.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, svc, resources := s.CreateService(t, depOpt)

		// Setup auth
		tenantID := util.IDString()
		partitionID := util.IDString()
		profileID := util.IDString()
		ctx = s.WithAuthClaims(ctx, tenantID, partitionID, profileID)
		s.SeedTenantRole(ctx, svc, tenantID, partitionID, profileID, authz.RoleOwner)

		ledgerServer := handlers.NewLedgerServer(
			resources.LedgerBusiness,
			resources.AccountBusiness,
			resources.TransactionBusiness,
		)

		// Create HTTP test server with connect handler and inject test claims
		path, handler := ledgerv1connect.NewLedgerServiceHandler(ledgerServer)
		claimsMiddleware := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
		mux := http.NewServeMux()
		mux.Handle(path, claimsMiddleware)
		srv := httptest.NewServer(mux)
		defer srv.Close()

		client := ledgerv1connect.NewLedgerServiceClient(srv.Client(), srv.URL)

		// Setup: create ledgers, accounts, and a transaction
		_, err := ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "stream-asset", Type: ledgerv1.LedgerType_ASSET},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateLedger(ctx, &connect.Request[ledgerv1.CreateLedgerRequest]{
			Msg: &ledgerv1.CreateLedgerRequest{Id: "stream-income", Type: ledgerv1.LedgerType_INCOME},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{Id: "stream-asset-acc", LedgerId: "stream-asset", Currency: "USD"},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateAccount(ctx, &connect.Request[ledgerv1.CreateAccountRequest]{
			Msg: &ledgerv1.CreateAccountRequest{Id: "stream-income-acc", LedgerId: "stream-income", Currency: "USD"},
		})
		require.NoError(t, err)

		_, err = ledgerServer.CreateTransaction(ctx, &connect.Request[ledgerv1.CreateTransactionRequest]{
			Msg: &ledgerv1.CreateTransactionRequest{
				Id:       "stream-txn-1",
				Currency: "USD",
				Type:     ledgerv1.TransactionType_NORMAL,
				Cleared:  true,
				Entries: []*ledgerv1.TransactionEntry{
					{
						AccountId: "stream-asset-acc",
						Credit:    false,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
					{
						AccountId: "stream-income-acc",
						Credit:    true,
						Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
					},
				},
			},
		})
		require.NoError(t, err)

		// Test SearchLedgers
		ledgerStream, err := client.SearchLedgers(ctx, connect.NewRequest(&commonv1.SearchRequest{Query: "{}"}))
		require.NoError(t, err)
		var foundLedgers []*ledgerv1.Ledger
		for ledgerStream.Receive() {
			foundLedgers = append(foundLedgers, ledgerStream.Msg().GetData()...)
		}
		require.NoError(t, ledgerStream.Err())
		assert.GreaterOrEqual(t, len(foundLedgers), 2, "Should find at least 2 ledgers")

		// Test SearchAccounts
		accountStream, err := client.SearchAccounts(ctx, connect.NewRequest(&commonv1.SearchRequest{Query: "{}"}))
		require.NoError(t, err)
		var foundAccounts []*ledgerv1.Account
		for accountStream.Receive() {
			foundAccounts = append(foundAccounts, accountStream.Msg().GetData()...)
		}
		require.NoError(t, accountStream.Err())
		assert.GreaterOrEqual(t, len(foundAccounts), 2, "Should find at least 2 accounts")

		// Test SearchTransactions
		txnStream, err := client.SearchTransactions(ctx, connect.NewRequest(&commonv1.SearchRequest{Query: "{}"}))
		require.NoError(t, err)
		var foundTxns []*ledgerv1.Transaction
		for txnStream.Receive() {
			foundTxns = append(foundTxns, txnStream.Msg().GetData()...)
		}
		require.NoError(t, txnStream.Err())
		assert.GreaterOrEqual(t, len(foundTxns), 1, "Should find at least 1 transaction")

		// Test SearchTransactionEntries
		entryStream, err := client.SearchTransactionEntries(
			ctx,
			connect.NewRequest(&commonv1.SearchRequest{Query: "{}"}),
		)
		require.NoError(t, err)
		var foundEntries []*ledgerv1.TransactionEntry
		for entryStream.Receive() {
			foundEntries = append(foundEntries, entryStream.Msg().GetData()...)
		}
		require.NoError(t, entryStream.Err())
		assert.GreaterOrEqual(t, len(foundEntries), 2, "Should find at least 2 entries")
	})
}

func TestLedgerHandlersSuite(t *testing.T) {
	suite.Run(t, &LedgerHandlersTestSuite{})
}
