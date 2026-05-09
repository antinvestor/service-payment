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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util"
)

// constAccountQuery uses a LATERAL subquery to compute balances scoped to
// each matched account, avoiding a full-table aggregation of transaction_entries.
// The LATERAL join runs the balance aggregation once per matched account row,
// bounded by the outer WHERE + OFFSET/LIMIT.
// Soft-deleted transactions (deleted_at IS NOT NULL) are excluded from balances.
const constAccountQuery = `SELECT
    a.id,
    a.currency,
    a.data,
    COALESCE(bs.balance, 0) AS total_balance,
    COALESCE(bs.un_cleared_balance, 0) AS total_uncleared_balance,
    COALESCE(bs.reserved_balance, 0) AS total_reserved_balance,
    a.ledger_id,
    a.ledger_type,
    a.created_at,
    a.modified_at,
    a.version,
    a.tenant_id,
    a.partition_id,
    a.access_id,
    a.deleted_at
FROM accounts a
LEFT JOIN LATERAL (
    SELECT
        COALESCE(SUM(CASE
            WHEN t.transaction_type IN ('NORMAL', 'REVERSAL')
                AND t.cleared_at IS NOT NULL
                AND t.cleared_at != '0001-01-01 00:00:00'
            THEN e.amount ELSE 0 END), 0) AS balance,
        COALESCE(SUM(CASE
            WHEN t.transaction_type IN ('NORMAL', 'REVERSAL')
                AND (t.cleared_at IS NULL OR t.cleared_at = '0001-01-01 00:00:00')
            THEN e.amount ELSE 0 END), 0) AS un_cleared_balance,
        COALESCE(SUM(CASE
            WHEN t.transaction_type = 'RESERVATION'
            THEN e.amount ELSE 0 END), 0) AS reserved_balance
    FROM transaction_entries e
    INNER JOIN transactions t ON e.transaction_id = t.id
        AND t.deleted_at IS NULL
    WHERE e.account_id = a.id
) bs ON true `

type AccountRepository interface {
	datastore.BaseRepository[*models.Account]
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.Account], error)
	ListByID(ctx context.Context, ids ...string) (map[string]*models.Account, error)
	HasTransactionEntries(ctx context.Context, accountID string) (bool, error)
	CountByLedgerID(ctx context.Context, ledgerID string) (int64, error)
}

// accountRepository provides all functions related to ledger account.
type accountRepository struct {
	datastore.BaseRepository[*models.Account]
}

// NewAccountRepository provides instance of `accountRepository`.
func NewAccountRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) AccountRepository {
	return &accountRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Account](
			ctx, dbPool, workMan, func() *models.Account { return &models.Account{} },
		),
	}
}

// GetByID returns an acccount with the given Reference.
func (a *accountRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Account, error) {
	if id == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	accList, err := a.ListByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return accList[id], nil
}

// ListByID returns a list of acccounts with the given list of ids.
func (a *accountRepository) ListByID(
	ctx context.Context,
	ids ...string,
) (map[string]*models.Account, error) {
	if len(ids) == 0 {
		return nil, apperrors.ErrAccountsNotFound.Extend("No Accounts were specified")
	}

	accountsMap := map[string]*models.Account{}

	queryMap := map[string]any{
		"query": map[string]any{
			"must": map[string]any{
				"fields": []map[string]any{
					{
						"id": map[string][]string{
							"in": ids,
						},
					},
				},
			},
		},
	}

	queryBytes, err := json.Marshal(queryMap)
	if err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err).Extend("Json marshalling error")
	}

	query := string(queryBytes)

	jobResult, err := a.SearchAsESQ(ctx, query)
	if err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err).Extend(fmt.Sprintf("db query error [%s]", query))
	}

	for {
		result, ok := jobResult.ReadResult(ctx)

		if !ok {
			return accountsMap, nil
		}

		if result.IsError() {
			return nil, apperrors.ErrSystemFailure.Override(result.Error())
		}

		for _, acc := range result.Item() {
			accountsMap[acc.ID] = acc
		}
	}
}

// buildTenancyClause returns a SQL WHERE fragment and args for tenant/partition
// scoping, extracted from the authentication claims in the context.
// When no claims are present (e.g. internal cross-tenant services), it returns
// an empty clause so the query remains unscoped — matching frame's TenancyPartition behavior.
func buildTenancyClause(ctx context.Context, tableAlias string) (string, []interface{}) {
	authClaim := security.ClaimsFromContext(ctx)
	if authClaim == nil || security.IsTenancyChecksOnClaimSkipped(ctx) {
		return "", nil
	}

	prefix := ""
	if tableAlias != "" {
		prefix = tableAlias + "."
	}

	clause := fmt.Sprintf("%stenant_id = ? AND %spartition_id = ?", prefix, prefix)
	return clause, []interface{}{authClaim.GetTenantID(), authClaim.GetPartitionID()}
}

