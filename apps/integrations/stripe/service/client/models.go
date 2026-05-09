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

// PaymentIntentRequest represents a request to create a PaymentIntent.
type PaymentIntentRequest struct {
	Amount      int64
	Currency    string
	Description string
	Metadata    map[string]string
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
	Amount      int64
	Currency    string
	Destination string
	Description string
	Metadata    map[string]string
}

// PayoutResponse represents the response from creating a Payout.
type PayoutResponse struct {
	ID       string
	Status   string
	Amount   int64
	Currency string
}

// WebhookEvent represents a parsed Stripe webhook event.
type WebhookEvent struct {
	ID   string
	Type string
	Data map[string]any
}
