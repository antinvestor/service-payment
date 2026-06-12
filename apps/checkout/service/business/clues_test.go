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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCluesFromProperties(t *testing.T) {
	props, err := structpb.NewStruct(map[string]any{
		"checkout": map[string]any{
			"lastMethod":    "mobile_money",
			"lastProvider":  "mpesa",
			"lastContactId": "contact-1",
			"lastCurrency":  "KES",
		},
	})
	require.NoError(t, err)

	clues := business.CluesFromProperties(props)
	assert.Equal(t, "mpesa", clues.LastProvider)
	assert.Equal(t, "contact-1", clues.LastContactID)
	assert.Equal(t, "KES", clues.LastCurrency)

	assert.Equal(t, business.Clues{}, business.CluesFromProperties(nil))

	empty, err := structpb.NewStruct(map[string]any{"other": "x"})
	require.NoError(t, err)
	assert.Equal(t, business.Clues{}, business.CluesFromProperties(empty))
}

func TestCluesToProperties(t *testing.T) {
	c := business.Clues{
		LastMethod:    "mobile_money",
		LastProvider:  "mpesa",
		LastContactID: "c1",
		LastCurrency:  "KES",
		LastPaidAt:    "2026-06-12T00:00:00Z",
	}
	props := c.ToProperties()
	checkout := props.GetFields()["checkout"].GetStructValue()
	require.NotNil(t, checkout)
	assert.Equal(t, "mpesa", checkout.GetFields()["lastProvider"].GetStringValue())
	assert.Equal(t, "c1", checkout.GetFields()["lastContactId"].GetStringValue())
}

func TestGuestCookieRoundTrip(t *testing.T) {
	secret := []byte("test-secret")
	val := business.EncodeGuestHints(secret, business.GuestHints{Phone: "254712345678", Method: "mpesa"})

	hints, ok := business.DecodeGuestHints(secret, val)
	require.True(t, ok)
	assert.Equal(t, "254712345678", hints.Phone)
	assert.Equal(t, "mpesa", hints.Method)
}

func TestGuestCookieTamperRejected(t *testing.T) {
	secret := []byte("test-secret")
	val := business.EncodeGuestHints(secret, business.GuestHints{Phone: "254712345678", Method: "mpesa"})

	_, ok := business.DecodeGuestHints(secret, val+"x")
	assert.False(t, ok)
	_, ok = business.DecodeGuestHints([]byte("other-secret"), val)
	assert.False(t, ok)
	_, ok = business.DecodeGuestHints(secret, "garbage")
	assert.False(t, ok)
	_, ok = business.DecodeGuestHints(secret, "")
	assert.False(t, ok)
}
