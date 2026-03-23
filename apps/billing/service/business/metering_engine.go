package business

import (
	"context"
	"fmt"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util/decimalx"
)

// MeteringEngine aggregates raw usage events into windowed usage per component.
type MeteringEngine interface {
	MeterUsage(ctx context.Context, sub *models.Subscription, components []*models.Component,
		billingRun *models.BillingRun) ([]*models.MeteredUsage, error)
}

type meteringEngine struct {
	workMan   workerpool.Manager
	eventRepo repository.UsageEventRepository
	meterRepo repository.MeteredUsageRepository
}

func NewMeteringEngine(
	workMan workerpool.Manager,
	eventRepo repository.UsageEventRepository,
	meterRepo repository.MeteredUsageRepository,
) MeteringEngine {
	return &meteringEngine{
		workMan:   workMan,
		eventRepo: eventRepo,
		meterRepo: meterRepo,
	}
}

func (e *meteringEngine) MeterUsage(
	ctx context.Context,
	sub *models.Subscription,
	components []*models.Component,
	billingRun *models.BillingRun,
) ([]*models.MeteredUsage, error) {
	var metered []*models.MeteredUsage

	for _, comp := range components {
		events, err := e.eventRepo.ListBySubscriptionAndPeriod(
			ctx, sub.GetID(), comp.MetricKey, billingRun.PeriodStart, billingRun.PeriodEnd)
		if err != nil {
			return nil, err
		}

		qty := aggregate(events, comp.AggregationType)

		mu := &models.MeteredUsage{
			SubscriptionID:    sub.GetID(),
			ComponentID:       comp.GetID(),
			MetricKey:         comp.MetricKey,
			WindowStart:       billingRun.PeriodStart,
			WindowEnd:         billingRun.PeriodEnd,
			WindowGranularity: models.WindowGranularityMonth,
			AggregationType:   comp.AggregationType,
			Quantity:          qty.Ptr(),
			EventCount:        int64(len(events)),
			BillingRunID:      billingRun.GetID(),
		}
		mu.GenID(ctx)
		mu.ID = fmt.Sprintf("%s_%s", billingRun.GetID(), comp.GetID())

		if createErr := e.meterRepo.Create(ctx, mu); createErr != nil {
			return nil, createErr
		}

		metered = append(metered, mu)
	}

	return metered, nil
}

func aggregate(events []*models.UsageEvent, aggType string) decimalx.Decimal {
	if len(events) == 0 {
		return decimalx.Zero()
	}

	switch aggType {
	case models.AggregationTypeSum:
		return aggregateSum(events)
	case models.AggregationTypeCount:
		return decimalx.NewFromInt64(int64(len(events)))
	case models.AggregationTypeMax:
		return aggregateMax(events)
	case models.AggregationTypeMin:
		return aggregateMin(events)
	case models.AggregationTypeAvg:
		return aggregateAvg(events)
	case models.AggregationTypeLast:
		return aggregateLast(events)
	default:
		return decimalx.Zero()
	}
}

func aggregateSum(events []*models.UsageEvent) decimalx.Decimal {
	sum := decimalx.Zero()
	for _, ev := range events {
		if ev.Quantity != nil {
			sum = sum.Add(*ev.Quantity)
		}
	}
	return sum
}

func aggregateMax(events []*models.UsageEvent) decimalx.Decimal {
	maxVal := decimalx.Zero()
	initialized := false
	for _, ev := range events {
		if ev.Quantity != nil {
			if !initialized || ev.Quantity.GreaterThan(maxVal) {
				maxVal = *ev.Quantity
				initialized = true
			}
		}
	}
	return maxVal
}

func aggregateMin(events []*models.UsageEvent) decimalx.Decimal {
	minVal := decimalx.Zero()
	initialized := false
	for _, ev := range events {
		if ev.Quantity != nil {
			if !initialized || ev.Quantity.LessThan(minVal) {
				minVal = *ev.Quantity
				initialized = true
			}
		}
	}
	return minVal
}

func aggregateAvg(events []*models.UsageEvent) decimalx.Decimal {
	sum := decimalx.Zero()
	validCount := int64(0)
	for _, ev := range events {
		if ev.Quantity != nil {
			sum = sum.Add(*ev.Quantity)
			validCount++
		}
	}
	if validCount == 0 {
		return decimalx.Zero()
	}
	return sum.Div(decimalx.NewFromInt64(validCount))
}

func aggregateLast(events []*models.UsageEvent) decimalx.Decimal {
	last := events[len(events)-1]
	if last.Quantity != nil {
		return *last.Quantity
	}
	return decimalx.Zero()
}
