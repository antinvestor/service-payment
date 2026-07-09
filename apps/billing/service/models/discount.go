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
	"github.com/pitabwire/util/decimalx"
)

// Discount represents a discount rule.
type Discount struct {
	data.BaseModel
	Name            string            `gorm:"type:varchar(255);not null"           json:"name"`
	DiscountType    string            `gorm:"type:varchar(50);not null"            json:"discount_type"`
	Value           *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"value"`
	Currency        string            `gorm:"type:varchar(10)"                     json:"currency"`
	ApplicableTo    data.JSONMap      `gorm:"type:jsonb"                           json:"applicable_to"`
	StartAt         *time.Time        `gorm:"type:timestamp"                       json:"start_at"`
	EndAt           *time.Time        `gorm:"type:timestamp"                       json:"end_at"`
	MaxApplications int               `                                            json:"max_applications"`
	Data            data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}

// DiscountedLine represents a discount applied to a rated line (derived, rebuildable).
type DiscountedLine struct {
	data.BaseModel
	BillingRunID string            `gorm:"type:varchar(50);not null;index"      json:"billing_run_id"`
	RatedLineID  string            `gorm:"type:varchar(50);not null;index"      json:"rated_line_id"`
	DiscountID   string            `gorm:"type:varchar(50);not null;index"      json:"discount_id"`
	Description  string            `gorm:"type:text"                            json:"description"`
	Amount       *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"amount"`
	Currency     string            `gorm:"type:varchar(10);not null"            json:"currency"`
	Data         data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
