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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect/v2"
	collectionv1 "github.com/antinvestor/service-payments/apps/billing/gen/collection/v1"
	"github.com/antinvestor/service-payments/apps/billing/gen/collection/v1/collectionv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/handlers"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePaymentClient struct {
	paymentv1connect.PaymentServiceClient
	statusReqs []*commonv1.StatusUpdateRequest
	err        error
}

func (f *fakePaymentClient) StatusUpdate(
	_ context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusUpdateResponse, error) {
	f.statusReqs = append(f.statusReqs, req)
	if f.err != nil {
		return nil, f.err
	}
	return &commonv1.StatusUpdateResponse{}, nil
}

func (f *fakePaymentClient) last() *commonv1.StatusUpdateRequest {
	if len(f.statusReqs) == 0 {
		return nil
	}
	return f.statusReqs[len(f.statusReqs)-1]
}

type fakeBillingClient struct {
	collectionv1connect.CollectionServiceClient
	syncReqs []*collectionv1.SyncProviderSubscriptionRequest
	err      error
}

func (f *fakeBillingClient) SyncProviderSubscription(
	_ context.Context,
	req *collectionv1.SyncProviderSubscriptionRequest,
) (*collectionv1.SyncProviderSubscriptionResponse, error) {
	f.syncReqs = append(f.syncReqs, req)
	if f.err != nil {
		return nil, f.err
	}
	return &collectionv1.SyncProviderSubscriptionResponse{SubscriptionId: "bsub", SubscriptionState: "ACTIVE", Applied: true}, nil
}

type fakePolarClient struct {
	client.PolarClient
	event *client.WebhookEvent
}

func (f *fakePolarClient) VerifyWebhookSignature(_ []byte, _ map[string]string, _ string) (*client.WebhookEvent, error) {
	return f.event, nil
}

func newWebhookServer(paymentCli *fakePaymentClient, billingCli *fakeBillingClient, event *client.WebhookEvent) *handlers.PolarWebhookServer {
	cfg := &config.PolarConfig{WebhookSecret: "test-secret"}
	var billing collectionv1connect.CollectionServiceClient
	if billingCli != nil {
		billing = billingCli
	}
	return handlers.NewPolarWebhookServer(paymentCli, billing, &fakePolarClient{event: event}, cfg)
}

func postWebhook(t *testing.T, srv *handlers.PolarWebhookServer) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhook/polar", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Webhook-Id", "msg_123")
	req.Header.Set("Webhook-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("Webhook-Signature", "v1,dummysig")
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

func subEvent(eventType, subID, status, periodEnd string, cancelAtEnd bool, meta map[string]any) *client.WebhookEvent {
	return &client.WebhookEvent{
		Type: eventType,
		Data: map[string]any{
			"id":                   subID,
			"status":               status,
			"product_id":           "prod_xyz",
			"customer_id":          "cust_123",
			"current_period_start": "2026-06-01T00:00:00Z",
			"current_period_end":   periodEnd,
			"cancel_at_period_end": cancelAtEnd,
			"modified_at":          "2026-06-02T10:00:00Z",
			"metadata":             meta,
		},
	}
}

// ─── subscription events → billing mirror ────────────────────────────────────

func TestSubscriptionActive_MirrorsToBillingAndSettlesPrompt(t *testing.T) {
	event := subEvent("subscription.active", "sub_1", "active", "2026-07-01T00:00:00Z", false, map[string]any{
		"prompt_id":       "prompt-1",
		"subscription_id": "bsub-1",
		"tenant_id":       "t1",
		"partition_id":    "p1",
	})
	pay, bill := &fakePaymentClient{}, &fakeBillingClient{}
	rr := postWebhook(t, newWebhookServer(pay, bill, event))
	require.Equal(t, http.StatusOK, rr.Code)

	require.Len(t, bill.syncReqs, 1)
	sync := bill.syncReqs[0]
	assert.Equal(t, "polar", sync.GetProvider())
	assert.Equal(t, collection.SubscriptionEventUpdated, sync.GetEvent())
	assert.Equal(t, "sub_1", sync.GetExternalSubscriptionId())
	assert.Equal(t, "cust_123", sync.GetExternalCustomerId())
	assert.Equal(t, "prod_xyz", sync.GetExternalProductId())
	assert.Equal(t, "active", sync.GetStatus())
	assert.Equal(t, "bsub-1", sync.GetSubscriptionId())
	assert.Equal(t, "2026-07-01T00:00:00Z", sync.GetCurrentPeriodEnd())
	assert.Equal(t, "2026-06-02T10:00:00Z", sync.GetOccurredAt())
	assert.False(t, sync.GetCancelAtPeriodEnd())

	st := pay.last()
	require.NotNil(t, st, "activation settles the prompt that opened the subscription")
	assert.Equal(t, "prompt-1", st.GetId())
	assert.Equal(t, "sub_1", st.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, st.GetState())
	assert.Equal(t, "sub_1", st.GetExtras().AsMap()[collection.ExtraExternalSubscriptionID])
}

func TestSubscriptionCanceled_DoesNotFailThePrompt(t *testing.T) {
	// Polar: a scheduled cancel keeps status active with cancel_at_period_end.
	event := subEvent("subscription.canceled", "sub_2", "active", "2026-07-15T00:00:00Z", true, map[string]any{
		"prompt_id": "prompt-2", "subscription_id": "bsub-2",
	})
	pay, bill := &fakePaymentClient{}, &fakeBillingClient{}
	rr := postWebhook(t, newWebhookServer(pay, bill, event))
	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, bill.syncReqs, 1)
	assert.True(t, bill.syncReqs[0].GetCancelAtPeriodEnd())
	assert.Equal(t, "active", bill.syncReqs[0].GetStatus())
	// Still active: the prompt is (re)confirmed successful, never failed.
	require.NotNil(t, pay.last())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.last().GetStatus())

	// Fully ended: billing learns, the prompt is left alone.
	event = subEvent("subscription.revoked", "sub_2", "revoked", "", false, map[string]any{"prompt_id": "prompt-2"})
	pay, bill = &fakePaymentClient{}, &fakeBillingClient{}
	rr = postWebhook(t, newWebhookServer(pay, bill, event))
	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, bill.syncReqs, 1)
	assert.Equal(t, "revoked", bill.syncReqs[0].GetStatus())
	assert.Nil(t, pay.last(), "a revoked subscription never marks the paid prompt failed")
}

