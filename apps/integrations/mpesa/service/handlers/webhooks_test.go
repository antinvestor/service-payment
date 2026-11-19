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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect/v2"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/handlers"
	"github.com/pitabwire/frame/v3/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPaymentClient struct {
	paymentv1connect.PaymentServiceClient
	stored    map[string]data.JSONMap // prompt id -> latest status extras
	statusErr error
	updates   []*commonv1.StatusUpdateRequest
}

func (s *stubPaymentClient) Status(
	_ context.Context,
	req *commonv1.StatusRequest,
) (*commonv1.StatusResponse, error) {
	if s.statusErr != nil {
		return nil, s.statusErr
	}
	extras, ok := s.stored[req.GetId()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, "record not found")
	}
	// The payment service keys statuses by entity id and type.
	if want := req.GetExtras().GetFields()["entity_type"].GetStringValue(); want != "" && extras.GetString("entity_type") != want {
		return nil, connect.NewError(connect.CodeNotFound, "record not found")
	}
	return &commonv1.StatusResponse{
		Id:     req.GetId(),
		Status: commonv1.STATUS_IN_PROCESS,
		Extras: extras.ToProtoStruct(),
	}, nil
}

func (s *stubPaymentClient) StatusUpdate(
	_ context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusUpdateResponse, error) {
	s.updates = append(s.updates, req)
	return &commonv1.StatusUpdateResponse{}, nil
}

type stubMpesa struct {
	client.MpesaClient
	queryResp  *client.STKQueryResponse
	queryErr   error
	queried    []*client.STKQueryRequest
	queryCreds []*client.MpesaCredentials
}

func (m *stubMpesa) STKPushQuery(
	_ context.Context,
	creds *client.MpesaCredentials,
	req *client.STKQueryRequest,
) (*client.STKQueryResponse, error) {
	m.queried = append(m.queried, req)
	m.queryCreds = append(m.queryCreds, creds)
	return m.queryResp, m.queryErr
}

func credsFor(connections *[]string) handlers.CredentialsResolver {
	return func(_ context.Context, connection string) (*client.MpesaCredentials, error) {
		*connections = append(*connections, connection)
		return &client.MpesaCredentials{Shortcode: "174379", Passkey: "pk", ConsumerKey: "k", ConsumerSecret: "s"}, nil
	}
}

func stkBody(t *testing.T, checkoutID string, resultCode int, amount any) []byte {
	t.Helper()
	cb := map[string]any{
		"MerchantRequestID": "mr-1",
		"CheckoutRequestID": checkoutID,
		"ResultCode":        resultCode,
		"ResultDesc":        "desc",
	}
	if amount != nil {
		cb["CallbackMetadata"] = map[string]any{"Item": []map[string]any{
			{"Name": "Amount", "Value": amount},
			{"Name": "MpesaReceiptNumber", "Value": "RCP123"},
		}}
	}
	body, err := json.Marshal(map[string]any{"Body": map[string]any{"stkCallback": cb}})
	require.NoError(t, err)
	return body
}

