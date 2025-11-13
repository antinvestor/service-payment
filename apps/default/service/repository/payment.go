package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/workerpool"
	"gorm.io/gorm/clause"
)

type paymentRepository struct {
	datastore.BaseRepository[*models.Payment]
}

func NewPaymentRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) PaymentRepository {
	repo := paymentRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Payment](
			ctx, dbPool, workMan, func() *models.Payment { return &models.Payment{} },
		),
	}
	return &repo
}

func (pr *paymentRepository) GetByPartitionAndID(
	ctx context.Context,
	partitionID string,
	id string,
) (*models.Payment, error) {
	payment := models.Payment{}
	err := pr.Pool().DB(ctx, true).First(&payment, "partition_id = ? AND id = ?", partitionID, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (pr *paymentRepository) GetByProfileID(ctx context.Context, profileID string) ([]*models.Payment, error) {
	var payments []*models.Payment
	err := pr.Pool().DB(ctx, true).
		Where("sender_profile_id = ? OR recipient_profile_id = ?", profileID, profileID).
		Find(&payments).Error
	return payments, err
}

func (pr *paymentRepository) Search(ctx context.Context, query *data.SearchQuery) (workerpool.JobResultPipe[[]*models.Payment], error) {
	return pr.BaseSearch(ctx, query, func(db *gorm.DB) *gorm.DB {
		if query.Query != "" {
			searchQ := fmt.Sprintf("%%%s%%", strings.TrimSpace(query.Query))
			db = db.Where("id ILIKE ? OR reference_id ILIKE ? OR transaction_id ILIKE ?", 
				searchQ, searchQ, searchQ)
		}
		
		// Add profile filter if provided
		if profileID, ok := query.Filters["profile_id"]; ok && profileID != "" {
			db = db.Where("sender_profile_id = ? OR recipient_profile_id = ?", profileID, profileID)
		}
		
		return db
	})
}

// Legacy methods for backward compatibility
func (pr *paymentRepository) SearchLegacy(ctx context.Context, query string) ([]*models.Payment, error) {
	query = strings.TrimSpace(query)
	var payments []*models.Payment
	paymentQuery := pr.Pool().DB(ctx, true)
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
