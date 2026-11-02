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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Meta keys the integration writes on every charge. Flutterwave returns
// them on the verified charge, so they are trusted only after a re-fetch.
const (
	MetaPromptID         = "prompt_id"
	MetaTenantID         = "tenant_id"
	MetaPartitionID      = "partition_id"
	MetaExpectedAmount   = "expected_amount"
	MetaExpectedCurrency = "expected_currency"
)

type promptHandler struct {
	credentialResolver
	statusEmitter
	fwCli   client.FlutterwaveClient
	cfg     *config.FlutterwaveConfig
	metrics *integrationobs.Metrics
}

// NewPromptHandler handles InitiatePrompt via Flutterwave v4 orchestrator charges.
//
// Every write is idempotent on the prompt: the charge reference and the
// X-Idempotency-Key are derived from the prompt id, the charge is looked up
// before one is created, and an unknown outcome (timeout, 5xx) is never
// reported as failed — the message is redelivered until Flutterwave's answer
// is known.
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

// fail reports a definite failure.
func (h *promptHandler) fail(ctx context.Context, promptID, chargeID, reason string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, "prompt", reason)
	if extras == nil {
		extras = map[string]any{}
	}
	extras["entity_type"] = "prompt"
	extras["provider"] = "flutterwave"
	extras["error_code"] = reason
	return h.emitStatus(ctx, promptID, chargeID, commonv1.STATUS_FAILED, extras)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "flutterwave.prompt")
	defer logger.Release()

	prompt := &paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil
	}
	promptID := prompt.GetId()
	if promptID == "" {
		logger.Error("prompt without id dropped")
		h.metrics.QueueFailed(ctx, "prompt", "validation_error")
		return nil
	}
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		if errors.Is(err, errCredentialsUnavailable) {
			logger.WithError(err).Warn("credentials unavailable; redelivering")
			return err
		}
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, promptID, "", "credentials_error", map[string]any{"error": err.Error()})
	}

	// Authorization step for an existing charge (PIN / OTP / AVS) — no new charge.
	if strings.EqualFold(extraString(prompt.GetExtra(), "action"), "authorize") {
		return h.handleAuthorize(ctx, logger, prompt, creds)
	}

	// Tokenized / subscription renewal: charge saved payment_method_id.
	if pmd := extraString(prompt.GetExtra(), "payment_method_id"); pmd != "" {
		if cus := extraString(prompt.GetExtra(), "customer_id"); cus != "" {
			return h.handleTokenCharge(ctx, logger, prompt, creds, cus, pmd, headers)
		}
	}
	return h.handleCharge(ctx, logger, prompt, creds, headers)
}

// Expectation is what a prompt asked for; a charge is only successful when
// Flutterwave collected exactly this.
type Expectation struct {
	Amount   float64
	Currency string
}

