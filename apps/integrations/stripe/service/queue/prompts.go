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

package queue

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/config"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type promptHandler struct {
	eventsMan   frameEvents.Manager
	stripeCli   client.StripeClient
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.StripeConfig
	metrics     *integrationobs.Metrics
}

// NewPromptHandler creates a queue worker for handling PaymentIntent creation prompts.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	stripeCli client.StripeClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.StripeConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		eventsMan:   eventsMan,
		stripeCli:   stripeCli,
		settingsCli: settingsCli,
		cfg:         cfg,
		metrics:     integrationobs.NewMetrics("stripe"),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "stripe.prompt")
	defer logger.Release()
	logger.Debug("queue handler started")

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil
	}

	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := extractCredentials(ctx, headers, h.settingsCli, h.cfg)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "prompt", "credentials_error")
		emitStatus(ctx, h.eventsMan, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	var amount int64
	if prompt.GetAmount() != nil {
		amount = moneyToCents(prompt.GetAmount().GetUnits(), prompt.GetAmount().GetNanos())
	}

	currency := "usd"
	if prompt.GetAmount() != nil && prompt.GetAmount().GetCurrencyCode() != "" {
		currency = prompt.GetAmount().GetCurrencyCode()
	}

	metadata := map[string]string{
		"prompt_id": promptID,
	}
	if prompt.GetExtra() != nil {
		for k, v := range prompt.GetExtra().GetFields() {
			metadata[k] = v.GetStringValue()
		}
	}

	// Include tenant/partition in Stripe metadata so they round-trip via webhooks
	if v := headers["tenant_id"]; v != "" {
		metadata["tenant_id"] = v
	}
	if v := headers["partition_id"]; v != "" {
		metadata["partition_id"] = v
	}

	req := &client.PaymentIntentRequest{
		Amount:      amount,
		Currency:    currency,
		Description: "Payment prompt " + promptID,
		Metadata:    metadata,
	}

	resp, err := h.stripeCli.CreatePaymentIntent(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("create payment intent failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		emitStatus(ctx, h.eventsMan, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	logger.WithField("payment_intent_id", resp.ID).Debug("payment intent created")

	emitStatus(ctx, h.eventsMan, promptID, resp.ID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"payment_intent_id": resp.ID,
		"client_secret":     resp.ClientSecret,
		"stripe_status":     resp.Status,
		"entity_type":       "prompt",
	})

	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}
