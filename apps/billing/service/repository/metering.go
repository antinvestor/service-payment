//nolint:dupl // Structurally identical to rating.go but type-specific; Go generics cannot abstract this.
package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

// MeteredUsageRepository provides operations for metered usage records.
type MeteredUsageRepository interface {
	datastore.BaseRepository[*models.MeteredUsage]
	ListByBillingRunID(ctx context.Context, billingRunID string) ([]*models.MeteredUsage, error)
}

type meteredUsageRepository struct {
	datastore.BaseRepository[*models.MeteredUsage]
}

func NewMeteredUsageRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) MeteredUsageRepository {
	return &meteredUsageRepository{
		BaseRepository: datastore.NewBaseRepository[*models.MeteredUsage](
			ctx, dbPool, workMan, func() *models.MeteredUsage { return &models.MeteredUsage{} },
		),
	}
}

func (r *meteredUsageRepository) ListByBillingRunID(
	ctx context.Context,
	billingRunID string,
) ([]*models.MeteredUsage, error) {
	if billingRunID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.MeteredUsage
	result := r.Pool().DB(ctx, true).Where("billing_run_id = ?", billingRunID).Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}
