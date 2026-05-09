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

package models_test

import (
	"context"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// --- FromLedgerType / ToLedgerType ---

func TestFromLedgerType(t *testing.T) {
	tests := []struct {
		input    ledgerv1.LedgerType
		expected string
	}{
		{ledgerv1.LedgerType_ASSET, "ASSET"},
		{ledgerv1.LedgerType_LIABILITY, "LIABILITY"},
		{ledgerv1.LedgerType_INCOME, "INCOME"},
		{ledgerv1.LedgerType_EXPENSE, "EXPENSE"},
		{ledgerv1.LedgerType_CAPITAL, "CAPITAL"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, models.FromLedgerType(tc.input))
		})
	}
}

func TestToLedgerType(t *testing.T) {
	tests := []struct {
		input    string
		expected ledgerv1.LedgerType
	}{
		{"ASSET", ledgerv1.LedgerType_ASSET},
		{"LIABILITY", ledgerv1.LedgerType_LIABILITY},
		{"INCOME", ledgerv1.LedgerType_INCOME},
		{"EXPENSE", ledgerv1.LedgerType_EXPENSE},
		{"CAPITAL", ledgerv1.LedgerType_CAPITAL},
		{"UNKNOWN", ledgerv1.LedgerType(0)},
		{"", ledgerv1.LedgerType(0)},
	}

	for _, tc := range tests {
		name := tc.input
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, models.ToLedgerType(tc.input))
		})
	}
}

// --- Ledger.ToAPI ---

func TestLedgerToAPI(t *testing.T) {
	lg := &models.Ledger{
		BaseModel: data.BaseModel{ID: "ledger-1"},
		Type:      "ASSET",
		ParentID:  "parent-1",
		Data:      data.JSONMap{"key": "value"},
	}

	api := lg.ToAPI()
	assert.Equal(t, "ledger-1", api.GetId())
	assert.Equal(t, ledgerv1.LedgerType_ASSET, api.GetType())
	assert.Equal(t, "parent-1", api.GetParent())
	assert.NotNil(t, api.GetData())
}

func TestLedgerToAPI_NoParent(t *testing.T) {
	lg := &models.Ledger{
		BaseModel: data.BaseModel{ID: "ledger-2"},
		Type:      "LIABILITY",
	}

	api := lg.ToAPI()
	assert.Equal(t, "ledger-2", api.GetId())
	assert.Equal(t, ledgerv1.LedgerType_LIABILITY, api.GetType())
	assert.Empty(t, api.GetParent())
}

// --- Account.ToAPI ---

func TestAccountToAPI_ValidBalance(t *testing.T) {
	acc := &models.Account{
		BaseModel:        data.BaseModel{ID: "acc-1"},
		LedgerID:         "ledger-1",
		Currency:         "USD",
		Balance:          decimalx.NewFromInt64(100).Ptr(),
		UnClearedBalance: decimalx.NewFromInt64(50).Ptr(),
		ReservedBalance:  decimalx.NewFromInt64(25).Ptr(),
		Data:             data.JSONMap{"name": "checking"},
	}

	api := acc.ToAPI()
	assert.Equal(t, "acc-1", api.GetId())
	assert.Equal(t, "ledger-1", api.GetLedger())
	assert.Equal(t, "USD", api.GetBalance().GetCurrencyCode())
	assert.Equal(t, int64(100), api.GetBalance().GetUnits())
	assert.Equal(t, int64(50), api.GetUnclearedBalance().GetUnits())
	assert.Equal(t, int64(25), api.GetReservedBalance().GetUnits())
}

func TestAccountToAPI_InvalidBalance(t *testing.T) {
	acc := &models.Account{
		BaseModel: data.BaseModel{ID: "acc-2"},
		Currency:  "EUR",
		// Balance.Valid is false (zero value)
	}

	api := acc.ToAPI()
	assert.Equal(t, int64(0), api.GetBalance().GetUnits())
	assert.Equal(t, "EUR", api.GetBalance().GetCurrencyCode())
}

// --- TransactionFromAPI ---

