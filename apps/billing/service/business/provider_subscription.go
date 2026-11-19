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
	"strings"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/workerpool"
	"github.com/pitabwire/util"
)

// Provider-managed subscriptions: the provider (Polar, Stripe) charges the
// payer on its own schedule and reports every change by webhook; billing
// mirrors that state, fans lifecycle events out to product systems and
// never raises its own renewal charges for these subscriptions.

// Subscription.Data keys for provider-managed subscriptions. The Polar names
// are kept for rows written before the generalisation.
const (
	PolarDataKeyProductID         = "polarProductId"
	PolarDataKeySubscriptionID    = "polarSubscriptionId"
	PolarDataKeyStatus            = "polarStatus"
	PolarDataKeyCurrentPeriodEnd  = models.SubDataCurrentPeriodEnd
	PolarDataKeyCollectionMode    = models.SubDataCollectionMode
	PolarDataKeyCancelAtPeriodEnd = models.SubDataCancelAtPeriodEnd

	polarCollectionModeValue = "polar"

	// Polar subscription status values (also the normalized set in
	// pkg/collection used by every provider).
	PolarStatusActive   = collection.SubscriptionStatusActive
	PolarStatusCanceled = collection.SubscriptionStatusCanceled
	PolarStatusRevoked  = collection.SubscriptionStatusRevoked
	PolarStatusPastDue  = collection.SubscriptionStatusPastDue
	PolarStatusUnpaid   = collection.SubscriptionStatusUnpaid

	providerPolar = "polar"
)

// ProviderProductFor returns the payment provider and the provider's product
// reference configured on the plan. Plans declare either
// data.paymentProvider + data.providerProductId, or the legacy
// data.polarProductId. Empty provider means billing collects the plan itself.
func ProviderProductFor(plan *models.Plan) (provider, productID string) {
	if plan == nil || plan.Data == nil {
		return "", ""
	}
	if p, _ := plan.Data[models.PlanDataPaymentProvider].(string); strings.TrimSpace(p) != "" {
		id, _ := plan.Data[models.PlanDataProviderProductID].(string)
		if strings.TrimSpace(id) == "" {
			id, _ = plan.Data[PolarDataKeyProductID].(string)
		}
		return strings.ToLower(strings.TrimSpace(p)), strings.TrimSpace(id)
	}
	if id, _ := plan.Data[PolarDataKeyProductID].(string); strings.TrimSpace(id) != "" {
		return providerPolar, strings.TrimSpace(id)
	}
	return "", ""
}

// PolarProductID returns the polar product ID configured on the plan, or "".
func PolarProductID(plan *models.Plan) string {
	p, id := ProviderProductFor(plan)
	if p != providerPolar {
		return ""
	}
	return id
}

// IsProviderCollected reports whether a provider charges this plan.
func IsProviderCollected(plan *models.Plan) bool {
	p, id := ProviderProductFor(plan)
	return p != "" && id != ""
}

// IsPolarCollected reports whether Polar charges this plan.
func IsPolarCollected(plan *models.Plan) bool {
	return PolarProductID(plan) != ""
}

// IsProviderManaged reports whether a subscription row is charged by a provider.
func IsProviderManaged(sub *models.Subscription) bool {
	if sub == nil || sub.Data == nil {
		return false
	}
	mode, _ := sub.Data[models.SubDataCollectionMode].(string)
	return mode == models.CollectionModeProvider || mode == polarCollectionModeValue
}

// SubscriptionProvider returns the provider route charging a subscription.
func SubscriptionProvider(sub *models.Subscription) string {
	if sub == nil || sub.Data == nil {
		return ""
	}
	if p, _ := sub.Data[models.SubDataPaymentProvider].(string); p != "" {
		return p
	}
	if mode, _ := sub.Data[models.SubDataCollectionMode].(string); mode == polarCollectionModeValue {
		return providerPolar
	}
	return ""
}

// ExternalSubscriptionID returns the provider's id for a subscription.
func ExternalSubscriptionID(sub *models.Subscription) string {
	if sub == nil || sub.Data == nil {
		return ""
	}
	if id, _ := sub.Data[models.SubDataExternalSubscriptionID].(string); id != "" {
		return id
	}
	id, _ := sub.Data[PolarDataKeySubscriptionID].(string)
	return id
}

// StartPolarInput carries the parameters needed to open a provider-managed
// subscription.
type StartPolarInput struct {
	ProfileID        string
	PlanID           string
	CatalogVersionID string
	Currency         string
	// Data is merged into Subscription.Data (external entity, metadata).
	Data map[string]any
}

