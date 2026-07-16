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
	"fmt"
	"strconv"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
)

// SubscriptionRenewResult is the outcome of processing one subscription
// (Trustage per-sub fire). There is no batch renew path.
type SubscriptionRenewResult struct {
	SubscriptionID string `json:"subscription_id"`
	Action         string `json:"action"` // renewed | failed | skipped | past_due | finalized
	Collected      bool   `json:"collected"`
	Failed         bool   `json:"failed"`
	Reason         string `json:"reason,omitempty"`
	NextRenewAt    string `json:"next_renew_at,omitempty"`
}

// processOutcome is internal bookkeeping for a single charge attempt.
type processOutcome struct {
	collected  bool
	failed     bool
	hostedNeed bool
}

// RenewalProcessor bills a single ACTIVE subscription via FLAT renew invoice + COF.
// Invoked only by Trustage one-shots: POST /_internal/billing/subscriptions/{id}/renew.
// No bulk scan of subscriptions.
type RenewalProcessor interface {
	ProcessSubscription(ctx context.Context, subscriptionID string) (*SubscriptionRenewResult, error)
}

// RenewalSweeper is an alias kept for call-site compatibility.
type RenewalSweeper = RenewalProcessor

type renewalProcessor struct {
	subBiz      SubscriptionBusiness
	planRepo    repository.PlanRepository
	compRepo    repository.ComponentRepository
	runRepo     repository.BillingRunRepository
	invoiceRepo repository.InvoiceRepository
	invoiceEng  InvoiceEngine
	instruments InstrumentSource
	collector   PaymentCollector
	scheduler   RenewalScheduler
	cfg         RenewalConfig
	obs         *observability.Metrics
}

// NewRenewalSweeper constructs the single-subscription rebill worker.
// scheduler may be nil → NoopRenewalScheduler.
// There is no bulk subscription list — only ProcessSubscription(id).
func NewRenewalSweeper(
	subBiz SubscriptionBusiness,
	planRepo repository.PlanRepository,
	compRepo repository.ComponentRepository,
	runRepo repository.BillingRunRepository,
	invoiceRepo repository.InvoiceRepository,
	invoiceEng InvoiceEngine,
	instruments InstrumentSource,
	collector PaymentCollector,
	cfg RenewalConfig,
	scheduler ...RenewalScheduler,
) RenewalProcessor {
	var sched RenewalScheduler = NoopRenewalScheduler{}
	if len(scheduler) > 0 && scheduler[0] != nil {
		sched = scheduler[0]
	}
	return &renewalProcessor{
		subBiz:      subBiz,
		planRepo:    planRepo,
		compRepo:    compRepo,
		runRepo:     runRepo,
		invoiceRepo: invoiceRepo,
		invoiceEng:  invoiceEng,
		instruments: instruments,
		collector:   collector,
		scheduler:   sched,
		cfg:         cfg,
		obs:         observability.NewMetrics(),
	}
}

