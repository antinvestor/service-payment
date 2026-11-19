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
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameclient "github.com/pitabwire/frame/v3/client"
	"github.com/pitabwire/util"
)

const (
	httpTimeout       = 30 * time.Second
	tokenExpiryBuffer = 60 * time.Second // refresh before the token actually expires
	maxResponseBytes  = 1 << 20
)

type tokenEntry struct {
	token  string
	expiry time.Time
}

type airtelClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
	tokens     map[string]*tokenEntry // keyed by environment:clientID
	metrics    *integrationobs.Metrics
}

// NewClient creates an Airtel Money client on a frame HTTP client with a
// fixed outbound timeout (set once here, never per call).
func NewClient(ctx context.Context) AirtelClient {
	return NewClientWithHTTP(frameclient.NewHTTPClient(ctx, frameclient.WithHTTPTimeout(httpTimeout)))
}

// NewClientWithHTTP wraps a caller-supplied HTTP client (tests).
func NewClientWithHTTP(httpClient *http.Client) AirtelClient {
	return &airtelClient{
		httpClient: httpClient,
		tokens:     make(map[string]*tokenEntry),
		metrics:    integrationobs.NewMetrics("airtel"),
	}
}

// generateToken returns a cached bearer token for the client id and
// environment, fetching a new one when none is valid. The HTTP call runs
// outside the lock so other credentials are not blocked by a slow token
// endpoint.
func (c *airtelClient) generateToken(ctx context.Context, creds *AirtelCredentials) (string, error) {
	cacheKey := creds.Environment + ":" + creds.ClientID

	c.mu.RLock()
	entry, ok := c.tokens[cacheKey]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiry) {
		return entry.token, nil
	}

	token, expiry, err := c.fetchToken(ctx, creds)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.tokens[cacheKey] = &tokenEntry{token: token, expiry: expiry}
	c.mu.Unlock()
	return token, nil
}

func (c *airtelClient) fetchToken(ctx context.Context, creds *AirtelCredentials) (string, time.Time, error) {
	body, err := json.Marshal(map[string]string{
		"client_id":     creds.ClientID,
		"client_secret": creds.ClientSecret,
		"grant_type":    "client_credentials",
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("marshal token payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, creds.BaseURL()+"/auth/oauth2/token", bytes.NewReader(body))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	status, respBody, err := c.send(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token request: %w", err)
	}
	if err = httpStatusError("token", status, respBody); err != nil {
		return "", time.Time{}, err
	}
	var tokenResp TokenResponse
	if err = json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("decode token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", time.Time{}, errors.New("token response carried no access token")
	}
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - tokenExpiryBuffer)
	return tokenResp.AccessToken, expiry, nil
}

// send performs a request; transport failures and unreadable bodies are
// outcome-unknown (collection.Uncertain).
func (c *airtelClient) send(req *http.Request) (int, []byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, collection.Uncertain(fmt.Errorf("execute request: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return 0, nil, collection.Uncertain(fmt.Errorf("read response: %w", err))
	}
	return resp.StatusCode, body, nil
}

// httpStatusError classifies an Airtel HTTP status: 5xx is outcome-unknown
// and wrapped as collection.Uncertain; any other non-2xx answer is an
// APIError (IsNotFound, IsRejected).
func httpStatusError(operation string, status int, body []byte) error {
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		return nil
	}
	apiErr := &APIError{Operation: operation, StatusCode: status, Message: string(body)}
	if status >= http.StatusInternalServerError {
		return collection.Uncertain(apiErr)
	}
	return apiErr
}

// postJSON sends an authenticated JSON request with Airtel's market headers
// and decodes the answer into out.
func (c *airtelClient) postJSON(
	ctx context.Context,
	creds *AirtelCredentials,
	operation, path, country, currency string,
	payload, out any,
) error {
	logger := util.Log(ctx).WithField("type", "airtel."+operation)
	defer logger.Release()

	token, err := c.generateToken(ctx, creds)
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s payload: %w", operation, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, creds.BaseURL()+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create %s request: %w", operation, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Country", country)
	httpReq.Header.Set("X-Currency", currency)

	status, respBody, err := c.send(httpReq)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	logger.WithFields(map[string]any{"status": status, "response": string(respBody)}).Debug(operation + " response")
	if err = httpStatusError(operation, status, respBody); err != nil {
		return err
	}
	if err = json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s response: %w", operation, err)
	}
	return nil
}

