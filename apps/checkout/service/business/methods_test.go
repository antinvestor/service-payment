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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMethodsJSON = `[
  {"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"],"countries":["KE"]},
  {"key":"mtn_momo","name":"MTN MoMo","route":"mtn","prefixes":["256","260"],"currencies":["UGX","ZMW"],"countries":["UG","ZM"]},
  {"key":"pawapay","name":"Mobile Money","route":"pawapay","prefixes":[],"currencies":[]},
  {"key":"card","name":"Card","route":"polar","prefixes":[],"currencies":[],"redirect":true}
]`

func TestParseMethodRegistry(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)
	require.Len(t, reg.Methods, 4)
	assert.Equal(t, "mpesa", reg.Methods[0].Key)
	assert.Equal(t, "mtn", reg.Methods[1].Route)

	_, err = business.ParseMethodRegistry("not json")
	require.Error(t, err)

	_, err = business.ParseMethodRegistry("[]")
	require.Error(t, err, "empty registry must be rejected")
}

func TestAvailableMethods(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	all := reg.Available(nil)
	assert.Len(t, all, 4)

	restricted := reg.Available([]string{"mpesa", "unknown"})
	require.Len(t, restricted, 1)
	assert.Equal(t, "mpesa", restricted[0].Key)
}

func TestGetMethod(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	m, ok := reg.Get("mtn_momo")
	require.True(t, ok)
	assert.Equal(t, "MTN MoMo", m.Name)

	_, ok = reg.Get("unknown")
	assert.False(t, ok)
}

func TestPreselect(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)
	methods := reg.Available(nil)

	t.Run("clue wins", func(t *testing.T) {
		m := business.Preselect(methods, "mtn_momo", "254712345678")
		assert.Equal(t, "mtn_momo", m.Key)
	})
	t.Run("phone prefix when no clue", func(t *testing.T) {
		m := business.Preselect(methods, "", "254712345678")
		assert.Equal(t, "mpesa", m.Key)
	})
	t.Run("plus prefix stripped before matching", func(t *testing.T) {
		m := business.Preselect(methods, "", "+260763456789")
		assert.Equal(t, "mtn_momo", m.Key)
	})
	t.Run("first local method as fallback", func(t *testing.T) {
		m := business.Preselect(methods, "", "")
		assert.Equal(t, "mpesa", m.Key)
	})
	t.Run("unknown clue falls through to prefix", func(t *testing.T) {
		m := business.Preselect(methods, "nope", "260763456789")
		assert.Equal(t, "mtn_momo", m.Key)
	})
}

func TestParseMethodRegistry_RedirectFlag(t *testing.T) {
	reg, err := business.ParseMethodRegistry(
		`[{"key":"card","name":"Card","route":"polar","prefixes":[],"currencies":[],"redirect":true}]`,
	)
	require.NoError(t, err)
	require.Len(t, reg.Methods, 1)
	assert.True(t, reg.Methods[0].Redirect)
	assert.Equal(t, "polar", reg.Methods[0].Route)
}

func TestPreselect_SkipsRedirectForPhonePrefix(t *testing.T) {
	reg, err := business.ParseMethodRegistry(
		`[{"key":"card","name":"Card","route":"polar","prefixes":["254"],"currencies":[],"redirect":true},` +
			`{"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]}]`,
	)
	require.NoError(t, err)
	methods := reg.Available(nil)

	m := business.Preselect(methods, "", "254712345678")
	assert.Equal(t, "mpesa", m.Key)

	m = business.Preselect(methods, "card", "254712345678")
	assert.Equal(t, "card", m.Key)
}

func TestResolve_LocationFiltersAndKeepsCard(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	// KE phone → mpesa + universal (pawapay, card); not mtn (UG/ZM).
	res := reg.Resolve(business.MethodFilter{
		Currency: "KES",
		Phone:    "254712345678",
	})
	keys := methodKeys(res.Available)
	assert.Contains(t, keys, "mpesa")
	assert.Contains(t, keys, "card")
	assert.Contains(t, keys, "pawapay")
	assert.NotContains(t, keys, "mtn_momo")
	assert.Equal(t, "mpesa", res.Selected.Key)
	assert.Equal(t, "location_phone", res.Reason)
}

