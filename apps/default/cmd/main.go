package main

import (
	"context"
	"fmt"
	"strings"

	apis "github.com/antinvestor/apis/go/common"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/ledger/v1"
	partitionv1 "buf.build/gen/go/antinvestor/partition/protocolbuffers/go/partition/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	aconfig "github.com/antinvestor/service-payments/config"
	"github.com/antinvestor/service-payments/service/business"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/antinvestor/service-payments/service/handlers"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/profiling/service"
	_ "gorm.io/driver/postgres"

	"github.com/pitabwire/frame"
)

func main() {

	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.PaymentConfig](ctx)

	if err != nil {
		panic(fmt.Sprintf("could not load config: %v", err))
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "service_payment"
	}

	ctx, svc := frame.NewService(frame.WithConfig(&cfg), frame.WithRegisterServerOauth2Client(), frame.WithDatastore())
	defer svc.Stop(ctx)
	logger := svc.Log(ctx).WithField("type", "main")

	// Run migrations if DO_MIGRATION=true
	if cfg.DoDatabaseMigrate() {
		err = svc.MigrateDatastore(ctx, cfg.GetDatabaseMigrationPath(),
			&models.Route{}, &models.Payment{}, &models.Status{}, &models.Prompt{},
			&models.Cost{}, &models.PaymentLink{})
		if err != nil {
			logger.WithError(err).Fatal("could not migrate successfully")
		}
		logger.Info("Migrations completed successfully")
		return
	}
	logger.Info("Skipping migrations")

	// Initialize database pool and work manager
	workMan := svc.WorkManager()
	dbPool := svc.DatastoreManager().GetPool(ctx, datastore.DefaultPoolName)

	// Initialize repositories
	paymentRepo := repository.NewPaymentRepository(ctx, dbPool, workMan)
	accountRepo := repository.NewAccountRepository(ctx, dbPool, workMan)
	costRepo := repository.NewCostRepository(ctx, dbPool, workMan)
	statusRepo := repository.NewStatusRepository(ctx, dbPool, workMan)
	promptRepo := repository.NewPromptRepository(ctx, dbPool, workMan)
	paymentLinkRepo := repository.NewPaymentLinkRepository(ctx, dbPool, workMan)
	routeRepo := repository.NewRouteRepository(ctx, dbPool, workMan)

	// OAuth2 and svc clients
	oauth2ServiceHost := cfg.GetOauth2ServiceURI()
	oauth2ServiceURL := fmt.Sprintf("%s/oauth2/token", oauth2ServiceHost)
	oauth2ServiceSecret := cfg.Oauth2ServiceClientSecret

	audienceList := []string{}
	if cfg.Oauth2ServiceAudience != "" {
		audienceList = strings.Split(cfg.Oauth2ServiceAudience, ",")
	}

	profileCli, err := profilev1.NewProfileClient(ctx,
		apis.WithEndpoint(cfg.ProfileServiceURI),
		apis.WithTokenEndpoint(oauth2ServiceURL),
		apis.WithTokenUsername(svc.JwtClientID()),
		apis.WithTokenPassword(oauth2ServiceSecret),
		apis.WithAudiences(audienceList...))
	if err != nil {
		logger.WithError(err).Fatal("could not setup profile client")
	}

	partitionCli, err := partitionV1.NewPartitionsClient(
		ctx,
		apis.WithEndpoint(cfg.PartitionServiceURI),
		apis.WithTokenEndpoint(oauth2ServiceURL),
		apis.WithTokenUsername(svc.JwtClientID()),
		apis.WithTokenPassword(oauth2ServiceSecret),
		apis.WithAudiences(audienceList...))
	if err != nil {
		logger.WithError(err).Fatal("could not setup partition client")
	}

	ledgerCli, err := ledgerv1.NewLedgerClient(ctx,
		apis.WithEndpoint(cfg.LedgerServiceURI),
		apis.WithTokenEndpoint(oauth2ServiceURL),
		apis.WithTokenUsername(svc.JwtClientID()),
		apis.WithTokenPassword(oauth2ServiceSecret),
		apis.WithAudiences(audienceList...))
	if err != nil {
		logger.WithError(err).Fatal("could not setup ledger client")
	}

	// Initialize business layer with proper dependency injection
	paymentBusiness, err := business.NewPaymentBusiness(
		ctx, svc, profileCli, partitionCli, ledgerCli,
		paymentRepo, statusRepo, costRepo, accountRepo, promptRepo, paymentLinkRepo,
	)
	if err != nil {
		logger.WithError(err).Fatal("could not initialize payment business")
	}

	jwtAudience := cfg.Oauth2JwtVerifyAudience
	if jwtAudience == "" {
		jwtAudience = serviceName
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{}
	streamInterceptors := []grpc.StreamServerInterceptor{}

	if cfg.SecurelyRunService {
		logger.Info("Running svc securely with TLS")
		unaryInterceptors = append(
			[]grpc.UnaryServerInterceptor{
				svc.UnaryAuthInterceptor(jwtAudience, cfg.Oauth2JwtVerifyIssuer),
			},
			unaryInterceptors...)
		streamInterceptors = append(
			[]grpc.StreamServerInterceptor{
				svc.StreamAuthInterceptor(jwtAudience, cfg.Oauth2JwtVerifyIssuer),
			},
			streamInterceptors...)
	} else {
		logger.Warn("Service is running insecurely: secure by setting SECURELY_RUN_SERVICE=True")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	implementation := &handlers.PaymentServer{
		Service:         svc,
		PaymentBusiness: paymentBusiness,
		ProfileCli:      profileCli,
		PartitionCli:    partitionCli,
		LedgerCli:       ledgerCli,
	}

	paymentv1.RegisterPaymentServiceServer(grpcServer, implementation)

	serviceOptions := []frame.Option{
		frame.WithDatastore(),
		frame.WithGRPCServer(grpcServer),
		frame.WithEnableGRPCServerReflection(),
		frame.WithRegisterEvents(
			&events.PaymentSave{Service: svc},
			&events.PaymentInQueue{Service: svc},
			&events.PaymentOutQueue{Service: svc},
			&events.PaymentInRoute{Service: svc},
			&events.PaymentOutRoute{Service: svc, ProfileCli: profileCli},
			&events.PromptSave{Service: svc},
			&events.PaymentLinkSave{Service: svc},
			&events.StatusSave{Service: svc},
		),
	}

	// Use NATS for pub/sub messaging
	natsURL := cfg.NatsURL
	promptTopic := cfg.PromptTopic
	paymentLinkTopic := cfg.PaymentLinkTopic

	serviceOptions = append(serviceOptions,
		frame.WithRegisterPublisher(promptTopic, natsURL+promptTopic),
		frame.WithRegisterPublisher(paymentLinkTopic, natsURL+paymentLinkTopic),
	)

	svc.Init(ctx, serviceOptions...)

	logger.WithField("server http port", cfg.HTTPServerPort).
		WithField("server grpc port", cfg.GrpcServerPort).
		Info("Initiating server operations")

	if runErr := svc.Run(ctx, ":8081"); runErr != nil {
		logger.WithError(runErr).Fatal("could not run Server")
	}
}