// resultError turns a 2xx answer whose status.success is false into a
// rejection.
func resultError(operation string, success bool, code, message string) error {
	if success {
		return nil
	}
	return &APIError{Operation: operation, StatusCode: http.StatusOK, Code: code, Message: message}
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *airtelClient) CollectionPush(
	ctx context.Context,
	creds *AirtelCredentials,
	req *CollectionRequest,
) (_ *CollectionResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "collection_push")
	defer func() { done(retErr) }()

	payload := collectionPayload{
		Reference: req.Reference,
		Subscriber: subscriberPayload{
			Country:  req.CountryCode,
			Currency: req.Currency,
			Msisdn:   req.PhoneNumber,
		},
		Transaction: transactionPayload{
			Amount:   req.Amount,
			Country:  req.CountryCode,
			Currency: req.Currency,
			ID:       req.Reference,
		},
	}
	var collResp CollectionResponse
	if err := c.postJSON(ctx, creds, "collection_push", "/merchant/v2/payments/", req.CountryCode, req.Currency, payload, &collResp); err != nil {
		return nil, err
	}
	if err := resultError("collection_push", collResp.Status.Success, collResp.Status.Code, collResp.Status.Message); err != nil {
		return &collResp, err
	}
	return &collResp, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *airtelClient) Disburse(
	ctx context.Context,
	creds *AirtelCredentials,
	req *DisbursementRequest,
) (_ *DisbursementResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "disburse")
	defer func() { done(retErr) }()

	if req.EncryptedPin == "" {
		return nil, ErrPinNotConfigured
	}
	payload := disbursementPayload{
		Payee:     payeePayload{Msisdn: req.PhoneNumber},
		Reference: req.Reference,
		Pin:       req.EncryptedPin,
		Transaction: transactionPayload{
			Amount:   req.Amount,
			Country:  req.CountryCode,
			Currency: req.Currency,
			ID:       req.Reference,
		},
	}
	var disbResp DisbursementResponse
	if err := c.postJSON(ctx, creds, "disburse", "/standard/v2/disbursements/", req.CountryCode, req.Currency, payload, &disbResp); err != nil {
		return nil, err
	}
	if err := resultError("disburse", disbResp.Status.Success, disbResp.Status.Code, disbResp.Status.Message); err != nil {
		return &disbResp, err
	}
	return &disbResp, nil
}

// TransactionStatus queries a collection by our transaction id.
func (c *airtelClient) TransactionStatus(
	ctx context.Context,
	creds *AirtelCredentials,
	transactionID string,
) (*StatusResponse, error) {
	return c.getStatus(ctx, creds, "transaction_status", "/standard/v1/payments/"+url.PathEscape(transactionID))
}

// DisbursementStatus queries a disbursement by our transaction id.
func (c *airtelClient) DisbursementStatus(
	ctx context.Context,
	creds *AirtelCredentials,
	transactionID string,
) (*StatusResponse, error) {
	return c.getStatus(ctx, creds, "disbursement_status", "/standard/v1/disbursements/"+url.PathEscape(transactionID))
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *airtelClient) getStatus(
	ctx context.Context,
	creds *AirtelCredentials,
	metricName, path string,
) (_ *StatusResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, metricName)
	defer func() { done(retErr) }()
	token, err := c.generateToken(ctx, creds)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, creds.BaseURL()+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create status request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Country", creds.CountryCode)
	req.Header.Set("X-Currency", creds.Currency)

	status, respBody, err := c.send(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", metricName, err)
	}
	if err = httpStatusError(metricName, status, respBody); err != nil {
		return nil, err
	}
	var statusResp StatusResponse
	if err = json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("decode status response: %w", err)
	}
	if err = resultError(metricName, statusResp.Status.Success, statusResp.Status.Code, statusResp.Status.Message); err != nil {
		return nil, err
	}
	return &statusResp, nil
}
