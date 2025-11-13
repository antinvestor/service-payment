package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/util"
)

type PromptSave struct {
	promptRepo repository.PromptRepository
}

// NewPromptSave creates a new PromptSave event handler with the required dependencies
func NewPromptSave(promptRepo repository.PromptRepository) *PromptSave {
	return &PromptSave{
		promptRepo: promptRepo,
	}
}

func (e *PromptSave) Name() string {
	return EventNamePromptSave
}

func (e *PromptSave) PayloadType() any {
	return &models.Prompt{}
}

func (e *PromptSave) Validate(ctx context.Context, payload any) error {
	logger := util.Log(ctx).WithField("function", "PromptSave.Validate")

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

	logger := util.Log(ctx).WithField("payload", prompt).WithField("type", e.Name())
	logger.Debug("handling event")

	// Attempt to save to database
	err := e.promptRepo.Create(ctx, prompt)
	if err != nil {
		logger.WithError(err).Error("could not save prompt to db")
		// Return the error so the caller knows the save failed
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
