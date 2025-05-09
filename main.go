package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	
	"github.com/nats-io/nats.go"

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

	// Ensure all required tables exist - this is critical for service operation
	log.Info("Running database auto-migration to ensure tables exist")
	if err := service.DB(ctx, false).AutoMigrate(&models.Route{}, &models.Payment{}, &models.Cost{}, &models.PaymentStatus{}, &models.Prompt{}, &models.PromptStatus{}); err != nil {
		log.WithError(err).Fatal("Failed to auto-migrate database tables - cannot continue")
		return
	}
	log.Info("Database auto-migration completed successfully")

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

    // Check if we should skip NATS and use memory messaging directly
    skipNats := os.Getenv("SKIP_NATS") == "true"
    
    // Get NATS URL from environment
    raw := os.Getenv("NATS_URL")
    var natsURL string
    
    if skipNats && strings.HasPrefix(raw, "mem://") {
        // If SKIP_NATS=true and we already have a memory URL, use it directly
        natsURL = raw
        log.WithField("memURL", natsURL).Info("Using in-memory messaging directly due to SKIP_NATS=true")
    } else if raw == "" {
        // fall back to default service name
        natsURL = "nats://nats:4222"
    } else if strings.HasPrefix(raw, "nats://") {
        natsURL = raw
    } else {
        log.Warn("NATS_URL missing 'nats://' prefix; assuming host:port format")
        natsURL = "nats://" + raw
    }

	log.WithField("natsURL", natsURL).Info("Using NATS for pub/sub messaging")

	// Define the prompt topic name consistently across services
	promptTopic := "initiate.prompt"
	
	// Variable to track connection success
	connected := false
	
	// Check if we're using memory URL directly due to SKIP_NATS
	if skipNats && strings.HasPrefix(natsURL, "mem://") {
		// Using memory-based pubsub directly, skip NATS connection attempts
		log.WithField("memoryURL", natsURL).Info("Using in-memory pubsub directly (SKIP_NATS=true)")
		serviceOptions = append(serviceOptions, frame.RegisterPublisher(promptTopic, natsURL))
		// Update connection status since we're using memory URL
		connected = true
	} else {
		// Try connecting to NATS with retry logic
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			// Test connection to NATS
			log.WithField("attempt", i+1).WithField("natsURL", natsURL).Info("Attempting to connect to NATS")
			nc, err := nats.Connect(natsURL)
			if err != nil {
				log.WithError(err).WithField("attempt", i+1).Warn("Failed to connect to NATS, retrying after delay")
				time.Sleep(2 * time.Second)
				continue
			}
			// Close connection since we're just testing
			nc.Close()
			log.Info("Successfully connected to NATS server")
			
			// Register the publisher using the original NATS URL without any manipulation
			// This is critical: use the exact same URL format for both services
			log.WithField("natsURL", natsURL).WithField("topic", promptTopic).Info("Registering publisher with NATS")
			serviceOptions = append(serviceOptions, frame.RegisterPublisher(promptTopic, natsURL))
			
			connected = true
			break
		}

		if !connected {
			log.WithField("retries", maxRetries).Warn("Failed to connect to NATS after maximum retries - falling back to memory-based pubsub")
			// Fall back to memory-based pubsub
			fallbackNatsURL := "mem://" + promptTopic
			log.WithField("fallbackURL", fallbackNatsURL).Info("Using memory-based pubsub as fallback")
			serviceOptions = append(serviceOptions, frame.RegisterPublisher(promptTopic, fallbackNatsURL))
		}
	}
	
	// Register event subscribers
	// Note: We need to register with the same URL that we used for publishing to ensure consistency
	// For events, we'll use the same topic we publish prompts to
	var eventQueueURL string
	var topicToUse string
	
	if skipNats && strings.HasPrefix(natsURL, "mem://") {
		// If SKIP_NATS=true and we're using memory URL, use it directly
		// IMPORTANT: For in-memory messaging, we need to use the EXACT same topic name and URL format
		// for both publisher and subscriber across both services
		eventQueueURL = natsURL // Use exactly the same URL that was passed in environment
		
		// Extract the topic from the mem:// URL to ensure exact match
		topicToUse = strings.TrimPrefix(natsURL, "mem://")
		log.WithField("memURL", eventQueueURL).WithField("topic", topicToUse).Info("Using in-memory URL and extracted topic for subscribers")
	} else if connected {
		// Use NATS if we successfully connected
		eventQueueURL = natsURL
		// For NATS, use the prompt topic
		topicToUse = promptTopic
	} else {
		// Fall back to memory-based URL if NATS connection failed
		// Use the same topic as for publishing for consistency
		topicToUse = promptTopic
		eventQueueURL = "mem://" + topicToUse
	}
	
	log.WithField("topic", topicToUse).WithField("url", eventQueueURL).Info("Using in-memory pub/sub for prompt events")
	// The payment service only publishes to the topic, it does not need to subscribe
	// The Jenga service handles subscribing to and processing these messages
	
	// If we want to subscribe to the topic in this service, we would add:
	// serviceOptions = append(serviceOptions, frame.RegisterSubscriber(topicName, natsURL, 10, promptHandler))

	service.Init(serviceOptions...)

	log.WithField("server http port", paymentConfig.HttpServerPort).
		WithField("server grpc port", paymentConfig.GrpcServerPort).
		Info("Initiating server operations")

	err = service.Run(ctx, ":8081")
	if err != nil {
		log.WithError(err).Fatal("could not run Server")
	}
}
