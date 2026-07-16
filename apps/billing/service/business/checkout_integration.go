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

	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	checkoutv1 "github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1"
	"github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1/checkoutv1connect"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/util"
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
type CheckoutIntegration interface {
	CreateInvoiceCheckout(ctx context.Context, invoiceID string, opts CheckoutOptions) (*InvoiceCheckout, error)
	SettleFromCheckout(ctx context.Context, sessionRef string) (*models.Invoice, error)
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

	resp, err := c.checkoutCli.CreateCheckoutSession(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}

	data := resp.Msg.GetData()

	util.Log(ctx).
		WithField("invoice_id", invoiceID).
		WithField("session_ref", data.GetRef()).
		Info("checkout session created for invoice")

	return &InvoiceCheckout{
		SessionRef: data.GetRef(),
		PageURL:    data.GetPageUrl(),
	}, nil
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
		connect.NewRequest(&checkoutv1.GetCheckoutSessionRequest{Ref: sessionRef}),
	)
	if err != nil {
		return nil, fmt.Errorf("get checkout session: %w", err)
	}

	session := resp.Msg.GetData()
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
