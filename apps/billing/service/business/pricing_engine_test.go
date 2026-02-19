package business_test

import (
	"testing"

	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/pitabwire/frame/data"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newComponent(name, pricingModel string, tiers []*models.Tier) *models.Component {
	comp := &models.Component{
		BaseModel:       data.BaseModel{ID: "comp_" + name},
		Name:            name,
		MetricKey:       "metric_" + name,
		PricingModel:    pricingModel,
		AggregationType: models.AggregationTypeSum,
		UnitName:        "unit",
		Tiers:           tiers,
	}
	return comp
}

func newTier(sortOrder int, lower, upper, unitPrice, flatFee string) *models.Tier {
	t := &models.Tier{
		SortOrder:  sortOrder,
		LowerBound: decimal.NewNullDecimal(decimal.RequireFromString(lower)),
		UnitPrice:  decimal.NewNullDecimal(decimal.RequireFromString(unitPrice)),
	}
	if upper != "" {
		t.UpperBound = decimal.NewNullDecimal(decimal.RequireFromString(upper))
	}
	if flatFee != "" {
		t.FlatFee = decimal.NewNullDecimal(decimal.RequireFromString(flatFee))
	}
	return t
}

func TestPricingEngine_RateFlat(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("flat_api", models.PricingModelFlat, []*models.Tier{
		newTier(0, "0", "", "0", "99.99"),
	})

	amount, unitPrice, desc := pe.RateFlat(comp)
	assert.True(t, amount.Equal(decimal.RequireFromString("99.99")))
	assert.True(t, unitPrice.Equal(decimal.RequireFromString("99.99")))
	assert.Contains(t, desc, "flat fee")
}

func TestPricingEngine_RateFlat_NoTiers(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("flat_empty", models.PricingModelFlat, nil)
	amount, _, desc := pe.RateFlat(comp)
	assert.True(t, amount.IsZero())
	assert.Contains(t, desc, "no tiers")
}

func TestPricingEngine_RatePerUnit(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("api_calls", models.PricingModelPerUnit, []*models.Tier{
		newTier(0, "0", "", "0.01", ""),
	})

	qty := decimal.NewFromInt(1000)
	amount, unitPrice, _ := pe.RatePerUnit(qty, comp)
	assert.True(t, amount.Equal(decimal.RequireFromString("10")))
	assert.True(t, unitPrice.Equal(decimal.RequireFromString("0.01")))
}

func TestPricingEngine_RatePerUnit_ZeroQuantity(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("api_calls", models.PricingModelPerUnit, []*models.Tier{
		newTier(0, "0", "", "0.01", ""),
	})

	amount, _, _ := pe.RatePerUnit(decimal.Zero, comp)
	assert.True(t, amount.IsZero())
}

func TestPricingEngine_RateTiered(t *testing.T) {
	pe := business.NewPricingEngine()

	// Tier 1: 0-100 @ $0.10/unit
	// Tier 2: 100-500 @ $0.05/unit
	// Tier 3: 500+ @ $0.02/unit
	comp := newComponent("storage", models.PricingModelTiered, []*models.Tier{
		newTier(0, "0", "100", "0.10", ""),
		newTier(1, "100", "500", "0.05", ""),
		newTier(2, "500", "", "0.02", ""),
	})

	// 250 units: 100 @ $0.10 + 150 @ $0.05 = $10 + $7.50 = $17.50
	qty := decimal.NewFromInt(250)
	amount, _, _ := pe.RateTiered(qty, comp)
	assert.True(t, amount.Equal(decimal.RequireFromString("17.5")))
}

func TestPricingEngine_RateTiered_WithinFirstTier(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("storage", models.PricingModelTiered, []*models.Tier{
		newTier(0, "0", "100", "0.10", ""),
		newTier(1, "100", "500", "0.05", ""),
	})

	qty := decimal.NewFromInt(50)
	amount, _, _ := pe.RateTiered(qty, comp)
	assert.True(t, amount.Equal(decimal.RequireFromString("5")))
}

