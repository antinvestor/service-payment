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
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/handlers"
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

type stubMtn struct {
	client.MtnClient
	rtp       *client.RequestToPayStatus
	transfer  *client.TransferStatus
	err       error
	queriedID []string
}

func (m *stubMtn) GetRequestToPayStatus(
	_ context.Context, _ *client.MtnCredentials, referenceID string,
) (*client.RequestToPayStatus, error) {
	m.queriedID = append(m.queriedID, referenceID)
	return m.rtp, m.err
}

func (m *stubMtn) GetTransferStatus(
	_ context.Context, _ *client.MtnCredentials, referenceID string,
) (*client.TransferStatus, error) {
	m.queriedID = append(m.queriedID, referenceID)
	return m.transfer, m.err
}

func resolver(connections *[]string) handlers.CredentialsResolver {
	return func(_ context.Context, connection string) (*client.MtnCredentials, error) {
		*connections = append(*connections, connection)
		return &client.MtnCredentials{SubscriptionKey: "k", APIUser: "u", APIKey: "a"}, nil
	}
}

func issued(entityType string) data.JSONMap {
	return data.JSONMap{
		"entity_type":                     entityType,
		client.ExtraReferenceID:           "ref-1",
		client.ExtraRequestedAmount:       "5000",
		client.ExtraRequestedCurrency:     "UGX",
		client.ExtraCredentialsConnection: "conn-1",
	}
}

func callbackBody(t *testing.T, externalID, status string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"financialTransactionId": "fin-forged",
		"externalId":             externalID,
		"amount":                 "5000",
		"currency":               "UGX",
		"status":                 status,
	})
	require.NoError(t, err)
	return body
}

func TestCollectionCallback_RecordsOnlyWhatMTNConfirms(t *testing.T) {
	tests := []struct {
		name         string
		stored       data.JSONMap
		bodyStatus   string
		rtp          *client.RequestToPayStatus
		queryErr     error
		wantHTTP     int
		wantStatus   commonv1.STATUS
		wantNoUpdate bool
		wantVerify   string
	}{
		{
			name:       "confirmed success",
			stored:     issued("prompt"),
			bodyStatus: "SUCCESSFUL",
			rtp:        &client.RequestToPayStatus{Status: "SUCCESSFUL", Amount: "5000", Currency: "UGX", ExternalID: "prompt-1", FinancialTransactionID: "fin-1"},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_SUCCESSFUL,
			wantVerify: "status_query",
		},
		{
			name:       "forged success, MTN says pending",
			stored:     issued("prompt"),
			bodyStatus: "SUCCESSFUL",
			rtp:        &client.RequestToPayStatus{Status: "PENDING", Amount: "5000", Currency: "UGX", ExternalID: "prompt-1"},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_IN_PROCESS,
			wantVerify: "status_query",
		},
		{
			name:       "forged success, MTN says failed",
			stored:     issued("prompt"),
			bodyStatus: "SUCCESSFUL",
			rtp:        &client.RequestToPayStatus{Status: "FAILED", ExternalID: "prompt-1"},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name:       "confirmed amount differs",
			stored:     issued("prompt"),
			bodyStatus: "SUCCESSFUL",
			rtp:        &client.RequestToPayStatus{Status: "SUCCESSFUL", Amount: "50", Currency: "UGX", ExternalID: "prompt-1"},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED,
			wantVerify: "amount_mismatch",
		},
		{
			name:       "confirmed currency differs",
			stored:     issued("prompt"),
			bodyStatus: "SUCCESSFUL",
			rtp:        &client.RequestToPayStatus{Status: "SUCCESSFUL", Amount: "5000", Currency: "EUR", ExternalID: "prompt-1"},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED,
			wantVerify: "amount_mismatch",
		},
		{
			name:         "MTN reports another externalId",
			stored:       issued("prompt"),
			bodyStatus:   "SUCCESSFUL",
			rtp:          &client.RequestToPayStatus{Status: "SUCCESSFUL", Amount: "5000", Currency: "UGX", ExternalID: "prompt-other"},
			wantHTTP:     http.StatusConflict,
			wantNoUpdate: true,
		},
		{
			name:         "query fails: nothing recorded, retryable",
			stored:       issued("prompt"),
			bodyStatus:   "SUCCESSFUL",
			queryErr:     errors.New("timeout"),
			wantHTTP:     http.StatusServiceUnavailable,
			wantNoUpdate: true,
		},
		{
			name:         "no reference id recorded",
			stored:       data.JSONMap{"entity_type": "prompt"},
			bodyStatus:   "SUCCESSFUL",
			wantHTTP:     http.StatusConflict,
			wantNoUpdate: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := &stubPaymentClient{stored: map[string]data.JSONMap{"prompt-1": tt.stored}}
			mtn := &stubMtn{rtp: tt.rtp, err: tt.queryErr}
			var connections []string
			srv := handlers.NewMtnWebhookServer(pay, mtn, resolver(&connections))

			req := httptest.NewRequest(http.MethodPost, "/webhook/mtn/collection?tenant_id=t1",
				bytes.NewReader(callbackBody(t, "prompt-1", tt.bodyStatus)))
			rr := httptest.NewRecorder()
			srv.NewRouterV1().ServeHTTP(rr, req)

			require.Equal(t, tt.wantHTTP, rr.Code, rr.Body.String())
			if tt.wantNoUpdate {
				assert.Empty(t, pay.updates)
				return
			}
			assert.Equal(t, []string{"ref-1"}, mtn.queriedID)
			assert.Equal(t, []string{"conn-1"}, connections)
			require.Len(t, pay.updates, 1)
			got := pay.updates[0]
			assert.Equal(t, "prompt-1", got.GetId())
			assert.Equal(t, tt.wantStatus, got.GetStatus())
			extras := got.GetExtras().GetFields()
			assert.Equal(t, "prompt", extras["entity_type"].GetStringValue())
			assert.Equal(t, "ref-1", extras[client.ExtraReferenceID].GetStringValue(), "binding carried forward")
			assert.NotEqual(t, "fin-forged", got.GetExternalId(), "body values are never recorded")
			if tt.wantVerify != "" {
				assert.Equal(t, tt.wantVerify, extras["verification"].GetStringValue())
			}
		})
	}
}

