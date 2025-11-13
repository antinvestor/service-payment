package events

import (
	"context"
	"errors"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"github.com/antinvestor/apis/go/profile"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"

	"strings"
)

type PaymentInQueue struct {
	qMan        queue.Manager
	eventMan    events.Manager
	paymentRepo repository.PaymentRepository
	ProfileCli  *profile.Client
}

func (event *PaymentInQueue) Name() string {
	return "payment.in.queue"
}

func (event *PaymentInQueue) PayloadType() any {
	pType := ""
	return &pType
}

func (event *PaymentInQueue) Validate(_ context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New(" payload is not of type string")
	}

	return nil
}

func (event *PaymentInQueue) Execute(ctx context.Context, payload any) error {
	paymentIDPtr, ok := payload.(*string)
	if !ok {
		return errors.New("payload is not of type *string")
	}
	paymentID := *paymentIDPtr
	logger := util.Log(ctx).WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling event")

	p, err := event.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Queue a payment for further processing by peripheral services
	err = event.qMan.Publish(ctx, p.RouteID, p)
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

	// Unified status
	status := models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_ACTIVE),
		Status:     int32(commonv1.STATUS_IN_PROCESS),
		Extra:      make(map[string]interface{}),
	}
	status.GenID(ctx)

	// Queue out payment status for further processing
	err = event.eventMan.Emit(ctx, EventNameStatusSave, &status)
	if err != nil {
		return err
	}

	return nil
}
