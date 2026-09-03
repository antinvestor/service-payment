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
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/config"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/credentials"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/handlers"
	"github.com/pitabwire/frame/v2/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey = "513fc4c3aaeb2a8f292a740ea178d830"
	testSecret = "webhook-secret"
)

// fakePaymentClient captures StatusUpdate calls; all other client methods are
// inherited from the embedded nil interface and must not be called.
type fakePaymentClient struct {
	paymentv1connect.PaymentServiceClient

	statusReq *commonv1.StatusUpdateRequest
	claims    *security.AuthenticationClaims
	err       error
}

func (f *fakePaymentClient) StatusUpdate(
	ctx context.Context,
	req *connect.Request[commonv1.StatusUpdateRequest],
) (*connect.Response[commonv1.StatusUpdateResponse], error) {
	f.statusReq = req.Msg
	f.claims = security.ClaimsFromContext(ctx)
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&commonv1.StatusUpdateResponse{}), nil
}

// fakeYellowcardClient serves the verified records the webhook server fetches.
type fakeYellowcardClient struct {
	client.YellowcardClient

	receive *client.Receive
	send    *client.Send
	err     error

	gotSequenceID string
	gotKind       string
}

func (f *fakeYellowcardClient) GetReceiveBySequenceID(_ context.Context, _ *client.Credentials, id string) (*client.Receive, error) {
	f.gotSequenceID, f.gotKind = id, "receive"
	return f.receive, f.err
}

func (f *fakeYellowcardClient) GetSendBySequenceID(_ context.Context, _ *client.Credentials, id string) (*client.Send, error) {
	f.gotSequenceID, f.gotKind = id, "send"
	return f.send, f.err
}

func testResolver(secret string) *credentials.Resolver {
	return credentials.NewResolver(nil, &config.YellowcardConfig{
		APIKey:        testAPIKey,
		SecretKey:     "api-secret",
		WebhookSecret: secret,
		Environment:   "sandbox",
	})
}

func newServer(paymentCli *fakePaymentClient, ycCli *fakeYellowcardClient) *handlers.YellowcardWebhookServer {
	return handlers.NewYellowcardWebhookServer(paymentCli, ycCli, testResolver(testSecret))
}

func post(t *testing.T, server *handlers.YellowcardWebhookServer, path, body, signature string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set(client.HeaderWebhookSignature, signature)
	}
	rr := httptest.NewRecorder()
	server.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

func signed(body string) string {
	return client.WebhookSignature(testSecret, []byte(body))
}

func receiveEvent(event, status string) string {
	return `{"id":"00e97bc4-1429-4ce7-acb5-841f9d9ed059","sequenceId":"prompt-1","status":"` + status +
		`","apiKey":"` + testAPIKey + `","event":"` + event + `","sessionId":"s1","executedAt":"2023-02-20T14:25:30.459Z"}`
}

func verifiedReceive(status, errorCode string) *client.Receive {
	return &client.Receive{
		ID: "00e97bc4-1429-4ce7-acb5-841f9d9ed059", SequenceID: "prompt-1", Status: status, ErrorCode: errorCode,
		Country: "UG", Currency: "UGX", ConvertedAmount: 5000, Rate: 3800.5, Reference: "REF1",
	}
}

func TestReceiveWebhook_Complete(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{receive: verifiedReceive("complete", "")}
	body := receiveEvent("RECEIVE.COMPLETE", "complete")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives?tenant_id=t1&partition_id=p1", body, signed(body))

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Equal(t, "prompt-1", ycCli.gotSequenceID)
	require.NotNil(t, paymentCli.statusReq)
	st := paymentCli.statusReq
	assert.Equal(t, "prompt-1", st.GetId())
	assert.Equal(t, "00e97bc4-1429-4ce7-acb5-841f9d9ed059", st.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, st.GetState())
	ex := st.GetExtras().AsMap()
	assert.Equal(t, "prompt", ex["entity_type"])
	assert.Equal(t, "yellowcard", ex["provider"])
	assert.Equal(t, "RECEIVE.COMPLETE", ex["event"])
	assert.Equal(t, "5000", ex["local_amount"])
	assert.Equal(t, "REF1", ex["provider_transaction_id"])
	_, hasFailure := ex["failure_code"]
	assert.False(t, hasFailure)

	require.NotNil(t, paymentCli.claims)
	assert.Equal(t, "t1", paymentCli.claims.TenantID)
	assert.Equal(t, "p1", paymentCli.claims.PartitionID)
}

func TestReceiveWebhook_Failed(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{receive: verifiedReceive("failed", "REFUSED")}
	body := receiveEvent("RECEIVE.FAILED", "failed")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, signed(body))

	require.Equal(t, http.StatusOK, rr.Code)
	st := paymentCli.statusReq
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, st.GetState())
	ex := st.GetExtras().AsMap()
	assert.Equal(t, "REFUSED", ex["failure_code"])
	assert.Equal(t, "The customer did not approve the payment", ex["failure_message"])
	assert.Nil(t, paymentCli.claims, "no tenancy without query params")
}

func TestReceiveWebhook_Expired_UsesStatusAsCode(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{receive: verifiedReceive("expired", "")}
	body := receiveEvent("RECEIVE.EXPIRED", "expired")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, signed(body))
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, commonv1.STATUS_FAILED, paymentCli.statusReq.GetStatus())
	assert.Equal(t, "EXPIRED", paymentCli.statusReq.GetExtras().AsMap()["failure_code"])
}

