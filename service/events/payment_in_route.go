package events

import (
	"context"
	"errors"
	"fmt"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"

	"strings"

	"github.com/pitabwire/frame"
)

func filterContactFromProfileByID(profile *profileV1.ProfileObject, contactID string) *profileV1.ContactObject {

	for _, contact := range profile.GetContacts() {
		if contact.GetId() == contactID {
			return contact
		}
	}

	return nil
}

type PaymentInRoute struct {
	Service *frame.Service
}

func (event *PaymentInRoute) Name() string {
	return "payment.in.route"
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
	paymentID := *payload.(*string)
	logger := event.Service.L(ctx).WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling event")

	paymentRepo := repository.NewPaymentRepository(ctx, event.Service)

	n, err := paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	route, err := routePayment(ctx, event.Service, models.RouteModeReceive, n)
	if err != nil {
		logger.WithError(err).Warn("could not route payment")

		if strings.Contains(err.Error(), "no routes matched for payment") {
			pStatus := models.PaymentStatus{
				PaymentID: n.GetID(),
				State:     int32(commonv1.STATE_INACTIVE),
				Status:    int32(commonv1.STATUS_FAILED),
				Extra: frame.DBPropertiesFromMap(map[string]string{
					"error": err.Error(),
				}),
			}

			pStatus.GenID(ctx)

			eventStatus := PaymentStatusSave{}
			err = event.Service.Emit(ctx, eventStatus.Name(), pStatus)
			if err != nil {
				logger.WithError(err).Warn("could not emit status for save")
				return err
			}

			return nil
		}

		return err
	}

	n.RouteID = route.ID

	err = paymentRepo.Save(ctx, n)
	if err != nil {
		logger.WithError(err).Warn("could not save routed payment to db")
		return err
	}

	evt := PaymentInQueue{}
	err = event.Service.Emit(ctx, evt.Name(), n.GetID())
	if err != nil {
		logger.WithError(err).Warn("could not queue out payment")
		return err
	}

	pStatus := models.PaymentStatus{
		PaymentID: n.GetID(),
		State:     int32(commonv1.STATE_ACTIVE),
		Status:    int32(commonv1.STATUS_QUEUED),
	}

	pStatus.GenID(ctx)

	// Queue out payment status for further processing
	eventStatus := PaymentStatusSave{}
	err = event.Service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Warn("could not emit status for save")
		return err
	}

	return nil
}

func routePayment(ctx context.Context, service *frame.Service, routeMode string, payment *models.Payment) (*models.Route, error) {

	routeRepository := repository.NewRouteRepository(ctx, service)
	if payment.RouteID != "" {
		route, err := routeRepository.GetByID(ctx, payment.RouteID)
		if err != nil {
			return nil, err
		}
		return route, nil
	}

	routes, err := routeRepository.GetByModeTypeAndPartitionID(ctx,
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

func loadRoute(ctx context.Context, service *frame.Service, routeId string) (*models.Route, error) {

	if routeId == "" {
		return nil, fmt.Errorf("no route id provided")
	}

	routeRepository := repository.NewRouteRepository(ctx, service)

	route, err := routeRepository.GetByID(ctx, routeId)
	if err != nil {
		return nil, err
	}

	err = service.AddPublisher(ctx, route.ID, route.Uri)
	if err != nil {
		return route, err
	}

	return route, nil

}

func selectRoute(_ context.Context, routes []*models.Route) (*models.Route, error) {
	//TODO: find a simple way of routing payments mostly by settings
	// or contact and profile preferences
	if len(routes) == 0 {
		return nil, errors.New("no routes matched for payment")
	}

	return routes[0], nil
}
