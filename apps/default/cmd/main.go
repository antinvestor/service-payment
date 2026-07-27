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
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v1/ledgerv1connect"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	paymentpbv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"buf.build/gen/go/antinvestor/tenancy/connectrpc/go/tenancy/v1/tenancyv1connect"
	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	apis "github.com/antinvestor/common/v2"
	"github.com/antinvestor/common/v2/connection"
	"github.com/antinvestor/common/v2/permissions"
	"github.com/antinvestor/common/v2/servicecatalog"
	aconfig "github.com/antinvestor/service-payments/apps/default/config"
	"github.com/antinvestor/service-payments/apps/default/service/authz"
	"github.com/antinvestor/service-payments/apps/default/service/business"
	"github.com/antinvestor/service-payments/apps/default/service/events"
	"github.com/antinvestor/service-payments/apps/default/service/handlers"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/v2"
	"github.com/pitabwire/frame/v2/config"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/frame/v2/security/authorizer"
	connectInterceptors "github.com/pitabwire/frame/v2/security/interceptors/connect"
	"github.com/pitabwire/frame/v2/setup"
	"github.com/pitabwire/util"
)

//go:embed spec/payment.openapi.yaml
var paymentAPISpecFile []byte

func main() { //nolint:funlen // service wiring requires sequential setup
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
		frame.WithDatastore(),
	)

	svc.Setup().RegisterFunc(setup.NameMigrate, func(ctx context.Context) error {
		return repository.Migrate(ctx, svc.DatastoreManager(), cfg.GetDatabaseMigrationPath())
	})
	defer svc.Stop(ctx)
	log := svc.Log(ctx)

	sm := svc.SecurityManager()
	workMan := svc.WorkManager()

	evtsMan := svc.EventsManager()
	qMan := svc.QueueManager()

	dbManager := svc.DatastoreManager()
	dbPool := dbManager.GetPool(ctx, datastore.DefaultPoolName)

	// Handle database migration if requested
	// Setup clients
	profileCli := setupProfileClient(ctx, cfg)
	ledgerCli := setupLedgerClient(ctx, cfg)
	tenancyCli := setupTenancyClient(ctx, cfg)

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
		ctx, workMan, qMan, cfg.InitiatePromptTopicName, cfg.PaymentLinkTopicName,
		evtsMan, profileCli, tenancyCli, ledgerCli,
		paymentRepo, statusRepo, costRepo, accountRepo, promptRepo, paymentLinkRepo,
	)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not initialize payment business")
	}

	// Setup Connect server
	connectHandler := setupConnectServer(
		ctx,
		sm,
		paymentBusiness,
		profileCli,
		ledgerCli,
		tenancyCli,
	)

	// Register permission manifest for the payment service namespace.
	paymentSD := paymentpbv1.File_v1_payment_proto.Services().ByName("PaymentService")

	// Setup HTTP handlers
	serviceOptions := []frame.Option{
		frame.WithDatastore(),
		frame.WithHTTPHandler(connectHandler),
		frame.WithPermissionRegistration(paymentSD),
	}

	// Register queue publishers (default + optional per-route prompt fan-out).
	promptPublisher := frame.WithRegisterPublisher(
		cfg.InitiatePromptTopicName,
		cfg.InitiatePromptTopicURI,
	)
	paymentLinkPublisher := frame.WithRegisterPublisher(
		cfg.PaymentLinkTopicName,
		cfg.PaymentLinkTopicURI,
	)
	serviceOptions = append(serviceOptions, promptPublisher, paymentLinkPublisher)
	serviceOptions = append(serviceOptions, registerPromptRoutePublishers(ctx, cfg)...)

	// Register event handlers with proper constructors
	serviceOptions = append(serviceOptions,
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
			events.NewStatusSave(statusRepo, paymentRepo),
		))

	// Initialize the service with all options
	svc.Init(ctx, serviceOptions...)

	if frame.ShouldRunSetup(&cfg) {
		if setupErr := svc.RunSetupForProcess(ctx, &cfg); setupErr != nil {
			util.Log(ctx).WithError(setupErr).Fatal("setup plan failed")
		}
		return
	}

	// Start the service
	err = svc.Run(ctx, "")
	if err != nil {
		log.WithError(err).Fatal("could not run Server")
	}
}