func post(srv *handlers.MpesaWebhookServer, target string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

// testSecret is the per-record callback secret the workers put in the URL;
// records store only its digest.
const testSecret = "0123456789abcdef0123456789abcdef"

func pendingPrompt(checkoutID string) data.JSONMap {
	return data.JSONMap{
		"entity_type":                     "prompt",
		client.ExtraCheckoutRequestID:     checkoutID,
		client.ExtraCallbackSecretHash:    client.HashCallbackSecret(testSecret),
		client.ExtraRequestedAmount:       "100",
		client.ExtraCredentialsConnection: "tenant-conn",
	}
}

func TestSTKCallback_PromptIDRecordsUnderPrompt(t *testing.T) {
	tests := []struct {
		name          string
		stored        data.JSONMap
		resultCode    int
		amount        any
		queryResp     *client.STKQueryResponse
		queryErr      error
		wantHTTP      int
		wantStatus    commonv1.STATUS
		wantNoUpdate  bool
		wantQueried   bool
		wantVerifyTag string
	}{
		{
			name:          "confirmed success",
			stored:        pendingPrompt("ws_CO_1"),
			amount:        100,
			queryResp:     &client.STKQueryResponse{ResultCode: "0", ResultDesc: "processed"},
			wantHTTP:      http.StatusOK,
			wantStatus:    commonv1.STATUS_SUCCESSFUL,
			wantQueried:   true,
			wantVerifyTag: "stk_query",
		},
		{
			name:          "query says cancelled",
			stored:        pendingPrompt("ws_CO_1"),
			amount:        100,
			queryResp:     &client.STKQueryResponse{ResultCode: "1032", ResultDesc: "cancelled by user"},
			wantHTTP:      http.StatusOK,
			wantStatus:    commonv1.STATUS_FAILED,
			wantQueried:   true,
			wantVerifyTag: "stk_query_mismatch",
		},
		{
			name:          "query has no result yet",
			stored:        pendingPrompt("ws_CO_1"),
			amount:        100.0,
			queryResp:     &client.STKQueryResponse{ResponseCode: "0"},
			wantHTTP:      http.StatusOK,
			wantStatus:    commonv1.STATUS_IN_PROCESS,
			wantQueried:   true,
			wantVerifyTag: "stk_query_pending",
		},
		{
			name:         "query unavailable records nothing",
			stored:       pendingPrompt("ws_CO_1"),
			amount:       100,
			queryErr:     errors.New("500.001.1001 the transaction is being processed"),
			wantHTTP:     http.StatusServiceUnavailable,
			wantNoUpdate: true,
			wantQueried:  true,
		},
		{
			name:          "amount mismatch fails without query",
			stored:        pendingPrompt("ws_CO_1"),
			amount:        1,
			wantHTTP:      http.StatusOK,
			wantStatus:    commonv1.STATUS_FAILED,
			wantVerifyTag: "amount_mismatch",
		},
		{
			name:          "missing amount on success fails",
			stored:        pendingPrompt("ws_CO_1"),
			wantHTTP:      http.StatusOK,
			wantStatus:    commonv1.STATUS_FAILED,
			wantVerifyTag: "amount_mismatch",
		},
		{
			name:       "failure callback recorded without query",
			stored:     pendingPrompt("ws_CO_1"),
			resultCode: 1032,
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name:         "checkout request id belongs to another prompt",
			stored:       pendingPrompt("ws_CO_other"),
			amount:       100,
			queryResp:    &client.STKQueryResponse{ResultCode: "0"},
			wantHTTP:     http.StatusConflict,
			wantNoUpdate: true,
		},
		{
			name:         "prompt without pending STK request has no secret: forbidden",
			stored:       data.JSONMap{"entity_type": "prompt"},
			amount:       100,
			queryResp:    &client.STKQueryResponse{ResultCode: "0"},
			wantHTTP:     http.StatusForbidden,
			wantNoUpdate: true,
		},
		{
			name:         "prompt bound but no secret recorded: forbidden",
			stored:       data.JSONMap{"entity_type": "prompt", client.ExtraCheckoutRequestID: "ws_CO_1"},
			amount:       100,
			queryResp:    &client.STKQueryResponse{ResultCode: "0"},
			wantHTTP:     http.StatusForbidden,
			wantNoUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := &stubPaymentClient{stored: map[string]data.JSONMap{"prompt-1": tt.stored}}
			mp := &stubMpesa{queryResp: tt.queryResp, queryErr: tt.queryErr}
			var connections []string
			srv := handlers.NewMpesaWebhookServer(pay, mp, credsFor(&connections))

			rr := post(srv, "/webhook/mpesa/stk?tenant_id=t1&partition_id=p1&prompt_id=prompt-1&cb_secret="+testSecret,
				stkBody(t, "ws_CO_1", tt.resultCode, tt.amount))

			require.Equal(t, tt.wantHTTP, rr.Code, rr.Body.String())
			if tt.wantQueried {
				require.Len(t, mp.queried, 1)
				assert.Equal(t, "ws_CO_1", mp.queried[0].CheckoutRequestID)
				assert.Equal(t, "174379", mp.queried[0].BusinessShortCode)
				assert.Equal(t, []string{"tenant-conn"}, connections)
			} else {
				assert.Empty(t, mp.queried)
			}
			if tt.wantNoUpdate {
				assert.Empty(t, pay.updates)
				return
			}
			require.Len(t, pay.updates, 1)
			got := pay.updates[0]
			assert.Equal(t, "prompt-1", got.GetId())
			assert.Equal(t, tt.wantStatus, got.GetStatus())
			extras := got.GetExtras().GetFields()
			assert.Equal(t, "prompt", extras["entity_type"].GetStringValue())
			assert.Equal(t, "ws_CO_1", extras[client.ExtraCheckoutRequestID].GetStringValue())
			assert.Equal(t, "100", extras[client.ExtraRequestedAmount].GetStringValue())
			assert.Equal(t, "tenant-conn", extras[client.ExtraCredentialsConnection].GetStringValue())
			assert.Equal(t, client.HashCallbackSecret(testSecret), extras[client.ExtraCallbackSecretHash].GetStringValue(), "digest carried forward")
			for k, v := range extras {
				assert.NotEqual(t, testSecret, v.GetStringValue(), "extra %q carries the secret in clear", k)
			}
			if tt.wantVerifyTag != "" {
				assert.Equal(t, tt.wantVerifyTag, extras["verification"].GetStringValue())
			}
		})
	}
}

