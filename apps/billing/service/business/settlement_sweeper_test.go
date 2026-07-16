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

	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	checkoutModels "github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (ts *CollectionSuite) TestSettlementSweeper_SettlesCompletedSession() {
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

		sweeper := business.NewSettlementSweeper(resources.InvoiceRepo, coll, 50)
		result, err := sweeper.Sweep(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.GreaterOrEqual(t, result.Candidates, 1)
		assert.GreaterOrEqual(t, result.Settled, 1)

		paid, err := resources.InvoiceEngine.GetInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.InvoiceStatePaid, paid.State)

		// Second sweep is a no-op (invoice no longer ISSUED).
		again, err := sweeper.Sweep(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, again.Settled)
	})
}

func (ts *CollectionSuite) TestSettlementSweeper_SkipsPendingSession() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "22.00")
		coll := newCollectionBiz(resources, ch)

		opened, err := coll.CollectPayment(ctx, business.CollectPaymentInput{InvoiceID: invoice.GetID()})
		require.NoError(t, err)

		// Leave session pending.
		_ = opened
		sweeper := business.NewSettlementSweeper(resources.InvoiceRepo, coll, 50)
		result, err := sweeper.Sweep(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Skipped, 1)

		still, err := resources.InvoiceEngine.GetInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.InvoiceStateIssued, still.State)
	})
}

func TestSettlementSweeper_NilCollection(t *testing.T) {
	sweeper := business.NewSettlementSweeper(nil, nil, 10)
	_, err := sweeper.Sweep(context.Background())
	require.Error(t, err)
}

func TestPostCashIdempotencyMarker(t *testing.T) {
	// Smoke: invoice Data can hold ledgerPaymentTxnId without panicking.
	inv := &models.Invoice{
		Data: data.JSONMap{"ledgerPaymentTxnId": "billing_pay_x"},
	}
	_, ok := inv.Data["ledgerPaymentTxnId"].(string)
	assert.True(t, ok)
}
