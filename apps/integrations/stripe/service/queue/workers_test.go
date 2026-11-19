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

package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/config"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type recordingEvents struct {
	frameEvents.Manager
	emitted []*commonv1.StatusUpdateRequest
}

func (r *recordingEvents) Emit(_ context.Context, _ string, payload any) error {
	if req, ok := payload.(*commonv1.StatusUpdateRequest); ok {
		r.emitted = append(r.emitted, req)
	}
	return nil
}

type cancelCall struct {
	id    string
	atEnd bool
}

type fakeStripe struct {
	client.StripeClient
	sessionReq *client.CheckoutSessionRequest
	sessionErr error
	intentReq  *client.PaymentIntentRequest
	intentErr  error
	payoutReq  *client.PayoutRequest
	payoutErr  error
	cancels    []cancelCall
	cancelErr  error
}

func (f *fakeStripe) CreateCheckoutSession(_ context.Context, _ *client.StripeCredentials, req *client.CheckoutSessionRequest) (*client.CheckoutSessionResponse, error) {
	f.sessionReq = req
	if f.sessionErr != nil {
		return nil, f.sessionErr
	}
	return &client.CheckoutSessionResponse{
		ID: "cs_1", URL: "https://checkout.stripe.com/c/pay/cs_1", Status: "open", Mode: req.Mode, ExpiresAt: 1_780_000_000,
	}, nil
}

func (f *fakeStripe) CreatePaymentIntent(_ context.Context, _ *client.StripeCredentials, req *client.PaymentIntentRequest) (*client.PaymentIntentResponse, error) {
	f.intentReq = req
	if f.intentErr != nil {
		return nil, f.intentErr
	}
	return &client.PaymentIntentResponse{ID: "pi_1", ClientSecret: "pi_1_secret", Status: "requires_payment_method"}, nil
}

func (f *fakeStripe) CreatePayout(_ context.Context, _ *client.StripeCredentials, req *client.PayoutRequest) (*client.PayoutResponse, error) {
	f.payoutReq = req
	if f.payoutErr != nil {
		return nil, f.payoutErr
	}
	return &client.PayoutResponse{ID: "po_1", Status: "pending", Amount: req.Amount, Currency: req.Currency}, nil
}

func (f *fakeStripe) CancelSubscription(_ context.Context, _ *client.StripeCredentials, id string, atEnd bool) (*client.SubscriptionResponse, error) {
	f.cancels = append(f.cancels, cancelCall{id: id, atEnd: atEnd})
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	return &client.SubscriptionResponse{ID: id, Status: "active", CancelAtPeriodEnd: atEnd}, nil
}

type fakeStatus struct {
	paymentv1connect.PaymentServiceClient
	extras map[string]any
}

func (f *fakeStatus) Status(_ context.Context, _ *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	st, _ := structpb.NewStruct(f.extras)
	return &commonv1.StatusResponse{Status: commonv1.STATUS_IN_PROCESS, Extras: st}, nil
}

func promptPayload(t *testing.T, amount *commonv1.Money, extras map[string]any) []byte {
	t.Helper()
	st, err := structpb.NewStruct(extras)
	require.NoError(t, err)
	raw, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:     "prompt-1",
		Source: &commonv1.ContactLink{ProfileId: "prof-1"},
		Amount: amount,
		Extra:  st,
	})
	require.NoError(t, err)
	return raw
}

func usd(units int64, nanos int32) *commonv1.Money {
	return &commonv1.Money{CurrencyCode: "USD", Units: units, Nanos: nanos}
}

func testCfg() *config.StripeConfig {
	return &config.StripeConfig{APIKey: "sk_test", WebhookSecret: "whsec_test"}
}

func rejection(status int, msg string) error {
	return &stripe.Error{HTTPStatusCode: status, Msg: msg}
}

// ─── prompt worker ───────────────────────────────────────────────────────────

