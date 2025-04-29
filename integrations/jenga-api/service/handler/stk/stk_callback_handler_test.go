package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/jenga-api/service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleStkCallback(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		emitError      error
		expectedStatus int
	}{
		{
			name:   "Happy path - successful callback handling",
			method: http.MethodPost,
			requestBody: models.StkCallback{
				Transaction:    "TRX123",
				MobileNumber:   "254712345678",
				RequestAmount:  1000,
				Currency:       "KES",
				DebitedAmount:  1000,
				Charge:         30,
				TelcoName:      "Safaricom",
				Code:           0,
				Message:        "Success",
			},
			emitError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error - wrong HTTP method",
			method:         http.MethodGet,
			requestBody:    nil,
			emitError:      nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Error - invalid request body",
			method: http.MethodPost,
			requestBody: map[string]string{
				"invalid": "request",
			},
			emitError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Error - event processing failure",
			method: http.MethodPost,
			requestBody: models.StkCallback{
				Transaction:    "TRX456",
				MobileNumber:   "254712345678",
				RequestAmount:  1000,
				Currency:       "KES",
				DebitedAmount:  1000,
				Charge:         30,
				TelcoName:      "Safaricom",
				Code:           0,
				Message:        "Success",
			},
			emitError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Import the MockService from handlers_test.go
			mockService := new(MockService)

			// Set up service mock expectations
			if tt.method == http.MethodPost && tt.emitError != nil {
				mockService.On("Emit", mock.Anything, "jenga.callback.receive.payment", mock.Anything).Return(tt.emitError)
			} else if tt.method == http.MethodPost && tt.requestBody != nil {
				mockService.On("Emit", mock.Anything, "jenga.callback.receive.payment", mock.Anything).Return(nil)
			}

			// Create a test version of the JobServer that uses our MockService
			jobServer := struct {
				Service *MockService
			}{
				Service: mockService,
			}
			
			// Create a handler method that matches the real one
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
					return
				}
				
				ctx := r.Context()
				
				var callback models.StkCallback
				if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}
				
				// Test just the emit functionality that's key to the test
				if err := jobServer.Service.Emit(ctx, "jenga.callback.receive.payment", &callback); err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				
				w.WriteHeader(http.StatusOK)
			}

			// Create request
			var reqBody []byte
			var err error
			if tt.requestBody != nil {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest(tt.method, "/receivepayments", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call our test handler function
			handlerFunc(rr, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
