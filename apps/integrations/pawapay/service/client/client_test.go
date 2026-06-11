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

package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCreds(serverURL string) *client.Credentials {
	return &client.Credentials{
		APIToken:    "TEST_API_TOKEN",
		Environment: "sandbox",
		BaseURL:     serverURL,
	}
}

func TestResolveBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		creds    client.Credentials
		expected string
	}{
		{
			name:     "sandbox by default",
			creds:    client.Credentials{},
			expected: "https://api.sandbox.pawapay.io",
		},
		{
			name:     "production",
			creds:    client.Credentials{Environment: "production"},
			expected: "https://api.pawapay.io",
		},
		{
			name:     "explicit override wins",
			creds:    client.Credentials{Environment: "production", BaseURL: "http://localhost:1234"},
			expected: "http://localhost:1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.creds.ResolveBaseURL())
		})
	}
}

func TestInitiatePayout(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectError    bool
		expectedStatus string
		expectedCode   string
	}{
		{
			name:           "accepted",
			responseStatus: http.StatusOK,
			responseBody:   `{"payoutId":"f4401bd2-1568-4140-bf2d-eb77d2b2b639","status":"ACCEPTED","created":"2020-10-19T11:17:01Z"}`,
			expectedStatus: client.InitiationStatusAccepted,
		},
		{
			name:           "duplicate ignored",
			responseStatus: http.StatusOK,
			responseBody:   `{"payoutId":"f4401bd2-1568-4140-bf2d-eb77d2b2b639","status":"DUPLICATE_IGNORED","created":"2020-10-19T11:17:01Z"}`,
			expectedStatus: client.InitiationStatusDuplicateIgnored,
		},
		{
			name:           "rejected with failure reason",
			responseStatus: http.StatusOK,
			responseBody:   `{"payoutId":"f4401bd2-1568-4140-bf2d-eb77d2b2b639","status":"REJECTED","failureReason":{"failureCode":"PROVIDER_TEMPORARILY_UNAVAILABLE","failureMessage":"The provider is down"}}`,
			expectedStatus: client.InitiationStatusRejected,
			expectedCode:   "PROVIDER_TEMPORARILY_UNAVAILABLE",
		},
		{
			name:           "authentication failure",
			responseStatus: http.StatusUnauthorized,
			responseBody:   `{"status":"REJECTED","failureReason":{"failureCode":"AUTHENTICATION_ERROR","failureMessage":"The API token in the request is invalid."}}`,
			expectError:    true,
		},
		{
			name:           "server error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"failureReason":{"failureCode":"UNKNOWN_ERROR","failureMessage":"Unable to process request."}}`,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/v2/payouts", r.URL.Path)
				assert.Equal(t, "Bearer TEST_API_TOKEN", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var payload map[string]any
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
				assert.Equal(t, "f4401bd2-1568-4140-bf2d-eb77d2b2b639", payload["payoutId"])
				assert.Equal(t, "123.45", payload["amount"])
				assert.Equal(t, "ZMW", payload["currency"])
				recipient, _ := payload["recipient"].(map[string]any)
				assert.Equal(t, "MMO", recipient["type"])
				accountDetails, _ := recipient["accountDetails"].(map[string]any)
				assert.Equal(t, "260763456789", accountDetails["phoneNumber"])
				assert.Equal(t, "MTN_MOMO_ZMB", accountDetails["provider"])

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, err := w.Write([]byte(tt.responseBody))
				assert.NoError(t, err)
			}))
			defer server.Close()

			pawapayCli := client.NewClient()
			resp, err := pawapayCli.InitiatePayout(context.Background(), testCreds(server.URL), &client.PayoutRequest{
				PayoutID:    "f4401bd2-1568-4140-bf2d-eb77d2b2b639",
				Amount:      "123.45",
				Currency:    "ZMW",
				PhoneNumber: "260763456789",
				Provider:    "MTN_MOMO_ZMB",
			})

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedStatus, resp.Status)
			if tt.expectedCode != "" {
				require.NotNil(t, resp.FailureReason)
				assert.Equal(t, tt.expectedCode, resp.FailureReason.FailureCode)
			}
		})
	}
}

