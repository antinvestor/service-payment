package handlers

import (
	"context"
	partitionv1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments-v1/service/business"
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
	response, err := paymentBusiness.Dispatch(ctx, req.GetData())
	if err != nil {
		return nil, err
	}
	return &paymentV1.SendResponse{Data: response}, nil
}
