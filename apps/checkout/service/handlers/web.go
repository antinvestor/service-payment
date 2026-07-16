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
	partitions   business.PartitionAllowlists
	cfg          *config.CheckoutConfig
	spawnLimiter *rateLimiter
}

// NewWebServer creates a WebServer with a real clock rate limiter.
// partitions may be empty (no partition-level method filtering).
func NewWebServer(
	biz *business.CheckoutBusiness,
	renderer *Renderer,
	registry *business.MethodRegistry,
	cfg *config.CheckoutConfig,
	partitions business.PartitionAllowlists,
) *WebServer {
	if partitions == nil {
		partitions = business.PartitionAllowlists{}
	}
	return &WebServer{
		business:     biz,
		renderer:     renderer,
		registry:     registry,
		partitions:   partitions,
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
	mux.HandleFunc("GET /c/{ref}/crypto", s.HandleCardCrypto)
	mux.HandleFunc("POST /c/{ref}/authorize", s.HandleAuthorize)
	mux.HandleFunc("GET /l/{ref}", s.HandleLink)
	// Cluster product services create sessions here (token auth), then redirect
	// the browser to pay.*/c/{ref} — never Flutterwave multipay.
	mux.HandleFunc("POST /internal/v1/sessions", s.HandleInternalCreateSession)
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

// extractContacts builds []ContactChoice and the preferred msisdn from prefill.
// Preferred phone (last successful payment contact) is selected first when present.
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

	// Prefer explicit phone prefer id from prefill when set.
	if clueContactID == "" {
		if pid, _ := prefill["cluePhoneContactId"].(string); pid != "" {
			clueContactID = pid
		}
	}

	contacts := make([]ContactChoice, 0, len(list))
	cluePhone := ""
	// Also honour prefill.phone when chips exist.
	if p, _ := prefill["phone"].(string); p != "" {
		cluePhone = p
	}

	for _, raw := range list {
		m, isMap := raw.(map[string]any)
		if !isMap {
			continue
		}
		cid, _ := m["contactId"].(string)
		msisdn, _ := m["msisdn"].(string)
		preferred, _ := m["preferred"].(bool)
		contacts = append(contacts, ContactChoice{
			ContactID: cid,
			Masked:    MaskMsisdn(msisdn),
			Preferred: preferred || (clueContactID != "" && cid == clueContactID),
		})
		if cid == clueContactID || preferred {
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
			Embed:    m.IsEmbedded(),
			Redirect: m.Redirect,
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

// buildMethods returns Link-style method choices for a session:
// location + partition config + cached last-used preference.
func (s *WebServer) buildMethods(session *models.CheckoutSession, r *http.Request, cluePhone string) []MethodChoice {
	clueKey := ""
	phone := cluePhone
	guestMethod := ""

	if session.Prefill != nil {
		// Prefer last provider key; fall back to lastMethod if it is a registry key.
		if v, _ := session.Prefill["clueProvider"].(string); v != "" {
			clueKey = v
		} else if v, _ := session.Prefill["clueMethod"].(string); v != "" {
			clueKey = v
		}
	}

	// Guest device cookie supplies last method + phone locality hints
	// (never echo raw phone to HTML — privacy).
	hints := s.guestHintsFromCookie(r)
	if session.PayerProfileID == "" {
		guestMethod = hints.Method
		if phone == "" {
			phone = hints.Phone
		}
	} else if phone == "" && hints.Phone != "" {
		// Recognized payer with no contacts: still use device phone for locality only.
		phone = hints.Phone
	}

	// Location priority: edge geo headers → profile last country → guest cookie country.
	country := business.DetectCountryFromHeaders(r.Header.Get)
	if country == "" && session.Prefill != nil {
		if c, _ := session.Prefill["country"].(string); c != "" {
			country = strings.ToUpper(strings.TrimSpace(c))
		}
	}
	if country == "" && hints.Country != "" {
		country = strings.ToUpper(strings.TrimSpace(hints.Country))
	}

	filter := business.MethodFilter{
		Currency:           session.Currency,
		Phone:              phone,
		Country:            country,
		SessionRestriction: restrictionKeys(session.Methods),
		PartitionAllowlist: s.partitions.ForPartition(session.PartitionID),
		ClueMethod:         clueKey,
		GuestMethod:        guestMethod,
	}

	resolved := s.registry.Resolve(filter)
	if len(resolved.Available) == 0 {
		return nil
	}
	return methodChoices(resolved.Available, resolved.Selected)
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
	email := ""
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
		email, _ = session.Prefill["email"].(string)
		if email != "" {
			data.MaskedEmail = maskEmail(email)
		}
		if pmd, _ := session.Prefill["paymentMethodId"].(string); pmd != "" {
			data.HasSavedCard = true
		}
	}
	data.NeedEmail = strings.TrimSpace(email) == ""
	data.NeedName = strings.TrimSpace(data.PayerName) == ""
	data.Methods = s.buildMethods(session, r, cluePhone)

	// Card form when selected method is embedded card and encryption is configured.
	selectedEmbed := false
	for _, m := range data.Methods {
		if m.Selected && m.Embed {
			selectedEmbed = true
			break
		}
	}
	if selectedEmbed && s.cfg.ResolvedCardEncryptionKey() != "" {
		data.ShowCardForm = true
		data.CardCryptoURL = "/c/" + ref + "/crypto"
	}
	data.AuthorizeURL = "/c/" + ref + "/authorize"

	// Surface next_action on processing sessions (PIN / OTP / 3DS).
	if session.Metadata != nil {
		if na, _ := session.Metadata["_next_action"].(string); na != "" {
			data.NextAction = na
		}
		if note, _ := session.Metadata["_payment_instruction"].(string); note != "" {
			data.PaymentInstruction = note
		}
	}
	return data
}

// maskEmail shows first char + domain for Link-style identity strip.
func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return "•••"
	}
	local := email[:at]
	domain := email[at:]
	if len(local) == 1 {
		return local + "•••" + domain
	}
	return string(local[0]) + "•••" + domain
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
		MethodKey:       r.FormValue("method"),
		PhoneNumber:     strings.TrimSpace(r.FormValue("phone")),
		ContactID:       r.FormValue("contact_id"),
		Amount:          r.FormValue("amount"),
		Email:           strings.TrimSpace(r.FormValue("email")),
		GuestEmail:      strings.TrimSpace(r.FormValue("email")),
		Name:            strings.TrimSpace(r.FormValue("name")),
		PaymentMethodID: strings.TrimSpace(r.FormValue("payment_method_id")),
		CustomerID:      strings.TrimSpace(r.FormValue("customer_id")),
	}
	// Encrypted card fields (browser AES-GCM) — never clear PAN.
	if enc := strings.TrimSpace(r.FormValue("encrypted_card_number")); enc != "" {
		in.Card = &business.EncryptedCardInput{
			EncryptedCardNumber:  enc,
			EncryptedExpiryMonth: strings.TrimSpace(r.FormValue("encrypted_expiry_month")),
			EncryptedExpiryYear:  strings.TrimSpace(r.FormValue("encrypted_expiry_year")),
			EncryptedCVV:         strings.TrimSpace(r.FormValue("encrypted_cvv")),
			Nonce:                strings.TrimSpace(r.FormValue("card_nonce")),
		}
	}
	// Saved card one-click: use profile token when form checkbox set.
	if r.FormValue("use_saved_card") == "1" && in.PaymentMethodID == "" {
		if session, err := s.business.GetSessionByRef(r.Context(), ref); err == nil && session.Prefill != nil {
			if pmd, _ := session.Prefill["paymentMethodId"].(string); pmd != "" {
				in.PaymentMethodID = pmd
			}
			if cus, _ := session.Prefill["providerCustomerId"].(string); cus != "" {
				in.CustomerID = cus
			}
		}
	}

	_, payErr := s.business.Pay(r.Context(), ref, in)
	if payErr != nil {
		s.handlePayError(w, r, ref, payErr)
		return
	}

	// Success: cache device hints (method + optional phone locality) like Link.
	if in.MethodKey != "" || in.PhoneNumber != "" {
		country := business.InferCountryFromPhone(in.PhoneNumber)
		if country == "" {
			country = business.DetectCountryFromHeaders(r.Header.Get)
		}
		hints := business.GuestHints{
			Phone:   in.PhoneNumber,
			Method:  in.MethodKey,
			Country: country,
		}
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

	payload := map[string]string{
		"status": session.Status,
	}
	if session.Status == models.SessionStatusFailed {
		lang := pickLang(r, "")
		payload["failure_reason"] = T(lang, "failed_title")
	}
	// Surface next steps for embedded card (3DS URL, PIN/OTP) while processing.
	// Never emit our own pay.* URL as redirect_url — that loops confirm forever.
	if session.Status == models.SessionStatusProcessing && session.Metadata != nil {
		publicBase := ""
		if s.cfg != nil {
			publicBase = s.cfg.PublicBaseURL
		}
		if redirectURL, ok := session.Metadata["_redirect_url"].(string); ok && redirectURL != "" {
			if business.IsExternalAuthRedirect(redirectURL, publicBase) {
				payload["redirect_url"] = redirectURL
			}
		}
		if na, ok := session.Metadata["_next_action"].(string); ok && na != "" {
			if na == "redirect_url" && payload["redirect_url"] == "" {
				// No external target — ignore self/empty redirect next_action.
			} else {
				payload["next_action"] = na
			}
		}
		if note, ok := session.Metadata["_payment_instruction"].(string); ok && note != "" {
			payload["payment_instruction"] = note
		}
	}
	writeJSON(w, http.StatusOK, payload)
}

// HandleCardCrypto returns the AES-256 key for browser-side card encryption.
// Only the encryption key (not secret/client secrets) is exposed — same model as FW docs.
func (s *WebServer) HandleCardCrypto(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")
	if _, err := s.business.GetSessionByRef(r.Context(), ref); err != nil {
		if isNotFoundErr(err) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
		return
	}
	key := s.cfg.ResolvedCardEncryptionKey()
	if key == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "card_encryption_not_configured",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"encryption_key": key,
		"algorithm":      "AES-GCM",
		"nonce_length":   "12",
	})
}

