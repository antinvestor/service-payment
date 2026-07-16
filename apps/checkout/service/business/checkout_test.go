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
	"errors"
	"fmt"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/checkout/config"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/antinvestor/service-payments/apps/checkout/service/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Fakes — embed the interface (nil); only override what the test touches.
// ---------------------------------------------------------------------------

type fakeSessionRepo struct {
	repository.SessionRepository
	sessions    map[string]*models.CheckoutSession
	byStatus    map[string][]*models.CheckoutSession
	createErr   error
	updateErr   error
	getRefErr   error
	lastUpdate  *models.CheckoutSession
	updateCount int
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{
		sessions: make(map[string]*models.CheckoutSession),
		byStatus: make(map[string][]*models.CheckoutSession),
	}
}

func (f *fakeSessionRepo) Create(_ context.Context, s *models.CheckoutSession) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.sessions[s.Ref] = s
	return nil
}

func (f *fakeSessionRepo) Update(
	_ context.Context,
	s *models.CheckoutSession,
	_ ...string,
) (int64, error) {
	if f.updateErr != nil {
		return 0, f.updateErr
	}
	f.sessions[s.Ref] = s
	f.lastUpdate = s
	f.updateCount++
	return 1, nil
}

func (f *fakeSessionRepo) GetByRef(_ context.Context, ref string) (*models.CheckoutSession, error) {
	if f.getRefErr != nil {
		return nil, f.getRefErr
	}
	s, ok := f.sessions[ref]
	if !ok {
		return nil, fmt.Errorf("session %q: %w", ref, gorm.ErrRecordNotFound)
	}
	return s, nil
}

func (f *fakeSessionRepo) GetByOrderRef(_ context.Context, orderRef string) (*models.CheckoutSession, error) {
	for _, s := range f.sessions {
		if s != nil && s.OrderRef == orderRef {
			return s, nil
		}
	}
	return nil, fmt.Errorf("session order_ref %q: %w", orderRef, gorm.ErrRecordNotFound)
}

func (f *fakeSessionRepo) ListByStatus(
	_ context.Context,
	status string,
	limit int,
) ([]*models.CheckoutSession, error) {
	list := f.byStatus[status]
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

// ---------------------------------------------------------------------------

type fakeLinkRepo struct {
	repository.LinkRepository
	links     map[string]*models.CheckoutLink
	createErr error
	getRefErr error
}

func newFakeLinkRepo() *fakeLinkRepo {
	return &fakeLinkRepo{links: make(map[string]*models.CheckoutLink)}
}

func (f *fakeLinkRepo) Create(_ context.Context, l *models.CheckoutLink) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.links[l.Ref] = l
	return nil
}

func (f *fakeLinkRepo) Update(
	_ context.Context,
	l *models.CheckoutLink,
	_ ...string,
) (int64, error) {
	f.links[l.Ref] = l
	return 1, nil
}

func (f *fakeLinkRepo) GetByRef(_ context.Context, ref string) (*models.CheckoutLink, error) {
	if f.getRefErr != nil {
		return nil, f.getRefErr
	}
	l, ok := f.links[ref]
	if !ok {
		return nil, fmt.Errorf("link %q: %w", ref, gorm.ErrRecordNotFound)
	}
	return l, nil
}

// ---------------------------------------------------------------------------

type fakePaymentClient struct {
	paymentv1connect.PaymentServiceClient
	promptResp *connect.Response[paymentv1.InitiatePromptResponse]
	promptErr  error
	statusResp *connect.Response[commonv1.StatusResponse]
	statusErr  error
	lastPrompt *paymentv1.InitiatePromptRequest
}

func (f *fakePaymentClient) InitiatePrompt(
	_ context.Context,
	req *connect.Request[paymentv1.InitiatePromptRequest],
) (*connect.Response[paymentv1.InitiatePromptResponse], error) {
	f.lastPrompt = req.Msg
	if f.promptErr != nil {
		return nil, f.promptErr
	}
	if f.promptResp != nil {
		return f.promptResp, nil
	}
	return connect.NewResponse(&paymentv1.InitiatePromptResponse{
		Data: &commonv1.StatusResponse{Id: "prompt-123"},
	}), nil
}

func (f *fakePaymentClient) Status(
	_ context.Context,
	_ *connect.Request[commonv1.StatusRequest],
) (*connect.Response[commonv1.StatusResponse], error) {
	if f.statusErr != nil {
		return nil, f.statusErr
	}
	if f.statusResp != nil {
		return f.statusResp, nil
	}
	return connect.NewResponse(&commonv1.StatusResponse{Status: commonv1.STATUS_IN_PROCESS}), nil
}

// ---------------------------------------------------------------------------

type fakeProfileClient struct {
	profilev1connect.ProfileServiceClient
	profile   *profilev1.ProfileObject
	getErr    error
	updateReq *profilev1.UpdateRequest
	updateErr error
}

//nolint:revive,staticcheck // Method name matches the generated proto interface (buf.build protobuf naming).
func (f *fakeProfileClient) GetById(
	_ context.Context,
	_ *connect.Request[profilev1.GetByIdRequest],
) (*connect.Response[profilev1.GetByIdResponse], error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return connect.NewResponse(&profilev1.GetByIdResponse{Data: f.profile}), nil
}

