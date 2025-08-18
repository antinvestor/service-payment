package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"gorm.io/gorm/clause"

	"github.com/pitabwire/frame"
)

type StatusSave struct {
	Service *frame.Service
}

func (e *StatusSave) Name() string {
	return "status.save"
}

func (e *StatusSave) PayloadType() any {
	return &models.Status{}
}

func (e *StatusSave) Validate(_ context.Context, payload any) error {
	status, ok := payload.(*models.Status)
	if !ok {
		return errors.New("payload is not of type models.Status")
	}
	if status.GetID() == "" {
		return errors.New("status Id should already have been set")
	}
	return nil
}

func (e *StatusSave) Execute(ctx context.Context, payload any) error {
	status, ok := payload.(*models.Status)
	if !ok {
		return errors.New("payload is not of type models.Status")
	}

	logger := e.Service.Log(ctx).WithField("payload", status).WithField("type", e.Name())
	logger.Debug("handling event")

	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(status)

	err := result.Error
	if err != nil {
		logger.WithError(err).Warn("could not save status to db")
		return err
	}
	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")

	return nil
}
