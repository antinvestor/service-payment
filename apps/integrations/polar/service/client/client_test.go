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

package client_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCreds returns PolarCredentials pointed at the given test server URL.
// BaseURLOverride ensures the client does not contact the real polar API.
func testCreds(serverURL string) *client.PolarCredentials {
	return &client.PolarCredentials{
		APIKey:          "test-api-key",
		Environment:     "sandbox",
		BaseURLOverride: serverURL,
	}
}

// TestGetSubscription verifies GET /v1/subscriptions/{id} path, method, auth
// header, and that the response is correctly decoded into a Subscription.
func TestGetSubscription(t *testing.T) {
	const subID = "sub_01abc"

	responseBody := `{
		"id": "sub_01abc",
		"status": "active",
		"product_id": "prod_xyz",
		"customer_id": "cust_123",
		"current_period_end": "2026-07-01T00:00:00Z",
		"cancel_at_period_end": false,
		"metadata": {"prompt_id": "p1", "tenant_id": "t1"}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/subscriptions/"+subID, r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(responseBody))
		assert.NoError(t, err)
	}))
	defer server.Close()

	polarCli := client.NewClient(context.Background())
	sub, err := polarCli.GetSubscription(context.Background(), testCreds(server.URL), subID)

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, "sub_01abc", sub.ID)
	assert.Equal(t, "active", sub.Status)
	assert.Equal(t, "prod_xyz", sub.ProductID)
	assert.Equal(t, "cust_123", sub.CustomerID)
	assert.Equal(t, "2026-07-01T00:00:00Z", sub.CurrentPeriodEnd)
	assert.False(t, sub.CancelAtPeriodEnd)
	assert.Equal(t, "t1", sub.Metadata["tenant_id"])
}

// TestGetSubscription_Error verifies that a non-200 response returns an error.
func TestGetSubscription_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	polarCli := client.NewClient(context.Background())
	sub, err := polarCli.GetSubscription(context.Background(), testCreds(server.URL), "missing-id")

	require.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "404")
}

// TestCancelSubscription_AtPeriodEnd verifies PATCH /v1/subscriptions/{id}
// with cancel_at_period_end=true body is sent correctly.
func TestCancelSubscription_AtPeriodEnd(t *testing.T) {
	const subID = "sub_01abc"

	responseBody := `{
		"id": "sub_01abc",
		"status": "active",
		"product_id": "prod_xyz",
		"customer_id": "cust_123",
		"current_period_end": "2026-07-01T00:00:00Z",
		"cancel_at_period_end": true,
		"metadata": {}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/subscriptions/"+subID, r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		assert.Equal(t, true, payload["cancel_at_period_end"])

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(responseBody))
		assert.NoError(t, err)
	}))
	defer server.Close()

	polarCli := client.NewClient(context.Background())
	sub, err := polarCli.CancelSubscription(context.Background(), testCreds(server.URL), subID, true)

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, "sub_01abc", sub.ID)
	assert.True(t, sub.CancelAtPeriodEnd)
}

// TestCancelSubscription_Immediate verifies DELETE /v1/subscriptions/{id}/revoke
// is called for immediate cancellation.
func TestCancelSubscription_Immediate(t *testing.T) {
	const subID = "sub_01abc"

	responseBody := `{
		"id": "sub_01abc",
		"status": "canceled",
		"product_id": "prod_xyz",
		"customer_id": "cust_123",
		"current_period_end": "2026-06-13T00:00:00Z",
		"cancel_at_period_end": false,
		"metadata": {}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/subscriptions/"+subID+"/revoke", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(responseBody))
		assert.NoError(t, err)
	}))
	defer server.Close()

	polarCli := client.NewClient(context.Background())
	sub, err := polarCli.CancelSubscription(context.Background(), testCreds(server.URL), subID, false)

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, "canceled", sub.Status)
	assert.False(t, sub.CancelAtPeriodEnd)
}

// TestCancelSubscription_Error verifies that a non-200 response returns an error.
func TestCancelSubscription_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error": "already canceled"}`))
	}))
	defer server.Close()

	polarCli := client.NewClient(context.Background())
	sub, err := polarCli.CancelSubscription(context.Background(), testCreds(server.URL), "sub_01abc", true)

	require.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "422")
}

