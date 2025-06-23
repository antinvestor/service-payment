package coreapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/antinvestor/jenga-api/service/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateBearerToken(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectError    bool
		expectedToken  *BearerTokenResponse
	}{
		{
			name:           "Success - 200 OK",
			responseStatus: http.StatusOK,
			responseBody:   `{"accessToken":"test-token","refreshToken":"refresh-token","expiresIn":"3600","issuedAt":"2023-01-01T00:00:00Z","tokenType":"Bearer"}`,
			expectError:    false,
			expectedToken: &BearerTokenResponse{
				AccessToken:  "test-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    "3600",
				IssuedAt:     "2023-01-01T00:00:00Z",
				TokenType:    "Bearer",
			},
		},
		{
			name:           "Error - 401 Unauthorized",
			responseStatus: http.StatusUnauthorized,
			responseBody:   `{"error":"Invalid credentials"}`,
			expectError:    true,
			expectedToken:  nil,
		},
		{
			name:           "Error - 500 Server Error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error":"Internal server error"}`,
			expectError:    true,
			expectedToken:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method
				assert.Equal(t, http.MethodPost, r.Method)

				// Check headers
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.NotEmpty(t, r.Header.Get("Api-Key"))

				// Set response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, err := w.Write([]byte(tt.responseBody))
				assert.NoError(t, err)
			}))
			defer server.Close()

			// Create client pointing to test server
			client := &Client{
				MerchantCode:   "TEST_MERCHANT",
				ConsumerSecret: "TEST_SECRET",
				ApiKey:         "TEST_API_KEY",
				HttpClient:     server.Client(),
				Env:            server.URL, // Use test server URL
			}

			// Call the method
			token, err := client.GenerateBearerToken()

			// Check expectations
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestInitiateSTKUSSD(t *testing.T) {
	// Enable test mode to skip actual signature validation
	TestMode = true
	defer func() { TestMode = false }()
	tests := []struct {
		name             string
		request          models.STKUSSDRequest
		responseStatus   int
		responseBody     string
		expectError      bool
		expectedResponse *models.STKUSSDResponse
	}{
		{
			name: "Success - 200 OK",
			request: models.STKUSSDRequest{
				Merchant: models.Merchant{
					AccountNumber: "12345",
					CountryCode:   "KE",
				},
				Payment: models.Payment{
					Ref:          "REF123",
					Amount:       "100",
					Currency:     "KES",
					Telco:        "Safaricom",
					MobileNumber: "254712345678",
				},
			},
			responseStatus: http.StatusOK,
			responseBody:   `{"status":true,"code":200,"message":"Success","reference":"REF123","transactionId":"TRX123"}`,
			expectError:    false,
			expectedResponse: &models.STKUSSDResponse{
				Status:        true,
				Code:          200,
				Message:       "Success",
				Reference:     "REF123",
				TransactionID: "TRX123",
			},
		},
		{
			name: "Error - 400 Bad Request",
			request: models.STKUSSDRequest{
				Merchant: models.Merchant{
					AccountNumber: "12345",
					CountryCode:   "KE",
				},
				Payment: models.Payment{
					Ref:          "REF123",
					Amount:       "100",
					Currency:     "KES",
					Telco:        "Unknown",
					MobileNumber: "254712345678",
				},
			},
			responseStatus: http.StatusBadRequest,
			responseBody:   `{"status":false,"code":400,"message":"Invalid request parameters"}`,
			expectError:    false, // Not expecting an error because we still parse the response
			expectedResponse: &models.STKUSSDResponse{
				Status:  false,
				Code:    400,
				Message: "Invalid request parameters",
			},
		},
	}

	// Create a temporary file for private key testing
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-private-key")
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()
	// Write dummy key content
	_, err = tmpFile.WriteString("-----BEGIN PRIVATE KEY-----\nMIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKNwapOQ6rQJHetP\n-----END PRIVATE KEY-----")
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and headers
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.NotEmpty(t, r.Header.Get("Signature"))

				// Check request body
				var requestBody models.STKUSSDRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.request, requestBody)

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, err = w.Write([]byte(tt.responseBody))
				assert.NoError(t, err)
			}))
			defer server.Close()

			// Create client pointing to test server
			client := &Client{
				MerchantCode:    "TEST_MERCHANT",
				ConsumerSecret:  "TEST_SECRET",
				ApiKey:          "TEST_API_KEY",
				HttpClient:      server.Client(),
				Env:             server.URL, // Use test server URL
				JengaPrivateKey: tmpFile.Name(),
			}

			// Call the method
			response, err := client.InitiateSTKUSSD(tt.request, "test-token")

			// Check expectations
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, response)
			}
		})
	}
}

func TestInitiateAccountBalance(t *testing.T) {
	// Enable test mode to skip actual signature validation
	TestMode = true
	defer func() { TestMode = false }()
	tests := []struct {
		name             string
		countryCode      string
		accountId        string
		responseStatus   int
		responseBody     string
		expectError      bool
		expectedResponse *models.BalanceResponse
	}{
		{
			name:           "Success - 200 OK",
			countryCode:    "KE",
			accountId:      "12345",
			responseStatus: http.StatusOK,
			responseBody:   `{"status":true,"code":200,"message":"Success","data":{"balances":[{"amount":"1000.00","type":"available"}],"currency":"KES"}}`,
			expectError:    false,
			expectedResponse: &models.BalanceResponse{
				Status:  true,
				Code:    200,
				Message: "Success",
				Data: struct {
					Balances []struct {
						Amount string `json:"amount"`
						Type   string `json:"type"`
					} `json:"balances"`
					Currency string `json:"currency"`
				}{
					Balances: []struct {
						Amount string `json:"amount"`
						Type   string `json:"type"`
					}{
						{
							Amount: "1000.00",
							Type:   "available",
						},
					},
					Currency: "KES",
				},
			},
		},
		{
			name:           "Error - 404 Not Found",
			countryCode:    "KE",
			accountId:      "invalid",
			responseStatus: http.StatusNotFound,
			responseBody:   `{"status":false,"code":404,"message":"Account not found"}`,
			expectError:    false, // Not expecting an error because we still parse the response
			expectedResponse: &models.BalanceResponse{
				Status:  false,
				Code:    404,
				Message: "Account not found",
			},
		},
	}

	// Create a temporary file for private key testing
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-private-key")
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()
	// Write dummy key content
	_, err = tmpFile.WriteString("-----BEGIN PRIVATE KEY-----\nMIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKNwapOQ6rQJHetP\n-----END PRIVATE KEY-----")
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and headers
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.NotEmpty(t, r.Header.Get("Signature"))

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, err = w.Write([]byte(tt.responseBody))
				assert.NoError(t, err)
			}))
			defer server.Close()

			// Create client pointing to test server
			client := &Client{
				MerchantCode:    "TEST_MERCHANT",
				ConsumerSecret:  "TEST_SECRET",
				ApiKey:          "TEST_API_KEY",
				HttpClient:      server.Client(),
				Env:             server.URL, // Use test server URL
				JengaPrivateKey: tmpFile.Name(),
			}

			// Call the method
			response, err := client.InitiateAccountBalance(tt.countryCode, tt.accountId, "test-token")

			// Check expectations
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.Status, response.Status)
				assert.Equal(t, tt.expectedResponse.Code, response.Code)
				assert.Equal(t, tt.expectedResponse.Message, response.Message)

				// Only check data if it's not an error response
				if tt.expectedResponse.Status {
					assert.Equal(t, tt.expectedResponse.Data.Currency, response.Data.Currency)
					assert.Equal(t, len(tt.expectedResponse.Data.Balances), len(response.Data.Balances))
					if len(tt.expectedResponse.Data.Balances) > 0 {
						assert.Equal(t, tt.expectedResponse.Data.Balances[0].Amount, response.Data.Balances[0].Amount)
						assert.Equal(t, tt.expectedResponse.Data.Balances[0].Type, response.Data.Balances[0].Type)
					}
				}
			}
		})
	}
}

func TestGenerateSignature(t *testing.T) {
	// Disable test mode for the signature generation tests
	// since these tests specifically validate the signature generation logic
	oldTestMode := TestMode
	TestMode = false
	defer func() { TestMode = oldTestMode }()
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-private-key")
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()

	// Write dummy key content
	_, err = tmpFile.WriteString("-----BEGIN PRIVATE KEY-----\nMIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKNwapOQ6rQJHetP\n-----END PRIVATE KEY-----")
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	tests := []struct {
		name        string
		message     string
		keyPath     string
		expectError bool
	}{
		{
			name:        "Error - Invalid private key file path",
			message:     "test message",
			keyPath:     "nonexistent.pem",
			expectError: true,
		},
		// Note: We can't easily test the success case with actual signature verification
		// since we're using a dummy private key. We can at least test the function doesn't crash
		// with a valid file path
		{
			name:        "File exists but contains invalid key",
			message:     "test message",
			keyPath:     tmpFile.Name(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := GenerateSignature(tt.message, tt.keyPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, signature)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, signature)
			}
		})
	}
}