// PolarSubscriptionStart is returned by StartPolarSubscription: the PENDING
// billing Subscription and the provider product the checkout must charge.
type PolarSubscriptionStart struct {
	Subscription   *models.Subscription
	Provider       string
	PolarProductID string
}

// MirrorInput carries provider-reported state to apply to the matching
// billing Subscription.
type MirrorInput struct {
	// Provider route (polar, stripe). Empty means polar for older callers.
	Provider string
	// Event is one of the collection.SubscriptionEvent* names; empty means
	// a status update.
	Event string
	// SubscriptionID is billing's id when the provider metadata carried it.
	SubscriptionID string
	// PolarSubscriptionID is the provider's subscription id.
	PolarSubscriptionID string
	ExternalCustomerID  string
	ExternalProductID   string
	ProfileID           string
	PlanID              string
	// PolarStatus is the normalized provider status.
	PolarStatus        string
	CancelAtPeriodEnd  bool
	CurrentPeriodStart string // RFC3339 or empty
	CurrentPeriodEnd   string // RFC3339 or empty
	// ExternalOrderID identifies the charge for billed events.
	ExternalOrderID string
	// OccurredAt is the provider timestamp; stale events are ignored.
	OccurredAt string
}

// PolarSubscriptionBusiness manages provider-managed subscriptions.
type PolarSubscriptionBusiness interface {
	// StartPolarSubscription returns the PENDING subscription for a provider
	// plan, reusing an open one for the same profile and plan.
	StartPolarSubscription(ctx context.Context, plan *models.Plan, in StartPolarInput) (*PolarSubscriptionStart, error)

	// MirrorSubscriptionState applies provider state to the billing
	// Subscription and fans lifecycle events out. It is idempotent and
	// ignores events older than the last one applied.
	MirrorSubscriptionState(ctx context.Context, in MirrorInput) (*models.Subscription, error)
}

// ErrStaleProviderEvent marks an event older than the state already applied.
var ErrStaleProviderEvent = errors.New("provider event is older than the applied state")

// ErrProviderSubscriptionMismatch is returned when a provider event names a
// billing subscription that belongs to another provider, product or
// provider subscription: webhook metadata can never re-point a row.
var ErrProviderSubscriptionMismatch = errors.New("provider event does not match the billing subscription it names")

type polarSubscriptionBusiness struct {
	workMan  workerpool.Manager
	subRepo  repository.SubscriptionRepository
	notifier SubscriptionLifecycleNotifier
	obs      *observability.Metrics
}

// NewPolarSubscriptionBusiness constructs a PolarSubscriptionBusiness. The
// optional notifier fans lifecycle events out to product systems.
func NewPolarSubscriptionBusiness(
	workMan workerpool.Manager,
	subRepo repository.SubscriptionRepository,
	optional ...SubscriptionLifecycleNotifier,
) PolarSubscriptionBusiness {
	notifier := NoopSubscriptionLifecycleNotifier()
	for _, n := range optional {
		if n != nil {
			notifier = n
		}
	}
	return &polarSubscriptionBusiness{
		workMan:  workMan,
		subRepo:  subRepo,
		notifier: notifier,
		obs:      observability.NewMetrics(),
	}
}

