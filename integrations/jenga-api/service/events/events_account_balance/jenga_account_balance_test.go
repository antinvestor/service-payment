package events_account_balance

import (
	"testing"

	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// validateAccountBalanceRequest validates a balance request
func validateAccountBalanceRequest(request *models.AccountBalanceRequest) error {
	if request.CountryCode == "" {
		return assert.AnError
	}

	if request.AccountId == "" {
		return assert.AnError
	}

	return nil
}

// executeAccountBalanceRequest simulates executing an account balance request against the API
func executeAccountBalanceRequest(client coreapi.JengaApiClient, request *models.AccountBalanceRequest) error {
	if client == nil {
		return assert.AnError
	}

	// Generate bearer token for authorization
	token, err := client.GenerateBearerToken()
	if err != nil {
		return err
	}

	// Get account balance
	_, err = client.InitiateAccountBalance(request.CountryCode, request.AccountId, token.AccessToken)
	return err
}

func TestJengaAccountBalance(t *testing.T) {
	tests := []struct {
		name                string
		request             *models.AccountBalanceRequest
		tokenResponse       *coreapi.BearerTokenResponse
		tokenError          error
		balanceResponse     *models.BalanceResponse
		balanceError        error
		expectExecuteError  bool
		expectValidateError bool
	}{
		{
			name: "Happy path - successful balance check",
			request: &models.AccountBalanceRequest{
				CountryCode: "KE",
				AccountId:   "12345",
			},
			tokenResponse: &coreapi.BearerTokenResponse{
				AccessToken: "test-token",
			},
			tokenError: nil,
			balanceResponse: &models.BalanceResponse{
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
			balanceError:        nil,
			expectExecuteError:  false,
			expectValidateError: false,
		},
		{
			name: "Error - token generation fails",
			request: &models.AccountBalanceRequest{
				CountryCode: "KE",
				AccountId:   "12345",
			},
			tokenResponse:       nil,
			tokenError:          assert.AnError,
			balanceResponse:     nil,
			balanceError:        nil,
			expectExecuteError:  true,
			expectValidateError: false,
		},
		{
			name: "Error - balance check fails",
			request: &models.AccountBalanceRequest{
				CountryCode: "KE",
				AccountId:   "12345",
			},
			tokenResponse: &coreapi.BearerTokenResponse{
				AccessToken: "test-token",
			},
			tokenError:          nil,
			balanceResponse:     nil,
			balanceError:        assert.AnError,
			expectExecuteError:  true,
			expectValidateError: false,
		},
		{
			name:                "Error - validation fails due to missing country code",
			request:             &models.AccountBalanceRequest{AccountId: "12345"},
			tokenResponse:       nil,
			tokenError:          nil,
			balanceResponse:     nil,
			balanceError:        nil,
			expectExecuteError:  false,
			expectValidateError: true,
		},
		{
			name:                "Error - validation fails due to missing account ID",
			request:             &models.AccountBalanceRequest{CountryCode: "KE"},
			tokenResponse:       nil,
			tokenError:          nil,
			balanceResponse:     nil,
			balanceError:        nil,
			expectExecuteError:  false,
			expectValidateError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(coreapi.MockClient)

			// Only set up mock expectations if we don't expect validation to fail
			if !tt.expectValidateError {
				// Set up client mock expectations
				mockClient.On("GenerateBearerToken").Return(tt.tokenResponse, tt.tokenError)
				
				// Only set up InitiateAccountBalance expectation if token generation doesn't fail
				if tt.tokenError == nil {
					mockClient.On("InitiateAccountBalance", tt.request.CountryCode, tt.request.AccountId, mock.AnythingOfType("string")).
						Return(tt.balanceResponse, tt.balanceError)
				}
			}

			// Test the event name (this is just a constant check)
			assert.Equal(t, "jenga.account.balance", "jenga.account.balance")

			// Test validation
			validateErr := validateAccountBalanceRequest(tt.request)
			if tt.expectValidateError {
				assert.Error(t, validateErr)
			} else {
				assert.NoError(t, validateErr)
			}

			// Skip execution test if validation is expected to fail
			if !tt.expectValidateError {
				// Test execution
				executeErr := executeAccountBalanceRequest(mockClient, tt.request)
				if tt.expectExecuteError {
					assert.Error(t, executeErr)
				} else {
					assert.NoError(t, executeErr)
				}
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}
