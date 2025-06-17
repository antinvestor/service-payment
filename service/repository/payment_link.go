package repository

import (
	"context"
	"strings"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
)

type PaymentLinkRepository interface {
	GetByID(ctx context.Context, id string) (*models.PaymentLink, error)
	GetByPartitionAndID(ctx context.Context, partitionID string, id string) (*models.PaymentLink, error)
	Search(ctx context.Context, query string) ([]*models.PaymentLink, error)
	Save(ctx context.Context, link *models.PaymentLink) error
}

type paymentLinkRepository struct {
	abstractRepository
}

func NewPaymentLinkRepository(ctx context.Context, service *frame.Service) PaymentLinkRepository {
	return &paymentLinkRepository{abstractRepository{service: service}}
}

func (repo *paymentLinkRepository) GetByID(ctx context.Context, id string) (*models.PaymentLink, error) {
	link := models.PaymentLink{}
	err := repo.readDb(ctx).First(&link, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (repo *paymentLinkRepository) GetByPartitionAndID(ctx context.Context, partitionID string, id string) (*models.PaymentLink, error) {
	link := models.PaymentLink{}
	err := repo.readDb(ctx).First(&link, "partition_id = ? AND id = ?", partitionID, id).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (repo *paymentLinkRepository) Search(ctx context.Context, query string) ([]*models.PaymentLink, error) {
	var links []*models.PaymentLink
	err := repo.readDb(ctx).Where("name ILIKE ?", "%"+strings.ToLower(query)+"%").Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (repo *paymentLinkRepository) Save(ctx context.Context, link *models.PaymentLink) error {
	return repo.writeDb(ctx).Save(link).Error
}
