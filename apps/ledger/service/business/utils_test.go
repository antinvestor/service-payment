package business

import (
	"sort"
	"testing"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestOrderedEntries_LenSwapLess(t *testing.T) {
	entries := OrderedEntries{
		{AccountID: "b", Amount: decimal.NewNullDecimal(decimal.NewFromInt(200))},
		{AccountID: "a", Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
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
		{AccountID: "a", Amount: decimal.NewNullDecimal(decimal.NewFromInt(200))},
		{AccountID: "a", Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}

	// Same account, compare by amount
	assert.False(t, entries.Less(0, 1))
	assert.True(t, entries.Less(1, 0))
}

func TestOrderedEntries_Sort(t *testing.T) {
	entries := OrderedEntries{
		{AccountID: "c", Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
		{AccountID: "a", Amount: decimal.NewNullDecimal(decimal.NewFromInt(200))},
		{AccountID: "a", Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
		{AccountID: "b", Amount: decimal.NewNullDecimal(decimal.NewFromInt(50))},
	}

	sort.Sort(entries)
	assert.Equal(t, "a", entries[0].AccountID)
	assert.True(t, entries[0].Amount.Decimal.Equal(decimal.NewFromInt(100)))
	assert.Equal(t, "a", entries[1].AccountID)
	assert.True(t, entries[1].Amount.Decimal.Equal(decimal.NewFromInt(200)))
	assert.Equal(t, "b", entries[2].AccountID)
	assert.Equal(t, "c", entries[3].AccountID)
}

func TestContainsSameElements_Identical(t *testing.T) {
	entries := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
		{AccountID: "b", Credit: false, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	entries[0].ID = "e1"
	entries[1].ID = "e2"

	other := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
		{AccountID: "b", Credit: false, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	other[0].ID = "e1"
	other[1].ID = "e2"

	assert.True(t, containsSameElements(entries, other))
}

func TestContainsSameElements_DifferentLengths(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	e2 := []*models.TransactionEntry{
		{AccountID: "a", Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
		{AccountID: "b", Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_DifferentCredit(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: false, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	e2[0].ID = "e1"

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_DifferentAmount(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(200))},
	}
	e2[0].ID = "e1"

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_MissingEntry(t *testing.T) {
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	e2[0].ID = "e2" // Different ID

	assert.False(t, containsSameElements(e1, e2))
}

func TestContainsSameElements_NegativeAmount(t *testing.T) {
	// Test with negative amounts - abs comparison
	e1 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(-100))},
	}
	e1[0].ID = "e1"

	e2 := []*models.TransactionEntry{
		{AccountID: "a", Credit: true, Amount: decimal.NewNullDecimal(decimal.NewFromInt(100))},
	}
	e2[0].ID = "e1"

	assert.True(t, containsSameElements(e1, e2))
}
