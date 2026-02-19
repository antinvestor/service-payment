package models

import (
	"github.com/pitabwire/frame/data"
	"github.com/shopspring/decimal"
)

// RatedLine represents a priced line item produced by the pricing engine (derived, rebuildable).
type RatedLine struct {
	data.BaseModel
	BillingRunID   string              `gorm:"type:varchar(50);not null;index"      json:"billing_run_id"`
	SubscriptionID string              `gorm:"type:varchar(50);not null;index"      json:"subscription_id"`
	ComponentID    string              `gorm:"type:varchar(50);not null"            json:"component_id"`
	MeteredUsageID string              `gorm:"type:varchar(50)"                     json:"metered_usage_id"`
	Description    string              `gorm:"type:text"                            json:"description"`
	Quantity       decimal.NullDecimal `gorm:"type:numeric(29,9)"                   json:"quantity"`
	UnitPrice      decimal.NullDecimal `gorm:"type:numeric(29,9)"                   json:"unit_price"`
	Amount         decimal.NullDecimal `gorm:"type:numeric(29,9);not null"          json:"amount"`
	Currency       string              `gorm:"type:varchar(10);not null"            json:"currency"`
	PricingModel   string              `gorm:"type:varchar(50)"                     json:"pricing_model"`
	TierInfo       data.JSONMap        `gorm:"type:jsonb"                           json:"tier_info"`
	Data           data.JSONMap        `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
