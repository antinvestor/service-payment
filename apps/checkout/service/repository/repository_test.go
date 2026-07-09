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

package repository_test

import (
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/antinvestor/service-payments/apps/checkout/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RepositorySuite struct {
	tests.BaseTestSuite
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

func (rs *RepositorySuite) TestSessionSaveAndGetByRef() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := rs.CreateService(t, dep)

		sessionRepo := resources.SessionRepository

		ref := util.RandomAlphaNumericString(32)
		expiresAt := time.Now().Add(30 * time.Minute)

		session := &models.CheckoutSession{
			Ref:       ref,
			Name:      "Test order",
			Amount:    "150",
			Currency:  "KES",
			Status:    models.SessionStatusPending,
			ExpiresAt: expiresAt,
		}

		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err, "failed to create checkout session")
		require.NotEmpty(t, session.ID, "session ID should be set after create")

		fetched, err := sessionRepo.GetByRef(ctx, ref)
		require.NoError(t, err, "failed to get checkout session by ref")

		assert.Equal(t, session.ID, fetched.ID, "session ID should match")
		assert.Equal(t, "KES", fetched.Currency, "currency should match")
		assert.Equal(t, models.SessionStatusPending, fetched.Status, "status should be pending")
	})
}

func (rs *RepositorySuite) TestLinkSaveAndGetByRef() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := rs.CreateService(t, dep)

		linkRepo := resources.LinkRepository

		ref := util.RandomAlphaNumericString(32)

		link := &models.CheckoutLink{
			Ref:          ref,
			Name:         "Test link",
			Currency:     "KES",
			AmountOption: models.AmountOptionVariable,
			Active:       true,
		}

		err := linkRepo.Create(ctx, link)
		require.NoError(t, err, "failed to create checkout link")
		require.NotEmpty(t, link.ID, "link ID should be set after create")

		fetched, err := linkRepo.GetByRef(ctx, ref)
		require.NoError(t, err, "failed to get checkout link by ref")

		assert.Equal(t, link.ID, fetched.ID, "link ID should match")
		assert.True(t, fetched.IsUsable(time.Now()), "link should be usable")
	})
}
