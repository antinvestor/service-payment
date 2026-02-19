package client

import "context"

// PolarClient defines the interface for Polar.sh API operations.
type PolarClient interface {
	// CreateCheckout creates a Polar checkout session for requesting money
	CreateCheckout(ctx context.Context, creds *PolarCredentials, req *CheckoutRequest) (*CheckoutResponse, error)

	// VerifyWebhookSignature verifies a Polar webhook signature
	VerifyWebhookSignature(payload []byte, headers map[string]string, webhookSecret string) (*WebhookEvent, error)
}
