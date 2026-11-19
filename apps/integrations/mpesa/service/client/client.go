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
	"strings"
	"sync"
	"time"

	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameclient "github.com/pitabwire/frame/v3/client"
	"github.com/pitabwire/util"
)

const (
	httpTimeout      = 30 * time.Second
	tokenCachePeriod = 50 * time.Minute // tokens last ~60min, refresh early
	maxResponseBytes = 1 << 20
)

type tokenEntry struct {
	token  string
	expiry time.Time
}

type client struct {
	httpClient *http.Client
	mu         sync.RWMutex
	tokens     map[string]*tokenEntry // keyed by environment:consumerKey
	metrics    *integrationobs.Metrics
}

// NewClient creates a Daraja client on a frame HTTP client with a fixed
// outbound timeout (set once here, never per call).
func NewClient(ctx context.Context) MpesaClient {
	return NewClientWithHTTP(frameclient.NewHTTPClient(ctx, frameclient.WithHTTPTimeout(httpTimeout)))
}

// NewClientWithHTTP wraps a caller-supplied HTTP client (tests).
func NewClientWithHTTP(httpClient *http.Client) MpesaClient {
	return &client{
		httpClient: httpClient,
		tokens:     make(map[string]*tokenEntry),
		metrics:    integrationobs.NewMetrics("mpesa"),
	}
}

// generateToken returns a cached bearer token, fetching a new one when none
// is valid. The HTTP call runs outside the lock so other credentials are not
// blocked by a slow token endpoint.
func (c *client) generateToken(ctx context.Context, creds *MpesaCredentials) (string, error) {
	cacheKey := creds.Environment + ":" + creds.ConsumerKey

	c.mu.RLock()
	entry, ok := c.tokens[cacheKey]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiry) {
		return entry.token, nil
	}

	token, err := c.fetchToken(ctx, creds)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.tokens[cacheKey] = &tokenEntry{token: token, expiry: time.Now().Add(tokenCachePeriod)}
	c.mu.Unlock()
	return token, nil
}

func (c *client) fetchToken(ctx context.Context, creds *MpesaCredentials) (string, error) {
	url := creds.BaseURL() + "/oauth/v1/generate?grant_type=client_credentials"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	auth := base64.StdEncoding.EncodeToString([]byte(creds.ConsumerKey + ":" + creds.ConsumerSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", collection.Uncertain(fmt.Errorf("execute token request: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", collection.Uncertain(fmt.Errorf("read token response: %w", err))
	}
	if err = statusError("token", resp.StatusCode, body); err != nil {
		return "", err
	}
	var tokenResp TokenResponse
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", errors.New("token response carried no access token")
	}
	return tokenResp.AccessToken, nil
}

// statusError classifies a Daraja HTTP status: 5xx is outcome-unknown (the
// request may have been applied) and wrapped as collection.Uncertain; any
// other non-200 answer is an APIError the worker may fail on.
func statusError(operation string, status int, body []byte) error {
	if status == http.StatusOK {
		return nil
	}
	apiErr := &APIError{Operation: operation, StatusCode: status, Body: string(body)}
	if status >= http.StatusInternalServerError {
		return collection.Uncertain(apiErr)
	}
	return apiErr
}

// doAuthenticatedPost handles the common pattern of: generate token, marshal payload,
// POST to endpoint, read response, check status, and unmarshal into result.
//
//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *client) doAuthenticatedPost(
	ctx context.Context,
	creds *MpesaCredentials,
	endpoint string,
	payload any,
	result any,
	logType string,
) (retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, logType)
	defer func() { done(retErr) }()
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
		return collection.Uncertain(fmt.Errorf("execute %s request: %w", logType, err))
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return collection.Uncertain(fmt.Errorf("read %s response: %w", logType, err))
	}
	logger.WithFields(map[string]any{
		"status":   resp.StatusCode,
		"response": string(respBody),
	}).Debug(logType + " response")

	if err = statusError(logType, resp.StatusCode, respBody); err != nil {
		return err
	}
	if err = json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("decode %s response: %w", logType, err)
	}
	return nil
}

// acceptedError turns a 200 answer whose ResponseCode is not "0" into a
// rejection: Daraja did not queue the request.
func acceptedError(operation, responseCode, description string) error {
	if code := strings.TrimSpace(responseCode); code != "" && code != "0" {
		return &APIError{Operation: operation, StatusCode: http.StatusOK, ResponseCode: code, Body: description}
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
	if err = acceptedError("stk push", stkResp.ResponseCode, stkResp.ResponseDescription); err != nil {
		return nil, err
	}
	return &stkResp, nil
}

func (c *client) STKPushQuery(
	ctx context.Context,
	creds *MpesaCredentials,
	req *STKQueryRequest,
) (*STKQueryResponse, error) {
	var queryResp STKQueryResponse
	err := c.doAuthenticatedPost(
		ctx, creds, "/mpesa/stkpushquery/v1/query", req, &queryResp, "stk push query",
	)
	if err != nil {
		return nil, err
	}
	return &queryResp, nil
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
	if err := acceptedError("b2c", b2cResp.ResponseCode, b2cResp.ResponseDescription); err != nil {
		return nil, err
	}
	return &b2cResp, nil
}

// TransactionStatus submits a Transaction Status query
// (POST /mpesa/transactionstatus/v1/query). Daraja only acknowledges it;
// the status itself is posted to req.ResultURL later.
func (c *client) TransactionStatus(
	ctx context.Context,
	creds *MpesaCredentials,
	req *TransactionStatusRequest,
) (*TransactionStatusResponse, error) {
	var resp TransactionStatusResponse
	err := c.doAuthenticatedPost(ctx, creds, "/mpesa/transactionstatus/v1/query", req, &resp, "transaction status")
	if err != nil {
		return nil, err
	}
	if err = acceptedError("transaction status", string(resp.ResponseCode), resp.ResponseDescription); err != nil {
		return nil, err
	}
	return &resp, nil
}
