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

package observability_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antinvestor/service-payments/apps/ledger/service/observability"
)

// TestNewMetrics verifies that NewMetrics constructs without panic and that
// StartSpan/EndSpan round-trips cleanly with and without an error.
func TestNewMetrics(t *testing.T) {
	m := observability.NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics() returned nil")
	}

	ctx := context.Background()

	// Round-trip with no error.
	ctx2, span := m.StartSpan(ctx, "test-span-ok")
	if ctx2 == nil {
		t.Error("StartSpan returned nil context")
	}
	if span == nil {
		t.Error("StartSpan returned nil span")
	}
	m.EndSpan(ctx2, span, nil) // must not panic

	// Round-trip with an error.
	ctx3, span2 := m.StartSpan(ctx, "test-span-err")
	m.EndSpan(ctx3, span2, errors.New("test error")) // must not panic
}

// TestRecordHelpers smoke-tests the recording helpers do not panic.
func TestRecordHelpers(_ *testing.T) {
	m := observability.NewMetrics()
	ctx := context.Background()

	m.RecordTransactionPosted(ctx, "USD", 2, false, 42.0)
	m.RecordTransactionPosted(ctx, "KES", 4, true, 10.5)
	m.RecordTransactionFailed(ctx, "validation")
	m.RecordTransactionFailed(ctx, "conflict")
	m.RecordTransactionFailed(ctx, "system")
	m.RecordAccountCreated(ctx)
	m.RecordLedgerCreated(ctx)
	m.RecordBookCreated(ctx)
}
