package handlers

import (
	"context"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionv1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/business"
	"github.com/pitabwire/frame"
)

type PaymentServer struct {
	Service      *frame.Service
	ProfileCli   *profileV1.ProfileClient
	PartitionCli *partitionv1.PartitionClient

	paymentV1.UnimplementedPaymentServiceServer
}

func (ps *PaymentServer) newPaymentBusiness(ctx context.Context) (business.PaymentBusiness, error) {
	return business.NewPaymentBusiness(ctx, ps.Service, ps.ProfileCli, ps.PartitionCli)
}

func (ps *PaymentServer) Send(ctx context.Context, req *paymentV1.SendRequest) (*paymentV1.SendResponse, error) {
	paymentBusiness, err := ps.newPaymentBusiness(ctx)
	if err != nil {
		return nil, err
	}
	response, err := paymentBusiness.Send(ctx, req.GetData())
	if err != nil {
		return nil, err
	}
	return &paymentV1.SendResponse{Data: response}, nil
}

func (ps *PaymentServer) Status(ctx context.Context, req *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	paymentBusiness, err := ps.newPaymentBusiness(ctx)
	if err != nil {
		return nil, err
	}
	return paymentBusiness.Status(ctx, req)
}

// StatusUpdate request to allow continuation of payment processing.
func (ps *PaymentServer) StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusUpdateResponse, error) {
	paymentBusiness, err := ps.newPaymentBusiness(ctx)
	if err != nil {
		return nil, err
	}
	response, err := paymentBusiness.StatusUpdate(ctx, req)
	if err != nil {
		return nil, err
	}

	return &commonv1.StatusUpdateResponse{Data: response}, nil
}

// Release method for releasing queued payments and returns if payment status if released.
func (ps *PaymentServer) Release(ctx context.Context, req *paymentV1.ReleaseRequest) (*paymentV1.ReleaseResponse, error) {
	paymentBusiness, err := ps.newPaymentBusiness(ctx)
	if err != nil {
		return nil, err
	}
	response, err := paymentBusiness.Release(ctx, req)

	if err != nil {
		return nil, err
	}

	return &paymentV1.ReleaseResponse{Data: response}, nil
}

// Receive method is for client request for particular Payment responses from system.
func (ps *PaymentServer) Receive(ctx context.Context, req *paymentV1.ReceiveRequest) (*paymentV1.ReceiveResponse, error) {
	paymentBusiness, err := ps.newPaymentBusiness(ctx)
	if err != nil {
		return nil, err
	}
	response, err := paymentBusiness.Receive(ctx, req.GetData())

	if err != nil {
		return nil, err
	}

	return &paymentV1.ReceiveResponse{Data: response}, nil
}

// InitiatePrompt method for client request for particular Prompt responses from system.
func (ps *PaymentServer) InitiatePrompt(ctx context.Context, req *paymentV1.InitiatePromptRequest) (*paymentV1.InitiatePromptResponse, error) {
	paymentBusiness, err := ps.newPaymentBusiness(ctx)
	if err != nil {
		return nil, err
	}
	response, err := paymentBusiness.InitiatePrompt(ctx, req)

	if err != nil {
		return nil, err
	}

	return &paymentV1.InitiatePromptResponse{Data: response}, nil
}




