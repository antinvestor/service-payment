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

// UsageEvent represents a raw usage event (append-only).
type UsageEvent struct {
	data.BaseModel
	EventID        string            `gorm:"type:varchar(255);not null;uniqueIndex"                 json:"event_id"`
	SubscriptionID string            `gorm:"type:varchar(50);not null;index"                        json:"subscription_id"`
	ProfileID      string            `gorm:"type:varchar(100);not null;index"                       json:"profile_id"`
	MetricKey      string            `gorm:"type:varchar(255);not null;index"                       json:"metric_key"`
	Quantity       *decimalx.Decimal `gorm:"type:numeric(29,9);not null"                            json:"quantity"`
	TrueCreatedAt  time.Time         `gorm:"column:true_created_at;type:timestamptz;not null;index" json:"true_created_at"`
	Properties     data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops"                   json:"properties"`
}
