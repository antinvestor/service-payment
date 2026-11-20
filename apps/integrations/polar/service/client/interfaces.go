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
	// CreateCheckout creates a Polar checkout session for the products.
	// Subscription vs one-time is decided by the product on Polar's side.
	CreateCheckout(ctx context.Context, creds *PolarCredentials, req *CheckoutRequest) (*CheckoutResponse, error)

	// GetCheckout reads a checkout session (status, metadata).
	GetCheckout(ctx context.Context, creds *PolarCredentials, checkoutID string) (*CheckoutResponse, error)

	// VerifyWebhookSignature verifies a Standard Webhooks signature and
	// decodes the event. headers carry webhook-id, webhook-timestamp and
	// webhook-signature (any case).
	VerifyWebhookSignature(payload []byte, headers map[string]string, webhookSecret string) (*WebhookEvent, error)

	// GetSubscription fetches the current state of a Polar subscription by ID.
	GetSubscription(ctx context.Context, creds *PolarCredentials, subscriptionID string) (*Subscription, error)

	// CancelSubscription cancels a Polar subscription: at the end of the
	// current period (PATCH cancel_at_period_end=true) or immediately
	// (DELETE /revoke).
	CancelSubscription(
		ctx context.Context,
		creds *PolarCredentials,
		subscriptionID string,
		atPeriodEnd bool,
	) (*Subscription, error)
}
