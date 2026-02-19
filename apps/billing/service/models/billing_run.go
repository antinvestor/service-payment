package models

import (
	"time"

	"github.com/pitabwire/frame/data"
)

// BillingRun represents a billing workflow execution (state machine).
type BillingRun struct {
	data.BaseModel
	SubscriptionID   string       `gorm:"type:varchar(50);not null;index"        json:"subscription_id"`
	CustomerID       string       `gorm:"type:varchar(100);not null;index"       json:"customer_id"`
	CatalogVersionID string       `gorm:"type:varchar(50);not null"              json:"catalog_version_id"`
	State            string       `gorm:"type:varchar(50);not null"              json:"state"`
	PeriodStart      time.Time    `gorm:"type:timestamp;not null"                json:"period_start"`
	PeriodEnd        time.Time    `gorm:"type:timestamp;not null"                json:"period_end"`
	StartedAt        *time.Time   `gorm:"type:timestamp"                         json:"started_at"`
	CompletedAt      *time.Time   `gorm:"type:timestamp"                         json:"completed_at"`
	FailedAt         *time.Time   `gorm:"type:timestamp"                         json:"failed_at"`
	ErrorMessage     string       `gorm:"type:text"                              json:"error_message"`
	InvoiceID        string       `gorm:"type:varchar(50)"                       json:"invoice_id"`
	Idempotency      string       `gorm:"type:varchar(255);not null;uniqueIndex" json:"idempotency"`
	Data             data.JSONMap `gorm:"type:jsonb;index:,gin:jsonb_path_ops"   json:"data"`
}
