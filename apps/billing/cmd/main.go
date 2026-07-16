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
	"encoding/json"
	"net/http"
	"strings"

	//nolint:gosec // G108: Profiling endpoint deliberately exposed for monitoring and debugging purposes
	_ "net/http/pprof"

	"buf.build/gen/go/antinvestor/billing/connectrpc/go/v1/billingv1connect"
	billingpb "buf.build/gen/go/antinvestor/billing/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
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
	"github.com/antinvestor/service-trustage/gen/go/workflow/v1/workflowv1connect"
	"github.com/pitabwire/frame/v2"
	"github.com/pitabwire/frame/v2/config"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/security"
	securityconnect "github.com/pitabwire/frame/v2/security/interceptors/connect"
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

	// Payment + profile clients for silent COF renewals (Flutterwave v4 token charges).
	paymentCli, paymentErr := setupPaymentClient(ctx, cfg)
	if paymentErr != nil {
		log.WithError(paymentErr).Warn("payment client unavailable — COF renewals disabled until configured")
	}
	profileCli, profileErr := setupProfileClient(ctx, cfg)
	if profileErr != nil {
		log.WithError(profileErr).Warn("profile client unavailable — instrument pin falls back to subscription data only")
	}
	instrumentSrc := business.NewInstrumentSource(profileCli)
	paymentCollector := business.NewPaymentCollector(paymentCli)
	renewalCfg := business.NewRenewalConfigFromEnv(
		cfg.RenewalLeadHours,
		cfg.RenewalMaxAttempts,
		cfg.RenewalRetryDelaysCSV,
		cfg.RenewalDefaultRoute,
	)

	// Per-entity Trustage one-shots (renew per subscription, settle per invoice).
	var renewalScheduler business.RenewalScheduler = business.NoopRenewalScheduler{}
	var settlementScheduler business.SettlementScheduler = business.NoopSettlementScheduler{}
	settlementCfg := business.NewSettlementConfigFromEnv(
		cfg.SettlementMaxAttempts,
		cfg.SettlementRetryDelaysMinutesCSV,
	)
	trustageCli, trustageErr := setupTrustageClient(ctx, cfg)
	if trustageErr != nil {
		log.WithError(trustageErr).Warn("trustage client unavailable — per-entity schedules disabled")
	} else if trustageCli != nil && strings.TrimSpace(cfg.BillingInternalBaseURL) != "" {
		renewalScheduler = business.NewTrustageRenewalScheduler(trustageCli, business.TrustageSchedulerConfig{
			BillingBaseURL:        cfg.BillingInternalBaseURL,
			AdminTokenPlaceholder: "${BILLING_INTERNAL_ADMIN_TOKEN}",
			Renewal:               renewalCfg,
			SubBiz:                subscriptionBus,
		})
		settlementScheduler = business.NewTrustageSettlementScheduler(trustageCli, business.SettlementSchedulerConfig{
			BillingBaseURL:        cfg.BillingInternalBaseURL,
			AdminTokenPlaceholder: "${BILLING_INTERNAL_ADMIN_TOKEN}",
			Settlement:            settlementCfg,
			InvoiceRepo:           invoiceRepo,
		})
		log.WithField("billing_base", cfg.BillingInternalBaseURL).
			Info("trustage: per-subscription renew + per-invoice settle schedulers enabled")
	} else {
		log.Info("trustage: URL or BILLING_INTERNAL_BASE_URL unset — using noop schedulers")
	}

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
		business.CollectionOptions{
			Instruments:         instrumentSrc,
			Scheduler:           renewalScheduler,
			SettlementScheduler: settlementScheduler,
		},
	)
	settlementProcessor := business.NewSettlementProcessor(
		invoiceRepo, collectionBiz, settlementCfg, settlementScheduler,
	)
	renewalSweeper := business.NewRenewalSweeper(
		subscriptionBus, planRepo, componentRepo,
		billingRunRepo, invoiceRepo, invoiceEng,
		instrumentSrc, paymentCollector, renewalCfg, renewalScheduler,
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
	// Trustage / ops internal HTTP (renew + settle) mounted alongside Connect.
	rootMux := http.NewServeMux()
	rootMux.Handle("/", connectHandler)
	mountInternalBillingRoutes(rootMux, &cfg, settlementProcessor, renewalSweeper)

	// Setup HTTP handlers and register permissions with Keto
	billingSD := billingpb.File_v1_billing_proto.Services().ByName("BillingService")
	collectionSD := collectionv1.File_collection_v1_collection_proto.Services().ByName("CollectionService")
	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(rootMux),
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

	// Renew + settle: per-entity Trustage one-shots only — no bulk tickers.

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

// setupPaymentClient dials service-payment for InitiatePrompt COF charges.
func setupPaymentClient(
	ctx context.Context,
	cfg aconfig.BillingConfig,
) (paymentv1connect.PaymentServiceClient, error) {
	if strings.TrimSpace(cfg.PaymentServiceURI) == "" {
		return nil, nil
	}
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.PaymentServiceURI,
		WorkloadAPITargetPath: cfg.PaymentServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServicePayment,
	}, paymentv1connect.NewPaymentServiceClient)
}

