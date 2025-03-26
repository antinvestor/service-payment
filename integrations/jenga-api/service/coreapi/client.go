package coreapi

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
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

// GenerateSignatureBillGoodsAndServices GenerateSignature generates a RSA signature for the payment request
func (c *Client) GenerateSignatureBillGoodsAndServices(billerCode, amount, billRef, partnerId string) (string, error) {
	// Format message as per Jenga API requirements
	message := fmt.Sprintf("%s%s%s%s", billerCode, amount, billRef, partnerId)
	//log message
	fmt.Println("------------------------------message--------------------------------")
	fmt.Println("****************************"+message+"*******************************")
	
	// Get private key path from environment or config
	privateKeyPath := c.JengaPrivateKey
	if privateKeyPath == "" {
		privateKeyPath = "app/keys/privatekey.pem" // default path
	}

	// Generate signature
	//signature, err := GenerateSignature(message, privateKeyPath)
	signature, err := GenerateSignature(billerCode+amount+billRef+partnerId, "app/keys/privatekey.pem")
	if err != nil {
		return "", fmt.Errorf("failed to generate signature: %v", err)
	}
	//log signature

	return signature, nil
}

// SignData generates a SHA-256 signature with RSA private key
func GenerateSignature(message, privateKeyPath string) (string, error) {
	// Read private key file
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %v", err)
	}

	// Decode PEM format
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return "", fmt.Errorf("failed to decode private key PEM")
	}

	// Parse RSA private key
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse RSA private key: %v", err)
	}

	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("failed to cast parsed key to RSA private key")
	}

	// Compute SHA-256 hash
	hashed := sha256.Sum256([]byte(message))

	// Sign the hash using RSA PKCS1v15
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %v", err)
	}

	// Encode to Base64
	return base64.StdEncoding.EncodeToString(signature), nil
}
func GenerateBalanceSignature(countryCode, accountId string) (string, error) {
	// Generate signature
	signature, err := GenerateSignature(countryCode+accountId, "app/keys/privatekey.pem")
	if err != nil {
		return "", fmt.Errorf("failed to generate signature: %v", err)
	}
	return signature, nil
}

func (c *Client) InitiateAccountBalance(countryCode string, accountId string, accessToken string) (*models.BalanceResponse, error) {
	//https://uat.finserve.africa/v3-apis/account-api/v3.0/accounts/balances/KE/00201XXXX14605
	url := fmt.Sprintf("%s/v3-apis/account-api/v3.0/accounts/balances/%s/%s", c.Env, countryCode, accountId)
	
	//https://uat.finserve.africa/v3-apis/account-api/v3.0/accounts/balances/{countryCode}/{accountId}
    


	// Generate the signature for the request
	signature, err := GenerateBalanceSignature(countryCode, accountId)
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to initiate account balance: %s response Message: %s", resp.Status, resp.Body)
	}
	var balanceResponse models.BalanceResponse

	if err := json.NewDecoder(resp.Body).Decode(&balanceResponse); err != nil {
		return nil, err
	}
	//print balance response
	fmt.Println("------------------------------balance response--------------------------------")
	fmt.Println(balanceResponse)

	return &balanceResponse, nil
}

// InitiateBillGoodsAndServices initiates a bill payment request for goods and services

func (c *Client) InitiateBillGoodsAndServices(request models.PaymentRequest, accessToken string) (*models.PaymentResponse, error) {
	//https://uat.finserve.africa/v3-apis/transaction-api/v3.0/bills/pay
	url := fmt.Sprintf("%s/v3-apis/transaction-api/v3.0/bills/pay", c.Env)
	// Generate the signature for the request
	signature, err := c.GenerateSignatureBillGoodsAndServices(
		request.Biller.BillerCode,
		request.Bill.Amount,
		request.Payer.Reference,
		request.PartnerID,
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to initiate bill payment: %s", resp.Status)
	}
	var paymentResponse models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return nil, err
	}
	return &paymentResponse, nil
}

// GenerateSignature generates a HMAC-SHA256 signature for the STK/USSD push request
func (c *Client) GenerateSignature(accountNumber, ref, mobileNumber, telco, amount, currency string) string {
	data := accountNumber + ref + mobileNumber + telco + amount + currency
	mac := hmac.New(sha256.New, []byte(c.ConsumerSecret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// InitiateSTKUSSD initiates an STK/USSD push request
func (c *Client) InitiateSTKUSSD(request models.STKUSSDRequest, accessToken string) (*models.STKUSSDResponse, error) {
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

	var stkUssdResponse models.STKUSSDResponse
	if err := json.NewDecoder(resp.Body).Decode(&stkUssdResponse); err != nil {
		return nil, err
	}
	return &stkUssdResponse, nil
}

// FetchBillers fetches billers from the Jenga API
func (c *Client) FetchBillers() ([]models.Biller, error) {
	// Generate bearer token
	token, err := c.GenerateBearerToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3-apis/transaction-api/v3.0/billers",c.Env)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch billers: %s", resp.Status)
	}

	var result struct {
		Status bool `json:"status"`
		Data   struct {
			Billers []models.Biller `json:"billers"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Data.Billers, nil
}
