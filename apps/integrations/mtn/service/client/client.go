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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameclient "github.com/pitabwire/frame/v3/client"
	"github.com/pitabwire/util"
)

type tokenEntry struct {
	token  string
	expiry time.Time
}

const (
	httpClientTimeout = 30 * time.Second
	tokenExpiryBuffer = 60 * time.Second // refresh before the token actually expires
	maxResponseBytes  = 1 << 20
)

type mtnClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
	tokens     map[string]*tokenEntry // keyed by environment:product:apiUser
	metrics    *integrationobs.Metrics
}

// NewClient creates an MTN MoMo API client on a frame HTTP client with a
// fixed outbound timeout (set once here, never per call).
func NewClient(ctx context.Context) MtnClient {
	return NewClientWithHTTP(frameclient.NewHTTPClient(ctx, frameclient.WithHTTPTimeout(httpClientTimeout)))
}

// NewClientWithHTTP wraps a caller-supplied HTTP client (tests).
func NewClientWithHTTP(httpClient *http.Client) MtnClient {
	return &mtnClient{
		httpClient: httpClient,
		tokens:     make(map[string]*tokenEntry),
		metrics:    integrationobs.NewMetrics("mtn"),
	}
}

// generateToken returns a cached bearer token for the product, fetching a
// new one when none is valid. The HTTP call runs outside the lock so other
// products and credentials are not blocked by a slow token endpoint.
func (c *mtnClient) generateToken(ctx context.Context, creds *MtnCredentials, product string) (string, error) {
	cacheKey := creds.Environment + ":" + product + ":" + creds.APIUser

	c.mu.RLock()
	entry, ok := c.tokens[cacheKey]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiry) {
		return entry.token, nil
	}

	token, expiry, err := c.fetchToken(ctx, creds, product)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.tokens[cacheKey] = &tokenEntry{token: token, expiry: expiry}
	c.mu.Unlock()
	return token, nil
}

func (c *mtnClient) fetchToken(ctx context.Context, creds *MtnCredentials, product string) (string, time.Time, error) {
	url := fmt.Sprintf("%s/%s/token/", creds.BaseURL(), product)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create token request: %w", err)
	}
	auth := base64.StdEncoding.EncodeToString([]byte(creds.APIUser + ":" + creds.APIKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Ocp-Apim-Subscription-Key", creds.SubscriptionKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, collection.Uncertain(fmt.Errorf("execute token request: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", time.Time{}, collection.Uncertain(fmt.Errorf("read token response: %w", err))
	}
	if err = statusError("token", resp.StatusCode, body, http.StatusOK); err != nil {
		return "", time.Time{}, err
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("decode token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", time.Time{}, errors.New("token response carried no access token")
	}
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - tokenExpiryBuffer)
	return tokenResp.AccessToken, expiry, nil
}

// statusError classifies an MTN answer: 5xx is outcome-unknown (the request
// may have been applied) and wrapped as collection.Uncertain; every other
// status that is not the expected one is an APIError the caller can inspect
// (IsDuplicate, IsNotFound, IsRejected).
func statusError(operation string, status int, body []byte, want int) error {
	if status == want {
		return nil
	}
	apiErr := &APIError{Operation: operation, StatusCode: status, Body: string(body)}
	if status >= http.StatusInternalServerError {
		return collection.Uncertain(apiErr)
	}
	return apiErr
}

