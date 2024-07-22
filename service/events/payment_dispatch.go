package events

import (
	"context"
	"errors"
	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/antinvestor/service-payments-v1/service/repository"
	"github.com/pitabwire/frame"
	"google.golang.org/protobuf/proto"
)

type PaymentDispatch struct {
	Service    *frame.Service
	PaymentCli *paymentV1.PaymentClient
}

func (event *PaymentDispatch) Name() string {
	return "payment.dispatch"
}

func (event *PaymentDispatch) PayloadType() any {
	pType := &paymentV1.Payment{}
	return &pType
}

func (event *PaymentDispatch) Validate(ctx context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New(" payload is not of type string")
	}

	return nil
}

func (event *PaymentDispatch) Execute(ctx context.Context, payload any) error {
	paymentID := *payload.(*string)

	logger := event.Service.L().WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling event")

	paymentRepo := repository.NewPaymentRepository(ctx, event.Service)
	payment, err := paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	paymentStatusRepo := repository.NewPaymentStatusRepository(ctx, event.Service)
	paymentStatus, err := paymentStatusRepo.GetByID(ctx, payment.StatusID)
	if err != nil {
		logger.WithError(err).WithField("status_id", payment.StatusID).Warn("could not get status")
		return err
	}

	apiPayment := payment.ToApi(payment)

	binaryProto, err := proto.Marshal(apiPayment)
	if err != nil {
		return err
	}

	err = event.Service.Publish(ctx, payment.RouteID, binaryProto)
	if err != nil {
		return err
	}

	logger.WithField("payment_id", payment.GetID()).
		WithField("route", payment.RouteID).
		Debug("Successfully queued payment dispatch")

	err = paymentRepo.Save(ctx, payment)
	if err != nil {
		return err
	}

	paymentStatus = &models.PaymentStatus{
		PaymentID: payment.GetID(),
		State:     int32(commonv1.STATE_ACTIVE),
		Status:    int32(commonv1.STATUS_IN_PROCESS),
	}

	paymentStatus.GenID(ctx)

	eventStatus := PaymentStatusSave{}
	err = event.Service.Emit(ctx, eventStatus.Name(), paymentStatus)
	if err != nil {
		return err
	}

	return nil
}
