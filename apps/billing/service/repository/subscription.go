package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/internal/apperrors"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

// SubscriptionRepository provides operations for subscriptions.
type SubscriptionRepository interface {
	datastore.BaseRepository[*models.Subscription]
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.Subscription], error)
	ListActiveByCustomer(ctx context.Context, customerID string) ([]*models.Subscription, error)
}

type subscriptionRepository struct {
	datastore.BaseRepository[*models.Subscription]
}

func NewSubscriptionRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) SubscriptionRepository {
	return &subscriptionRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Subscription](
			ctx, dbPool, workMan, func() *models.Subscription { return &models.Subscription{} },
		),
	}
}

func (r *subscriptionRepository) ListActiveByCustomer(
	ctx context.Context,
	customerID string,
) ([]*models.Subscription, error) {
	if customerID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.Subscription
	result := r.Pool().DB(ctx, true).
		Where("customer_id = ? AND state = ?", customerID, models.SubscriptionStateActive).
		Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}

//nolint:dupl // SearchAsESQ follows a generic pattern that Go cannot DRY with generics + worker pools.
func (r *subscriptionRepository) SearchAsESQ(
	ctx context.Context, query string,
) (workerpool.JobResultPipe[[]*models.Subscription], error) {
	job := workerpool.NewJob(
		func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.Subscription]) error {
			rawQuery, err := NewSearchRawQuery(ctx, query)
			if err != nil {
				return jobResult.WriteError(ctx, err)
			}

			sqlQuery := rawQuery.ToQueryConditions()
			for sqlQuery.canLoad() {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				var list []*models.Subscription
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
		},
	)

	if err := workerpool.SubmitJob(ctx, r.WorkManager(), job); err != nil {
		return nil, err
	}
	return job, nil
}