func TestSubscriptionEvent_UnknownAtBillingIsDropped(t *testing.T) {
	event := subEvent("subscription.updated", "sub_3", "active", "", false, map[string]any{})
	bill := &fakeBillingClient{err: connect.NewError(connect.CodeNotFound, "nope")}
	rr := postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, event))
	assert.Equal(t, http.StatusOK, rr.Code, "an unknown subscription must not make Polar retry forever")

	bill = &fakeBillingClient{err: connect.NewError(connect.CodeUnavailable, "down")}
	rr = postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, event))
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "a billing outage asks Polar to retry")
}

func TestSubscriptionEvent_WithoutBillingClientIsAcknowledged(t *testing.T) {
	event := subEvent("subscription.updated", "sub_4", "past_due", "", false, map[string]any{})
	rr := postWebhook(t, newWebhookServer(&fakePaymentClient{}, nil, event))
	assert.Equal(t, http.StatusOK, rr.Code)
}

// ─── order events → billed ───────────────────────────────────────────────────

func TestOrderPaid_RenewalIsBilled(t *testing.T) {
	event := &client.WebhookEvent{
		Type: "order.paid",
		Data: map[string]any{
			"id":              "ord_1",
			"paid":            true,
			"billing_reason":  "subscription_cycle",
			"subscription_id": "sub_1",
			"customer_id":     "cust_123",
			"product_id":      "prod_xyz",
			"total_amount":    float64(1999),
			"currency":        "usd",
			"created_at":      "2026-07-01T00:00:05Z",
			"subscription":    map[string]any{"current_period_end": "2026-08-01T00:00:00Z"},
			"metadata":        map[string]any{"subscription_id": "bsub-1", "prompt_id": "prompt-1"},
		},
	}
	pay, bill := &fakePaymentClient{}, &fakeBillingClient{}
	rr := postWebhook(t, newWebhookServer(pay, bill, event))
	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, bill.syncReqs, 1)
	sync := bill.syncReqs[0]
	assert.Equal(t, collection.SubscriptionEventBilled, sync.GetEvent())
	assert.Equal(t, "ord_1", sync.GetExternalOrderId())
	assert.Equal(t, "bsub-1", sync.GetSubscriptionId())
	assert.Equal(t, "2026-08-01T00:00:00Z", sync.GetCurrentPeriodEnd())
	require.NotNil(t, sync.GetAmount())
	assert.Equal(t, "USD", sync.GetAmount().GetCurrencyCode())
	assert.Equal(t, int64(19), sync.GetAmount().GetUnits())
	assert.Equal(t, int32(990_000_000), sync.GetAmount().GetNanos())
	assert.Nil(t, pay.last(), "subscription orders do not touch the prompt")
}

func TestOrderPaid_OneTimePurchaseSettlesPrompt(t *testing.T) {
	event := &client.WebhookEvent{
		Type: "order.created",
		Data: map[string]any{
			"id": "ord_2", "paid": true, "billing_reason": "purchase",
			"total_amount": float64(500), "currency": "usd",
			"metadata": map[string]any{"prompt_id": "prompt-9"},
		},
	}
	pay, bill := &fakePaymentClient{}, &fakeBillingClient{}
	rr := postWebhook(t, newWebhookServer(pay, bill, event))
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, bill.syncReqs)
	require.NotNil(t, pay.last())
	assert.Equal(t, "prompt-9", pay.last().GetId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.last().GetStatus())
	assert.Equal(t, "5.00", pay.last().GetExtras().AsMap()["Amount"])
}

// ─── checkout events ─────────────────────────────────────────────────────────

