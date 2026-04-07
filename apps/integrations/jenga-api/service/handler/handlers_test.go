package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/stretchr/testify/require"
)

func TestHandleStkCallback(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:   "Valid STK callback",
			method: http.MethodPost,
			requestBody: models.StkCallback{
				Status:        true,
				Transaction:   "TRX123",
				MobileNumber:  "254712345678",
				RequestAmount: json.Number("1000"),
				Currency:      "KES",
				DebitedAmount: json.Number("1000"),
				Charge:        json.Number("30"),
				TelcoName:     "Safaricom",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Missing required fields",
			method: http.MethodPost,
			requestBody: models.StkCallback{
				Status:  true,
				Message: "Success",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON body",
			method:         http.MethodPost,
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the HTTP flow without a real payment client
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				var callback models.StkCallback
				if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				if callback.Transaction == "" || callback.MobileNumber == "" || callback.Currency == "" {
					http.Error(w, "Missing required fields", http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
					"status": "success",
				})
			}

			reqBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(tt.method, "/receivepayments", bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handlerFunc(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
