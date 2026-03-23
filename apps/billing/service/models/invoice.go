package models

import (
	"time"

	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util/decimalx"
)

// Invoice represents a billing invoice (immutable after issuance).
type Invoice struct {
	data.BaseModel
	BillingRunID    string            `gorm:"type:varchar(50);not null;index"        json:"billing_run_id"`
	ProfileID       string            `gorm:"type:varchar(100);not null;index"       json:"profile_id"`
	SubscriptionID  string            `gorm:"type:varchar(50);not null;index"        json:"subscription_id"`
	InvoiceNumber   string            `gorm:"type:varchar(100);not null;uniqueIndex" json:"invoice_number"`
	State           string            `gorm:"type:varchar(50);not null"              json:"state"`
	Currency        string            `gorm:"type:varchar(10);not null"              json:"currency"`
	SubtotalAmount  *decimalx.Decimal `gorm:"type:numeric(29,9);not null"            json:"subtotal_amount"`
	DiscountAmount  *decimalx.Decimal `gorm:"type:numeric(29,9);not null"            json:"discount_amount"`
	CreditAmount    *decimalx.Decimal `gorm:"type:numeric(29,9);not null"            json:"credit_amount"`
	TotalAmount     *decimalx.Decimal `gorm:"type:numeric(29,9);not null"            json:"total_amount"`
	PeriodStart     time.Time         `gorm:"type:timestamp;not null"                json:"period_start"`
	PeriodEnd       time.Time         `gorm:"type:timestamp;not null"                json:"period_end"`
	IssuedAt        *time.Time        `gorm:"type:timestamp"                         json:"issued_at"`
	DueAt           *time.Time        `gorm:"type:timestamp"                         json:"due_at"`
	PaidAt          *time.Time        `gorm:"type:timestamp"                         json:"paid_at"`
	LedgerTxnID     string            `gorm:"type:varchar(50)"                       json:"ledger_txn_id"`
	CatalogSnapshot data.JSONMap      `gorm:"type:jsonb"                             json:"catalog_snapshot"`
	Data            data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops"   json:"data"`
	Lines           []*InvoiceLine    `gorm:"foreignKey:InvoiceID"                   json:"lines,omitempty"`
}

// InvoiceLine represents a line item on an invoice.
type InvoiceLine struct {
	data.BaseModel
	InvoiceID      string            `gorm:"type:varchar(50);not null;index"      json:"invoice_id"`
	RatedLineID    string            `gorm:"type:varchar(50)"                     json:"rated_line_id"`
	ComponentID    string            `gorm:"type:varchar(50)"                     json:"component_id"`
	Description    string            `gorm:"type:text"                            json:"description"`
	Quantity       *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"quantity"`
	UnitPrice      *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"unit_price"`
	Amount         *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"amount"`
	DiscountAmount *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"discount_amount"`
	CreditAmount   *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"credit_amount"`
	NetAmount      *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"net_amount"`
	Currency       string            `gorm:"type:varchar(10);not null"            json:"currency"`
	LineType       string            `gorm:"type:varchar(50)"                     json:"line_type"`
	Data           data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
