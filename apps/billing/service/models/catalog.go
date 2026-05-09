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

// CatalogVersion represents an immutable, versioned catalog of plans and pricing.
type CatalogVersion struct {
	data.BaseModel
	CatalogID   string       `gorm:"type:varchar(100);not null;index"     json:"catalog_id"`
	Version     int          `gorm:"not null"                             json:"version"`
	Name        string       `gorm:"type:varchar(255);not null"           json:"name"`
	Currency    string       `gorm:"type:varchar(10);not null"            json:"currency"`
	PublishedAt *time.Time   `gorm:"type:timestamp"                       json:"published_at"`
	EffectiveAt *time.Time   `gorm:"type:timestamp"                       json:"effective_at"`
	RetiredAt   *time.Time   `gorm:"type:timestamp"                       json:"retired_at"`
	Data        data.JSONMap `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
	Plans       []*Plan      `gorm:"foreignKey:CatalogVersionID"          json:"plans,omitempty"`
}

// Plan represents a billing plan within a catalog version.
type Plan struct {
	data.BaseModel
	CatalogVersionID string       `gorm:"type:varchar(50);not null;index"      json:"catalog_version_id"`
	ExternalID       string       `gorm:"type:varchar(255);index"              json:"external_id"`
	Name             string       `gorm:"type:varchar(255);not null"           json:"name"`
	Description      string       `gorm:"type:text"                            json:"description"`
	Data             data.JSONMap `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
	Components       []*Component `gorm:"foreignKey:PlanID"                    json:"components,omitempty"`
}

// Component represents a billable component within a plan.
type Component struct {
	data.BaseModel
	PlanID          string            `gorm:"type:varchar(50);not null;index"      json:"plan_id"`
	ExternalID      string            `gorm:"type:varchar(255);index"              json:"external_id"`
	Name            string            `gorm:"type:varchar(255);not null"           json:"name"`
	MetricKey       string            `gorm:"type:varchar(255);not null"           json:"metric_key"`
	PricingModel    string            `gorm:"type:varchar(50);not null"            json:"pricing_model"`
	AggregationType string            `gorm:"type:varchar(50);not null"            json:"aggregation_type"`
	UnitName        string            `gorm:"type:varchar(100)"                    json:"unit_name"`
	FreeQuantity    *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"free_quantity"`
	MinimumCharge   *decimalx.Decimal `gorm:"type:numeric(29,9)"                   json:"minimum_charge"`
	Data            data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
	Tiers           []*Tier           `gorm:"foreignKey:ComponentID"               json:"tiers,omitempty"`
}

// Tier represents a pricing tier within a component.
type Tier struct {
	data.BaseModel
	ComponentID string            `gorm:"type:varchar(50);not null;index" json:"component_id"`
	LowerBound  *decimalx.Decimal `gorm:"type:numeric(29,9);not null"     json:"lower_bound"`
	UpperBound  *decimalx.Decimal `gorm:"type:numeric(29,9)"              json:"upper_bound"`
	UnitPrice   *decimalx.Decimal `gorm:"type:numeric(29,9);not null"     json:"unit_price"`
	FlatFee     *decimalx.Decimal `gorm:"type:numeric(29,9)"              json:"flat_fee"`
	SortOrder   int               `gorm:"not null"                        json:"sort_order"`
}
