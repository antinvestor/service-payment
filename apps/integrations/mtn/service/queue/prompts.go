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
	"github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type promptHandler struct {
	emitter     *events.StatusEmitter
	mtnCli      client.MtnClient
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.MtnConfig
	metrics     *integrationobs.Metrics
}

// NewPromptHandler creates the queue worker for requestToPay collections.
// The X-Reference-Id is derived from the prompt id, so a redelivered prompt
// finds the request already issued instead of charging twice.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	mtnCli client.MtnClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		emitter:     events.NewStatusEmitter(eventsMan),
		mtnCli:      mtnCli,
		settingsCli: settingsCli,
		cfg:         cfg,
		metrics:     integrationobs.NewMetrics(providerName),
	}
}

func (h *promptHandler) fail(ctx context.Context, promptID, reason, message string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, kindPrompt, reason)
	return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, message, extras)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mtn.prompt")
	defer logger.Release()

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, kindPrompt, "unmarshal_error")
		return nil
	}
	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := extractCredentials(ctx, headers, h.settingsCli, h.cfg)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, promptID, "credentials_error", err.Error(), nil)
	}

	amount, err := mtnAmount(prompt.GetAmount())
	if err != nil {
		logger.WithError(err).Error("prompt amount cannot be sent to MTN")
		return h.fail(ctx, promptID, "rejected", "invalid amount: "+err.Error(), nil)
	}
	currency := requestCurrency(prompt.GetAmount(), creds)
	referenceID := client.ReferenceID(events.EntityTypePrompt, promptID)
	extras := requestExtras(events.EntityTypePrompt, referenceID, amount, currency, headers)
	logger = logger.WithField("reference_id", referenceID)

	// A redelivered prompt answers with the request already issued.
	if existing, getErr := h.mtnCli.GetRequestToPayStatus(ctx, creds, referenceID); getErr == nil {
		logger.Debug("requestToPay already issued")
		return h.emitRequestToPay(ctx, promptID, existing, extras)
	} else if !client.IsNotFound(getErr) {
		logger.WithError(getErr).Warn("requestToPay lookup failed, issuing")
	}

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}
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

	if err = h.mtnCli.RequestToPay(ctx, creds, req); err != nil {
		if client.IsRejected(err) {
			logger.WithError(err).Error("requestToPay refused by MTN")
			return h.fail(ctx, promptID, "provider_error", err.Error(), extras)
		}
		// Duplicate reference (409) or unknown outcome (timeout, 5xx): MTN
		// may hold the request. Report what MTN holds, or redeliver.
		existing, getErr := h.mtnCli.GetRequestToPayStatus(ctx, creds, referenceID)
		if getErr != nil {
			logger.WithError(err).Warn("requestToPay outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, kindPrompt, "provider_error")
			return fmt.Errorf("requestToPay %s outcome unknown: %w", promptID, err)
		}
		return h.emitRequestToPay(ctx, promptID, existing, extras)
	}

	logger.Debug("requestToPay initiated")
	if err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, referenceID, commonv1.STATUS_IN_PROCESS, extras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPrompt)
	return nil
}

// emitRequestToPay reports the state MTN holds for a collection request.
func (h *promptHandler) emitRequestToPay(ctx context.Context, promptID string, st *client.RequestToPayStatus, extras map[string]any) error {
	status := mapMtnStatus(st.Status)
	extras["mtn_status"] = st.Status
	if st.FinancialTransactionID != "" {
		extras["financial_transaction_id"] = st.FinancialTransactionID
	}
	if st.Reason != nil {
		extras["reason_code"] = st.Reason.Code
		extras["reason_message"] = st.Reason.Message
	}
	if status == commonv1.STATUS_FAILED {
		reason := "requestToPay failed"
		if st.Reason != nil && st.Reason.Message != "" {
			reason = st.Reason.Message
		}
		return h.fail(ctx, promptID, "rejected", reason, extras)
	}
	externalID := st.FinancialTransactionID
	if externalID == "" {
		externalID, _ = extras[client.ExtraReferenceID].(string)
	}
	if err := h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, externalID, status, extras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPrompt)
	return nil
}
