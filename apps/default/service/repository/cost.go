package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

type costRepository struct {
	datastore.BaseRepository[*models.Cost]
}

func NewCostRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) CostRepository {
	repo := costRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Cost](
			ctx, dbPool, workMan, func() *models.Cost { return &models.Cost{} },
		),
	}
	return &repo
}

func (cr *costRepository) GetByPaymentID(ctx context.Context, paymentID string) ([]*models.Cost, error) {
	var costs []*models.Cost
	err := cr.Pool().DB(ctx, true).Where("payment_id = ?", paymentID).Find(&costs).Error
	return costs, err
}
