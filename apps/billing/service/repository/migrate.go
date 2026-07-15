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

package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/pitabwire/frame/v2/datastore"
)

func Migrate(ctx context.Context, dbManager datastore.Manager, migrationPath string) error {
	dbPool := dbManager.GetPool(ctx, datastore.DefaultMigrationPoolName)

	return dbManager.Migrate(ctx, dbPool, migrationPath,
		&models.CatalogVersion{}, &models.Plan{}, &models.Component{}, &models.Tier{},
		&models.Subscription{},
		&models.IntegrationRoute{},
		&models.UsageEvent{},
		&models.MeteredUsage{},
		&models.RatedLine{},
		&models.Discount{}, &models.DiscountedLine{},
		&models.CreditGrant{}, &models.CreditEntry{},
		&models.Invoice{}, &models.InvoiceLine{},
		&models.BillingRun{})
}
