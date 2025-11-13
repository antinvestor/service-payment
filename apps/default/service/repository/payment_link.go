package repository

import (
	"context"
	"strings"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

type paymentLinkRepository struct {
	datastore.BaseRepository[*models.PaymentLink]
}

func NewPaymentLinkRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) PaymentLinkRepository {
	repo := paymentLinkRepository{
		BaseRepository: datastore.NewBaseRepository[*models.PaymentLink](
			ctx, dbPool, workMan, func() *models.PaymentLink { return &models.PaymentLink{} },
		),
	}
	return &repo
}

func (plr *paymentLinkRepository) GetByPartitionAndID(
	ctx context.Context,
	partitionID string,
	id string,
) (*models.PaymentLink, error) {
	link := &models.PaymentLink{}
	err := plr.Pool().DB(ctx, true).First(link, "partition_id = ? AND id = ?", partitionID, id).Error
	return link, err
}

func (plr *paymentLinkRepository) GetByProfileID(ctx context.Context, profileID string) ([]*models.PaymentLink, error) {
	// Note: PaymentLink doesn't have direct profile association, 
	// this method is included for interface consistency
	var links []*models.PaymentLink
	err := plr.Pool().DB(ctx, true).Find(&links).Error
	return links, err
}

// Legacy method for backward compatibility
func (plr *paymentLinkRepository) SearchLegacy(ctx context.Context, query string) ([]*models.PaymentLink, error) {
	var links []*models.PaymentLink
	err := plr.Pool().DB(ctx, true).Where("name ILIKE ?", "%"+strings.ToLower(query)+"%").Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}
