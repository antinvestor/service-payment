package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
)

type PromptStatusRepository interface {
	GetByID(ctx context.Context, id string) (*models.PromptStatus, error)
	GetByPromptID(ctx context.Context, promptId string) ([]models.PromptStatus, error)
	Save(ctx context.Context, promptStatus *models.PromptStatus) error
}

type promptStatusRepository struct {
	abstractRepository
}

func NewPromptStatusRepository(ctx context.Context, service *frame.Service) PromptStatusRepository {
	return &promptStatusRepository{abstractRepository{service: service}}
}

func (repo *promptStatusRepository) GetByID(ctx context.Context, id string) (*models.PromptStatus, error) {
	promptStatus := models.PromptStatus{}
	err := repo.readDb(ctx).First(&promptStatus, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &promptStatus, nil
}

func (repo *promptStatusRepository) GetByPromptID(ctx context.Context, promptId string) ([]models.PromptStatus, error) {
	var promptStatusList []models.PromptStatus

	err := repo.readDb(ctx).Find(&promptStatusList, "prompt_id = ?", promptId).Error
	if err != nil {
		return nil, err
	}
	return promptStatusList, nil
}

func (repo *promptStatusRepository) Save(ctx context.Context, promptStatus *models.PromptStatus) error {
	return repo.writeDb(ctx).Save(promptStatus).Error
}
