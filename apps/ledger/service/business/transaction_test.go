package business_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/ledger/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/structpb"
)

type TransactionBusinessSuite struct {
	tests.BaseTestSuite
	assetLedger  *models.Ledger
	incomeLedger *models.Ledger
}

func TestTransactionBusinessSuite(t *testing.T) {
	suite.Run(t, new(TransactionBusinessSuite))
}

func (ts *TransactionBusinessSuite) setupFixtures(ctx context.Context, resources *tests.ServiceResources) {
	// Create test ledgers using business layer
	ledgerBusiness := resources.LedgerBusiness
	accountBusiness := resources.AccountBusiness

	// Create asset ledger
	assetLedgerReq := &ledgerv1.CreateLedgerRequest{
		Id:   "test-asset-ledger",
		Type: ledgerv1.LedgerType_ASSET,
	}
	assetLedger, err := ledgerBusiness.CreateLedger(ctx, assetLedgerReq)
	ts.Require().NoError(err, "Unable to create asset ledger")
	ts.assetLedger = &models.Ledger{
		BaseModel: data.BaseModel{ID: assetLedger.GetId()},
		Type:      models.FromLedgerType(assetLedger.GetType()),
	}
	ts.T().Logf("Created asset ledger with ID: %s", assetLedger.GetId())

	// Create income ledger
	incomeLedgerReq := &ledgerv1.CreateLedgerRequest{
		Id:   "test-income-ledger",
		Type: ledgerv1.LedgerType_INCOME,
	}
	incomeLedger, err := ledgerBusiness.CreateLedger(ctx, incomeLedgerReq)
	ts.Require().NoError(err, "Unable to create income ledger")
	ts.incomeLedger = &models.Ledger{
		BaseModel: data.BaseModel{ID: incomeLedger.GetId()},
		Type:      models.FromLedgerType(incomeLedger.GetType()),
	}
	ts.T().Logf("Created income ledger with ID: %s", incomeLedger.GetId())

	// Create test accounts
	assetAccountReq := &ledgerv1.CreateAccountRequest{
		Id:       "asset-account",
		LedgerId: assetLedger.GetId(),
		Currency: "USD",
	}
	assetAccount, err := accountBusiness.CreateAccount(ctx, assetAccountReq)
	ts.Require().NoError(err, "Unable to create asset account")
	ts.T().Logf("Created asset account with ID: %s, Ledger: %s", assetAccount.GetId(), assetAccount.GetLedger())

	incomeAccountReq := &ledgerv1.CreateAccountRequest{
		Id:       "income-account",
		LedgerId: incomeLedger.GetId(),
		Currency: "USD",
	}
	incomeAccount, err := accountBusiness.CreateAccount(ctx, incomeAccountReq)
	ts.Require().NoError(err, "Unable to create income account")
	ts.T().Logf("Created income account with ID: %s, Ledger: %s", incomeAccount.GetId(), incomeAccount.GetLedger())

	// Verify accounts exist before creating transaction
	_, err = accountBusiness.GetAccount(ctx, "asset-account")
	ts.Require().NoError(err, "Asset account should exist before transaction creation")
	_, err = accountBusiness.GetAccount(ctx, "income-account")
	ts.Require().NoError(err, "Income account should exist before transaction creation")
}

func (ts *TransactionBusinessSuite) TestCreateTransactionWithBusinessValidation() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		timeNow := time.Now().UTC()
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "test-transaction-" + timeNow.Format("20060102150405"),
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
			TransactedAt: timeNow.Format(time.RFC3339),
			Cleared:      true,
		}

		transaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.NoError(t, err, "Error creating transaction through business layer")
		require.NotNil(t, transaction, "Transaction should be created")

		assert.Equal(
			t,
			"test-transaction-"+timeNow.Format("20060102150405"),
			transaction.GetId(),
			"Invalid transaction ID",
		)
		assert.Equal(t, "USD", transaction.GetCurrencyCode(), "Invalid currency")
		assert.Equal(t, ledgerv1.TransactionType_NORMAL, transaction.GetType(), "Invalid transaction type")
		assert.Len(t, transaction.GetEntries(), 2, "Should have 2 entries")
	})
}

