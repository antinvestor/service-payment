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
	"strconv"
	"strings"
	"time"
)

// DefaultRenewalRetryDelaysHours is day-0, +1d, +3d, +7d after period due.
var DefaultRenewalRetryDelaysHours = []int{0, 24, 72, 168}

// RenewalConfig controls automatic subscription rebill timing and retries.
// Renewals are always single-subscription (Trustage one-shot per sub).
type RenewalConfig struct {
	// LeadTime starts collection this long before period end.
	LeadTime time.Duration
	// RetryDelays hours after the first due moment for each attempt (index 0 = first try).
	RetryDelays []time.Duration
	// MaxAttempts including the first try. Exhaustion archives the Trustage reminder.
	MaxAttempts int
	// DefaultRoute payment route for COF (must be flutterwave for v4 token charges).
	DefaultRoute string
}

// ParseRetryDelaysHoursCSV parses "0,24,72,168" into durations.
func ParseRetryDelaysHoursCSV(csv string) []time.Duration {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]time.Duration, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		h, err := strconv.Atoi(p)
		if err != nil || h < 0 {
			continue
		}
		out = append(out, time.Duration(h)*time.Hour)
	}
	return out
}

// NewRenewalConfigFromEnv builds config from billing env-style fields.
func NewRenewalConfigFromEnv(
	leadHours, maxAttempts int,
	retryCSV, defaultRoute string,
) RenewalConfig {
	cfg := RenewalConfig{
		LeadTime:     time.Duration(leadHours) * time.Hour,
		MaxAttempts:  maxAttempts,
		DefaultRoute: strings.TrimSpace(defaultRoute),
	}
	if cfg.LeadTime <= 0 {
		cfg.LeadTime = 24 * time.Hour
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 4
	}
	if cfg.DefaultRoute == "" {
		cfg.DefaultRoute = "flutterwave"
	}
	delays := ParseRetryDelaysHoursCSV(retryCSV)
	if len(delays) == 0 {
		for _, h := range DefaultRenewalRetryDelaysHours {
			delays = append(delays, time.Duration(h)*time.Hour)
		}
	}
	cfg.RetryDelays = delays
	if len(cfg.RetryDelays) < cfg.MaxAttempts {
		// Extend with last delay step (+24h) if max attempts exceed CSV length.
		last := cfg.RetryDelays[len(cfg.RetryDelays)-1]
		for len(cfg.RetryDelays) < cfg.MaxAttempts {
			last += 24 * time.Hour
			cfg.RetryDelays = append(cfg.RetryDelays, last)
		}
	}
	return cfg
}

// FirstDueAt is when attempt 0 may run (periodEnd - leadTime).
func (c RenewalConfig) FirstDueAt(periodEnd time.Time) time.Time {
	firstDue := periodEnd.Add(-c.LeadTime)
	if firstDue.After(periodEnd) {
		return periodEnd
	}
	return firstDue
}

// AttemptDueAt is the absolute time when attemptIndex may first run.
func (c RenewalConfig) AttemptDueAt(periodEnd time.Time, attemptIndex int) (time.Time, bool) {
	if attemptIndex < 0 || attemptIndex >= c.MaxAttempts || attemptIndex >= len(c.RetryDelays) {
		return time.Time{}, false
	}
	return c.FirstDueAt(periodEnd).Add(c.RetryDelays[attemptIndex]), true
}

// AttemptDue reports whether attempt index (0-based) may run at now,
// given periodEnd and optional lastAttemptAt.
func (c RenewalConfig) AttemptDue(now, periodEnd time.Time, attemptIndex int, lastAttemptAt *time.Time) bool {
	dueAt, ok := c.AttemptDueAt(periodEnd, attemptIndex)
	if !ok {
		return false
	}
	if now.Before(dueAt) {
		return false
	}
	// Avoid thrashing: require spacing from last attempt for retries > 0.
	if attemptIndex > 0 && lastAttemptAt != nil {
		minGap := c.RetryDelays[attemptIndex] - c.RetryDelays[attemptIndex-1]
		if minGap < time.Hour {
			minGap = time.Hour
		}
		if now.Before(lastAttemptAt.Add(minGap)) {
			return false
		}
	}
	return true
}