func TestPrompt_OpensCheckoutSession(t *testing.T) {
	cases := []struct {
		name      string
		amount    *commonv1.Money
		extras    map[string]any
		wantMode  string
		wantPrice string
		wantMinor int64
		wantName  string
	}{
		{
			name:   "one-off payment from the prompt amount",
			amount: usd(10, 500_000_000),
			extras: map[string]any{
				collection.ExtraSuccessURL: "https://pay.example.com/c/sess",
				collection.ExtraOrderRef:   "INV-7",
			},
			wantMode: client.CheckoutModePayment, wantMinor: 1050, wantName: "INV-7",
		},
		{
			name:   "price id means a subscription",
			amount: usd(10, 0),
			extras: map[string]any{
				collection.ExtraSuccessURL: "https://pay.example.com/c/sess",
				collection.ExtraProductID:  "price_123",
			},
			wantMode: client.CheckoutModeSubscription, wantPrice: "price_123",
		},
		{
			name:   "recurring flag means a subscription",
			amount: usd(10, 0),
			extras: map[string]any{
				collection.ExtraSuccessURL: "https://pay.example.com/c/sess",
				collection.ExtraProductID:  "plan_legacy",
				collection.ExtraRecurring:  "true",
			},
			wantMode: client.CheckoutModeSubscription, wantPrice: "plan_legacy",
		},
		{
			name:   "price id with recurring=false charges once",
			amount: usd(10, 0),
			extras: map[string]any{
				collection.ExtraSuccessURL: "https://pay.example.com/c/sess",
				collection.ExtraProductID:  "price_once",
				collection.ExtraRecurring:  "false",
			},
			wantMode: client.CheckoutModePayment, wantPrice: "price_once",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := &recordingEvents{}
			st := &fakeStripe{}
			h := NewPromptHandler(ev, st, nil, &fakeStatus{}, testCfg())
			extras := map[string]any{
				collection.ExtraCustomerEmail:       "a@example.com",
				collection.ExtraCancelURL:           "https://pay.example.com/c/sess?cancel=1",
				collection.ExtraSessionRef:          "sess",
				collection.ExtraSubscriptionID:      "bsub-1",
				collection.ExtraEncryptedCardNumber: "ciphertext-must-not-leak",
				collection.ExtraPIN:                 "1234",
			}
			for k, v := range tc.extras {
				extras[k] = v
			}
			headers := map[string]string{"tenant_id": "t1", "partition_id": "p1"}
			require.NoError(t, h.Handle(context.Background(), headers, promptPayload(t, tc.amount, extras)))

			req := st.sessionReq
			require.NotNil(t, req, "a checkout session is opened")
			assert.Nil(t, st.intentReq, "no bare payment intent when a success_url is given")
			assert.Equal(t, "prompt:prompt-1", req.IdempotencyKey)
			assert.Equal(t, tc.wantMode, req.Mode)
			assert.Equal(t, tc.wantPrice, req.PriceID)
			assert.Equal(t, tc.wantMinor, req.Amount)
			if tc.wantMinor > 0 {
				assert.Equal(t, "USD", req.Currency)
				assert.Equal(t, tc.wantName, req.ProductName)
			}
			assert.Equal(t, "https://pay.example.com/c/sess", req.SuccessURL)
			assert.Equal(t, "https://pay.example.com/c/sess?cancel=1", req.CancelURL)
			assert.Equal(t, "a@example.com", req.CustomerEmail)
			assert.Equal(t, "prompt-1", req.ClientReferenceID)
			wantMeta := map[string]string{
				"prompt_id":                    "prompt-1",
				collection.ExtraSessionRef:     "sess",
				collection.ExtraSubscriptionID: "bsub-1",
				"tenant_id":                    "t1",
				"partition_id":                 "p1",
			}
			if ref, ok := tc.extras[collection.ExtraOrderRef].(string); ok {
				wantMeta[collection.ExtraOrderRef] = ref
			}
			assert.Equal(t, wantMeta, req.Metadata, "only whitelisted keys reach Stripe metadata")

			require.Len(t, ev.emitted, 1)
			status := ev.emitted[0]
			assert.Equal(t, commonv1.STATUS_IN_PROCESS, status.GetStatus())
			assert.Equal(t, "cs_1", status.GetExternalId())
			got := status.GetExtras().AsMap()
			assert.Equal(t, "cs_1", got[ExtraCheckoutSessionID])
			assert.Equal(t, "https://checkout.stripe.com/c/pay/cs_1", got[collection.ExtraCheckoutURL])
			assert.Equal(t, collection.NextActionRedirectURL, got[collection.ExtraNextAction])
			assert.Equal(t, collection.NextActionRedirectURL, got[collection.ExtraNextActionType])
			assert.Equal(t, "stripe", got[collection.ExtraProvider])
			assert.Equal(t, tc.wantMode, got[collection.ExtraMode])
			assert.Equal(t, time.Unix(1_780_000_000, 0).UTC().Format(time.RFC3339), got[collection.ExtraPaymentExpiresAt])
			assert.Equal(t, "prompt", got["entity_type"])
		})
	}
}

