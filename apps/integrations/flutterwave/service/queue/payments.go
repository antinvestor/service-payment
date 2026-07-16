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
	"net/url"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
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

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "flutterwave.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil
	}

	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "payment", "credentials_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": err.Error(), "entity_type": "payment",
		})
		return nil
	}

	amountWhole := moneyToWholeUnits(payment.GetAmount())
	currency := ""
	if payment.GetAmount() != nil {
		currency = strings.ToUpper(payment.GetAmount().GetCurrencyCode())
	}
	if currency == "" {
		currency = "KES"
	}
	if amountWhole <= 0 {
		h.metrics.QueueFailed(ctx, "payment", "validation_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": "invalid amount", "entity_type": "payment",
		})
		return nil
	}

	accountNumber := payment.GetRecipient().GetContactId()
	if acc := extraString(payment.GetExtra(), "account_number"); acc != "" {
		accountNumber = acc
	}
	bankCode := extraString(payment.GetExtra(), "account_bank")
	if bankCode == "" {
		bankCode = extraString(payment.GetExtra(), "bank_code")
	}
	beneficiaryName := extraString(payment.GetExtra(), "beneficiary_name")
	if beneficiaryName == "" {
		beneficiaryName = extraString(payment.GetExtra(), "recipient_name")
	}

	// MoMo payout vs bank.
	payoutType := extraString(payment.GetExtra(), "payout_type")
	if payoutType == "" {
		if bankCode == "" && accountNumber != "" {
			payoutType = "mobile_money"
		} else {
			payoutType = "bank"
		}
	}

	reference := extraString(payment.GetExtra(), "reference")
	if reference == "" {
		reference = "pay-" + sanitizeRef(paymentID)
	}
	if len(reference) > 42 {
		reference = reference[:42]
	}

	callbackURL := ""
	if h.cfg.PublicWebhookBase != "" {
		callbackURL = strings.TrimRight(h.cfg.PublicWebhookBase, "/") + "/webhook/flutterwave"
		callbackURL = appendTenantParams(callbackURL, headers)
	}

	meta := map[string]string{
		"payment_id": paymentID,
		"entity":     "payment",
	}
	if v := headers["tenant_id"]; v != "" {
		meta["tenant_id"] = v
	}
	if v := headers["partition_id"]; v != "" {
		meta["partition_id"] = v
	}

	var transferReq map[string]any
	var recipientID string
	apiVersion := "v4"
	if client.IsV3Credentials(creds) {
		apiVersion = "v3"
		v3 := map[string]any{
			"amount":    amountWhole,
			"currency":  currency,
			"reference": reference,
			"narration": firstNonEmptyStr(extraString(payment.GetExtra(), "narration"), "Payment disbursement"),
			"meta":      meta,
		}
		if payoutType == "mobile_money" {
			v3["account_bank"] = firstNonEmptyStr(bankCode, "MPS")
			v3["account_number"] = digitsOnly(accountNumber)
			v3["beneficiary_name"] = firstNonEmptyStr(beneficiaryName, "Beneficiary")
		} else {
			if bankCode == "" || accountNumber == "" {
				h.metrics.QueueFailed(ctx, "payment", "validation_error")
				h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
					"error": "account_bank and account_number required", "entity_type": "payment",
				})
				return nil
			}
			v3["account_bank"] = bankCode
			v3["account_number"] = accountNumber
			if beneficiaryName != "" {
				v3["beneficiary_name"] = beneficiaryName
			}
		}
		if callbackURL != "" {
			v3["callback_url"] = callbackURL
		}
		transferReq = v3
	} else {
		// v4: create recipient then transfer.
		if payoutType == "mobile_money" {
			corridor := resolveMoMoCorridor(accountNumber, currency)
			if corridor == nil {
				h.metrics.QueueFailed(ctx, "payment", "validation_error")
				h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
					"error": "unsupported mobile money corridor", "entity_type": "payment",
				})
				return nil
			}
			net := corridor.Network
			if o := extraString(payment.GetExtra(), "network"); o != "" {
				net = strings.ToUpper(o)
			}
			recipientID, err = h.fwCli.CreateTransferRecipient(ctx, creds, &client.TransferRecipientRequest{
				Type: "mobile_money",
				MobileMoney: &client.MobileMoneyRecipientDetails{
					Network: net,
					Country: isoFromDial(corridor.CountryCode),
					MSISDN:  digitsOnly(accountNumber),
				},
				Name: splitName(beneficiaryName),
			})
		} else {
			if bankCode == "" || accountNumber == "" {
				h.metrics.QueueFailed(ctx, "payment", "validation_error")
				h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
					"error": "account_bank and account_number required", "entity_type": "payment",
				})
				return nil
			}
			recipientID, err = h.fwCli.CreateTransferRecipient(ctx, creds, &client.TransferRecipientRequest{
				Type: resolveRecipientType(currency),
				Bank: &client.BankRecipientDetails{
					AccountNumber: accountNumber,
					Code:          bankCode,
				},
				Name: splitName(beneficiaryName),
			})
		}
		if err != nil {
			logger.WithError(err).Error("create transfer recipient failed")
			h.metrics.QueueFailed(ctx, "payment", "provider_error")
			h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
				"error": err.Error(), "entity_type": "payment",
			})
			return nil
		}
		transferReq = map[string]any{
			"action":    "instant",
			"reference": reference,
			"narration": firstNonEmptyStr(extraString(payment.GetExtra(), "narration"), "Payment disbursement"),
			"meta":      meta,
			"payment_instruction": map[string]any{
				"source_currency":      currency,
				"destination_currency": currency,
				"amount": map[string]any{
					"applies_to": "destination_currency",
					"value":      float64(amountWhole),
				},
				"recipient_id": recipientID,
			},
		}
		if callbackURL != "" {
			transferReq["callback_url"] = callbackURL
		}
	}

	tr, err := h.fwCli.CreateTransfer(ctx, creds, transferReq)
	if err != nil {
		logger.WithError(err).Error("create transfer failed")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": err.Error(), "entity_type": "payment", "reference": reference,
		})
		return nil
	}

	logger.WithField("transfer_id", tr.ID).WithField("status", tr.Status).Debug("transfer initiated")
	extras := map[string]any{
		"entity_type":     "payment",
		"provider":        "flutterwave",
		"api_version":     apiVersion,
		"transfer_id":     tr.ID,
		"transfer_status": tr.Status,
		"reference":       tr.Reference,
	}
	if recipientID != "" {
		extras["recipient_id"] = recipientID
	}
	h.emitStatus(ctx, paymentID, tr.ID, commonv1.STATUS_IN_PROCESS, extras)
	h.metrics.QueueProcessed(ctx, "payment")
	return nil
}

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
