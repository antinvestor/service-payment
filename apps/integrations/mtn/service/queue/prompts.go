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
	"github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/google/uuid"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type promptHandler struct {
	eventsMan   frameEvents.Manager
	mtnCli      client.MtnClient
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.MtnConfig
	metrics     *integrationobs.Metrics
}

// NewPromptHandler creates a queue worker for handling requestToPay collection prompts.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	mtnCli client.MtnClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		eventsMan:   eventsMan,
		mtnCli:      mtnCli,
		settingsCli: settingsCli,
		cfg:         cfg,
		metrics:     integrationobs.NewMetrics("mtn"),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mtn.prompt")
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

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}

	amount := formatMoneyAmount(prompt.GetAmount())

	currency := creds.Currency
	if prompt.GetAmount() != nil && prompt.GetAmount().GetCurrencyCode() != "" {
		currency = prompt.GetAmount().GetCurrencyCode()
	}

	referenceID := uuid.NewString()

	req := &client.RequestToPayRequest{
		ReferenceID: referenceID,
		Amount:      amount,
		Currency:    currency,
		ExternalID:  promptID,
		Payer: client.Party{
			PartyIDType: "MSISDN",
			PartyID:     phoneNumber,
		},
		PayerMessage: "Payment request",
		PayeeNote:    promptID,
		CallbackURL:  appendTenantParams(creds.CallbackURL+"/webhook/mtn/collection", headers),
	}

	err = h.mtnCli.RequestToPay(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("requestToPay failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		emitStatus(ctx, h.eventsMan, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	logger.WithField("reference_id", referenceID).Debug("requestToPay initiated")

	h.metrics.QueueProcessed(ctx, "prompt")
	emitStatus(ctx, h.eventsMan, promptID, referenceID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"reference_id": referenceID,
		"entity_type":  "prompt",
	})

	return nil
}
