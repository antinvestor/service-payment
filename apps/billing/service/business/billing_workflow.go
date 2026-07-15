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
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/workerpool"
	"github.com/pitabwire/util/decimalx"
)

// BillingWorkflow orchestrates the end-to-end billing pipeline.
// State machine: PENDING → METERING → RATING → DISCOUNTING → INVOICING → CREDITING → POSTING → COMPLETED.
type BillingWorkflow interface {
	RunBilling(ctx context.Context, subscriptionID string, periodStart, periodEnd time.Time) (*models.BillingRun, error)
	GetBillingRun(ctx context.Context, id string) (*models.BillingRun, error)
	// CollectInvoice creates a hosted checkout session for an issued invoice.
	CollectInvoice(ctx context.Context, invoiceID string) (*InvoiceCheckout, error)
	// SettleInvoiceFromCheckout marks an invoice as paid from a completed checkout session.
	SettleInvoiceFromCheckout(ctx context.Context, sessionRef string) (*models.Invoice, error)
}

type billingWorkflow struct {
	workMan         workerpool.Manager
	runRepo         repository.BillingRunRepository
	ratedLineRepo   repository.RatedLineRepository
	subBusiness     SubscriptionBusiness
	catalogBusiness CatalogBusiness
	compRepo        repository.ComponentRepository
	metering        MeteringEngine
	pricing         *PricingEngine
	discountEng     DiscountEngine
	creditEng       CreditEngine
	invoiceEng      InvoiceEngine
	ledgerInteg     LedgerIntegration
	checkout        CheckoutIntegration
	obs             *observability.Metrics
}

func NewBillingWorkflow(
	workMan workerpool.Manager,
	runRepo repository.BillingRunRepository,
	ratedLineRepo repository.RatedLineRepository,
	subBusiness SubscriptionBusiness,
	catalogBusiness CatalogBusiness,
	compRepo repository.ComponentRepository,
	metering MeteringEngine,
	pricing *PricingEngine,
	discountEng DiscountEngine,
	creditEng CreditEngine,
	invoiceEng InvoiceEngine,
	ledgerInteg LedgerIntegration,
	checkoutInteg CheckoutIntegration,
) BillingWorkflow {
	return &billingWorkflow{
		workMan:         workMan,
		runRepo:         runRepo,
		ratedLineRepo:   ratedLineRepo,
		subBusiness:     subBusiness,
		catalogBusiness: catalogBusiness,
		compRepo:        compRepo,
		metering:        metering,
		pricing:         pricing,
		discountEng:     discountEng,
		creditEng:       creditEng,
		invoiceEng:      invoiceEng,
		ledgerInteg:     ledgerInteg,
		checkout:        checkoutInteg,
		obs:             observability.NewMetrics(),
	}
}

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (w *billingWorkflow) RunBilling(
	ctx context.Context,
	subscriptionID string,
	periodStart, periodEnd time.Time,
) (result *models.BillingRun, err error) {
	ctx, span := w.obs.StartSpan(ctx, "RunBilling")
	start := time.Now()
	w.obs.RecordBillingRunStarted(ctx)
	defer func() {
		w.obs.EndSpan(ctx, span, err)
		if err == nil && result != nil && result.State == models.BillingRunStateCompleted {
			w.obs.RecordBillingRunCompleted(ctx, time.Since(start))
		}
	}()

	// Get subscription
	var sub *models.Subscription
	sub, err = w.subBusiness.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	if sub.State != models.SubscriptionStateActive {
		return nil, apperrors.ErrSubscriptionNotActive
	}

	// Create idempotent billing run
	idempotencyKey := fmt.Sprintf("%s:%s:%s",
		subscriptionID,
		periodStart.Format(time.RFC3339),
		periodEnd.Format(time.RFC3339))

	now := time.Now()
	run := &models.BillingRun{
		SubscriptionID:   subscriptionID,
		ProfileID:        sub.ProfileID,
		CatalogVersionID: sub.CatalogVersionID,
		State:            models.BillingRunStatePending,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		StartedAt:        &now,
		Idempotency:      idempotencyKey,
	}
	run.GenID(ctx)

	if createErr := w.runRepo.Create(ctx, run); createErr != nil {
		if data.ErrorIsDuplicateKey(createErr) {
			// Idempotent: return existing run
			existing, getErr := w.runRepo.GetByIdempotency(ctx, idempotencyKey)
			if getErr != nil {
				return nil, getErr
			}
			return existing, nil
		}
		return nil, createErr
	}

	// Execute pipeline steps
	var pipelineErr error

	// Step 1: Metering
	run, pipelineErr = w.stepMetering(ctx, run, sub)
	if pipelineErr != nil {
		return w.failRun(ctx, run, pipelineErr)
	}

	return run, nil
}