func TestCreateCheckout_PayloadAndErrors(t *testing.T) {
	var got map[string]any
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/checkouts/", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"chk_1","url":"https://polar.sh/c/chk_1","status":"open"}`))
	}))
	defer srv.Close()

	c := client.NewClient(context.Background())
	resp, err := c.CreateCheckout(context.Background(), testCreds(srv.URL), &client.CheckoutRequest{
		ProductIDs: []string{"prod_1"}, CustomerEmail: "a@b.c", ExternalCustomerID: "prof-1",
		SuccessURL: "https://pay/c/s", Metadata: map[string]string{"prompt_id": "p1"},
	})
	require.NoError(t, err)
	assert.Equal(t, "chk_1", resp.ID)
	assert.Equal(t, "Bearer test-api-key", auth)
	assert.Equal(t, []any{"prod_1"}, got["products"])
	assert.Equal(t, "prof-1", got["external_customer_id"])
	_, hasAmount := got["amount"]
	assert.False(t, hasAmount, "amount is only sent for pay-what-you-want")

	_, err = c.CreateCheckout(context.Background(), testCreds(srv.URL), &client.CheckoutRequest{})
	require.Error(t, err)
}

func TestClient_ErrorClassification(t *testing.T) {
	status := http.StatusBadGateway
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"error":"x"}`))
	}))
	defer srv.Close()
	c := client.NewClient(context.Background())

	_, err := c.GetSubscription(context.Background(), testCreds(srv.URL), "sub")
	require.Error(t, err)
	assert.True(t, collection.IsUncertain(err), "5xx leaves the outcome unknown")

	status = http.StatusNotFound
	_, err = c.GetSubscription(context.Background(), testCreds(srv.URL), "sub")
	require.Error(t, err)
	assert.False(t, collection.IsUncertain(err))
	assert.True(t, client.IsNotFound(err))
}

func TestVerifyWebhookSignature(t *testing.T) {
	c := client.NewClient(context.Background())
	body := []byte(`{"type":"checkout.updated","data":{"id":"c"}}`)
	sign := func(secret []byte, id, ts string) string {
		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(id + "." + ts + "." + string(body)))
		return base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	headers := func(sig string) map[string]string {
		return map[string]string{"Webhook-Id": "m1", "webhook-timestamp": ts, "WEBHOOK-SIGNATURE": sig}
	}

	ev, err := c.VerifyWebhookSignature(body, headers("v1,"+sign([]byte("plain"), "m1", ts)), "plain")
	require.NoError(t, err)
	assert.Equal(t, "checkout.updated", ev.Type)
	assert.Equal(t, "m1", ev.ID)
	assert.NotEmpty(t, ev.Timestamp)

	// Several signatures, one of them ours; the version tag must be v1.
	_, err = c.VerifyWebhookSignature(body, headers("v1,AAAA v1,"+sign([]byte("plain"), "m1", ts)), "plain")
	require.NoError(t, err)
	_, err = c.VerifyWebhookSignature(body, headers("v2,"+sign([]byte("plain"), "m1", ts)), "plain")
	require.Error(t, err)
	_, err = c.VerifyWebhookSignature(body, headers("v1,"+sign([]byte("other"), "m1", ts)), "plain")
	require.Error(t, err)
	_, err = c.VerifyWebhookSignature(body, headers("v1,"+sign([]byte("plain"), "m1", ts)), "")
	require.Error(t, err, "an empty secret never verifies")

	old := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	h := headers("v1," + sign([]byte("plain"), "m1", old))
	h["webhook-timestamp"] = old
	_, err = c.VerifyWebhookSignature(body, h, "plain")
	require.Error(t, err, "replayed timestamps are refused")

	raw := []byte("raw-key")
	_, err = c.VerifyWebhookSignature(body, headers("v1,"+sign(raw, "m1", ts)), "whsec_"+base64.StdEncoding.EncodeToString(raw))
	require.NoError(t, err)
}
