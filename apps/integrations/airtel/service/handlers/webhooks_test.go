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
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/handlers"
	"github.com/pitabwire/frame/v2/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPaymentClient struct {
	paymentv1connect.PaymentServiceClient
	stored    map[string]data.JSONMap
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
	return connect.NewResponse(&commonv1.StatusResponse{Id: req.Msg.GetId(), Extras: extras.ToProtoStruct()}), nil
}

func (s *stubPaymentClient) StatusUpdate(
	_ context.Context,
	req *connect.Request[commonv1.StatusUpdateRequest],
) (*connect.Response[commonv1.StatusUpdateResponse], error) {
	s.updates = append(s.updates, req.Msg)
	return connect.NewResponse(&commonv1.StatusUpdateResponse{}), nil
}

type stubAirtel struct {
	client.AirtelClient
	resp          *client.StatusResponse
	err           error
	collections   []string
	disbursements []string
	currencies    []string
}

func (a *stubAirtel) TransactionStatus(
	_ context.Context, creds *client.AirtelCredentials, id string,
) (*client.StatusResponse, error) {
	a.collections = append(a.collections, id)
	a.currencies = append(a.currencies, creds.Currency)
	return a.resp, a.err
}

func (a *stubAirtel) DisbursementStatus(
	_ context.Context, creds *client.AirtelCredentials, id string,
) (*client.StatusResponse, error) {
	a.disbursements = append(a.disbursements, id)
	a.currencies = append(a.currencies, creds.Currency)
	return a.resp, a.err
}

func airtelResp(id, status string) *client.StatusResponse {
	r := &client.StatusResponse{}
	r.Data.Transaction.ID = id
	r.Data.Transaction.Status = status
	r.Data.Transaction.AirtelMoneyID = "am-real"
	return r
}

func resolver(connections *[]string) handlers.CredentialsResolver {
	return func(_ context.Context, connection string) (*client.AirtelCredentials, error) {
		*connections = append(*connections, connection)
		return &client.AirtelCredentials{ClientID: "id", ClientSecret: "s", Currency: "KES", CountryCode: "UG"}, nil
	}
}

func issued(entityType string) data.JSONMap {
	return data.JSONMap{
		"entity_type":                     entityType,
		client.ExtraTransactionID:         "airtel-tx-1",
		client.ExtraRequestedAmount:       "5000",
		client.ExtraRequestedCurrency:     "UGX",
		client.ExtraCredentialsConnection: "conn-1",
	}
}

func callbackBody(t *testing.T, id, statusCode string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{"transaction": map[string]any{
		"id":              id,
		"message":         "forged",
		"status_code":     statusCode,
		"airtel_money_id": map[string]any{"id": "am-forged"},
	}})
	require.NoError(t, err)
	return body
}

func post(srv *handlers.AirtelWebhookServer, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

func TestCollectionCallback_RecordsOnlyWhatAirtelConfirms(t *testing.T) {
	tests := []struct {
		name         string
		stored       data.JSONMap
		resp         *client.StatusResponse
		queryErr     error
		wantHTTP     int
		wantStatus   commonv1.STATUS
		wantNoUpdate bool
	}{
		{name: "confirmed success", stored: issued("prompt"), resp: airtelResp("prompt-1", "TS"),
			wantHTTP: http.StatusOK, wantStatus: commonv1.STATUS_SUCCESSFUL},
		{name: "forged success, Airtel says in progress", stored: issued("prompt"), resp: airtelResp("prompt-1", "TIP"),
			wantHTTP: http.StatusOK, wantStatus: commonv1.STATUS_IN_PROCESS},
		{name: "forged success, Airtel says failed", stored: issued("prompt"), resp: airtelResp("prompt-1", "TF"),
			wantHTTP: http.StatusOK, wantStatus: commonv1.STATUS_FAILED},
		{name: "Airtel reports another transaction", stored: issued("prompt"), resp: airtelResp("prompt-2", "TS"),
			wantHTTP: http.StatusConflict, wantNoUpdate: true},
		{name: "empty Airtel status is retryable", stored: issued("prompt"), resp: airtelResp("prompt-1", ""),
			wantHTTP: http.StatusServiceUnavailable, wantNoUpdate: true},
		{name: "query fails: nothing recorded, retryable", stored: issued("prompt"), queryErr: errors.New("timeout"),
			wantHTTP: http.StatusServiceUnavailable, wantNoUpdate: true},
		{name: "no Airtel request recorded", stored: data.JSONMap{"entity_type": "prompt"}, resp: airtelResp("prompt-1", "TS"),
			wantHTTP: http.StatusConflict, wantNoUpdate: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := &stubPaymentClient{stored: map[string]data.JSONMap{"prompt-1": tt.stored}}
			airtel := &stubAirtel{resp: tt.resp, err: tt.queryErr}
			var connections []string
			srv := handlers.NewAirtelWebhookServer(pay, airtel, resolver(&connections))

			rr := post(srv, "/webhook/airtel/collection?tenant_id=t1", callbackBody(t, "prompt-1", "TS"))

			require.Equal(t, tt.wantHTTP, rr.Code, rr.Body.String())
			if tt.wantNoUpdate {
				assert.Empty(t, pay.updates)
				return
			}
			assert.Equal(t, []string{"prompt-1"}, airtel.collections)
			assert.Equal(t, []string{"UGX"}, airtel.currencies, "query uses the requested currency")
			assert.Equal(t, []string{"conn-1"}, connections)
			require.Len(t, pay.updates, 1)
			got := pay.updates[0]
			assert.Equal(t, "prompt-1", got.GetId())
			assert.Equal(t, tt.wantStatus, got.GetStatus())
			assert.Equal(t, "am-real", got.GetExternalId(), "body values are never recorded")
			extras := got.GetExtras().GetFields()
			assert.Equal(t, "status_query", extras["verification"].GetStringValue())
			assert.Equal(t, "prompt-1", extras[client.ExtraTransactionID].GetStringValue(), "binding carried forward")
			assert.Equal(t, "conn-1", extras[client.ExtraCredentialsConnection].GetStringValue())
		})
	}
}

func TestDisbursementCallback_UsesDisbursementStatus(t *testing.T) {
	pay := &stubPaymentClient{stored: map[string]data.JSONMap{"payment-1": issued("payment")}}
	airtel := &stubAirtel{resp: airtelResp("payment-1", "TS")}
	var connections []string
	srv := handlers.NewAirtelWebhookServer(pay, airtel, resolver(&connections))

	rr := post(srv, "/webhook/airtel/disbursement", callbackBody(t, "payment-1", "TS"))

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, []string{"payment-1"}, airtel.disbursements)
	assert.Empty(t, airtel.collections)
	require.Len(t, pay.updates, 1)
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.updates[0].GetStatus())
}

func TestCallback_StatusUnavailableIsRetryable(t *testing.T) {
	pay := &stubPaymentClient{statusErr: errors.New("payment service down")}
	var connections []string
	srv := handlers.NewAirtelWebhookServer(pay, &stubAirtel{}, resolver(&connections))

	rr := post(srv, "/webhook/airtel/collection", callbackBody(t, "prompt-1", "TS"))

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Empty(t, pay.updates)
}

func TestCallback_RejectsMissingTransactionID(t *testing.T) {
	pay := &stubPaymentClient{}
	srv := handlers.NewAirtelWebhookServer(pay, &stubAirtel{}, nil)

	rr := post(srv, "/webhook/airtel/collection", callbackBody(t, "", "TS"))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Empty(t, pay.updates)
}
