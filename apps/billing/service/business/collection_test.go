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
	billingTests "github.com/antinvestor/service-payments/apps/billing/tests"
	checkoutModels "github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CollectionSuite struct {
	billingTests.BaseTestSuite
}

func TestCollectionSuite(t *testing.T) {
	suite.Run(t, new(CollectionSuite))
}

func newCollectionBiz(
	resources *billingTests.ServiceResources,
	ch *checkoutTestHarness,
) business.CollectionBusiness {
	integ := business.NewCheckoutIntegration(
		ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "",
	)
	return business.NewCollectionBusiness(
		integ,
		resources.InvoiceEngine,
		resources.InvoiceRepo,
		resources.SubscriptionBusiness,
		resources.PlanRepo,
		resources.ComponentRepo,
		resources.BillingRunRepo,
		resources.PricingEngine,
		nil,
		business.CollectionLedgerAccounts{},
	)
}

func (ts *CollectionSuite) TestCollectPayment_OpensCheckout() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "120.00")
		coll := newCollectionBiz(resources, ch)

		result, err := coll.CollectPayment(ctx, business.CollectPaymentInput{InvoiceID: invoice.GetID()})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.AlreadyComplete)
		assert.NotEmpty(t, result.SessionRef)
		assert.Contains(t, result.PageURL, "/c/")
		assert.Equal(t, invoice.GetID(), result.InvoiceID)

		stored, err := resources.InvoiceRepo.GetByID(ctx, invoice.GetID())
		require.NoError(t, err)
		require.NotNil(t, stored.Data)
		assert.Equal(t, result.SessionRef, stored.Data[business.InvoiceDataCheckoutSessionRef])
	})
}

func (ts *CollectionSuite) TestCollectPayment_AlreadyPaid_Idempotent() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "40.00")
		_, err := resources.InvoiceEngine.RecordPayment(ctx, invoice.GetID())
		require.NoError(t, err)

		coll := newCollectionBiz(resources, ch)
		result, err := coll.CollectPayment(ctx, business.CollectPaymentInput{InvoiceID: invoice.GetID()})
		require.NoError(t, err)
		assert.True(t, result.AlreadyComplete)
		assert.Empty(t, result.PageURL)
		assert.Empty(t, result.SessionRef)
	})
}

func (ts *CollectionSuite) TestConfirmPayment_SettlesAndActivatesSubscription() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		sub, err := resources.SubscriptionBusiness.CreateSubscription(ctx, &models.Subscription{
			ProfileID:        profileID,
			PlanID:           util.IDString(),
			CatalogVersionID: util.IDString(),
			Currency:         "KES",
			State:            models.SubscriptionStatePending,
		})
		require.NoError(t, err)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "99.00")
		invoice.SubscriptionID = sub.GetID()
		_, err = resources.InvoiceRepo.Update(ctx, invoice)
		require.NoError(t, err)

		coll := newCollectionBiz(resources, ch)
		opened, err := coll.CollectPayment(ctx, business.CollectPaymentInput{InvoiceID: invoice.GetID()})
		require.NoError(t, err)

		stored := ch.sessionRepo.sessions[opened.SessionRef]
		require.NotNil(t, stored)
		stored.Status = checkoutModels.SessionStatusCompleted

		confirmed, err := coll.ConfirmPayment(ctx, opened.SessionRef)
		require.NoError(t, err)
		assert.True(t, confirmed.Paid)
		assert.Equal(t, models.InvoiceStatePaid, confirmed.InvoiceState)
		assert.Equal(t, sub.GetID(), confirmed.SubscriptionID)
		assert.Equal(t, models.SubscriptionStateActive, confirmed.SubscriptionState)

		again, err := coll.ConfirmPayment(ctx, opened.SessionRef)
		require.NoError(t, err)
		assert.True(t, again.Paid)
	})
}

func (ts *CollectionSuite) TestStartSubscription_WithFlatFee_OpensCheckout() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		catalogVersionID, planID := seedFlatPlan(t, ctx, resources, "500.00")
		coll := newCollectionBiz(resources, ch)

		result, err := coll.StartSubscription(ctx, business.StartSubscriptionInput{
			ProfileID:        profileID,
			PlanID:           planID,
			CatalogVersionID: catalogVersionID,
			Currency:         "KES",
			ReturnURL:        "https://app.example/done",
			PayerDisplayName: "Ada",
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.AlreadyComplete)
		assert.NotEmpty(t, result.SessionRef)
		assert.NotEmpty(t, result.InvoiceID)
		assert.NotEmpty(t, result.SubscriptionID)
		assert.Contains(t, result.PageURL, "/c/")

		sub, err := resources.SubscriptionBusiness.GetSubscription(ctx, result.SubscriptionID)
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStatePending, sub.State)

		sess := ch.sessionRepo.sessions[result.SessionRef]
		require.NotNil(t, sess)
		sess.Status = checkoutModels.SessionStatusCompleted

		confirmed, err := coll.ConfirmPayment(ctx, result.SessionRef)
		require.NoError(t, err)
		assert.True(t, confirmed.Paid)
		assert.Equal(t, models.SubscriptionStateActive, confirmed.SubscriptionState)
	})
}

