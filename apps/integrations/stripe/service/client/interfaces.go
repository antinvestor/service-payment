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

import (
	"context"

	"github.com/stripe/stripe-go/v82"
)

// StripeClient defines the interface for Stripe API operations.
//
// Transport failures and 5xx answers are returned wrapped in
// collection.Uncertain so workers redeliver; 4xx rejections are plain errors.
type StripeClient interface {
	// CreateCheckoutSession opens a hosted Checkout Session (payment or
	// subscription mode).
	CreateCheckoutSession(
		ctx context.Context,
		creds *StripeCredentials,
		req *CheckoutSessionRequest,
	) (*CheckoutSessionResponse, error)

	// CreatePaymentIntent creates a PaymentIntent whose client_secret is
	// confirmed by Stripe.js on our own page.
	CreatePaymentIntent(
		ctx context.Context,
		creds *StripeCredentials,
		req *PaymentIntentRequest,
	) (*PaymentIntentResponse, error)

	// CreatePayout creates a Payout to an external account of the platform
	// (or connected) Stripe account.
	CreatePayout(ctx context.Context, creds *StripeCredentials, req *PayoutRequest) (*PayoutResponse, error)

	// CancelSubscription cancels a subscription at the end of the current
	// period (update cancel_at_period_end=true) or immediately (DELETE).
	CancelSubscription(
		ctx context.Context,
		creds *StripeCredentials,
		subscriptionID string,
		atPeriodEnd bool,
	) (*SubscriptionResponse, error)

	// VerifyWebhookSignature verifies the Stripe-Signature header and
	// decodes the event.
	VerifyWebhookSignature(payload []byte, signature, webhookSecret string) (*stripe.Event, error)
}