// postTransaction handles the common POST pattern for both RequestToPay and Transfer.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *mtnClient) postTransaction(
	ctx context.Context,
	creds *MtnCredentials,
	product, endpoint, logType string,
	referenceID, callbackURL string,
	payload any,
) (retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, logType)
	defer func() { done(retErr) }()
	logger := util.Log(ctx).WithField("type", logType)
	defer logger.Release()

	token, err := c.generateToken(ctx, creds, product)
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := creds.BaseURL() + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Reference-Id", referenceID)
	httpReq.Header.Set("X-Target-Environment", creds.TargetEnvironment)
	httpReq.Header.Set("Ocp-Apim-Subscription-Key", creds.SubscriptionKey)
	if callbackURL != "" {
		httpReq.Header.Set("X-Callback-Url", callbackURL)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return collection.Uncertain(fmt.Errorf("execute request: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return collection.Uncertain(fmt.Errorf("read response: %w", err))
	}
	return statusError(logType, resp.StatusCode, respBody, http.StatusAccepted)
}

// getTransactionStatus handles the common GET status pattern for both collection and disbursement.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *mtnClient) getTransactionStatus(
	ctx context.Context,
	creds *MtnCredentials,
	product, endpoint string,
) (_ []byte, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "get_transaction_status")
	defer func() { done(retErr) }()
	token, err := c.generateToken(ctx, creds, product)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	url := creds.BaseURL() + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Target-Environment", creds.TargetEnvironment)
	req.Header.Set("Ocp-Apim-Subscription-Key", creds.SubscriptionKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, collection.Uncertain(fmt.Errorf("execute request: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, collection.Uncertain(fmt.Errorf("read response body: %w", err))
	}
	if err = statusError("get status", resp.StatusCode, respBody, http.StatusOK); err != nil {
		return nil, err
	}
	return respBody, nil
}

// RequestToPay issues a collection request. req.ReferenceID must be set
// (see ReferenceID); it is the idempotency key MTN deduplicates on.
func (c *mtnClient) RequestToPay(ctx context.Context, creds *MtnCredentials, req *RequestToPayRequest) error {
	if req.ReferenceID == "" {
		return errors.New("requestToPay needs a reference id")
	}

	payload := requestToPayPayload{
		Amount:       req.Amount,
		Currency:     req.Currency,
		ExternalID:   req.ExternalID,
		Payer:        req.Payer,
		PayerMessage: req.PayerMessage,
		PayeeNote:    req.PayeeNote,
	}

	return c.postTransaction(ctx, creds, "collection", "/collection/v1_0/requesttopay",
		"mtn.request_to_pay", req.ReferenceID, req.CallbackURL, payload)
}

func (c *mtnClient) GetRequestToPayStatus(
	ctx context.Context,
	creds *MtnCredentials,
	referenceID string,
) (*RequestToPayStatus, error) {
	endpoint := fmt.Sprintf("/collection/v1_0/requesttopay/%s", referenceID)
	body, err := c.getTransactionStatus(ctx, creds, "collection", endpoint)
	if err != nil {
		return nil, err
	}

	var status RequestToPayStatus
	if err = json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &status, nil
}

// Transfer issues a disbursement. req.ReferenceID must be set (see
// ReferenceID); it is the idempotency key MTN deduplicates on.
func (c *mtnClient) Transfer(ctx context.Context, creds *MtnCredentials, req *TransferRequest) error {
	if req.ReferenceID == "" {
		return errors.New("transfer needs a reference id")
	}

	payload := transferPayload{
		Amount:       req.Amount,
		Currency:     req.Currency,
		ExternalID:   req.ExternalID,
		Payee:        req.Payee,
		PayerMessage: req.PayerMessage,
		PayeeNote:    req.PayeeNote,
	}

	return c.postTransaction(ctx, creds, "disbursement", "/disbursement/v1_0/transfer",
		"mtn.transfer", req.ReferenceID, req.CallbackURL, payload)
}

func (c *mtnClient) GetTransferStatus(
	ctx context.Context,
	creds *MtnCredentials,
	referenceID string,
) (*TransferStatus, error) {
	endpoint := fmt.Sprintf("/disbursement/v1_0/transfer/%s", referenceID)
	body, err := c.getTransactionStatus(ctx, creds, "disbursement", endpoint)
	if err != nil {
		return nil, err
	}

	var status TransferStatus
	if err = json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &status, nil
}
