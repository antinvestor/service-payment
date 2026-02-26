package handlers

import (
	"context"
	"errors"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/ledger/v1/ledgerv1connect"
	"buf.build/gen/go/antinvestor/partition/connectrpc/go/partition/v1/partitionv1connect"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/default/service/authz"
	"github.com/antinvestor/service-payments/apps/default/service/business"
	"github.com/pitabwire/frame/security/authorizer"
)

var _ paymentv1connect.PaymentServiceHandler = (*PaymentServer)(nil)

type PaymentServer struct {
	authz           authz.Middleware
	PaymentBusiness business.PaymentBusiness
	ProfileCli      profilev1connect.ProfileServiceClient
	LedgerCli       ledgerv1connect.LedgerServiceClient
	PartitionCli    partitionv1connect.PartitionServiceClient
}

// NewPaymentServer creates a new PaymentServer with the required dependencies.
func NewPaymentServer(
	authzMiddleware authz.Middleware,
	paymentBusiness business.PaymentBusiness,
	profileCli profilev1connect.ProfileServiceClient,
	ledgerCli ledgerv1connect.LedgerServiceClient,
	partitionCli partitionv1connect.PartitionServiceClient,
) *PaymentServer {
	return &PaymentServer{
		authz:           authzMiddleware,
		PaymentBusiness: paymentBusiness,
		ProfileCli:      profileCli,
		PartitionCli:    partitionCli,
		LedgerCli:       ledgerCli,
	}
}

// toConnectError translates authorisation errors into ConnectRPC error codes.
func toConnectError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, authorizer.ErrInvalidSubject) || errors.Is(err, authorizer.ErrInvalidObject) {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	var permErr *authorizer.PermissionDeniedError
	if errors.As(err, &permErr) {
		return connect.NewError(connect.CodePermissionDenied, err)
	}

	return connect.NewError(connect.CodeInternal, err)
}

func (ps *PaymentServer) Send(
	ctx context.Context,
	req *connect.Request[paymentv1.SendRequest],
) (*connect.Response[paymentv1.SendResponse], error) {
	if err := ps.authz.CanSendPayment(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.Send(ctx, req.Msg.GetData())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&paymentv1.SendResponse{Data: response}), nil
}

func (ps *PaymentServer) Status(
	ctx context.Context,
	req *connect.Request[commonv1.StatusRequest],
) (*connect.Response[commonv1.StatusResponse], error) {
	if err := ps.authz.CanViewPaymentStatus(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.Status(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil
}

// StatusUpdate request to allow continuation of payment processing.
func (ps *PaymentServer) StatusUpdate(
	ctx context.Context,
	req *connect.Request[commonv1.StatusUpdateRequest],
) (*connect.Response[commonv1.StatusUpdateResponse], error) {
	if err := ps.authz.CanUpdatePaymentStatus(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.StatusUpdate(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&commonv1.StatusUpdateResponse{Data: response}), nil
}

// Release method for releasing queued payments and returns if payment status if released.
func (ps *PaymentServer) Release(
	ctx context.Context,
	req *connect.Request[paymentv1.ReleaseRequest],
) (*connect.Response[paymentv1.ReleaseResponse], error) {
	if err := ps.authz.CanReleasePayment(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.Release(ctx, req.Msg)

	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&paymentv1.ReleaseResponse{Data: response}), nil
}

// Receive method is for client request for particular Payment responses from system.
func (ps *PaymentServer) Receive(
	ctx context.Context,
	req *connect.Request[paymentv1.ReceiveRequest],
) (*connect.Response[paymentv1.ReceiveResponse], error) {
	if err := ps.authz.CanReceivePayment(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.Receive(ctx, req.Msg.GetData())

	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&paymentv1.ReceiveResponse{Data: response}), nil
}

// InitiatePrompt method for client request for particular Prompt responses from system.
func (ps *PaymentServer) InitiatePrompt(
	ctx context.Context,
	req *connect.Request[paymentv1.InitiatePromptRequest],
) (*connect.Response[paymentv1.InitiatePromptResponse], error) {
	if err := ps.authz.CanInitiatePrompt(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.InitiatePrompt(ctx, req.Msg)

	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&paymentv1.InitiatePromptResponse{Data: response}), nil
}

// CreatePaymentLink method for client request to create a payment link.
func (ps *PaymentServer) CreatePaymentLink(
	ctx context.Context,
	req *connect.Request[paymentv1.CreatePaymentLinkRequest],
) (*connect.Response[paymentv1.CreatePaymentLinkResponse], error) {
	if err := ps.authz.CanCreatePaymentLink(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.CreatePaymentLink(ctx, req.Msg)

	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&paymentv1.CreatePaymentLinkResponse{Data: response}), nil
}

// Search method for searching payments based on query criteria.
func (ps *PaymentServer) Search(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[paymentv1.SearchResponse],
) error {
	if err := ps.authz.CanSearchPayments(ctx); err != nil {
		return toConnectError(err)
	}
	resultPipe, err := ps.PaymentBusiness.Search(ctx, req.Msg)
	if err != nil {
		return err
	}

	// Read from the result pipe and stream back to client
	for {
		result, ok := resultPipe.ReadResult(ctx)
		if !ok {
			return nil
		}

		if result.IsError() {
			return result.Error()
		}

		// Stream each payment back to the client
		err = stream.Send(&paymentv1.SearchResponse{Data: result.Item()})
		if err != nil {
			return err
		}
	}
}

func (ps *PaymentServer) Reconcile(
	ctx context.Context,
	req *connect.Request[paymentv1.ReconcileRequest],
) (*connect.Response[paymentv1.ReconcileResponse], error) {
	if err := ps.authz.CanReconcile(ctx); err != nil {
		return nil, toConnectError(err)
	}
	response, err := ps.PaymentBusiness.Reconcile(ctx, req.Msg)

	if err != nil {
		return nil, err
	}

	return connect.NewResponse(response), nil
}
