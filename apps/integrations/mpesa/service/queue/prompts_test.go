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
	"net/http"
	"net/url"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"connectrpc.com/connect/v2"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSTKCallbackURL(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		promptID string
		want     map[string]string
	}{
		{
			name:     "tenant params and prompt id",
			headers:  map[string]string{"tenant_id": "t1", "partition_id": "p1"},
			promptID: "prompt-1",
			want:     map[string]string{"tenant_id": "t1", "partition_id": "p1", "prompt_id": "prompt-1"},
		},
		{
			name:     "prompt id without tenant",
			headers:  map[string]string{},
			promptID: "prompt-2",
			want:     map[string]string{"prompt_id": "prompt-2"},
		},
		{
			name:    "no prompt id keeps legacy url",
			headers: map[string]string{"tenant_id": "t1"},
			want:    map[string]string{"tenant_id": "t1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stkCallbackURL("https://hooks.example.com", tt.headers, tt.promptID)
			u, err := url.Parse(got)
			require.NoError(t, err)
			assert.Equal(t, "/webhook/mpesa/stk", u.Path)
			assert.Equal(t, "hooks.example.com", u.Host)
			q := u.Query()
			assert.Len(t, q, len(tt.want))
			for k, v := range tt.want {
				assert.Equal(t, v, q.Get(k), k)
			}
		})
	}
}

type recordingEvents struct {
	frameEvents.Manager
	emitted []*commonv1.StatusUpdateRequest
	err     error
}

func (r *recordingEvents) Emit(_ context.Context, _ string, payload any) error {
	if r.err != nil {
		return r.err
	}
	if req, ok := payload.(*commonv1.StatusUpdateRequest); ok {
		r.emitted = append(r.emitted, req)
	}
	return nil
}

type recordingMpesa struct {
	client.MpesaClient
	pushed  []*client.STKPushRequest
	b2c     []*client.B2CRequest
	pushErr error
	b2cErr  error
}

func (m *recordingMpesa) STKPush(
	_ context.Context,
	_ *client.MpesaCredentials,
	req *client.STKPushRequest,
) (*client.STKPushResponse, error) {
	m.pushed = append(m.pushed, req)
	if m.pushErr != nil {
		return nil, m.pushErr
	}
	return &client.STKPushResponse{MerchantRequestID: "mr-1", CheckoutRequestID: "ws_CO_1", ResponseCode: "0"}, nil
}

func (m *recordingMpesa) B2CPayment(
	_ context.Context,
	_ *client.MpesaCredentials,
	req *client.B2CRequest,
) (*client.B2CResponse, error) {
	m.b2c = append(m.b2c, req)
	if m.b2cErr != nil {
		return nil, m.b2cErr
	}
	return &client.B2CResponse{ConversationID: "AG_1", OriginatorConversationID: req.OriginatorConversationID, ResponseCode: "0"}, nil
}

func testConfig() *config.MpesaConfig {
	return &config.MpesaConfig{
		ConsumerKey: "k", ConsumerSecret: "s", Shortcode: "174379", Passkey: "pk",
		CallbackURL: "https://hooks.example.com", InitiatorName: "init", SecurityCredential: "sec",
	}
}

func promptPayload(t *testing.T, amount *commonv1.Money) []byte {
	t.Helper()
	payload, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:        "prompt-1",
		Recipient: &commonv1.ContactLink{ContactId: "254700000001"},
		Amount:    amount,
	})
	require.NoError(t, err)
	return payload
}

