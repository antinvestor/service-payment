package jengaApi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
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

// SendMobileRequest represents the request structure for sending money to mobile wallets
type SendMobileRequest struct {
	Source struct {
		CountryCode   string `json:"countryCode"`
		Name          string `json:"name"`
		AccountNumber string `json:"accountNumber"`
	} `json:"source"`
	Destination struct {
		Type         string `json:"type"`
		CountryCode  string `json:"countryCode"`
		Name         string `json:"name"`
		MobileNumber string `json:"mobileNumber"`
		WalletName   string `json:"walletName"`
	} `json:"destination"`
	Transfer struct {
		Type         string `json:"type"`
		Amount       string `json:"amount"`
		CurrencyCode string `json:"currencyCode"`
		Reference    string `json:"reference"`
		Date         string `json:"date"`
		Description  string `json:"description"`
		CallbackURL  string `json:"callbackUrl"`
	} `json:"transfer"`
}

// SendMobileResponse represents the response structure for a mobile wallet transaction
type SendMobileResponse struct {
	Status        bool   `json:"status"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Reference     string `json:"reference"`
	TransactionID string `json:"transactionId"`
}

// SendMobile sends money to a mobile wallet
func (c *Client) SendMobile(req *SendMobileRequest, token string) (*SendMobileResponse, error) {
	url := fmt.Sprintf("%s/v3-apis/transaction-api/v3.0/remittance/sendmobile", c.Env) // Adjust to live URL for production

	// Create the signature
	signature := c.generateSignature(req)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("signature", signature)

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send mobile payment: %s", resp.Status)
	}

	var sendMobileResponse SendMobileResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendMobileResponse); err != nil {
		return nil, err
	}
	return &sendMobileResponse, nil
}

// InternalBankTransferRequest represents the request structure for internal bank transfer
type InternalBankTransferRequest struct {
	Source struct {
		CountryCode   string `json:"countryCode"`
		Name          string `json:"name"`
		AccountNumber string `json:"accountNumber"`
	} `json:"source"`
	Destination struct {
		Type         string `json:"type"`
		CountryCode  string `json:"countryCode"`
		Name         string `json:"name"`
		AccountNumber string `json:"accountNumber"`
	} `json:"destination"`
	Transfer struct {
		Type         string `json:"type"`
		Amount       string `json:"amount"`
		CurrencyCode string `json:"currencyCode"`
		Reference    string `json:"reference"`
		Date         string `json:"date"`
		Description  string `json:"description"`
	} `json:"transfer"`
}

// InternalBankTransferResponse represents the response structure for an internal bank transfer
type InternalBankTransferResponse struct {
	Status        bool   `json:"status"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Reference     string `json:"reference"`
	Data          struct {
		TransactionID string `json:"transactionId"`
		Status        string `json:"status"`
	} `json:"data"`
}

// InternalBankTransfer performs an internal bank transfer within Equity Bank
func (c *Client) InternalBankTransfer(req *InternalBankTransferRequest, token string) (*InternalBankTransferResponse, error) {
	url := fmt.Sprintf("%s/v3-apis/transaction-api/v3.0/remittance/internalBankTransfer", c.Env) // Adjust to live URL for production

	// Create the signature
	signature := c.generateInternalTransferSignature(req)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("signature", signature)

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to perform internal bank transfer: %s", resp.Status)
	}

	var internalBankTransferResponse InternalBankTransferResponse
	if err := json.NewDecoder(resp.Body).Decode(&internalBankTransferResponse); err != nil {
		return nil, err
	}
	return &internalBankTransferResponse, nil
}

// generateSignature generates the required signature for the mobile payment request
func (c *Client) generateSignature(req *SendMobileRequest) string {
	signatureData := fmt.Sprintf("%s%s%s%s",
		req.Transfer.Amount,
		req.Transfer.CurrencyCode,
		req.Transfer.Reference,
		req.Source.AccountNumber,
	)

	h := hmac.New(sha256.New, []byte(c.ConsumerSecret))
	h.Write([]byte(signatureData))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// generateInternalTransferSignature generates the required signature for the internal bank transfer request
func (c *Client) generateInternalTransferSignature(req *InternalBankTransferRequest) string {
	signatureData := fmt.Sprintf("%s%s%s%s",
		req.Transfer.Amount,
		req.Transfer.CurrencyCode,
		req.Transfer.Reference,
		req.Source.AccountNumber,
	)

	h := hmac.New(sha256.New, []byte(c.ConsumerSecret))
	h.Write([]byte(signatureData))
	return fmt.Sprintf("%x", h.Sum(nil))
}
