package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/util"
)

type StatusSave struct {
	statusRepo repository.StatusRepository
}

// NewStatusSave creates a new StatusSave event handler with the required dependencies
func NewStatusSave(statusRepo repository.StatusRepository) *StatusSave {
	return &StatusSave{
		statusRepo: statusRepo,
	}
}

func (e *StatusSave) Name() string {
	return EventNameStatusSave
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

	logger := util.Log(ctx).WithField("payload", status).WithField("type", e.Name())
	logger.Debug("handling event")

	err := e.statusRepo.Create(ctx, status)
	if err != nil {
		logger.WithError(err).Warn("could not save status to db")
		return err
	}
	logger.Debug("successfully saved record to db")

	return nil
}
