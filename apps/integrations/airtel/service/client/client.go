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
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/util"
)

const (
	httpTimeout       = 30 * time.Second
	tokenExpiryBuffer = 60 // seconds to subtract from token expiry
	nanosPerUnit      = 1e9
)

type tokenEntry struct {
	token  string
	expiry time.Time
}

type airtelClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
	tokens     map[string]*tokenEntry // keyed by clientID
	metrics    *integrationobs.Metrics
}

// NewClient creates a new Airtel Money API client.
func NewClient() AirtelClient {
	return &airtelClient{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		tokens:  make(map[string]*tokenEntry),
		metrics: integrationobs.NewMetrics("airtel"),
	}
}

func (c *airtelClient) generateToken(ctx context.Context, creds *AirtelCredentials) (string, error) {
	cacheKey := creds.ClientID

	c.mu.RLock()
	if entry, ok := c.tokens[cacheKey]; ok && time.Now().Before(entry.expiry) {
		token := entry.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.tokens[cacheKey]; ok && time.Now().Before(entry.expiry) {
		return entry.token, nil
	}

	url := creds.BaseURL() + "/auth/oauth2/token"

	payload := map[string]string{
		"client_id":     creds.ClientID,
		"client_secret": creds.ClientSecret,
		"grant_type":    "client_credentials",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp TokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn-tokenExpiryBuffer) * time.Second)

	c.tokens[cacheKey] = &tokenEntry{
		token:  tokenResp.AccessToken,
		expiry: expiry,
	}

	return tokenResp.AccessToken, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *airtelClient) CollectionPush(
	ctx context.Context,
	creds *AirtelCredentials,
	req *CollectionRequest,
) (_ *CollectionResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "collection_push")
	defer func() { done(retErr) }()
	logger := util.Log(ctx).WithField("type", "airtel.collection_push")
	defer logger.Release()

	token, err := c.generateToken(ctx, creds)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

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

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal collection payload: %w", err)
	}

	url := creds.BaseURL() + "/merchant/v2/payments/"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create collection request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Country", req.CountryCode)
	httpReq.Header.Set("X-Currency", req.Currency)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute collection request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	logger.WithFields(map[string]any{
		"status":   resp.StatusCode,
		"response": string(respBody),
	}).Debug("collection response")

	var collResp CollectionResponse
	if err = json.Unmarshal(respBody, &collResp); err != nil {
		return nil, fmt.Errorf("decode collection response: %w", err)
	}

	if !collResp.Status.Success {
		return &collResp, fmt.Errorf("collection failed: %s - %s", collResp.Status.Code, collResp.Status.Message)
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
	logger := util.Log(ctx).WithField("type", "airtel.disburse")
	defer logger.Release()

	token, err := c.generateToken(ctx, creds)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	payload := disbursementPayload{
		Payee: payeePayload{
			Msisdn: req.PhoneNumber,
		},
		Reference: req.Reference,
		Transaction: transactionPayload{
			Amount:   req.Amount,
			Country:  req.CountryCode,
			Currency: req.Currency,
			ID:       req.Reference,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal disbursement payload: %w", err)
	}

	url := creds.BaseURL() + "/standard/v2/disbursements/"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create disbursement request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Country", req.CountryCode)
	httpReq.Header.Set("X-Currency", req.Currency)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute disbursement request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	logger.WithFields(map[string]any{
		"status":   resp.StatusCode,
		"response": string(respBody),
	}).Debug("disbursement response")

	var disbResp DisbursementResponse
	if err = json.Unmarshal(respBody, &disbResp); err != nil {
		return nil, fmt.Errorf("decode disbursement response: %w", err)
	}

	if !disbResp.Status.Success {
		return &disbResp, fmt.Errorf("disbursement failed: %s - %s", disbResp.Status.Code, disbResp.Status.Message)
	}

	return &disbResp, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *airtelClient) TransactionStatus(
	ctx context.Context,
	creds *AirtelCredentials,
	transactionID string,
) (_ *StatusResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "transaction_status")
	defer func() { done(retErr) }()
	token, err := c.generateToken(ctx, creds)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	url := fmt.Sprintf("%s/standard/v1/payments/%s", creds.BaseURL(), transactionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create status request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Country", creds.CountryCode)
	req.Header.Set("X-Currency", creds.Currency)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute status request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status check failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var statusResp StatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("decode status response: %w", err)
	}

	return &statusResp, nil
}
