package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
)

type PromptRepository interface {
	GetByID(ctx context.Context, id string) (*models.Prompt, error)
	
	Save(ctx context.Context, prompt *models.Prompt) error
}