func (f *fakeProfileClient) Update(
	_ context.Context,
	req *connect.Request[profilev1.UpdateRequest],
) (*connect.Response[profilev1.UpdateResponse], error) {
	f.updateReq = req.Msg
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return connect.NewResponse(&profilev1.UpdateResponse{}), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func defaultConfig() *config.CheckoutConfig {
	return &config.CheckoutConfig{
		SessionTTLMinutes:      30,
		MaxAttempts:            3,
		AttemptCooldownSeconds: 20,
		MethodsJSON:            `[{"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]}]`,
		PublicBaseURL:          "http://localhost:8080",
	}
}

func defaultRegistry() *business.MethodRegistry {
	reg, err := business.ParseMethodRegistry(
		`[{"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]},{"key":"mtn_momo","name":"MTN MoMo","route":"mtn","prefixes":["256"],"currencies":["UGX"]}]`,
	)
	if err != nil {
		panic(err)
	}
	return reg
}

func fixedNow() time.Time {
	return time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
}

func newBusiness(
	cfg *config.CheckoutConfig,
	reg *business.MethodRegistry,
	sessionRepo repository.SessionRepository,
	linkRepo repository.LinkRepository,
	payCli paymentv1connect.PaymentServiceClient,
	profCli profilev1connect.ProfileServiceClient,
) *business.CheckoutBusiness {
	b := business.NewCheckoutBusiness(cfg, reg, sessionRepo, linkRepo, payCli, profCli, nil)
	return b.WithClock(fixedNow)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// 1. CreateSession persists: 32-char ref, pending, ExpiresAt=now+TTL, metadata copied.
func TestCreateSession_Persists(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)

	ctx := context.Background()
	in := business.CreateSessionInput{
		Name:         "Test Payment",
		Description:  "A test",
		Amount:       "100.50",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		OrderRef:     "ORDER-001",
		Metadata:     map[string]string{"key1": "val1", "key2": "val2"},
		ReturnURL:    "https://example.com/return",
	}

	session, err := b.CreateSession(ctx, in)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Len(t, session.Ref, 32, "ref must be 32 chars")
	assert.Equal(t, models.SessionStatusPending, session.Status)
	assert.Equal(t, "Test Payment", session.Name)
	assert.Equal(t, "ORDER-001", session.OrderRef)
	assert.Equal(t, "KES", session.Currency)
	assert.Equal(t, "100.50", session.Amount)

	expectedExpiry := fixedNow().Add(30 * time.Minute)
	assert.Equal(t, expectedExpiry, session.ExpiresAt)

	require.NotNil(t, session.Metadata)
	assert.Equal(t, "val1", session.Metadata["key1"])
	assert.Equal(t, "val2", session.Metadata["key2"])

	// Verify persisted
	stored, ok := sessionRepo.sessions[session.Ref]
	require.True(t, ok, "session should be stored in repo")
	assert.Equal(t, session.Ref, stored.Ref)
}

// 2. CreateSession with payer: profile fetched; prefill has exact keys.
func TestCreateSession_WithPayer_PrefillKeys(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()

	props, _ := structpb.NewStruct(map[string]any{
		"name": "Alice Wanjiru",
		"checkout": map[string]any{
			"lastProvider":  "mpesa",
			"lastContactId": "contact-abc",
		},
	})
	profile := &profilev1.ProfileObject{
		Id:         "profile-001",
		Properties: props,
		Contacts: []*profilev1.ContactObject{
			{Id: "contact-abc", Type: profilev1.ContactType_MSISDN, Detail: "254712345678"},
		},
	}

	profCli := &fakeProfileClient{profile: profile}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		profCli,
	)

	ctx := context.Background()
	in := business.CreateSessionInput{
		Name:         "Payer Session",
		Amount:       "50.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		Payer: &business.PayerInput{
			ProfileID:   "profile-001",
			DisplayName: "Alice",
			Language:    "en",
			Contacts: []business.PayerContactInput{
				{ContactID: "contact-abc", Msisdn: "254712345678"},
			},
		},
	}

	session, err := b.CreateSession(ctx, in)
	require.NoError(t, err)
	require.NotNil(t, session)

	prefill := session.Prefill
	require.NotNil(t, prefill)
	assert.Equal(t, "Alice", prefill["displayName"], "displayName from caller value")
	assert.Equal(t, "en", prefill["language"])
	assert.Equal(t, "mpesa", prefill["clueProvider"])
	assert.Equal(t, "contact-abc", prefill["clueContactId"])

	contacts, ok := prefill["contacts"].([]any)
	require.True(t, ok, "contacts should be []any")
	require.Len(t, contacts, 1)
	c0, ok := contacts[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "contact-abc", c0["contactId"])
	assert.Equal(t, "254712345678", c0["msisdn"])

	assert.Equal(t, "profile-001", session.PayerProfileID)
}

// 2b. CreateSession with payer — displayName falls back to profile "name" when caller empty.
func TestCreateSession_WithPayer_DisplayNameFallback(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()

	props, _ := structpb.NewStruct(map[string]any{"name": "Bob Kariuki"})
	profCli := &fakeProfileClient{profile: &profilev1.ProfileObject{Properties: props}}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		profCli,
	)

	in := business.CreateSessionInput{
		Name:     "Test",
		Amount:   "10.00",
		Currency: "KES",
		Payer:    &business.PayerInput{ProfileID: "p1"}, // no DisplayName
	}
	session, err := b.CreateSession(context.Background(), in)
	require.NoError(t, err)
	assert.Equal(t, "Bob Kariuki", session.Prefill["displayName"])
}

// 2c. Profile contacts used when caller provides none.
func TestCreateSession_WithPayer_ProfileContactsFallback(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()

	// properties.email must NOT win over attached EMAIL contact.
	props, _ := structpb.NewStruct(map[string]any{"email": "props-only@example.com"})
	profCli := &fakeProfileClient{profile: &profilev1.ProfileObject{
		Properties: props,
		Contacts: []*profilev1.ContactObject{
			{Id: "c1", Type: profilev1.ContactType_MSISDN, Detail: "254700000001"},
			{
				Id:     "c2",
				Type:   profilev1.ContactType_EMAIL,
				Detail: "bob@example.com",
			}, // email for prefill; not listed as MSISDN contact chips
		},
	}}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		profCli,
	)

	in := business.CreateSessionInput{
		Name:     "Test",
		Amount:   "10.00",
		Currency: "KES",
		Payer:    &business.PayerInput{ProfileID: "p1"}, // no Contacts
	}
	session, err := b.CreateSession(context.Background(), in)
	require.NoError(t, err)

	contacts := session.Prefill["contacts"].([]any)
	require.Len(t, contacts, 1, "only MSISDN contacts included")
	c0 := contacts[0].(map[string]any)
	assert.Equal(t, "c1", c0["contactId"])
	assert.Equal(t, "254700000001", c0["msisdn"])
	// Email is resolved from the EMAIL contact, not properties.
	assert.Equal(t, "bob@example.com", session.Prefill["email"])
}

