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
	"errors"
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

type paymentHandler struct {
	credentialResolver
	emitter   *events.StatusEmitter
	airtelCli client.AirtelClient
	metrics   *integrationobs.Metrics
}

// NewPaymentHandler creates the queue worker for disbursements. The payment
// id is Airtel's transaction id, so a redelivered payout finds the one
// already sent (DisbursementStatus) instead of paying twice.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	airtelCli client.AirtelClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.AirtelConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		airtelCli:          airtelCli,
		metrics:            integrationobs.NewMetrics(providerName),
	}
}

func (h *paymentHandler) fail(ctx context.Context, paymentID, reason, message string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, kindPayment, reason)
	return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID, message, extras)
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "airtel.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, kindPayment, "unmarshal_error")
		return nil
	}
	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, paymentID, "credentials_error", err.Error(), nil)
	}

	amount, err := airtelAmount(payment.GetAmount())
	if err != nil {
		logger.WithError(err).Error("payout amount cannot be sent to Airtel")
		return h.fail(ctx, paymentID, "rejected", "invalid amount: "+err.Error(), nil)
	}
	currency := requestCurrency(payment.GetAmount(), creds)
	extras := withRequestBinding(map[string]any{}, amount, currency, headers)

	// A redelivered payout answers with the disbursement already sent.
	if existing, getErr := h.airtelCli.DisbursementStatus(ctx, creds, paymentID); getErr == nil && existing.Data.Transaction.Status != "" {
		logger.Debug("disbursement already sent")
		return h.emitKnown(ctx, paymentID, existing, extras)
	} else if getErr != nil && !client.IsNotFound(getErr) {
		logger.WithError(getErr).Warn("disbursement lookup failed, sending")
	}

	// Airtel's v2 disbursement needs the wallet PIN encrypted with Airtel's
	// public key; without both there is nothing to send.
	encryptedPin, err := client.EncryptPin(creds.DisbursementPin, creds.PublicKey)
	if err != nil {
		logger.WithError(err).Error("cannot authorise Airtel disbursement")
		reason := "credentials_error"
		if !errors.Is(err, client.ErrPinNotConfigured) {
			reason = "rejected"
		}
		return h.fail(ctx, paymentID, reason, "Airtel disbursement not configured: "+err.Error(), nil)
	}

	req := &client.DisbursementRequest{
		Reference:    paymentID,
		PhoneNumber:  payment.GetRecipient().GetContactId(),
		Amount:       amount,
		Currency:     currency,
		CountryCode:  creds.CountryCode,
		EncryptedPin: encryptedPin,
	}

	resp, err := h.airtelCli.Disburse(ctx, creds, req)
	if err != nil {
		if client.IsRejected(err) {
			logger.WithError(err).Error("disbursement refused by Airtel")
			return h.fail(ctx, paymentID, "provider_error", err.Error(), extras)
		}
		// Unknown outcome (timeout, 5xx): Airtel may hold the disbursement.
		// Report what it holds, or redeliver until it can be found.
		existing, getErr := h.airtelCli.DisbursementStatus(ctx, creds, paymentID)
		if getErr != nil || existing.Data.Transaction.Status == "" {
			logger.WithError(err).Warn("disbursement outcome unknown; redelivering")
			h.metrics.QueueFailed(ctx, kindPayment, "provider_error")
			return fmt.Errorf("disbursement %s outcome unknown: %w", paymentID, err)
		}
		return h.emitKnown(ctx, paymentID, existing, extras)
	}

	externalID := resp.Data.Transaction.ID
	logger.WithField("transaction_id", externalID).Debug("disbursement initiated")
	extras[client.ExtraTransactionID] = externalID
	extras["reference_id"] = resp.Data.Transaction.ReferenceID
	extras["status_code"] = resp.Status.Code
	extras["message"] = resp.Status.Message
	if err = h.emitter.Emit(ctx, events.EntityTypePayment, paymentID, externalID, commonv1.STATUS_IN_PROCESS, extras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPayment)
	return nil
}

func (h *paymentHandler) emitKnown(ctx context.Context, paymentID string, st *client.StatusResponse, extras map[string]any) error {
	if mapAirtelStatus(st.Data.Transaction.Status) == commonv1.STATUS_FAILED {
		h.metrics.QueueFailed(ctx, kindPayment, "rejected")
	} else {
		h.metrics.QueueProcessed(ctx, kindPayment)
	}
	return emitKnown(ctx, h.emitter, events.EntityTypePayment, paymentID, st, extras)
}
