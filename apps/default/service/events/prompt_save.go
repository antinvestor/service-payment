package events

import (
	"context"
	"errors"

	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/models"
	"gorm.io/gorm/clause"

	"github.com/pitabwire/frame"
)

type PromptSave struct {
	Service    *frame.Service
	ProfileCli *profileV1.ProfileClient
}

func (e *PromptSave) Name() string {
	return "prompt.save"
}

func (e *PromptSave) PayloadType() any {
	return &models.Prompt{}
}

func (e *PromptSave) Validate(ctx context.Context, payload any) error {
	logger := e.Service.Log(ctx).WithField("function", "PromptSave.Validate")

	prompt, ok := payload.(*models.Prompt)
	if !ok {
		logger.Error("Payload is not of type models.Prompt")
		return errors.New("payload is not of type models.Prompt")
	}

	// Log detailed ID information
	logger.
		WithField("prompt.ID", prompt.ID).
		WithField("prompt.GetID()", prompt.GetID()).
		WithField("prompt.BaseModel.ID", prompt.BaseModel.ID).
		Debug("Validating prompt ID")

	// Fix ID issues if possible
	if prompt.GetID() == "" {
		// If BaseModel ID is empty but explicit ID is set, try to use that
		if prompt.ID != "" {
			logger.Info("Using explicit ID field for validation")
			return nil
		}

		logger.Error("Prompt ID is not set and no fallback ID is available")
		return errors.New("prompt Id should already have been set")
	}

	// If we got here, the ID is valid
	logger.Debug("Prompt ID validation successful")
	return nil
}

func (e *PromptSave) Execute(ctx context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return errors.New("payload is not of type models.Prompt")
	}

	logger := e.Service.Log(ctx).WithField("payload", prompt).WithField("type", e.Name())
	logger.Debug("handling event")

	// Attempt to save to database
	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(prompt)

	err := result.Error
	if err != nil {
		logger.WithError(err).Error("could not save prompt to db")
		// Return the error so the caller knows the save failed
		return err
	}

	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")
	return nil
}
