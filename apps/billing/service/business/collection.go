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
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
)

// Invoice / subscription Data map keys used by simplified collection.
const (
	InvoiceDataCheckoutSessionRef = "checkoutSessionRef"
	InvoiceDataCollectionSource   = "collectionSource"
	CollectionSourceInvoice       = "invoice"
	CollectionSourceSubscription  = "subscription"
)

// ErrPlanHasNoUpfrontCharge is returned when StartSubscription finds no flat fee
// and the caller expected a checkout URL. Callers should treat already_complete=true
// as success instead.
var ErrPlanHasNoUpfrontCharge = errors.New("plan has no billable upfront fee")

// CollectPaymentInput starts checkout for an issued invoice.
type CollectPaymentInput struct {
	InvoiceID string
	ReturnURL string   // optional per-call override
	Methods   []string // optional payment method restriction
}

// StartSubscriptionInput creates a subscription and optional first-charge checkout.
type StartSubscriptionInput struct {
	ProfileID        string
	PlanID           string
	CatalogVersionID string
	Currency         string
	ReturnURL        string
	PayerDisplayName string
	Methods          []string // optional payment method restriction

	// External-entity integration (product systems, entitlements).
	// Stored on Subscription.Data and included in lifecycle queue events.
	ExternalEntityID   string
	ExternalEntityType string
	// IntegrationRouteID optionally pins lifecycle delivery to one route row.
	IntegrationRouteID string
	// Metadata is merged into Subscription.Data (product-defined keys).
	Metadata map[string]string
}

// CollectionResult is the merchant-facing result of CollectPayment / StartSubscription.
type CollectionResult struct {
	PageURL         string
	SessionRef      string
	InvoiceID       string
	SubscriptionID  string
	AlreadyComplete bool
}

// ConfirmPaymentResult is the result of ConfirmPayment.
type ConfirmPaymentResult struct {
	InvoiceID         string
	InvoiceState      string
	SubscriptionID    string
	SubscriptionState string
	Paid              bool
}

// CancelSubscriptionResult is the result of CancelSubscription.
type CancelSubscriptionResult struct {
	SubscriptionID    string
	SubscriptionState string
	VoidedInvoiceID   string
}

// CollectionBusiness is the simplified merchant API for payment collection.
//
// Happy paths:
//  1. CollectPayment(invoice) → redirect → ConfirmPayment(session)
//  2. StartSubscription(plan) → redirect (if fee) → ConfirmPayment(session)
//  3. CancelSubscription(id) for active or unpaid-pending subscriptions
type CollectionBusiness interface {
	CollectPayment(ctx context.Context, in CollectPaymentInput) (*CollectionResult, error)
	StartSubscription(ctx context.Context, in StartSubscriptionInput) (*CollectionResult, error)
	ConfirmPayment(ctx context.Context, sessionRef string) (*ConfirmPaymentResult, error)
	CancelSubscription(ctx context.Context, subscriptionID string) (*CancelSubscriptionResult, error)
}

// CollectionLedgerAccounts holds optional ledger account IDs for cash capture.
type CollectionLedgerAccounts struct {
	CashAccountID string
	ARAccountID   string
}

type collectionBusiness struct {
	checkout       CheckoutIntegration
	invoiceEng     InvoiceEngine
	invoiceRepo    repository.InvoiceRepository
	subBiz         SubscriptionBusiness
	planRepo       repository.PlanRepository
	compRepo       repository.ComponentRepository
	runRepo        repository.BillingRunRepository
	pricing        *PricingEngine
	ledger         LedgerIntegration
	ledgerAccounts CollectionLedgerAccounts
	instruments    InstrumentSource    // optional — pins COF on first settle
	scheduler      RenewalScheduler    // optional — per-sub Trustage renew
	settleSched    SettlementScheduler // optional — per-invoice Trustage settle
	obs            *observability.Metrics
}

// CollectionOptions are optional deps for NewCollectionBusiness.
type CollectionOptions struct {
	Instruments         InstrumentSource
	Scheduler           RenewalScheduler
	SettlementScheduler SettlementScheduler
}

