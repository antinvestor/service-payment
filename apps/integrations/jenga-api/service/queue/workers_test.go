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
	"net/url"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/queue"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

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

type fakeJenga struct {
	coreapi.JengaApiClient
	tokenErr error
	stkReq   *models.STKUSSDRequest
	stkResp  *models.STKUSSDResponse
	stkErr   error
	creds    *coreapi.Credentials
}

func (f *fakeJenga) GenerateBearerToken(_ context.Context, creds *coreapi.Credentials) (*coreapi.BearerTokenResponse, error) {
	f.creds = creds
	if f.tokenErr != nil {
		return nil, f.tokenErr
	}
	return &coreapi.BearerTokenResponse{AccessToken: "tok"}, nil
}

func (f *fakeJenga) InitiateSTKUSSD(_ context.Context, _ *coreapi.Credentials, req models.STKUSSDRequest, _ string) (*models.STKUSSDResponse, error) {
	f.stkReq = &req
	return f.stkResp, f.stkErr
}

type fakeStatus struct {
	paymentv1connect.PaymentServiceClient
	extras map[string]any
}

func (f *fakeStatus) Status(_ context.Context, req *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	if f.extras == nil {
		return nil, errors.New("not found")
	}
	st, _ := structpb.NewStruct(f.extras)
	return &commonv1.StatusResponse{Id: req.GetId(), Extras: st}, nil
}

// fakeSettings serves one tenant connection's Jenga credentials.
type fakeSettings struct {
	settingsv1connect.SettingsServiceClient
}

func (fakeSettings) Get(_ context.Context, _ *settingsv1.GetRequest) (*settingsv1.GetResponse, error) {
	return &settingsv1.GetResponse{Data: &settingsv1.SettingObject{Value: `{"` + config.HeaderMerchantCode + `":"MC-T",` +
		`"` + config.HeaderConsumerSecret + `":"cs-t","` + config.HeaderAPIKey + `":"ak-t",` +
		`"` + config.HeaderPrivateKeyPEM + `":"tenant-pem"}`}}, nil
}

func testConfig() *config.JengaConfig {
	return &config.JengaConfig{
		MerchantCode: "MC1", ConsumerSecret: "cs", ApiKey: "ak",
		JengaCallbackURL: "https://hooks.example.com", Environment: "sandbox", JengaPrivateKeyPEM: "pem",
	}
}

func promptPayload(t *testing.T, id, phone string, units int64, nanos int32, extra map[string]any) []byte {
	t.Helper()
	var st *structpb.Struct
	if extra != nil {
		var err error
		st, err = structpb.NewStruct(extra)
		require.NoError(t, err)
	}
	raw, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:               id,
		Source:           &commonv1.ContactLink{ContactId: phone},
		Amount:           &commonv1.Money{CurrencyCode: "KES", Units: units, Nanos: nanos},
		RecipientAccount: &paymentv1.Account{AccountNumber: "0011", CountryCode: "KE", Name: "Shop"},
		Extra:            st,
	})
	require.NoError(t, err)
	return raw
}

func extrasMap(req *commonv1.StatusUpdateRequest) map[string]any {
	return req.GetExtras().AsMap()
}

