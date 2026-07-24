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

package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/util"
)

const (
	httpClientTimeout     = 45 * time.Second
	tokenRefreshSkew      = 60 * time.Second // refresh ≥1m before expiry (FW recommendation)
	defaultOAuthURL       = "https://idp.flutterwave.com/realms/flutterwave/protocol/openid-connect/token"
	defaultSandboxBase    = "https://developersandbox-api.flutterwave.com"
	defaultProductionBase = "https://f4bexperience.flutterwave.com"
)

type tokenEntry struct {
	accessToken string
	expiresAt   time.Time
}

type flutterwaveClient struct {
	httpClient *http.Client
	metrics    *integrationobs.Metrics

	mu     sync.Mutex
	tokens map[string]*tokenEntry // key: clientID + ":" + env
}

// NewClient constructs a Flutterwave v4 HTTP client with OAuth token caching.
func NewClient() FlutterwaveClient {
	return &flutterwaveClient{
		httpClient: &http.Client{Timeout: httpClientTimeout},
		metrics:    integrationobs.NewMetrics("flutterwave"),
		tokens:     make(map[string]*tokenEntry),
	}
}

func (c *flutterwaveClient) CreateOrchestratorCharge(
	ctx context.Context,
	creds *Credentials,
	req *OrchestratorChargeRequest,
) (*Charge, error) {
	pmType := strings.ToLower(strings.TrimSpace(req.PaymentMethod.Type))
	if pmType == "" {
		pmType = "card"
	}
	hasEncryptedCard := req != nil && req.PaymentMethod.Card != nil &&
		strings.TrimSpace(req.PaymentMethod.Card.EncryptedCardNumber) != "" &&
		strings.TrimSpace(req.PaymentMethod.Card.Nonce) != ""

	// Direct mobile-money charge (STK / push) — orchestrator or classic v3.
	if pmType == "mobile_money" {
		if IsV3Credentials(creds) || hasStandardSecret(creds) {
			return c.createMoMoChargeV3(ctx, creds, req)
		}
		return c.orchestratorDirectCharge(ctx, creds, req)
	}

	// Embedded card (primary): AES-GCM encrypted fields + v4 OAuth orchestrator.
	// Docs: https://developer.flutterwave.com/docs/payment-orchestrator-flow
	// Docs: https://developer.flutterwave.com/docs/card
	// Works with pure OAuth client_id/client_secret — no FLWSECK required.
	if (pmType == "card" || pmType == "hosted" || pmType == "standard") && hasEncryptedCard {
		if IsV3Credentials(creds) && !hasOAuth(creds) {
			return nil, errors.New(
				"embedded card charges need Flutterwave v4 OAuth (FLUTTERWAVE_CLIENT_ID + CLIENT_SECRET); " +
					"FLWSECK_* alone cannot encrypt-orchestrator charge",
			)
		}
		// Ensure type is card for the API body.
		req.PaymentMethod.Type = "card"
		return c.orchestratorDirectCharge(ctx, creds, req)
	}

	// Opay / USSD / explicit bank_transfer on v4 OAuth.
	if pmType == "opay" || pmType == "ussd" || pmType == "bank_transfer" {
		if hasOAuth(creds) {
			return c.orchestratorDirectCharge(ctx, creds, req)
		}
		if hasStandardSecret(creds) || IsV3Credentials(creds) {
			// No v3 multipay for these shapes — require OAuth.
			return nil, fmt.Errorf("payment method %q requires Flutterwave v4 OAuth credentials", pmType)
		}
	}

	// Optional legacy: Flutterwave Standard multipay page (needs FLWSECK_*), only when
	// the caller explicitly requested hosted multipay without encrypted card fields.
	// Prefer this only when a classic secret is present — never required for OAuth tenants.
	if (pmType == "hosted" || pmType == "standard" || pmType == "payment_link") &&
		(hasStandardSecret(creds) || IsV3Credentials(creds)) {
		return c.createStandardPaymentV3(ctx, creds, req)
	}
	if pmType == "card" && !hasEncryptedCard && (hasStandardSecret(creds) || IsV3Credentials(creds)) {
		// Legacy multipay only as explicit fallback when FLWSECK is configured.
		return c.createStandardPaymentV3(ctx, creds, req)
	}

	if pmType == "card" || pmType == "hosted" || pmType == "standard" || pmType == "payment_link" {
		return nil, fmt.Errorf(
			"card collection requires encrypted card fields from hosted checkout " +
				"(encrypted_card_number + card_nonce via pay.*/c/{{session}}). " +
				"Configure CHECKOUT_CARD_ENCRYPTION_KEY / FLUTTERWAVE_ENCRYPTION_KEY and open " +
				"the pay page so the browser encrypts the card. " +
				"Pure v4 OAuth does not use FLWSECK multipay",
		)
	}

	return nil, fmt.Errorf("unsupported flutterwave payment method type %q", pmType)
}