// registerPromptRoutePublishers parses INITIATE_PROMPT_ROUTE_URIS JSON and
// registers one publisher per route as "prompt.<route>".
func registerPromptRoutePublishers(ctx context.Context, cfg aconfig.PaymentConfig) []frame.Option {
	raw := strings.TrimSpace(cfg.InitiatePromptRouteURIs)
	if raw == "" {
		return nil
	}
	var routes map[string]string
	if err := json.Unmarshal([]byte(raw), &routes); err != nil {
		util.Log(ctx).WithError(err).Error("INITIATE_PROMPT_ROUTE_URIS is not valid JSON — ignoring route fan-out")
		return nil
	}
	opts := make([]frame.Option, 0, len(routes))
	for route, uri := range routes {
		route = strings.TrimSpace(strings.ToLower(route))
		uri = strings.TrimSpace(uri)
		if route == "" || uri == "" {
			continue
		}
		ref := "prompt." + route
		util.Log(ctx).WithField("route", route).WithField("publisher", ref).Info("registering initiate-prompt route publisher")
		opts = append(opts, frame.WithRegisterPublisher(ref, uri))
	}
	return opts
}

// setupProfileClient creates and configures the profile service client.
func setupProfileClient(
	ctx context.Context,
	cfg aconfig.PaymentConfig,
) profilev1connect.ProfileServiceClient {
	profileCli, err := connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.ProfileServiceURI,
		WorkloadAPITargetPath: cfg.ProfileServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServiceProfile,
	}, profilev1connect.NewProfileServiceClient)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not setup profile client")
	}
	return profileCli
}

// setupLedgerClient creates and configures the ledger service client.
func setupLedgerClient(
	ctx context.Context,
	cfg aconfig.PaymentConfig,
) ledgerv1connect.LedgerServiceClient {
	ledgerCli, err := connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.LedgerServiceURI,
		WorkloadAPITargetPath: cfg.LedgerServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServiceLedger,
	}, ledgerv1connect.NewLedgerServiceClient)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not setup ledger client")
	}
	return ledgerCli
}

// setupTenancyClient creates and configures the partition service client.
func setupTenancyClient(
	ctx context.Context,
	cfg aconfig.PaymentConfig,
) tenancyv1connect.TenancyServiceClient {
	tenancyCli, err := connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.TenancyServiceURI,
		WorkloadAPITargetPath: cfg.TenancyServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServiceTenancy,
	}, tenancyv1connect.NewTenancyServiceClient)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not setup partition client")
	}
	return tenancyCli
}

// setupConnectServer initializes and configures the Connect RPC server.
func setupConnectServer(
	ctx context.Context,
	securityMan security.Manager,
	paymentBusiness business.PaymentBusiness,
	profileCli profilev1connect.ProfileServiceClient,
	ledgerCli ledgerv1connect.LedgerServiceClient,
	tenancyCli tenancyv1connect.TenancyServiceClient,
) http.Handler {
	auth := securityMan.GetAuthorizer(ctx)

	// Layer 1: TenancyAccessChecker verifies caller can access the partition.
	tenancyAccessChecker := authorizer.NewTenancyAccessChecker(auth, authz.NamespaceTenancyAccess)
	tenancyAccessInterceptor := connectInterceptors.NewTenancyAccessInterceptor(
		tenancyAccessChecker,
	)

	// Layer 2: FunctionAccessInterceptor enforces per-RPC permissions automatically.
	sd := paymentpbv1.File_v1_payment_proto.Services().ByName("PaymentService")
	procMap := permissions.BuildProcedureMap(sd)
	functionChecker := authorizer.NewFunctionChecker(auth, permissions.ForService(sd).Namespace)
	functionAccessInterceptor := connectInterceptors.NewFunctionAccessInterceptor(
		functionChecker,
		procMap,
	)

	defaultInterceptorList, err := connectInterceptors.DefaultList(
		ctx, securityMan.GetAuthenticator(ctx), tenancyAccessInterceptor, functionAccessInterceptor)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("main -- Could not create default interceptors")
	}

	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("main -- could not configure open telemetry")
	}
	interceptors := append([]connect.Interceptor{otelInterceptor}, defaultInterceptorList...)

	implementation := handlers.NewPaymentServer(
		paymentBusiness,
		profileCli,
		ledgerCli,
		tenancyCli,
	)

	_, serverHandler := paymentv1connect.NewPaymentServiceHandler(
		implementation, connect.WithInterceptors(interceptors...))

	mux := http.NewServeMux()
	mux.Handle("/", serverHandler)
	mux.Handle("/openapi.yaml", apis.NewOpenAPIHandler(paymentAPISpecFile, nil))

	return mux
}
