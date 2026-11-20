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
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
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

// fakeMtn answers status lookups from a map keyed by reference id and
// records every request issued.
type fakeMtn struct {
	client.MtnClient
	rtpStatus      map[string]*client.RequestToPayStatus
	transferStatus map[string]*client.TransferStatus
	lookupErr      error
	sendErr        error
	rtpRequests    []*client.RequestToPayRequest
	transfers      []*client.TransferRequest
	creds          []*client.MtnCredentials
}

func notFound() error {
	return &client.APIError{Operation: "get status", StatusCode: http.StatusNotFound}
}

func (m *fakeMtn) RequestToPay(_ context.Context, creds *client.MtnCredentials, req *client.RequestToPayRequest) error {
	m.rtpRequests = append(m.rtpRequests, req)
	m.creds = append(m.creds, creds)
	return m.sendErr
}

func (m *fakeMtn) GetRequestToPayStatus(_ context.Context, _ *client.MtnCredentials, ref string) (*client.RequestToPayStatus, error) {
	if m.lookupErr != nil {
		return nil, m.lookupErr
	}
	if st, ok := m.rtpStatus[ref]; ok {
		return st, nil
	}
	return nil, notFound()
}

func (m *fakeMtn) Transfer(_ context.Context, creds *client.MtnCredentials, req *client.TransferRequest) error {
	m.transfers = append(m.transfers, req)
	m.creds = append(m.creds, creds)
	return m.sendErr
}

func (m *fakeMtn) GetTransferStatus(_ context.Context, _ *client.MtnCredentials, ref string) (*client.TransferStatus, error) {
	if m.lookupErr != nil {
		return nil, m.lookupErr
	}
	if st, ok := m.transferStatus[ref]; ok {
		return st, nil
	}
	return nil, notFound()
}

func testConfig() *config.MtnConfig {
	return &config.MtnConfig{
		SubscriptionKey: "k", APIUser: "u", APIKey: "a", Currency: "EUR",
		Environment: config.EnvironmentSandbox, CallbackURL: "https://hooks.example.com",
	}
}

func promptPayload(t *testing.T, amount *commonv1.Money) []byte {
	t.Helper()
	payload, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:        "prompt-1",
		Recipient: &commonv1.ContactLink{ContactId: "256770000001"},
		Amount:    amount,
	})
	require.NoError(t, err)
	return payload
}

func paymentPayload(t *testing.T, amount *commonv1.Money) []byte {
	t.Helper()
	payload, err := proto.Marshal(&paymentv1.Payment{
		Id:        "payment-1",
		Recipient: &commonv1.ContactLink{ContactId: "256770000001"},
		Amount:    amount,
	})
	require.NoError(t, err)
	return payload
}

func TestReferenceID_Deterministic(t *testing.T) {
	ref := client.ReferenceID(events.EntityTypePrompt, "prompt-1")
	assert.Equal(t, ref, client.ReferenceID(events.EntityTypePrompt, "prompt-1"))
	assert.NotEqual(t, ref, client.ReferenceID(events.EntityTypePayment, "prompt-1"))
	assert.NotEqual(t, ref, client.ReferenceID(events.EntityTypePrompt, "prompt-2"))
	assert.Len(t, ref, 36)
}

func TestPromptHandler_RecordsVerificationBinding(t *testing.T) {
	evts := &recordingEvents{}
	mtn := &fakeMtn{}
	h := NewPromptHandler(evts, mtn, nil, testConfig())

	require.NoError(t, h.Handle(t.Context(), map[string]string{}, promptPayload(t, &commonv1.Money{CurrencyCode: "UGX", Units: 5000})))

	require.Len(t, mtn.rtpRequests, 1)
	assert.Equal(t, client.ReferenceID(events.EntityTypePrompt, "prompt-1"), mtn.rtpRequests[0].ReferenceID)
	assert.Equal(t, "5000", mtn.rtpRequests[0].Amount)
	assert.Equal(t, config.EnvironmentSandbox, mtn.creds[0].TargetEnvironment)
	require.Len(t, evts.emitted, 1)
	st := evts.emitted[0]
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, commonv1.STATE_ACTIVE, st.GetState())
	extras := st.GetExtras().GetFields()
	assert.Equal(t, mtn.rtpRequests[0].ReferenceID, extras[client.ExtraReferenceID].GetStringValue())
	assert.Equal(t, "5000", extras[client.ExtraRequestedAmount].GetStringValue())
	assert.Equal(t, "UGX", extras[client.ExtraRequestedCurrency].GetStringValue())
	assert.Equal(t, "prompt", extras["entity_type"].GetStringValue())
}

