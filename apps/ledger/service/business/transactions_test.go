package business_test

import (
	"context"
	"sync"
	"testing"
	"time"

	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"
)

// assertDecEqual compares two decimalx.Decimal values using the Equal method
// instead of reflect.DeepEqual (which fails due to pointer-wrapped internals).
func assertDecEqual(t *testing.T, expected, actual decimalx.Decimal, msgAndArgs ...interface{}) {
	t.Helper()
	assert.True(t, expected.Equal(actual), append([]interface{}{
		"expected %s but got %s", expected, actual,
	}, msgAndArgs...)...)
}

type TransactionsModelSuite struct {
	tests.BaseTestSuite
	ledger *models.Ledger
}

func (ts *TransactionsModelSuite) setupFixtures(ctx context.Context, resources *tests.ServiceResources) {
	// Create test ledgers using business layer

	// Create first ledger (Asset)
	createLedgerReq1 := &ledgerv1.CreateLedgerRequest{
		Id:   "test-ledger-asset",
		Type: ledgerv1.LedgerType_ASSET,
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"name": {Kind: &structpb.Value_StringValue{StringValue: "Test Asset Ledger"}},
			},
		},
	}

	ledger1, err := resources.LedgerBusiness.CreateLedger(ctx, createLedgerReq1)
	ts.Require().NoError(err, "Unable to create asset ledger")

	ts.ledger = &models.Ledger{
		BaseModel: data.BaseModel{ID: ledger1.GetId()},
		Type:      ledger1.GetType().String(),
	}

	// Create second ledger (Income)
	createLedgerReq2 := &ledgerv1.CreateLedgerRequest{
		Id:   "test-ledger-income",
		Type: ledgerv1.LedgerType_INCOME,
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"name": {Kind: &structpb.Value_StringValue{StringValue: "Test Income Ledger"}},
			},
		},
	}

	ledger2, err := resources.LedgerBusiness.CreateLedger(ctx, createLedgerReq2)
	ts.Require().NoError(err, "Unable to create income ledger")

	// Create test accounts using business layer
	accounts := []struct {
		id         string
		ledgerID   string
		ledgerType string
	}{
		{"a1", ledger1.GetId(), models.LedgerTypeAsset},
		{"a2", ledger2.GetId(), models.LedgerTypeIncome},
		{"a3", ledger1.GetId(), models.LedgerTypeAsset},
		{"a4", ledger1.GetId(), models.LedgerTypeAsset},
		{"b1", ledger1.GetId(), models.LedgerTypeAsset},
		{"b2", ledger1.GetId(), models.LedgerTypeAsset},
	}

	for _, acc := range accounts {
		createAccReq := &ledgerv1.CreateAccountRequest{
			Id:       acc.id,
			LedgerId: acc.ledgerID,
			Currency: "UGX",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"name": {Kind: &structpb.Value_StringValue{StringValue: "Test Account " + acc.id}},
				},
			},
		}

		_, err = resources.AccountBusiness.CreateAccount(ctx, createAccReq)
		ts.Require().NoError(err, "Unable to create account %s", acc.id)
	}
}

func (ts *TransactionsModelSuite) TestIsZeroSum() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		timeNow := time.Now().UTC()

		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t001"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
				{
					AccountID: "a2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
			},
		}
		valid := transaction.IsZeroSum()
		assert.True(t, valid, "Transaction should be zero summed")

		transaction.Entries[0].Amount = decimalx.NewFromInt64(200).Ptr()
		valid = transaction.IsZeroSum()
		assert.False(t, valid, "Transaction should not be zero summed")
	})
}

func (ts *TransactionsModelSuite) TestIsTrueDrCr() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, _ *definition.DependencyOption) {
		timeNow := time.Now().UTC()
		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t001"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(30).Ptr(),
				},
				{
					AccountID: "a2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(30).Ptr(),
				},
			},
		}
		valid := transaction.IsTrueDrCr()
		assert.True(t, valid, "Transaction should contain one dr and other cr entries")

		transaction.Entries[0].Credit = true
		valid = transaction.IsTrueDrCr()
		assert.False(t, valid, "Transaction should fail DrCr test")
	})
}

func (ts *TransactionsModelSuite) TestIsConflict() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		timeNow := time.Now().UTC()

		txnBusiness := res.TransactionBusiness
		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t0015"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
				{
					AccountID: "a2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
			},
		}
		done, err := txnBusiness.Transact(ctx, transaction)
		require.NoError(t, err)
		require.NotNil(t, done, "Transaction should be created")

		conflicts, err := txnBusiness.IsConflict(ctx, transaction)
		require.NoError(t, err, "Error while checking for conflict transaction")
		assert.False(t, conflicts, "Transaction should not conflict")

		transaction = &models.Transaction{
			BaseModel:    data.BaseModel{ID: "t0015"},
			Currency:     "UGX",
			TransactedAt: timeNow,
			ClearedAt:    timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(50).Ptr(),
				},
				{
					AccountID: "a2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(50).Ptr(),
				},
			},
		}

		conflicts, err = txnBusiness.IsConflict(ctx, transaction)
		require.NoError(t, err, "Error while checking for conflicting transaction")
		assert.True(t, conflicts, "Transaction should conflict since amounts are different from first received")

		transaction = &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t0015"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "b1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
				{
					AccountID: "b2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
			},
		}
		conflicts, err = txnBusiness.IsConflict(ctx, transaction)
		require.NoError(t, err, "Error while checking for conflicting transaction")
		assert.True(t, conflicts, "Transaction should conflict since accounts are different from first received")
	})
}

