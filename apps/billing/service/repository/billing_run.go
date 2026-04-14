package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

// BillingRunRepository provides operations for billing runs.
type BillingRunRepository interface {
	datastore.BaseRepository[*models.BillingRun]
	GetByIdempotency(ctx context.Context, idempotencyKey string) (*models.BillingRun, error)
}

type billingRunRepository struct {
	datastore.BaseRepository[*models.BillingRun]
}

func NewBillingRunRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) BillingRunRepository {
	return &billingRunRepository{
		BaseRepository: datastore.NewBaseRepository[*models.BillingRun](
			ctx, dbPool, workMan, func() *models.BillingRun { return &models.BillingRun{} },
		),
	}
}

func (r *billingRunRepository) GetByIdempotency(
	ctx context.Context,
	idempotencyKey string,
) (*models.BillingRun, error) {
	if idempotencyKey == "" {
		return nil, apperrors.ErrUnspecifiedReference
	}

	var run models.BillingRun
	result := r.Pool().DB(ctx, true).Where("idempotency = ?", idempotencyKey).First(&run)
	if result.Error != nil {
		if data.ErrorIsNoRows(result.Error) {
			return nil, apperrors.ErrBillingRunNotFound
		}
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return &run, nil
}
