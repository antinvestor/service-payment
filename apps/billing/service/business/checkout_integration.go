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

package business

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	checkoutv1 "github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1"
	"github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1/checkoutv1connect"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/pitabwire/util"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
)

// ErrCheckoutNotCompleted is returned when a checkout session has not yet
// reached the SESSION_STATUS_COMPLETED state.
var ErrCheckoutNotCompleted = errors.New("checkout session is not completed")

// InvoiceCheckout carries the result of a successfully created checkout
// session for an invoice.
type InvoiceCheckout struct {
	SessionRef string
	PageURL    string
}

// CheckoutOptions customises CreateInvoiceCheckout.
type CheckoutOptions struct {
	// ReturnURL overrides the integration-level default return URL when non-empty.
	ReturnURL string
	// Source is stored in session metadata (e.g. invoice | subscription).
	Source string
	// PayerDisplayName is used for checkout prefill when the invoice has a profile.
	PayerDisplayName string
	// Methods restricts payment method keys shown on the hosted page (empty = all).
	Methods []string
}

// CheckoutIntegration creates hosted checkout sessions for issued invoices
// and settles invoices from verified completed sessions.
// SubscriptionCheckoutInput opens hosted checkout for a provider-managed
// subscription: the provider charges its product, billing only needs the
// session to learn that the payer finished.
type SubscriptionCheckoutInput struct {
	Subscription *models.Subscription
	Plan         *models.Plan
	Provider     string
	ProductID    string
	Amount       *commonv1.Money
	Options      CheckoutOptions
}

// CheckoutSessionState is what billing reads back from a checkout session.
type CheckoutSessionState struct {
	Ref       string
	Completed bool
	// Status is the checkout session status name (pending, processing,
	// completed, failed, expired).
	Status   string
	PageURL  string
	OrderRef string
	Metadata map[string]string
	// PaidMethod / PaidRoute are the method key and payment route the
	// completed payment went through (empty until a payment is made).
	PaidMethod string
	PaidRoute  string
}

type CheckoutIntegration interface {
	CreateInvoiceCheckout(ctx context.Context, invoiceID string, opts CheckoutOptions) (*InvoiceCheckout, error)
	CreateSubscriptionCheckout(ctx context.Context, in SubscriptionCheckoutInput) (*InvoiceCheckout, error)
	SettleFromCheckout(ctx context.Context, sessionRef string) (*models.Invoice, error)
	// SessionState reads a checkout session without settling anything.
	SessionState(ctx context.Context, sessionRef string) (*CheckoutSessionState, error)
}

type checkoutIntegration struct {
	checkoutCli checkoutv1connect.CheckoutServiceClient
	invoiceRepo repository.InvoiceRepository
	invoiceEng  InvoiceEngine
	returnURL   string
	obs         *observability.Metrics
}

// NewCheckoutIntegration returns a CheckoutIntegration that calls the checkout
// service and settles invoices via the invoice engine.
func NewCheckoutIntegration(
	checkoutCli checkoutv1connect.CheckoutServiceClient,
	invoiceRepo repository.InvoiceRepository,
	invoiceEng InvoiceEngine,
	returnURL string,
) CheckoutIntegration {
	return &checkoutIntegration{
		checkoutCli: checkoutCli,
		invoiceRepo: invoiceRepo,
		invoiceEng:  invoiceEng,
		returnURL:   returnURL,
		obs:         observability.NewMetrics(),
	}
}

