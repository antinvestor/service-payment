package business

import (
	"context"
	"fmt"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/pitabwire/frame/workerpool"
	"github.com/shopspring/decimal"
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
			Quantity:          decimal.NewNullDecimal(qty),
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

func aggregate(events []*models.UsageEvent, aggType string) decimal.Decimal {
	if len(events) == 0 {
		return decimal.Zero
	}

	switch aggType {
	case models.AggregationTypeSum:
		return aggregateSum(events)
	case models.AggregationTypeCount:
		return decimal.NewFromInt(int64(len(events)))
	case models.AggregationTypeMax:
		return aggregateMax(events)
	case models.AggregationTypeMin:
		return aggregateMin(events)
	case models.AggregationTypeAvg:
		return aggregateAvg(events)
	case models.AggregationTypeLast:
		return aggregateLast(events)
	default:
		return decimal.Zero
	}
}

func aggregateSum(events []*models.UsageEvent) decimal.Decimal {
	sum := decimal.Zero
	for _, ev := range events {
		if ev.Quantity.Valid {
			sum = sum.Add(ev.Quantity.Decimal)
		}
	}
	return sum
}

func aggregateMax(events []*models.UsageEvent) decimal.Decimal {
	maxVal := decimal.Zero
	initialized := false
	for _, ev := range events {
		if ev.Quantity.Valid {
			if !initialized || ev.Quantity.Decimal.GreaterThan(maxVal) {
				maxVal = ev.Quantity.Decimal
				initialized = true
			}
		}
	}
	return maxVal
}

func aggregateMin(events []*models.UsageEvent) decimal.Decimal {
	minVal := decimal.Zero
	initialized := false
	for _, ev := range events {
		if ev.Quantity.Valid {
			if !initialized || ev.Quantity.Decimal.LessThan(minVal) {
				minVal = ev.Quantity.Decimal
				initialized = true
			}
		}
	}
	return minVal
}

func aggregateAvg(events []*models.UsageEvent) decimal.Decimal {
	sum := decimal.Zero
	validCount := int64(0)
	for _, ev := range events {
		if ev.Quantity.Valid {
			sum = sum.Add(ev.Quantity.Decimal)
			validCount++
		}
	}
	if validCount == 0 {
		return decimal.Zero
	}
	return sum.Div(decimal.NewFromInt(validCount))
}

func aggregateLast(events []*models.UsageEvent) decimal.Decimal {
	last := events[len(events)-1]
	if last.Quantity.Valid {
		return last.Quantity.Decimal
	}
	return decimal.Zero
}
