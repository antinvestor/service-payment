package business

import (
	"fmt"
	"sort"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/shopspring/decimal"
)

// infinityExponent is the exponent used to represent effectively infinite upper bounds in tier pricing.
const infinityExponent = 18

// PricingEngine is a pure, deterministic, side-effect free pricing calculator.
type PricingEngine struct{}

func NewPricingEngine() *PricingEngine {
	return &PricingEngine{}
}

// Rate prices all metered usage against their components.
func (pe *PricingEngine) Rate(
	metered []*models.MeteredUsage,
	components map[string]*models.Component,
	billingRunID string,
	subscriptionID string,
	currency string,
) []*models.RatedLine {
	var lines []*models.RatedLine

	for _, mu := range metered {
		comp, ok := components[mu.ComponentID]
		if !ok {
			continue
		}

		line := pe.rateComponent(mu, comp, billingRunID, subscriptionID, currency)
		if line != nil {
			lines = append(lines, line)
		}
	}

	return lines
}

func (pe *PricingEngine) rateComponent(
	mu *models.MeteredUsage,
	comp *models.Component,
	billingRunID string,
	subscriptionID string,
	currency string,
) *models.RatedLine {
	qty := mu.Quantity.Decimal

	// Apply free tier
	if comp.FreeQuantity.Valid && comp.FreeQuantity.Decimal.IsPositive() {
		qty = qty.Sub(comp.FreeQuantity.Decimal)
		if qty.IsNegative() {
			qty = decimal.Zero
		}
	}

	var amount decimal.Decimal
	var unitPrice decimal.Decimal
	var description string

	switch comp.PricingModel {
	case models.PricingModelFlat:
		amount, unitPrice, description = pe.RateFlat(comp)
	case models.PricingModelPerUnit:
		amount, unitPrice, description = pe.RatePerUnit(qty, comp)
	case models.PricingModelTiered:
		amount, unitPrice, description = pe.RateTiered(qty, comp)
	case models.PricingModelVolume:
		amount, unitPrice, description = pe.RateVolume(qty, comp)
	case models.PricingModelStairstep:
		amount, unitPrice, description = pe.RateStairstep(qty, comp)
	default:
		return nil
	}

	// Apply minimum charge
	if comp.MinimumCharge.Valid && amount.LessThan(comp.MinimumCharge.Decimal) {
		amount = comp.MinimumCharge.Decimal
		description = fmt.Sprintf("%s (minimum charge applied)", description)
	}

	return &models.RatedLine{
		BillingRunID:   billingRunID,
		SubscriptionID: subscriptionID,
		ComponentID:    comp.GetID(),
		MeteredUsageID: mu.GetID(),
		Description:    description,
		Quantity:       decimal.NewNullDecimal(qty),
		UnitPrice:      decimal.NewNullDecimal(unitPrice),
		Amount:         decimal.NewNullDecimal(amount),
		Currency:       currency,
		PricingModel:   comp.PricingModel,
	}
}

// RateFlat returns a flat fee regardless of usage.
func (pe *PricingEngine) RateFlat(comp *models.Component) (decimal.Decimal, decimal.Decimal, string) {
	if len(comp.Tiers) == 0 {
		return decimal.Zero, decimal.Zero, fmt.Sprintf("%s: flat (no tiers)", comp.Name)
	}

	tier := comp.Tiers[0]
	fee := tier.FlatFee.Decimal
	return fee, fee, fmt.Sprintf("%s: flat fee", comp.Name)
}

// RatePerUnit returns quantity * unit price.
func (pe *PricingEngine) RatePerUnit(
	qty decimal.Decimal,
	comp *models.Component,
) (decimal.Decimal, decimal.Decimal, string) {
	if len(comp.Tiers) == 0 {
		return decimal.Zero, decimal.Zero, fmt.Sprintf("%s: per-unit (no tiers)", comp.Name)
	}

	tier := comp.Tiers[0]
	up := tier.UnitPrice.Decimal
	return qty.Mul(up), up, fmt.Sprintf("%s: %s x %s per %s", comp.Name, qty.String(), up.String(), comp.UnitName)
}

