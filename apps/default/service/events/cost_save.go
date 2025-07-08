package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/gorm/clause"
)

type CostSave struct {
	Service *frame.Service
}

func (e *CostSave) Name() string {
	return "cost.save"
}

func (e *CostSave) PayloadType() any {
	return &models.Cost{}
}

func (e *CostSave) Validate(ctx context.Context, payload any) error {
	logger := e.Service.Log(ctx).WithField("function", "CostSave.Validate")

	cost, ok := payload.(*models.Cost)
	if !ok {
		logger.Error("Payload is not of type models.Cost")
		return errors.New("payload is not of type models.Cost")
	}

	if cost.ID == "" {
		logger.Error("Cost ID is not set")
		return errors.New("cost ID should already have been set")
	}

	logger.Debug("Cost ID validation successful")
	return nil
}

func (e *CostSave) Execute(ctx context.Context, payload any) error {
	cost := payload.(*models.Cost)

	logger := e.Service.Log(ctx).WithField("payload", cost).WithField("type", e.Name())
	logger.Debug("handling event")

	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(cost)

	err := result.Error
	if err != nil {
		logger.WithError(err).Error("could not save cost to db")
		return err
	}

	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")
	return nil
}
