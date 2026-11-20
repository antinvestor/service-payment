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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect/v2"
	collectionv1 "github.com/antinvestor/service-payments/apps/billing/gen/collection/v1"
	"github.com/antinvestor/service-payments/apps/billing/gen/collection/v1/collectionv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/config"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/handlers"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
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

type fakeStripeClient struct {
	client.StripeClient
	event *stripe.Event
}

func (f *fakeStripeClient) VerifyWebhookSignature(_ []byte, _, _ string) (*stripe.Event, error) {
	return f.event, nil
}

const eventCreated = int64(1_780_000_000)

var eventCreatedRFC3339 = time.Unix(eventCreated, 0).UTC().Format(time.RFC3339)

// event builds a stripe.Event carrying object as its data, the way
// stripe-go decodes a real webhook body.
func event(t *testing.T, eventType string, object map[string]any) *stripe.Event {
	t.Helper()
	raw, err := json.Marshal(object)
	require.NoError(t, err)
	return &stripe.Event{
		ID:      "evt_1",
		Type:    stripe.EventType(eventType),
		Created: eventCreated,
		Data:    &stripe.EventData{Raw: raw},
	}
}

func newWebhookServer(paymentCli *fakePaymentClient, billingCli *fakeBillingClient, ev *stripe.Event) *handlers.StripeWebhookServer {
	cfg := &config.StripeConfig{WebhookSecret: "whsec_test"}
	var billing collectionv1connect.CollectionServiceClient
	if billingCli != nil {
		billing = billingCli
	}
	return handlers.NewStripeWebhookServer(paymentCli, billing, &fakeStripeClient{event: ev}, cfg)
}

func postWebhook(t *testing.T, srv *handlers.StripeWebhookServer) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhook/stripe", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=1,v1=dummy")
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

func subscriptionObject(id, status string, cancelAtEnd bool, meta map[string]any) map[string]any {
	return map[string]any{
		"id":                   id,
		"object":               "subscription",
		"status":               status,
		"customer":             "cus_123",
		"cancel_at_period_end": cancelAtEnd,
		"metadata":             meta,
		"items": map[string]any{
			"object": "list",
			"data": []any{map[string]any{
				"id":                   "si_1",
				"current_period_start": 1_780_000_000,
				"current_period_end":   1_782_592_000,
				"price":                map[string]any{"id": "price_xyz", "object": "price"},
			}},
		},
	}
}

// ─── checkout session events ─────────────────────────────────────────────────

func TestCheckoutSession_StatusMapping(t *testing.T) {
	cases := []struct {
		eventType     string
		paymentStatus string
		want          commonv1.STATUS
	}{
		{"checkout.session.completed", "paid", commonv1.STATUS_SUCCESSFUL},
		{"checkout.session.completed", "no_payment_required", commonv1.STATUS_SUCCESSFUL},
		{"checkout.session.completed", "unpaid", commonv1.STATUS_IN_PROCESS},
		{"checkout.session.async_payment_succeeded", "paid", commonv1.STATUS_SUCCESSFUL},
		{"checkout.session.async_payment_failed", "unpaid", commonv1.STATUS_FAILED},
		{"checkout.session.expired", "unpaid", commonv1.STATUS_FAILED},
	}
	for _, tc := range cases {
		t.Run(tc.eventType+"/"+tc.paymentStatus, func(t *testing.T) {
			ev := event(t, tc.eventType, map[string]any{
				"id": "cs_1", "object": "checkout.session", "status": "complete",
				"payment_status": tc.paymentStatus, "mode": "subscription",
				"amount_total": 1999, "currency": "usd",
				"subscription": "sub_1", "customer": "cus_123", "payment_intent": "pi_1",
				"metadata": map[string]any{"prompt_id": "prompt-c", "tenant_id": "t1", "partition_id": "p1"},
			})
			pay := &fakePaymentClient{}
			rr := postWebhook(t, newWebhookServer(pay, nil, ev))
			require.Equal(t, http.StatusOK, rr.Code)
			st := pay.last()
			require.NotNil(t, st)
			assert.Equal(t, "prompt-c", st.GetId())
			assert.Equal(t, "cs_1", st.GetExternalId())
			assert.Equal(t, tc.want, st.GetStatus())
			extras := st.GetExtras().AsMap()
			assert.Equal(t, "prompt", extras["entity_type"])
			assert.Equal(t, "cs_1", extras["checkout_session_id"])
			switch tc.want {
			case commonv1.STATUS_SUCCESSFUL:
				assert.Equal(t, commonv1.STATE_ACTIVE, st.GetState())
				assert.Equal(t, "19.99", extras["Amount"])
				assert.Equal(t, "USD", extras["Currency"])
				assert.Equal(t, "sub_1", extras[collection.ExtraExternalSubscriptionID])
				assert.Equal(t, "cus_123", extras[collection.ExtraCustomerID])
				assert.Equal(t, "pi_1", extras["payment_intent_id"])
			case commonv1.STATUS_FAILED:
				assert.Equal(t, commonv1.STATE_INACTIVE, st.GetState())
				assert.NotEmpty(t, extras["error"])
			default:
				assert.Nil(t, extras["Amount"], "nothing charged yet")
			}
		})
	}
}

