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
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	credentialResolver
	statusEmitter
	fwCli   client.FlutterwaveClient
	cfg     *config.FlutterwaveConfig
	metrics *integrationobs.Metrics
}

// NewPaymentHandler handles outbound Payment Send via Flutterwave v4 transfers.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	fwCli client.FlutterwaveClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.FlutterwaveConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		fwCli:              fwCli,
		cfg:                cfg,
		metrics:            integrationobs.NewMetrics("flutterwave"),
	}
}

// fail reports a definite payout failure.
func (h *paymentHandler) fail(ctx context.Context, paymentID, reason string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, "payment", reason)
	if extras == nil {
		extras = map[string]any{}
	}
	extras["entity_type"] = "payment"
	extras["provider"] = "flutterwave"
	extras["error_code"] = reason
	return h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, extras)
}

// Handle pays the payee's verified mobile money number (the recipient
// contact the payment service resolved). Caller extras never choose the
// destination. Every write carries an idempotency key derived from the
// payment id, so a redelivered payout is performed once; a payout whose
// outcome is unknown is redelivered, never reported as failed.
//
//nolint:funlen // one payout flow: validate, recipient, transfer, report.
func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "flutterwave.payment")
	defer logger.Release()

	payment := &paymentv1.Payment{}
	if err := proto.Unmarshal(payload, payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil
	}
	paymentID := payment.GetId()
	if paymentID == "" {
		logger.Error("payment without id dropped")
		h.metrics.QueueFailed(ctx, "payment", "validation_error")
		return nil
	}
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		if errors.Is(err, errCredentialsUnavailable) {
			logger.WithError(err).Warn("credentials unavailable; redelivering")
			return err
		}
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, paymentID, "credentials_error", map[string]any{"error": err.Error()})
	}

	amount, currency, err := providerAmount(payment.GetAmount())
	if err != nil {
		return h.fail(ctx, paymentID, "invalid_amount", map[string]any{"error": err.Error()})
	}
	phone := digitsOnly(payment.GetRecipient().GetContactId())
	corridor := resolveMoMoCorridor(phone, currency)
	if len(phone) < 10 || corridor == nil { //nolint:mnd // shortest international MSISDN
		return h.fail(ctx, paymentID, "invalid_recipient", map[string]any{
			"error": "the payee's contact is not a mobile money number that receives " + currency,
		})
	}
	beneficiary := firstNonEmptyStr(
		extraString(payment.GetExtra(), "beneficiary_name"),
		extraString(payment.GetExtra(), "recipient_name"),
		"Beneficiary",
	)
	reference := chargeReference("pay", paymentID)
	narration := firstNonEmptyStr(extraString(payment.GetExtra(), "narration"), "Payment disbursement")
	tenantID, partitionID := promptTenancy(ctx, payment.GetExtra(), headers)
	meta := map[string]string{"payment_id": paymentID, "entity": "payment"}
	if tenantID != "" {
		meta[MetaTenantID] = tenantID
	}
	if partitionID != "" {
		meta[MetaPartitionID] = partitionID
	}
	callbackURL := ""
	if h.cfg.PublicWebhookBase != "" {
		callbackURL = strings.TrimRight(h.cfg.PublicWebhookBase, "/") + "/webhook/flutterwave"
	}

	var transferReq map[string]any
	recipientID := ""
	apiVersion := "v4"
	if client.IsV3Credentials(creds) {
		apiVersion = "v3"
		// v3 mobile money transfers are wired for M-Pesa Kenya only.
		if corridor.Currency != "KES" || corridor.Network != "MPESA" {
			return h.fail(ctx, paymentID, "unsupported_corridor", map[string]any{
				"error": "v3 keys pay out to M-Pesa Kenya only; use v4 OAuth credentials",
			})
		}
		transferReq = map[string]any{
			"amount":           amount,
			"currency":         currency,
			"reference":        reference,
			"narration":        narration,
			"meta":             meta,
			"account_bank":     "MPS",
			"account_number":   phone,
			"beneficiary_name": beneficiary,
		}
		if callbackURL != "" {
			transferReq["callback_url"] = callbackURL
		}
	} else {
		recipientID, err = h.fwCli.CreateTransferRecipient(
			client.WithIdempotencyKey(ctx, reference, "recipient"), creds, &client.TransferRecipientRequest{
				Type: "mobile_money",
				MobileMoney: &client.MobileMoneyRecipientDetails{
					Network: corridor.Network,
					Country: isoFromDial(corridor.CountryCode),
					MSISDN:  phone,
				},
				Name: splitName(beneficiary),
			})
		if err != nil {
			if client.IsDuplicate(err) {
				// A recipient this payment already made: its transfer may have
				// gone out too (a redelivery after the idempotency window).
				// Never fail it; leave it for the transfer webhook or an operator.
				logger.WithError(err).Error("transfer recipient already exists; payout left in process for review")
				return h.emitStatus(ctx, paymentID, "", commonv1.STATUS_IN_PROCESS, map[string]any{
					"entity_type": "payment", "provider": "flutterwave", "reference": reference,
					"transfer_status": "RECIPIENT_EXISTS", "needs_review": true,
				})
			}
			if client.IsRefusal(err) {
				logger.WithError(err).Error("transfer recipient refused")
				return h.fail(ctx, paymentID, "provider_rejected", map[string]any{"error": err.Error()})
			}
			// Nothing has been paid yet: try again later.
			logger.WithError(err).Warn("transfer recipient not created; redelivering")
			return fmt.Errorf("create recipient for %s: %w", paymentID, err)
		}
		transferReq = map[string]any{
			"action":    "instant",
			"reference": reference,
			"narration": narration,
			"meta":      meta,
			"payment_instruction": map[string]any{
				"source_currency":      currency,
				"destination_currency": currency,
				"amount": map[string]any{
					"applies_to": "destination_currency",
					"value":      amount,
				},
				"recipient_id": recipientID,
			},
		}
		if callbackURL != "" {
			transferReq["callback_url"] = callbackURL
		}
	}

	extras := map[string]any{
		"entity_type": "payment",
		"provider":    "flutterwave",
		"api_version": apiVersion,
		"reference":   reference,
	}
	if recipientID != "" {
		extras["recipient_id"] = recipientID
	}

	tr, err := h.fwCli.CreateTransfer(client.WithIdempotencyKey(ctx, reference, "transfer"), creds, transferReq)
	if err != nil {
		switch {
		case client.IsDuplicate(err):
			// Flutterwave already holds a transfer with this reference: it is
			// in flight or done. Its webhook settles it; never send again.
			logger.WithError(err).Warn("transfer already exists for this payment")
			extras["transfer_status"] = "DUPLICATE_REFERENCE"
			extras["needs_review"] = true
			return h.emitStatus(ctx, paymentID, "", commonv1.STATUS_IN_PROCESS, extras)
		case client.IsRejected(err):
			logger.WithError(err).Error("transfer refused by Flutterwave")
			extras["error"] = err.Error()
			return h.fail(ctx, paymentID, "provider_rejected", extras)
		default:
			// The transfer may have been made. The redelivered message sends
			// the same idempotency key and gets Flutterwave's original answer.
			logger.WithError(err).Warn("transfer outcome unknown; redelivering")
			return fmt.Errorf("transfer %s outcome unknown: %w", paymentID, err)
		}
	}

	status := TransferStatus(tr.Status)
	extras["transfer_id"] = tr.ID
	extras["transfer_status"] = tr.Status
	if status == commonv1.STATUS_FAILED {
		extras["error"] = "transfer " + strings.ToLower(tr.Status)
		h.metrics.QueueFailed(ctx, "payment", "transfer_failed")
	} else {
		h.metrics.QueueProcessed(ctx, "payment")
	}
	logger.WithField("transfer_id", tr.ID).WithField("status", tr.Status).Debug("transfer initiated")
	return h.emitStatus(ctx, paymentID, tr.ID, status, extras)
}

// transferStatus maps a Flutterwave transfer status.
func TransferStatus(s string) commonv1.STATUS {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "SUCCESSFUL", "SUCCESS", "SUCCEEDED":
		return commonv1.STATUS_SUCCESSFUL
	case "FAILED", "CANCELLED", "CANCELED":
		return commonv1.STATUS_FAILED
	}
	return commonv1.STATUS_IN_PROCESS
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func isoFromDial(dial string) string {
	switch dial {
	case "254":
		return "KE"
	case "233":
		return "GH"
	case "256":
		return "UG"
	case "255":
		return "TZ"
	case "250":
		return "RW"
	case "234":
		return "NG"
	case "260":
		return "ZM"
	default:
		return ""
	}
}