// ProcessSubscription is the Trustage per-subscription entry point.
// Loads only the given id — never scans other subscriptions.
func (s *renewalProcessor) ProcessSubscription(
	ctx context.Context,
	subscriptionID string,
) (result *SubscriptionRenewResult, err error) {
	ctx, span := s.obs.StartSpan(ctx, "ProcessSubscriptionRenewal")
	defer func() { s.obs.EndSpan(ctx, span, err) }()

	out := &SubscriptionRenewResult{SubscriptionID: subscriptionID}
	if subscriptionID == "" {
		return out, fmt.Errorf("subscription id required")
	}
	sub, getErr := s.subBiz.GetSubscription(ctx, subscriptionID)
	if getErr != nil {
		return out, getErr
	}
	now := time.Now().UTC()

	// Soft-cancel finalize when period is over (this sub only).
	if cancelAtPeriodEnd(sub.Data) {
		pe := currentPeriodEnd(sub)
		if pe.IsZero() || !pe.After(now) {
			if _, cErr := s.subBiz.CancelSubscription(ctx, subscriptionID); cErr != nil {
				out.Action = "error"
				out.Reason = cErr.Error()
				return out, cErr
			}
			_ = s.scheduler.CancelReminder(ctx, subscriptionID)
			out.Action = "finalized"
			out.Reason = "cancel_at_period_end"
			return out, nil
		}
		out.Action = "skipped"
		out.Reason = "cancel_at_period_end_pending"
		s.syncReminder(ctx, subscriptionID)
		return out, nil
	}

	if sub.State != models.SubscriptionStateActive {
		_ = s.scheduler.CancelReminder(ctx, subscriptionID)
		out.Action = "skipped"
		out.Reason = "not_active:" + sub.State
		return out, nil
	}

	if skip, reason := s.shouldSkip(sub, now); skip {
		out.Action = "skipped"
		out.Reason = reason
		if reason == "max_attempts" {
			_ = s.scheduler.CancelReminder(ctx, subscriptionID)
		} else {
			// Not due yet — keep / re-arm Trustage for the planned next fire.
			s.syncReminder(ctx, subscriptionID)
		}
		return out, nil
	}

	outcome := &processOutcome{}
	if procErr := s.processOne(ctx, sub, now, outcome); procErr != nil {
		out.Action = "failed"
		out.Failed = true
		out.Reason = procErr.Error()
		s.syncReminder(ctx, subscriptionID)
		return out, procErr
	}

	fresh, _ := s.subBiz.GetSubscription(ctx, subscriptionID)
	if fresh == nil {
		fresh = sub
	}
	attempts := renewAttemptCount(fresh.Data)
	switch {
	case outcome.collected:
		out.Action = "renewed"
		out.Collected = true
	case outcome.failed || outcome.hostedNeed:
		out.Failed = true
		if attempts >= s.cfg.MaxAttempts {
			out.Action = "past_due"
			out.Reason = "max_attempts"
		} else {
			out.Action = "failed"
			out.Reason = "charge_failed"
		}
	default:
		out.Action = "skipped"
	}
	s.syncReminder(ctx, subscriptionID)
	if reloaded, gErr := s.subBiz.GetSubscription(ctx, subscriptionID); gErr == nil && reloaded.Data != nil {
		if n, ok := reloaded.Data[models.SubDataNextRenewAt].(string); ok && n != "" {
			out.NextRenewAt = n
		}
	}
	return out, nil
}

func (s *renewalProcessor) syncReminder(ctx context.Context, subscriptionID string) {
	if s.scheduler == nil {
		return
	}
	fresh, err := s.subBiz.GetSubscription(ctx, subscriptionID)
	if err != nil {
		util.Log(ctx).WithError(err).WithField("subscription_id", subscriptionID).
			Warn("renewal: reload for reminder sync failed")
		return
	}
	if syncErr := s.scheduler.SyncReminder(ctx, fresh); syncErr != nil {
		util.Log(ctx).WithError(syncErr).WithField("subscription_id", subscriptionID).
			Warn("renewal: trustage reminder sync failed")
	}
}

func (s *renewalProcessor) shouldSkip(sub *models.Subscription, now time.Time) (bool, string) {
	if sub == nil {
		return true, "nil"
	}
	if cancelAtPeriodEnd(sub.Data) {
		return true, "cancel_at_period_end"
	}
	periodEnd := currentPeriodEnd(sub)
	if periodEnd.IsZero() {
		periodEnd = sub.BillingAnchor.AddDate(0, 1, 0)
		if periodEnd.IsZero() || periodEnd.Before(sub.StartAt) {
			periodEnd = sub.StartAt.AddDate(0, 1, 0)
		}
	}
	attempts := renewAttemptCount(sub.Data)
	lastAt := lastRenewAttemptAt(sub.Data)
	if attempts >= s.cfg.MaxAttempts {
		return true, "max_attempts"
	}
	if !s.cfg.AttemptDue(now, periodEnd, attempts, lastAt) {
		return true, "not_due_for_attempt"
	}
	return false, ""
}

