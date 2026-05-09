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

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/stretchr/testify/require"
)

func (ss *SearchSuite) TestSearchAccountsWithMustFields() {
	ss.WithTestDependencies(ss.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ss.CreateService(t, dep)
		ss.setupFixtures(ctx, resources)

		query := `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"eq": "acc1"}},
                    {"balance": {"gt": 0}}
                ]
            }
        }
    }`
		jobResult, err := resources.AccountRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		accounts, err := resultsToSlice[*models.Account](jobResult)

		require.NoError(t, err, "Error querying must fields")
		require.Len(t, accounts, 1, "Accounts count doesn't match")
		if len(accounts) > 0 {
			require.Equal(t, "acc1", accounts[0].ID, "Account Reference doesn't match")
		}
		query = `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"eq": "acc2"}},
                    {"balance": {"gt": 0}}
                ]
            }
        }
    }`

		jobResult, err = resources.AccountRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		accounts, err = resultsToSlice[*models.Account](jobResult)

		require.NoError(t, err, "Error querying must fields")
		require.Empty(t, accounts, "No account should exist for given query")
	})
}

func (ss *SearchSuite) TestSearchTransactionsWithMustFields() {
	ss.WithTestDependencies(ss.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ss.CreateService(t, dep)
		ss.setupFixtures(ctx, resources)

		query := `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"eq": "txn1"}},
                    {"transacted_at": {"gte": "2017-08-08"}}
                ]
            }
        }
    }`

		resultChannel, err := resources.TransactionRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		transactions, err := resultsToSlice[*models.Transaction](resultChannel)
		require.NoError(t, err, "Error in building search query")
		require.Len(t, transactions, 1, "Transactions count doesn't match")
		if len(transactions) > 0 {
			require.Equal(t, "txn1", transactions[0].ID, "Transaction Reference doesn't match")
		}
		query = `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"eq": "txn2"}},
                    {"transacted_at": {"lt": "2017-08-08"}}
                ]
            }
        }
    }`

		resultChannel, err = resources.TransactionRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		transactions, err = resultsToSlice[*models.Transaction](resultChannel)
		require.NoError(t, err, "Error in building search query")
		require.Empty(t, transactions, "No transaction should exist for given query")
	})
}

func (ss *SearchSuite) TestSearchAccountsWithFieldOperators() {
	ss.WithTestDependencies(ss.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ss.CreateService(t, dep)
		ss.setupFixtures(ctx, resources)

		query := `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"eq": "acc1"}}
                ]
            }
        }
    }`

		resultPipe, err := resources.AccountRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		accounts, err := resultsToSlice[*models.Account](resultPipe)

		require.NoError(t, err, "Error in building search query")
		require.Len(t, accounts, 1, "Accounts count doesn't match")
		if len(accounts) > 0 {
			require.Equal(t, "acc1", accounts[0].ID, "Account Reference doesn't match")
		}

		query = `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"ne": "acc1"}}
                ]
            }
        }
    }`

		resultPipe, err = resources.AccountRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		accounts, err = resultsToSlice[*models.Account](resultPipe)

		require.NoError(t, err, "Error in building search query")
		require.Len(t, accounts, 1, "Accounts count doesn't match")
		require.Equal(t, "acc2", accounts[0].ID, "Account Reference doesn't match")

		query = `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"like": "%c1"}}
                ]
            }
        }
    }`

		resultPipe, err = resources.AccountRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		accounts, err = resultsToSlice[*models.Account](resultPipe)

		require.NoError(t, err, "Error in building search query")
		require.Len(t, accounts, 1, "Accounts count doesn't match")
		require.Equal(t, "acc1", accounts[0].ID, "Account Reference doesn't match")

		query = `{
        "query": {
            "must": {
                "fields": [
                    {"id": {"notlike": "%c1"}}
                ]
            }
        }
    }`

		resultPipe, err = resources.AccountRepository.SearchAsESQ(ctx, query)
		require.NoError(t, err)
		accounts, err = resultsToSlice[*models.Account](resultPipe)

		require.NoError(t, err, "Error in building search query")
		require.Len(t, accounts, 1, "Accounts count doesn't match")
		require.Equal(t, "acc2", accounts[0].ID, "Account Reference doesn't match")
	})
}
