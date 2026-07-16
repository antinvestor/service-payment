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

// DefaultSettlementRetryDelaysMinutes: first check ~2m after checkout open,
// then 5m, 15m, 30m, 1h, 2h — spread out so abandoned pays are recovered
// without hammering checkout.
var DefaultSettlementRetryDelaysMinutes = []int{2, 5, 15, 30, 60, 120}

// SettlementConfig controls per-invoice Trustage settle reminders.
type SettlementConfig struct {
	// RetryDelays from checkout open / last plan: attempt i fires at
	// firstScheduledBase + RetryDelays[i] when using absolute schedule,
	// or from now when re-arming after a not-yet-complete poll.
	RetryDelays []time.Duration
	// MaxAttempts caps settle polls; after that the Trustage reminder is archived.
	MaxAttempts int
}

// ParseRetryDelaysMinutesCSV parses "2,5,15,30" into durations.
func ParseRetryDelaysMinutesCSV(csv string) []time.Duration {
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
		m, err := strconv.Atoi(p)
		if err != nil || m < 0 {
			continue
		}
		out = append(out, time.Duration(m)*time.Minute)
	}
	return out
}

// NewSettlementConfigFromEnv builds settlement dunning from env-style fields.
func NewSettlementConfigFromEnv(maxAttempts int, retryCSV string) SettlementConfig {
	cfg := SettlementConfig{MaxAttempts: maxAttempts}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 6
	}
	delays := ParseRetryDelaysMinutesCSV(retryCSV)
	if len(delays) == 0 {
		for _, m := range DefaultSettlementRetryDelaysMinutes {
			delays = append(delays, time.Duration(m)*time.Minute)
		}
	}
	cfg.RetryDelays = delays
	if len(cfg.RetryDelays) < cfg.MaxAttempts {
		last := cfg.RetryDelays[len(cfg.RetryDelays)-1]
		for len(cfg.RetryDelays) < cfg.MaxAttempts {
			// Stretch by +30m steps when max exceeds CSV.
			last += 30 * time.Minute
			cfg.RetryDelays = append(cfg.RetryDelays, last)
		}
	}
	return cfg
}

// NextSettleAt returns when attemptIndex (0-based) should fire.
// base is typically checkout-open time (now when first scheduled).
func (c SettlementConfig) NextSettleAt(base time.Time, attemptIndex int) (time.Time, bool) {
	if attemptIndex < 0 || attemptIndex >= c.MaxAttempts || attemptIndex >= len(c.RetryDelays) {
		return time.Time{}, false
	}
	return base.UTC().Add(c.RetryDelays[attemptIndex]), true
}

// RescheduleAfterPoll is the next fire after a "not completed" poll at `now`
// for the upcoming attempt index (already incremented).
// Uses the delay gap between attemptIndex-1 and attemptIndex, floored at 1m.
func (c SettlementConfig) RescheduleAfterPoll(now time.Time, nextAttemptIndex int) (time.Time, bool) {
	if nextAttemptIndex < 0 || nextAttemptIndex >= c.MaxAttempts || nextAttemptIndex >= len(c.RetryDelays) {
		return time.Time{}, false
	}
	var gap time.Duration
	if nextAttemptIndex == 0 {
		gap = c.RetryDelays[0]
	} else {
		gap = c.RetryDelays[nextAttemptIndex] - c.RetryDelays[nextAttemptIndex-1]
	}
	if gap < time.Minute {
		gap = time.Minute
	}
	return now.UTC().Add(gap), true
}
