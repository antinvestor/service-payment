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
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pitabwire/util"
)

const (
	httpTimeout      = 30 * time.Second
	tokenCachePeriod = 50 * time.Minute // tokens last ~60min, refresh early
)

type tokenEntry struct {
	token  string
	expiry time.Time
}

type client struct {
	httpClient *http.Client
	mu         sync.RWMutex
	tokens     map[string]*tokenEntry // keyed by consumerKey
}

// NewClient creates a new M-Pesa Daraja API client.
func NewClient() MpesaClient {
	return &client{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		tokens: make(map[string]*tokenEntry),
	}
}

func (c *client) generateToken(ctx context.Context, creds *MpesaCredentials) (string, error) {
	cacheKey := creds.ConsumerKey

	c.mu.RLock()
	if entry, ok := c.tokens[cacheKey]; ok && time.Now().Before(entry.expiry) {
		token := entry.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, ok := c.tokens[cacheKey]; ok && time.Now().Before(entry.expiry) {
		return entry.token, nil
	}

	url := creds.BaseURL() + "/oauth/v1/generate?grant_type=client_credentials"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(creds.ConsumerKey + ":" + creds.ConsumerSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	c.tokens[cacheKey] = &tokenEntry{
		token:  tokenResp.AccessToken,
		expiry: time.Now().Add(tokenCachePeriod),
	}

	return tokenResp.AccessToken, nil
}

// doAuthenticatedPost handles the common pattern of: generate token, marshal payload,
// POST to endpoint, read response, check status, and unmarshal into result.
func (c *client) doAuthenticatedPost(
	ctx context.Context,
	creds *MpesaCredentials,
	endpoint string,
	payload any,
	result any,
	logType string,
) error {
	logger := util.Log(ctx).WithField("type", logType)
	defer logger.Release()

	token, err := c.generateToken(ctx, creds)
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s payload: %w", logType, err)
	}

	apiURL := creds.BaseURL() + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create %s request: %w", logType, err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute %s request: %w", logType, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.WithFields(map[string]any{
		"status":   resp.StatusCode,
		"response": string(respBody),
	}).Debug(logType + " response")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s failed with status %d: %s", logType, resp.StatusCode, string(respBody))
	}

	if err = json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("decode %s response: %w", logType, err)
	}

	return nil
}

func (c *client) STKPush(ctx context.Context, creds *MpesaCredentials, req *STKPushRequest) (*STKPushResponse, error) {
	payload := stkPushPayload{
		BusinessShortCode: req.BusinessShortCode,
		Password:          req.Password,
		Timestamp:         req.Timestamp,
		TransactionType:   req.TransactionType,
		Amount:            req.Amount,
		PartyA:            req.PartyA,
		PartyB:            req.PartyB,
		PhoneNumber:       req.PhoneNumber,
		CallBackURL:       req.CallBackURL,
		AccountReference:  req.AccountReference,
		TransactionDesc:   req.TransactionDesc,
	}

	var stkResp STKPushResponse
	err := c.doAuthenticatedPost(
		ctx, creds, "/mpesa/stkpush/v1/processrequest", payload, &stkResp, "stk push",
	)
	if err != nil {
		return nil, err
	}

	return &stkResp, nil
}

func (c *client) B2CPayment(ctx context.Context, creds *MpesaCredentials, req *B2CRequest) (*B2CResponse, error) {
	payload := b2cPayload{
		OriginatorConversationID: req.OriginatorConversationID,
		InitiatorName:            req.InitiatorName,
		SecurityCredential:       req.SecurityCredential,
		CommandID:                req.CommandID,
		Amount:                   req.Amount,
		PartyA:                   req.PartyA,
		PartyB:                   req.PartyB,
		Remarks:                  req.Remarks,
		QueueTimeOutURL:          req.QueueTimeOutURL,
		ResultURL:                req.ResultURL,
		Occasion:                 req.Occasion,
	}

	var b2cResp B2CResponse
	if err := c.doAuthenticatedPost(ctx, creds, "/mpesa/b2c/v1/paymentrequest", payload, &b2cResp, "b2c"); err != nil {
		return nil, err
	}

	return &b2cResp, nil
}
