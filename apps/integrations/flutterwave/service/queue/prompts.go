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
	"google.golang.org/protobuf/types/known/structpb"
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

	// Authorization step for an existing charge (PIN / OTP / AVS) — no new charge.
	if action := strings.ToLower(extraString(prompt.GetExtra(), "action")); action == "authorize" {
		return h.handleAuthorize(ctx, logger, &prompt, creds, promptID)
	}

	// Tokenized / subscription renewal: charge saved payment_method_id.
	if pmd := extraString(prompt.GetExtra(), "payment_method_id"); pmd != "" {
		if cus := extraString(prompt.GetExtra(), "customer_id"); cus != "" {
			return h.handleTokenCharge(ctx, logger, &prompt, creds, promptID, cus, pmd, headers)
		}
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

	// Build payment method: prefer encrypted card (embedded) → MoMo → default.
	var pm client.PaymentMethodInput
	mode := "collection"
	if card := cardFromExtras(prompt.GetExtra()); card != nil {
		pm = client.PaymentMethodInput{Type: "card", Card: card}
		mode = "card"
	} else if phone != "" {
		if corridor := resolveMoMoCorridor(phone, currency); corridor != nil {
			// Only prefer MoMo when the caller did not force card.
			forced := strings.ToLower(extraString(prompt.GetExtra(), "payment_method_type"))
			if forced == "" || forced == "mobile_money" || forced == "momo" {
				currency = corridor.Currency
				pm = buildMoMoPaymentMethod(corridor, extraString(prompt.GetExtra(), "network"))
				mode = "mobile_money"
			}
		}
	}
	if pm.Type == "" {
		pm = h.defaultCollectionMethod(&prompt, currency)
		mode = pm.Type
	}
	// Card without encrypted fields cannot charge on pure v4 OAuth — fail early
	// with an actionable error (do not ask operators for FLWSECK multipay).
	if strings.EqualFold(pm.Type, "card") && (pm.Card == nil || pm.Card.EncryptedCardNumber == "") {
		// Allow legacy multipay only when FLWSECK is actually present.
		if !client.IsV3Credentials(creds) && !hasStandardSecretCreds(creds) {
			forced := strings.ToLower(extraString(prompt.GetExtra(), "payment_method_type"))
			if forced == "hosted" || forced == "standard" || forced == "payment_link" {
				// explicit multipay request without secret → clear error
			}
			h.metrics.QueueFailed(ctx, "prompt", "card_encryption_required")
			h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
				"error": "card payment requires encrypted card fields from hosted checkout " +
					"(open pay.stawi.org session and submit the card form; " +
					"set CHECKOUT_CARD_ENCRYPTION_KEY / FLUTTERWAVE_ENCRYPTION_KEY). " +
					"v4 OAuth does not use FLWSECK Standard multipay",
				"entity_type": "prompt",
				"mode":        "card",
				"hint":        "use_embedded_checkout",
			})
			return nil
		}
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

	h.emitChargeStatus(ctx, logger, promptID, ch, mode, reference)
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

// handleTokenCharge charges a previously saved payment method (subscription renewals).
func (h *promptHandler) handleTokenCharge(
	ctx context.Context,
	logger *util.LogEntry,
	prompt *paymentv1.InitiatePromptRequest,
	creds *client.Credentials,
	promptID, customerID, paymentMethodID string,
	headers map[string]string,
) error {
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
	reference := extraString(prompt.GetExtra(), "tx_ref")
	if reference == "" {
		reference = "renew-" + sanitizeRef(promptID)
	}
	if len(reference) > 42 {
		reference = reference[:42]
	}
	redirectURL := extraString(prompt.GetExtra(), "success_url")
	if redirectURL == "" {
		redirectURL = extraString(prompt.GetExtra(), "redirect_url")
	}
	meta := map[string]string{"prompt_id": promptID, "entity": "prompt", "mode": "recurring"}
	for _, k := range []string{"session_ref", "invoice_id", "subscription_id"} {
		if v := extraString(prompt.GetExtra(), k); v != "" {
			meta[k] = v
		}
	}
	if v := headers["tenant_id"]; v != "" {
		meta["tenant_id"] = v
	}

	// Recurring defaults true for token charges (subscription renewals).
	recurring := !strings.EqualFold(extraString(prompt.GetExtra(), "recurring"), "false")

	ch, err := h.fwCli.CreateCharge(ctx, creds, &client.ChargeRequest{
		Amount:          amount,
		Currency:        currency,
		Reference:       reference,
		CustomerID:      customerID,
		PaymentMethodID: paymentMethodID,
		RedirectURL:     redirectURL,
		Recurring:       recurring,
		Meta:            meta,
	})
	if err != nil {
		logger.WithError(err).Error("token charge failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": err.Error(), "entity_type": "prompt", "reference": reference,
		})
		return nil
	}
	h.emitChargeStatus(ctx, logger, promptID, ch, "recurring", reference)
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