// NewCollectionBusiness constructs the simplified collection orchestrator.
// ledger and ledgerAccounts may be nil/empty — cash posting is then skipped.
// instruments may be nil — then ConfirmPayment skips COF pin (renewals fall back to profile).
//
// Optional trailing args: InstrumentSource and/or CollectionOptions.
func NewCollectionBusiness(
	checkout CheckoutIntegration,
	invoiceEng InvoiceEngine,
	invoiceRepo repository.InvoiceRepository,
	subBiz SubscriptionBusiness,
	planRepo repository.PlanRepository,
	compRepo repository.ComponentRepository,
	runRepo repository.BillingRunRepository,
	pricing *PricingEngine,
	ledger LedgerIntegration,
	ledgerAccounts CollectionLedgerAccounts,
	optional ...any,
) CollectionBusiness {
	var inst InstrumentSource
	var sched RenewalScheduler = NoopRenewalScheduler{}
	var settleSched SettlementScheduler = NoopSettlementScheduler{}
	for _, o := range optional {
		switch v := o.(type) {
		case InstrumentSource:
			if v != nil {
				inst = v
			}
		case RenewalScheduler:
			if v != nil {
				sched = v
			}
		case SettlementScheduler:
			if v != nil {
				settleSched = v
			}
		case CollectionOptions:
			if v.Instruments != nil {
				inst = v.Instruments
			}
			if v.Scheduler != nil {
				sched = v.Scheduler
			}
			if v.SettlementScheduler != nil {
				settleSched = v.SettlementScheduler
			}
		}
	}
	return &collectionBusiness{
		checkout:       checkout,
		invoiceEng:     invoiceEng,
		invoiceRepo:    invoiceRepo,
		subBiz:         subBiz,
		planRepo:       planRepo,
		instruments:    inst,
		scheduler:      sched,
		settleSched:    settleSched,
		compRepo:       compRepo,
		runRepo:        runRepo,
		pricing:        pricing,
		ledger:         ledger,
		ledgerAccounts: ledgerAccounts,
		obs:            observability.NewMetrics(),
	}
}

// CollectPayment opens hosted checkout for an issued invoice.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (b *collectionBusiness) CollectPayment(
	ctx context.Context,
	in CollectPaymentInput,
) (result *CollectionResult, err error) {
	ctx, span := b.obs.StartSpan(ctx, "CollectPayment")
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	if in.InvoiceID == "" {
		return nil, ErrInvoiceIDRequired
	}
	if b.checkout == nil {
		return nil, errors.New("checkout integration not configured")
	}

	invoice, err := b.invoiceRepo.GetByID(ctx, in.InvoiceID)
	if err != nil {
		return nil, err
	}

	// Already paid → idempotent success, no checkout needed.
	if invoice.State == models.InvoiceStatePaid {
		return &CollectionResult{
			InvoiceID:       invoice.GetID(),
			SubscriptionID:  invoice.SubscriptionID,
			AlreadyComplete: true,
		}, nil
	}

	if invoice.State != models.InvoiceStateIssued {
		return nil, apperrors.ErrInvoiceNotPayable
	}

	// Zero-total invoices settle without payment rails.
	if isZeroAmount(invoice.TotalAmount) {
		invoice, err = b.invoiceEng.RecordPayment(ctx, invoice.GetID())
		if err != nil {
			return nil, err
		}
		b.postCashIfNeeded(ctx, invoice)
		if actErr := b.activateSubscriptionIfPending(ctx, invoice.SubscriptionID); actErr != nil {
			return nil, actErr
		}
		return &CollectionResult{
			InvoiceID:       invoice.GetID(),
			SubscriptionID:  invoice.SubscriptionID,
			AlreadyComplete: true,
		}, nil
	}

	checkout, err := b.checkout.CreateInvoiceCheckout(ctx, invoice.GetID(), CheckoutOptions{
		ReturnURL: in.ReturnURL,
		Source:    CollectionSourceInvoice,
		Methods:   in.Methods,
	})
	if err != nil {
		return nil, err
	}

	if persistErr := b.persistCheckoutSession(ctx, invoice, checkout.SessionRef, CollectionSourceInvoice); persistErr != nil {
		util.Log(ctx).WithError(persistErr).Warn("could not persist checkout session on invoice")
	}

	return &CollectionResult{
		PageURL:        checkout.PageURL,
		SessionRef:     checkout.SessionRef,
		InvoiceID:      invoice.GetID(),
		SubscriptionID: invoice.SubscriptionID,
	}, nil
}

