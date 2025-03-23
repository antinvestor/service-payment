package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
)

type PaymentStatusRepository interface {
	GetByID(ctx context.Context, id string) (*models.PaymentStatus, error)
	GetByPaymentID(ctx context.Context, paymentId string) ([]models.PaymentStatus, error)
	Save(ctx context.Context, paymentStatus *models.PaymentStatus) error
}

type paymentStatusRepository struct {
	abstractRepository
}

func NewPaymentStatusRepository(ctx context.Context, service *frame.Service) PaymentStatusRepository {
	return &paymentStatusRepository{abstractRepository{service: service}}
}

func (repo *paymentStatusRepository) GetByID(ctx context.Context, id string) (*models.PaymentStatus, error) {
	paymentStatus := models.PaymentStatus{}
	err := repo.readDb(ctx).First(&paymentStatus, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &paymentStatus, nil
}

func (repo *paymentStatusRepository) GetByPaymentID(ctx context.Context, paymentId string) ([]models.PaymentStatus, error) {
	var paymentStatusList []models.PaymentStatus

	err := repo.readDb(ctx).Find(&paymentStatusList,
		"payment_id = ? ", paymentId).Error
	if err != nil {
		return nil, err
	}
	return paymentStatusList, nil
}

func (repo *paymentStatusRepository) Save(ctx context.Context, paymentStatus *models.PaymentStatus) error {
	return repo.writeDb(ctx).Save(paymentStatus).Error
}