func TestTransactionFromAPI(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	strct, err := structpb.NewStruct(map[string]any{"ref": "pay-123"})
	require.NoError(t, err)

	apiTxn := &ledgerv1.Transaction{
		Id:           "txn-1",
		CurrencyCode: "USD",
		Type:         ledgerv1.TransactionType_NORMAL,
		TransactedAt: now.Format(time.RFC3339),
		Cleared:      true,
		Data:         strct,
		Entries: []*ledgerv1.TransactionEntry{
			{
				AccountId: "acc-1",
				Credit:    false,
				Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
			},
			{
				AccountId: "acc-2",
				Credit:    true,
				Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 100},
			},
		},
	}

	txn := models.TransactionFromAPI(ctx, apiTxn)
	assert.Equal(t, "txn-1", txn.ID)
	assert.Equal(t, "USD", txn.Currency)
	assert.Equal(t, "NORMAL", txn.TransactionType)
	assert.False(t, txn.TransactedAt.IsZero())
	assert.False(t, txn.ClearedAt.IsZero())
	assert.Len(t, txn.Entries, 2)
	assert.Equal(t, "acc-1", txn.Entries[0].AccountID)
	assert.Equal(t, "acc-2", txn.Entries[1].AccountID)
}

func TestTransactionFromAPI_EmptyEntries(t *testing.T) {
	ctx := context.Background()
	apiTxn := &ledgerv1.Transaction{
		Id:           "txn-empty",
		CurrencyCode: "USD",
	}

	txn := models.TransactionFromAPI(ctx, apiTxn)
	assert.Equal(t, "txn-empty", txn.ID)
	assert.Nil(t, txn.Entries)
}

func TestTransactionFromAPI_NotCleared(t *testing.T) {
	ctx := context.Background()
	apiTxn := &ledgerv1.Transaction{
		Id:           "txn-uncleared",
		CurrencyCode: "USD",
		Cleared:      false,
	}

	txn := models.TransactionFromAPI(ctx, apiTxn)
	assert.True(t, txn.ClearedAt.IsZero())
}

func TestTransactionFromAPI_InvalidTransactedAt(t *testing.T) {
	ctx := context.Background()
	apiTxn := &ledgerv1.Transaction{
		Id:           "txn-bad-time",
		CurrencyCode: "USD",
		TransactedAt: "not-a-date",
	}

	txn := models.TransactionFromAPI(ctx, apiTxn)
	assert.True(t, txn.TransactedAt.IsZero())
}

// --- Transaction.ToAPI ---

func TestTransactionToAPI(t *testing.T) {
	now := time.Now().UTC()
	txn := &models.Transaction{
		BaseModel:       data.BaseModel{ID: "txn-1"},
		Currency:        "USD",
		TransactionType: "NORMAL",
		TransactedAt:    now,
		ClearedAt:       now,
		Entries: []*models.TransactionEntry{
			{
				BaseModel:     data.BaseModel{ID: "e1"},
				AccountID:     "acc-1",
				TransactionID: "txn-1",
				Amount:        decimalx.NewFromInt64(100).Ptr(),
				Credit:        false,
			},
		},
	}

	api := txn.ToAPI()
	assert.Equal(t, "txn-1", api.GetId())
	assert.Equal(t, "USD", api.GetCurrencyCode())
	assert.Equal(t, ledgerv1.TransactionType_NORMAL, api.GetType())
	assert.True(t, api.GetCleared())
	assert.NotEmpty(t, api.GetTransactedAt())
	assert.Len(t, api.GetEntries(), 1)
}

func TestTransactionToAPI_NotCleared(t *testing.T) {
	txn := &models.Transaction{
		BaseModel:       data.BaseModel{ID: "txn-2"},
		Currency:        "USD",
		TransactionType: "NORMAL",
	}

	api := txn.ToAPI()
	assert.False(t, api.GetCleared())
	assert.Empty(t, api.GetTransactedAt())
}

func TestTransactionToAPI_UnknownType(t *testing.T) {
	txn := &models.Transaction{
		BaseModel:       data.BaseModel{ID: "txn-3"},
		TransactionType: "UNKNOWN_TYPE",
	}

	api := txn.ToAPI()
	assert.Equal(t, ledgerv1.TransactionType(0), api.GetType())
}

// --- TransactionEntryFromAPI ---

func TestTransactionEntryFromAPI(t *testing.T) {
	entry := models.TransactionEntryFromAPI(&ledgerv1.TransactionEntry{
		AccountId: "acc-1",
		Credit:    true,
		Amount:    &commonv1.Money{CurrencyCode: "USD", Units: 50, Nanos: 500000000},
	})

	assert.Equal(t, "acc-1", entry.AccountID)
	assert.True(t, entry.Credit)
	assert.NotNil(t, entry.Amount)
	expected, _ := decimalx.NewFromString("50.5")
	assert.True(t, entry.Amount.Equal(expected))
}

