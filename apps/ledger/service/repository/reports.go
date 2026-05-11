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

package repository

import (
	"context"
	"time"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/util/decimalx"
)

// ReportRepository exposes aggregate read operations that derive accounting
// reports from existing entries. It deliberately does not embed
// BaseRepository: there is no domain entity called "report" — only queries
// shaped against accounts, transactions and transaction_entries.
//
// All queries route through frame's pool, which wraps every gorm.DB result
// with scopes.TenancyPartition. The primary `Model(&models.Account{})` (or
// equivalent) sets Statement.Table so the scope's tenant_id and partition_id
// filters resolve to the correct table; JOINed tables inherit isolation
// through the account-side WHERE.
type ReportRepository interface {
	// AggregateTrialBalance returns one row per account that has at least
	// one cleared NORMAL or REVERSAL entry, with raw |amount| debit and
	// credit totals plus the DEADCLIC-signed net. NetBalance per account
	// uses the stored signed amount (sign already adjusted at posting
	// time), so the textbook trial-balance check is on TotalDebits vs
	// TotalCredits, not on NetBalance.
	AggregateTrialBalance(
		ctx context.Context, params models.TrialBalanceParams,
	) ([]*models.TrialBalanceLine, error)

	// StatementOpeningBalance returns the DEADCLIC-signed sum of entries
	// strictly before the supplied instant. Returns zero when before is
	// nil — no entries are "before the dawn of time". Used to seed running
	// balances on paginated statements.
	StatementOpeningBalance(
		ctx context.Context, accountID string, before *time.Time,
	) (decimalx.Decimal, error)

	// AccountStatementEntries returns entries for the account in
	// chronological order, paginated, with their parent transaction's
	// metadata hydrated. The repository does not compute RunningBalance —
	// that's a business-layer concern dependent on the opening balance.
	AccountStatementEntries(
		ctx context.Context, params models.StatementParams,
	) ([]*models.StatementEntryRow, error)
}

type reportRepository struct {
	dbPool pool.Pool
}

// NewReportRepository constructs the report repository wired to the supplied
// pool. The pool is held directly (not via a BaseRepository wrapper) because
// the report queries do not map cleanly to a single entity's lifecycle.
func NewReportRepository(dbPool pool.Pool) ReportRepository {
	return &reportRepository{dbPool: dbPool}
}

// committedStatuses returns the Transaction.Status values whose entries
// are part of "the books" — they posted and (in the case of reversed) had
// their offset posted too. Pending and terminal-but-uncommitted states
// (draft, voided, failed) never contributed to balance and are excluded.
// A function (rather than a var) keeps the slice immutable across callers.
func committedStatuses() []string {
	return []string{
		models.TransactionStatusPosted,
		models.TransactionStatusReversed,
	}
}

func (r *reportRepository) AggregateTrialBalance(
	ctx context.Context, params models.TrialBalanceParams,
) ([]*models.TrialBalanceLine, error) {
	db := r.dbPool.DB(ctx, true).Model(&models.Account{}).
		Select(`
			accounts.id AS account_id,
			accounts.ledger_id,
			accounts.ledger_type,
			accounts.currency,
			COALESCE(SUM(CASE WHEN NOT transaction_entries.credit
				THEN ABS(transaction_entries.amount) ELSE 0 END), 0) AS total_debits,
			COALESCE(SUM(CASE WHEN transaction_entries.credit
				THEN ABS(transaction_entries.amount) ELSE 0 END), 0) AS total_credits,
			COALESCE(SUM(transaction_entries.amount), 0) AS net_balance
		`).
		Joins(`INNER JOIN transaction_entries
			ON transaction_entries.account_id = accounts.id
			AND transaction_entries.deleted_at IS NULL`).
		Joins(`INNER JOIN transactions
			ON transactions.id = transaction_entries.transaction_id
			AND transactions.deleted_at IS NULL
			AND transactions.transaction_type IN ('NORMAL', 'REVERSAL')
			AND transactions.status IN ?`, committedStatuses()).
		Group("accounts.id, accounts.ledger_id, accounts.ledger_type, accounts.currency").
		Order("accounts.ledger_type, accounts.id")

	if params.Currency != "" {
		db = db.Where("accounts.currency = ?", params.Currency)
	}
	if params.LedgerID != "" {
		db = db.Where("accounts.ledger_id = ?", params.LedgerID)
	}
	if params.LedgerType != "" {
		db = db.Where("accounts.ledger_type = ?", params.LedgerType)
	}
	if len(params.BookIDs) > 0 {
		db = db.Where("accounts.book_id IN ?", params.BookIDs)
	}
	if params.AsOf != nil {
		db = db.Where("transactions.transacted_at <= ?", *params.AsOf)
	}

	var rows []*models.TrialBalanceLine
	if err := db.Scan(&rows).Error; err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return rows, nil
}

func (r *reportRepository) StatementOpeningBalance(
	ctx context.Context, accountID string, before *time.Time,
) (decimalx.Decimal, error) {
	if accountID == "" {
		return decimalx.Zero(), apperrors.ErrUnspecifiedID
	}
	if before == nil {
		return decimalx.Zero(), nil
	}

	// Use the entry table as the model anchor; tenancy applies to
	// transaction_entries via the auto-scope.
	var sum decimalx.Decimal
	err := r.dbPool.DB(ctx, true).Model(&models.TransactionEntry{}).
		Select("COALESCE(SUM(transaction_entries.amount), 0) AS sum").
		Joins(`INNER JOIN transactions
			ON transactions.id = transaction_entries.transaction_id
			AND transactions.deleted_at IS NULL
			AND transactions.transaction_type IN ('NORMAL', 'REVERSAL')
			AND transactions.status IN ?`, committedStatuses()).
		Where("transaction_entries.account_id = ?", accountID).
		Where("transactions.transacted_at < ?", *before).
		Scan(&sum).Error
	if err != nil {
		return decimalx.Zero(), apperrors.ErrSystemFailure.Override(err)
	}
	return sum, nil
}

func (r *reportRepository) AccountStatementEntries(
	ctx context.Context, params models.StatementParams,
) ([]*models.StatementEntryRow, error) {
	if params.AccountID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	limit := params.Limit
	if limit <= 0 || limit > 1000 {
		// Cap unbounded statements to keep result sets predictable.
		limit = 100
	}

	db := r.dbPool.DB(ctx, true).Model(&models.TransactionEntry{}).
		Select(`
			transaction_entries.id AS entry_id,
			transaction_entries.transaction_id,
			transaction_entries.amount,
			transaction_entries.credit,
			transaction_entries.currency,
			transactions.transacted_at,
			transactions.cleared_at,
			transactions.transaction_type,
			transactions.data AS transaction_data
		`).
		Joins(`INNER JOIN transactions
			ON transactions.id = transaction_entries.transaction_id
			AND transactions.deleted_at IS NULL
			AND transactions.transaction_type IN ('NORMAL', 'REVERSAL')
			AND transactions.status IN ?`, committedStatuses()).
		Where("transaction_entries.account_id = ?", params.AccountID).
		Order("transactions.transacted_at ASC, transaction_entries.id ASC").
		Limit(limit).
		Offset(params.Offset)

	if params.From != nil {
		db = db.Where("transactions.transacted_at >= ?", *params.From)
	}
	if params.To != nil {
		db = db.Where("transactions.transacted_at <= ?", *params.To)
	}

	var rows []*models.StatementEntryRow
	if err := db.Scan(&rows).Error; err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return rows, nil
}