// CreateInvoiceCheckout creates a hosted checkout session for an issued
// invoice and returns the session reference and page URL.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (c *checkoutIntegration) CreateInvoiceCheckout(
	ctx context.Context,
	invoiceID string,
	opts CheckoutOptions,
) (result *InvoiceCheckout, err error) {
	ctx, span := c.obs.StartSpan(ctx, "CreateInvoiceCheckout")
	defer func() {
		c.obs.EndSpan(ctx, span, err)
	}()

	if invoiceID == "" {
		return nil, ErrInvoiceIDRequired
	}

	invoice, err := c.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	if invoice.State != models.InvoiceStateIssued {
		return nil, apperrors.ErrInvoiceNotPayable
	}

	money, err := moneyFromDecimal(invoice.TotalAmount, invoice.Currency)
	if err != nil {
		return nil, fmt.Errorf("cannot build checkout amount: %w", err)
	}

	metadata := map[string]string{
		"invoiceId":     invoice.GetID(),
		"invoiceNumber": invoice.InvoiceNumber,
	}
	if invoice.SubscriptionID != "" {
		metadata["subscriptionId"] = invoice.SubscriptionID
	}
	if opts.Source != "" {
		metadata["source"] = opts.Source
	}

	req := &checkoutv1.CreateCheckoutSessionRequest{
		Name:     "Invoice " + invoice.InvoiceNumber,
		OrderRef: invoice.GetID(),
		Amount:   money,
		Metadata: metadata,
		Methods:  opts.Methods,
	}

	if invoice.ProfileID != "" {
		req.Payer = &checkoutv1.PayerPrefill{
			ProfileId:   invoice.ProfileID,
			DisplayName: opts.PayerDisplayName,
		}
	}

	returnURL := c.returnURL
	if opts.ReturnURL != "" {
		returnURL = opts.ReturnURL
	}
	if returnURL != "" {
		req.ReturnUrl = returnURL
	}

	resp, err := c.checkoutCli.CreateCheckoutSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}

	data := resp.GetData()

	util.Log(ctx).
		WithField("invoice_id", invoiceID).
		WithField("session_ref", data.GetRef()).
		Info("checkout session created for invoice")

	return &InvoiceCheckout{
		SessionRef: data.GetRef(),
		PageURL:    data.GetPageUrl(),
	}, nil
}

// CreateSubscriptionCheckout opens hosted checkout restricted to the
// provider's rail. The session carries product_id so the provider adapter
// charges the catalog product, and subscriptionId so provider webhooks map
// back to this subscription.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (c *checkoutIntegration) CreateSubscriptionCheckout(
	ctx context.Context,
	in SubscriptionCheckoutInput,
) (result *InvoiceCheckout, err error) {
	ctx, span := c.obs.StartSpan(ctx, "CreateSubscriptionCheckout")
	defer func() { c.obs.EndSpan(ctx, span, err) }()

	if in.Subscription == nil {
		return nil, ErrSubscriptionIDRequired
	}
	if in.Provider == "" || in.ProductID == "" {
		return nil, ErrPlanNotPolarCollected
	}
	// Only the provider's own rail can open a provider-managed subscription:
	// a caller may narrow the list, never widen it onto another rail.
	providerKey := in.Provider
	if p, ok := collection.ProviderByRoute(in.Provider); ok {
		providerKey = p.MethodKey
	}
	methods := []string{providerKey}
	if len(in.Options.Methods) > 0 {
		allowed := false
		for _, m := range in.Options.Methods {
			if strings.EqualFold(strings.TrimSpace(m), providerKey) {
				allowed = true
			}
		}
		if !allowed {
			return nil, fmt.Errorf("%w: plan is charged by %s, methods must include %q",
				ErrPlanNotPolarCollected, in.Provider, providerKey)
		}
	}
	name := "Subscription"
	if in.Plan != nil && in.Plan.Name != "" {
		name = in.Plan.Name
	}
	metadata := map[string]string{
		"subscriptionId":                 in.Subscription.GetID(),
		"source":                         CollectionSourceSubscription,
		models.SubDataCollectionMode:     models.CollectionModeProvider,
		models.SubDataPaymentProvider:    in.Provider,
		collection.ExtraProductID:        in.ProductID,
		models.PlanDataProviderProductID: in.ProductID,
	}
	if in.Plan != nil {
		metadata["planId"] = in.Plan.GetID()
	}
	req := &checkoutv1.CreateCheckoutSessionRequest{
		Name:     name,
		OrderRef: in.Subscription.GetID(),
		Amount:   in.Amount,
		Metadata: metadata,
		Methods:  methods,
	}
	if in.Subscription.ProfileID != "" {
		req.Payer = &checkoutv1.PayerPrefill{
			ProfileId:   in.Subscription.ProfileID,
			DisplayName: in.Options.PayerDisplayName,
		}
	}
	returnURL := c.returnURL
	if in.Options.ReturnURL != "" {
		returnURL = in.Options.ReturnURL
	}
	if returnURL != "" {
		req.ReturnUrl = returnURL
	}

	resp, err := c.checkoutCli.CreateCheckoutSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}
	data := resp.GetData()
	util.Log(ctx).
		WithField("subscription_id", in.Subscription.GetID()).
		WithField("provider", in.Provider).
		WithField("session_ref", data.GetRef()).
		Info("checkout session created for provider subscription")
	return &InvoiceCheckout{SessionRef: data.GetRef(), PageURL: data.GetPageUrl()}, nil
}

