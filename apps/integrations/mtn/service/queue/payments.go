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
	"net/url"

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

const (
	providerName = "mtn"
	kindPayment  = "payment"
	kindPrompt   = "prompt"
)

type paymentHandler struct {
	emitter     *events.StatusEmitter
	mtnCli      client.MtnClient
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.MtnConfig
	metrics     *integrationobs.Metrics
}

// NewPaymentHandler creates the queue worker for disbursement transfers.
// The X-Reference-Id is derived from the payment id, so a redelivered
// payout finds the transfer already issued instead of paying twice.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	mtnCli client.MtnClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		emitter:     events.NewStatusEmitter(eventsMan),
		mtnCli:      mtnCli,
		settingsCli: settingsCli,
		cfg:         cfg,
		metrics:     integrationobs.NewMetrics(providerName),
	}
}

func (h *paymentHandler) fail(ctx context.Context, paymentID, reason, message string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, kindPayment, reason)
	return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID, message, extras)
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mtn.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, kindPayment, "unmarshal_error")
		return nil
	}
	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := extractCredentials(ctx, headers, h.settingsCli, h.cfg)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, paymentID, "credentials_error", err.Error(), nil)
	}

	amount, err := mtnAmount(payment.GetAmount())
	if err != nil {
		logger.WithError(err).Error("payout amount cannot be sent to MTN")
		return h.fail(ctx, paymentID, "rejected", "invalid amount: "+err.Error(), nil)
	}
	currency := requestCurrency(payment.GetAmount(), creds)
	referenceID := client.ReferenceID(events.EntityTypePayment, paymentID)
	extras := requestExtras(events.EntityTypePayment, referenceID, amount, currency, headers)
	logger = logger.WithField("reference_id", referenceID)

	// A redelivered payout answers with the transfer already issued.
	if existing, getErr := h.mtnCli.GetTransferStatus(ctx, creds, referenceID); getErr == nil {
		logger.Debug("transfer already issued")
		return h.emitTransfer(ctx, paymentID, existing, extras)
	} else if !client.IsNotFound(getErr) {
		logger.WithError(getErr).Warn("transfer lookup failed, issuing")
	}

	req := &client.TransferRequest{
		ReferenceID: referenceID,
		Amount:      amount,
		Currency:    currency,
		ExternalID:  paymentID,
		Payee: client.Party{
			PartyIDType: "MSISDN",
			PartyID:     payment.GetRecipient().GetContactId(),
		},
		PayerMessage: "Payment disbursement",
		PayeeNote:    paymentID,
		CallbackURL:  appendTenantParams(creds.CallbackURL+"/webhook/mtn/disbursement", headers),
	}

	if err = h.mtnCli.Transfer(ctx, creds, req); err != nil {
		if client.IsRejected(err) {
			logger.WithError(err).Error("transfer refused by MTN")
			return h.fail(ctx, paymentID, "provider_error", err.Error(), extras)
		}
		// Duplicate reference (409) or unknown outcome (timeout, 5xx): MTN
		// may hold the transfer. Report what MTN holds, or redeliver.
		existing, getErr := h.mtnCli.GetTransferStatus(ctx, creds, referenceID)
		if getErr != nil {
			logger.WithError(err).Warn("transfer outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, kindPayment, "provider_error")
			return fmt.Errorf("transfer %s outcome unknown: %w", paymentID, err)
		}
		return h.emitTransfer(ctx, paymentID, existing, extras)
	}

	logger.Debug("transfer initiated")
	if err = h.emitter.Emit(ctx, events.EntityTypePayment, paymentID, referenceID, commonv1.STATUS_IN_PROCESS, extras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPayment)
	return nil
}

// emitTransfer reports the state MTN holds for a transfer.
func (h *paymentHandler) emitTransfer(ctx context.Context, paymentID string, st *client.TransferStatus, extras map[string]any) error {
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
		reason := "transfer failed"
		if st.Reason != nil && st.Reason.Message != "" {
			reason = st.Reason.Message
		}
		return h.fail(ctx, paymentID, "rejected", reason, extras)
	}
	externalID := st.FinancialTransactionID
	if externalID == "" {
		externalID, _ = extras[client.ExtraReferenceID].(string)
	}
	if err := h.emitter.Emit(ctx, events.EntityTypePayment, paymentID, externalID, status, extras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPayment)
	return nil
}

// appendTenantParams appends tenant_id and partition_id query params to a callback URL.
func appendTenantParams(callbackURL string, headers map[string]string) string {
	tenantID := headers["tenant_id"]
	partitionID := headers["partition_id"]
	if tenantID == "" && partitionID == "" {
		return callbackURL
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return callbackURL
	}
	q := u.Query()
	if tenantID != "" {
		q.Set("tenant_id", tenantID)
	}
	if partitionID != "" {
		q.Set("partition_id", partitionID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}
