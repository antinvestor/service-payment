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
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/pitabwire/frame/v3/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CollectionProviderSuite struct {
	billingTests.BaseTestSuite
}

func TestCollectionProviderSuite(t *testing.T) {
	suite.Run(t, new(CollectionProviderSuite))
}

type recordingCommands struct {
	published []collection.ProviderCommand
}

func (r *recordingCommands) Publish(_ context.Context, cmd collection.ProviderCommand) error {
	r.published = append(r.published, cmd)
	return nil
}

func (r *recordingCommands) Configured(string) bool { return true }

func newProviderCollectionBiz(
	resources *billingTests.ServiceResources,
	ch *checkoutTestHarness,
	commands business.ProviderCommandPublisher,
) business.CollectionBusiness {
	integ := business.NewCheckoutIntegration(ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "")
	return business.NewCollectionBusiness(
		integ, resources.InvoiceEngine, resources.InvoiceRepo, resources.SubscriptionBusiness,
		resources.PlanRepo, resources.ComponentRepo, resources.BillingRunRepo, resources.PricingEngine,
		nil, business.CollectionLedgerAccounts{},
		business.CollectionOptions{
			ProviderSubscriptions: resources.PolarSubscriptionBusiness,
			ProviderCommands:      commands,
		},
	)
}

// makeProviderPlan is seedFlatPlan plus the Polar product on the plan.
func makeProviderPlan(t *testing.T, ctx context.Context, resources *billingTests.ServiceResources) (string, string) {
	t.Helper()
	cvID, planID := seedFlatPlan(t, ctx, resources, "15.00")
	plan, err := resources.PlanRepo.GetByID(ctx, planID)
	require.NoError(t, err)
	plan.Data = map[string]any{models.PlanDataPaymentProvider: "polar", models.PlanDataProviderProductID: "prod_pro"}
	_, err = resources.PlanRepo.Update(ctx, plan)
	require.NoError(t, err)
	return cvID, planID
}

func (ts *CollectionProviderSuite) TestStartSubscription_ProviderPlan_OpensProviderCheckout() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)
		cvID, planID := makeProviderPlan(t, ctx, resources)
		commands := &recordingCommands{}
		coll := newProviderCollectionBiz(resources, ch, commands)

		result, err := coll.StartSubscription(ctx, business.StartSubscriptionInput{
			ProfileID: profileID, PlanID: planID, CatalogVersionID: cvID, Currency: "KES",
			ReturnURL: "https://app.example/done", ExternalEntityID: "ws-1", ExternalEntityType: "workspace",
		})
		require.NoError(t, err)
		assert.Empty(t, result.InvoiceID, "the provider invoices; billing raises none")
		assert.NotEmpty(t, result.SessionRef)
		assert.Contains(t, result.PageURL, "/c/")

		sub, err := resources.SubscriptionBusiness.GetSubscription(ctx, result.SubscriptionID)
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStatePending, sub.State)
		assert.Equal(t, models.CollectionModeProvider, sub.Data[models.SubDataCollectionMode])
		assert.Equal(t, "polar", sub.Data[models.SubDataPaymentProvider])
		assert.Equal(t, "prod_pro", sub.Data[models.SubDataExternalProductID])
		assert.Equal(t, "ws-1", sub.Data[models.SubDataExternalEntityID])
		assert.Equal(t, result.SessionRef, sub.Data[models.SubDataCheckoutSessionRef])

		sess := ch.sessionRepo.sessions[result.SessionRef]
		require.NotNil(t, sess)
		assert.Equal(t, "prod_pro", sess.Metadata[collection.ExtraProductID])
		assert.Equal(t, sub.GetID(), sess.Metadata["subscriptionId"])
		assert.Equal(t, models.CollectionModeProvider, sess.Metadata[models.SubDataCollectionMode])
		keys, _ := sess.Methods["keys"].([]any)
		assert.Equal(t, []any{"polar"}, keys, "only the provider's rail is offered")

		// Opening it again before paying reuses the pending subscription.
		again, err := coll.StartSubscription(ctx, business.StartSubscriptionInput{
			ProfileID: profileID, PlanID: planID, CatalogVersionID: cvID, Currency: "KES",
		})
		require.NoError(t, err)
		assert.Equal(t, result.SubscriptionID, again.SubscriptionID)

		// The payer finished on the provider page: confirm activates.
		sess.Status = checkoutModels.SessionStatusCompleted
		confirmed, err := coll.ConfirmPayment(ctx, result.SessionRef)
		require.NoError(t, err)
		assert.True(t, confirmed.Paid)
		assert.Equal(t, models.SubscriptionStateActive, confirmed.SubscriptionState)
		assert.Empty(t, confirmed.InvoiceID)

		// Cancelling asks the provider to stop charging once its id is known.
		_, err = resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			SubscriptionID: sub.GetID(), PolarSubscriptionID: "psub_x", PolarStatus: collection.SubscriptionStatusActive,
			CurrentPeriodEnd: rfc3339In(24 * 30 * 3600e9),
		})
		require.NoError(t, err)
		cancelled, err := coll.CancelSubscription(ctx, sub.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateActive, cancelled.SubscriptionState, "soft cancel keeps access until period end")
		require.Len(t, commands.published, 1)
		assert.Equal(t, collection.CommandCancelSubscription, commands.published[0].Command)
		assert.Equal(t, "psub_x", commands.published[0].ExternalSubscriptionID)
		assert.True(t, commands.published[0].AtPeriodEnd)
	})
}