func TestTransactionEntryFromAPI_NilAmount(t *testing.T) {
	entry := models.TransactionEntryFromAPI(&ledgerv1.TransactionEntry{
		AccountId: "acc-2",
	})

	assert.Equal(t, "acc-2", entry.AccountID)
	assert.NotNil(t, entry.Amount)
	assert.True(t, entry.Amount.IsZero())
}

// --- TransactionEntry.ToAPI ---

func TestTransactionEntryToAPI_ValidAmount(t *testing.T) {
	te := &models.TransactionEntry{
		BaseModel:     data.BaseModel{ID: "te-1"},
		AccountID:     "acc-1",
		TransactionID: "txn-1",
		Amount:        decimalx.NewFromInt64(200).Ptr(),
		Credit:        true,
	}

	api := te.ToAPI()
	assert.Equal(t, "te-1", api.GetId())
	assert.Equal(t, "acc-1", api.GetAccountId())
	assert.Equal(t, "txn-1", api.GetTransactionId())
	assert.True(t, api.GetCredit())
	assert.NotNil(t, api.GetAmount())
	assert.Equal(t, int64(200), api.GetAmount().GetUnits())
}

func TestTransactionEntryToAPI_InvalidAmount(t *testing.T) {
	te := &models.TransactionEntry{
		BaseModel: data.BaseModel{ID: "te-2"},
		AccountID: "acc-2",
		// Amount.Valid is false
	}

	api := te.ToAPI()
	assert.Nil(t, api.GetAmount())
}

// --- TransactionEntry.Equal ---

func TestTransactionEntryEqual_Same(t *testing.T) {
	te1 := &models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}
	te2 := models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}

	assert.True(t, te1.Equal(te2))
}

func TestTransactionEntryEqual_DifferentAmount(t *testing.T) {
	te1 := &models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}
	te2 := models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(200).Ptr(),
	}

	assert.False(t, te1.Equal(te2))
}

func TestTransactionEntryEqual_DifferentAccount(t *testing.T) {
	te1 := &models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}
	te2 := models.TransactionEntry{
		AccountID: "acc-2",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}

	assert.False(t, te1.Equal(te2))
}

func TestTransactionEntryEqual_DifferentCredit(t *testing.T) {
	te1 := &models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}
	te2 := models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    false,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}

	assert.False(t, te1.Equal(te2))
}

func TestTransactionEntryEqual_InvalidAmount(t *testing.T) {
	te1 := &models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		// Amount.Valid is false
	}
	te2 := models.TransactionEntry{
		AccountID: "acc-1",
		Credit:    true,
		Amount:    decimalx.NewFromInt64(100).Ptr(),
	}

	assert.False(t, te1.Equal(te2))
}

// --- Transaction.IsZeroSum ---

func TestIsZeroSum_Balanced(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
			{Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
		},
	}
	assert.True(t, txn.IsZeroSum())
}

func TestIsZeroSum_Unbalanced(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
			{Credit: true, Amount: decimalx.NewFromInt64(50).Ptr()},
		},
	}
	assert.False(t, txn.IsZeroSum())
}

func TestIsZeroSum_SingleEntry(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
		},
	}
	assert.False(t, txn.IsZeroSum())
}

func TestIsZeroSum_Empty(t *testing.T) {
	txn := &models.Transaction{}
	assert.True(t, txn.IsZeroSum())
}

func TestIsZeroSum_MultipleEntries(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: false, Amount: decimalx.NewFromInt64(50).Ptr()},
			{Credit: false, Amount: decimalx.NewFromInt64(50).Ptr()},
			{Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
		},
	}
	assert.True(t, txn.IsZeroSum())
}

// --- Transaction.IsTrueDrCr ---

func TestIsTrueDrCr_Valid(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: false},
			{Credit: true},
		},
	}
	assert.True(t, txn.IsTrueDrCr())
}

func TestIsTrueDrCr_AllDebits(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: false},
			{Credit: false},
		},
	}
	assert.False(t, txn.IsTrueDrCr())
}

func TestIsTrueDrCr_AllCredits(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: true},
			{Credit: true},
		},
	}
	assert.False(t, txn.IsTrueDrCr())
}

func TestIsTrueDrCr_SingleEntry(t *testing.T) {
	txn := &models.Transaction{
		Entries: []*models.TransactionEntry{
			{Credit: true},
		},
	}
	assert.False(t, txn.IsTrueDrCr())
}

func TestIsTrueDrCr_Empty(t *testing.T) {
	txn := &models.Transaction{}
	assert.False(t, txn.IsTrueDrCr())
}
