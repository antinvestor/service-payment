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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRetryDelaysHoursCSV(t *testing.T) {
	t.Parallel()
	delays := business.ParseRetryDelaysHoursCSV("0, 24,72, 168")
	require.Len(t, delays, 4)
	assert.Equal(t, 0*time.Hour, delays[0])
	assert.Equal(t, 24*time.Hour, delays[1])
	assert.Equal(t, 72*time.Hour, delays[2])
	assert.Equal(t, 168*time.Hour, delays[3])

	assert.Nil(t, business.ParseRetryDelaysHoursCSV(""))
	assert.Nil(t, business.ParseRetryDelaysHoursCSV("  "))
	// Invalid tokens skipped
	d2 := business.ParseRetryDelaysHoursCSV("0,abc,-1,48")
	require.Len(t, d2, 2)
	assert.Equal(t, 0*time.Hour, d2[0])
	assert.Equal(t, 48*time.Hour, d2[1])
}

func TestNewRenewalConfigFromEnv_Defaults(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(0, 0, "", "")
	assert.Equal(t, 24*time.Hour, cfg.LeadTime)
	assert.Equal(t, 4, cfg.MaxAttempts)
	assert.Equal(t, "flutterwave", cfg.DefaultRoute)
	require.Len(t, cfg.RetryDelays, 4)
	assert.Equal(t, 0*time.Hour, cfg.RetryDelays[0])
	assert.Equal(t, 24*time.Hour, cfg.RetryDelays[1])
	assert.Equal(t, 72*time.Hour, cfg.RetryDelays[2])
	assert.Equal(t, 168*time.Hour, cfg.RetryDelays[3])
}

func TestNewRenewalConfigFromEnv_ExtendsRetries(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(12, 6, "0,24", "flutterwave")
	assert.Equal(t, 12*time.Hour, cfg.LeadTime)
	assert.Equal(t, 6, cfg.MaxAttempts)
	require.GreaterOrEqual(t, len(cfg.RetryDelays), 6)
	// Extended with +24h steps from last known delay.
	assert.Equal(t, 48*time.Hour, cfg.RetryDelays[2])
	assert.Equal(t, 72*time.Hour, cfg.RetryDelays[3])
}

func TestRenewalConfig_AttemptDue_SpreadSchedule(t *testing.T) {
	t.Parallel()
	cfg := business.NewRenewalConfigFromEnv(24, 4, "0,24,72,168", "flutterwave")
	periodEnd := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	// Lead: first attempt due at periodEnd - 24h = July 9 12:00
	firstDue := periodEnd.Add(-24 * time.Hour)

	// Before lead window
	assert.False(t, cfg.AttemptDue(firstDue.Add(-time.Hour), periodEnd, 0, nil))
	// Exactly at first due
	assert.True(t, cfg.AttemptDue(firstDue, periodEnd, 0, nil))
	// Attempt 1 needs +24h after first due and min gap from last attempt
	last0 := firstDue
	assert.False(t, cfg.AttemptDue(firstDue.Add(12*time.Hour), periodEnd, 1, &last0))
	assert.True(t, cfg.AttemptDue(firstDue.Add(24*time.Hour), periodEnd, 1, &last0))
	// Attempt 2 at +72h from first due
	last1 := firstDue.Add(24 * time.Hour)
	assert.False(t, cfg.AttemptDue(firstDue.Add(48*time.Hour), periodEnd, 2, &last1))
	assert.True(t, cfg.AttemptDue(firstDue.Add(72*time.Hour), periodEnd, 2, &last1))
	// Beyond max attempts
	assert.False(t, cfg.AttemptDue(firstDue.Add(200*time.Hour), periodEnd, 4, nil))
}