// 2d. No EMAIL contact → prefill email stays empty (do not use properties.email).
func TestCreateSession_WithPayer_EmailFromContactNotProperties(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()

	props, _ := structpb.NewStruct(map[string]any{"email": "props-only@example.com", "name": "Bob"})
	profCli := &fakeProfileClient{profile: &profilev1.ProfileObject{
		Properties: props,
		Contacts: []*profilev1.ContactObject{
			{Id: "c1", Type: profilev1.ContactType_MSISDN, Detail: "254700000001"},
		},
	}}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		profCli,
	)

	session, err := b.CreateSession(context.Background(), business.CreateSessionInput{
		Name: "Test", Amount: "10.00", Currency: "KES",
		Payer: &business.PayerInput{ProfileID: "p1"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Bob", session.Prefill["displayName"])
	assert.Equal(t, "", session.Prefill["email"], "email must come from EMAIL contact, not properties")
}

// 3. CreateSession fixed without valid amount → error.
func TestCreateSession_Validation(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)
	ctx := context.Background()

	tests := []struct {
		name         string
		in           business.CreateSessionInput
		wantSentinel error  // use require.ErrorIs when set
		wantContains string // use assert.ErrorContains when set
	}{
		{
			name:         "missing name",
			in:           business.CreateSessionInput{Amount: "10.00", Currency: "KES"},
			wantContains: "name",
		},
		{
			name: "fixed without amount",
			in: business.CreateSessionInput{
				Name: "Test", Currency: "KES",
				AmountOption: models.AmountOptionFixed,
			},
			wantContains: "amount",
		},
		{
			name: "fixed with invalid amount",
			in: business.CreateSessionInput{
				Name: "Test", Currency: "KES",
				AmountOption: models.AmountOptionFixed, Amount: "abc",
			},
			wantContains: "amount",
		},
		{
			name: "missing currency",
			in: business.CreateSessionInput{
				Name: "Test", Amount: "10.00", AmountOption: models.AmountOptionFixed,
			},
			wantContains: "currency",
		},
		{
			name: "unknown method restriction",
			in: business.CreateSessionInput{
				Name: "Test", Amount: "10.00", Currency: "KES",
				AmountOption: models.AmountOptionFixed,
				Methods:      []string{"stripe"},
			},
			wantSentinel: business.ErrUnknownMethod,
		},
		{
			name: "javascript: return_url rejected",
			in: business.CreateSessionInput{
				Name: "Test", Amount: "10.00", Currency: "KES",
				AmountOption: models.AmountOptionFixed,
				ReturnURL:    "javascript:alert(1)",
			},
			wantContains: "return_url",
		},
		{
			name: "protocol-relative return_url rejected",
			in: business.CreateSessionInput{
				Name: "Test", Amount: "10.00", Currency: "KES",
				AmountOption: models.AmountOptionFixed,
				ReturnURL:    "//evil.com/steal",
			},
			wantContains: "return_url",
		},
		{
			name: "https return_url accepted",
			in: business.CreateSessionInput{
				Name: "Test", Amount: "10.00", Currency: "KES",
				AmountOption: models.AmountOptionFixed,
				ReturnURL:    "https://example.com/return",
			},
			// no error expected — leave wantContains and wantSentinel empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := b.CreateSession(ctx, tt.in)
			if tt.wantSentinel == nil && tt.wantContains == "" {
				// test cases that expect success
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tt.wantSentinel != nil {
				require.ErrorIs(t, err, tt.wantSentinel)
			}
			if tt.wantContains != "" {
				assert.ErrorContains(t, err, tt.wantContains)
			}
		})
	}
}

// 3b-extra. CreateSession with empty ReturnURL is allowed.
func TestCreateSession_EmptyReturnURL_Allowed(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)

	in := business.CreateSessionInput{
		Name:         "Test",
		Amount:       "10.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		ReturnURL:    "", // empty must be allowed
	}
	_, err := b.CreateSession(context.Background(), in)
	require.NoError(t, err, "empty ReturnURL must not fail validation")
}

// 3b-extra. CreateLink return_url validation: javascript rejected, //evil rejected, https accepted, empty allowed.
func TestCreateLink_ReturnURL_Validation(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name      string
		returnURL string
		wantErr   bool
	}{
		{"javascript rejected", "javascript:alert(1)", true},
		{"protocol-relative rejected", "//evil.com", true},
		{"https accepted", "https://example.com/done", false},
		{"http accepted", "http://localhost/done", false},
		{"empty allowed", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionRepo := newFakeSessionRepo()
			linkRepo := newFakeLinkRepo()
			b := newBusiness(
				defaultConfig(),
				defaultRegistry(),
				sessionRepo,
				linkRepo,
				&fakePaymentClient{},
				&fakeProfileClient{},
			)

			in := business.CreateLinkInput{
				Name:         "Link",
				Currency:     "KES",
				AmountOption: models.AmountOptionFixed,
				Amount:       "10.00",
				ReturnURL:    tc.returnURL,
			}
			_, err := b.CreateLink(ctx, in)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "return_url")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// 3b. CreateSession with empty AmountOption defaults to fixed and persists it.
func TestCreateSession_EmptyAmountOption_DefaultsToFixed(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)

	in := business.CreateSessionInput{
		Name:         "Default AmountOption",
		Amount:       "25.00",
		Currency:     "KES",
		AmountOption: "", // deliberately empty — should be normalised to fixed
	}
	session, err := b.CreateSession(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, models.AmountOptionFixed, session.AmountOption,
		"empty AmountOption must be stored as fixed")

	// Verify the persisted copy also carries the normalised value
	stored := sessionRepo.sessions[session.Ref]
	require.NotNil(t, stored)
	assert.Equal(t, models.AmountOptionFixed, stored.AmountOption)
}

// 3c. CreateSession with profile error still proceeds (tolerate).
func TestCreateSession_ProfileErrorTolerated(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	profCli := &fakeProfileClient{getErr: errors.New("profile service down")}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		profCli,
	)

	in := business.CreateSessionInput{
		Name:     "Test",
		Amount:   "10.00",
		Currency: "KES",
		Payer:    &business.PayerInput{ProfileID: "p1", DisplayName: "Eve"},
	}
	session, err := b.CreateSession(context.Background(), in)
	require.NoError(t, err, "profile errors should be tolerated")
	assert.Equal(t, "Eve", session.Prefill["displayName"])
}

// 4. GetSessionByRef expiry flip persists expired.
func TestGetSessionByRef_ExpiryFlip(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)

	past := fixedNow().Add(-1 * time.Minute)
	s := &models.CheckoutSession{
		Ref:       "abc123",
		Status:    models.SessionStatusPending,
		ExpiresAt: past,
	}
	sessionRepo.sessions["abc123"] = s

	ctx := context.Background()
	got, err := b.GetSessionByRef(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusExpired, got.Status)
	// Repo update was called
	require.NotNil(t, sessionRepo.lastUpdate)
	assert.Equal(t, models.SessionStatusExpired, sessionRepo.lastUpdate.Status)
}