// RateTiered applies graduated tiered pricing where each tier prices only the units within its bounds.
func (pe *PricingEngine) RateTiered(
	qty decimal.Decimal,
	comp *models.Component,
) (decimal.Decimal, decimal.Decimal, string) {
	tiers := sortTiers(comp.Tiers)
	total := decimal.Zero
	remaining := qty

	for _, tier := range tiers {
		if remaining.IsZero() || remaining.IsNegative() {
			break
		}

		lower := tier.LowerBound.Decimal
		upper := decimal.New(1, infinityExponent) // effectively infinity
		if tier.UpperBound.Valid && tier.UpperBound.Decimal.IsPositive() {
			upper = tier.UpperBound.Decimal
		}

		tierWidth := upper.Sub(lower)
		unitsInTier := decimal.Min(remaining, tierWidth)

		tierAmount := unitsInTier.Mul(tier.UnitPrice.Decimal)
		if tier.FlatFee.Valid {
			tierAmount = tierAmount.Add(tier.FlatFee.Decimal)
		}

		total = total.Add(tierAmount)
		remaining = remaining.Sub(unitsInTier)
	}

	avgUnit := decimal.Zero
	if !qty.IsZero() {
		avgUnit = total.Div(qty)
	}

	return total, avgUnit, fmt.Sprintf("%s: tiered %s %s", comp.Name, qty.String(), comp.UnitName)
}

// RateVolume applies volume pricing where the tier is selected based on total quantity
// and ALL units are priced at that tier's rate.
func (pe *PricingEngine) RateVolume(
	qty decimal.Decimal,
	comp *models.Component,
) (decimal.Decimal, decimal.Decimal, string) {
	tiers := sortTiers(comp.Tiers)

	var selectedTier *models.Tier
	for i := range tiers {
		tier := tiers[i]
		lower := tier.LowerBound.Decimal
		upper := decimal.New(1, infinityExponent)
		if tier.UpperBound.Valid && tier.UpperBound.Decimal.IsPositive() {
			upper = tier.UpperBound.Decimal
		}

		if qty.GreaterThanOrEqual(lower) && qty.LessThan(upper) {
			selectedTier = tier
			break
		}
	}

	if selectedTier == nil && len(tiers) > 0 {
		selectedTier = tiers[len(tiers)-1]
	}

	if selectedTier == nil {
		return decimal.Zero, decimal.Zero, fmt.Sprintf("%s: volume (no matching tier)", comp.Name)
	}

	up := selectedTier.UnitPrice.Decimal
	total := qty.Mul(up)
	if selectedTier.FlatFee.Valid {
		total = total.Add(selectedTier.FlatFee.Decimal)
	}

	return total, up, fmt.Sprintf("%s: volume %s %s @ %s", comp.Name, qty.String(), comp.UnitName, up.String())
}

// RateStairstep applies stair-step pricing where a flat fee is charged based on the tier
// the quantity falls into (quantity is not multiplied).
func (pe *PricingEngine) RateStairstep(
	qty decimal.Decimal,
	comp *models.Component,
) (decimal.Decimal, decimal.Decimal, string) {
	tiers := sortTiers(comp.Tiers)

	for _, tier := range tiers {
		lower := tier.LowerBound.Decimal
		upper := decimal.New(1, infinityExponent)
		if tier.UpperBound.Valid && tier.UpperBound.Decimal.IsPositive() {
			upper = tier.UpperBound.Decimal
		}

		if qty.GreaterThanOrEqual(lower) && qty.LessThan(upper) {
			fee := tier.FlatFee.Decimal
			return fee, fee, fmt.Sprintf("%s: stairstep %s %s", comp.Name, qty.String(), comp.UnitName)
		}
	}

	if len(tiers) > 0 {
		last := tiers[len(tiers)-1]
		fee := last.FlatFee.Decimal
		return fee, fee, fmt.Sprintf("%s: stairstep %s %s (max tier)", comp.Name, qty.String(), comp.UnitName)
	}

	return decimal.Zero, decimal.Zero, fmt.Sprintf("%s: stairstep (no tiers)", comp.Name)
}

func sortTiers(tiers []*models.Tier) []*models.Tier {
	sorted := make([]*models.Tier, len(tiers))
	copy(sorted, tiers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SortOrder < sorted[j].SortOrder
	})
	return sorted
}
