package main

import (
	"fmt"
	"os"
	"strings"

	apis "github.com/antinvestor/apis/go/common"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/config"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/antinvestor/service-payments/service/handlers"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"google.golang.org/grpc"
	_ "gorm.io/driver/postgres"
)

func main() {
	serviceName := "service_payment"
	paymentConfig, err := frame.ConfigFromEnv[config.PaymentConfig]()

	if err != nil {
		panic(fmt.Sprintf("could not load config: %v", err))
	}
	ctx, service := frame.NewService(serviceName, frame.WithConfig(&paymentConfig), frame.WithDatastore())
	defer service.Stop(ctx)
	logger := service.Log(ctx).WithField("type", "main")

	// Run migrations if DO_MIGRATION=true
	if paymentConfig.DO_MIGRATION {
		err = service.MigrateDatastore(ctx, paymentConfig.GetDatabaseMigrationPath(),
			&models.Route{}, &models.Payment{}, &models.Status{}, &models.Prompt{},
			&models.Cost{}, &models.PaymentLink{})
		if err != nil {
			logger.WithError(err).Fatal("could not migrate successfully")
		}
		logger.Info("Migrations completed successfully")
		return
	}

	// Ensure all required tables exist
	db := service.DB(ctx, false)
	if db == nil {
		logger.WithField("DATABASE_URL", os.Getenv("DATABASE_URL")).Fatal("Database connection is nil - check DATABASE_URL and database availability")
		return
	}
	if err := db.AutoMigrate(&models.Route{}, &models.Payment{}, &models.Cost{}, &models.Status{}, &models.Prompt{},  &models.PaymentLink{}); err != nil {
		logger.WithError(err).Fatal("Failed to auto-migrate database tables - cannot continue")
		return
	}

	// OAuth2 and service clients
	oauth2ServiceHost := paymentConfig.GetOauth2ServiceURI()
	oauth2ServiceURL := fmt.Sprintf("%s/oauth2/token", oauth2ServiceHost)
	oauth2ServiceSecret := paymentConfig.Oauth2ServiceClientSecret

	audienceList := []string{}
	if paymentConfig.Oauth2ServiceAudience != "" {
		audienceList = strings.Split(paymentConfig.Oauth2ServiceAudience, ",")
	}

	profileCli, err := profileV1.NewProfileClient(ctx,
		apis.WithEndpoint(paymentConfig.ProfileServiceURI),
		apis.WithTokenEndpoint(oauth2ServiceURL),
		apis.WithTokenUsername(service.JwtClientID()),
		apis.WithTokenPassword(oauth2ServiceSecret),
		apis.WithAudiences(audienceList...))
	if err != nil {
		logger.WithError(err).Fatal("could not setup profile client")
	}

	partitionCli, err := partitionV1.NewPartitionsClient(
		ctx,
		apis.WithEndpoint(paymentConfig.PartitionServiceURI),
		apis.WithTokenEndpoint(oauth2ServiceURL),
		apis.WithTokenUsername(service.JwtClientID()),
		apis.WithTokenPassword(oauth2ServiceSecret),
		apis.WithAudiences(audienceList...))
	if err != nil {
		logger.WithError(err).Fatal("could not setup partition client")
	}

	jwtAudience := paymentConfig.Oauth2JwtVerifyAudience
	if jwtAudience == "" {
		jwtAudience = serviceName
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{}
	streamInterceptors := []grpc.StreamServerInterceptor{}

	if paymentConfig.SecurelyRunService {
		logger.Info("Running service securely with TLS")
		unaryInterceptors = append([]grpc.UnaryServerInterceptor{service.UnaryAuthInterceptor(jwtAudience, paymentConfig.Oauth2JwtVerifyIssuer)}, unaryInterceptors...)
		streamInterceptors = append([]grpc.StreamServerInterceptor{service.StreamAuthInterceptor(jwtAudience, paymentConfig.Oauth2JwtVerifyIssuer)}, streamInterceptors...)
	} else {
		logger.Warn("Service is running insecurely: secure by setting SECURELY_RUN_SERVICE=True")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	implementation := &handlers.PaymentServer{
		Service:      service,
		ProfileCli:   profileCli,
		PartitionCli: partitionCli,
	}

	paymentV1.RegisterPaymentServiceServer(grpcServer, implementation)

	serviceOptions := []frame.Option{
		frame.WithDatastore(),
		frame.WithGRPCServer(grpcServer),
		frame.WithEnableGRPCServerReflection(),
		frame.WithRegisterEvents(
			&events.PaymentSave{Service: service},
			&events.PaymentInQueue{Service: service},
			&events.PaymentOutQueue{Service: service},
			&events.PaymentInRoute{Service: service},
			&events.PaymentOutRoute{Service: service, ProfileCli: profileCli},
			&events.PromptSave{Service: service},
			&events.PaymentLinkSave{Service: service},
			&events.StatusSave{Service: service},
		),
	}

	// Use NATS for pub/sub messaging
	natsURL := paymentConfig.NATS_URL
	promptTopic := paymentConfig.PromptTopic
	paymentLinkTopic := paymentConfig.PaymentLinkTopic


	serviceOptions = append(serviceOptions,
		frame.WithRegisterPublisher(promptTopic, natsURL + promptTopic),
		frame.WithRegisterPublisher(paymentLinkTopic, natsURL + paymentLinkTopic),
	)

	service.Init(ctx, serviceOptions...)

	logger.WithField("server http port", paymentConfig.HTTPServerPort).
		WithField("server grpc port", paymentConfig.GrpcServerPort).
		Info("Initiating server operations")

	if err := service.Run(ctx, ":8081"); err != nil {
		logger.WithError(err).Fatal("could not run Server")
	}
}
