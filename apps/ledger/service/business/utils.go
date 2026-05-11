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

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/pitabwire/util/decimalx"
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
		amtI := decimalx.DerefOr(entries[i].Amount, decimalx.Zero())
		amtJ := decimalx.DerefOr(entries[j].Amount, decimalx.Zero())
		return amtI.LessThan(amtJ)
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

		amount1 := decimalx.DerefOr(entry.Amount, decimalx.Zero()).Abs()
		amount2 := decimalx.DerefOr(entry2.Amount, decimalx.Zero()).Abs()
		if !amount1.Equal(amount2) {
			return false
		}
	}
	return true
}

// entriesEquivalent compares two entry lists by content (account, side,
// currency, |amount|) ignoring entry IDs. Used for idempotency_key dedup
// where two retries of the same logical request carry distinct
// Transaction.IDs — and therefore distinct deterministic entry IDs — but
// represent identical postings.
//
// Amounts are compared numerically via decimalx.Equal because the same
// decimal value can have multiple string representations (e.g. "100" vs
// "100.000000000" round-tripped through numeric(29,9)). Grouping by
// (account, side, currency) and sorting amounts within each group yields
// stable O(N log N) comparison.
func entriesEquivalent(l1, l2 []*models.TransactionEntry) bool {
	if len(l1) != len(l2) {
		return false
	}

	type groupKey struct {
		AccountID string
		Side      string
		Currency  string
	}
	keyOf := func(e *models.TransactionEntry) groupKey {
		side := "D"
		if e.Credit {
			side = "C"
		}
		return groupKey{AccountID: e.AccountID, Side: side, Currency: e.Currency}
	}

	group := func(list []*models.TransactionEntry) map[groupKey][]decimalx.Decimal {
		out := make(map[groupKey][]decimalx.Decimal, len(list))
		for _, e := range list {
			amt := decimalx.DerefOr(e.Amount, decimalx.Zero()).Abs()
			out[keyOf(e)] = append(out[keyOf(e)], amt)
		}
		for _, amts := range out {
			sort.Slice(amts, func(i, j int) bool { return amts[i].LessThan(amts[j]) })
		}
		return out
	}

	g1 := group(l1)
	g2 := group(l2)
	if len(g1) != len(g2) {
		return false
	}
	for k, amts1 := range g1 {
		amts2, ok := g2[k]
		if !ok || len(amts1) != len(amts2) {
			return false
		}
		for i := range amts1 {
			if !amts1[i].Equal(amts2[i]) {
				return false
			}
		}
	}
	return true
}
