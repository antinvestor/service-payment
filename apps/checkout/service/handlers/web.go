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

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/antinvestor/service-payments/apps/checkout/config"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/antinvestor/service-payments/apps/checkout/service/web"
	"github.com/pitabwire/util"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// rateLimiter — token bucket per key, mutex-protected
// ---------------------------------------------------------------------------

// rateLimiterPruneWindow is how long an entry must be idle before being pruned.
const (
	rateLimiterPruneWindow = 10 * time.Minute
	secondsPerMinute       = 60.0
	guestCookieMaxAge      = 180 * 24 * 60 * 60 // 180 days in seconds
	splitFirst             = 2                  // SplitN: keep first two parts only
)

type rateLimiterEntry struct {
	tokens   float64
	lastSeen time.Time
}

// rateLimiter implements a per-key token bucket rate limiter.
type rateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*rateLimiterEntry
	capacity float64 // max tokens
	rate     float64 // tokens per second
	now      func() time.Time
}

func newRateLimiter(perMinute int, nowFn func() time.Time) *rateLimiter {
	if nowFn == nil {
		nowFn = time.Now
	}
	capacity := float64(perMinute)
	rate := capacity / secondsPerMinute
	return &rateLimiter{
		entries:  make(map[string]*rateLimiterEntry),
		capacity: capacity,
		rate:     rate,
		now:      nowFn,
	}
}

// Allow returns true when the key has remaining capacity, consuming one token.
// It also prunes entries older than rateLimiterPruneWindow on each access.
func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	r.pruneOlderThan(now, rateLimiterPruneWindow)

	entry, exists := r.entries[key]
	if !exists {
		entry = &rateLimiterEntry{
			tokens:   r.capacity,
			lastSeen: now,
		}
		r.entries[key] = entry
	} else {
		// Refill tokens based on elapsed time
		elapsed := now.Sub(entry.lastSeen).Seconds()
		entry.tokens += elapsed * r.rate
		if entry.tokens > r.capacity {
			entry.tokens = r.capacity
		}
		entry.lastSeen = now
	}

	if entry.tokens < 1 {
		return false
	}
	entry.tokens--
	return true
}

// pruneOlderThan removes entries that haven't been seen within the window.
// Caller must hold the mutex.
func (r *rateLimiter) pruneOlderThan(now time.Time, window time.Duration) {
	for k, e := range r.entries {
		if now.Sub(e.lastSeen) > window {
			delete(r.entries, k)
		}
	}
}

// ---------------------------------------------------------------------------
// WebServer
// ---------------------------------------------------------------------------

// WebServer serves the public checkout pages.
type WebServer struct {
	business     *business.CheckoutBusiness
	renderer     *Renderer
	registry     *business.MethodRegistry
	cfg          *config.CheckoutConfig
	spawnLimiter *rateLimiter
}

// NewWebServer creates a WebServer with a real clock rate limiter.
func NewWebServer(
	biz *business.CheckoutBusiness,
	renderer *Renderer,
	registry *business.MethodRegistry,
	cfg *config.CheckoutConfig,
) *WebServer {
	return &WebServer{
		business:     biz,
		renderer:     renderer,
		registry:     registry,
		cfg:          cfg,
		spawnLimiter: newRateLimiter(cfg.LinkSpawnPerMinute, nil),
	}
}

// NewRouter returns a ServeMux with all checkout routes registered.
func (s *WebServer) NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(mustSub(web.Static, "static"))))
	mux.HandleFunc("GET /c/{ref}", s.HandlePage)
	mux.HandleFunc("POST /c/{ref}/pay", s.HandlePay)
	mux.HandleFunc("GET /c/{ref}/status", s.HandleStatus)
	mux.HandleFunc("GET /l/{ref}", s.HandleLink)
	return mux
}

// mustSub panics if fs.Sub fails; the embedded FS is always valid.
func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(fmt.Sprintf("handlers: fs.Sub(%q) failed: %v", dir, err))
	}
	return sub
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// signingSecret returns the signing secret as bytes.
func (s *WebServer) signingSecret() []byte {
	return []byte(s.cfg.SigningSecret)
}

// pickLang resolves a display language from request signals.
// Priority: ?lang query (if "en"/"fr") → prefill["language"] → Accept-Language → "en".
func pickLang(r *http.Request, prefillLang string) string {
	if q := r.URL.Query().Get("lang"); q == "en" || q == "fr" {
		return q
	}
	if prefillLang == "en" || prefillLang == "fr" {
		return prefillLang
	}
	if al := r.Header.Get("Accept-Language"); al != "" {
		// First tag: strip region/quality (e.g. "fr-FR,fr;q=0.9" → "fr")
		tag := strings.FieldsFunc(al, func(c rune) bool { return c == ',' || c == ';' })[0]
		tag = strings.ToLower(strings.SplitN(tag, "-", splitFirst)[0])
		if tag == "en" || tag == "fr" {
			return tag
		}
	}
	return "en"
}

