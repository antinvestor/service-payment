package main

import (
	"context"
	"log"
	"os"
	"time"

	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/config"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events/events_account_balance"
	"github.com/antinvestor/jenga-api/service/events/events_stk"
	handler "github.com/antinvestor/jenga-api/service/handler"
	"github.com/antinvestor/jenga-api/service/router"
	"github.com/pitabwire/frame"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Set config file path
	if err := os.Setenv("CONFIG_FILE", "config.yaml"); err != nil {
		log.Fatalf("failed to set config file env: %v", err)
	}

	serviceName := "service_jenga_api"
	var jengaConfig config.JengaConfig
	err := frame.ConfigFillFromEnv(&jengaConfig)
	if err != nil {
		log.Fatalf("failed to process config: %v", err)
		return
	}

	log.Printf("Jenga Config: %+v", jengaConfig)
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

	// Create context with timeout for gRPC connection
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize the payment client variable
	var paymentClient *paymentV1.PaymentClient
	
	// Set up a direct connection to the gRPC server using Dial instead of DialContext
	conn, err := grpc.Dial(
		paymentServiceEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*16), // 16MB max message size
			grpc.MaxCallSendMsgSize(1024*1024*16),
		),
		// Not using WithBlock to avoid connection timeout
	)
	if err != nil {
		log.Printf("Warning: Failed to create payment client connection: %v", err)
		// Continue execution - we'll handle the nil client in the handlers
	} else {
		// Note: we're not deferring conn.Close() here since we need the connection to stay open
		// Create the payment service client
		paymentServiceClient := paymentV1.NewPaymentServiceClient(conn)

		// Create a new PaymentClient with the service client
		paymentClient = &paymentV1.PaymentClient{
			Client: paymentServiceClient,
		}
		log.Printf("Successfully connected to payment service at %s", paymentServiceEndpoint)
	}
	// Initialize JobServer
	js := &handler.JobServer{
		Service: service,
	}

	// Initialize router
	router := router.NewRouter(js)

	// Create service options
	var accountBalance = &events_account_balance.JengaAccountBalance{Service: service, Client: clientApi}
	var callbackReceive = &events_stk.JengaCallbackReceivePayment{Service: service, PaymentClient: paymentClient}
	var stkussd = &events_stk.JengaSTKUSSD{Service: service, Client: clientApi, PaymentClient: paymentClient}

	serviceOptions := []frame.Option{
		frame.HttpHandler(router),
		frame.RegisterEvents(accountBalance, callbackReceive, stkussd),
	}

	// Start the service
	service.Init(serviceOptions...)

	log.Printf("Starting Jenga API service on :8080")
	if err := service.Run(ctx, ":8080"); err != nil {
		log.Fatalf("failed to run service: %v", err)
	}

}
