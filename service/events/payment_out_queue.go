package events

import (
	"context"
	"errors"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/antinvestor/service-payments-v1/service/repository"
	"github.com/pitabwire/frame"
	"github.com/sirupsen/logrus"
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

	logger := event.Service.L().WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling payment event")

	// Fetch payment record by ID
	paymentRepo := repository.NewPaymentRepository(ctx, event.Service)
	payment, err := paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Fetch payment status
	paymentStatusRepo := repository.NewPaymentStatusRepository(ctx, event.Service)
	pStatus, err := paymentStatusRepo.GetByID(ctx, payment.ID)
	if err != nil {
		logger.WithError(err).WithField("status_id", pStatus.PaymentID).Warn("could not get payment status")
		return err
	}
	
	paymentMap, err := event.formatOutboundPayment(ctx, logger, payment)
	if err != nil {
		return err
	}

	apiPayment := payment.ToApi(pStatus, paymentMap)

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

	// Update payment status
	pStatus = &models.PaymentStatus{
		PaymentID: payment.GetID(),
		State:     int32(commonv1.STATE_ACTIVE),
		Status:    int32(commonv1.STATUS_IN_PROCESS),
	}

	pStatus.GenID(ctx)

	// Emit payment status event
	eventStatus := PaymentStatusSave{}
	err = event.Service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		return err
	}

	return nil
}

func (event *PaymentOutQueue) formatOutboundPayment(ctx context.Context, logger *logrus.Entry, p *models.Payment) (map[string]string, error) {

	paymentMap := make(map[string]string)
	paymentMap["id"] = p.GetID()
	paymentMap["created_at"] = p.CreatedAt.Format(time.RFC3339Nano)
	if p.Amount.Valid {
		paymentMap["amount"] = p.Amount.Decimal.String()
	} else {
		paymentMap["amount"] = "0"
	}
	paymentMap["currency"] = p.Currency

	return paymentMap, nil
 }
 