func TestPromptHandler_Outcomes(t *testing.T) {
	ref := client.ReferenceID(events.EntityTypePrompt, "prompt-1")
	tests := []struct {
		name        string
		amount      *commonv1.Money
		mtn         *fakeMtn
		settingsHdr bool
		wantErr     bool
		wantSent    int
		wantStatus  commonv1.STATUS
		wantNoEmit  bool
	}{
		{
			name:   "already issued: reports MTN's state without resending",
			amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			mtn: &fakeMtn{rtpStatus: map[string]*client.RequestToPayStatus{
				ref: {Status: "SUCCESSFUL", FinancialTransactionID: "fin-1"},
			}},
			wantStatus: commonv1.STATUS_SUCCESSFUL,
		},
		{
			name:       "rejected by MTN: FAILED",
			amount:     &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			mtn:        &fakeMtn{sendErr: &client.APIError{Operation: "rtp", StatusCode: http.StatusBadRequest}},
			wantSent:   1,
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name:       "outcome unknown and no record at MTN: redelivered",
			amount:     &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			mtn:        &fakeMtn{sendErr: collection.Uncertain(errors.New("timeout"))},
			wantSent:   1,
			wantErr:    true,
			wantNoEmit: true,
		},
		{
			name:   "duplicate reference: reports what MTN holds",
			amount: &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			mtn: &fakeMtn{
				sendErr:   &client.APIError{Operation: "rtp", StatusCode: http.StatusConflict},
				rtpStatus: map[string]*client.RequestToPayStatus{ref: {Status: "PENDING"}},
			},
			// The pre-send lookup already finds the record, so nothing is sent.
			wantStatus: commonv1.STATUS_IN_PROCESS,
		},
		{
			name:       "fractional UGX amount: FAILED, never rounded",
			amount:     &commonv1.Money{CurrencyCode: "UGX", Units: 5000, Nanos: 500_000_000},
			mtn:        &fakeMtn{},
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name:        "settings connection without settings client: FAILED credentials",
			amount:      &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
			mtn:         &fakeMtn{},
			settingsHdr: true,
			wantStatus:  commonv1.STATUS_FAILED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evts := &recordingEvents{}
			h := NewPromptHandler(evts, tt.mtn, nil, testConfig())
			headers := map[string]string{}
			if tt.settingsHdr {
				headers[config.HeaderConnectionCredentials] = "conn-1"
			}

			err := h.Handle(t.Context(), headers, promptPayload(t, tt.amount))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, tt.mtn.rtpRequests, tt.wantSent)
			if tt.wantNoEmit {
				assert.Empty(t, evts.emitted)
				return
			}
			require.Len(t, evts.emitted, 1)
			assert.Equal(t, tt.wantStatus, evts.emitted[0].GetStatus())
			assert.Equal(t, events.StateFor(tt.wantStatus), evts.emitted[0].GetState())
			if tt.wantStatus == commonv1.STATUS_FAILED {
				assert.NotEmpty(t, evts.emitted[0].GetExtras().GetFields()[events.ExtraError].GetStringValue())
			}
		})
	}
}

func TestPromptHandler_EmitErrorIsReturned(t *testing.T) {
	evts := &recordingEvents{err: errors.New("bus down")}
	h := NewPromptHandler(evts, &fakeMtn{}, nil, testConfig())

	err := h.Handle(t.Context(), map[string]string{}, promptPayload(t, &commonv1.Money{CurrencyCode: "UGX", Units: 10}))

	require.Error(t, err)
}