func (a *accountRepository) searchAccounts(ctx context.Context, sqlQuery *SearchSQLQuery) ([]*models.Account, error) {
	// Build WHERE clause combining tenancy scoping with the search conditions.
	// Raw SQL bypasses GORM's automatic TenancyPartition scope, so we must
	// inject tenant filtering explicitly.
	var whereParts []string
	var allArgs []interface{}

	tenancySQL, tenancyArgs := buildTenancyClause(ctx, "a")
	if tenancySQL != "" {
		whereParts = append(whereParts, tenancySQL)
		allArgs = append(allArgs, tenancyArgs...)
	}

	// Exclude soft-deleted accounts
	whereParts = append(whereParts, "a.deleted_at IS NULL")

	if sqlQuery.sql != "" {
		whereParts = append(whereParts, sqlQuery.sql)
		allArgs = append(allArgs, sqlQuery.args...)
	}

	whereClause := "1=1"
	if len(whereParts) > 0 {
		whereClause = fmt.Sprintf("(%s)", joinAND(whereParts))
	}

	fullSQL := fmt.Sprintf(`%s WHERE %s ORDER BY a.created_at DESC LIMIT ? OFFSET ?`,
		constAccountQuery, whereClause)
	allArgs = append(allArgs, sqlQuery.batchSize, sqlQuery.offset)

	rows, err := a.Pool().DB(ctx, true).Raw(fullSQL, allArgs...).Rows()
	if err != nil {
		return nil, err
	}

	defer util.CloseAndLogOnError(ctx, rows, "could not close account rows")

	var accountList []*models.Account
	for rows.Next() {
		acc := models.Account{}
		err = rows.Scan(
			&acc.ID, &acc.Currency, &acc.Data, &acc.Balance, &acc.UnClearedBalance, &acc.ReservedBalance,
			&acc.LedgerID, &acc.LedgerType, &acc.CreatedAt, &acc.ModifiedAt, &acc.Version, &acc.TenantID,
			&acc.PartitionID, &acc.AccessID, &acc.DeletedAt)
		if err != nil {
			return accountList, err
		}
		accountList = append(accountList, &acc)
	}

	return accountList, nil
}

// joinAND joins SQL fragments with " AND ".
func joinAND(parts []string) string {
	result := parts[0]
	var resultSb230 strings.Builder
	for _, p := range parts[1:] {
		resultSb230.WriteString(" AND " + p)
	}
	result += resultSb230.String()
	return result
}

func (a *accountRepository) paginateAccountSearch(
	ctx context.Context,
	sqlQuery *SearchSQLQuery,
	jobResult workerpool.JobResultPipe[[]*models.Account],
) error {
	for sqlQuery.canLoad() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		accountList, dbErr := a.searchAccounts(ctx, sqlQuery)
		if dbErr != nil {
			if data.ErrorIsNoRows(dbErr) {
				return jobResult.WriteError(ctx, apperrors.ErrAccountsNotFound)
			}
			return jobResult.WriteError(
				ctx,
				apperrors.ErrSystemFailure.Override(dbErr).Extend("Query execution error"),
			)
		}

		dbErr = jobResult.WriteResult(ctx, accountList)
		if dbErr != nil {
			return dbErr
		}

		if sqlQuery.stop(len(accountList)) {
			break
		}
	}
	return nil
}

// HasTransactionEntries returns true if the account has any transaction entries.
func (a *accountRepository) HasTransactionEntries(ctx context.Context, accountID string) (bool, error) {
	var count int64
	err := a.Pool().DB(ctx, true).
		Model(&models.TransactionEntry{}).
		Where("account_id = ?", accountID).
		Limit(1).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountByLedgerID returns the number of non-deleted accounts in the given ledger.
func (a *accountRepository) CountByLedgerID(ctx context.Context, ledgerID string) (int64, error) {
	var count int64
	err := a.Pool().DB(ctx, true).
		Model(&models.Account{}).
		Where("ledger_id = ? AND deleted_at IS NULL", ledgerID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *accountRepository) SearchAsESQ(
	ctx context.Context,
	query string,
) (workerpool.JobResultPipe[[]*models.Account], error) {
	job := workerpool.NewJob(func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.Account]) error {
		rawQuery, aerr := NewSearchRawQuery(ctx, query)
		if aerr != nil {
			return jobResult.WriteError(ctx, aerr)
		}

		return a.paginateAccountSearch(ctx, rawQuery.ToQueryConditions(), jobResult)
	})

	err := workerpool.SubmitJob(ctx, a.WorkManager(), job)
	if err != nil {
		return nil, err
	}

	return job, nil
}
