package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/default/service/models"
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
	err := sr.Pool().DB(ctx, true).
		Where("entity_id = ? AND entity_type = ?", entityID, entityType).
		Order("created_at DESC").
		First(status).Error
	return status, err
}

func (sr *statusRepository) GetByEntityIDList(
	ctx context.Context,
	entityIDs []string,
	entityType string,
) (map[string]*models.Status, error) {
	var statusList []*models.Status
	err := sr.Pool().DB(ctx, true).
		Where("entity_id IN ? AND entity_type = ?", entityIDs, entityType).
		Order("created_at DESC").
		Find(&statusList).Error
	if err != nil {
		return nil, err
	}

	// Keep only the latest status per entity (first occurrence due to DESC order)
	statusMap := map[string]*models.Status{}
	for _, status := range statusList {
		if _, exists := statusMap[status.EntityID]; !exists {
			statusMap[status.EntityID] = status
		}
	}

	return statusMap, err
}
