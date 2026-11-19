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
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
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
	"google.golang.org/protobuf/types/known/structpb"
)

// Extras returned on the prompt status once Stripe holds the collection.
const (
	ExtraCheckoutSessionID = "checkout_session_id"
	ExtraPaymentIntentID   = "payment_intent_id"
	extraStripeStatus      = "stripe_status"
	// extraDescription is the optional customer-facing line item name.
	extraDescription = "description"

	providerName = "stripe"
	// pricePrefix is how Stripe names catalog prices ("price_…").
	pricePrefix = "price_"
)

// metadataKeys is the whitelist of prompt extras forwarded to Stripe
// metadata (prompt_id and the queue scope are always added). Everything else
// in the extras (card ciphertext, PINs, …) stays here.
var metadataKeys = []string{
	collection.ExtraSessionRef, collection.ExtraSubscriptionID, collection.ExtraInvoiceID,
	collection.ExtraOrderRef, collection.ExtraCustomerExternalID,
}

type promptHandler struct {
	credentialResolver
	emitter    *events.StatusEmitter
	stripeCli  client.StripeClient
	paymentCli paymentv1connect.PaymentServiceClient
	metrics    *integrationobs.Metrics
}

// NewPromptHandler creates the worker that turns a prompt into a hosted
// Checkout Session (or a bare PaymentIntent when no success_url is given).
// paymentCli (optional) lets a redelivered prompt find the session it
// already opened instead of opening a second one.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	stripeCli client.StripeClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	paymentCli paymentv1connect.PaymentServiceClient,
	cfg *config.StripeConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		stripeCli:          stripeCli,
		paymentCli:         paymentCli,
		metrics:            integrationobs.NewMetrics(providerName),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "stripe.prompt")
	defer logger.Release()

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil
	}
	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	if existing := h.openedReference(ctx, promptID); existing != "" {
		logger.WithField("stripe_id", existing).Info("Stripe object already opened for prompt, skipping")
		h.metrics.QueueProcessed(ctx, "prompt")
		return nil
	}

	creds, err := h.resolve(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "prompt", "credentials_error")
		return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, err.Error(), nil)
	}

	extras := extraStrings(prompt.GetExtra())
	metadata := map[string]string{"prompt_id": promptID}
	for _, key := range metadataKeys {
		if v := extras[key]; v != "" {
			metadata[key] = v
		}
	}
	scopeMetadata(headers, metadata)

	successURL := firstNonEmpty(extras[collection.ExtraSuccessURL], extras[collection.ExtraRedirectURL])
	if successURL != "" {
		return h.openCheckoutSession(ctx, logger, creds, &prompt, extras, metadata, successURL)
	}
	return h.createPaymentIntent(ctx, logger, creds, &prompt, metadata)
}