func TestPromptHandler_STKPush(t *testing.T) {
	tests := []struct {
		name       string
		amount     *commonv1.Money
		pushErr    error
		wantPushed bool
		wantErr    bool
		wantNoEmit bool
		wantStatus commonv1.STATUS
	}{
		{name: "KES is pushed", amount: &commonv1.Money{CurrencyCode: "KES", Units: 100}, wantPushed: true, wantStatus: commonv1.STATUS_IN_PROCESS},
		{name: "unset currency is pushed", amount: &commonv1.Money{Units: 100}, wantPushed: true, wantStatus: commonv1.STATUS_IN_PROCESS},
		{name: "non-KES currency is refused", amount: &commonv1.Money{CurrencyCode: "USD", Units: 100}, wantStatus: commonv1.STATUS_FAILED},
		{name: "fractional shillings are refused, never rounded", amount: &commonv1.Money{CurrencyCode: "KES", Units: 100, Nanos: 500_000_000}, wantStatus: commonv1.STATUS_FAILED},
		{name: "zero amount is refused", amount: &commonv1.Money{CurrencyCode: "KES"}, wantStatus: commonv1.STATUS_FAILED},
		{
			name:       "Daraja rejection fails the prompt",
			amount:     &commonv1.Money{CurrencyCode: "KES", Units: 100},
			pushErr:    &client.APIError{Operation: "stk push", StatusCode: http.StatusBadRequest, Body: "Bad Request - Invalid PhoneNumber"},
			wantPushed: true,
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name:       "unknown outcome is redelivered, nothing recorded",
			amount:     &commonv1.Money{CurrencyCode: "KES", Units: 100},
			pushErr:    collection.Uncertain(errors.New("timeout")),
			wantPushed: true,
			wantErr:    true,
			wantNoEmit: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evts := &recordingEvents{}
			mp := &recordingMpesa{pushErr: tt.pushErr}
			h := NewPromptHandler(evts, mp, nil, testConfig(), nil)

			err := h.Handle(t.Context(), map[string]string{"tenant_id": "t1"}, promptPayload(t, tt.amount))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if !tt.wantPushed {
				assert.Empty(t, mp.pushed)
			} else {
				require.Len(t, mp.pushed, 1)
				u, perr := url.Parse(mp.pushed[0].CallBackURL)
				require.NoError(t, perr)
				assert.Equal(t, "prompt-1", u.Query().Get(client.CallbackParamPromptID))
				assert.Equal(t, "t1", u.Query().Get("tenant_id"))
				assert.Equal(t, "100", mp.pushed[0].Amount)
			}
			if tt.wantNoEmit {
				assert.Empty(t, evts.emitted)
				return
			}
			require.Len(t, evts.emitted, 1)
			st := evts.emitted[0]
			assert.Equal(t, "prompt-1", st.GetId())
			assert.Equal(t, tt.wantStatus, st.GetStatus())
			assert.Equal(t, events.StateFor(tt.wantStatus), st.GetState())
			extras := st.GetExtras().GetFields()
			assert.Equal(t, "prompt", extras["entity_type"].GetStringValue())
			if tt.wantStatus == commonv1.STATUS_FAILED {
				assert.NotEmpty(t, extras[events.ExtraError].GetStringValue())
				return
			}
			assert.Equal(t, "ws_CO_1", extras[client.ExtraCheckoutRequestID].GetStringValue())
			assert.Equal(t, "100", extras[client.ExtraRequestedAmount].GetStringValue())
			assert.Equal(t, "KES", extras[client.ExtraRequestedCurrency].GetStringValue())
		})
	}
}

func TestPromptHandler_EmitErrorIsReturned(t *testing.T) {
	evts := &recordingEvents{err: errors.New("bus down")}
	h := NewPromptHandler(evts, &recordingMpesa{}, nil, testConfig(), nil)

	require.Error(t, h.Handle(t.Context(), map[string]string{}, promptPayload(t, &commonv1.Money{CurrencyCode: "KES", Units: 100})))
}

type stubStatusReader struct {
	extras map[string]any
	err    error
}

func (s *stubStatusReader) Status(
	_ context.Context,
	req *commonv1.StatusRequest,
) (*commonv1.StatusResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	extras, err := structpb.NewStruct(s.extras)
	if err != nil {
		return nil, err
	}
	return &commonv1.StatusResponse{Id: req.GetId(), Extras: extras}, nil
}

func TestPromptHandler_IdempotentPerPrompt(t *testing.T) {
	tests := []struct {
		name       string
		reader     *stubStatusReader
		wantPushed bool
	}{
		{
			name:   "already pushed: redelivery is skipped",
			reader: &stubStatusReader{extras: map[string]any{"entity_type": "prompt", client.ExtraCheckoutRequestID: "ws_CO_first"}},
		},
		{
			name:       "only queued: push",
			reader:     &stubStatusReader{extras: map[string]any{"transaction_ref": "tx-1"}},
			wantPushed: true,
		},
		{
			name:       "no status yet: push",
			reader:     &stubStatusReader{err: connect.NewError(connect.CodeNotFound, "record not found")},
			wantPushed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evts := &recordingEvents{}
			mp := &recordingMpesa{}
			h := NewPromptHandler(evts, mp, nil, testConfig(), tt.reader)

			require.NoError(t, h.Handle(t.Context(), map[string]string{}, promptPayload(t, &commonv1.Money{CurrencyCode: "KES", Units: 100})))

			if !tt.wantPushed {
				assert.Empty(t, mp.pushed)
				assert.Empty(t, evts.emitted, "the existing binding must not be overwritten")
				return
			}
			assert.Len(t, mp.pushed, 1)
			require.Len(t, evts.emitted, 1)
			assert.Equal(t, commonv1.STATUS_IN_PROCESS, evts.emitted[0].GetStatus())
		})
	}
}

func TestCredentials_SettingsConnectionWithoutClient(t *testing.T) {
	r := credentialResolver{cfg: testConfig()}
	_, err := r.extractCredentials(t.Context(), map[string]string{config.HeaderConnectionCredentials: "conn-1"})
	require.Error(t, err)
}
