package events

import (
	"context"
	"errors"
	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/gorm/clause"
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

func (event *PaymentSave) Validate(ctx context.Context, payload any) error {
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

	payment := payload.(*models.Payment)

	logger := event.Service.L().WithField("type", event.Name())
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

	return nil

}