func (s *renewalProcessor) processOne(
	ctx context.Context,
	sub *models.Subscription,
	now time.Time,
	out *processOutcome,
) error {
	periodEnd := currentPeriodEnd(sub)
	if periodEnd.IsZero() {
		periodEnd = sub.BillingAnchor.AddDate(0, 1, 0)
		if periodEnd.Before(now.Add(-365 * 24 * time.Hour)) {
			periodEnd = now
		}
	}
	periodStart := periodEnd
	nextEnd := periodStart.AddDate(0, 1, 0)
	periodKey := fmt.Sprintf("renew:%s:%s", sub.GetID(), periodStart.UTC().Format("2006-01-02"))

	invoice, err := s.ensureRenewalInvoice(ctx, sub, periodStart, nextEnd, periodKey)
	if err != nil {
		return err
	}
	if invoice.State == models.InvoiceStatePaid {
		_ = s.markPeriodPaid(ctx, sub, invoice, nextEnd)
		out.collected = true
		return nil
	}
	if invoice.State != models.InvoiceStateIssued {
		issued, issErr := s.invoiceEng.IssueInvoice(ctx, invoice.GetID())
		if issErr != nil {
			return issErr
		}
		invoice = issued
	}

	attempts := renewAttemptCount(sub.Data) + 1
	_, _ = s.subBiz.PatchSubscriptionData(ctx, sub.GetID(), map[string]any{
		models.SubDataLastRenewAttemptAt: now.Format(time.RFC3339),
		models.SubDataRenewAttemptCount:  attempts,
		models.SubDataLastRenewInvoiceID: invoice.GetID(),
		models.SubDataRenewPeriodKey:     periodKey,
	})

	inst, instErr := s.instruments.ResolveInstrument(ctx, sub)
	if instErr != nil || inst == nil {
		out.hostedNeed = true
		util.Log(ctx).WithError(instErr).WithField("subscription_id", sub.GetID()).
			Info("renewal: no saved instrument — product should open hosted collect")
		s.notifyPaymentFailed(ctx, sub, invoice.GetID(), "no_saved_instrument")
		if attempts >= s.cfg.MaxAttempts {
			s.notifyPastDue(ctx, sub, invoice.GetID())
		}
		return nil
	}

	if s.collector == nil {
		return fmt.Errorf("payment collector not configured")
	}

	promptID, chargeErr := s.collector.CollectCOF(ctx, invoice, inst, s.cfg.DefaultRoute)
	if chargeErr != nil {
		out.failed = true
		s.notifyPaymentFailed(ctx, sub, invoice.GetID(), chargeErr.Error())
		if attempts >= s.cfg.MaxAttempts {
			s.notifyPastDue(ctx, sub, invoice.GetID())
		}
		return chargeErr
	}

	st, waitErr := WaitCOF(ctx, s.collector, promptID, 45*time.Second, 2*time.Second)
	if waitErr != nil {
		out.failed = true
		s.notifyPaymentFailed(ctx, sub, invoice.GetID(), waitErr.Error())
		if attempts >= s.cfg.MaxAttempts {
			s.notifyPastDue(ctx, sub, invoice.GetID())
		}
		return waitErr
	}
	if st != commonv1.STATUS_SUCCESSFUL {
		out.failed = true
		s.notifyPaymentFailed(ctx, sub, invoice.GetID(), "charge_"+st.String())
		if attempts >= s.cfg.MaxAttempts {
			s.notifyPastDue(ctx, sub, invoice.GetID())
		}
		return nil
	}

	paidInv, payErr := s.invoiceEng.RecordPayment(ctx, invoice.GetID())
	if payErr != nil {
		return payErr
	}
	if markErr := s.markPeriodPaid(ctx, sub, paidInv, nextEnd); markErr != nil {
		return markErr
	}
	out.collected = true
	return nil
}