// firstWord returns the first whitespace-delimited word of s.
func firstWord(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, ' '); idx >= 0 {
		return s[:idx]
	}
	return s
}

// buildReturnURL appends session=<ref>&status=<status> to the session's ReturnURL.
// Returns "" if ReturnURL is empty or fails the safe-URL check (http/https only).
// This is defense-in-depth: business validation is the primary gate.
func buildReturnURL(returnURL, ref, status string) string {
	if !business.IsSafeReturnURL(returnURL) {
		return ""
	}
	u, err := url.Parse(returnURL)
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("session", ref)
	q.Set("status", status)
	u.RawQuery = q.Encode()
	return u.String()
}

// extractContacts builds []ContactChoice and the clue msisdn from prefill.
// The clue msisdn is the raw msisdn for method preselect logic (not for display).
func extractContacts(prefill map[string]any, clueContactID string) ([]ContactChoice, string) {
	contactsRaw, hasContacts := prefill["contacts"]
	if !hasContacts {
		return nil, ""
	}
	list, isList := contactsRaw.([]any)
	if !isList {
		return nil, ""
	}

	contacts := make([]ContactChoice, 0, len(list))
	cluePhone := ""

	for _, raw := range list {
		m, isMap := raw.(map[string]any)
		if !isMap {
			continue
		}
		cid, _ := m["contactId"].(string)
		msisdn, _ := m["msisdn"].(string)
		contacts = append(contacts, ContactChoice{ContactID: cid, Masked: MaskMsisdn(msisdn)})
		if cid == clueContactID {
			cluePhone = msisdn
		}
	}

	// Fallback clue: first contact's phone if the clue contact wasn't found
	if cluePhone == "" && len(list) > 0 {
		if m, isMap := list[0].(map[string]any); isMap {
			cluePhone, _ = m["msisdn"].(string)
		}
	}

	return contacts, cluePhone
}

// methodChoices builds []MethodChoice from available methods, marking the preselected one.
func methodChoices(available []business.Method, selected business.Method) []MethodChoice {
	choices := make([]MethodChoice, 0, len(available))
	for _, m := range available {
		choices = append(choices, MethodChoice{
			Key:      m.Key,
			Name:     m.Name,
			Selected: m.Key == selected.Key,
		})
	}
	return choices
}

// restrictionKeys extracts the restriction key list from a session's Methods JSONMap.
func restrictionKeys(methods map[string]any) []string {
	if methods == nil {
		return nil
	}
	keysRaw, hasKeys := methods["keys"]
	if !hasKeys {
		return nil
	}
	list, isList := keysRaw.([]any)
	if !isList {
		return nil
	}
	keys := make([]string, 0, len(list))
	for _, k := range list {
		if ks, isStr := k.(string); isStr {
			keys = append(keys, ks)
		}
	}
	return keys
}

// guestHintsFromCookie reads the co_hints cookie and decodes it.
// Returns zero-value GuestHints when absent or invalid.
func (s *WebServer) guestHintsFromCookie(r *http.Request) business.GuestHints {
	cookie, err := r.Cookie("co_hints")
	if err != nil {
		return business.GuestHints{}
	}
	hints, ok := business.DecodeGuestHints(s.signingSecret(), cookie.Value)
	if !ok {
		return business.GuestHints{}
	}
	return hints
}

// buildMethods returns the method choices for a session, using prefill and optional guest hints.
func (s *WebServer) buildMethods(session *models.CheckoutSession, r *http.Request, cluePhone string) []MethodChoice {
	restriction := restrictionKeys(session.Methods)
	available := s.registry.Available(restriction)
	if len(available) == 0 {
		return nil
	}

	clueKey := ""
	phone := cluePhone

	if session.Prefill != nil {
		clueKey, _ = session.Prefill["clueProvider"].(string)
	}

	// Guest: supplement with cookie hints for method preselect only
	// (never echo hint phone to HTML — privacy)
	if session.PayerProfileID == "" {
		hints := s.guestHintsFromCookie(r)
		if clueKey == "" {
			clueKey = hints.Method
		}
		if phone == "" {
			phone = hints.Phone
		}
	}

	selected := business.Preselect(available, clueKey, phone)
	return methodChoices(available, selected)
}