func TestCheckoutSession_ForeignSessionIsIgnored(t *testing.T) {
	ev := event(t, "checkout.session.completed", map[string]any{"id": "cs_x", "object": "checkout.session", "payment_status": "paid"})
	pay := &fakePaymentClient{}
	rr := postWebhook(t, newWebhookServer(pay, nil, ev))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, pay.last())
}

// ─── payment intent / payout events ──────────────────────────────────────────

func TestPaymentIntent_Events(t *testing.T) {
	ev := event(t, "payment_intent.succeeded", map[string]any{
		"id": "pi_1", "object": "payment_intent", "status": "succeeded",
		"amount": 5000, "amount_received": 5000, "currency": "eur",
		"metadata": map[string]any{"prompt_id": "prompt-p"},
	})
	pay := &fakePaymentClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, nil, ev)).Code)
	st := pay.last()
	require.NotNil(t, st)
	assert.Equal(t, "prompt-p", st.GetId())
	assert.Equal(t, "pi_1", st.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, "50.00", st.GetExtras().AsMap()["Amount"])
	assert.Equal(t, "EUR", st.GetExtras().AsMap()["Currency"])

	ev = event(t, "payment_intent.payment_failed", map[string]any{
		"id": "pi_2", "object": "payment_intent", "status": "requires_payment_method",
		"last_payment_error": map[string]any{"message": "Your card was declined.", "code": "card_declined"},
		"metadata":           map[string]any{"prompt_id": "prompt-p"},
	})
	pay = &fakePaymentClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, nil, ev)).Code)
	st = pay.last()
	require.NotNil(t, st)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "Your card was declined.", st.GetExtras().AsMap()["error"])

	// A payment intent without our prompt id is not ours.
	ev = event(t, "payment_intent.succeeded", map[string]any{"id": "pi_3", "object": "payment_intent"})
	pay = &fakePaymentClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, nil, ev)).Code)
	assert.Nil(t, pay.last())
}

func TestPayout_Events(t *testing.T) {
	ev := event(t, "payout.paid", map[string]any{
		"id": "po_1", "object": "payout", "status": "paid", "amount": 1500, "currency": "jpy",
		"metadata": map[string]any{"payment_id": "pay-1"},
	})
	pay := &fakePaymentClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, nil, ev)).Code)
	st := pay.last()
	require.NotNil(t, st)
	assert.Equal(t, "pay-1", st.GetId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, "payment", st.GetExtras().AsMap()["entity_type"])
	assert.Equal(t, "1500", st.GetExtras().AsMap()["Amount"])

	ev = event(t, "payout.failed", map[string]any{
		"id": "po_2", "object": "payout", "status": "failed", "failure_code": "account_closed",
		"failure_message": "The bank account has been closed.",
		"metadata":        map[string]any{"payment_id": "pay-1"},
	})
	pay = &fakePaymentClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, nil, ev)).Code)
	st = pay.last()
	require.NotNil(t, st)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "The bank account has been closed.", st.GetExtras().AsMap()["error"])
}

// ─── subscription events → billing mirror ────────────────────────────────────

