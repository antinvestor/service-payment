// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package business

import (
	"context"
	"fmt"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
	"github.com/pitabwire/util/decimalx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreditEngine manages prepaid credit grants and consumption.
type CreditEngine interface {
	GrantCredit(ctx context.Context, grant *models.CreditGrant) (*models.CreditGrant, error)
	ApplyCredits(ctx context.Context, profileID string, currency string, amount decimalx.Decimal,
		billingRunID string, invoiceID string) (decimalx.Decimal, []*models.CreditEntry, error)
	ExpireCredits(ctx context.Context, profileID string, currency string) ([]*models.CreditEntry, error)
	GetBalance(ctx context.Context, profileID string, currency string) (decimalx.Decimal, error)
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
	if grant.ProfileID == "" {
		return nil, ErrCreditProfileIDRequired
	}
	if grant.OriginalAmount == nil || grant.OriginalAmount.IsZero() ||
		grant.OriginalAmount.IsNegative() {
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
	profileID string,
	currency string,
	amount decimalx.Decimal,
	billingRunID string,
	invoiceID string,
) (decimalx.Decimal, []*models.CreditEntry, error) {
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
				"profile_id = ? AND currency = ? AND remaining_amount > 0 AND (expires_at IS NULL OR expires_at > NOW())",
				profileID,
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

			available := decimalx.DerefOr(grant.RemainingAmount, decimalx.Zero())
			if available.IsZero() || available.IsNegative() {
				continue
			}

			consume := decimalx.Min(remaining, available)
			sub := available.Sub(consume)
			grant.RemainingAmount = &sub

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
				Amount:        consume.Ptr(),
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
	profileID string,
	currency string,
) ([]*models.CreditEntry, error) {
	var entries []*models.CreditEntry

	err := e.dbPool.DB(ctx, false).Transaction(func(tx *gorm.DB) error {
		// Use SELECT ... FOR UPDATE to lock the grant rows
		var grants []*models.CreditGrant
		result := tx.
			Where(
				"profile_id = ? AND currency = ? AND remaining_amount > 0 AND expires_at IS NOT NULL AND expires_at <= NOW()",
				profileID,
				currency,
			).
			Order("expires_at ASC").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Find(&grants)
		if result.Error != nil {
			return result.Error
		}

		for _, grant := range grants {
			remainingAmt := decimalx.DerefOr(grant.RemainingAmount, decimalx.Zero())
			if remainingAmt.IsZero() {
				continue
			}

			expireAmount := remainingAmt
			zero := decimalx.Zero()
			grant.RemainingAmount = &zero

			if err := tx.Model(grant).
				Where("id = ? AND version = ?", grant.GetID(), grant.GetVersion()).
				Updates(grant).Error; err != nil {
				return err
			}

			entry := &models.CreditEntry{
				CreditGrantID: grant.GetID(),
				EntryType:     models.CreditEntryTypeExpire,
				Amount:        expireAmount.Ptr(),
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

func (e *creditEngine) GetBalance(ctx context.Context, profileID string, currency string) (decimalx.Decimal, error) {
	if profileID == "" {
		return decimalx.Zero(), apperrors.ErrUnspecifiedID
	}

	grants, err := e.grantRepo.ListActiveByProfile(ctx, profileID, currency)
	if err != nil {
		return decimalx.Zero(), err
	}

	total := decimalx.Zero()
	for _, grant := range grants {
		if grant.RemainingAmount != nil {
			total = total.Add(*grant.RemainingAmount)
		}
	}

	return total, nil
}
