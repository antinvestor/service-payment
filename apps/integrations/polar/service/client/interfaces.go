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

import "context"

// PolarClient defines the interface for Polar.sh API operations.
type PolarClient interface {
	// CreateCheckout creates a Polar checkout session for requesting money.
	// For subscription products the product id determines recurrence — Polar
	// decides subscription vs one-time by the product; no extra checkout field
	// is required from our side.
	CreateCheckout(ctx context.Context, creds *PolarCredentials, req *CheckoutRequest) (*CheckoutResponse, error)

	// VerifyWebhookSignature verifies a Polar webhook signature
	VerifyWebhookSignature(payload []byte, headers map[string]string, webhookSecret string) (*WebhookEvent, error)

	// GetSubscription fetches the current state of a Polar subscription by ID.
	// Endpoint: GET /v1/subscriptions/{id}
	GetSubscription(ctx context.Context, creds *PolarCredentials, subscriptionID string) (*Subscription, error)

	// CancelSubscription cancels a Polar subscription.
	// When atPeriodEnd is true the subscription is cancelled at the end of the
	// current billing period (PATCH /v1/subscriptions/{id} with
	// cancel_at_period_end=true); when false the subscription is revoked
	// immediately (DELETE /v1/subscriptions/{id}/revoke).
	CancelSubscription(
		ctx context.Context,
		creds *PolarCredentials,
		subscriptionID string,
		atPeriodEnd bool,
	) (*Subscription, error)
}
