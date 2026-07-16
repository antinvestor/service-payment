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
	"strings"

	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
)

// MoMoCorridor maps phone/currency to v4 mobile_money payment method fields.
type MoMoCorridor struct {
	CountryCode string // dialling code without +
	Network     string // MTN | AIRTEL | …
	Currency    string
	// NationalNumber is phone without country code (7–10 digits).
	NationalNumber string
}

// resolveMoMoCorridor maps MSISDN + currency to Flutterwave v4 mobile_money.
func resolveMoMoCorridor(phone, currency string) *MoMoCorridor {
	phone = digitsOnly(phone)
	currency = strings.ToUpper(strings.TrimSpace(currency))

	type rule struct {
		prefix, countryCode, network, currency string
	}
	rules := []rule{
		{"254", "254", "SAFARICOM", "KES"}, // M-Pesa Kenya — network label per FW dashboard
		{"233", "233", "MTN", "GHS"},
		{"256", "256", "MTN", "UGX"},
		{"255", "255", "VODACOM", "TZS"},
		{"250", "250", "MTN", "RWF"},
		{"234", "234", "MTN", "NGN"},
		{"260", "260", "MTN", "ZMW"},
	}
	for _, r := range rules {
		if strings.HasPrefix(phone, r.prefix) || currency == r.currency {
			if currency != "" && currency != r.currency && !strings.HasPrefix(phone, r.prefix) {
				continue
			}
			national := phone
			if strings.HasPrefix(phone, r.prefix) {
				national = strings.TrimPrefix(phone, r.prefix)
			}
			// Drop leading 0 if present on national part.
			national = strings.TrimPrefix(national, "0")
			net := r.network
			// KE often uses M-PESA branding; allow SAFARICOM/MPESA.
			if r.currency == "KES" {
				net = "MPESA"
			}
			return &MoMoCorridor{
				CountryCode:    r.countryCode,
				Network:        net,
				Currency:       r.currency,
				NationalNumber: national,
			}
		}
	}
	return nil
}

// buildMoMoPaymentMethod constructs a v4 payment_method for orchestrator.
func buildMoMoPaymentMethod(c *MoMoCorridor, networkOverride string) client.PaymentMethodInput {
	net := c.Network
	if networkOverride != "" {
		net = strings.ToUpper(networkOverride)
	}
	return client.PaymentMethodInput{
		Type: "mobile_money",
		MobileMoney: &client.MobileMoneyDetails{
			CountryCode: c.CountryCode,
			Network:     net,
			PhoneNumber: c.NationalNumber,
		},
	}
}

// resolveRecipientType picks transfers/recipients type for a currency.
func resolveRecipientType(currency string) string {
	switch strings.ToUpper(currency) {
	case "NGN":
		return "bank_ngn"
	case "KES":
		return "bank_kes"
	case "GHS":
		return "bank_ghs"
	case "UGX":
		return "bank_ugx"
	case "TZS":
		return "bank_tzs"
	case "ZAR":
		return "bank_zar"
	default:
		return "bank_" + strings.ToLower(currency)
	}
}

func digitsOnly(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+")
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// splitName best-effort splits "First Last" into CustomerName for Flutterwave v4.
// v4 requires each name part to be 2–50 chars of letters/spaces/,'-./ only.
func splitName(full string) *client.CustomerName {
	full = strings.TrimSpace(full)
	if full == "" {
		return &client.CustomerName{First: "Customer", Last: "Payer"}
	}
	raw := strings.Fields(full)
	parts := make([]string, 0, len(raw))
	for _, p := range raw {
		if s := sanitizeNamePart(p); s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return &client.CustomerName{First: "Customer", Last: "Payer"}
	}
	n := &client.CustomerName{First: parts[0]}
	if len(parts) == 1 {
		n.Last = parts[0]
		return n
	}
	n.Last = parts[len(parts)-1]
	if len(parts) > 2 {
		// Only set middle when every middle token is valid (v4 rejects short/invalid middle).
		mid := sanitizeNamePart(strings.Join(parts[1:len(parts)-1], " "))
		if len(mid) >= 2 {
			n.Middle = mid
		}
	}
	return n
}

// sanitizeNamePart keeps letters/spaces/,'-./ and enforces min length 2.
func sanitizeNamePart(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
			b.WriteRune(r)
		case r == ' ' || r == ',' || r == '.' || r == '\'' || r == '-':
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	// Collapse repeated spaces
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}
	if len(out) < 2 {
		return ""
	}
	if len(out) > 50 {
		out = out[:50]
	}
	return out
}
