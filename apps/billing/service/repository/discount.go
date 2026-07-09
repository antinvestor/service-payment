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
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

// DiscountRepository provides operations for discount rules.
type DiscountRepository interface {
	datastore.BaseRepository[*models.Discount]
	ListActive(ctx context.Context, at time.Time) ([]*models.Discount, error)
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.Discount], error)
}

type discountRepository struct {
	datastore.BaseRepository[*models.Discount]
}

func NewDiscountRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) DiscountRepository {
	return &discountRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Discount](
			ctx, dbPool, workMan, func() *models.Discount { return &models.Discount{} },
		),
	}
}

func (r *discountRepository) ListActive(ctx context.Context, at time.Time) ([]*models.Discount, error) {
	var list []*models.Discount
	result := r.Pool().DB(ctx, true).
		Where("(start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at > ?)", at, at).
		Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}

//nolint:dupl // SearchAsESQ follows a generic pattern that Go cannot DRY with generics + worker pools.
func (r *discountRepository) SearchAsESQ(
	ctx context.Context, query string,
) (workerpool.JobResultPipe[[]*models.Discount], error) {
	job := workerpool.NewJob(func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.Discount]) error {
		rawQuery, err := NewSearchRawQuery(ctx, query)
		if err != nil {
			return jobResult.WriteError(ctx, err)
		}

		sqlQuery := rawQuery.ToQueryConditions()
		for sqlQuery.canLoad() {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			var list []*models.Discount
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

// DiscountedLineRepository provides operations for discounted line records.
type DiscountedLineRepository interface {
	datastore.BaseRepository[*models.DiscountedLine]
	ListByBillingRunID(ctx context.Context, billingRunID string) ([]*models.DiscountedLine, error)
}

type discountedLineRepository struct {
	datastore.BaseRepository[*models.DiscountedLine]
}

func NewDiscountedLineRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) DiscountedLineRepository {
	return &discountedLineRepository{
		BaseRepository: datastore.NewBaseRepository[*models.DiscountedLine](
			ctx, dbPool, workMan, func() *models.DiscountedLine { return &models.DiscountedLine{} },
		),
	}
}

func (r *discountedLineRepository) ListByBillingRunID(
	ctx context.Context,
	billingRunID string,
) ([]*models.DiscountedLine, error) {
	if billingRunID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.DiscountedLine
	result := r.Pool().DB(ctx, true).Where("billing_run_id = ?", billingRunID).Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}
