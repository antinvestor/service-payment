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

package coreapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	models "github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/util"
)

const (
	// defaultTokenTTL is used when Jenga does not say how long a token lasts
	// (tokens last about an hour).
	defaultTokenTTL = 50 * time.Minute
	// tokenRefreshMargin refreshes a token before Jenga expires it.
	tokenRefreshMargin = 5 * time.Minute
	// maxResponseBody bounds what is read from a provider response.
	maxResponseBody = 1 << 20
)

// APIError is a Jenga rejection: the request reached Jenga and was refused,
// so nothing was applied.
type APIError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("jenga api error %d code=%d: %s", e.HTTPStatus, e.Code, e.Message)
}

// IsRejected reports whether err is a Jenga rejection (as opposed to an
// outcome-unknown transport or server failure).
func IsRejected(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

type tokenEntry struct {
	token  string
	expiry time.Time
}

// Client represents the Jenga API client with per-credential token caching.
type Client struct {
	HTTPClient *http.Client
	mu         sync.Mutex
	tokens     map[string]*tokenEntry // keyed by credentials fingerprint
	signer     *signer
	metrics    *integrationobs.Metrics
	now        func() time.Time
}

// New creates a new instance of the Jenga API client.
// The httpClient must be provided by the caller (constructed with a timeout).
func New(httpClient *http.Client) *Client {
	return &Client{
		HTTPClient: httpClient,
		tokens:     make(map[string]*tokenEntry),
		signer:     newSigner(),
		metrics:    integrationobs.NewMetrics("jenga"),
		now:        time.Now,
	}
}

// BearerTokenResponse represents the response structure for bearer token generation.
type BearerTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	IssuedAt     string `json:"issuedAt"`
	TokenType    string `json:"tokenType"`
}

// tokenCacheKey fingerprints the credentials a token was issued for, so
// tenants sharing a merchant code with different keys never share tokens.
func tokenCacheKey(creds *Credentials) string {
	sum := sha256.Sum256([]byte(creds.ResolveBaseURL() + "\x00" + creds.MerchantCode + "\x00" +
		creds.APIKey + "\x00" + creds.ConsumerSecret))
	return hex.EncodeToString(sum[:])
}

// GenerateBearerToken returns a bearer token for the credentials, fetching
// one from Jenga only when the cached token is missing or about to expire.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *Client) GenerateBearerToken(
	ctx context.Context,
	creds *Credentials,
) (_ *BearerTokenResponse, retErr error) {
	cacheKey := tokenCacheKey(creds)
	c.mu.Lock()
	if entry, ok := c.tokens[cacheKey]; ok && c.now().Before(entry.expiry) {
		token := entry.token
		c.mu.Unlock()
		return &BearerTokenResponse{AccessToken: token, TokenType: "Bearer"}, nil
	}
	c.mu.Unlock()

	ctx, done := c.metrics.ObserveProviderCall(ctx, "generate_bearer_token")
	defer func() { done(retErr) }()

	body := map[string]string{
		"merchantCode":   creds.MerchantCode,
		"consumerSecret": creds.ConsumerSecret,
	}
	headers := map[string]string{"Api-Key": creds.APIKey}
	var tokenResponse BearerTokenResponse
	if err := c.post(ctx, creds.ResolveBaseURL()+"/authentication/api/v3/authenticate/merchant",
		body, headers, &tokenResponse, "token"); err != nil {
		return nil, err
	}
	if tokenResponse.AccessToken == "" {
		return nil, &APIError{HTTPStatus: http.StatusOK, Message: "token response carried no accessToken"}
	}

	ttl := defaultTokenTTL
	if secs, err := strconv.Atoi(tokenResponse.ExpiresIn); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
		if ttl > tokenRefreshMargin {
			ttl -= tokenRefreshMargin
		}
	}
	c.mu.Lock()
	c.tokens[cacheKey] = &tokenEntry{token: tokenResponse.AccessToken, expiry: c.now().Add(ttl)}
	c.mu.Unlock()

	util.Log(ctx).WithField("operation", "GenerateBearerToken").Debug("bearer token generated and cached")
	return &tokenResponse, nil
}

// post sends a JSON request and decodes a 2xx answer into result. Transport
// failures and 5xx answers are outcome-unknown (collection.Uncertain); 4xx
// answers are rejections (*APIError).
func (c *Client) post(
	ctx context.Context,
	url string,
	payload any,
	headers map[string]string,
	result any,
	op string,
) error {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s request: %w", op, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create %s request: %w", op, err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req) //nolint:bodyclose // closed via defer util.CloseAndLogOnError
	if err != nil {
		return collection.Uncertain(fmt.Errorf("execute %s request: %w", op, err))
	}
	defer util.CloseAndLogOnError(ctx, resp.Body, "close "+op+" response body")

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return collection.Uncertain(fmt.Errorf("read %s response: %w", op, err))
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		util.Log(ctx).WithFields(map[string]any{
			"operation":   op,
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("jenga request failed")
		apiErr := &APIError{HTTPStatus: resp.StatusCode, Message: string(respBody)}
		var parsed struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &parsed) == nil && parsed.Message != "" {
			apiErr.Code, apiErr.Message = parsed.Code, parsed.Message
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			return collection.Uncertain(apiErr)
		}
		return apiErr
	}

	if err = json.Unmarshal(respBody, result); err != nil {
		return collection.Uncertain(fmt.Errorf("decode %s response: %w", op, err))
	}
	return nil
}

