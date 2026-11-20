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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameclient "github.com/pitabwire/frame/v3/client"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

const httpClientTimeout = 30 * time.Second

type stripeClient struct {
	mu       sync.RWMutex
	backends *stripe.Backends
	clients  map[string]*stripe.Client // keyed by API key
	metrics  *integrationobs.Metrics
}

// NewClient creates a Stripe client whose backends run on a frame HTTP
// client with a fixed outbound timeout (set once here, never per call).
func NewClient(ctx context.Context) StripeClient {
	return NewClientWithHTTP(frameclient.NewHTTPClient(ctx, frameclient.WithHTTPTimeout(httpClientTimeout)))
}

// NewClientWithHTTP wraps a caller-supplied HTTP client (tests).
func NewClientWithHTTP(httpClient *http.Client) StripeClient {
	backend := func(kind stripe.SupportedBackend) stripe.Backend {
		return stripe.GetBackendWithConfig(kind, &stripe.BackendConfig{HTTPClient: httpClient})
	}
	return &stripeClient{
		backends: &stripe.Backends{
			API:         backend(stripe.APIBackend),
			Connect:     backend(stripe.ConnectBackend),
			Uploads:     backend(stripe.UploadsBackend),
			MeterEvents: backend(stripe.MeterEventsBackend),
		},
		clients: make(map[string]*stripe.Client),
		metrics: integrationobs.NewMetrics("stripe"),
	}
}

func (c *stripeClient) getClient(apiKey string) *stripe.Client {
	c.mu.RLock()
	sc, ok := c.clients[apiKey]
	c.mu.RUnlock()
	if ok {
		return sc
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if sc, ok = c.clients[apiKey]; ok {
		return sc
	}
	sc = stripe.NewClient(apiKey, stripe.WithBackends(c.backends))
	c.clients[apiKey] = sc
	return sc
}

// classify wraps a stripe-go error: transport failures and 5xx/429 answers
// leave the outcome unknown (collection.Uncertain, the worker redelivers
// under the same idempotency key); 4xx answers are rejections.
func classify(op string, err error) error {
	if err == nil {
		return nil
	}
	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		wrapped := fmt.Errorf("%s: %w", op, err)
		if stripeErr.HTTPStatusCode >= http.StatusInternalServerError ||
			stripeErr.HTTPStatusCode == http.StatusTooManyRequests ||
			stripeErr.HTTPStatusCode == 0 {
			return collection.Uncertain(wrapped)
		}
		return wrapped
	}
	// Not a Stripe answer at all: the request may or may not have landed.
	return collection.Uncertain(fmt.Errorf("%s: %w", op, err))
}

// IsNotFound reports whether Stripe answered 404.
func IsNotFound(err error) bool {
	var stripeErr *stripe.Error
	return errors.As(err, &stripeErr) && stripeErr.HTTPStatusCode == http.StatusNotFound
}