// Past-TTL processing session that was already paid must complete, not expire.
func TestGetSessionByRef_ProcessingPastTTL_Paid_Completes(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Id:     "pay-1",
			Status: commonv1.STATUS_SUCCESSFUL,
		}),
	}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, linkRepo, payCli, &fakeProfileClient{}).
		WithSynchronousClues()

	past := fixedNow().Add(-1 * time.Minute)
	s := &models.CheckoutSession{
		Ref:       "paid-stuck",
		Status:    models.SessionStatusProcessing,
		PromptID:  "prompt-paid",
		ExpiresAt: past,
		Currency:  "KES",
	}
	sessionRepo.sessions["paid-stuck"] = s

	got, err := b.GetSessionByRef(context.Background(), "paid-stuck")
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusCompleted, got.Status)
	assert.Equal(t, "pay-1", got.PaymentID)
}

// Expired session that was paid while stuck is recovered to completed.
func TestGetSessionByRef_Expired_Paid_Recovers(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Id:     "pay-2",
			Status: commonv1.STATUS_SUCCESSFUL,
		}),
	}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, linkRepo, payCli, &fakeProfileClient{}).
		WithSynchronousClues()

	s := &models.CheckoutSession{
		Ref:       "expired-paid",
		Status:    models.SessionStatusExpired,
		PromptID:  "prompt-expired-paid",
		ExpiresAt: fixedNow().Add(-10 * time.Minute),
		Currency:  "KES",
	}
	sessionRepo.sessions["expired-paid"] = s

	got, err := b.GetSessionByRef(context.Background(), "expired-paid")
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusCompleted, got.Status)
}

// 4b. GetSessionByRef does not expire processing sessions within TTL.
func TestGetSessionByRef_NoExpiry_WhenNotPast(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)

	future := fixedNow().Add(10 * time.Minute)
	s := &models.CheckoutSession{
		Ref:       "ref1",
		Status:    models.SessionStatusPending,
		ExpiresAt: future,
	}
	sessionRepo.sessions["ref1"] = s

	got, err := b.GetSessionByRef(context.Background(), "ref1")
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusPending, got.Status, "should NOT be expired")
	assert.Nil(t, sessionRepo.lastUpdate, "no update should be issued")
}

// 5. CreateLink + SpawnSession copies fields, LinkID set, fresh 32-char ref.
func TestCreateLink_And_SpawnSession(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)
	ctx := context.Background()

	expiry := fixedNow().Add(24 * time.Hour)
	in := business.CreateLinkInput{
		Name:         "Product Link",
		Description:  "Buy now",
		Amount:       "200.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		OrderRef:     "ORDER-LINK",
		ReturnURL:    "https://example.com/return",
		Metadata:     map[string]string{"product": "widget"},
		ExpiresAt:    &expiry,
	}

	link, err := b.CreateLink(ctx, in)
	require.NoError(t, err)
	require.NotNil(t, link)
	assert.Len(t, link.Ref, 12, "link ref must be 12 chars")
	assert.True(t, link.Active)
	assert.Equal(t, "Product Link", link.Name)
	assert.Equal(t, &expiry, link.ExpiresAt)

	session, err := b.SpawnSession(ctx, link.Ref)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Len(t, session.Ref, 32, "session ref must be 32 chars")
	assert.Equal(t, link.GetID(), session.LinkID)
	assert.Equal(t, "Product Link", session.Name)
	assert.Equal(t, "200.00", session.Amount)
	assert.Equal(t, "KES", session.Currency)
	assert.Equal(t, "ORDER-LINK", session.OrderRef)
	assert.Equal(t, models.SessionStatusPending, session.Status)
}

// 5b. Inactive link → ErrLinkUnusable.
func TestSpawnSession_InactiveLink(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)
	ctx := context.Background()

	link := &models.CheckoutLink{Ref: "link-inactive", Active: false, Currency: "KES"}
	linkRepo.links["link-inactive"] = link

	_, err := b.SpawnSession(ctx, "link-inactive")
	require.ErrorIs(t, err, business.ErrLinkUnusable)
}

// 5c. Expired link → ErrLinkUnusable.
func TestSpawnSession_ExpiredLink(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		&fakeProfileClient{},
	)
	ctx := context.Background()

	past := fixedNow().Add(-1 * time.Minute)
	link := &models.CheckoutLink{Ref: "link-exp", Active: true, ExpiresAt: &past, Currency: "KES"}
	linkRepo.links["link-exp"] = link

	_, err := b.SpawnSession(ctx, "link-exp")
	require.ErrorIs(t, err, business.ErrLinkUnusable)
}

// 6. Pay happy path.
func TestPay_HappyPath(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:          "sess-pay",
		Status:       models.SessionStatusPending,
		ExpiresAt:    future,
		Amount:       "150.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		OrderRef:     "ORD-42",
		Metadata:     map[string]any{"invoice": "INV-99"},
	}
	sessionRepo.sessions["sess-pay"] = s

	in := business.PayInput{
		MethodKey:   "mpesa",
		PhoneNumber: "254712345678",
	}
	updated, err := b.Pay(ctx, "sess-pay", in)
	require.NoError(t, err)
	require.NotNil(t, updated)

	// Session updated to processing
	assert.Equal(t, models.SessionStatusProcessing, updated.Status)
	assert.Equal(t, 1, updated.Attempts)
	assert.Equal(t, "prompt-123", updated.PromptID)
	assert.NotNil(t, updated.LastAttemptAt)

	// Verify prompt request
	require.NotNil(t, payCli.lastPrompt)
	assert.Equal(t, "mpesa", payCli.lastPrompt.GetRoute())
	assert.Equal(t, "254712345678", payCli.lastPrompt.GetSource().GetContactId())
	assert.Equal(t, int64(150), payCli.lastPrompt.GetAmount().GetUnits())
	assert.Equal(t, "KES", payCli.lastPrompt.GetAmount().GetCurrencyCode())

	// Check Extra fields
	extraMap := payCli.lastPrompt.GetExtra().AsMap()
	assert.Equal(t, "sess-pay", extraMap["session_ref"])
	assert.Equal(t, "ORD-42", extraMap["order_ref"])
	assert.Equal(t, "mpesa", extraMap["provider"])
	assert.Equal(t, "INV-99", extraMap["meta_invoice"])
}

// 6b. Pay with recognized payer (ProfileID + ContactID).
func TestPay_RecognizedPayer(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:            "sess-recog",
		Status:         models.SessionStatusPending,
		ExpiresAt:      future,
		Amount:         "75.00",
		Currency:       "KES",
		AmountOption:   models.AmountOptionFixed,
		PayerProfileID: "profile-xyz",
		Prefill: map[string]any{
			"contacts": []any{
				map[string]any{"contactId": "cid-1", "msisdn": "254711111111"},
			},
		},
	}
	sessionRepo.sessions["sess-recog"] = s

	in := business.PayInput{
		MethodKey: "mpesa",
		ContactID: "cid-1",
	}
	updated, err := b.Pay(ctx, "sess-recog", in)
	require.NoError(t, err)
	require.NotNil(t, updated)

	require.NotNil(t, payCli.lastPrompt)
	assert.Equal(t, "254711111111", payCli.lastPrompt.GetSource().GetContactId())
	assert.Equal(t, "profile-xyz", payCli.lastPrompt.GetSource().GetProfileId())
}

