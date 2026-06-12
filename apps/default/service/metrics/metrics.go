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
	"go.opentelemetry.io/otel/trace"
)

const (
	unknownAttrValue = "unknown"
	pkgName          = "service-payments"
)

// Metrics holds the tracer and all business instruments for the payment service.
// It is the single observability handle to pass into business constructors.
type Metrics struct {
	tracer telemetry.Tracer

	// Payment lifecycle instruments.
	paymentInitiated  telemetry.Counter
	paymentSucceeded  telemetry.Counter
	paymentFailed     telemetry.Counter
	paymentAmount     telemetry.FloatCounter
	paymentProcessing telemetry.Histogram

	// Prompt instruments.
	promptInitiated telemetry.Counter
	promptFailed    telemetry.Counter
	promptLatency   telemetry.Histogram

	// Payment link instruments.
	paymentLinkCreated telemetry.Counter
}

// NewMetrics constructs the single shared Metrics handle for the default payment
// service. The OTel SDK deduplicates instruments by name so multiple
// constructions are safe and record into the same series.
func NewMetrics() *Metrics {
	t := telemetry.NewTracer(pkgName)
	bm := telemetry.NewBusinessMetrics(pkgName)
	return &Metrics{
		tracer: t,

		paymentInitiated: bm.Counter(
			"payments_initiated_total",
			"Payments accepted for processing",
		),
		paymentSucceeded: bm.Counter(
			"payments_succeeded_total",
			"Payments that reached a successful terminal status",
		),
		paymentFailed: bm.Counter(
			"payments_failed_total",
			"Payments that reached a failed terminal status",
		),
		paymentAmount: bm.FloatCounter(
			"payments_amount_total",
			"Value of successful payments in major currency units",
		),
		paymentProcessing: bm.Histogram(
			"payments_processing_duration_ms",
			"Time from payment initiation to terminal status",
		),

		promptInitiated: bm.Counter(
			"prompts_initiated_total",
			"Prompt requests accepted for processing",
		),
		promptFailed: bm.Counter(
			"prompts_failed_total",
			"Prompt requests that failed",
		),
		promptLatency: bm.Histogram(
			"prompts_latency_ms",
			"Time to complete a prompt initiation",
		),

		paymentLinkCreated: bm.Counter(
			"payment_links_created_total",
			"Payment links created successfully",
		),
	}
}

// StartSpan starts a new traced span and returns the enriched context and span.
func (m *Metrics) StartSpan(
	ctx context.Context,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return m.tracer.Start(ctx, name, opts...)
}

// EndSpan ends a span, recording any error on it.
func (m *Metrics) EndSpan(ctx context.Context, span trace.Span, err error) {
	m.tracer.End(ctx, span, err)
}

// RecordPaymentInitiated counts a payment accepted for processing.
func (m *Metrics) RecordPaymentInitiated(ctx context.Context, p *models.Payment) {
	if m == nil {
		return
	}
	m.paymentInitiated.Add(ctx, 1, attribute.String("provider", providerOf(p)))
}

// RecordPaymentTerminal records a payment reaching a terminal status.
func (m *Metrics) RecordPaymentTerminal(ctx context.Context, p *models.Payment, status *models.Status) {
	if m == nil || p == nil || status == nil {
		return
	}
	recordTerminalOutcome(ctx, p, status, m.paymentSucceeded, m.paymentFailed, m.paymentAmount, m.paymentProcessing)
}

// RecordPromptInitiated counts a prompt request accepted for processing.
func (m *Metrics) RecordPromptInitiated(ctx context.Context) {
	if m == nil {
		return
	}
	m.promptInitiated.Add(ctx, 1)
}

// RecordPromptFailed counts a prompt request that failed.
func (m *Metrics) RecordPromptFailed(ctx context.Context) {
	if m == nil {
		return
	}
	m.promptFailed.Add(ctx, 1)
}

// RecordPromptLatency records how long a prompt initiation took.
func (m *Metrics) RecordPromptLatency(ctx context.Context, elapsed time.Duration) {
	if m == nil {
		return
	}
	m.promptLatency.Record(ctx, float64(elapsed.Milliseconds()))
}

// RecordPaymentLinkCreated counts a payment link created successfully.
func (m *Metrics) RecordPaymentLinkCreated(ctx context.Context) {
	if m == nil {
		return
	}
	m.paymentLinkCreated.Add(ctx, 1)
}

// --- Legacy PaymentMetrics (kept for backward-compatibility) ---

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
	bm := telemetry.NewBusinessMetrics(pkgName)
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
	recordTerminalOutcome(ctx, p, status, m.succeeded, m.failed, m.amount, m.processing)
}

// recordTerminalOutcome is the shared algorithm for recording a terminal payment
// outcome. It accepts the instrument handles explicitly so it can be reused by
// both the unified Metrics and the legacy PaymentMetrics types.
func recordTerminalOutcome(
	ctx context.Context,
	p *models.Payment,
	status *models.Status,
	succeeded telemetry.Counter,
	failed telemetry.Counter,
	amount telemetry.FloatCounter,
	processing telemetry.Histogram,
) {
	providerAttr := attribute.String("provider", providerOf(p))
	elapsedMs := float64(time.Since(p.CreatedAt).Milliseconds())

	outcome := commonv1.STATUS(status.Status)
	if outcome == commonv1.STATUS_SUCCESSFUL {
		succeeded.Add(ctx, 1, providerAttr)
		processing.Record(ctx, elapsedMs, providerAttr)

		if p.Amount == nil {
			return
		}
		amt, err := strconv.ParseFloat(p.Amount.String(), 64)
		if err != nil {
			return
		}
		currency := p.Currency
		if currency == "" {
			currency = unknownAttrValue
		}
		amount.Add(ctx, amt, providerAttr, attribute.String("currency", currency))
		return
	}

	if outcome == commonv1.STATUS_FAILED {
		failed.Add(ctx, 1, providerAttr, attribute.String("reason", failureReason(status)))
		processing.Record(ctx, elapsedMs, providerAttr)
	}
}