func TestPrompt_WithoutSuccessURLCreatesPaymentIntent(t *testing.T) {
	ev := &recordingEvents{}
	st := &fakeStripe{}
	h := NewPromptHandler(ev, st, nil, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), map[string]string{"tenant_id": "t1"},
		promptPayload(t, usd(20, 0), map[string]any{collection.ExtraInvoiceID: "inv-1", collection.ExtraOTP: "0000"})))

	assert.Nil(t, st.sessionReq)
	require.NotNil(t, st.intentReq)
	assert.Equal(t, "pi:prompt-1", st.intentReq.IdempotencyKey)
	assert.Equal(t, int64(2000), st.intentReq.Amount)
	assert.Equal(t, "USD", st.intentReq.Currency)
	assert.Equal(t, map[string]string{"prompt_id": "prompt-1", collection.ExtraInvoiceID: "inv-1", "tenant_id": "t1"},
		st.intentReq.Metadata)

	require.Len(t, ev.emitted, 1)
	got := ev.emitted[0].GetExtras().AsMap()
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, ev.emitted[0].GetStatus())
	assert.Equal(t, "pi_1", got[ExtraPaymentIntentID])
	assert.Equal(t, "pi_1_secret", got[collection.ExtraClientSecret])
}

func TestPrompt_RedeliverySkipsWhenStripeObjectExists(t *testing.T) {
	for _, key := range []string{ExtraCheckoutSessionID, ExtraPaymentIntentID} {
		t.Run(key, func(t *testing.T) {
			ev := &recordingEvents{}
			st := &fakeStripe{}
			h := NewPromptHandler(ev, st, nil, &fakeStatus{extras: map[string]any{key: "existing"}}, testCfg())
			require.NoError(t, h.Handle(context.Background(), nil,
				promptPayload(t, usd(1, 0), map[string]any{collection.ExtraSuccessURL: "https://x"})))
			assert.Nil(t, st.sessionReq, "a second session must not be opened")
			assert.Nil(t, st.intentReq)
			assert.Empty(t, ev.emitted)
		})
	}
}

func TestPrompt_UncertainErrorRedelivers_RejectionFails(t *testing.T) {
	extras := map[string]any{collection.ExtraSuccessURL: "https://x"}

	ev := &recordingEvents{}
	st := &fakeStripe{sessionErr: collection.Uncertain(rejection(502, "bad gateway"))}
	h := NewPromptHandler(ev, st, nil, nil, testCfg())
	require.Error(t, h.Handle(context.Background(), nil, promptPayload(t, usd(1, 0), extras)), "unknown outcome is redelivered")
	assert.Empty(t, ev.emitted)

	st = &fakeStripe{sessionErr: errors.New("dial tcp: i/o timeout")}
	st.sessionErr = collection.Uncertain(st.sessionErr)
	h = NewPromptHandler(ev, st, nil, nil, testCfg())
	require.Error(t, h.Handle(context.Background(), nil, promptPayload(t, usd(1, 0), extras)))
	assert.Empty(t, ev.emitted)

	st = &fakeStripe{sessionErr: rejection(400, "No such price")}
	h = NewPromptHandler(ev, st, nil, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, promptPayload(t, usd(1, 0), extras)))
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, ev.emitted[0].GetState())
	assert.Contains(t, ev.emitted[0].GetExtras().AsMap()["error"], "No such price")
}

