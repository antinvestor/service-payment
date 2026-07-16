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
	workflowv1 "github.com/antinvestor/service-trustage/gen/go/workflow/v1"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

// WorkflowClient is the Trustage surface needed for per-subscription reminders.
// Satisfied by workflowv1connect.WorkflowServiceClient.
type WorkflowClient interface {
	ListWorkflows(context.Context, *connect.Request[workflowv1.ListWorkflowsRequest]) (*connect.Response[workflowv1.ListWorkflowsResponse], error)
	CreateWorkflow(context.Context, *connect.Request[workflowv1.CreateWorkflowRequest]) (*connect.Response[workflowv1.CreateWorkflowResponse], error)
	ActivateWorkflow(context.Context, *connect.Request[workflowv1.ActivateWorkflowRequest]) (*connect.Response[workflowv1.ActivateWorkflowResponse], error)
	ArchiveWorkflow(context.Context, *connect.Request[workflowv1.ArchiveWorkflowRequest]) (*connect.Response[workflowv1.ArchiveWorkflowResponse], error)
}

// RenewalScheduler keeps one Trustage one-shot reminder per subscription in
// sync with subscription properties (period end, attempt count, cancel flag).
// Failures reschedule to the next dunning delay; cancel / max attempts archive.
type RenewalScheduler interface {
	// SyncReminder plans from sub properties and ensure/cancels the Trustage workflow.
	SyncReminder(ctx context.Context, sub *models.Subscription) error
	// CancelReminder archives the Trustage workflow for the subscription.
	CancelReminder(ctx context.Context, subscriptionID string) error
}

// NoopRenewalScheduler is used when Trustage is not configured.
type NoopRenewalScheduler struct{}

func (NoopRenewalScheduler) SyncReminder(context.Context, *models.Subscription) error { return nil }
func (NoopRenewalScheduler) CancelReminder(context.Context, string) error             { return nil }

// TrustageSchedulerConfig configures HTTP callbacks from Trustage into billing.
type TrustageSchedulerConfig struct {
	// BillingBaseURL is the in-cluster base for billing (no trailing slash),
	// e.g. http://service-payment-billing.finance.svc:80
	BillingBaseURL string
	// AdminTokenEnv is the Trustage secret placeholder for X-Admin-Token
	// (literal "${BILLING_INTERNAL_ADMIN_TOKEN}" so Trustage expands at fire time).
	AdminTokenPlaceholder string
	// Renewal config for planning next fire.
	Renewal RenewalConfig
	// Optional: patch nextRenewAt onto subscription after ensure.
	SubBiz SubscriptionBusiness
}

type trustageRenewalScheduler struct {
	client WorkflowClient
	cfg    TrustageSchedulerConfig
}

// NewTrustageRenewalScheduler builds a scheduler. client may be nil → noop behaviour.
func NewTrustageRenewalScheduler(client WorkflowClient, cfg TrustageSchedulerConfig) RenewalScheduler {
	if client == nil || strings.TrimSpace(cfg.BillingBaseURL) == "" {
		return NoopRenewalScheduler{}
	}
	if cfg.AdminTokenPlaceholder == "" {
		cfg.AdminTokenPlaceholder = "${BILLING_INTERNAL_ADMIN_TOKEN}"
	}
	cfg.BillingBaseURL = strings.TrimRight(cfg.BillingBaseURL, "/")
	return &trustageRenewalScheduler{client: client, cfg: cfg}
}

func (s *trustageRenewalScheduler) SyncReminder(ctx context.Context, sub *models.Subscription) error {
	if sub == nil {
		return nil
	}
	plan := PlanRenewalReminder(sub, s.cfg.Renewal, time.Now().UTC())
	switch plan.Action {
	case ReminderCancel:
		if err := s.CancelReminder(ctx, sub.GetID()); err != nil {
			return err
		}
		s.patchNext(ctx, sub.GetID(), "", true)
		return nil
	case ReminderEnsure:
		if err := s.ensure(ctx, sub.GetID(), plan.NextAt, plan.Kind); err != nil {
			return err
		}
		s.patchNext(ctx, sub.GetID(), plan.NextAt.Format(time.RFC3339), false)
		util.Log(ctx).
			WithField("subscription_id", sub.GetID()).
			WithField("next_at", plan.NextAt.Format(time.RFC3339)).
			WithField("reason", plan.Reason).
			WithField("kind", plan.Kind).
			Info("renewal: trustage reminder synced")
		return nil
	default:
		return nil
	}
}

func (s *trustageRenewalScheduler) CancelReminder(ctx context.Context, subscriptionID string) error {
	if subscriptionID == "" {
		return nil
	}
	name := TrustageRenewWorkflowName(subscriptionID)
	if err := archiveWorkflowByName(ctx, s.client, name); err != nil {
		return err
	}
	util.Log(ctx).WithField("subscription_id", subscriptionID).
		WithField("workflow", name).Info("renewal: trustage reminder cancelled")
	return nil
}

