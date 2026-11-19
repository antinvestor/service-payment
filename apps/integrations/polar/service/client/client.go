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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameclient "github.com/pitabwire/frame/v3/client"
	"github.com/pitabwire/util"
)

const (
	httpClientTimeout         = 30 * time.Second
	webhookTimestampTolerance = 300 // seconds
	maxResponseBytes          = 1 << 20
)

// APIError is a Polar answer with a non-2xx status.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("polar api status %d: %s", e.StatusCode, e.Body)
}

// IsNotFound reports whether Polar answered 404.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

type polarClient struct {
	httpClient *http.Client
	metrics    *integrationobs.Metrics
}

// NewClient creates a Polar client on a frame HTTP client with a fixed
// outbound timeout (set once here, never per call).
func NewClient(ctx context.Context) PolarClient {
	return NewClientWithHTTP(frameclient.NewHTTPClient(ctx, frameclient.WithHTTPTimeout(httpClientTimeout)))
}

// NewClientWithHTTP wraps a caller-supplied HTTP client (tests).
func NewClientWithHTTP(httpClient *http.Client) PolarClient {
	return &polarClient{httpClient: httpClient, metrics: integrationobs.NewMetrics("polar")}
}

// do sends one request. Transport failures and 5xx answers are reported as
// outcome-unknown (collection.Uncertain) so workers redeliver instead of
// failing a payment Polar may have accepted.
func (c *polarClient) do(ctx context.Context, creds *PolarCredentials, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode polar request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, creds.BaseURL()+path, reader)
	if err != nil {
		return fmt.Errorf("build polar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+creds.APIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return collection.Uncertain(fmt.Errorf("polar %s %s: %w", method, path, err))
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if readErr != nil {
		return collection.Uncertain(fmt.Errorf("read polar response: %w", readErr))
	}
	util.Log(ctx).WithFields(map[string]any{"method": method, "path": path, "status": resp.StatusCode}).
		Debug("polar response")
	if resp.StatusCode >= http.StatusInternalServerError {
		return collection.Uncertain(&APIError{StatusCode: resp.StatusCode, Body: string(respBody)})
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}
	if out != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode polar response: %w", err)
		}
	}
	return nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *polarClient) CreateCheckout(
	ctx context.Context,
	creds *PolarCredentials,
	req *CheckoutRequest,
) (_ *CheckoutResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "create_checkout")
	defer func() { done(retErr) }()

	if len(req.ProductIDs) == 0 {
		return nil, errors.New("polar checkout needs at least one product id")
	}
	payload := checkoutPayload{
		Products:           req.ProductIDs,
		CustomerEmail:      req.CustomerEmail,
		CustomerName:       req.CustomerName,
		ExternalCustomerID: req.ExternalCustomerID,
		SuccessURL:         req.SuccessURL,
		Amount:             req.Amount,
		Metadata:           req.Metadata,
	}
	var out CheckoutResponse
	if err := c.do(ctx, creds, http.MethodPost, "/v1/checkouts/", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *polarClient) GetCheckout(
	ctx context.Context,
	creds *PolarCredentials,
	checkoutID string,
) (_ *CheckoutResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "get_checkout")
	defer func() { done(retErr) }()
	var out CheckoutResponse
	if err := c.do(ctx, creds, http.MethodGet, "/v1/checkouts/"+checkoutID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *polarClient) GetSubscription(
	ctx context.Context,
	creds *PolarCredentials,
	subscriptionID string,
) (_ *Subscription, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "get_subscription")
	defer func() { done(retErr) }()
	var out Subscription
	if err := c.do(ctx, creds, http.MethodGet, "/v1/subscriptions/"+subscriptionID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *polarClient) CancelSubscription(
	ctx context.Context,
	creds *PolarCredentials,
	subscriptionID string,
	atPeriodEnd bool,
) (_ *Subscription, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "cancel_subscription")
	defer func() { done(retErr) }()
	var out Subscription
	var err error
	if atPeriodEnd {
		err = c.do(ctx, creds, http.MethodPatch, "/v1/subscriptions/"+subscriptionID,
			map[string]any{"cancel_at_period_end": true}, &out)
	} else {
		err = c.do(ctx, creds, http.MethodDelete, "/v1/subscriptions/"+subscriptionID+"/revoke", nil, &out)
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyWebhookSignature implements Standard Webhooks verification:
// base64(HMAC-SHA256(secret, "{id}.{timestamp}.{body}")) must match one of
// the "v1,<sig>" entries of webhook-signature, and the timestamp must be
// within five minutes of now.
func (c *polarClient) VerifyWebhookSignature(
	payload []byte,
	headers map[string]string,
	webhookSecret string,
) (*WebhookEvent, error) {
	webhookID := headerValue(headers, "webhook-id")
	timestamp := headerValue(headers, "webhook-timestamp")
	signature := headerValue(headers, "webhook-signature")
	if webhookID == "" || timestamp == "" || signature == "" {
		return nil, errors.New("missing required webhook headers (webhook-id, webhook-timestamp, webhook-signature)")
	}
	secretBytes, err := decodeWebhookSecret(webhookSecret)
	if err != nil {
		return nil, err
	}

	ts, err := parseTimestamp(timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook timestamp: %w", err)
	}
	if math.Abs(time.Since(ts).Seconds()) > webhookTimestampTolerance {
		return nil, errors.New("webhook timestamp too old or too new")
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(webhookID + "." + timestamp + "."))
	mac.Write(payload)
	expected := mac.Sum(nil)

	verified := false
	for _, sig := range strings.Fields(signature) {
		version, value, ok := strings.Cut(sig, ",")
		if !ok || version != "v1" {
			continue
		}
		got, decErr := base64.StdEncoding.DecodeString(value)
		if decErr != nil {
			continue
		}
		if hmac.Equal(got, expected) {
			verified = true
			break
		}
	}
	if !verified {
		return nil, errors.New("webhook signature verification failed")
	}

	var event WebhookEvent
	if unmarshalErr := json.Unmarshal(payload, &event); unmarshalErr != nil {
		return nil, fmt.Errorf("decode webhook event: %w", unmarshalErr)
	}
	event.ID = webhookID
	event.Timestamp = ts.UTC().Format(time.RFC3339)
	return &event, nil
}

func headerValue(headers map[string]string, name string) string {
	for k, v := range headers {
		if strings.EqualFold(k, name) && v != "" {
			return v
		}
	}
	return ""
}

func parseTimestamp(ts string) (time.Time, error) {
	epoch, err := strconv.ParseInt(strings.TrimSpace(ts), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(epoch, 0), nil
}

// decodeWebhookSecret returns the HMAC key. Polar's dashboard shows a plain
// secret: its bytes are the key. A Standard Webhooks "whsec_" secret carries
// the key base64-encoded. An empty secret is refused: it would verify any
// payload.
func decodeWebhookSecret(secret string) ([]byte, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, errors.New("webhook secret is not configured")
	}
	if rest, ok := strings.CutPrefix(secret, "whsec_"); ok {
		key, err := base64.StdEncoding.DecodeString(rest)
		if err != nil {
			return nil, fmt.Errorf("decode whsec_ webhook secret: %w", err)
		}
		return key, nil
	}
	return []byte(secret), nil
}
