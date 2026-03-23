package main

import (
	"context"

	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	apis "github.com/antinvestor/apis/go/common"
	"github.com/antinvestor/apis/go/payment"
	"github.com/antinvestor/apis/go/settings"
	aconfig "github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/handlers"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/queue"
	"github.com/antinvestor/service-payments/internal/events"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/config"
	"github.com/pitabwire/util"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.MpesaConfig](ctx)
	if err != nil {
		util.Log(ctx).With("err", err).Error("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "integration_payment_mpesa"
	}

	ctx, svc := frame.NewServiceWithContext(ctx, frame.WithConfig(&cfg))
	defer svc.Stop(ctx)

	logger := svc.Log(ctx)
	eventsMan := svc.EventsManager()

	paymentCli, err := setupPaymentClient(ctx, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not setup payment client")
	}

	settingsCli, err := setupSettingsClient(ctx, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not setup settings client")
	}

	mpesaCli := client.NewClient()

	webhookServer := handlers.NewMpesaWebhookServer(paymentCli)
	paymentWorker := queue.NewPaymentHandler(eventsMan, mpesaCli, settingsCli, &cfg)
	promptWorker := queue.NewPromptHandler(eventsMan, mpesaCli, settingsCli, &cfg)

	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(webhookServer.NewRouterV1()),
		frame.WithRegisterEvents(events.NewPaymentStatusUpdate(ctx, paymentCli)),
		frame.WithRegisterSubscriber(cfg.QueuePaymentName, cfg.QueuePaymentURI, paymentWorker),
		frame.WithRegisterSubscriber(cfg.QueuePromptName, cfg.QueuePromptURI, promptWorker),
	}

	svc.Init(ctx, serviceOptions...)

	logger.Info("Initiating M-Pesa integration server operations")
	err = svc.Run(ctx, "")
	if err != nil {
		logger.WithError(err).Error("could not run Server")
	}
}

func setupPaymentClient(
	ctx context.Context,
	cfg aconfig.MpesaConfig,
) (paymentv1connect.PaymentServiceClient, error) {
	return payment.NewClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.PaymentServiceURI,
		WorkloadAPITargetPath: cfg.PaymentServiceWorkloadAPITargetPath,
		Audiences:             []string{"service_payment"},
	})
}

func setupSettingsClient(
	ctx context.Context,
	cfg aconfig.MpesaConfig,
) (settingsv1connect.SettingsServiceClient, error) {
	return settings.NewClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.SettingsServiceURI,
		WorkloadAPITargetPath: cfg.SettingsServiceWorkloadAPITargetPath,
		Audiences:             []string{"service_setting"},
	})
}