// 7. Pay guards.
func TestPay_Guards(t *testing.T) {
	ctx := context.Background()
	future := fixedNow().Add(20 * time.Minute)

	makeSession := func(status string, attempts int, lastAttempt *time.Time, opts ...func(*models.CheckoutSession)) *models.CheckoutSession {
		s := &models.CheckoutSession{
			Ref:           "sess",
			Status:        status,
			ExpiresAt:     future,
			Amount:        "10.00",
			Currency:      "KES",
			AmountOption:  models.AmountOptionFixed,
			Attempts:      attempts,
			LastAttemptAt: lastAttempt,
		}
		for _, o := range opts {
			o(s)
		}
		return s
	}

	justNow := fixedNow()

	tests := []struct {
		name    string
		session *models.CheckoutSession
		in      business.PayInput
		wantErr error
	}{
		{
			name:    "completed → ErrSessionGone",
			session: makeSession(models.SessionStatusCompleted, 0, nil),
			in:      business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"},
			wantErr: business.ErrSessionGone,
		},
		{
			name:    "expired → ErrSessionGone",
			session: makeSession(models.SessionStatusExpired, 0, nil),
			in:      business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"},
			wantErr: business.ErrSessionGone,
		},
		{
			name:    "max attempts reached → ErrTooManyAttempts",
			session: makeSession(models.SessionStatusPending, 3, nil),
			in:      business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"},
			wantErr: business.ErrTooManyAttempts,
		},
		{
			name:    "within cooldown → ErrCooldown",
			session: makeSession(models.SessionStatusPending, 1, &justNow),
			in:      business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"},
			wantErr: business.ErrCooldown,
		},
		{
			name:    "unknown method → ErrUnknownMethod",
			session: makeSession(models.SessionStatusPending, 0, nil),
			in:      business.PayInput{MethodKey: "stripe", PhoneNumber: "254700000001"},
			wantErr: business.ErrUnknownMethod,
		},
		{
			name: "method not in session restriction → ErrUnknownMethod",
			session: makeSession(
				models.SessionStatusPending,
				0,
				nil,
				func(s *models.CheckoutSession) {
					s.Methods = map[string]any{"keys": []any{"mtn_momo"}}
				},
			),
			in:      business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"},
			wantErr: business.ErrUnknownMethod,
		},
		{
			name: "variable without amount → ErrAmountRequired",
			session: makeSession(
				models.SessionStatusPending,
				0,
				nil,
				func(s *models.CheckoutSession) {
					s.AmountOption = models.AmountOptionVariable
					s.Amount = ""
				},
			),
			in:      business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"},
			wantErr: business.ErrAmountRequired,
		},
		{
			name:    "guest without phone → ErrContactRequired",
			session: makeSession(models.SessionStatusPending, 0, nil),
			in:      business.PayInput{MethodKey: "mpesa"},
			wantErr: business.ErrContactRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := newFakeSessionRepo()
			linkRepo := newFakeLinkRepo()
			payCli := &fakePaymentClient{}
			b := newBusiness(
				defaultConfig(),
				defaultRegistry(),
				sessionRepo,
				linkRepo,
				payCli,
				&fakeProfileClient{},
			)

			sessionRepo.sessions["sess"] = tt.session

			_, err := b.Pay(ctx, "sess", tt.in)
			require.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

// 7b. After cooldown, retry is allowed.
func TestPay_AfterCooldown_Allowed(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	past := fixedNow().Add(-25 * time.Second) // beyond 20s cooldown
	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:           "sess2",
		Status:        models.SessionStatusPending,
		ExpiresAt:     future,
		Amount:        "10.00",
		Currency:      "KES",
		AmountOption:  models.AmountOptionFixed,
		Attempts:      1,
		LastAttemptAt: &past,
	}
	sessionRepo.sessions["sess2"] = s

	_, err := b.Pay(
		ctx,
		"sess2",
		business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"},
	)
	require.NoError(t, err)
}

// 8a. RefreshStatus SUCCESSFUL → completed, profile Update called (when _contact_id set).
func TestRefreshStatus_Successful_WithClue(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Id:         "payment-external-456",
			ExternalId: "payment-external-456",
			Status:     commonv1.STATUS_SUCCESSFUL,
		}),
	}
	profCli := &fakeProfileClient{}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, linkRepo, payCli, profCli).WithSynchronousClues()
	ctx := context.Background()

	promptID := "prompt-123"
	s := &models.CheckoutSession{
		Ref:            "sess-refresh",
		Status:         models.SessionStatusProcessing,
		PromptID:       promptID,
		Currency:       "KES",
		PayerProfileID: "profile-001",
		Metadata: map[string]any{
			"_method":     "mpesa",
			"_contact_id": "contact-xyz",
		},
	}
	sessionRepo.sessions["sess-refresh"] = s

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusCompleted, updated.Status)
	assert.NotEmpty(t, updated.PaymentID)

	// Profile update was called with clue data
	require.NotNil(t, profCli.updateReq)
	assert.Equal(t, "profile-001", profCli.updateReq.GetId())
	props := profCli.updateReq.GetProperties().AsMap()
	checkout, ok := props["checkout"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "mpesa", checkout["lastProvider"])
	assert.Equal(t, "contact-xyz", checkout["lastContactId"])
	assert.Equal(t, "KES", checkout["lastCurrency"])
}

// 8b. RefreshStatus SUCCESSFUL without _contact_id still writes method clues
// (Link-style last-used method, including card/redirect without MSISDN contact).
func TestRefreshStatus_Successful_NoContactId_WritesMethodClue(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Status: commonv1.STATUS_SUCCESSFUL,
		}),
	}
	profCli := &fakeProfileClient{}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, linkRepo, payCli, profCli).WithSynchronousClues()
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:            "sess-noclue",
		Status:         models.SessionStatusProcessing,
		PromptID:       "p-1",
		Currency:       "KES",
		PayerProfileID: "profile-001",
		Metadata:       map[string]any{"_method": "mpesa"}, // no _contact_id
	}
	sessionRepo.sessions["sess-noclue"] = s

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusCompleted, updated.Status)
	require.NotNil(t, profCli.updateReq, "profile Update must store last method without contact")
	props := profCli.updateReq.GetProperties().AsMap()
	checkout, ok := props["checkout"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "mpesa", checkout["lastProvider"])
	assert.Equal(t, "mobile_money", checkout["lastMethod"])
	assert.Equal(t, "KES", checkout["lastCurrency"])
}

