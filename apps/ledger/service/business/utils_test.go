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

package business

import (
	"sort"
	"strings"
	"testing"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
)

func TestOrderedEntries_LenSwapLess(t *testing.T) {
	entries := OrderedEntries{
		{AccountID: "b", Amount: decimalx.NewFromInt64(200).Ptr()},
		{AccountID: "a", Amount: decimalx.NewFromInt64(100).Ptr()},
	}

	assert.Equal(t, 2, entries.Len())

	entries.Swap(0, 1)
	assert.Equal(t, "a", entries[0].AccountID)
	assert.Equal(t, "b", entries[1].AccountID)

	// Less: different account IDs
	assert.True(t, entries.Less(0, 1))
	assert.False(t, entries.Less(1, 0))
}

func TestOrderedEntries_LessSameAccount(t *testing.T) {
	entries := OrderedEntries{
		{AccountID: "a", Amount: decimalx.NewFromInt64(200).Ptr()},
		{AccountID: "a", Amount: decimalx.NewFromInt64(100).Ptr()},
	}

	// Same account, compare by amount
	assert.False(t, entries.Less(0, 1))
	assert.True(t, entries.Less(1, 0))
}

func TestOrderedEntries_Sort(t *testing.T) {
	entries := OrderedEntries{
		{AccountID: "c", Amount: decimalx.NewFromInt64(100).Ptr()},
		{AccountID: "a", Amount: decimalx.NewFromInt64(200).Ptr()},
		{AccountID: "a", Amount: decimalx.NewFromInt64(100).Ptr()},
		{AccountID: "b", Amount: decimalx.NewFromInt64(50).Ptr()},
	}

	sort.Sort(entries)
	assert.Equal(t, "a", entries[0].AccountID)
	assert.True(t, entries[0].Amount.Equal(decimalx.NewFromInt64(100)))
	assert.Equal(t, "a", entries[1].AccountID)
	assert.True(t, entries[1].Amount.Equal(decimalx.NewFromInt64(200)))
	assert.Equal(t, "b", entries[2].AccountID)
	assert.Equal(t, "c", entries[3].AccountID)
}

func TestContainsSameElements_Identical(t *testing.T) {
	entries := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
		{AccountID: "b", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	entries[0].ID = "e1"
	entries[1].ID = "e2"

	other := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
		{AccountID: "b", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	other[0].ID = "e1"
	other[1].ID = "e2"

	assert.True(t, containsSameElements(entries, other))
}

func TestContainsSameElements_DifferentLengths(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	e2 := []*models.TransactionEntry{
		{AccountID: "a", Amount: decimalx.NewFromInt64(100).Ptr()},
		{AccountID: "b", Amount: decimalx.NewFromInt64(100).Ptr()},
	}

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_DifferentCredit(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	e2[0].ID = "e1"

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_DifferentAmount(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(200).Ptr()},
	}
	e2[0].ID = "e1"

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_MissingEntry(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	e2[0].ID = "e2" // Different ID

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_NegativeAmount(t *testing.T) {
	// Test with negative amounts - abs comparison
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(-100).Ptr()},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
	}
	e2[0].ID = "e1"

	assert.True(t, containsSameElements(e1, e2))
}

func TestPostingEntryID_ReadableWhenItFits(t *testing.T) {
	assert.Equal(t, "txn-1_acc-1_0", postingEntryID("txn-1", "acc-1", 0))
	assert.Equal(t, "txn-1_acc-1_0_REVERSAL", reversalEntryID("txn-1_acc-1_0"))
}

func TestPostingEntryID_MaxLengthIDsFitColumn(t *testing.T) {
	txnID := strings.Repeat("t", 40)
	accountID := strings.Repeat("a", 40)

	seen := map[string]struct{}{}
	for i := range 1000 {
		id := postingEntryID(txnID, accountID, i)
		assert.LessOrEqual(t, len(id), maxEntryIDLength)
		assert.NotContains(t, id, "_", "derived ids never contain the readable separator")
		assert.Equal(t, id, postingEntryID(txnID, accountID, i), "derivation must be deterministic")
		seen[id] = struct{}{}

		rev := reversalEntryID(id)
		assert.LessOrEqual(t, len(rev), maxEntryIDLength)
		assert.NotEqual(t, id, rev)
		seen[rev] = struct{}{}
	}
	assert.Len(t, seen, 2000, "entry ids must be unique per (transaction, account, index)")

	assert.NotEqual(t,
		postingEntryID(txnID, accountID, 0),
		postingEntryID(txnID, strings.Repeat("b", 40), 0))
	assert.NotEqual(t,
		postingEntryID(txnID, accountID, 0),
		postingEntryID(strings.Repeat("u", 40), accountID, 0))
}

func TestReversalEntryID_LongReadableOriginal(t *testing.T) {
	// A readable original id at the column limit gains a suffix that would
	// overflow, so the reversal id is derived.
	original := strings.Repeat("x", maxEntryIDLength)
	rev := reversalEntryID(original)
	assert.Len(t, rev, len(entryIDHashPrefix)+entryIDHashLength)
	assert.Equal(t, rev, reversalEntryID(original))
}
