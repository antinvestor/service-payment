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
	"connectrpc.com/otelconnect"
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
	"github.com/pitabwire/frame/workerpool"
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

	checkoutBiz := business.NewCheckoutBusiness(&cfg, registry, sessionRepo, linkRepo, paymentCli, profileCli, workMan)

	renderer, err := handlers.NewRenderer([]byte(cfg.SigningSecret))
	if err != nil {
		log.WithError(err).Error("could not create renderer")
		return
	}

	webServer := handlers.NewWebServer(checkoutBiz, renderer, registry, &cfg)
	rpcServer := handlers.NewCheckoutServer(checkoutBiz, &cfg)
	rpcHandler := setupConnectServer(ctx, svc.SecurityManager(), rpcServer)

	mux := webServer.NewRouter()
	mux.Handle("/"+checkoutv1connect.CheckoutServiceName+"/", rpcHandler)

	sd := checkoutv1.File_checkout_v1_checkout_proto.Services().ByName("CheckoutService")
	svc.Init(ctx,
		frame.WithHTTPHandler(mux),
		frame.WithPermissionRegistration(sd),
	)

	// runSweeper uses a raw goroutine only as a periodic scheduler — Frame exposes
	// no cron/ticker primitive. Each tick submits the sweep as a frame workerpool
	// job so the work runs on the managed pool, not an ad-hoc goroutine.
	go runSweeper(ctx, &cfg, checkoutBiz, workMan)

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

// setupConnectServer wires the interceptor stack:
// otelconnect → TenancyAccessInterceptor → FunctionAccessInterceptor → DefaultList.
func setupConnectServer(
	ctx context.Context,
	securityMan security.Manager,
	implementation checkoutv1connect.CheckoutServiceHandler,
) http.Handler {
	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not configure open telemetry")
	}

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

	// Prepend otelconnect as the outermost interceptor so every RPC is traced.
	allInterceptors := append([]connect.Interceptor{otelInterceptor}, defaultInterceptorList...)

	_, serverHandler := checkoutv1connect.NewCheckoutServiceHandler(
		implementation, connect.WithInterceptors(allInterceptors...))

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

// runSweeper ticks on SweepIntervalSeconds and submits a sweep job to the frame
// workerpool on each tick. It exits when ctx is cancelled (service shutdown).
//
// The raw goroutine here is intentional: Frame exposes no cron/periodic-job
// primitive, so the ticker loop IS the scheduler. The actual sweep work executes
// on the managed worker pool — not in this goroutine — conforming to the repo
// no-raw-goroutines-for-work rule. A missed tick on shutdown is harmless.
func runSweeper(
	ctx context.Context,
	cfg *aconfig.CheckoutConfig,
	biz *business.CheckoutBusiness,
	workMan workerpool.Manager,
) {
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
			job := workerpool.NewJob(func(jobCtx context.Context, _ workerpool.JobResultPipe[any]) error {
				if err := biz.SweepProcessing(jobCtx); err != nil {
					util.Log(jobCtx).WithError(err).Warn("checkout sweep failed")
				}
				return nil
			})
			if err := workerpool.SubmitJob(ctx, workMan, job); err != nil {
				util.Log(ctx).WithError(err).Warn("could not submit sweep job")
			}
		}
	}
}
