package repository_test

import (
	"context"
	"fmt"
	"testing"

	models "github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LedgersSuite struct {
	tests.BaseTestSuite
	ledger *models.Ledger
}

func (ls *LedgersSuite) setupFixtures(ctx context.Context, svc *frame.Service) {
	// Create test ledger.
	dbPool := svc.DatastoreManager().GetPool(ctx, datastore.DefaultPoolName)
	workMan := svc.WorkManager()
	ledgersDB := repository.NewLedgerRepository(ctx, dbPool, workMan)

	lg := &models.Ledger{Type: models.LedgerTypeAsset}
	err := ledgersDB.Create(ctx, lg)
	if err != nil {
		ls.T().Fatalf("Error creating ledger: %v", err)
	}
	ls.ledger = lg
}

func (ls *LedgersSuite) TestLedgersInfoAPI() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, svc, _ := ls.CreateService(t, dep)
		ls.setupFixtures(ctx, svc)

		dbPool := svc.DatastoreManager().GetPool(ctx, datastore.DefaultPoolName)
		workMan := svc.WorkManager()
		ledgersDB := repository.NewLedgerRepository(ctx, dbPool, workMan)
		lg, err := ledgersDB.GetByID(ctx, ls.ledger.ID)
		require.NoError(t, err, "Error while getting ledger")
		assert.Equal(t, ls.ledger.ID, lg.ID, "Invalid ledger id")
	})
}

func (ls *LedgersSuite) TestLedgerSearch() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ls.CreateService(t, dep)

		ledgerRepo := resources.LedgerRepository

		// Create ledgers with different types
		asset := &models.Ledger{Type: models.LedgerTypeAsset}
		err := ledgerRepo.Create(ctx, asset)
		require.NoError(t, err)

		liability := &models.Ledger{Type: models.LedgerTypeLiability}
		err = ledgerRepo.Create(ctx, liability)
		require.NoError(t, err)

		// Search for all ledgers
		result, err := ledgerRepo.SearchAsESQ(ctx, "{}")
		require.NoError(t, err)

		var allLedgers []*models.Ledger
		for {
			res, ok := result.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allLedgers = append(allLedgers, res.Item()...)
		}
		assert.GreaterOrEqual(t, len(allLedgers), 2, "Should find at least 2 ledgers")
	})
}

func (ls *LedgersSuite) TestLedgerSearchByType() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ls.CreateService(t, dep)

		ledgerRepo := resources.LedgerRepository

		// Create ledgers with specific types
		income := &models.Ledger{Type: models.LedgerTypeIncome}
		err := ledgerRepo.Create(ctx, income)
		require.NoError(t, err)

		expense := &models.Ledger{Type: models.LedgerTypeExpense}
		err = ledgerRepo.Create(ctx, expense)
		require.NoError(t, err)

		// Search by type
		query := `{"query": {"must": {"fields": [{"type": {"eq": "INCOME"}}]}}}`
		result, err := ledgerRepo.SearchAsESQ(ctx, query)
		require.NoError(t, err)

		var found []*models.Ledger
		for {
			res, ok := result.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			found = append(found, res.Item()...)
		}
		assert.GreaterOrEqual(t, len(found), 1)
		for _, lg := range found {
			assert.Equal(t, models.LedgerTypeIncome, lg.Type)
		}
	})
}

func (ls *LedgersSuite) TestLedgerSearchWithPagination() {
	ls.WithTestDependencies(ls.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ls.CreateService(t, dep)

		ledgerRepo := resources.LedgerRepository

		// Create enough ledgers to test pagination (use size=2 to force multiple batches)
		for i := range 5 {
			lg := &models.Ledger{Type: models.LedgerTypeAsset}
			lg.ID = fmt.Sprintf("page-ledger-%d", i)
			err := ledgerRepo.Create(ctx, lg)
			require.NoError(t, err)
		}

		// Search with size=3 and a filter condition (empty query ignores size)
		query := `{"size": 3, "query": {"must": {"fields": [{"type": {"eq": "ASSET"}}]}}}`
		result, err := ledgerRepo.SearchAsESQ(ctx, query)
		require.NoError(t, err)

		var capped []*models.Ledger
		for {
			res, ok := result.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			capped = append(capped, res.Item()...)
		}
		assert.LessOrEqual(t, len(capped), 3)

		// Search without size limit to get all
		result2, err := ledgerRepo.SearchAsESQ(ctx, `{}`)
		require.NoError(t, err)

		var allLedgers []*models.Ledger
		for {
			res, ok := result2.ReadResult(ctx)
			if !ok {
				break
			}
			if res.IsError() {
				t.Fatalf("unexpected error: %v", res.Error())
			}
			allLedgers = append(allLedgers, res.Item()...)
		}
		assert.GreaterOrEqual(t, len(allLedgers), 5)
	})
}

func TestLedgersSuite(t *testing.T) {
	suite.Run(t, new(LedgersSuite))
}