func TestCheckoutUpdated_StatusMapping(t *testing.T) {
	cases := []struct {
		polar string
		want  commonv1.STATUS
	}{
		{"open", commonv1.STATUS_IN_PROCESS},
		{"confirmed", commonv1.STATUS_IN_PROCESS},
		{"succeeded", commonv1.STATUS_SUCCESSFUL},
		{"failed", commonv1.STATUS_FAILED},
		{"expired", commonv1.STATUS_FAILED},
	}
	for _, tc := range cases {
		t.Run(tc.polar, func(t *testing.T) {
			event := &client.WebhookEvent{
				Type: "checkout.updated",
				Data: map[string]any{
					"id": "chk_1", "status": tc.polar, "total_amount": float64(1000), "currency": "usd",
					"subscription_id": "sub_1",
					"metadata":        map[string]any{"prompt_id": "prompt-c"},
				},
			}
			pay := &fakePaymentClient{}
			rr := postWebhook(t, newWebhookServer(pay, nil, event))
			require.Equal(t, http.StatusOK, rr.Code)
			require.NotNil(t, pay.last())
			assert.Equal(t, tc.want, pay.last().GetStatus())
			assert.Equal(t, "chk_1", pay.last().GetExternalId())
			if tc.want == commonv1.STATUS_SUCCESSFUL {
				extras := pay.last().GetExtras().AsMap()
				assert.Equal(t, "10.00", extras["Amount"])
				assert.Equal(t, "USD", extras["Currency"])
				assert.Equal(t, "sub_1", extras[collection.ExtraExternalSubscriptionID])
			}
		})
	}
}

func TestCheckoutCreated_ForeignCheckoutIsIgnored(t *testing.T) {
	event := &client.WebhookEvent{Type: "checkout.created", Data: map[string]any{"id": "chk_x", "status": "open"}}
	pay := &fakePaymentClient{}
	rr := postWebhook(t, newWebhookServer(pay, nil, event))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, pay.last())

	event = &client.WebhookEvent{Type: "checkout.created", Data: map[string]any{
		"id": "chk_1", "status": "open", "url": "https://polar.sh/checkout/chk_1",
		"metadata": map[string]any{"prompt_id": "prompt-c"},
	}}
	pay = &fakePaymentClient{}
	rr = postWebhook(t, newWebhookServer(pay, nil, event))
	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, pay.last())
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, pay.last().GetStatus())
	assert.Equal(t, "https://polar.sh/checkout/chk_1", pay.last().GetExtras().AsMap()[collection.ExtraCheckoutURL])
}

func TestWebhook_PaymentServiceErrorRetries(t *testing.T) {
	event := &client.WebhookEvent{Type: "checkout.updated", Data: map[string]any{
		"id": "chk_1", "status": "succeeded", "metadata": map[string]any{"prompt_id": "prompt-c"},
	}}
	pay := &fakePaymentClient{err: connect.NewError(connect.CodePermissionDenied, "denied")}
	rr := postWebhook(t, newWebhookServer(pay, nil, event))
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// ─── signature verification with the real client ─────────────────────────────

func signedRequest(t *testing.T, secret []byte, body string) *http.Request {
	t.Helper()
	msgID := "msg_hmac_test"
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(msgID + "." + ts + "." + body))
	req := httptest.NewRequest(http.MethodPost, "/webhook/polar", strings.NewReader(body))
	req.Header.Set("Webhook-Id", msgID)
	req.Header.Set("Webhook-Timestamp", ts)
	req.Header.Set("Webhook-Signature", "v1,"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	return req
}

func TestHandleWebhook_RealSignature(t *testing.T) {
	body := `{"type":"checkout.updated","data":{"id":"chk_s","status":"succeeded","metadata":{"prompt_id":"p-hmac"}}}`

	t.Run("plain dashboard secret", func(t *testing.T) {
		pay := &fakePaymentClient{}
		srv := handlers.NewPolarWebhookServer(pay, nil, client.NewClient(context.Background()), &config.PolarConfig{WebhookSecret: "plain-secret"})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest(t, []byte("plain-secret"), body))
		require.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, pay.last())
		assert.Equal(t, "p-hmac", pay.last().GetId())
	})
	t.Run("whsec secret", func(t *testing.T) {
		raw := []byte("test-secret-bytes")
		pay := &fakePaymentClient{}
		srv := handlers.NewPolarWebhookServer(pay, nil, client.NewClient(context.Background()),
			&config.PolarConfig{WebhookSecret: "whsec_" + base64.StdEncoding.EncodeToString(raw)})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest(t, raw, body))
		assert.Equal(t, http.StatusOK, rr.Code)
	})
	t.Run("wrong secret", func(t *testing.T) {
		pay := &fakePaymentClient{}
		srv := handlers.NewPolarWebhookServer(pay, nil, client.NewClient(context.Background()), &config.PolarConfig{WebhookSecret: "other"})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest(t, []byte("plain-secret"), body))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Nil(t, pay.last())
	})
	t.Run("empty secret never verifies", func(t *testing.T) {
		pay := &fakePaymentClient{}
		srv := handlers.NewPolarWebhookServer(pay, nil, client.NewClient(context.Background()), &config.PolarConfig{})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest(t, []byte(""), body))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
