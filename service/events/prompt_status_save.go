package events

import (
	"context"
	"errors"

	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/gorm/clause"
)

type PromptStatusSave struct {
	Service    *frame.Service
	ProfileCli *profileV1.ProfileClient
}

func (e *PromptStatusSave) Name() string {
	return "promptStatus.save"
}

func (e *PromptStatusSave) PayloadType() any {
	return &models.PromptStatus{}
}

func (e *PromptStatusSave) Validate(_ context.Context, payload any) error {
	promptStatus, ok := payload.(*models.PromptStatus)
	if !ok {
		return errors.New(" payload is not of type models.PromptStatus")
	}
	if promptStatus.GetID() == "" {
		return errors.New(" promptStatus Id should already have been set ")
	}
	return nil
}

func (e *PromptStatusSave) Execute(ctx context.Context, payload any) error {
	promptStatus := payload.(*models.PromptStatus)

	logger := e.Service.L(ctx).WithField("payload", promptStatus).WithField("type", e.Name())
	logger.Debug("handling event")

	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(promptStatus)

	err := result.Error
	if err != nil {
		logger.WithError(err).Warn("could not save prompt status to db")
		return err
	}
	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")

	return nil
}
