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

// Subscription represents a profile's subscription to a plan.
type Subscription struct {
	data.BaseModel
	ProfileID        string       `gorm:"type:varchar(100);not null;index"     json:"profile_id"`
	CatalogVersionID string       `gorm:"type:varchar(50);not null;index"      json:"catalog_version_id"`
	PlanID           string       `gorm:"type:varchar(50);not null;index"      json:"plan_id"`
	State            string       `gorm:"type:varchar(50);not null"            json:"state"`
	StartAt          time.Time    `gorm:"type:timestamp;not null"              json:"start_at"`
	EndAt            *time.Time   `gorm:"type:timestamp"                       json:"end_at"`
	CancelledAt      *time.Time   `gorm:"type:timestamp"                       json:"cancelled_at"`
	BillingAnchor    time.Time    `gorm:"type:timestamp;not null"              json:"billing_anchor"`
	Currency         string       `gorm:"type:varchar(10);not null"            json:"currency"`
	Data             data.JSONMap `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}
