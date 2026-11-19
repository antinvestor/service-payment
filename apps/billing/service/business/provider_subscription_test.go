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
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	billingTests "github.com/antinvestor/service-payments/apps/billing/tests"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/pitabwire/frame/v3/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProviderSubscriptionSuite struct {
	billingTests.BaseTestSuite
}

func TestProviderSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(ProviderSubscriptionSuite))
}

func rfc3339In(d time.Duration) string {
	return time.Now().Add(d).UTC().Format(time.RFC3339)
}

func (ts *ProviderSubscriptionSuite) TestMirror_ByBillingSubscriptionID() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)
		_, start := startPendingPolarSub(ctx, t, resources, profileID)

		// Webhook metadata carries billing's subscription id: no profile or plan needed.
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			Provider:            "polar",
			SubscriptionID:      start.Subscription.GetID(),
			PolarSubscriptionID: "psub_1",
			ExternalCustomerID:  "cust_1",
			PolarStatus:         collection.SubscriptionStatusActive,
			CurrentPeriodEnd:    rfc3339In(30 * 24 * time.Hour),
			OccurredAt:          rfc3339In(0),
		})
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateActive, sub.State)
		assert.Equal(t, "psub_1", sub.Data[models.SubDataExternalSubscriptionID])
		assert.Equal(t, "psub_1", sub.Data[business.PolarDataKeySubscriptionID], "legacy key kept for older readers")
		assert.Equal(t, "cust_1", sub.Data[models.SubDataProviderCustomerID])
		assert.Equal(t, "polar", sub.Data[models.SubDataPaymentProvider])

		// From now on the provider id alone finds the row.
		again, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			Provider:            "polar",
			PolarSubscriptionID: "psub_1",
			PolarStatus:         collection.SubscriptionStatusActive,
			CancelAtPeriodEnd:   true,
			CurrentPeriodEnd:    rfc3339In(30 * 24 * time.Hour),
			OccurredAt:          rfc3339In(time.Second),
		})
		require.NoError(t, err)
		assert.Equal(t, sub.GetID(), again.GetID())
		assert.Equal(t, models.SubscriptionStateActive, again.State, "cancel at period end keeps access")
		assert.Equal(t, true, again.Data[models.SubDataCancelAtPeriodEnd])

		// The payer changes their mind: the flag clears.
		back, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			Provider:            "polar",
			PolarSubscriptionID: "psub_1",
			PolarStatus:         collection.SubscriptionStatusActive,
			OccurredAt:          rfc3339In(2 * time.Second),
		})
		require.NoError(t, err)
		_, flagged := back.Data[models.SubDataCancelAtPeriodEnd]
		assert.False(t, flagged)
	})
}

func (ts *ProviderSubscriptionSuite) TestMirror_StaleEventsAreIgnored() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)
		_, start := startPendingPolarSub(ctx, t, resources, profileID)

		_, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			SubscriptionID: start.Subscription.GetID(), PolarSubscriptionID: "psub_2",
			PolarStatus: collection.SubscriptionStatusRevoked, OccurredAt: rfc3339In(0),
		})
		require.NoError(t, err)

		// An "active" that happened before the revoke arrives late.
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: "psub_2",
			PolarStatus:         collection.SubscriptionStatusActive, OccurredAt: rfc3339In(-time.Minute),
		})
		require.ErrorIs(t, err, business.ErrStaleProviderEvent)
		require.NotNil(t, sub)
		assert.Equal(t, models.SubscriptionStateExpired, sub.State, "a stale activation never revives a revoked subscription")
	})
}

func (ts *ProviderSubscriptionSuite) TestMirror_BilledAndPastDue() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)
		_, start := startPendingPolarSub(ctx, t, resources, profileID)

		periodEnd := rfc3339In(30 * 24 * time.Hour)
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			Event: collection.SubscriptionEventBilled, SubscriptionID: start.Subscription.GetID(),
			PolarSubscriptionID: "psub_3", ExternalOrderID: "ord_1", CurrentPeriodEnd: periodEnd,
		})
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateActive, sub.State, "a paid order activates a pending subscription")
		assert.Equal(t, "ord_1", sub.Data[models.SubDataLastProviderOrderID])
		require.NotNil(t, sub.EndAt)
		assert.Equal(t, periodEnd, sub.EndAt.UTC().Format(time.RFC3339))

		// The same order again changes nothing.
		again, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			Event: collection.SubscriptionEventBilled, PolarSubscriptionID: "psub_3", ExternalOrderID: "ord_1",
		})
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateActive, again.State)

		// Past due keeps access (the provider retries), records the state.
		pd, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: "psub_3", PolarStatus: collection.SubscriptionStatusPastDue, OccurredAt: rfc3339In(0),
		})
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateActive, pd.State)
		assert.Equal(t, collection.SubscriptionStatusPastDue, pd.Data[models.SubDataProviderStatus])

		// Unpaid ends it.
		gone, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: "psub_3", PolarStatus: collection.SubscriptionStatusUnpaid, OccurredAt: rfc3339In(time.Second),
		})
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateExpired, gone.State)
		assert.NotNil(t, gone.CancelledAt)
	})
}