func TestSTKCallback_PromptStatusUnavailable(t *testing.T) {
	pay := &stubPaymentClient{statusErr: connect.NewError(connect.CodeUnavailable, "down")}
	mp := &stubMpesa{queryResp: &client.STKQueryResponse{ResultCode: "0"}}
	var connections []string
	srv := handlers.NewMpesaWebhookServer(pay, mp, credsFor(&connections))

	rr := post(srv, "/webhook/mpesa/stk?prompt_id=prompt-1&cb_secret="+testSecret, stkBody(t, "ws_CO_1", 0, 100))

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Empty(t, pay.updates)
	assert.Empty(t, mp.queried)
}

func TestSTKCallback_WithoutPromptIDIsRejected(t *testing.T) {
	pay := &stubPaymentClient{}
	mp := &stubMpesa{queryResp: &client.STKQueryResponse{ResultCode: "0"}}
	var connections []string
	srv := handlers.NewMpesaWebhookServer(pay, mp, credsFor(&connections))

	for _, resultCode := range []int{0, 1032} {
		rr := post(srv, "/webhook/mpesa/stk?tenant_id=t1", stkBody(t, "ws_CO_1", resultCode, 100))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	}
	assert.Empty(t, pay.updates, "nothing is recorded under a bare CheckoutRequestID")
	assert.Empty(t, mp.queried)
}

func TestSTKCallback_ForgedSecretIsForbidden(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{name: "missing secret", target: "/webhook/mpesa/stk?prompt_id=prompt-1"},
		{name: "wrong secret", target: "/webhook/mpesa/stk?prompt_id=prompt-1&cb_secret=guess"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := &stubPaymentClient{stored: map[string]data.JSONMap{"prompt-1": pendingPrompt("ws_CO_1")}}
			mp := &stubMpesa{queryResp: &client.STKQueryResponse{ResultCode: "0"}}
			srv := handlers.NewMpesaWebhookServer(pay, mp, nil)

			for _, resultCode := range []int{0, 1032} {
				rr := post(srv, tt.target, stkBody(t, "ws_CO_1", resultCode, 100))
				assert.Equal(t, http.StatusForbidden, rr.Code)
			}
			assert.Empty(t, pay.updates)
			assert.Empty(t, mp.queried, "Daraja is never queried for a forged callback")
		})
	}
}

func TestSTKCallback_FailureNeedsBinding(t *testing.T) {
	pay := &stubPaymentClient{stored: map[string]data.JSONMap{"prompt-1": pendingPrompt("ws_CO_other")}}
	srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

	rr := post(srv, "/webhook/mpesa/stk?prompt_id=prompt-1&cb_secret="+testSecret, stkBody(t, "ws_CO_forged", 1032, nil))

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Empty(t, pay.updates)
}

