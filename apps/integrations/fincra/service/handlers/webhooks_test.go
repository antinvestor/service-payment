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
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/config"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/credentials"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/fincratest"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/handlers"
	"github.com/pitabwire/frame/v3/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePaymentClient captures StatusUpdate calls.
type fakePaymentClient struct {
	paymentv1connect.PaymentServiceClient

	statusReq *commonv1.StatusUpdateRequest
	claims    *security.AuthenticationClaims
	err       error
}

func (f *fakePaymentClient) StatusUpdate(
	ctx context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusUpdateResponse, error) {
	f.statusReq = req
	f.claims = security.ClaimsFromContext(ctx)
	if f.err != nil {
		return nil, f.err
	}
	return &commonv1.StatusUpdateResponse{}, nil
}

type harness struct {
	srv     *fincratest.Server
	payment *fakePaymentClient
	cfg     *config.FincraConfig
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	srv := fincratest.New(t)
	return &harness{
		srv:     srv,
		payment: &fakePaymentClient{},
		cfg: &config.FincraConfig{
			SecretKey: fincratest.SecretKey, PublicKey: fincratest.PublicKey, BusinessID: fincratest.BusinessID,
			WebhookSecret: fincratest.WebhookSecret, Environment: "sandbox", BaseURL: srv.URL, FeeBearer: "business",
		},
	}
}

func (h *harness) post(t *testing.T, target, body, signature string) *httptest.ResponseRecorder {
	t.Helper()
	server := handlers.NewFincraWebhookServer(h.payment, client.NewClient(), credentials.NewResolver(nil, h.cfg))
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set(client.HeaderWebhookSignature, signature)
	}
	rr := httptest.NewRecorder()
	server.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

func signed(body string) string {
	return client.WebhookSignature(fincratest.WebhookSecret, []byte(body))
}

func (h *harness) checkout(status, variance string, received float64) {
	h.srv.PutCheckout(&client.CheckoutPayment{
		ID: "1093510", Amount: 100, AmountExpected: 100, AmountReceived: client.FlexFloat(received), Currency: "NGN",
		Reference: "fcr-p-fc99410576", MerchantReference: "prompt-1", Status: status, VarianceType: variance,
		Metadata: map[string]any{"prompt_id": "prompt-1", "tenant_id": "tenant-meta", "partition_id": "partition-meta"},
	})
}

// chargeSuccessful is a checkout's charge.successful webhook whose
// reference is Fincra's pay code, as in the hosted checkout docs.
const chargeSuccessful = `{"event":"charge.successful","data":{"id":601227,"amount":100,"amountReceived":100,` +
	`"currency":"NGN","status":"success","reference":"fcr-p-fc99410576","type":"bank_transfer",` +
	`"metadata":{"prompt_id":"prompt-1"},"chargeReference":"fcr-bt-1"}}`

func TestCheckoutSuccess(t *testing.T) {
	h := newHarness(t)
	h.checkout("success", "", 100)

	rr := h.post(t, "/webhook/fincra?tenant_id=tenant-query&partition_id=partition-query", chargeSuccessful, signed(chargeSuccessful))

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	st := h.payment.statusReq
	require.NotNil(t, st)
	assert.Equal(t, "prompt-1", st.GetId())
	assert.Equal(t, "fcr-p-fc99410576", st.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, st.GetState())
	ex := st.GetExtras().AsMap()
	assert.Equal(t, "prompt", ex["entity_type"])
	assert.Equal(t, "fincra", ex["provider"])
	assert.Equal(t, "100", ex["amount"], "checkout verifies the paid amount from this key")
	assert.Equal(t, "NGN", ex["currency"])

	require.NotNil(t, h.payment.claims)
	assert.Equal(t, "tenant-meta", h.payment.claims.GetTenantID(), "tenancy stored on Fincra's record wins")
	assert.Equal(t, "partition-meta", h.payment.claims.GetPartitionID())
}

func TestCheckoutUnderpaymentFails(t *testing.T) {
	h := newHarness(t)
	h.checkout("success", "underpayment", 73)

	rr := h.post(t, "/webhook/fincra", chargeSuccessful, signed(chargeSuccessful))

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	st := h.payment.statusReq
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, st.GetState())
	ex := st.GetExtras().AsMap()
	assert.Equal(t, "AMOUNT_MISMATCH", ex["failure_code"])
	assert.Equal(t, "underpayment", ex["variance_type"])
	assert.Equal(t, "73", ex["amount_received"])
	assert.NotContains(t, ex, "amount", "an underpayment never reports the session amount as paid")
}