// hasOAuth reports v4 client-credentials (UUID-style client id/secret).
func hasOAuth(c *Credentials) bool {
	if c == nil {
		return false
	}
	if strings.TrimSpace(c.ClientID) == "" || strings.TrimSpace(c.ClientSecret) == "" {
		return false
	}
	// v3 keys are never valid OAuth client secrets.
	if strings.HasPrefix(c.ClientSecret, "FLWSECK_") || strings.HasPrefix(c.ClientSecret, "FLWSECK-") {
		return false
	}
	if strings.HasPrefix(c.ClientID, "FLWPUBK_") {
		return false
	}
	return true
}

// orchestratorDirectCharge is POST /orchestration/direct-charges (v4 OAuth).
// Retries transient provider failures so a committed pay still collects.
func (c *flutterwaveClient) orchestratorDirectCharge(
	ctx context.Context,
	creds *Credentials,
	req *OrchestratorChargeRequest,
) (*Charge, error) {
	var out *Charge
	err := withRetry(ctx, RetryPolicy{}, func() error {
		var env APIEnvelope[Charge]
		if err := c.doJSON(ctx, creds, http.MethodPost, "/orchestration/direct-charges", req, &env); err != nil {
			return err
		}
		if !strings.EqualFold(env.Status, "success") && env.Data.ID == "" {
			return apiError("orchestrator charge", env.Status, env.Message, env.Error)
		}
		cp := env.Data
		out = &cp
		return nil
	})
	return out, err
}

func (c *flutterwaveClient) CreateCustomer(
	ctx context.Context,
	creds *Credentials,
	req *CreateCustomerRequest,
) (*Customer, error) {
	var env APIEnvelope[Customer]
	if err := c.doJSON(ctx, creds, http.MethodPost, "/customers", req, &env); err != nil {
		return nil, err
	}
	if env.Data.ID == "" {
		return nil, apiError("create customer", env.Status, env.Message, env.Error)
	}
	return &env.Data, nil
}

func (c *flutterwaveClient) CreatePaymentMethod(
	ctx context.Context,
	creds *Credentials,
	req *PaymentMethodInput,
) (*PaymentMethod, error) {
	var env APIEnvelope[PaymentMethod]
	if err := c.doJSON(ctx, creds, http.MethodPost, "/payment-methods", req, &env); err != nil {
		return nil, err
	}
	if env.Data.ID == "" {
		return nil, apiError("create payment method", env.Status, env.Message, env.Error)
	}
	return &env.Data, nil
}

func (c *flutterwaveClient) CreateCharge(
	ctx context.Context,
	creds *Credentials,
	req *ChargeRequest,
) (*Charge, error) {
	var out *Charge
	err := withRetry(ctx, RetryPolicy{}, func() error {
		var env APIEnvelope[Charge]
		if err := c.doJSON(ctx, creds, http.MethodPost, "/charges", req, &env); err != nil {
			return err
		}
		if !strings.EqualFold(env.Status, "success") && env.Data.ID == "" {
			return apiError("create charge", env.Status, env.Message, env.Error)
		}
		cp := env.Data
		out = &cp
		return nil
	})
	return out, err
}

func (c *flutterwaveClient) UpdateCharge(
	ctx context.Context,
	creds *Credentials,
	chargeID string,
	req *UpdateChargeRequest,
) (*Charge, error) {
	if strings.TrimSpace(chargeID) == "" {
		return nil, errors.New("charge id is required")
	}
	var env APIEnvelope[Charge]
	path := "/charges/" + url.PathEscape(chargeID)
	if err := c.doJSON(ctx, creds, http.MethodPut, path, req, &env); err != nil {
		return nil, err
	}
	if env.Data.ID == "" {
		return nil, apiError("update charge", env.Status, env.Message, env.Error)
	}
	return &env.Data, nil
}

// hasStandardSecret reports whether credentials include a classic secret key
// usable for POST /v3/payments (hosted Standard checkout).
func hasStandardSecret(c *Credentials) bool {
	if c == nil {
		return false
	}
	sec := firstNonEmpty(c.SecretKey, c.ClientSecret)
	return strings.HasPrefix(sec, "FLWSECK_") || strings.HasPrefix(sec, "FLWSECK-")
}