// 8c. RefreshStatus FAILED → failed.
func TestRefreshStatus_Failed(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Status: commonv1.STATUS_FAILED,
		}),
	}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:      "sess-fail",
		Status:   models.SessionStatusProcessing,
		PromptID: "p-2",
	}
	sessionRepo.sessions["sess-fail"] = s

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusFailed, updated.Status)
}

// 8d. RefreshStatus transport error → still processing, nil error.
func TestRefreshStatus_TransportError(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{statusErr: errors.New("connection refused")}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:      "sess-transport",
		Status:   models.SessionStatusProcessing,
		PromptID: "p-3",
	}
	sessionRepo.sessions["sess-transport"] = s

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusProcessing, updated.Status)
}

// 8e. RefreshStatus on non-processing session returns unchanged.
func TestRefreshStatus_NotProcessing_NoOp(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:    "sess-noprocess",
		Status: models.SessionStatusPending,
	}
	sessionRepo.sessions["sess-noprocess"] = s

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusPending, updated.Status)
	assert.Nil(t, payCli.lastPrompt, "InitiatePrompt must not be called")
	// No repository Update should be issued for a non-processing session
	assert.Equal(
		t,
		0,
		sessionRepo.updateCount,
		"session repo Update must not be called on no-op refresh",
	)
}

// ---------------------------------------------------------------------------
// Finding 1: Empty msisdn guard
// ---------------------------------------------------------------------------

// 1a. Recognised payer with unknown ContactID and no PhoneNumber → validation error, no prompt.
func TestPay_RecognizedPayer_UnknownContactID_NoPhone_Error(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:            "sess-recog-nomsisdn",
		Status:         models.SessionStatusPending,
		ExpiresAt:      future,
		Amount:         "50.00",
		Currency:       "KES",
		AmountOption:   models.AmountOptionFixed,
		PayerProfileID: "profile-xyz",
		Prefill: map[string]any{
			"contacts": []any{
				map[string]any{"contactId": "cid-known", "msisdn": "254711111111"},
			},
		},
	}
	sessionRepo.sessions["sess-recog-nomsisdn"] = s

	// ContactID that is NOT in prefill, and no PhoneNumber provided
	in := business.PayInput{
		MethodKey: "mpesa",
		ContactID: "cid-unknown", // not in prefill
		// PhoneNumber: ""         // no fallback
	}
	_, err := b.Pay(ctx, "sess-recog-nomsisdn", in)
	require.Error(t, err)
	require.ErrorIs(t, err, business.ErrContactRequired)
	// No prompt should have been sent
	assert.Nil(t, payCli.lastPrompt, "InitiatePrompt must not be called when msisdn is empty")
}

// ---------------------------------------------------------------------------
// Finding 2: Attempt counting on failed prompts
// ---------------------------------------------------------------------------

// 2a. Prompt error → Attempts incremented, status still pending, error returned.
func TestPay_PromptError_AttemptsIncremented(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		promptErr: errors.New("provider rejected"),
	}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:          "sess-prompt-fail",
		Status:       models.SessionStatusPending,
		ExpiresAt:    future,
		Amount:       "100.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
	}
	sessionRepo.sessions["sess-prompt-fail"] = s

	_, err := b.Pay(ctx, "sess-prompt-fail", business.PayInput{
		MethodKey:   "mpesa",
		PhoneNumber: "254712345678",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initiate prompt")

	// Attempt must be recorded even though prompt failed
	stored := sessionRepo.sessions["sess-prompt-fail"]
	assert.Equal(t, 1, stored.Attempts, "attempt must be incremented even on prompt failure")
	assert.NotNil(t, stored.LastAttemptAt, "LastAttemptAt must be set even on prompt failure")
	assert.Equal(
		t,
		models.SessionStatusPending,
		stored.Status,
		"status must remain pending after prompt failure",
	)
}

// ---------------------------------------------------------------------------
// Finding 3: VARIABLE happy path + invalid variable amount
// ---------------------------------------------------------------------------

// 3a. VARIABLE session with valid in.Amount stores amount and prompt Money is correct.
func TestPay_Variable_ValidAmount_StoresAndPrompts(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:          "sess-variable",
		Status:       models.SessionStatusPending,
		ExpiresAt:    future,
		Amount:       "",
		Currency:     "KES",
		AmountOption: models.AmountOptionVariable,
	}
	sessionRepo.sessions["sess-variable"] = s

	updated, err := b.Pay(ctx, "sess-variable", business.PayInput{
		MethodKey:   "mpesa",
		PhoneNumber: "254712345678",
		Amount:      "75.50",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	// Session amount stored as supplied string
	assert.Equal(t, "75.50", updated.Amount)

	// Prompt Money: units=75, nanos=500_000_000
	require.NotNil(t, payCli.lastPrompt)
	assert.Equal(t, int64(75), payCli.lastPrompt.GetAmount().GetUnits())
	assert.Equal(t, int32(500_000_000), payCli.lastPrompt.GetAmount().GetNanos())
	assert.Equal(t, "KES", payCli.lastPrompt.GetAmount().GetCurrencyCode())
}

// 3b. FIXED session ignores in.Amount; prompt uses session's own amount.
func TestPay_Fixed_IgnoresInAmount(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:          "sess-fixed-ignore",
		Status:       models.SessionStatusPending,
		ExpiresAt:    future,
		Amount:       "200.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
	}
	sessionRepo.sessions["sess-fixed-ignore"] = s

	updated, err := b.Pay(ctx, "sess-fixed-ignore", business.PayInput{
		MethodKey:   "mpesa",
		PhoneNumber: "254712345678",
		Amount:      "999.99", // should be ignored for FIXED sessions
	})
	require.NoError(t, err)

	// Session amount must stay as original
	assert.Equal(t, "200.00", updated.Amount)

	// Prompt must use session's 200.00
	require.NotNil(t, payCli.lastPrompt)
	assert.Equal(t, int64(200), payCli.lastPrompt.GetAmount().GetUnits())
	assert.Equal(t, int32(0), payCli.lastPrompt.GetAmount().GetNanos())
}

// 3c. VARIABLE session with invalid non-empty amount (too many decimals) → error mentioning amount.
func TestPay_Variable_InvalidAmount_Error(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:          "sess-variable-bad",
		Status:       models.SessionStatusPending,
		ExpiresAt:    future,
		Amount:       "",
		Currency:     "KES",
		AmountOption: models.AmountOptionVariable,
	}
	sessionRepo.sessions["sess-variable-bad"] = s

	_, err := b.Pay(ctx, "sess-variable-bad", business.PayInput{
		MethodKey:   "mpesa",
		PhoneNumber: "254712345678",
		Amount:      "12.345", // 3 decimal places → invalid
	})
	require.Error(t, err)
	require.ErrorIs(t, err, business.ErrAmountRequired)
	assert.Nil(t, payCli.lastPrompt, "no prompt should be sent for invalid amount")
}

