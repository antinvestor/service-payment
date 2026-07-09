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
	Contacts    []PayerContactInput
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

// PayInput carries the fields submitted by the payer on the payment form.
type PayInput struct {
	MethodKey   string
	PhoneNumber string // guest payers
	ContactID   string // recognised payer: which prefill contact to charge
	Amount      string // VARIABLE sessions only
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
//nolint:gocognit,nestif // Profile enrichment requires sequential nil-guard checks; extracting further would reduce clarity.
func (b *CheckoutBusiness) applyPayer(
	ctx context.Context,
	session *models.CheckoutSession,
	payer *PayerInput,
) {
	session.PayerProfileID = payer.ProfileID

	displayName := payer.DisplayName
	language := payer.Language
	var clueProvider, clueContactID string
	var profileContacts []*profilev1.ContactObject

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
			clueProvider = clues.LastProvider
			clueContactID = clues.LastContactID
			profileContacts = profile.GetContacts()
		}
	}

	// Build contacts list: caller first; fall back to profile MSISDN contacts
	var contacts []any
	if len(payer.Contacts) > 0 {
		for _, c := range payer.Contacts {
			contacts = append(contacts, map[string]any{
				"contactId": c.ContactID,
				"msisdn":    c.Msisdn,
			})
		}
	} else {
		for _, c := range profileContacts {
			if c.GetType() == profilev1.ContactType_MSISDN {
				contacts = append(contacts, map[string]any{
					"contactId": c.GetId(),
					"msisdn":    c.GetDetail(),
				})
			}
		}
	}

	prefill := data.JSONMap{
		"displayName":   displayName,
		"language":      language,
		"clueProvider":  clueProvider,
		"clueContactId": clueContactID,
		"contacts":      contacts,
	}
	session.Prefill = prefill
}

// ---------------------------------------------------------------------------
// GetSessionByRef
// ---------------------------------------------------------------------------

// GetSessionByRef retrieves a session by ref, flipping status to expired if needed.
func (b *CheckoutBusiness) GetSessionByRef(
	ctx context.Context,
	ref string,
) (*models.CheckoutSession, error) {
	session, err := b.sessionRepo.GetByRef(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("get session by ref: %w", err)
	}

	// Expire sessions that are pending/processing and past expiry
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

	// Resolve msisdn and contactRef
	msisdn, contactID, payerErr := b.resolvePayer(session, in)
	if payerErr != nil {
		b.obs.RecordPayFailure(ctx, "contact_required")
		err = payerErr
		return nil, err
	}

	// All guards passed — record the attempt.
	b.obs.RecordPayAttempt(ctx, in.MethodKey, session.AmountOption == models.AmountOptionVariable)

	// Record attempt BEFORE calling provider so that even a rejected prompt counts
	// toward cooldown and max-attempts throttling.
	now := b.now()
	session.Attempts++
	session.LastAttemptAt = &now

	// Store internal keys for clue write-back
	if session.Metadata == nil {
		session.Metadata = make(data.JSONMap)
	}
	session.Metadata["_method"] = in.MethodKey
	if contactID != "" {
		session.Metadata["_contact_id"] = contactID
	}

	if _, updateErr := b.sessionRepo.Update(ctx, session); updateErr != nil {
		err = fmt.Errorf("persist attempt increment: %w", updateErr)
		return nil, err
	}

	// Build and send prompt; on failure the attempt is already persisted.
	promptID, promptErr := b.sendPrompt(ctx, session, method, msisdn, in.MethodKey)
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
func (b *CheckoutBusiness) sendPrompt(
	ctx context.Context,
	session *models.CheckoutSession,
	method Method,
	msisdn string,
	methodKey string,
) (string, error) {
	extraMap := map[string]any{
		"session_ref": session.Ref,
		"order_ref":   session.OrderRef,
		"provider":    methodKey,
	}
	for k, v := range session.Metadata {
		if len(k) > 0 && k[0] == '_' {
			continue
		}
		if vStr, isStr := v.(string); isStr {
			extraMap["meta_"+k] = vStr
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

// resolvePayer returns (msisdn, contactID) from the PayInput.
// contactID is non-empty only for recognized payers.
func (b *CheckoutBusiness) resolvePayer(
	session *models.CheckoutSession,
	in PayInput,
) (string, string, error) {
	// Recognized payer
	if session.PayerProfileID != "" && in.ContactID != "" {
		msisdn := b.findMsisdnFromPrefill(session, in.ContactID)
		if msisdn == "" {
			msisdn = strings.TrimSpace(in.PhoneNumber) // fallback
		}
		if msisdn == "" {
			return "", "", fmt.Errorf("%w: contact not found in prefill", ErrContactRequired)
		}
		return msisdn, in.ContactID, nil
	}

	// Guest payer
	if in.PhoneNumber == "" {
		return "", "", ErrContactRequired
	}
	return in.PhoneNumber, "", nil
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
func (b *CheckoutBusiness) RefreshStatus(
	ctx context.Context,
	session *models.CheckoutSession,
) (*models.CheckoutSession, error) {
	if session.Status != models.SessionStatusProcessing || session.PromptID == "" {
		return session, nil
	}

	resp, err := b.paymentCli.Status(
		ctx,
		connect.NewRequest(&commonv1.StatusRequest{Id: session.PromptID}),
	)
	if err != nil {
		util.Log(ctx).
			WithError(err).
			Warn("payment status poll transport error — staying processing")
		return session, nil
	}

	status := resp.Msg.GetStatus()
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

// writeClues persists checkout clues to the payer's profile (best-effort).
func (b *CheckoutBusiness) writeClues(ctx context.Context, session *models.CheckoutSession) {
	if session.PayerProfileID == "" {
		return
	}
	// Only write when a linked contact was used
	contactID := ""
	if session.Metadata != nil {
		if cid, ok := session.Metadata["_contact_id"].(string); ok {
			contactID = cid
		}
	}
	if contactID == "" {
		return
	}

	methodKey := ""
	if session.Metadata != nil {
		if mk, ok := session.Metadata["_method"].(string); ok {
			methodKey = mk
		}
	}

	clues := Clues{
		LastMethod:    methodCategoryMobileMoney,
		LastProvider:  methodKey,
		LastContactID: contactID,
		LastCurrency:  session.Currency,
		LastPaidAt:    b.now().Format(time.RFC3339),
	}

	_, err := b.profileCli.Update(ctx, connect.NewRequest(&profilev1.UpdateRequest{
		Id:         session.PayerProfileID,
		Properties: clues.ToProperties(),
	}))
	if err != nil {
		util.Log(ctx).WithError(err).Warn("could not write checkout clues to profile")
		b.obs.RecordClueWritebackFailure(ctx)
	}
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
