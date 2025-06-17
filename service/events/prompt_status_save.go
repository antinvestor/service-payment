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

func (e *PromptStatusSave) Validate(ctx context.Context, payload any) error {
	logger := e.Service.Log(ctx).WithField("function", "PromptStatusSave.Validate")

	promptStatus, ok := payload.(*models.PromptStatus)
	if !ok {
		logger.Error("Payload is not of type models.PromptStatus")
		return errors.New("payload is not of type models.PromptStatus")
	}

	// Log detailed ID information
	logger.
		WithField("promptStatus.ID", promptStatus.ID).
		WithField("promptStatus.GetID()", promptStatus.GetID()).
		WithField("promptStatus.PromptID", promptStatus.PromptID).
		Debug("Validating prompt status ID")

	// Check for ID validity
	if promptStatus.GetID() == "" {
		// If BaseModel ID is empty but explicit ID is set, try to use that
		if promptStatus.ID != "" {
			logger.Info("Using explicit ID field as fallback")
			return nil
		}

		// If PromptID is set but ID isn't, use PromptID as ID
		if promptStatus.PromptID != "" {
			promptStatus.ID = promptStatus.PromptID
			logger.Info("Setting ID from PromptID field")
			return nil
		}

		logger.Error("PromptStatus ID is not set and no fallback ID is available")
		return errors.New("promptStatus Id should already have been set")
	}

	// If we got here, the ID is valid
	logger.Debug("PromptStatus ID validation successful")
	return nil
}

func (e *PromptStatusSave) Execute(ctx context.Context, payload any) error {
	promptStatus := payload.(*models.PromptStatus)

	logger := e.Service.Log(ctx).WithField("payload", promptStatus).WithField("type", e.Name())
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
