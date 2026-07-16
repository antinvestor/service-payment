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
	"fmt"
	"net/url"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/checkout/config"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/antinvestor/service-payments/apps/checkout/service/observability"
	"github.com/antinvestor/service-payments/apps/checkout/service/repository"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/workerpool"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

// Sentinel errors.
var (
	ErrSessionGone     = errors.New("checkout session is no longer payable")
	ErrTooManyAttempts = errors.New("too many payment attempts for this session")
	ErrCooldown        = errors.New("please wait before retrying")
	ErrUnknownMethod   = errors.New("unknown payment method")
	ErrLinkUnusable    = errors.New("checkout link is not usable")
	ErrAmountRequired  = errors.New("an amount is required for this payment")
	ErrContactRequired = errors.New("a phone contact is required to pay")
)

// Ref length constants.
const (
	sessionRefLen       = 32
	linkRefLen          = 12
	sweepBatch          = 50
	clueWriteTimeoutSec = 10 // detached goroutine budget for profile clue write-back
)

// IsSafeReturnURL returns true when rawURL is a non-empty URL with scheme http
// or https (case-insensitive).  Empty string is NOT safe (caller decides).
func IsSafeReturnURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == "http" || scheme == "https"
}

// IsExternalAuthRedirect returns true when rawURL is a safe http(s) URL that is
// NOT our own hosted checkout (pay.*). Providers sometimes echo the success_url
// we sent them as next_action.redirect_url even after the charge already
// succeeded; treating that as a 3DS redirect loops the confirm page forever.
func IsExternalAuthRedirect(rawURL, publicBaseURL string) bool {
	if !IsSafeReturnURL(rawURL) {
		return false
	}
	if strings.TrimSpace(publicBaseURL) == "" {
		return true
	}
	candidate, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"))
	if err != nil || base.Host == "" {
		return true
	}
	// Same host as pay.stawi.org (optionally with /c/... path) is our session, not a bank ACS.
	return !strings.EqualFold(candidate.Host, base.Host)
}

// methodCategoryMobileMoney is the clue category written to the payer profile on
// a successful payment.  The registry models individual providers (e.g. "mpesa",
// "mtn_momo"), not categories, so the category is hard-coded here today.
const methodCategoryMobileMoney = "mobile_money"

// Input types.

// PayerContactInput is one phone contact for a known payer.
type PayerContactInput struct {
	ContactID string
	Msisdn    string
}

// PayerInput carries caller-provided payer information used during session creation.
type PayerInput struct {
	ProfileID   string
	DisplayName string
	Language    string
	// Email is used when the profile has no EMAIL contact / property yet.
	Email    string
	Contacts []PayerContactInput
}

// CreateSessionInput carries all fields needed to create a CheckoutSession.
type CreateSessionInput struct {
	Name         string
	Description  string
	Amount       string // decimal; may be "" for variable
	Currency     string
	AmountOption string // models.AmountOptionFixed|Variable
	OrderRef     string
	Metadata     map[string]string
	ReturnURL    string
	Payer        *PayerInput
	Methods      []string // restriction list; empty = all
}

// CreateLinkInput carries all fields needed to create a CheckoutLink.
type CreateLinkInput struct {
	Name         string
	Description  string
	Amount       string
	Currency     string
	AmountOption string
	OrderRef     string
	ReturnURL    string
	Metadata     map[string]string
	ExpiresAt    *time.Time
}

// EncryptedCardInput is AES-256-GCM card material produced in the browser.
// Raw PAN never touches our servers when the encryption key is served to the client.
type EncryptedCardInput struct {
	EncryptedCardNumber  string
	EncryptedExpiryMonth string
	EncryptedExpiryYear  string
	EncryptedCVV         string
	Nonce                string
}

// PayInput carries the fields submitted by the payer on the payment form.
type PayInput struct {
	MethodKey   string
	PhoneNumber string // guest payers
	ContactID   string // recognised payer: which prefill contact to charge
	Amount      string // VARIABLE sessions only
	// Email / name only when profile did not already store them.
	Email string
	Name  string
	// Encrypted card (embedded checkout).
	Card *EncryptedCardInput
	// Saved instrument for Link-style one-click / subscription renewals.
	PaymentMethodID string
	CustomerID      string
	// Optional guest email when not on profile (card requires email for FW).
	GuestEmail string
}

// AuthorizeInput completes PIN / OTP / AVS on a processing charge.
type AuthorizeInput struct {
	// pin | otp | avs
	Type string
	// PIN: either clear (server encrypts) or pre-encrypted.
	PIN          string
	EncryptedPIN string
	Nonce        string
	// OTP
	OTP string
	// AVS
	City       string
	Country    string
	Line1      string
	Line2      string
	PostalCode string
	State      string
}

// ---------------------------------------------------------------------------
// CheckoutBusiness
// ---------------------------------------------------------------------------

// CheckoutBusiness orchestrates the full checkout lifecycle.
type CheckoutBusiness struct {
	cfg         *config.CheckoutConfig
	registry    *MethodRegistry
	sessionRepo repository.SessionRepository
	linkRepo    repository.LinkRepository
	paymentCli  paymentv1connect.PaymentServiceClient
	profileCli  profilev1connect.ProfileServiceClient
	obs         *observability.Metrics
	now         func() time.Time
	cluesSync   bool               // true in tests only: makes writeClues run synchronously
	workMan     workerpool.Manager // nil in tests → falls back to synchronous execution
}

// NewCheckoutBusiness creates a CheckoutBusiness with a real clock.
// workMan is the frame worker-pool manager (pass svc.WorkManager() in production,
// nil in tests — a nil workMan causes the clue write-back to run synchronously).
func NewCheckoutBusiness(
	cfg *config.CheckoutConfig,
	registry *MethodRegistry,
	sessionRepo repository.SessionRepository,
	linkRepo repository.LinkRepository,
	paymentCli paymentv1connect.PaymentServiceClient,
	profileCli profilev1connect.ProfileServiceClient,
	workMan workerpool.Manager,
) *CheckoutBusiness {
	return &CheckoutBusiness{
		cfg:         cfg,
		registry:    registry,
		sessionRepo: sessionRepo,
		linkRepo:    linkRepo,
		paymentCli:  paymentCli,
		profileCli:  profileCli,
		obs:         observability.NewMetrics(),
		now:         time.Now,
		workMan:     workMan,
	}
}

