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
	"testing"

	models "github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/require"
)

// TestLedgerTenantIsolation verifies Postgres RLS hides one tenant's
// ledgers from another tenant. The suite drops application queries to
// an unprivileged role, so the tenancy policies installed during
// migration are actually enforced.
func (ls *LedgersSuite) TestLedgerTenantIsolation() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ls.CreateService(t, dep)

		ledgerRepo := resources.LedgerRepository

		ctxA := ls.WithAuthClaims(ctx, "tenant-a", "partition-a", util.IDString())
		ctxB := ls.WithAuthClaims(ctx, "tenant-b", "partition-b", util.IDString())

		lg := &models.Ledger{Type: models.LedgerTypeAsset}
		require.NoError(t, ledgerRepo.Create(ctxA, lg))
		require.Equal(t, "tenant-a", lg.TenantID)

		// Owning tenant reads its ledger back.
		got, err := ledgerRepo.GetByID(ctxA, lg.ID)
		require.NoError(t, err)
		require.Equal(t, lg.ID, got.ID)

		// Another tenant must not be able to read it.
		_, err = ledgerRepo.GetByID(ctxB, lg.ID)
		require.Error(t, err, "tenant B must not read tenant A's ledger")

		// Cross-tenant search must come back empty.
		require.Empty(t, collectLedgers(ctxB, t, ledgerRepo),
			"tenant B search must not return tenant A's ledgers")
	})
}

// TestAccountTenantIsolation verifies accounts are isolated per tenant.
func (ls *LedgersSuite) TestAccountTenantIsolation() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ls.CreateService(t, dep)

		ledgerRepo := resources.LedgerRepository
		accountRepo := resources.AccountRepository

		ctxA := ls.WithAuthClaims(ctx, "tenant-a", "partition-a", util.IDString())
		ctxB := ls.WithAuthClaims(ctx, "tenant-b", "partition-b", util.IDString())

		lg := &models.Ledger{Type: models.LedgerTypeAsset}
		require.NoError(t, ledgerRepo.Create(ctxA, lg))

		acc := &models.Account{
			Currency:   "KES",
			LedgerID:   lg.ID,
			LedgerType: models.LedgerTypeAsset,
		}
		require.NoError(t, accountRepo.Create(ctxA, acc))

		// Owning tenant reads its account back.
		got, err := accountRepo.GetByID(ctxA, acc.ID)
		require.NoError(t, err)
		require.Equal(t, acc.ID, got.ID)

		// Another tenant must not be able to read it. GetByID's contract
		// is (nil, nil) when no row is visible.
		invisible, err := accountRepo.GetByID(ctxB, acc.ID)
		require.NoError(t, err)
		require.Nil(t, invisible, "tenant B must not read tenant A's account")

		listed, err := accountRepo.ListByID(ctxB, acc.ID)
		require.NoError(t, err)
		require.Empty(t, listed, "tenant B must not list tenant A's accounts")
	})
}

// TestBookTenantIsolation verifies books are isolated per tenant, which
// requires the books table to keep the RLS policy frame installs on it.
func (ls *LedgersSuite) TestBookTenantIsolation() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ls.CreateService(t, dep)

		bookRepo := resources.BookRepository

		ctxA := ls.WithAuthClaims(ctx, "tenant-a", "partition-a", util.IDString())
		ctxB := ls.WithAuthClaims(ctx, "tenant-b", "partition-b", util.IDString())

		book := &models.Book{Name: "Group A", Type: models.BookTypeGroup}
		book.GenID(ctxA)
		require.NoError(t, bookRepo.Create(ctxA, book))

		got, err := bookRepo.GetByID(ctxA, book.ID)
		require.NoError(t, err)
		require.Equal(t, book.ID, got.ID)

		_, err = bookRepo.GetByID(ctxB, book.ID)
		require.Error(t, err, "tenant B must not read tenant A's book")

		listed, err := bookRepo.ListByType(ctxB, models.BookTypeGroup)
		require.NoError(t, err)
		require.Empty(t, listed, "tenant B must not list tenant A's books")

		// Tenant B cannot attach its records to tenant A's book.
		foreignBook := book.ID
		lg := &models.Ledger{Type: models.LedgerTypeAsset, BookID: &foreignBook}
		require.Error(t, resources.LedgerRepository.Create(ctxB, lg),
			"tenant B must not reference tenant A's book")

		// The same id is free in tenant B's own tenancy.
		mine := &models.Book{BaseModel: data.BaseModel{ID: book.ID}, Name: "Group B", Type: models.BookTypeGroup}
		mine.GenID(ctxB)
		require.NoError(t, bookRepo.Create(ctxB, mine))

		got, err = bookRepo.GetByID(ctxA, book.ID)
		require.NoError(t, err)
		require.Equal(t, "Group A", got.Name)
		got, err = bookRepo.GetByID(ctxB, book.ID)
		require.NoError(t, err)
		require.Equal(t, "Group B", got.Name)
	})
}

// TestMigrationRerunKeepsBookTenancyKey verifies a second migration run, as
// every deploy does, leaves book ids scoped to the tenancy: AutoMigrate must
// not restore a global key on books.
func (ls *LedgersSuite) TestMigrationRerunKeepsBookTenancyKey() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, svc, resources := ls.CreateService(t, dep)

		require.NoError(t, repository.Migrate(ctx, svc.DatastoreManager(), "../../migrations/0001"))

		ctxA := ls.WithAuthClaims(ctx, "tenant-a", "partition-a", util.IDString())
		ctxB := ls.WithAuthClaims(ctx, "tenant-b", "partition-b", util.IDString())
		for _, c := range []context.Context{ctxA, ctxB} {
			book := &models.Book{BaseModel: data.BaseModel{ID: "shared-book-id"}, Name: "Group", Type: models.BookTypeGroup}
			book.GenID(c)
			require.NoError(t, resources.BookRepository.Create(c, book))
		}
	})
}

func collectLedgers(ctx context.Context, t *testing.T, repo repository.LedgerRepository) []*models.Ledger {
	t.Helper()

	result, err := repo.SearchAsESQ(ctx, "{}")
	require.NoError(t, err)

	var all []*models.Ledger
	for {
		res, ok := result.ReadResult(ctx)
		if !ok {
			break
		}
		if res.IsError() {
			t.Fatalf("unexpected search error: %v", res.Error())
		}
		all = append(all, res.Item()...)
	}
	return all
}
