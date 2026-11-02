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
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func paymentPayload(t *testing.T, id, phone string, units int64, currency string, extra map[string]any) []byte {
	t.Helper()
	var st *structpb.Struct
	if extra != nil {
		var err error
		st, err = structpb.NewStruct(extra)
		require.NoError(t, err)
	}
	raw, err := proto.Marshal(&paymentv1.Payment{
		Id:        id,
		Route:     "yellowcard",
		Source:    &commonv1.ContactLink{ProfileId: "merchant-1"},
		Recipient: &commonv1.ContactLink{ContactId: phone},
		Amount:    &commonv1.Money{CurrencyCode: currency, Units: units},
		Outbound:  true,
		Extra:     st,
	})
	require.NoError(t, err)
	return raw
}

func TestPayment_MomoHappyPath(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{channels: channels, networks: networks,
		sendResp: &client.Send{ID: "snd-1", SequenceID: "pay-1", Status: "processing", ChannelID: "c-momo-out",
			Country: "UG", Currency: "UGX", ConvertedAmount: 20000, Rate: 3800,
			Destination: client.Destination{AccountType: "momo", AccountNumber: "+256701234567", NetworkID: "n-mtn"}}}
	ev := &fakeEvents{}
	cfg := testConfig()
	cfg.CustomerType = "institution"
	cfg.BusinessID = "biz-1"
	cfg.BusinessName = "Acme Ltd"
	h := queue.NewPaymentHandler(ev, fc, client.NewCatalog(fc, 0), nil, cfg)

	require.NoError(t, h.Handle(t.Context(), map[string]string{"tenant_id": "t1"},
		paymentPayload(t, "pay-1", "+256701234567", 20000, "UGX", map[string]any{"network": "MTN", "reason": "bills", "recipient_name": "Jane"})))

	require.NotNil(t, fc.sendReq)
	req := fc.sendReq
	assert.Equal(t, "pay-1", req.SequenceID)
	assert.Equal(t, int64(20000), req.LocalAmount)
	assert.Equal(t, "c-momo-out", req.ChannelID)
	assert.Equal(t, "momo", req.Destination.AccountType)
	assert.Equal(t, "+256701234567", req.Destination.AccountNumber)
	assert.Equal(t, "n-mtn", req.Destination.NetworkID)
	assert.Equal(t, "Jane", req.Destination.AccountName)
	assert.Equal(t, "bills", req.Reason)
	assert.Equal(t, "institution", req.CustomerType)
	assert.Equal(t, "biz-1", req.Sender.BusinessID)
	assert.Equal(t, "Acme Ltd", req.Sender.BusinessName)
	assert.True(t, req.ForceAccept)

	st := ev.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, "snd-1", st.GetExternalId())
	ex := extrasMap(st)
	assert.Equal(t, "payment", ex["entity_type"])
	assert.Equal(t, "MTN Uganda", ex["network"])
}

func TestPayment_BankHappyPath(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{channels: channels, networks: networks,
		sendResp: &client.Send{ID: "snd-2", SequenceID: "pay-2", Status: "created", Country: "UG", Currency: "UGX"}}
	ev := &fakeEvents{}
	h := queue.NewPaymentHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil,
		paymentPayload(t, "pay-2", "", 50000, "UGX", map[string]any{
			"account_number": "9030001234567", "account_name": "Ken Adams", "country": "UG", "bank_code": "031",
		})))

	require.NotNil(t, fc.sendReq)
	req := fc.sendReq
	assert.Equal(t, "bank", req.Destination.AccountType)
	assert.Equal(t, "9030001234567", req.Destination.AccountNumber)
	assert.Equal(t, "n-stanbic", req.Destination.NetworkID)
	assert.Equal(t, "Stanbic Bank", req.Destination.AccountBank)
	assert.Equal(t, "Ken Adams", req.Destination.AccountName)
	assert.Equal(t, "c-bank-out", req.ChannelID)
	assert.Equal(t, "other", req.Reason)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, ev.last(t).GetStatus())
}

func TestPayment_ProviderError(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{channels: channels, networks: networks,
		sendErr: &client.APIError{HTTPStatus: 400, Code: "PaymentValidationError", Message: "insufficient balance"}}
	ev := &fakeEvents{}
	h := queue.NewPaymentHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, paymentPayload(t, "pay-3", "+256701234567", 100, "UGX", nil)))
	st := ev.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, st.GetState())
	assert.Equal(t, "payment", extrasMap(st)["entity_type"])
	assert.Equal(t, "PaymentValidationError", extrasMap(st)["failure_code"])
}

func TestPayment_MissingCredentials(t *testing.T) {
	cfg := testConfig()
	cfg.APIKey = ""
	fc := &fakeClient{}
	ev := &fakeEvents{}
	h := queue.NewPaymentHandler(ev, fc, client.NewCatalog(fc, 0), nil, cfg)

	require.NoError(t, h.Handle(t.Context(), nil, paymentPayload(t, "pay-4", "+256701234567", 100, "UGX", nil)))
	assert.Nil(t, fc.sendReq)
	assert.Equal(t, "CREDENTIALS", extrasMap(ev.last(t))["failure_code"])
}

func TestPayment_BadPayload(t *testing.T) {
	fc := &fakeClient{}
	ev := &fakeEvents{}
	h := queue.NewPaymentHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())
	require.NoError(t, h.Handle(t.Context(), nil, []byte{0xff}))
	assert.Empty(t, ev.emitted)
}
