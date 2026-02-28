package handlers

import (
	"context"
	"errors"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/ledger/v1/ledgerv1connect"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/ledger/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/service/business"
	"github.com/antinvestor/service-payments/internal/apperrors"
	"github.com/pitabwire/frame/security/authorizer"
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

type LedgerServer struct {
	Ledger      business.LedgerBusiness
	Account     business.AccountBusiness
	Transaction business.TransactionBusiness
	authz       authz.Middleware
}

// NewLedgerServer creates a new LedgerServer with injected dependencies.
func NewLedgerServer(
	ledgerBusiness business.LedgerBusiness,
	accountBusiness business.AccountBusiness,
	transactionBusiness business.TransactionBusiness,
	authzMiddleware authz.Middleware,
) ledgerv1connect.LedgerServiceHandler {
	return &LedgerServer{
		Ledger:      ledgerBusiness,
		Account:     accountBusiness,
		Transaction: transactionBusiness,
		authz:       authzMiddleware,
	}
}

// SearchLedgers finds ledgers in the chart of accounts.
// Supports filtering by type, parent, and custom properties.
func (ledgerSrv *LedgerServer) SearchLedgers(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[ledgerv1.SearchLedgersResponse],
) error {
	if err := ledgerSrv.authz.CanViewLedger(ctx); err != nil {
		return authorizer.ToConnectError(err)
	}

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
	if err := ledgerSrv.authz.CanManageLedger(ctx); err != nil {
		return nil, authorizer.ToConnectError(err)
	}

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
	if err := ledgerSrv.authz.CanManageLedger(ctx); err != nil {
		return nil, authorizer.ToConnectError(err)
	}

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
