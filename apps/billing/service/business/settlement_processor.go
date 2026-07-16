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
	"strconv"
	"strings"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util"
)

// InvoiceSettleResult is the outcome of processing one invoice (Trustage fire).
type InvoiceSettleResult struct {
	InvoiceID    string `json:"invoice_id"`
	SessionRef   string `json:"session_ref,omitempty"`
	Action       string `json:"action"` // settled | pending | skipped | exhausted | error
	Paid         bool   `json:"paid"`
	Reason       string `json:"reason,omitempty"`
	NextSettleAt string `json:"next_settle_at,omitempty"`
	Attempt      int    `json:"attempt"`
}

// SettlementProcessor settles a single open checkout invoice.
// Invoked only by Trustage: POST /_internal/billing/invoices/{id}/settle.
// No bulk scan of invoices.
type SettlementProcessor interface {
	ProcessInvoice(ctx context.Context, invoiceID string) (*InvoiceSettleResult, error)
}

// SettlementSweeper is a historical alias for SettlementProcessor.
type SettlementSweeper = SettlementProcessor

type settlementProcessor struct {
	invoiceRepo repository.InvoiceRepository
	collection  CollectionBusiness
	scheduler   SettlementScheduler
	cfg         SettlementConfig
	obs         *observability.Metrics
}

// NewSettlementProcessor constructs the single-invoice settle worker.
func NewSettlementProcessor(
	invoiceRepo repository.InvoiceRepository,
	collection CollectionBusiness,
	cfg SettlementConfig,
	scheduler ...SettlementScheduler,
) SettlementProcessor {
	var sched SettlementScheduler = NoopSettlementScheduler{}
	if len(scheduler) > 0 && scheduler[0] != nil {
		sched = scheduler[0]
	}
	return &settlementProcessor{
		invoiceRepo: invoiceRepo,
		collection:  collection,
		scheduler:   sched,
		cfg:         cfg,
		obs:         observability.NewMetrics(),
	}
}

// NewSettlementSweeper keeps the old constructor name for call sites/tests.
// batchSize is ignored (no bulk path).
func NewSettlementSweeper(
	invoiceRepo repository.InvoiceRepository,
	collection CollectionBusiness,
	_ int,
	scheduler ...SettlementScheduler,
) SettlementProcessor {
	return NewSettlementProcessor(invoiceRepo, collection, NewSettlementConfigFromEnv(0, ""), scheduler...)
}

