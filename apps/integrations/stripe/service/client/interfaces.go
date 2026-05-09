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

// StripeClient defines the interface for Stripe API operations.
type StripeClient interface {
	// CreatePaymentIntent creates a Stripe PaymentIntent for requesting money
	CreatePaymentIntent(
		ctx context.Context,
		creds *StripeCredentials,
		req *PaymentIntentRequest,
	) (*PaymentIntentResponse, error)

	// CreatePayout creates a Stripe Payout for sending money
	CreatePayout(ctx context.Context, creds *StripeCredentials, req *PayoutRequest) (*PayoutResponse, error)

	// VerifyWebhookSignature verifies and parses a Stripe webhook event
	VerifyWebhookSignature(payload []byte, signature, webhookSecret string) (*WebhookEvent, error)
}
