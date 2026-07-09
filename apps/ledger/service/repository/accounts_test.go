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
	"context"
	"fmt"
	"testing"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AccountsSuite struct {
	tests.BaseTestSuite
	ledger *models.Ledger
}

func TestAccountsSuite(t *testing.T) {
	suite.Run(t, new(AccountsSuite))
}

func (as *AccountsSuite) setupFixtures(ctx context.Context, resources *tests.ServiceResources) {
	// Create test accounts using cached repositories
	ledgersDB := resources.LedgerRepository
	accountsDB := resources.AccountRepository

	as.ledger = &models.Ledger{Type: models.LedgerTypeAsset}
	err := ledgersDB.Create(ctx, as.ledger)
	as.Require().NoError(err, "Unable to create ledger for account")

	account := &models.Account{LedgerID: as.ledger.ID, Currency: "UGX", LedgerType: models.LedgerTypeAsset}
	account.ID = "100"
	err = accountsDB.Create(ctx, account)
	if err != nil {
		as.Require().NoError(err, "Unable to create account")
	}
}

func (as *AccountsSuite) TestAccountsInfoAPI() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		// Use cached account repository from dependencies
		account, err := resources.AccountRepository.GetByID(ctx, "100")
		require.NoError(t, err, "Error getting account info api account")
		assert.Equal(t, "100", account.ID, "Invalid account Reference")
		assert.True(t, account.Balance != nil && account.Balance.IsZero(), "Invalid account balance")
	})
}

func (as *AccountsSuite) TestAccountSearch() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountRepo := resources.AccountRepository

		// Create additional accounts
		acc2 := &models.Account{LedgerID: as.ledger.ID, Currency: "USD", LedgerType: models.LedgerTypeAsset}
		acc2.ID = "200"
		err := accountRepo.Create(ctx, acc2)
		require.NoError(t, err)

		// Search all accounts
		result, err := accountRepo.SearchAsESQ(ctx, "{}")
		require.NoError(t, err)

		var allAccounts []*models.Account
		for {
			res, ok := result.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allAccounts = append(allAccounts, res.Item()...)
		}
		assert.GreaterOrEqual(t, len(allAccounts), 2, "Should find at least 2 accounts")
	})
}

func (as *AccountsSuite) TestAccountListByID_MultipleAccounts() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountRepo := resources.AccountRepository

		// Create additional accounts
		acc2 := &models.Account{LedgerID: as.ledger.ID, Currency: "EUR", LedgerType: models.LedgerTypeAsset}
		acc2.ID = "300"
		err := accountRepo.Create(ctx, acc2)
		require.NoError(t, err)

		// List by IDs
		accs, err := accountRepo.ListByID(ctx, "100", "300")
		require.NoError(t, err)
		assert.Len(t, accs, 2)
		assert.NotNil(t, accs["100"])
		assert.NotNil(t, accs["300"])
	})
}

func (as *AccountsSuite) TestAccountListByID_EmptyIDs() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)

		_, err := resources.AccountRepository.ListByID(ctx)
		require.Error(t, err)
	})
}

func (as *AccountsSuite) TestAccountGetByID_EmptyID() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)

		acc, err := resources.AccountRepository.GetByID(ctx, "")
		require.Error(t, err)
		assert.Nil(t, acc)
	})
}

func (as *AccountsSuite) TestHasTransactionEntries_NoEntries() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		has, err := resources.AccountRepository.HasTransactionEntries(ctx, "100")
		require.NoError(t, err)
		assert.False(t, has)
	})
}

func (as *AccountsSuite) TestCountByLedgerID() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		count, err := resources.AccountRepository.CountByLedgerID(ctx, as.ledger.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func (as *AccountsSuite) TestAccountSearchWithPagination() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountRepo := resources.AccountRepository

		// Create additional accounts
		for i := range 5 {
			acc := &models.Account{
				LedgerID:   as.ledger.ID,
				Currency:   "USD",
				LedgerType: models.LedgerTypeAsset,
			}
			acc.ID = fmt.Sprintf("page-acc-%d", i)
			err := accountRepo.Create(ctx, acc)
			require.NoError(t, err)
		}

		// Search with size limiting total results to 3
		query := `{"size": 3, "query": {"must": {"fields": [{"ledger_id": {"eq": "` + as.ledger.ID + `"}}]}}}`
		result, err := accountRepo.SearchAsESQ(ctx, query)
		require.NoError(t, err)

		var allAccounts []*models.Account
		for {
			res, ok := result.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allAccounts = append(allAccounts, res.Item()...)
		}
		// size=3 caps total results at 3
		assert.LessOrEqual(t, len(allAccounts), 3)
		assert.GreaterOrEqual(t, len(allAccounts), 1)

		// Now search without size limit to get all
		result2, err := accountRepo.SearchAsESQ(
			ctx,
			`{"query": {"must": {"fields": [{"ledger_id": {"eq": "`+as.ledger.ID+`"}}]}}}`,
		)
		require.NoError(t, err)

		var allAccounts2 []*models.Account
		for {
			res, ok := result2.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allAccounts2 = append(allAccounts2, res.Item()...)
		}
		// Should find all 6 (5 created + 1 fixture)
		assert.GreaterOrEqual(t, len(allAccounts2), 6)
	})
}

func (as *AccountsSuite) TestCountByLedgerID_NoAccounts() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)

		// Create a ledger with no accounts
		lg := &models.Ledger{Type: models.LedgerTypeAsset}
		lg.ID = "empty-ledger"
		err := resources.LedgerRepository.Create(ctx, lg)
		require.NoError(t, err)

		count, err := resources.AccountRepository.CountByLedgerID(ctx, "empty-ledger")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
