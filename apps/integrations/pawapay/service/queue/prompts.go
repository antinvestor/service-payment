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
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/config"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/credentials"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type promptHandler struct {
	statusEmitter
	credsResolver *credentials.Resolver
	pawapayCli    client.PawapayClient
	metrics       *integrationobs.Metrics
}

// NewPromptHandler creates a queue worker for handling deposit (collection)
// prompt requests. A pawaPay deposit triggers a PIN prompt on the customer's
// phone to authorise the payment, equivalent to an M-Pesa STK push.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	pawapayCli client.PawapayClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.PawapayConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		statusEmitter: statusEmitter{eventsMan: eventsMan},
		credsResolver: credentials.NewResolver(settingsCli, cfg),
		pawapayCli:    pawapayCli,
		metrics:       integrationobs.NewMetrics("pawapay"),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "pawapay.prompt")
	defer logger.Release()
	logger.Debug("queue handler started")

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil // non-retriable
	}

	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "prompt", "credentials_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}

	extraProvider := extraString(prompt.GetExtra(), "provider")

	provider, phoneNumber, err := resolveProvider(ctx, h.pawapayCli, creds, extraProvider, phoneNumber)
	if err != nil {
		logger.WithError(err).Error("failed to resolve provider")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	depositID := paymentUUID(promptID)

	depositReq := &client.DepositRequest{
		DepositID:            depositID,
		Amount:               formatMoneyAmount(prompt.GetAmount()),
		Currency:             prompt.GetAmount().GetCurrencyCode(),
		PhoneNumber:          phoneNumber,
		Provider:             provider,
		PreAuthorisationCode: extraString(prompt.GetExtra(), "pre_authorisation_code"),
		ClientReferenceID:    promptID,
		CustomerMessage:      sanitizeCustomerMessage(extraString(prompt.GetExtra(), "customer_message")),
		Metadata:             paymentMetadata("prompt", promptID, headers),
	}

	resp, err := h.pawapayCli.InitiateDeposit(ctx, creds, depositReq)
	if err != nil {
		logger.WithError(err).Error("deposit initiation failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	if resp.Status == client.InitiationStatusRejected {
		logger.WithField("failure_reason", resp.FailureReason).Error("deposit rejected")
		h.metrics.QueueFailed(ctx, "prompt", "rejected")
		h.emitStatus(ctx, promptID, depositID, commonv1.STATUS_FAILED, map[string]any{
			"deposit_id":      depositID,
			"failure_code":    failureCode(resp.FailureReason),
			"failure_message": failureMessage(resp.FailureReason),
			"entity_type":     "prompt",
		})
		return nil
	}

	logger.WithField("deposit_id", depositID).WithField("status", resp.Status).Debug("deposit initiated")

	h.metrics.QueueProcessed(ctx, "prompt")
	h.emitStatus(ctx, promptID, depositID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"deposit_id":  depositID,
		"provider":    provider,
		"status":      resp.Status,
		"created":     resp.Created,
		"entity_type": "prompt",
	})

	return nil
}
