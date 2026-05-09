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

package main

import (
	"context"
	"net/http"

	//nolint:gosec // G108: Profiling endpoint deliberately exposed for monitoring and debugging purposes
	_ "net/http/pprof"

	"buf.build/gen/go/antinvestor/billing/connectrpc/go/v1/billingv1connect"
	billingpb "buf.build/gen/go/antinvestor/billing/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/antinvestor/common/timescale"
	aconfig "github.com/antinvestor/service-payments/apps/billing/config"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/handlers"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"

	// Ledger integration dependencies.
	ledgerBusiness "github.com/antinvestor/service-payments/apps/ledger/service/business"
	ledgerRepo "github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/security"
	securityconnect "github.com/pitabwire/frame/security/interceptors/connect"
	"github.com/pitabwire/util"
)

//nolint:funlen // DI wiring function that is inherently long
func main() {
	ctx := context.Background()

	// Create frame service
	cfg, err := config.LoadWithOIDC[aconfig.BillingConfig](ctx)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "service_billing"
	}

	_, service := frame.NewServiceWithContext(
		ctx,
		frame.WithConfig(&cfg),
		frame.WithDatastore(),
		frame.WithTranslation("en"),
	)
	defer service.Stop(ctx)

	log := service.Log(ctx)

	// Get the default database pool and work manager
	dbManager := service.DatastoreManager()
	dbPool := dbManager.GetPool(ctx, datastore.DefaultPoolName)

	// Register hypertables (no-op WARN if timescaledb extension is absent).
	ensureHypertables(ctx, dbPool)
	workMan := service.WorkManager()

	// Create billing repositories (15 repos, bottom-up)
	catalogVersionRepo := repository.NewCatalogVersionRepository(ctx, dbPool, workMan)
	planRepo := repository.NewPlanRepository(ctx, dbPool, workMan)
	componentRepo := repository.NewComponentRepository(ctx, dbPool, workMan)
	tierRepo := repository.NewTierRepository(ctx, dbPool, workMan)
	subscriptionRepo := repository.NewSubscriptionRepository(ctx, dbPool, workMan)
	usageEventRepo := repository.NewUsageEventRepository(ctx, dbPool, workMan)
	meteredUsageRepo := repository.NewMeteredUsageRepository(ctx, dbPool, workMan)
	ratedLineRepo := repository.NewRatedLineRepository(ctx, dbPool, workMan)
	discountRepo := repository.NewDiscountRepository(ctx, dbPool, workMan)
	discountedLineRepo := repository.NewDiscountedLineRepository(ctx, dbPool, workMan)
	creditGrantRepo := repository.NewCreditGrantRepository(ctx, dbPool, workMan)
	creditEntryRepo := repository.NewCreditEntryRepository(ctx, dbPool, workMan)
	invoiceRepo := repository.NewInvoiceRepository(ctx, dbPool, workMan)
	invoiceLineRepo := repository.NewInvoiceLineRepository(ctx, dbPool, workMan)
	billingRunRepo := repository.NewBillingRunRepository(ctx, dbPool, workMan)

	// Create ledger repositories for integration
	lLedgerRepo := ledgerRepo.NewLedgerRepository(ctx, dbPool, workMan)
	lAccountRepo := ledgerRepo.NewAccountRepository(ctx, dbPool, workMan)
	lTransactionRepo := ledgerRepo.NewTransactionRepository(ctx, dbPool, workMan, lAccountRepo)

	// Create ledger business layers for integration
	_ = ledgerBusiness.NewLedgerBusiness(workMan, lLedgerRepo, lAccountRepo)
	_ = ledgerBusiness.NewAccountBusiness(workMan, lLedgerRepo, lAccountRepo)
	ledgerTxnBusiness := ledgerBusiness.NewTransactionBusiness(workMan, lAccountRepo, lTransactionRepo)

	// Create billing business layers
	catalogBus := business.NewCatalogBusiness(workMan, catalogVersionRepo, planRepo, componentRepo, tierRepo)
	subscriptionBus := business.NewSubscriptionBusiness(workMan, subscriptionRepo)
	usageIngestionBus := business.NewUsageIngestionBusiness(workMan, usageEventRepo)
	meteringEng := business.NewMeteringEngine(workMan, usageEventRepo, meteredUsageRepo)
	pricingEng := business.NewPricingEngine()
	discountEng := business.NewDiscountEngine(workMan, discountRepo, discountedLineRepo)
	creditEng := business.NewCreditEngine(workMan, dbPool, creditGrantRepo, creditEntryRepo)
	invoiceEng := business.NewInvoiceEngine(workMan, dbPool, invoiceRepo, invoiceLineRepo)
	ledgerInteg := business.NewLedgerIntegration(ledgerTxnBusiness)
	billingWorkflow := business.NewBillingWorkflow(
		workMan, billingRunRepo, ratedLineRepo, subscriptionBus, catalogBus, componentRepo,
		meteringEng, pricingEng, discountEng, creditEng, invoiceEng, ledgerInteg)

	// Create handler with injected business layer
	billingServer := handlers.NewBillingServer(
		catalogBus, subscriptionBus, usageIngestionBus, meteringEng, pricingEng,
		discountEng, creditEng, invoiceEng, billingWorkflow, ledgerInteg,
		catalogVersionRepo, usageEventRepo, invoiceRepo, discountRepo, subscriptionRepo)

	// Handle database migration if requested
	if handleDatabaseMigration(ctx, dbManager, cfg, log) {
		return
	}

	// Setup Connect server with injected dependencies
	connectHandler := setupConnectServer(ctx, service.SecurityManager(), billingServer)

	// Setup HTTP handlers and register permissions with Keto
	sd := billingpb.File_v1_billing_proto.Services().ByName("BillingService")
	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(connectHandler),
		frame.WithPermissionRegistration(sd),
	}
	service.Init(ctx, serviceOptions...)

	// Startup service
	err = service.Run(ctx, "")
	if err != nil {
		log.WithError(err).Error("could not run Server")
	}
}

// ensureHypertables registers TimescaleDB hypertables idempotently.
// Errors are logged as warnings so the service continues when TimescaleDB
// is not yet available.
func ensureHypertables(ctx context.Context, dbPool pool.Pool) {
	if tsErr := timescale.Ensure(ctx, dbPool.DB(ctx, false), models.Hypertables()); tsErr != nil {
		util.Log(ctx).WithError(tsErr).Warn("timescale hypertable setup skipped — will retry after cluster migration")
	}
}

// handleDatabaseMigration performs database migration if configured to do so.
func handleDatabaseMigration(
	ctx context.Context,
	dbManager datastore.Manager,
	cfg aconfig.BillingConfig,
	log *util.LogEntry,
) bool {
	if cfg.DoDatabaseMigrate() {
		err := repository.Migrate(ctx, dbManager, cfg.GetDatabaseMigrationPath())
		if err != nil {
			log.WithError(err).Fatal("main -- Could not migrate successfully")
		}
		return true
	}
	return false
}

// setupConnectServer initializes and configures the connect server.
func setupConnectServer(
	ctx context.Context,
	securityMan security.Manager,
	implementation billingv1connect.BillingServiceHandler,
) http.Handler {
	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not configure open telemetry")
	}

	validateInterceptor := securityconnect.NewValidationInterceptor()

	authenticator := securityMan.GetAuthenticator(ctx)
	authInterceptor := securityconnect.NewAuthInterceptor(authenticator)

	_, serverHandler := billingv1connect.NewBillingServiceHandler(
		implementation, connect.WithInterceptors(authInterceptor, otelInterceptor, validateInterceptor))

	return serverHandler
}
