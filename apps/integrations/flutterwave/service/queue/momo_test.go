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

package queue //nolint:testpackage // exercises unexported corridor helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveMoMoCorridor(t *testing.T) {
	c := resolveMoMoCorridor("254712345678", "KES")
	require.NotNil(t, c)
	assert.Equal(t, "254", c.CountryCode)
	assert.Equal(t, "712345678", c.NationalNumber)
	assert.Equal(t, "KES", c.Currency)

	c = resolveMoMoCorridor("+2339012345678", "GHS")
	require.NotNil(t, c)
	assert.Equal(t, "233", c.CountryCode)
	assert.Equal(t, "MTN", c.Network)

	assert.Nil(t, resolveMoMoCorridor("", "USD"))
}

func TestBuildMoMoPaymentMethod(t *testing.T) {
	c := resolveMoMoCorridor("254712345678", "KES")
	pm := buildMoMoPaymentMethod(c, "")
	assert.Equal(t, "mobile_money", pm.Type)
	require.NotNil(t, pm.MobileMoney)
	assert.Equal(t, "712345678", pm.MobileMoney.PhoneNumber)
}

func TestSplitName(t *testing.T) {
	n := splitName("King Leo James")
	require.NotNil(t, n)
	assert.Equal(t, "King", n.First)
	assert.Equal(t, "Leo", n.Middle)
	assert.Equal(t, "James", n.Last)
}
