package models

import (
	"time"

	"github.com/pitabwire/frame/data"
	"github.com/shopspring/decimal"
)

// CreditGrant represents a prepaid credit grant to a customer.
type CreditGrant struct {
	data.BaseModel
	CustomerID      string              `gorm:"type:varchar(100);not null;index"     json:"customer_id"`
	Name            string              `gorm:"type:varchar(255);not null"           json:"name"`
	OriginalAmount  decimal.NullDecimal `gorm:"type:numeric(29,9);not null"          json:"original_amount"`
	RemainingAmount decimal.NullDecimal `gorm:"type:numeric(29,9);not null"          json:"remaining_amount"`
	Currency        string              `gorm:"type:varchar(10);not null"            json:"currency"`
	ExpiresAt       *time.Time          `gorm:"type:timestamp"                       json:"expires_at"`
	Priority        int                 `gorm:"not null;default:0"                   json:"priority"`
	Data            data.JSONMap        `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}

// CreditEntry represents a ledger entry for credit operations.
type CreditEntry struct {
	data.BaseModel
	CreditGrantID string              `gorm:"type:varchar(50);not null;index"      json:"credit_grant_id"`
	BillingRunID  string              `gorm:"type:varchar(50);index"               json:"billing_run_id"`
	InvoiceID     string              `gorm:"type:varchar(50);index"               json:"invoice_id"`
	EntryType     string              `gorm:"type:varchar(50);not null"            json:"entry_type"`
	Amount        decimal.NullDecimal `gorm:"type:numeric(29,9);not null"          json:"amount"`
	Currency      string              `gorm:"type:varchar(10);not null"            json:"currency"`
	LedgerTxnID   string              `gorm:"type:varchar(50)"                     json:"ledger_txn_id"`
	Description   string              `gorm:"type:text"                            json:"description"`
	Data          data.JSONMap        `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