func (ts *CollectionSuite) TestStartSubscription_NoUpfrontFee_ActivatesImmediately() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		catalogVersionID, planID := seedUsageOnlyPlan(t, ctx, resources)
		coll := newCollectionBiz(resources, ch)

		result, err := coll.StartSubscription(ctx, business.StartSubscriptionInput{
			ProfileID:        profileID,
			PlanID:           planID,
			CatalogVersionID: catalogVersionID,
			Currency:         "KES",
		})
		require.NoError(t, err)
		assert.True(t, result.AlreadyComplete)
		assert.Empty(t, result.PageURL)
		assert.Empty(t, result.SessionRef)
		assert.NotEmpty(t, result.SubscriptionID)

		sub, err := resources.SubscriptionBusiness.GetSubscription(ctx, result.SubscriptionID)
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateActive, sub.State)
	})
}

// seedFlatPlan creates a catalog version + plan with one FLAT component/tier.
func seedFlatPlan(
	t *testing.T,
	ctx context.Context,
	resources *billingTests.ServiceResources,
	fee string,
) (catalogVersionID, planID string) {
	t.Helper()

	cv, err := resources.CatalogBusiness.CreateCatalogVersion(ctx, &models.CatalogVersion{
		CatalogID: util.IDString(),
		Name:      "test-catalog",
		Currency:  "KES",
	})
	require.NoError(t, err)

	plan, err := resources.CatalogBusiness.CreatePlan(ctx, &models.Plan{
		CatalogVersionID: cv.GetID(),
		Name:             "Pro",
	})
	require.NoError(t, err)

	comp, err := resources.CatalogBusiness.CreateComponent(ctx, &models.Component{
		PlanID:          plan.GetID(),
		Name:            "Base",
		MetricKey:       "base",
		PricingModel:    models.PricingModelFlat,
		AggregationType: models.AggregationTypeSum,
	})
	require.NoError(t, err)

	feeDec, err := decimalx.NewFromString(fee)
	require.NoError(t, err)
	_, err = resources.CatalogBusiness.CreateTier(ctx, &models.Tier{
		ComponentID: comp.GetID(),
		SortOrder:   0,
		LowerBound:  decimalx.Zero().Ptr(),
		FlatFee:     feeDec.Ptr(),
		UnitPrice:   decimalx.Zero().Ptr(),
	})
	require.NoError(t, err)

	return cv.GetID(), plan.GetID()
}

// seedUsageOnlyPlan creates a plan with only PER_UNIT pricing (no upfront fee).
func seedUsageOnlyPlan(
	t *testing.T,
	ctx context.Context,
	resources *billingTests.ServiceResources,
) (catalogVersionID, planID string) {
	t.Helper()

	cv, err := resources.CatalogBusiness.CreateCatalogVersion(ctx, &models.CatalogVersion{
		CatalogID: util.IDString(),
		Name:      "usage-catalog",
		Currency:  "KES",
	})
	require.NoError(t, err)

	plan, err := resources.CatalogBusiness.CreatePlan(ctx, &models.Plan{
		CatalogVersionID: cv.GetID(),
		Name:             "Pay as you go",
	})
	require.NoError(t, err)

	comp, err := resources.CatalogBusiness.CreateComponent(ctx, &models.Component{
		PlanID:          plan.GetID(),
		Name:            "API calls",
		MetricKey:       "api_calls",
		PricingModel:    models.PricingModelPerUnit,
		AggregationType: models.AggregationTypeSum,
		UnitName:        "call",
	})
	require.NoError(t, err)

	_, err = resources.CatalogBusiness.CreateTier(ctx, &models.Tier{
		ComponentID: comp.GetID(),
		SortOrder:   0,
		LowerBound:  decimalx.Zero().Ptr(),
		UnitPrice:   decimalx.NewFromInt64(1).Ptr(),
	})
	require.NoError(t, err)

	return cv.GetID(), plan.GetID()
}
