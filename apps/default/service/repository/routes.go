package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
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