// pageDataFor builds a PageData from a session and request (common fields).
func (s *WebServer) pageDataFor(session *models.CheckoutSession, r *http.Request) PageData {
	ref := session.Ref

	// Resolve language
	prefillLang := ""
	if session.Prefill != nil {
		prefillLang, _ = session.Prefill["language"].(string)
	}
	lang := pickLang(r, prefillLang)

	data := PageData{
		Lang:         lang,
		SessionRef:   ref,
		MerchantName: session.Name,
		Description:  session.Description,
		Currency:     session.Currency,
		Variable:     session.AmountOption == models.AmountOptionVariable,
		CSRF:         CSRFToken(s.signingSecret(), ref),
		Status:       session.Status,
	}

	// AmountDisplay: only for fixed, non-empty amounts
	if session.Amount != "" && session.AmountOption != models.AmountOptionVariable {
		if money, err := business.MoneyFromAmount(session.Amount, session.Currency); err == nil {
			data.AmountDisplay = business.FormatMoney(money)
		}
		// on parse error leave AmountDisplay "" — defensive
	}

	// Payer block (cluePhone used internally for method preselect — never echoed to HTML)
	cluePhone := ""
	if session.Prefill != nil {
		clueContactID, _ := session.Prefill["clueContactId"].(string)
		var contacts []ContactChoice
		contacts, cluePhone = extractContacts(session.Prefill, clueContactID)
		data.Contacts = contacts
		if cluePhone != "" {
			data.MaskedPhone = MaskMsisdn(cluePhone)
		}
		displayName, _ := session.Prefill["displayName"].(string)
		data.PayerName = firstWord(displayName)
	}

	data.Methods = s.buildMethods(session, r, cluePhone)
	return data
}

// clientIP extracts the real client IP from X-Forwarded-For or RemoteAddr.
// The service runs behind a trusted gateway proxy that appends the connecting
// IP as the LAST comma-separated hop.  Leftmost values are attacker-controlled
// and MUST NOT be trusted for rate-limiting or security decisions.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[len(parts)-1])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// renderError writes a plain-text HTTP error response.
func renderError(w http.ResponseWriter, code int, msg string) {
	http.Error(w, msg, code)
}

// renderPage renders a named page with the given data, writing the status code first.
func (s *WebServer) renderPage(w http.ResponseWriter, r *http.Request, code int, page string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	if err := s.renderer.Render(w, page, data); err != nil {
		// Template rendering failed after headers sent — log and emit a comment.
		util.Log(r.Context()).WithError(err).Error("template render failed")
		_, _ = w.Write([]byte("<!-- render error -->"))
	}
}

// ---------------------------------------------------------------------------
// HandlePage — GET /c/{ref}
// ---------------------------------------------------------------------------

// HandlePage renders the appropriate checkout page for the session.
func (s *WebServer) HandlePage(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")
	session, err := s.business.GetSessionByRef(r.Context(), ref)
	if err != nil {
		if isNotFoundErr(err) {
			renderError(w, http.StatusNotFound, "not found")
			return
		}
		renderError(w, http.StatusInternalServerError, "internal error")
		return
	}

	switch session.Status {
	case models.SessionStatusCompleted:
		data := s.pageDataFor(session, r)
		data.RedirectURL = buildReturnURL(session.ReturnURL, ref, "completed")
		s.renderPage(w, r, http.StatusOK, "done", data)

	case models.SessionStatusExpired:
		// Minimal data — nothing sensitive
		data := PageData{Lang: pickLang(r, "")}
		s.renderPage(w, r, http.StatusOK, "gone", data)

	case models.SessionStatusProcessing:
		data := s.pageDataFor(session, r)
		data.PollURL = "/c/" + ref + "/status"
		data.ReturnURL = buildReturnURL(session.ReturnURL, ref, "completed")
		s.renderPage(w, r, http.StatusOK, "confirm", data)

	case models.SessionStatusFailed:
		data := s.pageDataFor(session, r)
		data.FailureReason = T(data.Lang, "failed_title")
		s.renderPage(w, r, http.StatusOK, "pay", data)

	default: // pending
		data := s.pageDataFor(session, r)
		s.renderPage(w, r, http.StatusOK, "pay", data)
	}
}

// ---------------------------------------------------------------------------
// HandlePay — POST /c/{ref}/pay
// ---------------------------------------------------------------------------

