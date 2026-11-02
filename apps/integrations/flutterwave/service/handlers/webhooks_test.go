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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/fwtest"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/handlers"
	"github.com/pitabwire/frame/v3/security"
	"github.com/stretchr/testify/require"
)

const secret = "whsec"

// updates records the status updates sent to the payment service and the
// tenancy each was attributed to.
type updates struct {
	paymentv1connect.PaymentServiceClient

	mu      sync.Mutex
	got     []*commonv1.StatusUpdateRequest
	tenants []string
}

func (u *updates) StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusUpdateResponse, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	tenant := ""
	if c := security.ClaimsFromContext(ctx); c != nil {
		tenant = c.GetTenantID()
	}
	u.got = append(u.got, req)
	u.tenants = append(u.tenants, tenant)
	return &commonv1.StatusUpdateResponse{}, nil
}

func (u *updates) count() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.got)
}

func (u *updates) last(t *testing.T) (*commonv1.StatusUpdateRequest, string) {
	t.Helper()
	u.mu.Lock()
	defer u.mu.Unlock()
	require.NotEmpty(t, u.got, "no status update was sent")
	return u.got[len(u.got)-1], u.tenants[len(u.tenants)-1]
}

type fixture struct {
	fw      *fwtest.Server
	updates *updates
	mux     http.Handler
}

func newFixture(t *testing.T, webhookSecret string) *fixture {
	t.Helper()
	fw := fwtest.New(t)
	up := &updates{}
	cfg := &config.FlutterwaveConfig{
		ClientID:          "client-id",
		ClientSecret:      "client-secret",
		Environment:       "sandbox",
		SandboxAPIBaseURL: fw.URL,
		OAuthTokenURL:     fw.URL + "/token",
		WebhookSecret:     webhookSecret,
	}
	srv := handlers.NewFlutterwaveWebhookServer(up, client.NewClient(), cfg)
	return &fixture{fw: fw, updates: up, mux: srv.NewRouterV1()}
}

