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
	"strconv"
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

type promptHandler struct {
	credentialResolver
	statusEmitter
	fwCli   client.FlutterwaveClient
	cfg     *config.FlutterwaveConfig
	metrics *integrationobs.Metrics
}

// NewPromptHandler handles InitiatePrompt via Flutterwave v4 orchestrator charges.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	fwCli client.FlutterwaveClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.FlutterwaveConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		fwCli:              fwCli,
		cfg:                cfg,
		metrics:            integrationobs.NewMetrics("flutterwave"),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "flutterwave.prompt")
	defer logger.Release()

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil
	}

	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "prompt", "credentials_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": err.Error(), "entity_type": "prompt",
		})
		return nil
	}

	amountStr, currency := formatMoneyAmount(prompt.GetAmount())
	if currency == "" {
		currency = "KES"
	}
	amount, _ := strconv.ParseFloat(amountStr, 64)
	if amount <= 0 {
		h.metrics.QueueFailed(ctx, "prompt", "validation_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": "invalid amount", "entity_type": "prompt",
		})
		return nil
	}

	phone := prompt.GetRecipient().GetContactId()
	if phone == "" {
		phone = prompt.GetSource().GetContactId()
	}

	email := extraString(prompt.GetExtra(), "customer_email")
	if email == "" {
		email = extraString(prompt.GetExtra(), "email")
	}
	if email == "" {
		email = "payer+" + sanitizeRef(promptID) + "@payments.local"
	}
	name := extraString(prompt.GetExtra(), "customer_name")
	if name == "" {
		name = extraString(prompt.GetExtra(), "display_name")
	}

	redirectURL := extraString(prompt.GetExtra(), "success_url")
	if redirectURL == "" {
		redirectURL = extraString(prompt.GetExtra(), "redirect_url")
	}
	if redirectURL == "" {
		redirectURL = headerOrDefault(headers, config.HeaderRedirectURL, h.cfg.DefaultRedirectURL)
	}

	reference := extraString(prompt.GetExtra(), "tx_ref")
	if reference == "" {
		reference = "prompt-" + sanitizeRef(promptID)
	}
	// FW reference: 6–42 alphanumeric + hyphen
	if len(reference) > 42 {
		reference = reference[:42]
	}

	meta := map[string]string{
		"prompt_id": promptID,
		"entity":    "prompt",
	}
	if v := headers["tenant_id"]; v != "" {
		meta["tenant_id"] = v
	}
	if v := headers["partition_id"]; v != "" {
		meta["partition_id"] = v
	}
	for _, k := range []string{"session_ref", "invoice_id", "subscription_id"} {
		if v := extraString(prompt.GetExtra(), k); v != "" {
			meta[k] = v
		}
	}

	// Build payment method: prefer MoMo when phone is present.
	var pm client.PaymentMethodInput
	mode := "collection"
	if phone != "" {
		if corridor := resolveMoMoCorridor(phone, currency); corridor != nil {
			currency = corridor.Currency
			pm = buildMoMoPaymentMethod(corridor, extraString(prompt.GetExtra(), "network"))
			mode = "mobile_money"
		}
	}
	if pm.Type == "" {
		pm = h.defaultCollectionMethod(&prompt, currency)
		mode = pm.Type
	}

	customer := client.CustomerInput{
		Email: email,
		Name:  splitName(name),
		Meta:  meta,
	}
	if phone != "" {
		if corridor := resolveMoMoCorridor(phone, currency); corridor != nil {
			customer.Phone = &client.CustomerPhone{
				CountryCode: corridor.CountryCode,
				Number:      corridor.NationalNumber,
			}
		}
	}

	req := &client.OrchestratorChargeRequest{
		Amount:        amount,
		Currency:      currency,
		Reference:     reference,
		RedirectURL:   redirectURL,
		Customer:      customer,
		PaymentMethod: pm,
		Meta:          meta,
	}

	ch, err := h.fwCli.CreateOrchestratorCharge(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("orchestrator charge failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": err.Error(), "entity_type": "prompt", "reference": reference,
		})
		return nil
	}

	apiVersion := "v4"
	if client.IsV3Credentials(creds) {
		apiVersion = "v3"
	}
	extras := map[string]any{
		"entity_type":   "prompt",
		"provider":      "flutterwave",
		"api_version":   apiVersion,
		"mode":          mode,
		"reference":     ch.Reference,
		"charge_id":     ch.ID,
		"charge_status": ch.Status,
		"tx_ref":        ch.Reference,
	}
	if url := client.ExtractRedirectURL(ch); url != "" {
		// Checkout reads extras.checkout_url for browser redirect.
		extras["checkout_url"] = url
	}
	if note := client.ExtractPaymentInstructionNote(ch); note != "" {
		extras["payment_instruction"] = note
	}
	// Bank transfer details when present.
	if ch.NextAction != nil {
		if t, _ := ch.NextAction["type"].(string); t == "requires_bank_transfer" {
			if bt, ok := ch.NextAction["requires_bank_transfer"].(map[string]any); ok {
				extras["bank_transfer"] = bt
			}
		}
	}

	logger.WithField("charge_id", ch.ID).WithField("status", ch.Status).Debug("charge initiated")
	h.emitStatus(ctx, promptID, ch.ID, commonv1.STATUS_IN_PROCESS, extras)
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

func (h *promptHandler) defaultCollectionMethod(
	prompt *paymentv1.InitiatePromptRequest,
	_ string,
) client.PaymentMethodInput {
	methodType := extraString(prompt.GetExtra(), "payment_method_type")
	if methodType == "" {
		methodType = h.cfg.DefaultCollectionMethod
	}
	if methodType == "" {
		// Hosted multipayment page (Standard) — not bank_transfer.
		// v4 orchestrator rejects bank_transfer on many accounts.
		methodType = "card"
	}
	switch strings.ToLower(methodType) {
	case "opay":
		return client.PaymentMethodInput{Type: "opay"}
	case "ussd":
		bank := extraString(prompt.GetExtra(), "account_bank")
		if bank == "" {
			bank = "044" // Access Bank default for sandbox demos
		}
		return client.PaymentMethodInput{
			Type: "ussd",
			USSD: &client.USSDDetails{AccountBank: bank},
		}
	case "bank_transfer", "banktransfer":
		// Explicit bank transfer only — not the SPA default.
		display := extraString(prompt.GetExtra(), "account_display_name")
		if display == "" {
			display = "Payment"
		}
		return client.PaymentMethodInput{
			Type: "bank_transfer",
			BankTransfer: &client.BankTransferDetails{
				AccountType:        "dynamic",
				AccountExpiresIn:   3600,
				AccountDisplayName: display,
			},
		}
	case "card", "hosted", "standard", "payment_link":
		// Marker for hosted Standard multipayment page (redirect to Flutterwave).
		return client.PaymentMethodInput{Type: "card"}
	default:
		// Unknown → hosted card page, never bank_transfer (v4 400).
		return client.PaymentMethodInput{Type: "card"}
	}
}

func sanitizeRef(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) < 6 {
		out = out + "xxxxxx"
	}
	return out
}
