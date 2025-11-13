package handlers

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"github.com/antinvestor/apis/go/ledger"
	"github.com/antinvestor/apis/go/partition"
	"github.com/antinvestor/apis/go/profile"

	"github.com/antinvestor/service-payments/service/business"
	"github.com/pitabwire/frame"
)

type PaymentServer struct {
	Service         *frame.Service
	PaymentBusiness business.PaymentBusiness
	ProfileCli      profile.Client
	PartitionCli    partition.Client
	LedgerCli       ledger.Client

	paymentv1connect.UnimplementedPaymentServiceHandler
}

func (ps *PaymentServer) Send(ctx context.Context, req *paymentv1.SendRequest) (*paymentv1.SendResponse, error) {
	response, err := ps.PaymentBusiness.Send(ctx, req.GetData())
	if err != nil {
		return nil, err
	}
	return &paymentv1.SendResponse{Data: response}, nil
}

func (ps *PaymentServer) Status(ctx context.Context, req *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	return ps.PaymentBusiness.Status(ctx, req)
}

// StatusUpdate request to allow continuation of payment processing.
func (ps *PaymentServer) StatusUpdate(
	ctx context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusUpdateResponse, error) {
	response, err := ps.PaymentBusiness.StatusUpdate(ctx, req)
	if err != nil {
		return nil, err
	}

	return &commonv1.StatusUpdateResponse{Data: response}, nil
}

// Release method for releasing queued payments and returns if payment status if released.
func (ps *PaymentServer) Release(
	ctx context.Context,
	req *paymentv1.ReleaseRequest,
) (*paymentv1.ReleaseResponse, error) {
	response, err := ps.PaymentBusiness.Release(ctx, req)

	if err != nil {
		return nil, err
	}

	return &paymentv1.ReleaseResponse{Data: response}, nil
}

// Receive method is for client request for particular Payment responses from system.
func (ps *PaymentServer) Receive(
	ctx context.Context,
	req *paymentv1.ReceiveRequest,
) (*paymentv1.ReceiveResponse, error) {

	response, err := ps.PaymentBusiness.Receive(ctx, req.GetData())

	if err != nil {
		return nil, err
	}

	return &paymentv1.ReceiveResponse{Data: response}, nil
}

// InitiatePrompt method for client request for particular Prompt responses from system.
func (ps *PaymentServer) InitiatePrompt(
	ctx context.Context,
	req *paymentv1.InitiatePromptRequest,
) (*paymentv1.InitiatePromptResponse, error) {
	response, err := ps.PaymentBusiness.InitiatePrompt(ctx, req)

	if err != nil {
		return nil, err
	}

	return &paymentv1.InitiatePromptResponse{Data: response}, nil
}

// CreatePaymentLink method for client request to create a payment link.
func (ps *PaymentServer) CreatePaymentLink(
	ctx context.Context,
	req *paymentv1.CreatePaymentLinkRequest,
) (*paymentv1.CreatePaymentLinkResponse, error) {
	response, err := ps.PaymentBusiness.CreatePaymentLink(ctx, req)

	if err != nil {
		return nil, err
	}

	return &paymentv1.CreatePaymentLinkResponse{Data: response}, nil
}
