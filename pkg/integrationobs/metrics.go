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

// Package integrationobs provides shared observability instrumentation for
// payment-provider integrations. One Metrics instance per integration service,
// labelled by bounded provider + operation attributes.
package integrationobs

import (
	"context"
	"time"

	"github.com/pitabwire/frame/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

const pkgName = "payment_integration"

// Metrics instruments a payment-provider integration: outbound provider API
// calls and inbound queue message handling. One instance per integration
// service, labelled by bounded provider + operation attributes.
type Metrics struct {
	tracer           telemetry.Tracer
	providerAttr     attribute.KeyValue
	providerLatency  telemetry.Histogram
	providerErrors   telemetry.Counter
	providerRequests telemetry.Counter
	queueProcessed   telemetry.Counter
	queueFailed      telemetry.Counter
	webhookReceived  telemetry.Counter
	webhookRejected  telemetry.Counter
}

// NewMetrics constructs a Metrics handle for the named provider integration.
// Instruments are namespaced under "payment_integration/...".
// The OTel SDK deduplicates instruments by name so multiple constructions are safe.
func NewMetrics(provider string) *Metrics {
	t := telemetry.NewTracer(pkgName)
	bm := telemetry.NewBusinessMetrics(pkgName)
	return &Metrics{
		tracer:       t,
		providerAttr: attribute.String("provider", provider),
		providerLatency: bm.Histogram(
			"provider_call_duration_ms",
			"Duration of outbound provider API calls in milliseconds",
		),
		providerErrors: bm.Counter(
			"provider_call_errors_total",
			"Outbound provider API calls that returned an error",
		),
		providerRequests: bm.Counter(
			"provider_calls_total",
			"Total outbound provider API calls",
		),
		queueProcessed: bm.Counter(
			"queue_messages_processed_total",
			"Queue messages processed successfully",
		),
		queueFailed: bm.Counter(
			"queue_messages_failed_total",
			"Queue messages that failed processing",
		),
		webhookReceived: bm.Counter(
			"webhook_callbacks_received_total",
			"Inbound webhook callbacks received",
		),
		webhookRejected: bm.Counter(
			"webhook_callbacks_rejected_total",
			"Inbound webhook callbacks rejected",
		),
	}
}

// ObserveProviderCall wraps a provider HTTP call. It starts a trace span and
// returns (ctx, done). The caller must call done(err) when the call completes.
// done records latency, increments the request counter, and on non-nil error
// also increments the error counter. The span is ended by done.
func (m *Metrics) ObserveProviderCall(ctx context.Context, operation string) (context.Context, func(error)) {
	if m == nil {
		return ctx, func(error) {}
	}
	opAttr := attribute.String("operation", operation)
	ctx, span := m.tracer.Start(ctx, pkgName+"/"+operation)
	start := time.Now()
	return ctx, func(err error) {
		elapsed := float64(time.Since(start).Milliseconds())
		m.providerRequests.Add(ctx, 1, m.providerAttr, opAttr)
		m.providerLatency.Record(ctx, elapsed, m.providerAttr, opAttr)
		if err != nil {
			m.providerErrors.Add(ctx, 1, m.providerAttr, opAttr)
		}
		m.tracer.End(ctx, span, err)
	}
}

// QueueProcessed records a successfully processed queue message of the given kind
// ("payment" or "prompt").
func (m *Metrics) QueueProcessed(ctx context.Context, kind string) {
	if m == nil {
		return
	}
	m.queueProcessed.Add(ctx, 1, m.providerAttr, attribute.String("kind", kind))
}

// QueueFailed records a queue message that failed processing.
// reason should be one of: unmarshal_error, credentials_error, provider_error, rejected.
func (m *Metrics) QueueFailed(ctx context.Context, kind, reason string) {
	if m == nil {
		return
	}
	m.queueFailed.Add(ctx, 1, m.providerAttr, attribute.String("kind", kind), attribute.String("reason", reason))
}

// WebhookReceived records an inbound webhook callback of the given kind.
func (m *Metrics) WebhookReceived(ctx context.Context, kind string) {
	if m == nil {
		return
	}
	m.webhookReceived.Add(ctx, 1, m.providerAttr, attribute.String("kind", kind))
}

// WebhookRejected records a webhook callback rejected due to reason.
// reason should be one of: decode_error, missing_id, verification_failed, unknown_payment, status_update_error.
func (m *Metrics) WebhookRejected(ctx context.Context, kind, reason string) {
	if m == nil {
		return
	}
	m.webhookRejected.Add(ctx, 1, m.providerAttr, attribute.String("kind", kind), attribute.String("reason", reason))
}
