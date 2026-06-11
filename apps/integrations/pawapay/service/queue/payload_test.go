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

//nolint:testpackage // exercises unexported payload helpers
package queue

import (
	"context"
	"errors"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentUUID(t *testing.T) {
	t.Run("existing UUID passes through unchanged", func(t *testing.T) {
		id := "f4401bd2-1568-4140-bf2d-eb77d2b2b639"
		assert.Equal(t, id, paymentUUID(id))
	})

	t.Run("non-UUID ID is mapped deterministically", func(t *testing.T) {
		first := paymentUUID("payment-internal-id-001")
		second := paymentUUID("payment-internal-id-001")
		other := paymentUUID("payment-internal-id-002")

		assert.Equal(t, first, second, "same input must map to the same UUID for idempotency")
		assert.NotEqual(t, first, other)
	})

	t.Run("derived ID is a valid v4 UUID", func(t *testing.T) {
		u, err := uuid.Parse(paymentUUID("payment-internal-id-001"))
		require.NoError(t, err)
		assert.Equal(t, uuid.Version(4), u.Version())
		assert.Equal(t, uuid.RFC4122, u.Variant())
	})
}

func TestFormatMoneyAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   *commonv1.Money
		expected string
	}{
		{name: "nil amount", amount: nil, expected: "0"},
		{name: "whole units", amount: &commonv1.Money{Units: 150}, expected: "150"},
		{name: "two decimal places", amount: &commonv1.Money{Units: 123, Nanos: 450000000}, expected: "123.45"},
		{name: "trailing zero trimmed", amount: &commonv1.Money{Units: 123, Nanos: 400000000}, expected: "123.4"},
		{name: "sub-cent rounds half up", amount: &commonv1.Money{Units: 10, Nanos: 555000000}, expected: "10.56"},
		{name: "rounding carries into units", amount: &commonv1.Money{Units: 9, Nanos: 999000000}, expected: "10"},
		{name: "zero", amount: &commonv1.Money{}, expected: "0"},
		{
			name:     "below one keeps single leading zero",
			amount:   &commonv1.Money{Units: 0, Nanos: 500000000},
			expected: "0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arg interface {
				GetUnits() int64
				GetNanos() int32
			}
			if tt.amount != nil {
				arg = tt.amount
			}
			assert.Equal(t, tt.expected, formatMoneyAmount(arg))
		})
	}
}

func TestSanitizeCustomerMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty input omitted", input: "", expected: ""},
		{name: "valid message preserved", input: "Order 12345", expected: "Order 12345"},
		{name: "special characters dropped", input: "Pay: order #123!", expected: "Pay order 123"},
		{name: "too short after cleaning omitted", input: "a!@#", expected: ""},
		{
			name:     "long message truncated to 22",
			input:    "This is a very long narration indeed",
			expected: "This is a very long na",
		},
		{name: "whitespace collapsed", input: "  Order   12345  ", expected: "Order 12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeCustomerMessage(tt.input)
			assert.Equal(t, tt.expected, got)
			if got != "" {
				assert.GreaterOrEqual(t, len(got), 4)
				assert.LessOrEqual(t, len(got), 22)
			}
		})
	}
}

func TestPaymentMetadata(t *testing.T) {
	md := paymentMetadata("payment", "pay-123", map[string]string{
		"tenant_id":    "tenant-1",
		"partition_id": "partition-1",
	})

	assert.Equal(t, "payment", md["entityType"])
	assert.Equal(t, "pay-123", md["paymentId"])
	assert.Equal(t, "tenant-1", md["tenantId"])
	assert.Equal(t, "partition-1", md["partitionId"])
}

// stubPawapayClient overrides PredictProvider for resolveProvider tests.
type stubPawapayClient struct {
	client.PawapayClient

	prediction *client.ProviderPrediction
	err        error
	called     bool
}

func (s *stubPawapayClient) PredictProvider(
	_ context.Context,
	_ *client.Credentials,
	_ string,
) (*client.ProviderPrediction, error) {
	s.called = true
	if s.err != nil {
		return nil, s.err
	}
	return s.prediction, nil
}

func TestResolveProvider(t *testing.T) {
	t.Run("extra provider takes precedence", func(t *testing.T) {
		stub := &stubPawapayClient{}
		provider, msisdn, err := resolveProvider(
			context.Background(), stub, &client.Credentials{Provider: "DEFAULT_PROVIDER"},
			"EXTRA_PROVIDER", "260763456789",
		)

		require.NoError(t, err)
		assert.Equal(t, "EXTRA_PROVIDER", provider)
		assert.Equal(t, "260763456789", msisdn)
		assert.False(t, stub.called)
	})

	t.Run("credential default used when no extra", func(t *testing.T) {
		stub := &stubPawapayClient{}
		provider, _, err := resolveProvider(
			context.Background(), stub, &client.Credentials{Provider: "DEFAULT_PROVIDER"},
			"", "260763456789",
		)

		require.NoError(t, err)
		assert.Equal(t, "DEFAULT_PROVIDER", provider)
		assert.False(t, stub.called)
	})

	t.Run("prediction used as last resort and sanitises msisdn", func(t *testing.T) {
		stub := &stubPawapayClient{
			prediction: &client.ProviderPrediction{
				Country:     "ZMB",
				Provider:    "MTN_MOMO_ZMB",
				PhoneNumber: "260763456789",
			},
		}
		provider, msisdn, err := resolveProvider(
			context.Background(), stub, &client.Credentials{}, "", "+260 763 456 789",
		)

		require.NoError(t, err)
		assert.Equal(t, "MTN_MOMO_ZMB", provider)
		assert.Equal(t, "260763456789", msisdn)
		assert.True(t, stub.called)
	})

	t.Run("prediction failure surfaces error", func(t *testing.T) {
		stub := &stubPawapayClient{err: errors.New("prediction unavailable")}
		_, _, err := resolveProvider(context.Background(), stub, &client.Credentials{}, "", "12345")

		require.Error(t, err)
	})
}