func TestPrompt_HappyPathBindsCallback(t *testing.T) {
	ev := &fakeEvents{}
	jc := &fakeJenga{stkResp: &models.STKUSSDResponse{Status: true, Code: 0, Message: "ok", TransactionID: "JTX-1"}}
	h := queue.NewPromptHandler(ev, jc, &fakeStatus{}, fakeSettings{}, testConfig())
	headers := map[string]string{"tenant_id": "t1", "partition_id": "pa1", config.HeaderConnectionCredentials: "conn-1"}

	require.NoError(t, h.Handle(t.Context(), headers, promptPayload(t, "p1", "254712345678", 150, 500_000_000, nil)))

	require.NotNil(t, jc.stkReq)
	assert.Equal(t, "150.50", jc.stkReq.Payment.Amount)
	assert.Equal(t, "KES", jc.stkReq.Payment.Currency)
	assert.Equal(t, "254712345678", jc.stkReq.Payment.MobileNumber)
	assert.Equal(t, "0011", jc.stkReq.Merchant.AccountNumber)
	assert.NotEqual(t, "p1", jc.stkReq.Payment.Ref, "the reference must not be the prompt id")
	assert.Len(t, jc.stkReq.Payment.Ref, 24)
	assert.Equal(t, "MC-T", jc.creds.MerchantCode, "tenant credentials from settings")
	assert.Equal(t, "tenant-pem", jc.creds.PrivateKeyPEM, "tenant connections sign with their own key")

	cb, err := url.Parse(jc.stkReq.Payment.CallBackUrl)
	require.NoError(t, err)
	assert.Equal(t, models.STKCallbackPath, cb.Path)
	assert.Equal(t, "p1", cb.Query().Get(models.CallbackParamPromptID))
	assert.Equal(t, "t1", cb.Query().Get("tenant_id"))
	assert.Equal(t, "pa1", cb.Query().Get("partition_id"))
	assert.Equal(t, "conn-1", cb.Query().Get("connection"))
	secret := cb.Query().Get(models.CallbackParamSecret)
	assert.Len(t, secret, 32)

	require.Len(t, ev.emitted, 2)
	queued, inProcess := ev.emitted[0], ev.emitted[1]
	assert.Equal(t, commonv1.STATUS_QUEUED, queued.GetStatus())
	assert.Equal(t, "p1", queued.GetId())
	assert.Equal(t, jc.stkReq.Payment.Ref, extrasMap(queued)[models.ExtraTransactionRef])
	assert.Equal(t, models.HashCallbackSecret(secret), extrasMap(queued)[models.ExtraCallbackSecretHash])
	for _, st := range ev.emitted {
		assert.NotContains(t, extrasMap(st), "callback_secret", "the clear secret is never stored")
		for _, v := range extrasMap(st) {
			assert.NotEqual(t, secret, v, "the clear secret is never stored")
		}
	}

	assert.Equal(t, commonv1.STATUS_IN_PROCESS, inProcess.GetStatus())
	assert.Equal(t, "JTX-1", inProcess.GetExternalId())
	ex := extrasMap(inProcess)
	assert.Equal(t, "prompt", ex["entity_type"])
	assert.Equal(t, "jenga", ex[collection.ExtraProvider])
	assert.Equal(t, "JTX-1", ex[models.ExtraTransactionID])
	assert.Equal(t, "150.50", ex[models.ExtraRequestedAmount])
	assert.Equal(t, "KES", ex[models.ExtraRequestedCurrency])
	assert.Equal(t, "conn-1", ex[models.ExtraCredentialsConnection])
	assert.Equal(t, models.HashCallbackSecret(secret), ex[models.ExtraCallbackSecretHash])
}

func TestPrompt_Outcomes(t *testing.T) {
	tests := []struct {
		name       string
		payload    []byte
		jenga      *fakeJenga
		status     *fakeStatus
		cfg        *config.JengaConfig
		wantErr    bool
		wantEmits  int
		wantStatus commonv1.STATUS
		wantCode   string
		wantPushed bool
	}{
		{
			name:      "non-proto payload is dropped without a status",
			payload:   []byte(`{"id":"p1"}`),
			jenga:     &fakeJenga{},
			wantEmits: 0,
		},
		{
			name:       "rejected by Jenga fails the prompt",
			payload:    promptPayload(t, "p1", "254712345678", 100, 0, nil),
			jenga:      &fakeJenga{stkErr: &coreapi.APIError{HTTPStatus: 400, Code: 400, Message: "bad telco"}},
			wantEmits:  2,
			wantStatus: commonv1.STATUS_FAILED,
			wantCode:   "PROVIDER_ERROR",
			wantPushed: true,
		},
		{
			name:       "unknown outcome redelivers after the binding is recorded",
			payload:    promptPayload(t, "p1", "254712345678", 100, 0, nil),
			jenga:      &fakeJenga{stkErr: collection.Uncertain(errors.New("timeout"))},
			wantErr:    true,
			wantEmits:  1,
			wantStatus: commonv1.STATUS_QUEUED,
			wantPushed: true,
		},
		{
			name:       "token transport failure redelivers",
			payload:    promptPayload(t, "p1", "254712345678", 100, 0, nil),
			jenga:      &fakeJenga{tokenErr: collection.Uncertain(errors.New("timeout"))},
			wantErr:    true,
			wantEmits:  1,
			wantStatus: commonv1.STATUS_QUEUED,
		},
		{
			name:       "credentials refused fails the prompt",
			payload:    promptPayload(t, "p1", "254712345678", 100, 0, nil),
			jenga:      &fakeJenga{tokenErr: &coreapi.APIError{HTTPStatus: 401, Message: "invalid"}},
			wantEmits:  2,
			wantStatus: commonv1.STATUS_FAILED,
			wantCode:   "CREDENTIALS_ERROR",
		},
		{
			name:       "missing credentials fails the prompt",
			payload:    promptPayload(t, "p1", "254712345678", 100, 0, nil),
			jenga:      &fakeJenga{},
			cfg:        &config.JengaConfig{JengaCallbackURL: "https://hooks.example.com"},
			wantEmits:  1,
			wantStatus: commonv1.STATUS_FAILED,
			wantCode:   "CREDENTIALS_ERROR",
		},
		{
			name:       "fractional amount beyond cents is refused",
			payload:    promptPayload(t, "p1", "254712345678", 1, 1, nil),
			jenga:      &fakeJenga{},
			wantEmits:  1,
			wantStatus: commonv1.STATUS_FAILED,
			wantCode:   "INVALID_AMOUNT",
		},
		{
			name:       "missing phone is refused",
			payload:    promptPayload(t, "p1", "", 100, 0, nil),
			jenga:      &fakeJenga{},
			wantEmits:  1,
			wantStatus: commonv1.STATUS_FAILED,
			wantCode:   "MISSING_PHONE_NUMBER",
		},
		{
			name:    "already pushed prompt is not pushed again",
			payload: promptPayload(t, "p1", "254712345678", 100, 0, nil),
			jenga:   &fakeJenga{},
			status:  &fakeStatus{extras: map[string]any{models.ExtraTransactionID: "JTX-1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &fakeEvents{}
			cfg := tt.cfg
			if cfg == nil {
				cfg = testConfig()
			}
			status := tt.status
			if status == nil {
				status = &fakeStatus{}
			}
			h := queue.NewPromptHandler(ev, tt.jenga, status, nil, cfg)
			err := h.Handle(t.Context(), nil, tt.payload)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantPushed, tt.jenga.stkReq != nil)
			require.Len(t, ev.emitted, tt.wantEmits)
			if tt.wantEmits == 0 {
				return
			}
			last := ev.emitted[len(ev.emitted)-1]
			assert.Equal(t, tt.wantStatus, last.GetStatus())
			if tt.wantStatus == commonv1.STATUS_FAILED {
				assert.Equal(t, commonv1.STATE_INACTIVE, last.GetState())
				assert.Equal(t, tt.wantCode, extrasMap(last)["failure_code"])
			}
		})
	}
}

