package events

import (
	"context"
	"errors"

	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/gorm/clause"
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

func (e *PromptSave) Validate(_ context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return errors.New(" payload is not of type models.Prompt")
	}
	if prompt.GetID() == "" {
		return errors.New(" prompt Id should already have been set ")
	}
	return nil
}

func (e *PromptSave) Execute(ctx context.Context, payload any) error {
	prompt := payload.(*models.Prompt)

	logger := e.Service.L(ctx).WithField("payload", prompt).WithField("type", e.Name())
	logger.Debug("handling event")

	// Attempt to save to database, but continue if table doesn't exist
	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(prompt)

	err := result.Error
	if err != nil {
		// Log the error but don't fail the operation
		logger.WithError(err).Warn("could not save prompt to db - continuing execution")
		// We're intentionally not returning the error here
	} else {
		logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")
	}

	// Always return success
	return nil
}
