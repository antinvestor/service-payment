package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"

	"github.com/pitabwire/frame"
)

type StatusRepository interface {
	GetByEntity(ctx context.Context, entityID, entityType string) (*models.Status, error)
	Save(ctx context.Context, status *models.Status) error
}

type statusRepository struct {
	abstractRepository
}

func NewStatusRepository(_ context.Context, service *frame.Service) StatusRepository {
	return &statusRepository{abstractRepository{service: service}}
}

func (repo *statusRepository) GetByEntity(ctx context.Context, entityID, entityType string) (*models.Status, error) {
	status := models.Status{}
	err := repo.readDB(ctx).First(&status, "entity_id = ? AND entity_type = ?", entityID, entityType).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (repo *statusRepository) Save(ctx context.Context, status *models.Status) error {
	return repo.writeDB(ctx).Save(status).Error
}
