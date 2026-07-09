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

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

type routeRepository struct {
	datastore.BaseRepository[*models.Route]
}

func NewRouteRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) RouteRepository {
	repo := routeRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Route](
			ctx, dbPool, workMan, func() *models.Route { return &models.Route{} },
		),
	}
	return &repo
}

func (rr *routeRepository) GetByMode(ctx context.Context, mode string) ([]*models.Route, error) {
	var routes []*models.Route
	err := rr.Pool().DB(ctx, true).Find(&routes,
		"mode = ? OR ( mode = ?)", mode, models.RouteModeTransceive).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (rr *routeRepository) GetByModeTypeAndPartitionID(
	ctx context.Context,
	mode string,
	routeType string,
	partitionID string,
) ([]*models.Route, error) {
	var routes []*models.Route
	err := rr.Pool().DB(ctx, true).Find(&routes,
		"partition_id = ? AND ( route_type = ? OR route_type = ? ) AND (mode = ? OR ( mode = ?))",
		partitionID, "any", routeType, mode, models.RouteModeTransceive).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}