// StartSubscription creates a subscription and opens checkout for any upfront flat fee.
//
//nolint:nonamedreturns,funlen // named err for span; orchestration is intentionally sequential
func (b *collectionBusiness) StartSubscription(
	ctx context.Context,
	in StartSubscriptionInput,
) (result *CollectionResult, err error) {
	ctx, span := b.obs.StartSpan(ctx, "StartSubscription")
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	if err = b.validateStartSubscriptionInput(in); err != nil {
		return nil, err
	}
	if b.checkout == nil {
		return nil, errors.New("checkout integration not configured")
	}

	plan, err := b.planRepo.GetByID(ctx, in.PlanID)
	if err != nil {
		return nil, err
	}
	if plan.CatalogVersionID != in.CatalogVersionID {
		return nil, fmt.Errorf("%w: plan does not belong to catalog version", ErrCatalogVersionRequired)
	}

	components, err := b.compRepo.ListByPlanID(ctx, in.PlanID)
	if err != nil {
		return nil, err
	}

	// Create subscription as PENDING when a charge is expected; ACTIVE when free.
	upfront := b.rateUpfrontFlatFees(components)
	initialState := models.SubscriptionStatePending
	if upfront.IsZero() {
		initialState = models.SubscriptionStateActive
	}

	sub, err := b.subBiz.CreateSubscription(ctx, &models.Subscription{
		ProfileID:        in.ProfileID,
		PlanID:           in.PlanID,
		CatalogVersionID: in.CatalogVersionID,
		Currency:         in.Currency,
		State:            initialState,
		Data:             buildSubscriptionData(in),
	})
	if err != nil {
		return nil, err
	}

	// Free plan: no invoice, no checkout.
	if upfront.IsZero() {
		util.Log(ctx).
			WithField("subscription_id", sub.GetID()).
			Info("subscription activated with no upfront charge")
		return &CollectionResult{
			SubscriptionID:  sub.GetID(),
			AlreadyComplete: true,
		}, nil
	}

	invoice, err := b.createUpfrontInvoice(ctx, sub, plan, components, upfront)
	if err != nil {
		return nil, err
	}

	issued, err := b.invoiceEng.IssueInvoice(ctx, invoice.GetID())
	if err != nil {
		return nil, err
	}

	// Link invoice on the subscription so cancel can void it if unpaid.
	if _, patchErr := b.subBiz.PatchSubscriptionData(ctx, sub.GetID(), map[string]any{
		models.SubDataSignupInvoiceID: issued.GetID(),
	}); patchErr != nil {
		util.Log(ctx).WithError(patchErr).Warn("could not link signup invoice on subscription")
	}

	checkout, err := b.checkout.CreateInvoiceCheckout(ctx, issued.GetID(), CheckoutOptions{
		ReturnURL:        in.ReturnURL,
		Source:           CollectionSourceSubscription,
		PayerDisplayName: in.PayerDisplayName,
		Methods:          in.Methods,
	})
	if err != nil {
		return nil, err
	}

	if persistErr := b.persistCheckoutSession(ctx, issued, checkout.SessionRef, CollectionSourceSubscription); persistErr != nil {
		util.Log(ctx).WithError(persistErr).Warn("could not persist checkout session on invoice")
	}

	return &CollectionResult{
		PageURL:        checkout.PageURL,
		SessionRef:     checkout.SessionRef,
		InvoiceID:      issued.GetID(),
		SubscriptionID: sub.GetID(),
	}, nil
}

