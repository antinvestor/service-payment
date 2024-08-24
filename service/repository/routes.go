package repository

import (
	"context"

	"github.com/antinvestor/service-payments-v1/service/models"
	"github.com/pitabwire/frame"
)

type RouteRepository interface {
	GetByID(ctx context.Context, id string) (*models.Route, error)
	GetByModeTypeAndPartitionID(ctx context.Context, mode string, routeType string, partitionId string) ([]*models.Route, error)
	GetByMode(ctx context.Context, mode string) ([]*models.Route, error)
	Save(ctx context.Context, channel *models.Route) error
}

type routeRepository struct {
	abstractRepository
}

func NewRouteRepository(_ context.Context, service *frame.Service) RouteRepository {
	return &routeRepository{abstractRepository{service: service}}
}

func (repo *routeRepository) GetByID(ctx context.Context, id string) (*models.Route, error) {
	route := models.Route{}
	err := repo.readDb(ctx).First(&route, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (repo *routeRepository) GetByMode(ctx context.Context, mode string) ([]*models.Route, error) {
	var routes []*models.Route

	err := repo.readDb(ctx).Find(&routes,
		"mode = ? OR ( mode = ?)", mode, models.RouteModeTransceive).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (repo *routeRepository) GetByModeTypeAndPartitionID(ctx context.Context, mode string, routeType string, partitionId string) ([]*models.Route, error) {
	var routes []*models.Route

	err := repo.readDb(ctx).Find(&routes,
		"partition_id = ? AND ( route_type = ? OR route_type = ? ) AND (mode = ? OR ( mode = ?))",
		partitionId, "any", routeType, mode, models.RouteModeTransceive).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (repo *routeRepository) Save(ctx context.Context, route *models.Route) error {
	return repo.writeDb(ctx).Save(route).Error
}