// handleAuthorize completes PIN / OTP / AVS on a pending charge.
func (h *promptHandler) handleAuthorize(
	ctx context.Context,
	logger *util.LogEntry,
	prompt *paymentv1.InitiatePromptRequest,
	creds *client.Credentials,
	promptID string,
) error {
	chargeID := extraString(prompt.GetExtra(), "charge_id")
	if chargeID == "" {
		chargeID = extraString(prompt.GetExtra(), "external_id")
	}
	if chargeID == "" {
		h.metrics.QueueFailed(ctx, "prompt", "validation_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error": "charge_id required for authorize", "entity_type": "prompt",
		})
		return nil
	}

	authType := strings.ToLower(extraString(prompt.GetExtra(), "authorization_type"))
	if authType == "" {
		authType = strings.ToLower(extraString(prompt.GetExtra(), "auth_type"))
	}
	req := &client.UpdateChargeRequest{
		Authorization: client.ChargeAuthorization{Type: authType},
	}
	switch authType {
	case "pin":
		req.Authorization.PIN = &client.PINAuth{
			Nonce:        extraString(prompt.GetExtra(), "nonce"),
			EncryptedPIN: extraString(prompt.GetExtra(), "encrypted_pin"),
		}
		// Server-side encrypt when clear PIN + encryption key available.
		if req.Authorization.PIN.EncryptedPIN == "" {
			if pin := extraString(prompt.GetExtra(), "pin"); pin != "" && creds.EncryptionKey != "" {
				nonce, enc, err := client.EncryptPIN(pin, creds.EncryptionKey)
				if err != nil {
					h.emitStatus(ctx, promptID, chargeID, commonv1.STATUS_FAILED, map[string]any{
						"error": err.Error(), "entity_type": "prompt",
					})
					return nil
				}
				req.Authorization.PIN.Nonce = nonce
				req.Authorization.PIN.EncryptedPIN = enc
			}
		}
	case "otp":
		req.Authorization.OTP = &client.OTPAuth{
			Code: extraString(prompt.GetExtra(), "otp"),
		}
	case "avs":
		req.Authorization.AVS = &client.AVSAuth{
			Address: &client.CustomerAddress{
				City:       extraString(prompt.GetExtra(), "avs_city"),
				Country:    extraString(prompt.GetExtra(), "avs_country"),
				Line1:      extraString(prompt.GetExtra(), "avs_line1"),
				Line2:      extraString(prompt.GetExtra(), "avs_line2"),
				PostalCode: extraString(prompt.GetExtra(), "avs_postal_code"),
				State:      extraString(prompt.GetExtra(), "avs_state"),
			},
		}
	default:
		h.metrics.QueueFailed(ctx, "prompt", "validation_error")
		h.emitStatus(ctx, promptID, chargeID, commonv1.STATUS_FAILED, map[string]any{
			"error": "unknown authorization_type", "entity_type": "prompt",
		})
		return nil
	}

	ch, err := h.fwCli.UpdateCharge(ctx, creds, chargeID, req)
	if err != nil {
		logger.WithError(err).Error("authorize charge failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		h.emitStatus(ctx, promptID, chargeID, commonv1.STATUS_FAILED, map[string]any{
			"error": err.Error(), "entity_type": "prompt",
		})
		return nil
	}
	h.emitChargeStatus(ctx, logger, promptID, ch, "authorize", ch.Reference)
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