func TestInitiateDeposit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/deposits", r.URL.Path)

		var payload map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "8917c345-4791-4285-a416-62f24b6982db", payload["depositId"])
		assert.Equal(t, "15", payload["amount"])
		assert.Equal(t, "OTP123", payload["preAuthorisationCode"])
		assert.Equal(t, "INV-123456", payload["clientReferenceId"])
		assert.Equal(t, "Order payment", payload["customerMessage"])

		payer, _ := payload["payer"].(map[string]any)
		assert.Equal(t, "MMO", payer["type"])

		metadata, _ := payload["metadata"].([]any)
		if assert.Len(t, metadata, 1) {
			assert.Equal(t, map[string]any{"orderId": "ORD-1"}, metadata[0])
		}

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(
			[]byte(
				`{"depositId":"8917c345-4791-4285-a416-62f24b6982db","status":"ACCEPTED","created":"2020-10-19T11:17:01Z"}`,
			),
		)
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.InitiateDeposit(context.Background(), testCreds(server.URL), &client.DepositRequest{
		DepositID:            "8917c345-4791-4285-a416-62f24b6982db",
		Amount:               "15",
		Currency:             "ZMW",
		PhoneNumber:          "260763456789",
		Provider:             "MTN_MOMO_ZMB",
		PreAuthorisationCode: "OTP123",
		ClientReferenceID:    "INV-123456",
		CustomerMessage:      "Order payment",
		Metadata:             map[string]string{"orderId": "ORD-1"},
	})

	require.NoError(t, err)
	assert.Equal(t, client.InitiationStatusAccepted, resp.Status)
	assert.Equal(t, "8917c345-4791-4285-a416-62f24b6982db", resp.DepositID)
}

func TestGetDeposit(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		expectedSearch string
		expectedStatus string
	}{
		{
			name:           "completed deposit found",
			responseBody:   `{"status":"FOUND","data":{"depositId":"8917c345-4791-4285-a416-62f24b6982db","status":"COMPLETED","amount":"123.00","currency":"ZMW","country":"ZMB","payer":{"type":"MMO","accountDetails":{"phoneNumber":"260763456789","provider":"MTN_MOMO_ZMB"}},"created":"2020-10-19T08:17:01Z","providerTransactionId":"12356789","metadata":{"orderId":"ORD-123456789"}}}`,
			expectedSearch: client.SearchStatusFound,
			expectedStatus: client.PaymentStatusCompleted,
		},
		{
			name:           "not found",
			responseBody:   `{"status":"NOT_FOUND"}`,
			expectedSearch: client.SearchStatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v2/deposits/8917c345-4791-4285-a416-62f24b6982db", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(tt.responseBody))
				assert.NoError(t, err)
			}))
			defer server.Close()

			pawapayCli := client.NewClient()
			resp, err := pawapayCli.GetDeposit(
				context.Background(), testCreds(server.URL), "8917c345-4791-4285-a416-62f24b6982db",
			)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedSearch, resp.Status)
			if tt.expectedStatus != "" {
				require.NotNil(t, resp.Data)
				assert.Equal(t, tt.expectedStatus, resp.Data.Status)
				assert.Equal(t, "12356789", resp.Data.ProviderTransactionID)
				assert.Equal(t, "ORD-123456789", resp.Data.Metadata["orderId"])
			}
		})
	}
}

func TestGetPayoutFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/payouts/8917c345-4791-4285-a416-62f24b6982db", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(
			[]byte(
				`{"status":"FOUND","data":{"payoutId":"8917c345-4791-4285-a416-62f24b6982db","status":"FAILED","amount":"123.00","currency":"ZMW","country":"ZMB","recipient":{"type":"MMO","accountDetails":{"phoneNumber":"260763456789","provider":"MTN_MOMO_ZMB"}},"created":"2020-10-19T08:17:01Z","failureReason":{"failureCode":"RECIPIENT_NOT_FOUND","failureMessage":"Recipient not found"}}}`,
			),
		)
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.GetPayout(
		context.Background(), testCreds(server.URL), "8917c345-4791-4285-a416-62f24b6982db",
	)

	require.NoError(t, err)
	require.NotNil(t, resp.Data)
	assert.Equal(t, client.PaymentStatusFailed, resp.Data.Status)
	require.NotNil(t, resp.Data.FailureReason)
	assert.Equal(t, "RECIPIENT_NOT_FOUND", resp.Data.FailureReason.FailureCode)
}

