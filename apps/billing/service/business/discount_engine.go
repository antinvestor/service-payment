package business

import (
	"context"
	"fmt"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/pitabwire/frame/workerpool"
	"github.com/shopspring/decimal"
)

const percentageDivisor = 100

// DiscountEngine applies discounts to rated lines without mutating them.
type DiscountEngine interface {
	CreateDiscount(ctx context.Context, disc *models.Discount) (*models.Discount, error)
	ApplyDiscounts(
		ctx context.Context,
		ratedLines []*models.RatedLine,
		billingRunID string,
		at time.Time,
	) ([]*models.DiscountedLine, error)
}

type discountEngine struct {
	workMan      workerpool.Manager
	discountRepo repository.DiscountRepository
	discLineRepo repository.DiscountedLineRepository
}

func NewDiscountEngine(
	workMan workerpool.Manager,
	discountRepo repository.DiscountRepository,
	discLineRepo repository.DiscountedLineRepository,
) DiscountEngine {
	return &discountEngine{
		workMan:      workMan,
		discountRepo: discountRepo,
		discLineRepo: discLineRepo,
	}
}

func (e *discountEngine) CreateDiscount(ctx context.Context, disc *models.Discount) (*models.Discount, error) {
	if disc.Name == "" {
		return nil, ErrDiscountNameRequired
	}
	if disc.DiscountType != models.DiscountTypePercentage && disc.DiscountType != models.DiscountTypeFixed {
		return nil, ErrDiscountTypeInvalid
	}
	if !disc.Value.Valid || disc.Value.Decimal.IsZero() || disc.Value.Decimal.IsNegative() {
		return nil, ErrDiscountValueRequired
	}
	if disc.DiscountType == models.DiscountTypePercentage {
		hundred := decimal.NewFromInt(percentageDivisor)
		if disc.Value.Decimal.GreaterThan(hundred) {
			return nil, ErrDiscountPercentageOutOfRange
		}
	}

	disc.GenID(ctx)

	if err := e.discountRepo.Create(ctx, disc); err != nil {
		return nil, err
	}

	return disc, nil
}

func (e *discountEngine) ApplyDiscounts(
	ctx context.Context,
	ratedLines []*models.RatedLine,
	billingRunID string,
	at time.Time,
) ([]*models.DiscountedLine, error) {
	discounts, err := e.discountRepo.ListActive(ctx, at)
	if err != nil {
		return nil, err
	}

	if len(discounts) == 0 {
		return nil, nil
	}

	var discountedLines []*models.DiscountedLine

	for _, rl := range ratedLines {
		for _, disc := range discounts {
			discAmount := calculateDiscount(rl, disc)
			if discAmount.IsZero() {
				continue
			}

			dl := &models.DiscountedLine{
				BillingRunID: billingRunID,
				RatedLineID:  rl.GetID(),
				DiscountID:   disc.GetID(),
				Description:  fmt.Sprintf("Discount: %s", disc.Name),
				Amount:       decimal.NewNullDecimal(discAmount),
				Currency:     rl.Currency,
			}
			dl.GenID(ctx)
			dl.ID = fmt.Sprintf("%s_%s_%s", billingRunID, rl.GetID(), disc.GetID())

			if createErr := e.discLineRepo.Create(ctx, dl); createErr != nil {
				return nil, createErr
			}

			discountedLines = append(discountedLines, dl)
		}
	}

	return discountedLines, nil
}

func calculateDiscount(rl *models.RatedLine, disc *models.Discount) decimal.Decimal {
	amount := rl.Amount.Decimal

	switch disc.DiscountType {
	case models.DiscountTypePercentage:
		pct := disc.Value.Decimal.Div(decimal.NewFromInt(percentageDivisor))
		return amount.Mul(pct)
	case models.DiscountTypeFixed:
		fixed := disc.Value.Decimal
		if fixed.GreaterThan(amount) {
			return amount
		}
		return fixed
	default:
		return decimal.Zero
	}
}
