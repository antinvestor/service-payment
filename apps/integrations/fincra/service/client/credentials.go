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

package client

import "strings"

const (
	// SandboxBaseURL is the Fincra test environment API host.
	SandboxBaseURL = "https://sandboxapi.fincra.com"
	// ProductionBaseURL is the Fincra live environment API host.
	ProductionBaseURL = "https://api.fincra.com"

	// SandboxCheckoutBaseURL hosts sandbox checkout pages (/pay/{payCode}).
	SandboxCheckoutBaseURL = "https://sandbox-checkout.fincra.com"
	// ProductionCheckoutBaseURL hosts live checkout pages (/pay/{payCode}).
	ProductionCheckoutBaseURL = "https://checkout.fincra.com"

	// EnvironmentProduction selects the live API.
	EnvironmentProduction = "production"

	// CollectionModeCheckout collects through Fincra's hosted checkout page.
	CollectionModeCheckout = "checkout"
	// CollectionModeDirect collects through the direct charge API.
	CollectionModeDirect = "direct"

	// FeeBearerBusiness makes the merchant pay Fincra's fees.
	FeeBearerBusiness = "business"
	// FeeBearerCustomer adds Fincra's fees to what the payer pays.
	FeeBearerCustomer = "customer"
)

// Credentials are the per-connection settings used to call Fincra.
type Credentials struct {
	// SecretKey authenticates API requests (api-key header).
	SecretKey string
	// PublicKey identifies the account (x-pub-key header).
	PublicKey string
	// BusinessID is the parent business id (x-business-id header, payout body).
	BusinessID string
	// WebhookSecret is the key Fincra signs webhooks with.
	WebhookSecret string
	// Environment is sandbox or production.
	Environment string
	// BaseURL overrides the environment host.
	BaseURL string
	// CollectionMode is checkout or direct.
	CollectionMode string
	// FeeBearer is business or customer.
	FeeBearer string
	// Country is the default payout country.
	Country string
	// Connection is the settings connection the credentials came from, or
	// "" for the service's own configured account.
	Connection string
	// TenantID and PartitionID are the tenancy a settings connection is
	// bound to (set only for connections).
	TenantID    string
	PartitionID string
}

// IsProduction reports whether the credentials target the live API.
func (c *Credentials) IsProduction() bool {
	switch strings.ToLower(strings.TrimSpace(c.Environment)) {
	case EnvironmentProduction, "live":
		return true
	}
	return false
}

// ResolveBaseURL returns the API host for the credentials' environment.
func (c *Credentials) ResolveBaseURL() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	if c.IsProduction() {
		return ProductionBaseURL
	}
	return SandboxBaseURL
}

// CheckoutLink is the hosted payment page for a pay code.
func (c *Credentials) CheckoutLink(payCode string) string {
	if payCode == "" {
		return ""
	}
	base := SandboxCheckoutBaseURL
	if c.IsProduction() {
		base = ProductionCheckoutBaseURL
	}
	return base + "/pay/" + payCode
}

// ResolveCollectionMode returns checkout unless direct charge is configured.
func (c *Credentials) ResolveCollectionMode() string {
	if strings.EqualFold(strings.TrimSpace(c.CollectionMode), CollectionModeDirect) {
		return CollectionModeDirect
	}
	return CollectionModeCheckout
}

// ResolveFeeBearer returns business unless customer is configured.
func (c *Credentials) ResolveFeeBearer() string {
	if strings.EqualFold(strings.TrimSpace(c.FeeBearer), FeeBearerCustomer) {
		return FeeBearerCustomer
	}
	return FeeBearerBusiness
}
