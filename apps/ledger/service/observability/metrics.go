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

// Package observability defines the tenant-scoped business instruments and
// tracing helpers for the ledger service. Instruments are created through
// frame's telemetry.BusinessMetrics factory, so every measurement
// transparently carries tenant_id and partition_id derived from the context
// claims.
package observability

import (
	"context"

	"github.com/pitabwire/frame/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const pkgName = "service_payment_ledger"

// Metrics holds pre-allocated OTel instruments for the ledger service.
// Instruments are created once at startup through the BusinessMetrics
// factory, so every measurement is transparently tenant-scoped.
type Metrics struct {
	tracer telemetry.Tracer

	// Transaction instruments.
	txPostedTotal  telemetry.Counter
	txFailedTotal  telemetry.Counter
	txPostLatency  telemetry.Histogram
	txEntriesTotal telemetry.Counter

	// Account instruments.
	accountCreatedTotal telemetry.Counter

	// Ledger/Book instruments.
	ledgerCreatedTotal telemetry.Counter
	bookCreatedTotal   telemetry.Counter
}

// NewMetrics creates and registers all OTel instruments for the ledger service.
// The underlying OTel SDK deduplicates instruments by name, so multiple
// constructions are safe and record into the same series.
func NewMetrics() *Metrics {
	t := telemetry.NewTracer(pkgName)
	bm := telemetry.NewBusinessMetrics(pkgName)

	return &Metrics{
		tracer: t,

		txPostedTotal: bm.Counter(
			pkgName+"/transactions/posted_total",
			"Transactions successfully posted to the ledger",
		),
		txFailedTotal: bm.Counter(
			pkgName+"/transactions/failed_total",
			"Transactions that failed to post (bounded by reason)",
		),
		txPostLatency: bm.Histogram(
			pkgName+"/transactions/post_latency_ms",
			"Latency of the transaction posting path in milliseconds",
		),
		txEntriesTotal: bm.Counter(
			pkgName+"/transactions/entries_total",
			"Total number of double-entry lines posted",
		),

		accountCreatedTotal: bm.Counter(
			pkgName+"/accounts/created_total",
			"Accounts created",
		),

		ledgerCreatedTotal: bm.Counter(
			pkgName+"/ledgers/created_total",
			"Ledgers created",
		),
		bookCreatedTotal: bm.Counter(
			pkgName+"/books/created_total",
			"Books created",
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

// EndSpan ends a span, records the error status, and is the deferred
// counterpart to StartSpan.
func (m *Metrics) EndSpan(ctx context.Context, span trace.Span, err error) {
	m.tracer.End(ctx, span, err)
}

// RecordTransactionPosted increments the posted counter, records entry count,
// and records post latency. It carries bounded, non-PII span attributes:
// currency, entry_count, and reversal flag.
func (m *Metrics) RecordTransactionPosted(
	ctx context.Context,
	currency string,
	entryCount int,
	reversal bool,
	latencyMs float64,
) {
	if m == nil {
		return
	}
	currencyAttr := attribute.String("currency", currency)
	m.txPostedTotal.Add(ctx, 1, currencyAttr, attribute.Bool("reversal", reversal))
	m.txEntriesTotal.Add(ctx, int64(entryCount), currencyAttr)
	m.txPostLatency.Record(ctx, latencyMs, currencyAttr)
}

// RecordTransactionFailed increments the failed counter with a bounded reason.
// Valid reason values are kept to a small, known set to prevent cardinality
// explosion (e.g. "validation", "conflict", "system").
func (m *Metrics) RecordTransactionFailed(ctx context.Context, reason string) {
	if m == nil {
		return
	}
	m.txFailedTotal.Add(ctx, 1, attribute.String("reason", reason))
}

// RecordAccountCreated increments the account-created counter.
func (m *Metrics) RecordAccountCreated(ctx context.Context) {
	if m == nil {
		return
	}
	m.accountCreatedTotal.Add(ctx, 1)
}

// RecordLedgerCreated increments the ledger-created counter.
func (m *Metrics) RecordLedgerCreated(ctx context.Context) {
	if m == nil {
		return
	}
	m.ledgerCreatedTotal.Add(ctx, 1)
}

// RecordBookCreated increments the book-created counter.
func (m *Metrics) RecordBookCreated(ctx context.Context) {
	if m == nil {
		return
	}
	m.bookCreatedTotal.Add(ctx, 1)
}
