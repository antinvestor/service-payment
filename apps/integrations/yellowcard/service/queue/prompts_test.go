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
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/config"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/queue"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// fakeEvents captures emitted status updates.
type fakeEvents struct {
	frameEvents.Manager
	emitted []*commonv1.StatusUpdateRequest
}

func (f *fakeEvents) Emit(_ context.Context, name string, payload any) error {
	if name != events.PaymentStatusUpdateEvent {
		return errors.New("unexpected event " + name)
	}
	req, ok := payload.(*commonv1.StatusUpdateRequest)
	if !ok {
		return errors.New("unexpected payload")
	}
	f.emitted = append(f.emitted, req)
	return nil
}

func (f *fakeEvents) last(t *testing.T) *commonv1.StatusUpdateRequest {
	t.Helper()
	require.NotEmpty(t, f.emitted, "expected a status update to be emitted")
	return f.emitted[len(f.emitted)-1]
}

// fakeClient serves catalog data and records submissions.
type fakeClient struct {
	client.YellowcardClient

	channels []client.Channel
	networks []client.Network

	receiveReq  *client.ReceiveRequest
	receiveResp *client.Receive
	receiveErr  error
	lookupResp  *client.Receive
	lookupErr   error

	sendReq  *client.SendRequest
	sendResp *client.Send
	sendErr  error
	sendLook *client.Send

	gotCreds *client.Credentials
}

func (f *fakeClient) GetChannels(_ context.Context, creds *client.Credentials, _ string) ([]client.Channel, error) {
	f.gotCreds = creds
	return f.channels, nil
}

func (f *fakeClient) GetNetworks(_ context.Context, _ *client.Credentials, _ string) ([]client.Network, error) {
	return f.networks, nil
}

func (f *fakeClient) SubmitReceive(_ context.Context, creds *client.Credentials, req *client.ReceiveRequest) (*client.Receive, error) {
	f.gotCreds, f.receiveReq = creds, req
	return f.receiveResp, f.receiveErr
}

func (f *fakeClient) GetReceiveBySequenceID(_ context.Context, _ *client.Credentials, _ string) (*client.Receive, error) {
	return f.lookupResp, f.lookupErr
}

func (f *fakeClient) SubmitSend(_ context.Context, creds *client.Credentials, req *client.SendRequest) (*client.Send, error) {
	f.gotCreds, f.sendReq = creds, req
	return f.sendResp, f.sendErr
}

func (f *fakeClient) GetSendBySequenceID(_ context.Context, _ *client.Credentials, _ string) (*client.Send, error) {
	return f.sendLook, nil
}

func testConfig() *config.YellowcardConfig {
	return &config.YellowcardConfig{
		APIKey:              "key",
		SecretKey:           "secret",
		Environment:         "sandbox",
		CustomerType:        "retail",
		DefaultRedirectURL:  "https://pay.example.com/return",
		CatalogCacheSeconds: 0,
	}
}

func ugandaCatalog() ([]client.Channel, []client.Network) {
	channels := []client.Channel{
		{ID: "c-momo-in", Country: "UG", Currency: "UGX", Status: "active", APIStatus: "active", ChannelType: "momo", RampType: "deposit"},
		{ID: "c-bank-in", Country: "UG", Currency: "UGX", Status: "active", APIStatus: "active", ChannelType: "bank", RampType: "deposit"},
		{ID: "c-momo-out", Country: "UG", Currency: "UGX", Status: "active", APIStatus: "active", ChannelType: "momo", RampType: "withdraw"},
		{ID: "c-bank-out", Country: "UG", Currency: "UGX", Status: "active", APIStatus: "active", ChannelType: "bank", RampType: "withdraw"},
	}
	networks := []client.Network{
		{ID: "n-mtn", Code: "MTN", Name: "MTN Uganda", Country: "UG", Status: "active", AccountNumberType: "momo", ChannelIDs: []string{"c-momo-in", "c-momo-out"}},
		{ID: "n-airtel", Code: "AIRTEL", Name: "Airtel Uganda", Country: "UG", Status: "active", AccountNumberType: "momo", ChannelIDs: []string{"c-momo-in", "c-momo-out"}},
		{ID: "n-stanbic", Code: "031", Name: "Stanbic Bank", Country: "UG", Status: "active", AccountNumberType: "bank", ChannelIDs: []string{"c-bank-in", "c-bank-out"}},
	}
	return channels, networks
}

