package business

import (
	"context"
	"fmt"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/internal/apperrors"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreditEngine manages prepaid credit grants and consumption.
type CreditEngine interface {
	GrantCredit(ctx context.Context, grant *models.CreditGrant) (*models.CreditGrant, error)
	ApplyCredits(ctx context.Context, customerID string, currency string, amount decimal.Decimal,
		billingRunID string, invoiceID string) (decimal.Decimal, []*models.CreditEntry, error)
	ExpireCredits(ctx context.Context, customerID string, currency string) ([]*models.CreditEntry, error)
	GetBalance(ctx context.Context, customerID string, currency string) (decimal.Decimal, error)
}

type creditEngine struct {
	workMan   workerpool.Manager
	dbPool    pool.Pool
	grantRepo repository.CreditGrantRepository
	entryRepo repository.CreditEntryRepository
}

func NewCreditEngine(
	workMan workerpool.Manager,
	dbPool pool.Pool,
	grantRepo repository.CreditGrantRepository,
	entryRepo repository.CreditEntryRepository,
) CreditEngine {
	return &creditEngine{
		workMan:   workMan,
		dbPool:    dbPool,
		grantRepo: grantRepo,
		entryRepo: entryRepo,
	}
}

func (e *creditEngine) GrantCredit(ctx context.Context, grant *models.CreditGrant) (*models.CreditGrant, error) {
	if grant.CustomerID == "" {
		return nil, ErrCreditCustomerIDRequired
	}
	if !grant.OriginalAmount.Valid || grant.OriginalAmount.Decimal.IsZero() ||
		grant.OriginalAmount.Decimal.IsNegative() {
		return nil, ErrCreditAmountRequired
	}
	if grant.Currency == "" {
		return nil, ErrCreditCurrencyRequired
	}

	grant.GenID(ctx)
	grant.RemainingAmount = grant.OriginalAmount

	// Wrap grant + entry creation in a transaction
	err := e.dbPool.DB(ctx, false).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(grant).Error; err != nil {
			return err
		}

		entry := &models.CreditEntry{
			CreditGrantID: grant.GetID(),
			EntryType:     models.CreditEntryTypeGrant,
			Amount:        grant.OriginalAmount,
			Currency:      grant.Currency,
			Description:   fmt.Sprintf("Credit grant: %s", grant.Name),
		}
		entry.GenID(ctx)

		return tx.Create(entry).Error
	})
	if err != nil {
		return nil, err
	}

	return grant, nil
}

// ApplyCredits consumes credits against an amount, ordered by priority ASC, expires_at ASC.
// Returns the remaining amount after credits and the credit entries created.
func (e *creditEngine) ApplyCredits(
	ctx context.Context,
	customerID string,
	currency string,
	amount decimal.Decimal,
	billingRunID string,
	invoiceID string,
) (decimal.Decimal, []*models.CreditEntry, error) {
	if amount.IsZero() || amount.IsNegative() {
		return amount, nil, nil
	}

	remaining := amount
	var entries []*models.CreditEntry

	err := e.dbPool.DB(ctx, false).Transaction(func(tx *gorm.DB) error {
		// Use SELECT ... FOR UPDATE to lock the grant rows
		var grants []*models.CreditGrant
		result := tx.
			Where(
				"customer_id = ? AND currency = ? AND remaining_amount > 0 AND (expires_at IS NULL OR expires_at > NOW())",
				customerID,
				currency,
			).
			Order("priority ASC, expires_at ASC").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Find(&grants)
		if result.Error != nil {
			return result.Error
		}

		for _, grant := range grants {
			if remaining.IsZero() {
				break
			}

			available := grant.RemainingAmount.Decimal
			if available.IsZero() || available.IsNegative() {
				continue
			}

			consume := decimal.Min(remaining, available)
			grant.RemainingAmount = decimal.NewNullDecimal(available.Sub(consume))

			if err := tx.Model(grant).
				Where("id = ? AND version = ?", grant.GetID(), grant.GetVersion()).
				Updates(grant).Error; err != nil {
				return err
			}

			entry := &models.CreditEntry{
				CreditGrantID: grant.GetID(),
				BillingRunID:  billingRunID,
				InvoiceID:     invoiceID,
				EntryType:     models.CreditEntryTypeConsume,
				Amount:        decimal.NewNullDecimal(consume),
				Currency:      currency,
				Description:   fmt.Sprintf("Credit consumed from: %s", grant.Name),
			}
			entry.GenID(ctx)

			if err := tx.Create(entry).Error; err != nil {
				return err
			}

			entries = append(entries, entry)
			remaining = remaining.Sub(consume)
		}

		return nil
	})
	if err != nil {
		return amount, nil, err
	}

	return remaining, entries, nil
}

func (e *creditEngine) ExpireCredits(
	ctx context.Context,
	customerID string,
	currency string,
) ([]*models.CreditEntry, error) {
	var entries []*models.CreditEntry

	err := e.dbPool.DB(ctx, false).Transaction(func(tx *gorm.DB) error {
		// Use SELECT ... FOR UPDATE to lock the grant rows
		var grants []*models.CreditGrant
		result := tx.
			Where(
				"customer_id = ? AND currency = ? AND remaining_amount > 0 AND expires_at IS NOT NULL AND expires_at <= NOW()",
				customerID,
				currency,
			).
			Order("expires_at ASC").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Find(&grants)
		if result.Error != nil {
			return result.Error
		}

		for _, grant := range grants {
			if grant.RemainingAmount.Decimal.IsZero() {
				continue
			}

			expireAmount := grant.RemainingAmount.Decimal
			grant.RemainingAmount = decimal.NewNullDecimal(decimal.Zero)

			if err := tx.Model(grant).
				Where("id = ? AND version = ?", grant.GetID(), grant.GetVersion()).
				Updates(grant).Error; err != nil {
				return err
			}

			entry := &models.CreditEntry{
				CreditGrantID: grant.GetID(),
				EntryType:     models.CreditEntryTypeExpire,
				Amount:        decimal.NewNullDecimal(expireAmount),
				Currency:      currency,
				Description:   fmt.Sprintf("Credit expired: %s", grant.Name),
			}
			entry.GenID(ctx)

			if err := tx.Create(entry).Error; err != nil {
				return err
			}

			entries = append(entries, entry)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (e *creditEngine) GetBalance(ctx context.Context, customerID string, currency string) (decimal.Decimal, error) {
	if customerID == "" {
		return decimal.Zero, apperrors.ErrUnspecifiedID
	}

	grants, err := e.grantRepo.ListActiveByCustomer(ctx, customerID, currency)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, grant := range grants {
		if grant.RemainingAmount.Valid {
			total = total.Add(grant.RemainingAmount.Decimal)
		}
	}

	return total, nil
}
