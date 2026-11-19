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

package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/pkg/collection"
	frameclient "github.com/pitabwire/frame/v3/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func TestBuildCheckoutSessionParams(t *testing.T) {
	meta := map[string]string{"prompt_id": "p1", "subscription_id": "b1"}

	t.Run("ad-hoc payment", func(t *testing.T) {
		params, err := BuildCheckoutSessionParams(&CheckoutSessionRequest{
			IdempotencyKey: "prompt:p1", Mode: CheckoutModePayment,
			Amount: 1050, Currency: "USD", ProductName: "INV-7",
			SuccessURL: "https://s", CustomerEmail: "a@b.c", ClientReferenceID: "p1", Metadata: meta,
		})
		require.NoError(t, err)
		assert.Equal(t, "payment", *params.Mode)
		assert.Equal(t, "prompt:p1", *params.IdempotencyKey)
		assert.Equal(t, "https://s", *params.SuccessURL)
		assert.Equal(t, "https://s", *params.CancelURL, "cancel falls back to success")
		assert.Equal(t, "a@b.c", *params.CustomerEmail)
		assert.Equal(t, "p1", *params.ClientReferenceID)
		require.Len(t, params.LineItems, 1)
		item := params.LineItems[0]
		assert.Nil(t, item.Price)
		assert.Equal(t, int64(1), *item.Quantity)
		assert.Equal(t, "usd", *item.PriceData.Currency)
		assert.Equal(t, int64(1050), *item.PriceData.UnitAmount)
		assert.Equal(t, "INV-7", *item.PriceData.ProductData.Name)
		assert.Equal(t, meta, params.Metadata)
		require.NotNil(t, params.PaymentIntentData, "payment intent webhooks must carry the same keys")
		assert.Equal(t, meta, params.PaymentIntentData.Metadata)
		assert.Nil(t, params.SubscriptionData)
	})

	t.Run("subscription with price", func(t *testing.T) {
		params, err := BuildCheckoutSessionParams(&CheckoutSessionRequest{
			Mode: CheckoutModeSubscription, PriceID: "price_1", SuccessURL: "https://s", CancelURL: "https://c", Metadata: meta,
		})
		require.NoError(t, err)
		assert.Equal(t, "subscription", *params.Mode)
		assert.Equal(t, "https://c", *params.CancelURL)
		assert.Equal(t, "price_1", *params.LineItems[0].Price)
		assert.Nil(t, params.LineItems[0].PriceData)
		require.NotNil(t, params.SubscriptionData, "subscription webhooks must carry the same keys")
		assert.Equal(t, meta, params.SubscriptionData.Metadata)
		assert.Nil(t, params.PaymentIntentData)
		assert.Nil(t, params.IdempotencyKey)
	})

	t.Run("rejections", func(t *testing.T) {
		cases := []struct {
			name string
			req  *CheckoutSessionRequest
		}{
			{"nil", nil},
			{"no success url", &CheckoutSessionRequest{Amount: 1, Currency: "usd"}},
			{"subscription without price", &CheckoutSessionRequest{Mode: CheckoutModeSubscription, SuccessURL: "https://s", Amount: 1, Currency: "usd"}},
			{"payment without amount", &CheckoutSessionRequest{SuccessURL: "https://s", Currency: "usd"}},
			{"payment without currency", &CheckoutSessionRequest{SuccessURL: "https://s", Amount: 1}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := BuildCheckoutSessionParams(tc.req)
				assert.Error(t, err)
			})
		}
	})
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		uncertain bool
	}{
		{"nil", nil, false},
		{"400 rejection", &stripe.Error{HTTPStatusCode: 400, Msg: "No such price"}, false},
		{"402 card declined", &stripe.Error{HTTPStatusCode: 402, Code: stripe.ErrorCodeCardDeclined}, false},
		{"404", &stripe.Error{HTTPStatusCode: 404}, false},
		{"429 lock", &stripe.Error{HTTPStatusCode: 429}, true},
		{"500", &stripe.Error{HTTPStatusCode: 500}, true},
		{"503", &stripe.Error{HTTPStatusCode: 503}, true},
		{"stripe error without status", &stripe.Error{}, true},
		{"timeout", &net.DNSError{IsTimeout: true}, true},
		{"context deadline", context.DeadlineExceeded, true},
		{"plain error", errors.New("connection reset"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classify("op", tc.err)
			if tc.err == nil {
				assert.NoError(t, got)
				return
			}
			require.Error(t, got)
			assert.Equal(t, tc.uncertain, collection.IsUncertain(got))
			assert.ErrorIs(t, got, tc.err, "the Stripe error stays reachable")
		})
	}
	assert.True(t, IsNotFound(classify("op", &stripe.Error{HTTPStatusCode: 404})))
	assert.False(t, IsNotFound(classify("op", &stripe.Error{HTTPStatusCode: 400})))
}

