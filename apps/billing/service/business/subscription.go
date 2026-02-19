package business

import (
	"context"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/internal/apperrors"
	"github.com/pitabwire/frame/workerpool"
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
}

func NewSubscriptionBusiness(
	workMan workerpool.Manager,
	subRepo repository.SubscriptionRepository,
) SubscriptionBusiness {
	return &subscriptionBusiness{
		workMan: workMan,
		subRepo: subRepo,
	}
}

func (b *subscriptionBusiness) CreateSubscription(
	ctx context.Context,
	sub *models.Subscription,
) (*models.Subscription, error) {
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

	if err := b.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (b *subscriptionBusiness) GetSubscription(ctx context.Context, id string) (*models.Subscription, error) {
	if id == "" {
		return nil, ErrSubscriptionIDRequired
	}

	return b.subRepo.GetByID(ctx, id)
}

func (b *subscriptionBusiness) CancelSubscription(ctx context.Context, id string) (*models.Subscription, error) {
	if id == "" {
		return nil, ErrSubscriptionIDRequired
	}

	sub, err := b.subRepo.GetByID(ctx, id)
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
