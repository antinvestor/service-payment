package business

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/workerpool"
)

// UsageIngestionBusiness defines the business interface for usage event ingestion.
type UsageIngestionBusiness interface {
	IngestEvent(ctx context.Context, event *models.UsageEvent) (*models.UsageEvent, error)
	IngestBatch(ctx context.Context, events []*models.UsageEvent) ([]*models.UsageEvent, error)
}

type usageIngestionBusiness struct {
	workMan   workerpool.Manager
	eventRepo repository.UsageEventRepository
}

func NewUsageIngestionBusiness(
	workMan workerpool.Manager,
	eventRepo repository.UsageEventRepository,
) UsageIngestionBusiness {
	return &usageIngestionBusiness{
		workMan:   workMan,
		eventRepo: eventRepo,
	}
}

func (b *usageIngestionBusiness) IngestEvent(
	ctx context.Context,
	event *models.UsageEvent,
) (*models.UsageEvent, error) {
	if event.EventID == "" {
		return nil, ErrUsageEventIDRequired
	}
	if event.SubscriptionID == "" {
		return nil, ErrUsageSubscriptionIDRequired
	}
	if event.MetricKey == "" {
		return nil, ErrUsageMetricKeyRequired
	}
	if event.Quantity == nil || event.Quantity.IsZero() {
		return nil, ErrUsageQuantityRequired
	}

	event.GenID(ctx)

	// Idempotent: BaseRepository uses ON CONFLICT DO NOTHING
	err := b.eventRepo.Create(ctx, event)
	if err != nil {
		if data.ErrorIsDuplicateKey(err) {
			return nil, apperrors.ErrUsageEventDuplicate
		}
		return nil, err
	}

	return event, nil
}

func (b *usageIngestionBusiness) IngestBatch(
	ctx context.Context,
	events []*models.UsageEvent,
) ([]*models.UsageEvent, error) {
	var ingested []*models.UsageEvent
	for _, event := range events {
		result, err := b.IngestEvent(ctx, event)
		if err != nil {
			// Skip duplicates in batch mode
			var appErr apperrors.ApplicationError
			if ok := isApplicationError(
				err,
				&appErr,
			); ok &&
				appErr.ErrorCode() == apperrors.DefaultCodeOffset+apperrors.ErrorCodeUsageEventDuplicate {
				continue
			}
			return ingested, err
		}
		ingested = append(ingested, result)
	}
	return ingested, nil
}

func isApplicationError(err error, target *apperrors.ApplicationError) bool {
	return errors.As(err, target)
}
