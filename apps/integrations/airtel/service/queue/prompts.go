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
	"fmt"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/config"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type promptHandler struct {
	credentialResolver
	emitter   *events.StatusEmitter
	airtelCli client.AirtelClient
	metrics   *integrationobs.Metrics
}

// NewPromptHandler creates the queue worker for USSD push collections. The
// prompt id is Airtel's transaction id, so a redelivered prompt finds the
// push already sent (TransactionStatus) instead of prompting twice.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	airtelCli client.AirtelClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.AirtelConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		airtelCli:          airtelCli,
		metrics:            integrationobs.NewMetrics(providerName),
	}
}

func (h *promptHandler) fail(ctx context.Context, promptID, reason, message string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, kindPrompt, reason)
	return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, message, extras)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "airtel.prompt")
	defer logger.Release()

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, kindPrompt, "unmarshal_error")
		return nil
	}
	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, promptID, "credentials_error", err.Error(), nil)
	}

	amount, err := airtelAmount(prompt.GetAmount())
	if err != nil {
		logger.WithError(err).Error("prompt amount cannot be sent to Airtel")
		return h.fail(ctx, promptID, "rejected", "invalid amount: "+err.Error(), nil)
	}
	currency := requestCurrency(prompt.GetAmount(), creds)
	extras := withRequestBinding(map[string]any{}, amount, currency, headers)

	// A redelivered prompt answers with the push already sent.
	if existing, getErr := h.airtelCli.TransactionStatus(ctx, creds, promptID); getErr == nil && existing.Data.Transaction.Status != "" {
		logger.Debug("collection push already sent")
		return h.emitKnown(ctx, promptID, existing, extras)
	} else if getErr != nil && !client.IsNotFound(getErr) {
		logger.WithError(getErr).Warn("collection lookup failed, sending")
	}

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}
	req := &client.CollectionRequest{
		Reference:   promptID,
		PhoneNumber: phoneNumber,
		Amount:      amount,
		Currency:    currency,
		CountryCode: creds.CountryCode,
	}

	resp, err := h.airtelCli.CollectionPush(ctx, creds, req)
	if err != nil {
		if client.IsRejected(err) {
			logger.WithError(err).Error("collection push refused by Airtel")
			return h.fail(ctx, promptID, "provider_error", err.Error(), extras)
		}
		// Unknown outcome (timeout, 5xx): Airtel may have pushed the prompt.
		// Report what it holds, or redeliver until it can be found.
		existing, getErr := h.airtelCli.TransactionStatus(ctx, creds, promptID)
		if getErr != nil || existing.Data.Transaction.Status == "" {
			logger.WithError(err).Warn("collection push outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, kindPrompt, "provider_error")
			return fmt.Errorf("collection push %s outcome unknown: %w", promptID, err)
		}
		return h.emitKnown(ctx, promptID, existing, extras)
	}

	externalID := resp.Data.Transaction.ID
	logger.WithField("transaction_id", externalID).Debug("collection push initiated")
	extras[client.ExtraTransactionID] = externalID
	extras["status_code"] = resp.Status.Code
	extras["message"] = resp.Status.Message
	if err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, externalID, commonv1.STATUS_IN_PROCESS, extras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPrompt)
	return nil
}

func (h *promptHandler) emitKnown(ctx context.Context, promptID string, st *client.StatusResponse, extras map[string]any) error {
	if mapAirtelStatus(st.Data.Transaction.Status) == commonv1.STATUS_FAILED {
		h.metrics.QueueFailed(ctx, kindPrompt, "rejected")
	} else {
		h.metrics.QueueProcessed(ctx, kindPrompt)
	}
	return emitKnown(ctx, h.emitter, events.EntityTypePrompt, promptID, st, extras)
}