func (ts *TransactionBusinessSuite) TestCreateTransactionNonZeroSum() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "invalid-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 200, Nanos: 0}, // Non-zero sum
				},
			},
		}

		transaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.Error(t, err, "Should fail with non-zero sum transaction")
		assert.Nil(t, transaction, "Transaction should not be created")
		assert.Contains(t, err.Error(), "non zero sum", "Error should mention zero sum validation")
	})
}

func (ts *TransactionBusinessSuite) TestCreateTransactionInvalidDebitCredit() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create transaction with invalid debit/credit combination (both debit but equal amounts)
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "invalid-dr-cr-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false, // Debit
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    false, // Also debit - invalid but amounts equal for zero sum
					Amount: &money.Money{
						CurrencyCode: "USD",
						Units:        -100,
						Nanos:        0,
					}, // Negative amount to make zero sum
				},
			},
		}

		transaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.Error(t, err, "Should fail with invalid debit/credit entry")
		assert.Nil(t, transaction, "Transaction should not be created")
		assert.Contains(t, err.Error(), "invalid count of dr/cr", "Error should mention debit/credit validation")
	})
}

func (ts *TransactionBusinessSuite) TestCreateTransactionWithNonExistentAccount() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "orphan-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "non-existent-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
		}

		transaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.Error(t, err, "Should fail with non-existent account")
		assert.Nil(t, transaction, "Transaction should not be created")
		assert.Contains(t, err.Error(), "not found", "Error should mention account not found")
	})
}

func (ts *TransactionBusinessSuite) TestCreateTransactionWithCurrencyMismatch() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create transaction with different currency than accounts
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "currency-mismatch-transaction",
			Currency: "EUR", // Different from account currency (USD)
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "EUR", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "EUR", Units: 100, Nanos: 0},
				},
			},
		}

		transaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.Error(t, err, "Should fail with currency mismatch")
		assert.Nil(t, transaction, "Transaction should not be created")
		assert.Contains(t, err.Error(), "currency", "Error should mention currency mismatch")
	})
}

func (ts *TransactionBusinessSuite) TestCreateReservationTransaction() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "reservation-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_RESERVATION,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "reservation-entry",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 500, Nanos: 0},
				},
			},
		}

		transaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.NoError(t, err, "Error creating reservation transaction")
		require.NotNil(t, transaction, "Reservation transaction should be created")

		assert.Equal(t, ledgerv1.TransactionType_RESERVATION, transaction.GetType(), "Should be reservation type")
		assert.Len(t, transaction.GetEntries(), 1, "Reservation should have exactly 1 entry")
	})
}

func (ts *TransactionBusinessSuite) TestCreateReservationTransactionInvalidEntries() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create reservation transaction with multiple entries (invalid)
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "invalid-reservation-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_RESERVATION,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
		}

		transaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.Error(t, err, "Should fail with invalid reservation transaction")
		assert.Nil(t, transaction, "Reservation transaction should not be created")
		assert.Contains(t, err.Error(), "invalid count of dr/cr", "Error should mention entry validation")
	})
}

func (ts *TransactionBusinessSuite) TestReverseTransaction() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// First create a normal transaction
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "original-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 1000, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 1000, Nanos: 0},
				},
			},
		}

		originalTransaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.NoError(t, err, "Error creating original transaction")

		// Now reverse it
		reverseReq := &ledgerv1.ReverseTransactionRequest{
			Id: originalTransaction.GetId(),
		}

		reversedTransaction, err := transactionBusiness.ReverseTransaction(ctx, reverseReq)
		require.NoError(t, err, "Error reversing transaction")
		require.NotNil(t, reversedTransaction, "Reversed transaction should be created")

		assert.Equal(t, ledgerv1.TransactionType_REVERSAL, reversedTransaction.GetType(), "Should be reversal type")
		assert.Contains(t, reversedTransaction.GetId(), "_REVERSAL", "Reversal transaction ID should contain _REVERSAL")
		assert.Len(t, reversedTransaction.GetEntries(), 2, "Reversal should have same number of entries")
	})
}

