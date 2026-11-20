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

package coreapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCreds(serverURL string) *coreapi.Credentials {
	return &coreapi.Credentials{
		MerchantCode:   "TEST_MERCHANT",
		ConsumerSecret: "TEST_SECRET",
		APIKey:         "TEST_API_KEY",
		BaseURL:        serverURL,
	}
}

func TestGenerateBearerToken(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectError    bool
		expectedToken  *coreapi.BearerTokenResponse
	}{
		{
			name:           "Success - 200 OK",
			responseStatus: http.StatusOK,
			responseBody:   `{"accessToken":"test-token","refreshToken":"refresh-token","expiresIn":"3600","issuedAt":"2023-01-01T00:00:00Z","tokenType":"Bearer"}`,
			expectError:    false,
			expectedToken: &coreapi.BearerTokenResponse{
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.NotEmpty(t, r.Header.Get("Api-Key"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, err := w.Write([]byte(tt.responseBody))
				assert.NoError(t, err)
			}))
			defer server.Close()
			client := coreapi.New(server.Client())
			creds := testCreds(server.URL)

			token, err := client.GenerateBearerToken(context.Background(), creds)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, token)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestInitiateSTKUSSD(t *testing.T) {
	coreapi.SetTestMode(true)
	defer coreapi.SetTestMode(false)

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
			name: "Error - 400 Bad Request returns error",
			request: models.STKUSSDRequest{
				Merchant: models.Merchant{AccountNumber: "12345", CountryCode: "KE"},
				Payment: models.Payment{
					Ref: "REF123", Amount: "100", Currency: "KES",
					Telco: "Unknown", MobileNumber: "254712345678",
				},
			},
			responseStatus:   http.StatusBadRequest,
			responseBody:     `{"status":false,"code":400,"message":"Invalid request parameters"}`,
			expectError:      true,
			expectedResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.NotEmpty(t, r.Header.Get("Signature"))

				var requestBody models.STKUSSDRequest
				decodeErr := json.NewDecoder(r.Body).Decode(&requestBody)
				assert.NoError(t, decodeErr)
				assert.Equal(t, tt.request, requestBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, writeErr := w.Write([]byte(tt.responseBody))
				assert.NoError(t, writeErr)
			}))
			defer server.Close()
			client := coreapi.New(server.Client())
			creds := testCreds(server.URL)

			response, respErr := client.InitiateSTKUSSD(context.Background(), creds, tt.request, "test-token")

			if tt.expectError {
				require.Error(t, respErr)
				require.Nil(t, response)
			} else {
				require.NoError(t, respErr)
				require.Equal(t, tt.expectedResponse, response)
			}
		})
	}
}

func TestGenerateSignature(t *testing.T) {
	coreapi.SetTestMode(false)
	defer coreapi.SetTestMode(false)

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-private-key")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(
		"-----BEGIN PRIVATE KEY-----\nMIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKNwapOQ6rQJHetP\n-----END PRIVATE KEY-----",
	)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

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
		{
			name:        "File exists but contains invalid key",
			message:     "test message",
			keyPath:     tmpFile.Name(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, sigErr := coreapi.GenerateSignature(tt.message, tt.keyPath)

			if tt.expectError {
				require.Error(t, sigErr)
				require.Empty(t, signature)
			} else {
				require.NoError(t, sigErr)
				require.NotEmpty(t, signature)
			}
		})
	}
}

func TestOutcomeClassificationAndTokenCache(t *testing.T) {
	coreapi.SetTestMode(true)
	defer coreapi.SetTestMode(false)

	tokenCalls := 0
	stkStatus := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/authentication/api/v3/authenticate/merchant" {
			tokenCalls++
			_, _ = w.Write([]byte(`{"accessToken":"tok","expiresIn":"3600"}`))
			return
		}
		w.WriteHeader(stkStatus)
		_, _ = w.Write([]byte(`{"status":false,"code":500,"message":"upstream"}`))
	}))
	defer server.Close()
	client := coreapi.New(server.Client())
	creds := testCreds(server.URL)

	for range 3 {
		tok, err := client.GenerateBearerToken(context.Background(), creds)
		require.NoError(t, err)
		assert.Equal(t, "tok", tok.AccessToken)
	}
	assert.Equal(t, 1, tokenCalls, "token is cached per credentials")

	other := testCreds(server.URL)
	other.APIKey = "OTHER"
	_, err := client.GenerateBearerToken(context.Background(), other)
	require.NoError(t, err)
	assert.Equal(t, 2, tokenCalls, "different credentials get their own token")

	req := models.STKUSSDRequest{Payment: models.Payment{Ref: "R", Amount: "1", Currency: "KES"}}
	stkStatus = http.StatusInternalServerError
	_, err = client.InitiateSTKUSSD(context.Background(), creds, req, "tok")
	require.Error(t, err)
	assert.True(t, collection.IsUncertain(err), "5xx is outcome-unknown")
	assert.False(t, coreapi.IsRejected(err) && !collection.IsUncertain(err))

	stkStatus = http.StatusBadRequest
	_, err = client.InitiateSTKUSSD(context.Background(), creds, req, "tok")
	require.Error(t, err)
	assert.True(t, coreapi.IsRejected(err), "4xx is a rejection")
	assert.False(t, collection.IsUncertain(err))

	server.Close()
	_, err = client.InitiateSTKUSSD(context.Background(), creds, req, "tok")
	require.Error(t, err)
	assert.True(t, collection.IsUncertain(err), "transport failure is outcome-unknown")
}
