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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// APIError is a Fincra error response. Fincra answers errors in several
// shapes: {"success":false,"error":"...","errorType":"..."},
// {"status":false,"message":"..."} and {"message":"..."}.
type APIError struct {
	HTTPStatus int
	ErrorType  string
	Message    string
}

func (e *APIError) Error() string {
	code := e.ErrorType
	if code == "" {
		code = http.StatusText(e.HTTPStatus)
	}
	return fmt.Sprintf("fincra %d %s: %s", e.HTTPStatus, code, e.Message)
}

// Code is the error type, or the HTTP status text when Fincra gave none.
func (e *APIError) Code() string {
	if e.ErrorType != "" {
		return e.ErrorType
	}
	if e.HTTPStatus != 0 {
		return strings.ToUpper(strings.ReplaceAll(http.StatusText(e.HTTPStatus), " ", "_"))
	}
	return "PROVIDER_ERROR"
}

// parseAPIError builds an APIError from an error response body.
func parseAPIError(status int, body []byte) *APIError {
	apiErr := &APIError{HTTPStatus: status}
	var raw struct {
		Error     json.RawMessage `json:"error"`
		ErrorType string          `json:"errorType"`
		Message   json.RawMessage `json:"message"`
	}
	if json.Unmarshal(body, &raw) != nil {
		apiErr.Message = strings.TrimSpace(string(body))
		return apiErr
	}
	apiErr.ErrorType = raw.ErrorType
	apiErr.Message = firstText(raw.Error, raw.Message)
	if apiErr.Message == "" {
		apiErr.Message = strings.TrimSpace(string(body))
	}
	return apiErr
}

// firstText returns the first field that is a non-empty string, or the
// first element of a string array (validation messages).
func firstText(fields ...json.RawMessage) string {
	for _, f := range fields {
		if len(f) == 0 {
			continue
		}
		var s string
		if json.Unmarshal(f, &s) == nil && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
		var list []string
		if json.Unmarshal(f, &list) == nil && len(list) > 0 {
			return strings.Join(list, "; ")
		}
	}
	return ""
}

// IsNotFound reports whether Fincra has no record for the lookup.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.HTTPStatus == http.StatusNotFound || apiErr.ErrorType == ErrTypeNotFound
}

// IsRejected reports whether Fincra definitively refused a request: an API
// error with a 4xx status other than timeout and rate limiting. Any other
// failure of a money-moving call (transport error, 5xx, 408, 425, 429)
// leaves its outcome unknown — Fincra may have accepted it.
func IsRejected(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.HTTPStatus {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
		return false
	}
	return apiErr.HTTPStatus >= http.StatusBadRequest && apiErr.HTTPStatus < http.StatusInternalServerError
}

// IsDuplicate reports whether Fincra refused a reference it has seen before.
func IsDuplicate(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.ErrorType == ErrTypeDuplicateReference || apiErr.HTTPStatus == http.StatusConflict {
		return true
	}
	msg := strings.ToLower(apiErr.Message)
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "already exist")
}
