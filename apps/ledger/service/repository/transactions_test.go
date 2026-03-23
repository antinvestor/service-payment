package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	"github.com/antinvestor/service-payments/internal/utility"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TransactionsSuite tests are merged into a single test method to minimize
// database container creation (avoids Postgres resource exhaustion).
type TransactionsSuite struct {
	tests.BaseTestSuite
}

func TestTransactionsSuite(t *testing.T) {
	suite.Run(t, new(TransactionsSuite))
}

func (ts *TransactionsSuite) TestAllTransactionRepoOps() { //nolint:gocognit
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		// Setup fixtures
		ledger := &models.Ledger{Type: models.LedgerTypeAsset}
		err := resources.LedgerRepository.Create(ctx, ledger)
		require.NoError(t, err)

		acc1 := &models.Account{LedgerID: ledger.ID, Currency: "USD", LedgerType: models.LedgerTypeAsset}
		acc1.ID = "txn-test-acc"
		err = resources.AccountRepository.Create(ctx, acc1)
		require.NoError(t, err)

		acc2 := &models.Account{LedgerID: ledger.ID, Currency: "USD", LedgerType: models.LedgerTypeAsset}
		acc2.ID = "txn-test-acc-2"
		err = resources.AccountRepository.Create(ctx, acc2)
		require.NoError(t, err)

		txnRepo := resources.TransactionRepository

		// --- Test: Transaction Search ---
		txn := &models.Transaction{
			Currency:        "USD",
			TransactionType: "NORMAL",
			TransactedAt:    time.Now(),
			Entries: []*models.TransactionEntry{
				{AccountID: "txn-test-acc", Amount: utility.DecPtr(decimalx.NewFromInt64(100)), Credit: false},
				{AccountID: "txn-test-acc-2", Amount: utility.DecPtr(decimalx.NewFromInt64(-100)), Credit: true},
			},
		}
		txn.ID = "search-test-txn"
		txn.Entries[0].ID = "entry-1"
		txn.Entries[0].TransactionID = txn.ID
		txn.Entries[1].ID = "entry-2"
		txn.Entries[1].TransactionID = txn.ID

		err = txnRepo.Create(ctx, txn)
		require.NoError(t, err)

		result, err := txnRepo.SearchAsESQ(ctx, "{}")
		require.NoError(t, err)

		var allTxns []*models.Transaction
		for {
			res, ok := result.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allTxns = append(allTxns, res.Item()...)
		}
		assert.GreaterOrEqual(t, len(allTxns), 1)

		// Verify entries were loaded
		for _, foundTxn := range allTxns {
			if foundTxn.ID == "search-test-txn" {
				assert.Len(t, foundTxn.Entries, 2)
			}
		}

		// --- Test: Search Entries ---
		entryResult, err := txnRepo.SearchEntries(ctx, "{}")
		require.NoError(t, err)

		var allEntries []*models.TransactionEntry
		for {
			res, ok := entryResult.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allEntries = append(allEntries, res.Item()...)
		}
		assert.GreaterOrEqual(t, len(allEntries), 1)

		// --- Test: Search with Filter ---
		for _, id := range []string{"filter-txn-1", "filter-txn-2"} {
			ftxn := &models.Transaction{Currency: "USD", TransactionType: "NORMAL", TransactedAt: time.Now()}
			ftxn.ID = id
			err = txnRepo.Create(ctx, ftxn)
			require.NoError(t, err)
		}

		filterResult, err := txnRepo.SearchAsESQ(
			ctx,
			`{"query": {"must": {"fields": [{"id": {"eq": "filter-txn-1"}}]}}}`,
		)
		require.NoError(t, err)

		var found []*models.Transaction
		for {
			res, ok := filterResult.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			found = append(found, res.Item()...)
		}
		assert.Len(t, found, 1)
		assert.Equal(t, "filter-txn-1", found[0].ID)

		// --- Test: HasTransactionEntries returns true ---
		has, err := resources.AccountRepository.HasTransactionEntries(ctx, "txn-test-acc")
		require.NoError(t, err)
		assert.True(t, has)

		// --- Test: Search with pagination ---
		for i := range 5 {
			ptxn := &models.Transaction{Currency: "USD", TransactionType: "NORMAL", TransactedAt: time.Now()}
			ptxn.ID = fmt.Sprintf("page-txn-%d", i)
			err = txnRepo.Create(ctx, ptxn)
			require.NoError(t, err)
		}

		// size=3 with a filter condition caps total results
		pageResult, err := txnRepo.SearchAsESQ(
			ctx,
			`{"size": 3, "query": {"must": {"fields": [{"currency": {"eq": "USD"}}]}}}`,
		)
		require.NoError(t, err)

		var cappedTxns []*models.Transaction
		for {
			res, ok := pageResult.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			cappedTxns = append(cappedTxns, res.Item()...)
		}
		assert.LessOrEqual(t, len(cappedTxns), 3)

		// Search without size limit to get all
		allResult, err := txnRepo.SearchAsESQ(ctx, `{}`)
		require.NoError(t, err)

		var allTxns2 []*models.Transaction
		for {
			res, ok := allResult.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allTxns2 = append(allTxns2, res.Item()...)
		}
		assert.GreaterOrEqual(t, len(allTxns2), 8) // search-test-txn + 2 filter + 5 page
	})
}
