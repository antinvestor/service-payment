package coreapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	models "github.com/antinvestor/jenga-api/service/models"
)

// Client represents the Jenga API client
type Client struct {
	MerchantCode   string
	ConsumerSecret string
	ApiKey         string
	HttpClient     *http.Client
	Env            string
}

// New creates a new instance of the Jenga API client
func New(merchantCode, consumerSecret, apiKey, env string) *Client {
	return &Client{
		MerchantCode:   merchantCode,
		ConsumerSecret: consumerSecret,
		ApiKey:         apiKey,
		HttpClient:     &http.Client{},
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to generate token: %s", resp.Status)
	}

	var tokenResponse BearerTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}
	return &tokenResponse, nil
}


// STKUSSDRequest represents the structure for the STK/USSD push request
type STKUSSDRequest struct {
	Merchant Merchant `json:"merchant"`
	Payment  Payment  `json:"payment"`
}



// STKUSSDResponse represents the response structure for the STK/USSD push initiation
type STKUSSDResponse struct {
	Status        bool   `json:"status"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Reference     string `json:"reference"`
	TransactionID string `json:"transactionId"`
}

// GenerateSignature generates a HMAC-SHA256 signature for the STK/USSD push request
func (c *Client) GenerateSignature(accountNumber, ref, mobileNumber, telco, amount, currency string) string {
	data := accountNumber + ref + mobileNumber + telco + amount + currency
	mac := hmac.New(sha256.New, []byte(c.ConsumerSecret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}



// InitiateSTKUSSD initiates an STK/USSD push request
func (c *Client) InitiateSTKUSSD(request STKUSSDRequest, accessToken string) (*STKUSSDResponse, error) {
	url := fmt.Sprintf("%s/v3-apis/payment-api/v3.0/stkussdpush/initiate", c.Env)

	// Generate the signature for the request
	signature := c.GenerateSignature(
		request.Merchant.AccountNumber,
		request.Payment.Ref,
		request.Payment.MobileNumber,
		request.Payment.Telco,
		request.Payment.Amount,
		request.Payment.Currency,
	)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to initiate STK/USSD push: %s", resp.Status)
	}

	var stkUssdResponse STKUSSDResponse
	if err := json.NewDecoder(resp.Body).Decode(&stkUssdResponse); err != nil {
		return nil, err
	}
	return &stkUssdResponse, nil
}

