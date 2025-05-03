package main

import (
	"fmt"
	"strings"

	apis "github.com/antinvestor/apis/go/common"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	"github.com/antinvestor/service-payments/config"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/antinvestor/service-payments/service/handlers"
	"github.com/antinvestor/service-payments/service/models"

	"github.com/sirupsen/logrus"

	paymentV1 "github.com/antinvestor/apis/go/payment/v1"

	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/pitabwire/frame"
	"google.golang.org/grpc"
)

func main() {
	serviceName := "service_payment"

	var paymentConfig config.PaymentConfig
	err := frame.ConfigFillFromEnv(&paymentConfig)
	if err != nil {
		logrus.WithError(err).Fatal("could not process configs")
		return
	}

	ctx, service := frame.NewService(serviceName, frame.Config(&paymentConfig))
	defer service.Stop(ctx)

	log := service.L(ctx).WithField("type", "main")

	log.Info("starting service...")
	serviceOptions := []frame.Option{frame.Datastore(ctx)}
	
	// Initialize service with database connection
	service.Init(serviceOptions...)

	if paymentConfig.DoDatabaseMigrate() {
		err = service.MigrateDatastore(ctx, paymentConfig.GetDatabaseMigrationPath(),
			&models.Route{}, &models.Payment{}, &models.PaymentStatus{}, &models.Prompt{}, &models.PromptStatus{})

		if err != nil {
			log.WithError(err).Fatal("could not migrate successfully")
		}
		return
	}

	err = service.RegisterForJwt(ctx)
	if err != nil {
		log.WithError(err).Fatal("main -- could not register for jwt")
	}

	// Ensure prompt tables exist without requiring full migration
	if err := service.DB(ctx, false).AutoMigrate(&models.Route{}, &models.Payment{}, &models.PaymentStatus{}, &models.Prompt{}, &models.PromptStatus{}, ); err != nil {
		log.WithError(err).Warn("Failed to auto-migrate prompt tables - some features may not work")
		// Continue execution, don't fail the entire service
	}

	oauth2ServiceHost := paymentConfig.GetOauth2ServiceURI()
	oauth2ServiceURL := fmt.Sprintf("%s/oauth2/token", oauth2ServiceHost)
	oauth2ServiceSecret := paymentConfig.Oauth2ServiceClientSecret

	audienceList := make([]string, 0)
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
		log.WithError(err).Fatal("could not setup profile client")
	}

	partitionCli, err := partitionV1.NewPartitionsClient(
		ctx,
		apis.WithEndpoint(paymentConfig.PartitionServiceURI),
		apis.WithTokenEndpoint(oauth2ServiceURL),
		apis.WithTokenUsername(service.JwtClientID()),
		apis.WithTokenPassword(oauth2ServiceSecret),
		apis.WithAudiences(audienceList...))
	if err != nil {
		log.WithError(err).Fatal("could not setup partition client")
	}

	jwtAudience := paymentConfig.Oauth2JwtVerifyAudience
	if jwtAudience == "" {
		jwtAudience = serviceName
	}

	// Skip the validator for now since there's a type incompatibility
	// The grpc-ecosystem interceptor expects a different validator type than what we're creating
	unaryInterceptors := []grpc.UnaryServerInterceptor{}

	streamInterceptors := []grpc.StreamServerInterceptor{}

	// Check if the service should run securely
	if paymentConfig.SecurelyRunService {
		log.Info("Running service securely with TLS")
		unaryInterceptors = append([]grpc.UnaryServerInterceptor{service.UnaryAuthInterceptor(jwtAudience, paymentConfig.Oauth2JwtVerifyIssuer)}, unaryInterceptors...)
		streamInterceptors = append([]grpc.StreamServerInterceptor{service.StreamAuthInterceptor(jwtAudience, paymentConfig.Oauth2JwtVerifyIssuer)}, streamInterceptors...)
	} else {
		log.Warn("Service is running insecurely: secure by setting SECURELY_RUN_SERVICE=True")
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

	grpcServerOpt := frame.GrpcServer(grpcServer)
	serviceOptions = append(serviceOptions, grpcServerOpt, frame.EnableGrpcServerReflection())

	serviceOptions = append(serviceOptions,
		frame.RegisterEvents(
			&events.PaymentSave{Service: service},
			&events.PaymentStatusSave{Service: service},
			&events.PaymentInQueue{Service: service},
			&events.PaymentOutQueue{Service: service},
			&events.PaymentInRoute{Service: service},
			&events.PaymentOutRoute{Service: service, ProfileCli: profileCli},
			&events.PromptSave{Service: service},
			&events.PromptStatusSave{Service: service},
		))

	service.Init(serviceOptions...)

	log.WithField("server http port", paymentConfig.HttpServerPort).
		WithField("server grpc port", paymentConfig.GrpcServerPort).
		Info("Initiating server operations")

	err = service.Run(ctx, ":8081")
	if err != nil {
		log.WithError(err).Fatal("could not run Server")
	}
}
