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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/config"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

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

type fakeAirtel struct {
	client.AirtelClient
	known         map[string]*client.StatusResponse
	lookupErr     error
	sendErr       error
	collections   []*client.CollectionRequest
	disbursements []*client.DisbursementRequest
}

func statusOf(id, status string) *client.StatusResponse {
	r := &client.StatusResponse{}
	r.Data.Transaction.ID = id
	r.Data.Transaction.Status = status
	r.Data.Transaction.AirtelMoneyID = "AM.1"
	r.Status.Success = true
	return r
}

func (a *fakeAirtel) lookup(id string) (*client.StatusResponse, error) {
	if a.lookupErr != nil {
		return nil, a.lookupErr
	}
	if st, ok := a.known[id]; ok {
		return st, nil
	}
	return nil, &client.APIError{Operation: "status", StatusCode: http.StatusNotFound}
}

func (a *fakeAirtel) TransactionStatus(_ context.Context, _ *client.AirtelCredentials, id string) (*client.StatusResponse, error) {
	return a.lookup(id)
}

func (a *fakeAirtel) DisbursementStatus(_ context.Context, _ *client.AirtelCredentials, id string) (*client.StatusResponse, error) {
	return a.lookup(id)
}

func (a *fakeAirtel) CollectionPush(_ context.Context, _ *client.AirtelCredentials, req *client.CollectionRequest) (*client.CollectionResponse, error) {
	a.collections = append(a.collections, req)
	if a.sendErr != nil {
		return nil, a.sendErr
	}
	r := &client.CollectionResponse{}
	r.Data.Transaction.ID = req.Reference
	r.Data.Transaction.Status = "TIP"
	r.Status.Success = true
	return r, nil
}

func (a *fakeAirtel) Disburse(_ context.Context, _ *client.AirtelCredentials, req *client.DisbursementRequest) (*client.DisbursementResponse, error) {
	a.disbursements = append(a.disbursements, req)
	if a.sendErr != nil {
		return nil, a.sendErr
	}
	r := &client.DisbursementResponse{}
	r.Data.Transaction.ID = req.Reference
	r.Data.Transaction.Status = "TIP"
	r.Status.Success = true
	return r, nil
}

func publicKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func testConfig(t *testing.T) *config.AirtelConfig {
	t.Helper()
	return &config.AirtelConfig{
		ClientID: "id", ClientSecret: "s", CountryCode: "UG", Currency: "UGX",
		DisbursementPin: "1234", PublicKey: publicKeyPEM(t),
	}
}

func promptPayload(t *testing.T, amount *commonv1.Money) []byte {
	t.Helper()
	payload, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id: "prompt-1", Recipient: &commonv1.ContactLink{ContactId: "256700000001"}, Amount: amount,
	})
	require.NoError(t, err)
	return payload
}

func paymentPayload(t *testing.T, amount *commonv1.Money) []byte {
	t.Helper()
	payload, err := proto.Marshal(&paymentv1.Payment{
		Id: "payment-1", Recipient: &commonv1.ContactLink{ContactId: "256700000001"}, Amount: amount,
	})
	require.NoError(t, err)
	return payload
}

