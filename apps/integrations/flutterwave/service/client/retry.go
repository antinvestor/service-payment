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

package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// RetryPolicy bounds transient provider failures so we still "win" collection
// without hammering Flutterwave on permanent client errors.
type RetryPolicy struct {
	// MaxAttempts includes the first try. Default 3.
	MaxAttempts int
	// BaseDelay is the first backoff. Default 200ms.
	BaseDelay time.Duration
	// MaxDelay caps exponential backoff. Default 2s.
	MaxDelay time.Duration
}

func (p RetryPolicy) normalised() RetryPolicy {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 3
	}
	if p.BaseDelay <= 0 {
		p.BaseDelay = 200 * time.Millisecond
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = 2 * time.Second
	}
	return p
}

// withRetry runs fn with exponential backoff on retryable errors.
func withRetry(ctx context.Context, policy RetryPolicy, fn func() error) error {
	policy = policy.normalised()
	var last error
	delay := policy.BaseDelay
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = fn()
		if last == nil {
			return nil
		}
		if !isRetryable(last) || attempt == policy.MaxAttempts {
			return last
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		delay *= 2
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
	}
	return last
}

// isRetryable reports transient transport / rate-limit / server errors.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	// HTTP status phrases we surface from doJSON.
	for _, needle := range []string{
		"http 408", "http 425", "http 429",
		"http 500", "http 502", "http 503", "http 504",
		"timeout", "temporary", "connection reset", "eof",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	// Permanent client encryption / validation must not retry.
	if strings.Contains(msg, "client_encryption") ||
		strings.Contains(msg, "validation") ||
		strings.Contains(msg, "http 400") ||
		strings.Contains(msg, "http 401") ||
		strings.Contains(msg, "http 403") ||
		strings.Contains(msg, "http 404") ||
		strings.Contains(msg, "http 422") {
		return false
	}
	return false
}

// FormatRetryError wraps the last error with attempt context for logs.
func FormatRetryError(attempts int, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("after %d attempts: %w", attempts, err)
}
