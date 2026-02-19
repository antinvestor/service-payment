package models

import (
	"time"

	"github.com/pitabwire/frame/data"
	"github.com/shopspring/decimal"
)

// MeteredUsage represents aggregated usage for a billing period (derived, rebuildable).
type MeteredUsage struct {
	data.BaseModel
	SubscriptionID    string              `gorm:"type:varchar(50);not null;index" json:"subscription_id"`
	ComponentID       string              `gorm:"type:varchar(50);not null;index" json:"component_id"`
	MetricKey         string              `gorm:"type:varchar(255);not null"      json:"metric_key"`
	WindowStart       time.Time           `gorm:"type:timestamp;not null"         json:"window_start"`
	WindowEnd         time.Time           `gorm:"type:timestamp;not null"         json:"window_end"`
	WindowGranularity string              `gorm:"type:varchar(50);not null"       json:"window_granularity"`
	AggregationType   string              `gorm:"type:varchar(50);not null"       json:"aggregation_type"`
	Quantity          decimal.NullDecimal `gorm:"type:numeric(29,9);not null"     json:"quantity"`
	EventCount        int64               `gorm:"not null"                        json:"event_count"`
	BillingRunID      string              `gorm:"type:varchar(50);not null;index" json:"billing_run_id"`
}
