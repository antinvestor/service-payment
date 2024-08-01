package business

import (
	"context"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments-v1/service/events"
	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
)

type PaymentBusiness interface {
	Dispatch(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
	//QueueIn(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
}

func NewPaymentBusiness(_ context.Context, service *frame.Service, profileCli *profileV1.ProfileClient, partitionCli *partitionV1.PartitionClient) (PaymentBusiness, error) {
	//initialize the service
	if service == nil || profileCli == nil || partitionCli == nil {
		return nil, ErrorInitializationFail
	}
	return &paymentBusiness{
		service:      service,
		profileCli:   profileCli,
		partitionCli: partitionCli,
	}, nil
}

type paymentBusiness struct {
	service      *frame.Service
	profileCli   *profileV1.ProfileClient
	partitionCli *partitionV1.PartitionClient
}



func (pb *paymentBusiness) Dispatch(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
	logger := pb.service.L().WithField("request", message)
	//authClaim := frame.ClaimsFromContext(ctx)
	//logger.WithField("auth claim", authClaim).Info("handling send request")

	p := &models.Payment{
		SenderProfileType:     message.GetSource().GetProfileType(),
		SenderProfileID:       message.GetSource().GetProfileId(),
		SenderContactID:       message.GetSource().GetContactId(),
		RecipientProfileType:  message.GetRecipient().GetProfileType(),
		RecipientProfileID:    message.GetRecipient().GetProfileId(),
		RecipientContactID:    message.GetRecipient().GetContactId(),
		ReferenceId:           message.GetReferenceId(),
		BatchId:               message.GetBatchId(),
		ExternalTransactionId: message.GetExternalTransactionId(),
		Route:                 message.GetRoute(),
		Source:                message.GetSource(),
		Recipient:             message.GetRecipient(),
		State:                 message.GetState(),
		Status:                message.GetStatus(),
		Outbound:              message.GetOutbound(),
	}

	if message.GetId() == "" {
		p.GenID(ctx)
	}


	if p.ValidXID(message.GetId()) {
		p.Id = message.GetId()
	}

	// Validate and set amount
	if message.GetAmount().Units == 0 || message.GetAmount().CurrencyCode == "" {
		logger.Error("amount or cost is missing")
		return nil, ErrorPaymentDoesNotExist
	}
	p.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetAmount().Units)),
	}
	p.Currency = message.GetAmount().CurrencyCode

	// Validate and set cost
	if message.GetCost().Units == 0 || message.GetCost().CurrencyCode == "" {
		logger.Error("amount or cost is missing")
		return nil, ErrorPaymentDoesNotExist
	}
	p.Cost = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
	}



	pStatus := models.PaymentStatus{
		PaymentID: p.Id,
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	if err := pb.emitPaymentEvent(ctx, p); err != nil {
		return nil, err
	}

	if err := pb.emitPaymentStatusEvent(ctx, pStatus); err != nil {
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) emitPaymentEvent(ctx context.Context, p *models.Payment) error {
	event := events.PaymentSave{}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.L().WithError(err).Warn("could not emit payment event")
		return err
	}
	return nil
}

func (pb *paymentBusiness) emitPaymentStatusEvent(ctx context.Context, pStatus models.PaymentStatus) error {
	eventStatus := events.PaymentStatusSave{}
	if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
		pb.service.L().WithError(err).Warn("could not emit payment status event")
		return err
	}
	return nil
}
