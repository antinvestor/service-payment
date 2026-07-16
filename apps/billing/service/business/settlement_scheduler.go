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
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	workflowv1 "github.com/antinvestor/service-trustage/gen/go/workflow/v1"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

// Invoice data keys for settlement bookkeeping.
const (
	InvoiceDataSettleAttemptCount  = "settleAttemptCount"
	InvoiceDataLastSettleAttemptAt = "lastSettleAttemptAt"
	InvoiceDataNextSettleAt        = "nextSettleAt"
	InvoiceDataTrustageSettleWF    = "trustageSettleWorkflow"
)

// SettlementScheduler keeps one Trustage one-shot per open checkout invoice.
type SettlementScheduler interface {
	// ScheduleFirst arms the first settle poll after hosted checkout opens.
	ScheduleFirst(ctx context.Context, invoice *models.Invoice) error
	// EnsureAt creates/updates the Trustage one-shot at nextAt.
	EnsureAt(ctx context.Context, invoiceID string, nextAt time.Time) error
	// CancelReminder archives the Trustage workflow for the invoice.
	CancelReminder(ctx context.Context, invoiceID string) error
}

// NoopSettlementScheduler when Trustage is not configured.
type NoopSettlementScheduler struct{}

func (NoopSettlementScheduler) ScheduleFirst(context.Context, *models.Invoice) error { return nil }
func (NoopSettlementScheduler) EnsureAt(context.Context, string, time.Time) error    { return nil }
func (NoopSettlementScheduler) CancelReminder(context.Context, string) error         { return nil }

// SettlementSchedulerConfig wires Trustage callbacks for per-invoice settle.
type SettlementSchedulerConfig struct {
	BillingBaseURL        string
	AdminTokenPlaceholder string
	Settlement            SettlementConfig
	InvoiceRepo           repository.InvoiceRepository // optional: persist nextSettleAt
}

type trustageSettlementScheduler struct {
	client WorkflowClient
	cfg    SettlementSchedulerConfig
}

// NewTrustageSettlementScheduler builds a per-invoice settle scheduler.
func NewTrustageSettlementScheduler(client WorkflowClient, cfg SettlementSchedulerConfig) SettlementScheduler {
	if client == nil || strings.TrimSpace(cfg.BillingBaseURL) == "" {
		return NoopSettlementScheduler{}
	}
	if cfg.AdminTokenPlaceholder == "" {
		cfg.AdminTokenPlaceholder = "${BILLING_INTERNAL_ADMIN_TOKEN}"
	}
	cfg.BillingBaseURL = strings.TrimRight(cfg.BillingBaseURL, "/")
	return &trustageSettlementScheduler{client: client, cfg: cfg}
}

// TrustageSettleWorkflowName is the deterministic Trustage workflow name.
func TrustageSettleWorkflowName(invoiceID string) string {
	return "billing.invoice.settle." + invoiceID
}

func (s *trustageSettlementScheduler) ScheduleFirst(ctx context.Context, invoice *models.Invoice) error {
	if invoice == nil {
		return nil
	}
	// Attempt 0 due after first delay from now (checkout just opened).
	next, ok := s.cfg.Settlement.NextSettleAt(time.Now().UTC(), 0)
	if !ok {
		return nil
	}
	if err := s.EnsureAt(ctx, invoice.GetID(), next); err != nil {
		return err
	}
	s.patchNext(ctx, invoice.GetID(), next.Format(time.RFC3339), false)
	util.Log(ctx).
		WithField("invoice_id", invoice.GetID()).
		WithField("next_at", next.Format(time.RFC3339)).
		Info("settlement: trustage first reminder scheduled")
	return nil
}

