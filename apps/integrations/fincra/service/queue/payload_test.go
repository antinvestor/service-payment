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

package queue

import (
	"context"
	"encoding/json"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/pitabwire/frame/v3/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNormalizeMSISDN(t *testing.T) {
	assert.Equal(t, "254700000000", normalizeMSISDN("+254 700 000 000"))
	assert.Equal(t, "233240000000", normalizeMSISDN("00233240000000"))
	assert.Empty(t, normalizeMSISDN("0700"))
	assert.Empty(t, normalizeMSISDN("user@example.com"))
}

func TestDetectNetwork(t *testing.T) {
	tests := map[string]string{
		"254700000000":  "MPESA",
		"254110000000":  "MPESA",
		"254731234567":  "AIRTEL",
		"254751234567":  "AIRTEL",
		"254758000000":  "MPESA",
		"254763000000":  "EQUITEL",
		"256770000000":  "MTN",
		"256701234567":  "AIRTEL",
		"233240000000":  "MTN",
		"233201234567":  "VODAFONE",
		"233541234567":  "MTN",
		"233561234567":  "AIRTELTIGO",
		"260970000000":  "AIRTEL",
		"260951234567":  "ZAMTEL",
		"255751234567":  "MPESA",
		"255671234567":  "TIGO",
		"250781234567":  "MTN",
		"2250701234567": "",
	}
	for msisdn, want := range tests {
		assert.Equal(t, want, detectNetwork(msisdn), msisdn)
	}
}

func TestCorridorForPhone(t *testing.T) {
	c, ok := corridorForPhone("2250701234567")
	require.True(t, ok)
	assert.Equal(t, corridor{"CI", "XOF"}, c)
	c, ok = corridorForPhone("201055555000")
	require.True(t, ok)
	assert.Equal(t, "EG", c.Country)
	_, ok = corridorForPhone("4479460000")
	assert.False(t, ok)
}

func TestWholeUnits(t *testing.T) {
	v, ok := wholeUnits(&commonv1.Money{CurrencyCode: "NGN", Units: 350})
	assert.True(t, ok)
	assert.Equal(t, int64(350), v)
	_, ok = wholeUnits(&commonv1.Money{Units: 10, Nanos: 500_000_000})
	assert.False(t, ok, "a fractional amount is refused, never rounded")
	_, ok = wholeUnits(&commonv1.Money{})
	assert.False(t, ok)
	_, ok = wholeUnits((*commonv1.Money)(nil))
	assert.False(t, ok)
}

func TestPayoutAmount(t *testing.T) {
	tests := []struct {
		name  string
		money *commonv1.Money
		whole bool
		want  json.Number
		ok    bool
	}{
		{"whole", &commonv1.Money{Units: 1000}, false, "1000", true},
		{"cents", &commonv1.Money{Units: 10, Nanos: 50_000_000}, false, "10.05", true},
		{"half", &commonv1.Money{Units: 10, Nanos: 500_000_000}, false, "10.50", true},
		{"sub-cent refused", &commonv1.Money{Units: 10, Nanos: 5_000_000}, false, "", false},
		{"fraction where whole required", &commonv1.Money{Units: 10, Nanos: 500_000_000}, true, "", false},
		{"whole where whole required", &commonv1.Money{Units: 10}, true, "10", true},
		{"zero", &commonv1.Money{}, false, "", false},
		{"negative", &commonv1.Money{Units: -5}, false, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := payoutAmount(tt.money, tt.whole)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
	assert.True(t, requiresWholePayout("UGX", "MTN"))
	assert.True(t, requiresWholePayout("KES", "AIRTEL"))
	assert.False(t, requiresWholePayout("KES", "MPESA"))
}

func TestPlainText(t *testing.T) {
	in := `<p>You will receive a prompt on mobile number <b>254700000000</b>. <br/> Enter your M-Pesa PIN.</p>` +
		` <ol> <li>Go to your SIM Tool Kit</li> <li>Select M-PESA &amp; pay</li> </ol>`
	assert.Equal(t,
		"You will receive a prompt on mobile number 254700000000.\nEnter your M-Pesa PIN.\nGo to your SIM Tool Kit\nSelect M-PESA & pay",
		plainText(in))
	assert.Equal(t, "Enter the OTP sent to your phone", plainText("Enter the OTP sent to your phone"))
}

func TestWebhookURL(t *testing.T) {
	assert.Empty(t, webhookURL("", "t1", "p1", ""))
	assert.Equal(t, "https://pay.example.com/webhook/fincra", webhookURL("https://pay.example.com/", "", "", ""))
	assert.Equal(t, "https://pay.example.com/webhook/fincra?connection=acme&partition_id=p1&tenant_id=t1",
		webhookURL("https://pay.example.com", "t1", "p1", "acme"))
}

func TestCollectionMethod(t *testing.T) {
	method := func(m map[string]any) string {
		s, err := structpb.NewStruct(m)
		require.NoError(t, err)
		return collectionMethod(s)
	}
	assert.Equal(t, "mobile_money", method(map[string]any{"payment_method_type": "mobile_money"}))
	assert.Equal(t, "mobile_money", method(map[string]any{"channel_type": "momo"}))
	assert.Equal(t, "bank_transfer", method(map[string]any{"channel_type": "bank"}))
	assert.Equal(t, "card", method(map[string]any{"payment_method_type": "card"}))
	assert.Empty(t, method(map[string]any{}))
}

func TestBuildCustomer(t *testing.T) {
	extra, _ := structpb.NewStruct(map[string]any{"customer_name": "Ada L", "customer_email": "ada@example.com"})
	c, err := buildCustomer(extra, "254700000000", "fallback@example.com")
	require.NoError(t, err)
	assert.Equal(t, "Ada L", c.Name)
	assert.Equal(t, "ada@example.com", c.Email)
	assert.Equal(t, "254700000000", c.PhoneNumber)

	c, err = buildCustomer(nil, "", "fallback@example.com")
	require.NoError(t, err)
	assert.Equal(t, "fallback@example.com", c.Email)
	assert.Equal(t, "Customer", c.Name)

	_, err = buildCustomer(nil, "", "")
	assert.ErrorIs(t, err, errMissingEmail)
}

func TestTenancy(t *testing.T) {
	claims := &security.AuthenticationClaims{TenantID: "t-claims", PartitionID: "p-claims"}
	ctx := claims.ClaimsToContext(context.Background())

	tenant, partition := tenancy(ctx, map[string]string{"tenant_id": "t-header"})
	assert.Equal(t, "t-header", tenant)
	assert.Equal(t, "p-claims", partition)

	tenant, partition = tenancy(context.Background(), nil)
	assert.Empty(t, tenant)
	assert.Empty(t, partition)
}

func TestSplitName(t *testing.T) {
	first, last := splitName("Mary Jane Atieno")
	assert.Equal(t, "Mary Jane", first)
	assert.Equal(t, "Atieno", last)
	first, last = splitName("Cher")
	assert.Equal(t, "Cher", first)
	assert.Equal(t, "Cher", last)
}

func TestParseCodeMap(t *testing.T) {
	m := parseCodeMap(`{"kes":"MPESA","mtn":" MTN_GH ","x":""}`)
	assert.Equal(t, map[string]string{"KES": "MPESA", "MTN": "MTN_GH"}, m)
	assert.Empty(t, parseCodeMap("not json"))
}
