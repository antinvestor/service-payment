// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

// SubscriptionRepository provides operations for subscriptions.
type SubscriptionRepository interface {
	datastore.BaseRepository[*models.Subscription]
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.Subscription], error)
	ListActiveByProfile(ctx context.Context, profileID string) ([]*models.Subscription, error)
	// GetByPolarSubscriptionID finds a subscription by the polar subscription ID stored in its Data map.
	// Returns apperrors.ErrSubscriptionNotFound when no match exists.
	GetByPolarSubscriptionID(ctx context.Context, polarSubscriptionID string) (*models.Subscription, error)
	// GetPendingByProfileAndPlan returns the first PENDING subscription for the given profile+plan combination.
	// Returns apperrors.ErrSubscriptionNotFound when no match exists.
	GetPendingByProfileAndPlan(ctx context.Context, profileID, planID string) (*models.Subscription, error)
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

func (r *subscriptionRepository) ListActiveByProfile(
	ctx context.Context,
	profileID string,
) ([]*models.Subscription, error) {
	if profileID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.Subscription
	result := r.Pool().DB(ctx, true).
		Where("profile_id = ? AND state = ?", profileID, models.SubscriptionStateActive).
		Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}

func (r *subscriptionRepository) GetByPolarSubscriptionID(
	ctx context.Context,
	polarSubscriptionID string,
) (*models.Subscription, error) {
	if polarSubscriptionID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var sub models.Subscription
	result := r.Pool().DB(ctx, true).
		Where("data->>'polarSubscriptionId' = ?", polarSubscriptionID).
		First(&sub)
	if result.Error != nil {
		if data.ErrorIsNoRows(result.Error) {
			return nil, apperrors.ErrSubscriptionNotFound
		}
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return &sub, nil
}

func (r *subscriptionRepository) GetPendingByProfileAndPlan(
	ctx context.Context,
	profileID, planID string,
) (*models.Subscription, error) {
	if profileID == "" || planID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var sub models.Subscription
	result := r.Pool().DB(ctx, true).
		Where("profile_id = ? AND plan_id = ? AND state = ?", profileID, planID, models.SubscriptionStatePending).
		Order("created_at DESC").
		First(&sub)
	if result.Error != nil {
		if data.ErrorIsNoRows(result.Error) {
			return nil, apperrors.ErrSubscriptionNotFound
		}
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return &sub, nil
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
