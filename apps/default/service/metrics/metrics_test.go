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

package metrics_test

import (
	"context"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/default/service/metrics"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func dataPointAttrSets(aggregation metricdata.Aggregation) []attribute.Set {
	var out []attribute.Set
	switch d := aggregation.(type) {
	case metricdata.Sum[int64]:
		for _, dp := range d.DataPoints {
			out = append(out, dp.Attributes)
		}
	case metricdata.Sum[float64]:
		for _, dp := range d.DataPoints {
			out = append(out, dp.Attributes)
		}
	case metricdata.Histogram[float64]:
		for _, dp := range d.DataPoints {
			out = append(out, dp.Attributes)
		}
	}
	return out
}

func metricAttrSets(rm metricdata.ResourceMetrics, name string) []attribute.Set {
	var out []attribute.Set
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				out = append(out, dataPointAttrSets(m.Data)...)
			}
		}
	}
	return out
}

func requireAttr(t *testing.T, set attribute.Set, key, want string) {
	t.Helper()
	v, ok := set.Value(attribute.Key(key))
	require.True(t, ok, "attribute %q must be present", key)
	require.Equal(t, want, v.AsString(), "attribute %q", key)
}

// Payment lifecycle metrics must transparently carry tenant_id and
// partition_id from the request context claims alongside their business
// attributes.
func TestPaymentMetricsTenantScoped(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prev := otel.GetMeterProvider()
	otel.SetMeterProvider(provider)
	t.Cleanup(func() { otel.SetMeterProvider(prev) })

	claims := &security.AuthenticationClaims{TenantID: "tenant-x", PartitionID: "part-y"}
	claims.Subject = "user-tenant-x"
	ctx := claims.ClaimsToContext(context.Background())

	pm := metrics.NewPaymentMetrics()

	amount := decimalx.New(12345, -2) // 123.45 major units
	p := &models.Payment{
		RouteID:  "mpesa-route",
		Amount:   amount.Ptr(),
		Currency: "KES",
	}
	p.CreatedAt = time.Now().Add(-time.Second)

	pm.RecordInitiated(ctx, p)
	pm.RecordTerminal(ctx, p, &models.Status{
		EntityType: "payment",
		Status:     int32(commonv1.STATUS_SUCCESSFUL.Number()),
	})
	pm.RecordTerminal(ctx, p, &models.Status{
		EntityType: "payment",
		Status:     int32(commonv1.STATUS_FAILED.Number()),
		Extra:      data.JSONMap{"error": "insufficient funds"},
	})

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	for _, name := range []string{
		"payments_initiated_total",
		"payments_succeeded_total",
		"payments_failed_total",
		"payments_amount_total",
		"payments_processing_duration_ms",
	} {
		sets := metricAttrSets(rm, name)
		require.Len(t, sets, 1, "%s must have exactly one datapoint", name)
		requireAttr(t, sets[0], "tenant_id", "tenant-x")
		requireAttr(t, sets[0], "partition_id", "part-y")
		requireAttr(t, sets[0], "provider", "mpesa-route")
	}

	amountSets := metricAttrSets(rm, "payments_amount_total")
	requireAttr(t, amountSets[0], "currency", "KES")

	failedSets := metricAttrSets(rm, "payments_failed_total")
	requireAttr(t, failedSets[0], "reason", "provider_error")

	// The amount counter must record major units.
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "payments_amount_total" {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[float64])
			require.True(t, ok)
			require.InDelta(t, 123.45, sum.DataPoints[0].Value, 0.0001)
		}
	}
}
