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
	billingTests "github.com/antinvestor/service-payments/apps/billing/tests"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// Unit-level tests for pure helpers (no DB needed)
// ---------------------------------------------------------------------------

func TestPolarProductID(t *testing.T) {
	tests := []struct {
		name    string
		plan    *models.Plan
		wantID  string
		wantPol bool
	}{
		{
			name:    "nil plan",
			plan:    nil,
			wantID:  "",
			wantPol: false,
		},
		{
			name:    "plan with nil Data",
			plan:    &models.Plan{},
			wantID:  "",
			wantPol: false,
		},
		{
			name: "plan with empty Data",
			plan: &models.Plan{
				Data: data.JSONMap{},
			},
			wantID:  "",
			wantPol: false,
		},
		{
			name: "plan with non-string polarProductId",
			plan: &models.Plan{
				Data: data.JSONMap{"polarProductId": 12345},
			},
			wantID:  "",
			wantPol: false,
		},
		{
			name: "plan with empty string polarProductId",
			plan: &models.Plan{
				Data: data.JSONMap{"polarProductId": ""},
			},
			wantID:  "",
			wantPol: false,
		},
		{
			name: "plan with valid polarProductId",
			plan: &models.Plan{
				Data: data.JSONMap{"polarProductId": "prod_abc123"},
			},
			wantID:  "prod_abc123",
			wantPol: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID := business.PolarProductID(tt.plan)
			assert.Equal(t, tt.wantID, gotID, "PolarProductID")

			gotPol := business.IsPolarCollected(tt.plan)
			assert.Equal(t, tt.wantPol, gotPol, "IsPolarCollected")
		})
	}
}

// ---------------------------------------------------------------------------
// Integration test suite (real PG via BaseTestSuite)
// ---------------------------------------------------------------------------

type PolarSubscriptionSuite struct {
	billingTests.BaseTestSuite
}

func TestPolarSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(PolarSubscriptionSuite))
}

// makePolarPlan creates a minimal catalog version and a plan within the given context (which
// must carry auth claims so RLS policies allow reads back).
// If polarProductID is non-empty the plan's Data map carries "polarProductId".
func makePolarPlan(
	ctx context.Context,
	t *testing.T,
	resources *billingTests.ServiceResources,
	polarProductID string,
) (*models.CatalogVersion, *models.Plan) {
	t.Helper()

	cv := &models.CatalogVersion{
		CatalogID: util.IDString(),
		Version:   1,
		Name:      "test catalog",
		Currency:  "KES",
	}
	cv.GenID(ctx)
	created, err := resources.CatalogBusiness.CreateCatalogVersion(ctx, cv)
	require.NoError(t, err)

	planData := data.JSONMap{}
	if polarProductID != "" {
		planData[business.PolarDataKeyProductID] = polarProductID
	}

	plan := &models.Plan{
		CatalogVersionID: created.GetID(),
		Name:             "test plan",
		Data:             planData,
	}
	plan.GenID(ctx)
	createdPlan, err := resources.CatalogBusiness.CreatePlan(ctx, plan)
	require.NoError(t, err)

	return created, createdPlan
}

// startPendingPolarSub creates a PENDING polar subscription via StartPolarSubscription.
// ctx must carry auth claims so that the subscription is written with the correct tenant/partition.
func startPendingPolarSub(
	ctx context.Context,
	t *testing.T,
	resources *billingTests.ServiceResources,
	profileID string,
) (*models.Plan, *business.PolarSubscriptionStart) {
	t.Helper()

	_, plan := makePolarPlan(ctx, t, resources, "prod_mirror_test")
	in := business.StartPolarInput{
		ProfileID:        profileID,
		PlanID:           plan.GetID(),
		CatalogVersionID: plan.CatalogVersionID,
		Currency:         "KES",
	}
	result, err := resources.PolarSubscriptionBusiness.StartPolarSubscription(ctx, plan, in)
	require.NoError(t, err)
	return plan, result
}

// ---------------------------------------------------------------------------
// StartPolarSubscription tests
// ---------------------------------------------------------------------------

