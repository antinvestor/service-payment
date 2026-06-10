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

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TenantIsolationSuite struct {
	tests.BaseTestSuite
}

func TestTenantIsolationSuite(t *testing.T) {
	suite.Run(t, new(TenantIsolationSuite))
}

// TestSubscriptionTenantIsolation verifies Postgres RLS hides one
// tenant's subscriptions from another tenant. The suite drops
// application queries to an unprivileged role, so the tenancy
// policies installed during migration are actually enforced.
func (ts *TenantIsolationSuite) TestSubscriptionTenantIsolation() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		subBus := resources.SubscriptionBusiness

		ctxA := ts.WithAuthClaims(ctx, "tenant-a", "partition-a", util.IDString())
		ctxB := ts.WithAuthClaims(ctx, "tenant-b", "partition-b", util.IDString())

		profileID := util.IDString()
		sub, err := subBus.CreateSubscription(ctxA, &models.Subscription{
			ProfileID:        profileID,
			PlanID:           util.IDString(),
			CatalogVersionID: util.IDString(),
			Currency:         "KES",
		})
		require.NoError(t, err)
		require.Equal(t, "tenant-a", sub.TenantID)

		// Owning tenant reads its subscription back.
		got, err := subBus.GetSubscription(ctxA, sub.ID)
		require.NoError(t, err)
		require.Equal(t, sub.ID, got.ID)

		// Another tenant must not be able to read it.
		_, err = subBus.GetSubscription(ctxB, sub.ID)
		require.Error(t, err, "tenant B must not read tenant A's subscription")

		// Even for the same profile, listings stay within the tenant.
		list, err := subBus.ListActiveByProfile(ctxB, profileID)
		require.NoError(t, err)
		require.Empty(t, list, "tenant B must not list tenant A's subscriptions")
	})
}
