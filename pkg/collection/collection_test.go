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

package collection_test

import (
	"context"
	"errors"
	"net"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/require"
)

func TestMinorUnits(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		money   *commonv1.Money
		want    int64
		wantErr error
	}{
		{"two decimals", &commonv1.Money{CurrencyCode: "KES", Units: 1500, Nanos: 500_000_000}, 150050, nil},
		{"zero decimal currency", &commonv1.Money{CurrencyCode: "UGX", Units: 1500}, 1500, nil},
		{"zero decimal refuses cents", &commonv1.Money{CurrencyCode: "UGX", Units: 1, Nanos: 500_000_000}, 0, collection.ErrFractionalAmount},
		{"three decimals", &commonv1.Money{CurrencyCode: "KWD", Units: 1, Nanos: 5_000_000}, 1005, nil},
		{"sub-cent refused", &commonv1.Money{CurrencyCode: "USD", Units: 1, Nanos: 1_000_000}, 0, collection.ErrFractionalAmount},
		{"nil", nil, 0, collection.ErrNonPositiveAmount},
		{"zero", &commonv1.Money{CurrencyCode: "USD"}, 0, collection.ErrNonPositiveAmount},
		{"negative", &commonv1.Money{CurrencyCode: "USD", Units: -1}, 0, collection.ErrNonPositiveAmount},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := collection.MinorUnits(tc.money)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestWholeUnitsAndDecimalString(t *testing.T) {
	t.Parallel()
	_, err := collection.WholeUnits(&commonv1.Money{CurrencyCode: "KES", Units: 10, Nanos: 500_000_000})
	require.ErrorIs(t, err, collection.ErrFractionalAmount)
	u, err := collection.WholeUnits(&commonv1.Money{CurrencyCode: "KES", Units: 10})
	require.NoError(t, err)
	require.Equal(t, int64(10), u)

	require.Equal(t, "1500.50", collection.DecimalString(&commonv1.Money{CurrencyCode: "KES", Units: 1500, Nanos: 500_000_000}))
	require.Equal(t, "1500.00", collection.DecimalString(&commonv1.Money{CurrencyCode: "KES", Units: 1500}))
	require.Equal(t, "1500", collection.DecimalString(&commonv1.Money{CurrencyCode: "UGX", Units: 1500}))
	require.Equal(t, "0.05", collection.DecimalString(&commonv1.Money{CurrencyCode: "USD", Nanos: 50_000_000}))

	m := collection.FromMinorUnits(150050, "kes")
	require.Equal(t, "KES", m.GetCurrencyCode())
	require.Equal(t, int64(1500), m.GetUnits())
	require.Equal(t, int32(500_000_000), m.GetNanos())
	m = collection.FromMinorUnits(1500, "UGX")
	require.Equal(t, int64(1500), m.GetUnits())
	require.Equal(t, int32(0), m.GetNanos())
}

func TestIsUncertain(t *testing.T) {
	t.Parallel()
	require.False(t, collection.IsUncertain(nil))
	require.False(t, collection.IsUncertain(errors.New("provider refused: insufficient funds")))
	require.True(t, collection.IsUncertain(collection.Uncertain(errors.New("502 bad gateway"))))
	require.True(t, collection.IsUncertain(context.DeadlineExceeded))
	require.True(t, collection.IsUncertain(&net.OpError{Op: "dial", Err: errors.New("timeout")}))
	wrapped := errors.New("outer")
	require.ErrorIs(t, collection.Uncertain(wrapped), wrapped)
}

func TestProviderRegistry(t *testing.T) {
	t.Parallel()
	seenRoutes := map[string]bool{}
	seenKeys := map[string]bool{}
	for _, p := range collection.Providers {
		require.NotEmpty(t, p.Route)
		require.NotEmpty(t, p.MethodKey)
		require.NotEmpty(t, p.Name)
		require.False(t, seenRoutes[p.Route], "duplicate route %s", p.Route)
		require.False(t, seenKeys[p.MethodKey], "duplicate method key %s", p.MethodKey)
		seenRoutes[p.Route] = true
		seenKeys[p.MethodKey] = true
		require.Equal(t, len(p.Countries) > 0, len(p.Prefixes) > 0, "%s: countries and prefixes go together", p.Route)
	}
	p, ok := collection.ProviderByRoute("MPESA")
	require.True(t, ok)
	require.True(t, p.WholeAmounts)
	require.Equal(t, collection.FlowPush, p.Flow)
	_, ok = collection.ProviderByRoute("nope")
	require.False(t, ok)
	card, ok := collection.ProviderByMethodKey("card")
	require.True(t, ok)
	require.Equal(t, "fincra", card.Route)
	require.True(t, collection.RouteRequiresWholeAmount("fincra"))
	require.False(t, collection.RouteRequiresWholeAmount("stripe"))
	subs := collection.SubscriptionProviders()
	require.NotEmpty(t, subs)
	for _, s := range subs {
		require.True(t, s.Subscriptions)
	}
}
