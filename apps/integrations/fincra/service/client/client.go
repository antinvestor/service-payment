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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/util"
)

const (
	httpTimeout = 30 * time.Second

	maxResponseBody = 4 << 20 // 4 MiB

	retryMaxAttempts = 3
	retryBaseDelay   = 200 * time.Millisecond
	retryMaxDelay    = 2 * time.Second

	headerAPIKey     = "api-key"
	headerPublicKey  = "x-pub-key"
	headerBusinessID = "x-business-id"
)

type client struct {
	httpClient *http.Client
	metrics    *integrationobs.Metrics
}

// NewClient creates a Fincra API client.
func NewClient() FincraClient {
	return NewClientWithHTTP(&http.Client{Timeout: httpTimeout})
}

// NewClientWithHTTP creates a Fincra API client over the given HTTP client.
func NewClientWithHTTP(httpClient *http.Client) FincraClient {
	return &client{
		httpClient: httpClient,
		metrics:    integrationobs.NewMetrics("fincra"),
	}
}

// errRetryable marks transient transport or server failures on idempotent reads.
type errRetryable struct{ err error }

func (e *errRetryable) Error() string { return e.err.Error() }
func (e *errRetryable) Unwrap() error { return e.err }

// do performs one JSON request and decodes the envelope's data into result.
// Only GET requests are retried; money-moving POSTs carry our reference,
// which Fincra refuses to reuse, so the caller looks the record up instead.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *client) do(
	ctx context.Context,
	creds *Credentials,
	method, endpoint string,
	payload any,
	result any,
	op string,
) (retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, op)
	defer func() { done(retErr) }()

	attempts := 1
	if method == http.MethodGet {
		attempts = retryMaxAttempts
	}

	delay := retryBaseDelay
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		lastErr = c.doOnce(ctx, creds, method, endpoint, payload, result, op)
		var transient *errRetryable
		if lastErr == nil || !errors.As(lastErr, &transient) || attempt == attempts {
			break
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		delay = min(delay*2, retryMaxDelay)
	}

	var transient *errRetryable
	if errors.As(lastErr, &transient) {
		return transient.err
	}
	return lastErr
}

func (c *client) doOnce(
	ctx context.Context,
	creds *Credentials,
	method, endpoint string,
	payload any,
	result any,
	op string,
) error {
	logger := util.Log(ctx).WithField("type", "fincra."+op)
	defer logger.Release()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal %s payload: %w", op, err)
		}
		body = bytes.NewReader(raw)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, creds.ResolveBaseURL()+endpoint, body)
	if err != nil {
		return fmt.Errorf("create %s request: %w", op, err)
	}
	httpReq.Header.Set("Accept", "application/json")
	if payload != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set(headerAPIKey, creds.SecretKey)
	if creds.PublicKey != "" {
		httpReq.Header.Set(headerPublicKey, creds.PublicKey)
	}
	if creds.BusinessID != "" {
		httpReq.Header.Set(headerBusinessID, creds.BusinessID)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) || errors.Is(err, io.EOF) {
			return &errRetryable{err: fmt.Errorf("execute %s request: %w", op, err)}
		}
		return fmt.Errorf("execute %s request: %w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return &errRetryable{err: fmt.Errorf("read %s response: %w", op, err)}
	}
	logger.WithField("status", resp.StatusCode).Debug(op + " response")

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		apiErr := parseAPIError(resp.StatusCode, respBody)
		if isRetryableStatus(resp.StatusCode) {
			return &errRetryable{err: apiErr}
		}
		return apiErr
	}

	env := envelope[json.RawMessage]{}
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, &env); err != nil {
			return fmt.Errorf("decode %s response: %w", op, err)
		}
	}
	if (env.Status != nil && !*env.Status) || (env.Success != nil && !*env.Success) {
		return parseAPIError(resp.StatusCode, respBody)
	}
	if result == nil {
		return nil
	}
	if len(env.Data) == 0 || bytes.Equal(env.Data, []byte("null")) {
		return &APIError{HTTPStatus: resp.StatusCode, ErrorType: "EMPTY_RESPONSE", Message: op + " returned no data"}
	}
	if err = json.Unmarshal(env.Data, result); err != nil {
		return fmt.Errorf("decode %s data: %w", op, err)
	}
	return nil
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
		return true
	}
	return status >= http.StatusInternalServerError
}

// --- hosted checkout ---

func (c *client) CreateCheckout(ctx context.Context, creds *Credentials, req *CheckoutRequest) (*CheckoutLink, error) {
	var out CheckoutLink
	if err := c.do(ctx, creds, http.MethodPost, "/checkout/payments", req, &out, "checkout create"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) GetCheckout(ctx context.Context, creds *Credentials, merchantReference string) (*CheckoutPayment, error) {
	var out CheckoutPayment
	endpoint := "/checkout/payments/merchant-reference/" + url.PathEscape(merchantReference)
	if err := c.do(ctx, creds, http.MethodGet, endpoint, nil, &out, "checkout verify"); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- direct charge ---

func (c *client) InitiateCharge(ctx context.Context, creds *Credentials, req *ChargeRequest) (*Charge, error) {
	var out Charge
	if err := c.do(ctx, creds, http.MethodPost, "/checkout/charges", req, &out, "charge initiate"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) AuthorizeCharge(
	ctx context.Context,
	creds *Credentials,
	chargeID string,
	req *AuthorizeRequest,
) (*Charge, error) {
	var out Charge
	endpoint := "/checkout/charges/" + url.PathEscape(chargeID) + "/authorize"
	if err := c.do(ctx, creds, http.MethodPost, endpoint, req, &out, "charge authorize"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) GetCharge(ctx context.Context, creds *Credentials, merchantReference string) (*Charge, error) {
	var out Charge
	endpoint := "/checkout/charges/merchant-reference/" + url.PathEscape(merchantReference)
	if err := c.do(ctx, creds, http.MethodGet, endpoint, nil, &out, "charge verify"); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- payouts ---

func (c *client) CreatePayout(ctx context.Context, creds *Credentials, req *PayoutRequest) (*Payout, error) {
	var out Payout
	if err := c.do(ctx, creds, http.MethodPost, "/disbursements/payouts", req, &out, "payout create"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) GetPayout(ctx context.Context, creds *Credentials, customerReference string) (*Payout, error) {
	var out Payout
	endpoint := "/disbursements/payouts/customer-reference/" + url.PathEscape(customerReference)
	if err := c.do(ctx, creds, http.MethodGet, endpoint, nil, &out, "payout fetch"); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- catalog ---

func (c *client) ListProviders(
	ctx context.Context,
	creds *Credentials,
	country, currency, destination string,
) ([]Provider, error) {
	q := url.Values{}
	if country != "" {
		q.Set("country", country)
	}
	if currency != "" {
		q.Set("currency", currency)
	}
	if destination != "" {
		q.Set("paymentDestination", destination)
	}
	var out []Provider
	if err := c.do(ctx, creds, http.MethodGet, "/core/banks?"+q.Encode(), nil, &out, "providers"); err != nil {
		return nil, err
	}
	return out, nil
}
