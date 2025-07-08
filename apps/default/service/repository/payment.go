package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
)

type PaymentRepository interface {
	GetByPartitionAndID(ctx context.Context, partitionID string, id string) (*models.Payment, error)
	GetByID(ctx context.Context, id string) (*models.Payment, error)
	Search(ctx context.Context, query string) ([]*models.Payment, error)
	Save(ctx context.Context, payment *models.Payment) error
}

type paymentRepository struct {
	abstractRepository
}

func NewPaymentRepository(ctx context.Context, service *frame.Service) PaymentRepository {
	return &paymentRepository{abstractRepository{service: service}}
}

func (repo *paymentRepository) GetByPartitionAndID(ctx context.Context, partitionID string, id string) (*models.Payment, error) {
	payment := models.Payment{}
	err := repo.readDb(ctx).First(&payment, "partition_id = ? AND id = ?", partitionID, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (repo *paymentRepository) GetByID(ctx context.Context, id string) (*models.Payment, error) {
	payment := models.Payment{}
	err := repo.readDb(ctx).First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (repo *paymentRepository) Search(ctx context.Context, query string) ([]*models.Payment, error) {
	query = strings.TrimSpace(query)
	var payments []*models.Payment
	paymentQuery := repo.readDb(ctx)
	if query != "" {
		searchQ := fmt.Sprintf("%%%s%%", query)

		paymentQuery = paymentQuery.
			Where(" id ILIKE ? OR external_id ILIKE ?", searchQ, searchQ)
	}

	err := paymentQuery.Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}
func (repo *paymentRepository) Save(ctx context.Context, payment *models.Payment) error {
	return repo.writeDb(ctx).Save(payment).Error
}