// signedHeaders are the headers of an authenticated, signed Jenga call.
func signedHeaders(accessToken, signature string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Signature":     signature,
	}
}

// InitiateSTKUSSD initiates an STK/USSD push request.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *Client) InitiateSTKUSSD(
	ctx context.Context,
	creds *Credentials,
	request models.STKUSSDRequest,
	accessToken string,
) (_ *models.STKUSSDResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "initiate_stk_ussd")
	defer func() { done(retErr) }()
	logger := util.Log(ctx).WithFields(map[string]any{
		"operation":   "InitiateSTKUSSD",
		"payment_ref": request.Payment.Ref,
		"mobile":      request.Payment.MobileNumber,
		"amount":      request.Payment.Amount,
		"currency":    request.Payment.Currency,
	})

	signature, err := c.signer.sign(creds,
		request.Merchant.AccountNumber,
		request.Payment.Ref,
		request.Payment.MobileNumber,
		request.Payment.Telco,
		request.Payment.Amount,
		request.Payment.Currency,
	)
	if err != nil {
		return nil, fmt.Errorf("generate STK signature: %w", err)
	}

	var stkUssdResponse models.STKUSSDResponse
	if err = c.post(ctx, creds.ResolveBaseURL()+"/v3-apis/payment-api/v3.0/stkussdpush/initiate",
		request, signedHeaders(accessToken, signature), &stkUssdResponse, "STK push"); err != nil {
		return nil, err
	}
	if !stkUssdResponse.Status {
		return nil, &APIError{HTTPStatus: http.StatusOK, Code: stkUssdResponse.Code, Message: stkUssdResponse.Message}
	}

	logger.WithFields(map[string]any{
		"transaction_id": stkUssdResponse.TransactionID,
		"response_code":  stkUssdResponse.Code,
	}).Info("STK/USSD push initiated")

	return &stkUssdResponse, nil
}

// CreatePaymentLink creates a payment link using the Jenga API.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *Client) CreatePaymentLink(
	ctx context.Context,
	creds *Credentials,
	request models.PaymentLinkRequest,
	accessToken string,
) (_ *models.PaymentLinkResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "create_payment_link")
	defer func() { done(retErr) }()

	signature, err := c.signer.sign(creds,
		request.PaymentLink.ExpiryDate,
		fmt.Sprint(request.PaymentLink.Amount),
		request.PaymentLink.Currency,
		request.PaymentLink.AmountOption,
		request.PaymentLink.ExternalRef,
	)
	if err != nil {
		return nil, fmt.Errorf("generate payment link signature: %w", err)
	}

	var paymentLinkResponse models.PaymentLinkResponse
	if err = c.post(ctx, creds.ResolveBaseURL()+"/api-checkout/api/v1/create/payment-link",
		request, signedHeaders(accessToken, signature), &paymentLinkResponse, "payment link"); err != nil {
		return nil, err
	}
	if !paymentLinkResponse.Status {
		return nil, &APIError{HTTPStatus: http.StatusOK, Code: paymentLinkResponse.Code, Message: paymentLinkResponse.Message}
	}

	util.Log(ctx).WithField("response_code", paymentLinkResponse.Code).Info("payment link created")
	return &paymentLinkResponse, nil
}

// InitiateTillsPay initiates a tills/pay request (a payment to a merchant till).
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *Client) InitiateTillsPay(
	ctx context.Context,
	creds *Credentials,
	request models.TillsPayRequest,
	accessToken string,
) (_ *models.TillsPayResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "initiate_tills_pay")
	defer func() { done(retErr) }()

	signature, err := c.signer.sign(creds,
		request.Merchant.Till,
		request.Partner.ID,
		request.Payment.Amount,
		request.Payment.Currency,
		request.Payment.Ref,
	)
	if err != nil {
		return nil, fmt.Errorf("generate tills pay signature: %w", err)
	}

	var tillsPayResponse models.TillsPayResponse
	if err = c.post(ctx, creds.ResolveBaseURL()+"/v3-apis/transaction-api/v3.0/tills/pay",
		request, signedHeaders(accessToken, signature), &tillsPayResponse, "tills pay"); err != nil {
		return nil, err
	}
	if !tillsPayResponse.Status {
		return nil, &APIError{HTTPStatus: http.StatusOK, Code: tillsPayResponse.Code, Message: tillsPayResponse.Message}
	}

	util.Log(ctx).WithFields(map[string]any{
		"transaction_id": tillsPayResponse.TransactionID,
		"merchant_name":  tillsPayResponse.MerchantName,
	}).Info("tills pay completed")

	return &tillsPayResponse, nil
}