func (ts *TransactionsModelSuite) TestTransact() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		timeNow := time.Now().UTC()
		txnBusiness := res.TransactionBusiness

		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t003"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
				{
					AccountID: "a2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
			},
			Data: map[string]interface{}{
				"tag1": "val1",
				"tag2": "val2",
			},
		}
		done, err := txnBusiness.Transact(ctx, transaction)
		require.NoError(t, err)
		require.NotNil(t, done, "Transaction should be created")

		_, getErr := txnBusiness.GetTransaction(ctx, "t003")
		require.NoError(t, getErr, "Error while checking for existing transaction")
	})
}

func (ts *TransactionsModelSuite) TestReserveTransaction() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		accountRepo := res.AccountRepository
		txnBusiness := res.TransactionBusiness

		initialAcc, err := accountRepo.GetByID(ctx, "a3")
		require.NoError(t, err)

		timeNow := time.Now().UTC()
		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t031"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_RESERVATION.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a3",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(98).Ptr(),
				},
			},
			Data: map[string]interface{}{
				"tag1": "val1",
				"tag2": "val2",
			},
		}
		done, err := txnBusiness.Transact(ctx, transaction)
		require.NoError(t, err)
		require.NotNil(t, done, "Transaction should be created")

		exists, err := txnBusiness.GetTransaction(ctx, "t031")
		require.NoError(t, err, "Error while checking for existing transaction")
		assert.Equal(t, "t031", exists.GetId(), "Transaction should exist")

		finalAcc, err := accountRepo.GetByID(ctx, "a3")
		require.NoError(t, err)

		assertDecEqual(
			t,
			decimalx.NewFromInt64(0),
			decimalx.DerefOr(finalAcc.Balance, decimalx.Zero()).
				Sub(decimalx.DerefOr(initialAcc.Balance, decimalx.Zero())),
			"Reservation Balance should be consistent",
		)

		assertDecEqual(
			t,
			decimalx.NewFromInt64(98),
			decimalx.DerefOr(finalAcc.ReservedBalance, decimalx.Zero()),
			"reserved balance should be consistent",
		)
	})
}

func (ts *TransactionsModelSuite) TestTransactBalanceCheck() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		accountRepo := res.AccountRepository
		txnBusiness := res.TransactionBusiness

		initialAccMap, err := accountRepo.ListByID(ctx, "a3", "a4")
		require.NoError(t, err)

		timeNow := time.Now().UTC()
		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t008"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a3",
					Amount:    decimalx.NewFromInt64(51).Ptr(),
					Credit:    false,
				},
				{
					AccountID: "a4",
					Amount:    decimalx.NewFromInt64(51).Ptr(),
					Credit:    true,
				},
			},
			Data: map[string]interface{}{
				"tag1": "transaction balance check",
			},
		}
		done, err1 := txnBusiness.Transact(ctx, transaction)
		require.NoError(t, err1)
		require.NotNil(t, done, "Transaction should be created")

		finalAccMap, err2 := accountRepo.ListByID(ctx, "a3", "a4")
		require.NoError(t, err2)

		assertDecEqual(
			t,
			decimalx.NewFromInt64(51),
			decimalx.DerefOr(finalAccMap["a3"].Balance, decimalx.Zero()).
				Sub(decimalx.DerefOr(initialAccMap["a3"].Balance, decimalx.Zero())),
			"Debited Balance should be equal",
		)
		negFiftyOne := decimalx.NewFromInt64(51).Neg()
		assertDecEqual(
			t,
			negFiftyOne,
			decimalx.DerefOr(finalAccMap["a4"].Balance, decimalx.Zero()).
				Sub(decimalx.DerefOr(initialAccMap["a4"].Balance, decimalx.Zero())),
			"Credited Balance should be equal",
		)
	})
}

func (ts *TransactionsModelSuite) TestDuplicateTransactions() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		txnBusiness := res.TransactionBusiness

		timeNow := time.Now().UTC()

		var wg sync.WaitGroup
		wg.Add(5)
		for i := 1; i <= 5; i++ {
			go func(txnID string) {
				// Create a fresh transaction instance for each goroutine
				transactionCopy := &models.Transaction{
					BaseModel:       data.BaseModel{ID: txnID},
					Currency:        "UGX",
					TransactionType: ledgerv1.TransactionType_NORMAL.String(),
					TransactedAt:    timeNow,
					ClearedAt:       timeNow,
					Entries: []*models.TransactionEntry{
						{
							AccountID: "a1",
							Credit:    false,
							Amount:    decimalx.NewFromInt64(100).Ptr(),
						},
						{
							AccountID: "a2",
							Credit:    true,
							Amount:    decimalx.NewFromInt64(100).Ptr(),
						},
					},
				}

				trxn, err := txnBusiness.Transact(ctx, transactionCopy)
				if err != nil {
					t.Logf("Transaction creation failed: %v", err)
				}
				assert.NotNil(t, trxn, "Transaction creation should be success")
				wg.Done()
			}("t005")
		}
		wg.Wait()

		exists, err := txnBusiness.GetTransaction(ctx, "t005")
		require.NoError(t, err, "Error while checking for existing transaction")
		assert.Equal(t, "t005", exists.GetId(), "Transaction should exist")
	})
}

