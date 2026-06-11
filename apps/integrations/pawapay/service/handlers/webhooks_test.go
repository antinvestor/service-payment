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

package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/config"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/credentials"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePaymentClient captures StatusUpdate calls; all other client methods are
// inherited from the embedded nil interface and must not be called.
type fakePaymentClient struct {
	paymentv1connect.PaymentServiceClient

	statusReq *commonv1.StatusUpdateRequest
	err       error
}

func (f *fakePaymentClient) StatusUpdate(
	_ context.Context,
	req *connect.Request[commonv1.StatusUpdateRequest],
) (*connect.Response[commonv1.StatusUpdateResponse], error) {
	f.statusReq = req.Msg
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&commonv1.StatusUpdateResponse{}), nil
}

// fakePawapayClient serves the verified records the webhook server fetches
// from the pawaPay API; only the status check methods are implemented.
type fakePawapayClient struct {
	client.PawapayClient

	deposit *client.DepositStatusResult
	payout  *client.PayoutStatusResult
	refund  *client.RefundStatusResult
	err     error

	gotCreds *client.Credentials
	gotID    string
}

func (f *fakePawapayClient) GetDeposit(
	_ context.Context, creds *client.Credentials, depositID string,
) (*client.DepositStatusResult, error) {
	f.gotCreds, f.gotID = creds, depositID
	return f.deposit, f.err
}

func (f *fakePawapayClient) GetPayout(
	_ context.Context, creds *client.Credentials, payoutID string,
) (*client.PayoutStatusResult, error) {
	f.gotCreds, f.gotID = creds, payoutID
	return f.payout, f.err
}

func (f *fakePawapayClient) GetRefund(
	_ context.Context, creds *client.Credentials, refundID string,
) (*client.RefundStatusResult, error) {
	f.gotCreds, f.gotID = creds, refundID
	return f.refund, f.err
}

func testResolver(apiToken string) *credentials.Resolver {
	return credentials.NewResolver(nil, &config.PawapayConfig{
		APIToken:    apiToken,
		Environment: "sandbox",
	})
}

func newServer(
	paymentCli *fakePaymentClient,
	pawapayCli *fakePawapayClient,
) *handlers.PawapayWebhookServer {
	return handlers.NewPawapayWebhookServer(paymentCli, pawapayCli, testResolver("TEST_TOKEN"))
}

func postCallback(
	t *testing.T,
	server *handlers.PawapayWebhookServer,
	path, body string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

func verifiedDeposit(status string) *client.DepositStatusResult {
	return &client.DepositStatusResult{
		Status: client.SearchStatusFound,
		Data: &client.Deposit{
			DepositID: "8917c345-4791-4285-a416-62f24b6982db",
			Status:    status,
			Amount:    "123.45",
			Currency:  "ZMW",
			Country:   "ZMB",
			Payer: client.NewMMOParty(
				"260763456789", "MTN_MOMO_ZMB",
			),
			ProviderTransactionID: "ABC123",
			FailureReason: &client.FailureReason{
				FailureCode:    "INSUFFICIENT_BALANCE",
				FailureMessage: "Not enough funds",
			},
			Metadata: map[string]any{
				"paymentId":   "prompt-001",
				"entityType":  "prompt",
				"tenantId":    "t1",
				"partitionId": "p1",
			},
		},
	}
}

func TestHandleDepositCallback(t *testing.T) {
	depositCallbackBody := `{"depositId":"8917c345-4791-4285-a416-62f24b6982db","status":"COMPLETED"}`

	tests := []struct {
		name               string
		body               string
		verified           *client.DepositStatusResult
		lookupErr          error
		statusUpdateErr    error
		expectedHTTPStatus int
		expectedID         string
		expectedStatus     commonv1.STATUS
		expectedState      commonv1.STATE
	}{
		{
			name:               "verified completed deposit maps to successful",
			body:               depositCallbackBody,
			verified:           verifiedDeposit(client.PaymentStatusCompleted),
			expectedHTTPStatus: http.StatusOK,
			expectedID:         "prompt-001",
			expectedStatus:     commonv1.STATUS_SUCCESSFUL,
			expectedState:      commonv1.STATE_ACTIVE,
		},
		{
			name: "forged COMPLETED body cannot override verified FAILED record",
			body: depositCallbackBody, // attacker claims COMPLETED
			verified: verifiedDeposit(
				client.PaymentStatusFailed,
			), // pawaPay says FAILED
			expectedHTTPStatus: http.StatusOK,
			expectedID:         "prompt-001",
			expectedStatus:     commonv1.STATUS_FAILED,
			expectedState:      commonv1.STATE_INACTIVE,
		},
		{
			name:               "verified processing deposit maps to in process",
			body:               depositCallbackBody,
			verified:           verifiedDeposit(client.PaymentStatusProcessing),
			expectedHTTPStatus: http.StatusOK,
			expectedID:         "prompt-001",
			expectedStatus:     commonv1.STATUS_IN_PROCESS,
			expectedState:      commonv1.STATE_ACTIVE,
		},
		{
			name:               "deposit unknown to pawaPay is rejected",
			body:               depositCallbackBody,
			verified:           &client.DepositStatusResult{Status: client.SearchStatusNotFound},
			expectedHTTPStatus: http.StatusNotFound,
		},
		{
			name:               "pawaPay lookup failure rejects callback",
			body:               depositCallbackBody,
			lookupErr:          errors.New("pawapay unreachable"),
			expectedHTTPStatus: http.StatusBadGateway,
		},
		{
			name:               "invalid JSON rejected",
			body:               `not json`,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name:               "missing deposit id rejected",
			body:               `{"status":"COMPLETED"}`,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name:               "status update failure returns 500",
			body:               depositCallbackBody,
			verified:           verifiedDeposit(client.PaymentStatusCompleted),
			statusUpdateErr:    errors.New("downstream unavailable"),
			expectedHTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentCli := &fakePaymentClient{err: tt.statusUpdateErr}
			pawapayCli := &fakePawapayClient{deposit: tt.verified, err: tt.lookupErr}
			server := newServer(paymentCli, pawapayCli)

			rr := postCallback(t, server, "/webhook/pawapay/deposits", tt.body)

			require.Equal(t, tt.expectedHTTPStatus, rr.Code)
			if tt.expectedHTTPStatus != http.StatusOK {
				if tt.expectedHTTPStatus != http.StatusInternalServerError {
					assert.Nil(t, paymentCli.statusReq, "no status update may be sent for rejected callbacks")
				}
				return
			}

			require.NotNil(t, paymentCli.statusReq)
			assert.Equal(t, tt.expectedID, paymentCli.statusReq.GetId())
			assert.Equal(t, "8917c345-4791-4285-a416-62f24b6982db", paymentCli.statusReq.GetExternalId())
			assert.Equal(t, tt.expectedStatus, paymentCli.statusReq.GetStatus())
			assert.Equal(t, tt.expectedState, paymentCli.statusReq.GetState())
			assert.Equal(t, "TEST_TOKEN", pawapayCli.gotCreds.APIToken)
			assert.Equal(t, "8917c345-4791-4285-a416-62f24b6982db", pawapayCli.gotID)
		})
	}
}

