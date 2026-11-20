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

// PolarCredentials holds per-request credentials for Polar API.
type PolarCredentials struct {
	APIKey         string
	WebhookSecret  string
	OrganizationID string
	Environment    string
	// BaseURLOverride overrides the default base URL derived from Environment.
	// Intended for testing only; leave empty in production.
	BaseURLOverride string
}

// BaseURL returns the Polar API base URL based on environment.
// If BaseURLOverride is set it takes precedence (useful in tests).
func (c *PolarCredentials) BaseURL() string {
	if c.BaseURLOverride != "" {
		return c.BaseURLOverride
	}
	if c.Environment == "production" {
		return "https://api.polar.sh"
	}
	return "https://sandbox-api.polar.sh"
}

// CheckoutRequest creates a Polar checkout session for one or more products.
// Polar prices the products itself: for subscription products the checkout
// opens a subscription, for one-time products a single order.
type CheckoutRequest struct {
	ProductIDs    []string
	CustomerEmail string
	CustomerName  string
	// ExternalCustomerID is our stable payer id (profile id) so Polar
	// customers map back to profiles and renewals carry it.
	ExternalCustomerID string
	SuccessURL         string
	// Amount in minor units for pay-what-you-want products only.
	Amount   int64
	Metadata map[string]string
}

// checkoutPayload is the JSON payload sent to POST /v1/checkouts/.
type checkoutPayload struct {
	Products           []string          `json:"products"`
	CustomerEmail      string            `json:"customer_email,omitempty"`
	CustomerName       string            `json:"customer_name,omitempty"`
	ExternalCustomerID string            `json:"external_customer_id,omitempty"`
	SuccessURL         string            `json:"success_url,omitempty"`
	Amount             int64             `json:"amount,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// CheckoutResponse represents the response from creating a checkout session.
type CheckoutResponse struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	ClientSecret string `json:"client_secret"`
	Status       string `json:"status"`
	ExpiresAt    string `json:"expires_at"`
}

// WebhookEvent represents a parsed Polar webhook event.
type WebhookEvent struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
	// ID and Timestamp come from the Standard Webhooks headers.
	ID        string `json:"-"`
	Timestamp string `json:"-"`
}

// Subscription represents a Polar subscription object returned by the API.
// See https://docs.polar.sh/api-reference/subscriptions/get
type Subscription struct {
	ID                 string         `json:"id"`
	Status             string         `json:"status"`
	ProductID          string         `json:"product_id"`
	CustomerID         string         `json:"customer_id"`
	CurrentPeriodStart string         `json:"current_period_start"`
	CurrentPeriodEnd   string         `json:"current_period_end"`
	CancelAtPeriodEnd  bool           `json:"cancel_at_period_end"`
	EndsAt             string         `json:"ends_at"`
	Metadata           map[string]any `json:"metadata"`
}