func (ts *CollectionProviderSuite) TestConfirmPayment_ProviderSession_NotCompleted() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)
		cvID, planID := makeProviderPlan(t, ctx, resources)
		coll := newProviderCollectionBiz(resources, ch, &recordingCommands{})

		result, err := coll.StartSubscription(ctx, business.StartSubscriptionInput{
			ProfileID: profileID, PlanID: planID, CatalogVersionID: cvID, Currency: "KES",
		})
		require.NoError(t, err)
		_, err = coll.ConfirmPayment(ctx, result.SessionRef)
		require.ErrorIs(t, err, business.ErrCheckoutNotCompleted)
	})
}

func (ts *CollectionProviderSuite) TestConfirmPayment_RefusesUnboundOrForeignSessions() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ch := newCheckoutTestHarness(t)
		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)
		cvID, planID := makeProviderPlan(t, ctx, resources)
		coll := newProviderCollectionBiz(resources, ch, &recordingCommands{})

		result, err := coll.StartSubscription(ctx, business.StartSubscriptionInput{
			ProfileID: profileID, PlanID: planID, CatalogVersionID: cvID, Currency: "KES",
		})
		require.NoError(t, err)

		// A completed session crafted by another caller that merely names the
		// subscription in its metadata must not activate it.
		forged := &checkoutModels.CheckoutSession{
			Ref: "forged-session-ref-000000000000000", Name: "x", Amount: "1.00", Currency: "KES",
			AmountOption: checkoutModels.AmountOptionFixed, Status: checkoutModels.SessionStatusCompleted,
			OrderRef: result.SubscriptionID,
			Metadata: map[string]any{"subscriptionId": result.SubscriptionID, models.SubDataCollectionMode: models.CollectionModeProvider},
		}
		ch.sessionRepo.sessions[forged.Ref] = forged
		_, err = coll.ConfirmPayment(ctx, forged.Ref)
		require.ErrorIs(t, err, business.ErrCheckoutSessionNotBound)
		sub, err := resources.SubscriptionBusiness.GetSubscription(ctx, result.SubscriptionID)
		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStatePending, sub.State)

		// The real session, but paid on another rail: refused as well.
		real := ch.sessionRepo.sessions[result.SessionRef]
		real.Status = checkoutModels.SessionStatusCompleted
		real.Metadata["_route"] = "mpesa"
		_, err = coll.ConfirmPayment(ctx, result.SessionRef)
		require.ErrorIs(t, err, business.ErrCheckoutSessionNotBound)

		// A caller cannot widen the rail onto M-PESA for a Polar plan.
		_, err = coll.StartSubscription(ctx, business.StartSubscriptionInput{
			ProfileID: util.IDString(), PlanID: planID, CatalogVersionID: cvID, Currency: "KES", Methods: []string{"mpesa"},
		})
		require.Error(t, err)
	})
}
