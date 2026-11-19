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

	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMethodRegistry_FromRoutes(t *testing.T) {
	reg, err := business.BuildMethodRegistry("fincra, mpesa,mtn,airtel", "")
	require.NoError(t, err)
	require.Len(t, reg.Methods, 4)
	assert.Equal(t, "card", reg.Methods[0].Key, "fincra's default method key is card")
	assert.True(t, reg.Methods[0].Redirect)
	assert.True(t, reg.Methods[0].IsCard())
	assert.Equal(t, collection.PaymentMethodTypeCard, reg.Methods[0].ResolvedMethodType())
	mpesa := reg.Methods[1]
	assert.Equal(t, "mpesa", mpesa.Key)
	assert.Equal(t, []string{"KES"}, mpesa.Currencies)
	assert.Equal(t, []string{"254"}, mpesa.Prefixes)
	assert.True(t, mpesa.RequiresWholeAmount())
	assert.False(t, mpesa.IsCard())

	_, err = business.BuildMethodRegistry("fincra,unknownpsp", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknownpsp")

	_, err = business.BuildMethodRegistry("", "")
	require.Error(t, err)
}

func TestBuildMethodRegistry_OverrideMergesWithDescriptor(t *testing.T) {
	reg, err := business.BuildMethodRegistry("fincra,mpesa",
		`[{"key":"card","name":"Visa / Mastercard","route":"fincra","redirect":true},
		  {"key":"polar","route":"polar","redirect":true}]`)
	require.NoError(t, err)
	require.Len(t, reg.Methods, 3)
	card, ok := reg.Get("card")
	require.True(t, ok)
	assert.Equal(t, "Visa / Mastercard", card.Name)
	assert.True(t, card.RequiresWholeAmount(), "fincra's whole-amount rule survives the override")
	assert.Equal(t, collection.PaymentMethodTypeCard, card.ResolvedMethodType())
	polar, ok := reg.Get("polar")
	require.True(t, ok)
	assert.Equal(t, "Polar", polar.Name, "name inherited from the descriptor")
	assert.True(t, polar.RequiresProductID)
	assert.True(t, polar.Subscriptions)

	_, err = business.BuildMethodRegistry("fincra", `[{"key":"x"}]`)
	require.Error(t, err, "an override needs key and route")
}

func TestResolve_DisjointAllowlistsOfferNothing(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)
	res := reg.Resolve(business.MethodFilter{
		Currency:           "KES",
		SessionRestriction: []string{"card"},
		PartitionAllowlist: []string{"mpesa"},
	})
	assert.Empty(t, res.Available)
	assert.Equal(t, "allowlists_disjoint", res.Reason)
}

func TestResolve_ProductRailsNeedProductID(t *testing.T) {
	reg, err := business.BuildMethodRegistry("fincra,polar", "")
	require.NoError(t, err)
	without := reg.Resolve(business.MethodFilter{Currency: "USD"})
	for _, m := range without.Available {
		assert.NotEqual(t, "polar", m.Key, "polar is hidden without a product reference")
	}
	with := reg.Resolve(business.MethodFilter{Currency: "USD", HasProductID: true})
	keys := make([]string, 0, len(with.Available))
	for _, m := range with.Available {
		keys = append(keys, m.Key)
	}
	assert.Contains(t, keys, "polar")
}

func TestMethodForPay_MatchesPageResolution(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	// Partition excludes mpesa: Pay must refuse it even though the key exists.
	_, err = reg.MethodForPay(business.MethodFilter{
		Currency:           "KES",
		PartitionAllowlist: []string{"card"},
	}, "mpesa")
	require.ErrorIs(t, err, business.ErrUnknownMethod)

	// A typed guest phone ranks but does not widen a merchant restriction.
	_, err = reg.MethodForPay(business.MethodFilter{
		Currency:           "KES",
		Phone:              "254700000001",
		SessionRestriction: []string{"card"},
	}, "mpesa")
	require.ErrorIs(t, err, business.ErrUnknownMethod)

	// A profile contact does widen it (the page shows the local rail too).
	m, err := reg.MethodForPay(business.MethodFilter{
		Currency:           "KES",
		Phones:             []string{"254700000001"},
		SessionRestriction: []string{"card"},
	}, "mpesa")
	require.NoError(t, err)
	assert.Equal(t, "mpesa", m.Key)

	_, err = reg.MethodForPay(business.MethodFilter{Currency: "USD"}, "mpesa")
	require.ErrorIs(t, err, business.ErrMethodCurrencyUnsupported)
	_, err = reg.MethodForPay(business.MethodFilter{Currency: "KES", Amount: "10.50"}, "mpesa")
	require.ErrorIs(t, err, business.ErrMethodAmountUnsupported)
	_, err = reg.MethodForPay(business.MethodFilter{Currency: "KES"}, "nope")
	require.ErrorIs(t, err, business.ErrUnknownMethod)
}