func promptPayload(t *testing.T, id, phone string, units int64, currency string, extra map[string]any) []byte {
	t.Helper()
	var st *structpb.Struct
	if extra != nil {
		var err error
		st, err = structpb.NewStruct(extra)
		require.NoError(t, err)
	}
	raw, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:        id,
		Route:     "yellowcard",
		Recipient: &commonv1.ContactLink{ContactId: phone, ProfileId: "profile-1"},
		Source:    &commonv1.ContactLink{ContactId: phone, ProfileId: "profile-1"},
		Amount:    &commonv1.Money{CurrencyCode: currency, Units: units},
		Extra:     st,
	})
	require.NoError(t, err)
	return raw
}

func extrasMap(req *commonv1.StatusUpdateRequest) map[string]any {
	return req.GetExtras().AsMap()
}

func TestPrompt_MomoHappyPath(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{
		channels: channels, networks: networks,
		receiveResp: &client.Receive{
			ID: "rcv-1", SequenceID: "prompt-1", Status: "processing", ChannelID: "c-momo-in",
			Country: "UG", Currency: "UGX", ConvertedAmount: 5000, Rate: 3800,
			Source:    client.Source{AccountType: "momo", AccountNumber: "+256701234567", NetworkID: "n-airtel"},
			ExpiresAt: "2030-01-01T00:00:00Z",
		},
	}
	ev := &fakeEvents{}
	cfg := testConfig()
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, cfg)

	headers := map[string]string{"tenant_id": "t1", "partition_id": "p1"}
	err := h.Handle(t.Context(), headers, promptPayload(t, "prompt-1", "+256 701 234567", 5000, "UGX", map[string]any{
		"customer_name": "Jane Doe", "customer_email": "jane@example.com", "network": "airtel",
		"success_url": "https://pay.example.com/c/abc",
	}))
	require.NoError(t, err)

	require.NotNil(t, fc.receiveReq)
	req := fc.receiveReq
	assert.Equal(t, "prompt-1", req.SequenceID)
	assert.Equal(t, int64(5000), req.LocalAmount)
	assert.Equal(t, "UG", req.Country)
	assert.Equal(t, "UGX", req.Currency)
	assert.Equal(t, "c-momo-in", req.ChannelID)
	assert.Equal(t, "momo", req.Source.AccountType)
	assert.Equal(t, "+256701234567", req.Source.AccountNumber)
	assert.Equal(t, "n-airtel", req.Source.NetworkID)
	assert.Equal(t, "Jane Doe", req.Recipient.Name)
	assert.Equal(t, "jane@example.com", req.Recipient.Email)
	assert.Equal(t, "UG", req.Recipient.Country)
	assert.Equal(t, "retail", req.CustomerType)
	assert.Equal(t, "profile-1", req.CustomerUID)
	assert.Equal(t, "https://pay.example.com/c/abc", req.RedirectURL)
	assert.Equal(t, "other", req.Reason)
	assert.True(t, req.ForceAccept)
	assert.Equal(t, "key", fc.gotCreds.APIKey)

	st := ev.last(t)
	assert.Equal(t, "prompt-1", st.GetId())
	assert.Equal(t, "rcv-1", st.GetExternalId())
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, st.GetState())
	ex := extrasMap(st)
	assert.Equal(t, "prompt", ex["entity_type"])
	assert.Equal(t, "yellowcard", ex["provider"])
	assert.Equal(t, "rcv-1", ex["receive_id"])
	assert.Equal(t, "n-airtel", ex["network_id"])
	assert.Equal(t, "Airtel Uganda", ex["network"])
	assert.Contains(t, ex["payment_instruction"], "UGX 5,000")
	_, hasNext := ex["next_action"]
	assert.False(t, hasNext)
}

func TestPrompt_BankHappyPath(t *testing.T) {
	channels := []client.Channel{
		{ID: "c-bank-in", Country: "NG", Currency: "NGN", Status: "active", ChannelType: "bank", RampType: "deposit"},
	}
	fc := &fakeClient{
		channels: channels,
		receiveResp: &client.Receive{
			ID: "rcv-2", SequenceID: "prompt-2", Status: "processing", ChannelID: "c-bank-in",
			Country: "NG", Currency: "NGN", ConvertedAmount: 9757.8, Reference: "JJ8094861",
			BankInfo:  &client.BankInfo{Name: "PAGA", AccountNumber: "01234567890", AccountName: "Ken Adams"},
			ExpiresAt: "2030-01-01T04:00:00Z",
		},
	}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	err := h.Handle(t.Context(), nil, promptPayload(t, "prompt-2", "+2348012345678", 9758, "NGN", map[string]any{
		"customer_name": "Ken Adams", "customer_email": "ken@example.com",
	}))
	require.NoError(t, err)

	require.NotNil(t, fc.receiveReq)
	assert.Equal(t, "bank", fc.receiveReq.Source.AccountType)
	assert.Empty(t, fc.receiveReq.Source.NetworkID)
	assert.Equal(t, "https://pay.example.com/return", fc.receiveReq.RedirectURL, "config default redirect")

	ex := extrasMap(ev.last(t))
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, ev.last(t).GetStatus())
	assert.Equal(t, "PAGA", ex["bank_name"])
	assert.Equal(t, "01234567890", ex["bank_account_number"])
	assert.Equal(t, "Ken Adams", ex["bank_account_name"])
	assert.Equal(t, "JJ8094861", ex["payment_reference"])
	assert.Equal(t, "2030-01-01T04:00:00Z", ex["payment_expires_at"])
	assert.Contains(t, ex["payment_instruction"], "01234567890")
}

