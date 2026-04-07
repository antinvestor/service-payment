package coreapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	models "github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/util"
)

// Client represents the Jenga API client.
type Client struct {
	MerchantCode    string
	ConsumerSecret  string
	APIKey          string
	HTTPClient      *http.Client
	Env             string
	JengaPrivateKey string
}

// New creates a new instance of the Jenga API client.
// The httpClient must be provided by the caller (e.g. from Frame's HTTPClientManager).
func New(httpClient *http.Client, merchantCode, consumerSecret, apiKey, env, privateKeyPath string) *Client {
	return &Client{
		MerchantCode:    merchantCode,
		ConsumerSecret:  consumerSecret,
		APIKey:          apiKey,
		HTTPClient:      httpClient,
		Env:             env,
		JengaPrivateKey: privateKeyPath,
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

// GenerateBearerToken generates a Bearer token for authorization.
func (c *Client) GenerateBearerToken(ctx context.Context) (*BearerTokenResponse, error) {
	logger := util.Log(ctx).WithField("operation", "GenerateBearerToken")

	url := fmt.Sprintf("%s/authentication/api/v3/authenticate/merchant", c.Env)
	body := map[string]string{
		"merchantCode":   c.MerchantCode,
		"consumerSecret": c.ConsumerSecret,
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
	req.Header.Set("Api-Key", c.APIKey)

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

	logger.Debug("bearer token generated successfully")
	return &tokenResponse, nil
}

func (c *Client) GeneratePaymentSignature(args ...string) (string, error) {
	privateKeyPath := c.JengaPrivateKey
	if privateKeyPath == "" {
		privateKeyPath = "/app/keys/privatekey.pem"
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

	url := fmt.Sprintf("%s/v3-apis/payment-api/v3.0/stkussdpush/initiate", c.Env)

	signature, err := c.GeneratePaymentSignature(
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

	var stkUssdResponse models.STKUSSDResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&stkUssdResponse); decodeErr != nil {
		return nil, fmt.Errorf("decode STK response: %w", decodeErr)
	}

	logger.WithFields(map[string]any{
		"response_status": stkUssdResponse.Status,
		"response_code":   stkUssdResponse.Code,
		"transaction_id":  stkUssdResponse.TransactionID,
	}).Info("STK/USSD push initiated")

	return &stkUssdResponse, nil
}

// CreatePaymentLink creates a payment link using the Jenga API.
func (c *Client) CreatePaymentLink(
	ctx context.Context,
	request models.PaymentLinkRequest,
	accessToken string,
) (*models.PaymentLinkResponse, error) {
	logger := util.Log(ctx).WithFields(map[string]any{
		"operation":    "CreatePaymentLink",
		"external_ref": request.PaymentLink.ExternalRef,
		"amount":       request.PaymentLink.Amount,
		"currency":     request.PaymentLink.Currency,
	})

	url := fmt.Sprintf("%s/api-checkout/api/v1/create/payment-link", c.Env)

	signature, err := c.GeneratePaymentSignature(
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

	var paymentLinkResponse models.PaymentLinkResponse
	if unmarshalErr := json.Unmarshal(respBody, &paymentLinkResponse); unmarshalErr != nil {
		logger.WithFields(map[string]any{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("failed to parse payment link response")
		return nil, fmt.Errorf(
			"parse payment link response: %w (status: %s, body: %s)",
			unmarshalErr, resp.Status, string(respBody),
		)
	}

	logger.WithFields(map[string]any{
		"response_status": paymentLinkResponse.Status,
		"response_code":   paymentLinkResponse.Code,
	}).Info("payment link creation completed")

	return &paymentLinkResponse, nil
}

// InitiateTillsPay initiates a tills/pay request.
func (c *Client) InitiateTillsPay(
	ctx context.Context,
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

	url := fmt.Sprintf("%s/v3-apis/transaction-api/v3.0/tills/pay", c.Env)

	signature, err := c.GeneratePaymentSignature(
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

	var tillsPayResponse models.TillsPayResponse
	if unmarshalErr := json.Unmarshal(respBody, &tillsPayResponse); unmarshalErr != nil {
		logger.WithFields(map[string]any{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("failed to parse tills pay response")
		return nil, fmt.Errorf(
			"parse tills pay response: %w (status: %s, body: %s)",
			unmarshalErr, resp.Status, string(respBody),
		)
	}

	logger.WithFields(map[string]any{
		"response_status": tillsPayResponse.Status,
		"transaction_id":  tillsPayResponse.TransactionID,
		"merchant_name":   tillsPayResponse.MerchantName,
	}).Info("tills pay completed")

	return &tillsPayResponse, nil
}
