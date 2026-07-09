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
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util/decimalx"
)

// RatedLine represents a priced line item produced by the pricing engine (derived, rebuildable).
type RatedLine struct {
	data.BaseModel
	BillingRunID   string            `gorm:"type:varchar(50);not null;index"      json:"billing_run_id"`
	SubscriptionID string            `gorm:"type:varchar(50);not null;index"      json:"subscription_id"`
	ComponentID    string            `gorm:"type:varchar(50);not null"            json:"component_id"`
	MeteredUsageID string            `gorm:"type:varchar(50)"                     json:"metered_usage_id"`
	Description    string            `gorm:"type:text"                            json:"description"`
	Quantity       *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"quantity"`
	UnitPrice      *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"unit_price"`
	Amount         *decimalx.Decimal `gorm:"type:numeric(29,9);not null"          json:"amount"`
	Currency       string            `gorm:"type:varchar(10);not null"            json:"currency"`
	PricingModel   string            `gorm:"type:varchar(50)"                     json:"pricing_model"`
	TierInfo       data.JSONMap      `gorm:"type:jsonb"                           json:"tier_info"`
	Data           data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
