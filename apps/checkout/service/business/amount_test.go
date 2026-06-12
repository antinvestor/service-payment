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

package business_test

import (
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		in        string
		units     int64
		nanos     int32
		expectErr bool
	}{
		{in: "150", units: 150},
		{in: "123.45", units: 123, nanos: 450000000},
		{in: "0.5", units: 0, nanos: 500000000},
		{in: "0", expectErr: true},
		{in: "-5", expectErr: true},
		{in: "12.345", expectErr: true},
		{in: "abc", expectErr: true},
		{in: "", expectErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			units, nanos, err := business.ParseAmount(tt.in)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.units, units)
			assert.Equal(t, tt.nanos, nanos)
		})
	}
}

func TestMoneyFromAmount(t *testing.T) {
	m, err := business.MoneyFromAmount("123.45", "KES")
	require.NoError(t, err)
	assert.Equal(t, "KES", m.GetCurrencyCode())
	assert.Equal(t, int64(123), m.GetUnits())
	assert.Equal(t, int32(450000000), m.GetNanos())

	_, err = business.MoneyFromAmount("bad", "KES")
	require.Error(t, err)
}

func TestFormatMoney(t *testing.T) {
	assert.Equal(t, "KES 123.45",
		business.FormatMoney(&commonv1.Money{CurrencyCode: "KES", Units: 123, Nanos: 450000000}))
	assert.Equal(t, "KES 150.00",
		business.FormatMoney(&commonv1.Money{CurrencyCode: "KES", Units: 150}))
}

func TestAmountString(t *testing.T) {
	assert.Equal(t, "123.45", business.AmountString(&commonv1.Money{Units: 123, Nanos: 450000000}))
	assert.Equal(t, "150", business.AmountString(&commonv1.Money{Units: 150}))
}
