package events

import (
	"context"
	"errors"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/antinvestor/service-payments-v1/service/repository"

	"strings"

	"github.com/pitabwire/frame"
)

type PaymentInQueue struct {
	Service    *frame.Service
	ProfileCli *profileV1.ProfileClient
}

func (event *PaymentInQueue) Name() string {
	return "payment.in.queue"
}

func (event *PaymentInQueue) PayloadType() any {
	pType := ""
	return &pType
}

func (event *PaymentInQueue) Validate(ctx context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New(" payload is not of type string")
	}

	return nil
}

func (event *PaymentInQueue) Execute(ctx context.Context, payload any) error {
	paymentID := *payload.(*string)
	logger := event.Service.L(ctx).WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling event")

	paymentRepo := repository.NewPaymentRepository(ctx, event.Service)

	p, err := paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Queue a payment for further processing by peripheral services
	err = event.Service.Publish(ctx, p.RouteID, p)
	if err != nil {

		if !strings.Contains(err.Error(), "reference does not exist") {

			if p.RouteID != "" {
				_, err = loadRoute(ctx, event.Service, p.RouteID)
				if err != nil {
					return err
				}
			}

			return err
		}
	}

	logger.
		WithField("payment", p.ID).
		WithField("route", p.RouteID).
		Debug(" Successfully routed in payment")

	pStatus := models.PaymentStatus{
		PaymentID: p.GetID(),
		State:          int32(commonv1.STATE_ACTIVE),
		Status:         int32(commonv1.STATUS_IN_PROCESS),
	}

	pStatus.GenID(ctx)

	// Queue out payment status for further processing
	eventStatus := PaymentStatusSave{}
	err = event.Service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		return err
	}

	return nil
}
