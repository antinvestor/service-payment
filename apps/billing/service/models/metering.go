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

// MeteredUsage represents aggregated usage for a billing period (derived, rebuildable).
type MeteredUsage struct {
	data.BaseModel
	SubscriptionID    string            `gorm:"type:varchar(50);not null;index" json:"subscription_id"`
	ComponentID       string            `gorm:"type:varchar(50);not null;index" json:"component_id"`
	MetricKey         string            `gorm:"type:varchar(255);not null"      json:"metric_key"`
	WindowStart       time.Time         `gorm:"type:timestamp;not null"         json:"window_start"`
	WindowEnd         time.Time         `gorm:"type:timestamp;not null"         json:"window_end"`
	WindowGranularity string            `gorm:"type:varchar(50);not null"       json:"window_granularity"`
	AggregationType   string            `gorm:"type:varchar(50);not null"       json:"aggregation_type"`
	Quantity          *decimalx.Decimal `gorm:"type:numeric(29,9);not null"     json:"quantity"`
	EventCount        int64             `gorm:"not null"                        json:"event_count"`
	BillingRunID      string            `gorm:"type:varchar(50);not null;index" json:"billing_run_id"`
}
