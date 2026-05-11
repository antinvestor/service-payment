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

package models

import (
	"time"

	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util/decimalx"
)

// TrialBalanceParams filters the per-account trial-balance aggregation.
// Zero-value fields are not applied as filters.
type TrialBalanceParams struct {
	Currency   string
	LedgerID   string
	LedgerType string
	// BookIDs scopes the aggregation to a set of books. Empty slice means
	// "no book filter" (every book in the caller's tenancy). When the
	// caller wants a consolidated report across an organization and all
	// its groups + members, the report business pre-expands the root id
	// into all descendants and supplies them here.
	BookIDs []string
	// AsOf restricts the aggregation to transactions transacted at or before
	// this instant. Nil means "no upper bound" — include everything posted.
	AsOf *time.Time
}

// TrialBalanceLine is one row of the trial balance: a single account's
// raw debit and credit totals plus the DEADCLIC-signed net balance.
//
// TotalDebits and TotalCredits are unsigned magnitudes (|amount| grouped by
// the Credit flag) so a textbook trial-balance check is `sum(TotalDebits) ==
// sum(TotalCredits)` per currency across all rows. NetBalance is the natural
// balance — positive for normal-debit accounts (assets, expenses) when in
// surplus, positive for normal-credit accounts (liabilities, equity, income)
// when in surplus.
type TrialBalanceLine struct {
	AccountID    string
	LedgerID     string
	LedgerType   string
	Currency     string
	TotalDebits  decimalx.Decimal
	TotalCredits decimalx.Decimal
	NetBalance   decimalx.Decimal
}

// StatementParams scopes an account statement query.
//
// AccountID is required. From/To are inclusive bounds on transacted_at; nil
// means "no bound on that side". Limit and Offset paginate the entries;
// callers receive a stable chronological order (transacted_at, entry_id).
type StatementParams struct {
	AccountID string
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
}

// StatementEntryRow is a single entry on an account statement, hydrated with
// its parent transaction's metadata so the caller need not perform a second
// query per row. RunningBalance is populated by the business layer; the
// repository returns rows with RunningBalance unset.
type StatementEntryRow struct {
	EntryID         string
	TransactionID   string
	Amount          decimalx.Decimal
	Credit          bool
	Currency        string
	TransactedAt    time.Time
	ClearedAt       time.Time
	TransactionType string
	TransactionData data.JSONMap
	RunningBalance  decimalx.Decimal
}
