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
	"errors"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/util"
)

// Migrate runs database migrations for all payment service models.
func Migrate(ctx context.Context, dsMan datastore.Manager, migrationPath string) error {
	logger := util.Log(ctx).WithField("function", "Migrate")

	if migrationPath == "" {
		return errors.New("migration path cannot be empty")
	}

	logger.WithField("migration_path", migrationPath).Debug("starting database migration")

	// Get the default database pool
	pool := dsMan.GetPool(ctx, datastore.DefaultPoolName)
	if pool == nil {
		return errors.New("could not get database pool")
	}

	// Run migrations for all dbModels
	dbModels := []any{
		&models.Payment{},
		&models.Cost{},
		&models.Status{},
		&models.Route{},
		&models.Account{},
		&models.Prompt{},
		&models.PaymentLink{},
	}

	db := pool.DB(ctx, false)
	if db == nil {
		return errors.New("could not get database connection")
	}

	// AutoMigrate all dbModels
	err := db.AutoMigrate(dbModels...)
	if err != nil {
		logger.WithError(err).Error("failed to run auto migration")
		return err
	}

	logger.Debug("database migration completed successfully")
	return nil
}