func sign(body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (f *fixture) post(t *testing.T, event map[string]any, headers map[string]string, query string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(event)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/webhook/flutterwave"+query, bytes.NewReader(body))
	if headers == nil {
		headers = map[string]string{"Flutterwave-Signature": sign(body)}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)
	return rec
}

// storeCharge puts a charge on the fake account as the worker would have made it.
func (f *fixture) storeCharge(id, status string, amount float64, meta map[string]any) {
	f.fw.Charges[id] = &client.Charge{
		ID: id, Amount: amount, Currency: "KES", Reference: "prompt-" + id, Status: status, Meta: meta,
	}
}

func promptMeta(promptID string) map[string]any {
	return map[string]any{
		"prompt_id": promptID, "tenant_id": "tenant-1", "partition_id": "partition-1",
		"expected_amount": "1500.00", "expected_currency": "KES",
	}
}

func chargeEvent(id, bodyStatus string, meta map[string]any) map[string]any {
	return map[string]any{
		"type": "charge.completed",
		"data": map[string]any{"id": id, "status": bodyStatus, "meta": meta},
	}
}

func TestWebhookReportsTheVerifiedCharge(t *testing.T) {
	f := newFixture(t, secret)
	f.storeCharge("chg_1", "succeeded", 1500, promptMeta("prompt-1"))

	// The body claims failure and another tenant; only the verified charge counts.
	forged := map[string]any{"prompt_id": "prompt-1", "tenant_id": "tenant-evil"}
	rec := f.post(t, chargeEvent("chg_1", "failed", forged), nil, "?tenant_id=tenant-evil")
	require.Equal(t, http.StatusOK, rec.Code)

	got, tenant := f.updates.last(t)
	require.Equal(t, "prompt-1", got.GetId())
	require.Equal(t, commonv1.STATUS_SUCCESSFUL, got.GetStatus())
	require.Equal(t, "chg_1", got.GetExternalId())
	require.Equal(t, "tenant-1", tenant, "tenancy comes from the verified charge")
	ex := got.GetExtras().AsMap()
	require.Equal(t, "1500.00", ex["amount"])
	require.Equal(t, "KES", ex["currency"])
}

func TestWebhookFailsAnUnderpaidCharge(t *testing.T) {
	f := newFixture(t, secret)
	f.storeCharge("chg_2", "succeeded", 150, promptMeta("prompt-2"))

	require.Equal(t, http.StatusOK, f.post(t, chargeEvent("chg_2", "succeeded", nil), nil, "").Code)
	got, _ := f.updates.last(t)
	require.Equal(t, commonv1.STATUS_FAILED, got.GetStatus())
	require.Equal(t, "amount_mismatch", got.GetExtras().AsMap()["error_code"])
}

func TestWebhookRefusals(t *testing.T) {
	t.Run("no secret configured", func(t *testing.T) {
		f := newFixture(t, "")
		f.storeCharge("chg_3", "succeeded", 1500, promptMeta("prompt-3"))
		rec := f.post(t, chargeEvent("chg_3", "succeeded", nil), map[string]string{}, "")
		require.Equal(t, http.StatusServiceUnavailable, rec.Code)
		require.Zero(t, f.updates.count())
	})
	t.Run("bad signature", func(t *testing.T) {
		f := newFixture(t, secret)
		rec := f.post(t, chargeEvent("chg_3", "succeeded", nil), map[string]string{"Flutterwave-Signature": "bad"}, "")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	t.Run("the secret itself as signature", func(t *testing.T) {
		f := newFixture(t, secret)
		rec := f.post(t, chargeEvent("chg_3", "succeeded", nil), map[string]string{"Flutterwave-Signature": secret}, "")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	t.Run("v3 verif-hash on a v4 account", func(t *testing.T) {
		f := newFixture(t, secret)
		rec := f.post(t, chargeEvent("chg_3", "succeeded", nil), map[string]string{"Verif-Hash": secret}, "")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	t.Run("unsigned", func(t *testing.T) {
		f := newFixture(t, secret)
		rec := f.post(t, chargeEvent("chg_3", "succeeded", nil), map[string]string{}, "")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestWebhookForAChargeNotOnTheAccountIsIgnored(t *testing.T) {
	f := newFixture(t, secret)
	require.Equal(t, http.StatusOK, f.post(t, chargeEvent("chg_missing", "succeeded", promptMeta("p")), nil, "").Code)
	require.Zero(t, f.updates.count())
}

func TestWebhookIsRetriedWhileFlutterwaveIsUnreachable(t *testing.T) {
	f := newFixture(t, secret)
	f.storeCharge("chg_4", "succeeded", 1500, promptMeta("prompt-4"))
	f.fw.FailN("GET /charges/chg_4", http.StatusServiceUnavailable, 3)

	rec := f.post(t, chargeEvent("chg_4", "succeeded", nil), nil, "")
	require.Equal(t, http.StatusBadGateway, rec.Code, "Flutterwave resends it")
	require.Zero(t, f.updates.count())
}

func TestWebhookReportsTheVerifiedTransfer(t *testing.T) {
	f := newFixture(t, secret)
	f.fw.Transfers["trf_1"] = &client.Transfer{
		ID: "trf_1", Reference: "pay-payment-1", Status: "SUCCESSFUL", DestinationCurrency: "UGX",
		Amount: map[string]any{"value": 50000.0},
		Meta:   map[string]any{"payment_id": "payment-1", "tenant_id": "tenant-1"},
	}
	event := map[string]any{
		"type": "transfer.disburse",
		"data": map[string]any{"id": "trf_1", "status": "FAILED", "meta": map[string]any{"payment_id": "payment-evil"}},
	}
	require.Equal(t, http.StatusOK, f.post(t, event, nil, "").Code)
	got, tenant := f.updates.last(t)
	require.Equal(t, "payment-1", got.GetId())
	require.Equal(t, commonv1.STATUS_SUCCESSFUL, got.GetStatus())
	require.Equal(t, "tenant-1", tenant)
}

func TestTransferReversalNeverRemarksAPayout(t *testing.T) {
	f := newFixture(t, secret)
	event := map[string]any{"type": "transfer.reversal", "data": map[string]any{"id": "trf_9", "reference": "pay-x"}}
	require.Equal(t, http.StatusOK, f.post(t, event, nil, "").Code)
	require.Zero(t, f.updates.count())
}

func TestReturnPageClaimsNothing(t *testing.T) {
	f := newFixture(t, secret)
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/webhook/flutterwave/return?charge_id=chg_1", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotContains(t, rec.Body.String(), "Payment received")
	require.Zero(t, f.fw.Count("GET /charges/chg_1"), "a browser visit never calls Flutterwave")
}
