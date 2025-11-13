package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/util"
)

type CostSave struct {
	costRepo repository.CostRepository
}

// NewCostSave creates a new CostSave event handler with the required dependencies
func NewCostSave(costRepo repository.CostRepository) *CostSave {
	return &CostSave{
		costRepo: costRepo,
	}
}

func (e *CostSave) Name() string {
	return EventNameCostSave
}

func (e *CostSave) PayloadType() any {
	return &models.Cost{}
}

func (e *CostSave) Validate(ctx context.Context, payload any) error {
	logger := util.Log(ctx).WithField("function", "CostSave.Validate")

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
	cost, ok := payload.(*models.Cost)
	if !ok {
		return errors.New("payload is not of type models.Cost")
	}

	logger := util.Log(ctx).WithField("payload", cost).WithField("type", e.Name())
	logger.Debug("handling event")

	err := e.costRepo.Create(ctx, cost)
	if err != nil {
		logger.WithError(err).Error("could not save cost to db")
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
