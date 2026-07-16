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
	"fmt"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
)

// RenewalReminderAction is what SyncReminder should do for a subscription.
type RenewalReminderAction string

const (
	// ReminderEnsure creates/updates the Trustage one-shot at NextAt.
	ReminderEnsure RenewalReminderAction = "ensure"
	// ReminderCancel archives the Trustage workflow (cancel / max dunning).
	ReminderCancel RenewalReminderAction = "cancel"
	// ReminderNone leaves Trustage alone (nil scheduler / nothing to do).
	ReminderNone RenewalReminderAction = "none"
)

// RenewalPlan is computed solely from subscription properties + dunning config.
type RenewalPlan struct {
	Action RenewalReminderAction
	// NextAt is when Trustage should fire (ensure only).
	NextAt time.Time
	// Reason is for logs / operators.
	Reason string
	// Kind distinguishes rebill vs period-end hard-cancel.
	Kind string // "renew" | "finalize_cancel"
}

// PlanRenewalReminder decides the next Trustage reminder from sub.Data / state.
//
// Rules:
//   - not ACTIVE → cancel
//   - cancelAtPeriodEnd → ensure at period end (finalize); if already past → cancel
//     (caller ProcessSubscription finalizes first)
//   - renewAttemptCount >= maxAttempts → cancel (past_due, no more auto-tries)
//   - else ensure at AttemptDueAt(periodEnd, attemptCount)
//   - if that time is in the past, schedule ASAP (now + 1m) so Trustage catches up
func PlanRenewalReminder(sub *models.Subscription, cfg RenewalConfig, now time.Time) RenewalPlan {
	if sub == nil {
		return RenewalPlan{Action: ReminderNone, Reason: "nil_subscription"}
	}
	now = now.UTC()
	if sub.State != models.SubscriptionStateActive {
		return RenewalPlan{Action: ReminderCancel, Reason: "not_active:" + sub.State}
	}

	periodEnd := currentPeriodEnd(sub)
	if periodEnd.IsZero() {
		periodEnd = sub.BillingAnchor.UTC().AddDate(0, 1, 0)
		if periodEnd.IsZero() || periodEnd.Before(sub.StartAt) {
			periodEnd = sub.StartAt.UTC().AddDate(0, 1, 0)
		}
	}

	// Soft-cancel: no rebill — only a finalize reminder at period end.
	if cancelAtPeriodEnd(sub.Data) {
		if !periodEnd.After(now) {
			// Period already over; ProcessSubscription should hard-cancel.
			// No further Trustage reminder after finalize.
			return RenewalPlan{Action: ReminderCancel, Reason: "cancel_at_period_end_due", Kind: "finalize_cancel"}
		}
		return RenewalPlan{
			Action: ReminderEnsure,
			NextAt: periodEnd.UTC(),
			Reason: "finalize_cancel_at_period_end",
			Kind:   "finalize_cancel",
		}
	}

	attempts := renewAttemptCount(sub.Data)
	if attempts >= cfg.MaxAttempts {
		return RenewalPlan{Action: ReminderCancel, Reason: fmt.Sprintf("max_attempts:%d", attempts)}
	}

	dueAt, ok := cfg.AttemptDueAt(periodEnd, attempts)
	if !ok {
		return RenewalPlan{Action: ReminderCancel, Reason: "no_attempt_slot"}
	}
	// Missed fire / catch-up: don't leave forever in the past.
	if dueAt.Before(now) {
		dueAt = now.Add(time.Minute)
	}
	return RenewalPlan{
		Action: ReminderEnsure,
		NextAt: dueAt.UTC(),
		Reason: fmt.Sprintf("renew_attempt_%d", attempts),
		Kind:   "renew",
	}
}

// CronForOneShot encodes a UTC instant as a yearly cron "M H D M *" so Trustage
// fires once on that calendar minute. After fire, billing re-plans and replaces
// the workflow with the next one-shot (or archives it).
func CronForOneShot(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("%d %d %d %d *", t.Minute(), t.Hour(), t.Day(), int(t.Month()))
}

// TrustageRenewWorkflowName is the deterministic Trustage workflow name.
func TrustageRenewWorkflowName(subscriptionID string) string {
	return "billing.subscription.renew." + subscriptionID
}