func (ts *PolarSubscriptionSuite) TestStartPolarSubscription_PolarPlan_CreatesPendingSubscription() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		_, plan := makePolarPlan(ctx, t, resources, "prod_polar_abc")

		in := business.StartPolarInput{
			ProfileID:        profileID,
			PlanID:           plan.GetID(),
			CatalogVersionID: plan.CatalogVersionID,
			Currency:         "KES",
		}
		result, err := resources.PolarSubscriptionBusiness.StartPolarSubscription(ctx, plan, in)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Subscription)
		assert.Equal(t, "prod_polar_abc", result.PolarProductID)
		assert.Equal(t, models.SubscriptionStatePending, result.Subscription.State)
		assert.Equal(t, profileID, result.Subscription.ProfileID)
		assert.Equal(t, plan.GetID(), result.Subscription.PlanID)

		sub := result.Subscription
		assert.Equal(t, "polar", sub.Data[business.PolarDataKeyCollectionMode])
		assert.Equal(t, "prod_polar_abc", sub.Data[business.PolarDataKeyProductID])
	})
}

func (ts *PolarSubscriptionSuite) TestStartPolarSubscription_NonPolarPlan_ReturnsError() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		_, plan := makePolarPlan(ctx, t, resources, "") // no polar product ID

		in := business.StartPolarInput{
			ProfileID:        profileID,
			PlanID:           plan.GetID(),
			CatalogVersionID: plan.CatalogVersionID,
			Currency:         "KES",
		}
		result, err := resources.PolarSubscriptionBusiness.StartPolarSubscription(ctx, plan, in)

		require.Error(t, err)
		require.ErrorIs(t, err, business.ErrPlanNotPolarCollected)
		assert.Nil(t, result)
	})
}

func (ts *PolarSubscriptionSuite) TestStartPolarSubscription_NilPlan_ReturnsError() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		_, _, resources := ts.CreateService(t, dep)

		in := business.StartPolarInput{ProfileID: util.IDString()}
		result, err := resources.PolarSubscriptionBusiness.StartPolarSubscription(t.Context(), nil, in)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// ---------------------------------------------------------------------------
// MirrorSubscriptionState tests
// ---------------------------------------------------------------------------

func (ts *PolarSubscriptionSuite) TestMirrorSubscriptionState_Active_TransitionsToActive() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		plan, start := startPendingPolarSub(ctx, t, resources, profileID)
		polarSubID := "polarsub_" + util.IDString()
		periodEnd := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)

		in := business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusActive,
			CurrentPeriodEnd:    periodEnd,
		}
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, in)

		require.NoError(t, err)
		require.NotNil(t, sub)
		assert.Equal(t, models.SubscriptionStateActive, sub.State)
		assert.NotNil(t, sub.EndAt)
		assert.Equal(t, polarSubID, sub.Data[business.PolarDataKeySubscriptionID])
		assert.Equal(t, business.PolarStatusActive, sub.Data[business.PolarDataKeyStatus])
		_ = start
	})
}

func (ts *PolarSubscriptionSuite) TestMirrorSubscriptionState_CanceledFuturePeriodEnd_KeepsActiveWithFlag() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		plan, _ := startPendingPolarSub(ctx, t, resources, profileID)
		polarSubID := "polarsub_" + util.IDString()

		// First: activate.
		futureEnd := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		_, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusActive,
			CurrentPeriodEnd:    futureEnd,
		})
		require.NoError(t, err)

		// Then: canceled but still within period.
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusCanceled,
			CurrentPeriodEnd:    futureEnd,
		})

		require.NoError(t, err)
		require.NotNil(t, sub)
		assert.Equal(t, models.SubscriptionStateActive, sub.State,
			"should remain ACTIVE when period end is in the future")
		assert.Equal(t, "true", sub.Data[business.PolarDataKeyCancelAtPeriodEnd],
			"cancelAtPeriodEnd flag must be set")
	})
}