func TestWebhook_BodyTooLarge(t *testing.T) {
	pay := &stubPaymentClient{}
	srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

	rr := post(srv, "/webhook/mpesa/c2b/confirmation", bytes.Repeat([]byte(" "), 1<<20+1))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Empty(t, pay.updates)
}

func TestSTKCallback_RejectsMissingCheckoutRequestID(t *testing.T) {
	pay := &stubPaymentClient{}
	srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

	rr := post(srv, "/webhook/mpesa/stk?prompt_id=prompt-1&cb_secret="+testSecret, stkBody(t, "", 0, 100))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Empty(t, pay.updates)
}

func c2bBody(t *testing.T, billRef, amount string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"TransactionType": "Pay Bill", "TransID": "RKTQDM7W6S", "TransTime": "20191122063845",
		"TransAmount": amount, "BusinessShortCode": "600638", "BillRefNumber": billRef,
		"MSISDN": "254700000001", "FirstName": "John",
	})
	require.NoError(t, err)
	return body
}

func guard(t *testing.T, cidrs string, hops int) *handlers.SourceGuard {
	t.Helper()
	g, err := handlers.NewSourceGuard(cidrs, hops)
	require.NoError(t, err)
	return g
}

func postFrom(srv *handlers.MpesaWebhookServer, target, remoteAddr string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	req.RemoteAddr = remoteAddr
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

func TestC2BConfirmation_SettlesOnlyVerifiedPrompts(t *testing.T) {
	prompt := data.JSONMap{"entity_type": "prompt", client.ExtraRequestedAmount: "100", client.ExtraRequestedCurrency: "KES"}
	tests := []struct {
		name       string
		stored     map[string]data.JSONMap
		statusErr  error
		guard      *handlers.SourceGuard
		remote     string
		billRef    string
		amount     string
		wantHTTP   int
		wantStatus commonv1.STATUS
		wantVerify string
	}{
		{
			name: "allowlisted source with matching amount settles", stored: map[string]data.JSONMap{"prompt-1": prompt},
			guard: guard(t, "196.201.214.0/24", 0), remote: "196.201.214.200:4433", billRef: "prompt-1", amount: "100.00",
			wantHTTP: http.StatusOK, wantStatus: commonv1.STATUS_SUCCESSFUL, wantVerify: "c2b_source_amount",
		},
		{
			name: "no allowlist configured: unverified", stored: map[string]data.JSONMap{"prompt-1": prompt},
			remote: "196.201.214.200:4433", billRef: "prompt-1", amount: "100",
			wantHTTP: http.StatusOK, wantStatus: commonv1.STATUS_IN_PROCESS, wantVerify: "c2b_unverified",
		},
		{
			name: "allowlisted source but amount differs: unverified", stored: map[string]data.JSONMap{"prompt-1": prompt},
			guard: guard(t, "196.201.214.0/24", 0), remote: "196.201.214.200:4433", billRef: "prompt-1", amount: "10",
			wantHTTP: http.StatusOK, wantStatus: commonv1.STATUS_IN_PROCESS, wantVerify: "c2b_unverified",
		},
		{
			name: "allowlisted source but no recorded amount: unverified", stored: map[string]data.JSONMap{"prompt-1": {"entity_type": "prompt"}},
			guard: guard(t, "196.201.214.0/24", 0), remote: "196.201.214.200:4433", billRef: "prompt-1", amount: "100",
			wantHTTP: http.StatusOK, wantStatus: commonv1.STATUS_IN_PROCESS, wantVerify: "c2b_unverified",
		},
		{
			name: "source outside allowlist is refused", stored: map[string]data.JSONMap{"prompt-1": prompt},
			guard: guard(t, "196.201.214.0/24", 0), remote: "203.0.113.9:4433", billRef: "prompt-1", amount: "100",
			wantHTTP: http.StatusForbidden,
		},
		{
			name: "payments are never settled by C2B", stored: map[string]data.JSONMap{"payment-1": {"entity_type": "payment", client.ExtraRequestedAmount: "100"}},
			guard: guard(t, "196.201.214.0/24", 0), remote: "196.201.214.200:4433", billRef: "payment-1", amount: "100",
			wantHTTP: http.StatusConflict,
		},
		{
			name: "unknown bill reference is refused", stored: map[string]data.JSONMap{},
			remote: "196.201.214.200:4433", billRef: "anything", amount: "100", wantHTTP: http.StatusConflict,
		},
		{
			name: "missing bill reference", stored: map[string]data.JSONMap{},
			remote: "196.201.214.200:4433", amount: "100", wantHTTP: http.StatusBadRequest,
		},
		{
			name: "lookup failure is retryable", statusErr: connect.NewError(connect.CodeUnavailable, "down"),
			remote: "196.201.214.200:4433", billRef: "prompt-1", amount: "100", wantHTTP: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := &stubPaymentClient{stored: tt.stored, statusErr: tt.statusErr}
			var opts []handlers.Option
			if tt.guard != nil {
				opts = append(opts, handlers.WithSourceGuard(tt.guard))
			}
			srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil, opts...)

			rr := postFrom(srv, "/webhook/mpesa/c2b/confirmation?tenant_id=t1", tt.remote, c2bBody(t, tt.billRef, tt.amount))

			require.Equal(t, tt.wantHTTP, rr.Code, rr.Body.String())
			if tt.wantHTTP != http.StatusOK {
				assert.Empty(t, pay.updates)
				return
			}
			require.Len(t, pay.updates, 1)
			got := pay.updates[0]
			assert.Equal(t, tt.billRef, got.GetId())
			assert.Equal(t, tt.wantStatus, got.GetStatus())
			assert.Equal(t, commonv1.STATE_ACTIVE, got.GetState())
			assert.Equal(t, "RKTQDM7W6S", got.GetExternalId())
			extras := got.GetExtras().GetFields()
			assert.Equal(t, "prompt", extras["entity_type"].GetStringValue())
			assert.Equal(t, tt.wantVerify, extras["verification"].GetStringValue())
		})
	}
}

