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

package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/checkout/config"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/handlers"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/antinvestor/service-payments/apps/checkout/service/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Fakes — copied from business/checkout_test.go embedding approach
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
	promptResp  *connect.Response[paymentv1.InitiatePromptResponse]
	promptErr   error
	statusResp  *connect.Response[commonv1.StatusResponse]
	statusErr   error
	lastPrompt  *paymentv1.InitiatePromptRequest
	promptCount int
}

func (f *fakePaymentClient) InitiatePrompt(
	_ context.Context,
	req *connect.Request[paymentv1.InitiatePromptRequest],
) (*connect.Response[paymentv1.InitiatePromptResponse], error) {
	f.lastPrompt = req.Msg
	f.promptCount++
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
}

// ---------------------------------------------------------------------------
// Test harness helpers
// ---------------------------------------------------------------------------

//nolint:gochecknoglobals // test-only constant: fixed secret shared by all sub-tests in this package.
var testSecret = []byte("test-secret-for-web-handlers")

func testConfig() *config.CheckoutConfig {
	return &config.CheckoutConfig{
		SessionTTLMinutes:      30,
		MaxAttempts:            3,
		AttemptCooldownSeconds: 20,
		LinkSpawnPerMinute:     10,
		MethodsJSON:            `[{"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]},{"key":"mtn_momo","name":"MTN MoMo","route":"mtn","prefixes":["256"],"currencies":["UGX"]}]`,
		PublicBaseURL:          "http://localhost:8080",
		SigningSecret:          string(testSecret),
	}
}

func testRegistry() *business.MethodRegistry {
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

type testHarness struct {
	sessionRepo *fakeSessionRepo
	linkRepo    *fakeLinkRepo
	payCli      *fakePaymentClient
	biz         *business.CheckoutBusiness
	server      *handlers.WebServer
	router      http.Handler
}

func newHarness(t *testing.T) *testHarness {
	t.Helper()
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}

	cfg := testConfig()
	reg := testRegistry()

	biz := business.NewCheckoutBusiness(
		cfg,
		reg,
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
		nil,
	)
	biz = biz.WithClock(fixedNow)

	renderer, err := handlers.NewRenderer(testSecret)
	require.NoError(t, err)

	srv := handlers.NewWebServer(biz, renderer, reg, cfg)
	return &testHarness{
		sessionRepo: sessionRepo,
		linkRepo:    linkRepo,
		payCli:      payCli,
		biz:         biz,
		server:      srv,
		router:      srv.NewRouter(),
	}
}

func (h *testHarness) addPendingSession(ref string) *models.CheckoutSession {
	s := &models.CheckoutSession{
		Ref:          ref,
		Name:         "Test Merchant",
		Description:  "Test Description",
		Amount:       "100.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		Status:       models.SessionStatusPending,
		ExpiresAt:    fixedNow().Add(30 * time.Minute),
		Prefill: map[string]any{
			"displayName":   "Alice Wanjiru",
			"language":      "en",
			"clueProvider":  "mpesa",
			"clueContactId": "contact-1",
			"contacts": []any{
				map[string]any{
					"contactId": "contact-1",
					"msisdn":    "254712345678",
				},
			},
		},
		PayerProfileID: "profile-001",
	}
	h.sessionRepo.sessions[ref] = s
	return s
}

func (h *testHarness) addGuestPendingSession(ref string) {
	s := &models.CheckoutSession{
		Ref:          ref,
		Name:         "Guest Merchant",
		Description:  "Guest Description",
		Amount:       "50.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		Status:       models.SessionStatusPending,
		ExpiresAt:    fixedNow().Add(30 * time.Minute),
		// No PayerProfileID, no Prefill — guest
	}
	h.sessionRepo.sessions[ref] = s
}

