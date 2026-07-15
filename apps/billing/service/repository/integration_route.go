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
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

// IntegrationRouteRepository manages billing → external-entity delivery routes.
type IntegrationRouteRepository interface {
	datastore.BaseRepository[*models.IntegrationRoute]
	// ListForLifecycle returns routes for a partition that accept the given event.
	// Matches mode lifecycle|any and route_type event|any.
	ListForLifecycle(ctx context.Context, partitionID, eventType string) ([]*models.IntegrationRoute, error)
}

type integrationRouteRepository struct {
	datastore.BaseRepository[*models.IntegrationRoute]
}

// NewIntegrationRouteRepository constructs the repository.
func NewIntegrationRouteRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) IntegrationRouteRepository {
	return &integrationRouteRepository{
		BaseRepository: datastore.NewBaseRepository[*models.IntegrationRoute](
			ctx, dbPool, workMan, func() *models.IntegrationRoute { return &models.IntegrationRoute{} },
		),
	}
}

func (r *integrationRouteRepository) ListForLifecycle(
	ctx context.Context,
	partitionID, eventType string,
) ([]*models.IntegrationRoute, error) {
	var routes []*models.IntegrationRoute
	q := r.Pool().DB(ctx, true).
		Where("(mode = ? OR mode = ?)", models.IntegrationRouteModeLifecycle, models.IntegrationRouteModeAny).
		Where("(route_type = ? OR route_type = ?)", models.IntegrationRouteTypeAny, eventType)
	if partitionID != "" {
		q = q.Where("partition_id = ?", partitionID)
	}
	if err := q.Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}
