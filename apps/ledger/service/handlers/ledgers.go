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

package handlers

import (
	"context"
	"errors"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v1/ledgerv1connect"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/ledger/service/business"
	"github.com/antinvestor/service-payments/pkg/apperrors"
)

// ToConnectError translates application errors into appropriate ConnectRPC
// error codes so clients receive meaningful status codes instead of generic
// Internal errors.
func ToConnectError(err error) error {
	if err == nil {
		return nil
	}

	var appErr apperrors.ApplicationError
	if !errors.As(err, &appErr) {
		return err
	}

	code := appErr.ErrorCode()

	switch code - apperrors.DefaultCodeOffset {
	case apperrors.ErrorCodeUnspecifiedID,
		apperrors.ErrorCodeUnspecifiedReference,
		apperrors.ErrorCodeBadDataSupplied,
		apperrors.ErrorCodeTransactionEntryHasZeroAmount,
		apperrors.ErrorCodeTransactionHasNonZeroSum,
		apperrors.ErrorCodeTransactionHasInvalidDrCrEntry,
		apperrors.ErrorCodeTransactionAccountsDifferCurrency,
		apperrors.ErrorCodeSearchQueryHasInvalidFormat,
		apperrors.ErrorCodeSearchQueryHasInvalidKeys:
		return connect.NewError(connect.CodeInvalidArgument, appErr)

	case apperrors.ErrorCodeLedgerNotFound,
		apperrors.ErrorCodeAccountNotFound,
		apperrors.ErrorCodeAccountsNotFound,
		apperrors.ErrorCodeTransactionNotFound,
		apperrors.ErrorCodeTransactionEntriesNotFound,
		apperrors.ErrorCodeSearchNamespaceUnknown:
		return connect.NewError(connect.CodeNotFound, appErr)

	case apperrors.ErrorCodeAccountWithReferenceExists,
		apperrors.ErrorCodeTransactionAlreadyExists,
		apperrors.ErrorCodeTransactionIsConflicting:
		return connect.NewError(connect.CodeAlreadyExists, appErr)

	case apperrors.ErrorCodeTransactionTypeNotReversible:
		return connect.NewError(connect.CodeFailedPrecondition, appErr)

	case apperrors.ErrorCodeAccountsCurrencyUnknown:
		return connect.NewError(connect.CodeInvalidArgument, appErr)

	default:
		return connect.NewError(connect.CodeInternal, appErr)
	}
}

// LedgerServer is the ConnectRPC LedgerService implementation. It is a
// thin adapter that translates Connect request envelopes into business-
// layer calls and shapes the responses back into proto. Every domain rule
// lives in the business layer; handlers only enforce request validation
// and error translation.
type LedgerServer struct {
	Ledger      business.LedgerBusiness
	Account     business.AccountBusiness
	Transaction business.TransactionBusiness
	Report      business.ReportBusiness
	Book        business.BookBusiness
}

// NewLedgerServer creates a new LedgerServer with injected dependencies.
// All five business interfaces are required — Report and Book power the
// trial-balance, account-statement and book-CRUD RPCs respectively.
func NewLedgerServer(
	ledgerBusiness business.LedgerBusiness,
	accountBusiness business.AccountBusiness,
	transactionBusiness business.TransactionBusiness,
	reportBusiness business.ReportBusiness,
	bookBusiness business.BookBusiness,
) ledgerv1connect.LedgerServiceHandler {
	return &LedgerServer{
		Ledger:      ledgerBusiness,
		Account:     accountBusiness,
		Transaction: transactionBusiness,
		Report:      reportBusiness,
		Book:        bookBusiness,
	}
}

// SearchLedgers finds ledgers in the chart of accounts.
// Supports filtering by type, parent, and custom properties.
func (ledgerSrv *LedgerServer) SearchLedgers(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[ledgerv1.SearchLedgersResponse],
) error {
	// Search ledgers using business layer
	return ToConnectError(
		ledgerSrv.Ledger.SearchLedgers(ctx, req.Msg, func(_ context.Context, batch []*ledgerv1.Ledger) error {
			return stream.Send(&ledgerv1.SearchLedgersResponse{
				Data: batch,
			})
		}),
	)
}

// CreateLedger creates a new ledger in the chart of accounts.
// Ledgers can be hierarchical with parent-child relationships.
func (ledgerSrv *LedgerServer) CreateLedger(
	ctx context.Context,
	req *connect.Request[ledgerv1.CreateLedgerRequest],
) (*connect.Response[ledgerv1.CreateLedgerResponse], error) {
	// Create the ledger using business layer
	createdLedger, err := ledgerSrv.Ledger.CreateLedger(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}

	response := &ledgerv1.CreateLedgerResponse{
		Data: createdLedger,
	}

	return connect.NewResponse(response), nil
}

// UpdateLedger updates an existing ledger's metadata.
// The ledger type and reference cannot be changed.
func (ledgerSrv *LedgerServer) UpdateLedger(
	ctx context.Context,
	req *connect.Request[ledgerv1.UpdateLedgerRequest],
) (*connect.Response[ledgerv1.UpdateLedgerResponse], error) {
	// Update the ledger using business layer
	updatedLedger, err := ledgerSrv.Ledger.UpdateLedger(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}

	response := &ledgerv1.UpdateLedgerResponse{
		Data: updatedLedger,
	}

	return connect.NewResponse(response), nil
}