func (h *promptHandler) openCheckoutSession(
	ctx context.Context,
	logger *util.LogEntry,
	creds *client.StripeCredentials,
	prompt *paymentv1.InitiatePromptRequest,
	extras, metadata map[string]string,
	successURL string,
) error {
	promptID := prompt.GetId()
	priceID := extras[collection.ExtraProductID]
	recurring := strings.EqualFold(extras[collection.ExtraRecurring], "true")

	req := &client.CheckoutSessionRequest{
		IdempotencyKey:    "prompt:" + promptID,
		Mode:              client.CheckoutModePayment,
		SuccessURL:        successURL,
		CancelURL:         firstNonEmpty(extras[collection.ExtraCancelURL], successURL),
		CustomerEmail:     firstNonEmpty(extras[collection.ExtraCustomerEmail], extras[collection.ExtraEmail]),
		ClientReferenceID: promptID,
		Metadata:          metadata,
	}
	switch {
	case priceID != "":
		// A catalog price: Stripe prices it, and a price id (or an explicit
		// recurring flag) means a subscription unless recurring is "false".
		req.PriceID = priceID
		if recurring || (strings.HasPrefix(priceID, pricePrefix) &&
			!strings.EqualFold(extras[collection.ExtraRecurring], "false")) {
			req.Mode = client.CheckoutModeSubscription
		}
	case recurring:
		logger.Error("recurring prompt carries no product_id; Stripe subscriptions need a price")
		h.metrics.QueueFailed(ctx, "prompt", "rejected")
		return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID,
			"Stripe subscriptions need product_id: the plan's Stripe price", nil)
	default:
		amount, err := collection.MinorUnits(prompt.GetAmount())
		if err != nil {
			logger.WithError(err).Error("prompt amount cannot be charged")
			h.metrics.QueueFailed(ctx, "prompt", "rejected")
			return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, err.Error(), nil)
		}
		req.Amount = amount
		req.Currency = prompt.GetAmount().GetCurrencyCode()
		// The line item name the payer sees on the hosted page.
		req.ProductName = firstNonEmpty(extras[extraDescription], extras[collection.ExtraOrderRef],
			"Payment "+promptID)
	}

	resp, err := h.stripeCli.CreateCheckoutSession(ctx, creds, req)
	if err != nil {
		if collection.IsUncertain(err) {
			// Stripe may have opened the session: redeliver, the idempotency
			// key then returns the same session on the next attempt.
			logger.WithError(err).Warn("create checkout session outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, "prompt", "provider_error")
			return err
		}
		logger.WithError(err).Error("create checkout session refused")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, err.Error(), nil)
	}

	logger.WithField("checkout_session_id", resp.ID).WithField("mode", resp.Mode).Debug("checkout session created")
	statusExtras := map[string]any{
		ExtraCheckoutSessionID:         resp.ID,
		collection.ExtraCheckoutURL:    resp.URL,
		collection.ExtraNextAction:     collection.NextActionRedirectURL,
		collection.ExtraNextActionType: collection.NextActionRedirectURL,
		collection.ExtraProvider:       providerName,
		collection.ExtraMode:           resp.Mode,
		extraStripeStatus:              resp.Status,
	}
	if resp.ExpiresAt > 0 {
		statusExtras[collection.ExtraPaymentExpiresAt] = time.Unix(resp.ExpiresAt, 0).UTC().Format(time.RFC3339)
	}
	if priceID != "" {
		statusExtras[collection.ExtraProductID] = priceID
	}
	err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, resp.ID, commonv1.STATUS_IN_PROCESS, statusExtras)
	if err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

// createPaymentIntent serves callers that confirm on their own page with
// Stripe.js: no hosted page, the client_secret is the next step.
func (h *promptHandler) createPaymentIntent(
	ctx context.Context,
	logger *util.LogEntry,
	creds *client.StripeCredentials,
	prompt *paymentv1.InitiatePromptRequest,
	metadata map[string]string,
) error {
	promptID := prompt.GetId()
	amount, err := collection.MinorUnits(prompt.GetAmount())
	if err != nil {
		logger.WithError(err).Error("prompt amount cannot be charged")
		h.metrics.QueueFailed(ctx, "prompt", "rejected")
		return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, err.Error(), nil)
	}
	resp, err := h.stripeCli.CreatePaymentIntent(ctx, creds, &client.PaymentIntentRequest{
		IdempotencyKey: "pi:" + promptID,
		Amount:         amount,
		Currency:       prompt.GetAmount().GetCurrencyCode(),
		Description:    "Payment prompt " + promptID,
		Metadata:       metadata,
	})
	if err != nil {
		if collection.IsUncertain(err) {
			logger.WithError(err).Warn("create payment intent outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, "prompt", "provider_error")
			return err
		}
		logger.WithError(err).Error("create payment intent refused")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, err.Error(), nil)
	}

	logger.WithField("payment_intent_id", resp.ID).Debug("payment intent created")
	err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, resp.ID, commonv1.STATUS_IN_PROCESS, map[string]any{
		ExtraPaymentIntentID:         resp.ID,
		collection.ExtraClientSecret: resp.ClientSecret,
		collection.ExtraProvider:     providerName,
		extraStripeStatus:            resp.Status,
	})
	if err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

// openedReference returns the checkout session or payment intent id already
// recorded on the prompt's status, or "" when none (or when no payment
// client is configured).
func (h *promptHandler) openedReference(ctx context.Context, promptID string) string {
	if h.paymentCli == nil || promptID == "" {
		return ""
	}
	statusExtras, _ := structpb.NewStruct(map[string]any{events.ExtraEntityType: events.EntityTypePrompt})
	resp, err := h.paymentCli.Status(ctx, &commonv1.StatusRequest{Id: promptID, Extras: statusExtras})
	if err != nil {
		return ""
	}
	fields := resp.GetExtras().GetFields()
	for _, key := range []string{ExtraCheckoutSessionID, ExtraPaymentIntentID} {
		if f, ok := fields[key]; ok && f.GetStringValue() != "" {
			return f.GetStringValue()
		}
	}
	return ""
}