// TestCreateCheckoutSession_Transport drives the real stripe-go client at a
// local server: the Idempotency-Key header and the form body are what Stripe
// sees, and 5xx answers are reported as outcome-unknown.
func TestCreateCheckoutSession_Transport(t *testing.T) {
	var gotIdempotency, gotAuth string
	var gotForm url.Values
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIdempotency = r.Header.Get("Idempotency-Key")
		gotAuth = r.Header.Get("Authorization")
		require.NoError(t, r.ParseForm())
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(`{"id":"cs_test","object":"checkout.session","url":"https://checkout.stripe.com/c/pay/cs_test","status":"open","mode":"subscription","expires_at":1780000000}`))
			return
		}
		_, _ = w.Write([]byte(`{"error":{"type":"api_error","message":"boom"}}`))
	}))
	defer server.Close()

	httpClient := frameclient.NewHTTPClient(context.Background(), frameclient.WithHTTPTimeout(5*time.Second))
	backend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		HTTPClient: httpClient, URL: stripe.String(server.URL), MaxNetworkRetries: stripe.Int64(0),
	})
	sc := &stripeClient{
		backends: &stripe.Backends{API: backend, Connect: backend, Uploads: backend, MeterEvents: backend},
		clients:  map[string]*stripe.Client{},
	}
	creds := &StripeCredentials{APIKey: "sk_test_123"}
	req := &CheckoutSessionRequest{
		IdempotencyKey: "prompt:p1", Mode: CheckoutModeSubscription, PriceID: "price_1",
		SuccessURL: "https://s", Metadata: map[string]string{"prompt_id": "p1"},
	}

	resp, err := sc.CreateCheckoutSession(context.Background(), creds, req)
	require.NoError(t, err)
	assert.Equal(t, "cs_test", resp.ID)
	assert.Equal(t, "https://checkout.stripe.com/c/pay/cs_test", resp.URL)
	assert.Equal(t, "subscription", resp.Mode)
	assert.Equal(t, int64(1780000000), resp.ExpiresAt)
	assert.Equal(t, "prompt:p1", gotIdempotency)
	assert.Equal(t, "Bearer sk_test_123", gotAuth)
	assert.Equal(t, "subscription", gotForm.Get("mode"))
	assert.Equal(t, "price_1", gotForm.Get("line_items[0][price]"))
	assert.Equal(t, "1", gotForm.Get("line_items[0][quantity]"))
	assert.Equal(t, "p1", gotForm.Get("metadata[prompt_id]"))
	assert.Equal(t, "p1", gotForm.Get("subscription_data[metadata][prompt_id]"))

	status = http.StatusBadGateway
	_, err = sc.CreateCheckoutSession(context.Background(), creds, req)
	require.Error(t, err)
	assert.True(t, collection.IsUncertain(err), "a 5xx leaves the outcome unknown")

	status = http.StatusBadRequest
	_, err = sc.CreateCheckoutSession(context.Background(), creds, req)
	require.Error(t, err)
	assert.False(t, collection.IsUncertain(err), "a 4xx is a rejection")
}

func TestVerifyWebhookSignature(t *testing.T) {
	c := NewClient(context.Background())
	body := []byte(`{"id":"evt_1","object":"event","api_version":"2019-01-01","type":"payout.paid","data":{"object":{"id":"po_1","object":"payout"}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: body, Secret: "whsec_x"})

	ev, err := c.VerifyWebhookSignature(body, signed.Header, "whsec_x")
	require.NoError(t, err, "an old event API version is tolerated")
	assert.Equal(t, "evt_1", ev.ID)
	assert.Equal(t, "payout.paid", string(ev.Type))
	assert.Equal(t, "po_1", ev.Data.Object["id"])

	_, err = c.VerifyWebhookSignature(body, signed.Header, "whsec_other")
	assert.Error(t, err)
	_, err = c.VerifyWebhookSignature(body, signed.Header, "")
	assert.Error(t, err, "an empty secret verifies nothing")
}