func TestPrompt_ExplicitChannelTypeOverridesAuto(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{channels: channels, networks: networks,
		receiveResp: &client.Receive{ID: "r", SequenceID: "p", Status: "pending", Country: "UG", Currency: "UGX"}}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p", "+256701234567", 100, "UGX",
		map[string]any{"payment_method_type": "bank_transfer"})))
	assert.Equal(t, "bank", fc.receiveReq.Source.AccountType)
	assert.Equal(t, "c-bank-in", fc.receiveReq.ChannelID)
}

func TestPrompt_ProviderError(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{channels: channels, networks: networks,
		receiveErr: &client.APIError{HTTPStatus: 400, Code: "PaymentValidationError", Message: "amount must be between 0 and 1000"}}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p", "+256701234567", 100, "UGX", nil)))
	st := ev.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, st.GetState())
	ex := extrasMap(st)
	assert.Equal(t, "PaymentValidationError", ex["failure_code"])
	assert.Equal(t, "amount must be between 0 and 1000", ex["failure_message"])
	assert.Equal(t, "prompt", ex["entity_type"])
}

func TestPrompt_RejectedByProvider(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{channels: channels, networks: networks,
		receiveResp: &client.Receive{ID: "r", SequenceID: "p", Status: "failed", ErrorCode: "INVALID_RECIPIENT", Country: "UG", Currency: "UGX"}}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p", "+256701234567", 100, "UGX", nil)))
	st := ev.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "r", st.GetExternalId())
	assert.Equal(t, "INVALID_RECIPIENT", extrasMap(st)["failure_code"])
}

func TestPrompt_MissingCredentials(t *testing.T) {
	cfg := testConfig()
	cfg.SecretKey = ""
	fc := &fakeClient{}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, cfg)

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p", "+256701234567", 100, "UGX", nil)))
	assert.Nil(t, fc.receiveReq)
	assert.Equal(t, commonv1.STATUS_FAILED, ev.last(t).GetStatus())
	assert.Equal(t, "CREDENTIALS", extrasMap(ev.last(t))["failure_code"])
}

func TestPrompt_NoChannel(t *testing.T) {
	fc := &fakeClient{channels: []client.Channel{
		{ID: "x", Country: "UG", Currency: "UGX", Status: "inactive", ChannelType: "momo", RampType: "deposit"},
	}}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p", "+256701234567", 100, "UGX", nil)))
	assert.Nil(t, fc.receiveReq)
	assert.Equal(t, "CHANNEL_UNAVAILABLE", extrasMap(ev.last(t))["failure_code"])
}

func TestPrompt_UnknownCountry(t *testing.T) {
	fc := &fakeClient{}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p", "+11234567890", 100, "USD", nil)))
	assert.Equal(t, "INVALID_COUNTRY", extrasMap(ev.last(t))["failure_code"])
}

func TestPrompt_DuplicateSequence(t *testing.T) {
	channels, networks := ugandaCatalog()
	fc := &fakeClient{channels: channels, networks: networks,
		receiveErr: &client.APIError{HTTPStatus: 400, Code: "PaymentValidationError", Message: "sequenceId already exists"},
		lookupResp: &client.Receive{ID: "r-existing", SequenceID: "p", Status: "pending", Country: "UG", Currency: "UGX", ConvertedAmount: 100}}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p", "+256701234567", 100, "UGX", nil)))
	st := ev.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, "r-existing", st.GetExternalId())
}

func TestPrompt_BadPayload(t *testing.T) {
	fc := &fakeClient{}
	ev := &fakeEvents{}
	h := queue.NewPromptHandler(ev, fc, client.NewCatalog(fc, 0), nil, testConfig())
	require.NoError(t, h.Handle(t.Context(), nil, []byte{0xff, 0x01, 0x02}))
	assert.Empty(t, ev.emitted)
}
