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
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

// CatalogVersionRepository provides operations for catalog versions.
type CatalogVersionRepository interface {
	datastore.BaseRepository[*models.CatalogVersion]
	GetEffectiveByCatalogID(ctx context.Context, catalogID string, at time.Time) (*models.CatalogVersion, error)
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.CatalogVersion], error)
}

type catalogVersionRepository struct {
	datastore.BaseRepository[*models.CatalogVersion]
}

func NewCatalogVersionRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) CatalogVersionRepository {
	return &catalogVersionRepository{
		BaseRepository: datastore.NewBaseRepository[*models.CatalogVersion](
			ctx, dbPool, workMan, func() *models.CatalogVersion { return &models.CatalogVersion{} },
		),
	}
}

func (r *catalogVersionRepository) GetEffectiveByCatalogID(
	ctx context.Context,
	catalogID string,
	at time.Time,
) (*models.CatalogVersion, error) {
	if catalogID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var cv models.CatalogVersion
	result := r.Pool().DB(ctx, true).
		Where("catalog_id = ? AND published_at IS NOT NULL AND effective_at IS NOT NULL AND effective_at <= ? AND (retired_at IS NULL OR retired_at > ?)",
			catalogID, at, at).
		Order("effective_at DESC").
		First(&cv)
	if result.Error != nil {
		if data.ErrorIsNoRows(result.Error) {
			return nil, apperrors.ErrCatalogVersionNotFound
		}
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return &cv, nil
}

//nolint:dupl // SearchAsESQ follows a generic pattern that Go cannot DRY with generics + worker pools.
func (r *catalogVersionRepository) SearchAsESQ(
	ctx context.Context, query string,
) (workerpool.JobResultPipe[[]*models.CatalogVersion], error) {
	job := workerpool.NewJob(
		func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.CatalogVersion]) error {
			rawQuery, err := NewSearchRawQuery(ctx, query)
			if err != nil {
				return jobResult.WriteError(ctx, err)
			}

			sqlQuery := rawQuery.ToQueryConditions()
			for sqlQuery.canLoad() {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				var list []*models.CatalogVersion
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

// PlanRepository provides operations for plans.
type PlanRepository interface {
	datastore.BaseRepository[*models.Plan]
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.Plan], error)
}

type planRepository struct {
	datastore.BaseRepository[*models.Plan]
}

func NewPlanRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) PlanRepository {
	return &planRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Plan](
			ctx, dbPool, workMan, func() *models.Plan { return &models.Plan{} },
		),
	}
}

//nolint:dupl // SearchAsESQ follows a generic pattern that Go cannot DRY with generics + worker pools.
func (r *planRepository) SearchAsESQ(
	ctx context.Context, query string,
) (workerpool.JobResultPipe[[]*models.Plan], error) {
	job := workerpool.NewJob(func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.Plan]) error {
		rawQuery, err := NewSearchRawQuery(ctx, query)
		if err != nil {
			return jobResult.WriteError(ctx, err)
		}

		sqlQuery := rawQuery.ToQueryConditions()
		for sqlQuery.canLoad() {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			var list []*models.Plan
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

// ComponentRepository provides operations for components.
type ComponentRepository interface {
	datastore.BaseRepository[*models.Component]
	ListByPlanID(ctx context.Context, planID string) ([]*models.Component, error)
}

type componentRepository struct {
	datastore.BaseRepository[*models.Component]
}

func NewComponentRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) ComponentRepository {
	return &componentRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Component](
			ctx, dbPool, workMan, func() *models.Component { return &models.Component{} },
		),
	}
}

func (r *componentRepository) ListByPlanID(ctx context.Context, planID string) ([]*models.Component, error) {
	if planID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.Component
	result := r.Pool().DB(ctx, true).Where("plan_id = ?", planID).Preload("Tiers").Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}

// TierRepository provides operations for tiers.
type TierRepository interface {
	datastore.BaseRepository[*models.Tier]
	ListByComponentID(ctx context.Context, componentID string) ([]*models.Tier, error)
}

type tierRepository struct {
	datastore.BaseRepository[*models.Tier]
}

func NewTierRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) TierRepository {
	return &tierRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Tier](
			ctx, dbPool, workMan, func() *models.Tier { return &models.Tier{} },
		),
	}
}

func (r *tierRepository) ListByComponentID(ctx context.Context, componentID string) ([]*models.Tier, error) {
	if componentID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.Tier
	result := r.Pool().DB(ctx, true).Where("component_id = ?", componentID).Order("sort_order ASC").Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}