func (c *flutterwaveClient) GetCharge(
	ctx context.Context,
	creds *Credentials,
	chargeID string,
) (*Charge, error) {
	if IsV3Credentials(creds) {
		return c.getChargeV3(ctx, creds, chargeID)
	}
	var env APIEnvelope[Charge]
	if err := c.doJSON(ctx, creds, http.MethodGet, "/charges/"+url.PathEscape(chargeID), nil, &env); err != nil {
		return nil, err
	}
	if env.Data.ID == "" {
		return nil, apiError("get charge", env.Status, env.Message, env.Error)
	}
	return &env.Data, nil
}

func (c *flutterwaveClient) CreateTransferRecipient(
	ctx context.Context,
	creds *Credentials,
	req *TransferRecipientRequest,
) (string, error) {
	var env APIEnvelope[map[string]any]
	if err := c.doJSON(ctx, creds, http.MethodPost, "/transfers/recipients", req, &env); err != nil {
		return "", err
	}
	if id, ok := env.Data["id"].(string); ok && id != "" {
		return id, nil
	}
	return "", apiError("create recipient", env.Status, env.Message, env.Error)
}

func (c *flutterwaveClient) CreateTransfer(
	ctx context.Context,
	creds *Credentials,
	req map[string]any,
) (*Transfer, error) {
	if IsV3Credentials(creds) {
		return c.createTransferV3(ctx, creds, req)
	}
	var env APIEnvelope[Transfer]
	if err := c.doJSON(ctx, creds, http.MethodPost, "/transfers", req, &env); err != nil {
		return nil, err
	}
	if env.Data.ID == "" {
		return nil, apiError("create transfer", env.Status, env.Message, env.Error)
	}
	return &env.Data, nil
}

func (c *flutterwaveClient) CreateDirectTransfer(
	ctx context.Context,
	creds *Credentials,
	req *DirectTransferRequest,
) (*Transfer, error) {
	var env APIEnvelope[Transfer]
	if err := c.doJSON(ctx, creds, http.MethodPost, "/direct-transfers", req, &env); err != nil {
		return nil, err
	}
	if env.Data.ID == "" {
		return nil, apiError("direct transfer", env.Status, env.Message, env.Error)
	}
	return &env.Data, nil
}

func (c *flutterwaveClient) GetTransfer(
	ctx context.Context,
	creds *Credentials,
	transferID string,
) (*Transfer, error) {
	var env APIEnvelope[Transfer]
	if err := c.doJSON(ctx, creds, http.MethodGet, "/transfers/"+url.PathEscape(transferID), nil, &env); err != nil {
		return nil, err
	}
	if env.Data.ID == "" {
		return nil, apiError("get transfer", env.Status, env.Message, env.Error)
	}
	return &env.Data, nil
}

// VerifyWebhookSignature validates webhook authenticity.
// v4: flutterwave-signature = base64(HMAC-SHA256(secret, body))
// v3: verif-hash header may equal the secret hash exactly (also accepted when equal).
func (c *flutterwaveClient) VerifyWebhookSignature(rawBody []byte, signatureHeader, secretHash string) bool {
	if secretHash == "" || signatureHeader == "" {
		return false
	}
	// Exact match (v3 verif-hash style).
	if hmac.Equal([]byte(signatureHeader), []byte(secretHash)) {
		return true
	}
	mac := hmac.New(sha256.New, []byte(secretHash))
	_, _ = mac.Write(rawBody)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signatureHeader))
}

// --- HTTP / OAuth ---

