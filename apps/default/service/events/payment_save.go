package events

import (
	"context"
	"errors"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	"github.com/antinvestor/service-payments/service/models"
	"gorm.io/gorm/clause"

	"github.com/pitabwire/frame"
)

type PaymentSave struct {
	Service *frame.Service
}

func (event *PaymentSave) Name() string {
	return "payment.save"
}

func (event *PaymentSave) PayloadType() any {
	return &models.Payment{}
}

func (event *PaymentSave) Validate(_ context.Context, payload any) error {
	payment, ok := payload.(*models.Payment)
	if !ok {
		return errors.New(" payload is not of type models.Payment")
	}

	if payment.GetID() == "" {
		return errors.New(" payment Id should already have been set ")
	}

	return nil
}

func (event *PaymentSave) Execute(ctx context.Context, payload any) error {
	payment, ok := payload.(*models.Payment)
	if !ok {
		return errors.New("payload is not of type models.Payment")
	}

	logger := event.Service.Log(ctx).WithField("type", event.Name())
	logger.WithField("payload", payment).Debug("handling event")

	result := event.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(payment)

	err := result.Error

	if err != nil {
		logger.WithError(err).Warn("could not save to db")
		return err
	}
	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")

	if !payment.OutBound {
		// Use the parent event's Service field instead of creating a new uninitialized event
		inRouteEvent := PaymentInRoute{Service: event.Service}
		err = event.Service.Emit(ctx, inRouteEvent.Name(), payment.GetID())
		if err != nil {
			return err
		}

		return nil
	}

	if payment.IsReleased() {
		// Use the parent event's Service field instead of creating a new uninitialized event
		outRouteEvent := PaymentOutRoute{Service: event.Service}
		err = event.Service.Emit(ctx, outRouteEvent.Name(), payment.GetID())
		if err != nil {
			logger.WithError(err).Warn("could not emit for queue out")
			return err
		}
	} else {
		status := models.Status{
			EntityID:   payment.GetID(),
			EntityType: "payment",
			State:      int32(commonv1.STATE_CHECKED.Number()),
			Status:     int32(commonv1.STATUS_QUEUED.Number()),
			Extra:      make(map[string]interface{}),
		}
		status.GenID(ctx)
		// Queue out payment status for further processing
		statusEvent := StatusSave{Service: event.Service}
		err = event.Service.Emit(ctx, statusEvent.Name(), &status)
		if err != nil {
			logger.WithError(err).Warn("could not emit status")
			return err
		}
	}
	return nil
}
