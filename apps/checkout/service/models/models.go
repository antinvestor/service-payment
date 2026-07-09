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

	"github.com/pitabwire/frame/v2/data"
)

// Session statuses.
const (
	SessionStatusPending    = "pending"
	SessionStatusProcessing = "processing"
	SessionStatusCompleted  = "completed"
	SessionStatusFailed     = "failed"
	SessionStatusExpired    = "expired"
)

// Amount options.
const (
	AmountOptionFixed    = "fixed"
	AmountOptionVariable = "variable"
)

// CheckoutSession is a single-payer, single-use payment session.
type CheckoutSession struct {
	data.BaseModel

	Ref            string       `gorm:"type:varchar(64);uniqueIndex;not null"             json:"ref"`
	LinkID         string       `gorm:"type:varchar(50);index"                            json:"link_id"`
	Name           string       `gorm:"type:varchar(100);not null"                        json:"name"`
	Description    string       `gorm:"type:varchar(500)"                                 json:"description"`
	Amount         string       `gorm:"type:varchar(40)"                                  json:"amount"` // decimal string
	Currency       string       `gorm:"type:varchar(10);not null"                         json:"currency"`
	AmountOption   string       `gorm:"type:varchar(20);not null;default:'fixed'"         json:"amount_option"`
	OrderRef       string       `gorm:"type:varchar(250);index"                           json:"order_ref"`
	Metadata       data.JSONMap `gorm:"type:jsonb"                                        json:"metadata"`
	ReturnURL      string       `gorm:"type:varchar(500)"                                 json:"return_url"`
	PayerProfileID string       `gorm:"type:varchar(50);index"                            json:"payer_profile_id"`
	Prefill        data.JSONMap `gorm:"type:jsonb"                                        json:"prefill"`
	Methods        data.JSONMap `gorm:"type:jsonb"                                        json:"methods"` // {"keys": [...]} restriction
	PromptID       string       `gorm:"type:varchar(50);index"                            json:"prompt_id"`
	PaymentID      string       `gorm:"type:varchar(50)"                                  json:"payment_id"`
	Attempts       int          `gorm:"not null;default:0"                                json:"attempts"`
	LastAttemptAt  *time.Time   `                                                         json:"last_attempt_at"`
	Status         string       `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	ExpiresAt      time.Time    `gorm:"index"                                             json:"expires_at"`
}

func (*CheckoutSession) TableName() string { return "checkout_sessions" }

// IsTerminal reports whether the session can no longer accept payment.
func (s *CheckoutSession) IsTerminal() bool {
	return s.Status == SessionStatusCompleted || s.Status == SessionStatusExpired
}

// CheckoutLink is a reusable template that spawns CheckoutSessions.
type CheckoutLink struct {
	data.BaseModel

	Ref          string       `gorm:"type:varchar(64);uniqueIndex;not null"     json:"ref"`
	Name         string       `gorm:"type:varchar(100);not null"                json:"name"`
	Description  string       `gorm:"type:varchar(500)"                         json:"description"`
	Amount       string       `gorm:"type:varchar(40)"                          json:"amount"`
	Currency     string       `gorm:"type:varchar(10);not null"                 json:"currency"`
	AmountOption string       `gorm:"type:varchar(20);not null;default:'fixed'" json:"amount_option"`
	OrderRef     string       `gorm:"type:varchar(250)"                         json:"order_ref"`
	Metadata     data.JSONMap `gorm:"type:jsonb"                                json:"metadata"`
	ReturnURL    string       `gorm:"type:varchar(500)"                         json:"return_url"`
	ExpiresAt    *time.Time   `                                                 json:"expires_at"`
	Active       bool         `gorm:"not null;default:true"                     json:"active"`
}

func (*CheckoutLink) TableName() string { return "checkout_links" }

// IsUsable reports whether new sessions may be spawned from this link.
func (l *CheckoutLink) IsUsable(now time.Time) bool {
	if !l.Active {
		return false
	}
	if l.ExpiresAt != nil && now.After(*l.ExpiresAt) {
		return false
	}
	return true
}
