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
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePaymentClient captures the most recent StatusUpdate call.
type fakePaymentClient struct {
	paymentv1connect.PaymentServiceClient
	statusReq *commonv1.StatusUpdateRequest
	err       error
}

func (f *fakePaymentClient) StatusUpdate(
	_ context.Context,
	req *connect.Request[commonv1.StatusUpdateRequest],
) (*connect.Response[commonv1.StatusUpdateResponse], error) {
	f.statusReq = req.Msg
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&commonv1.StatusUpdateResponse{}), nil
}

// fakePolarClient implements client.PolarClient. VerifyWebhookSignature
// always succeeds, returning the pre-configured event without any HMAC
// verification, so tests remain deterministic.
type fakePolarClient struct {
	client.PolarClient
	event *client.WebhookEvent
}

func (f *fakePolarClient) VerifyWebhookSignature(
	_ []byte,
	_ map[string]string,
	_ string,
) (*client.WebhookEvent, error) {
	return f.event, nil
}

func newWebhookServer(paymentCli *fakePaymentClient, event *client.WebhookEvent) *handlers.PolarWebhookServer {
	polarCli := &fakePolarClient{event: event}
	cfg := &config.PolarConfig{WebhookSecret: "whsec_" + base64.StdEncoding.EncodeToString([]byte("test-secret"))}
	return handlers.NewPolarWebhookServer(paymentCli, polarCli, cfg)
}

func postWebhook(t *testing.T, srv *handlers.PolarWebhookServer) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhook/polar", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Webhook-Id", "msg_123")
	req.Header.Set("Webhook-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	// Provide a dummy signature — fakePolarClient bypasses real verification.
	req.Header.Set("Webhook-Signature", "v1,dummysig")
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)
	return rr
}

// buildSubEvent constructs a minimal WebhookEvent for the given type, embedding
// the provided data fields. metadata is merged under the "metadata" key.
func buildSubEvent(
	eventType, subID, status, currentPeriodEnd string,
	meta map[string]string,
) *client.WebhookEvent {
	metaAny := make(map[string]any, len(meta))
	for k, v := range meta {
		metaAny[k] = v
	}
	return &client.WebhookEvent{
		Type: eventType,
		Data: map[string]any{
			"id":                   subID,
			"status":               status,
			"product_id":           "prod_xyz",
			"customer_id":          "cust_123",
			"current_period_end":   currentPeriodEnd,
			"cancel_at_period_end": false,
			"metadata":             metaAny,
		},
	}
}

// ─── subscription.created ─────────────────────────────────────────────────────

func TestHandleSubscriptionCreated(t *testing.T) {
	const (
		subID    = "sub_01abc"
		promptID = "prompt-001"
	)
	event := buildSubEvent("subscription.created", subID, "active", "2026-07-01T00:00:00Z", map[string]string{
		"prompt_id":    promptID,
		"tenant_id":    "t1",
		"partition_id": "p1",
	})

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, promptID, paymentCli.statusReq.GetId())
	assert.Equal(t, subID, paymentCli.statusReq.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, paymentCli.statusReq.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, paymentCli.statusReq.GetState())
}

// ─── subscription.updated ─────────────────────────────────────────────────────

func TestHandleSubscriptionUpdated_Active(t *testing.T) {
	event := buildSubEvent("subscription.updated", "sub_02", "active", "", map[string]string{
		"prompt_id": "p2",
	})

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, paymentCli.statusReq.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, paymentCli.statusReq.GetState())
}

func TestHandleSubscriptionUpdated_Canceled(t *testing.T) {
	event := buildSubEvent("subscription.updated", "sub_02", "canceled", "", map[string]string{
		"prompt_id": "p2",
	})

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, commonv1.STATUS_FAILED, paymentCli.statusReq.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, paymentCli.statusReq.GetState())
}

// ─── subscription.active ──────────────────────────────────────────────────────

func TestHandleSubscriptionActive(t *testing.T) {
	const (
		subID            = "sub_active_01"
		promptID         = "prompt-active"
		currentPeriodEnd = "2026-08-01T00:00:00Z"
	)
	event := buildSubEvent("subscription.active", subID, "active", currentPeriodEnd, map[string]string{
		"prompt_id":    promptID,
		"tenant_id":    "t2",
		"partition_id": "p2",
	})

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, promptID, paymentCli.statusReq.GetId())
	assert.Equal(t, subID, paymentCli.statusReq.GetExternalId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, paymentCli.statusReq.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, paymentCli.statusReq.GetState())

	extras := paymentCli.statusReq.GetExtras().AsMap()
	assert.Equal(t, "subscription.active", extras["polar_event_type"])
	assert.Equal(t, subID, extras["subscription_id"])
	assert.Equal(t, "active", extras["subscription_status"])
	assert.Equal(t, currentPeriodEnd, extras["current_period_end"])
	assert.Equal(t, "prompt", extras["entity_type"])
}

// ─── subscription.canceled ────────────────────────────────────────────────────