// WithClock returns a copy of b with an injectable clock for tests.
func (b *CheckoutBusiness) WithClock(now func() time.Time) *CheckoutBusiness {
	cp := *b
	cp.now = now
	return &cp
}

// WithSynchronousClues returns a copy of b that runs writeClues synchronously.
// FOR TESTS ONLY — production code always uses the detached goroutine path.
func (b *CheckoutBusiness) WithSynchronousClues() *CheckoutBusiness {
	cp := *b
	cp.cluesSync = true
	return &cp
}

// ---------------------------------------------------------------------------
// CreateSession
// ---------------------------------------------------------------------------

// CreateSession validates input, optionally fetches a profile for prefill, and
// persists a new pending CheckoutSession.
func (b *CheckoutBusiness) CreateSession(
	ctx context.Context,
	in CreateSessionInput,
) (*models.CheckoutSession, error) {
	ctx, span := b.obs.StartSpan(ctx, "CreateSession")
	var err error
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	// Normalise defaults before validation so the stored session always carries
	// the resolved value (a local default inside the validator would not survive
	// back to this scope and would leave the session with an empty AmountOption).
	if in.AmountOption == "" {
		in.AmountOption = models.AmountOptionFixed
	}

	if err = b.validateSessionInput(in); err != nil {
		return nil, err
	}

	session := &models.CheckoutSession{
		Ref:          util.RandomAlphaNumericString(sessionRefLen),
		Name:         in.Name,
		Description:  in.Description,
		Amount:       in.Amount,
		Currency:     in.Currency,
		AmountOption: in.AmountOption,
		OrderRef:     in.OrderRef,
		ReturnURL:    in.ReturnURL,
		Status:       models.SessionStatusPending,
		ExpiresAt:    b.now().Add(time.Duration(b.cfg.SessionTTLMinutes) * time.Minute),
	}

	// Copy metadata
	if len(in.Metadata) > 0 {
		m := make(data.JSONMap, len(in.Metadata))
		for k, v := range in.Metadata {
			m[k] = v
		}
		session.Metadata = m
	}

	// Methods restriction
	if len(in.Methods) > 0 {
		keys := make([]any, len(in.Methods))
		for i, k := range in.Methods {
			keys[i] = k
		}
		session.Methods = data.JSONMap{"keys": keys}
	}

	// Payer prefill
	if in.Payer != nil {
		b.applyPayer(ctx, session, in.Payer)
	}

	if err = b.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	b.obs.RecordSessionCreated(ctx, in.AmountOption, in.Payer != nil)
	return session, nil
}

// validateSessionInput checks the required fields on CreateSessionInput.
// AmountOption is normalised by the caller (CreateSession) before this is invoked.
func (b *CheckoutBusiness) validateSessionInput(in CreateSessionInput) error {
	if in.Name == "" {
		return errors.New("session name is required")
	}
	if in.AmountOption == models.AmountOptionFixed {
		if _, _, err := ParseAmount(in.Amount); err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}
	}
	if in.Currency == "" || len(in.Currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}
	if in.ReturnURL != "" && !IsSafeReturnURL(in.ReturnURL) {
		return errors.New("return_url must use http or https scheme")
	}
	for _, key := range in.Methods {
		if _, ok := b.registry.Get(key); !ok {
			return fmt.Errorf("%w: %s", ErrUnknownMethod, key)
		}
	}
	return nil
}

// applyPayer enriches a session with prefill data from the payer + optional profile.
//
// Prefill is the source of truth for checkout UI and provider prompts (Flutterwave, etc.):
// display name, email, MSISDN contacts, last-used method, and country — so the payer
// does not re-type identity they already have on their profile.
//
//nolint:gocognit,nestif // Profile enrichment requires sequential nil-guard checks; extracting further would reduce clarity.
func (b *CheckoutBusiness) applyPayer(
	ctx context.Context,
	session *models.CheckoutSession,
	payer *PayerInput,
) {
	session.PayerProfileID = payer.ProfileID

	displayName := payer.DisplayName
	language := payer.Language
	var clueProvider, clueMethod, clueContactID, clueEmailContactID, cluePhoneContactID, clueCountry string
	var paymentMethodID, providerCustomerID string
	var profileContacts []*profilev1.ContactObject
	email := strings.TrimSpace(payer.Email)

	// Fetch profile if ID provided
	if payer.ProfileID != "" && b.profileCli != nil {
		resp, err := b.profileCli.GetById(
			ctx,
			connect.NewRequest(&profilev1.GetByIdRequest{Id: payer.ProfileID}),
		)
		if err != nil {
			util.Log(ctx).
				WithError(err).
				Warn("could not fetch profile for prefill — continuing without it")
		} else if resp.Msg.GetData() != nil {
			profile := resp.Msg.GetData()
			// Fallback display name from "name" property
			if displayName == "" {
				if props := profile.GetProperties(); props != nil {
					if v, ok := props.GetFields()["name"]; ok {
						displayName = v.GetStringValue()
					}
				}
			}
			clues := CluesFromProperties(profile.GetProperties())
			// lastProvider is the registry key (e.g. mpesa); lastMethod may be a category.
			clueProvider = clues.LastProvider
			if clueProvider == "" {
				clueProvider = clues.LastMethod
			}
			clueMethod = clues.LastMethod
			clueContactID = clues.LastContactID
			clueEmailContactID = clues.LastEmailContactID
			cluePhoneContactID = clues.LastPhoneContactID
			if cluePhoneContactID == "" {
				cluePhoneContactID = clueContactID
			}
			clueCountry = clues.LastCountry
			paymentMethodID = clues.PaymentMethodID
			providerCustomerID = clues.ProviderCustomerID
			profileContacts = profile.GetContacts()
		}
	}

	// Profiles have multiple contacts (emails + phones). Classify carefully and
	// prefer previously successful payment contacts when present.
	// ContactType_EMAIL is protobuf zero (0) — shape corrects mislabels.
	emailPick := pickEmailFromCallerContacts(payer.Contacts, clueEmailContactID)
	if email == "" {
		email = emailPick.Detail
	}
	if emailPick.ContactID == "" || email == "" {
		emailPick = pickEmailFromContacts(profileContacts, clueEmailContactID)
		if email == "" {
			email = emailPick.Detail
		}
	}
	if clueEmailContactID == "" && emailPick.ContactID != "" {
		clueEmailContactID = emailPick.ContactID
	}

	// Phone chips + default MSISDN: prefer last successful payment phone contact.
	var phoneMaps []map[string]any
	if len(payer.Contacts) > 0 {
		phoneMaps = normalizeCallerPhoneContacts(payer.Contacts, cluePhoneContactID)
	}
	if len(phoneMaps) == 0 {
		phoneMaps = phoneContactsFromProfile(profileContacts, cluePhoneContactID)
	}
	// Preferred phone detail for locality / masked display.
	phone := ""
	phoneContactID := cluePhoneContactID
	if len(phoneMaps) > 0 {
		if phoneContactID == "" {
			if id, _ := phoneMaps[0]["contactId"].(string); id != "" {
				phoneContactID = id
			}
		}
		// Ensure preferred flag on the chosen phone when we have an id.
		for i := range phoneMaps {
			id, _ := phoneMaps[i]["contactId"].(string)
			if id != "" && id == phoneContactID {
				phoneMaps[i]["preferred"] = true
				if m, _ := phoneMaps[i]["msisdn"].(string); m != "" {
					phone = m
				}
				// Move preferred to front.
				if i > 0 {
					phoneMaps[0], phoneMaps[i] = phoneMaps[i], phoneMaps[0]
				}
				break
			}
		}
		if phone == "" {
			phone, _ = phoneMaps[0]["msisdn"].(string)
			phoneContactID, _ = phoneMaps[0]["contactId"].(string)
		}
	}
	if cluePhoneContactID == "" {
		cluePhoneContactID = phoneContactID
	}
	if clueContactID == "" {
		clueContactID = cluePhoneContactID
	}

	contacts := make([]any, 0, len(phoneMaps))
	for _, m := range phoneMaps {
		contacts = append(contacts, m)
	}

	prefill := data.JSONMap{
		"displayName":        displayName,
		"language":           language,
		"clueProvider":       clueProvider,
		"clueMethod":         clueMethod,
		"clueContactId":      clueContactID,
		"clueEmailContactId": clueEmailContactID,
		"cluePhoneContactId": cluePhoneContactID,
		"country":            clueCountry,
		"email":              email,
		"emailContactId":     emailPick.ContactID,
		"phone":              phone,
		"contacts":           contacts,
	}
	if emailPick.ContactID != "" {
		prefill["emailContactId"] = emailPick.ContactID
	}
	if paymentMethodID != "" {
		prefill["paymentMethodId"] = paymentMethodID
	}
	if providerCustomerID != "" {
		prefill["providerCustomerId"] = providerCustomerID
	}
	session.Prefill = prefill
}

