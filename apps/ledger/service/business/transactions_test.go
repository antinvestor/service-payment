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
	"fmt"
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

// TestIdempotencyKeyConcurrent verifies that under N parallel posts carrying
// the same idempotency_key (but distinct Transaction.IDs and identical
// entries), exactly one row reaches the database and every caller is
// returned a transaction tied to that row. This is the core durability
// guarantee that lets webhook redelivery, queue replays and at-least-once
// transports replay safely.
func (ts *TransactionsModelSuite) TestIdempotencyKeyConcurrent() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		const parallelism = 10
		idemKey := "webhook:mtn:abc123"
		timeNow := time.Now().UTC()

		var (
			wg        sync.WaitGroup
			mu        sync.Mutex
			returnIDs = make(map[string]struct{}, parallelism)
			errs      []error
		)
		wg.Add(parallelism)
		for i := range parallelism {
			go func(i int) {
				defer wg.Done()

				// Distinct Transaction.ID per goroutine; the idempotency_key
				// is the only thing tying them together.
				txn := &models.Transaction{
					BaseModel: data.BaseModel{
						ID: ts.uniqueTxnID("idem_concurrent", i),
					},
					Currency:        "UGX",
					TransactionType: ledgerv1.TransactionType_NORMAL.String(),
					IdempotencyKey:  idemKey,
					TransactedAt:    timeNow,
					ClearedAt:       timeNow,
					Entries: []*models.TransactionEntry{
						{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
						{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
					},
				}

				out, err := res.TransactionBusiness.Transact(ctx, txn)
				mu.Lock()
				if err != nil {
					errs = append(errs, err)
				} else {
					returnIDs[out.GetID()] = struct{}{}
				}
				mu.Unlock()
			}(i)
		}
		wg.Wait()

		require.Empty(t, errs, "no goroutine should fail; idempotent replay must succeed for all")
		require.Len(t, returnIDs, 1,
			"all goroutines must converge on a single canonical transaction id; got %d distinct ids", len(returnIDs))

		// Verify exactly one row carries the idempotency_key.
		stored, err := res.TransactionRepository.GetByIdempotencyKey(ctx, idemKey)
		require.NoError(t, err)
		require.NotNil(t, stored)
		_, ok := returnIDs[stored.GetID()]
		require.True(t, ok, "stored transaction id must match the id returned to callers")
	})
}

// TestIdempotencyKeyConflictDifferentEntries verifies that reusing an
// idempotency_key with materially different entries is rejected — the
// system must not silently return the prior posting when the caller's
// intent has changed.
func (ts *TransactionsModelSuite) TestIdempotencyKeyConflictDifferentEntries() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		idemKey := "webhook:mtn:conflict-test"
		timeNow := time.Now().UTC()

		first := &models.Transaction{
			BaseModel:       data.BaseModel{ID: ts.uniqueTxnID("idem_conflict_a", 0)},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			IdempotencyKey:  idemKey,
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, first)
		require.NoError(t, err)

		// Same idempotency_key, different amount — should be rejected.
		second := &models.Transaction{
			BaseModel:       data.BaseModel{ID: ts.uniqueTxnID("idem_conflict_b", 0)},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			IdempotencyKey:  idemKey,
			TransactedAt:    timeNow,
			ClearedAt:       timeNow,
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(250).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(250).Ptr()},
			},
		}
		_, err = res.TransactionBusiness.Transact(ctx, second)
		require.Error(t, err, "reuse of idempotency_key with different entries must be rejected")
	})
}

