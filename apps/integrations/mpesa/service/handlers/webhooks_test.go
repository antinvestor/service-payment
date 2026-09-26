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
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/handlers"
	"github.com/pitabwire/frame/v2/data"
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
	req *connect.Request[commonv1.StatusRequest],
) (*connect.Response[commonv1.StatusResponse], error) {
	if s.statusErr != nil {
		return nil, s.statusErr
	}
	extras, ok := s.stored[req.Msg.GetId()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("record not found"))
	}
	return connect.NewResponse(&commonv1.StatusResponse{
		Id:     req.Msg.GetId(),
		Status: commonv1.STATUS_IN_PROCESS,
		Extras: extras.ToProtoStruct(),
	}), nil
}

func (s *stubPaymentClient) StatusUpdate(
	_ context.Context,
	req *connect.Request[commonv1.StatusUpdateRequest],
) (*connect.Response[commonv1.StatusUpdateResponse], error) {
	s.updates = append(s.updates, req.Msg)
	return connect.NewResponse(&commonv1.StatusUpdateResponse{}), nil
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

func pendingPrompt(checkoutID string) data.JSONMap {
	return data.JSONMap{
		"entity_type":                     "prompt",
		client.ExtraCheckoutRequestID:     checkoutID,
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
			name:         "prompt without pending STK request",
			stored:       data.JSONMap{"entity_type": "prompt"},
			amount:       100,
			queryResp:    &client.STKQueryResponse{ResultCode: "0"},
			wantHTTP:     http.StatusConflict,
			wantNoUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := &stubPaymentClient{stored: map[string]data.JSONMap{"prompt-1": tt.stored}}
			mp := &stubMpesa{queryResp: tt.queryResp, queryErr: tt.queryErr}
			var connections []string
			srv := handlers.NewMpesaWebhookServer(pay, mp, credsFor(&connections))

			rr := post(srv, "/webhook/mpesa/stk?tenant_id=t1&partition_id=p1&prompt_id=prompt-1",
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
			if tt.wantVerifyTag != "" {
				assert.Equal(t, tt.wantVerifyTag, extras["verification"].GetStringValue())
			}
		})
	}
}

func TestSTKCallback_PromptStatusUnavailable(t *testing.T) {
	pay := &stubPaymentClient{statusErr: connect.NewError(connect.CodeUnavailable, errors.New("down"))}
	mp := &stubMpesa{queryResp: &client.STKQueryResponse{ResultCode: "0"}}
	var connections []string
	srv := handlers.NewMpesaWebhookServer(pay, mp, credsFor(&connections))

	rr := post(srv, "/webhook/mpesa/stk?prompt_id=prompt-1", stkBody(t, "ws_CO_1", 0, 100))

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Empty(t, pay.updates)
	assert.Empty(t, mp.queried)
}

func TestSTKCallback_LegacyWithoutPromptID(t *testing.T) {
	t.Run("success is confirmed and recorded under CheckoutRequestID", func(t *testing.T) {
		pay := &stubPaymentClient{}
		mp := &stubMpesa{queryResp: &client.STKQueryResponse{ResultCode: "0"}}
		var connections []string
		srv := handlers.NewMpesaWebhookServer(pay, mp, credsFor(&connections))

		rr := post(srv, "/webhook/mpesa/stk?tenant_id=t1", stkBody(t, "ws_CO_1", 0, 100))

		require.Equal(t, http.StatusOK, rr.Code)
		require.Len(t, pay.updates, 1)
		assert.Equal(t, "ws_CO_1", pay.updates[0].GetId())
		assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.updates[0].GetStatus())
		assert.Equal(t, []string{""}, connections, "legacy callbacks use configured credentials")
	})

	t.Run("forged success is not recorded as successful", func(t *testing.T) {
		pay := &stubPaymentClient{}
		mp := &stubMpesa{queryResp: &client.STKQueryResponse{ResultCode: "2001", ResultDesc: "invalid"}}
		var connections []string
		srv := handlers.NewMpesaWebhookServer(pay, mp, credsFor(&connections))

		rr := post(srv, "/webhook/mpesa/stk", stkBody(t, "ws_CO_forged", 0, 100))

		require.Equal(t, http.StatusOK, rr.Code)
		require.Len(t, pay.updates, 1)
		assert.Equal(t, commonv1.STATUS_FAILED, pay.updates[0].GetStatus())
	})
}

func TestSTKCallback_RejectsMissingCheckoutRequestID(t *testing.T) {
	pay := &stubPaymentClient{}
	srv := handlers.NewMpesaWebhookServer(pay, &stubMpesa{}, nil)

	rr := post(srv, "/webhook/mpesa/stk?prompt_id=prompt-1", stkBody(t, "", 0, 100))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Empty(t, pay.updates)
}