// ---------------------------------------------------------------------------
// GetSessionByRef
// ---------------------------------------------------------------------------

// GetSessionByRef retrieves a session by ref, flipping status to expired if needed.
// Processing sessions that already succeeded at the payment provider are completed
// first (never expire a paid charge as "link expired"). Already-expired sessions
// with a PromptID are recovered the same way so users who paid while stuck on
// confirm still land on the success page.
func (b *CheckoutBusiness) GetSessionByRef(
	ctx context.Context,
	ref string,
) (*models.CheckoutSession, error) {
	session, err := b.sessionRepo.GetByRef(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("get session by ref: %w", err)
	}

	// Recover paid-but-stuck / paid-but-expired sessions before showing "gone".
	if session.PromptID != "" &&
		(session.Status == models.SessionStatusProcessing ||
			session.Status == models.SessionStatusExpired) {
		if refreshed, refreshErr := b.RefreshStatus(ctx, session); refreshErr == nil {
			session = refreshed
			if session.Status == models.SessionStatusCompleted {
				return session, nil
			}
		}
	}

	// Expire unpaid sessions that are pending/processing and past expiry.
	if (session.Status == models.SessionStatusPending || session.Status == models.SessionStatusProcessing) &&
		b.now().After(session.ExpiresAt) {
		session.Status = models.SessionStatusExpired
		if _, updateErr := b.sessionRepo.Update(ctx, session); updateErr != nil {
			util.Log(ctx).WithError(updateErr).Warn("could not mark session expired")
		}
	}
	return session, nil
}

// ---------------------------------------------------------------------------
// CreateLink
// ---------------------------------------------------------------------------

// CreateLink validates and persists a new reusable CheckoutLink.
func (b *CheckoutBusiness) CreateLink(
	ctx context.Context,
	in CreateLinkInput,
) (*models.CheckoutLink, error) {
	ctx, span := b.obs.StartSpan(ctx, "CreateLink")
	var err error
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	if in.Name == "" {
		return nil, errors.New("link name is required")
	}
	if in.Currency == "" || len(in.Currency) != 3 {
		return nil, errors.New("currency must be a 3-letter code")
	}
	amtOption := in.AmountOption
	if amtOption == "" {
		amtOption = models.AmountOptionFixed
	}
	if amtOption == models.AmountOptionFixed && in.Amount != "" {
		if _, _, parseErr := ParseAmount(in.Amount); parseErr != nil {
			return nil, fmt.Errorf("invalid amount: %w", parseErr)
		}
	}
	if in.ReturnURL != "" && !IsSafeReturnURL(in.ReturnURL) {
		return nil, errors.New("return_url must use http or https scheme")
	}

	link := &models.CheckoutLink{
		Ref:          util.RandomAlphaNumericString(linkRefLen),
		Name:         in.Name,
		Description:  in.Description,
		Amount:       in.Amount,
		Currency:     in.Currency,
		AmountOption: amtOption,
		OrderRef:     in.OrderRef,
		ReturnURL:    in.ReturnURL,
		ExpiresAt:    in.ExpiresAt,
		Active:       true,
	}

	if len(in.Metadata) > 0 {
		m := make(data.JSONMap, len(in.Metadata))
		for k, v := range in.Metadata {
			m[k] = v
		}
		link.Metadata = m
	}

	if err = b.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}
	b.obs.RecordLinkCreated(ctx)
	return link, nil
}

// ---------------------------------------------------------------------------
// SpawnSession
// ---------------------------------------------------------------------------

