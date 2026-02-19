package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/pitabwire/frame/datastore"
)

func Migrate(ctx context.Context, dbManager datastore.Manager, migrationPath string) error {
	dbPool := dbManager.GetPool(ctx, datastore.DefaultMigrationPoolName)

	return dbManager.Migrate(ctx, dbPool, migrationPath,
		&models.CatalogVersion{}, &models.Plan{}, &models.Component{}, &models.Tier{},
		&models.Subscription{},
		&models.UsageEvent{},
		&models.MeteredUsage{},
		&models.RatedLine{},
		&models.Discount{}, &models.DiscountedLine{},
		&models.CreditGrant{}, &models.CreditEntry{},
		&models.Invoice{}, &models.InvoiceLine{},
		&models.BillingRun{})
}
