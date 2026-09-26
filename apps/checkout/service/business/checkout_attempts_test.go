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

package business_test

import (
	"context"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func successWith(t *testing.T, id string, extras map[string]any) *commonv1.StatusResponse {
	t.Helper()
	st, err := structpb.NewStruct(extras)
	require.NoError(t, err)
	return &commonv1.StatusResponse{Id: id, Status: commonv1.STATUS_SUCCESSFUL, Extras: st}
}

func processingSession(ref, promptID string, lastAttempt time.Time, history ...any) *models.CheckoutSession {
	s := &models.CheckoutSession{
		Ref:           ref,
		Status:        models.SessionStatusProcessing,
		ExpiresAt:     fixedNow().Add(20 * time.Minute),
		Amount:        "100.00",
		Currency:      "KES",
		AmountOption:  models.AmountOptionFixed,
		PromptID:      promptID,
		Attempts:      1,
		LastAttemptAt: &lastAttempt,
		Metadata:      map[string]any{},
	}
	if len(history) > 0 {
		s.Metadata["_prompt_ids"] = history
	}
	return s
}

func TestPay_ProcessingSession_SettlesPreviousPromptFirst(t *testing.T) {
	cfg := defaultConfig()
	cfg.PromptTimeoutSeconds = 180

	tests := []struct {
		name          string
		sentAgo       time.Duration
		statusByID    map[string]*commonv1.StatusResponse
		wantErr       error
		wantStatus    string
		wantNewPrompt bool
	}{
		{
			name:    "previous prompt succeeded: complete, no new attempt",
			sentAgo: 30 * time.Second,
			statusByID: map[string]*commonv1.StatusResponse{
				"prompt-old": successWith(t, "prompt-old", map[string]any{"Amount": "100"}),
			},
			wantErr:    business.ErrSessionGone,
			wantStatus: models.SessionStatusCompleted,
		},
		{
			name:       "previous prompt still live: refuse",
			sentAgo:    30 * time.Second,
			wantErr:    business.ErrPaymentInProgress,
			wantStatus: models.SessionStatusProcessing,
		},
		{
			name:          "previous prompt timed out: allow a new attempt",
			sentAgo:       4 * time.Minute,
			wantStatus:    models.SessionStatusProcessing,
			wantNewPrompt: true,
		},
		{
			name:    "previous prompt failed: allow a new attempt",
			sentAgo: 30 * time.Second,
			statusByID: map[string]*commonv1.StatusResponse{
				"prompt-old": {Id: "prompt-old", Status: commonv1.STATUS_FAILED},
			},
			wantStatus:    models.SessionStatusProcessing,
			wantNewPrompt: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := newFakeSessionRepo()
			statusByID := tt.statusByID
			if statusByID == nil {
				statusByID = map[string]*commonv1.StatusResponse{}
			}
			payCli := &fakePaymentClient{statusByID: statusByID}
			b := newBusiness(cfg, defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})
			sessionRepo.sessions["sess-p"] = processingSession("sess-p", "prompt-old", fixedNow().Add(-tt.sentAgo), "prompt-old")

			updated, err := b.Pay(context.Background(), "sess-p", business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"})

			stored := sessionRepo.sessions["sess-p"]
			assert.Equal(t, tt.wantStatus, stored.Status)
			if !tt.wantNewPrompt {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, payCli.lastPrompt, "no new prompt may be sent")
				assert.Equal(t, 1, stored.Attempts)
				assert.Equal(t, "prompt-old", stored.PromptID)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, payCli.lastPrompt)
			assert.Equal(t, "prompt-123", updated.PromptID)
			assert.Equal(t, []any{"prompt-old", "prompt-123"}, updated.Metadata["_prompt_ids"])
		})
	}
}

func TestPay_ErrPaymentInProgress_IsNotAMethodError(t *testing.T) {
	assert.NotErrorIs(t, business.ErrPaymentInProgress, business.ErrUnknownMethod)
}

func TestRefreshStatus_LateSuccessOnEarlierPrompt(t *testing.T) {
	for _, status := range []string{models.SessionStatusProcessing, models.SessionStatusFailed, models.SessionStatusExpired} {
		t.Run(status, func(t *testing.T) {
			sessionRepo := newFakeSessionRepo()
			payCli := &fakePaymentClient{statusByID: map[string]*commonv1.StatusResponse{
				"prompt-1": successWith(t, "prompt-1", map[string]any{"Amount": "100"}),
				"prompt-2": {Id: "prompt-2", Status: commonv1.STATUS_FAILED},
			}}
			b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})
			s := processingSession("sess-late", "prompt-2", fixedNow(), "prompt-1", "prompt-2")
			s.Status = status
			sessionRepo.sessions["sess-late"] = s

			updated, err := b.RefreshStatus(context.Background(), s)

			require.NoError(t, err)
			assert.Equal(t, models.SessionStatusCompleted, updated.Status)
			assert.Equal(t, "prompt-1", updated.PromptID, "session points at the prompt that paid")
			assert.Equal(t, []string{"prompt-1"}, payCli.polled, "earlier prompts are polled, stopping at the first verified success")
		})
	}
}

