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
	"fmt"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/workerpool"
	"github.com/pitabwire/util"
)

// SubscriptionBusiness defines the business interface for subscription operations.
type SubscriptionBusiness interface {
	CreateSubscription(ctx context.Context, sub *models.Subscription) (*models.Subscription, error)
	GetSubscription(ctx context.Context, id string) (*models.Subscription, error)
	CancelSubscription(ctx context.Context, id string) (*models.Subscription, error)
	ListActiveByProfile(ctx context.Context, profileID string) ([]*models.Subscription, error)
	// ActivateSubscription transitions a PENDING subscription to ACTIVE after payment.
	// Idempotent: already-ACTIVE subscriptions are returned unchanged.
	ActivateSubscription(ctx context.Context, id string) (*models.Subscription, error)
	// CancelPendingSubscription cancels a PENDING subscription awaiting first payment.
	// Idempotent for already-CANCELLED. Rejects ACTIVE (use CancelSubscription).
	CancelPendingSubscription(ctx context.Context, id string) (*models.Subscription, error)
	// PatchSubscriptionData merges keys into the subscription Data map and persists.
	PatchSubscriptionData(ctx context.Context, id string, patch map[string]any) (*models.Subscription, error)
	// NotifyBilled emits subscription.billed to external integrators after invoice pay.
	NotifyBilled(ctx context.Context, sub *models.Subscription, invoiceID string)
}

type subscriptionBusiness struct {
	workMan  workerpool.Manager
	subRepo  repository.SubscriptionRepository
	notifier SubscriptionLifecycleNotifier
	obs      *observability.Metrics
}

func NewSubscriptionBusiness(
	workMan workerpool.Manager,
	subRepo repository.SubscriptionRepository,
	notifier ...SubscriptionLifecycleNotifier,
) SubscriptionBusiness {
	var n SubscriptionLifecycleNotifier
	if len(notifier) > 0 && notifier[0] != nil {
		n = notifier[0]
	} else {
		n = NoopSubscriptionLifecycleNotifier()
	}
	return &subscriptionBusiness{
		workMan:  workMan,
		subRepo:  subRepo,
		notifier: n,
		obs:      observability.NewMetrics(),
	}
}

// notifyLifecycle is best-effort; never fails the primary business operation.
func (b *subscriptionBusiness) notifyLifecycle(
	ctx context.Context,
	eventType string,
	sub *models.Subscription,
	invoiceID string,
) {
	if b.notifier == nil || sub == nil {
		return
	}
	if err := b.notifier.NotifyFromSubscription(ctx, eventType, sub, invoiceID); err != nil {
		util.Log(ctx).
			WithError(err).
			WithField("event", eventType).
			WithField("subscription_id", sub.GetID()).
			Warn("subscription lifecycle notify failed")
	}
}

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (b *subscriptionBusiness) CreateSubscription(
	ctx context.Context,
	sub *models.Subscription,
) (result *models.Subscription, err error) {
	ctx, span := b.obs.StartSpan(ctx, "CreateSubscription")
	defer func() {
		b.obs.EndSpan(ctx, span, err)
		if err == nil {
			b.obs.RecordSubscriptionCreated(ctx)
		}
	}()

	if sub.ProfileID == "" {
		return nil, ErrSubscriptionProfileRequired
	}
	if sub.PlanID == "" {
		return nil, ErrSubscriptionPlanRequired
	}
	if sub.CatalogVersionID == "" {
		return nil, ErrCatalogVersionRequired
	}
	if sub.Currency == "" {
		return nil, ErrSubscriptionCurrencyRequired
	}

	sub.GenID(ctx)
	if sub.State == "" {
		sub.State = models.SubscriptionStateActive
	}
	if sub.StartAt.IsZero() {
		sub.StartAt = time.Now()
	}
	if sub.BillingAnchor.IsZero() {
		sub.BillingAnchor = sub.StartAt
	}

	if err = b.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	// External-entity integrators: created (+ activated when free path starts ACTIVE).
	b.notifyLifecycle(ctx, models.SubscriptionEventCreated, sub, "")
	if sub.State == models.SubscriptionStateActive {
		b.notifyLifecycle(ctx, models.SubscriptionEventActivated, sub, "")
	}

	return sub, nil
}

func (b *subscriptionBusiness) GetSubscription(
	ctx context.Context,
	id string,
) (*models.Subscription, error) {
	if id == "" {
		return nil, ErrSubscriptionIDRequired
	}

	return b.subRepo.GetByID(ctx, id)
}