func (ts *TransactionBusinessSuite) TestGetTransaction() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create a transaction first
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "get-test-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 250, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 250, Nanos: 0},
				},
			},
		}

		createdTransaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.NoError(t, err, "Error creating transaction")

		// Now retrieve it
		retrievedTransaction, err := transactionBusiness.GetTransaction(ctx, "get-test-transaction")
		require.NoError(t, err, "Error retrieving transaction")

		assert.Equal(
			t,
			createdTransaction.GetId(),
			retrievedTransaction.GetId(),
			"Retrieved transaction should match created transaction",
		)
		assert.Equal(
			t,
			createdTransaction.GetCurrencyCode(),
			retrievedTransaction.GetCurrencyCode(),
			"Currency should match",
		)
		assert.Equal(t, createdTransaction.GetType(), retrievedTransaction.GetType(), "Type should match")
	})
}

func (ts *TransactionBusinessSuite) TestSearchTransactions() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create a transaction
		_, err := transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "search-txn-1",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		// Search with empty query
		searchReq := &commonv1.SearchRequest{Query: "{}"}
		var foundTxns []*ledgerv1.Transaction
		err = transactionBusiness.SearchTransactions(
			ctx,
			searchReq,
			func(_ context.Context, batch []*ledgerv1.Transaction) error {
				foundTxns = append(foundTxns, batch...)
				return nil
			},
		)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(foundTxns), 1)
	})
}

func (ts *TransactionBusinessSuite) TestSearchTransactionsWithFilter() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		_, err := transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "search-filter-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		searchReq := &commonv1.SearchRequest{
			Query: `{"query": {"must": {"fields": [{"id": {"eq": "search-filter-txn"}}]}}}`,
		}
		var foundTxns []*ledgerv1.Transaction
		err = transactionBusiness.SearchTransactions(
			ctx,
			searchReq,
			func(_ context.Context, batch []*ledgerv1.Transaction) error {
				foundTxns = append(foundTxns, batch...)
				return nil
			},
		)
		require.NoError(t, err)
		assert.Len(t, foundTxns, 1)
	})
}

