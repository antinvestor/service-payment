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
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	eventsMan   frameEvents.Manager
	stripeCli   client.StripeClient
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.StripeConfig
	metrics     *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for handling Stripe Payout disbursements.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	stripeCli client.StripeClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.StripeConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		eventsMan:   eventsMan,
		stripeCli:   stripeCli,
		settingsCli: settingsCli,
		cfg:         cfg,
		metrics:     integrationobs.NewMetrics("stripe"),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "stripe.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil
	}

	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := extractCredentials(ctx, headers, h.settingsCli, h.cfg)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "payment", "credentials_error")
		emitStatus(ctx, h.eventsMan, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	var amount int64
	if payment.GetAmount() != nil {
		amount = moneyToCents(payment.GetAmount().GetUnits(), payment.GetAmount().GetNanos())
	}

	currency := "usd"
	if payment.GetAmount() != nil && payment.GetAmount().GetCurrencyCode() != "" {
		currency = payment.GetAmount().GetCurrencyCode()
	}

	destination := ""
	if payment.GetExtra() != nil {
		if dest, ok := payment.GetExtra().GetFields()["stripe_destination"]; ok {
			destination = dest.GetStringValue()
		}
	}

	metadata := map[string]string{
		"payment_id": paymentID,
	}

	// Include tenant/partition in Stripe metadata so they round-trip via webhooks
	if v := headers["tenant_id"]; v != "" {
		metadata["tenant_id"] = v
	}
	if v := headers["partition_id"]; v != "" {
		metadata["partition_id"] = v
	}

	req := &client.PayoutRequest{
		Amount:      amount,
		Currency:    currency,
		Destination: destination,
		Description: "Payment disbursement " + paymentID,
		Metadata:    metadata,
	}

	resp, err := h.stripeCli.CreatePayout(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("create payout failed")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		emitStatus(ctx, h.eventsMan, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	logger.WithField("payout_id", resp.ID).Debug("payout created")

	emitStatus(ctx, h.eventsMan, paymentID, resp.ID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"payout_id":     resp.ID,
		"stripe_status": resp.Status,
		"entity_type":   "payment",
	})

	h.metrics.QueueProcessed(ctx, "payment")
	return nil
}
