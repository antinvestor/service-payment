package events

import (
	"context"
	"errors"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/frame"
	"google.golang.org/protobuf/proto"
)

type PaymentOutQueue struct {
	Service *frame.Service
}

func (event *PaymentOutQueue) Name() string {
	return "payment.out.queue"
}

func (event *PaymentOutQueue) PayloadType() any {
	pType := ""
	return &pType
}

func (event *PaymentOutQueue) Validate(ctx context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New("payload is not of type string")
	}
	return nil
}

func (event *PaymentOutQueue) Execute(ctx context.Context, payload any) error {
	paymentID := *payload.(*string)

	logger := event.Service.Log(ctx).WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling payment event")

	// Fetch payment record by ID
	paymentRepo := repository.NewPaymentRepository(ctx, event.Service)
	payment, err := paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Fetch payment status
	statusRepo := repository.NewStatusRepository(ctx, event.Service)
	status, err := statusRepo.GetByEntity(ctx, payment.ID, "payment")
	if err != nil {
		logger.WithError(err).WithField("status_id", payment.ID).Warn("could not get payment status")
		return err
	}

	paymentMap := event.formatOutboundPayment(payment)

	apiPayment := payment.ToApi(status, paymentMap)

	binaryProto, err := proto.Marshal(apiPayment)
	if err != nil {
		return err
	}

	// Publish the payment message for further processing
	err = event.Service.Publish(ctx, payment.RouteID, binaryProto)
	if err != nil {
		return err
	}

	logger.WithField("payment_id", payment.GetID()).
		WithField("route", payment.RouteID).
		Debug("Payment message successfully queued")

	// Save payment status
	err = paymentRepo.Save(ctx, payment)
	if err != nil {
		return err
	}

	// Update payment status using unified Status
	status = &models.Status{
		EntityID:   payment.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_ACTIVE),
		Status:     int32(commonv1.STATUS_IN_PROCESS),
		Extra:      make(map[string]interface{}),
	}
	status.GenID(ctx)

	// Emit status event
	statusEvent := StatusSave{Service: event.Service}
	err = event.Service.Emit(ctx, statusEvent.Name(), status)
	if err != nil {
		return err
	}

	return nil
}

func (event *PaymentOutQueue) formatOutboundPayment(p *models.Payment) map[string]string {
	paymentMap := make(map[string]string)
	paymentMap["id"] = p.GetID()
	paymentMap["created_at"] = p.CreatedAt.Format(time.RFC3339Nano)
	if p.Amount.Valid {
		paymentMap["amount"] = p.Amount.Decimal.String()
	} else {
		paymentMap["amount"] = "0"
	}
	paymentMap["currency"] = p.Currency

	return paymentMap
}