//nolint:dupl,nonamedreturns // same span+counter pattern; named err captured by deferred span-end closure
func (b *subscriptionBusiness) CancelSubscription(
	ctx context.Context,
	id string,
) (result *models.Subscription, err error) {
	ctx, span := b.obs.StartSpan(ctx, "CancelSubscription")
	defer func() {
		b.obs.EndSpan(ctx, span, err)
		if err == nil {
			b.obs.RecordSubscriptionCancelled(ctx)
		}
	}()

	if id == "" {
		return nil, ErrSubscriptionIDRequired
	}

	var sub *models.Subscription
	sub, err = b.subRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if sub.State != models.SubscriptionStateActive {
		return nil, apperrors.ErrSubscriptionNotActive
	}

	now := time.Now()
	sub.State = models.SubscriptionStateCancelled
	sub.CancelledAt = &now

	_, err = b.subRepo.Update(ctx, sub)
	if err != nil {
		return nil, err
	}

	b.notifyLifecycle(ctx, models.SubscriptionEventCancelled, sub, "")
	return sub, nil
}

func (b *subscriptionBusiness) ListActiveByProfile(
	ctx context.Context,
	profileID string,
) ([]*models.Subscription, error) {
	if profileID == "" {
		return nil, ErrSubscriptionProfileRequired
	}

	return b.subRepo.ListActiveByProfile(ctx, profileID)
}

// ActivateSubscription moves a PENDING subscription to ACTIVE after successful payment.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (b *subscriptionBusiness) ActivateSubscription(
	ctx context.Context,
	id string,
) (result *models.Subscription, err error) {
	ctx, span := b.obs.StartSpan(ctx, "ActivateSubscription")
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	if id == "" {
		return nil, ErrSubscriptionIDRequired
	}

	var sub *models.Subscription
	sub, err = b.subRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Idempotent: already active is success.
	if sub.State == models.SubscriptionStateActive {
		return sub, nil
	}
	if sub.State != models.SubscriptionStatePending {
		return nil, apperrors.ErrSubscriptionNotActive.Extend(
			fmt.Sprintf("cannot activate subscription in state %s", sub.State),
		)
	}

	sub.State = models.SubscriptionStateActive
	_, err = b.subRepo.Update(ctx, sub)
	if err != nil {
		return nil, err
	}
	b.notifyLifecycle(ctx, models.SubscriptionEventActivated, sub, "")
	return sub, nil
}

// CancelPendingSubscription cancels a PENDING subscription before first payment.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (b *subscriptionBusiness) CancelPendingSubscription(
	ctx context.Context,
	id string,
) (result *models.Subscription, err error) {
	ctx, span := b.obs.StartSpan(ctx, "CancelPendingSubscription")
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	if id == "" {
		return nil, ErrSubscriptionIDRequired
	}

	var sub *models.Subscription
	sub, err = b.subRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if sub.State == models.SubscriptionStateCancelled {
		return sub, nil
	}
	if sub.State != models.SubscriptionStatePending {
		return nil, apperrors.ErrSubscriptionNotActive.Extend(
			fmt.Sprintf("CancelPendingSubscription only applies to PENDING (got %s)", sub.State),
		)
	}

	now := time.Now()
	sub.State = models.SubscriptionStateCancelled
	sub.CancelledAt = &now
	_, err = b.subRepo.Update(ctx, sub)
	if err != nil {
		return nil, err
	}
	b.obs.RecordSubscriptionCancelled(ctx)
	b.notifyLifecycle(ctx, models.SubscriptionEventCancelled, sub, "")
	return sub, nil
}

// NotifyBilled publishes subscription.billed for product/external-entity integrators.
func (b *subscriptionBusiness) NotifyBilled(
	ctx context.Context,
	sub *models.Subscription,
	invoiceID string,
) {
	b.notifyLifecycle(ctx, models.SubscriptionEventBilled, sub, invoiceID)
}

// PatchSubscriptionData merges patch keys into subscription.Data and saves.
func (b *subscriptionBusiness) PatchSubscriptionData(
	ctx context.Context,
	id string,
	patch map[string]any,
) (*models.Subscription, error) {
	if id == "" {
		return nil, ErrSubscriptionIDRequired
	}
	sub, err := b.subRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub.Data == nil {
		sub.Data = make(map[string]any)
	}
	for k, v := range patch {
		sub.Data[k] = v
	}
	_, err = b.subRepo.Update(ctx, sub)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