func (ts *TransactionBusinessSuite) TestSearchAndDeleteOperations() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create a transaction for search tests
		_, err := transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "search-ops-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		// Test SearchTransactions with empty query
		searchReq := &commonv1.SearchRequest{Query: "{}"}
		var foundTxns []*ledgerv1.Transaction
		err = transactionBusiness.SearchTransactions(
			ctx,
			searchReq,
			func(_ context.Context, batch []*ledgerv1.Transaction) error {
				foundTxns = append(foundTxns, batch...)
				return nil
			},
		)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(foundTxns), 1)

		// Test SearchTransactions with filter
		filterReq := &commonv1.SearchRequest{
			Query: `{"query": {"must": {"fields": [{"id": {"eq": "search-ops-txn"}}]}}}`,
		}
		var filteredTxns []*ledgerv1.Transaction
		err = transactionBusiness.SearchTransactions(
			ctx,
			filterReq,
			func(_ context.Context, batch []*ledgerv1.Transaction) error {
				filteredTxns = append(filteredTxns, batch...)
				return nil
			},
		)
		require.NoError(t, err)
		assert.Len(t, filteredTxns, 1)

		// Test SearchEntries
		var foundEntries []*ledgerv1.TransactionEntry
		err = transactionBusiness.SearchEntries(
			ctx,
			searchReq,
			func(_ context.Context, batch []*ledgerv1.TransactionEntry) error {
				foundEntries = append(foundEntries, batch...)
				return nil
			},
		)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(foundEntries), 2)

		// Test DeleteTransaction (should fail — use reversal instead)
		err = transactionBusiness.DeleteTransaction(ctx, "some-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be deleted")

		// Test DeleteTransaction empty ID
		err = transactionBusiness.DeleteTransaction(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction ID is required")

		// Test SearchTransactions with invalid query (triggers search error result)
		err = transactionBusiness.SearchTransactions(
			ctx,
			&commonv1.SearchRequest{Query: "invalid json"},
			func(_ context.Context, batch []*ledgerv1.Transaction) error {
				return nil
			},
		)
		require.Error(t, err)

		// Test SearchEntries with invalid query
		err = transactionBusiness.SearchEntries(
			ctx,
			&commonv1.SearchRequest{Query: "invalid json"},
			func(_ context.Context, batch []*ledgerv1.TransactionEntry) error {
				return nil
			},
		)
		require.Error(t, err)

		// Test SearchTransactions with consumer error
		consumerErr := errors.New("consumer failed")
		err = transactionBusiness.SearchTransactions(
			ctx,
			searchReq,
			func(_ context.Context, batch []*ledgerv1.Transaction) error {
				return consumerErr
			},
		)
		require.Error(t, err)
		assert.Equal(t, consumerErr, err)

		// Test SearchEntries with consumer error
		err = transactionBusiness.SearchEntries(
			ctx,
			searchReq,
			func(_ context.Context, batch []*ledgerv1.TransactionEntry) error {
				return consumerErr
			},
		)
		require.Error(t, err)
		assert.Equal(t, consumerErr, err)

		// Test SearchTransactions with nil query (triggers default)
		var emptyQueryTxns []*ledgerv1.Transaction
		err = transactionBusiness.SearchTransactions(
			ctx,
			&commonv1.SearchRequest{},
			func(_ context.Context, batch []*ledgerv1.Transaction) error {
				emptyQueryTxns = append(emptyQueryTxns, batch...)
				return nil
			},
		)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(emptyQueryTxns), 1)

		// Test UpdateTransaction with non-existent ID
		_, err = transactionBusiness.UpdateTransaction(ctx, &ledgerv1.UpdateTransactionRequest{
			Id: "non-existent-txn-update",
		})
		require.Error(t, err)
	})
}

func (ts *TransactionBusinessSuite) TestValidationErrors() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Test missing reference
		txn, err := transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Currency: "USD",
		})
		require.Error(t, err)
		assert.Nil(t, txn)
		assert.Contains(t, err.Error(), "reference is required")

		// Test missing currency
		txn, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id: "no-currency-txn",
		})
		require.Error(t, err)
		assert.Nil(t, txn)
		assert.Contains(t, err.Error(), "currency is required")

		// Test no entries
		txn, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "no-entries-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
		})
		require.Error(t, err)
		assert.Nil(t, txn)

		// Test zero amount entry
		txn, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "zero-amount-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 0}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 0}},
			},
		})
		require.Error(t, err)
		assert.Nil(t, txn)
		assert.Contains(t, err.Error(), "zero amount")

		// Test reverse empty ID
		_, err = transactionBusiness.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction ID is required")

		// Test update empty ID
		_, err = transactionBusiness.UpdateTransaction(ctx, &ledgerv1.UpdateTransactionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction ID is required")

		// Test get empty ID
		_, err = transactionBusiness.GetTransaction(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction ID is required")
	})
}