func (ts *TransactionsModelSuite) TestTransactionReversaL() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		txnBusiness := res.TransactionBusiness

		timeNow := time.Now().UTC()

		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t053"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
				{
					AccountID: "a2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
			},
		}

		trxn, err := txnBusiness.Transact(ctx, transaction)
		require.NoError(t, err)
		require.NotNil(t, trxn, "Transaction creation should be success")

		reversal, err := txnBusiness.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
			Id: trxn.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, reversal, "Transaction reversal should be success")

		_, getErr := txnBusiness.GetTransaction(ctx, "t053_REVERSAL")
		require.NoError(t, getErr, "Error while checking for existing transaction")
	})
}

func (ts *TransactionsModelSuite) TestUnClearedTransactions() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		accountRepo := res.AccountRepository
		txnBusiness := res.TransactionBusiness

		initialAccMap, err := accountRepo.ListByID(ctx, "b1", "b2")
		require.NoError(t, err)

		timeNow := time.Now().UTC()

		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t051"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "b1",
					Credit:    false,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
				{
					AccountID: "b2",
					Credit:    true,
					Amount:    decimalx.NewFromInt64(100).Ptr(),
				},
			},
		}

		done, err1 := txnBusiness.Transact(ctx, transaction)
		require.NoError(t, err1)
		require.NotNil(t, done, "Transaction should be created")

		finalAccMap, err2 := accountRepo.ListByID(ctx, "b1", "b2")
		require.NoError(t, err2)

		assertDecEqual(
			t,
			decimalx.Zero(),
			decimalx.DerefOr(finalAccMap["b1"].Balance, decimalx.Zero()).
				Sub(decimalx.DerefOr(initialAccMap["b1"].Balance, decimalx.Zero())),
			"Debited Balance should be equal",
		)
		assertDecEqual(
			t,
			decimalx.Zero(),
			decimalx.DerefOr(finalAccMap["b2"].Balance, decimalx.Zero()).
				Sub(decimalx.DerefOr(initialAccMap["b2"].Balance, decimalx.Zero())),
			"Credited Balance should be equal",
		)

		assertDecEqual(
			t,
			decimalx.NewFromInt64(100),
			decimalx.DerefOr(finalAccMap["b1"].UnClearedBalance, decimalx.Zero()),
			"b1 Uncleared balance should be equal",
		)
		negHundred := decimalx.NewFromInt64(100).Neg()
		assertDecEqual(
			t,
			negHundred,
			decimalx.DerefOr(finalAccMap["b2"].UnClearedBalance, decimalx.Zero()),
			"b2 Uncleared balance should be equal",
		)

		assertDecEqual(
			t,
			decimalx.Zero(),
			decimalx.DerefOr(finalAccMap["b1"].ReservedBalance, decimalx.Zero()),
			"b1 reserved balance should be zero",
		)
		assertDecEqual(
			t,
			decimalx.Zero(),
			decimalx.DerefOr(finalAccMap["b2"].ReservedBalance, decimalx.Zero()),
			"b2 reserved balance should be zero",
		)
	})
}

func (ts *TransactionsModelSuite) TestTransactWithBoundaryValues() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		txnBusiness := res.TransactionBusiness

		timeNow := time.Now().UTC()

		// In-boundary value transaction
		boundaryValue, _ := decimalx.NewFromString("9223372036854775807.999999999") // Max +ve for 2^64
		transaction := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "t004"},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{
					AccountID: "a3",
					Credit:    false,
					Amount:    boundaryValue.Ptr(),
				},
				{
					AccountID: "a4",
					Credit:    true,
					Amount:    boundaryValue.Ptr(),
				},
			},
			Data: map[string]interface{}{
				"tag1": "val1",
				"tag2": "val2",
			},
		}
		done, _ := txnBusiness.Transact(ctx, transaction)
		require.NotNil(t, done, "Transaction should be created")
		exists, err := txnBusiness.GetTransaction(ctx, "t004")
		require.NoError(t, err, "Error while checking for existing transaction")
		assert.Equal(t, "t004", exists.GetId(), "Transaction should exist")

		// Out-of-boundary value transaction
		// Note: Not able write test case for out of boundary value here,
		// due to overflow error while compilation.
		// The test case is written in `package controllers` using JSON
	})
}

func TestTransactionsModelSuite(t *testing.T) {
	suite.Run(t, new(TransactionsModelSuite))
}
