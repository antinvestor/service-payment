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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect"
	handlers "github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/handler"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubPaymentClient serves stored prompt statuses and records updates. Any
// other method (Receive included) is inherited from the nil embedded
// interface and panics if called, which is the point: callbacks must never
// record inbound payments.
type stubPaymentClient struct {
	paymentv1connect.PaymentServiceClient
	stored    map[string]data.JSONMap
	statusErr error
	updateErr error
	updates   []*commonv1.StatusUpdateRequest
	claims    *security.AuthenticationClaims
}

func (s *stubPaymentClient) Status(_ context.Context, req *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	if s.statusErr != nil {
		return nil, s.statusErr
	}
	extras, ok := s.stored[req.GetId()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("record not found"))
	}
	return &commonv1.StatusResponse{Id: req.GetId(), Status: commonv1.STATUS_IN_PROCESS, Extras: extras.ToProtoStruct()}, nil
}

func (s *stubPaymentClient) StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusUpdateResponse, error) {
	s.updates = append(s.updates, req)
	s.claims = security.ClaimsFromContext(ctx)
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	return &commonv1.StatusUpdateResponse{}, nil
}

func pendingPrompt(ref, amount string) data.JSONMap {
	return data.JSONMap{
		"entity_type":                     "prompt",
		models.ExtraTransactionRef:        ref,
		models.ExtraRequestedAmount:       amount,
		models.ExtraRequestedCurrency:     "KES",
		models.ExtraCredentialsConnection: "tenant-conn",
		models.ExtraTransactionID:         "JTX-1",
		models.ExtraCallbackSecret:        secret,
	}
}

const secret = "0123456789abcdef0123456789abcdef"

// withSecret appends the minted per-prompt secret to a callback target.
func withSecret(target string) string {
	if strings.Contains(target, "?") {
		return target + "&cb_secret=" + secret
	}
	return target + "?cb_secret=" + secret
}

func post(cli *stubPaymentClient, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handlers.NewWebhookServer(cli).NewRouter().ServeHTTP(rr, req)
	return rr
}

func stkBody(ref string, ok bool, amount string) string {
	status := "false"
	if ok {
		status = "true"
	}
	return `{"status":` + status + `,"code":0,"message":"done","transactionReference":"` + ref +
		`","telcoReference":"TLC1","mobileNumber":"254712345678","currency":"KES","requestAmount":` + amount +
		`,"debitedAmount":` + amount + `,"charge":0,"telco":"Safaricom"}`
}

func TestStkCallbackBinding(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		body       string
		stored     map[string]data.JSONMap
		statusErr  error
		updateErr  error
		wantHTTP   int
		wantStatus commonv1.STATUS
		wantExtra  map[string]string
	}{
		{
			name:       "bound success is recorded under the prompt",
			target:     withSecret("/receivepayments?prompt_id=p1&tenant_id=t1&partition_id=pa1"),
			body:       stkBody("REF-1", true, "100.00"),
			stored:     map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100.00")},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_SUCCESSFUL,
			wantExtra:  map[string]string{"amount": "100.00", "transaction_id": "JTX-1", "credentials_connection": "tenant-conn"},
		},
		{
			name:       "equal amounts in different notation match",
			target:     withSecret("/receivepayments?prompt_id=p1"),
			body:       stkBody("REF-1", true, "100"),
			stored:     map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100.00")},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_SUCCESSFUL,
		},
		{
			name:       "failed push is recorded FAILED",
			target:     withSecret("/receivepayments?prompt_id=p1"),
			body:       stkBody("REF-1", false, "100"),
			stored:     map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED,
			wantExtra:  map[string]string{"failure_code": "JENGA_0"},
		},
		{
			name:       "amount mismatch never succeeds",
			target:     withSecret("/receivepayments?prompt_id=p1"),
			body:       stkBody("REF-1", true, "1"),
			stored:     map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED,
			wantExtra:  map[string]string{"verification": "amount_mismatch"},
		},
		{
			name:     "missing secret is forged",
			target:   "/receivepayments?prompt_id=p1",
			body:     stkBody("REF-1", true, "100"),
			stored:   map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			wantHTTP: http.StatusForbidden,
		},
		{
			name:     "wrong secret is forged",
			target:   "/receivepayments?prompt_id=p1&cb_secret=ffffffffffffffffffffffffffffffff",
			body:     stkBody("REF-1", true, "100"),
			stored:   map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			wantHTTP: http.StatusForbidden,
		},
		{
			name:     "prompt without a stored secret cannot be settled",
			target:   withSecret("/receivepayments?prompt_id=p1"),
			body:     stkBody("REF-1", true, "100"),
			stored:   map[string]data.JSONMap{"p1": {"entity_type": "prompt", models.ExtraTransactionRef: "REF-1", models.ExtraRequestedAmount: "100"}},
			wantHTTP: http.StatusForbidden,
		},
		{
			name:     "no prompt id is unbound",
			target:   "/receivepayments",
			body:     stkBody("REF-1", true, "100"),
			stored:   map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			wantHTTP: http.StatusConflict,
		},
		{
			name:     "unknown prompt is unbound",
			target:   "/receivepayments?prompt_id=nope",
			body:     stkBody("REF-1", true, "100"),
			stored:   map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			wantHTTP: http.StatusConflict,
		},
		{
			name:     "different transaction reference is unbound",
			target:   withSecret("/receivepayments?prompt_id=p1"),
			body:     stkBody("REF-FORGED", true, "100"),
			stored:   map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			wantHTTP: http.StatusConflict,
		},
		{
			name:     "prompt without a push is unbound",
			target:   withSecret("/receivepayments?prompt_id=p1"),
			body:     stkBody("REF-1", true, "100"),
			stored:   map[string]data.JSONMap{"p1": {"entity_type": "prompt", models.ExtraCallbackSecret: secret}},
			wantHTTP: http.StatusConflict,
		},
		{
			name:      "lookup failure asks for redelivery",
			target:    withSecret("/receivepayments?prompt_id=p1"),
			body:      stkBody("REF-1", true, "100"),
			statusErr: errors.New("payment service down"),
			wantHTTP:  http.StatusServiceUnavailable,
		},
		{
			name:      "status update failure is a server error",
			target:    withSecret("/receivepayments?prompt_id=p1"),
			body:      stkBody("REF-1", true, "100"),
			stored:    map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")},
			updateErr: errors.New("boom"),
			wantHTTP:  http.StatusInternalServerError,
		},
		{
			name:     "missing transaction reference",
			target:   withSecret("/receivepayments?prompt_id=p1"),
			body:     `{"status":true,"mobileNumber":"254712345678","currency":"KES"}`,
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "invalid json",
			target:   withSecret("/receivepayments?prompt_id=p1"),
			body:     "not json",
			wantHTTP: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &stubPaymentClient{stored: tt.stored, statusErr: tt.statusErr, updateErr: tt.updateErr}
			rr := post(cli, tt.target, tt.body)
			require.Equal(t, tt.wantHTTP, rr.Code, rr.Body.String())
			if tt.wantHTTP != http.StatusOK {
				if tt.updateErr == nil {
					assert.Empty(t, cli.updates, "rejected callbacks must not change any status")
				}
				return
			}
			require.Len(t, cli.updates, 1)
			got := cli.updates[0]
			assert.Equal(t, "p1", got.GetId())
			assert.Equal(t, "REF-1", got.GetExternalId())
			assert.Equal(t, tt.wantStatus, got.GetStatus())
			if tt.wantStatus == commonv1.STATUS_FAILED {
				assert.Equal(t, commonv1.STATE_INACTIVE, got.GetState())
			} else {
				assert.Equal(t, commonv1.STATE_ACTIVE, got.GetState())
			}
			extras := got.GetExtras().AsMap()
			assert.Equal(t, "prompt", extras["entity_type"])
			assert.Equal(t, "REF-1", extras[models.ExtraTransactionRef])
			for k, v := range tt.wantExtra {
				assert.Equal(t, v, extras[k], k)
			}
			if strings.Contains(tt.target, "tenant_id=t1") {
				require.NotNil(t, cli.claims)
				assert.Equal(t, "t1", cli.claims.GetTenantID())
				assert.Equal(t, "pa1", cli.claims.GetPartitionID())
			}
		})
	}
}