// validCSRF returns a valid CSRF token for ref using testSecret.
func validCSRF(ref string) string {
	return handlers.CSRFToken(testSecret, ref)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// 1. GET /c/unknown → 404.
func TestHandlePage_UnknownSession_404(t *testing.T) {
	h := newHarness(t)

	req := httptest.NewRequest(http.MethodGet, "/c/unknownref", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// 2. Recognized pending session: 200, first name, masked digits, csrf, preselected method, no raw msisdn.
func TestHandlePage_RecognizedPending_PageData(t *testing.T) {
	h := newHarness(t)
	h.addPendingSession("sess-recognized")

	req := httptest.NewRequest(http.MethodGet, "/c/sess-recognized", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()

	// First name only — "Alice"
	assert.Contains(t, body, "Alice", "should contain first name")
	// Masked phone — should NOT contain the raw msisdn
	assert.NotContains(t, body, "254712345678", "raw msisdn must not appear in page")
	// CSRF token input
	assert.Contains(t, body, `name="csrf"`, "must have CSRF input")
	// Preselected method (mpesa should be checked/selected)
	assert.Contains(t, body, "mpesa", "method key must appear")
}

// 3. Guest pending session: body has name="phone" input.
func TestHandlePage_GuestPending_PhoneInput(t *testing.T) {
	h := newHarness(t)
	h.addGuestPendingSession("sess-guest")

	req := httptest.NewRequest(http.MethodGet, "/c/sess-guest", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `name="phone"`, "guest page must have phone input")
}

// 4. Expired session: gone page, body lacks amount and PayerName.
func TestHandlePage_ExpiredSession_GonePage(t *testing.T) {
	h := newHarness(t)
	s := &models.CheckoutSession{
		Ref:          "sess-expired",
		Name:         "Expired Merchant",
		Amount:       "999.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		Status:       models.SessionStatusExpired,
		ExpiresAt:    fixedNow().Add(-1 * time.Hour),
		Prefill: map[string]any{
			"displayName": "Eve Tester",
		},
	}
	h.sessionRepo.sessions["sess-expired"] = s

	req := httptest.NewRequest(http.MethodGet, "/c/sess-expired", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	// Gone page: must not contain amount or payer name (minimal data)
	assert.NotContains(t, body, "999.00", "gone page must not show amount")
	assert.NotContains(t, body, "Eve Tester", "gone page must not show payer name")
}

// 5. POST /pay no/wrong csrf → 403, business untouched.
func TestHandlePay_BadCSRF_403(t *testing.T) {
	h := newHarness(t)
	h.addGuestPendingSession("sess-csrf")

	form := url.Values{
		"csrf":   {"bad-token"},
		"method": {"mpesa"},
		"phone":  {"254712345678"},
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/c/sess-csrf/pay",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, 0, h.payCli.promptCount, "business must not be called on CSRF failure")
}

// 5b. POST /pay missing csrf → 403.
func TestHandlePay_MissingCSRF_403(t *testing.T) {
	h := newHarness(t)
	h.addGuestPendingSession("sess-no-csrf")

	form := url.Values{
		"method": {"mpesa"},
		"phone":  {"254712345678"},
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/c/sess-no-csrf/pay",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// 6. POST /pay happy guest path → 303, prompt called, Set-Cookie co_hints httpOnly.
func TestHandlePay_GuestHappyPath(t *testing.T) {
	h := newHarness(t)
	h.addGuestPendingSession("sess-pay-ok")

	form := url.Values{
		"csrf":   {validCSRF("sess-pay-ok")},
		"method": {"mpesa"},
		"phone":  {"254712345678"},
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/c/sess-pay-ok/pay",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusSeeOther, rec.Code, "should redirect to session page")
	assert.Equal(t, "/c/sess-pay-ok", rec.Header().Get("Location"))
	assert.Equal(t, 1, h.payCli.promptCount, "prompt must be called once")

	// co_hints cookie set with httpOnly
	cookies := rec.Result().Cookies()
	var hintsCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "co_hints" {
			hintsCookie = c
			break
		}
	}
	require.NotNil(t, hintsCookie, "co_hints cookie must be set")
	assert.True(t, hintsCookie.HttpOnly, "co_hints must be httpOnly")
	assert.True(t, hintsCookie.Secure, "co_hints must be Secure")
	assert.Equal(t, http.SameSiteLaxMode, hintsCookie.SameSite, "co_hints must be SameSite=Lax")
	assert.NotEmpty(t, hintsCookie.Value, "co_hints must have a value")
}

// 7. POST /pay when attempts exhausted → 429 with failure banner in body.
func TestHandlePay_TooManyAttempts_429(t *testing.T) {
	h := newHarness(t)
	s := &models.CheckoutSession{
		Ref:          "sess-exhausted",
		Name:         "Test Merchant",
		Amount:       "100.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		Status:       models.SessionStatusPending,
		ExpiresAt:    fixedNow().Add(30 * time.Minute),
		Attempts:     3, // MaxAttempts = 3 in testConfig
	}
	h.sessionRepo.sessions["sess-exhausted"] = s

	form := url.Values{
		"csrf":   {validCSRF("sess-exhausted")},
		"method": {"mpesa"},
		"phone":  {"254712345678"},
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/c/sess-exhausted/pay",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	// Re-rendered pay page must contain CSRF input and the localized too_many_attempts text.
	body := rec.Body.String()
	assert.Contains(t, body, `name="csrf"`, "re-rendered pay page must have CSRF input")
	assert.Contains(
		t,
		body,
		"Too many attempts",
		"body must contain localised too_many_attempts text",
	)
}

// 8. GET /c/{ref}/status completed session → JSON status completed.
func TestHandleStatus_CompletedSession(t *testing.T) {
	h := newHarness(t)
	s := &models.CheckoutSession{
		Ref:      "sess-status",
		Status:   models.SessionStatusCompleted,
		Currency: "KES",
		// No PromptID so RefreshStatus is a no-op
	}
	h.sessionRepo.sessions["sess-status"] = s

	req := httptest.NewRequest(http.MethodGet, "/c/sess-status/status", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var payload map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	assert.Equal(t, "completed", payload["status"])
	assert.NotContains(t, payload, "failure_reason", "failure_reason field must be absent")
}

// 8b. GET /c/{ref}/status unknown → 404 JSON.
func TestHandleStatus_Unknown_404JSON(t *testing.T) {
	h := newHarness(t)

	req := httptest.NewRequest(http.MethodGet, "/c/no-such-session/status", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	var payload map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	assert.Equal(t, "not_found", payload["error"])
}

// 8c. GET /c/{ref}/status failed session → JSON with status=failed and failure_reason.
func TestHandleStatus_FailedSession_IncludesFailureReason(t *testing.T) {
	h := newHarness(t)
	s := &models.CheckoutSession{
		Ref:      "sess-failed-status",
		Status:   models.SessionStatusFailed,
		Currency: "KES",
	}
	h.sessionRepo.sessions["sess-failed-status"] = s

	req := httptest.NewRequest(http.MethodGet, "/c/sess-failed-status/status", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var payload map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	assert.Equal(t, "failed", payload["status"])
	assert.NotEmpty(t, payload["failure_reason"], "failure_reason must be present for failed sessions")
}

// 9a. GET /l/{ref} active link → 303 to /c/<new session ref>.
func TestHandleLink_ActiveLink_Redirect(t *testing.T) {
	h := newHarness(t)
	link := &models.CheckoutLink{
		Ref:          "link-active",
		Name:         "Test Link",
		Currency:     "KES",
		Amount:       "50.00",
		AmountOption: models.AmountOptionFixed,
		Active:       true,
	}
	h.linkRepo.links["link-active"] = link

	req := httptest.NewRequest(http.MethodGet, "/l/link-active", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	location := rec.Header().Get("Location")
	assert.True(t, strings.HasPrefix(location, "/c/"), "redirect must go to /c/<ref>")
}

// 9b. GET /l/{ref} inactive link → 410 gone page.
func TestHandleLink_InactiveLink_410(t *testing.T) {
	h := newHarness(t)
	link := &models.CheckoutLink{
		Ref:      "link-inactive",
		Name:     "Inactive Link",
		Currency: "KES",
		Active:   false,
	}
	h.linkRepo.links["link-inactive"] = link

	req := httptest.NewRequest(http.MethodGet, "/l/link-inactive", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusGone, rec.Code)
}

// 9c. Rate limit: with LinkSpawnPerMinute=2, third call from same IP → 429.
func TestHandleLink_RateLimit_429(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}

	// Tight rate limit for test
	cfg := testConfig()
	cfg.LinkSpawnPerMinute = 2

	reg := testRegistry()
	biz := business.NewCheckoutBusiness(
		cfg,
		reg,
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
		nil,
	)
	biz = biz.WithClock(fixedNow)

	renderer, err := handlers.NewRenderer(testSecret)
	require.NoError(t, err)

	srv := handlers.NewWebServer(biz, renderer, reg, cfg)
	router := srv.NewRouter()

	// Add an active link
	link := &models.CheckoutLink{
		Ref:          "link-rate",
		Name:         "Rate Test Link",
		Currency:     "KES",
		Amount:       "10.00",
		AmountOption: models.AmountOptionFixed,
		Active:       true,
	}
	linkRepo.links["link-rate"] = link

	makeReq := func() *http.Response {
		req := httptest.NewRequest(http.MethodGet, "/l/link-rate", nil)
		req.RemoteAddr = "10.0.0.1:1234" // same IP
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Result()
	}

	// First two calls should succeed (303 redirect)
	resp1 := makeReq()
	assert.Equal(t, http.StatusSeeOther, resp1.StatusCode, "first call should succeed")

	resp2 := makeReq()
	assert.Equal(t, http.StatusSeeOther, resp2.StatusCode, "second call should succeed")

	// Third call should be rate-limited
	resp3 := makeReq()
	assert.Equal(
		t,
		http.StatusTooManyRequests,
		resp3.StatusCode,
		"third call should be rate limited",
	)
}

// 10. Static: GET /static/checkout.css → 200 text/css.
func TestStatic_CheckoutCSS(t *testing.T) {
	h := newHarness(t)

	req := httptest.NewRequest(http.MethodGet, "/static/checkout.css", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	ct := rec.Header().Get("Content-Type")
	assert.Contains(t, ct, "text/css", "should serve text/css")
}

// 11. POST /pay guest with valid CSRF + method but EMPTY phone → 400 + banner, no prompt sent.
func TestHandlePay_GuestEmptyPhone_400(t *testing.T) {
	h := newHarness(t)
	h.addGuestPendingSession("sess-no-phone")

	form := url.Values{
		"csrf":   {validCSRF("sess-no-phone")},
		"method": {"mpesa"},
		"phone":  {""}, // empty phone
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/c/sess-no-phone/pay",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "empty phone must yield 400")
	body := rec.Body.String()
	// The re-rendered pay page should contain the localized contact_required text
	assert.True(
		t,
		strings.Contains(body, `class="banner`) ||
			strings.Contains(body, "Enter your phone number"),
		"response body must contain a failure banner or the localized contact_required text",
	)
	assert.Equal(t, 0, h.payCli.promptCount, "no prompt must be sent for empty phone")
}

// 12. Rate-limiter X-Forwarded-For: two requests with the same XFF first-IP share a bucket;
// a different XFF first-IP gets its own bucket.
func TestRateLimiter_XFF_Buckets(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}

	cfg := testConfig()
	cfg.LinkSpawnPerMinute = 1 // one token → second request from same bucket is rate-limited

	reg := testRegistry()
	biz := business.NewCheckoutBusiness(
		cfg,
		reg,
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
		nil,
	)
	biz = biz.WithClock(fixedNow)

	renderer, err := handlers.NewRenderer(testSecret)
	require.NoError(t, err)

	srv := handlers.NewWebServer(biz, renderer, reg, cfg)
	router := srv.NewRouter()

	// Active link
	link := &models.CheckoutLink{
		Ref:          "link-xff",
		Name:         "XFF Rate Test",
		Currency:     "KES",
		Amount:       "10.00",
		AmountOption: models.AmountOptionFixed,
		Active:       true,
	}
	linkRepo.links["link-xff"] = link

	makeReq := func(xff string) *http.Response {
		req := httptest.NewRequest(http.MethodGet, "/l/link-xff", nil)
		req.Header.Set("X-Forwarded-For", xff)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Result()
	}

	// Rightmost hop is the gateway-appended value and is trusted.
	// Leftmost values are attacker-controlled and must be ignored.

	// First request: rightmost=1.2.3.4, leftmost=10.0.0.9 (spoofed) → consumes the one token.
	resp1 := makeReq("10.0.0.9, 1.2.3.4")
	assert.Equal(
		t,
		http.StatusSeeOther,
		resp1.StatusCode,
		"first request (rightmost=1.2.3.4) should succeed",
	)

	// Second request: different spoofed leftmost but SAME rightmost=1.2.3.4 → same bucket, rate-limited.
	resp2 := makeReq("99.99.99.99, 1.2.3.4")
	assert.Equal(
		t,
		http.StatusTooManyRequests,
		resp2.StatusCode,
		"spoofed-different-leftmost but same rightmost must share a bucket and be rate-limited",
	)

	// Request with a different rightmost hop → different bucket, succeeds.
	resp3 := makeReq("10.0.0.9, 5.6.7.8")
	assert.Equal(
		t,
		http.StatusSeeOther,
		resp3.StatusCode,
		"different rightmost IP gets its own bucket",
	)
}

// 13. Session with javascript: ReturnURL: done page must NOT include a meta-refresh
// or any link/redirect target containing "javascript:".
func TestHandlePage_CompletedSession_JavascriptReturnURL_NoRefresh(t *testing.T) {
	h := newHarness(t)
	s := &models.CheckoutSession{
		Ref:          "sess-jsurl",
		Name:         "Merchant",
		Amount:       "100.00",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		Status:       models.SessionStatusCompleted,
		ReturnURL:    "javascript:alert(1)", // malicious URL stored (simulate bypass)
	}
	h.sessionRepo.sessions["sess-jsurl"] = s

	req := httptest.NewRequest(http.MethodGet, "/c/sess-jsurl", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	// Defense-in-depth: buildReturnURL must strip the javascript: URL.
	assert.NotContains(
		t,
		body,
		"javascript:",
		"page must not contain javascript: scheme in any context",
	)
	assert.NotContains(t, body, "meta-refresh", "page must not contain meta-refresh")
	// The meta http-equiv refresh tag should not be present.
	assert.NotContains(
		t,
		body,
		`http-equiv="refresh"`,
		"page must not emit a meta refresh redirect",
	)
}