func TestPaymentHandler_Outcomes(t *testing.T) {
	tests := []struct {
		name       string
		cfg        func(t *testing.T) *config.AirtelConfig
		amount     *commonv1.Money
		airtel     *fakeAirtel
		wantSent   int
		wantErr    bool
		wantNoEmit bool
		wantStatus commonv1.STATUS
		wantPin    bool
	}{
		{
			name: "disbursement sent with encrypted pin", cfg: testConfig,
			amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000}, airtel: &fakeAirtel{},
			wantSent: 1, wantStatus: commonv1.STATUS_IN_PROCESS, wantPin: true,
		},
		{
			name: "missing pin: FAILED, nothing sent",
			cfg: func(t *testing.T) *config.AirtelConfig {
				c := testConfig(t)
				c.DisbursementPin = ""
				return c
			},
			amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000}, airtel: &fakeAirtel{},
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name: "missing public key: FAILED, nothing sent",
			cfg: func(t *testing.T) *config.AirtelConfig {
				c := testConfig(t)
				c.PublicKey = ""
				return c
			},
			amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000}, airtel: &fakeAirtel{},
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name: "fractional amount: FAILED, never rounded", cfg: testConfig,
			amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000, Nanos: 500_000_000}, airtel: &fakeAirtel{},
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name: "already sent: reports Airtel's state", cfg: testConfig,
			amount:     &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			airtel:     &fakeAirtel{known: map[string]*client.StatusResponse{"payment-1": statusOf("payment-1", "TS")}},
			wantStatus: commonv1.STATUS_SUCCESSFUL,
		},
		{
			name: "rejected by Airtel: FAILED", cfg: testConfig,
			amount:   &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			airtel:   &fakeAirtel{sendErr: &client.APIError{Operation: "disburse", StatusCode: http.StatusOK, Code: "DP008", Message: "insufficient"}},
			wantSent: 1, wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name: "unknown outcome, nothing at Airtel: redelivered", cfg: testConfig,
			amount:   &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			airtel:   &fakeAirtel{sendErr: collection.Uncertain(errors.New("timeout"))},
			wantSent: 1, wantErr: true, wantNoEmit: true,
		},
		{
			name: "lookup failure does not block sending", cfg: testConfig,
			amount:   &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			airtel:   &fakeAirtel{lookupErr: collection.Uncertain(errors.New("timeout"))},
			wantSent: 1, wantStatus: commonv1.STATUS_IN_PROCESS, wantPin: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evts := &recordingEvents{}
			h := NewPaymentHandler(evts, tt.airtel, nil, tt.cfg(t))

			err := h.Handle(t.Context(), map[string]string{}, paymentPayload(t, tt.amount))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Len(t, tt.airtel.disbursements, tt.wantSent)
			if tt.wantPin {
				assert.NotEmpty(t, tt.airtel.disbursements[0].EncryptedPin)
				assert.NotEqual(t, "1234", tt.airtel.disbursements[0].EncryptedPin)
				assert.Equal(t, "5000", tt.airtel.disbursements[0].Amount)
			}
			if tt.wantNoEmit {
				assert.Empty(t, evts.emitted)
				return
			}
			require.Len(t, evts.emitted, 1)
			st := evts.emitted[0]
			assert.Equal(t, "payment-1", st.GetId())
			assert.Equal(t, tt.wantStatus, st.GetStatus())
			assert.Equal(t, events.StateFor(tt.wantStatus), st.GetState())
			extras := st.GetExtras().GetFields()
			assert.Equal(t, "payment", extras["entity_type"].GetStringValue())
			if tt.wantStatus == commonv1.STATUS_FAILED {
				assert.NotEmpty(t, extras[events.ExtraError].GetStringValue())
			} else {
				assert.Equal(t, "payment-1", extras[client.ExtraTransactionID].GetStringValue())
				assert.Equal(t, "5000", extras[client.ExtraRequestedAmount].GetStringValue())
				assert.Equal(t, "UGX", extras[client.ExtraRequestedCurrency].GetStringValue())
			}
		})
	}
}

func TestPromptHandler_Outcomes(t *testing.T) {
	tests := []struct {
		name       string
		amount     *commonv1.Money
		airtel     *fakeAirtel
		wantSent   int
		wantErr    bool
		wantNoEmit bool
		wantStatus commonv1.STATUS
	}{
		{
			name: "push sent", amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000}, airtel: &fakeAirtel{},
			wantSent: 1, wantStatus: commonv1.STATUS_IN_PROCESS,
		},
		{
			name: "already pushed and failed: FAILED without resending", amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			airtel:     &fakeAirtel{known: map[string]*client.StatusResponse{"prompt-1": statusOf("prompt-1", "TF")}},
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name: "rejected: FAILED", amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			airtel:   &fakeAirtel{sendErr: &client.APIError{Operation: "collection_push", StatusCode: http.StatusBadRequest}},
			wantSent: 1, wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name: "unknown outcome then found at Airtel: reports it", amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			airtel: &fakeAirtel{sendErr: collection.Uncertain(errors.New("502"))},
			// The pre-send lookup finds nothing; after the failed send the
			// fake still finds nothing, so the message is redelivered.
			wantSent: 1, wantErr: true, wantNoEmit: true,
		},
		{
			name: "zero amount refused", amount: &commonv1.Money{CurrencyCode: "UGX"}, airtel: &fakeAirtel{},
			wantStatus: commonv1.STATUS_FAILED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evts := &recordingEvents{}
			h := NewPromptHandler(evts, tt.airtel, nil, testConfig(t))

			err := h.Handle(t.Context(), map[string]string{config.HeaderConnectionCredentials: ""}, promptPayload(t, tt.amount))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Len(t, tt.airtel.collections, tt.wantSent)
			if tt.wantNoEmit {
				assert.Empty(t, evts.emitted)
				return
			}
			require.Len(t, evts.emitted, 1)
			assert.Equal(t, tt.wantStatus, evts.emitted[0].GetStatus())
			assert.Equal(t, events.StateFor(tt.wantStatus), evts.emitted[0].GetState())
			assert.Equal(t, "prompt", evts.emitted[0].GetExtras().GetFields()["entity_type"].GetStringValue())
		})
	}
}

func TestPromptHandler_EmitErrorIsReturned(t *testing.T) {
	evts := &recordingEvents{err: errors.New("bus down")}
	h := NewPromptHandler(evts, &fakeAirtel{}, nil, testConfig(t))
	require.Error(t, h.Handle(t.Context(), map[string]string{}, promptPayload(t, &commonv1.Money{CurrencyCode: "UGX", Units: 10})))
}

func TestCredentials_SettingsConnectionWithoutClient(t *testing.T) {
	resolve := NewCredentialResolver(nil, testConfig(t))
	_, err := resolve(t.Context(), "conn-1")
	require.Error(t, err)

	creds, err := resolve(t.Context(), "")
	require.NoError(t, err)
	assert.Equal(t, "1234", creds.DisbursementPin)
}