// SpawnSession creates a CheckoutSession from a reusable CheckoutLink.
func (b *CheckoutBusiness) SpawnSession(
	ctx context.Context,
	linkRef string,
) (*models.CheckoutSession, error) {
	ctx, span := b.obs.StartSpan(ctx, "SpawnSession")
	var err error
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	link, err := b.linkRepo.GetByRef(ctx, linkRef)
	if err != nil {
		return nil, fmt.Errorf("get link by ref: %w", err)
	}
	if !link.IsUsable(b.now()) {
		err = ErrLinkUnusable
		return nil, err
	}

	session := &models.CheckoutSession{
		Ref:          util.RandomAlphaNumericString(sessionRefLen),
		LinkID:       link.GetID(),
		Name:         link.Name,
		Description:  link.Description,
		Amount:       link.Amount,
		Currency:     link.Currency,
		AmountOption: link.AmountOption,
		OrderRef:     link.OrderRef,
		ReturnURL:    link.ReturnURL,
		Status:       models.SessionStatusPending,
		ExpiresAt:    b.now().Add(time.Duration(b.cfg.SessionTTLMinutes) * time.Minute),
		Metadata:     link.Metadata,
	}

	if err = b.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create spawned session: %w", err)
	}
	b.obs.RecordSessionSpawned(ctx)
	return session, nil
}

// ---------------------------------------------------------------------------
// Pay
// ---------------------------------------------------------------------------

// Pay executes a payment attempt against an existing session.
//
//nolint:funlen // Guard chain + observability instrumentation makes this function inherently long; extraction would reduce clarity without reducing complexity.
//nolint:gocognit // Pay orchestrates method resolve, payer, redirect vs prompt.
func (b *CheckoutBusiness) Pay(
	ctx context.Context,
	ref string,
	in PayInput,
) (*models.CheckoutSession, error) {
	ctx, span := b.obs.StartSpan(ctx, "Pay")
	var err error
	payStart := b.now()
	defer func() {
		b.obs.RecordPayLatency(ctx, float64(b.now().Sub(payStart).Milliseconds()), in.MethodKey)
		b.obs.EndSpan(ctx, span, err)
	}()

	session, sessionErr := b.GetSessionByRef(ctx, ref)
	if sessionErr != nil {
		err = sessionErr
		return nil, err
	}

	// Terminal status check.
	// A failed session intentionally remains payable: the page offers "Try again"
	// after a failed prompt, bounded by the MaxAttempts cap.
	if session.Status == models.SessionStatusCompleted ||
		session.Status == models.SessionStatusExpired {
		b.obs.RecordPayFailure(ctx, "session_gone")
		err = ErrSessionGone
		return nil, err
	}

	// Max attempts check
	if session.Attempts >= b.cfg.MaxAttempts {
		b.obs.RecordPayFailure(ctx, "too_many_attempts")
		err = ErrTooManyAttempts
		return nil, err
	}

	// Cooldown check
	cooldown := time.Duration(b.cfg.AttemptCooldownSeconds) * time.Second
	if session.LastAttemptAt != nil && b.now().Before(session.LastAttemptAt.Add(cooldown)) {
		b.obs.RecordPayFailure(ctx, "cooldown")
		err = ErrCooldown
		return nil, err
	}

	// Amount: variable sessions require in.Amount
	if session.AmountOption == models.AmountOptionVariable {
		if in.Amount == "" {
			b.obs.RecordPayFailure(ctx, "amount_required")
			err = ErrAmountRequired
			return nil, err
		}
		if _, _, parseErr := ParseAmount(in.Amount); parseErr != nil {
			b.obs.RecordPayFailure(ctx, "amount_required")
			err = fmt.Errorf("%w: %w", ErrAmountRequired, parseErr)
			return nil, err
		}
		session.Amount = in.Amount
	}

	// Method validation and restriction check
	method, methodErr := b.resolveMethod(session, in.MethodKey)
	if methodErr != nil {
		b.obs.RecordPayFailure(ctx, "unknown_method")
		err = methodErr
		return nil, err
	}

	// Resolve msisdn and contactRef.
	// Embedded card and redirect methods do not require phone — profile contacts
	// are still resolved so we never re-ask for data we already store.
	var msisdn, contactID string
	if !method.Redirect && !method.IsEmbedded() {
		var payerErr error
		msisdn, contactID, payerErr = b.resolvePayer(session, in)
		if payerErr != nil {
			b.obs.RecordPayFailure(ctx, "contact_required")
			err = payerErr
			return nil, err
		}
	} else {
		// Best-effort profile phone for card / redirect rails (optional).
		msisdn, contactID = b.resolvePayerOptional(session, in)
	}

	// Embedded card must include encrypted fields or a saved payment method.
	if method.IsEmbedded() && in.PaymentMethodID == "" {
		if in.Card == nil || strings.TrimSpace(in.Card.EncryptedCardNumber) == "" ||
			strings.TrimSpace(in.Card.Nonce) == "" {
			b.obs.RecordPayFailure(ctx, "card_required")
			err = fmt.Errorf("%w: encrypted card fields required", ErrUnknownMethod)
			return nil, err
		}
	}

	// All guards passed — record the attempt.
	b.obs.RecordPayAttempt(ctx, in.MethodKey, session.AmountOption == models.AmountOptionVariable)

	// Record attempt BEFORE calling provider so that even a rejected prompt counts
	// toward cooldown and max-attempts throttling.
	now := b.now()
	session.Attempts++
	session.LastAttemptAt = &now

	// Store internal keys for clue write-back (preferred payment contacts).
	if session.Metadata == nil {
		session.Metadata = make(data.JSONMap)
	}
	session.Metadata["_method"] = in.MethodKey
	if contactID != "" {
		session.Metadata["_contact_id"] = contactID
		session.Metadata["_phone_contact_id"] = contactID
	}
	// Email contact used for this pay (from prefill preference).
	if session.Prefill != nil {
		if eid, _ := session.Prefill["emailContactId"].(string); eid != "" {
			session.Metadata["_email_contact_id"] = eid
		} else if eid, _ := session.Prefill["clueEmailContactId"].(string); eid != "" {
			session.Metadata["_email_contact_id"] = eid
		}
		// If card path and no phone selected, still remember preferred phone for next time.
		if contactID == "" {
			if pid, _ := session.Prefill["cluePhoneContactId"].(string); pid != "" {
				session.Metadata["_phone_contact_id"] = pid
			}
		}
	}

	if _, updateErr := b.sessionRepo.Update(ctx, session); updateErr != nil {
		err = fmt.Errorf("persist attempt increment: %w", updateErr)
		return nil, err
	}

	// Build and send prompt; on failure the attempt is already persisted.
	promptID, promptErr := b.sendPrompt(ctx, session, method, msisdn, in)
	if promptErr != nil {
		b.obs.RecordPayFailure(ctx, "prompt_error")
		err = promptErr
		return nil, err
	}

	session.PromptID = promptID
	session.Status = models.SessionStatusProcessing

	if _, updateErr := b.sessionRepo.Update(ctx, session); updateErr != nil {
		err = fmt.Errorf("persist prompt id: %w", updateErr)
		return nil, err
	}
	return session, nil
}

