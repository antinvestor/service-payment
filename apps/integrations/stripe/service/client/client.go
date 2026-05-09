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
	"fmt"
	"sync"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type stripeClient struct {
	mu      sync.RWMutex
	clients map[string]*stripe.Client // keyed by API key
}

// NewClient creates a new Stripe API client.
func NewClient() StripeClient {
	return &stripeClient{
		clients: make(map[string]*stripe.Client),
	}
}

func (c *stripeClient) getClient(apiKey string) *stripe.Client {
	c.mu.RLock()
	if sc, ok := c.clients[apiKey]; ok {
		c.mu.RUnlock()
		return sc
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if sc, ok := c.clients[apiKey]; ok {
		return sc
	}

	sc := stripe.NewClient(apiKey)
	c.clients[apiKey] = sc
	return sc
}

func (c *stripeClient) CreatePaymentIntent(
	ctx context.Context,
	creds *StripeCredentials,
	req *PaymentIntentRequest,
) (*PaymentIntentResponse, error) {
	sc := c.getClient(creds.APIKey)

	params := &stripe.PaymentIntentCreateParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentCreateAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	if len(req.Metadata) > 0 {
		params.Metadata = req.Metadata
	}

	pi, err := sc.V1PaymentIntents.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create payment intent: %w", err)
	}

	return &PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
	}, nil
}

func (c *stripeClient) CreatePayout(
	ctx context.Context,
	creds *StripeCredentials,
	req *PayoutRequest,
) (*PayoutResponse, error) {
	sc := c.getClient(creds.APIKey)

	params := &stripe.PayoutCreateParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
	}

	if req.Destination != "" {
		params.Destination = stripe.String(req.Destination)
	}

	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	if len(req.Metadata) > 0 {
		params.Metadata = req.Metadata
	}

	po, err := sc.V1Payouts.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create payout: %w", err)
	}

	return &PayoutResponse{
		ID:       po.ID,
		Status:   string(po.Status),
		Amount:   po.Amount,
		Currency: string(po.Currency),
	}, nil
}

func (c *stripeClient) VerifyWebhookSignature(payload []byte, signature, webhookSecret string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("verify webhook signature: %w", err)
	}

	data := make(map[string]any)
	if event.Data != nil && event.Data.Raw != nil {
		data["raw"] = string(event.Data.Raw)
	}

	return &WebhookEvent{
		ID:   event.ID,
		Type: string(event.Type),
		Data: data,
	}, nil
}
