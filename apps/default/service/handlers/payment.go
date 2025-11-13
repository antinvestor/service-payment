package handlers

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/ledger/v1/ledgerv1connect"
	"buf.build/gen/go/antinvestor/partition/connectrpc/go/partition/v1/partitionv1connect"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/default/service/business"
)

var _ paymentv1connect.PaymentServiceHandler = (*PaymentServer)(nil)

type PaymentServer struct {
	PaymentBusiness business.PaymentBusiness
	ProfileCli      profilev1connect.ProfileServiceClient
	LedgerCli       ledgerv1connect.LedgerServiceClient
	PartitionCli    partitionv1connect.PartitionServiceClient
}

// NewPaymentServer creates a new PaymentServer with the required dependencies.
func NewPaymentServer(
	paymentBusiness business.PaymentBusiness,
	profileCli profilev1connect.ProfileServiceClient,
	ledgerCli ledgerv1connect.LedgerServiceClient,
	partitionCli partitionv1connect.PartitionServiceClient,
) *PaymentServer {
	return &PaymentServer{
		PaymentBusiness: paymentBusiness,
		ProfileCli:      profileCli,
		PartitionCli:    partitionCli,
		LedgerCli:       ledgerCli,
	}
}

func (ps *PaymentServer) Send(
	ctx context.Context,
	req *connect.Request[paymentv1.SendRequest],
) (*connect.Response[paymentv1.SendResponse], error) {
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
	response, err := ps.PaymentBusiness.Reconcile(ctx, req.Msg)

	if err != nil {
		return nil, err
	}

	return connect.NewResponse(response), nil
}
