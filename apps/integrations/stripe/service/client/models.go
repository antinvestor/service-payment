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

// StripeCredentials holds per-request credentials for Stripe API.
type StripeCredentials struct {
	APIKey        string
	WebhookSecret string
}

// Checkout Session modes (https://docs.stripe.com/api/checkout/sessions/create#create_checkout_session-mode).
const (
	CheckoutModePayment      = "payment"
	CheckoutModeSubscription = "subscription"
)

// CheckoutSessionRequest describes a hosted Checkout Session.
//
// Exactly one of PriceID (a catalog price, "price_…") or Amount+Currency
// (ad-hoc price_data) is used for the single line item.
type CheckoutSessionRequest struct {
	// IdempotencyKey is sent as Idempotency-Key so a redelivered prompt
	// returns the session Stripe already created.
	IdempotencyKey string
	Mode           string
	PriceID        string
	// Amount is in the currency's minor units; ignored when PriceID is set.
	Amount   int64
	Currency string
	// ProductName is the customer-facing line item name for ad-hoc prices.
	ProductName       string
	SuccessURL        string
	CancelURL         string
	CustomerEmail     string
	ClientReferenceID string
	// Metadata is set on the session and copied to the subscription
	// (subscription mode) or payment intent (payment mode) so their
	// webhooks carry the same correlation keys.
	Metadata map[string]string
}

// CheckoutSessionResponse is the subset of a Checkout Session the workers use.
type CheckoutSessionResponse struct {
	ID        string
	URL       string
	Status    string
	Mode      string
	ExpiresAt int64
}

// PaymentIntentRequest represents a request to create a PaymentIntent.
type PaymentIntentRequest struct {
	IdempotencyKey string
	Amount         int64
	Currency       string
	Description    string
	Metadata       map[string]string
}

// PaymentIntentResponse represents the response from creating a PaymentIntent.
type PaymentIntentResponse struct {
	ID           string
	ClientSecret string
	Status       string
	Amount       int64
	Currency     string
}

// PayoutRequest represents a request to create a Payout.
type PayoutRequest struct {
	IdempotencyKey string
	Amount         int64
	Currency       string
	Destination    string
	Description    string
	Metadata       map[string]string
}

// PayoutResponse represents the response from creating a Payout.
type PayoutResponse struct {
	ID       string
	Status   string
	Amount   int64
	Currency string
}

// SubscriptionResponse is the subset of a Stripe subscription the command
// worker reports.
type SubscriptionResponse struct {
	ID                string
	Status            string
	CancelAtPeriodEnd bool
}
