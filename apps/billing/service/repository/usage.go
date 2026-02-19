package repository

import (
	"context"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/internal/apperrors"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

// UsageEventRepository provides operations for usage events.
type UsageEventRepository interface {
	datastore.BaseRepository[*models.UsageEvent]
	ListBySubscriptionAndPeriod(
		ctx context.Context,
		subscriptionID string,
		metricKey string,
		start, end time.Time,
	) ([]*models.UsageEvent, error)
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.UsageEvent], error)
}

type usageEventRepository struct {
	datastore.BaseRepository[*models.UsageEvent]
}

func NewUsageEventRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) UsageEventRepository {
	return &usageEventRepository{
		BaseRepository: datastore.NewBaseRepository[*models.UsageEvent](
			ctx, dbPool, workMan, func() *models.UsageEvent { return &models.UsageEvent{} },
		),
	}
}

func (r *usageEventRepository) ListBySubscriptionAndPeriod(
	ctx context.Context, subscriptionID string, metricKey string, start, end time.Time,
) ([]*models.UsageEvent, error) {
	if subscriptionID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.UsageEvent
	result := r.Pool().DB(ctx, true).
		Where("subscription_id = ? AND metric_key = ? AND timestamp >= ? AND timestamp < ?",
			subscriptionID, metricKey, start, end).
		Order("timestamp ASC").
		Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}

//nolint:dupl // SearchAsESQ follows a generic pattern that Go cannot DRY with generics + worker pools.
func (r *usageEventRepository) SearchAsESQ(
	ctx context.Context, query string,
) (workerpool.JobResultPipe[[]*models.UsageEvent], error) {
	job := workerpool.NewJob(func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.UsageEvent]) error {
		rawQuery, err := NewSearchRawQuery(ctx, query)
		if err != nil {
			return jobResult.WriteError(ctx, err)
		}

		sqlQuery := rawQuery.ToQueryConditions()
		for sqlQuery.canLoad() {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			var list []*models.UsageEvent
			result := r.Pool().DB(ctx, true).Where(sqlQuery.sql, sqlQuery.args...).
				Offset(sqlQuery.offset).Limit(sqlQuery.batchSize).Find(&list)
			if result.Error != nil {
				return jobResult.WriteError(ctx, apperrors.ErrSystemFailure.Override(result.Error))
			}

			if err := jobResult.WriteResult(ctx, list); err != nil {
				return err
			}

			if sqlQuery.stop(len(list)) {
				break
			}
		}
		return nil
	})

	if err := workerpool.SubmitJob(ctx, r.WorkManager(), job); err != nil {
		return nil, err
	}
	return job, nil
}