func (ts *TransactionBusinessSuite) TestReverseNonNormalAndClearance() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create a reservation transaction
		_, err := transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "non-reversible-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_RESERVATION,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		// Try to reverse it — should fail for non-NORMAL type
		_, err = transactionBusiness.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
			Id: "non-reversible-txn",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not reversible")

		// Create an uncleared transaction for clearance test
		_, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "clearance-test-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Cleared:  false,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		// Update with clearance timestamp
		clearanceTime := time.Now().UTC().Format("2006-01-02T15:04:05.999999999")
		updatedTxn, err := transactionBusiness.UpdateTransaction(ctx, &ledgerv1.UpdateTransactionRequest{
			Id:        "clearance-test-txn",
			ClearedAt: clearanceTime,
		})
		require.NoError(t, err)
		assert.True(t, updatedTxn.GetCleared())

		// Test processClearanceUpdate with invalid time format
		_, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "clearance-invalid-time-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Cleared:  false,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		_, err = transactionBusiness.UpdateTransaction(ctx, &ledgerv1.UpdateTransactionRequest{
			Id:        "clearance-invalid-time-txn",
			ClearedAt: "not-a-valid-time",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing time")

		// Test IsConflict — create a transaction and check conflict detection
		_, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "conflict-check-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		// IsConflict with different entries (should be a conflict)
		diffModel := models.TransactionFromAPI(ctx, &ledgerv1.Transaction{
			Id:           "conflict-check-txn",
			CurrencyCode: "USD",
			Type:         ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 999}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 999}},
			},
		})

		isConflict, err := transactionBusiness.IsConflict(ctx, diffModel)
		require.NoError(t, err)
		assert.True(t, isConflict, "Different entries should be a conflict")

		// IsConflict with non-existent transaction ID
		nonExistent := &models.Transaction{}
		nonExistent.ID = "non-existent-txn-for-conflict"
		_, err = transactionBusiness.IsConflict(ctx, nonExistent)
		require.Error(t, err)

		// Test sort comparator branches: entries with same account but different Credit
		// (covers the ei.Credit != ej.Credit and amount comparison branches in sort.Slice)
		_, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "same-account-multi-entry-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "asset-account", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 200}},
				{AccountId: "asset-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "income-account", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)
	})
}

func (ts *TransactionBusinessSuite) TestUpdateTransaction() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create a transaction first
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "update-test-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 300, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 300, Nanos: 0},
				},
			},
		}

		_, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.NoError(t, err, "Error creating transaction")

		// Update the transaction data
		updateTransactionReq := &ledgerv1.UpdateTransactionRequest{
			Id: "update-test-transaction",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"reference": {Kind: &structpb.Value_StringValue{StringValue: "Updated reference"}},
					"category":  {Kind: &structpb.Value_StringValue{StringValue: "Payment"}},
				},
			},
		}

		updatedTransaction, err := transactionBusiness.UpdateTransaction(ctx, updateTransactionReq)
		require.NoError(t, err, "Error updating transaction")
		require.NotNil(t, updatedTransaction, "Updated transaction should not be nil")

		// Verify the update
		assert.Equal(t, "Updated reference", updatedTransaction.GetData().GetFields()["reference"].GetStringValue())
		assert.Equal(t, "Payment", updatedTransaction.GetData().GetFields()["category"].GetStringValue())
	})
}

func (ts *TransactionBusinessSuite) TestDuplicateTransactionExactDuplicate() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create first transaction
		createTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "duplicate-test-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
			TransactedAt: time.Now().UTC().Format(time.RFC3339),
			Cleared:      true,
		}

		// Create the same transaction twice - should be idempotent
		firstTransaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.NoError(t, err, "Error creating first transaction")
		require.NotNil(t, firstTransaction, "First transaction should be created")

		secondTransaction, err := transactionBusiness.CreateTransaction(ctx, createTransactionReq)
		require.NoError(t, err, "Error creating duplicate transaction")
		require.NotNil(t, secondTransaction, "Duplicate transaction should be returned")

		// Should return the same transaction (idempotent behavior)
		assert.Equal(t, firstTransaction.GetId(), secondTransaction.GetId(), "Should return same transaction ID")
		assert.Equal(
			t,
			firstTransaction.GetCurrencyCode(),
			secondTransaction.GetCurrencyCode(),
			"Should have same currency",
		)
		assert.Len(t, secondTransaction.GetEntries(), 2, "Should have 2 entries")
	})
}