func (s *renewalProcessor) ensureRenewalInvoice(
	ctx context.Context,
	sub *models.Subscription,
	periodStart, periodEnd time.Time,
	periodKey string,
) (*models.Invoice, error) {
	run := &models.BillingRun{
		SubscriptionID:   sub.GetID(),
		ProfileID:        sub.ProfileID,
		CatalogVersionID: sub.CatalogVersionID,
		State:            models.BillingRunStateCompleted,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		StartedAt:        ptrTime(time.Now().UTC()),
		CompletedAt:      ptrTime(time.Now().UTC()),
		Idempotency:      periodKey,
		Data: data.JSONMap{
			"source": "subscription_renewal",
			"planId": sub.PlanID,
		},
	}
	run.GenID(ctx)
	if err := s.runRepo.Create(ctx, run); err != nil {
		if data.ErrorIsDuplicateKey(err) {
			existing, getErr := s.runRepo.GetByIdempotency(ctx, periodKey)
			if getErr != nil {
				return nil, getErr
			}
			if existing.InvoiceID != "" {
				return s.invoiceRepo.GetByID(ctx, existing.InvoiceID)
			}
			return nil, fmt.Errorf("renew run exists without invoice: %s", existing.GetID())
		}
		return nil, err
	}

	components, err := s.compRepo.ListByPlanID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}
	ratedLines := make([]*models.RatedLine, 0, len(components))
	for _, comp := range components {
		amount := flatFeeAmount(comp)
		if !amount.IsPositive() {
			continue
		}
		rl := &models.RatedLine{
			BillingRunID:   run.GetID(),
			SubscriptionID: sub.GetID(),
			ComponentID:    comp.GetID(),
			Description:    comp.Name + ": renewal",
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

	invoice, err := s.invoiceEng.GenerateInvoice(ctx, run, ratedLines, nil, decimalx.Zero())
	if err != nil {
		return nil, err
	}
	run.InvoiceID = invoice.GetID()
	if _, uErr := s.runRepo.Update(ctx, run); uErr != nil {
		return nil, uErr
	}
	return s.invoiceEng.IssueInvoice(ctx, invoice.GetID())
}

func (s *renewalProcessor) markPeriodPaid(
	ctx context.Context,
	sub *models.Subscription,
	invoice *models.Invoice,
	nextPeriodEnd time.Time,
) error {
	_, err := s.subBiz.PatchSubscriptionData(ctx, sub.GetID(), map[string]any{
		models.SubDataCurrentPeriodEnd:   nextPeriodEnd.UTC().Format(time.RFC3339),
		models.SubDataRenewAttemptCount:  0,
		models.SubDataLastRenewInvoiceID: invoice.GetID(),
	})
	if err != nil {
		return err
	}
	fresh, getErr := s.subBiz.GetSubscription(ctx, sub.GetID())
	if getErr == nil {
		s.subBiz.NotifyBilled(ctx, fresh, invoice.GetID())
	}
	return nil
}

func (s *renewalProcessor) notifyPaymentFailed(ctx context.Context, sub *models.Subscription, invoiceID, reason string) {
	if s.subBiz == nil || sub == nil {
		return
	}
	_, _ = s.subBiz.PatchSubscriptionData(ctx, sub.GetID(), map[string]any{
		"lastRenewError": reason,
	})
	fresh, err := s.subBiz.GetSubscription(ctx, sub.GetID())
	if err != nil {
		return
	}
	s.subBiz.NotifyLifecycle(ctx, models.SubscriptionEventPaymentFailed, fresh, invoiceID)
}

func (s *renewalProcessor) notifyPastDue(ctx context.Context, sub *models.Subscription, invoiceID string) {
	if s.subBiz == nil || sub == nil {
		return
	}
	fresh, err := s.subBiz.GetSubscription(ctx, sub.GetID())
	if err != nil {
		return
	}
	s.subBiz.NotifyLifecycle(ctx, models.SubscriptionEventPastDue, fresh, invoiceID)
}

// --- subscription data helpers ---

func cancelAtPeriodEnd(d data.JSONMap) bool {
	if d == nil {
		return false
	}
	v, ok := d[models.SubDataCancelAtPeriodEnd]
	if !ok {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(t, "true") || t == "1"
	default:
		return false
	}
}

func currentPeriodEnd(sub *models.Subscription) time.Time {
	if sub == nil {
		return time.Time{}
	}
	if sub.Data != nil {
		if s, ok := sub.Data[models.SubDataCurrentPeriodEnd].(string); ok && s != "" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t.UTC()
			}
		}
	}
	if sub.EndAt != nil {
		return sub.EndAt.UTC()
	}
	if !sub.BillingAnchor.IsZero() {
		return sub.BillingAnchor.UTC().AddDate(0, 1, 0)
	}
	return sub.StartAt.UTC().AddDate(0, 1, 0)
}

func renewAttemptCount(d data.JSONMap) int {
	if d == nil {
		return 0
	}
	v, ok := d[models.SubDataRenewAttemptCount]
	if !ok {
		return 0
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(t)
		return n
	default:
		return 0
	}
}

func lastRenewAttemptAt(d data.JSONMap) *time.Time {
	if d == nil {
		return nil
	}
	s, _ := d[models.SubDataLastRenewAttemptAt].(string)
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	u := t.UTC()
	return &u
}

func ptrTime(t time.Time) *time.Time { return &t }