// HandlePay processes a payment form submission.
func (s *WebServer) HandlePay(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")

	if err := r.ParseForm(); err != nil {
		renderError(w, http.StatusBadRequest, "bad request")
		return
	}

	// CSRF check
	token := r.FormValue("csrf")
	if !VerifyCSRF(s.signingSecret(), ref, token) {
		renderError(w, http.StatusForbidden, "forbidden")
		return
	}

	in := business.PayInput{
		MethodKey:   r.FormValue("method"),
		PhoneNumber: strings.TrimSpace(r.FormValue("phone")),
		ContactID:   r.FormValue("contact_id"),
		Amount:      r.FormValue("amount"),
	}

	_, payErr := s.business.Pay(r.Context(), ref, in)
	if payErr != nil {
		s.handlePayError(w, r, ref, payErr)
		return
	}

	// Success: set guest cookie if phone provided
	if in.PhoneNumber != "" {
		hints := business.GuestHints{Phone: in.PhoneNumber, Method: in.MethodKey}
		cookieVal := business.EncodeGuestHints(s.signingSecret(), hints)
		http.SetCookie(w, &http.Cookie{
			Name:     "co_hints",
			Value:    cookieVal,
			Path:     "/",
			MaxAge:   guestCookieMaxAge,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	//nolint:gosec // G710: ref is retrieved from the database; it is not user-controlled redirect input.
	http.Redirect(w, r, "/c/"+ref, http.StatusSeeOther)
}

// handlePayError maps business errors to HTTP responses for HandlePay.
// Error messages are localised via translation keys — the raw error text is
// never reflected to the page to prevent user-input leakage.
func (s *WebServer) handlePayError(w http.ResponseWriter, r *http.Request, ref string, payErr error) {
	switch {
	case errors.Is(payErr, business.ErrSessionGone):
		//nolint:gosec // G710: ref is retrieved from the database; not user-controlled redirect input.
		http.Redirect(w, r, "/c/"+ref, http.StatusSeeOther)
		return

	case errors.Is(payErr, business.ErrTooManyAttempts):
		s.reRenderPayWithError(w, r, ref, "too_many_attempts", http.StatusTooManyRequests)
		return

	case errors.Is(payErr, business.ErrCooldown):
		s.reRenderPayWithError(w, r, ref, "cooldown", http.StatusTooManyRequests)
		return

	case errors.Is(payErr, business.ErrUnknownMethod):
		s.reRenderPayWithError(w, r, ref, "bad_method", http.StatusBadRequest)
		return

	case errors.Is(payErr, business.ErrAmountRequired):
		s.reRenderPayWithError(w, r, ref, "amount_required", http.StatusBadRequest)
		return

	case errors.Is(payErr, business.ErrContactRequired):
		s.reRenderPayWithError(w, r, ref, "contact_required", http.StatusBadRequest)
		return

	default:
		if isNotFoundErr(payErr) {
			renderError(w, http.StatusNotFound, "not found")
			return
		}
		s.reRenderPayWithError(w, r, ref, "failed_title", http.StatusInternalServerError)
	}
}

// reRenderPayWithError reloads the session and re-renders the pay page with a localised
// failure reason. msgKey is a translation key; it is resolved after the session's language
// is determined so the message always appears in the payer's language.
func (s *WebServer) reRenderPayWithError(w http.ResponseWriter, r *http.Request, ref, msgKey string, code int) {
	session, sessionErr := s.business.GetSessionByRef(r.Context(), ref)
	if sessionErr != nil || session == nil {
		renderError(w, code, "")
		return
	}
	data := s.pageDataFor(session, r)
	data.FailureReason = T(data.Lang, msgKey)
	s.renderPage(w, r, code, "pay", data)
}

// ---------------------------------------------------------------------------
// HandleStatus — GET /c/{ref}/status
// ---------------------------------------------------------------------------

// HandleStatus returns the session status as JSON for polling.
func (s *WebServer) HandleStatus(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")

	session, err := s.business.GetSessionByRef(r.Context(), ref)
	if err != nil {
		if isNotFoundErr(err) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
		return
	}

	// Refresh if processing
	if session.Status == models.SessionStatusProcessing {
		if refreshed, refreshErr := s.business.RefreshStatus(r.Context(), session); refreshErr == nil {
			session = refreshed
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": session.Status,
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// HandleLink — GET /l/{ref}
// ---------------------------------------------------------------------------

// HandleLink spawns a session from a link and redirects to the session page.
func (s *WebServer) HandleLink(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.spawnLimiter.Allow(ip) {
		renderError(w, http.StatusTooManyRequests, "too many requests")
		return
	}

	ref := r.PathValue("ref")
	session, err := s.business.SpawnSession(r.Context(), ref)
	if err != nil {
		if errors.Is(err, business.ErrLinkUnusable) {
			data := PageData{Lang: pickLang(r, "")}
			s.renderPage(w, r, http.StatusGone, "gone", data)
			return
		}
		if isNotFoundErr(err) {
			renderError(w, http.StatusNotFound, "not found")
			return
		}
		renderError(w, http.StatusInternalServerError, "internal error")
		return
	}

	//nolint:gosec // G710: session.Ref is generated by the database, not a user-supplied redirect target.
	http.Redirect(w, r, "/c/"+session.Ref, http.StatusSeeOther)
}

// ---------------------------------------------------------------------------
// isNotFoundErr checks whether an error wraps gorm.ErrRecordNotFound.
// ---------------------------------------------------------------------------

func isNotFoundErr(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