func TestSourceGuard_Allowlist(t *testing.T) {
	tests := []struct {
		name   string
		cidrs  string
		hops   int
		remote string
		xff    string
		wantOK bool
		wantIP string
	}{
		{name: "empty allowlist admits everyone", remote: "203.0.113.9:1", wantOK: true, wantIP: "203.0.113.9"},
		{name: "remote addr inside", cidrs: "196.201.214.0/24,196.201.212.0/24", remote: "196.201.212.7:1", wantOK: true, wantIP: "196.201.212.7"},
		{name: "remote addr outside", cidrs: "196.201.214.0/24", remote: "203.0.113.9:1", wantIP: "203.0.113.9"},
		{name: "bare ip entry", cidrs: "196.201.214.5", remote: "196.201.214.5:1", wantOK: true, wantIP: "196.201.214.5"},
		{name: "XFF ignored without trusted hops", cidrs: "196.201.214.0/24", remote: "10.0.0.2:1", xff: "196.201.214.5", wantIP: "10.0.0.2"},
		{name: "one trusted hop reads the last XFF entry", cidrs: "196.201.214.0/24", hops: 1, remote: "10.0.0.2:1", xff: "196.201.214.5", wantOK: true, wantIP: "196.201.214.5"},
		{name: "spoofed entry before the trusted hop is ignored", cidrs: "196.201.214.0/24", hops: 1, remote: "10.0.0.2:1", xff: "196.201.214.5, 203.0.113.9", wantIP: "203.0.113.9"},
		{name: "two trusted hops", cidrs: "196.201.214.0/24", hops: 2, remote: "10.0.0.3:1", xff: "196.201.214.5, 10.0.0.2", wantOK: true, wantIP: "196.201.214.5"},
		{name: "fewer XFF entries than hops falls back to remote addr", cidrs: "196.201.214.0/24", hops: 2, remote: "10.0.0.3:1", xff: "196.201.214.5", wantIP: "10.0.0.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := guard(t, tt.cidrs, tt.hops)
			req := httptest.NewRequest(http.MethodPost, "/webhook/mpesa/stk", nil)
			req.RemoteAddr = tt.remote
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			assert.Equal(t, tt.wantIP, g.ClientIP(req).String())
			assert.Equal(t, tt.wantOK, g.Allowed(req))
		})
	}
	_, err := handlers.NewSourceGuard("not-a-cidr/99", 0)
	require.Error(t, err)
}