// ConfirmPayment settles a completed checkout session and activates any pending subscription.
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (b *collectionBusiness) ConfirmPayment(
	ctx context.Context,
	sessionRef string,
) (result *ConfirmPaymentResult, err error) {
	ctx, span := b.obs.StartSpan(ctx, "ConfirmPayment")
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	if sessionRef == "" {
		return nil, apperrors.ErrUnspecifiedReference
	}
	if b.checkout == nil {
		return nil, errors.New("checkout integration not configured")
	}

	// SettleFromCheckout is idempotent. Cash ledger posting is gated by
	// invoice.Data["ledgerPaymentTxnId"] so repeat confirms are safe.
	invoice, err := b.checkout.SettleFromCheckout(ctx, sessionRef)
	if err != nil {
		return nil, err
	}

	if invoice.State == models.InvoiceStatePaid {
		b.postCashIfNeeded(ctx, invoice)
		// Paid → no more settlement polls for this invoice.
		if b.settleSched != nil {
			_ = b.settleSched.CancelReminder(ctx, invoice.GetID())
		}
	}

	// Subscription activation is best-effort after a successful settle.
	// A missing or already-terminal subscription must not roll back payment capture.
	subState := ""
	if invoice.SubscriptionID != "" {
		if actErr := b.activateSubscriptionIfPending(ctx, invoice.SubscriptionID); actErr != nil {
			util.Log(ctx).
				WithError(actErr).
				WithField("subscription_id", invoice.SubscriptionID).
				Warn("could not activate subscription after payment")
		}
		if sub, getErr := b.subBiz.GetSubscription(ctx, invoice.SubscriptionID); getErr == nil {
			subState = sub.State
			// Notify product integrators that this subscription invoice was paid.
			if invoice.State == models.InvoiceStatePaid {
				b.bootstrapPeriodAndInstrument(ctx, sub, invoice)
				// Reload after patch for accurate NotifyBilled data.
				if fresh, g2 := b.subBiz.GetSubscription(ctx, invoice.SubscriptionID); g2 == nil {
					sub = fresh
					subState = sub.State
				}
				b.subBiz.NotifyBilled(ctx, sub, invoice.GetID())
				// First successful pay → schedule per-sub Trustage renew reminder.
				b.syncRenewalReminder(ctx, sub)
			}
		}
	}

	// Persist session ref if missing (settle path may not have written it).
	if invoice.Data == nil || invoice.Data[InvoiceDataCheckoutSessionRef] == nil {
		_ = b.persistCheckoutSession(ctx, invoice, sessionRef, "")
	}

	return &ConfirmPaymentResult{
		InvoiceID:         invoice.GetID(),
		InvoiceState:      invoice.State,
		SubscriptionID:    invoice.SubscriptionID,
		SubscriptionState: subState,
		Paid:              invoice.State == models.InvoiceStatePaid,
	}, nil
}

// bootstrapPeriodAndInstrument pins currentPeriodEnd and COF instrument after
// a successful interactive pay so silent renewals need no browser.
func (b *collectionBusiness) bootstrapPeriodAndInstrument(
	ctx context.Context,
	sub *models.Subscription,
	invoice *models.Invoice,
) {
	if sub == nil {
		return
	}
	patch := map[string]any{}
	// First period end = now + 1 calendar month (matches product monthly plans).
	if _, has := sub.Data[models.SubDataCurrentPeriodEnd]; !has {
		pe := time.Now().UTC().AddDate(0, 1, 0)
		if !sub.BillingAnchor.IsZero() {
			pe = sub.BillingAnchor.UTC().AddDate(0, 1, 0)
		}
		patch[models.SubDataCurrentPeriodEnd] = pe.Format(time.RFC3339)
	}
	patch[models.SubDataRenewAttemptCount] = 0
	if invoice != nil {
		patch[models.SubDataLastRenewInvoiceID] = invoice.GetID()
	}
	// Pin instrument from profile checkout clues when not already on the sub.
	if instrumentFromData(sub.Data) == nil && b.instruments != nil {
		if inst, err := b.instruments.ResolveInstrument(ctx, sub); err == nil && inst != nil {
			patch[models.SubDataPaymentMethodID] = inst.PaymentMethodID
			patch[models.SubDataProviderCustomerID] = inst.CustomerID
			if inst.Provider != "" {
				patch[models.SubDataPaymentProvider] = inst.Provider
			} else {
				patch[models.SubDataPaymentProvider] = "flutterwave"
			}
		}
	}
	if len(patch) == 0 {
		return
	}
	if _, err := b.subBiz.PatchSubscriptionData(ctx, sub.GetID(), patch); err != nil {
		util.Log(ctx).WithError(err).WithField("subscription_id", sub.GetID()).
			Warn("could not bootstrap period/instrument after payment")
	}
}