// BuildCheckoutSessionParams turns a request into stripe-go params. It is
// exported for tests: the mapping is where the Stripe contract lives.
func BuildCheckoutSessionParams(req *CheckoutSessionRequest) (*stripe.CheckoutSessionCreateParams, error) {
	if req == nil {
		return nil, errors.New("checkout session request is nil")
	}
	if req.SuccessURL == "" {
		return nil, errors.New("checkout session needs a success_url")
	}
	mode := req.Mode
	if mode == "" {
		mode = CheckoutModePayment
	}
	item := &stripe.CheckoutSessionCreateLineItemParams{Quantity: stripe.Int64(1)}
	switch {
	case req.PriceID != "":
		item.Price = stripe.String(req.PriceID)
	case mode == CheckoutModeSubscription:
		return nil, errors.New("subscription checkout needs a Stripe price id (product_id)")
	default:
		if req.Amount <= 0 || req.Currency == "" {
			return nil, errors.New("checkout session needs a price id or a positive amount and currency")
		}
		item.PriceData = &stripe.CheckoutSessionCreateLineItemPriceDataParams{
			Currency:   stripe.String(strings.ToLower(req.Currency)),
			UnitAmount: stripe.Int64(req.Amount),
			ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
				Name: stripe.String(req.ProductName),
			},
		}
	}

	params := &stripe.CheckoutSessionCreateParams{
		Mode:       stripe.String(mode),
		LineItems:  []*stripe.CheckoutSessionCreateLineItemParams{item},
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(firstNonEmpty(req.CancelURL, req.SuccessURL)),
	}
	if req.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}
	if req.ClientReferenceID != "" {
		params.ClientReferenceID = stripe.String(req.ClientReferenceID)
	}
	if len(req.Metadata) > 0 {
		params.Metadata = cloneMetadata(req.Metadata)
		if mode == CheckoutModeSubscription {
			params.SubscriptionData = &stripe.CheckoutSessionCreateSubscriptionDataParams{
				Metadata: cloneMetadata(req.Metadata),
			}
		} else {
			params.PaymentIntentData = &stripe.CheckoutSessionCreatePaymentIntentDataParams{
				Metadata: cloneMetadata(req.Metadata),
			}
		}
	}
	if req.IdempotencyKey != "" {
		params.SetIdempotencyKey(req.IdempotencyKey)
	}
	return params, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *stripeClient) CreateCheckoutSession(
	ctx context.Context,
	creds *StripeCredentials,
	req *CheckoutSessionRequest,
) (_ *CheckoutSessionResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "create_checkout_session")
	defer func() { done(retErr) }()

	params, err := BuildCheckoutSessionParams(req)
	if err != nil {
		return nil, err
	}
	sess, err := c.getClient(creds.APIKey).V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return nil, classify("create checkout session", err)
	}
	return &CheckoutSessionResponse{
		ID:        sess.ID,
		URL:       sess.URL,
		Status:    string(sess.Status),
		Mode:      string(sess.Mode),
		ExpiresAt: sess.ExpiresAt,
	}, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *stripeClient) CreatePaymentIntent(
	ctx context.Context,
	creds *StripeCredentials,
	req *PaymentIntentRequest,
) (_ *PaymentIntentResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "create_payment_intent")
	defer func() { done(retErr) }()

	params := &stripe.PaymentIntentCreateParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(strings.ToLower(req.Currency)),
		AutomaticPaymentMethods: &stripe.PaymentIntentCreateAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}
	if len(req.Metadata) > 0 {
		params.Metadata = cloneMetadata(req.Metadata)
	}
	if req.IdempotencyKey != "" {
		params.SetIdempotencyKey(req.IdempotencyKey)
	}

	pi, err := c.getClient(creds.APIKey).V1PaymentIntents.Create(ctx, params)
	if err != nil {
		return nil, classify("create payment intent", err)
	}
	return &PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
	}, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *stripeClient) CreatePayout(
	ctx context.Context,
	creds *StripeCredentials,
	req *PayoutRequest,
) (_ *PayoutResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "create_payout")
	defer func() { done(retErr) }()

	params := &stripe.PayoutCreateParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(strings.ToLower(req.Currency)),
	}
	if req.Destination != "" {
		params.Destination = stripe.String(req.Destination)
	}
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}
	if len(req.Metadata) > 0 {
		params.Metadata = cloneMetadata(req.Metadata)
	}
	if req.IdempotencyKey != "" {
		params.SetIdempotencyKey(req.IdempotencyKey)
	}

	po, err := c.getClient(creds.APIKey).V1Payouts.Create(ctx, params)
	if err != nil {
		return nil, classify("create payout", err)
	}
	return &PayoutResponse{
		ID:       po.ID,
		Status:   string(po.Status),
		Amount:   po.Amount,
		Currency: string(po.Currency),
	}, nil
}

//nolint:nonamedreturns // named retErr captured by deferred metrics done callback
func (c *stripeClient) CancelSubscription(
	ctx context.Context,
	creds *StripeCredentials,
	subscriptionID string,
	atPeriodEnd bool,
) (_ *SubscriptionResponse, retErr error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "cancel_subscription")
	defer func() { done(retErr) }()

	sc := c.getClient(creds.APIKey)
	var sub *stripe.Subscription
	var err error
	if atPeriodEnd {
		sub, err = sc.V1Subscriptions.Update(ctx, subscriptionID, &stripe.SubscriptionUpdateParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		})
	} else {
		sub, err = sc.V1Subscriptions.Cancel(ctx, subscriptionID, &stripe.SubscriptionCancelParams{})
	}
	if err != nil {
		return nil, classify("cancel subscription", err)
	}
	return &SubscriptionResponse{
		ID:                sub.ID,
		Status:            string(sub.Status),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
	}, nil
}

// VerifyWebhookSignature checks the Stripe-Signature header against the
// endpoint secret. The API version pinned on the webhook endpoint may lag
// stripe-go's, so version mismatches are tolerated; an empty secret is
// refused because it would verify nothing.
func (c *stripeClient) VerifyWebhookSignature(payload []byte, signature, webhookSecret string) (*stripe.Event, error) {
	if strings.TrimSpace(webhookSecret) == "" {
		return nil, errors.New("webhook secret is not configured")
	}
	event, err := webhook.ConstructEventWithOptions(payload, signature, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return nil, fmt.Errorf("verify webhook signature: %w", err)
	}
	return &event, nil
}

func cloneMetadata(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