func TestPrompt_RejectsWhatStripeCannotCharge(t *testing.T) {
	cases := []struct {
		name   string
		amount *commonv1.Money
		extras map[string]any
	}{
		{"recurring without a price", usd(5, 0), map[string]any{
			collection.ExtraSuccessURL: "https://x", collection.ExtraRecurring: "true",
		}},
		{"fractional cents", usd(5, 1_000), map[string]any{collection.ExtraSuccessURL: "https://x"}},
		{"zero amount", usd(0, 0), map[string]any{collection.ExtraSuccessURL: "https://x"}},
		{"nil amount on payment intent", nil, map[string]any{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := &recordingEvents{}
			st := &fakeStripe{}
			h := NewPromptHandler(ev, st, nil, nil, testCfg())
			require.NoError(t, h.Handle(context.Background(), nil, promptPayload(t, tc.amount, tc.extras)))
			assert.Nil(t, st.sessionReq)
			assert.Nil(t, st.intentReq)
			require.Len(t, ev.emitted, 1)
			assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())
		})
	}
}

func TestPrompt_MissingAPIKeyFails(t *testing.T) {
	ev := &recordingEvents{}
	h := NewPromptHandler(ev, &fakeStripe{}, nil, nil, &config.StripeConfig{})
	require.NoError(t, h.Handle(context.Background(), nil, promptPayload(t, usd(1, 0), map[string]any{})))
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())

	// A settings connection without a settings client must not panic.
	ev = &recordingEvents{}
	h = NewPromptHandler(ev, &fakeStripe{}, nil, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), map[string]string{config.HeaderConnectionCredentials: "conn-1"},
		promptPayload(t, usd(1, 0), map[string]any{})))
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())

	// Header credentials override the service default.
	st := &fakeStripe{}
	h = NewPromptHandler(&recordingEvents{}, st, nil, nil, &config.StripeConfig{})
	require.NoError(t, h.Handle(context.Background(), map[string]string{config.HeaderAPIKey: "sk_live"},
		promptPayload(t, usd(1, 0), map[string]any{})))
	assert.NotNil(t, st.intentReq)
}

// ─── payout worker ───────────────────────────────────────────────────────────

func paymentPayload(t *testing.T, amount *commonv1.Money, extras map[string]any) []byte {
	t.Helper()
	st, err := structpb.NewStruct(extras)
	require.NoError(t, err)
	raw, err := proto.Marshal(&paymentv1.Payment{Id: "pay-1", Amount: amount, Extra: st})
	require.NoError(t, err)
	return raw
}

func TestPayout_CreatesPayoutToDestination(t *testing.T) {
	ev := &recordingEvents{}
	st := &fakeStripe{}
	h := NewPaymentHandler(ev, st, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), map[string]string{"tenant_id": "t1"},
		paymentPayload(t, &commonv1.Money{CurrencyCode: "JPY", Units: 1500}, map[string]any{ExtraStripeDestination: "ba_1"})))

	require.NotNil(t, st.payoutReq)
	assert.Equal(t, "payout:pay-1", st.payoutReq.IdempotencyKey)
	assert.Equal(t, int64(1500), st.payoutReq.Amount, "JPY has no minor unit")
	assert.Equal(t, "JPY", st.payoutReq.Currency)
	assert.Equal(t, "ba_1", st.payoutReq.Destination)
	assert.Equal(t, map[string]string{"payment_id": "pay-1", "tenant_id": "t1"}, st.payoutReq.Metadata)

	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, ev.emitted[0].GetStatus())
	assert.Equal(t, "po_1", ev.emitted[0].GetExternalId())
	got := ev.emitted[0].GetExtras().AsMap()
	assert.Equal(t, "po_1", got[ExtraPayoutID])
	assert.Equal(t, "payment", got["entity_type"])
}