// resolveMethod validates that the method key exists and is allowed by the session.
func (b *CheckoutBusiness) resolveMethod(
	session *models.CheckoutSession,
	methodKey string,
) (Method, error) {
	method, ok := b.registry.Get(methodKey)
	if !ok {
		return Method{}, fmt.Errorf("%w: %s", ErrUnknownMethod, methodKey)
	}
	if err := b.checkMethodRestriction(session, methodKey); err != nil {
		return Method{}, err
	}
	return method, nil
}

// checkMethodRestriction returns ErrUnknownMethod when the session restricts methods and
// the key is not in the restriction list.
func (b *CheckoutBusiness) checkMethodRestriction(
	session *models.CheckoutSession,
	methodKey string,
) error {
	if session.Methods == nil {
		return nil
	}
	keysRaw, hasKeys := session.Methods["keys"]
	if !hasKeys {
		return nil
	}
	keysList, isList := keysRaw.([]any)
	if !isList {
		return nil
	}
	for _, k := range keysList {
		if ks, isStr := k.(string); isStr && ks == methodKey {
			return nil
		}
	}
	return fmt.Errorf("%w: %s not in session restriction", ErrUnknownMethod, methodKey)
}

// sendPrompt builds the prompt request, calls the payment service, and returns the promptID.
//
// Prefill identity (email, name) and return URLs are forwarded in Extra so provider
// integrations (Flutterwave v4) can build the customer object without re-prompting.
// Encrypted card fields are portable keys so the payment route can switch PSPs.
//
//nolint:gocognit,funlen // Builds prompt extras from prefill, metadata, card, and success URL.
func (b *CheckoutBusiness) sendPrompt(
	ctx context.Context,
	session *models.CheckoutSession,
	method Method,
	msisdn string,
	in PayInput,
) (string, error) {
	methodKey := in.MethodKey
	extraMap := map[string]any{
		"session_ref": session.Ref,
		"order_ref":   session.OrderRef,
		"provider":    methodKey,
	}
	if method.IsEmbedded() || strings.EqualFold(methodKey, "card") {
		extraMap["payment_method_type"] = "card"
	}
	// Provider return after 3DS / hosted: land back on our checkout session page.
	if b.cfg != nil && b.cfg.PublicBaseURL != "" && session.Ref != "" {
		successURL := strings.TrimRight(b.cfg.PublicBaseURL, "/") + "/c/" + session.Ref
		extraMap["success_url"] = successURL
		extraMap["redirect_url"] = successURL
	}
	// Profile / prefill identity — only fill gaps from form input.
	email := strings.TrimSpace(in.Email)
	if email == "" {
		email = strings.TrimSpace(in.GuestEmail)
	}
	name := strings.TrimSpace(in.Name)
	if session.Prefill != nil {
		if email == "" {
			if v, _ := session.Prefill["email"].(string); strings.TrimSpace(v) != "" {
				email = strings.TrimSpace(v)
			}
		}
		if name == "" {
			if v, _ := session.Prefill["displayName"].(string); strings.TrimSpace(v) != "" {
				name = strings.TrimSpace(v)
			}
		}
	}
	if email != "" {
		extraMap["customer_email"] = email
		extraMap["email"] = email
	}
	if name != "" {
		extraMap["customer_name"] = name
		extraMap["display_name"] = name
	}
	// Embedded encrypted card (never clear PAN).
	if in.Card != nil {
		extraMap["encrypted_card_number"] = in.Card.EncryptedCardNumber
		extraMap["encrypted_expiry_month"] = in.Card.EncryptedExpiryMonth
		extraMap["encrypted_expiry_year"] = in.Card.EncryptedExpiryYear
		extraMap["encrypted_cvv"] = in.Card.EncryptedCVV
		extraMap["card_nonce"] = in.Card.Nonce
		extraMap["nonce"] = in.Card.Nonce
	}
	// Tokenized / Link-style saved card (subscriptions + returning payers).
	if strings.TrimSpace(in.PaymentMethodID) != "" {
		extraMap["payment_method_id"] = strings.TrimSpace(in.PaymentMethodID)
	}
	if strings.TrimSpace(in.CustomerID) != "" {
		extraMap["customer_id"] = strings.TrimSpace(in.CustomerID)
	}
	// Saved instrument from profile clues when form did not send one.
	if session.Prefill != nil && extraMap["payment_method_id"] == nil {
		if pmd, _ := session.Prefill["paymentMethodId"].(string); strings.TrimSpace(pmd) != "" {
			extraMap["payment_method_id"] = strings.TrimSpace(pmd)
			if cus, _ := session.Prefill["providerCustomerId"].(string); strings.TrimSpace(cus) != "" {
				extraMap["customer_id"] = strings.TrimSpace(cus)
			}
		}
	}
	if session.OrderRef != "" {
		extraMap["invoice_id"] = session.OrderRef
	}
	for k, v := range session.Metadata {
		if len(k) > 0 && k[0] == '_' {
			continue
		}
		if vStr, isStr := v.(string); isStr {
			extraMap["meta_"+k] = vStr
			switch k {
			case "invoiceId", "invoice_id":
				extraMap["invoice_id"] = vStr
			case "subscriptionId", "subscription_id":
				extraMap["subscription_id"] = vStr
			case "source":
				extraMap["collection_source"] = vStr
			case "paymentMethodId", "payment_method_id":
				if extraMap["payment_method_id"] == nil {
					extraMap["payment_method_id"] = vStr
				}
			case "customerId", "customer_id", "providerCustomerId":
				if extraMap["customer_id"] == nil {
					extraMap["customer_id"] = vStr
				}
			}
		}
	}

	extra, extraErr := structpb.NewStruct(extraMap)
	if extraErr != nil {
		return "", fmt.Errorf("build extra struct: %w", extraErr)
	}

	money, moneyErr := MoneyFromAmount(session.Amount, session.Currency)
	if moneyErr != nil {
		return "", fmt.Errorf("build money: %w", moneyErr)
	}

	promptReq := &paymentv1.InitiatePromptRequest{
		Source: &commonv1.ContactLink{
			ProfileId: session.PayerProfileID,
			ContactId: msisdn,
		},
		Recipient: &commonv1.ContactLink{
			ProfileId: session.PayerProfileID,
			ContactId: msisdn,
		},
		Amount: money,
		Route:  method.Route,
		Extra:  extra,
	}

	resp, respErr := b.paymentCli.InitiatePrompt(ctx, connect.NewRequest(promptReq))
	if respErr != nil {
		return "", fmt.Errorf("initiate prompt: %w", respErr)
	}
	return resp.Msg.GetData().GetId(), nil
}

