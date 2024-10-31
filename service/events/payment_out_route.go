package events

import (
	"context"
	"errors"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments-v1/service/repository"
	"github.com/antinvestor/service-payments-v1/service/models"

	"strings"

	"github.com/pitabwire/frame"
)

type PaymentOutRoute struct {
	Service    *frame.Service
	ProfileCli *profileV1.ProfileClient
}

func (event *PaymentOutRoute) Name() string {
	return "payment.out.route"
}

func (event *PaymentOutRoute) PayloadType() any {
	pType := ""
	return &pType
}

func (event *PaymentOutRoute) Validate(ctx context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New(" payload is not of type string")
	}

	return nil
}

func (event *PaymentOutRoute) Execute(ctx context.Context, payload any) error {

	paymentId := *payload.(*string)

	logger := event.Service.L(ctx).WithField("payload", paymentId).WithField("type", event.Name())
	logger.Debug("handling event")

	paymentRepo := repository.NewPaymentRepository(ctx, event.Service)

	p, err := paymentRepo.GetByID(ctx, paymentId)
	if err != nil {
		logger.WithError(err).Warn("could not get payment from db")
		return err
	}

	pr, err := event.ProfileCli.GetProfileByID(ctx, p.RecipientProfileID)
	if err != nil {
		logger.WithError(err).WithField("profile_id", p.RecipientProfileID).Warn("could not get profile by id")
		return err
	}

	contact := filterContactFromProfileByID(pr, p.RecipientContactID)
	switch contact.Type {
	case profileV1.ContactType_PHONE:
		p.PaymentType = models.RouteTypeShortForm
	default:
		p.PaymentType = models.RouteTypeAny
	}

	route, err := routePayment(ctx, event.Service, models.RouteModeTransmit, p)
	if err != nil {
		logger.WithError(err).Error("could not route payment")

		if strings.Contains(err.Error(), "no routes matched for payment") {
			pStatus := models.PaymentStatus{
				PaymentID: p.GetID(),
				State:          int32(commonv1.STATE_INACTIVE),
				Status:         int32(commonv1.STATUS_FAILED),
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

	p.RouteID = route.ID
	err = paymentRepo.Save(ctx, p)
	if err != nil {
		logger.WithError(err).Warn("could not save routed payment to db")
		return err
	}

	evt := PaymentOutQueue{}
	err = event.Service.Emit(ctx, evt.Name(), p.GetID())
	if err != nil {
		logger.WithError(err).Warn("could not queue out payment")
		return err
	}

	pStatus := models.PaymentStatus{
		PaymentID: p.GetID(),
		State:          int32(commonv1.STATE_ACTIVE),
		Status:         int32(commonv1.STATUS_QUEUED),
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