func TestCheckoutOverpaymentIsPaid(t *testing.T) {
	h := newHarness(t)
	h.checkout("success", "overpayment", 111)

	rr := h.post(t, "/webhook/fincra", chargeSuccessful, signed(chargeSuccessful))

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, h.payment.statusReq.GetStatus())
	assert.Equal(t, "100", h.payment.statusReq.GetExtras().AsMap()["amount"])
}

func TestStillPendingAsksForRedelivery(t *testing.T) {
	h := newHarness(t)
	h.checkout("pending", "", 0)

	rr := h.post(t, "/webhook/fincra", chargeSuccessful, signed(chargeSuccessful))

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Nil(t, h.payment.statusReq, "a body claiming success is never recorded over Fincra's own record")
}

func TestDirectChargeVerifiedByMerchantReference(t *testing.T) {
	h := newHarness(t)
	h.srv.PutCharge(&client.Charge{
		ID: "2586", Reference: "prompt-2", Status: "success", Amount: 600, AmountExpected: 600,
		Currency: "KES", Type: "mobile_money", ChargeReference: "fcr-ch-9",
	})
	body := `{"event":"charge.successful","data":{"id":2586,"status":"success","reference":"prompt-2","type":"mobile_money"}}`

	rr := h.post(t, "/webhook/fincra?tenant_id=t1&partition_id=p1", body, signed(body))

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	st := h.payment.statusReq
	assert.Equal(t, "prompt-2", st.GetId())
	assert.Equal(t, "2586", st.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	ex := st.GetExtras().AsMap()
	assert.Equal(t, "600", ex["amount"])
	assert.Equal(t, "fcr-ch-9", ex["provider_transaction_id"])
	assert.Equal(t, "t1", h.payment.claims.GetTenantID(), "direct charges take tenancy from the webhook URL")
}

func TestDirectChargeWithCustomerFeesStatesNoAmount(t *testing.T) {
	h := newHarness(t)
	h.cfg.FeeBearer = "customer"
	h.srv.PutCharge(&client.Charge{
		ID: "892892", Reference: "prompt-2", Status: "success", Amount: 101.51, AmountExpected: 101.51,
		Currency: "NGN", Type: "bank_transfer",
	})
	body := `{"event":"charge.successful","data":{"id":892892,"status":"success","reference":"prompt-2"}}`

	rr := h.post(t, "/webhook/fincra", body, signed(body))

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, h.payment.statusReq.GetStatus())
	assert.NotContains(t, h.payment.statusReq.GetExtras().AsMap(), "amount")
}

func TestFailedCharge(t *testing.T) {
	h := newHarness(t)
	h.srv.PutCharge(&client.Charge{ID: "5", Reference: "prompt-2", Status: "failed", Message: "Customer declined"})
	body := `{"event":"charge.failed","data":{"id":5,"status":"failed","reference":"prompt-2"}}`

	rr := h.post(t, "/webhook/fincra", body, signed(body))

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	ex := h.payment.statusReq.GetExtras().AsMap()
	assert.Equal(t, commonv1.STATUS_FAILED, h.payment.statusReq.GetStatus())
	assert.Equal(t, "FAILED", ex["failure_code"])
	assert.Equal(t, "Customer declined", ex["failure_message"])
}