// TestConcurrentPostingsBalanceIntegrity verifies that N parallel posts
// each touching the same pair of accounts produce a final balance equal
// to the deterministic sum of their amounts. Exercises the atomic Create
// override and the LATERAL balance derivation together: if either the
// posting was non-atomic (orphan transactions) or the balance read missed
// concurrently committed rows, the final assertion would fail.
func (ts *TransactionsModelSuite) TestConcurrentPostingsBalanceIntegrity() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		const (
			parallelism = 20
			amountEach  = int64(100)
		)
		expectedDelta := decimalx.NewFromInt64(amountEach * parallelism)

		initialMap, err := res.AccountRepository.ListByID(ctx, "a1")
		require.NoError(t, err)
		initialBalance := decimalx.DerefOr(initialMap["a1"].Balance, decimalx.Zero())

		timeNow := time.Now().UTC()
		var (
			wg   sync.WaitGroup
			mu   sync.Mutex
			errs []error
		)
		wg.Add(parallelism)
		for i := range parallelism {
			go func(i int) {
				defer wg.Done()
				txn := &models.Transaction{
					BaseModel:       data.BaseModel{ID: ts.uniqueTxnID("concurrent_balance", i)},
					Currency:        "UGX",
					TransactionType: ledgerv1.TransactionType_NORMAL.String(),
					IdempotencyKey:  ts.uniqueTxnID("concurrent_balance_key", i),
					TransactedAt:    timeNow,
					ClearedAt:       timeNow,
					Entries: []*models.TransactionEntry{
						{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(amountEach).Ptr()},
						{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(amountEach).Ptr()},
					},
				}
				if _, e := res.TransactionBusiness.Transact(ctx, txn); e != nil {
					mu.Lock()
					errs = append(errs, e)
					mu.Unlock()
				}
			}(i)
		}
		wg.Wait()
		require.Empty(t, errs, "all posts should succeed under concurrent load")

		finalMap, err := res.AccountRepository.ListByID(ctx, "a1")
		require.NoError(t, err)
		finalBalance := decimalx.DerefOr(finalMap["a1"].Balance, decimalx.Zero())
		delta := finalBalance.Sub(initialBalance)
		assertDecEqual(t, expectedDelta, delta,
			"final balance must equal initial + sum of all concurrent posts")
	})
}

// uniqueTxnID returns a compact per-run unique ID under the varchar(50)
// constraint on transactions.id. Stable within a goroutine via the i suffix.
func (ts *TransactionsModelSuite) uniqueTxnID(prefix string, i int) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano()%1_000_000, i)
}

// TestStatusCreatedClearedYieldsPosted verifies the create path: ClearedAt
// non-zero translates to status=posted with PostedAt populated, both at
// creation time.
func (ts *TransactionsModelSuite) TestStatusCreatedClearedYieldsPosted() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		now := time.Now().UTC()
		txnID := ts.uniqueTxnID("status_cleared", 0)
		txn := &models.Transaction{
			BaseModel:       data.BaseModel{ID: txnID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    now,
			ClearedAt:       now,
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, txn)
		require.NoError(t, err)

		stored, err := res.TransactionRepository.GetByID(ctx, txnID)
		require.NoError(t, err)
		assert.Equal(t, models.TransactionStatusPosted, stored.Status)
		require.NotNil(t, stored.PostedAt)
		assert.False(t, stored.PostedAt.IsZero())
	})
}

// TestStatusCreatedUnclearedYieldsPending verifies that omitting ClearedAt
// at creation parks the transaction in 'pending' with no PostedAt.
func (ts *TransactionsModelSuite) TestStatusCreatedUnclearedYieldsPending() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		txnID := ts.uniqueTxnID("status_uncl", 0)
		txn := &models.Transaction{
			BaseModel:       data.BaseModel{ID: txnID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    time.Now().UTC(),
			// ClearedAt intentionally left zero.
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, txn)
		require.NoError(t, err)

		stored, err := res.TransactionRepository.GetByID(ctx, txnID)
		require.NoError(t, err)
		assert.Equal(t, models.TransactionStatusPending, stored.Status)
		assert.Nil(t, stored.PostedAt)
	})
}

// TestReversalMarksOriginalAndCarriesLineage proves the atomic reversal
// pathway: the new REVERSAL row carries reversed_transaction_id pointing
// back at the original, and the original's status flips from posted to
// reversed in the same DB transaction.
func (ts *TransactionsModelSuite) TestReversalMarksOriginalAndCarriesLineage() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		now := time.Now().UTC()
		origID := ts.uniqueTxnID("rev_orig", 0)
		original := &models.Transaction{
			BaseModel:       data.BaseModel{ID: origID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    now,
			ClearedAt:       now,
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, original)
		require.NoError(t, err)

		reversedAPI, err := res.TransactionBusiness.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
			Id: origID,
		})
		require.NoError(t, err)
		require.NotNil(t, reversedAPI)

		// Original is now reversed.
		afterOrig, err := res.TransactionRepository.GetByID(ctx, origID)
		require.NoError(t, err)
		assert.Equal(t, models.TransactionStatusReversed, afterOrig.Status,
			"original transaction must transition to 'reversed' after reversal commits")

		// REVERSAL row carries reversed_transaction_id back at the original.
		reversalID := reversedAPI.GetId()
		reversalStored, err := res.TransactionRepository.GetByID(ctx, reversalID)
		require.NoError(t, err)
		assert.Equal(t, models.TransactionStatusPosted, reversalStored.Status)
		require.NotNil(t, reversalStored.ReversedTransactionID,
			"reversal must carry reversed_transaction_id FK to the original")
		assert.Equal(t, origID, *reversalStored.ReversedTransactionID)
	})
}

