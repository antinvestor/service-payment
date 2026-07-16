// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	apis "github.com/antinvestor/common/v2"
	"github.com/antinvestor/common/v2/connection"
	"github.com/antinvestor/common/v2/servicecatalog"
	aconfig "github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/handlers"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/queue"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/pitabwire/frame/v2"
	"github.com/pitabwire/frame/v2/config"
	"github.com/pitabwire/util"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.FlutterwaveConfig](ctx)
	if err != nil {
		util.Log(ctx).With("err", err).Error("could not process configs")
		return
	}

	if cfg.Name() == "" {
		cfg.ServiceName = "integration_payment_flutterwave"
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
		logger.WithError(err).Warn("could not setup settings client — per-tenant credentials disabled")
	}

	fwCli := client.NewClient()

	webhookServer := handlers.NewFlutterwaveWebhookServer(paymentCli, fwCli, &cfg)
	paymentWorker := queue.NewPaymentHandler(eventsMan, fwCli, settingsCli, &cfg)
	promptWorker := queue.NewPromptHandler(eventsMan, fwCli, settingsCli, &cfg)

	serviceOptions := []frame.Option{
		frame.WithHTTPHandler(webhookServer.NewRouterV1()),
		frame.WithRegisterEvents(events.NewPaymentStatusUpdate(ctx, paymentCli)),
		frame.WithRegisterSubscriber(cfg.QueuePaymentName, cfg.QueuePaymentURI, paymentWorker),
		frame.WithRegisterSubscriber(cfg.QueuePromptName, cfg.QueuePromptURI, promptWorker),
	}

	if cfg.EnableSubscriptionWorker && cfg.QueueSubscriptionURI != "" {
		subWorker := queue.NewSubscriptionHandler(fwCli, settingsCli, &cfg)
		serviceOptions = append(serviceOptions,
			frame.WithRegisterSubscriber(cfg.QueueSubscriptionName, cfg.QueueSubscriptionURI, subWorker),
		)
		logger.WithField("queue", cfg.QueueSubscriptionURI).Info("subscription lifecycle worker enabled")
	}

	svc.Init(ctx, serviceOptions...)

	logger.Info("Initiating Flutterwave integration server operations")
	testCreds := &client.Credentials{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		PublicKey:    cfg.PublicKey,
		SecretKey:    firstNonEmptyCfg(cfg.SecretKey, cfg.ClientSecret),
	}
	if client.IsV3Credentials(testCreds) {
		logger.Info("Flutterwave classic (v3) secret-key mode enabled")
	} else if cfg.ClientID == "" || cfg.ClientSecret == "" {
		logger.Warn("FLUTTERWAVE_CLIENT_ID / FLUTTERWAVE_CLIENT_SECRET empty — OAuth will fail until set or provided per-message")
	} else {
		logger.Info("Flutterwave v4 OAuth mode enabled")
	}

	err = svc.Run(ctx, "")
	if err != nil {
		logger.WithError(err).Error("could not run Server")
	}
}

func setupPaymentClient(
	ctx context.Context,
	cfg aconfig.FlutterwaveConfig,
) (paymentv1connect.PaymentServiceClient, error) {
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.PaymentServiceURI,
		WorkloadAPITargetPath: cfg.PaymentServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServicePayment,
	}, paymentv1connect.NewPaymentServiceClient)
}

func setupSettingsClient(
	ctx context.Context,
	cfg aconfig.FlutterwaveConfig,
) (settingsv1connect.SettingsServiceClient, error) {
	return connection.NewServiceClient(ctx, &cfg, apis.ServiceTarget{
		Endpoint:              cfg.SettingsServiceURI,
		WorkloadAPITargetPath: cfg.SettingsServiceWorkloadAPITargetPath,
		ServiceID:             servicecatalog.ServiceSettings,
	}, settingsv1connect.NewSettingsServiceClient)
}

func firstNonEmptyCfg(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
