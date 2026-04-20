package models

import (
	"time"

	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util/decimalx"
)

// UsageEvent represents a raw usage event (append-only).
type UsageEvent struct {
	data.BaseModel
	EventID        string            `gorm:"type:varchar(255);not null;uniqueIndex" json:"event_id"`
	SubscriptionID string            `gorm:"type:varchar(50);not null;index"        json:"subscription_id"`
	ProfileID      string            `gorm:"type:varchar(100);not null;index"       json:"profile_id"`
	MetricKey      string            `gorm:"type:varchar(255);not null;index"       json:"metric_key"`
	Quantity       *decimalx.Decimal `gorm:"type:numeric(29,9);not null"            json:"quantity"`
	TrueCreatedAt  time.Time         `gorm:"column:true_created_at;type:timestamptz;not null;index" json:"true_created_at"`
	Properties     data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops"   json:"properties"`
}