func TestPaymentHandler_Outcomes(t *testing.T) {
	ref := client.ReferenceID(events.EntityTypePayment, "payment-1")
	tests := []struct {
		name       string
		amount     *commonv1.Money
		mtn        *fakeMtn
		wantErr    bool
		wantSent   int
		wantStatus commonv1.STATUS
		wantNoEmit bool
		wantExtID  string
	}{
		{
			name:       "transfer issued with deterministic reference",
			amount:     &commonv1.Money{CurrencyCode: "EUR", Units: 10, Nanos: 500_000_000},
			mtn:        &fakeMtn{},
			wantSent:   1,
			wantStatus: commonv1.STATUS_IN_PROCESS,
			wantExtID:  ref,
		},
		{
			name:   "already issued and failed at MTN: FAILED with reason",
			amount: &commonv1.Money{CurrencyCode: "EUR", Units: 10},
			mtn: &fakeMtn{transferStatus: map[string]*client.TransferStatus{
				ref: {Status: "FAILED", Reason: &struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				}{Code: "PAYEE_NOT_FOUND", Message: "no such payee"}},
			}},
			wantStatus: commonv1.STATUS_FAILED,
		},
		{
			name:       "5xx then lookup fails: redelivered",
			amount:     &commonv1.Money{CurrencyCode: "EUR", Units: 10},
			mtn:        &fakeMtn{sendErr: collection.Uncertain(&client.APIError{StatusCode: 502})},
			wantSent:   1,
			wantErr:    true,
			wantNoEmit: true,
		},
		{
			name:       "lookup error does not block sending",
			amount:     &commonv1.Money{CurrencyCode: "EUR", Units: 10},
			mtn:        &fakeMtn{lookupErr: collection.Uncertain(errors.New("timeout"))},
			wantSent:   1,
			wantStatus: commonv1.STATUS_IN_PROCESS,
			wantExtID:  ref,
		},
		{
			name:       "zero amount: FAILED",
			amount:     &commonv1.Money{CurrencyCode: "EUR"},
			mtn:        &fakeMtn{},
			wantStatus: commonv1.STATUS_FAILED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evts := &recordingEvents{}
			h := NewPaymentHandler(evts, tt.mtn, nil, testConfig())

			err := h.Handle(t.Context(), map[string]string{"tenant_id": "t1"}, paymentPayload(t, tt.amount))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Len(t, tt.mtn.transfers, tt.wantSent)
			if tt.wantSent > 0 {
				assert.Equal(t, ref, tt.mtn.transfers[0].ReferenceID)
				assert.Contains(t, tt.mtn.transfers[0].CallbackURL, "tenant_id=t1")
			}
			if tt.wantNoEmit {
				assert.Empty(t, evts.emitted)
				return
			}
			require.Len(t, evts.emitted, 1)
			assert.Equal(t, tt.wantStatus, evts.emitted[0].GetStatus())
			assert.Equal(t, events.StateFor(tt.wantStatus), evts.emitted[0].GetState())
			if tt.wantExtID != "" {
				assert.Equal(t, tt.wantExtID, evts.emitted[0].GetExternalId())
			}
			assert.Equal(t, "payment", evts.emitted[0].GetExtras().GetFields()["entity_type"].GetStringValue())
		})
	}
}

func TestMtnAmount(t *testing.T) {
	tests := []struct {
		name    string
		money   *commonv1.Money
		want    string
		wantErr bool
	}{
		{name: "whole UGX", money: &commonv1.Money{CurrencyCode: "UGX", Units: 5000}, want: "5000"},
		{name: "whole EUR", money: &commonv1.Money{CurrencyCode: "EUR", Units: 10}, want: "10"},
		{name: "EUR cents", money: &commonv1.Money{CurrencyCode: "EUR", Units: 10, Nanos: 500_000_000}, want: "10.50"},
		{name: "fractional UGX refused", money: &commonv1.Money{CurrencyCode: "UGX", Units: 1, Nanos: 500_000_000}, wantErr: true},
		{name: "sub-cent EUR refused", money: &commonv1.Money{CurrencyCode: "EUR", Units: 1, Nanos: 1_000_000}, wantErr: true},
		{name: "nil refused", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mtnAmount(tt.money)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequestExtras_Connection(t *testing.T) {
	extras := requestExtras("payment", "ref-1", "10", "UGX",
		map[string]string{config.HeaderConnectionCredentials: "conn-1"})
	assert.Equal(t, "conn-1", extras[client.ExtraCredentialsConnection])
	assert.Equal(t, "payment", extras["entity_type"])

	extras = requestExtras("payment", "ref-1", "10", "UGX", map[string]string{})
	assert.NotContains(t, extras, client.ExtraCredentialsConnection)
}

func TestExtractCredentials_TargetEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.MtnConfig
		headers map[string]string
		want    string
		wantErr bool
	}{
		{name: "sandbox default", cfg: testConfig(), want: "sandbox"},
		{
			name:    "production requires a market code",
			cfg:     &config.MtnConfig{SubscriptionKey: "k", APIUser: "u", APIKey: "a", Environment: config.EnvironmentProduction},
			wantErr: true,
		},
		{
			name: "production uses the configured market code",
			cfg: &config.MtnConfig{SubscriptionKey: "k", APIUser: "u", APIKey: "a",
				Environment: config.EnvironmentProduction, TargetEnvironment: "mtnuganda"},
			want: "mtnuganda",
		},
		{
			name:    "header override wins",
			cfg:     testConfig(),
			headers: map[string]string{config.HeaderTargetEnvironment: "mtnghana"},
			want:    "mtnghana",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds, err := extractCredentials(t.Context(), tt.headers, nil, tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, creds.TargetEnvironment)
		})
	}
}
