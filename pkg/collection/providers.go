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

package collection

import "strings"

// Flow is how the payer completes a collection on a rail.
type Flow string

const (
	// FlowPush sends a prompt (STK, USSD, requestToPay) to the payer's phone;
	// the payer approves on the device and the adapter reports the outcome.
	FlowPush Flow = "push"
	// FlowRedirect hands the payer to a provider-hosted page: the adapter
	// returns ExtraCheckoutURL and the outcome arrives by webhook.
	FlowRedirect Flow = "redirect"
	// FlowEmbed collects the instrument on our page (encrypted card fields)
	// and the adapter charges it directly, possibly with a 3DS redirect.
	FlowEmbed Flow = "embed"
)

// Provider describes one payment integration: the route key the payment
// service publishes prompts and payouts on, how a collection completes, and
// the rails and currencies it serves.
//
// Product apps (checkout, billing) read this registry instead of hardcoding
// provider names, so adding an integration is one entry here plus the worker.
type Provider struct {
	// Route is the payment route key: the prompt queue key in
	// INITIATE_PROMPT_ROUTE_URIS and the Route.Name of payout routes.
	Route string
	// Name is the human label shown on the hosted page.
	Name string
	// MethodKey is the default checkout method key for this provider.
	MethodKey string
	// MethodType is the portable instrument class (PaymentMethodType*).
	MethodType string
	Flow       Flow

	// Currencies the provider collects in (empty = any).
	Currencies []string
	// Countries the provider serves (ISO 3166-1 alpha-2; empty = any).
	Countries []string
	// Prefixes are E.164 country prefixes of the mobile money networks the
	// provider reaches (empty = not phone-local).
	Prefixes []string

	// WholeAmounts marks rails that only charge whole currency units.
	WholeAmounts bool
	// Collect reports whether the provider takes money (prompts).
	Collect bool
	// Payout reports whether the provider sends money (payments queue).
	Payout bool
	// Subscriptions reports provider-managed recurring billing: the
	// provider charges renewals itself and reports them by webhook.
	Subscriptions bool
	// SavedInstrument reports card-on-file charges with
	// ExtraPaymentMethodID + ExtraCustomerID (silent renewals).
	SavedInstrument bool
	// Refunds reports provider refunds.
	Refunds bool
	// RequiresProductID marks providers that charge a catalog product
	// (ExtraProductID) rather than an arbitrary amount.
	RequiresProductID bool
	// Dormant providers are built but not offered: they are listed for
	// documentation and never enabled unless named explicitly.
	Dormant bool
}

// IsUniversal reports whether the provider has no locality constraint.
func (p Provider) IsUniversal() bool {
	return len(p.Prefixes) == 0 && len(p.Countries) == 0
}

// Providers is the registry of every integration under apps/integrations.
// Order is the default display order on the hosted page.
var Providers = []Provider{
	{
		Route: "fincra", Name: "Card", MethodKey: "card",
		MethodType: PaymentMethodTypeCard, Flow: FlowRedirect,
		WholeAmounts: true, Collect: true, Payout: true,
	},
	{
		Route: "mpesa", Name: "M-PESA", MethodKey: "mpesa",
		MethodType: PaymentMethodTypeMobileMoney, Flow: FlowPush,
		Currencies: []string{"KES"}, Countries: []string{"KE"}, Prefixes: []string{"254"},
		WholeAmounts: true, Collect: true, Payout: true,
	},
	{
		Route: "mtn", Name: "MTN MoMo", MethodKey: "mtn_momo",
		MethodType: PaymentMethodTypeMobileMoney, Flow: FlowPush,
		Currencies:   []string{"UGX", "ZMW", "GHS", "RWF", "XAF", "XOF"},
		Countries:    []string{"UG", "ZM", "GH", "RW", "CM", "CI", "BJ"},
		Prefixes:     []string{"256", "260", "233", "250", "237", "225", "229"},
		WholeAmounts: true, Collect: true, Payout: true,
	},
	{
		Route: "airtel", Name: "Airtel Money", MethodKey: "airtel_money",
		MethodType: PaymentMethodTypeMobileMoney, Flow: FlowPush,
		Currencies:   []string{"TZS", "UGX", "KES", "ZMW", "MWK", "RWF"},
		Countries:    []string{"TZ", "UG", "KE", "ZM", "MW", "RW"},
		Prefixes:     []string{"255", "256", "254", "260", "265", "250"},
		WholeAmounts: true, Collect: true, Payout: true,
	},
	{
		Route: "pawapay", Name: "Mobile Money", MethodKey: "pawapay",
		MethodType: PaymentMethodTypeMobileMoney, Flow: FlowPush,
		Countries: []string{"BJ", "BF", "CM", "CD", "CG", "CI", "GA", "GH", "KE", "MW", "ML", "MZ", "NG", "RW", "SN", "SL", "TZ", "UG", "ZM"},
		Prefixes:  []string{"229", "226", "237", "243", "242", "225", "241", "233", "254", "265", "223", "258", "234", "250", "221", "232", "255", "256", "260"},
		Collect:   true, Payout: true, Refunds: true,
	},
	{
		Route: "yellowcard", Name: "Yellow Card", MethodKey: "yellowcard",
		MethodType: PaymentMethodTypeMobileMoney, Flow: FlowPush,
		WholeAmounts: true, Collect: true, Payout: true, Dormant: true,
	},
	{
		Route: "stripe", Name: "Card (Stripe)", MethodKey: "stripe",
		MethodType: PaymentMethodTypeCard, Flow: FlowRedirect,
		Collect: true, SavedInstrument: true, Subscriptions: true, Refunds: true,
	},
	{
		Route: "polar", Name: "Polar", MethodKey: "polar",
		MethodType: PaymentMethodTypeCard, Flow: FlowRedirect,
		Collect: true, Subscriptions: true, RequiresProductID: true,
	},
	{
		Route: "jenga", Name: "Jenga", MethodKey: "jenga",
		MethodType: PaymentMethodTypeMobileMoney, Flow: FlowPush,
		Currencies: []string{"KES"}, Countries: []string{"KE"}, Prefixes: []string{"254"},
		Collect: true, Payout: true, Dormant: true,
	},
}

// ProviderByRoute returns the descriptor for a payment route key.
func ProviderByRoute(route string) (Provider, bool) {
	route = strings.ToLower(strings.TrimSpace(route))
	for _, p := range Providers {
		if p.Route == route {
			return p, true
		}
	}
	return Provider{}, false
}

// ProviderByMethodKey returns the descriptor whose default method key matches.
func ProviderByMethodKey(key string) (Provider, bool) {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, p := range Providers {
		if p.MethodKey == key {
			return p, true
		}
	}
	return Provider{}, false
}

// RouteRequiresWholeAmount reports whether a route only charges whole units.
// Unknown routes are not restricted.
func RouteRequiresWholeAmount(route string) bool {
	p, ok := ProviderByRoute(route)
	return ok && p.WholeAmounts
}

// SubscriptionProviders lists routes whose provider manages recurring billing.
func SubscriptionProviders() []Provider {
	var out []Provider
	for _, p := range Providers {
		if p.Subscriptions {
			out = append(out, p)
		}
	}
	return out
}
