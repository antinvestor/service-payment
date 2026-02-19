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
