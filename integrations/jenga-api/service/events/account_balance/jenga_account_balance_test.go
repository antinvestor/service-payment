package events

import (
	"context"
	"testing"

	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)



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
			// Create mock service and client
			mockService := &frame.Service{}
			mockClient := new(coreapi.MockClient)

			// Set up client mock expectations
			mockClient.On("GenerateBearerToken").Return(tt.tokenResponse, tt.tokenError)
			mockClient.On("InitiateAccountBalance", tt.request.CountryCode, tt.request.AccountId, mock.AnythingOfType("string")).
				Return(tt.balanceResponse, tt.balanceError)

			// Create the event handler
			event := &JengaAccountBalance{
				Service: mockService,
				Client:  mockClient,
			}

			// Test the Name method
			assert.Equal(t, "jenga.account.balance", event.Name())

			// Test the PayloadType method
			payloadType := event.PayloadType()
			_, ok := payloadType.(*models.AccountBalanceRequest)
			assert.True(t, ok)

			// Test the Validate method
			err := event.Validate(context.Background(), tt.request)
			if tt.expectValidateError {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			// Test the Execute method
			err = event.Execute(context.Background(), tt.request)

			if tt.expectExecuteError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}
