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
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func mustStruct(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	require.NoError(t, err)
	return s
}

func TestNormalizeMSISDN(t *testing.T) {
	tests := map[string]string{
		"+256 701 234 567": "+256701234567",
		"256701234567":     "+256701234567",
		"0256701234567":    "+256701234567",
		"(27) 82-123-4567": "+27821234567",
		"12345":            "",
		"":                 "",
	}
	for in, want := range tests {
		assert.Equal(t, want, normalizeMSISDN(in), in)
	}
}

func TestCorridorForPhone(t *testing.T) {
	tests := []struct {
		phone   string
		country string
		ok      bool
	}{
		{"+256701234567", "UG", true},
		{"+267712345678", "BW", true}, // 267 must win over 27
		{"+27821234567", "ZA", true},
		{"+2348012345678", "NG", true},
		{"+254712345678", "KE", true},
		{"+11234567890", "", false},
	}
	for _, tt := range tests {
		c, ok := corridorForPhone(tt.phone)
		assert.Equal(t, tt.ok, ok, tt.phone)
		assert.Equal(t, tt.country, c.Country, tt.phone)
	}
}

func TestResolveCorridor(t *testing.T) {
	creds := &client.Credentials{Country: "ZM", Currency: "ZMW"}

	c, err := resolveCorridor("+256701234567", nil, "", creds)
	require.NoError(t, err)
	assert.Equal(t, corridor{"UG", "UGX"}, c)

	c, err = resolveCorridor("+256701234567", nil, "usd", creds)
	require.NoError(t, err)
	assert.Equal(t, corridor{"UG", "USD"}, c, "payment currency wins over the table")

	c, err = resolveCorridor("+256701234567", mustStruct(t, map[string]any{"country": "ng", "currency": "ngn"}), "UGX", creds)
	require.NoError(t, err)
	assert.Equal(t, corridor{"NG", "NGN"}, c, "extras win over everything")

	c, err = resolveCorridor("+11234567890", nil, "", creds)
	require.NoError(t, err)
	assert.Equal(t, corridor{"ZM", "ZMW"}, c, "credential defaults as last resort")

	c, err = resolveCorridor("", mustStruct(t, map[string]any{"country": "RW"}), "", &client.Credentials{})
	require.NoError(t, err)
	assert.Equal(t, corridor{"RW", "RWF"}, c, "country from extras with currency from table")

	_, err = resolveCorridor("+11234567890", nil, "", &client.Credentials{})
	require.ErrorIs(t, err, errUnknownCountry)
}

func TestMoneyToLocalAmount(t *testing.T) {
	assert.Equal(t, int64(10), moneyToLocalAmount(&commonv1.Money{Units: 10, Nanos: 490_000_000}))
	assert.Equal(t, int64(11), moneyToLocalAmount(&commonv1.Money{Units: 10, Nanos: 500_000_000}))
	assert.Equal(t, int64(5000), moneyToLocalAmount(&commonv1.Money{Units: 5000}))
	assert.Equal(t, int64(0), moneyToLocalAmount((*commonv1.Money)(nil)))
}

func TestBuildParty(t *testing.T) {
	retail := &client.Credentials{CustomerType: "retail"}

	p := buildParty(mustStruct(t, map[string]any{
		"customer_name": "Jane Doe", "customer_email": "jane@example.com",
	}), "+256701234567", "UG", retail)
	assert.Equal(t, client.Party{Name: "Jane Doe", Email: "jane@example.com", Country: "UG", Phone: "+256701234567"}, p)

	p = buildParty(mustStruct(t, map[string]any{
		"display_name": "Jane Doe", "email": "jane@example.com",
		"customer_address": "Plot 1", "customer_dob": "02/01/1997",
		"customer_id_type": "NIN", "customer_id_number": "123",
		"customer_additional_id_type": "BVN", "customer_additional_id_number": "456",
	}), "+2348012345678", "NG", retail)
	assert.Equal(t, "Jane Doe", p.Name)
	assert.Equal(t, "Plot 1", p.Address)
	assert.Equal(t, "02/01/1997", p.DOB)
	assert.Equal(t, "NIN", p.IDType)
	assert.Equal(t, "123", p.IDNumber)
	assert.Equal(t, "BVN", p.AdditionalIDType)
	assert.Equal(t, "456", p.AdditionalIDNumber)

	inst := &client.Credentials{CustomerType: "institution", BusinessID: "biz", BusinessName: "Acme"}
	p = buildParty(mustStruct(t, map[string]any{"customer_email": "ops@acme.example"}), "+256701234567", "UG", inst)
	assert.Equal(t, client.Party{BusinessID: "biz", BusinessName: "Acme", Email: "ops@acme.example"}, p)
}

func TestCustomerUID(t *testing.T) {
	assert.Equal(t, "c1", customerUID(nil, mustStruct(t, map[string]any{"customer_id": "c1"}), "p1"))
	assert.Equal(t, "p1", customerUID(nil, nil, "p1"))
	assert.Equal(t, "t1", customerUID(map[string]string{"tenant_id": "t1"}, nil, ""))
	assert.Equal(t, "anonymous", customerUID(nil, nil, ""))
}

