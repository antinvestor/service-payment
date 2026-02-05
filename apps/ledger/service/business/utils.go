package business

import (
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
)

// DefaultTimestampLayout is the timestamp layout followed in Ledger.
const DefaultTimestampLayout = "2006-01-02T15:04:05.999999999"

// OrderedEntries implements sort.Interface for []*TransactionEntry based on
// the AccountID and Amount fields.
type OrderedEntries []models.TransactionEntry

func (entries OrderedEntries) Len() int      { return len(entries) }
func (entries OrderedEntries) Swap(i, j int) { entries[i], entries[j] = entries[j], entries[i] }
func (entries OrderedEntries) Less(i, j int) bool {
	if entries[i].AccountID == entries[j].AccountID {
		return entries[i].Amount.Decimal.LessThan(entries[j].Amount.Decimal)
	}
	return entries[i].AccountID < entries[j].AccountID
}

func containsSameElements(l1 []*models.TransactionEntry, l2 []*models.TransactionEntry) bool {
	if len(l1) != len(l2) {
		return false
	}

	// Key by entry ID (deterministic: {txnID}_{accountID}) to handle
	// transactions with multiple entries for the same account.
	l1Map := make(map[string]*models.TransactionEntry, len(l1))
	for _, entry := range l1 {
		l1Map[entry.ID] = entry
	}

	for _, entry2 := range l2 {
		entry, ok := l1Map[entry2.ID]
		if !ok {
			return false
		}

		if entry.Credit != entry2.Credit {
			return false
		}

		amount1 := entry.Amount.Decimal.Abs()
		amount2 := entry2.Amount.Decimal.Abs()
		if !amount1.Equal(amount2) {
			return false
		}
	}
	return true
}