func TestHandleDepositCallbackFailsClosedWithoutCredentials(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	pawapayCli := &fakePawapayClient{deposit: verifiedDeposit(client.PaymentStatusCompleted)}
	server := handlers.NewPawapayWebhookServer(paymentCli, pawapayCli, testResolver(""))

	rr := postCallback(t, server, "/webhook/pawapay/deposits",
		`{"depositId":"8917c345-4791-4285-a416-62f24b6982db","status":"COMPLETED"}`)

	require.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Nil(t, paymentCli.statusReq)
}

func TestHandleDepositCallbackExtrasFromVerifiedRecord(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	pawapayCli := &fakePawapayClient{deposit: verifiedDeposit(client.PaymentStatusFailed)}
	server := newServer(paymentCli, pawapayCli)

	// The posted body lies about every field; extras must come from the verified record.
	rr := postCallback(t, server, "/webhook/pawapay/deposits", `{
		"depositId":"8917c345-4791-4285-a416-62f24b6982db",
		"status":"COMPLETED",
		"amount":"999999","currency":"USD",
		"metadata":{"paymentId":"attacker-controlled","entityType":"payment"}
	}`)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)

	assert.Equal(t, "prompt-001", paymentCli.statusReq.GetId(), "id must come from verified metadata")

	extras := paymentCli.statusReq.GetExtras().AsMap()
	assert.Equal(t, "INSUFFICIENT_BALANCE", extras["failure_code"])
	assert.Equal(t, "Not enough funds", extras["failure_message"])
	assert.Equal(t, "ABC123", extras["provider_transaction_id"])
	assert.Equal(t, "MTN_MOMO_ZMB", extras["provider"])
	assert.Equal(t, "260763456789", extras["phone_number"])
	assert.Equal(t, "prompt", extras["entity_type"])
	assert.Equal(t, "123.45", extras["amount"], "amount must come from the verified record")
	assert.Equal(t, "ZMW", extras["currency"], "currency must come from the verified record")
}