func TestReceiveWebhook_Pending(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	rcv := verifiedReceive("pending", "")
	rcv.BankInfo = &client.BankInfo{Name: "PAGA", AccountNumber: "0123", AccountName: "Ken"}
	ycCli := &fakeYellowcardClient{receive: rcv}
	body := receiveEvent("RECEIVE.PENDING", "pending")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, signed(body))
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, paymentCli.statusReq.GetStatus())
	ex := paymentCli.statusReq.GetExtras().AsMap()
	assert.Equal(t, "0123", ex["bank_account_number"])
	assert.Equal(t, "REF1", ex["payment_reference"])
}

func TestWebhook_BadSignature(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{receive: verifiedReceive("complete", "")}
	body := receiveEvent("RECEIVE.COMPLETE", "complete")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, client.WebhookSignature("wrong", []byte(body)))
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Nil(t, paymentCli.statusReq)
	assert.Empty(t, ycCli.gotSequenceID, "provider must not be consulted for unsigned webhooks")

	rr = post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestWebhook_APIKeyMismatch(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{receive: verifiedReceive("complete", "")}
	body := strings.Replace(receiveEvent("RECEIVE.COMPLETE", "complete"), testAPIKey, "someone-else", 1)

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, signed(body))
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Nil(t, paymentCli.statusReq)
}

func TestWebhook_UnknownSequence(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{err: &client.APIError{HTTPStatus: 404, Code: "CollectionNotFoundError"}}
	body := receiveEvent("RECEIVE.COMPLETE", "complete")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, signed(body))
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Nil(t, paymentCli.statusReq)
}

func TestWebhook_ProviderDown(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{err: errors.New("connection refused")}
	body := receiveEvent("RECEIVE.COMPLETE", "complete")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, signed(body))
	assert.Equal(t, http.StatusBadGateway, rr.Code)
	assert.Nil(t, paymentCli.statusReq)
}

func TestWebhook_StatusUpdateError(t *testing.T) {
	paymentCli := &fakePaymentClient{err: errors.New("db down")}
	ycCli := &fakeYellowcardClient{receive: verifiedReceive("complete", "")}
	body := receiveEvent("RECEIVE.COMPLETE", "complete")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard/receives", body, signed(body))
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestWebhook_NoCredentials(t *testing.T) {
	server := handlers.NewYellowcardWebhookServer(&fakePaymentClient{}, &fakeYellowcardClient{},
		credentials.NewResolver(nil, &config.YellowcardConfig{}))
	body := receiveEvent("RECEIVE.COMPLETE", "complete")
	rr := post(t, server, "/webhook/yellowcard/receives", body, signed(body))
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestWebhook_BadBody(t *testing.T) {
	server := newServer(&fakePaymentClient{}, &fakeYellowcardClient{})
	rr := post(t, server, "/webhook/yellowcard/receives", "{not json", signed("{not json"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	rr = post(t, server, "/webhook/yellowcard/receives", `{"status":"complete"}`, signed(`{"status":"complete"}`))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCatchAll_DispatchesSend(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{send: &client.Send{
		ID: "snd-1", SequenceID: "pay-1", Status: "complete", Country: "UG", Currency: "UGX", ConvertedAmount: 20000,
	}}
	body := `{"id":"snd-1","sequenceId":"pay-1","status":"complete","apiKey":"` + testAPIKey + `","event":"SEND.COMPLETE"}`

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard", body, signed(body))
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Equal(t, "send", ycCli.gotKind)
	assert.Equal(t, "pay-1", paymentCli.statusReq.GetId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, paymentCli.statusReq.GetStatus())
	assert.Equal(t, "payment", paymentCli.statusReq.GetExtras().AsMap()["entity_type"])
}

func TestCatchAll_LegacyEventNames(t *testing.T) {
	paymentCli := &fakePaymentClient{}
	ycCli := &fakeYellowcardClient{receive: verifiedReceive("complete", "")}
	body := receiveEvent("COLLECTION.COMPLETE", "complete")

	rr := post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard", body, signed(body))
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "receive", ycCli.gotKind)

	body = receiveEvent("SETTLEMENT.COMPLETE", "complete")
	rr = post(t, newServer(paymentCli, ycCli), "/webhook/yellowcard", body, signed(body))
	assert.Equal(t, http.StatusBadRequest, rr.Code, "unrelated event families are rejected")
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	newServer(&fakePaymentClient{}, &fakeYellowcardClient{}).NewRouterV1().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMapStatus(t *testing.T) {
	tests := []struct {
		status string
		want   commonv1.STATUS
		state  commonv1.STATE
	}{
		{"complete", commonv1.STATUS_SUCCESSFUL, commonv1.STATE_ACTIVE},
		{"COMPLETE", commonv1.STATUS_SUCCESSFUL, commonv1.STATE_ACTIVE},
		{"failed", commonv1.STATUS_FAILED, commonv1.STATE_INACTIVE},
		{"expired", commonv1.STATUS_FAILED, commonv1.STATE_INACTIVE},
		{"cancelled", commonv1.STATUS_FAILED, commonv1.STATE_INACTIVE},
		{"refunded", commonv1.STATUS_FAILED, commonv1.STATE_INACTIVE},
		{"pending_refund", commonv1.STATUS_IN_PROCESS, commonv1.STATE_ACTIVE},
		{"created", commonv1.STATUS_IN_PROCESS, commonv1.STATE_ACTIVE},
		{"processing", commonv1.STATUS_IN_PROCESS, commonv1.STATE_ACTIVE},
		{"", commonv1.STATUS_UNKNOWN, commonv1.STATE_ACTIVE},
	}
	for _, tt := range tests {
		st, state := handlers.MapStatus(tt.status)
		assert.Equal(t, tt.want, st, tt.status)
		assert.Equal(t, tt.state, state, tt.status)
	}
}
