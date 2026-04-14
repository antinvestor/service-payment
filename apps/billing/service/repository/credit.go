package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

// CreditGrantRepository provides operations for credit grants.
type CreditGrantRepository interface {
	datastore.BaseRepository[*models.CreditGrant]
	ListActiveByProfile(ctx context.Context, profileID string, currency string) ([]*models.CreditGrant, error)
	ListExpirableByProfile(ctx context.Context, profileID string, currency string) ([]*models.CreditGrant, error)
}

type creditGrantRepository struct {
	datastore.BaseRepository[*models.CreditGrant]
}

func NewCreditGrantRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) CreditGrantRepository {
	return &creditGrantRepository{
		BaseRepository: datastore.NewBaseRepository[*models.CreditGrant](
			ctx, dbPool, workMan, func() *models.CreditGrant { return &models.CreditGrant{} },
		),
	}
}

func (r *creditGrantRepository) ListActiveByProfile(
	ctx context.Context,
	profileID string,
	currency string,
) ([]*models.CreditGrant, error) {
	if profileID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.CreditGrant
	result := r.Pool().DB(ctx, true).
		Where("profile_id = ? AND currency = ? AND remaining_amount > 0 AND (expires_at IS NULL OR expires_at > NOW())",
			profileID, currency).
		Order("priority ASC, expires_at ASC").
		Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}

func (r *creditGrantRepository) ListExpirableByProfile(
	ctx context.Context,
	profileID string,
	currency string,
) ([]*models.CreditGrant, error) {
	if profileID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.CreditGrant
	result := r.Pool().DB(ctx, true).
		Where("profile_id = ? AND currency = ? AND remaining_amount > 0 AND expires_at IS NOT NULL AND expires_at <= NOW()",
			profileID, currency).
		Order("expires_at ASC").
		Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}

// CreditEntryRepository provides operations for credit entries.
type CreditEntryRepository interface {
	datastore.BaseRepository[*models.CreditEntry]
	ListByGrantID(ctx context.Context, grantID string) ([]*models.CreditEntry, error)
}

type creditEntryRepository struct {
	datastore.BaseRepository[*models.CreditEntry]
}

func NewCreditEntryRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) CreditEntryRepository {
	return &creditEntryRepository{
		BaseRepository: datastore.NewBaseRepository[*models.CreditEntry](
			ctx, dbPool, workMan, func() *models.CreditEntry { return &models.CreditEntry{} },
		),
	}
}

func (r *creditEntryRepository) ListByGrantID(ctx context.Context, grantID string) ([]*models.CreditEntry, error) {
	if grantID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.CreditEntry
	result := r.Pool().DB(ctx, true).Where("credit_grant_id = ?", grantID).Order("created_at ASC").Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}
