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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type fakePolar struct {
	client.PolarClient
	checkoutReq *client.CheckoutRequest
	checkoutErr error
	cancelCalls []struct {
		id    string
		atEnd bool
	}
	cancelErr error
}

func (f *fakePolar) CreateCheckout(_ context.Context, _ *client.PolarCredentials, req *client.CheckoutRequest) (*client.CheckoutResponse, error) {
	f.checkoutReq = req
	if f.checkoutErr != nil {
		return nil, f.checkoutErr
	}
	return &client.CheckoutResponse{ID: "chk_1", URL: "https://polar.sh/checkout/chk_1", Status: "open", ExpiresAt: "2026-06-01T01:00:00Z"}, nil
}

func (f *fakePolar) CancelSubscription(_ context.Context, _ *client.PolarCredentials, id string, atEnd bool) (*client.Subscription, error) {
	f.cancelCalls = append(f.cancelCalls, struct {
		id    string
		atEnd bool
	}{id, atEnd})
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	return &client.Subscription{ID: id, Status: "active", CancelAtPeriodEnd: atEnd}, nil
}

type fakeStatus struct {
	paymentv1connect.PaymentServiceClient
	extras map[string]any
}

func (f *fakeStatus) Status(_ context.Context, _ *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	st, _ := structpb.NewStruct(f.extras)
	return &commonv1.StatusResponse{Status: commonv1.STATUS_IN_PROCESS, Extras: st}, nil
}

func promptPayload(t *testing.T, extras map[string]any) []byte {
	t.Helper()
	st, err := structpb.NewStruct(extras)
	require.NoError(t, err)
	raw, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:     "prompt-1",
		Source: &commonv1.ContactLink{ProfileId: "prof-1"},
		Amount: &commonv1.Money{CurrencyCode: "USD", Units: 10},
		Extra:  st,
	})
	require.NoError(t, err)
	return raw
}

func testCfg() *config.PolarConfig {
	return &config.PolarConfig{APIKey: "key", Environment: "sandbox", WebhookSecret: "s"}
}

func TestPrompt_OpensCheckoutFromPortableExtras(t *testing.T) {
	ev := &recordingEvents{}
	polar := &fakePolar{}
	h := NewPromptHandler(ev, polar, nil, &fakeStatus{}, testCfg())

	err := h.Handle(context.Background(), map[string]string{"tenant_id": "t1", "partition_id": "p1"}, promptPayload(t, map[string]any{
		collection.ExtraProductID:      "prod_1",
		collection.ExtraCustomerEmail:  "a@example.com",
		collection.ExtraCustomerName:   "Alice",
		collection.ExtraSuccessURL:     "https://pay.example.com/c/sess",
		collection.ExtraSessionRef:     "sess",
		collection.ExtraSubscriptionID: "bsub-1",
	}))
	require.NoError(t, err)
	require.NotNil(t, polar.checkoutReq)
	assert.Equal(t, []string{"prod_1"}, polar.checkoutReq.ProductIDs)
	assert.Equal(t, "a@example.com", polar.checkoutReq.CustomerEmail)
	assert.Equal(t, "Alice", polar.checkoutReq.CustomerName)
	assert.Equal(t, "prof-1", polar.checkoutReq.ExternalCustomerID)
	assert.Equal(t, "https://pay.example.com/c/sess", polar.checkoutReq.SuccessURL)
	assert.Equal(t, "prompt-1", polar.checkoutReq.Metadata["prompt_id"])
	assert.Equal(t, "bsub-1", polar.checkoutReq.Metadata[collection.ExtraSubscriptionID])
	assert.Equal(t, "t1", polar.checkoutReq.Metadata["tenant_id"])

	require.Len(t, ev.emitted, 1)
	st := ev.emitted[0]
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, "chk_1", st.GetExternalId())
	extras := st.GetExtras().AsMap()
	assert.Equal(t, "https://polar.sh/checkout/chk_1", extras[collection.ExtraCheckoutURL])
	assert.Equal(t, collection.NextActionRedirectURL, extras[collection.ExtraNextAction])
	assert.Equal(t, "prompt", extras["entity_type"])
}

