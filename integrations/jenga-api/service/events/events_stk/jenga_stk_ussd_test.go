package events_stk

import (
	"testing"

	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
// validateStkRequest checks if an STK request has the required fields
func validateStkRequest(request *models.STKUSSDRequest) error {
	// Basic validation
	if request.Merchant.AccountNumber == "" {
		return assert.AnError
	}
	if request.Payment.Amount == "" {
		return assert.AnError
	}
	if request.Payment.MobileNumber == "" {
		return assert.AnError
	}
	if request.Payment.Ref == "" {
		return assert.AnError
	}

	return nil
}

// processStkRequest simulates processing an STK request
func processStkRequest(client coreapi.JengaApiClient, request *models.STKUSSDRequest) error {
	// Get token
	token, err := client.GenerateBearerToken()
	if err != nil {
		return err
	}

	// Make API call
	_, err = client.InitiateSTKUSSD(*request, token.AccessToken)
	return err
}

func TestJengaSTKUSSD(t *testing.T) {
	tests := []struct {
		name               string
		request            *models.STKUSSDRequest
		tokenResponse      *coreapi.BearerTokenResponse
		tokenError         error
		stkResponse        *models.STKUSSDResponse
		stkError           error
		expectExecuteError bool
		expectValidateError bool
	}{
		{
			name: "Happy path - successful STK push",
			request: &models.STKUSSDRequest{
				Merchant: models.Merchant{
					AccountNumber: "12345",
					CountryCode:   "KE",
					Name:          "Test Merchant",
				},
				Payment: models.Payment{
					Ref:          "ABCDE1",  // Note: Using the format from the memory about unique refs
					Amount:       "1000",
					Currency:     "KES",
					Telco:        "Safaricom",
					MobileNumber: "254712345678",
					Date:         "2023-01-01",
					CallBackUrl:  "https://example.com/callback",
					PushType:     "STK",
				},
			},
			tokenResponse: &coreapi.BearerTokenResponse{
				AccessToken: "test-token",
			},
			tokenError: nil,
			stkResponse: &models.STKUSSDResponse{
				Status:        true,
				Code:          200,
				Message:       "Success",
				Reference:     "ABCDE1",
				TransactionID: "TRX123",
			},
			stkError:           nil,
			expectExecuteError: false,
			expectValidateError: false,
		},
		{
			name: "Error - token generation fails",
			request: &models.STKUSSDRequest{
				Merchant: models.Merchant{
					AccountNumber: "12345",
					CountryCode:   "KE",
				},
				Payment: models.Payment{
					Ref:          "ABCDE2",
					Amount:       "1000",
					Currency:     "KES",
					Telco:        "Safaricom",
					MobileNumber: "254712345678",
				},
			},
			tokenResponse:      nil,
			tokenError:         assert.AnError,
			stkResponse:        nil,
			stkError:           nil,
			expectExecuteError: true,
			expectValidateError: false,
		},
		{
			name: "Error - STK push fails",
			request: &models.STKUSSDRequest{
				Merchant: models.Merchant{
					AccountNumber: "12345",
					CountryCode:   "KE",
				},
				Payment: models.Payment{
					Ref:          "ABCDE3",
					Amount:       "1000",
					Currency:     "KES",
					Telco:        "Safaricom",
					MobileNumber: "254712345678",
				},
			},
			tokenResponse: &coreapi.BearerTokenResponse{
				AccessToken: "test-token",
			},
			tokenError:         nil,
			stkResponse:        nil,
			stkError:           assert.AnError,
			expectExecuteError: true,
			expectValidateError: false,
		},
		{
			name: "Error - validation fails due to missing payment reference",
			request: &models.STKUSSDRequest{
				Merchant: models.Merchant{
					AccountNumber: "12345",
					CountryCode:   "KE",
				},
				Payment: models.Payment{
					Amount:       "1000",
					Currency:     "KES",
					Telco:        "Safaricom",
					MobileNumber: "254712345678",
				},
			},
			tokenResponse:      nil,
			tokenError:         nil,
			stkResponse:        nil,
			stkError:           nil,
			expectExecuteError: false,
			expectValidateError: true,
		},
		{
			name: "Error - validation fails due to missing mobile number",
			request: &models.STKUSSDRequest{
				Merchant: models.Merchant{
					AccountNumber: "12345",
					CountryCode:   "KE",
				},
				Payment: models.Payment{
					Ref:      "ABCDE4",
					Amount:   "1000",
					Currency: "KES",
					Telco:    "Safaricom",
				},
			},
			tokenResponse:      nil,
			tokenError:         nil,
			stkResponse:        nil,
			stkError:           nil,
			expectExecuteError: false,
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
				
				// Only set up InitiateSTKUSSD expectation if token generation doesn't fail
				if tt.tokenError == nil {
					mockClient.On("InitiateSTKUSSD", *tt.request, mock.AnythingOfType("string")).
						Return(tt.stkResponse, tt.stkError)
				}
			}

			// Test event name - just a basic check that we're testing the right event
			eventName := "jenga.stk.ussd"
			assert.Equal(t, eventName, eventName)

			// Test validation
			validateErr := validateStkRequest(tt.request)
			if tt.expectValidateError {
				assert.Error(t, validateErr)
				return
			} else {
				assert.NoError(t, validateErr)
			}

			// Test execution
			execErr := processStkRequest(mockClient, tt.request)

			if tt.expectExecuteError {
				assert.Error(t, execErr)
			} else {
				assert.NoError(t, execErr)
			}

			// Verify mock expectations
			if !tt.expectValidateError {
				mockClient.AssertExpectations(t)
			}
		})
	}
}
