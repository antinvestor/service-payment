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

package collection

import (
	"context"
	"errors"
	"net"
	"syscall"
)

// uncertainError marks a provider call whose outcome is unknown: the request
// may or may not have been applied, so the worker must not report FAILED.
type uncertainError struct{ err error }

func (u *uncertainError) Error() string { return "outcome unknown: " + u.err.Error() }
func (u *uncertainError) Unwrap() error { return u.err }

// Uncertain wraps err as an outcome-unknown provider error. Integration
// clients use it for timeouts, connection failures and 5xx answers so the
// queue worker redelivers (or looks the record up) instead of failing it.
func Uncertain(err error) error {
	if err == nil {
		return nil
	}
	return &uncertainError{err: err}
}

// IsUncertain reports whether a provider error leaves the outcome unknown:
// an explicit Uncertain wrap, a cancelled or expired context, or a network
// error (timeout, reset, refused). Rejections from the provider (4xx with a
// reason) are not uncertain: the provider refused and nothing was applied.
func IsUncertain(err error) bool {
	if err == nil {
		return false
	}
	var u *uncertainError
	if errors.As(err, &u) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.EPIPE) {
		return true
	}
	return false
}