// postCashIfNeeded posts Debit Cash / Credit AR once per invoice.
// Uses invoice.Data["ledgerPaymentTxnId"] as an idempotency marker.
func (b *collectionBusiness) postCashIfNeeded(ctx context.Context, invoice *models.Invoice) {
	if b.ledger == nil {
		return
	}
	if invoice.Data != nil {
		if _, ok := invoice.Data["ledgerPaymentTxnId"]; ok {
			return // already posted
		}
	}
	txnID, err := b.ledger.PostPaymentToLedger(
		ctx, invoice, b.ledgerAccounts.CashAccountID, b.ledgerAccounts.ARAccountID,
	)
	if err != nil {
		util.Log(ctx).WithError(err).WithField("invoice_id", invoice.GetID()).
			Warn("could not post payment to ledger")
		return
	}
	if txnID == "" {
		return
	}
	fresh, getErr := b.invoiceRepo.GetByID(ctx, invoice.GetID())
	if getErr != nil {
		return
	}
	if fresh.Data == nil {
		fresh.Data = make(data.JSONMap)
	}
	fresh.Data["ledgerPaymentTxnId"] = txnID
	if _, updateErr := b.invoiceRepo.Update(ctx, fresh); updateErr != nil {
		util.Log(ctx).WithError(updateErr).Warn("could not persist ledger payment txn id")
	}
}

func (b *collectionBusiness) validateStartSubscriptionInput(in StartSubscriptionInput) error {
	if in.ProfileID == "" {
		return ErrSubscriptionProfileRequired
	}
	if in.PlanID == "" {
		return ErrSubscriptionPlanRequired
	}
	if in.CatalogVersionID == "" {
		return ErrCatalogVersionRequired
	}
	if in.Currency == "" || len(in.Currency) != 3 {
		return ErrSubscriptionCurrencyRequired
	}
	return nil
}

