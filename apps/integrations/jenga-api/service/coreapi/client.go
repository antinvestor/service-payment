package coreapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	models "github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/util"
)

const tokenCacheTTL = 50 * time.Minute // Jenga tokens last ~60min, refresh early

type tokenEntry struct {
	token  string
	expiry time.Time
}

// Client represents the Jenga API client with per-tenant token caching.
type Client struct {
	HTTPClient *http.Client
	mu         sync.RWMutex
	tokens     map[string]*tokenEntry // keyed by MerchantCode
}

// New creates a new instance of the Jenga API client.
// The httpClient must be provided by the caller (e.g. from Frame's HTTPClientManager).
func New(httpClient *http.Client) *Client {
	return &Client{
		HTTPClient: httpClient,
		tokens:     make(map[string]*tokenEntry),
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

// GenerateBearerToken generates a Bearer token using the provided credentials.
// Tokens are cached per MerchantCode with a 50-minute TTL to avoid redundant auth calls.
func (c *Client) GenerateBearerToken(
	ctx context.Context,
	creds *Credentials,
) (*BearerTokenResponse, error) {
	cacheKey := creds.MerchantCode

	// Fast path: check cache under read lock
	c.mu.RLock()
	if entry, ok := c.tokens[cacheKey]; ok && time.Now().Before(entry.expiry) {
		token := entry.token
		c.mu.RUnlock()
		return &BearerTokenResponse{AccessToken: token}, nil
	}
	c.mu.RUnlock()

	// Slow path: acquire write lock and double-check
	c.mu.Lock()
	if entry, ok := c.tokens[cacheKey]; ok && time.Now().Before(entry.expiry) {
		token := entry.token
		c.mu.Unlock()
		return &BearerTokenResponse{AccessToken: token}, nil
	}
	c.mu.Unlock()

	logger := util.Log(ctx).WithField("operation", "GenerateBearerToken")

	url := fmt.Sprintf("%s/authentication/api/v3/authenticate/merchant", creds.Environment)
	body := map[string]string{
		"merchantCode":   creds.MerchantCode,
		"consumerSecret": creds.ConsumerSecret,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", creds.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute token request: %w", err)
	}
	defer util.CloseAndLogOnError(ctx, resp.Body, "close token response body")

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(map[string]any{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("bearer token generation failed")
		return nil, fmt.Errorf("generate token: status %s, body: %s", resp.Status, string(respBody))
	}

	var tokenResponse BearerTokenResponse
	if unmarshalErr := json.Unmarshal(respBody, &tokenResponse); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal token response: %w", unmarshalErr)
	}

	// Cache the token
	c.mu.Lock()
	c.tokens[cacheKey] = &tokenEntry{
		token:  tokenResponse.AccessToken,
		expiry: time.Now().Add(tokenCacheTTL),
	}
	c.mu.Unlock()

	logger.Debug("bearer token generated and cached")
	return &tokenResponse, nil
}

func generatePaymentSignature(privateKeyPath string, args ...string) (string, error) {
	if privateKeyPath == "" {
		privateKeyPath = "/keys/privatekey.pem"
	}

	signature, err := GenerateSignature(strings.Join(args, ""), privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("generate signature: %w", err)
	}
	return signature, nil
}

// InitiateSTKUSSD initiates an STK/USSD push request.
func (c *Client) InitiateSTKUSSD(
	ctx context.Context,
	creds *Credentials,
	request models.STKUSSDRequest,
	accessToken string,
) (*models.STKUSSDResponse, error) {
	logger := util.Log(ctx).WithFields(map[string]any{
		"operation":   "InitiateSTKUSSD",
		"payment_ref": request.Payment.Ref,
		"mobile":      request.Payment.MobileNumber,
		"amount":      request.Payment.Amount,
		"currency":    request.Payment.Currency,
	})

	url := fmt.Sprintf("%s/v3-apis/payment-api/v3.0/stkussdpush/initiate", creds.Environment)

	signature, err := generatePaymentSignature(
		creds.PrivateKeyPath,
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

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal STK request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create STK request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Signature", signature)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute STK request: %w", err)
	}
	defer util.CloseAndLogOnError(ctx, resp.Body, "close STK response body")

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read STK response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(map[string]any{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("STK/USSD push request failed")
		return nil, fmt.Errorf(
			"STK push failed: status %s, body: %s",
			resp.Status,
			string(respBody),
		)
	}

	var stkUssdResponse models.STKUSSDResponse
	if decodeErr := json.Unmarshal(respBody, &stkUssdResponse); decodeErr != nil {
		return nil, fmt.Errorf("decode STK response: %w", decodeErr)
	}

	if !stkUssdResponse.Status {
		return nil, fmt.Errorf(
			"STK push rejected: code=%d message=%s",
			stkUssdResponse.Code,
			stkUssdResponse.Message,
		)
	}

	logger.WithFields(map[string]any{
		"transaction_id": stkUssdResponse.TransactionID,
		"response_code":  stkUssdResponse.Code,
	}).Info("STK/USSD push initiated")

	return &stkUssdResponse, nil
}

// CreatePaymentLink creates a payment link using the Jenga API.
func (c *Client) CreatePaymentLink(
	ctx context.Context,
	creds *Credentials,
	request models.PaymentLinkRequest,
	accessToken string,
) (*models.PaymentLinkResponse, error) {
	logger := util.Log(ctx).WithFields(map[string]any{
		"operation":    "CreatePaymentLink",
		"external_ref": request.PaymentLink.ExternalRef,
		"amount":       request.PaymentLink.Amount,
		"currency":     request.PaymentLink.Currency,
	})

	url := fmt.Sprintf("%s/api-checkout/api/v1/create/payment-link", creds.Environment)

	signature, err := generatePaymentSignature(
		creds.PrivateKeyPath,
		request.PaymentLink.ExpiryDate,
		fmt.Sprint(request.PaymentLink.Amount),
		request.PaymentLink.Currency,
		request.PaymentLink.AmountOption,
		request.PaymentLink.ExternalRef,
	)
	if err != nil {
		return nil, fmt.Errorf("generate payment link signature: %w", err)
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal payment link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create payment link request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Signature", signature)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute payment link request: %w", err)
	}
	defer util.CloseAndLogOnError(ctx, resp.Body, "close payment link response body")

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read payment link response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(map[string]any{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("payment link creation request failed")
		return nil, fmt.Errorf(
			"payment link failed: status %s, body: %s",
			resp.Status,
			string(respBody),
		)
	}

	var paymentLinkResponse models.PaymentLinkResponse
	if unmarshalErr := json.Unmarshal(respBody, &paymentLinkResponse); unmarshalErr != nil {
		return nil, fmt.Errorf(
			"parse payment link response: %w (status: %s, body: %s)",
			unmarshalErr, resp.Status, string(respBody),
		)
	}

	if !paymentLinkResponse.Status {
		return nil, fmt.Errorf(
			"payment link rejected: code=%d message=%s",
			paymentLinkResponse.Code,
			paymentLinkResponse.Message,
		)
	}

	logger.WithField("response_code", paymentLinkResponse.Code).Info("payment link created")

	return &paymentLinkResponse, nil
}

// InitiateTillsPay initiates a tills/pay request.
func (c *Client) InitiateTillsPay(
	ctx context.Context,
	creds *Credentials,
	request models.TillsPayRequest,
	accessToken string,
) (*models.TillsPayResponse, error) {
	logger := util.Log(ctx).WithFields(map[string]any{
		"operation":   "InitiateTillsPay",
		"till":        request.Merchant.Till,
		"payment_ref": request.Payment.Ref,
		"amount":      request.Payment.Amount,
		"currency":    request.Payment.Currency,
	})

	url := fmt.Sprintf("%s/v3-apis/transaction-api/v3.0/tills/pay", creds.Environment)

	signature, err := generatePaymentSignature(
		creds.PrivateKeyPath,
		request.Merchant.Till,
		request.Partner.ID,
		request.Payment.Amount,
		request.Payment.Currency,
		request.Payment.Ref,
	)
	if err != nil {
		return nil, fmt.Errorf("generate tills pay signature: %w", err)
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal tills pay request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create tills pay request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Signature", signature)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute tills pay request: %w", err)
	}
	defer util.CloseAndLogOnError(ctx, resp.Body, "close tills pay response body")

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tills pay response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(map[string]any{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("tills pay request failed")
		return nil, fmt.Errorf(
			"tills pay failed: status %s, body: %s",
			resp.Status,
			string(respBody),
		)
	}

	var tillsPayResponse models.TillsPayResponse
	if unmarshalErr := json.Unmarshal(respBody, &tillsPayResponse); unmarshalErr != nil {
		return nil, fmt.Errorf(
			"parse tills pay response: %w (status: %s, body: %s)",
			unmarshalErr, resp.Status, string(respBody),
		)
	}

	if !tillsPayResponse.Status {
		return nil, fmt.Errorf(
			"tills pay rejected: code=%d message=%s",
			tillsPayResponse.Code,
			tillsPayResponse.Message,
		)
	}

	logger.WithFields(map[string]any{
		"transaction_id": tillsPayResponse.TransactionID,
		"merchant_name":  tillsPayResponse.MerchantName,
	}).Info("tills pay completed")

	return &tillsPayResponse, nil
}