func (ts *TransactionBusinessSuite) TestDuplicateTransactionConflictingEntries() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create first transaction
		firstTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "conflicting-duplicate-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
		}

		firstTransaction, err := transactionBusiness.CreateTransaction(ctx, firstTransactionReq)
		require.NoError(t, err, "Error creating first transaction")
		require.NotNil(t, firstTransaction, "First transaction should be created")

		// Try to create transaction with same ID but different entries (conflict)
		conflictingTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "conflicting-duplicate-transaction", // Same ID
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 200, Nanos: 0}, // Different amount
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 200, Nanos: 0}, // Different amount
				},
			},
		}

		conflictingTransaction, err := transactionBusiness.CreateTransaction(ctx, conflictingTransactionReq)
		require.Error(t, err, "Should fail with conflicting transaction")
		assert.Nil(t, conflictingTransaction, "Conflicting transaction should not be created")
		assert.Contains(t, err.Error(), "conflict", "Error should mention conflict")
	})
}

func (ts *TransactionBusinessSuite) TestDuplicateTransactionConflictingAccounts() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create first transaction
		firstTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "conflicting-accounts-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
		}

		firstTransaction, err := transactionBusiness.CreateTransaction(ctx, firstTransactionReq)
		require.NoError(t, err, "Error creating first transaction")
		require.NotNil(t, firstTransaction, "First transaction should be created")

		// Create additional accounts for conflicting test
		createAdditionalAccountReq := &ledgerv1.CreateAccountRequest{
			Id:       "additional-account",
			LedgerId: ts.assetLedger.ID,
			Currency: "USD",
		}
		_, err = resources.AccountBusiness.CreateAccount(ctx, createAdditionalAccountReq)
		require.NoError(t, err, "Error creating additional account")

		// Try to create transaction with same ID but different accounts
		conflictingAccountsReq := &ledgerv1.CreateTransactionRequest{
			Id:       "conflicting-accounts-transaction", // Same ID
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "additional-account", // Different account
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
		}

		conflictingTransaction, err := transactionBusiness.CreateTransaction(ctx, conflictingAccountsReq)
		require.Error(t, err, "Should fail with conflicting accounts")
		assert.Nil(t, conflictingTransaction, "Conflicting transaction should not be created")
		assert.Contains(t, err.Error(), "conflict", "Error should mention conflict")
	})
}

func (ts *TransactionBusinessSuite) TestDuplicateTransactionDifferentEntryOrder() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create first transaction
		firstTransactionReq := &ledgerv1.CreateTransactionRequest{
			Id:       "order-test-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry1",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry2",
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
		}

		firstTransaction, err := transactionBusiness.CreateTransaction(ctx, firstTransactionReq)
		require.NoError(t, err, "Error creating first transaction")
		require.NotNil(t, firstTransaction, "First transaction should be created")

		// Create transaction with same entries but different order
		reversedOrderReq := &ledgerv1.CreateTransactionRequest{
			Id:       "order-test-transaction", // Same ID
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "entry2", // Different entry ID order
					AccountId: "income-account",
					Credit:    true,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
				{
					Id:        "entry1", // Different entry ID order
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
				},
			},
		}

		reversedOrderTransaction, err := transactionBusiness.CreateTransaction(ctx, reversedOrderReq)
		require.NoError(t, err, "Should succeed with same entries in different order")
		require.NotNil(t, reversedOrderTransaction, "Transaction should be returned")

		// Should return the same transaction (idempotent behavior)
		assert.Equal(t, firstTransaction.GetId(), reversedOrderTransaction.GetId(), "Should return same transaction ID")
	})
}

