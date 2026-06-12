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
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/workerpool"
	"github.com/pitabwire/util"
)

// Data map keys used by polar-collected subscriptions.
const (
	PolarDataKeyProductID         = "polarProductId"
	PolarDataKeySubscriptionID    = "polarSubscriptionId"
	PolarDataKeyStatus            = "polarStatus"
	PolarDataKeyCurrentPeriodEnd  = "currentPeriodEnd"
	PolarDataKeyCollectionMode    = "collectionMode"
	PolarDataKeyCancelAtPeriodEnd = "cancelAtPeriodEnd"

	polarCollectionModeValue = "polar"

	// Polar subscription status values sent by the Polar webhook.
	PolarStatusActive   = "active"
	PolarStatusCanceled = "canceled"
	PolarStatusRevoked  = "revoked"
	PolarStatusPastDue  = "past_due"
	PolarStatusUnpaid   = "unpaid"
)

// PolarProductID returns the polar product ID configured on the plan,
// or "" when the plan has no polar product mapping.
func PolarProductID(plan *models.Plan) string {
	if plan == nil || plan.Data == nil {
		return ""
	}
	v, _ := plan.Data[PolarDataKeyProductID].(string)
	return v
}

// IsPolarCollected reports whether billing should delegate charge collection
// for the plan to Polar. Plans with a non-empty polarProductId in their Data
// map are polar-collected; all others use billing's internal invoice cycle.
func IsPolarCollected(plan *models.Plan) bool {
	return PolarProductID(plan) != ""
}

// StartPolarInput carries the parameters needed to initiate a polar-collected
// subscription checkout.
type StartPolarInput struct {
	ProfileID        string
	PlanID           string
	CatalogVersionID string
	Currency         string
}

// PolarSubscriptionStart is returned by StartPolarSubscription. It contains
// the newly created billing Subscription (PENDING) and the Polar product ID
// that the caller must use to initiate a Polar checkout session.
type PolarSubscriptionStart struct {
	Subscription   *models.Subscription
	PolarProductID string
}

// MirrorInput carries the webhook-derived state that billing should mirror
// onto the matching billing Subscription.
type MirrorInput struct {
	// PolarSubscriptionID is the identifier Polar assigned to the subscription.
	// Used as the primary lookup key; falls back to ProfileID+PlanID if empty.
	PolarSubscriptionID string
	ProfileID           string
	PlanID              string
	// PolarStatus is the status string from the Polar webhook payload
	// (e.g. "active", "canceled", "revoked", "past_due", "unpaid").
	PolarStatus      string
	CurrentPeriodEnd string // RFC 3339 timestamp or empty
}

// PolarSubscriptionBusiness manages the lifecycle of polar-collected subscriptions.
// It is intentionally separate from SubscriptionBusiness to keep the polar-specific
// logic isolated and the base interface unchanged.
//
// Transport note: Polar webhooks are emitted by the payment/polar service, NOT billing.
// Billing has no inbound queue or event subscriber (see apps/billing/cmd/main.go — no
// frame.WithRegisterSubscriber is wired). MirrorSubscriptionState is therefore exposed
// as a business method; the payment/polar side MUST call it (via an RPC wrapper) when
// it receives a Polar subscription.updated webhook.
// TODO: wire an RPC handler on the billing service that calls MirrorSubscriptionState,
// or add a billing subscriber once a shared queue topic is established.
type PolarSubscriptionBusiness interface {
	// StartPolarSubscription creates a PENDING billing Subscription linked to the
	// plan's Polar product and returns the polar product ID for initiating checkout.
	// Returns ErrPlanNotPolarCollected when the plan has no polarProductId.
	StartPolarSubscription(ctx context.Context, plan *models.Plan, in StartPolarInput) (*PolarSubscriptionStart, error)

	// MirrorSubscriptionState maps a Polar webhook payload to the billing Subscription
	// state machine. It is idempotent: applying the same event twice yields the same result.
	// Returns ErrPolarSubscriptionNotFound when no subscription can be matched.
	MirrorSubscriptionState(ctx context.Context, in MirrorInput) (*models.Subscription, error)
}

type polarSubscriptionBusiness struct {
	workMan workerpool.Manager
	subRepo repository.SubscriptionRepository
	obs     *observability.Metrics
}

// NewPolarSubscriptionBusiness constructs a PolarSubscriptionBusiness.
func NewPolarSubscriptionBusiness(
	workMan workerpool.Manager,
	subRepo repository.SubscriptionRepository,
) PolarSubscriptionBusiness {
	return &polarSubscriptionBusiness{
		workMan: workMan,
		subRepo: subRepo,
		obs:     observability.NewMetrics(),
	}
}

// StartPolarSubscription validates the plan is polar-collected, creates a PENDING
// subscription with polar metadata in Data, and returns the polar product ID.
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
	polarProductID := PolarProductID(plan)
	if polarProductID == "" {
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

	now := time.Now()
	sub := &models.Subscription{
		ProfileID:        in.ProfileID,
		PlanID:           in.PlanID,
		CatalogVersionID: in.CatalogVersionID,
		State:            models.SubscriptionStatePending,
		Currency:         in.Currency,
		StartAt:          now,
		BillingAnchor:    now,
		Data: map[string]any{
			PolarDataKeyCollectionMode: polarCollectionModeValue,
			PolarDataKeyProductID:      polarProductID,
		},
	}
	sub.GenID(ctx)

	if err = b.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	util.Log(ctx).
		WithField("subscription_id", sub.GetID()).
		WithField("polar_product_id", polarProductID).
		Info("polar-collected subscription created in PENDING state")

	return &PolarSubscriptionStart{
		Subscription:   sub,
		PolarProductID: polarProductID,
	}, nil
}

