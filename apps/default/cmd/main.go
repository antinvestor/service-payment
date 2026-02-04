package main

import (
	"context"
	"net/http"

	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/ledger/v1/ledgerv1connect"
	"buf.build/gen/go/antinvestor/partition/connectrpc/go/partition/v1/partitionv1connect"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"connectrpc.com/connect"
	apis "github.com/antinvestor/apis/go/common"
	"github.com/antinvestor/apis/go/ledger"
	"github.com/antinvestor/apis/go/partition"
	"github.com/antinvestor/apis/go/profile"
	aconfig "github.com/antinvestor/service-payments/apps/default/config"
	"github.com/antinvestor/service-payments/apps/default/service/business"
	"github.com/antinvestor/service-payments/apps/default/service/events"
	"github.com/antinvestor/service-payments/apps/default/service/handlers"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/security"
	connectInterceptors "github.com/pitabwire/frame/security/interceptors/connect"
	"github.com/pitabwire/frame/security/openid"
	"github.com/pitabwire/util"
)

func main() {
	ctx := context.Background()

	// Initialize configuration
	cfg, err := config.LoadWithOIDC[aconfig.PaymentConfig](ctx)
	if err != nil {
		util.Log(ctx).With("err", err).Error("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "service_payment"
	}

	// Create service
	ctx, svc := frame.NewServiceWithContext(
		ctx,
		frame.WithConfig(&cfg),
		frame.WithRegisterServerOauth2Client(),
		frame.WithDatastore(),
	)
	defer svc.Stop(ctx)
	log := svc.Log(ctx)

	sm := svc.SecurityManager()
	workMan := svc.WorkManager()

	evtsMan := svc.EventsManager()
	qMan := svc.QueueManager()

	dbManager := svc.DatastoreManager()
	dbPool := dbManager.GetPool(ctx, datastore.DefaultPoolName)

	// Handle database migration if requested
	if handleDatabaseMigration(ctx, dbManager, cfg) {
		return
	}

	// Setup clients
	profileCli := setupProfileClient(ctx, sm, cfg)
	ledgerCli := setupLedgerClient(ctx, sm, cfg)
	partitionCli := setupPartitionClient(ctx, sm, cfg)

	// Initialize repositories
	paymentRepo := repository.NewPaymentRepository(ctx, dbPool, workMan)
	accountRepo := repository.NewAccountRepository(ctx, dbPool, workMan)
	costRepo := repository.NewCostRepository(ctx, dbPool, workMan)
	statusRepo := repository.NewStatusRepository(ctx, dbPool, workMan)
	promptRepo := repository.NewPromptRepository(ctx, dbPool, workMan)
	paymentLinkRepo := repository.NewPaymentLinkRepository(ctx, dbPool, workMan)
	routeRepo := repository.NewRouteRepository(ctx, dbPool, workMan)

	// Initialize business layer
	paymentBusiness, err := business.NewPaymentBusiness(
		ctx, workMan, evtsMan, profileCli, partitionCli, ledgerCli,
		paymentRepo, statusRepo, costRepo, accountRepo, promptRepo, paymentLinkRepo,
	)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not initialize payment business")
	}

	// Setup Connect server
	connectHandler := setupConnectServer(ctx, sm, paymentBusiness, profileCli, ledgerCli, partitionCli)

	// Setup HTTP handlers
	serviceOptions := []frame.Option{frame.WithDatastore(), frame.WithHTTPHandler(connectHandler)}

	// Register queue publishers

	promptPublisher := frame.WithRegisterPublisher(cfg.InitiatePromptTopicName, cfg.InitiatePromptTopicURI)
	paymentLinkPublisher := frame.WithRegisterPublisher(cfg.PaymentLinkTopicName, cfg.PaymentLinkTopicURI)

	// Register event handlers with proper constructors
	serviceOptions = append(serviceOptions,
		promptPublisher, paymentLinkPublisher,
		frame.WithRegisterEvents(
			events.NewPaymentSave(paymentRepo, evtsMan),
			events.NewPaymentInQueue(qMan, evtsMan, paymentRepo, routeRepo, profileCli),
			events.NewPaymentOutQueue(qMan, evtsMan, paymentRepo, statusRepo),
			events.NewPaymentInRoute(evtsMan, paymentRepo, routeRepo),
			events.NewPaymentOutRoute(qMan, evtsMan, paymentRepo, routeRepo, statusRepo),
			events.NewPromptSave(promptRepo),
			events.NewPaymentLinkSave(paymentLinkRepo),
			events.NewAccountSave(accountRepo),
			events.NewCostSave(costRepo),
			events.NewStatusSave(statusRepo),
		))

	// Initialize the service with all options
	svc.Init(ctx, serviceOptions...)

	// Start the service
	err = svc.Run(ctx, "")
	if err != nil {
		log.WithError(err).Fatal("could not run Server")
	}
}