func (ts *TransactionBusinessSuite) TestDuplicateReservationTransaction() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create first reservation transaction
		reservationReq := &ledgerv1.CreateTransactionRequest{
			Id:       "duplicate-reservation-transaction",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_RESERVATION,
			Entries: []*ledgerv1.TransactionEntry{
				{
					Id:        "reservation-entry",
					AccountId: "asset-account",
					Credit:    false,
					Amount:    &money.Money{CurrencyCode: "USD", Units: 500, Nanos: 0},
				},
			},
		}

		firstReservation, err := transactionBusiness.CreateTransaction(ctx, reservationReq)
		require.NoError(t, err, "Error creating first reservation transaction")
		require.NotNil(t, firstReservation, "First reservation should be created")

		// Try to create the same reservation transaction again
		duplicateReservation, err := transactionBusiness.CreateTransaction(ctx, reservationReq)
		require.NoError(t, err, "Error creating duplicate reservation transaction")
		require.NotNil(t, duplicateReservation, "Duplicate reservation should be returned")

		// Should return the same reservation (idempotent behavior)
		assert.Equal(t, firstReservation.GetId(), duplicateReservation.GetId(), "Should return same reservation ID")
		assert.Equal(
			t,
			ledgerv1.TransactionType_RESERVATION,
			duplicateReservation.GetType(),
			"Should be reservation type",
		)
	})
}

// setupConcurrentTestAccounts creates additional accounts needed for concurrent testing.
func (ts *TransactionBusinessSuite) setupConcurrentTestAccounts(
	ctx context.Context,
	resources *tests.ServiceResources,
	numAccounts int,
) {
	for i := range numAccounts {
		accountReq := &ledgerv1.CreateAccountRequest{
			Id:       fmt.Sprintf("concurrent-account-%d", i),
			LedgerId: ts.assetLedger.ID,
			Currency: "USD",
		}
		_, err := resources.AccountBusiness.CreateAccount(ctx, accountReq)
		ts.Require().NoError(err, "Error creating concurrent account %d", i)
	}
}

// createTestTransaction creates a test transaction request with the specified parameters.
func (ts *TransactionBusinessSuite) createTestTransaction(
	transactionID string,
	accountIndex int,
	amount int64,
) *ledgerv1.CreateTransactionRequest {
	return &ledgerv1.CreateTransactionRequest{
		Id:       transactionID,
		Currency: "USD",
		Type:     ledgerv1.TransactionType_NORMAL,
		Entries: []*ledgerv1.TransactionEntry{
			{
				Id:        "entry1",
				AccountId: fmt.Sprintf("concurrent-account-%d", accountIndex),
				Credit:    false,
				Amount:    &money.Money{CurrencyCode: "USD", Units: amount, Nanos: 0},
			},
			{
				Id:        "entry2",
				AccountId: "income-account",
				Credit:    true,
				Amount:    &money.Money{CurrencyCode: "USD", Units: amount, Nanos: 0},
			},
		},
	}
}

// runConcurrentTransactionTest executes the concurrent transaction logic for a single goroutine.
func (ts *TransactionBusinessSuite) runConcurrentTransactionTest(
	ctx context.Context,
	transactionBusiness *tests.ServiceResources,
	goroutineID, numTransactions int,
	results *concurrentTestResults,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for j := range numTransactions {
		transactionID := fmt.Sprintf("concurrent-txn-%d-%d", goroutineID, j)

		// Create transaction with unique ID
		createReq := ts.createTestTransaction(transactionID, 0, 100)

		// First attempt - should succeed
		transaction, err := transactionBusiness.TransactionBusiness.CreateTransaction(ctx, createReq)
		if err != nil {
			results.mu.Lock()
			results.errorCount++
			results.mu.Unlock()
			continue
		}

		// Second attempt with same ID - should be idempotent
		duplicateTransaction, err := transactionBusiness.TransactionBusiness.CreateTransaction(ctx, createReq)
		if err != nil {
			ts.handleDuplicateError(err, results)
			continue
		}

		// Verify idempotent behavior
		results.verifyIdempotentBehavior(transactionID, transaction, duplicateTransaction)

		// Test conflicting transaction with same ID
		conflictReq := ts.createTestTransaction(transactionID, 1, 200) // Different account and amount
		_, conflictErr := transactionBusiness.TransactionBusiness.CreateTransaction(ctx, conflictReq)
		if conflictErr != nil && strings.Contains(conflictErr.Error(), "conflict") {
			results.mu.Lock()
			results.conflictCount++
			results.mu.Unlock()
		}
	}
}

