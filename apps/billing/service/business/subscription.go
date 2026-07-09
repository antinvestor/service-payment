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
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/workerpool"
)

// SubscriptionBusiness defines the business interface for subscription operations.
type SubscriptionBusiness interface {
	CreateSubscription(ctx context.Context, sub *models.Subscription) (*models.Subscription, error)
	GetSubscription(ctx context.Context, id string) (*models.Subscription, error)
	CancelSubscription(ctx context.Context, id string) (*models.Subscription, error)
	ListActiveByProfile(ctx context.Context, profileID string) ([]*models.Subscription, error)
}

type subscriptionBusiness struct {
	workMan workerpool.Manager
	subRepo repository.SubscriptionRepository
	obs     *observability.Metrics
}

func NewSubscriptionBusiness(
	workMan workerpool.Manager,
	subRepo repository.SubscriptionRepository,
) SubscriptionBusiness {
	return &subscriptionBusiness{
		workMan: workMan,
		subRepo: subRepo,
		obs:     observability.NewMetrics(),
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
