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

// Client represents the Jenga API client
type Client struct {
	MerchantCode    string
	ConsumerSecret  string
	ApiKey          string
	HttpClient      *http.Client
	Env             string
	JengaPrivateKey string
}

// New creates a new instance of the Jenga API client
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

// BearerTokenResponse represents the response structure for bearer token generation
type BearerTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	IssuedAt     string `json:"issuedAt"`
	TokenType    string `json:"tokenType"`
}

// GenerateBearerToken generates a Bearer token for authorization
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
	defer resp.Body.Close()

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


func (c *Client)GeneratePaymentSignature(args ...string) (string, error) {
	// Generate signature
	// Use the private key path stored in the client configuration
	privateKeyPath := c.JengaPrivateKey
	if privateKeyPath == "" {
		privateKeyPath = "app/keys/privatekey.pem" // Fallback to default path
	}

	signature, err := GenerateSignature(strings.Join(args, ""), privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate signature: %v", err)
	}
	return signature, nil
}

// InitiateSTKUSSD initiates an STK/USSD push request
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
	defer resp.Body.Close()

	var stkUssdResponse models.STKUSSDResponse
	if err := json.NewDecoder(resp.Body).Decode(&stkUssdResponse); err != nil {
		return nil, err
	}
	return &stkUssdResponse, nil
}

func (c *Client) InitiateAccountBalance(countryCode string, accountId string, accessToken string) (*models.BalanceResponse, error) {
	//https://uat.finserve.africa/v3-apis/account-api/v3.0/accounts/balances/KE/00201XXXX14605
	url := fmt.Sprintf("%s/v3-apis/account-api/v3.0/accounts/balances/%s/%s", c.Env, countryCode, accountId)
	
	// Generate the signature for the request
	signature, err := GenerateBalanceSignature(countryCode, accountId, c.JengaPrivateKey)
	if err != nil {
		return nil, err
	}
	//print signature
	fmt.Println("------------------------------signature--------------------------------")
	fmt.Println(signature)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
	defer resp.Body.Close()
	// Read the response body for all status codes
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	// Always parse the response, even for error status codes
	var balanceResponse models.BalanceResponse
	if err := json.Unmarshal(respBodyBytes, &balanceResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v (status: %s, body: %s)", err, resp.Status, string(respBodyBytes))
	}
	//print balance response
	fmt.Println("------------------------------balance response--------------------------------")
	fmt.Println(balanceResponse)

	return &balanceResponse, nil
}