// handleDuplicateError processes errors from duplicate transaction attempts.
func (ts *TransactionBusinessSuite) handleDuplicateError(
	err error,
	results *concurrentTestResults,
) {
	results.mu.Lock()
	defer results.mu.Unlock()

	if strings.Contains(err.Error(), "conflict") {
		results.conflictCount++
	} else {
		results.errorCount++
	}
}

// concurrentTestResults tracks the results of concurrent transaction testing.
type concurrentTestResults struct {
	mu                  sync.Mutex
	successCount        int
	duplicateCount      int
	conflictCount       int
	errorCount          int
	createdTransactions map[string]*ledgerv1.Transaction
}

// verifyIdempotentBehavior checks that duplicate transactions return the same result.
func (r *concurrentTestResults) verifyIdempotentBehavior(
	transactionID string,
	transaction, duplicateTransaction *ledgerv1.Transaction,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if transaction.GetId() == duplicateTransaction.GetId() {
		if _, exists := r.createdTransactions[transactionID]; !exists {
			r.successCount++
			r.createdTransactions[transactionID] = transaction
		}
		r.duplicateCount++ // Count successful idempotent calls
	} else {
		// This shouldn't happen - different transaction returned for same ID
		r.errorCount++
	}
}

func (ts *TransactionBusinessSuite) TestConcurrentTransactionStress() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ts.setupFixtures(ctx, resources)

		transactionBusiness := resources.TransactionBusiness

		// Create additional accounts for concurrent testing
		ts.setupConcurrentTestAccounts(ctx, resources, 10)

		// Test parameters
		numGoroutines := 50
		numTransactions := 10
		var wg sync.WaitGroup

		results := &concurrentTestResults{
			createdTransactions: make(map[string]*ledgerv1.Transaction),
		}

		// Launch multiple goroutines creating transactions concurrently
		for i := range numGoroutines {
			wg.Add(1)
			go ts.runConcurrentTransactionTest(ctx, resources, i, numTransactions, results, &wg)
		}

		wg.Wait()

		// Verify results
		expectedTransactions := numGoroutines * numTransactions
		t.Logf("Concurrent transaction test results:")
		t.Logf("- Expected unique transactions: %d", expectedTransactions)
		t.Logf("- Successful transactions: %d", results.successCount)
		t.Logf("- Duplicate (idempotent) calls: %d", results.duplicateCount)
		t.Logf("- Conflicting transactions rejected: %d", results.conflictCount)
		t.Logf("- Other errors: %d", results.errorCount)

		assert.Equal(t, expectedTransactions, results.successCount, "All unique transactions should be created")
		assert.Equal(
			t,
			expectedTransactions,
			results.duplicateCount,
			"All duplicate calls should be handled idempotently",
		)
		assert.Positive(t, results.conflictCount, "Conflicting transactions should be rejected")
		assert.Equal(t, 0, results.errorCount, "There should be no unexpected errors")

		// Verify all transactions were actually created
		for transactionID, transaction := range results.createdTransactions {
			retrieved, err := transactionBusiness.GetTransaction(ctx, transactionID)
			require.NoError(t, err, "Should be able to retrieve created transaction %s", transactionID)
			assert.Equal(t, transaction.GetId(), retrieved.GetId(), "Retrieved transaction should match")
			assert.Len(t, retrieved.GetEntries(), 2, "Transaction should have 2 entries")
		}
	})
}