// SessionState reads a checkout session's status and metadata.
func (c *checkoutIntegration) SessionState(ctx context.Context, sessionRef string) (*CheckoutSessionState, error) {
	if sessionRef == "" {
		return nil, apperrors.ErrUnspecifiedReference
	}
	resp, err := c.checkoutCli.GetCheckoutSession(ctx, &checkoutv1.GetCheckoutSessionRequest{Ref: sessionRef})
	if err != nil {
		return nil, fmt.Errorf("get checkout session: %w", err)
	}
	session := resp.GetData()
	return &CheckoutSessionState{
		Ref:        session.GetRef(),
		Completed:  session.GetStatus() == checkoutv1.SessionStatus_SESSION_STATUS_COMPLETED,
		Status:     sessionStatusName(session.GetStatus()),
		PageURL:    session.GetPageUrl(),
		OrderRef:   session.GetOrderRef(),
		Metadata:   session.GetMetadata(),
		PaidMethod: session.GetPaidMethod(),
		PaidRoute:  session.GetPaidRoute(),
	}, nil
}

// sessionStatusName maps the proto enum to the checkout status names.
func sessionStatusName(s checkoutv1.SessionStatus) string {
	switch s {
	case checkoutv1.SessionStatus_SESSION_STATUS_PROCESSING:
		return "processing"
	case checkoutv1.SessionStatus_SESSION_STATUS_COMPLETED:
		return "completed"
	case checkoutv1.SessionStatus_SESSION_STATUS_FAILED:
		return "failed"
	case checkoutv1.SessionStatus_SESSION_STATUS_EXPIRED:
		return "expired"
	default:
		return "pending"
	}
}

// SettleFromCheckout verifies that a checkout session is completed, matches
// the invoice amount, and then calls RecordPayment to mark the invoice paid.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (c *checkoutIntegration) SettleFromCheckout(
	ctx context.Context,
	sessionRef string,
) (result *models.Invoice, err error) {
	ctx, span := c.obs.StartSpan(ctx, "SettleFromCheckout")
	defer func() {
		c.obs.EndSpan(ctx, span, err)
	}()

	if sessionRef == "" {
		return nil, apperrors.ErrUnspecifiedReference
	}

	resp, err := c.checkoutCli.GetCheckoutSession(
		ctx,
		&checkoutv1.GetCheckoutSessionRequest{Ref: sessionRef},
	)
	if err != nil {
		return nil, fmt.Errorf("get checkout session: %w", err)
	}

	session := resp.GetData()
	if session.GetStatus() != checkoutv1.SessionStatus_SESSION_STATUS_COMPLETED {
		return nil, ErrCheckoutNotCompleted
	}

	invoiceID := session.GetOrderRef()
	if invoiceID == "" {
		return nil, fmt.Errorf("checkout session %q has no order_ref", sessionRef)
	}

	invoice, err := c.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	// Idempotent: already paid → return without error.
	if invoice.State == models.InvoiceStatePaid {
		return invoice, nil
	}

	// Verify session amount matches invoice total.
	if !moneyMatchesDecimal(session.GetAmount(), invoice.TotalAmount, invoice.Currency) {
		return nil, fmt.Errorf(
			"amount mismatch: checkout session amount does not match invoice total for invoice %s",
			invoiceID,
		)
	}

	invoice, err = c.invoiceEng.RecordPayment(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	util.Log(ctx).
		WithField("invoice_id", invoiceID).
		WithField("session_ref", sessionRef).
		Info("invoice settled from checkout session")

	return invoice, nil
}
