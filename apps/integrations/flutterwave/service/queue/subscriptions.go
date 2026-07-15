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
	"encoding/json"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
)

// SubscriptionLifecycleEvent mirrors billing's lifecycle payload.
type SubscriptionLifecycleEvent struct {
	Event            string         `json:"event"`
	SubscriptionID   string         `json:"subscription_id"`
	ProfileID        string         `json:"profile_id"`
	PlanID           string         `json:"plan_id"`
	State            string         `json:"state"`
	Currency         string         `json:"currency"`
	ExternalEntityID string         `json:"external_entity_id,omitempty"`
	InvoiceID        string         `json:"invoice_id,omitempty"`
	Data             map[string]any `json:"data,omitempty"`
}

type subscriptionHandler struct {
	credentialResolver
	fwCli   client.FlutterwaveClient
	cfg     *config.FlutterwaveConfig
	metrics *integrationobs.Metrics
}

// NewSubscriptionHandler consumes billing lifecycle events.
//
// Flutterwave v4 does not expose v3-style payment-plans APIs. Subscriptions
// are owned by service_billing; charges complete via the prompt queue + webhooks.
// This worker logs lifecycle correlation and is the extension point for future
// COF/recurring charge orchestration.
func NewSubscriptionHandler(
	fwCli client.FlutterwaveClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.FlutterwaveConfig,
) queue.SubscribeWorker {
	return &subscriptionHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		fwCli:              fwCli,
		cfg:                cfg,
		metrics:            integrationobs.NewMetrics("flutterwave"),
	}
}

func (h *subscriptionHandler) Handle(ctx context.Context, _ map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "flutterwave.subscription")
	defer logger.Release()

	var evt SubscriptionLifecycleEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		logger.WithError(err).Error("failed to unmarshal lifecycle event")
		h.metrics.QueueFailed(ctx, "subscription", "unmarshal_error")
		return nil
	}
	logger = logger.WithFields(map[string]any{
		"event":           evt.Event,
		"subscription_id": evt.SubscriptionID,
	})

	switch evt.Event {
	case "subscription.created", "subscription.activated":
		logger.Info("subscription lifecycle — first collection uses prompt/orchestrator charge")
	case "subscription.billed":
		logger.Info("subscription billed — await charge.completed webhook for payment confirmation")
	case "subscription.cancelled":
		logger.Info("subscription cancelled in billing — Flutterwave v4 has no separate cancel-plan RPC")
	default:
		logger.Debug("ignoring lifecycle event")
	}
	h.metrics.QueueProcessed(ctx, "subscription")
	return nil
}
