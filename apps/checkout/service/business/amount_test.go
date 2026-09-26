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
	tests := []struct {
		name  string
		money *commonv1.Money
		want  string
	}{
		{name: "two decimals", money: &commonv1.Money{CurrencyCode: "KES", Units: 123, Nanos: 450000000}, want: "KES 123.45"},
		{name: "whole", money: &commonv1.Money{CurrencyCode: "KES", Units: 150}, want: "KES 150.00"},
		{name: "sub-cent carries into units", money: &commonv1.Money{CurrencyCode: "KES", Units: 10, Nanos: 995000000}, want: "KES 11.00"},
		{name: "sub-cent rounds down", money: &commonv1.Money{CurrencyCode: "KES", Units: 10, Nanos: 994999999}, want: "KES 10.99"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, business.FormatMoney(tt.money))
		})
	}
}

func TestAmountString(t *testing.T) {
	tests := []struct {
		name  string
		money *commonv1.Money
		want  string
	}{
		{name: "two decimals", money: &commonv1.Money{Units: 123, Nanos: 450000000}, want: "123.45"},
		{name: "whole", money: &commonv1.Money{Units: 150}, want: "150"},
		{name: "one decimal", money: &commonv1.Money{Units: 12, Nanos: 500000000}, want: "12.5"},
		{name: "cents only", money: &commonv1.Money{Units: 0, Nanos: 50000000}, want: "0.05"},
		{name: "0.995 carries to 1", money: &commonv1.Money{Units: 0, Nanos: 995000000}, want: "1"},
		{name: "9.995 carries to 10", money: &commonv1.Money{Units: 9, Nanos: 995000000}, want: "10"},
		{name: "12.345 rounds half up", money: &commonv1.Money{Units: 12, Nanos: 345000000}, want: "12.35"},
		{name: "12.3449 rounds down", money: &commonv1.Money{Units: 12, Nanos: 344900000}, want: "12.34"},
		{name: "12.999999999 carries", money: &commonv1.Money{Units: 12, Nanos: 999999999}, want: "13"},
		{name: "below half a cent is zero", money: &commonv1.Money{Units: 0, Nanos: 4999999}, want: "0"},
		{name: "negative", money: &commonv1.Money{Units: -1, Nanos: -995000000}, want: "-2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := business.AmountString(tt.money)
			assert.Equal(t, tt.want, got)
			if tt.money.GetUnits() >= 0 && tt.want != "0" {
				_, _, err := business.ParseAmount(got)
				require.NoError(t, err, "AmountString output must round-trip through ParseAmount")
			}
		})
	}
}