func TestInitiateBulkPayouts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/payouts/bulk", r.URL.Path)

		var payload []map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Len(t, payload, 2)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(
			[]byte(
				`[{"payoutId":"a","status":"ACCEPTED"},{"payoutId":"b","status":"REJECTED","failureReason":{"failureCode":"AMOUNT_OUT_OF_BOUNDS","failureMessage":"too large"}}]`,
			),
		)
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.InitiateBulkPayouts(context.Background(), testCreds(server.URL), []*client.PayoutRequest{
		{PayoutID: "a", Amount: "10", Currency: "ZMW", PhoneNumber: "260763456789", Provider: "MTN_MOMO_ZMB"},
		{PayoutID: "b", Amount: "999999", Currency: "ZMW", PhoneNumber: "260763456780", Provider: "MTN_MOMO_ZMB"},
	})

	require.NoError(t, err)
	require.Len(t, resp, 2)
	assert.Equal(t, client.InitiationStatusAccepted, resp[0].Status)
	assert.Equal(t, client.InitiationStatusRejected, resp[1].Status)
}

func TestInitiateRefund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/refunds", r.URL.Path)

		var payload map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "11111111-1568-4140-bf2d-eb77d2b2b639", payload["refundId"])
		assert.Equal(t, "f4401bd2-1568-4140-bf2d-eb77d2b2b639", payload["depositId"])
		assert.Equal(t, "50", payload["amount"])

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(
			[]byte(
				`{"refundId":"11111111-1568-4140-bf2d-eb77d2b2b639","status":"ACCEPTED","created":"2020-10-19T11:17:01Z"}`,
			),
		)
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.InitiateRefund(context.Background(), testCreds(server.URL), &client.RefundRequest{
		RefundID:  "11111111-1568-4140-bf2d-eb77d2b2b639",
		DepositID: "f4401bd2-1568-4140-bf2d-eb77d2b2b639",
		Amount:    "50",
		Currency:  "ZMW",
	})

	require.NoError(t, err)
	assert.Equal(t, client.InitiationStatusAccepted, resp.Status)
}

func TestManualActions(t *testing.T) {
	tests := []struct {
		name         string
		call         func(cli client.PawapayClient, creds *client.Credentials) (*client.ManualActionResponse, error)
		expectedPath string
	}{
		{
			name: "resend deposit callback",
			call: func(cli client.PawapayClient, creds *client.Credentials) (*client.ManualActionResponse, error) {
				return cli.ResendDepositCallback(context.Background(), creds, "id-1")
			},
			expectedPath: "/v2/deposits/resend-callback/id-1",
		},
		{
			name: "resend payout callback",
			call: func(cli client.PawapayClient, creds *client.Credentials) (*client.ManualActionResponse, error) {
				return cli.ResendPayoutCallback(context.Background(), creds, "id-1")
			},
			expectedPath: "/v2/payouts/resend-callback/id-1",
		},
		{
			name: "cancel enqueued payout",
			call: func(cli client.PawapayClient, creds *client.Credentials) (*client.ManualActionResponse, error) {
				return cli.CancelEnqueuedPayout(context.Background(), creds, "id-1")
			},
			expectedPath: "/v2/payouts/fail-enqueued/id-1",
		},
		{
			name: "resend refund callback",
			call: func(cli client.PawapayClient, creds *client.Credentials) (*client.ManualActionResponse, error) {
				return cli.ResendRefundCallback(context.Background(), creds, "id-1")
			},
			expectedPath: "/v2/refunds/resend-callback/id-1",
		},
		{
			name: "cancel enqueued refund",
			call: func(cli client.PawapayClient, creds *client.Credentials) (*client.ManualActionResponse, error) {
				return cli.CancelEnqueuedRefund(context.Background(), creds, "id-1")
			},
			expectedPath: "/v2/refunds/fail-enqueued/id-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, tt.expectedPath, r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{"status":"ACCEPTED"}`))
				assert.NoError(t, err)
			}))
			defer server.Close()

			pawapayCli := client.NewClient()
			resp, err := tt.call(pawapayCli, testCreds(server.URL))

			require.NoError(t, err)
			assert.Equal(t, client.InitiationStatusAccepted, resp.Status)
		})
	}
}

func TestCreatePaymentPageSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/paymentpage", r.URL.Path)

		var payload map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "https://merchant.com/done", payload["returnUrl"])
		amountDetails, _ := payload["amountDetails"].(map[string]any)
		assert.Equal(t, "100", amountDetails["amount"])

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"redirectUrl":"https://paywith.pawapay.io/?token=abc"}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.CreatePaymentPageSession(
		context.Background(),
		testCreds(server.URL),
		&client.PaymentPageRequest{
			DepositID: "8917c345-4791-4285-a416-62f24b6982db",
			ReturnURL: "https://merchant.com/done",
			Amount:    "100",
			Currency:  "ZMW",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "https://paywith.pawapay.io/?token=abc", resp.RedirectURL)
}

func TestPredictProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/predict-provider", r.URL.Path)

		var payload map[string]string
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "+260 763-456789", payload["phoneNumber"])

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"country":"ZMB","provider":"MTN_MOMO_ZMB","phoneNumber":"260763456789"}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.PredictProvider(context.Background(), testCreds(server.URL), "+260 763-456789")

	require.NoError(t, err)
	assert.Equal(t, "MTN_MOMO_ZMB", resp.Provider)
	assert.Equal(t, "260763456789", resp.PhoneNumber)
}

func TestWalletBalances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/wallet-balances", r.URL.Path)
		assert.Equal(t, "ZMB", r.URL.Query().Get("country"))

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(
			[]byte(`{"balances":[{"country":"ZMB","balance":"21798.03","currency":"ZMW","provider":""}]}`),
		)
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.WalletBalances(context.Background(), testCreds(server.URL), "ZMB")

	require.NoError(t, err)
	require.Len(t, resp.Balances, 1)
	assert.Equal(t, "21798.03", resp.Balances[0].Balance)
}

func TestActiveConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/active-conf", r.URL.Path)
		assert.Equal(t, "ZMB", r.URL.Query().Get("country"))
		assert.Equal(t, "DEPOSIT", r.URL.Query().Get("operationType"))

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{
			"companyName":"Merchant Inc.",
			"signatureConfiguration":{"signedRequestsOnly":false,"signedCallbacks":true},
			"countries":[{"country":"ZMB","prefix":"260","providers":[{
				"provider":"MTN_MOMO_ZMB","displayName":"MTN","nameDisplayedToCustomer":"Merchant",
				"currencies":[{"currency":"ZMW","displayName":"K","operationTypes":{
					"DEPOSIT":{"minAmount":"1","maxAmount":"20000","decimalsInAmount":"TWO_PLACES","status":"OPERATIONAL","authType":"PROVIDER_AUTH","pinPrompt":"AUTOMATIC"}
				}}]
			}]}]
		}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.ActiveConfiguration(context.Background(), testCreds(server.URL), "ZMB", "DEPOSIT")

	require.NoError(t, err)
	assert.Equal(t, "Merchant Inc.", resp.CompanyName)
	assert.True(t, resp.SignatureConfiguration.SignedCallbacks)
	require.Len(t, resp.Countries, 1)
	require.Len(t, resp.Countries[0].Providers, 1)
	depositConf := resp.Countries[0].Providers[0].Currencies[0].OperationTypes["DEPOSIT"]
	assert.Equal(t, "20000", depositConf.MaxAmount)
	assert.Equal(t, "PROVIDER_AUTH", depositConf.AuthType)
}

func TestAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/availability", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(
			[]byte(
				`[{"country":"ZMB","providers":[{"provider":"MTN_MOMO_ZMB","operationTypes":[{"operationType":"DEPOSIT","status":"OPERATIONAL"},{"operationType":"PAYOUT","status":"DELAYED"}]}]}]`,
			),
		)
		assert.NoError(t, err)
	}))
	defer server.Close()

	pawapayCli := client.NewClient()
	resp, err := pawapayCli.Availability(context.Background(), testCreds(server.URL), "", "")

	require.NoError(t, err)
	require.Len(t, resp, 1)
	require.Len(t, resp[0].Providers, 1)
	assert.Equal(t, "DELAYED", resp[0].Providers[0].OperationTypes[1].Status)
}
