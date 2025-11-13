package events

import (
	"context"
	"errors"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util"

	"strings"

	"github.com/pitabwire/frame"
)

type PaymentOutRoute struct {
	Service    *frame.Service
	ProfileCli *profilev1.ProfileClient
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
	paymentPtr, ok := payload.(*string)
	if !ok {
		return errors.New("payload is not of type *string")
	}
	if paymentPtr == nil {
		return errors.New("payload is nil")
	}
	paymentID := *paymentPtr

	logger := util.Log(ctx).WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling event")

	paymentRepo := repository.NewPaymentRepository(ctx, event.Service)

	p, err := paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		logger.WithError(err).Warn("could not get payment from db")
		return err
	}

	route, err := routePayment(ctx, event.Service, models.RouteModeTransmit, p)
	if err != nil {
		logger.WithError(err).Error("could not route payment")

		if strings.Contains(err.Error(), "no routes matched for payment") {
			status := models.Status{
				EntityID:   p.GetID(),
				EntityType: "payment",
				State:      int32(commonv1.STATE_INACTIVE),
				Status:     int32(commonv1.STATUS_FAILED),
				Extra:      data.JSONMap{"error": err.Error()},
			}
			status.GenID(ctx)

			err = event.Service.Emit(ctx, EventNameStatusSave, &status)
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

	status := models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_ACTIVE),
		Status:     int32(commonv1.STATUS_QUEUED),
		Extra:      make(map[string]interface{}),
	}
	status.GenID(ctx)

	err = event.Service.Emit(ctx, EventNameStatusSave, &status)
	if err != nil {
		logger.WithError(err).Warn("could not emit status for save")
		return err
	}

	return nil
}
