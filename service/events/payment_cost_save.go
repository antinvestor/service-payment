package events

import (
	"context"
	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/pitabwire/frame"
	"github.com/pkg/errors"
	"gorm.io/gorm/clause"
)

type CostSave struct {
	Service *frame.Service
}

func (event *CostSave) Name() string {
	return "cost.save"
}

func (event *CostSave) PayloadType() any {
	return &models.Cost{}
}

func (event *CostSave) Validate(ctx context.Context, payload any) error {
	cost, ok := payload.(*models.Cost)
	if !ok {
		return errors.New(" payload is not of type models.Cost")
	}

	if cost.GetID() == "" {
		return errors.New(" cost Id should already have been set ")
	}

	return nil
}

func (event *CostSave) Execute(ctx context.Context, payload any) error {
	cost := payload.(*models.Cost)

	logger := event.Service.L().WithField("type", event.Name())
	logger.WithField("payload", cost).Debug("handling event")

	result := event.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(cost)

	err := result.Error
	if err != nil {
		logger.WithError(err).Warn("could not save to db")
		return err
	}
	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")

	return nil
}
