package models

import (
	"time"

	"github.com/pitabwire/frame/data"
)

// Subscription represents a profile's subscription to a plan.
type Subscription struct {
	data.BaseModel
	ProfileID        string       `gorm:"type:varchar(100);not null;index"     json:"profile_id"`
	CatalogVersionID string       `gorm:"type:varchar(50);not null;index"      json:"catalog_version_id"`
	PlanID           string       `gorm:"type:varchar(50);not null;index"      json:"plan_id"`
	State            string       `gorm:"type:varchar(50);not null"            json:"state"`
	StartAt          time.Time    `gorm:"type:timestamp;not null"              json:"start_at"`
	EndAt            *time.Time   `gorm:"type:timestamp"                       json:"end_at"`
	CancelledAt      *time.Time   `gorm:"type:timestamp"                       json:"cancelled_at"`
	BillingAnchor    time.Time    `gorm:"type:timestamp;not null"              json:"billing_anchor"`
	Currency         string       `gorm:"type:varchar(10);not null"            json:"currency"`
	Data             data.JSONMap `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
