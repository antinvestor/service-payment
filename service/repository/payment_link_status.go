package repository

import (
	"context"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
)

type PaymentLinkStatusRepository interface {
	GetByID(ctx context.Context, id string) (*models.PaymentLinkStatus, error)
	GetByPaymentLinkID(ctx context.Context, paymentLinkID string) ([]models.PaymentLinkStatus, error)
	Save(ctx context.Context, status *models.PaymentLinkStatus) error
}

type paymentLinkStatusRepository struct {
	abstractRepository
}

func NewPaymentLinkStatusRepository(ctx context.Context, service *frame.Service) PaymentLinkStatusRepository {
	return &paymentLinkStatusRepository{abstractRepository{service: service}}
}

func (repo *paymentLinkStatusRepository) GetByID(ctx context.Context, id string) (*models.PaymentLinkStatus, error) {
	status := models.PaymentLinkStatus{}
	err := repo.readDb(ctx).First(&status, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (repo *paymentLinkStatusRepository) GetByPaymentLinkID(ctx context.Context, paymentLinkID string) ([]models.PaymentLinkStatus, error) {
	var statusList []models.PaymentLinkStatus

	err := repo.readDb(ctx).Find(&statusList, "payment_link_id = ?", paymentLinkID).Error
	if err != nil {
		return nil, err
	}
	return statusList, nil
}

func (repo *paymentLinkStatusRepository) Save(ctx context.Context, status *models.PaymentLinkStatus) error {
	return repo.writeDb(ctx).Save(status).Error
}
