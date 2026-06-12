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
  {"key":"pawapay","name":"Mobile Money","route":"pawapay","prefixes":[],"currencies":[]}
]`

func TestParseMethodRegistry(t *testing.T) {
	reg, err := business.ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)
	require.Len(t, reg.Methods, 3)
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
	assert.Len(t, all, 3)

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

	_, ok = reg.Get("card")
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
	t.Run("first method as fallback", func(t *testing.T) {
		m := business.Preselect(methods, "", "")
		assert.Equal(t, "mpesa", m.Key)
	})
	t.Run("unknown clue falls through to prefix", func(t *testing.T) {
		m := business.Preselect(methods, "card", "260763456789")
		assert.Equal(t, "mtn_momo", m.Key)
	})
}
