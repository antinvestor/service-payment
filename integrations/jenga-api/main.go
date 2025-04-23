package main

import (
	"log"
	"os"

	commonV1 "github.com/antinvestor/apis/go/common"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/config"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events"
	handler "github.com/antinvestor/jenga-api/service/handler"
	"github.com/antinvestor/jenga-api/service/router"
	"github.com/pitabwire/frame"
)

func main() {
	// Set config file path
	if err := os.Setenv("CONFIG_FILE", "config.yaml"); err != nil {
		log.Fatalf("failed to set config file env: %v", err)
	}

	serviceName := "service_jenga_api"
	var jengaConfig config.JengaConfig
	err := frame.ConfigProcess("", &jengaConfig)
	if err != nil {
		log.Fatalf("failed to process config: %v", err)
		return
	}
	//initialize jenga client
	if jengaConfig.MerchantCode == "" {
		log.Fatalf("MerchantCode is required")
		return
	}
	clientApi := coreapi.New(jengaConfig.MerchantCode, jengaConfig.ConsumerSecret, jengaConfig.ApiKey, jengaConfig.Env, jengaConfig.JengaPrivateKey)

	// Initialize payment client
	ctx, service := frame.NewService(serviceName, frame.Config(&jengaConfig))
	defer service.Stop(ctx)
	// Use environment variable for the gRPC endpoint or default to container service name
	paymentServiceEndpoint := os.Getenv("PAYMENT_SERVICE_ENDPOINT")
	if paymentServiceEndpoint == "" {
		// When running in Docker, we should use the service name from docker-compose
		// as defined in docker-compose.yml for the payment service
		paymentServiceEndpoint = "localhost:50051"
	}

	log.Printf("Attempting to connect to payment service at: %s", paymentServiceEndpoint)

	// When running in Docker environment, make sure networkss are properly linked
	// or that host IPs are accessible depending on your Docker network setup

	// Use the proper client initialization function
	paymentClient, err := paymentV1.NewPaymentsClient(
		ctx,
		commonV1.WithEndpoint(paymentServiceEndpoint), // Important for non-TLS connections
	)

	if err != nil {
		log.Printf("Warning: Failed to create payment client: %v", err)
		// Continue execution - we'll handle the nil client in the handlers
	} else {
		log.Printf("Successfully connected to payment service at %s", paymentServiceEndpoint)
	}
	// Initialize JobServer
	js := &handler.JobServer{
		Service: service,
	}

	// Initialize router
	router := router.NewRouter(js)

	// Create service options
	serviceOptions := []frame.Option{
		frame.HttpHandler(router),
		frame.RegisterEvents(
			&events.JengaAccountBalance{Service: service, Client: clientApi},
			&events.JengaCallbackReceivePayment{Service: service, PaymentClient: paymentClient},
			&events.JengaSTKUSSD{Service: service, Client: clientApi, PaymentClient: paymentClient},
		),
	}

	// Start the service
	service.Init(serviceOptions...)

	log.Printf("Starting Jenga API service on :8080")
	err = service.Run(ctx, "")
	if err != nil {
		log.Fatalf("failed to run service: %v", err)
	}

}
