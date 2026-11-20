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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Extras returned on the prompt status once a checkout exists.
const (
	ExtraCheckoutID = "checkout_id"
	extraPolarEvent = "polar_event_type"
)

type promptHandler struct {
	credentialResolver
	emitter    *events.StatusEmitter
	polarCli   client.PolarClient
	paymentCli paymentv1connect.PaymentServiceClient
	metrics    *integrationobs.Metrics
}

// NewPromptHandler creates the worker that turns a prompt into a Polar
// checkout session. paymentCli (optional) lets a redelivered prompt find the
// checkout it already opened instead of opening a second one.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	polarCli client.PolarClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	paymentCli paymentv1connect.PaymentServiceClient,
	cfg *config.PolarConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		polarCli:           polarCli,
		paymentCli:         paymentCli,
		metrics:            integrationobs.NewMetrics("polar"),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "polar.prompt")
	defer logger.Release()

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil
	}
	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	if existing := h.openedCheckoutID(ctx, promptID); existing != "" {
		logger.WithField("checkout_id", existing).Info("checkout already opened for prompt, skipping")
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
	productID := extras[collection.ExtraProductID]
	if productID == "" {
		logger.Error("prompt carries no product_id; Polar charges catalog products only")
		h.metrics.QueueFailed(ctx, "prompt", "rejected")
		return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID,
			"Polar needs product_id: the plan's Polar product", nil)
	}

	metadata := map[string]string{"prompt_id": promptID}
	for _, key := range []string{
		collection.ExtraSessionRef, collection.ExtraSubscriptionID, collection.ExtraInvoiceID,
		collection.ExtraOrderRef, collection.ExtraCustomerExternalID,
	} {
		if v := extras[key]; v != "" {
			metadata[key] = v
		}
	}
	// Tenant and partition round-trip through Polar so webhooks can be scoped.
	for _, key := range []string{"tenant_id", "partition_id"} {
		if v := headers[key]; v != "" {
			metadata[key] = v
		}
	}

	req := &client.CheckoutRequest{
		ProductIDs:         []string{productID},
		CustomerEmail:      firstNonEmpty(extras[collection.ExtraCustomerEmail], extras[collection.ExtraEmail]),
		CustomerName:       firstNonEmpty(extras[collection.ExtraCustomerName], extras[collection.ExtraDisplayName]),
		ExternalCustomerID: firstNonEmpty(extras[collection.ExtraCustomerExternalID], prompt.GetSource().GetProfileId()),
		SuccessURL:         firstNonEmpty(extras[collection.ExtraSuccessURL], extras[collection.ExtraRedirectURL]),
		Metadata:           metadata,
	}

	resp, err := h.polarCli.CreateCheckout(ctx, creds, req)
	if err != nil {
		if collection.IsUncertain(err) {
			// Polar may have opened the checkout: redeliver, the status
			// lookup above then finds it on the next attempt.
			logger.WithError(err).Warn("create checkout outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, "prompt", "provider_error")
			return err
		}
		logger.WithError(err).Error("create checkout refused")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, err.Error(), nil)
	}

	logger.WithField("checkout_id", resp.ID).Debug("checkout session created")
	err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, resp.ID, commonv1.STATUS_IN_PROCESS, map[string]any{
		ExtraCheckoutID:                  resp.ID,
		collection.ExtraCheckoutURL:      resp.URL,
		collection.ExtraNextAction:       collection.NextActionRedirectURL,
		collection.ExtraNextActionType:   collection.NextActionRedirectURL,
		collection.ExtraProvider:         "polar",
		collection.ExtraPaymentExpiresAt: resp.ExpiresAt,
		collection.ExtraProductID:        productID,
		"polar_status":                   resp.Status,
	})
	if err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

// openedCheckoutID returns the checkout id already recorded on the prompt's
// status, or "" when none (or when no payment client is configured).
func (h *promptHandler) openedCheckoutID(ctx context.Context, promptID string) string {
	if h.paymentCli == nil || promptID == "" {
		return ""
	}
	statusExtras, _ := structpb.NewStruct(map[string]any{events.ExtraEntityType: events.EntityTypePrompt})
	resp, err := h.paymentCli.Status(ctx, &commonv1.StatusRequest{Id: promptID, Extras: statusExtras})
	if err != nil {
		return ""
	}
	if f, ok := resp.GetExtras().GetFields()[ExtraCheckoutID]; ok {
		return f.GetStringValue()
	}
	return ""
}

func extraStrings(s *structpb.Struct) map[string]string {
	out := map[string]string{}
	for k, v := range s.GetFields() {
		if str := strings.TrimSpace(v.GetStringValue()); str != "" {
			out[k] = str
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
