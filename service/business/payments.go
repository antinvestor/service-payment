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

	authClaim := frame.ClaimsFromContext(ctx)

	logger.WithField("auth claim", authClaim).Info("handling send request")

	p := &models.Payment{

		SenderProfileType: message.GetSource().GetProfileType(),
		SenderProfileID:   message.GetSource().GetProfileId(),
		SenderContactID:   message.GetSource().GetContactId(),

		RecipientProfileType: message.GetRecipient().GetProfileType(),
		RecipientProfileID:   message.GetRecipient().GetProfileId(),
		RecipientContactID:   message.GetRecipient().GetContactId(),

		ReferenceId:           message.GetReferenceId(),
		BatchId:               message.GetBatchId(),
		ExternalTransactionId: message.GetExternalTransactionId(),
		Route:                 message.GetRoute(),
		Cost:                  message.GetCost(),
		Amount:                message.GetAmount(),
		Source:                message.GetSource(),
		Recipient:             message.GetRecipient(),
		State:                 message.GetState(),
		Status:                message.GetStatus(),
		Outbound:              message.GetOutbound(),
	}
	p.GenID(ctx)

	if p.ValidXID(message.GetId()) {
		p.Id = message.GetId()
	}

	if p.Amount == nil || p.Cost == nil {
		logger.Error("amount or cost is missing")
		return nil, ErrorPaymentDoesNotExist
	}

	if p.Amount.GetCurrencyCode() != p.Cost.GetCurrencyCode() {
		logger.Error("currency code mismatch")
		return nil, ErrorInvalidPaymentRequest
	}

	pStatus := models.PaymentStatus{
		PaymentID: p.GetID(),
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	//dispatch payment event
	event := events.PaymentSave{}
	err := pb.service.Emit(ctx, event.Name(), p)
	if err != nil {
		logger.WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	//dispatch payment status event
	eventStatus := events.PaymentStatusSave{}
	err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Warn("could not emit payment status event")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}