// Authorize completes PIN / OTP / AVS for a processing session by re-using the
// same payment route with action=authorize (provider-portable extras).
func (b *CheckoutBusiness) Authorize(
	ctx context.Context,
	ref string,
	in AuthorizeInput,
) (*models.CheckoutSession, error) {
	session, err := b.GetSessionByRef(ctx, ref)
	if err != nil {
		return nil, err
	}
	if session.Status != models.SessionStatusProcessing {
		return nil, ErrSessionGone
	}
	chargeID := ""
	methodKey := "card"
	route := "flutterwave"
	if session.Metadata != nil {
		if v, _ := session.Metadata["_charge_id"].(string); v != "" {
			chargeID = v
		}
		if v, _ := session.Metadata["_method"].(string); v != "" {
			methodKey = v
		}
	}
	if m, ok := b.registry.Get(methodKey); ok {
		route = m.Route
	}
	if chargeID == "" {
		// Fall back to payment external id = charge id after StatusUpdate.
		chargeID = session.PaymentID
	}
	if chargeID == "" {
		return nil, fmt.Errorf("charge id not ready for authorization")
	}

	extraMap := map[string]any{
		"action":             "authorize",
		"authorization_type": strings.ToLower(strings.TrimSpace(in.Type)),
		"charge_id":          chargeID,
		"session_ref":        session.Ref,
		"provider":           methodKey,
	}
	switch strings.ToLower(in.Type) {
	case "pin":
		if in.EncryptedPIN != "" {
			extraMap["encrypted_pin"] = in.EncryptedPIN
			extraMap["nonce"] = in.Nonce
		} else {
			extraMap["pin"] = in.PIN
		}
	case "otp":
		extraMap["otp"] = in.OTP
	case "avs":
		extraMap["avs_city"] = in.City
		extraMap["avs_country"] = in.Country
		extraMap["avs_line1"] = in.Line1
		extraMap["avs_line2"] = in.Line2
		extraMap["avs_postal_code"] = in.PostalCode
		extraMap["avs_state"] = in.State
	default:
		return nil, fmt.Errorf("%w: authorization type", ErrUnknownMethod)
	}
	extra, err := structpb.NewStruct(extraMap)
	if err != nil {
		return nil, fmt.Errorf("build authorize extras: %w", err)
	}
	money, err := MoneyFromAmount(session.Amount, session.Currency)
	if err != nil {
		return nil, err
	}
	// Same prompt id so status updates continue on this session.
	promptReq := &paymentv1.InitiatePromptRequest{
		Id: session.PromptID,
		Source: &commonv1.ContactLink{
			ProfileId: session.PayerProfileID,
		},
		Recipient: &commonv1.ContactLink{
			ProfileId: session.PayerProfileID,
		},
		Amount: money,
		Route:  route,
		Extra:  extra,
	}
	if _, err := b.paymentCli.InitiatePrompt(ctx, connect.NewRequest(promptReq)); err != nil {
		return nil, fmt.Errorf("authorize prompt: %w", err)
	}
	return session, nil
}

// resolvePayer returns (msisdn, contactID) from the PayInput.
// contactID is non-empty only for recognized payers.
func (b *CheckoutBusiness) resolvePayer(
	session *models.CheckoutSession,
	in PayInput,
) (string, string, error) {
	msisdn, contactID := b.resolvePayerOptional(session, in)
	if msisdn == "" {
		return "", "", ErrContactRequired
	}
	return msisdn, contactID, nil
}

// resolvePayerOptional is like resolvePayer but never errors — used for redirect
// methods where phone is helpful but not mandatory.
func (b *CheckoutBusiness) resolvePayerOptional(
	session *models.CheckoutSession,
	in PayInput,
) (string, string) {
	// Explicit contact from form (recognized payer).
	if session.PayerProfileID != "" && in.ContactID != "" {
		msisdn := b.findMsisdnFromPrefill(session, in.ContactID)
		if msisdn == "" {
			msisdn = strings.TrimSpace(in.PhoneNumber)
		}
		// Caller chose a specific contact: do not silently substitute another.
		return msisdn, in.ContactID
	}

	// Guest typed phone.
	if phone := strings.TrimSpace(in.PhoneNumber); phone != "" {
		return phone, ""
	}

	// Auto-pick from profile prefill only when no contact was selected:
	// clue contact, else first MSISDN (zero re-entry path).
	if session.PayerProfileID != "" && session.Prefill != nil {
		if clueID, _ := session.Prefill["clueContactId"].(string); clueID != "" {
			if msisdn := b.findMsisdnFromPrefill(session, clueID); msisdn != "" {
				return msisdn, clueID
			}
		}
		if msisdn, contactID := b.firstPrefillContact(session); msisdn != "" {
			return msisdn, contactID
		}
	}
	return "", ""
}

// firstPrefillContact returns the first MSISDN contact from session prefill.
func (b *CheckoutBusiness) firstPrefillContact(session *models.CheckoutSession) (msisdn, contactID string) {
	if session.Prefill == nil {
		return "", ""
	}
	contactsRaw, ok := session.Prefill["contacts"]
	if !ok {
		return "", ""
	}
	contacts, ok := contactsRaw.([]any)
	if !ok {
		return "", ""
	}
	for _, raw := range contacts {
		c, isMap := raw.(map[string]any)
		if !isMap {
			continue
		}
		phone, _ := c["msisdn"].(string)
		cid, _ := c["contactId"].(string)
		if strings.TrimSpace(phone) != "" {
			return strings.TrimSpace(phone), cid
		}
	}
	return "", ""
}