// handleDatabaseMigration performs database migration if configured to do so.
func handleDatabaseMigration(
	ctx context.Context,
	dbManager datastore.Manager,
	cfg aconfig.PaymentConfig,
) bool {
	if cfg.DoDatabaseMigrate() {
		err := repository.Migrate(ctx, dbManager, cfg.GetDatabaseMigrationPath())
		if err != nil {
			util.Log(ctx).WithError(err).Fatal("main -- Could not migrate successfully")
		}
		return true
	}
	return false
}

// setupProfileClient creates and configures the profile service client.
func setupProfileClient(
	ctx context.Context,
	clHolder security.InternalOauth2ClientHolder,
	cfg aconfig.PaymentConfig,
) profilev1connect.ProfileServiceClient {
	profileCli, err := profile.NewClient(ctx,
		apis.WithEndpoint(cfg.ProfileServiceURI),
		apis.WithTokenEndpoint(cfg.GetOauth2TokenEndpoint()),
		apis.WithTokenUsername(clHolder.JwtClientID()),
		apis.WithTokenPassword(clHolder.JwtClientSecret()),
		apis.WithScopes(openid.ConstSystemScopeInternal),
		apis.WithAudiences("service_profile"))
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not setup profile client")
	}
	return profileCli
}

// setupLedgerClient creates and configures the ledger service client.
func setupLedgerClient(
	ctx context.Context,
	clHolder security.InternalOauth2ClientHolder,
	cfg aconfig.PaymentConfig,
) ledgerv1connect.LedgerServiceClient {
	ledgerCli, err := ledger.NewClient(ctx,
		apis.WithEndpoint(cfg.LedgerServiceURI),
		apis.WithTokenEndpoint(cfg.GetOauth2TokenEndpoint()),
		apis.WithTokenUsername(clHolder.JwtClientID()),
		apis.WithTokenPassword(clHolder.JwtClientSecret()),
		apis.WithScopes(openid.ConstSystemScopeInternal),
		apis.WithAudiences("service_ledger"))
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not setup ledger client")
	}
	return ledgerCli
}

// setupPartitionClient creates and configures the partition service client.
func setupPartitionClient(
	ctx context.Context,
	clHolder security.InternalOauth2ClientHolder,
	cfg aconfig.PaymentConfig,
) partitionv1connect.PartitionServiceClient {
	partitionCli, err := partition.NewClient(ctx,
		apis.WithEndpoint(cfg.PartitionServiceURI),
		apis.WithTokenEndpoint(cfg.GetOauth2TokenEndpoint()),
		apis.WithTokenUsername(clHolder.JwtClientID()),
		apis.WithTokenPassword(clHolder.JwtClientSecret()),
		apis.WithScopes(openid.ConstSystemScopeInternal),
		apis.WithAudiences("service_tenancy"))
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not setup partition client")
	}
	return partitionCli
}

// setupConnectServer initializes and configures the Connect RPC server.
func setupConnectServer(
	ctx context.Context,
	securityMan security.Manager,
	paymentBusiness business.PaymentBusiness,
	profileCli profilev1connect.ProfileServiceClient,
	ledgerCli ledgerv1connect.LedgerServiceClient,
	partitionCli partitionv1connect.PartitionServiceClient,
) http.Handler {
	defaultInterceptorList, err := connectInterceptors.DefaultList(ctx, securityMan.GetAuthenticator(ctx))
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("main -- Could not create default interceptors")
	}

	implementation := handlers.NewPaymentServer(
		paymentBusiness,
		profileCli,
		ledgerCli,
		partitionCli,
	)

	_, serverHandler := paymentv1connect.NewPaymentServiceHandler(
		implementation, connect.WithInterceptors(defaultInterceptorList...))

	mux := http.NewServeMux()
	mux.Handle("/", serverHandler)

	return mux
}
