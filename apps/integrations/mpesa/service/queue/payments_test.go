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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func paymentPayload(t *testing.T, amount *commonv1.Money) []byte {
	t.Helper()
	payload, err := proto.Marshal(&paymentv1.Payment{
		Id:        "payment-1",
		Recipient: &commonv1.ContactLink{ContactId: "254700000001"},
		Amount:    amount,
	})
	require.NoError(t, err)
	return payload
}

func TestPaymentHandler_B2C(t *testing.T) {
	notFound := &stubStatusReader{err: connect.NewError(connect.CodeNotFound, "record not found")}
	tests := []struct {
		name         string
		cfg          *config.MpesaConfig
		amount       *commonv1.Money
		reader       *stubStatusReader
		b2cErr       error
		wantSent     bool
		wantErr      bool
		wantStatuses []commonv1.STATUS
	}{
		{
			name:         "KES payout: binding recorded before sending, then accepted",
			cfg:          testConfig(),
			amount:       &commonv1.Money{CurrencyCode: "KES", Units: 250},
			reader:       notFound,
			wantSent:     true,
			wantStatuses: []commonv1.STATUS{commonv1.STATUS_IN_PROCESS, commonv1.STATUS_IN_PROCESS},
		},
		{
			name:         "non-KES currency refused",
			cfg:          testConfig(),
			amount:       &commonv1.Money{CurrencyCode: "UGX", Units: 250},
			reader:       notFound,
			wantStatuses: []commonv1.STATUS{commonv1.STATUS_FAILED},
		},
		{
			name:         "fractional shillings refused, never rounded",
			cfg:          testConfig(),
			amount:       &commonv1.Money{CurrencyCode: "KES", Units: 250, Nanos: 500_000_000},
			reader:       notFound,
			wantStatuses: []commonv1.STATUS{commonv1.STATUS_FAILED},
		},
		{
			name: "missing initiator credentials refused",
			cfg: &config.MpesaConfig{ConsumerKey: "k", ConsumerSecret: "s", Shortcode: "174379",
				CallbackURL: "https://hooks.example.com"},
			amount:       &commonv1.Money{CurrencyCode: "KES", Units: 250},
			reader:       notFound,
			wantStatuses: []commonv1.STATUS{commonv1.STATUS_FAILED},
		},
		{
			name:         "Daraja rejection: FAILED after the binding",
			cfg:          testConfig(),
			amount:       &commonv1.Money{CurrencyCode: "KES", Units: 250},
			reader:       notFound,
			b2cErr:       &client.APIError{Operation: "b2c", StatusCode: http.StatusBadRequest, Body: "invalid initiator"},
			wantSent:     true,
			wantStatuses: []commonv1.STATUS{commonv1.STATUS_IN_PROCESS, commonv1.STATUS_FAILED},
		},
		{
			name:         "Daraja non-zero response code: FAILED",
			cfg:          testConfig(),
			amount:       &commonv1.Money{CurrencyCode: "KES", Units: 250},
			reader:       notFound,
			b2cErr:       &client.APIError{Operation: "b2c", StatusCode: http.StatusOK, ResponseCode: "1", Body: "not accepted"},
			wantSent:     true,
			wantStatuses: []commonv1.STATUS{commonv1.STATUS_IN_PROCESS, commonv1.STATUS_FAILED},
		},
		{
			name:         "unknown outcome: binding kept, message redelivered",
			cfg:          testConfig(),
			amount:       &commonv1.Money{CurrencyCode: "KES", Units: 250},
			reader:       notFound,
			b2cErr:       collection.Uncertain(errors.New("timeout")),
			wantSent:     true,
			wantErr:      true,
			wantStatuses: []commonv1.STATUS{commonv1.STATUS_IN_PROCESS},
		},
		{
			name:   "already issued: redelivery never pays again",
			cfg:    testConfig(),
			amount: &commonv1.Money{CurrencyCode: "KES", Units: 250},
			reader: &stubStatusReader{extras: map[string]any{
				"entity_type": "payment", client.ExtraOriginatorConversationID: "0123456789abcdef0123456789abcdef",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evts := &recordingEvents{}
			mp := &recordingMpesa{b2cErr: tt.b2cErr}
			h := NewPaymentHandler(evts, mp, nil, tt.cfg, tt.reader)

			err := h.Handle(t.Context(), map[string]string{"tenant_id": "t1"}, paymentPayload(t, tt.amount))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			var got []commonv1.STATUS
			for _, st := range evts.emitted {
				got = append(got, st.GetStatus())
				assert.Equal(t, "payment-1", st.GetId())
				assert.Equal(t, events.StateFor(st.GetStatus()), st.GetState())
				assert.Equal(t, "payment", st.GetExtras().GetFields()["entity_type"].GetStringValue())
			}
			assert.Equal(t, tt.wantStatuses, got)
			if !tt.wantSent {
				assert.Empty(t, mp.b2c)
				return
			}
			require.Len(t, mp.b2c, 1)
			req := mp.b2c[0]
			assert.NotEqual(t, "payment-1", req.OriginatorConversationID, "originator id is never the payment id")
			assert.Len(t, req.OriginatorConversationID, 32)
			assert.Equal(t, "250", req.Amount)
			assert.Equal(t, "init", req.InitiatorName)
			secret := ""
			for _, raw := range []string{req.ResultURL, req.QueueTimeOutURL} {
				u, perr := url.Parse(raw)
				require.NoError(t, perr)
				assert.Equal(t, "payment-1", u.Query().Get(client.CallbackParamPaymentID))
				assert.Equal(t, "t1", u.Query().Get("tenant_id"))
				secret = u.Query().Get(client.CallbackParamSecret)
				assert.Len(t, secret, 32)
			}
			// The first status, written before the call, carries the binding
			// the result callback checks: the random originator id and the
			// digest (never the clear value) of the callback secret.
			first := evts.emitted[0].GetExtras().GetFields()
			assert.Equal(t, req.OriginatorConversationID, first[client.ExtraOriginatorConversationID].GetStringValue())
			assert.Equal(t, client.HashCallbackSecret(secret), first[client.ExtraCallbackSecretHash].GetStringValue())
			assert.Equal(t, "250", first[client.ExtraRequestedAmount].GetStringValue())
			assert.Equal(t, "KES", first[client.ExtraRequestedCurrency].GetStringValue())
			for _, st := range evts.emitted {
				assertNoSecretInClear(t, st.GetExtras().GetFields(), secret)
			}
		})
	}
}

func TestB2CCallbackURL(t *testing.T) {
	got := b2cCallbackURL("https://hooks.example.com", "/webhook/mpesa/b2c", map[string]string{"partition_id": "p1"}, "payment-9", "s3cret")
	u, err := url.Parse(got)
	require.NoError(t, err)
	assert.Equal(t, "/webhook/mpesa/b2c", u.Path)
	assert.Equal(t, "payment-9", u.Query().Get(client.CallbackParamPaymentID))
	assert.Equal(t, "s3cret", u.Query().Get(client.CallbackParamSecret))
	assert.Equal(t, "p1", u.Query().Get("partition_id"))
}
