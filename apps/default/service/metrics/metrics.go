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

// Package metrics defines the tenant-scoped business instruments for the
// payment lifecycle. Instruments are created through frame's
// telemetry.BusinessMetrics factory, so every measurement transparently
// carries tenant_id and partition_id derived from the context claims.
package metrics

import (
	"context"
	"strconv"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

const unknownAttrValue = "unknown"

// PaymentMetrics groups the business instruments for the payment lifecycle.
type PaymentMetrics struct {
	initiated  telemetry.Counter
	succeeded  telemetry.Counter
	failed     telemetry.Counter
	amount     telemetry.FloatCounter
	processing telemetry.Histogram
}

// NewPaymentMetrics creates the payment business instruments. The underlying
// OTel SDK deduplicates instruments by name, so multiple constructions are
// safe and record into the same series.
func NewPaymentMetrics() *PaymentMetrics {
	bm := telemetry.NewBusinessMetrics("service-payments")
	return &PaymentMetrics{
		initiated: bm.Counter(
			"payments_initiated_total",
			"Payments accepted for processing",
		),
		succeeded: bm.Counter(
			"payments_succeeded_total",
			"Payments that reached a successful terminal status",
		),
		failed: bm.Counter(
			"payments_failed_total",
			"Payments that reached a failed terminal status",
		),
		amount: bm.FloatCounter(
			"payments_amount_total",
			"Value of successful payments in major currency units",
		),
		processing: bm.Histogram(
			"payments_processing_duration_ms",
			"Time from payment initiation to terminal status",
		),
	}
}

func providerOf(p *models.Payment) string {
	if p == nil || p.RouteID == "" {
		return unknownAttrValue
	}
	return p.RouteID
}

// failureReason derives a bounded-cardinality reason from the status extras.
func failureReason(status *models.Status) string {
	if step := status.Extra.GetString("step"); step != "" {
		return step
	}
	if status.Extra.GetString("error") != "" {
		return "provider_error"
	}
	return unknownAttrValue
}

// RecordInitiated counts a payment accepted for processing.
func (m *PaymentMetrics) RecordInitiated(ctx context.Context, p *models.Payment) {
	if m == nil {
		return
	}
	m.initiated.Add(ctx, 1, attribute.String("provider", providerOf(p)))
}

// IsTerminalStatus reports whether a status value is a terminal payment
// outcome worth recording.
func IsTerminalStatus(status int32) bool {
	s := commonv1.STATUS(status)
	return s == commonv1.STATUS_SUCCESSFUL || s == commonv1.STATUS_FAILED
}

// RecordTerminal records a payment reaching a terminal status: success or
// failure counters, the amount moved on success, and the processing duration.
func (m *PaymentMetrics) RecordTerminal(ctx context.Context, p *models.Payment, status *models.Status) {
	if m == nil || p == nil || status == nil {
		return
	}

	providerAttr := attribute.String("provider", providerOf(p))
	elapsedMs := float64(time.Since(p.CreatedAt).Milliseconds())

	outcome := commonv1.STATUS(status.Status)
	if outcome == commonv1.STATUS_SUCCESSFUL {
		m.succeeded.Add(ctx, 1, providerAttr)
		m.processing.Record(ctx, elapsedMs, providerAttr)

		if p.Amount == nil {
			return
		}
		amount, err := strconv.ParseFloat(p.Amount.String(), 64)
		if err != nil {
			return
		}
		currency := p.Currency
		if currency == "" {
			currency = unknownAttrValue
		}
		m.amount.Add(ctx, amount, providerAttr, attribute.String("currency", currency))
		return
	}

	if outcome == commonv1.STATUS_FAILED {
		m.failed.Add(ctx, 1, providerAttr, attribute.String("reason", failureReason(status)))
		m.processing.Record(ctx, elapsedMs, providerAttr)
	}
}