func TestSourceGuard_RefusesEveryRoute(t *testing.T) {
	pay := &stubPaymentClient{stored: map[string]data.JSONMap{"prompt-1": pendingPrompt("ws_CO_1")}}
	srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil, handlers.WithSourceGuard(guard(t, "196.201.214.0/24", 0)))

	for _, target := range []string{
		"/webhook/mpesa/stk?prompt_id=prompt-1&cb_secret=" + testSecret,
		"/webhook/mpesa/c2b/validation", "/webhook/mpesa/c2b/confirmation",
		"/webhook/mpesa/b2c?payment_id=p&cb_secret=x", "/webhook/mpesa/b2c/timeout?payment_id=p&cb_secret=x",
	} {
		rr := postFrom(srv, target, "203.0.113.9:1", stkBody(t, "ws_CO_1", 0, 100))
		assert.Equal(t, http.StatusForbidden, rr.Code, target)
	}
	assert.Empty(t, pay.updates)
}

func b2cBody(t *testing.T, originatorID string, resultCode int, amount any) []byte {
	t.Helper()
	result := map[string]any{
		"ResultType": 0, "ResultCode": resultCode, "ResultDesc": "desc",
		"OriginatorConversationID": originatorID, "ConversationID": "AG_1", "TransactionID": "LGR019G3J2",
	}
	if amount != nil {
		result["ResultParameters"] = map[string]any{"ResultParameter": []map[string]any{
			{"Key": client.ResultParamTransactionAmount, "Value": amount},
			{"Key": client.ResultParamTransactionReceipt, "Value": "LGR019G3J2"},
		}}
	}
	body, err := json.Marshal(map[string]any{"Result": result})
	require.NoError(t, err)
	return body
}

const testOriginator = "fedcba9876543210fedcba9876543210"

func issuedPayout() data.JSONMap {
	return data.JSONMap{
		"entity_type":                        "payment",
		client.ExtraOriginatorConversationID: testOriginator,
		client.ExtraCallbackSecretHash:       client.HashCallbackSecret(testSecret),
		client.ExtraRequestedAmount:          "250",
		client.ExtraRequestedCurrency:        "KES",
		client.ExtraCredentialsConnection:    "tenant-conn",
	}
}

const b2cTarget = "/webhook/mpesa/b2c?payment_id=payment-1&cb_secret=" + testSecret

