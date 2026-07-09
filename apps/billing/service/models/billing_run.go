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

// BillingRun represents a billing workflow execution (state machine).
type BillingRun struct {
	data.BaseModel
	SubscriptionID   string       `gorm:"type:varchar(50);not null;index"        json:"subscription_id"`
	ProfileID        string       `gorm:"type:varchar(100);not null;index"       json:"profile_id"`
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