func TestSubscriptionEvent_MirrorsToBillingAndSettlesPrompt(t *testing.T) {
	ev := event(t, "customer.subscription.created", subscriptionObject("sub_1", "active", false, map[string]any{
		"prompt_id": "prompt-1", "subscription_id": "bsub-1", "profile_id": "prof-1",
		"tenant_id": "t1", "partition_id": "p1",
	}))
	pay, bill := &fakePaymentClient{}, &fakeBillingClient{}
	rr := postWebhook(t, newWebhookServer(pay, bill, ev))
	require.Equal(t, http.StatusOK, rr.Code)

	require.Len(t, bill.syncReqs, 1)
	sync := bill.syncReqs[0]
	assert.Equal(t, "stripe", sync.GetProvider())
	assert.Equal(t, collection.SubscriptionEventUpdated, sync.GetEvent())
	assert.Equal(t, "sub_1", sync.GetExternalSubscriptionId())
	assert.Equal(t, "cus_123", sync.GetExternalCustomerId())
	assert.Equal(t, "price_xyz", sync.GetExternalProductId())
	assert.Equal(t, collection.SubscriptionStatusActive, sync.GetStatus())
	assert.Equal(t, "bsub-1", sync.GetSubscriptionId())
	assert.Equal(t, "prof-1", sync.GetProfileId())
	assert.False(t, sync.GetCancelAtPeriodEnd())
	assert.Equal(t, "2026-05-28T20:26:40Z", sync.GetCurrentPeriodStart())
	assert.Equal(t, "2026-06-27T20:26:40Z", sync.GetCurrentPeriodEnd())
	assert.Equal(t, eventCreatedRFC3339, sync.GetOccurredAt())
	assert.Equal(t, "t1", sync.GetMetadata()["tenant_id"])

	st := pay.last()
	require.NotNil(t, st, "activation settles the prompt that opened the subscription")
	assert.Equal(t, "prompt-1", st.GetId())
	assert.Equal(t, "sub_1", st.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, "sub_1", st.GetExtras().AsMap()[collection.ExtraExternalSubscriptionID])
	assert.Equal(t, "cus_123", st.GetExtras().AsMap()[collection.ExtraCustomerID])
}

func TestSubscriptionEvent_StatusNormalization(t *testing.T) {
	cases := []struct {
		eventType    string
		stripeStatus string
		want         string
		settles      bool
	}{
		{"customer.subscription.created", "trialing", collection.SubscriptionStatusTrialing, true},
		{"customer.subscription.created", "incomplete", collection.SubscriptionStatusIncomplete, false},
		{"customer.subscription.updated", "Past_Due", collection.SubscriptionStatusPastDue, false},
		{"customer.subscription.updated", "incomplete_expired", collection.SubscriptionStatusIncompleteExpired, false},
		{"customer.subscription.updated", "unpaid", collection.SubscriptionStatusUnpaid, false},
		{"customer.subscription.paused", "paused", collection.SubscriptionStatusPaused, false},
		{"customer.subscription.resumed", "active", collection.SubscriptionStatusActive, true},
		{"customer.subscription.deleted", "canceled", collection.SubscriptionStatusCanceled, false},
	}
	for _, tc := range cases {
		t.Run(tc.eventType+"/"+tc.stripeStatus, func(t *testing.T) {
			ev := event(t, tc.eventType, subscriptionObject("sub_2", tc.stripeStatus, false, map[string]any{"prompt_id": "prompt-2"}))
			pay, bill := &fakePaymentClient{}, &fakeBillingClient{}
			require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, bill, ev)).Code)
			require.Len(t, bill.syncReqs, 1)
			assert.Equal(t, tc.want, bill.syncReqs[0].GetStatus())
			if tc.settles {
				require.NotNil(t, pay.last())
				assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.last().GetStatus())
			} else {
				assert.Nil(t, pay.last(), "a non-active subscription never touches the paid prompt")
			}
		})
	}
}

func TestSubscriptionEvent_ScheduledCancelAndLegacyPeriod(t *testing.T) {
	// Scheduled cancel: still active, cancel_at_period_end set.
	ev := event(t, "customer.subscription.updated", subscriptionObject("sub_3", "active", true, map[string]any{}))
	bill := &fakeBillingClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, ev)).Code)
	require.Len(t, bill.syncReqs, 1)
	assert.True(t, bill.syncReqs[0].GetCancelAtPeriodEnd())
	assert.Equal(t, collection.SubscriptionStatusActive, bill.syncReqs[0].GetStatus())

	// Older API versions report the period on the subscription itself.
	ev = event(t, "customer.subscription.updated", map[string]any{
		"id": "sub_4", "object": "subscription", "status": "active",
		"current_period_start": 1_780_000_000, "current_period_end": 1_782_592_000,
		"metadata": map[string]any{},
	})
	bill = &fakeBillingClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, ev)).Code)
	require.Len(t, bill.syncReqs, 1)
	assert.Equal(t, "2026-06-27T20:26:40Z", bill.syncReqs[0].GetCurrentPeriodEnd())
}