func TestInjectTenantPrecedence(t *testing.T) {
	depositCallbackBody := `{"depositId":"8917c345-4791-4285-a416-62f24b6982db","status":"COMPLETED"}`

	t.Run("spoofed query tenant conflicting with verified metadata is rejected", func(t *testing.T) {
		paymentCli := &fakePaymentClient{}
		pawapayCli := &fakePawapayClient{deposit: verifiedDeposit(client.PaymentStatusCompleted)}
		server := newServer(paymentCli, pawapayCli)

		// verified metadata says tenantId=t1; the attacker claims another tenant
		rr := postCallback(t, server,
			"/webhook/pawapay/deposits?tenant_id=attacker-tenant&partition_id=p1",
			depositCallbackBody)

		require.Equal(t, http.StatusForbidden, rr.Code)
		assert.Nil(t, paymentCli.statusReq, "no status update may be sent on tenant mismatch")
	})

	t.Run("matching query tenant is accepted", func(t *testing.T) {
		paymentCli := &fakePaymentClient{}
		pawapayCli := &fakePawapayClient{deposit: verifiedDeposit(client.PaymentStatusCompleted)}
		server := newServer(paymentCli, pawapayCli)

		rr := postCallback(t, server,
			"/webhook/pawapay/deposits?tenant_id=t1&partition_id=p1",
			depositCallbackBody)

		require.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, paymentCli.statusReq)
	})

	t.Run("query tenant fills in when verified metadata has none", func(t *testing.T) {
		verified := verifiedDeposit(client.PaymentStatusCompleted)
		verified.Data.Metadata = map[string]any{"paymentId": "prompt-001"} // no tenant info
		paymentCli := &fakePaymentClient{}
		pawapayCli := &fakePawapayClient{deposit: verified}
		server := newServer(paymentCli, pawapayCli)

		rr := postCallback(t, server,
			"/webhook/pawapay/deposits?tenant_id=t9&partition_id=p9",
			`{"depositId":"8917c345-4791-4285-a416-62f24b6982db","status":"COMPLETED"}`)

		require.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, paymentCli.statusReq)
	})
}

func TestHandlePayoutCallback(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		expectedStatus commonv1.STATUS
		expectedState  commonv1.STATE
	}{
		{
			name:           "completed",
			status:         client.PaymentStatusCompleted,
			expectedStatus: commonv1.STATUS_SUCCESSFUL,
			expectedState:  commonv1.STATE_ACTIVE,
		},
		{
			name:           "failed",
			status:         client.PaymentStatusFailed,
			expectedStatus: commonv1.STATUS_FAILED,
			expectedState:  commonv1.STATE_INACTIVE,
		},
		{
			name:           "enqueued",
			status:         client.PaymentStatusEnqueued,
			expectedStatus: commonv1.STATUS_IN_PROCESS,
			expectedState:  commonv1.STATE_ACTIVE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentCli := &fakePaymentClient{}
			pawapayCli := &fakePawapayClient{payout: &client.PayoutStatusResult{
				Status: client.SearchStatusFound,
				Data: &client.Payout{
					PayoutID:  "f4401bd2-1568-4140-bf2d-eb77d2b2b639",
					Status:    tt.status,
					Amount:    "55",
					Currency:  "ZMW",
					Country:   "ZMB",
					Recipient: client.NewMMOParty("260763456789", "MTN_MOMO_ZMB"),
					Metadata:  map[string]any{"paymentId": "payment-001", "entityType": "payment"},
				},
			}}
			server := newServer(paymentCli, pawapayCli)

			rr := postCallback(t, server, "/webhook/pawapay/payouts",
				`{"payoutId":"f4401bd2-1568-4140-bf2d-eb77d2b2b639","status":"COMPLETED"}`)

			require.Equal(t, http.StatusOK, rr.Code)
			require.NotNil(t, paymentCli.statusReq)
			assert.Equal(t, "payment-001", paymentCli.statusReq.GetId())
			assert.Equal(t, "f4401bd2-1568-4140-bf2d-eb77d2b2b639", paymentCli.statusReq.GetExternalId())
			assert.Equal(t, tt.expectedStatus, paymentCli.statusReq.GetStatus())
			assert.Equal(t, tt.expectedState, paymentCli.statusReq.GetState())

			extras := paymentCli.statusReq.GetExtras().AsMap()
			assert.Equal(t, "payment", extras["entity_type"])
		})
	}
}

func TestHandleRefundCallback(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	pawapayCli := &fakePawapayClient{refund: &client.RefundStatusResult{
		Status: client.SearchStatusFound,
		Data: &client.Refund{
			RefundID:  "11111111-1568-4140-bf2d-eb77d2b2b639",
			Status:    client.PaymentStatusCompleted,
			Amount:    "50",
			Currency:  "ZMW",
			Country:   "ZMB",
			Recipient: client.NewMMOParty("260763456789", "MTN_MOMO_ZMB"),
			Metadata:  map[string]any{"paymentId": "refund-001"},
		},
	}}
	server := newServer(paymentCli, pawapayCli)

	rr := postCallback(t, server, "/webhook/pawapay/refunds",
		`{"refundId":"11111111-1568-4140-bf2d-eb77d2b2b639","status":"COMPLETED"}`)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, "refund-001", paymentCli.statusReq.GetId())
	assert.Equal(t, "11111111-1568-4140-bf2d-eb77d2b2b639", paymentCli.statusReq.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, paymentCli.statusReq.GetStatus())

	extras := paymentCli.statusReq.GetExtras().AsMap()
	assert.Equal(t, "11111111-1568-4140-bf2d-eb77d2b2b639", extras["refund_id"])
	assert.Equal(t, "payment", extras["entity_type"])
}