func TestPricingEngine_RateVolume(t *testing.T) {
	pe := business.NewPricingEngine()

	// Volume: 0-100 @ $0.10, 100-500 @ $0.05, 500+ @ $0.02
	comp := newComponent("bandwidth", models.PricingModelVolume, []*models.Tier{
		newTier(0, "0", "100", "0.10", ""),
		newTier(1, "100", "500", "0.05", ""),
		newTier(2, "500", "", "0.02", ""),
	})

	// 250 units: all at $0.05 = $12.50
	qty := decimal.NewFromInt(250)
	amount, unitPrice, _ := pe.RateVolume(qty, comp)
	assert.True(t, amount.Equal(decimal.RequireFromString("12.5")))
	assert.True(t, unitPrice.Equal(decimal.RequireFromString("0.05")))
}

func TestPricingEngine_RateStairstep(t *testing.T) {
	pe := business.NewPricingEngine()

	// Stairstep: 0-10 flat $5, 10-50 flat $20, 50+ flat $50
	comp := newComponent("seats", models.PricingModelStairstep, []*models.Tier{
		newTier(0, "0", "10", "0", "5"),
		newTier(1, "10", "50", "0", "20"),
		newTier(2, "50", "", "0", "50"),
	})

	// 25 seats: falls in tier 2, flat fee $20
	qty := decimal.NewFromInt(25)
	amount, _, _ := pe.RateStairstep(qty, comp)
	assert.True(t, amount.Equal(decimal.RequireFromString("20")))
}

func TestPricingEngine_Rate_FreeTier(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("api", models.PricingModelPerUnit, []*models.Tier{
		newTier(0, "0", "", "0.01", ""),
	})
	comp.FreeQuantity = decimal.NewNullDecimal(decimal.NewFromInt(100))

	metered := []*models.MeteredUsage{
		{
			BaseModel:   data.BaseModel{ID: "mu_1"},
			ComponentID: comp.GetID(),
			Quantity:    decimal.NewNullDecimal(decimal.NewFromInt(250)),
		},
	}

	compMap := map[string]*models.Component{comp.GetID(): comp}
	lines := pe.Rate(metered, compMap, "br_1", "sub_1", "USD")

	assert.Len(t, lines, 1)
	// 250 - 100 free = 150 billable @ $0.01 = $1.50
	assert.True(t, lines[0].Amount.Decimal.Equal(decimal.RequireFromString("1.5")))
}

func TestPricingEngine_Rate_FreeTier_UnderFree(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("api", models.PricingModelPerUnit, []*models.Tier{
		newTier(0, "0", "", "0.01", ""),
	})
	comp.FreeQuantity = decimal.NewNullDecimal(decimal.NewFromInt(100))

	metered := []*models.MeteredUsage{
		{
			BaseModel:   data.BaseModel{ID: "mu_1"},
			ComponentID: comp.GetID(),
			Quantity:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
		},
	}

	compMap := map[string]*models.Component{comp.GetID(): comp}
	lines := pe.Rate(metered, compMap, "br_1", "sub_1", "USD")

	assert.Len(t, lines, 1)
	assert.True(t, lines[0].Amount.Decimal.IsZero())
}

func TestPricingEngine_Rate_MinimumCharge(t *testing.T) {
	pe := business.NewPricingEngine()

	comp := newComponent("api", models.PricingModelPerUnit, []*models.Tier{
		newTier(0, "0", "", "0.001", ""),
	})
	comp.MinimumCharge = decimal.NewNullDecimal(decimal.RequireFromString("5.00"))

	metered := []*models.MeteredUsage{
		{
			BaseModel:   data.BaseModel{ID: "mu_1"},
			ComponentID: comp.GetID(),
			Quantity:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
		},
	}

	compMap := map[string]*models.Component{comp.GetID(): comp}
	lines := pe.Rate(metered, compMap, "br_1", "sub_1", "USD")

	assert.Len(t, lines, 1)
	// 100 * 0.001 = $0.10, but minimum is $5.00
	assert.True(t, lines[0].Amount.Decimal.Equal(decimal.RequireFromString("5")))
	assert.Contains(t, lines[0].Description, "minimum charge")
}