func TestSubscriptionEvent_BillingErrors(t *testing.T) {
	ev := event(t, "customer.subscription.updated", subscriptionObject("sub_5", "active", false, map[string]any{}))

	bill := &fakeBillingClient{err: connect.NewError(connect.CodeNotFound, "nope")}
	rr := postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, ev))
	assert.Equal(t, http.StatusOK, rr.Code, "an unknown subscription must not make Stripe retry for days")

	bill = &fakeBillingClient{err: connect.NewError(connect.CodeUnavailable, "down")}
	rr = postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, ev))
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "a billing outage asks Stripe to retry")

	rr = postWebhook(t, newWebhookServer(&fakePaymentClient{}, nil, ev))
	assert.Equal(t, http.StatusOK, rr.Code, "no billing client: acknowledged")
}

// ─── invoice events → billed / payment_failed ────────────────────────────────

func invoiceObject(id string, amountPaid, amountDue int64, withParent bool) map[string]any {
	inv := map[string]any{
		"id": id, "object": "invoice", "status": "paid",
		"amount_paid": amountPaid, "amount_due": amountDue, "currency": "usd",
		"customer":     "cus_123",
		"period_start": 1_780_000_000, "period_end": 1_782_592_000,
		"billing_reason": "subscription_cycle",
		"metadata":       map[string]any{},
		"lines": map[string]any{
			"object": "list",
			"data": []any{map[string]any{
				"id":      "il_1",
				"object":  "line_item",
				"period":  map[string]any{"start": 1_782_592_000, "end": 1_785_270_400},
				"pricing": map[string]any{"type": "price_details", "price_details": map[string]any{"price": "price_xyz", "product": "prod_1"}},
			}},
		},
	}
	if withParent {
		inv["parent"] = map[string]any{
			"type": "subscription_details",
			"subscription_details": map[string]any{
				"subscription": "sub_1",
				"metadata":     map[string]any{"subscription_id": "bsub-1", "prompt_id": "prompt-1", "tenant_id": "t1"},
			},
		}
	}
	return inv
}

func TestInvoicePaid_RenewalIsBilled(t *testing.T) {
	ev := event(t, "invoice.paid", invoiceObject("in_1", 1999, 1999, true))
	pay, bill := &fakePaymentClient{}, &fakeBillingClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, bill, ev)).Code)
	require.Len(t, bill.syncReqs, 1)
	sync := bill.syncReqs[0]
	assert.Equal(t, collection.SubscriptionEventBilled, sync.GetEvent())
	assert.Equal(t, "sub_1", sync.GetExternalSubscriptionId())
	assert.Equal(t, "in_1", sync.GetExternalOrderId())
	assert.Equal(t, "bsub-1", sync.GetSubscriptionId())
	assert.Equal(t, "cus_123", sync.GetExternalCustomerId())
	assert.Equal(t, "price_xyz", sync.GetExternalProductId())
	assert.Equal(t, "2026-06-27T20:26:40Z", sync.GetCurrentPeriodStart(), "line period is the billed period")
	assert.Equal(t, "2026-07-28T20:26:40Z", sync.GetCurrentPeriodEnd())
	assert.Equal(t, eventCreatedRFC3339, sync.GetOccurredAt())
	require.NotNil(t, sync.GetAmount())
	assert.Equal(t, "USD", sync.GetAmount().GetCurrencyCode())
	assert.Equal(t, int64(19), sync.GetAmount().GetUnits())
	assert.Equal(t, int32(990_000_000), sync.GetAmount().GetNanos())
	assert.Equal(t, "t1", sync.GetMetadata()["tenant_id"])
	assert.Nil(t, pay.last(), "invoices do not touch the prompt")
}

func TestInvoicePaymentFailed_ReportsPaymentFailed(t *testing.T) {
	ev := event(t, "invoice.payment_failed", invoiceObject("in_2", 0, 1999, true))
	bill := &fakeBillingClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, ev)).Code)
	require.Len(t, bill.syncReqs, 1)
	assert.Equal(t, collection.SubscriptionEventPaymentFailed, bill.syncReqs[0].GetEvent())
	assert.Equal(t, "in_2", bill.syncReqs[0].GetExternalOrderId())
	require.NotNil(t, bill.syncReqs[0].GetAmount())
	assert.Equal(t, int64(19), bill.syncReqs[0].GetAmount().GetUnits(), "amount due is what failed")
}