func generalBody(ref, status, amount string) string {
	return `{"callbackType":"PAYMENT","customer":{"name":"Jane","mobileNumber":"254712345678","reference":"C1"},` +
		`"transaction":{"date":"2026-01-01","reference":"` + ref + `","paymentMode":"MPESA","amount":` + amount +
		`,"currency":"KES","billNumber":"B1","status":"` + status + `","remarks":"ok"},` +
		`"bank":{"reference":"BNK1","transactionType":"C","account":"0011"}}`
}

func TestGeneralCallbackOnlySettlesRaisedPrompts(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		body       string
		wantHTTP   int
		wantStatus commonv1.STATUS
	}{
		{
			name:       "bound success",
			target:     withSecret("/webhook/jenga/callback?prompt_id=p1"),
			body:       generalBody("REF-1", "SUCCESS", "100"),
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_SUCCESSFUL,
		},
		{
			name:       "bound failure",
			target:     withSecret("/webhook/jenga/callback?prompt_id=p1"),
			body:       generalBody("REF-1", "FAILED", "100"),
			wantHTTP:   http.StatusOK,
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name:     "forged inbound payment is refused and never received",
			target:   "/webhook/jenga/callback",
			body:     generalBody("ANY", "SUCCESS", "100000"),
			wantHTTP: http.StatusConflict,
		},
		{
			name:     "forged reference for a known prompt is refused",
			target:   withSecret("/webhook/jenga/callback?prompt_id=p1"),
			body:     generalBody("OTHER", "SUCCESS", "100"),
			wantHTTP: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &stubPaymentClient{stored: map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")}}
			rr := post(cli, tt.target, tt.body)
			require.Equal(t, tt.wantHTTP, rr.Code, rr.Body.String())
			if tt.wantHTTP != http.StatusOK {
				assert.Empty(t, cli.updates)
				return
			}
			require.Len(t, cli.updates, 1)
			assert.Equal(t, tt.wantStatus, cli.updates[0].GetStatus())
			assert.Equal(t, "p1", cli.updates[0].GetId())
		})
	}
}

func TestCallbackBodyLimit(t *testing.T) {
	cli := &stubPaymentClient{stored: map[string]data.JSONMap{"p1": pendingPrompt("REF-1", "100")}}
	big := `{"status":true,"transactionReference":"REF-1","message":"` + strings.Repeat("x", 2<<20) + `"}`
	req := httptest.NewRequest(http.MethodPost, withSecret("/receivepayments?prompt_id=p1"), bytes.NewReader([]byte(big)))
	rr := httptest.NewRecorder()
	handlers.NewWebhookServer(cli).NewRouter().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Empty(t, cli.updates)
}