func TestB2CResult_BindsToIssuedPayout(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		stored       data.JSONMap
		statusErr    error
		originatorID string
		resultCode   int
		amount       any
		wantHTTP     int
		wantStatus   commonv1.STATUS
		wantVerify   string
	}{
		{
			name: "confirmed payout", target: b2cTarget, stored: issuedPayout(),
			originatorID: testOriginator, amount: 250, wantHTTP: http.StatusOK,
			wantStatus: commonv1.STATUS_SUCCESSFUL, wantVerify: "b2c_binding_amount",
		},
		{
			name: "failed payout", target: b2cTarget, stored: issuedPayout(),
			originatorID: testOriginator, resultCode: 2001, wantHTTP: http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED, wantVerify: "b2c_binding",
		},
		{
			name: "paid amount differs", target: b2cTarget, stored: issuedPayout(),
			originatorID: testOriginator, amount: 25, wantHTTP: http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED, wantVerify: "amount_mismatch",
		},
		{
			name: "success without amount parameter", target: b2cTarget, stored: issuedPayout(),
			originatorID: testOriginator, wantHTTP: http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED, wantVerify: "amount_mismatch",
		},
		{
			name: "originator belongs to another payout", target: b2cTarget, stored: issuedPayout(),
			originatorID: "payment-other", amount: 250, wantHTTP: http.StatusConflict,
		},
		{
			name: "payment without an issued payout has no secret: forbidden", target: b2cTarget,
			stored: data.JSONMap{"entity_type": "payment"}, originatorID: testOriginator, amount: 250, wantHTTP: http.StatusForbidden,
		},
		{
			name: "forged secret", target: "/webhook/mpesa/b2c?payment_id=payment-1&cb_secret=guess", stored: issuedPayout(),
			originatorID: testOriginator, amount: 250, wantHTTP: http.StatusForbidden,
		},
		{
			name: "missing secret", target: "/webhook/mpesa/b2c?payment_id=payment-1", stored: issuedPayout(),
			originatorID: testOriginator, amount: 250, wantHTTP: http.StatusForbidden,
		},
		{
			name: "missing payment_id", target: "/webhook/mpesa/b2c?cb_secret=" + testSecret, stored: issuedPayout(),
			originatorID: testOriginator, amount: 250, wantHTTP: http.StatusBadRequest,
		},
		{
			name: "lookup failure is retryable", target: b2cTarget,
			statusErr: connect.NewError(connect.CodeUnavailable, "down"), originatorID: testOriginator, amount: 250,
			wantHTTP: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := &stubPaymentClient{stored: map[string]data.JSONMap{"payment-1": tt.stored}, statusErr: tt.statusErr}
			srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

			rr := post(srv, tt.target, b2cBody(t, tt.originatorID, tt.resultCode, tt.amount))

			require.Equal(t, tt.wantHTTP, rr.Code, rr.Body.String())
			if tt.wantHTTP != http.StatusOK {
				assert.Empty(t, pay.updates)
				return
			}
			require.Len(t, pay.updates, 1)
			got := pay.updates[0]
			assert.Equal(t, "payment-1", got.GetId())
			assert.Equal(t, tt.wantStatus, got.GetStatus())
			assert.Equal(t, commonv1.STATUS_SUCCESSFUL == tt.wantStatus, got.GetState() == commonv1.STATE_ACTIVE)
			assert.Equal(t, "LGR019G3J2", got.GetExternalId())
			extras := got.GetExtras().GetFields()
			assert.Equal(t, "payment", extras["entity_type"].GetStringValue())
			assert.Equal(t, tt.wantVerify, extras["verification"].GetStringValue())
			assert.Equal(t, testOriginator, extras[client.ExtraOriginatorConversationID].GetStringValue())
			assert.Equal(t, "250", extras[client.ExtraRequestedAmount].GetStringValue(), "binding carried forward")
			for k, v := range extras {
				assert.NotEqual(t, testSecret, v.GetStringValue(), "extra %q carries the secret in clear", k)
			}
		})
	}
}

func TestB2CTimeout_LeavesPaymentInProcess(t *testing.T) {
	t.Run("bound payout is left as is", func(t *testing.T) {
		pay := &stubPaymentClient{stored: map[string]data.JSONMap{"payment-1": issuedPayout()}}
		srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

		rr := post(srv, "/webhook/mpesa/b2c/timeout?payment_id=payment-1&cb_secret="+testSecret, b2cBody(t, testOriginator, 1, nil))

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, pay.updates, "a timeout never records FAILED")
	})
	t.Run("unbound timeout is refused", func(t *testing.T) {
		pay := &stubPaymentClient{stored: map[string]data.JSONMap{}}
		srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

		rr := post(srv, "/webhook/mpesa/b2c/timeout?payment_id=payment-1&cb_secret="+testSecret, b2cBody(t, testOriginator, 1, nil))

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Empty(t, pay.updates)
	})
	t.Run("forged timeout is forbidden", func(t *testing.T) {
		pay := &stubPaymentClient{stored: map[string]data.JSONMap{"payment-1": issuedPayout()}}
		srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

		rr := post(srv, "/webhook/mpesa/b2c/timeout?payment_id=payment-1&cb_secret=nope", b2cBody(t, testOriginator, 1, nil))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Empty(t, pay.updates)
	})
}
