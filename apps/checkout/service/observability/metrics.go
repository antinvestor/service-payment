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

// Package observability defines the telemetry instruments for the checkout
// service. Instruments are created through frame's telemetry factory so every
// measurement transparently carries tenant_id and partition_id derived from
// the context security claims.
package observability

import (
	"context"

	"github.com/pitabwire/frame/v2/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const pkgName = "service_payment_checkout"

// Metrics holds pre-allocated OTel instruments for the checkout service.
// Construct once at startup via NewMetrics; the underlying SDK deduplicates
// instruments by name so multiple constructions are safe.
type Metrics struct {
	tracer telemetry.Tracer

	// Session instruments.
	sessionsCreated     telemetry.Counter
	sessionsSpawnedLink telemetry.Counter

	// Link instruments.
	linksCreated telemetry.Counter

	// Pay instruments.
	payAttempts  telemetry.Counter
	payFailures  telemetry.Counter
	payLatencyMs telemetry.Histogram

	// Outcome instruments.
	outcomesCompleted telemetry.Counter
	outcomesFailed    telemetry.Counter

	// Sweep instruments.
	sweepLatencyMs telemetry.Histogram
	sweepExpired   telemetry.Counter

	// Clue write-back instruments.
	clueWritebackFailures telemetry.Counter
}

// NewMetrics creates and registers all OTel instruments for the checkout service.
func NewMetrics() *Metrics {
	t := telemetry.NewTracer(pkgName)
	bm := telemetry.NewBusinessMetrics(pkgName)

	return &Metrics{
		tracer: t,

		sessionsCreated: bm.Counter(
			pkgName+"/sessions/created",
			"Number of checkout sessions created",
		),
		sessionsSpawnedLink: bm.Counter(
			pkgName+"/sessions/spawned_from_link",
			"Number of checkout sessions spawned from a reusable link",
		),

		linksCreated: bm.Counter(
			pkgName+"/links/created",
			"Number of reusable checkout links created",
		),

		payAttempts: bm.Counter(
			pkgName+"/pay/attempts",
			"Number of pay attempts that passed initial guards",
		),
		payFailures: bm.Counter(
			pkgName+"/pay/failures",
			"Number of pay attempts that failed after guard-pass, labelled by reason",
		),
		payLatencyMs: bm.Histogram(
			pkgName+"/pay/latency",
			"Latency of the Pay method in milliseconds",
		),

		outcomesCompleted: bm.Counter(
			pkgName+"/outcomes/completed",
			"Number of checkout sessions that reached the completed status",
		),
		outcomesFailed: bm.Counter(
			pkgName+"/outcomes/failed",
			"Number of checkout sessions that reached the failed status",
		),

		sweepLatencyMs: bm.Histogram(
			pkgName+"/sweep/latency",
			"Latency of a single SweepProcessing run in milliseconds",
		),
		sweepExpired: bm.Counter(
			pkgName+"/sweep/expired",
			"Number of sessions expired by the sweep worker",
		),

		clueWritebackFailures: bm.Counter(
			pkgName+"/clue_writeback/failures",
			"Number of profile clue write-back errors (best-effort, not fatal)",
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

// EndSpan ends a span, recording any error onto the span.
func (m *Metrics) EndSpan(ctx context.Context, span trace.Span, err error) {
	m.tracer.End(ctx, span, err)
}

// RecordSessionCreated increments the sessions/created counter.
// amountOption is "fixed" or "variable"; hasPayer indicates whether a payer
// profile was attached. No PII attributes are allowed here.
func (m *Metrics) RecordSessionCreated(ctx context.Context, amountOption string, hasPayer bool) {
	if m == nil {
		return
	}
	m.sessionsCreated.Add(ctx, 1,
		attribute.String("amount_option", amountOption),
		attribute.Bool("has_payer", hasPayer),
	)
}

// RecordSessionSpawned increments the sessions/spawned_from_link counter.
func (m *Metrics) RecordSessionSpawned(ctx context.Context) {
	if m == nil {
		return
	}
	m.sessionsSpawnedLink.Add(ctx, 1)
}

// RecordLinkCreated increments the links/created counter.
func (m *Metrics) RecordLinkCreated(ctx context.Context) {
	if m == nil {
		return
	}
	m.linksCreated.Add(ctx, 1)
}

// RecordPayAttempt increments the pay/attempts counter.
// methodKey is the payment method key (e.g. "mpesa"); variable indicates
// whether it is a variable-amount session.
func (m *Metrics) RecordPayAttempt(ctx context.Context, methodKey string, variable bool) {
	if m == nil {
		return
	}
	m.payAttempts.Add(ctx, 1,
		attribute.String("method", methodKey),
		attribute.Bool("variable", variable),
	)
}

// RecordPayFailure increments the pay/failures counter with a bounded reason label.
// reason must be one of the sentinel names defined in the business package
// (session_gone, too_many_attempts, cooldown, unknown_method, amount_required,
// contact_required, prompt_error) or "unknown".
func (m *Metrics) RecordPayFailure(ctx context.Context, reason string) {
	if m == nil {
		return
	}
	m.payFailures.Add(ctx, 1, attribute.String("reason", reason))
}

// RecordPayLatency records the Pay method duration in milliseconds.
func (m *Metrics) RecordPayLatency(ctx context.Context, elapsedMs float64, methodKey string) {
	if m == nil {
		return
	}
	m.payLatencyMs.Record(ctx, elapsedMs, attribute.String("method", methodKey))
}

// RecordOutcomeCompleted increments the outcomes/completed counter.
func (m *Metrics) RecordOutcomeCompleted(ctx context.Context) {
	if m == nil {
		return
	}
	m.outcomesCompleted.Add(ctx, 1)
}

// RecordOutcomeFailed increments the outcomes/failed counter.
func (m *Metrics) RecordOutcomeFailed(ctx context.Context) {
	if m == nil {
		return
	}
	m.outcomesFailed.Add(ctx, 1)
}

// RecordSweepLatency records the SweepProcessing run duration in milliseconds.
func (m *Metrics) RecordSweepLatency(ctx context.Context, elapsedMs float64) {
	if m == nil {
		return
	}
	m.sweepLatencyMs.Record(ctx, elapsedMs)
}

// RecordSweepExpired increments the sweep/expired counter.
func (m *Metrics) RecordSweepExpired(ctx context.Context) {
	if m == nil {
		return
	}
	m.sweepExpired.Add(ctx, 1)
}

// RecordClueWritebackFailure increments the clue_writeback/failures counter.
func (m *Metrics) RecordClueWritebackFailure(ctx context.Context) {
	if m == nil {
		return
	}
	m.clueWritebackFailures.Add(ctx, 1)
}
