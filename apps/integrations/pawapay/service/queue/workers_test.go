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

package queue_test

import (
	"context"
	"errors"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/config"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/credentials"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/queue"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v3/events"
	frameQueue "github.com/pitabwire/frame/v3/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// fakeEvents captures emitted status updates.
type fakeEvents struct {
	frameEvents.Manager
	emitted []*commonv1.StatusUpdateRequest
	err     error
}

func (f *fakeEvents) Emit(_ context.Context, name string, payload any) error {
	if name != events.PaymentStatusUpdateEvent {
		return errors.New("unexpected event " + name)
	}
	req, ok := payload.(*commonv1.StatusUpdateRequest)
	if !ok {
		return errors.New("unexpected payload")
	}
	if f.err != nil {
		return f.err
	}
	f.emitted = append(f.emitted, req)
	return nil
}

// fakeClient serves the initiation answers and lookups the workers make.
type fakeClient struct {
	client.PawapayClient

	depositResp *client.DepositInitiationResponse
	depositErr  error
	depositGet  *client.DepositStatusResult
	depositGetE error
	depositReq  *client.DepositRequest

	payoutResp *client.PayoutInitiationResponse
	payoutErr  error
	payoutGet  *client.PayoutStatusResult
	payoutGetE error
	payoutReq  *client.PayoutRequest

	// heldBefore makes lookups find the record before any initiation call;
	// otherwise a record is only found once the worker has initiated it.
	heldBefore bool
	lookups    int
}

func (f *fakeClient) held(initiated bool) bool { return f.heldBefore || initiated }

func (f *fakeClient) InitiateDeposit(
	_ context.Context, _ *client.Credentials, req *client.DepositRequest,
) (*client.DepositInitiationResponse, error) {
	f.depositReq = req
	return f.depositResp, f.depositErr
}

func (f *fakeClient) GetDeposit(
	_ context.Context, _ *client.Credentials, _ string,
) (*client.DepositStatusResult, error) {
	f.lookups++
	if f.depositGetE == nil && !f.held(f.depositReq != nil) {
		return &client.DepositStatusResult{Status: notFound}, nil
	}
	return f.depositGet, f.depositGetE
}

func (f *fakeClient) InitiatePayout(
	_ context.Context, _ *client.Credentials, req *client.PayoutRequest,
) (*client.PayoutInitiationResponse, error) {
	f.payoutReq = req
	return f.payoutResp, f.payoutErr
}

func (f *fakeClient) GetPayout(
	_ context.Context, _ *client.Credentials, _ string,
) (*client.PayoutStatusResult, error) {
	f.lookups++
	if f.payoutGetE == nil && !f.held(f.payoutReq != nil) {
		return &client.PayoutStatusResult{Status: notFound}, nil
	}
	return f.payoutGet, f.payoutGetE
}

func testResolver(token string) *credentials.Resolver {
	return credentials.NewResolver(nil, &config.PawapayConfig{
		APIToken: token, Environment: "sandbox", Provider: "MTN_MOMO_ZMB",
	})
}

var (
	notFound     = client.SearchStatusNotFound
	errUncertain = collection.Uncertain(errors.New("deposit failed with status 502: bad gateway"))
	errRejected  = errors.New("deposit failed with status 400: invalid phone")
	rejection    = &client.FailureReason{FailureCode: "INVALID_PHONE_NUMBER", FailureMessage: "bad number"}
)

func promptPayload(t *testing.T, id string) []byte {
	t.Helper()
	raw, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:        id,
		Recipient: &commonv1.ContactLink{ContactId: "260763456789"},
		Amount:    &commonv1.Money{CurrencyCode: "ZMW", Units: 50},
	})
	require.NoError(t, err)
	return raw
}

func paymentPayload(t *testing.T, id string) []byte {
	t.Helper()
	raw, err := proto.Marshal(&paymentv1.Payment{
		Id:        id,
		Recipient: &commonv1.ContactLink{ContactId: "260763456789"},
		Amount:    &commonv1.Money{CurrencyCode: "ZMW", Units: 50},
		Outbound:  true,
	})
	require.NoError(t, err)
	return raw
}

