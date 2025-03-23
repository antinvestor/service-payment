package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/gorm/clause"
)

type PaymentStatusSave struct {
	Service *frame.Service
}

func (e *PaymentStatusSave) Name() string {
	return "paymentStatus.save"
}

func (e *PaymentStatusSave) PayloadType() any {
	return &models.PaymentStatus{}
}

func (e *PaymentStatusSave) Validate(_ context.Context, payload any) error {
	paymentStatus, ok := payload.(*models.PaymentStatus)
	if !ok {
		return errors.New(" payload is not of type models.PaymentStatus")
	}
	if paymentStatus.GetID() == "" {
		return errors.New(" paymentStatus Id should already have been set ")
	}
	return nil
}

func (e *PaymentStatusSave) Execute(ctx context.Context, payload any) error {
	pStatus := payload.(*models.PaymentStatus)

	logger := e.Service.L(ctx).WithField("payload", pStatus).WithField("type", e.Name())
	logger.Debug("handling event")

	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(pStatus)

	err := result.Error
	if err != nil {
		logger.WithError(err).Warn("could not save payment status to db")
		return err
	}
	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")

	return nil
}
