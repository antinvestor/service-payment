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

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
)

type subscriptionHandler struct {
	credentialResolver
	polarCli client.PolarClient
	metrics  *integrationobs.Metrics
}

// NewSubscriptionHandler creates the worker that executes billing's
// provider commands (collection.ProviderCommand) against Polar.
func NewSubscriptionHandler(
	polarCli client.PolarClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.PolarConfig,
) queue.SubscribeWorker {
	return &subscriptionHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		polarCli:           polarCli,
		metrics:            integrationobs.NewMetrics("polar"),
	}
}

func (h *subscriptionHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "polar.subscription")
	defer logger.Release()

	var cmd collection.ProviderCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		logger.WithError(err).Error("failed to unmarshal provider command")
		h.metrics.QueueFailed(ctx, "subscription", "unmarshal_error")
		return nil
	}
	logger = logger.WithField("command", cmd.Command).
		WithField("subscription_id", cmd.SubscriptionID).
		WithField("external_subscription_id", cmd.ExternalSubscriptionID)

	creds, err := h.resolve(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "subscription", "credentials_error")
		return nil
	}

	switch cmd.Command {
	case collection.CommandCancelSubscription:
		if cmd.ExternalSubscriptionID == "" {
			logger.Error("cancel command without external subscription id")
			h.metrics.QueueFailed(ctx, "subscription", "rejected")
			return nil
		}
		_, err = h.polarCli.CancelSubscription(ctx, creds, cmd.ExternalSubscriptionID, cmd.AtPeriodEnd)
		if err != nil {
			if collection.IsUncertain(err) {
				logger.WithError(err).Warn("cancel outcome unknown; redelivering")
				h.metrics.QueueFailed(ctx, "subscription", "provider_error")
				return err
			}
			// Already cancelled / unknown at Polar: the webhook mirror holds
			// the truth, nothing more to do here.
			logger.WithError(err).Warn("cancel refused by Polar")
			h.metrics.QueueFailed(ctx, "subscription", "rejected")
			return nil
		}
		logger.WithField("at_period_end", cmd.AtPeriodEnd).Info("subscription cancelled at Polar")
	default:
		logger.Warn("unknown provider command, skipping")
		h.metrics.QueueFailed(ctx, "subscription", "rejected")
		return nil
	}
	h.metrics.QueueProcessed(ctx, "subscription")
	return nil
}