//nolint:funlen,nonamedreturns // HTTP client glue with metrics; retErr captured by defer.
func (c *flutterwaveClient) doJSON(
	ctx context.Context,
	creds *Credentials,
	method, path string,
	payload any,
	result any,
) (retErr error) {
	op := strings.TrimPrefix(path, "/")
	if i := strings.Index(op, "/"); i > 0 {
		op = op[:i]
	}
	ctx, done := c.metrics.ObserveProviderCall(ctx, op)
	defer func() { done(retErr) }()

	token, err := c.accessToken(ctx, creds)
	if err != nil {
		return err
	}

	base := resolveAPIBase(creds)
	urlStr := strings.TrimRight(base, "/") + path

	var body io.Reader
	if payload != nil {
		raw, mErr := json.Marshal(payload)
		if mErr != nil {
			return fmt.Errorf("marshal request: %w", mErr)
		}
		body = bytes.NewReader(raw)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("X-Trace-Id", newTraceID())
	httpReq.Header.Set("X-Idempotency-Key", newIdempotencyKey())

	logger := util.Log(ctx).WithField("type", "flutterwave.v4.api").WithField("path", path)
	defer logger.Release()

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	logger.WithFields(map[string]any{
		"http_status": resp.StatusCode,
		"body_len":    len(respBody),
	}).Debug("flutterwave response")

	if resp.StatusCode == http.StatusUnauthorized {
		// Force token refresh once.
		c.invalidateToken(creds)
		token, err = c.accessToken(ctx, creds)
		if err != nil {
			return err
		}
		// Retry once with new token.
		httpReq2, _ := http.NewRequestWithContext(ctx, method, urlStr, replayBody(payload))
		if httpReq2 != nil {
			httpReq2.Header.Set("Authorization", "Bearer "+token)
			httpReq2.Header.Set("Content-Type", "application/json")
			httpReq2.Header.Set("Accept", "application/json")
			httpReq2.Header.Set("X-Trace-Id", newTraceID())
			httpReq2.Header.Set("X-Idempotency-Key", newIdempotencyKey())
			resp2, err2 := c.httpClient.Do(httpReq2)
			if err2 != nil {
				return fmt.Errorf("retry request: %w", err2)
			}
			defer func() { _ = resp2.Body.Close() }()
			respBody, _ = io.ReadAll(resp2.Body)
			resp = resp2
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("flutterwave http %d: %s", resp.StatusCode, truncate(string(respBody), 512))
	}
	if result == nil || len(respBody) == 0 {
		return nil
	}
	if err = json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *flutterwaveClient) accessToken(ctx context.Context, creds *Credentials) (string, error) {
	if IsV3Credentials(creds) {
		return "", errors.New("v3 secret-key mode does not use OAuth access tokens")
	}
	if creds == nil || creds.ClientID == "" || creds.ClientSecret == "" {
		return "", errors.New("flutterwave client_id and client_secret are required (v4 OAuth)")
	}
	key := creds.ClientID + ":" + strings.ToLower(creds.Environment)

	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.tokens[key]; ok && time.Until(ent.expiresAt) > tokenRefreshSkew {
		return ent.accessToken, nil
	}

	tokenURL := creds.OAuthTokenURL
	if tokenURL == "" {
		tokenURL = defaultOAuthURL
	}

	form := url.Values{}
	form.Set("client_id", creds.ClientID)
	form.Set("client_secret", creds.ClientSecret)
	form.Set("grant_type", "client_credentials")

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build oauth request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("oauth token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth token http %d: %s", resp.StatusCode, truncate(string(body), 256))
	}
	var tok oauthTokenResponse
	if err = json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("decode oauth token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", errors.New("oauth response missing access_token")
	}
	expiresIn := tok.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 600
	}
	c.tokens[key] = &tokenEntry{
		accessToken: tok.AccessToken,
		expiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	return tok.AccessToken, nil
}

func (c *flutterwaveClient) invalidateToken(creds *Credentials) {
	if creds == nil {
		return
	}
	key := creds.ClientID + ":" + strings.ToLower(creds.Environment)
	c.mu.Lock()
	delete(c.tokens, key)
	c.mu.Unlock()
}

func resolveAPIBase(creds *Credentials) string {
	if creds != nil && creds.APIBaseURL != "" {
		return creds.APIBaseURL
	}
	if creds != nil && strings.EqualFold(creds.Environment, "production") {
		return defaultProductionBase
	}
	return defaultSandboxBase
}

func apiError(op, status, message string, errObj *struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}) error {
	if errObj != nil && errObj.Message != "" {
		return fmt.Errorf("flutterwave %s: %s (%s/%s)", op, errObj.Message, errObj.Type, errObj.Code)
	}
	return fmt.Errorf("flutterwave %s: status=%s message=%s", op, status, message)
}

func newTraceID() string {
	return "trace-" + randomHex(16)
}

func newIdempotencyKey() string {
	return "idem-" + randomHex(16)
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func replayBody(payload any) io.Reader {
	if payload == nil {
		return nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return bytes.NewReader(raw)
}

// ExtractRedirectURL pulls next_action.redirect_url.url from a charge.
func ExtractRedirectURL(ch *Charge) string {
	if ch == nil || ch.NextAction == nil {
		return ""
	}
	if t, _ := ch.NextAction["type"].(string); t == "redirect_url" {
		if ru, ok := ch.NextAction["redirect_url"].(map[string]any); ok {
			if u, ok := ru["url"].(string); ok {
				return u
			}
		}
	}
	return ""
}

// ExtractPaymentInstructionNote pulls next_action.payment_instruction.note.
func ExtractPaymentInstructionNote(ch *Charge) string {
	if ch == nil || ch.NextAction == nil {
		return ""
	}
	if t, _ := ch.NextAction["type"].(string); t == "payment_instruction" {
		if pi, ok := ch.NextAction["payment_instruction"].(map[string]any); ok {
			if n, ok := pi["note"].(string); ok {
				return n
			}
		}
	}
	return ""
}