// buildSubscriptionData packs external-entity integration fields into Data.
func buildSubscriptionData(in StartSubscriptionInput) data.JSONMap {
	out := data.JSONMap{}
	for k, v := range in.Metadata {
		k = strings.TrimSpace(k)
		if k == "" || v == "" {
			continue
		}
		// Reserved keys are set from first-class fields below.
		if k == models.SubDataExternalEntityID ||
			k == models.SubDataExternalEntityType ||
			k == models.SubDataIntegrationRouteID ||
			k == models.SubDataSignupInvoiceID {
			continue
		}
		out[k] = v
	}
	if id := strings.TrimSpace(in.ExternalEntityID); id != "" {
		out[models.SubDataExternalEntityID] = id
	}
	if t := strings.TrimSpace(in.ExternalEntityType); t != "" {
		out[models.SubDataExternalEntityType] = t
	}
	if r := strings.TrimSpace(in.IntegrationRouteID); r != "" {
		out[models.SubDataIntegrationRouteID] = r
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// rateUpfrontFlatFees sums FLAT component fees on the plan (first-period charge).
func (b *collectionBusiness) rateUpfrontFlatFees(components []*models.Component) decimalx.Decimal {
	total := decimalx.Zero()
	for _, comp := range components {
		amount := flatFeeAmount(comp)
		if amount.IsPositive() {
			total = total.Add(amount)
		}
	}
	return total
}

// flatFeeAmount returns the positive flat fee for a FLAT component, or zero.
func flatFeeAmount(comp *models.Component) decimalx.Decimal {
	if comp == nil || comp.PricingModel != models.PricingModelFlat {
		return decimalx.Zero()
	}
	if len(comp.Tiers) == 0 || comp.Tiers[0].FlatFee == nil {
		return decimalx.Zero()
	}
	fee := *comp.Tiers[0].FlatFee
	if !fee.IsPositive() {
		return decimalx.Zero()
	}
	return fee
}

// createUpfrontInvoice builds a completed billing run + issued-ready draft invoice
// for the subscription's first flat-fee charge.
func (b *collectionBusiness) createUpfrontInvoice(
	ctx context.Context,
	sub *models.Subscription,
	plan *models.Plan,
	components []*models.Component,
	_ decimalx.Decimal,
) (*models.Invoice, error) {
	now := time.Now()
	// First period: billing anchor → +1 month (calendar).
	periodEnd := now.AddDate(0, 1, 0)
	idempotency := fmt.Sprintf("signup:%s:%s", sub.GetID(), now.Format("2006-01-02"))

	run := &models.BillingRun{
		SubscriptionID:   sub.GetID(),
		ProfileID:        sub.ProfileID,
		CatalogVersionID: sub.CatalogVersionID,
		State:            models.BillingRunStateCompleted,
		PeriodStart:      now,
		PeriodEnd:        periodEnd,
		StartedAt:        &now,
		CompletedAt:      &now,
		Idempotency:      idempotency,
		Data: data.JSONMap{
			"source": CollectionSourceSubscription,
			"planId": plan.GetID(),
		},
	}
	run.GenID(ctx)
	if err := b.runRepo.Create(ctx, run); err != nil {
		if data.ErrorIsDuplicateKey(err) {
			existing, getErr := b.runRepo.GetByIdempotency(ctx, idempotency)
			if getErr != nil {
				return nil, getErr
			}
			if existing.InvoiceID != "" {
				return b.invoiceRepo.GetByID(ctx, existing.InvoiceID)
			}
			return nil, fmt.Errorf("signup billing run exists without invoice: %s", existing.GetID())
		}
		return nil, err
	}

	ratedLines := make([]*models.RatedLine, 0, len(components))
	for _, comp := range components {
		amount := flatFeeAmount(comp)
		if !amount.IsPositive() {
			continue
		}
		description := comp.Name + ": flat fee"
		rl := &models.RatedLine{
			BillingRunID:   run.GetID(),
			SubscriptionID: sub.GetID(),
			ComponentID:    comp.GetID(),
			Description:    description,
			Quantity:       decimalx.NewFromInt64(1).Ptr(),
			UnitPrice:      amount.Ptr(),
			Amount:         amount.Ptr(),
			Currency:       sub.Currency,
			PricingModel:   comp.PricingModel,
		}
		rl.GenID(ctx)
		ratedLines = append(ratedLines, rl)
	}

	if len(ratedLines) == 0 {
		return nil, ErrPlanHasNoUpfrontCharge
	}

	invoice, err := b.invoiceEng.GenerateInvoice(ctx, run, ratedLines, nil, decimalx.Zero())
	if err != nil {
		return nil, err
	}

	run.InvoiceID = invoice.GetID()
	if _, updateErr := b.runRepo.Update(ctx, run); updateErr != nil {
		return nil, updateErr
	}

	return invoice, nil
}

func (b *collectionBusiness) persistCheckoutSession(
	ctx context.Context,
	invoice *models.Invoice,
	sessionRef, source string,
) error {
	if invoice == nil || sessionRef == "" {
		return nil
	}
	// Re-load to avoid overwriting concurrent changes with a stale pointer.
	fresh, err := b.invoiceRepo.GetByID(ctx, invoice.GetID())
	if err != nil {
		return err
	}
	if fresh.Data == nil {
		fresh.Data = make(data.JSONMap)
	}
	fresh.Data[InvoiceDataCheckoutSessionRef] = sessionRef
	if source != "" {
		fresh.Data[InvoiceDataCollectionSource] = source
	}
	_, err = b.invoiceRepo.Update(ctx, fresh)
	if err != nil {
		return err
	}
	// Per-invoice Trustage settle one-shot (abandoned browser recovery).
	if b.settleSched != nil {
		if sErr := b.settleSched.ScheduleFirst(ctx, fresh); sErr != nil {
			util.Log(ctx).WithError(sErr).WithField("invoice_id", fresh.GetID()).
				Warn("could not schedule trustage settlement reminder")
		}
	}
	return nil
}

// CancelSubscription cancels a subscription.
//   - PENDING: hard cancel + void open signup invoice
//   - ACTIVE: soft cancel (cancel_at_period_end) — access until period end;
//     RenewalSweeper hard-cancels after period end and emits cancelled
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (b *collectionBusiness) CancelSubscription(
	ctx context.Context,
	subscriptionID string,
) (result *CancelSubscriptionResult, err error) {
	ctx, span := b.obs.StartSpan(ctx, "CollectionCancelSubscription")
	defer func() { b.obs.EndSpan(ctx, span, err) }()

	if subscriptionID == "" {
		return nil, ErrSubscriptionIDRequired
	}

	sub, err := b.subBiz.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	voidedInvoiceID := ""
	if invID, voidErr := b.voidOpenSignupInvoice(ctx, sub); voidErr != nil {
		util.Log(ctx).WithError(voidErr).Warn("could not void open signup invoice on cancel")
	} else {
		voidedInvoiceID = invID
	}

	// Already cancelled → idempotent success.
	if sub.State == models.SubscriptionStateCancelled {
		return &CancelSubscriptionResult{
			SubscriptionID:    sub.GetID(),
			SubscriptionState: sub.State,
			VoidedInvoiceID:   voidedInvoiceID,
		}, nil
	}

	switch sub.State {
	case models.SubscriptionStatePending:
		cancelled, cErr := b.subBiz.CancelPendingSubscription(ctx, subscriptionID)
		if cErr != nil {
			return nil, cErr
		}
		// Never paid — no renew reminder (cancel any stray).
		if b.scheduler != nil {
			_ = b.scheduler.CancelReminder(ctx, cancelled.GetID())
		}
		return &CancelSubscriptionResult{
			SubscriptionID:    cancelled.GetID(),
			SubscriptionState: cancelled.State,
			VoidedInvoiceID:   voidedInvoiceID,
		}, nil
	case models.SubscriptionStateActive:
		// Soft cancel: keep ACTIVE, skip rebill, finalize after period end.
		if cancelAtPeriodEnd(sub.Data) {
			return &CancelSubscriptionResult{
				SubscriptionID:    sub.GetID(),
				SubscriptionState: sub.State,
				VoidedInvoiceID:   voidedInvoiceID,
			}, nil
		}
		pe := currentPeriodEnd(sub)
		if pe.IsZero() {
			pe = time.Now().UTC().AddDate(0, 1, 0)
		}
		patched, pErr := b.subBiz.PatchSubscriptionData(ctx, subscriptionID, map[string]any{
			models.SubDataCancelAtPeriodEnd: true,
			models.SubDataCurrentPeriodEnd:  pe.UTC().Format(time.RFC3339),
		})
		if pErr != nil {
			return nil, pErr
		}
		// Lifecycle: product mirrors cancel_at_period_end (still ACTIVE until End).
		b.subBiz.NotifyLifecycle(ctx, models.SubscriptionEventCancelled, patched, "")
		// Reschedule Trustage: drop rebill, keep finalize-at-period-end only.
		b.syncRenewalReminder(ctx, patched)
		return &CancelSubscriptionResult{
			SubscriptionID:    patched.GetID(),
			SubscriptionState: patched.State,
			VoidedInvoiceID:   voidedInvoiceID,
		}, nil
	default:
		return nil, apperrors.ErrSubscriptionNotActive.Extend(
			fmt.Sprintf("cannot cancel subscription in state %s", sub.State),
		)
	}
}

func (b *collectionBusiness) syncRenewalReminder(ctx context.Context, sub *models.Subscription) {
	if b.scheduler == nil || sub == nil {
		return
	}
	if err := b.scheduler.SyncReminder(ctx, sub); err != nil {
		util.Log(ctx).WithError(err).WithField("subscription_id", sub.GetID()).
			Warn("could not sync trustage renewal reminder")
	}
}

func (b *collectionBusiness) voidOpenSignupInvoice(
	ctx context.Context,
	sub *models.Subscription,
) (string, error) {
	if sub == nil || sub.Data == nil {
		return "", nil
	}
	invID, ok := sub.Data[models.SubDataSignupInvoiceID].(string)
	if !ok || invID == "" {
		return "", nil
	}
	inv, err := b.invoiceRepo.GetByID(ctx, invID)
	if err != nil {
		return "", err
	}
	if inv.State != models.InvoiceStateIssued && inv.State != models.InvoiceStateDraft {
		return "", nil
	}
	voided, err := b.invoiceEng.VoidInvoice(ctx, invID)
	if err != nil {
		return "", err
	}
	return voided.GetID(), nil
}

func (b *collectionBusiness) activateSubscriptionIfPending(ctx context.Context, subscriptionID string) error {
	if subscriptionID == "" {
		return nil
	}
	sub, err := b.subBiz.ActivateSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("activate subscription: %w", err)
	}
	util.Log(ctx).
		WithField("subscription_id", sub.GetID()).
		WithField("state", sub.State).
		Info("subscription activated after payment")
	return nil
}

// Note: callers of activateSubscriptionIfPending after successful settle should
// treat errors as non-fatal — payment capture is the source of truth.

func isZeroAmount(d *decimalx.Decimal) bool {
	return d == nil || d.IsZero() || d.IsNegative()
}