func (ts *ProviderSubscriptionSuite) TestStart_ReusesPendingAndRefusesNonProviderPlan() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		plan, first := startPendingPolarSub(ctx, t, resources, profileID)
		second, err := resources.PolarSubscriptionBusiness.StartPolarSubscription(ctx, plan, business.StartPolarInput{
			ProfileID: profileID, PlanID: plan.GetID(), CatalogVersionID: plan.CatalogVersionID, Currency: "KES",
		})
		require.NoError(t, err)
		assert.Equal(t, first.Subscription.GetID(), second.Subscription.GetID(), "an abandoned checkout reuses its pending row")
		assert.Equal(t, "polar", second.Provider)

		_, plain := makePolarPlan(ctx, t, resources, "")
		_, err = resources.PolarSubscriptionBusiness.StartPolarSubscription(ctx, plain, business.StartPolarInput{
			ProfileID: profileID, PlanID: plain.GetID(), CatalogVersionID: plain.CatalogVersionID, Currency: "KES",
		})
		require.ErrorIs(t, err, business.ErrPlanNotPolarCollected)
	})
}

func TestProviderProductFor(t *testing.T) {
	t.Parallel()
	p, id := business.ProviderProductFor(&models.Plan{Data: map[string]any{"polarProductId": "prod_1"}})
	assert.Equal(t, "polar", p)
	assert.Equal(t, "prod_1", id)
	p, id = business.ProviderProductFor(&models.Plan{Data: map[string]any{
		models.PlanDataPaymentProvider: "Stripe", models.PlanDataProviderProductID: "price_1",
	}})
	assert.Equal(t, "stripe", p)
	assert.Equal(t, "price_1", id)
	p, _ = business.ProviderProductFor(&models.Plan{})
	assert.Empty(t, p)
	assert.True(t, business.IsProviderManaged(&models.Subscription{Data: map[string]any{"collectionMode": "polar"}}))
	assert.True(t, business.IsProviderManaged(&models.Subscription{Data: map[string]any{"collectionMode": "provider"}}))
	assert.False(t, business.IsProviderManaged(&models.Subscription{}))
}

func (ts *ProviderSubscriptionSuite) TestMirror_RefusesEventsNamingOtherSubscriptions() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		// A subscription billing collects itself: webhook metadata must not claim it.
		own, err := resources.SubscriptionBusiness.CreateSubscription(ctx, &models.Subscription{
			ProfileID: profileID, PlanID: util.IDString(), CatalogVersionID: util.IDString(),
			Currency: "KES", State: models.SubscriptionStateActive,
		})
		require.NoError(t, err)
		_, err = resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			SubscriptionID: own.GetID(), PolarSubscriptionID: "psub_evil", PolarStatus: collection.SubscriptionStatusRevoked,
		})
		require.ErrorIs(t, err, business.ErrProviderSubscriptionMismatch)
		fresh, err := resources.SubscriptionBusiness.GetSubscription(ctx, own.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateActive, fresh.State)

		// A provider subscription already pinned to psub_a cannot be re-pointed.
		_, start := startPendingPolarSub(ctx, t, resources, profileID)
		_, err = resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			SubscriptionID: start.Subscription.GetID(), PolarSubscriptionID: "psub_a",
			PolarStatus: collection.SubscriptionStatusActive, OccurredAt: rfc3339In(0),
		})
		require.NoError(t, err)
		_, err = resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			SubscriptionID: start.Subscription.GetID(), PolarSubscriptionID: "psub_b",
			PolarStatus: collection.SubscriptionStatusRevoked, OccurredAt: rfc3339In(time.Second),
		})
		require.ErrorIs(t, err, business.ErrProviderSubscriptionMismatch)

		// Another product's event is refused too.
		_, err = resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: "psub_a", ExternalProductID: "prod_other",
			PolarStatus: collection.SubscriptionStatusRevoked, OccurredAt: rfc3339In(time.Second),
		})
		require.ErrorIs(t, err, business.ErrProviderSubscriptionMismatch)
	})
}
