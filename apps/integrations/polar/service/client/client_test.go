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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
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

	polarCli := client.NewClient()
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

	polarCli := client.NewClient()
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

	polarCli := client.NewClient()
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

	polarCli := client.NewClient()
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

	polarCli := client.NewClient()
	sub, err := polarCli.CancelSubscription(context.Background(), testCreds(server.URL), "sub_01abc", true)

	require.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "422")
}
