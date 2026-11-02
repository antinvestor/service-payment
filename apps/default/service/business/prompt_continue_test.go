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

package business

import (
	"context"
	"errors"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"connectrpc.com/connect/v2"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/queue"
	utilmoney "github.com/pitabwire/util/moneyx"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/antinvestor/service-payments/apps/default/service/metrics"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
)

// promptsByID holds raised prompts.
type promptsByID struct {
	repository.PromptRepository

	byID map[string]*models.Prompt
	err  error
}

func (r *promptsByID) GetByID(_ context.Context, id string) (*models.Prompt, error) {
	if r.err != nil {
		return nil, r.err
	}
	if p, ok := r.byID[id]; ok {
		return p, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// publishedPrompts records what reached the prompt queue.
type publishedPrompts struct {
	queue.Manager

	topics   []string
	payloads [][]byte
}

func (q *publishedPrompts) GetPublisher(string) (queue.Publisher, error) {
	return nil, errors.New("no route publisher")
}

func (q *publishedPrompts) Publish(_ context.Context, topic string, payload any, _ ...map[string]string) error {
	q.topics = append(q.topics, topic)
	q.payloads = append(q.payloads, payload.([]byte))
	return nil
}

func (q *publishedPrompts) last(t *testing.T) *paymentv1.InitiatePromptRequest {
	t.Helper()
	require.NotEmpty(t, q.payloads, "nothing was published")
	sent := &paymentv1.InitiatePromptRequest{}
	require.NoError(t, proto.Unmarshal(q.payloads[len(q.payloads)-1], sent))
	return sent
}

type continuationFixture struct {
	pb       *paymentBusiness
	prompt   *models.Prompt
	statuses *statusesByEntity
	queue    *publishedPrompts
}

func newContinuationFixture(status commonv1.STATUS, reported map[string]any) *continuationFixture {
	amount := utilmoney.FromMoney(&commonv1.Money{CurrencyCode: "KES", Units: 1500})
	prompt := &models.Prompt{
		SourceID:        "payer-1",
		SourceContactID: "+254700000000",
		RecipientID:     "payer-1",
		Amount:          &amount,
		Route:           "flutterwave",
		State:           int32(commonv1.STATE_CREATED.Number()),
		Status:          int32(commonv1.STATUS_QUEUED.Number()),
		Extra:           data.JSONMap{"currency": "KES", "customer_email": "payer@example.com"},
	}
	prompt.ID = "prompt-1"
	prompt.TenantID = "tenant-1"
	prompt.PartitionID = "partition-1"

	statuses := &statusesByEntity{byEntity: map[string]*models.Status{}}
	if status != commonv1.STATUS_UNKNOWN {
		latest := &models.Status{
			EntityID:   prompt.ID,
			EntityType: "prompt",
			State:      int32(commonv1.STATE_ACTIVE.Number()),
			Status:     int32(status.Number()),
			Extra:      data.JSONMap(reported),
		}
		latest.ID = "status-1"
		statuses.byEntity[prompt.ID] = latest
	}
	q := &publishedPrompts{}
	return &continuationFixture{
		pb: &paymentBusiness{
			qMan:            q,
			promptTopicName: "initiate.prompt",
			promptRepo:      &promptsByID{byID: map[string]*models.Prompt{prompt.ID: prompt}},
			statusRepo:      statuses,
			obs:             metrics.NewMetrics(),
		},
		prompt:   prompt,
		statuses: statuses,
		queue:    q,
	}
}

func authorizeRequest(t *testing.T, id string, extras map[string]any) *paymentv1.InitiatePromptRequest {
	t.Helper()
	base := map[string]any{"action": "authorize"}
	for k, v := range extras {
		base[k] = v
	}
	extra, err := structpb.NewStruct(base)
	require.NoError(t, err)
	return &paymentv1.InitiatePromptRequest{
		Id:     id,
		Route:  "flutterwave",
		Source: &commonv1.ContactLink{ProfileId: "payer-1"},
		Amount: &commonv1.Money{CurrencyCode: "KES", Units: 1500},
		Extra:  extra,
	}
}

func TestAuthorizeStepReachesTheProvider(t *testing.T) {
	f := newContinuationFixture(commonv1.STATUS_IN_PROCESS, map[string]any{
		"charge_id": "chg_123", "next_action": "requires_pin",
	})

	resp, err := f.pb.InitiatePrompt(context.Background(), authorizeRequest(t, "prompt-1", map[string]any{
		"authorization_type": "pin",
		"charge_id":          "chg_123",
		"encrypted_pin":      "cipher",
		"nonce":              "n0nce",
		"session_ref":        "sess-1",
	}))
	require.NoError(t, err)
	require.Equal(t, commonv1.STATUS_IN_PROCESS, resp.GetStatus(), "answers where the prompt stands")

	require.Equal(t, []string{"initiate.prompt"}, f.queue.topics)
	sent := f.queue.last(t)
	extras := sent.GetExtra().AsMap()
	require.Equal(t, "prompt-1", sent.GetId())
	require.Equal(t, "flutterwave", sent.GetRoute())
	require.Equal(t, "authorize", extras["action"])
	require.Equal(t, "pin", extras["authorization_type"])
	require.Equal(t, "chg_123", extras["charge_id"])
	require.Equal(t, "cipher", extras["encrypted_pin"])
	require.Equal(t, "n0nce", extras["nonce"])
	require.Equal(t, "tenant-1", extras["tenant_id"], "tenancy comes from the stored prompt")
	require.Equal(t, "payer@example.com", extras["customer_email"])
	require.NotContains(t, extras, "session_ref", "only authorization fields are forwarded")
	require.Equal(t, int64(1500), sent.GetAmount().GetUnits())

	// Nothing about the step is stored on the prompt.
	require.NotContains(t, f.prompt.Extra, "encrypted_pin")
	require.Len(t, f.statuses.byEntity, 1)
}

func TestAuthorizeStepFillsTheReportedCharge(t *testing.T) {
	f := newContinuationFixture(commonv1.STATUS_IN_PROCESS, map[string]any{"charge_id": "chg_123"})

	_, err := f.pb.InitiatePrompt(context.Background(), authorizeRequest(t, "prompt-1", map[string]any{
		"authorization_type": "otp", "otp": "123456",
	}))
	require.NoError(t, err)
	extras := f.queue.last(t).GetExtra().AsMap()
	require.Equal(t, "chg_123", extras["charge_id"])
	require.Equal(t, "123456", extras["otp"])
}

func TestAuthorizeStepIsRefused(t *testing.T) {
	tests := []struct {
		name     string
		status   commonv1.STATUS
		reported map[string]any
		req      func(t *testing.T) *paymentv1.InitiatePromptRequest
		code     connect.Code
	}{
		{
			name:   "prompt already paid",
			status: commonv1.STATUS_SUCCESSFUL,
			req: func(t *testing.T) *paymentv1.InitiatePromptRequest {
				return authorizeRequest(t, "prompt-1", map[string]any{"authorization_type": "otp", "otp": "1"})
			},
			code: connect.CodeFailedPrecondition,
		},
		{
			name:   "prompt already failed",
			status: commonv1.STATUS_FAILED,
			req: func(t *testing.T) *paymentv1.InitiatePromptRequest {
				return authorizeRequest(t, "prompt-1", map[string]any{"authorization_type": "otp", "otp": "1"})
			},
			code: connect.CodeFailedPrecondition,
		},
		{
			name:     "another charge",
			status:   commonv1.STATUS_IN_PROCESS,
			reported: map[string]any{"charge_id": "chg_123"},
			req: func(t *testing.T) *paymentv1.InitiatePromptRequest {
				return authorizeRequest(t, "prompt-1", map[string]any{
					"authorization_type": "otp", "otp": "1", "charge_id": "chg_other",
				})
			},
			code: connect.CodeInvalidArgument,
		},
		{
			name:   "another route",
			status: commonv1.STATUS_IN_PROCESS,
			req: func(t *testing.T) *paymentv1.InitiatePromptRequest {
				r := authorizeRequest(t, "prompt-1", map[string]any{"authorization_type": "otp", "otp": "1"})
				r.Route = "fincra"
				return r
			},
			code: connect.CodeInvalidArgument,
		},
		{
			name:   "prompt never raised",
			status: commonv1.STATUS_IN_PROCESS,
			req: func(t *testing.T) *paymentv1.InitiatePromptRequest {
				return authorizeRequest(t, "prompt-unknown", map[string]any{"authorization_type": "otp", "otp": "1"})
			},
			code: connect.CodeNotFound,
		},
		{
			name:   "no prompt named",
			status: commonv1.STATUS_IN_PROCESS,
			req: func(t *testing.T) *paymentv1.InitiatePromptRequest {
				return authorizeRequest(t, "", map[string]any{"authorization_type": "otp", "otp": "1"})
			},
			code: connect.CodeInvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newContinuationFixture(tt.status, tt.reported)
			_, err := f.pb.InitiatePrompt(context.Background(), tt.req(t))
			require.Error(t, err)
			require.Equal(t, tt.code, connect.CodeOf(err))
			require.Empty(t, f.queue.topics, "nothing reaches the provider")
		})
	}
}

func TestAskingAgainAnswersTheLatestStatus(t *testing.T) {
	f := newContinuationFixture(commonv1.STATUS_SUCCESSFUL, nil)
	again := &paymentv1.InitiatePromptRequest{
		Id:     "prompt-1",
		Route:  "flutterwave",
		Source: &commonv1.ContactLink{ProfileId: "payer-1"},
		Amount: &commonv1.Money{CurrencyCode: "KES", Units: 1500},
	}
	resp, err := f.pb.InitiatePrompt(context.Background(), again)
	require.NoError(t, err)
	require.Equal(t, commonv1.STATUS_SUCCESSFUL, resp.GetStatus(), "not the status the prompt was raised with")
	require.Empty(t, f.queue.topics, "the payer is not pushed again")
}

func TestAskingAgainWhenTheLookupFailsIsRefused(t *testing.T) {
	f := newContinuationFixture(commonv1.STATUS_IN_PROCESS, nil)
	f.pb.promptRepo = &promptsByID{err: errors.New("db down")}
	_, err := f.pb.InitiatePrompt(context.Background(), &paymentv1.InitiatePromptRequest{
		Id:     "prompt-1",
		Source: &commonv1.ContactLink{ProfileId: "payer-1"},
	})
	require.Error(t, err)
	require.Empty(t, f.queue.topics)
}

func TestCardDataIsSentButNeverStored(t *testing.T) {
	extra, err := structpb.NewStruct(map[string]any{
		"encrypted_card_number": "c1", "encrypted_cvv": "c2", "card_nonce": "n",
		"encrypted_expiry_month": "m", "encrypted_expiry_year": "y",
		"customer_email": "payer@example.com",
	})
	require.NoError(t, err)

	var stored data.JSONMap
	stored = storedExtras(stored.FromProtoStruct(extra))
	require.Equal(t, data.JSONMap{"customer_email": "payer@example.com"}, stored)

	prompt := &models.Prompt{Extra: stored}
	sent := prompt.ToAPI(nil)
	withTransientExtras(sent, extra)
	got := sent.GetExtra().AsMap()
	for _, key := range []string{"encrypted_card_number", "encrypted_cvv", "card_nonce", "encrypted_expiry_month"} {
		require.Contains(t, got, key)
	}
}