func (h *promptHandler) emitChargeStatus(
	ctx context.Context,
	logger *util.LogEntry,
	promptID string,
	ch *client.Charge,
	mode, reference string,
) {
	apiVersion := "v4"
	extras := map[string]any{
		"entity_type":   "prompt",
		"provider":      "flutterwave",
		"api_version":   apiVersion,
		"mode":          mode,
		"reference":     ch.Reference,
		"charge_id":     ch.ID,
		"charge_status": ch.Status,
		"tx_ref":        firstNonEmpty(ch.Reference, reference),
	}
	if ch.CustomerID != "" {
		extras["customer_id"] = ch.CustomerID
	}
	if pmd := client.PaymentMethodIDFromCharge(ch); pmd != "" {
		extras["payment_method_id"] = pmd
	}

	na := client.ExtractNextAction(ch)
	if na.Type != "" {
		extras["next_action"] = string(na.Type)
		extras["next_action_type"] = string(na.Type)
	}
	if na.RedirectURL != "" {
		// Checkout reads extras.checkout_url for 3DS / legacy redirect.
		extras["checkout_url"] = na.RedirectURL
		extras["auth_redirect_url"] = na.RedirectURL
	}
	if na.Note != "" {
		extras["payment_instruction"] = na.Note
	}
	if len(na.Fields) > 0 {
		fields := make([]any, len(na.Fields))
		for i, f := range na.Fields {
			fields[i] = f
		}
		extras["avs_fields"] = fields
	}
	if ch.NextAction != nil {
		if t, _ := ch.NextAction["type"].(string); t == "requires_bank_transfer" {
			if bt, ok := ch.NextAction["requires_bank_transfer"].(map[string]any); ok {
				extras["bank_transfer"] = bt
			}
		}
	}

	// Terminal statuses from provider on first response.
	status := commonv1.STATUS_IN_PROCESS
	switch strings.ToLower(ch.Status) {
	case "succeeded", "successful", "success":
		status = commonv1.STATUS_SUCCESSFUL
	case "failed", "voided":
		status = commonv1.STATUS_FAILED
	}

	logger.WithField("charge_id", ch.ID).WithField("status", ch.Status).
		WithField("next_action", na.Type).Debug("charge initiated")
	h.emitStatus(ctx, promptID, ch.ID, status, extras)
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
		// Without encrypted card fields the client falls back to Standard (if FLWSECK)
		// or returns a clear error directing operators to embedded card encryption.
		return client.PaymentMethodInput{Type: "card"}
	default:
		return client.PaymentMethodInput{Type: "card"}
	}
}

func hasStandardSecretCreds(c *client.Credentials) bool {
	if c == nil {
		return false
	}
	sec := c.SecretKey
	if sec == "" {
		sec = c.ClientSecret
	}
	return strings.HasPrefix(sec, "FLWSECK_") || strings.HasPrefix(sec, "FLWSECK-")
}

// cardFromExtras builds CardDetails from portable prompt extras.
// Keys are provider-agnostic so checkout can switch PSPs without UI changes.
func cardFromExtras(extra *structpb.Struct) *client.CardDetails {
	if extra == nil {
		return nil
	}
	nonce := extraString(extra, "card_nonce")
	if nonce == "" {
		nonce = extraString(extra, "nonce")
	}
	encNum := extraString(extra, "encrypted_card_number")
	encMon := extraString(extra, "encrypted_expiry_month")
	encYear := extraString(extra, "encrypted_expiry_year")
	encCVV := extraString(extra, "encrypted_cvv")
	if encNum == "" || nonce == "" {
		return nil
	}
	card := &client.CardDetails{
		EncryptedCardNumber:  encNum,
		EncryptedExpiryMonth: encMon,
		EncryptedExpiryYear:  encYear,
		EncryptedCVV:         encCVV,
		Nonce:                nonce,
	}
	if strings.EqualFold(extraString(extra, "cof_enabled"), "true") {
		card.COF = &client.CardCOF{
			Enabled:     true,
			AgreementID: extraString(extra, "cof_agreement_id"),
		}
	}
	return card
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
