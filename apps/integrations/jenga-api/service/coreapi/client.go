package coreapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	models "github.com/antinvestor/jenga-api/service/models"
)

// Client represents the Jenga API client.
type Client struct {
	MerchantCode    string
	ConsumerSecret  string
	ApiKey          string       //nolint:staticcheck // API field name
	HttpClient      *http.Client //nolint:staticcheck // API field name
	Env             string
	JengaPrivateKey string
}

// New creates a new instance of the Jenga API client.
func New(merchantCode, consumerSecret, apiKey, env string, jengaPrivateKey string) *Client {
	// Create a custom transport with TLS configuration
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	// Create HTTP client with the custom transport
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	return &Client{
		MerchantCode:   merchantCode,
		ConsumerSecret: consumerSecret,
		ApiKey:         apiKey,
		HttpClient:     httpClient,
		Env:            env,
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
func (c *Client) GenerateBearerToken() (*BearerTokenResponse, error) {
	url := fmt.Sprintf("%s/authentication/api/v3/authenticate/merchant", c.Env)
	body := map[string]string{
		"merchantCode":   c.MerchantCode,
		"consumerSecret": c.ConsumerSecret,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("failed to close response body: %v\n", closeErr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to generate token: %s, body: %s", resp.Status, string(respBody))
	}

	var tokenResponse BearerTokenResponse
	if err := json.Unmarshal(respBody, &tokenResponse); err != nil {
		return nil, err
	}
	return &tokenResponse, nil
}

func (c *Client) GeneratePaymentSignature(args ...string) (string, error) {
	// Generate signature
	// Use the private key path stored in the client configuration
	privateKeyPath := c.JengaPrivateKey
	if privateKeyPath == "" {
		privateKeyPath = "app/keys/privatekey.pem" // Fallback to default path
	}

	signature, err := GenerateSignature(strings.Join(args, ""), privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate signature: %w", err)
	}
	return signature, nil
}

// InitiateSTKUSSD initiates an STK/USSD push request.
func (c *Client) InitiateSTKUSSD(request models.STKUSSDRequest, accessToken string) (*models.STKUSSDResponse, error) {
	url := fmt.Sprintf("%s/v3-apis/payment-api/v3.0/stkussdpush/initiate", c.Env)

	// Generate the signature for the request
	signature, err := c.GeneratePaymentSignature(
		request.Merchant.AccountNumber,
		request.Payment.Ref,
		request.Payment.MobileNumber,
		request.Payment.Telco,
		request.Payment.Amount,
		request.Payment.Currency,
	)
	if err != nil {
		return nil, err
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Signature", signature)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	var stkUssdResponse models.STKUSSDResponse
	if err := json.NewDecoder(resp.Body).Decode(&stkUssdResponse); err != nil {
		return nil, err
	}
	return &stkUssdResponse, nil
}

// CreatePaymentLink creates a payment link using the Jenga API.
func (c *Client) CreatePaymentLink(
	request models.PaymentLinkRequest,
	accessToken string,
) (*models.PaymentLinkResponse, error) {
	// Compose the endpoint URL
	url := fmt.Sprintf("%s/api-checkout/api/v1/create/payment-link", c.Env)

	// Prepare signature fields as per the formula:
	// paymentLink.expiryDate+paymentLink.amount+paymentLink.currency+paymentLink.amountOption+paymentLink.externalRef
	expiryDate := request.PaymentLink.ExpiryDate
	amount := fmt.Sprint(
		request.PaymentLink.Amount,
	) // Convert amount to string for signature generation.request.PaymentLink.Amount
	currency := request.PaymentLink.Currency
	amountOption := request.PaymentLink.AmountOption
	externalRef := request.PaymentLink.ExternalRef

	signature, err := c.GeneratePaymentSignature(
		expiryDate,
		amount,
		currency,
		amountOption,
		externalRef,
	)
	if err != nil {
		return nil, err
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Signature", signature)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("failed to close response body: %v\n", closeErr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymentLinkResponse models.PaymentLinkResponse
	if unmarshalErr := json.Unmarshal(respBody, &paymentLinkResponse); unmarshalErr != nil {
		return nil, fmt.Errorf(
			"failed to parse response: %w (status: %s, body: %s)",
			unmarshalErr,
			resp.Status,
			string(respBody),
		)
	}

	return &paymentLinkResponse, nil
}

// InitiateTillsPay initiates a tills/pay request.
func (c *Client) InitiateTillsPay(
	request models.TillsPayRequest,
	accessToken string,
) (*models.TillsPayResponse, error) {
	url := fmt.Sprintf("%s/v3-apis/transaction-api/v3.0/tills/pay", c.Env)

	// Generate the signature for the request
	//merchant.till+partner.id+payment.amount+payment.currency+payment.ref
	signature, err := c.GeneratePaymentSignature(
		request.Merchant.Till,
		request.Partner.ID,
		request.Payment.Amount,
		request.Payment.Currency,
		request.Payment.Ref,
	)
	fmt.Println("------------------------------signature--------------------------------")
	fmt.Println(signature)
	if err != nil {
		return nil, err
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Signature", signature)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("failed to close response body: %v\n", closeErr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tillsPayResponse models.TillsPayResponse
	if unmarshalErr := json.Unmarshal(respBody, &tillsPayResponse); unmarshalErr != nil {
		return nil, fmt.Errorf(
			"failed to parse response: %w (status: %s, body: %s)",
			unmarshalErr,
			resp.Status,
			string(respBody),
		)
	}

	return &tillsPayResponse, nil
}