//nolint:funlen // one charge flow: validate, look up, create, report.
func (h *promptHandler) handleCharge(
	ctx context.Context,
	logger *util.LogEntry,
	prompt *paymentv1.InitiatePromptRequest,
	creds *client.Credentials,
	headers map[string]string,
) error {
	promptID := prompt.GetId()
	amount, currency, err := providerAmount(prompt.GetAmount())
	if err != nil {
		return h.fail(ctx, promptID, "", "invalid_amount", map[string]any{"error": err.Error()})
	}
	want := Expectation{Amount: amount, Currency: currency}
	reference := chargeReference("prompt", promptID)

	// A charge already made for this prompt (redelivery, lost response) is
	// reported, never made again.
	if ch, done, lookupErr := h.existingCharge(ctx, logger, creds, reference); lookupErr != nil {
		return lookupErr
	} else if done {
		return h.reportCharge(ctx, logger, promptID, ch, "lookup", want)
	}

	phone := prompt.GetSource().GetContactId()
	if phone == "" {
		phone = prompt.GetRecipient().GetContactId()
	}
	email := firstNonEmpty(extraString(prompt.GetExtra(), "customer_email"), extraString(prompt.GetExtra(), "email"))
	if email == "" {
		email = "payer+" + sanitizeRef(promptID) + "@payments.local"
	}
	name := firstNonEmpty(extraString(prompt.GetExtra(), "customer_name"), extraString(prompt.GetExtra(), "display_name"))
	redirectURL := firstNonEmpty(
		extraString(prompt.GetExtra(), "success_url"),
		extraString(prompt.GetExtra(), "redirect_url"),
		headerOrDefault(headers, config.HeaderRedirectURL, h.cfg.DefaultRedirectURL),
	)

	meta := h.chargeMeta(ctx, prompt, headers, want)

	// Payment method: encrypted card (embedded checkout) → mobile money → default.
	var pm client.PaymentMethodInput
	var mode string
	forced := strings.ToLower(extraString(prompt.GetExtra(), "payment_method_type"))
	var corridor *MoMoCorridor
	if digits := digitsOnly(phone); len(digits) >= 10 { //nolint:mnd // shortest international MSISDN
		corridor = resolveMoMoCorridor(digits, currency)
	}
	switch card := cardFromExtras(prompt.GetExtra()); {
	case card != nil:
		pm = client.PaymentMethodInput{Type: "card", Card: card}
		mode = "card"
	case forced == "mobile_money" || forced == "momo":
		if corridor == nil {
			return h.fail(ctx, promptID, "", "invalid_corridor", map[string]any{
				"error": "no mobile money corridor charges " + currency + " for this number",
			})
		}
		pm = buildMoMoPaymentMethod(corridor, extraString(prompt.GetExtra(), "network"))
		mode = "mobile_money"
	case forced == "" && corridor != nil:
		pm = buildMoMoPaymentMethod(corridor, extraString(prompt.GetExtra(), "network"))
		mode = "mobile_money"
	default:
		pm = h.defaultCollectionMethod(prompt)
		mode = pm.Type
	}
	if strings.EqualFold(pm.Type, "card") && (pm.Card == nil || pm.Card.EncryptedCardNumber == "") &&
		!client.IsV3Credentials(creds) && !hasStandardSecretCreds(creds) {
		return h.fail(ctx, promptID, "", "card_encryption_required", map[string]any{
			"error": "card payment requires encrypted card fields from hosted checkout " +
				"(set CHECKOUT_CARD_ENCRYPTION_KEY / FLUTTERWAVE_ENCRYPTION_KEY and pay on the checkout page)",
			"mode": "card",
			"hint": "use_embedded_checkout",
		})
	}

	customer := client.CustomerInput{Email: email, Name: splitName(name), Meta: meta}
	if corridor != nil {
		customer.Phone = &client.CustomerPhone{CountryCode: corridor.CountryCode, Number: corridor.NationalNumber}
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
	ch, err := h.fwCli.CreateOrchestratorCharge(client.WithIdempotencyKey(ctx, reference, "charge"), creds, req)
	if err != nil {
		return h.afterCreateError(ctx, logger, promptID, creds, reference, err, want)
	}
	return h.reportCharge(ctx, logger, promptID, ch, mode, want)
}

// existingCharge finds the charge made with reference. done is true when
// one exists. An unanswered lookup redelivers the prompt (err), except when
// Flutterwave refuses the lookup itself: the idempotency key still protects
// the create.
func (h *promptHandler) existingCharge(
	ctx context.Context,
	logger *util.LogEntry,
	creds *client.Credentials,
	reference string,
) (*client.Charge, bool, error) {
	ch, err := h.fwCli.FindChargeByReference(ctx, creds, reference)
	switch {
	case err == nil:
		// A list item may omit next_action; report the full charge.
		full, getErr := h.fwCli.GetCharge(ctx, creds, ch.ID)
		if getErr != nil {
			return nil, false, fmt.Errorf("read charge %s: %w", ch.ID, getErr)
		}
		return full, true, nil
	case client.IsNotFound(err):
		return nil, false, nil
	case client.IsRejected(err):
		logger.WithError(err).Warn("charge lookup refused; relying on the idempotency key")
		return nil, false, nil
	default:
		logger.WithError(err).Warn("charge lookup failed; redelivering")
		return nil, false, fmt.Errorf("look up charge %s: %w", reference, err)
	}
}

// afterCreateError settles a failed create: a definite refusal fails the
// prompt; anything else may have charged, so the charge is looked up and
// reported, or the prompt is redelivered until it can be.
func (h *promptHandler) afterCreateError(
	ctx context.Context,
	logger *util.LogEntry,
	promptID string,
	creds *client.Credentials,
	reference string,
	createErr error,
	want Expectation,
) error {
	if client.IsRejected(createErr) {
		logger.WithError(createErr).Error("charge refused by Flutterwave")
		return h.fail(ctx, promptID, "", "provider_rejected", map[string]any{
			"error": createErr.Error(), "reference": reference,
		})
	}
	ch, found, err := h.existingCharge(ctx, logger, creds, reference)
	if err != nil || !found {
		logger.WithError(createErr).Warn("charge outcome unknown; redelivering")
		return fmt.Errorf("charge %s outcome unknown: %w", reference, createErr)
	}
	return h.reportCharge(ctx, logger, promptID, ch, "lookup", want)
}

// chargeMeta is the charge's meta: the prompt, its tenancy and the amount
// asked for, so a verified charge can be attributed and checked.
func (h *promptHandler) chargeMeta(
	ctx context.Context,
	prompt *paymentv1.InitiatePromptRequest,
	headers map[string]string,
	want Expectation,
) map[string]string {
	tenantID, partitionID := promptTenancy(ctx, prompt.GetExtra(), headers)
	meta := map[string]string{
		MetaPromptID:         prompt.GetId(),
		"entity":             "prompt",
		MetaExpectedAmount:   strconv.FormatFloat(want.Amount, 'f', 2, 64),
		MetaExpectedCurrency: want.Currency,
	}
	if tenantID != "" {
		meta[MetaTenantID] = tenantID
	}
	if partitionID != "" {
		meta[MetaPartitionID] = partitionID
	}
	for _, k := range []string{"session_ref", "invoice_id", "subscription_id"} {
		if v := extraString(prompt.GetExtra(), k); v != "" {
			meta[k] = v
		}
	}
	return meta
}

// promptTenancy is the prompt's tenant and partition: as the payment service
// stamped them on the prompt, else the message's claims, else its headers.
func promptTenancy(ctx context.Context, extra *structpb.Struct, headers map[string]string) (string, string) {
	tenantID := extraString(extra, "tenant_id")
	partitionID := extraString(extra, "partition_id")
	if claims := security.ClaimsFromContext(ctx); claims != nil {
		tenantID = firstNonEmpty(tenantID, claims.GetTenantID())
		partitionID = firstNonEmpty(partitionID, claims.GetPartitionID())
	}
	return firstNonEmpty(tenantID, headers["tenant_id"]), firstNonEmpty(partitionID, headers["partition_id"])
}

// handleTokenCharge charges a previously saved payment method (subscription renewals).
// Flutterwave v4 OAuth only — classic v3 secret-key multipay has no COF path.
func (h *promptHandler) handleTokenCharge(
	ctx context.Context,
	logger *util.LogEntry,
	prompt *paymentv1.InitiatePromptRequest,
	creds *client.Credentials,
	customerID, paymentMethodID string,
	headers map[string]string,
) error {
	promptID := prompt.GetId()
	if (client.IsV3Credentials(creds) && !hasOAuthCreds(creds)) ||
		strings.EqualFold(extraString(prompt.GetExtra(), "api_version"), "v3") {
		return h.fail(ctx, promptID, "", "validation_error", map[string]any{
			"error":       "token charge requires Flutterwave v4 OAuth credentials (not v3 secret keys)",
			"api_version": "v4",
			"mode":        "recurring",
		})
	}
	amount, currency, err := providerAmount(prompt.GetAmount())
	if err != nil {
		return h.fail(ctx, promptID, "", "invalid_amount", map[string]any{"error": err.Error()})
	}
	want := Expectation{Amount: amount, Currency: currency}
	reference := chargeReference("renew", promptID)

	// A renewal charges a saved card with no payer present: it must never
	// be charged twice.
	if ch, done, lookupErr := h.existingCharge(ctx, logger, creds, reference); lookupErr != nil {
		return lookupErr
	} else if done {
		return h.reportCharge(ctx, logger, promptID, ch, "recurring", want)
	}

	meta := h.chargeMeta(ctx, prompt, headers, want)
	meta["mode"] = "recurring"
	meta["api_version"] = "v4"
	recurring := !strings.EqualFold(extraString(prompt.GetExtra(), "recurring"), "false")

	ch, err := h.fwCli.CreateCharge(client.WithIdempotencyKey(ctx, reference, "charge"), creds, &client.ChargeRequest{
		Amount:          amount,
		Currency:        currency,
		Reference:       reference,
		CustomerID:      customerID,
		PaymentMethodID: paymentMethodID,
		RedirectURL: firstNonEmpty(
			extraString(prompt.GetExtra(), "success_url"), extraString(prompt.GetExtra(), "redirect_url")),
		Recurring: recurring,
		Meta:      meta,
	})
	if err != nil {
		return h.afterCreateError(ctx, logger, promptID, creds, reference, err, want)
	}
	return h.reportCharge(ctx, logger, promptID, ch, "recurring", want)
}

// hasOAuthCreds mirrors client.hasOAuth without exporting it — v4 client credentials only.
func hasOAuthCreds(c *client.Credentials) bool {
	if c == nil {
		return false
	}
	if strings.TrimSpace(c.ClientID) == "" || strings.TrimSpace(c.ClientSecret) == "" {
		return false
	}
	if strings.HasPrefix(c.ClientSecret, "FLWSECK_") || strings.HasPrefix(c.ClientSecret, "FLWSECK-") {
		return false
	}
	return !strings.HasPrefix(c.ClientID, "FLWPUBK_")
}

// handleAuthorize completes PIN / OTP / AVS on a pending charge. The
// payment service only forwards a step for the charge Flutterwave reported
// for this prompt; the charge is re-read first so a redelivered step on a
// settled charge just reports it.
//
//nolint:funlen // one authorization flow: verify, build, send, report.
func (h *promptHandler) handleAuthorize(
	ctx context.Context,
	logger *util.LogEntry,
	prompt *paymentv1.InitiatePromptRequest,
	creds *client.Credentials,
) error {
	promptID := prompt.GetId()
	want := expectationFromPrompt(prompt)
	chargeID := extraString(prompt.GetExtra(), "charge_id")
	if chargeID == "" {
		return h.fail(ctx, promptID, "", "validation_error", map[string]any{"error": "charge_id required for authorize"})
	}

	current, err := h.fwCli.GetCharge(ctx, creds, chargeID)
	if err != nil {
		if client.IsRejected(err) {
			return h.fail(ctx, promptID, chargeID, "charge_not_found", map[string]any{"error": err.Error()})
		}
		return fmt.Errorf("read charge %s: %w", chargeID, err)
	}
	if !chargeOfPrompt(current, promptID) {
		logger.WithField("charge_id", chargeID).WithField("charge_prompt", metaString(current.Meta, MetaPromptID)).
			WithField("charge_reference", current.Reference).
			Error("authorize step names a charge this prompt did not make; dropped")
		h.metrics.QueueFailed(ctx, "prompt", "charge_mismatch")
		return nil
	}
	// The payment service's prompt is the authority on the amount; the
	// charge's own record is used only when the prompt carries none.
	if want.Currency == "" {
		want = ExpectationFromMeta(current.Meta)
	}
	if isTerminalPaymentStatus(ChargeStatus(current.Status)) {
		return h.reportCharge(ctx, logger, promptID, current, "authorize", want)
	}

	authType := strings.ToLower(firstNonEmpty(
		extraString(prompt.GetExtra(), "authorization_type"), extraString(prompt.GetExtra(), "auth_type")))
	req := &client.UpdateChargeRequest{Authorization: client.ChargeAuthorization{Type: authType}}
	var secret string
	switch authType {
	case "pin":
		pin := &client.PINAuth{
			Nonce:        extraString(prompt.GetExtra(), "nonce"),
			EncryptedPIN: extraString(prompt.GetExtra(), "encrypted_pin"),
		}
		if pin.EncryptedPIN == "" || pin.Nonce == "" {
			clear := extraString(prompt.GetExtra(), "pin")
			if clear == "" || creds.EncryptionKey == "" {
				return h.fail(ctx, promptID, chargeID, "validation_error", map[string]any{
					"error": "an encrypted PIN and nonce are required",
				})
			}
			nonce, enc, encErr := client.EncryptPIN(clear, creds.EncryptionKey)
			if encErr != nil {
				return h.fail(ctx, promptID, chargeID, "validation_error", map[string]any{"error": encErr.Error()})
			}
			pin.Nonce, pin.EncryptedPIN = nonce, enc
			secret = clear
		} else {
			secret = pin.EncryptedPIN
		}
		req.Authorization.PIN = pin
	case "otp":
		req.Authorization.OTP = &client.OTPAuth{Code: extraString(prompt.GetExtra(), "otp")}
		secret = req.Authorization.OTP.Code
	case "avs":
		addr := &client.CustomerAddress{
			City:       extraString(prompt.GetExtra(), "avs_city"),
			Country:    extraString(prompt.GetExtra(), "avs_country"),
			Line1:      extraString(prompt.GetExtra(), "avs_line1"),
			Line2:      extraString(prompt.GetExtra(), "avs_line2"),
			PostalCode: extraString(prompt.GetExtra(), "avs_postal_code"),
			State:      extraString(prompt.GetExtra(), "avs_state"),
		}
		req.Authorization.AVS = &client.AVSAuth{Address: addr}
		secret = strings.Join([]string{addr.Line1, addr.Line2, addr.City, addr.State, addr.PostalCode, addr.Country}, "|")
	default:
		return h.fail(ctx, promptID, chargeID, "validation_error", map[string]any{"error": "unknown authorization_type"})
	}

	// The same step sent again (a redelivery) is performed once; a new PIN
	// or OTP is a new step.
	stepKey := sha256.Sum256([]byte(secret))
	stepCtx := client.WithIdempotencyKey(ctx, "prompt", promptID, "authorize", authType, hex.EncodeToString(stepKey[:8]))
	ch, err := h.fwCli.UpdateCharge(stepCtx, creds, chargeID, req)
	if err == nil {
		return h.reportCharge(ctx, logger, promptID, ch, "authorize", want)
	}

	// A refused step (a wrong OTP) or an unknown outcome leaves the charge
	// where Flutterwave holds it: report that, so the payer can try again
	// while it is still pending. Never fail a charge on a guess.
	logger.WithError(err).Warn("authorize step not accepted")
	after, getErr := h.fwCli.GetCharge(ctx, creds, chargeID)
	if getErr != nil {
		return fmt.Errorf("authorize %s outcome unknown: %w", chargeID, err)
	}
	if client.IsRejected(err) && !isTerminalPaymentStatus(ChargeStatus(after.Status)) {
		h.metrics.QueueFailed(ctx, "prompt", "authorization_rejected")
		return h.emitCharge(ctx, logger, promptID, after, "authorize", want, map[string]any{
			"authorization_error": err.Error(),
		})
	}
	return h.reportCharge(ctx, logger, promptID, after, "authorize", want)
}

// reportCharge reports a charge's status.
func (h *promptHandler) reportCharge(
	ctx context.Context,
	logger *util.LogEntry,
	promptID string,
	ch *client.Charge,
	mode string,
	want Expectation,
) error {
	if err := h.emitCharge(ctx, logger, promptID, ch, mode, want, nil); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

// emitCharge reports a charge. A charge that succeeded for an amount or
// currency other than the one asked for is reported as failed: the payer
// did not pay what was owed.
func (h *promptHandler) emitCharge(
	ctx context.Context,
	logger *util.LogEntry,
	promptID string,
	ch *client.Charge,
	mode string,
	want Expectation,
	more map[string]any,
) error {
	status, extras := ChargeReport(ch, mode, want)
	for k, v := range more {
		extras[k] = v
	}
	if extras["error_code"] == "amount_mismatch" {
		logger.WithField("charge_id", ch.ID).WithField("charged", ch.Amount).WithField("currency", ch.Currency).
			Error("charge succeeded for another amount than asked")
	}
	logger.WithField("charge_id", ch.ID).WithField("status", ch.Status).Debug("charge reported")
	return h.emitStatus(ctx, promptID, ch.ID, status, extras)
}

// ChargeReport is the status and extras to report for a charge. A charge
// that succeeded for an amount or currency other than the one asked for is
// reported as failed: the payer did not pay what was owed.
func ChargeReport(ch *client.Charge, mode string, want Expectation) (commonv1.STATUS, map[string]any) {
	status, extras := chargeExtras(ch, mode)
	if status == commonv1.STATUS_SUCCESSFUL && !want.metBy(ch) {
		status = commonv1.STATUS_FAILED
		extras["error_code"] = "amount_mismatch"
		extras["error"] = fmt.Sprintf("charged %s %.2f, expected %s %.2f",
			ch.Currency, ch.Amount, want.Currency, want.Amount)
	}
	return status, extras
}

// metBy reports whether a charge collected what was asked. An unknown
// expectation (none recorded) is met by any charge.
func (e Expectation) metBy(ch *client.Charge) bool {
	if e.Currency == "" {
		return true
	}
	return strings.EqualFold(ch.Currency, e.Currency) && sameAmount(ch.Amount, e.Amount)
}

// ExpectationFromMeta is the amount a charge was created for, as recorded
// in its meta (read back from Flutterwave, so trusted after a re-fetch).
func ExpectationFromMeta(meta map[string]any) Expectation {
	currency := strings.ToUpper(metaString(meta, MetaExpectedCurrency))
	amount, err := strconv.ParseFloat(metaString(meta, MetaExpectedAmount), 64)
	if currency == "" || err != nil {
		return Expectation{}
	}
	return Expectation{Amount: amount, Currency: currency}
}

func expectationFromPrompt(prompt *paymentv1.InitiatePromptRequest) Expectation {
	amount, currency, err := providerAmount(prompt.GetAmount())
	if err != nil {
		return Expectation{}
	}
	return Expectation{Amount: amount, Currency: currency}
}

// chargeStatus maps a Flutterwave charge status.
func ChargeStatus(s string) commonv1.STATUS {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "succeeded", "successful", "success", "completed":
		return commonv1.STATUS_SUCCESSFUL
	case "failed", "voided", "cancelled", "canceled":
		return commonv1.STATUS_FAILED
	}
	return commonv1.STATUS_IN_PROCESS
}

// chargeExtras are the status extras of a charge; the amount and currency
// are what was actually charged, which checkout compares to the session.
func chargeExtras(ch *client.Charge, mode string) (commonv1.STATUS, map[string]any) {
	status := ChargeStatus(ch.Status)
	extras := map[string]any{
		"entity_type":   "prompt",
		"provider":      "flutterwave",
		"api_version":   "v4",
		"mode":          mode,
		"reference":     ch.Reference,
		"charge_id":     ch.ID,
		"charge_status": ch.Status,
		"tx_ref":        ch.Reference,
	}
	if ch.Currency != "" {
		extras["amount"] = strconv.FormatFloat(ch.Amount, 'f', 2, 64)
		extras["currency"] = strings.ToUpper(ch.Currency)
	}
	if ch.CustomerID != "" {
		extras["customer_id"] = ch.CustomerID
	}
	if pmd := client.PaymentMethodIDFromCharge(ch); pmd != "" {
		extras["payment_method_id"] = pmd
	}
	// Only surface next_action while the payer still has something to do.
	// Flutterwave often echoes our success_url as next_action.redirect_url
	// even after the charge succeeded; that would loop the confirm page.
	if !isTerminalPaymentStatus(status) {
		mergeFlutterwaveNextActionExtras(extras, ch)
	}
	return status, extras
}

// chargeReference is the charge reference for an entity: derived from our
// id only, never from caller extras, so a lookup by reference can only find
// this entity's charge. Flutterwave allows 6–42 characters.
func chargeReference(kind, id string) string {
	ref := kind + "-" + sanitizeRef(id)
	if len(ref) <= 42 { //nolint:mnd // Flutterwave reference limit
		return ref
	}
	sum := sha256.Sum256([]byte(id))
	return kind + "-" + hex.EncodeToString(sum[:])[:32]
}

// chargeOfPrompt reports whether a charge was made for this prompt: its
// meta names the prompt, or its reference is the one derived from the
// prompt id. A charge proving neither is not this prompt's.
func chargeOfPrompt(ch *client.Charge, promptID string) bool {
	if owner := metaString(ch.Meta, MetaPromptID); owner != "" {
		return owner == promptID
	}
	return ch.Reference != "" &&
		(ch.Reference == chargeReference("prompt", promptID) || ch.Reference == chargeReference("renew", promptID))
}

func metaString(meta map[string]any, key string) string {
	if v, ok := meta[key].(string); ok {
		return v
	}
	return ""
}

func (h *promptHandler) defaultCollectionMethod(prompt *paymentv1.InitiatePromptRequest) client.PaymentMethodInput {
	methodType := firstNonEmpty(extraString(prompt.GetExtra(), "payment_method_type"), h.cfg.DefaultCollectionMethod, "card")
	switch strings.ToLower(methodType) {
	case "opay":
		return client.PaymentMethodInput{Type: "opay"}
	case "ussd":
		return client.PaymentMethodInput{
			Type: "ussd",
			USSD: &client.USSDDetails{AccountBank: extraString(prompt.GetExtra(), "account_bank")},
		}
	case "bank_transfer", "banktransfer":
		return client.PaymentMethodInput{
			Type: "bank_transfer",
			BankTransfer: &client.BankTransferDetails{
				AccountType:        "dynamic",
				AccountExpiresIn:   3600, //nolint:mnd // one hour to pay
				AccountDisplayName: firstNonEmpty(extraString(prompt.GetExtra(), "account_display_name"), "Payment"),
			},
		}
	default:
		// Without encrypted card fields the client falls back to Standard (if
		// FLWSECK) or refuses with an actionable error.
		return client.PaymentMethodInput{Type: "card"}
	}
}

func hasStandardSecretCreds(c *client.Credentials) bool {
	if c == nil {
		return false
	}
	sec := firstNonEmpty(c.SecretKey, c.ClientSecret)
	return strings.HasPrefix(sec, "FLWSECK_") || strings.HasPrefix(sec, "FLWSECK-")
}

// cardFromExtras builds CardDetails from portable prompt extras.
// Keys are provider-agnostic so checkout can switch PSPs without UI changes.
func cardFromExtras(extra *structpb.Struct) *client.CardDetails {
	if extra == nil {
		return nil
	}
	nonce := firstNonEmpty(extraString(extra, "card_nonce"), extraString(extra, "nonce"))
	encNum := extraString(extra, "encrypted_card_number")
	if encNum == "" || nonce == "" {
		return nil
	}
	card := &client.CardDetails{
		EncryptedCardNumber:  encNum,
		EncryptedExpiryMonth: extraString(extra, "encrypted_expiry_month"),
		EncryptedExpiryYear:  extraString(extra, "encrypted_expiry_year"),
		EncryptedCVV:         extraString(extra, "encrypted_cvv"),
		Nonce:                nonce,
	}
	if strings.EqualFold(extraString(extra, "cof_enabled"), "true") {
		card.COF = &client.CardCOF{Enabled: true, AgreementID: extraString(extra, "cof_agreement_id")}
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
	if len(out) < 6 { //nolint:mnd // Flutterwave minimum reference length
		out += "xxxxxx"
	}
	return out
}

func isTerminalPaymentStatus(status commonv1.STATUS) bool {
	return status == commonv1.STATUS_SUCCESSFUL || status == commonv1.STATUS_FAILED
}

// mergeFlutterwaveNextActionExtras adds redirect / AVS / bank-transfer hints for
// non-terminal charges only (caller must gate on status).
func mergeFlutterwaveNextActionExtras(extras map[string]any, ch *client.Charge) {
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
	mergeBankTransferExtra(extras, ch)
}

func mergeBankTransferExtra(extras map[string]any, ch *client.Charge) {
	if ch.NextAction == nil {
		return
	}
	t, _ := ch.NextAction["type"].(string)
	if t != "requires_bank_transfer" {
		return
	}
	if bt, ok := ch.NextAction["requires_bank_transfer"].(map[string]any); ok {
		extras["bank_transfer"] = bt
	}
}