// MirrorSubscriptionState applies a Polar webhook state to the matching billing Subscription.
// Lookup order: polarSubscriptionId in Data → ProfileID+PlanID PENDING fallback.
// Idempotent: re-applying the same status twice yields the same final state.
//
//nolint:nonamedreturns // named err captured by deferred closure
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

	if in.ProfileID == "" {
		return nil, ErrPolarMirrorProfileRequired
	}
	if in.PlanID == "" {
		return nil, ErrPolarMirrorPlanRequired
	}
	if in.PolarStatus == "" {
		return nil, ErrPolarMirrorPolarStatusRequired
	}

	// Locate the billing subscription: prefer an exact polarSubscriptionId match,
	// fall back to the most recent PENDING record for this profile+plan.
	sub, err := b.findSubscriptionForMirror(ctx, in)
	if err != nil {
		return nil, err
	}

	// Stamp polar metadata unconditionally (idempotent).
	if sub.Data == nil {
		sub.Data = make(map[string]any)
	}
	if in.PolarSubscriptionID != "" {
		sub.Data[PolarDataKeySubscriptionID] = in.PolarSubscriptionID
	}
	sub.Data[PolarDataKeyStatus] = in.PolarStatus
	if in.CurrentPeriodEnd != "" {
		sub.Data[PolarDataKeyCurrentPeriodEnd] = in.CurrentPeriodEnd
	}

	// Map polar status → billing state.
	b.applyPolarStatus(ctx, sub, in)

	_, err = b.subRepo.Update(ctx, sub)
	if err != nil {
		return nil, err
	}

	util.Log(ctx).
		WithField("subscription_id", sub.GetID()).
		WithField("polar_status", in.PolarStatus).
		WithField("billing_state", sub.State).
		Info("polar subscription state mirrored")

	return sub, nil
}

// findSubscriptionForMirror resolves the billing subscription for the incoming mirror event.
func (b *polarSubscriptionBusiness) findSubscriptionForMirror(
	ctx context.Context,
	in MirrorInput,
) (*models.Subscription, error) {
	// Primary lookup: exact polar subscription ID.
	if in.PolarSubscriptionID != "" {
		sub, err := b.subRepo.GetByPolarSubscriptionID(ctx, in.PolarSubscriptionID)
		if err == nil {
			return sub, nil
		}
		if !errors.Is(err, apperrors.ErrSubscriptionNotFound) {
			return nil, err
		}
		// Not found by polar ID — fall through to profile+plan PENDING lookup.
	}

	// Fallback: find the most recent PENDING subscription for this profile+plan.
	sub, err := b.subRepo.GetPendingByProfileAndPlan(ctx, in.ProfileID, in.PlanID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSubscriptionNotFound) {
			return nil, ErrPolarSubscriptionNotFound
		}
		return nil, err
	}

	return sub, nil
}

// applyPolarStatus maps the Polar status to billing state and sets timestamps.
// See design: active→ACTIVE; canceled with future period end→ACTIVE+cancelAtPeriodEnd;
// canceled past period end→CANCELLED; revoked/past_due/unpaid→EXPIRED.
// The sub is mutated in place.
func (b *polarSubscriptionBusiness) applyPolarStatus(
	ctx context.Context,
	sub *models.Subscription,
	in MirrorInput,
) {
	now := time.Now()

	switch in.PolarStatus {
	case PolarStatusActive:
		// First activation: set StartAt.
		if sub.State == models.SubscriptionStatePending {
			sub.StartAt = now
		}
		sub.State = models.SubscriptionStateActive
		// Update EndAt from current period end.
		if in.CurrentPeriodEnd != "" {
			t, parseErr := time.Parse(time.RFC3339, in.CurrentPeriodEnd)
			if parseErr == nil {
				sub.EndAt = &t
			}
		}
		// Remove any lingering cancel-at-period-end flag on re-activation.
		delete(sub.Data, PolarDataKeyCancelAtPeriodEnd)
		b.obs.RecordPolarStateActive(ctx)

	case PolarStatusCanceled:
		// Polar "canceled" = will not renew but remains active until period end.
		// If period end is in the future, keep ACTIVE and set the cancelAtPeriodEnd flag.
		periodEnd, err := parsePeriodEnd(in.CurrentPeriodEnd)
		if err == nil && periodEnd.After(now) {
			sub.State = models.SubscriptionStateActive
			sub.Data[PolarDataKeyCancelAtPeriodEnd] = "true"
			t := periodEnd
			sub.EndAt = &t
		} else {
			// Period already ended or no period info: hard cancel.
			sub.State = models.SubscriptionStateCancelled
			sub.CancelledAt = &now
			b.obs.RecordPolarStateCancelled(ctx)
		}

	case PolarStatusRevoked, PolarStatusPastDue, PolarStatusUnpaid:
		// Immediate termination.
		sub.State = models.SubscriptionStateExpired
		sub.CancelledAt = &now
		b.obs.RecordPolarStateCancelled(ctx)

	default:
		// Unknown status — log and leave state unchanged to avoid incorrect transitions.
		util.Log(ctx).
			WithField("polar_status", in.PolarStatus).
			Warn("unknown polar subscription status; billing state left unchanged")
	}
}

// parsePeriodEnd parses an RFC 3339 period-end timestamp. Returns a zero time on error.
func parsePeriodEnd(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty period end")
	}
	return time.Parse(time.RFC3339, s)
}