func TestInvoice_LegacyTopLevelSubscription_AndNoSubscription(t *testing.T) {
	inv := invoiceObject("in_3", 500, 500, false)
	inv["subscription"] = "sub_legacy"
	bill := &fakeBillingClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, event(t, "invoice.paid", inv))).Code)
	require.Len(t, bill.syncReqs, 1)
	assert.Equal(t, "sub_legacy", bill.syncReqs[0].GetExternalSubscriptionId())

	bill = &fakeBillingClient{}
	require.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(&fakePaymentClient{}, bill, event(t, "invoice.paid", invoiceObject("in_4", 500, 500, false)))).Code)
	assert.Empty(t, bill.syncReqs, "a one-off invoice is not billing's subscription")
}

// ─── transport ───────────────────────────────────────────────────────────────

func TestWebhook_Transport(t *testing.T) {
	ev := event(t, "checkout.session.completed", map[string]any{
		"id": "cs_1", "object": "checkout.session", "payment_status": "paid",
		"metadata": map[string]any{"prompt_id": "prompt-c"},
	})

	t.Run("payment service error retries", func(t *testing.T) {
		pay := &fakePaymentClient{err: connect.NewError(connect.CodePermissionDenied, "denied")}
		assert.Equal(t, http.StatusInternalServerError, postWebhook(t, newWebhookServer(pay, nil, ev)).Code)
	})
	t.Run("unhandled event acknowledged", func(t *testing.T) {
		pay := &fakePaymentClient{}
		other := event(t, "charge.refunded", map[string]any{"id": "ch_1", "object": "charge"})
		assert.Equal(t, http.StatusOK, postWebhook(t, newWebhookServer(pay, nil, other)).Code)
		assert.Nil(t, pay.last())
	})
	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/webhook/stripe", nil)
		rr := httptest.NewRecorder()
		newWebhookServer(&fakePaymentClient{}, nil, ev).NewRouterV1().ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
	t.Run("body over 1 MiB rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhook/stripe", strings.NewReader(strings.Repeat("x", 1<<20+1)))
		rr := httptest.NewRecorder()
		newWebhookServer(&fakePaymentClient{}, nil, ev).NewRouterV1().ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// ─── signature verification with the real client ─────────────────────────────

func signedRequest(secret, body string, at time.Time) *http.Request {
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: []byte(body), Secret: secret, Timestamp: at})
	req := httptest.NewRequest(http.MethodPost, "/webhook/stripe", strings.NewReader(body))
	req.Header.Set("Stripe-Signature", signed.Header)
	return req
}

func TestHandleWebhook_RealSignature(t *testing.T) {
	body := `{"id":"evt_sig","object":"event","api_version":"2020-08-27","type":"checkout.session.completed",` +
		`"created":1780000000,"data":{"object":{"id":"cs_s","object":"checkout.session","payment_status":"paid",` +
		`"amount_total":1000,"currency":"usd","metadata":{"prompt_id":"p-sig"}}}}`
	const secret = "whsec_real"
	realClient := client.NewClient(context.Background())

	t.Run("valid signature", func(t *testing.T) {
		pay := &fakePaymentClient{}
		srv := handlers.NewStripeWebhookServer(pay, nil, realClient, &config.StripeConfig{WebhookSecret: secret})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest(secret, body, time.Now()))
		require.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, pay.last())
		assert.Equal(t, "p-sig", pay.last().GetId())
		assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.last().GetStatus())
		assert.Equal(t, "10.00", pay.last().GetExtras().AsMap()["Amount"])
	})
	t.Run("wrong secret", func(t *testing.T) {
		pay := &fakePaymentClient{}
		srv := handlers.NewStripeWebhookServer(pay, nil, realClient, &config.StripeConfig{WebhookSecret: "whsec_other"})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest(secret, body, time.Now()))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Nil(t, pay.last())
	})
	t.Run("stale timestamp", func(t *testing.T) {
		srv := handlers.NewStripeWebhookServer(&fakePaymentClient{}, nil, realClient, &config.StripeConfig{WebhookSecret: secret})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest(secret, body, time.Now().Add(-time.Hour)))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
	t.Run("empty secret never verifies", func(t *testing.T) {
		srv := handlers.NewStripeWebhookServer(&fakePaymentClient{}, nil, realClient, &config.StripeConfig{})
		rr := httptest.NewRecorder()
		srv.NewRouterV1().ServeHTTP(rr, signedRequest("", body, time.Now()))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