func foundDeposit(status string, reason *client.FailureReason) *client.DepositStatusResult {
	return &client.DepositStatusResult{Status: client.SearchStatusFound, Data: &client.Deposit{
		DepositID: "dep-1", Status: status, Amount: "50", Currency: "ZMW", FailureReason: reason,
		Payer: client.NewMMOParty("260763456789", "MTN_MOMO_ZMB"),
	}}
}

func foundPayout(status string, reason *client.FailureReason) *client.PayoutStatusResult {
	return &client.PayoutStatusResult{Status: client.SearchStatusFound, Data: &client.Payout{
		PayoutID: "pay-1", Status: status, Amount: "50", Currency: "ZMW", FailureReason: reason,
		Recipient: client.NewMMOParty("260763456789", "MTN_MOMO_ZMB"),
	}}
}

// workerCase drives one worker through an initiation answer and a lookup
// answer, then checks what was emitted.
type workerCase struct {
	name        string
	initResp    string // initiation status; "" with initErr for a failed call
	initErr     error
	initReason  *client.FailureReason
	lookupFound string // record status pawaPay holds on lookup; "" = NOT_FOUND
	lookupErr   error
	heldBefore  bool
	emitErr     error
	token       string

	wantErr      bool
	wantNoEmit   bool
	wantNoInit   bool
	wantStatus   commonv1.STATUS
	wantState    commonv1.STATE
	wantFailCode string
}

func workerCases() []workerCase {
	return []workerCase{
		{
			name: "accepted is in process", initResp: client.InitiationStatusAccepted,
			wantStatus: commonv1.STATUS_IN_PROCESS, wantState: commonv1.STATE_ACTIVE,
		},
		{
			name: "rejected fails inactive with the reason", initResp: client.InitiationStatusRejected,
			initReason: rejection,
			wantStatus: commonv1.STATUS_FAILED, wantState: commonv1.STATE_INACTIVE, wantFailCode: "INVALID_PHONE_NUMBER",
		},
		{
			name: "plain rejection error fails", initErr: errRejected,
			wantStatus: commonv1.STATUS_FAILED, wantState: commonv1.STATE_INACTIVE,
		},
		{
			name: "uncertain outcome and nothing held redelivers", initErr: errUncertain,
			wantErr: true, wantNoEmit: true,
		},
		{
			name: "uncertain outcome and lookup error redelivers", initErr: errUncertain,
			lookupErr: errUncertain,
			wantErr:   true, wantNoEmit: true,
		},
		{
			name: "uncertain outcome reports what pawaPay holds", initErr: errUncertain,
			lookupFound: client.PaymentStatusCompleted,
			wantStatus:  commonv1.STATUS_SUCCESSFUL, wantState: commonv1.STATE_ACTIVE,
		},
		{
			name: "duplicate reports the held record", initResp: client.InitiationStatusDuplicateIgnored,
			lookupFound: client.PaymentStatusProcessing,
			wantStatus:  commonv1.STATUS_IN_PROCESS, wantState: commonv1.STATE_ACTIVE,
		},
		{
			name: "duplicate but not found redelivers", initResp: client.InitiationStatusDuplicateIgnored,
			wantErr: true, wantNoEmit: true,
		},
		{
			name: "record already held skips initiation", lookupFound: client.PaymentStatusFailed,
			heldBefore: true, initReason: rejection,
			wantNoInit: true, wantStatus: commonv1.STATUS_FAILED, wantState: commonv1.STATE_INACTIVE,
			wantFailCode: "INVALID_PHONE_NUMBER",
		},
		{
			name: "missing credentials fail", token: "-",
			wantNoInit: true, wantStatus: commonv1.STATUS_FAILED, wantState: commonv1.STATE_INACTIVE,
		},
		{
			name: "emit failure is returned", initResp: client.InitiationStatusAccepted, emitErr: errors.New("nats down"),
			wantErr: true, wantNoEmit: true,
		},
	}
}