// ---------------------------------------------------------------------------
// Finding 4: Clue write error swallowed on SUCCESSFUL
// ---------------------------------------------------------------------------

// 4a. RefreshStatus SUCCESSFUL + profile Update error → still returns completed, nil error.
func TestRefreshStatus_Successful_ClueWriteError_Swallowed(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Id:     "payment-ext-789",
			Status: commonv1.STATUS_SUCCESSFUL,
		}),
	}
	profCli := &fakeProfileClient{
		updateErr: errors.New("profile service unavailable"),
	}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, linkRepo, payCli, profCli).WithSynchronousClues()
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:            "sess-clue-err",
		Status:         models.SessionStatusProcessing,
		PromptID:       "prompt-clue",
		Currency:       "KES",
		PayerProfileID: "profile-002",
		Metadata: map[string]any{
			"_method":     "mpesa",
			"_contact_id": "contact-clue",
		},
	}
	sessionRepo.sessions["sess-clue-err"] = s

	updated, err := b.RefreshStatus(ctx, s)
	// Error from writeClues must be swallowed — it's best-effort
	require.NoError(t, err, "clue write error must not propagate")
	assert.Equal(t, models.SessionStatusCompleted, updated.Status)
}

// ---------------------------------------------------------------------------
// Nit: exact prefill key count
// ---------------------------------------------------------------------------

// Prefill includes display, language, method/contact clues, country, contacts.
func TestCreateSession_WithPayer_PrefillExactKeys(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	profCli := &fakeProfileClient{profile: &profilev1.ProfileObject{}}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		&fakePaymentClient{},
		profCli,
	)

	in := business.CreateSessionInput{
		Name:     "Key count test",
		Amount:   "10.00",
		Currency: "KES",
		Payer:    &business.PayerInput{ProfileID: "p1", DisplayName: "Tester", Language: "sw"},
	}
	session, err := b.CreateSession(context.Background(), in)
	require.NoError(t, err)

	prefill := session.Prefill
	require.NotNil(t, prefill)
	// Multi-contact prefer: email + phone prefer ids, phone default, contacts list.
	for _, key := range []string{
		"displayName", "language", "clueProvider", "clueMethod", "clueContactId",
		"clueEmailContactId", "cluePhoneContactId", "country", "email", "emailContactId",
		"phone", "contacts",
	} {
		_, ok := prefill[key]
		assert.True(t, ok, "missing prefill key %s", key)
	}
}

// 9. SweepProcessing expires overdue pending and refreshes processing.
func TestSweepProcessing(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Status: commonv1.STATUS_SUCCESSFUL,
		}),
	}
	b := newBusiness(
		defaultConfig(),
		defaultRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	past := fixedNow().Add(-1 * time.Hour)
	future := fixedNow().Add(1 * time.Hour)

	// Overdue pending session
	overdue := &models.CheckoutSession{
		Ref:       "overdue-pending",
		Status:    models.SessionStatusPending,
		ExpiresAt: past,
	}
	sessionRepo.sessions["overdue-pending"] = overdue
	sessionRepo.byStatus[models.SessionStatusPending] = []*models.CheckoutSession{overdue}

	// Processing session
	processing := &models.CheckoutSession{
		Ref:       "in-process",
		Status:    models.SessionStatusProcessing,
		PromptID:  "prompt-sweep",
		ExpiresAt: future,
	}
	sessionRepo.sessions["in-process"] = processing
	sessionRepo.byStatus[models.SessionStatusProcessing] = []*models.CheckoutSession{processing}

	err := b.SweepProcessing(ctx)
	require.NoError(t, err)

	// Overdue pending should be expired
	assert.Equal(t, models.SessionStatusExpired, sessionRepo.sessions["overdue-pending"].Status)

	// Processing should be refreshed (completed in this test)
	assert.Equal(t, models.SessionStatusCompleted, sessionRepo.sessions["in-process"].Status)
}

// 9. RefreshStatus with nil workMan (async path) must NOT panic and must still
// complete: when workMan is nil the clue write-back falls back to synchronous
// execution without panicking or returning an error.
func TestRefreshStatus_NilWorkMan_AsyncFallback_NoPanic(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Id:     "payment-nil-workman",
			Status: commonv1.STATUS_SUCCESSFUL,
		}),
	}
	profCli := &fakeProfileClient{}

	// Construct with nil workMan and cluesSync=false (the default) to exercise
	// the nil-workMan synchronous fallback path without WithSynchronousClues.
	b := business.NewCheckoutBusiness(
		defaultConfig(), defaultRegistry(),
		sessionRepo, linkRepo,
		payCli, profCli,
		nil, // nil workMan → synchronous fallback, must not panic
	).WithClock(fixedNow)
	// cluesSync is false — we are deliberately NOT calling WithSynchronousClues.

	s := &models.CheckoutSession{
		Ref:            "sess-nil-workman",
		Status:         models.SessionStatusProcessing,
		PromptID:       "prompt-nwm",
		Currency:       "KES",
		PayerProfileID: "profile-nwm",
		Metadata: map[string]any{
			"_method":     "mpesa",
			"_contact_id": "contact-nwm",
		},
	}
	sessionRepo.sessions["sess-nil-workman"] = s

	ctx := context.Background()
	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	// Session must have been completed — business logic ran fully.
	assert.Equal(t, models.SessionStatusCompleted, updated.Status)
	// Clue write-back executed synchronously via fallback — profile was updated.
	require.NotNil(t, profCli.updateReq)
	assert.Equal(t, "profile-nwm", profCli.updateReq.GetId())
}

// ---------------------------------------------------------------------------
// Redirect method tests
// ---------------------------------------------------------------------------

// redirectRegistry returns a registry that includes the card redirect method.
func redirectRegistry() *business.MethodRegistry {
	reg, err := business.ParseMethodRegistry(
		`[{"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]},` +
			`{"key":"card","name":"Card","route":"polar","prefixes":[],"currencies":[],"redirect":true}]`,
	)
	if err != nil {
		panic(err)
	}
	return reg
}

