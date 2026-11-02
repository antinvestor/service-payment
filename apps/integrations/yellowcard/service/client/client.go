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

	retryMaxAttempts = 3
	retryBaseDelay   = 200 * time.Millisecond
	retryMaxDelay    = 2 * time.Second
)

type client struct {
	httpClient *http.Client
	metrics    *integrationobs.Metrics
	now        func() time.Time
}

// NewClient creates a Yellow Card Payments API client.
func NewClient() YellowcardClient {
	return &client{
		httpClient: &http.Client{Timeout: httpTimeout},
		metrics:    integrationobs.NewMetrics("yellowcard"),
		now:        time.Now,
	}
}

// errRetryable marks transient transport or server failures on idempotent reads.
type errRetryable struct{ err error }

func (e *errRetryable) Error() string { return e.err.Error() }
func (e *errRetryable) Unwrap() error { return e.err }

// doRequest performs one signed JSON request. Only GET requests are retried;
// financial POSTs rely on the sequenceId for idempotency at Yellow Card.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *client) doRequest(
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
		delay *= 2
		if delay > retryMaxDelay {
			delay = retryMaxDelay
		}
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
	logger := util.Log(ctx).WithField("type", "yellowcard."+op)
	defer logger.Release()

	var (
		rawBody []byte
		body    io.Reader
	)
	if payload != nil {
		var err error
		rawBody, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal %s payload: %w", op, err)
		}
		body = bytes.NewReader(rawBody)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, creds.ResolveBaseURL()+endpoint, body)
	if err != nil {
		return fmt.Errorf("create %s request: %w", op, err)
	}
	httpReq.Header.Set("Accept", "application/json")
	if payload != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	SignRequest(httpReq, rawBody, creds.APIKey, creds.SecretKey, c.now())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) || errors.Is(err, io.EOF) {
			return &errRetryable{err: fmt.Errorf("execute %s request: %w", op, err)}
		}
		return fmt.Errorf("execute %s request: %w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	logger.WithFields(map[string]any{
		"status":   resp.StatusCode,
		"response": string(respBody),
	}).Debug(op + " response")

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		apiErr := &APIError{HTTPStatus: resp.StatusCode}
		if jsonErr := json.Unmarshal(respBody, apiErr); jsonErr != nil || apiErr.Code == "" {
			apiErr.Code = http.StatusText(resp.StatusCode)
			apiErr.Message = string(respBody)
		}
		if isRetryableStatus(resp.StatusCode) {
			return &errRetryable{err: apiErr}
		}
		return apiErr
	}

	if result == nil || len(respBody) == 0 {
		return nil
	}
	if err = json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("decode %s response: %w", op, err)
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

// --- receives ---

func (c *client) SubmitReceive(ctx context.Context, creds *Credentials, req *ReceiveRequest) (*Receive, error) {
	var out Receive
	if err := c.doRequest(ctx, creds, http.MethodPost, "/receive", req, &out, "receive submit"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) AcceptReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error) {
	return c.receiveAction(ctx, creds, id, "accept")
}

func (c *client) DenyReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error) {
	return c.receiveAction(ctx, creds, id, "deny")
}

func (c *client) CancelReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error) {
	return c.receiveAction(ctx, creds, id, "cancel")
}

func (c *client) RefundReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error) {
	return c.receiveAction(ctx, creds, id, "refund")
}

func (c *client) receiveAction(ctx context.Context, creds *Credentials, id, action string) (*Receive, error) {
	var out Receive
	endpoint := "/receive/" + url.PathEscape(id) + "/" + action
	if err := c.doRequest(ctx, creds, http.MethodPost, endpoint, struct{}{}, &out, "receive "+action); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) GetReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error) {
	var out Receive
	endpoint := "/receive/" + url.PathEscape(id)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &out, "receive lookup"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) GetReceiveBySequenceID(ctx context.Context, creds *Credentials, sequenceID string) (*Receive, error) {
	var out Receive
	endpoint := "/receive/sequence-id/" + url.PathEscape(sequenceID)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &out, "receive lookup"); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- sends ---

func (c *client) SubmitSend(ctx context.Context, creds *Credentials, req *SendRequest) (*Send, error) {
	var out Send
	if err := c.doRequest(ctx, creds, http.MethodPost, "/send", req, &out, "send submit"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) AcceptSend(ctx context.Context, creds *Credentials, id string) (*Send, error) {
	return c.sendAction(ctx, creds, id, "accept")
}

func (c *client) DenySend(ctx context.Context, creds *Credentials, id string) (*Send, error) {
	return c.sendAction(ctx, creds, id, "deny")
}

func (c *client) sendAction(ctx context.Context, creds *Credentials, id, action string) (*Send, error) {
	var out Send
	endpoint := "/send/" + url.PathEscape(id) + "/" + action
	if err := c.doRequest(ctx, creds, http.MethodPost, endpoint, struct{}{}, &out, "send "+action); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) GetSend(ctx context.Context, creds *Credentials, id string) (*Send, error) {
	var out Send
	endpoint := "/send/" + url.PathEscape(id)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &out, "send lookup"); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) GetSendBySequenceID(ctx context.Context, creds *Credentials, sequenceID string) (*Send, error) {
	var out Send
	endpoint := "/send/sequence-id/" + url.PathEscape(sequenceID)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &out, "send lookup"); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- catalog ---

func (c *client) GetChannels(ctx context.Context, creds *Credentials, country string) ([]Channel, error) {
	var out []Channel
	endpoint := "/channels" + countryQuery(country)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &out, "channels"); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) GetNetworks(ctx context.Context, creds *Credentials, country string) ([]Network, error) {
	var out []Network
	endpoint := "/networks" + countryQuery(country)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &out, "networks"); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) GetRates(ctx context.Context, creds *Credentials, currency string) ([]Rate, error) {
	var out ratesResponse
	endpoint := "/rates"
	if currency != "" {
		endpoint += "?currency=" + url.QueryEscape(currency)
	}
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &out, "rates"); err != nil {
		return nil, err
	}
	return out.Rates, nil
}

func countryQuery(country string) string {
	if country == "" {
		return ""
	}
	return "?country=" + url.QueryEscape(country)
}