// HandleAuthorize accepts PIN / OTP / AVS for a processing card charge.
func (s *WebServer) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")
	if err := r.ParseForm(); err != nil {
		renderError(w, http.StatusBadRequest, "bad request")
		return
	}
	token := r.FormValue("csrf")
	if !VerifyCSRF(s.signingSecret(), ref, token) {
		renderError(w, http.StatusForbidden, "forbidden")
		return
	}
	in := business.AuthorizeInput{
		Type:         strings.TrimSpace(r.FormValue("auth_type")),
		PIN:          strings.TrimSpace(r.FormValue("pin")),
		EncryptedPIN: strings.TrimSpace(r.FormValue("encrypted_pin")),
		Nonce:        strings.TrimSpace(r.FormValue("nonce")),
		OTP:          strings.TrimSpace(r.FormValue("otp")),
		City:         strings.TrimSpace(r.FormValue("avs_city")),
		Country:      strings.TrimSpace(r.FormValue("avs_country")),
		Line1:        strings.TrimSpace(r.FormValue("avs_line1")),
		Line2:        strings.TrimSpace(r.FormValue("avs_line2")),
		PostalCode:   strings.TrimSpace(r.FormValue("avs_postal_code")),
		State:        strings.TrimSpace(r.FormValue("avs_state")),
	}
	if _, err := s.business.Authorize(r.Context(), ref, in); err != nil {
		if isNotFoundErr(err) {
			renderError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "authorize_failed"})
		return
	}
	// Prefer redirect back to confirm page for form posts; JSON for XHR.
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		writeJSON(w, http.StatusOK, map[string]string{"status": "processing"})
		return
	}
	//nolint:gosec // G710: ref from path after session lookup
	http.Redirect(w, r, "/c/"+ref, http.StatusSeeOther)
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
