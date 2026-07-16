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
	"context"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	checkoutModels "github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (ts *CollectionSuite) TestSettlementProcessor_SettlesCompletedSession() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "88.00")
		coll := newCollectionBiz(resources, ch)

		opened, err := coll.CollectPayment(ctx, business.CollectPaymentInput{InvoiceID: invoice.GetID()})
		require.NoError(t, err)
		require.NotEmpty(t, opened.SessionRef)

		// Complete checkout without ConfirmPayment (abandoned browser).
		stored := ch.sessionRepo.sessions[opened.SessionRef]
		require.NotNil(t, stored)
		stored.Status = checkoutModels.SessionStatusCompleted

		proc := business.NewSettlementProcessor(
			resources.InvoiceRepo, coll, business.NewSettlementConfigFromEnv(6, "2,5,15"),
		)
		result, err := proc.ProcessInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "settled", result.Action)
		assert.True(t, result.Paid)

		paid, err := resources.InvoiceEngine.GetInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.InvoiceStatePaid, paid.State)

		// Second process is a no-op (already paid).
		again, err := proc.ProcessInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, "skipped", again.Action)
		assert.True(t, again.Paid)
	})
}

func (ts *CollectionSuite) TestSettlementProcessor_PendingRearms() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "22.00")
		coll := newCollectionBiz(resources, ch)

		opened, err := coll.CollectPayment(ctx, business.CollectPaymentInput{InvoiceID: invoice.GetID()})
		require.NoError(t, err)
		require.NotEmpty(t, opened.SessionRef)

		// Leave session pending — ProcessInvoice should re-arm, not pay.
		proc := business.NewSettlementProcessor(
			resources.InvoiceRepo, coll, business.NewSettlementConfigFromEnv(6, "2,5,15"),
		)
		result, err := proc.ProcessInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, "pending", result.Action)
		assert.False(t, result.Paid)
		assert.NotEmpty(t, result.NextSettleAt)

		still, err := resources.InvoiceEngine.GetInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.InvoiceStateIssued, still.State)
	})
}

func TestSettlementProcessor_NilCollection(t *testing.T) {
	t.Parallel()
	proc := business.NewSettlementProcessor(nil, nil, business.NewSettlementConfigFromEnv(0, ""))
	_, err := proc.ProcessInvoice(context.Background(), "inv_1")
	require.Error(t, err)
}

func TestSettlementConfig_RescheduleSpread(t *testing.T) {
	t.Parallel()
	cfg := business.NewSettlementConfigFromEnv(6, "2,5,15,30")
	require.Len(t, cfg.RetryDelays, 6) // extended to max
	base := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	n0, ok := cfg.NextSettleAt(base, 0)
	require.True(t, ok)
	assert.Equal(t, base.Add(2*time.Minute), n0)
	n1, ok := cfg.NextSettleAt(base, 1)
	require.True(t, ok)
	assert.Equal(t, base.Add(5*time.Minute), n1)

	now := base.Add(10 * time.Minute)
	next, ok := cfg.RescheduleAfterPoll(now, 2) // gap 15-5=10m
	require.True(t, ok)
	assert.Equal(t, now.Add(10*time.Minute), next)
}

func TestTrustageSettleWorkflowName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "billing.invoice.settle.inv_9", business.TrustageSettleWorkflowName("inv_9"))
}

func TestPostCashIdempotencyMarker(t *testing.T) {
	// Smoke: invoice Data can hold ledgerPaymentTxnId without panicking.
	inv := &models.Invoice{
		Data: data.JSONMap{"ledgerPaymentTxnId": "billing_pay_x"},
	}
	_, ok := inv.Data["ledgerPaymentTxnId"].(string)
	assert.True(t, ok)
}