func TestResolve_CachedProfileBeatsLocation(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	// User is in KE but previously paid with card — preselect card.
	res := reg.Resolve(business.MethodFilter{
		Currency:   "KES",
		Phone:      "254712345678",
		ClueMethod: "card",
	})
	assert.Equal(t, "card", res.Selected.Key)
	assert.Equal(t, "cached_profile", res.Reason)
	// Card should be ranked first.
	require.NotEmpty(t, res.Available)
	assert.Equal(t, "card", res.Available[0].Key)
}

func TestResolve_GuestCookieBeatsLocation(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	res := reg.Resolve(business.MethodFilter{
		Currency:    "KES",
		Phone:       "254712345678",
		GuestMethod: "pawapay",
	})
	assert.Equal(t, "pawapay", res.Selected.Key)
	assert.Equal(t, "cached_device", res.Reason)
}

func TestResolve_PartitionAllowlist(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	// Partition only allows card.
	res := reg.Resolve(business.MethodFilter{
		Currency:           "KES",
		Phone:              "254712345678",
		PartitionAllowlist: []string{"card"},
	})
	require.Len(t, res.Available, 1)
	assert.Equal(t, "card", res.Available[0].Key)
	assert.Equal(t, "card", res.Selected.Key)
}

func TestResolve_SessionAndPartitionIntersect(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	res := reg.Resolve(business.MethodFilter{
		Currency:           "KES",
		SessionRestriction: []string{"mpesa", "card"},
		PartitionAllowlist: []string{"card", "pawapay"},
	})
	// Intersection: card only.
	require.Len(t, res.Available, 1)
	assert.Equal(t, "card", res.Available[0].Key)
}

func TestResolve_CountryHeaderLocality(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)

	res := reg.Resolve(business.MethodFilter{
		Currency: "UGX",
		Country:  "UG",
	})
	keys := methodKeys(res.Available)
	assert.Contains(t, keys, "mtn_momo")
	assert.Contains(t, keys, "card")
	assert.Contains(t, keys, "pawapay")
	assert.NotContains(t, keys, "mpesa")
	assert.Equal(t, "location_country", res.Reason)
	assert.Equal(t, "mtn_momo", res.Selected.Key)
}

func TestParsePartitionAllowlists(t *testing.T) {
	p, err := business.ParsePartitionAllowlists(`{"part-a":["mpesa"],"*":["card"]}`)
	require.NoError(t, err)
	assert.Equal(t, []string{"mpesa"}, p.ForPartition("part-a"))
	assert.Equal(t, []string{"card"}, p.ForPartition("unknown"))
	assert.Equal(t, []string{"card"}, p.ForPartition(""))

	empty, err := business.ParsePartitionAllowlists("")
	require.NoError(t, err)
	assert.Nil(t, empty.ForPartition("x"))
}

func TestDetectCountryFromHeaders(t *testing.T) {
	h := func(k string) string {
		if k == "CF-IPCountry" {
			return "ke"
		}
		return ""
	}
	assert.Equal(t, "KE", business.DetectCountryFromHeaders(h))
	assert.Equal(t, "", business.DetectCountryFromHeaders(func(string) string { return "XX" }))
}

func TestInferCountryFromPhone(t *testing.T) {
	assert.Equal(t, "KE", business.InferCountryFromPhone("+254712345678"))
	assert.Equal(t, "UG", business.InferCountryFromPhone("256700000000"))
	assert.Equal(t, "", business.InferCountryFromPhone(""))
}

func methodKeys(methods []business.Method) []string {
	out := make([]string, len(methods))
	for i, m := range methods {
		out[i] = m.Key
	}
	return out
}
