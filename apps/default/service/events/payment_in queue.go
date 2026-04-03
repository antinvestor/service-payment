package events

import (
	"context"
	"errors"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
)

type PaymentInQueue struct {
	qMan        queue.Manager
	eventMan    events.Manager
	paymentRepo repository.PaymentRepository
	routeRepo   repository.RouteRepository
	profileCli  profilev1connect.ProfileServiceClient
}

// NewPaymentInQueue creates a new PaymentInQueue event handler with the required dependencies.
func NewPaymentInQueue(
	qMan queue.Manager,
	eventMan events.Manager,
	paymentRepo repository.PaymentRepository,
	routeRepo repository.RouteRepository,
	profileCli profilev1connect.ProfileServiceClient,
) *PaymentInQueue {
	return &PaymentInQueue{
		qMan:        qMan,
		eventMan:    eventMan,
		paymentRepo: paymentRepo,
		routeRepo:   routeRepo,
		profileCli:  profileCli,
	}
}

func (event *PaymentInQueue) Name() string {
	return EventNamePaymentInQueue
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
	logger := util.Log(ctx).WithFields(map[string]any{"payment_id": paymentID, "type": event.Name()})
	logger.Debug("handling event")

	p, err := event.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Queue a payment for further processing by peripheral services
	err = event.qMan.Publish(ctx, p.RouteID, p)
	if err != nil {
		return event.handlePublishError(ctx, err, p)
	}

	logger.WithField("route_id", p.RouteID).Debug("successfully routed in payment")

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

// handlePublishError handles errors from queue publish operation.
func (event *PaymentInQueue) handlePublishError(ctx context.Context, err error, p *models.Payment) error {
	// If the error is not about missing reference, return it
	if !strings.Contains(err.Error(), "reference does not exist") {
		return err
	}

	// Try to load the route if route ID is set
	if p.RouteID != "" {
		if _, loadErr := loadRoute(ctx, event.qMan, event.routeRepo, p.RouteID); loadErr != nil {
			return loadErr
		}
	}

	return err
}