func (s *trustageRenewalScheduler) patchNext(ctx context.Context, subID, nextRFC string, clear bool) {
	if s.cfg.SubBiz == nil || subID == "" {
		return
	}
	patch := map[string]any{}
	if clear {
		patch[models.SubDataNextRenewAt] = ""
	} else if nextRFC != "" {
		patch[models.SubDataNextRenewAt] = nextRFC
		patch[models.SubDataTrustageRenewWorkflow] = TrustageRenewWorkflowName(subID)
	}
	if len(patch) == 0 {
		return
	}
	if _, err := s.cfg.SubBiz.PatchSubscriptionData(ctx, subID, patch); err != nil {
		util.Log(ctx).WithError(err).WithField("subscription_id", subID).
			Warn("renewal: could not persist nextRenewAt")
	}
}

func (s *trustageRenewalScheduler) ensure(ctx context.Context, subscriptionID string, nextAt time.Time, kind string) error {
	name := TrustageRenewWorkflowName(subscriptionID)
	desiredCron := CronForOneShot(nextAt)

	id, active, cron, err := findWorkflow(ctx, s.client, name)
	if err != nil {
		return err
	}
	if active && cron == desiredCron {
		return nil
	}
	if id != "" {
		// Cadence/time changed or inactive wrong cron: archive then recreate.
		if aerr := archiveWorkflowByName(ctx, s.client, name); aerr != nil {
			return aerr
		}
	}

	dsl, err := buildRenewalDSL(s.cfg, subscriptionID, name, desiredCron, kind)
	if err != nil {
		return err
	}
	createResp, err := s.client.CreateWorkflow(ctx, connect.NewRequest(&workflowv1.CreateWorkflowRequest{Dsl: dsl}))
	if err != nil {
		if !isAlreadyExistsErr(err) {
			return fmt.Errorf("renewal schedule create %s: %w", name, err)
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

func buildRenewalDSL(
	cfg TrustageSchedulerConfig,
	subscriptionID, name, cronExpr, kind string,
) (*structpb.Struct, error) {
	url := fmt.Sprintf("%s/_internal/billing/subscriptions/%s/renew", cfg.BillingBaseURL, subscriptionID)
	desc := fmt.Sprintf("Subscription %s renewal reminder (%s)", subscriptionID, kind)
	m := map[string]any{
		"version":     "1.0",
		"name":        name,
		"description": desc,
		"timeout":     "3m",
		"on_error":    map[string]any{"action": "abort"},
		"steps": []any{
			map[string]any{
				"id":      "renew_subscription",
				"type":    "call",
				"name":    "Process subscription renew/finalize",
				"timeout": "2m",
				"retry": map[string]any{
					"max_attempts":     float64(2),
					"backoff_strategy": "exponential",
					"initial_backoff":  "15s",
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
						"body": map[string]any{"kind": kind},
					},
					"output_var": "renew_result",
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

func findWorkflow(ctx context.Context, client WorkflowClient, name string) (id string, active bool, cron string, err error) {
	resp, err := client.ListWorkflows(ctx, connect.NewRequest(&workflowv1.ListWorkflowsRequest{Name: name}))
	if err != nil {
		return "", false, "", fmt.Errorf("renewal schedule list %s: %w", name, err)
	}
	for _, it := range resp.Msg.GetItems() {
		if it.GetName() != name {
			continue
		}
		switch it.GetStatus() {
		case workflowv1.WorkflowStatus_WORKFLOW_STATUS_ACTIVE:
			return it.GetId(), true, cronFromDSL(it.GetDsl()), nil
		case workflowv1.WorkflowStatus_WORKFLOW_STATUS_ARCHIVED:
			continue
		default:
			id, cron = it.GetId(), cronFromDSL(it.GetDsl())
		}
	}
	return id, false, cron, nil
}

func cronFromDSL(dsl *structpb.Struct) string {
	if dsl == nil {
		return ""
	}
	sched := dsl.GetFields()["schedules"].GetListValue()
	if sched == nil || len(sched.GetValues()) == 0 {
		return ""
	}
	first := sched.GetValues()[0].GetStructValue()
	if first == nil {
		return ""
	}
	return first.GetFields()["cron_expr"].GetStringValue()
}

func activateWorkflow(ctx context.Context, client WorkflowClient, name, id string) error {
	if id == "" {
		return nil
	}
	if _, err := client.ActivateWorkflow(ctx, connect.NewRequest(&workflowv1.ActivateWorkflowRequest{Id: id})); err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil
		}
		return fmt.Errorf("renewal schedule activate %s: %w", name, err)
	}
	return nil
}

func archiveWorkflowByName(ctx context.Context, client WorkflowClient, name string) error {
	listResp, err := client.ListWorkflows(ctx, connect.NewRequest(&workflowv1.ListWorkflowsRequest{Name: name}))
	if err != nil {
		return fmt.Errorf("renewal schedule list %s: %w", name, err)
	}
	for _, it := range listResp.Msg.GetItems() {
		if it.GetName() != name || it.GetStatus() == workflowv1.WorkflowStatus_WORKFLOW_STATUS_ARCHIVED {
			continue
		}
		if _, err := client.ArchiveWorkflow(ctx, connect.NewRequest(&workflowv1.ArchiveWorkflowRequest{
			Id: it.GetId(),
		})); err != nil {
			return fmt.Errorf("renewal schedule archive %s: %w", name, err)
		}
	}
	return nil
}

func isAlreadyExistsErr(err error) bool {
	if err == nil {
		return false
	}
	if connect.CodeOf(err) == connect.CodeAlreadyExists {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "23505")
}
