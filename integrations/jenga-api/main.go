package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/config"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events/events_account_balance"
	"github.com/antinvestor/jenga-api/service/events/events_link_processing"
	"github.com/antinvestor/jenga-api/service/events/events_stk"
	"github.com/antinvestor/jenga-api/service/events/events_tills_pay"
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
	jengaConfig, err := frame.ConfigFromEnv[config.JengaConfig]()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
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
	ctx, service := frame.NewService(serviceName, frame.WithConfig(&jengaConfig))
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
	var clientConn *grpc.ClientConn
	var dialErr error

	// Set up a direct connection to the gRPC server using grpc.Dial
	clientConn, dialErr = grpc.Dial(
		paymentServiceEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*16), // 16MB max message size
			grpc.MaxCallSendMsgSize(1024*1024*16),
		),
	)
	if dialErr != nil {
		log.Printf("Warning: Failed to create payment client connection: %v", dialErr)
		// Continue execution - we'll handle the nil client in the handlers
	} else {
		// Note: we're not deferring clientConn.Close() here since we need the connection to stay open
		// Create the payment service client
		paymentServiceClient := paymentV1.NewPaymentServiceClient(clientConn)

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
	var initiatePrompt = &events_stk.InitiatePrompt{Service: service, Client: clientApi, PaymentClient: paymentClient}
	var createPaymentLink = &events_link_processing.CreatePaymentLink{Service: service, Client: clientApi, PaymentClient: paymentClient}
	var tillsPay = &events_tills_pay.JengaTillsPay{Service: service, Client: clientApi}
	// Create service options
	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(router),
		frame.WithRegisterEvents(accountBalance, callbackReceive, stkussd, tillsPay),
	}

	// Set NATS URL explicitly with proper format for cross-container communication
	// This matches the same approach used in the payment service
	raw := os.Getenv("NATS_URL")
	var natsURL string
	if raw == "" {
		// fall back to default service name
		natsURL = "nats://nats:4222"
	} else if strings.HasPrefix(raw, "nats://") {
		natsURL = raw
	} else {
		natsURL = "nats://" + raw
	}

	// CRITICAL: Define consistent prompt topic name - must EXACTLY match payment service
	promptTopic := "initiate.prompt"
	paymentLinkTopic := "create.payment.link"
	log.Printf("Using NATS URL: %s for topic: %s", natsURL, promptTopic)

	// Helper to ensure the NATS URL has the correct subject query parameter
	ensureSubject := func(baseURL, subject string) string {
		if !strings.Contains(baseURL, "nats://") {
			return baseURL
		}
		url := baseURL
		// Remove any existing subject parameter
		if strings.Contains(url, "subject=") {
			parts := strings.Split(url, "?")
			if len(parts) == 2 {
				// Remove subject from query string
				params := strings.Split(parts[1], "&")
				newParams := make([]string, 0, len(params))
				for _, p := range params {
					if !strings.HasPrefix(p, "subject=") {
						newParams = append(newParams, p)
					}
				}
				url = parts[0]
				if len(newParams) > 0 {
					url += "?" + strings.Join(newParams, "&")
				}
			}
		}
		// Add the correct subject parameter
		if strings.Contains(url, "?") {
			url += "&subject=" + subject
		} else {
			url += "?subject=" + subject
		}
		return url
	}

	// Check if we should skip NATS connection attempts
	skipNats := os.Getenv("SKIP_NATS") == "true"

	// Connect to NATS with retry logic - unless explicitly skipped
	connected := false
	maxRetries := 3
	var natsPromptURL, natsPaymentLinkURL string
	if skipNats {
		log.Printf("SKIP_NATS=true detected, skipping NATS connection attempts and using in-memory messaging")
		connected = false // Force using memory-based pubsub
	} else {
		natsPromptURL = ensureSubject(natsURL, promptTopic)
		natsPaymentLinkURL = ensureSubject(natsURL, paymentLinkTopic)
		for i := range maxRetries {
			// Test connection to NATS
			log.Printf("Attempt %d/%d: Connecting to NATS at %s", i+1, maxRetries, natsURL)
			nc, err := nats.Connect(natsURL)
			if err != nil {
				log.Printf("Failed to connect to NATS (attempt %d/%d): %v", i+1, maxRetries, err)
				time.Sleep(2 * time.Second) // Wait before retrying
				continue
			}
			// Close connection since we're just testing
			nc.Close()
			log.Printf("Successfully connected to NATS server")

			// Register the publisher using the NATS URL with correct subject
			log.Printf("Registering publisher for topic '%s' with NATS URL: %s", promptTopic, natsPromptURL)
			pubOption := frame.WithRegisterPublisher(promptTopic, natsPromptURL)
			serviceOptions = append(serviceOptions, pubOption)
			// Register publisher for create.payment.link as well
			log.Printf("Registering publisher for topic 'create.payment.link' with NATS URL: %s", natsPaymentLinkURL)
			pubCreatePaymentLinkOption := frame.WithRegisterPublisher(paymentLinkTopic, natsPaymentLinkURL)
			serviceOptions = append(serviceOptions, pubCreatePaymentLinkOption)

			connected = true
			break
		}
	}

	// If we couldn't connect after all retries, log a warning but continue with memory-based pubsub
	if !connected {
		log.Printf("WARNING: Failed to connect to NATS after %d attempts - falling back to memory-based pubsub", maxRetries)
		// Fall back to memory-based pubsub
		memURL := "mem://" + promptTopic
		log.Printf("Using memory-based pubsub as fallback: %s", memURL)
		serviceOptions = append(serviceOptions, frame.WithRegisterPublisher(promptTopic, memURL))
		// Register publisher for create.payment.link as well (memory fallback)
		memCreatePaymentLinkURL := "mem://create.payment.link"
		log.Printf("Using memory-based pubsub as fallback for create.payment.link: %s", memCreatePaymentLinkURL)
		serviceOptions = append(serviceOptions, frame.WithRegisterPublisher("create.payment.link", memCreatePaymentLinkURL))
	}

	// Register the subscriber using the same URL as we used for the publisher
	// If SKIP_NATS=true or NATS connection failed, use memory URL for both publisher and subscriber

	var promptSubURL, paymentLinkSubURL string

	if skipNats {
		// If SKIP_NATS=true, use the exact same memory URL from env for both pub and sub
		// CRITICAL: Must use the exact same URL format as the payment service
		promptSubURL = os.Getenv("NATS_URL")
		paymentLinkSubURL = os.Getenv("NATS_URL")
		log.Printf("SKIP_NATS=true: Registering subscriber with exact URL from env: %s", promptSubURL)
	} else if connected {
		// If we successfully connected to NATS, use the NATS URL with correct subject
		promptSubURL = natsPromptURL
		paymentLinkSubURL = natsPaymentLinkURL
		log.Printf("Registering subscriber for topic '%s' with NATS URL: %s", promptTopic, promptSubURL)
		log.Printf("Registering subscriber for topic '%s' with NATS URL: %s", paymentLinkTopic, paymentLinkSubURL)
	} else {
		// If NATS connection failed, use memory-based URL (same as publisher fallback)
		promptSubURL = "mem://" + promptTopic
		paymentLinkSubURL = "mem://" + paymentLinkTopic
		log.Printf("Registering subscriber for topic '%s' with memory URL: %s (NATS fallback)", promptTopic, promptSubURL)
		log.Printf("Registering subscriber for topic '%s' with memory URL: %s (NATS fallback)", paymentLinkTopic, paymentLinkSubURL)
	}

	// For in-memory messaging with SKIP_NATS=true, the URL already contains the topic
	// so we need to extract the actual topic name from the URL to ensure exact match
	var topicToSubscribe string
	if skipNats && strings.HasPrefix(promptSubURL, "mem://") {
		// Extract topic from the URL
		topicToSubscribe = strings.TrimPrefix(promptSubURL, "mem://")
		log.Printf("Using exact topic from memory URL: '%s'", topicToSubscribe)
	} else {
		topicToSubscribe = promptTopic
	}

	subOpt := frame.WithRegisterSubscriber(
		topicToSubscribe, // Must match payment service publisher exactly with same string
		promptSubURL,     // Use the same URL based on connection status
		initiatePrompt,   // The handler function for processing prompts
	)
	serviceOptions = append(serviceOptions, subOpt)
	// Register the create.payment.link subscriber
	createPaymentLinkSubOpt := frame.WithRegisterSubscriber(
		"create.payment.link", // Must match payment service publisher exactly with same string
		paymentLinkSubURL,     // Use the same URL based on connection status
		createPaymentLink,     // The handler function for processing payment links
	)
	serviceOptions = append(serviceOptions, createPaymentLinkSubOpt)
	service.Init(ctx, serviceOptions...)

	log.Printf("Starting Jenga API service on :8080")
	if err := service.Run(ctx, ":8080"); err != nil {
		log.Fatalf("failed to run service: %v", err)
	}
}