func TestPayout_Rejections(t *testing.T) {
	cases := []struct {
		name   string
		amount *commonv1.Money
		extras map[string]any
		cfg    *config.StripeConfig
	}{
		{"no destination", usd(10, 0), map[string]any{}, testCfg()},
		{"fractional amount", usd(10, 5), map[string]any{ExtraStripeDestination: "ba_1"}, testCfg()},
		{"no credentials", usd(10, 0), map[string]any{ExtraStripeDestination: "ba_1"}, &config.StripeConfig{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := &recordingEvents{}
			st := &fakeStripe{}
			h := NewPaymentHandler(ev, st, nil, tc.cfg)
			require.NoError(t, h.Handle(context.Background(), nil, paymentPayload(t, tc.amount, tc.extras)))
			assert.Nil(t, st.payoutReq, "nothing is sent to Stripe")
			require.Len(t, ev.emitted, 1)
			assert.Equal(t, "pay-1", ev.emitted[0].GetId())
			assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())
			assert.NotEmpty(t, ev.emitted[0].GetExtras().AsMap()["error"])
		})
	}
}

func TestPayout_UncertainErrorRedelivers_RejectionFails(t *testing.T) {
	extras := map[string]any{ExtraStripeDestination: "ba_1"}
	ev := &recordingEvents{}
	h := NewPaymentHandler(ev, &fakeStripe{payoutErr: collection.Uncertain(rejection(503, "down"))}, nil, testCfg())
	require.Error(t, h.Handle(context.Background(), nil, paymentPayload(t, usd(10, 0), extras)))
	assert.Empty(t, ev.emitted)

	h = NewPaymentHandler(ev, &fakeStripe{payoutErr: rejection(400, "Insufficient funds")}, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, paymentPayload(t, usd(10, 0), extras)))
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())
}

// ─── subscription command worker ─────────────────────────────────────────────

func TestSubscriptionCommand_Cancel(t *testing.T) {
	st := &fakeStripe{}
	h := NewSubscriptionHandler(st, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"stripe","command":"subscription.cancel","subscription_id":"b1","external_subscription_id":"sub_1","at_period_end":true}`)))
	require.Len(t, st.cancels, 1)
	assert.Equal(t, cancelCall{id: "sub_1", atEnd: true}, st.cancels[0])

	st = &fakeStripe{}
	h = NewSubscriptionHandler(st, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"stripe","command":"subscription.cancel","external_subscription_id":"sub_1"}`)))
	require.Len(t, st.cancels, 1)
	assert.False(t, st.cancels[0].atEnd, "no at_period_end revokes immediately")

	st = &fakeStripe{cancelErr: collection.Uncertain(errors.New("timeout"))}
	h = NewSubscriptionHandler(st, nil, testCfg())
	require.Error(t, h.Handle(context.Background(), nil, []byte(`{"provider":"stripe","command":"subscription.cancel","external_subscription_id":"sub_1"}`)))

	st = &fakeStripe{cancelErr: rejection(404, "No such subscription")}
	h = NewSubscriptionHandler(st, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"stripe","command":"subscription.cancel","external_subscription_id":"sub_1"}`)),
		"already gone at Stripe is not retried")

	st = &fakeStripe{}
	h = NewSubscriptionHandler(st, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"stripe","command":"subscription.cancel"}`)))
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"stripe","command":"nope"}`)))
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`not json`)))
	assert.Empty(t, st.cancels)

	// No credentials: acknowledged, never executed.
	st = &fakeStripe{}
	h = NewSubscriptionHandler(st, nil, &config.StripeConfig{})
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"stripe","command":"subscription.cancel","external_subscription_id":"sub_1"}`)))
	assert.Empty(t, st.cancels)
}