func TestResolveChannelType(t *testing.T) {
	assert.Equal(t, "bank", resolveChannelType(mustStruct(t, map[string]any{"channel_type": "BANK"}), nil))
	assert.Equal(t, "momo", resolveChannelType(mustStruct(t, map[string]any{"payment_method_type": "mobile_money"}), nil))
	assert.Equal(t, "bank", resolveChannelType(mustStruct(t, map[string]any{"payment_method_type": "bank_transfer"}), nil))
	assert.Equal(t, "momo", resolveChannelType(nil, &client.Credentials{ChannelType: "momo"}))
	assert.Equal(t, "", resolveChannelType(nil, &client.Credentials{ChannelType: "card"}))
	assert.Equal(t, "", resolveChannelType(nil, nil))
}

func TestInstructionsAndAmounts(t *testing.T) {
	assert.Equal(t, "UGX 5,000", formatLocalAmount(5000, "UGX"))
	assert.Equal(t, "NGN 1,234,567", formatLocalAmount(1234567, "NGN"))
	assert.Equal(t, "12", formatLocalAmount(12, ""))

	assert.Equal(t,
		"Approve the UGX 5,000 payment request sent to your phone (MTN Uganda). Enter your mobile money PIN to complete.",
		momoInstruction(5000, "UGX", "MTN Uganda"))

	info := &client.BankInfo{Name: "PAGA", AccountNumber: "0123", AccountName: "Ken Adams"}
	assert.Equal(t,
		"Transfer exactly NGN 9,758 to PAGA account 0123 (Ken Adams) using reference REF1. The account details expire at 2023-02-06T13:41:20.108Z.",
		bankInstruction(info, 9758, "NGN", "REF1", "2023-02-06T13:41:20.108Z"))
	assert.Contains(t, bankInstruction(nil, 10, "ZAR", "", ""), "Transfer ZAR 10")
	assert.Contains(t, bankInstruction(&client.BankInfo{PaymentLink: "https://x"}, 10, "ZAR", "", ""), "secure link")
}

func TestFailureExtras(t *testing.T) {
	e := failureExtras("REFUSED", "")
	assert.Equal(t, "REFUSED", e["failure_code"])
	assert.Equal(t, "The customer did not approve the payment", e["failure_message"])

	e = failureExtras("", "")
	assert.Equal(t, "OTHER_ERROR", e["failure_code"])

	e = providerFailureExtras(&client.APIError{HTTPStatus: 400, Code: "PaymentValidationError", Message: "bad"})
	assert.Equal(t, "PaymentValidationError", e["failure_code"])
	assert.Equal(t, "bad", e["failure_message"])

	assert.True(t, isTerminalFailure("EXPIRED"))
	assert.False(t, isTerminalFailure("pending"))
}

func TestReceiveExtras_Momo(t *testing.T) {
	r := &client.Receive{
		ID: "r1", SequenceID: "p1", Status: "processing", Country: "UG", Currency: "UGX",
		ConvertedAmount: 5000, Rate: 3800.5, Reference: "REF", ExpiresAt: "2030-01-01T00:00:00Z",
		Source: client.Source{NetworkID: "n1"},
	}
	e := receiveExtras(r, "momo", "MTN Uganda")
	assert.Equal(t, "prompt", e["entity_type"])
	assert.Equal(t, "yellowcard", e["provider"])
	assert.Equal(t, "r1", e["receive_id"])
	assert.Equal(t, "n1", e["network_id"])
	assert.Equal(t, "5000", e["local_amount"])
	assert.Equal(t, "3800.5", e["rate"])
	assert.Contains(t, e[collection.ExtraPaymentInstruction], "MTN Uganda")
	_, hasNext := e[collection.ExtraNextAction]
	assert.False(t, hasNext, "momo must not request a checkout next action")
	_, hasBank := e[collection.ExtraBankAccountNumber]
	assert.False(t, hasBank)
}

func TestReceiveExtras_Bank(t *testing.T) {
	r := &client.Receive{
		ID: "r1", SequenceID: "p1", Status: "processing", Country: "NG", Currency: "NGN",
		ConvertedAmount: 9757.8, Reference: "JJ8094861", ExpiresAt: "2023-02-06T17:31:20Z",
		BankInfo: &client.BankInfo{Name: "PAGA", AccountNumber: "01234567890", AccountName: "Ken Adams"},
	}
	e := receiveExtras(r, "bank", "")
	assert.Equal(t, "PAGA", e[collection.ExtraBankName])
	assert.Equal(t, "01234567890", e[collection.ExtraBankAccountNumber])
	assert.Equal(t, "Ken Adams", e[collection.ExtraBankAccountName])
	assert.Equal(t, "JJ8094861", e[collection.ExtraPaymentReference])
	assert.Equal(t, "2023-02-06T17:31:20Z", e[collection.ExtraPaymentExpiresAt])
	assert.Equal(t, "9758", e["local_amount"])
	assert.Contains(t, e[collection.ExtraPaymentInstruction], "01234567890")
	_, hasURL := e[collection.ExtraCheckoutURL]
	assert.False(t, hasURL)

	r.BankInfo = &client.BankInfo{PaymentLink: "https://pay.yellowcard.io/x"}
	e = receiveExtras(r, "bank", "")
	assert.Equal(t, "https://pay.yellowcard.io/x", e[collection.ExtraCheckoutURL])
	assert.Equal(t, "https://pay.yellowcard.io/x", e[collection.ExtraAuthRedirectURL])
	assert.Equal(t, collection.NextActionRedirectURL, e[collection.ExtraNextAction])
}