func (w *billingWorkflow) stepMetering(
	ctx context.Context,
	run *models.BillingRun,
	sub *models.Subscription,
) (*models.BillingRun, error) {
	run.State = models.BillingRunStateMetering
	if _, err := w.runRepo.Update(ctx, run); err != nil {
		return run, err
	}

	// Get components for the plan
	components, err := w.compRepo.ListByPlanID(ctx, sub.PlanID)
	if err != nil {
		return run, err
	}

	// Meter usage
	metered, err := w.metering.MeterUsage(ctx, sub, components, run)
	if err != nil {
		return run, err
	}

	// Step 2: Rating
	return w.stepRating(ctx, run, sub, metered, components)
}

func (w *billingWorkflow) stepRating(
	ctx context.Context,
	run *models.BillingRun,
	sub *models.Subscription,
	metered []*models.MeteredUsage,
	components []*models.Component,
) (*models.BillingRun, error) {
	run.State = models.BillingRunStateRating
	if _, err := w.runRepo.Update(ctx, run); err != nil {
		return run, err
	}

	// Build component map
	compMap := make(map[string]*models.Component)
	for _, c := range components {
		compMap[c.GetID()] = c
	}

	ratedLines := w.pricing.Rate(metered, compMap, run.GetID(), sub.GetID(), sub.Currency)

	// Persist rated lines for audit trail
	for _, rl := range ratedLines {
		if err := w.ratedLineRepo.Create(ctx, rl); err != nil {
			return run, err
		}
	}

	if len(ratedLines) == 0 {
		// No usage to bill - complete the run
		run.State = models.BillingRunStateCompleted
		completedAt := time.Now()
		run.CompletedAt = &completedAt
		_, err := w.runRepo.Update(ctx, run)
		return run, err
	}

	// Step 3: Discounting
	return w.stepDiscounting(ctx, run, sub, ratedLines)
}

func (w *billingWorkflow) stepDiscounting(
	ctx context.Context,
	run *models.BillingRun,
	sub *models.Subscription,
	ratedLines []*models.RatedLine,
) (*models.BillingRun, error) {
	run.State = models.BillingRunStateDiscounting
	if _, err := w.runRepo.Update(ctx, run); err != nil {
		return run, err
	}

	discountedLines, err := w.discountEng.ApplyDiscounts(ctx, ratedLines, run.GetID(), time.Now())
	if err != nil {
		return run, err
	}

	// Step 4: Invoicing (before crediting so we have an invoice ID for credit entries)
	return w.stepInvoicing(ctx, run, sub, ratedLines, discountedLines)
}

func (w *billingWorkflow) stepInvoicing(
	ctx context.Context,
	run *models.BillingRun,
	sub *models.Subscription,
	ratedLines []*models.RatedLine,
	discountedLines []*models.DiscountedLine,
) (*models.BillingRun, error) {
	run.State = models.BillingRunStateInvoicing
	if _, err := w.runRepo.Update(ctx, run); err != nil {
		return run, err
	}

	// Generate invoice initially with zero credit amount
	invoice, err := w.invoiceEng.GenerateInvoice(ctx, run, ratedLines, discountedLines, decimalx.Zero())
	if err != nil {
		return run, err
	}

	// Persist the invoice ID on the billing run immediately
	run.InvoiceID = invoice.GetID()
	if _, updateErr := w.runRepo.Update(ctx, run); updateErr != nil {
		return run, updateErr
	}

	// Step 5: Crediting (now we have the invoice ID)
	return w.stepCrediting(ctx, run, sub, invoice)
}

