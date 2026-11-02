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

// momoRules are the mobile money corridors, by country calling code. The
// corridor's currency is the only one its wallets are charged in.
var momoRules = []struct {
	prefix, network, currency string
}{
	{"254", "MPESA", "KES"},
	{"233", "MTN", "GHS"},
	{"256", "MTN", "UGX"},
	{"255", "VODACOM", "TZS"},
	{"250", "MTN", "RWF"},
	{"234", "MTN", "NGN"},
	{"260", "MTN", "ZMW"},
}

// networkPrefixes names the network of a national number's leading digits
// (longest prefix wins); the corridor's default network applies otherwise.
var networkPrefixes = map[string]map[string]string{
	"254": {
		"73": "AIRTEL", "750": "AIRTEL", "751": "AIRTEL", "752": "AIRTEL", "753": "AIRTEL", "754": "AIRTEL",
		"755": "AIRTEL", "756": "AIRTEL", "762": "AIRTEL", "78": "AIRTEL", "100": "AIRTEL", "101": "AIRTEL",
		"102": "AIRTEL",
	},
	"256": {"77": "MTN", "78": "MTN", "76": "MTN", "39": "MTN", "70": "AIRTEL", "75": "AIRTEL", "74": "AIRTEL", "20": "AIRTEL"},
	"233": {
		"24": "MTN", "54": "MTN", "55": "MTN", "59": "MTN", "25": "MTN", "53": "MTN",
		"20": "VODAFONE", "50": "VODAFONE",
		"26": "AIRTELTIGO", "56": "AIRTELTIGO", "27": "AIRTELTIGO", "57": "AIRTELTIGO",
	},
	"260": {"96": "MTN", "76": "MTN", "97": "AIRTEL", "77": "AIRTEL", "95": "ZAMTEL", "75": "ZAMTEL"},
	"255": {
		"74": "VODACOM", "75": "VODACOM", "76": "VODACOM",
		"68": "AIRTEL", "69": "AIRTEL", "78": "AIRTEL",
		"65": "TIGO", "67": "TIGO", "71": "TIGO",
	},
	"250": {"78": "MTN", "79": "MTN", "72": "AIRTEL", "73": "AIRTEL"},
}

// resolveMoMoCorridor maps an international MSISDN to its mobile money
// corridor. It never changes the currency: a number whose corridor charges
// another currency than the one asked for has no corridor (nil).
func resolveMoMoCorridor(phone, currency string) *MoMoCorridor {
	phone = strings.TrimLeft(digitsOnly(phone), "0")
	currency = strings.ToUpper(strings.TrimSpace(currency))
	for _, r := range momoRules {
		if !strings.HasPrefix(phone, r.prefix) {
			continue
		}
		if currency != "" && currency != r.currency {
			return nil
		}
		national := strings.TrimPrefix(strings.TrimPrefix(phone, r.prefix), "0")
		network := r.network
		best := 0
		for prefix, name := range networkPrefixes[r.prefix] {
			if len(prefix) > best && strings.HasPrefix(national, prefix) {
				best, network = len(prefix), name
			}
		}
		return &MoMoCorridor{
			CountryCode:    r.prefix,
			Network:        network,
			Currency:       r.currency,
			NationalNumber: national,
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