func TestPrompt_FailsWithoutProductID(t *testing.T) {
	ev := &recordingEvents{}
	polar := &fakePolar{}
	h := NewPromptHandler(ev, polar, nil, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, promptPayload(t, map[string]any{})))
	assert.Nil(t, polar.checkoutReq)
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, ev.emitted[0].GetState())
}

func TestPrompt_RedeliverySkipsWhenCheckoutExists(t *testing.T) {
	ev := &recordingEvents{}
	polar := &fakePolar{}
	h := NewPromptHandler(ev, polar, nil, &fakeStatus{extras: map[string]any{ExtraCheckoutID: "chk_old"}}, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, promptPayload(t, map[string]any{collection.ExtraProductID: "prod_1"})))
	assert.Nil(t, polar.checkoutReq, "a second checkout must not be opened")
	assert.Empty(t, ev.emitted)
}

func TestPrompt_UncertainErrorRedelivers_RejectionFails(t *testing.T) {
	ev := &recordingEvents{}
	polar := &fakePolar{checkoutErr: collection.Uncertain(errors.New("502"))}
	h := NewPromptHandler(ev, polar, nil, nil, testCfg())
	err := h.Handle(context.Background(), nil, promptPayload(t, map[string]any{collection.ExtraProductID: "prod_1"}))
	require.Error(t, err, "unknown outcome is redelivered")
	assert.Empty(t, ev.emitted)

	polar = &fakePolar{checkoutErr: &client.APIError{StatusCode: 422, Body: "bad product"}}
	h = NewPromptHandler(ev, polar, nil, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, promptPayload(t, map[string]any{collection.ExtraProductID: "prod_1"})))
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())
}

func TestPrompt_MissingAPIKeyFails(t *testing.T) {
	ev := &recordingEvents{}
	h := NewPromptHandler(ev, &fakePolar{}, nil, nil, &config.PolarConfig{Environment: "sandbox"})
	require.NoError(t, h.Handle(context.Background(), nil, promptPayload(t, map[string]any{collection.ExtraProductID: "p"})))
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())

	// A settings connection without a settings client must not panic.
	h = NewPromptHandler(ev, &fakePolar{}, nil, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), map[string]string{config.HeaderConnectionCredentials: "conn-1"},
		promptPayload(t, map[string]any{collection.ExtraProductID: "p"})))
}

func TestSubscriptionCommand_Cancel(t *testing.T) {
	polar := &fakePolar{}
	h := NewSubscriptionHandler(polar, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"polar","command":"subscription.cancel","subscription_id":"b1","external_subscription_id":"sub_1","at_period_end":true}`)))
	require.Len(t, polar.cancelCalls, 1)
	assert.Equal(t, "sub_1", polar.cancelCalls[0].id)
	assert.True(t, polar.cancelCalls[0].atEnd)

	polar = &fakePolar{cancelErr: collection.Uncertain(errors.New("timeout"))}
	h = NewSubscriptionHandler(polar, nil, testCfg())
	require.Error(t, h.Handle(context.Background(), nil, []byte(`{"provider":"polar","command":"subscription.cancel","external_subscription_id":"sub_1"}`)))

	polar = &fakePolar{cancelErr: &client.APIError{StatusCode: 404}}
	h = NewSubscriptionHandler(polar, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"polar","command":"subscription.cancel","external_subscription_id":"sub_1"}`)),
		"already gone at Polar is not retried")

	polar = &fakePolar{}
	h = NewSubscriptionHandler(polar, nil, testCfg())
	require.NoError(t, h.Handle(context.Background(), nil, []byte(`{"provider":"polar","command":"nope"}`)))
	assert.Empty(t, polar.cancelCalls)
}

func TestPayment_AlwaysRefused(t *testing.T) {
	ev := &recordingEvents{}
	raw, err := proto.Marshal(&paymentv1.Payment{Id: "pay-1"})
	require.NoError(t, err)
	require.NoError(t, NewPaymentHandler(ev).Handle(context.Background(), nil, raw))
	require.Len(t, ev.emitted, 1)
	assert.Equal(t, "pay-1", ev.emitted[0].GetId())
	assert.Equal(t, commonv1.STATUS_FAILED, ev.emitted[0].GetStatus())
	assert.Equal(t, "payment", ev.emitted[0].GetExtras().AsMap()["entity_type"])
}
