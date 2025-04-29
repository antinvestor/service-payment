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
			// Create mock service and client
			mockService := &frame.Service{}
			mockClient := new(coreapi.MockClient)

			// Set up client mock expectations
			mockClient.On("GenerateBearerToken").Return(tt.tokenResponse, tt.tokenError)
			mockClient.On("InitiateSTKUSSD", *tt.request, mock.AnythingOfType("string")).
				Return(tt.stkResponse, tt.stkError)

			// Create the event handler
			event := &JengaSTKUSSD{
				Service: mockService,
				Client:  mockClient,
			}

			// Test the Name method
			assert.Equal(t, "jenga.stk.ussd", event.Name())

			// Test the PayloadType method
			payloadType := event.PayloadType()
			_, ok := payloadType.(*models.STKUSSDRequest)
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
			if !tt.expectValidateError {
				mockClient.AssertExpectations(t)
			}
		})
	}
}
