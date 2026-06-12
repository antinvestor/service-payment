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
  {"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]},
  {"key":"mtn_momo","name":"MTN MoMo","route":"mtn","prefixes":["256","260"],"currencies":["UGX","ZMW"]},
  {"key":"pawapay","name":"Mobile Money","route":"pawapay","prefixes":[],"currencies":[]},
  {"key":"card","name":"Card","route":"polar","prefixes":[],"currencies":[],"redirect":true}
]`

const testMethodsJSONNoRedirect = `[
  {"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]},
  {"key":"mtn_momo","name":"MTN MoMo","route":"mtn","prefixes":["256","260"],"currencies":["UGX","ZMW"]},
  {"key":"pawapay","name":"Mobile Money","route":"pawapay","prefixes":[],"currencies":[]}
]`

func TestParseMethodRegistry(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)
	require.Len(t, reg.Methods, 4)
	assert.Equal(t, "mpesa", reg.Methods[0].Key)
	assert.Equal(t, "mtn", reg.Methods[1].Route)
	// card method has redirect=true
	cardMethod, ok := reg.Get("card")
	require.True(t, ok)
	assert.True(t, cardMethod.Redirect, "card method must have Redirect=true")
	// non-redirect methods have Redirect=false
	mpesa, ok := reg.Get("mpesa")
	require.True(t, ok)
	assert.False(t, mpesa.Redirect, "mpesa method must have Redirect=false")

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

	cardMethod, ok := reg.Get("card")
	require.True(t, ok, "card method must be found in registry")
	assert.True(t, cardMethod.Redirect, "card must be a redirect method")

	_, ok = reg.Get("stripe")
	assert.False(t, ok)
}

func TestPreselect(t *testing.T) {
	// Use non-redirect registry for the baseline tests to avoid changing existing behaviour.
	regNoRedirect, err := business.ParseMethodRegistry(testMethodsJSONNoRedirect)
	require.NoError(t, err)
	methods := regNoRedirect.Available(nil)

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
	t.Run("first method as fallback", func(t *testing.T) {
		m := business.Preselect(methods, "", "")
		assert.Equal(t, "mpesa", m.Key)
	})
	t.Run("unknown clue falls through to prefix", func(t *testing.T) {
		m := business.Preselect(methods, "stripe", "260763456789")
		assert.Equal(t, "mtn_momo", m.Key)
	})
}

func TestPreselect_RedirectMethod(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)
	methods := reg.Available(nil)

	t.Run("redirect method never selected by phone prefix", func(t *testing.T) {
		// card has no prefixes, but even if a redirect method had prefixes it must not be
		// matched by phone. Use the full methods list which includes the card (redirect) method.
		// A 254-prefix phone should select mpesa, not card.
		m := business.Preselect(methods, "", "254712345678")
		assert.Equal(t, "mpesa", m.Key, "redirect method must not be selected by phone prefix")
		assert.False(t, m.Redirect, "selected method must not be a redirect method")
	})

	t.Run("explicit clue selects redirect method", func(t *testing.T) {
		m := business.Preselect(methods, "card", "254712345678")
		assert.Equal(t, "card", m.Key, "explicit clue must select the redirect method")
		assert.True(t, m.Redirect, "selected method must be the redirect method")
	})

	t.Run("redirect method as first fallback when no other match", func(t *testing.T) {
		// Only redirect methods available — should still return first one.
		redirectOnly := []business.Method{
			{Key: "card", Name: "Card", Route: "polar", Redirect: true},
		}
		m := business.Preselect(redirectOnly, "", "")
		assert.Equal(t, "card", m.Key, "redirect method returned as first-entry fallback")
	})
}
