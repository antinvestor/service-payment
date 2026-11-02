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
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrInvalidRequest marks a request refused before it was sent: nothing
// reached Flutterwave, so the outcome is certain.
var ErrInvalidRequest = errors.New("invalid flutterwave request")

// ErrCredentialsRefused means Flutterwave's token service refused the
// configured keys. It is an outage, not a refused payment: work is retried
// until an operator fixes the keys, never failed (a charge made before the
// keys were revoked may exist).
var ErrCredentialsRefused = errors.New("flutterwave refused the configured credentials")

// ErrNotFound is returned by lookups that find nothing.
var ErrNotFound = errors.New("flutterwave: not found")

// HTTPError is a non-2xx answer from Flutterwave.
type HTTPError struct {
	Status int
	Body   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("flutterwave http %d: %s", e.Status, e.Body)
}

// APIError is a 2xx answer whose envelope reports failure: Flutterwave
// answered and refused.
type APIError struct {
	Op      string
	Status  string
	Message string
	Type    string
	Code    string
}

func (e *APIError) Error() string {
	if e.Type != "" || e.Code != "" {
		return fmt.Sprintf("flutterwave %s: %s (%s/%s)", e.Op, e.Message, e.Type, e.Code)
	}
	return fmt.Sprintf("flutterwave %s: status=%s message=%s", e.Op, e.Status, e.Message)
}

func invalidRequest(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
}

// IsRejected reports a definite refusal: Flutterwave answered that it will
// not do what was asked (or the request never left). Anything else — a
// timeout, a dropped connection, a 5xx, a conflict, a rate limit, an
// unreadable answer — leaves the outcome unknown: the charge or transfer may
// exist, so it must be looked up, never assumed failed.
func IsRejected(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrInvalidRequest) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return true
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.Status {
		case http.StatusRequestTimeout, http.StatusConflict, http.StatusTooEarly, http.StatusTooManyRequests:
			return false
		}
		return httpErr.Status >= 400 && httpErr.Status < 500 && !IsDuplicate(err)
	}
	return false
}

// IsNotFound reports a 404 or an empty lookup.
func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var httpErr *HTTPError
	return errors.As(err, &httpErr) && httpErr.Status == http.StatusNotFound
}

// IsDuplicate reports Flutterwave refusing something it already holds: the
// earlier request reached it, so the charge or transfer exists. The match
// is deliberately broad — mistaking a duplicate for a refusal would report
// a payout failed that was paid, and invite paying it again.
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) && httpErr.Status == http.StatusConflict {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "already")
}

// IsRefusal reports a definite 4xx or envelope refusal whatever its wording.
// It suits writes that move no money (a transfer recipient), where a
// refusal is safe to report even when it mentions an existing object.
func IsRefusal(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrInvalidRequest) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return true
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.Status {
		case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
			return false
		}
		return httpErr.Status >= 400 && httpErr.Status < 500
	}
	return false
}