// StartPolarSubscription validates the plan is provider-collected and returns
// a PENDING subscription carrying the provider metadata. Re-opening the same
// plan for a profile with a PENDING subscription returns that row, so an
// abandoned checkout never leaves a trail of pending rows.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (b *polarSubscriptionBusiness) StartPolarSubscription(
	ctx context.Context,
	plan *models.Plan,
	in StartPolarInput,
) (result *PolarSubscriptionStart, err error) {
	ctx, span := b.obs.StartSpan(ctx, "StartPolarSubscription")
	defer func() {
		b.obs.EndSpan(ctx, span, err)
		if err == nil {
			b.obs.RecordPolarSubscriptionStarted(ctx)
		}
	}()

	if plan == nil {
		return nil, ErrPlanIDRequired
	}
	provider, productID := ProviderProductFor(plan)
	if provider == "" || productID == "" {
		return nil, ErrPlanNotPolarCollected
	}
	if in.ProfileID == "" {
		return nil, ErrSubscriptionProfileRequired
	}
	if in.PlanID == "" {
		return nil, ErrSubscriptionPlanRequired
	}
	if in.CatalogVersionID == "" {
		return nil, ErrCatalogVersionRequired
	}
	if in.Currency == "" {
		return nil, ErrSubscriptionCurrencyRequired
	}

	if existing, getErr := b.subRepo.GetPendingByProfileAndPlan(ctx, in.ProfileID, in.PlanID); getErr == nil &&
		SubscriptionProvider(existing) == provider {
		util.Log(ctx).WithField("subscription_id", existing.GetID()).
			Debug("reusing pending provider subscription")
		return &PolarSubscriptionStart{Subscription: existing, Provider: provider, PolarProductID: productID}, nil
	}

	now := time.Now()
	subData := map[string]any{}
	for k, v := range in.Data {
		subData[k] = v
	}
	subData[models.SubDataCollectionMode] = models.CollectionModeProvider
	subData[models.SubDataPaymentProvider] = provider
	subData[models.SubDataExternalProductID] = productID
	if provider == providerPolar {
		subData[PolarDataKeyProductID] = productID
	}
	sub := &models.Subscription{
		ProfileID:        in.ProfileID,
		PlanID:           in.PlanID,
		CatalogVersionID: in.CatalogVersionID,
		State:            models.SubscriptionStatePending,
		Currency:         in.Currency,
		StartAt:          now,
		BillingAnchor:    now,
		Data:             subData,
	}
	sub.GenID(ctx)

	if err = b.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}
	_ = b.notifier.NotifyFromSubscription(ctx, models.SubscriptionEventCreated, sub, "")

	util.Log(ctx).
		WithField("subscription_id", sub.GetID()).
		WithField("provider", provider).
		WithField("product_id", productID).
		Info("provider-managed subscription created in PENDING state")

	return &PolarSubscriptionStart{Subscription: sub, Provider: provider, PolarProductID: productID}, nil
}

// MirrorSubscriptionState applies provider state to the matching billing
// Subscription. Lookup order: billing subscription id → provider subscription
// id → newest PENDING row for profile+plan.
//
//nolint:nonamedreturns,gocognit // named err captured by deferred closure; one state machine
func (b *polarSubscriptionBusiness) MirrorSubscriptionState(
	ctx context.Context,
	in MirrorInput,
) (result *models.Subscription, err error) {
	ctx, span := b.obs.StartSpan(ctx, "MirrorSubscriptionState")
	defer func() {
		b.obs.EndSpan(ctx, span, err)
		if err == nil {
			b.obs.RecordPolarSubscriptionMirrored(ctx)
		}
	}()

	if in.Provider == "" {
		in.Provider = providerPolar
	}
	if in.Event == "" {
		in.Event = collection.SubscriptionEventUpdated
	}
	if in.Event == collection.SubscriptionEventUpdated && in.PolarStatus == "" {
		return nil, ErrPolarMirrorPolarStatusRequired
	}
	if in.SubscriptionID == "" && in.PolarSubscriptionID == "" && (in.ProfileID == "" || in.PlanID == "") {
		return nil, ErrPolarMirrorProfileRequired
	}

	sub, err := b.findSubscriptionForMirror(ctx, in)
	if err != nil {
		return nil, err
	}
	if sub.Data == nil {
		sub.Data = make(map[string]any)
	}

	occurredAt := parseRFC3339(in.OccurredAt)
	if in.Event == collection.SubscriptionEventUpdated && !occurredAt.IsZero() {
		if last := parseRFC3339(stringData(sub.Data, models.SubDataProviderEventAt)); !last.IsZero() && occurredAt.Before(last) {
			util.Log(ctx).WithField("subscription_id", sub.GetID()).
				WithField("event_at", in.OccurredAt).Debug("ignoring stale provider event")
			return sub, ErrStaleProviderEvent
		}
	}

	// Provider identity: stamped unconditionally (idempotent).
	sub.Data[models.SubDataCollectionMode] = models.CollectionModeProvider
	sub.Data[models.SubDataPaymentProvider] = in.Provider
	if in.PolarSubscriptionID != "" {
		sub.Data[models.SubDataExternalSubscriptionID] = in.PolarSubscriptionID
		if in.Provider == providerPolar {
			sub.Data[PolarDataKeySubscriptionID] = in.PolarSubscriptionID
		}
	}
	if in.ExternalCustomerID != "" {
		sub.Data[models.SubDataProviderCustomerID] = in.ExternalCustomerID
	}
	if in.ExternalProductID != "" {
		sub.Data[models.SubDataExternalProductID] = in.ExternalProductID
	}
	if in.CurrentPeriodEnd != "" {
		sub.Data[models.SubDataCurrentPeriodEnd] = in.CurrentPeriodEnd
	}

	var events []string
	switch in.Event {
	case collection.SubscriptionEventBilled:
		events = b.applyBilled(sub, in)
	case collection.SubscriptionEventPaymentFailed:
		events = []string{models.SubscriptionEventPaymentFailed}
	default:
		events = b.applyProviderStatus(ctx, sub, in)
		if !occurredAt.IsZero() {
			sub.Data[models.SubDataProviderEventAt] = occurredAt.UTC().Format(time.RFC3339)
		}
	}

	if _, err = b.subRepo.Update(ctx, sub); err != nil {
		return nil, err
	}
	for _, evt := range events {
		invoiceID := ""
		if evt == models.SubscriptionEventBilled {
			invoiceID = in.ExternalOrderID
		}
		_ = b.notifier.NotifyFromSubscription(ctx, evt, sub, invoiceID)
	}

	util.Log(ctx).
		WithField("subscription_id", sub.GetID()).
		WithField("provider", in.Provider).
		WithField("event", in.Event).
		WithField("provider_status", in.PolarStatus).
		WithField("billing_state", sub.State).
		Info("provider subscription state mirrored")

	return sub, nil
}

