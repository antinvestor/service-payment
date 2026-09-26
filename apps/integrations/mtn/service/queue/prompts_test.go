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
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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

type recordingMtn struct {
	client.MtnClient
	requests []*client.RequestToPayRequest
}

func (m *recordingMtn) RequestToPay(_ context.Context, _ *client.MtnCredentials, req *client.RequestToPayRequest) error {
	m.requests = append(m.requests, req)
	return nil
}

func TestPromptHandler_RecordsVerificationBinding(t *testing.T) {
	evts := &recordingEvents{}
	mtn := &recordingMtn{}
	cfg := &config.MtnConfig{SubscriptionKey: "k", APIUser: "u", APIKey: "a", Currency: "EUR"}
	h := NewPromptHandler(evts, mtn, nil, cfg)

	payload, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:        "prompt-1",
		Recipient: &commonv1.ContactLink{ContactId: "256770000001"},
		Amount:    &commonv1.Money{CurrencyCode: "UGX", Units: 5000},
	})
	require.NoError(t, err)
	require.NoError(t, h.Handle(t.Context(), map[string]string{}, payload))

	require.Len(t, mtn.requests, 1)
	require.Len(t, evts.emitted, 1)
	st := evts.emitted[0]
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	extras := st.GetExtras().GetFields()
	assert.Equal(t, mtn.requests[0].ReferenceID, extras[client.ExtraReferenceID].GetStringValue())
	assert.Equal(t, "5000", extras[client.ExtraRequestedAmount].GetStringValue())
	assert.Equal(t, "UGX", extras[client.ExtraRequestedCurrency].GetStringValue())
	assert.Equal(t, "prompt", extras["entity_type"].GetStringValue())
}

func TestRequestExtras_Connection(t *testing.T) {
	extras := requestExtras("payment", "ref-1", "10", "UGX",
		map[string]string{config.HeaderConnectionCredentials: "conn-1"})
	assert.Equal(t, "conn-1", extras[client.ExtraCredentialsConnection])
	assert.Equal(t, "payment", extras["entity_type"])

	extras = requestExtras("payment", "ref-1", "10", "UGX", map[string]string{})
	assert.NotContains(t, extras, client.ExtraCredentialsConnection)
}
