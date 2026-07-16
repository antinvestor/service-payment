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
	"time"

	//nolint:gosec // G108: Profiling endpoint deliberately exposed for monitoring and debugging purposes
	_ "net/http/pprof"

	"buf.build/gen/go/antinvestor/billing/connectrpc/go/v1/billingv1connect"
	billingpb "buf.build/gen/go/antinvestor/billing/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	apis "github.com/antinvestor/common/v2"
	"github.com/antinvestor/common/v2/connection"
	"github.com/antinvestor/common/v2/servicecatalog"
	"github.com/antinvestor/common/v2/timescale"
	aconfig "github.com/antinvestor/service-payments/apps/billing/config"
	collectionv1 "github.com/antinvestor/service-payments/apps/billing/gen/collection/v1"
	"github.com/antinvestor/service-payments/apps/billing/gen/collection/v1/collectionv1connect"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/handlers"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1/checkoutv1connect"

	// Ledger integration dependencies.
	ledgerBusiness "github.com/antinvestor/service-payments/apps/ledger/service/business"
	ledgerRepo "github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/pitabwire/frame/v2"
	"github.com/pitabwire/frame/v2/config"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/security"
	securityconnect "github.com/pitabwire/frame/v2/security/interceptors/connect"
	"github.com/pitabwire/frame/v2/workerpool"
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

	// Subscription lifecycle fan-out to external entity integrators
	// (product apps / entitlements) — payment Send/Receive style.
	integrationRouteRepo := repository.NewIntegrationRouteRepository(ctx, dbPool, workMan)
	qMan := service.QueueManager()
	lifecycleNotifier := business.NewSubscriptionLifecycleNotifier(
		qMan,
		integrationRouteRepo,
		business.LifecycleNotifierConfig{
			DefaultTopicName: cfg.SubscriptionLifecycleTopicName,
			DefaultTopicURI:  cfg.SubscriptionLifecycleTopicURI,
		},
	)

	// Create billing business layers
	catalogBus := business.NewCatalogBusiness(workMan, catalogVersionRepo, planRepo, componentRepo, tierRepo)
	subscriptionBus := business.NewSubscriptionBusiness(workMan, subscriptionRepo, lifecycleNotifier)
	usageIngestionBus := business.NewUsageIngestionBusiness(workMan, usageEventRepo)
	meteringEng := business.NewMeteringEngine(workMan, usageEventRepo, meteredUsageRepo)
	pricingEng := business.NewPricingEngine()
	discountEng := business.NewDiscountEngine(workMan, discountRepo, discountedLineRepo)
	creditEng := business.NewCreditEngine(workMan, dbPool, creditGrantRepo, creditEntryRepo)
	invoiceEng := business.NewInvoiceEngine(workMan, dbPool, invoiceRepo, invoiceLineRepo)
	ledgerInteg := business.NewLedgerIntegration(ledgerTxnBusiness)

	checkoutCli, checkoutErr := setupCheckoutClient(ctx, cfg)
	if checkoutErr != nil {
		log.WithError(checkoutErr).Error("could not setup checkout client")
		return
	}
	checkoutInteg := business.NewCheckoutIntegration(checkoutCli, invoiceRepo, invoiceEng, cfg.CheckoutInvoiceReturnURL)

	billingWorkflow := business.NewBillingWorkflow(
		workMan, billingRunRepo, ratedLineRepo, subscriptionBus, catalogBus, componentRepo,
		meteringEng, pricingEng, discountEng, creditEng, invoiceEng, ledgerInteg, checkoutInteg)

	ledgerAccounts := business.CollectionLedgerAccounts{
		CashAccountID: cfg.LedgerCashAccountID,
		ARAccountID:   cfg.LedgerARAccountID,
	}
	collectionBiz := business.NewCollectionBusiness(
		checkoutInteg, invoiceEng, invoiceRepo, subscriptionBus,
		planRepo, componentRepo, billingRunRepo, pricingEng,
		ledgerInteg, ledgerAccounts,
	)
	settlementSweeper := business.NewSettlementSweeper(
		invoiceRepo, collectionBiz, cfg.SettlementSweepBatchSize,
	)

	// Create handlers with injected business layer
	billingServer := handlers.NewBillingServer(
		catalogBus, subscriptionBus, usageIngestionBus, meteringEng, pricingEng,
		discountEng, creditEng, invoiceEng, billingWorkflow, ledgerInteg,
		catalogVersionRepo, usageEventRepo, invoiceRepo, discountRepo, subscriptionRepo)
	collectionServer := handlers.NewCollectionServer(collectionBiz)

	// Handle database migration if requested
	if handleDatabaseMigration(ctx, dbManager, cfg, log) {
		return
	}

	// Setup Connect server with both BillingService and CollectionService
	connectHandler := setupConnectServer(ctx, service.SecurityManager(), billingServer, collectionServer)

	// Setup HTTP handlers and register permissions with Keto
	billingSD := billingpb.File_v1_billing_proto.Services().ByName("BillingService")
	collectionSD := collectionv1.File_collection_v1_collection_proto.Services().ByName("CollectionService")
	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(connectHandler),
		frame.WithPermissionRegistration(billingSD),
		frame.WithPermissionRegistration(collectionSD),
	}
	// Global default lifecycle queue (product services subscribe here).
	// Partition IntegrationRoute rows register additional publishers on demand.
	if cfg.SubscriptionLifecycleTopicName != "" && cfg.SubscriptionLifecycleTopicURI != "" {
		serviceOptions = append(serviceOptions, frame.WithRegisterPublisher(
			cfg.SubscriptionLifecycleTopicName,
			cfg.SubscriptionLifecycleTopicURI,
		))
	}
	service.Init(ctx, serviceOptions...)

	// Settlement sweeper recovers abandoned browser return/confirm paths.
	go runSettlementSweeper(ctx, &cfg, settlementSweeper, workMan)

	// Startup service
	err = service.Run(ctx, "")
	if err != nil {
		log.WithError(err).Error("could not run Server")
	}
}