// findSubscriptionForMirror resolves the billing subscription for an event.
// Every candidate must already be a provider-managed row of this provider
// whose product and provider subscription (when known) match the event, so
// metadata on a webhook can never take over another subscription.
func (b *polarSubscriptionBusiness) findSubscriptionForMirror(
	ctx context.Context,
	in MirrorInput,
) (*models.Subscription, error) {
	if in.SubscriptionID != "" {
		sub, err := b.subRepo.GetByID(ctx, in.SubscriptionID)
		if err == nil {
			if bindErr := providerBindingMatches(sub, in); bindErr != nil {
				return nil, bindErr
			}
			return sub, nil
		}
		if !data.ErrorIsNoRows(err) && !errors.Is(err, apperrors.ErrSubscriptionNotFound) {
			return nil, err
		}
	}
	if in.PolarSubscriptionID != "" {
		sub, err := b.subRepo.GetByExternalSubscriptionID(ctx, in.Provider, in.PolarSubscriptionID)
		if err == nil {
			if bindErr := providerBindingMatches(sub, in); bindErr != nil {
				return nil, bindErr
			}
			return sub, nil
		}
		if !errors.Is(err, apperrors.ErrSubscriptionNotFound) {
			return nil, err
		}
	}
	if in.ProfileID == "" || in.PlanID == "" {
		return nil, ErrPolarSubscriptionNotFound
	}
	sub, err := b.subRepo.GetPendingByProfileAndPlan(ctx, in.ProfileID, in.PlanID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSubscriptionNotFound) {
			return nil, ErrPolarSubscriptionNotFound
		}
		return nil, err
	}
	if bindErr := providerBindingMatches(sub, in); bindErr != nil {
		return nil, bindErr
	}
	return sub, nil
}

// providerBindingMatches checks that an event may be applied to sub: the
// row is provider-managed by in.Provider, its product is the event's
// product and, once a provider subscription id is pinned, the event names it.
func providerBindingMatches(sub *models.Subscription, in MirrorInput) error {
	if !IsProviderManaged(sub) {
		return ErrProviderSubscriptionMismatch
	}
	if p := SubscriptionProvider(sub); p != "" && p != in.Provider {
		return ErrProviderSubscriptionMismatch
	}
	if stored := ExternalSubscriptionID(sub); stored != "" && in.PolarSubscriptionID != "" && stored != in.PolarSubscriptionID {
		return ErrProviderSubscriptionMismatch
	}
	if in.ExternalProductID != "" {
		product := stringData(sub.Data, models.SubDataExternalProductID)
		if product == "" {
			product = stringData(sub.Data, PolarDataKeyProductID)
		}
		if product != "" && product != in.ExternalProductID {
			return ErrProviderSubscriptionMismatch
		}
	}
	return nil
}

// applyBilled records a successful provider charge for a period.
func (b *polarSubscriptionBusiness) applyBilled(sub *models.Subscription, in MirrorInput) []string {
	if in.ExternalOrderID != "" && stringData(sub.Data, models.SubDataLastProviderOrderID) == in.ExternalOrderID {
		return nil // already recorded
	}
	if in.ExternalOrderID != "" {
		sub.Data[models.SubDataLastProviderOrderID] = in.ExternalOrderID
	}
	sub.Data[models.SubDataRenewAttemptCount] = 0
	delete(sub.Data, models.SubDataProviderStatus)
	events := []string{}
	if sub.State == models.SubscriptionStatePending {
		sub.State = models.SubscriptionStateActive
		sub.StartAt = time.Now()
		events = append(events, models.SubscriptionEventActivated)
		b.obs.RecordPolarStateActive(context.Background())
	}
	if pe := parseRFC3339(in.CurrentPeriodEnd); !pe.IsZero() {
		sub.EndAt = &pe
	}
	return append(events, models.SubscriptionEventBilled)
}

