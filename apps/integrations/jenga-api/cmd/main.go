package main

import (
	"context"

	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	apis "github.com/antinvestor/common"
	"github.com/antinvestor/common/connection"
	aconfig "github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventscallback"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventslinkprocessing"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventsstk"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventstillspay"
	handler "github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/handler"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/util"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.JengaConfig](ctx)
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "service_jenga_api"
	}

	ctx, svc := frame.NewServiceWithContext(ctx, frame.WithConfig(&cfg))
	defer svc.Stop(ctx)

	logger := svc.Log(ctx)

	// Use Frame's HTTP client manager instead of creating our own
	httpClient := svc.HTTPClientManager().Client(ctx)

	//nolint:revive,staticcheck // clientApi more readable than clientAPI
	clientApi := coreapi.New(
		httpClient,
		cfg.MerchantCode,
		cfg.ConsumerSecret,
		cfg.ApiKey,
		cfg.Env,
		cfg.JengaPrivateKey,
	)

	evtsMan := svc.EventsManager()

	paymentCli, err := setupPaymentClient(ctx, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not setup payment client")
	}

	// Initialize event handlers
	initiatePrompt := eventsstk.NewInitiatePrompt(clientApi, paymentCli, cfg.JengaCallbackURL)
	createPaymentLink := eventslinkprocessing.NewCreatePaymentLink(clientApi, paymentCli)
	callbackHandler := eventscallback.NewJengaCallbackReceivePayment(paymentCli)
	tillsPayHandler := eventstillspay.NewJengaTillsPay(clientApi)

	js := handler.NewJobServer(evtsMan, clientApi, paymentCli)

	eventHandlers := []events.EventI{
		callbackHandler,
		initiatePrompt,
		createPaymentLink,
		tillsPayHandler,
	}

	// NATS configuration
	natsURL := cfg.NATS_URL
	promptTopic := initiatePrompt.Name()
	paymentLinkTopic := createPaymentLink.Name()

	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(js.NewRouter()),
		frame.WithRegisterEvents(eventHandlers...),
		frame.WithRegisterPublisher(promptTopic, natsURL+promptTopic),
		frame.WithRegisterPublisher(paymentLinkTopic, natsURL+paymentLinkTopic),
		frame.WithRegisterSubscriber(promptTopic, natsURL+promptTopic, initiatePrompt),
		frame.WithRegisterSubscriber(paymentLinkTopic, natsURL+paymentLinkTopic, createPaymentLink),
	}

	svc.Init(ctx, serviceOptions...)

	logger.Info("Initiating Jenga API integration server operations")
	if runErr := svc.Run(ctx, ""); runErr != nil {
		logger.WithError(runErr).Error("could not run Server")
	}
}

func setupPaymentClient(
	ctx context.Context,
	cfg aconfig.JengaConfig,
) (paymentv1connect.PaymentServiceClient, error) {
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.PaymentServiceURI,
		WorkloadAPITargetPath: cfg.PaymentServiceWorkloadAPITargetPath,
		Audiences:             []string{"service_payment"},
	}, paymentv1connect.NewPaymentServiceClient)
}