// findMsisdnFromPrefill looks up a msisdn for contactID in session prefill contacts.
func (b *CheckoutBusiness) findMsisdnFromPrefill(
	session *models.CheckoutSession,
	contactID string,
) string {
	if session.Prefill == nil {
		return ""
	}
	contactsRaw, ok := session.Prefill["contacts"]
	if !ok {
		return ""
	}
	contacts, ok := contactsRaw.([]any)
	if !ok {
		return ""
	}
	for _, raw := range contacts {
		c, isMap := raw.(map[string]any)
		if !isMap {
			continue
		}
		cid, hasCID := c["contactId"].(string)
		if hasCID && cid == contactID {
			if msisdn, hasMSISDN := c["msisdn"].(string); hasMSISDN {
				return msisdn
			}
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// RefreshStatus
// ---------------------------------------------------------------------------

// RefreshStatus polls the payment service for an update on a processing session.
// Also recovers expired sessions that still have a PromptID when the provider
// already reported SUCCESSFUL (paid while the confirm poll was stuck).
func (b *CheckoutBusiness) RefreshStatus(
	ctx context.Context,
	session *models.CheckoutSession,
) (*models.CheckoutSession, error) {
	if session.PromptID == "" {
		return session, nil
	}
	switch session.Status {
	case models.SessionStatusProcessing, models.SessionStatusExpired:
		// continue
	default:
		return session, nil
	}

	// Status rows are keyed by (entity_id, entity_type). InitiatePrompt writes
	// entity_type=prompt; omitting it queries entity_type='' and always 404s
	// ("record not found") even after Flutterwave reports SUCCESSFUL.
	statusExtras, _ := structpb.NewStruct(map[string]any{"entity_type": "prompt"})
	resp, err := b.paymentCli.Status(
		ctx,
		connect.NewRequest(&commonv1.StatusRequest{
			Id:     session.PromptID,
			Extras: statusExtras,
		}),
	)
	if err != nil {
		util.Log(ctx).
			WithError(err).
			Warn("payment status poll transport error — staying processing")
		return session, nil
	}

	status := resp.Msg.GetStatus()

	// Capture provider next steps (3DS URL, PIN/OTP, charge/token ids) while processing.
	if status != commonv1.STATUS_SUCCESSFUL && status != commonv1.STATUS_FAILED {
		b.captureProviderExtras(ctx, session, resp.Msg.GetExtras())
	} else if status == commonv1.STATUS_SUCCESSFUL {
		// Persist tokenized instrument for Link-style reuse / subscription renewals.
		b.captureProviderExtras(ctx, session, resp.Msg.GetExtras())
	}

	//nolint:exhaustive // Only terminal statuses need action; all others leave session unchanged.
	switch status {
	case commonv1.STATUS_SUCCESSFUL:
		session.Status = models.SessionStatusCompleted
		session.PaymentID = resp.Msg.GetId()
		if _, updateErr := b.sessionRepo.Update(ctx, session); updateErr != nil {
			return nil, fmt.Errorf("update session completed: %w", updateErr)
		}
		b.obs.RecordOutcomeCompleted(ctx)
		// writeClues is best-effort hint persistence. Loss is tolerable, so the
		// frame workerpool (not a durable queue) is the right tier of the async
		// decision tree: bounded parallel, survives nothing. Raw goroutines are
		// forbidden by repo Go patterns; when workMan is nil (tests) or cluesSync
		// is set (WithSynchronousClues), we fall back to synchronous execution.
		if b.cluesSync || b.workMan == nil {
			// Synchronous path: FOR TESTS ONLY (cluesSync via WithSynchronousClues,
			// or nil workMan when constructed without one).
			b.writeClues(ctx, session)
		} else {
			snapshot := *session // shallow copy — capture values before submission to avoid racing on caller's pointer
			job := workerpool.NewJob(func(jobCtx context.Context, _ workerpool.JobResultPipe[any]) error {
				clueCtx, cancel := context.WithTimeout(context.WithoutCancel(jobCtx), clueWriteTimeoutSec*time.Second)
				defer cancel()
				b.writeClues(clueCtx, &snapshot)
				return nil
			})
			if submitErr := workerpool.SubmitJob(ctx, b.workMan, job); submitErr != nil {
				util.Log(ctx).WithError(submitErr).Warn("could not submit clue write-back job")
			}
		}

	case commonv1.STATUS_FAILED:
		session.Status = models.SessionStatusFailed
		if _, updateErr := b.sessionRepo.Update(ctx, session); updateErr != nil {
			return nil, fmt.Errorf("update session failed: %w", updateErr)
		}
		b.obs.RecordOutcomeFailed(ctx)
	}
	return session, nil
}

// captureProviderExtras persists portable provider fields on the session:
// 3DS redirect, next_action type (pin/otp/avs), charge_id, payment_method_id.
func (b *CheckoutBusiness) captureProviderExtras(
	ctx context.Context,
	session *models.CheckoutSession,
	extras *structpb.Struct,
) {
	if extras == nil {
		return
	}
	if session.Metadata == nil {
		session.Metadata = make(data.JSONMap)
	}
	changed := false
	set := func(key, val string) {
		if val == "" {
			return
		}
		if existing, _ := session.Metadata[key].(string); existing == val {
			return
		}
		session.Metadata[key] = val
		changed = true
	}
	// 3DS / legacy hosted URL — never store our own pay.* success URL as a
	// redirect; that freezes the confirm spinner in a self-navigation loop.
	publicBase := ""
	if b.cfg != nil {
		publicBase = b.cfg.PublicBaseURL
	}
	if f, ok := extras.GetFields()["checkout_url"]; ok {
		urlStr := f.GetStringValue()
		if IsExternalAuthRedirect(urlStr, publicBase) {
			set("_redirect_url", urlStr)
		}
	}
	if f, ok := extras.GetFields()["auth_redirect_url"]; ok {
		urlStr := f.GetStringValue()
		if IsExternalAuthRedirect(urlStr, publicBase) {
			set("_redirect_url", urlStr)
		}
	}
	// Resolve whether any redirect URL in extras is an external bank/ACS target.
	pairedRedirect := ""
	if f, ok := extras.GetFields()["checkout_url"]; ok {
		pairedRedirect = f.GetStringValue()
	}
	if pairedRedirect == "" {
		if f, ok := extras.GetFields()["auth_redirect_url"]; ok {
			pairedRedirect = f.GetStringValue()
		}
	}
	externalRedirect := IsExternalAuthRedirect(pairedRedirect, publicBase)

	for _, k := range []struct{ from, to string }{
		{"next_action_type", "_next_action"},
		{"next_action", "_next_action"},
		{"charge_id", "_charge_id"},
		{"payment_method_id", "_payment_method_id"},
		{"customer_id", "_customer_id"},
		{"payment_instruction", "_payment_instruction"},
	} {
		if f, ok := extras.GetFields()[k.from]; ok {
			val := f.GetStringValue()
			// Drop redirect_url next_action when it only points at our own checkout.
			if k.to == "_next_action" && val == "redirect_url" && !externalRedirect {
				continue
			}
			set(k.to, val)
		}
	}
	if !changed {
		return
	}
	if _, updateErr := b.sessionRepo.Update(ctx, session); updateErr != nil {
		util.Log(ctx).WithError(updateErr).Warn("could not persist provider extras on session")
	}
}

// writeClues persists checkout clues to the payer's profile (best-effort).
// Always stores the method key so the next visit can preselect like Stripe Link,
// including card/redirect payments that may not have a linked contact.
//
// Also marks preferred payment contacts (phone + email) when the profile has
// multiple contacts so the next checkout reuses them without re-entry.
func (b *CheckoutBusiness) writeClues(ctx context.Context, session *models.CheckoutSession) {
	if session.PayerProfileID == "" || b.profileCli == nil {
		return
	}

	contactID := ""
	phoneContactID := ""
	emailContactID := ""
	methodKey := ""
	if session.Metadata != nil {
		if cid, ok := session.Metadata["_contact_id"].(string); ok {
			contactID = cid
		}
		if pid, ok := session.Metadata["_phone_contact_id"].(string); ok {
			phoneContactID = pid
		}
		if eid, ok := session.Metadata["_email_contact_id"].(string); ok {
			emailContactID = eid
		}
		if mk, ok := session.Metadata["_method"].(string); ok {
			methodKey = mk
		}
	}
	if phoneContactID == "" {
		phoneContactID = contactID
	}
	if methodKey == "" {
		return
	}

	// Category is informational; LastProvider is the registry key used for preselect.
	category := methodCategoryMobileMoney
	if m, ok := b.registry.Get(methodKey); ok && (m.Redirect || m.IsEmbedded()) {
		category = "card"
	}

	// Infer country from the payer MSISDN (not contact ID).
	msisdn := ""
	if phoneContactID != "" {
		msisdn = b.findMsisdnFromPrefill(session, phoneContactID)
	}
	country := InferCountryFromPhone(msisdn)
	if country == "" && session.Prefill != nil {
		if c, _ := session.Prefill["country"].(string); c != "" {
			country = strings.ToUpper(strings.TrimSpace(c))
		}
	}

	clues := Clues{
		LastMethod:         category,
		LastProvider:       methodKey,
		LastContactID:      phoneContactID,
		LastPhoneContactID: phoneContactID,
		LastEmailContactID: emailContactID,
		LastCurrency:       session.Currency,
		LastCountry:        country,
		LastPaidAt:         b.now().Format(time.RFC3339),
	}
	// Tokenized card for Stripe Link-style reuse and subscription renewals.
	if session.Metadata != nil {
		if pmd, _ := session.Metadata["_payment_method_id"].(string); pmd != "" {
			clues.PaymentMethodID = pmd
		}
		if cus, _ := session.Metadata["_customer_id"].(string); cus != "" {
			clues.ProviderCustomerID = cus
		}
	}

	_, err := b.profileCli.Update(ctx, connect.NewRequest(&profilev1.UpdateRequest{
		Id:         session.PayerProfileID,
		Properties: clues.ToProperties(),
	}))
	if err != nil {
		util.Log(ctx).WithError(err).Warn("could not write checkout clues to profile")
		b.obs.RecordClueWritebackFailure(ctx)
		return
	}
	util.Log(ctx).
		WithField("profile_id", session.PayerProfileID).
		WithField("phone_contact_id", phoneContactID).
		WithField("email_contact_id", emailContactID).
		WithField("method", methodKey).
		Debug("wrote preferred payment contacts to profile clues")
}

// ---------------------------------------------------------------------------
// SweepProcessing
// ---------------------------------------------------------------------------

// SweepProcessing polls processing sessions for status and expires stale pending ones.
func (b *CheckoutBusiness) SweepProcessing(ctx context.Context) error {
	ctx, span := b.obs.StartSpan(ctx, "SweepProcessing")
	sweepStart := b.now()
	var firstErr error
	defer func() {
		b.obs.RecordSweepLatency(ctx, float64(b.now().Sub(sweepStart).Milliseconds()))
		b.obs.EndSpan(ctx, span, firstErr)
	}()

	// Refresh processing sessions
	processing, err := b.sessionRepo.ListByStatus(ctx, models.SessionStatusProcessing, sweepBatch)
	if err != nil {
		firstErr = fmt.Errorf("list processing sessions: %w", err)
	} else {
		for _, s := range processing {
			if _, refreshErr := b.RefreshStatus(ctx, s); refreshErr != nil && firstErr == nil {
				firstErr = refreshErr
			}
		}
	}

	// Expire overdue pending sessions
	pending, listErr := b.sessionRepo.ListByStatus(ctx, models.SessionStatusPending, sweepBatch)
	if listErr != nil {
		if firstErr == nil {
			firstErr = fmt.Errorf("list pending sessions: %w", listErr)
		}
		return firstErr
	}
	for _, s := range pending {
		if !b.now().After(s.ExpiresAt) {
			continue
		}
		s.Status = models.SessionStatusExpired
		if _, updateErr := b.sessionRepo.Update(ctx, s); updateErr != nil && firstErr == nil {
			firstErr = updateErr
		} else if updateErr == nil {
			b.obs.RecordSweepExpired(ctx)
		}
	}

	return firstErr
}