// applyProviderStatus maps the normalized provider status onto the billing
// state machine. The sub is mutated in place; returned names are the
// lifecycle events to publish.
//
//nolint:gocognit,funlen // one state machine, kept in one place
func (b *polarSubscriptionBusiness) applyProviderStatus(
	ctx context.Context,
	sub *models.Subscription,
	in MirrorInput,
) []string {
	now := time.Now()
	before := sub.State
	wasCancelAtEnd := cancelAtPeriodEnd(sub.Data)
	wasPastDue := stringData(sub.Data, models.SubDataProviderStatus) == collection.SubscriptionStatusPastDue
	sub.Data[PolarDataKeyStatus] = in.PolarStatus
	sub.Data[models.SubDataProviderStatus] = in.PolarStatus
	periodEnd := parseRFC3339(in.CurrentPeriodEnd)

	var events []string
	activate := func() {
		if sub.State == models.SubscriptionStatePending {
			sub.StartAt = now
			events = append(events, models.SubscriptionEventActivated)
			b.obs.RecordPolarStateActive(ctx)
		}
		sub.State = models.SubscriptionStateActive
		sub.CancelledAt = nil
		if !periodEnd.IsZero() {
			sub.EndAt = &periodEnd
		}
	}
	cancelAtEnd := func() {
		sub.Data[models.SubDataCancelAtPeriodEnd] = true
		if !periodEnd.IsZero() {
			sub.EndAt = &periodEnd
		}
		if !wasCancelAtEnd {
			events = append(events, models.SubscriptionEventCancelled)
		}
	}
	terminate := func(state string) {
		sub.State = state
		if sub.CancelledAt == nil {
			sub.CancelledAt = &now
		}
		delete(sub.Data, models.SubDataCancelAtPeriodEnd)
		if before != state {
			events = append(events, models.SubscriptionEventCancelled)
			b.obs.RecordPolarStateCancelled(ctx)
		}
	}

	switch in.PolarStatus {
	case collection.SubscriptionStatusActive, collection.SubscriptionStatusTrialing:
		activate()
		if in.CancelAtPeriodEnd {
			cancelAtEnd()
		} else {
			delete(sub.Data, models.SubDataCancelAtPeriodEnd)
		}
	case collection.SubscriptionStatusPastDue:
		// Recoverable: the provider retries the charge; access stays on
		// until the provider gives up (unpaid / revoked).
		if sub.State == models.SubscriptionStatePending {
			activate()
		}
		if !wasPastDue {
			events = append(events, models.SubscriptionEventPastDue)
		}
	case collection.SubscriptionStatusIncomplete:
		// First payment not confirmed yet: stay / return to PENDING.
		if sub.State == models.SubscriptionStateActive {
			sub.State = models.SubscriptionStatePending
		}
	case collection.SubscriptionStatusPaused:
		// Provider paused collection; keep access decisions with the product.
		if !wasPastDue {
			events = append(events, models.SubscriptionEventPastDue)
		}
	case collection.SubscriptionStatusCanceled:
		// Cancelled, but when the paid period is still running the payer
		// keeps access until it ends (cancel-at-period-end semantics).
		if !periodEnd.IsZero() && periodEnd.After(now) && sub.State != models.SubscriptionStatePending {
			sub.State = models.SubscriptionStateActive
			cancelAtEnd()
		} else {
			terminate(models.SubscriptionStateCancelled)
		}
	case collection.SubscriptionStatusIncompleteExpired:
		terminate(models.SubscriptionStateCancelled)
	case collection.SubscriptionStatusRevoked, collection.SubscriptionStatusUnpaid:
		terminate(models.SubscriptionStateExpired)
	default:
		util.Log(ctx).
			WithField("provider_status", in.PolarStatus).
			Warn("unknown provider subscription status; billing state left unchanged")
	}
	return events
}

func parseRFC3339(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}

func stringData(d map[string]any, key string) string {
	if d == nil {
		return ""
	}
	v, _ := d[key].(string)
	return v
}
