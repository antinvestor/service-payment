package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"

	"github.com/pitabwire/frame"
)

type CostRepository interface {
	Get(ctx context.Context, id string) (*models.Cost, error)
	Save(ctx context.Context, cost *models.Cost) error
	Delete(ctx context.Context, id string) error
	GetByPaymentID(ctx context.Context, paymentID string) ([]*models.Cost, error)
}

type costRepository struct {
	abstractRepository
}

func NewCostRepository(_ context.Context, service *frame.Service) CostRepository {
	return &costRepository{
		abstractRepository: abstractRepository{service: service},
	}
}

func (r *costRepository) Get(ctx context.Context, id string) (*models.Cost, error) {
	var cost models.Cost
	if err := r.readDB(ctx).WithContext(ctx).Where("id = ?", id).First(&cost).Error; err != nil {
		return nil, err
	}
	return &cost, nil
}

func (r *costRepository) Save(ctx context.Context, cost *models.Cost) error {
	return r.writeDB(ctx).WithContext(ctx).Save(cost).Error
}

func (r *costRepository) Delete(ctx context.Context, id string) error {
	return r.writeDB(ctx).WithContext(ctx).Delete(&models.Cost{}, "id = ?", id).Error
}

func (r *costRepository) GetByPaymentID(ctx context.Context, paymentID string) ([]*models.Cost, error) {
	var costs []*models.Cost
	if err := r.readDB(ctx).WithContext(ctx).Where("payment_id = ?", paymentID).Find(&costs).Error; err != nil {
		return nil, err
	}
	return costs, nil
}