func TestPay_RedirectMethod_NoPhoneRequired(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(
		defaultConfig(),
		redirectRegistry(),
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	ctx := context.Background()

	future := fixedNow().Add(20 * time.Minute)
	s := &models.CheckoutSession{
		Ref:          "sess-card",
		Status:       models.SessionStatusPending,
		ExpiresAt:    future,
		Amount:       "100.00",
		Currency:     "USD",
		AmountOption: models.AmountOptionFixed,
		OrderRef:     "ORD-CARD-01",
	}
	sessionRepo.sessions["sess-card"] = s

	in := business.PayInput{MethodKey: "card"}
	updated, err := b.Pay(ctx, "sess-card", in)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, models.SessionStatusProcessing, updated.Status)
	assert.Equal(t, 1, updated.Attempts)

	require.NotNil(t, payCli.lastPrompt)
	assert.Equal(t, "polar", payCli.lastPrompt.GetRoute())
	assert.Empty(t, payCli.lastPrompt.GetSource().GetContactId())
}

func TestPay_RedirectMethod_Guest_Succeeds(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(defaultConfig(), redirectRegistry(), sessionRepo, linkRepo, payCli, &fakeProfileClient{})
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:          "sess-card-guest",
		Status:       models.SessionStatusPending,
		ExpiresAt:    fixedNow().Add(20 * time.Minute),
		Amount:       "50.00",
		Currency:     "USD",
		AmountOption: models.AmountOptionFixed,
	}
	sessionRepo.sessions["sess-card-guest"] = s

	updated, err := b.Pay(ctx, "sess-card-guest", business.PayInput{MethodKey: "card"})
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusProcessing, updated.Status)
	require.NotNil(t, payCli.lastPrompt)
	assert.Empty(t, payCli.lastPrompt.GetSource().GetContactId())
}

func TestPay_NonRedirectMethod_NoPhone_ErrContactRequired(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(defaultConfig(), redirectRegistry(), sessionRepo, linkRepo, payCli, &fakeProfileClient{})
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:          "sess-mpesa-nophone",
		Status:       models.SessionStatusPending,
		ExpiresAt:    fixedNow().Add(20 * time.Minute),
		Amount:       "10.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
	}
	sessionRepo.sessions["sess-mpesa-nophone"] = s

	_, err := b.Pay(ctx, "sess-mpesa-nophone", business.PayInput{MethodKey: "mpesa"})
	require.Error(t, err)
	require.ErrorIs(t, err, business.ErrContactRequired)
}

func TestRefreshStatus_CapturesCheckoutURL(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()

	extras, err := structpb.NewStruct(map[string]any{
		"checkout_url": "https://polar.sh/checkout/abc123",
	})
	require.NoError(t, err)

	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Status: commonv1.STATUS_IN_PROCESS,
			Extras: extras,
		}),
	}
	b := newBusiness(defaultConfig(), redirectRegistry(), sessionRepo, linkRepo, payCli, &fakeProfileClient{})
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:      "sess-redirect-url",
		Status:   models.SessionStatusProcessing,
		PromptID: "prompt-polar",
		Currency: "USD",
	}
	sessionRepo.sessions["sess-redirect-url"] = s

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusProcessing, updated.Status)
	require.NotNil(t, updated.Metadata)
	assert.Equal(t, "https://polar.sh/checkout/abc123", updated.Metadata["_redirect_url"])
}

func TestIsExternalAuthRedirect(t *testing.T) {
	t.Parallel()
	base := "https://pay.stawi.org"
	assert.True(t, business.IsExternalAuthRedirect("https://acs.bank.example/3ds", base))
	assert.False(t, business.IsExternalAuthRedirect("https://pay.stawi.org/c/abc", base))
	assert.False(t, business.IsExternalAuthRedirect("https://pay.stawi.org/c/abc/confirm", base))
	assert.False(t, business.IsExternalAuthRedirect("javascript:alert(1)", base))
	assert.False(t, business.IsExternalAuthRedirect("", base))
	// No public base configured → any safe URL is treated as external.
	assert.True(t, business.IsExternalAuthRedirect("https://acs.bank.example/3ds", ""))
}

func TestRefreshStatus_OwnCheckoutURL_NotCapturedAsRedirect(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()

	extras, err := structpb.NewStruct(map[string]any{
		"checkout_url":      "https://pay.example/c/sess-self",
		"auth_redirect_url": "https://pay.example/c/sess-self",
		"next_action":       "redirect_url",
		"next_action_type":  "redirect_url",
	})
	require.NoError(t, err)

	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Status: commonv1.STATUS_IN_PROCESS,
			Extras: extras,
		}),
	}
	cfg := defaultConfig()
	cfg.PublicBaseURL = "https://pay.example"
	b := newBusiness(cfg, redirectRegistry(), sessionRepo, linkRepo, payCli, &fakeProfileClient{})
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:      "sess-self",
		Status:   models.SessionStatusProcessing,
		PromptID: "prompt-self",
		Currency: "KES",
	}
	sessionRepo.sessions["sess-self"] = s

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusProcessing, updated.Status)
	if updated.Metadata != nil {
		assert.Empty(t, updated.Metadata["_redirect_url"], "own pay.* URL must not become 3DS redirect")
		assert.Empty(t, updated.Metadata["_next_action"], "self redirect_url next_action must be dropped")
	}
}

func TestRefreshStatus_CheckoutURL_Unsafe_Ignored(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()

	extras, err := structpb.NewStruct(map[string]any{
		"checkout_url": "javascript:alert(1)",
	})
	require.NoError(t, err)

	payCli := &fakePaymentClient{
		statusResp: connect.NewResponse(&commonv1.StatusResponse{
			Status: commonv1.STATUS_IN_PROCESS,
			Extras: extras,
		}),
	}
	b := newBusiness(defaultConfig(), redirectRegistry(), sessionRepo, linkRepo, payCli, &fakeProfileClient{})
	ctx := context.Background()

	s := &models.CheckoutSession{
		Ref:      "sess-unsafe-url",
		Status:   models.SessionStatusProcessing,
		PromptID: "prompt-unsafe",
		Currency: "USD",
	}
	sessionRepo.sessions["sess-unsafe-url"] = s
	initialCount := sessionRepo.updateCount

	updated, err := b.RefreshStatus(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, initialCount, sessionRepo.updateCount)
	if updated.Metadata != nil {
		assert.Empty(t, updated.Metadata["_redirect_url"])
	}
}