// TestReversalRejectsNonPosted verifies that an uncleared (pending) or
// already-reversed transaction cannot be reversed; the system refuses
// rather than silently producing a meaningless offset.
func (ts *TransactionsModelSuite) TestReversalRejectsNonPosted() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		pendingID := ts.uniqueTxnID("rev_pend", 0)
		pending := &models.Transaction{
			BaseModel:       data.BaseModel{ID: pendingID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    time.Now().UTC(),
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(50).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(50).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, pending)
		require.NoError(t, err)

		_, err = res.TransactionBusiness.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
			Id: pendingID,
		})
		require.Error(t, err, "reversing a pending transaction must be rejected")

		// And the same after marking it reversed: double-reversal must fail.
		postedID := ts.uniqueTxnID("rev_dbl", 0)
		now := time.Now().UTC()
		posted := &models.Transaction{
			BaseModel:       data.BaseModel{ID: postedID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    now,
			ClearedAt:       now,
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(75).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(75).Ptr()},
			},
		}
		_, err = res.TransactionBusiness.Transact(ctx, posted)
		require.NoError(t, err)

		_, err = res.TransactionBusiness.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
			Id: postedID,
		})
		require.NoError(t, err)

		_, err = res.TransactionBusiness.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
			Id: postedID,
		})
		require.Error(t, err, "second reversal on an already-reversed transaction must be rejected")
	})
}

// TestVoidPendingTransaction proves a pending transaction can be voided
// and the row carries the correct terminal status plus a non-zero
// voided_at timestamp.
func (ts *TransactionsModelSuite) TestVoidPendingTransaction() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		txnID := ts.uniqueTxnID("void_pending", 0)
		txn := &models.Transaction{
			BaseModel:       data.BaseModel{ID: txnID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    time.Now().UTC(),
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(75).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(75).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, txn)
		require.NoError(t, err)

		voided, err := res.TransactionBusiness.VoidTransaction(ctx, txnID)
		require.NoError(t, err)
		require.NotNil(t, voided)
		assert.Equal(t, models.TransactionStatusVoided, voided.Status)
		require.NotNil(t, voided.VoidedAt)
		assert.False(t, voided.VoidedAt.IsZero())
	})
}

// TestVoidRejectsPosted proves voiding a posted transaction is rejected;
// callers must reverse instead so the books carry an audit trail.
func (ts *TransactionsModelSuite) TestVoidRejectsPosted() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		txnID := ts.uniqueTxnID("void_posted", 0)
		now := time.Now().UTC()
		txn := &models.Transaction{
			BaseModel:       data.BaseModel{ID: txnID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    now,
			ClearedAt:       now,
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(75).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(75).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, txn)
		require.NoError(t, err)

		_, err = res.TransactionBusiness.VoidTransaction(ctx, txnID)
		require.Error(t, err, "voiding a posted transaction must be rejected")
	})
}

// TestMarkFailedFromPendingThenRejectAgain proves a pending transaction
// can be marked failed exactly once; subsequent attempts fail because the
// source-state set no longer matches.
func (ts *TransactionsModelSuite) TestMarkFailedFromPendingThenRejectAgain() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := ts.CreateService(t, depOpt)
		ts.setupFixtures(ctx, res)

		txnID := ts.uniqueTxnID("mark_failed", 0)
		txn := &models.Transaction{
			BaseModel:       data.BaseModel{ID: txnID},
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    time.Now().UTC(),
			Entries: []*models.TransactionEntry{
				{AccountID: "a1", Credit: false, Amount: decimalx.NewFromInt64(40).Ptr()},
				{AccountID: "a2", Credit: true, Amount: decimalx.NewFromInt64(40).Ptr()},
			},
		}
		_, err := res.TransactionBusiness.Transact(ctx, txn)
		require.NoError(t, err)

		failed, err := res.TransactionBusiness.MarkTransactionFailed(ctx, txnID)
		require.NoError(t, err)
		assert.Equal(t, models.TransactionStatusFailed, failed.Status)

		_, err = res.TransactionBusiness.MarkTransactionFailed(ctx, txnID)
		require.Error(t, err, "marking an already-failed transaction failed again must be rejected")
	})
}

func TestTransactionsModelSuite(t *testing.T) {
	suite.Run(t, new(TransactionsModelSuite))
}
