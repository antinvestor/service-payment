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
	"fmt"
	"sort"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/pitabwire/util/decimalx"
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
	qty := *mu.Quantity

	// Apply free tier
	if comp.FreeQuantity != nil && comp.FreeQuantity.IsPositive() {
		qty = qty.Sub(*comp.FreeQuantity)
		if qty.IsNegative() {
			qty = decimalx.Zero()
		}
	}

	var amount decimalx.Decimal
	var unitPrice decimalx.Decimal
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
	if comp.MinimumCharge != nil && amount.LessThan(*comp.MinimumCharge) {
		amount = *comp.MinimumCharge
		description = fmt.Sprintf("%s (minimum charge applied)", description)
	}

	return &models.RatedLine{
		BillingRunID:   billingRunID,
		SubscriptionID: subscriptionID,
		ComponentID:    comp.GetID(),
		MeteredUsageID: mu.GetID(),
		Description:    description,
		Quantity:       qty.Ptr(),
		UnitPrice:      unitPrice.Ptr(),
		Amount:         amount.Ptr(),
		Currency:       currency,
		PricingModel:   comp.PricingModel,
	}
}

// RateFlat returns a flat fee regardless of usage.
func (pe *PricingEngine) RateFlat(comp *models.Component) (decimalx.Decimal, decimalx.Decimal, string) {
	if len(comp.Tiers) == 0 {
		return decimalx.Zero(), decimalx.Zero(), fmt.Sprintf("%s: flat (no tiers)", comp.Name)
	}

	tier := comp.Tiers[0]
	fee := *tier.FlatFee
	return fee, fee, fmt.Sprintf("%s: flat fee", comp.Name)
}

// RatePerUnit returns quantity * unit price.
func (pe *PricingEngine) RatePerUnit(
	qty decimalx.Decimal,
	comp *models.Component,
) (decimalx.Decimal, decimalx.Decimal, string) {
	if len(comp.Tiers) == 0 {
		return decimalx.Zero(), decimalx.Zero(), fmt.Sprintf("%s: per-unit (no tiers)", comp.Name)
	}

	tier := comp.Tiers[0]
	up := *tier.UnitPrice
	return qty.Mul(up), up, fmt.Sprintf("%s: %s x %s per %s", comp.Name, qty.String(), up.String(), comp.UnitName)
}

// RateTiered applies graduated tiered pricing where each tier prices only the units within its bounds.
func (pe *PricingEngine) RateTiered(
	qty decimalx.Decimal,
	comp *models.Component,
) (decimalx.Decimal, decimalx.Decimal, string) {
	tiers := sortTiers(comp.Tiers)
	total := decimalx.Zero()
	remaining := qty

	for _, tier := range tiers {
		if remaining.IsZero() || remaining.IsNegative() {
			break
		}

		lower := *tier.LowerBound
		upper := decimalx.New(1, infinityExponent) // effectively infinity
		if tier.UpperBound != nil && tier.UpperBound.IsPositive() {
			upper = *tier.UpperBound
		}

		tierWidth := upper.Sub(lower)
		unitsInTier := decimalx.Min(remaining, tierWidth)

		tierAmount := unitsInTier.Mul(*tier.UnitPrice)
		if tier.FlatFee != nil {
			tierAmount = tierAmount.Add(*tier.FlatFee)
		}

		total = total.Add(tierAmount)
		remaining = remaining.Sub(unitsInTier)
	}

	avgUnit := decimalx.Zero()
	if !qty.IsZero() {
		avgUnit = total.Div(qty)
	}

	return total, avgUnit, fmt.Sprintf("%s: tiered %s %s", comp.Name, qty.String(), comp.UnitName)
}

// RateVolume applies volume pricing where the tier is selected based on total quantity
// and ALL units are priced at that tier's rate.
func (pe *PricingEngine) RateVolume(
	qty decimalx.Decimal,
	comp *models.Component,
) (decimalx.Decimal, decimalx.Decimal, string) {
	tiers := sortTiers(comp.Tiers)

	var selectedTier *models.Tier
	for i := range tiers {
		tier := tiers[i]
		lower := *tier.LowerBound
		upper := decimalx.New(1, infinityExponent)
		if tier.UpperBound != nil && tier.UpperBound.IsPositive() {
			upper = *tier.UpperBound
		}

		if !qty.LessThan(lower) && qty.LessThan(upper) {
			selectedTier = tier
			break
		}
	}

	if selectedTier == nil && len(tiers) > 0 {
		selectedTier = tiers[len(tiers)-1]
	}

	if selectedTier == nil {
		return decimalx.Zero(), decimalx.Zero(), fmt.Sprintf("%s: volume (no matching tier)", comp.Name)
	}

	up := *selectedTier.UnitPrice
	total := qty.Mul(up)
	if selectedTier.FlatFee != nil {
		total = total.Add(*selectedTier.FlatFee)
	}

	return total, up, fmt.Sprintf("%s: volume %s %s @ %s", comp.Name, qty.String(), comp.UnitName, up.String())
}

// RateStairstep applies stair-step pricing where a flat fee is charged based on the tier
// the quantity falls into (quantity is not multiplied).
func (pe *PricingEngine) RateStairstep(
	qty decimalx.Decimal,
	comp *models.Component,
) (decimalx.Decimal, decimalx.Decimal, string) {
	tiers := sortTiers(comp.Tiers)

	for _, tier := range tiers {
		lower := *tier.LowerBound
		upper := decimalx.New(1, infinityExponent)
		if tier.UpperBound != nil && tier.UpperBound.IsPositive() {
			upper = *tier.UpperBound
		}

		if !qty.LessThan(lower) && qty.LessThan(upper) {
			fee := *tier.FlatFee
			return fee, fee, fmt.Sprintf("%s: stairstep %s %s", comp.Name, qty.String(), comp.UnitName)
		}
	}

	if len(tiers) > 0 {
		last := tiers[len(tiers)-1]
		fee := *last.FlatFee
		return fee, fee, fmt.Sprintf("%s: stairstep %s %s (max tier)", comp.Name, qty.String(), comp.UnitName)
	}

	return decimalx.Zero(), decimalx.Zero(), fmt.Sprintf("%s: stairstep (no tiers)", comp.Name)
}

func sortTiers(tiers []*models.Tier) []*models.Tier {
	sorted := make([]*models.Tier, len(tiers))
	copy(sorted, tiers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SortOrder < sorted[j].SortOrder
	})
	return sorted
}
