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
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/config"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

// ExtraStripeDestination names the external account (ba_… / card_…) or
// connected account a payout goes to. Stripe Payouts only move the Stripe
// balance to the account holder's own bank, so this is required: a payout
// without it would silently pay the merchant's default bank account.
const (
	ExtraStripeDestination = "stripe_destination"
	ExtraPayoutID          = "payout_id"
)

type paymentHandler struct {
	credentialResolver
	emitter   *events.StatusEmitter
	stripeCli client.StripeClient
	metrics   *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for Stripe Payout disbursements.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	stripeCli client.StripeClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.StripeConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		stripeCli:          stripeCli,
		metrics:            integrationobs.NewMetrics(providerName),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "stripe.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil
	}
	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.resolve(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "payment", "credentials_error")
		return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID, err.Error(), nil)
	}

	extras := extraStrings(payment.GetExtra())
	destination := extras[ExtraStripeDestination]
	if destination == "" {
		logger.Error("payout carries no stripe_destination")
		h.metrics.QueueFailed(ctx, "payment", "rejected")
		return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID,
			"Stripe payouts need stripe_destination: the external or connected account to pay", nil)
	}
	amount, err := collection.MinorUnits(payment.GetAmount())
	if err != nil {
		logger.WithError(err).Error("payout amount cannot be sent")
		h.metrics.QueueFailed(ctx, "payment", "rejected")
		return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID, err.Error(), nil)
	}

	metadata := map[string]string{"payment_id": paymentID}
	scopeMetadata(headers, metadata)

	resp, err := h.stripeCli.CreatePayout(ctx, creds, &client.PayoutRequest{
		IdempotencyKey: "payout:" + paymentID,
		Amount:         amount,
		Currency:       payment.GetAmount().GetCurrencyCode(),
		Destination:    destination,
		Description:    "Payment disbursement " + paymentID,
		Metadata:       metadata,
	})
	if err != nil {
		if collection.IsUncertain(err) {
			logger.WithError(err).Warn("create payout outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, "payment", "provider_error")
			return err
		}
		logger.WithError(err).Error("create payout refused")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID, err.Error(), nil)
	}

	logger.WithField("payout_id", resp.ID).Debug("payout created")
	err = h.emitter.Emit(ctx, events.EntityTypePayment, paymentID, resp.ID, commonv1.STATUS_IN_PROCESS, map[string]any{
		ExtraPayoutID:            resp.ID,
		extraStripeStatus:        resp.Status,
		collection.ExtraProvider: providerName,
	})
	if err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, "payment")
	return nil
}
