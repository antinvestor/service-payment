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

package models

import (
	"time"

	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util/decimalx"
)

// CreditGrant represents a prepaid credit grant to a profile.
type CreditGrant struct {
	data.BaseModel
	ProfileID       string            `gorm:"type:varchar(100);not null;index"     json:"profile_id"`
	Name            string            `gorm:"type:varchar(255);not null"           json:"name"`
	OriginalAmount  *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"original_amount"`
	RemainingAmount *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"remaining_amount"`
	Currency        string            `gorm:"type:varchar(10);not null"            json:"currency"`
	ExpiresAt       *time.Time        `gorm:"type:timestamp"                       json:"expires_at"`
	Priority        int               `gorm:"not null;default:0"                   json:"priority"`
	Data            data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}

// CreditEntry represents a ledger entry for credit operations.
type CreditEntry struct {
	data.BaseModel
	CreditGrantID string            `gorm:"type:varchar(50);not null;index"      json:"credit_grant_id"`
	BillingRunID  string            `gorm:"type:varchar(50);index"               json:"billing_run_id"`
	InvoiceID     string            `gorm:"type:varchar(50);index"               json:"invoice_id"`
	EntryType     string            `gorm:"type:varchar(50);not null"            json:"entry_type"`
	Amount        *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"amount"`
	Currency      string            `gorm:"type:varchar(10);not null"            json:"currency"`
	LedgerTxnID   string            `gorm:"type:varchar(50)"                     json:"ledger_txn_id"`
	Description   string            `gorm:"type:text"                            json:"description"`
	Data          data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
