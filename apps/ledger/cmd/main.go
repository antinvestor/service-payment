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

	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v1/ledgerv1connect"
	ledgerpbv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/antinvestor/common/v2/permissions"
	aconfig "github.com/antinvestor/service-payments/apps/ledger/config"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/service/business"
	"github.com/antinvestor/service-payments/apps/ledger/service/handlers"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/pitabwire/frame/v2"
	"github.com/pitabwire/frame/v2/config"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/frame/v2/security/authorizer"
	connectInterceptors "github.com/pitabwire/frame/v2/security/interceptors/connect"
	"github.com/pitabwire/frame/v2/setup"
	"github.com/pitabwire/util"
)

func main() {
	ctx := context.Background()

	// Create frame service
	cfg, err := config.LoadWithOIDC[aconfig.LedgerConfig](ctx)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "service_ledger"
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
	workMan := service.WorkManager()

	service.Setup().RegisterFunc(setup.NameMigrate, func(ctx context.Context) error {
		return repository.Migrate(ctx, dbManager, cfg.GetDatabaseMigrationPath())
	})

	// Setup Job: migrate + publish service_ledger permission manifest only.
	sd := ledgerpbv1.File_v1_ledger_proto.Services().ByName("LedgerService")
	if frame.ShouldRunSetup(&cfg) {
		service.Init(ctx, frame.WithPermissionRegistration(sd))
		if setupErr := service.RunSetupForProcess(ctx, &cfg); setupErr != nil {
			log.WithError(setupErr).Fatal("setup plan failed")
		}
		log.Info("setup plan complete — exiting")
		return
	}

	// Create repositories with proper dependency injection
	ledgerRepo := repository.NewLedgerRepository(ctx, dbPool, workMan)
	accountRepo := repository.NewAccountRepository(ctx, dbPool, workMan)
	transactionRepo := repository.NewTransactionRepository(ctx, dbPool, workMan, accountRepo)
	reportRepo := repository.NewReportRepository(dbPool)
	bookRepo := repository.NewBookRepository(ctx, dbPool, workMan)

	ledgerBusiness := business.NewLedgerBusiness(workMan, ledgerRepo, accountRepo)
	accountBusiness := business.NewAccountBusiness(workMan, ledgerRepo, accountRepo)
	transactionBusiness := business.NewTransactionBusiness(workMan, accountRepo, transactionRepo)
	reportBusiness := business.NewReportBusiness(reportRepo)
	bookBusiness := business.NewBookBusiness(bookRepo)

	// Create handler with injected business layer
	ledgerServer := handlers.NewLedgerServer(
		ledgerBusiness, accountBusiness, transactionBusiness, reportBusiness, bookBusiness)

	// Setup Connect server with injected dependencies
	connectHandler := setupConnectServer(ctx, service.SecurityManager(), ledgerServer)

	// Runtime: HTTP only — permissions publish on the setup Job path above.
	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(connectHandler),
	}
	service.Init(ctx, serviceOptions...)

	// Startup service
	err = service.Run(ctx, "")
	if err != nil {
		log.WithError(err).Error("could not run Server")
	}
}

// setupConnectServer initializes and configures the connect server.
func setupConnectServer(
	ctx context.Context,
	securityMan security.Manager,
	implementation ledgerv1connect.LedgerServiceHandler,
) http.Handler {
	auth := securityMan.GetAuthorizer(ctx)

	// Layer 1: TenancyAccessChecker verifies caller can access the partition.
	// Self-heal missing #service tuples for internal service bots.
	tenancyAccessChecker := authorizer.NewTenancyAccessChecker(auth, authz.NamespaceTenancyAccess,
		authorizer.WithOnTenancyAccessDenied(authz.HealServiceTenancyAccess),
	)
	tenancyAccessInterceptor := connectInterceptors.NewTenancyAccessInterceptor(tenancyAccessChecker)

	// Layer 2: FunctionAccessInterceptor enforces per-RPC permissions automatically.
	sd := ledgerpbv1.File_v1_ledger_proto.Services().ByName("LedgerService")
	procMap := permissions.BuildProcedureMap(sd)
	functionChecker := authorizer.NewFunctionChecker(auth, permissions.ForService(sd).Namespace)
	functionAccessInterceptor := connectInterceptors.NewFunctionAccessInterceptor(functionChecker, procMap)

	// Layer 3: TenancyTxInterceptor is now automatically included in DefaultList.
	// It opens a request-scoped transaction after auth has populated the claims,
	// publishes app.tenant_id + app.partition_id from the claims via set_config,
	// and binds the transaction to the request context.

	otelInterceptor, otelErr := otelconnect.NewInterceptor()
	if otelErr != nil {
		util.Log(ctx).WithError(otelErr).Fatal("main -- Could not configure open telemetry interceptor")
	}

	defaultInterceptorList, err := connectInterceptors.DefaultList(
		ctx, securityMan.GetAuthenticator(ctx),
		tenancyAccessInterceptor, functionAccessInterceptor)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("main -- Could not create default interceptors")
	}

	allInterceptors := append([]connect.Interceptor{otelInterceptor}, defaultInterceptorList...)

	_, serverHandler := ledgerv1connect.NewLedgerServiceHandler(
		implementation, connect.WithInterceptors(allInterceptors...))

	return serverHandler
}
