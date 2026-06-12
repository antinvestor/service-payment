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

package integrationobs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antinvestor/service-payments/pkg/integrationobs"
)

func TestNewMetrics_DoesNotPanic(t *testing.T) {
	m := integrationobs.NewMetrics("testprovider")
	if m == nil {
		t.Fatal("expected non-nil Metrics")
	}
}

func TestObserveProviderCall_RoundTrip(t *testing.T) {
	m := integrationobs.NewMetrics("testprovider")
	ctx := context.Background()

	// success path
	ctx2, done := m.ObserveProviderCall(ctx, "test_op")
	if ctx2 == nil {
		t.Fatal("expected non-nil context")
	}
	done(nil) // must not panic

	// error path
	_, done2 := m.ObserveProviderCall(ctx, "test_op_err")
	done2(errors.New("some error")) // must not panic
}

func TestNilMetrics_DoesNotPanic(t *testing.T) {
	var m *integrationobs.Metrics
	ctx := context.Background()
	ctx2, done := m.ObserveProviderCall(ctx, "op")
	if ctx2 == nil {
		t.Fatal("expected context from nil Metrics")
	}
	done(nil)
	m.QueueProcessed(ctx, "payment")
	m.QueueFailed(ctx, "payment", "provider_error")
	m.WebhookReceived(ctx, "deposit")
	m.WebhookRejected(ctx, "deposit", "decode_error")
}
