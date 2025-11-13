package events

import (
	"context"
	"errors"
	"fmt"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
)

// filterContactFromProfileByID finds a contact by ID in a profile
// Currently unused but may be needed for future functionality
/*
func filterContactFromProfileByID(profile *profilev1.ProfileObject, contactID string) *profilev1.ContactObject {
	for _, contact := range profile.GetContacts() {
		if contact.GetId() == contactID {
			return contact
		}
	}

	return nil
}
*/

type PaymentInRoute struct {
	eventMan    events.Manager
	paymentRepo repository.PaymentRepository
	routeRepo   repository.RouteRepository
}

// NewPaymentInRoute creates a new PaymentInRoute event handler with the required dependencies
func NewPaymentInRoute(eventMan events.Manager, paymentRepo repository.PaymentRepository, routeRepo repository.RouteRepository) *PaymentInRoute {
	return &PaymentInRoute{
		eventMan:    eventMan,
		paymentRepo: paymentRepo,
		routeRepo:   routeRepo,
	}
}

func (event *PaymentInRoute) Name() string {
	return EventNamePaymentInRoute
}

func (event *PaymentInRoute) PayloadType() any {
	pType := ""
	return &pType
}

func (event *PaymentInRoute) Validate(ctx context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New(" payload is not of type string")
	}

	return nil
}

func (event *PaymentInRoute) Execute(ctx context.Context, payload any) error {
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

	route, err := routePayment(ctx, event.routeRepo, models.RouteModeReceive, p)
	if err != nil {
		logger.WithError(err).Warn("could not route payment")

		if strings.Contains(err.Error(), "no routes matched for payment") {
			status := models.Status{
				EntityID:   p.GetID(),
				EntityType: "payment",
				State:      int32(commonv1.STATE_INACTIVE),
				Status:     int32(commonv1.STATUS_FAILED),
				Extra:      data.JSONMap{"error": err.Error()},
			}
			status.GenID(ctx)

			err = event.eventMan.Emit(ctx, EventNameStatusSave, &status)
			if err != nil {
				logger.WithError(err).Warn("could not emit status for save")
				return err
			}
			return nil
		}

		return err
	}

	p.RouteID = route.ID

	_, err = event.paymentRepo.Update(ctx, p, "route_id")
	if err != nil {
		logger.WithError(err).Warn("could not save routed payment to db")
		return err
	}

	evt := PaymentInQueue{}
	err = event.eventMan.Emit(ctx, evt.Name(), p.GetID())
	if err != nil {
		logger.WithError(err).Warn("could not queue out payment")
		return err
	}

	status := models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_ACTIVE),
		Status:     int32(commonv1.STATUS_QUEUED),
		Extra:      make(map[string]interface{}),
	}
	status.GenID(ctx)
	err = event.eventMan.Emit(ctx, EventNameStatusSave, &status)
	if err != nil {
		logger.WithError(err).Warn("could not emit status for save")
		return err
	}

	return nil
}

func routePayment(
	ctx context.Context,
	routeRepo repository.RouteRepository,
	routeMode string,
	payment *models.Payment,
) (*models.Route, error) {
	if payment.RouteID != "" {
		route, err := routeRepo.GetByID(ctx, payment.RouteID)
		if err != nil {
			return nil, err
		}
		return route, nil
	}

	routes, err := routeRepo.GetByModeTypeAndPartitionID(ctx,
		routeMode, payment.PaymentType, payment.PartitionID)
	if err != nil {
		return nil, err
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes matched for payment : %s", payment.GetID())
	}

	route := routes[0]
	if len(routes) > 1 {
		route, err = selectRoute(ctx, routes)
		if err != nil {
			return nil, err
		}
	}

	return route, nil
}

func loadRoute(ctx context.Context, qMan queue.Manager, routeRepo repository.RouteRepository, routeID string) (*models.Route, error) {
	if routeID == "" {
		return nil, errors.New("no route id provided")
	}

	route, err := routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return nil, err
	}

	err = qMan.AddPublisher(ctx, route.ID, route.URI)
	if err != nil {
		return route, err
	}

	return route, nil
}

func selectRoute(_ context.Context, routes []*models.Route) (*models.Route, error) {
	// TODO: find a simple way of routing payments mostly by settings
	// or contact and profile preferences
	if len(routes) == 0 {
		return nil, errors.New("no routes matched for payment")
	}

	return routes[0], nil
}
