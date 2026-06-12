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
	"time"

	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"connectrpc.com/connect"
	apis "github.com/antinvestor/common"
	"github.com/antinvestor/common/connection"
	"github.com/antinvestor/common/permissions"
	aconfig "github.com/antinvestor/service-payments/apps/checkout/config"
	checkoutv1 "github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1"
	"github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1/checkoutv1connect"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/handlers"
	"github.com/antinvestor/service-payments/apps/checkout/service/repository"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/authorizer"
	connectInterceptors "github.com/pitabwire/frame/security/interceptors/connect"
	"github.com/pitabwire/util"
)

// namespaceTenancyAccess is the Keto namespace used by TenancyAccessChecker.
// All services in this repo share the same constant value.
const namespaceTenancyAccess = "tenancy_access"

func main() {
	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.CheckoutConfig](ctx)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "service_payment_checkout"
	}

	ctx, svc := frame.NewServiceWithContext(ctx, frame.WithConfig(&cfg), frame.WithDatastore())
	defer svc.Stop(ctx)

	log := svc.Log(ctx)

	dbManager := svc.DatastoreManager()

	// Migrate branch — mirrors ledger exactly.
	if handleDatabaseMigration(ctx, dbManager, cfg, log) {
		return
	}

	if cfg.SigningSecret == "" {
		log.Error("CHECKOUT_SIGNING_SECRET is required")
		return
	}

	registry, err := business.ParseMethodRegistry(cfg.MethodsJSON)
	if err != nil {
		log.WithError(err).Error("invalid CHECKOUT_METHODS")
		return
	}

	dbPool := dbManager.GetPool(ctx, datastore.DefaultPoolName)
	workMan := svc.WorkManager()

	sessionRepo := repository.NewSessionRepository(ctx, dbPool, workMan)
	linkRepo := repository.NewLinkRepository(ctx, dbPool, workMan)

	paymentCli, err := setupPaymentClient(ctx, cfg)
	if err != nil {
		log.WithError(err).Error("could not setup payment client")
		return
	}

	profileCli, err := setupProfileClient(ctx, cfg)
	if err != nil {
		log.WithError(err).Error("could not setup profile client")
		return
	}

	checkoutBiz := business.NewCheckoutBusiness(&cfg, registry, sessionRepo, linkRepo, paymentCli, profileCli)

	renderer, err := handlers.NewRenderer([]byte(cfg.SigningSecret))
	if err != nil {
		log.WithError(err).Error("could not create renderer")
		return
	}

	webServer := handlers.NewWebServer(checkoutBiz, renderer, registry, &cfg)
	rpcServer := handlers.NewCheckoutServer(checkoutBiz, &cfg)
	rpcHandler := setupConnectServer(ctx, svc.SecurityManager(), rpcServer)

	mux := webServer.NewRouter()
	mux.Handle("/", rpcHandler)

	sd := checkoutv1.File_checkout_v1_checkout_proto.Services().ByName("CheckoutService")
	svc.Init(ctx,
		frame.WithHTTPHandler(mux),
		frame.WithPermissionRegistration(sd),
	)

	// runSweeper is a deliberate, documented exception to the no-raw-goroutines rule.
	// Frame exposes no periodic-job primitive (WorkerPool handles one-shot jobs,
	// not recurring ticks). The sweeper is best-effort reconciliation — a missed
	// tick on shutdown is harmless, making a goroutine the correct tool here.
	go runSweeper(ctx, &cfg, checkoutBiz)

	log.Info("Initiating checkout server operations")
	if err = svc.Run(ctx, ""); err != nil {
		log.WithError(err).Error("could not run Server")
	}
}

// handleDatabaseMigration performs database migration when configured and returns true to exit.
func handleDatabaseMigration(
	ctx context.Context,
	dbManager datastore.Manager,
	cfg aconfig.CheckoutConfig,
	log *util.LogEntry,
) bool {
	if cfg.DoDatabaseMigrate() {
		if err := repository.Migrate(ctx, dbManager, cfg.GetDatabaseMigrationPath()); err != nil {
			log.WithError(err).Fatal("main -- Could not migrate successfully")
		}
		return true
	}
	return false
}

// setupConnectServer mirrors ledger's interceptor stack exactly:
// TenancyAccessInterceptor → FunctionAccessInterceptor → DefaultList.
func setupConnectServer(
	ctx context.Context,
	securityMan security.Manager,
	implementation checkoutv1connect.CheckoutServiceHandler,
) http.Handler {
	auth := securityMan.GetAuthorizer(ctx)

	// Layer 1: TenancyAccessChecker verifies caller can access the partition.
	tenancyAccessChecker := authorizer.NewTenancyAccessChecker(auth, namespaceTenancyAccess)
	tenancyAccessInterceptor := connectInterceptors.NewTenancyAccessInterceptor(tenancyAccessChecker)

	// Layer 2: FunctionAccessInterceptor enforces per-RPC permissions automatically.
	sd := checkoutv1.File_checkout_v1_checkout_proto.Services().ByName("CheckoutService")
	procMap := permissions.BuildProcedureMap(sd)
	functionChecker := authorizer.NewFunctionChecker(auth, permissions.ForService(sd).Namespace)
	functionAccessInterceptor := connectInterceptors.NewFunctionAccessInterceptor(functionChecker, procMap)

	// Layer 3: TenancyTxInterceptor is automatically included in DefaultList.
	defaultInterceptorList, err := connectInterceptors.DefaultList(
		ctx, securityMan.GetAuthenticator(ctx),
		tenancyAccessInterceptor, functionAccessInterceptor)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("main -- Could not create default interceptors")
	}

	_, serverHandler := checkoutv1connect.NewCheckoutServiceHandler(
		implementation, connect.WithInterceptors(defaultInterceptorList...))

	return serverHandler
}

func setupPaymentClient(
	ctx context.Context,
	cfg aconfig.CheckoutConfig,
) (paymentv1connect.PaymentServiceClient, error) {
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.PaymentServiceURI,
		WorkloadAPITargetPath: cfg.PaymentServiceWorkloadAPITargetPath,
		Audiences:             []string{"service_payment"},
	}, paymentv1connect.NewPaymentServiceClient)
}

func setupProfileClient(
	ctx context.Context,
	cfg aconfig.CheckoutConfig,
) (profilev1connect.ProfileServiceClient, error) {
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.ProfileServiceURI,
		WorkloadAPITargetPath: cfg.ProfileServiceWorkloadAPITargetPath,
		Audiences:             []string{"service_profile"},
	}, profilev1connect.NewProfileServiceClient)
}

// runSweeper ticks on SweepIntervalSeconds and calls SweepProcessing.
// It exits when ctx is cancelled (service shutdown).
//
// This is a deliberate raw goroutine: Frame exposes no cron/periodic-job
// primitive; WorkerPool handles one-shot jobs, not recurring ticks.
// A missed tick on shutdown is harmless, making a goroutine the right tool.
func runSweeper(ctx context.Context, cfg *aconfig.CheckoutConfig, biz *business.CheckoutBusiness) {
	interval := time.Duration(cfg.SweepIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := biz.SweepProcessing(ctx); err != nil {
				util.Log(ctx).WithError(err).Warn("checkout sweep failed")
			}
		}
	}
}