func runWorkerCase(
	t *testing.T,
	tc workerCase,
	build func(*fakeEvents, *fakeClient, *credentials.Resolver) frameQueue.SubscribeWorker,
	payload []byte,
	wantEntity string,
	initiated func(*fakeClient) bool,
) {
	t.Helper()
	token := "t"
	if tc.token == "-" {
		token = ""
	}
	ev := &fakeEvents{err: tc.emitErr}
	fc := &fakeClient{
		depositErr: tc.initErr, payoutErr: tc.initErr,
		depositGetE: tc.lookupErr, payoutGetE: tc.lookupErr, heldBefore: tc.heldBefore,
	}
	if tc.initResp != "" {
		fc.depositResp = &client.DepositInitiationResponse{DepositID: "dep-1", Status: tc.initResp, FailureReason: tc.initReason}
		fc.payoutResp = &client.PayoutInitiationResponse{PayoutID: "pay-1", Status: tc.initResp, FailureReason: tc.initReason}
	}
	if tc.lookupFound != "" {
		fc.depositGet = foundDeposit(tc.lookupFound, tc.initReason)
		fc.payoutGet = foundPayout(tc.lookupFound, tc.initReason)
	} else {
		fc.depositGet = &client.DepositStatusResult{Status: notFound}
		fc.payoutGet = &client.PayoutStatusResult{Status: notFound}
	}

	err := build(ev, fc, testResolver(token)).Handle(t.Context(), map[string]string{"tenant_id": "t1"}, payload)

	if tc.wantErr {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
	assert.Equal(t, !tc.wantNoInit, initiated(fc), "initiation call")
	if tc.wantNoEmit {
		assert.Empty(t, ev.emitted)
		return
	}
	require.Len(t, ev.emitted, 1)
	st := ev.emitted[0]
	assert.Equal(t, tc.wantStatus, st.GetStatus())
	assert.Equal(t, tc.wantState, st.GetState())
	ex := st.GetExtras().AsMap()
	assert.Equal(t, wantEntity, ex["entity_type"])
	assert.Equal(t, "pawapay", ex["provider"])
	if tc.wantFailCode != "" {
		assert.Equal(t, tc.wantFailCode, ex["failure_code"])
	}
}

func TestPromptWorker(t *testing.T) {
	for _, tc := range workerCases() {
		t.Run(tc.name, func(t *testing.T) {
			runWorkerCase(t, tc,
				func(ev *fakeEvents, fc *fakeClient, r *credentials.Resolver) frameQueue.SubscribeWorker {
					return queue.NewPromptHandler(ev, fc, r)
				},
				promptPayload(t, "prompt-1"), "prompt",
				func(fc *fakeClient) bool { return fc.depositReq != nil })
		})
	}
}

func TestPaymentWorker(t *testing.T) {
	for _, tc := range workerCases() {
		t.Run(tc.name, func(t *testing.T) {
			runWorkerCase(t, tc,
				func(ev *fakeEvents, fc *fakeClient, r *credentials.Resolver) frameQueue.SubscribeWorker {
					return queue.NewPaymentHandler(ev, fc, r)
				},
				paymentPayload(t, "payment-1"), "payment",
				func(fc *fakeClient) bool { return fc.payoutReq != nil })
		})
	}
}

func TestPromptWorkerRequestCarriesAmountMetadata(t *testing.T) {
	ev := &fakeEvents{}
	fc := &fakeClient{
		depositGet:  &client.DepositStatusResult{Status: notFound},
		depositResp: &client.DepositInitiationResponse{DepositID: "dep-1", Status: client.InitiationStatusAccepted},
	}
	h := queue.NewPromptHandler(ev, fc, testResolver("t"))

	require.NoError(t, h.Handle(t.Context(), map[string]string{"tenant_id": "t1"}, promptPayload(t, "prompt-1")))

	require.NotNil(t, fc.depositReq)
	assert.Equal(t, "50", fc.depositReq.Amount)
	assert.Equal(t, "50", fc.depositReq.Metadata["amount"])
	assert.Equal(t, "ZMW", fc.depositReq.Metadata["currency"])
	assert.Equal(t, "prompt-1", fc.depositReq.Metadata["paymentId"])
	assert.Equal(t, "t1", fc.depositReq.Metadata["tenantId"])
	assert.Equal(t, fc.depositReq.DepositID, ev.emitted[0].GetExternalId())
}

func TestWorkerUnmarshalErrorIsNotRetried(t *testing.T) {
	ev := &fakeEvents{}
	fc := &fakeClient{}
	require.NoError(t, queue.NewPromptHandler(ev, fc, testResolver("t")).Handle(t.Context(), nil, []byte{0xff}))
	require.NoError(t, queue.NewPaymentHandler(ev, fc, testResolver("t")).Handle(t.Context(), nil, []byte{0xff}))
	assert.Empty(t, ev.emitted)
	assert.Zero(t, fc.lookups)
}