func TestPrompt_RedeliveryReusesQueuedBinding(t *testing.T) {
	ev := &fakeEvents{}
	jc := &fakeJenga{stkResp: &models.STKUSSDResponse{Status: true, TransactionID: "JTX-2"}}
	oldHash := models.HashCallbackSecret("secret-from-first-delivery")
	prior := &fakeStatus{extras: map[string]any{
		models.ExtraTransactionRef: "ref-from-first-delivery", models.ExtraCallbackSecretHash: oldHash,
	}}
	h := queue.NewPromptHandler(ev, jc, prior, nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), nil, promptPayload(t, "p1", "254712345678", 100, 0, nil)))

	require.NotNil(t, jc.stkReq)
	assert.Equal(t, "ref-from-first-delivery", jc.stkReq.Payment.Ref, "the reference is kept")
	cb, err := url.Parse(jc.stkReq.Payment.CallBackUrl)
	require.NoError(t, err)
	newSecret := cb.Query().Get(models.CallbackParamSecret)
	assert.Len(t, newSecret, 32)
	require.Len(t, ev.emitted, 2)
	newHash := extrasMap(ev.emitted[0])[models.ExtraCallbackSecretHash]
	assert.Equal(t, models.HashCallbackSecret(newSecret), newHash, "binding re-emitted with the new hash")
	assert.NotEqual(t, oldHash, newHash, "a fresh secret is minted; the previous one no longer binds")
}

func TestPrompt_EmitFailureRedelivers(t *testing.T) {
	ev := &fakeEvents{err: errors.New("queue down")}
	jc := &fakeJenga{}
	h := queue.NewPromptHandler(ev, jc, &fakeStatus{}, nil, testConfig())

	require.Error(t, h.Handle(t.Context(), nil, promptPayload(t, "p1", "254712345678", 100, 0, nil)))
	assert.Nil(t, jc.stkReq, "no push before the binding is recorded")
}

func TestPayment_IsNotSupported(t *testing.T) {
	ev := &fakeEvents{}
	h := queue.NewPaymentHandler(ev)
	raw, err := proto.Marshal(&paymentv1.Payment{Id: "pay-1", Recipient: &commonv1.ContactLink{ContactId: "254712345678"},
		Amount: &commonv1.Money{CurrencyCode: "KES", Units: 100}})
	require.NoError(t, err)

	require.NoError(t, h.Handle(t.Context(), nil, raw))
	require.Len(t, ev.emitted, 1)
	st := ev.emitted[0]
	assert.Equal(t, "pay-1", st.GetId())
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, st.GetState())
	assert.Equal(t, "payment", extrasMap(st)["entity_type"])
	assert.Equal(t, "PAYOUT_NOT_SUPPORTED", extrasMap(st)["failure_code"])
}