func TestPayoutWebhooks(t *testing.T) {
	tests := []struct {
		name, status, reason string
		want                 commonv1.STATUS
		wantCode             int
	}{
		{"successful", "successful", "", commonv1.STATUS_SUCCESSFUL, http.StatusOK},
		{"failed", "failed", "Wrong beneficiary details", commonv1.STATUS_FAILED, http.StatusOK},
		{"still processing", "processing", "", commonv1.STATUS_UNKNOWN, http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.srv.PutPayout(&client.Payout{
				ID: "14380", CustomerReference: "payment-1", Reference: "bf2eb02e", Status: tt.status, Reason: tt.reason,
				AmountReceived: 500, DestinationCurrency: "KES",
			})
			body := `{"event":"payout.successful","data":{"id":14380,"status":"successful","reference":"bf2eb02e","customerReference":"payment-1"}}`

			rr := h.post(t, "/webhook/fincra?tenant_id=t1&partition_id=p1", body, signed(body))

			require.Equal(t, tt.wantCode, rr.Code, rr.Body.String())
			if tt.wantCode != http.StatusOK {
				assert.Nil(t, h.payment.statusReq)
				return
			}
			st := h.payment.statusReq
			assert.Equal(t, "payment-1", st.GetId())
			assert.Equal(t, "bf2eb02e", st.GetExternalId())
			assert.Equal(t, tt.want, st.GetStatus())
			ex := st.GetExtras().AsMap()
			assert.Equal(t, "payment", ex["entity_type"])
			assert.Equal(t, "500", ex["amount_received"])
			if tt.want == commonv1.STATUS_FAILED {
				assert.Equal(t, "PAYOUT_FAILED", ex["failure_code"])
				assert.Equal(t, tt.reason, ex["failure_message"])
			}
			assert.Equal(t, "t1", h.payment.claims.GetTenantID())
		})
	}
}

func TestRejections(t *testing.T) {
	payoutBody := `{"event":"payout.successful","data":{"customerReference":"payment-404"}}`
	noRefs := `{"event":"charge.successful","data":{"reference":"fcr-p-only"}}`
	tests := []struct {
		name      string
		mutate    func(h *harness)
		method    string
		body      string
		signature func(string) string
		want      int
	}{
		{"bad signature", nil, http.MethodPost, chargeSuccessful, func(string) string { return "ab12" }, http.StatusUnauthorized},
		{"no signature", nil, http.MethodPost, chargeSuccessful, func(string) string { return "" }, http.StatusUnauthorized},
		{"no webhook secret configured", func(h *harness) { h.cfg.WebhookSecret = "" }, http.MethodPost, chargeSuccessful, signed,
			http.StatusServiceUnavailable},
		{"no credentials", func(h *harness) { h.cfg.SecretKey = "" }, http.MethodPost, chargeSuccessful, signed,
			http.StatusServiceUnavailable},
		{"not json", nil, http.MethodPost, `nope`, signed, http.StatusBadRequest},
		{"unknown payout", nil, http.MethodPost, payoutBody, signed, http.StatusNotFound},
		{"only Fincra references", nil, http.MethodPost, noRefs, signed, http.StatusNotFound},
		{"unknown collection", nil, http.MethodPost, chargeSuccessful, signed, http.StatusNotFound},
		{"Fincra unreachable", func(h *harness) {
			h.srv.FailNext["/checkout/payments/"] = []int{500, 500, 500}
		}, http.MethodPost, chargeSuccessful, signed, http.StatusBadGateway},
		{"status update fails", func(h *harness) {
			h.checkout("success", "", 100)
			h.payment.err = errors.New("payment service down")
		}, http.MethodPost, chargeSuccessful, signed, http.StatusInternalServerError},
		{"wrong method", nil, http.MethodGet, "", signed, http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			if tt.mutate != nil {
				tt.mutate(h)
			}
			server := handlers.NewFincraWebhookServer(h.payment, client.NewClient(), credentials.NewResolver(nil, h.cfg))
			req := httptest.NewRequest(tt.method, "/webhook/fincra", strings.NewReader(tt.body))
			if sig := tt.signature(tt.body); sig != "" {
				req.Header.Set(client.HeaderWebhookSignature, sig)
			}
			rr := httptest.NewRecorder()
			server.NewRouterV1().ServeHTTP(rr, req)
			assert.Equal(t, tt.want, rr.Code, rr.Body.String())
		})
	}
}

func TestIgnoredEvents(t *testing.T) {
	h := newHarness(t)
	body := `{"event":"conversion.successful","data":{"reference":"c-1"}}`
	rr := h.post(t, "/webhook/fincra", body, signed(body))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, h.payment.statusReq)
	assert.Empty(t, h.srv.Calls, "nothing to verify")
}

func TestHealthz(t *testing.T) {
	h := newHarness(t)
	server := handlers.NewFincraWebhookServer(h.payment, client.NewClient(), credentials.NewResolver(nil, h.cfg))
	rr := httptest.NewRecorder()
	server.NewRouterV1().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	assert.Equal(t, http.StatusOK, rr.Code)
}