func (ts *PolarSubscriptionSuite) TestMirrorSubscriptionState_CanceledPastPeriodEnd_SetsCancelled() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		plan, _ := startPendingPolarSub(ctx, t, resources, profileID)
		polarSubID := "polarsub_" + util.IDString()

		// Activate with a past period end.
		pastEnd := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
		_, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusActive,
			CurrentPeriodEnd:    pastEnd,
		})
		require.NoError(t, err)

		// Canceled with period already ended.
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusCanceled,
			CurrentPeriodEnd:    pastEnd,
		})

		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateCancelled, sub.State)
		assert.NotNil(t, sub.CancelledAt)
	})
}

func (ts *PolarSubscriptionSuite) TestMirrorSubscriptionState_Revoked_SetsExpired() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		plan, _ := startPendingPolarSub(ctx, t, resources, profileID)
		polarSubID := "polarsub_" + util.IDString()

		// Activate first.
		_, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusActive,
			CurrentPeriodEnd:    time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339),
		})
		require.NoError(t, err)

		// Revoke.
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusRevoked,
		})

		require.NoError(t, err)
		assert.Equal(t, models.SubscriptionStateExpired, sub.State)
		assert.NotNil(t, sub.CancelledAt)
	})
}

func (ts *PolarSubscriptionSuite) TestMirrorSubscriptionState_Idempotent() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		plan, _ := startPendingPolarSub(ctx, t, resources, profileID)
		polarSubID := "polarsub_" + util.IDString()
		periodEnd := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)

		mirrorIn := business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusActive,
			CurrentPeriodEnd:    periodEnd,
		}

		// Apply twice.
		sub1, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, mirrorIn)
		require.NoError(t, err)

		sub2, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, mirrorIn)
		require.NoError(t, err)

		assert.Equal(t, sub1.State, sub2.State, "state must be same on repeated apply")
		assert.Equal(t, models.SubscriptionStateActive, sub2.State)
	})
}

func (ts *PolarSubscriptionSuite) TestMirrorSubscriptionState_UnknownSubscription_ReturnsError() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		_, _, resources := ts.CreateService(t, dep)

		in := business.MirrorInput{
			PolarSubscriptionID: "polarsub_nonexistent",
			ProfileID:           util.IDString(),
			PlanID:              util.IDString(),
			PolarStatus:         business.PolarStatusActive,
		}
		sub, err := resources.PolarSubscriptionBusiness.MirrorSubscriptionState(t.Context(), in)

		require.Error(t, err)
		require.ErrorIs(t, err, business.ErrPolarSubscriptionNotFound)
		assert.Nil(t, sub)
	})
}

// ---------------------------------------------------------------------------
// RunBilling polar-collected guard
// ---------------------------------------------------------------------------

func (ts *PolarSubscriptionSuite) TestRunBilling_PolarCollectedPlan_NoInvoiceGenerated() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		// Create a polar-collected plan.
		_, plan := makePolarPlan(ctx, t, resources, "prod_billing_guard_test")

		// Start subscription and activate it via mirror so RunBilling finds it ACTIVE.
		in := business.StartPolarInput{
			ProfileID:        profileID,
			PlanID:           plan.GetID(),
			CatalogVersionID: plan.CatalogVersionID,
			Currency:         "KES",
		}
		start, err := resources.PolarSubscriptionBusiness.StartPolarSubscription(ctx, plan, in)
		require.NoError(t, err)

		polarSubID := "polarsub_guard_" + util.IDString()
		periodEnd := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		_, err = resources.PolarSubscriptionBusiness.MirrorSubscriptionState(ctx, business.MirrorInput{
			PolarSubscriptionID: polarSubID,
			ProfileID:           profileID,
			PlanID:              plan.GetID(),
			PolarStatus:         business.PolarStatusActive,
			CurrentPeriodEnd:    periodEnd,
		})
		require.NoError(t, err)

		// RunBilling must complete immediately with no invoice.
		now := time.Now()
		run, err := resources.BillingWorkflow.RunBilling(
			ctx,
			start.Subscription.GetID(),
			now.AddDate(0, -1, 0),
			now,
		)

		require.NoError(t, err)
		require.NotNil(t, run)
		assert.Equal(t, models.BillingRunStateCompleted, run.State,
			"polar-collected run must be COMPLETED immediately")
		assert.Empty(t, run.InvoiceID,
			"no invoice must be generated for a polar-collected subscription")
	})
}
