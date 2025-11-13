package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

type statusRepository struct {
	datastore.BaseRepository[*models.Status]
}

func NewStatusRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) StatusRepository {
	repo := statusRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Status](
			ctx, dbPool, workMan, func() *models.Status { return &models.Status{} },
		),
	}
	return &repo
}

func (sr *statusRepository) GetByEntity(ctx context.Context, entityID, entityType string) (*models.Status, error) {
	status := &models.Status{}
	err := sr.Pool().DB(ctx, true).First(status, "entity_id = ? AND entity_type = ?", entityID, entityType).Error
	return status, err
}
