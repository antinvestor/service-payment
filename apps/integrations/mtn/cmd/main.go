package main

import (
	"context"

	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	apis "github.com/antinvestor/apis/go/common"
	"github.com/antinvestor/apis/go/payment"
	"github.com/antinvestor/apis/go/settings"
	aconfig "github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/handlers"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/queue"
	"github.com/antinvestor/service-payments/internal/events"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/security/openid"
	"github.com/pitabwire/util"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.MtnConfig](ctx)
	if err != nil {
		util.Log(ctx).With("err", err).Error("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "integration_payment_mtn"
	}

	ctx, svc := frame.NewServiceWithContext(ctx, frame.WithConfig(&cfg), frame.WithRegisterServerOauth2Client())
	defer svc.Stop(ctx)

	logger := svc.Log(ctx)
	sm := svc.SecurityManager()
	eventsMan := svc.EventsManager()

	paymentCli, err := setupPaymentClient(ctx, sm, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not setup payment client")
	}

	settingsCli, err := setupSettingsClient(ctx, sm, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not setup settings client")
	}

	mtnCli := client.NewClient()

	webhookServer := handlers.NewMtnWebhookServer(paymentCli)
	paymentWorker := queue.NewPaymentHandler(eventsMan, mtnCli, settingsCli, &cfg)
	promptWorker := queue.NewPromptHandler(eventsMan, mtnCli, settingsCli, &cfg)

	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(webhookServer.NewRouterV1()),
		frame.WithRegisterEvents(events.NewPaymentStatusUpdate(ctx, paymentCli)),
		frame.WithRegisterSubscriber(cfg.QueuePaymentName, cfg.QueuePaymentURI, paymentWorker),
		frame.WithRegisterSubscriber(cfg.QueuePromptName, cfg.QueuePromptURI, promptWorker),
	}

	svc.Init(ctx, serviceOptions...)

	logger.Info("Initiating MTN MoMo integration server operations")
	err = svc.Run(ctx, "")
	if err != nil {
		logger.WithError(err).Error("could not run Server")
	}
}

func setupPaymentClient(
	ctx context.Context,
	clHolder security.InternalOauth2ClientHolder,
	cfg aconfig.MtnConfig,
) (paymentv1connect.PaymentServiceClient, error) {
	return payment.NewClient(ctx,
		apis.WithEndpoint(cfg.PaymentServiceURI),
		apis.WithTokenEndpoint(cfg.GetOauth2TokenEndpoint()),
		apis.WithTokenUsername(clHolder.JwtClientID()),
		apis.WithTokenPassword(clHolder.JwtClientSecret()),
		apis.WithScopes(openid.ConstSystemScopeInternal),
		apis.WithAudiences("service_payment"))
}

func setupSettingsClient(
	ctx context.Context,
	clHolder security.InternalOauth2ClientHolder,
	cfg aconfig.MtnConfig,
) (settingsv1connect.SettingsServiceClient, error) {
	return settings.NewClient(ctx,
		apis.WithEndpoint(cfg.SettingsServiceURI),
		apis.WithTokenEndpoint(cfg.GetOauth2TokenEndpoint()),
		apis.WithTokenUsername(clHolder.JwtClientID()),
		apis.WithTokenPassword(clHolder.JwtClientSecret()),
		apis.WithScopes(openid.ConstSystemScopeInternal),
		apis.WithAudiences("service_settings"))
}
