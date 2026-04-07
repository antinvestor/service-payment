package main

import (
	"context"
	"net/http"

	//nolint:gosec // G108: Profiling endpoint deliberately exposed for monitoring and debugging purposes
	_ "net/http/pprof"

	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v1/ledgerv1connect"
	ledgerpbv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/common/permissions"
	aconfig "github.com/antinvestor/service-payments/apps/ledger/config"
	"github.com/antinvestor/service-payments/apps/ledger/service/authz"
	"github.com/antinvestor/service-payments/apps/ledger/service/business"
	"github.com/antinvestor/service-payments/apps/ledger/service/handlers"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/authorizer"
	connectInterceptors "github.com/pitabwire/frame/security/interceptors/connect"
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

	// Create repositories with proper dependency injection
	ledgerRepo := repository.NewLedgerRepository(ctx, dbPool, workMan)
	accountRepo := repository.NewAccountRepository(ctx, dbPool, workMan)
	transactionRepo := repository.NewTransactionRepository(ctx, dbPool, workMan, accountRepo)

	ledgerBusiness := business.NewLedgerBusiness(workMan, ledgerRepo, accountRepo)
	accountBusiness := business.NewAccountBusiness(workMan, ledgerRepo, accountRepo)
	transactionBusiness := business.NewTransactionBusiness(workMan, accountRepo, transactionRepo)

	// Create handler with injected business layer
	ledgerServer := handlers.NewLedgerServer(ledgerBusiness, accountBusiness, transactionBusiness)

	// Handle database migration if requested
	if handleDatabaseMigration(ctx, dbManager, cfg, log) {
		return
	}

	// Setup Connect server with injected dependencies
	connectHandler := setupConnectServer(ctx, service.SecurityManager(), ledgerServer)

	// Setup HTTP handlers
	serviceOptions := []frame.Option{frame.WithHTTPHandler(connectHandler)}
	service.Init(ctx, serviceOptions...)

	// Startup service
	err = service.Run(ctx, "")
	if err != nil {
		log.WithError(err).Error("could not run Server")
	}
}

// handleDatabaseMigration performs database migration if configured to do so.
func handleDatabaseMigration(
	ctx context.Context,
	dbManager datastore.Manager,
	cfg aconfig.LedgerConfig,
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
	implementation ledgerv1connect.LedgerServiceHandler,
) http.Handler {
	auth := securityMan.GetAuthorizer(ctx)

	// Layer 1: TenancyAccessChecker verifies caller can access the partition.
	tenancyAccessChecker := authorizer.NewTenancyAccessChecker(auth, authz.NamespaceTenancyAccess)
	tenancyAccessInterceptor := connectInterceptors.NewTenancyAccessInterceptor(tenancyAccessChecker)

	// Layer 2: FunctionAccessInterceptor enforces per-RPC permissions automatically.
	sd := ledgerpbv1.File_v1_ledger_proto.Services().ByName("LedgerService")
	procMap := permissions.BuildProcedureMap(sd)
	functionChecker := authorizer.NewFunctionChecker(auth, permissions.ForService(sd).Namespace)
	functionAccessInterceptor := connectInterceptors.NewFunctionAccessInterceptor(functionChecker, procMap)

	defaultInterceptorList, err := connectInterceptors.DefaultList(
		ctx, securityMan.GetAuthenticator(ctx), tenancyAccessInterceptor, functionAccessInterceptor)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("main -- Could not create default interceptors")
	}

	_, serverHandler := ledgerv1connect.NewLedgerServiceHandler(
		implementation, connect.WithInterceptors(defaultInterceptorList...))

	return serverHandler
}