func (s *trustageSettlementScheduler) EnsureAt(ctx context.Context, invoiceID string, nextAt time.Time) error {
	if invoiceID == "" {
		return nil
	}
	name := TrustageSettleWorkflowName(invoiceID)
	desiredCron := CronForOneShot(nextAt)

	id, active, cron, err := findWorkflow(ctx, s.client, name)
	if err != nil {
		return err
	}
	if active && cron == desiredCron {
		return nil
	}
	if id != "" {
		if aerr := archiveWorkflowByName(ctx, s.client, name); aerr != nil {
			return aerr
		}
	}

	dsl, err := buildSettleDSL(s.cfg, invoiceID, name, desiredCron)
	if err != nil {
		return err
	}
	createResp, err := s.client.CreateWorkflow(ctx, connect.NewRequest(&workflowv1.CreateWorkflowRequest{Dsl: dsl}))
	if err != nil {
		if !isAlreadyExistsErr(err) {
			return fmt.Errorf("settle schedule create %s: %w", name, err)
		}
		id, active, _, lerr := findWorkflow(ctx, s.client, name)
		if lerr != nil {
			return lerr
		}
		if active || id == "" {
			return nil
		}
		return activateWorkflow(ctx, s.client, name, id)
	}
	return activateWorkflow(ctx, s.client, name, createResp.Msg.GetWorkflow().GetId())
}

func (s *trustageSettlementScheduler) CancelReminder(ctx context.Context, invoiceID string) error {
	if invoiceID == "" {
		return nil
	}
	name := TrustageSettleWorkflowName(invoiceID)
	if err := archiveWorkflowByName(ctx, s.client, name); err != nil {
		return err
	}
	s.patchNext(ctx, invoiceID, "", true)
	util.Log(ctx).WithField("invoice_id", invoiceID).
		WithField("workflow", name).Info("settlement: trustage reminder cancelled")
	return nil
}

func (s *trustageSettlementScheduler) patchNext(ctx context.Context, invoiceID, nextRFC string, clear bool) {
	if s.cfg.InvoiceRepo == nil || invoiceID == "" {
		return
	}
	inv, err := s.cfg.InvoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return
	}
	if inv.Data == nil {
		inv.Data = make(data.JSONMap)
	}
	if clear {
		delete(inv.Data, InvoiceDataNextSettleAt)
	} else if nextRFC != "" {
		inv.Data[InvoiceDataNextSettleAt] = nextRFC
		inv.Data[InvoiceDataTrustageSettleWF] = TrustageSettleWorkflowName(invoiceID)
	}
	if _, uErr := s.cfg.InvoiceRepo.Update(ctx, inv); uErr != nil {
		util.Log(ctx).WithError(uErr).WithField("invoice_id", invoiceID).
			Warn("settlement: could not persist nextSettleAt")
	}
}

func buildSettleDSL(cfg SettlementSchedulerConfig, invoiceID, name, cronExpr string) (*structpb.Struct, error) {
	url := fmt.Sprintf("%s/_internal/billing/invoices/%s/settle", cfg.BillingBaseURL, invoiceID)
	m := map[string]any{
		"version":     "1.0",
		"name":        name,
		"description": fmt.Sprintf("Invoice %s checkout settlement reminder", invoiceID),
		"timeout":     "2m",
		"on_error":    map[string]any{"action": "abort"},
		"steps": []any{
			map[string]any{
				"id":      "settle_invoice",
				"type":    "call",
				"name":    "Process invoice settlement",
				"timeout": "90s",
				"retry": map[string]any{
					"max_attempts":     float64(2),
					"backoff_strategy": "exponential",
					"initial_backoff":  "10s",
				},
				"call": map[string]any{
					"action": "http.request",
					"input": map[string]any{
						"url":    url,
						"method": "POST",
						"headers": map[string]any{
							"Content-Type":  "application/json",
							"X-Admin-Token": cfg.AdminTokenPlaceholder,
						},
						"body": map[string]any{},
					},
					"output_var": "settle_result",
				},
			},
		},
		"schedules": []any{
			map[string]any{
				"name":      name + ".oneshot",
				"cron_expr": cronExpr,
				"timezone":  "UTC",
			},
		},
	}
	return structpb.NewStruct(m)
}