func (w *billingWorkflow) stepCrediting(
	ctx context.Context,
	run *models.BillingRun,
	sub *models.Subscription,
	invoice *models.Invoice,
) (*models.BillingRun, error) {
	run.State = models.BillingRunStateCrediting
	if _, err := w.runRepo.Update(ctx, run); err != nil {
		return run, err
	}

	// Calculate amount eligible for credits (subtotal - discounts)
	subtotal := decimalx.DerefOr(invoice.SubtotalAmount, decimalx.Zero())
	discount := decimalx.DerefOr(invoice.DiscountAmount, decimalx.Zero())
	amountAfterDiscount := subtotal.Sub(discount)
	if amountAfterDiscount.IsNegative() {
		amountAfterDiscount = decimalx.Zero()
	}

	if amountAfterDiscount.IsPositive() {
		remainingAmount, _, err := w.creditEng.ApplyCredits(
			ctx, sub.ProfileID, sub.Currency, amountAfterDiscount, run.GetID(), invoice.GetID())
		if err != nil {
			return run, err
		}

		creditAmount := amountAfterDiscount.Sub(remainingAmount)
		if creditAmount.IsPositive() {
			if updateErr := w.invoiceEng.UpdateInvoiceTotals(ctx, invoice, creditAmount); updateErr != nil {
				return run, updateErr
			}
		}
	}

	// Step 6: Posting to ledger
	return w.stepPosting(ctx, run)
}

func (w *billingWorkflow) stepPosting(
	ctx context.Context,
	run *models.BillingRun,
) (*models.BillingRun, error) {
	run.State = models.BillingRunStatePosting
	if _, err := w.runRepo.Update(ctx, run); err != nil {
		return run, err
	}

	// NOTE: Ledger posting will be implemented when AR/Revenue account IDs
	// are available from configuration. The invoice is generated and can be
	// posted retroactively.

	// Step 7: Complete
	run.State = models.BillingRunStateCompleted
	completedAt := time.Now()
	run.CompletedAt = &completedAt

	_, err := w.runRepo.Update(ctx, run)
	return run, err
}

func (w *billingWorkflow) failRun(
	ctx context.Context,
	run *models.BillingRun,
	pipelineErr error,
) (*models.BillingRun, error) {
	now := time.Now()
	run.State = models.BillingRunStateFailed
	run.FailedAt = &now
	run.ErrorMessage = pipelineErr.Error()

	_, updateErr := w.runRepo.Update(ctx, run)
	if updateErr != nil {
		return run, fmt.Errorf("pipeline error: %w; also failed to update run: %w", pipelineErr, updateErr)
	}

	return run, pipelineErr
}

func (w *billingWorkflow) GetBillingRun(ctx context.Context, id string) (*models.BillingRun, error) {
	if id == "" {
		return nil, ErrBillingRunIDRequired
	}

	return w.runRepo.GetByID(ctx, id)
}

// CollectInvoice creates a hosted checkout session for an issued invoice.
// Returns an error if the checkout integration is not configured.
func (w *billingWorkflow) CollectInvoice(
	ctx context.Context,
	invoiceID string,
) (*InvoiceCheckout, error) {
	if w.checkout == nil {
		return nil, errors.New("checkout integration not configured")
	}
	return w.checkout.CreateInvoiceCheckout(ctx, invoiceID, CheckoutOptions{Source: CollectionSourceInvoice})
}

// SettleInvoiceFromCheckout marks an invoice as paid from a completed checkout session.
// Returns an error if the checkout integration is not configured.
func (w *billingWorkflow) SettleInvoiceFromCheckout(
	ctx context.Context,
	sessionRef string,
) (*models.Invoice, error) {
	if w.checkout == nil {
		return nil, errors.New("checkout integration not configured")
	}
	return w.checkout.SettleFromCheckout(ctx, sessionRef)
}
