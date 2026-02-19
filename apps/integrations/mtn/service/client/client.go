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

	"github.com/google/uuid"
	"github.com/pitabwire/util"
)

type tokenEntry struct {
	token  string
	expiry time.Time
}

const (
	httpClientTimeout = 30 * time.Second
	tokenExpiryBuffer = 60 // seconds before actual expiry to refresh token
)

type mtnClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
	tokens     map[string]*tokenEntry // keyed by product:apiUser
}

// NewClient creates a new MTN MoMo API client.
func NewClient() MtnClient {
	return &mtnClient{
		httpClient: &http.Client{
			Timeout: httpClientTimeout,
		},
		tokens: make(map[string]*tokenEntry),
	}
}

func (c *mtnClient) generateToken(ctx context.Context, creds *MtnCredentials, product string) (string, error) {
	cacheKey := product + ":" + creds.APIUser

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

	url := fmt.Sprintf("%s/%s/token/", creds.BaseURL(), product)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(creds.APIUser + ":" + creds.APIKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Ocp-Apim-Subscription-Key", creds.SubscriptionKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
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

// postTransaction handles the common POST pattern for both RequestToPay and Transfer.
func (c *mtnClient) postTransaction(
	ctx context.Context,
	creds *MtnCredentials,
	product, endpoint, logType string,
	referenceID, callbackURL string,
	payload any,
) error {
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
	httpReq.Header.Set("X-Target-Environment", creds.Environment)
	httpReq.Header.Set("Ocp-Apim-Subscription-Key", creds.SubscriptionKey)
	if callbackURL != "" {
		httpReq.Header.Set("X-Callback-Url", callbackURL)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s failed with status %d: %s", logType, resp.StatusCode, string(respBody))
	}

	return nil
}

// getTransactionStatus handles the common GET status pattern for both collection and disbursement.
func (c *mtnClient) getTransactionStatus(
	ctx context.Context,
	creds *MtnCredentials,
	product, endpoint string,
) ([]byte, error) {
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
	req.Header.Set("X-Target-Environment", creds.Environment)
	req.Header.Set("Ocp-Apim-Subscription-Key", creds.SubscriptionKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get status failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return respBody, nil
}

func (c *mtnClient) RequestToPay(ctx context.Context, creds *MtnCredentials, req *RequestToPayRequest) error {
	if req.ReferenceID == "" {
		req.ReferenceID = uuid.NewString()
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

func (c *mtnClient) Transfer(ctx context.Context, creds *MtnCredentials, req *TransferRequest) error {
	if req.ReferenceID == "" {
		req.ReferenceID = uuid.NewString()
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