func TestRefreshStatus_FailedOrExpiredOnlyChangeOnSuccess(t *testing.T) {
	for _, status := range []string{models.SessionStatusFailed, models.SessionStatusExpired} {
		t.Run(status, func(t *testing.T) {
			sessionRepo := newFakeSessionRepo()
			payCli := &fakePaymentClient{statusByID: map[string]*commonv1.StatusResponse{
				"prompt-1": {Id: "prompt-1", Status: commonv1.STATUS_FAILED},
			}}
			b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})
			s := processingSession("sess-x", "prompt-1", fixedNow())
			s.Status = status
			sessionRepo.sessions["sess-x"] = s

			updated, err := b.RefreshStatus(context.Background(), s)

			require.NoError(t, err)
			assert.Equal(t, status, updated.Status)
		})
	}
}

func TestRefreshStatus_AmountAndCurrencyChecks(t *testing.T) {
	tests := []struct {
		name       string
		extras     map[string]any
		wantStatus string
	}{
		{name: "no amount reported", extras: map[string]any{}, wantStatus: models.SessionStatusCompleted},
		{name: "mpesa callback amount matches", extras: map[string]any{"Amount": "100", "requested_amount": "100"}, wantStatus: models.SessionStatusCompleted},
		{name: "mpesa callback amount differs", extras: map[string]any{"Amount": "1", "requested_amount": "100"}, wantStatus: models.SessionStatusFailed},
		{name: "flutterwave numeric amount matches", extras: map[string]any{"amount": 100.0, "currency": "KES"}, wantStatus: models.SessionStatusCompleted},
		{name: "flutterwave numeric amount differs", extras: map[string]any{"amount": 99.99, "currency": "KES"}, wantStatus: models.SessionStatusFailed},
		{name: "mtn currency differs", extras: map[string]any{"amount": "100", "currency": "UGX"}, wantStatus: models.SessionStatusFailed},
		{name: "airtel requested amount matches", extras: map[string]any{"requested_amount": "100", "requested_currency": "kes"}, wantStatus: models.SessionStatusCompleted},
		{name: "unparsable amount is not accepted", extras: map[string]any{"amount": "one hundred"}, wantStatus: models.SessionStatusFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := newFakeSessionRepo()
			payCli := &fakePaymentClient{statusByID: map[string]*commonv1.StatusResponse{
				"prompt-1": successWith(t, "prompt-1", tt.extras),
			}}
			b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})
			s := processingSession("sess-amt", "prompt-1", fixedNow())
			sessionRepo.sessions["sess-amt"] = s

			updated, err := b.RefreshStatus(context.Background(), s)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, updated.Status)
			if tt.wantStatus == models.SessionStatusFailed {
				assert.Equal(t, "prompt-1", updated.Metadata["_amount_mismatch_prompt"])
				assert.Empty(t, updated.PaymentID)
			}
		})
	}
}

func TestRefreshStatus_MismatchOnEarlierPromptIsIgnored(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	payCli := &fakePaymentClient{statusByID: map[string]*commonv1.StatusResponse{
		"prompt-1": successWith(t, "prompt-1", map[string]any{"amount": "1"}),
	}}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})
	s := processingSession("sess-m", "prompt-2", fixedNow(), "prompt-1", "prompt-2")
	sessionRepo.sessions["sess-m"] = s

	updated, err := b.RefreshStatus(context.Background(), s)

	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusProcessing, updated.Status, "current prompt is still in process")
}

func TestSweepProcessing_PollsUnexpiredFailedSessions(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	payCli := &fakePaymentClient{statusByID: map[string]*commonv1.StatusResponse{
		"prompt-1": successWith(t, "prompt-1", map[string]any{"Amount": "100"}),
	}}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})

	live := processingSession("failed-live", "prompt-1", fixedNow(), "prompt-1")
	live.Status = models.SessionStatusFailed
	expired := processingSession("failed-expired", "prompt-1", fixedNow(), "prompt-1")
	expired.Status = models.SessionStatusFailed
	expired.ExpiresAt = fixedNow().Add(-time.Minute)
	for _, s := range []*models.CheckoutSession{live, expired} {
		sessionRepo.sessions[s.Ref] = s
	}
	sessionRepo.byStatus[models.SessionStatusFailed] = []*models.CheckoutSession{live, expired}

	require.NoError(t, b.SweepProcessing(context.Background()))

	assert.Equal(t, models.SessionStatusCompleted, sessionRepo.sessions["failed-live"].Status)
	assert.Equal(t, models.SessionStatusFailed, sessionRepo.sessions["failed-expired"].Status)
}
