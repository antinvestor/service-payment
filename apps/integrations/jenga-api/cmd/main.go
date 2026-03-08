package main

import (
	"context"

	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	apis "github.com/antinvestor/apis/go/common"
	"github.com/antinvestor/apis/go/payment"
	aconfig "github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventscallback"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventslinkprocessing"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventsstk"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventstillspay"
	handler "github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/handler"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/router"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/util"
)

const internalSystemScope = "system_int"

func main() {
	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.JengaConfig](ctx)
	if err != nil {
		panic(err)
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "service_jenga_api"
	}

	ctx, svc := frame.NewServiceWithContext(ctx, frame.WithConfig(&cfg))
	defer svc.Stop(ctx)
	logger := util.Log(ctx).WithField("type", "main")

	// Setup Jenga API client
	//nolint:revive,staticcheck // clientApi more readable than clientAPI
	clientApi := coreapi.New(
		cfg.MerchantCode,
		cfg.ConsumerSecret,
		cfg.ApiKey,
		cfg.Env,
		cfg.JengaPrivateKey,
	)

	// Get managers
	evtsMan := svc.EventsManager()
	// Setup payment service client using Connect RPC
	paymentCli := setupPaymentClient(ctx, cfg)

	// Initialize event handlers using constructors
	initiatePrompt := eventsstk.NewInitiatePrompt(
		clientApi,
		paymentCli,
		cfg.JengaCallbackURL,
	)

	createPaymentLink := eventslinkprocessing.NewCreatePaymentLink(
		clientApi,
		paymentCli,
	)

	callbackHandler := eventscallback.NewJengaCallbackReceivePayment(
		paymentCli,
	)

	tillsPayHandler := eventstillspay.NewJengaTillsPay(
		clientApi,
	)

	// Initialize JobServer with dependencies
	js := handler.NewJobServer(
		evtsMan,
		clientApi,
		paymentCli,
	)

	httpRouter := router.NewRouter(js)

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
		frame.WithHTTPHandler(httpRouter),
		frame.WithRegisterEvents(eventHandlers...),
		frame.WithRegisterPublisher(promptTopic, natsURL+promptTopic),
		frame.WithRegisterPublisher(paymentLinkTopic, natsURL+paymentLinkTopic),
		frame.WithRegisterSubscriber(promptTopic, natsURL+promptTopic, initiatePrompt),
		frame.WithRegisterSubscriber(paymentLinkTopic, natsURL+paymentLinkTopic, createPaymentLink),
	}

	svc.Init(ctx, serviceOptions...)

	logger.Info("Jenga API service started successfully on port 8080")
	if runErr := svc.Run(ctx, ":8080"); runErr != nil {
		logger.WithError(runErr).Fatal("Failed to run Jenga API service")
	}
}

func setupPaymentClient(
	ctx context.Context,
	cfg aconfig.JengaConfig,
) paymentv1connect.PaymentServiceClient {
	ledgerCli, err := payment.NewClient(ctx,
		apis.WithEndpoint(cfg.PaymentServiceURI),
		apis.WithTokenEndpoint(cfg.GetOauth2TokenEndpoint()),
		apis.WithTokenUsername(cfg.GetOauth2ServiceClientID()),
		apis.WithTokenPassword(cfg.GetOauth2ServiceClientSecret()),
		apis.WithScopes(internalSystemScope),
		apis.WithAudiences("service_payment"))
	if err != nil {
		util.Log(ctx).WithError(err).Fatal("could not setup ledger client")
	}
	return ledgerCli
}
