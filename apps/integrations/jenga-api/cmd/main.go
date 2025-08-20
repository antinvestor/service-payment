package main

import (
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/config"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events/events_callback"
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

	serviceName := "service_jenga_api"
	jengaConfig, err := frame.ConfigFromEnv[config.JengaConfig]()
	if err != nil {
		panic(err)
	}
	//nolint:revive // clientApi more readable than clientAPI
	clientApi := coreapi.New(jengaConfig.MerchantCode, jengaConfig.ConsumerSecret, jengaConfig.ApiKey, jengaConfig.Env, jengaConfig.JengaPrivateKey)
	ctx, service := frame.NewService(serviceName, frame.WithConfig(&jengaConfig))
	defer service.Stop(ctx)
	logger := service.Log(ctx).WithField("type", "main")
	// Use environment variable for the gRPC endpoint or default to container service name
	paymentServiceEndpoint := jengaConfig.PaymentServiceURI

	// Initialize the payment client variable
	var paymentClient *paymentV1.PaymentClient
	var clientConn *grpc.ClientConn
	var dialErr error

	// Set up a direct connection to the gRPC server using grpc.DialContext
	clientConn, dialErr = grpc.DialContext(
		ctx,
		paymentServiceEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*16), // 16MB max message size
			grpc.MaxCallSendMsgSize(1024*1024*16),
		),
	)
	if dialErr != nil {
		logger.WithError(dialErr).Error("Failed to connect to payment service")
	} else {

		paymentServiceClient := paymentV1.NewPaymentServiceClient(clientConn)

		paymentClient = &paymentV1.PaymentClient{
			Client: paymentServiceClient,
		}
		logger.Info("Successfully connected to payment service at ", paymentServiceEndpoint)
	}
	// Initialize JobServer
	js := &handler.JobServer{
		Service: service,
	}
	router := router.NewRouter(js)
	initiatePrompt := &events_stk.InitiatePrompt{
		Service:       service,
		Client:        clientApi,
		PaymentClient: *paymentClient,
		CallbackURL:   jengaConfig.JengaCallbackURL,
	}
	createPaymentLink := &events_link_processing.CreatePaymentLink{Service: service, Client: clientApi, PaymentClient: *paymentClient}

	eventHandlers := []frame.EventI{
		&events_callback.JengaCallbackReceivePayment{Service: service, PaymentClient: paymentClient},
		initiatePrompt,
		createPaymentLink,
		&events_tills_pay.JengaTillsPay{Service: service, Client: clientApi},
	}

	// NATS-only configuration
	natsURL := jengaConfig.NATS_URL
	promptTopic := initiatePrompt.Name()
	paymentLinkTopic := createPaymentLink.Name()
	//TODO to ensure to put the topics and the urls in the config file
	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(router),
		frame.WithRegisterEvents(eventHandlers...),
		frame.WithRegisterPublisher(promptTopic, natsURL+promptTopic),
		frame.WithRegisterPublisher(paymentLinkTopic, natsURL+paymentLinkTopic),
		frame.WithRegisterSubscriber(promptTopic, natsURL+promptTopic, initiatePrompt),
		frame.WithRegisterSubscriber(paymentLinkTopic, natsURL+paymentLinkTopic, createPaymentLink),
	}

	service.Init(ctx, serviceOptions...)

	logger.Info("Jenga API service started successfully on port 8080")
	if runErr := service.Run(ctx, ":8080"); runErr != nil {
		logger.WithError(runErr).Fatal("Failed to run Jenga API service")
	}
}