// ProcessInvoice handles one invoice settlement attempt.
func (s *settlementProcessor) ProcessInvoice(
	ctx context.Context,
	invoiceID string,
) (result *InvoiceSettleResult, err error) {
	ctx, span := s.obs.StartSpan(ctx, "ProcessInvoiceSettlement")
	defer func() { s.obs.EndSpan(ctx, span, err) }()

	out := &InvoiceSettleResult{InvoiceID: invoiceID}
	if invoiceID == "" {
		return out, fmt.Errorf("invoice id required")
	}
	if s.collection == nil {
		return out, errors.New("collection business not configured")
	}

	inv, getErr := s.invoiceRepo.GetByID(ctx, invoiceID)
	if getErr != nil {
		return out, getErr
	}

	// Terminal / no session → drop Trustage reminder.
	if inv.State == models.InvoiceStatePaid {
		_ = s.scheduler.CancelReminder(ctx, invoiceID)
		out.Action = "skipped"
		out.Paid = true
		out.Reason = "already_paid"
		return out, nil
	}
	if inv.State != models.InvoiceStateIssued {
		_ = s.scheduler.CancelReminder(ctx, invoiceID)
		out.Action = "skipped"
		out.Reason = "not_issued:" + inv.State
		return out, nil
	}

	sessionRef, _ := inv.Data[InvoiceDataCheckoutSessionRef].(string)
	sessionRef = strings.TrimSpace(sessionRef)
	out.SessionRef = sessionRef
	if sessionRef == "" {
		_ = s.scheduler.CancelReminder(ctx, invoiceID)
		out.Action = "skipped"
		out.Reason = "no_checkout_session"
		return out, nil
	}

	now := time.Now().UTC()
	attempts := settleAttemptCount(inv.Data) + 1
	out.Attempt = attempts
	_ = s.patchInvoiceData(ctx, invoiceID, map[string]any{
		InvoiceDataLastSettleAttemptAt: now.Format(time.RFC3339),
		InvoiceDataSettleAttemptCount:  attempts,
	})

	confirmed, confirmErr := s.collection.ConfirmPayment(ctx, sessionRef)
	if confirmErr != nil {
		if errors.Is(confirmErr, ErrCheckoutNotCompleted) {
			return s.rearmPending(ctx, invoiceID, attempts, out, "checkout_not_completed")
		}
		// Transient/other error: still re-arm if attempts remain.
		util.Log(ctx).WithError(confirmErr).WithField("invoice_id", invoiceID).
			Warn("settlement: confirm failed")
		if attempts >= s.cfg.MaxAttempts {
			_ = s.scheduler.CancelReminder(ctx, invoiceID)
			out.Action = "exhausted"
			out.Reason = confirmErr.Error()
			return out, nil
		}
		return s.rearmPending(ctx, invoiceID, attempts, out, confirmErr.Error())
	}

	if confirmed != nil && confirmed.Paid {
		_ = s.scheduler.CancelReminder(ctx, invoiceID)
		out.Action = "settled"
		out.Paid = true
		util.Log(ctx).WithField("invoice_id", invoiceID).
			WithField("session_ref", sessionRef).
			Info("settlement: invoice settled via trustage reminder")
		return out, nil
	}

	// Confirmed but not paid (edge) — treat as pending.
	return s.rearmPending(ctx, invoiceID, attempts, out, "not_paid")
}

func (s *settlementProcessor) rearmPending(
	ctx context.Context,
	invoiceID string,
	attempts int,
	out *InvoiceSettleResult,
	reason string,
) (*InvoiceSettleResult, error) {
	if attempts >= s.cfg.MaxAttempts {
		_ = s.scheduler.CancelReminder(ctx, invoiceID)
		out.Action = "exhausted"
		out.Reason = reason
		util.Log(ctx).WithField("invoice_id", invoiceID).
			WithField("attempts", attempts).
			Info("settlement: max attempts — trustage reminder cancelled")
		return out, nil
	}
	// nextAttemptIndex for schedule = attempts (0-based next is current count)
	nextAt, ok := s.cfg.RescheduleAfterPoll(time.Now().UTC(), attempts)
	if !ok {
		_ = s.scheduler.CancelReminder(ctx, invoiceID)
		out.Action = "exhausted"
		out.Reason = reason
		return out, nil
	}
	if err := s.scheduler.EnsureAt(ctx, invoiceID, nextAt); err != nil {
		out.Action = "error"
		out.Reason = err.Error()
		return out, err
	}
	_ = s.patchInvoiceData(ctx, invoiceID, map[string]any{
		InvoiceDataNextSettleAt: nextAt.Format(time.RFC3339),
	})
	out.Action = "pending"
	out.Reason = reason
	out.NextSettleAt = nextAt.Format(time.RFC3339)
	return out, nil
}

func (s *settlementProcessor) patchInvoiceData(ctx context.Context, invoiceID string, patch map[string]any) error {
	inv, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return err
	}
	if inv.Data == nil {
		inv.Data = make(data.JSONMap)
	}
	for k, v := range patch {
		inv.Data[k] = v
	}
	_, err = s.invoiceRepo.Update(ctx, inv)
	return err
}

func settleAttemptCount(d data.JSONMap) int {
	if d == nil {
		return 0
	}
	v, ok := d[InvoiceDataSettleAttemptCount]
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
