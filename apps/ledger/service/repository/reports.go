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
	"strings"
	"time"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/util/decimalx"
	"gorm.io/gorm"
)

// ReportRepository exposes aggregate read operations that derive
// accounting reports from existing entries. It deliberately does not
// embed BaseRepository: there is no domain entity called "report" —
// only queries shaped against accounts, transactions and
// transaction_entries.
//
// Tenancy enforcement: every report query runs through frame's
// Pool.WithTenancy helper, which publishes app.tenant_id /
// app.partition_id session variables from the auth claims inside a
// transaction. Row-Level Security policies on each tenancy-scoped
// table read those variables via current_setting() and filter rows
// automatically. The SQL below therefore contains no tenant_id /
// partition_id references — application code stays unaware of
// tenancy entirely; frame and Postgres handle it between them.
type ReportRepository interface {
	AggregateTrialBalance(
		ctx context.Context, params models.TrialBalanceParams,
	) ([]*models.TrialBalanceLine, error)

	StatementOpeningBalance(
		ctx context.Context, accountID string, before *time.Time,
	) (decimalx.Decimal, error)

	AccountStatementEntries(
		ctx context.Context, params models.StatementParams,
	) ([]*models.StatementEntryRow, error)
}

type reportRepository struct {
	dbPool pool.Pool
}

// NewReportRepository constructs the report repository.
func NewReportRepository(dbPool pool.Pool) ReportRepository {
	return &reportRepository{dbPool: dbPool}
}

func (r *reportRepository) AggregateTrialBalance(
	ctx context.Context, params models.TrialBalanceParams,
) ([]*models.TrialBalanceLine, error) {
	var whereParts []string
	var args []interface{}

	whereParts = append(whereParts, "a.deleted_at IS NULL")
	whereParts = append(whereParts, "t.transaction_type IN ('NORMAL','REVERSAL')")
	whereParts = append(whereParts, "t.deleted_at IS NULL")
	whereParts = append(whereParts, "t.status IN ('posted','reversed')")
	whereParts = append(whereParts, "e.deleted_at IS NULL")

	if params.Currency != "" {
		whereParts = append(whereParts, "a.currency = ?")
		args = append(args, params.Currency)
	}
	if params.LedgerID != "" {
		whereParts = append(whereParts, "a.ledger_id = ?")
		args = append(args, params.LedgerID)
	}
	if params.LedgerType != "" {
		whereParts = append(whereParts, "a.ledger_type = ?")
		args = append(args, params.LedgerType)
	}
	if len(params.BookIDs) > 0 {
		whereParts = append(whereParts, "a.book_id IN ?")
		args = append(args, params.BookIDs)
	}
	if params.AsOf != nil {
		whereParts = append(whereParts, "t.transacted_at <= ?")
		args = append(args, *params.AsOf)
	}

	sqlText := `
SELECT
    a.id AS account_id,
    a.ledger_id AS ledger_id,
    a.ledger_type AS ledger_type,
    a.currency AS currency,
    COALESCE(SUM(CASE WHEN NOT e.credit THEN ABS(e.amount) ELSE 0 END), 0) AS total_debits,
    COALESCE(SUM(CASE WHEN e.credit THEN ABS(e.amount) ELSE 0 END), 0) AS total_credits,
    COALESCE(SUM(e.amount), 0) AS net_balance
FROM accounts a
INNER JOIN transaction_entries e ON e.account_id = a.id
INNER JOIN transactions t ON t.id = e.transaction_id
WHERE ` + strings.Join(whereParts, " AND ") + `
GROUP BY a.id, a.ledger_id, a.ledger_type, a.currency
ORDER BY a.ledger_type, a.id`

	var rows []*models.TrialBalanceLine
	if err := r.dbPool.WithTenancy(ctx, true, func(tx *gorm.DB) error {
		return tx.Raw(sqlText, args...).Scan(&rows).Error
	}); err != nil {
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

	sqlText := `
SELECT COALESCE(SUM(e.amount), 0) AS sum
FROM transaction_entries e
INNER JOIN transactions t ON t.id = e.transaction_id
WHERE e.account_id = ?
  AND t.transacted_at < ?
  AND t.transaction_type IN ('NORMAL','REVERSAL')
  AND t.deleted_at IS NULL
  AND t.status IN ('posted','reversed')
  AND e.deleted_at IS NULL`

	var sum decimalx.Decimal
	if err := r.dbPool.WithTenancy(ctx, true, func(tx *gorm.DB) error {
		return tx.Raw(sqlText, accountID, *before).Scan(&sum).Error
	}); err != nil {
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
		limit = 100
	}

	var whereParts []string
	var args []interface{}

	whereParts = append(whereParts, "e.account_id = ?")
	args = append(args, params.AccountID)
	whereParts = append(whereParts, "t.transaction_type IN ('NORMAL','REVERSAL')")
	whereParts = append(whereParts, "t.deleted_at IS NULL")
	whereParts = append(whereParts, "t.status IN ('posted','reversed')")
	whereParts = append(whereParts, "e.deleted_at IS NULL")

	if params.From != nil {
		whereParts = append(whereParts, "t.transacted_at >= ?")
		args = append(args, *params.From)
	}
	if params.To != nil {
		whereParts = append(whereParts, "t.transacted_at <= ?")
		args = append(args, *params.To)
	}

	args = append(args, limit, params.Offset)

	sqlText := `
SELECT
    e.id AS entry_id,
    e.transaction_id AS transaction_id,
    e.amount AS amount,
    e.credit AS credit,
    e.currency AS currency,
    t.transacted_at AS transacted_at,
    t.cleared_at AS cleared_at,
    t.transaction_type AS transaction_type,
    t.data AS transaction_data
FROM transaction_entries e
INNER JOIN transactions t ON t.id = e.transaction_id
WHERE ` + strings.Join(whereParts, " AND ") + `
ORDER BY t.transacted_at ASC, e.id ASC
LIMIT ? OFFSET ?`

	var rows []*models.StatementEntryRow
	if err := r.dbPool.WithTenancy(ctx, true, func(tx *gorm.DB) error {
		return tx.Raw(sqlText, args...).Scan(&rows).Error
	}); err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return rows, nil
}