// setupProfileClient dials profile for checkout instrument clues.
func setupProfileClient(
	ctx context.Context,
	cfg aconfig.BillingConfig,
) (profilev1connect.ProfileServiceClient, error) {
	if strings.TrimSpace(cfg.ProfileServiceURI) == "" {
		return nil, nil
	}
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.ProfileServiceURI,
		WorkloadAPITargetPath: cfg.ProfileServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServiceProfile,
	}, profilev1connect.NewProfileServiceClient)
}

// setupTrustageClient dials Trustage for per-subscription renew workflows.
func setupTrustageClient(
	ctx context.Context,
	cfg aconfig.BillingConfig,
) (workflowv1connect.WorkflowServiceClient, error) {
	if strings.TrimSpace(cfg.TrustageURL) == "" {
		return nil, nil
	}
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.TrustageURL,
		WorkloadAPITargetPath: cfg.TrustageWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServiceTrustage,
	}, workflowv1connect.NewWorkflowServiceClient)
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

// mountInternalBillingRoutes exposes Trustage-triggerable per-entity endpoints.
// Renew:  POST /_internal/billing/subscriptions/{id}/renew
// Settle: POST /_internal/billing/invoices/{id}/settle
// Auth: X-Admin-Token or Authorization: Bearer <BILLING_INTERNAL_ADMIN_TOKEN>.
func mountInternalBillingRoutes(
	mux *http.ServeMux,
	cfg *aconfig.BillingConfig,
	settlement business.SettlementProcessor,
	renewal business.RenewalProcessor,
) {
	token := strings.TrimSpace(cfg.InternalAdminToken)
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				http.Error(w, `{"error":"internal_routes_disabled"}`, http.StatusServiceUnavailable)
				return
			}
			got := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
			if got == "" {
				if authz := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(authz), "bearer ") {
					got = strings.TrimSpace(authz[7:])
				}
			}
			if got == "" || got != token {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	writeJSON := func(w http.ResponseWriter, status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}

	// Per-subscription Trustage callback — one subscription only.
	mux.HandleFunc("POST /_internal/billing/subscriptions/{id}/renew", auth(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "subscription id required"})
			return
		}
		result, err := renewal.ProcessSubscription(r.Context(), id)
		if err != nil {
			if result != nil && result.Action != "" && result.Action != "error" {
				writeJSON(w, http.StatusOK, result)
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"error":  err.Error(),
				"result": result,
			})
			return
		}
		writeJSON(w, http.StatusOK, result)
	}))

	// Per-invoice Trustage callback — one invoice only (no bulk scan).
	mux.HandleFunc("POST /_internal/billing/invoices/{id}/settle", auth(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invoice id required"})
			return
		}
		result, err := settlement.ProcessInvoice(r.Context(), id)
		if err != nil {
			if result != nil && result.Action != "" && result.Action != "error" {
				writeJSON(w, http.StatusOK, result)
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"error":  err.Error(),
				"result": result,
			})
			return
		}
		writeJSON(w, http.StatusOK, result)
	}))
}