// setupCheckoutClient creates a gRPC client for the checkout service.
func setupCheckoutClient(
	ctx context.Context,
	cfg aconfig.BillingConfig,
) (checkoutv1connect.CheckoutServiceClient, error) {
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.CheckoutServiceURI,
		WorkloadAPITargetPath: cfg.CheckoutServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServiceCheckout,
	}, checkoutv1connect.NewCheckoutServiceClient)
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

// setupConnectServer mounts BillingService and CollectionService on one mux.
func setupConnectServer(
	ctx context.Context,
	securityMan security.Manager,
	billingImpl billingv1connect.BillingServiceHandler,
	collectionImpl collectionv1connect.CollectionServiceHandler,
) http.Handler {
	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not configure open telemetry")
	}

	validateInterceptor := securityconnect.NewValidationInterceptor()

	authenticator := securityMan.GetAuthenticator(ctx)
	authInterceptor := securityconnect.NewAuthInterceptor(authenticator)
	opts := connect.WithInterceptors(authInterceptor, otelInterceptor, validateInterceptor)

	mux := http.NewServeMux()

	billingPath, billingHandler := billingv1connect.NewBillingServiceHandler(billingImpl, opts)
	mux.Handle(billingPath, billingHandler)

	collectionPath, collectionHandler := collectionv1connect.NewCollectionServiceHandler(collectionImpl, opts)
	mux.Handle(collectionPath, collectionHandler)

	return mux
}

// runSettlementSweeper ticks periodically and submits sweep work to the frame workerpool.
// The raw goroutine is only the scheduler; Frame has no cron primitive.
func runSettlementSweeper(
	ctx context.Context,
	cfg *aconfig.BillingConfig,
	sweeper business.SettlementSweeper,
	workMan workerpool.Manager,
) {
	intervalSec := cfg.SettlementSweepIntervalSeconds
	if intervalSec <= 0 {
		util.Log(ctx).Info("settlement sweeper disabled (interval <= 0)")
		return
	}
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	util.Log(ctx).
		WithField("interval_seconds", intervalSec).
		Info("settlement sweeper started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			job := workerpool.NewJob(func(jobCtx context.Context, _ workerpool.JobResultPipe[any]) error {
				result, sweepErr := sweeper.Sweep(jobCtx)
				if sweepErr != nil {
					util.Log(jobCtx).WithError(sweepErr).Warn("settlement sweep failed")
					return nil // do not retry the whole tick via workerpool
				}
				if result != nil && (result.Settled > 0 || result.Errors > 0) {
					util.Log(jobCtx).
						WithField("candidates", result.Candidates).
						WithField("settled", result.Settled).
						WithField("skipped", result.Skipped).
						WithField("errors", result.Errors).
						Info("settlement sweep completed")
				}
				return nil
			})
			if submitErr := workerpool.SubmitJob(ctx, workMan, job); submitErr != nil {
				util.Log(ctx).WithError(submitErr).Warn("could not submit settlement sweep job")
			}
		}
	}
}