func TestCallback_StatusUnavailableIsRetryable(t *testing.T) {
	pay := &stubPaymentClient{statusErr: errors.New("payment service down")}
	var connections []string
	srv := handlers.NewMtnWebhookServer(pay, &stubMtn{}, resolver(&connections))

	req := httptest.NewRequest(http.MethodPost, "/webhook/mtn/collection",
		bytes.NewReader(callbackBody(t, "prompt-1", "SUCCESSFUL")))
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Empty(t, pay.updates)
}

func TestDisbursementCallback_UsesTransferStatus(t *testing.T) {
	pay := &stubPaymentClient{stored: map[string]data.JSONMap{"payment-1": issued("payment")}}
	mtn := &stubMtn{transfer: &client.TransferStatus{
		Status: "SUCCESSFUL", Amount: "5000", Currency: "UGX", ExternalID: "payment-1", FinancialTransactionID: "fin-9",
	}}
	var connections []string
	srv := handlers.NewMtnWebhookServer(pay, mtn, resolver(&connections))

	req := httptest.NewRequest(http.MethodPost, "/webhook/mtn/disbursement",
		bytes.NewReader(callbackBody(t, "payment-1", "SUCCESSFUL")))
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, pay.updates, 1)
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.updates[0].GetStatus())
	assert.Equal(t, "fin-9", pay.updates[0].GetExternalId())
	assert.Equal(t, "payment", pay.updates[0].GetExtras().GetFields()["entity_type"].GetStringValue())
}

func TestCallback_RejectsMissingExternalID(t *testing.T) {
	pay := &stubPaymentClient{}
	srv := handlers.NewMtnWebhookServer(pay, &stubMtn{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook/mtn/collection",
		bytes.NewReader(callbackBody(t, "", "SUCCESSFUL")))
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Empty(t, pay.updates)
}