func TestHandleSubscriptionCanceled(t *testing.T) {
	const (
		subID            = "sub_cancel_01"
		promptID         = "prompt-cancel"
		currentPeriodEnd = "2026-07-15T00:00:00Z"
	)
	event := buildSubEvent("subscription.canceled", subID, "canceled", currentPeriodEnd, map[string]string{
		"prompt_id":    promptID,
		"tenant_id":    "t3",
		"partition_id": "p3",
	})

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)

	// ExternalId must be the polar subscription id (not prompt id)
	assert.Equal(t, promptID, paymentCli.statusReq.GetId())
	assert.Equal(t, subID, paymentCli.statusReq.GetExternalId())

	// Terminal status
	assert.Equal(t, commonv1.STATUS_FAILED, paymentCli.statusReq.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, paymentCli.statusReq.GetState())

	// Extras carry polar status + current_period_end for billing mirror
	extras := paymentCli.statusReq.GetExtras().AsMap()
	assert.Equal(t, "subscription.canceled", extras["polar_event_type"])
	assert.Equal(t, subID, extras["subscription_id"])
	assert.Equal(t, "canceled", extras["subscription_status"])
	assert.Equal(t, currentPeriodEnd, extras["current_period_end"])
	assert.Equal(t, "prompt", extras["entity_type"])
}

// TestHandleSubscriptionCanceled_FallsBackToSubIDWhenNoPromptID verifies that
// when metadata contains no prompt_id the polar subscription id is used as the
// StatusUpdate id.
func TestHandleSubscriptionCanceled_FallsBackToSubIDWhenNoPromptID(t *testing.T) {
	const subID = "sub_cancel_02"
	event := buildSubEvent("subscription.canceled", subID, "canceled", "", map[string]string{})

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, subID, paymentCli.statusReq.GetId(), "id must fall back to sub id when prompt_id is absent")
	assert.Equal(t, subID, paymentCli.statusReq.GetExternalId())
}

// ─── subscription.revoked ─────────────────────────────────────────────────────

func TestHandleSubscriptionRevoked(t *testing.T) {
	const (
		subID    = "sub_revoke_01"
		promptID = "prompt-revoke"
	)
	event := buildSubEvent("subscription.revoked", subID, "revoked", "2026-06-13T00:00:00Z", map[string]string{
		"prompt_id": promptID,
	})

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, promptID, paymentCli.statusReq.GetId())
	assert.Equal(t, subID, paymentCli.statusReq.GetExternalId())
	assert.Equal(t, commonv1.STATUS_FAILED, paymentCli.statusReq.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, paymentCli.statusReq.GetState())

	extras := paymentCli.statusReq.GetExtras().AsMap()
	assert.Equal(t, "subscription.revoked", extras["polar_event_type"])
	assert.Equal(t, subID, extras["subscription_id"])
	assert.Equal(t, "revoked", extras["subscription_status"])
	assert.Equal(t, "prompt", extras["entity_type"])
}

// ─── checkout events (regression guard) ───────────────────────────────────────

func TestHandleCheckoutCreated_Regression(t *testing.T) {
	event := &client.WebhookEvent{
		Type: "checkout.created",
		Data: map[string]any{
			"id":     "chk_001",
			"status": "open",
			"metadata": map[string]any{
				"prompt_id": "prompt-checkout",
			},
		},
	}

	paymentCli := &fakePaymentClient{}
	srv := newWebhookServer(paymentCli, event)
	rr := postWebhook(t, srv)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, "prompt-checkout", paymentCli.statusReq.GetId())
	assert.Equal(t, "chk_001", paymentCli.statusReq.GetExternalId())
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, paymentCli.statusReq.GetStatus())
}

// ─── HMAC signature verification (real client) ────────────────────────────────

// TestHandleWebhook_InvalidSignature ensures that a tampered or missing
// signature results in a 400 from the real (non-fake) polar client.
func TestHandleWebhook_InvalidSignature(t *testing.T) {
	realPolarCli := client.NewClient()
	cfg := &config.PolarConfig{WebhookSecret: "whsec_" + base64.StdEncoding.EncodeToString([]byte("test-secret"))}
	paymentCli := &fakePaymentClient{}
	srv := handlers.NewPolarWebhookServer(paymentCli, realPolarCli, cfg)

	req := httptest.NewRequest(http.MethodPost, "/webhook/polar", strings.NewReader(`{}`))
	req.Header.Set("Webhook-Id", "msg_123")
	req.Header.Set("Webhook-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("Webhook-Signature", "v1,invalidsig")
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, paymentCli.statusReq)
}

// TestHandleWebhook_ValidHMAC ensures that a correctly-signed webhook is
// accepted by the real polar client.
func TestHandleWebhook_ValidHMAC(t *testing.T) {
	rawSecret := []byte("test-secret-bytes")
	webhookSecret := "whsec_" + base64.StdEncoding.EncodeToString(rawSecret)

	msgID := "msg_hmac_test"
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	body := `{"type":"subscription.updated","data":{"id":"sub_hmac","status":"active","metadata":{"prompt_id":"p-hmac"}}}`

	signedContent := msgID + "." + ts + "." + body
	mac := hmac.New(sha256.New, rawSecret)
	mac.Write([]byte(signedContent))
	sig := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))

	realPolarCli := client.NewClient()
	cfg := &config.PolarConfig{WebhookSecret: webhookSecret}
	paymentCli := &fakePaymentClient{}
	srv := handlers.NewPolarWebhookServer(paymentCli, realPolarCli, cfg)

	req := httptest.NewRequest(http.MethodPost, "/webhook/polar", strings.NewReader(body))
	req.Header.Set("Webhook-Id", msgID)
	req.Header.Set("Webhook-Timestamp", ts)
	req.Header.Set("Webhook-Signature", sig)
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, paymentCli.statusReq)
	assert.Equal(t, "p-hmac", paymentCli.statusReq.GetId())
}
