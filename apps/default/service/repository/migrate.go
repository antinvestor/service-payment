package repository

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/util"
)

// Migrate runs database migrations for all payment service models.
func Migrate(ctx context.Context, dsMan datastore.Manager, migrationPath string) error {
	logger := util.Log(ctx).WithField("function", "Migrate")

	if migrationPath == "" {
		return errors.New("migration path cannot be empty")
	}

	logger.WithField("migration_path", migrationPath).Info("Starting database migration")

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
		logger.WithError(err).Error("Failed to run auto migration")
		return err
	}

	logger.Info("Database migration completed successfully")
	return nil
}
