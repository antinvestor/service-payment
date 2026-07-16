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

package business_test

import (
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/pitabwire/frame/v2/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronForOneShot(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 8, 15, 14, 30, 0, 0, time.UTC)
	assert.Equal(t, "30 14 15 8 *", business.CronForOneShot(ts))
}

func TestTrustageRenewWorkflowName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "billing.subscription.renew.sub_1", business.TrustageRenewWorkflowName("sub_1"))
}

func TestPlanRenewalReminder_EnsureNextPeriod(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(24, 4, "0,24,72,168", "flutterwave")
	periodEnd := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	sub := &models.Subscription{
		State: models.SubscriptionStateActive,
		Data: data.JSONMap{
			models.SubDataCurrentPeriodEnd:  periodEnd.Format(time.RFC3339),
			models.SubDataRenewAttemptCount: 0,
		},
	}
	plan := business.PlanRenewalReminder(sub, cfg, now)
	require.Equal(t, business.ReminderEnsure, plan.Action)
	// First due = periodEnd - 24h
	assert.Equal(t, periodEnd.Add(-24*time.Hour), plan.NextAt)
	assert.Equal(t, "renew", plan.Kind)
}

func TestPlanRenewalReminder_RetrySpread(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(24, 4, "0,24,72,168", "flutterwave")
	periodEnd := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	firstDue := periodEnd.Add(-24 * time.Hour) // Aug 19 12:00
	// Attempt index 1 is due at firstDue+24h = Aug 20 12:00. Plan while still before that.
	now := firstDue.Add(12 * time.Hour) // Aug 20 00:00 — after attempt 0, before attempt 1
	sub := &models.Subscription{
		State: models.SubscriptionStateActive,
		Data: data.JSONMap{
			models.SubDataCurrentPeriodEnd:  periodEnd.Format(time.RFC3339),
			models.SubDataRenewAttemptCount: 1, // next is attempt index 1 → +24h
		},
	}
	plan := business.PlanRenewalReminder(sub, cfg, now)
	require.Equal(t, business.ReminderEnsure, plan.Action)
	assert.Equal(t, firstDue.Add(24*time.Hour), plan.NextAt)
}

func TestPlanRenewalReminder_MaxAttemptsCancels(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(24, 4, "0,24,72,168", "flutterwave")
	sub := &models.Subscription{
		State: models.SubscriptionStateActive,
		Data: data.JSONMap{
			models.SubDataCurrentPeriodEnd:  time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
			models.SubDataRenewAttemptCount: 4,
		},
	}
	plan := business.PlanRenewalReminder(sub, cfg, time.Now().UTC())
	assert.Equal(t, business.ReminderCancel, plan.Action)
	assert.Contains(t, plan.Reason, "max_attempts")
}

func TestPlanRenewalReminder_SoftCancel(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(24, 4, "0,24,72,168", "flutterwave")
	pe := time.Now().UTC().Add(10 * 24 * time.Hour)
	sub := &models.Subscription{
		State: models.SubscriptionStateActive,
		Data: data.JSONMap{
			models.SubDataCancelAtPeriodEnd: true,
			models.SubDataCurrentPeriodEnd:  pe.Format(time.RFC3339),
		},
	}
	plan := business.PlanRenewalReminder(sub, cfg, time.Now().UTC())
	require.Equal(t, business.ReminderEnsure, plan.Action)
	assert.Equal(t, "finalize_cancel", plan.Kind)
	assert.Equal(t, pe.UTC().Truncate(time.Second), plan.NextAt.Truncate(time.Second))
}

func TestPlanRenewalReminder_NotActive(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(24, 4, "0,24,72,168", "flutterwave")
	sub := &models.Subscription{State: models.SubscriptionStateCancelled}
	plan := business.PlanRenewalReminder(sub, cfg, time.Now().UTC())
	assert.Equal(t, business.ReminderCancel, plan.Action)
}
