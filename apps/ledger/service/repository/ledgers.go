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
	"fmt"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util"
)

type LedgerRepository interface {
	datastore.BaseRepository[*models.Ledger]
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.Ledger], error)
}

// LedgerRepository provides all functions related to ledger Ledger.
type ledgerRepository struct {
	datastore.BaseRepository[*models.Ledger]
}

// NewLedgerRepository provides instance of `LedgerRepository`.
func NewLedgerRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) LedgerRepository {
	return &ledgerRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Ledger](
			ctx, dbPool, workMan, func() *models.Ledger { return &models.Ledger{} },
		),
	}
}

// Query constants for ledger repository.
const constLedgerQuery = `SELECT
    l.id, l.parent_id, l.type, l.data,
    l.created_at, l.modified_at, l.version,
    l.tenant_id, l.partition_id, l.access_id, l.deleted_at
FROM ledgers l`

func (l *ledgerRepository) searchLedgers(ctx context.Context, sqlQuery *SearchSQLQuery) ([]*models.Ledger, error) {
	var whereParts []string
	var allArgs []interface{}

	tenancySQL, tenancyArgs := buildTenancyClause(ctx, "l")
	if tenancySQL != "" {
		whereParts = append(whereParts, tenancySQL)
		allArgs = append(allArgs, tenancyArgs...)
	}

	whereParts = append(whereParts, "l.deleted_at IS NULL")

	if sqlQuery.sql != "" {
		whereParts = append(whereParts, sqlQuery.sql)
		allArgs = append(allArgs, sqlQuery.args...)
	}

	whereClause := "1=1"
	if len(whereParts) > 0 {
		whereClause = fmt.Sprintf("(%s)", joinAND(whereParts))
	}

	fullSQL := fmt.Sprintf(`%s WHERE %s ORDER BY l.created_at DESC LIMIT ? OFFSET ?`,
		constLedgerQuery, whereClause)
	allArgs = append(allArgs, sqlQuery.batchSize, sqlQuery.offset)

	rows, err := l.Pool().DB(ctx, true).Raw(fullSQL, allArgs...).Rows()
	if err != nil {
		return nil, err
	}

	defer util.CloseAndLogOnError(ctx, rows, "could not close ledger rows")

	ledgerList := make([]*models.Ledger, 0)
	for rows.Next() {
		ledger := new(models.Ledger)
		errR := rows.Scan(
			&ledger.ID, &ledger.ParentID, &ledger.Type, &ledger.Data,
			&ledger.CreatedAt, &ledger.ModifiedAt, &ledger.Version,
			&ledger.TenantID, &ledger.PartitionID, &ledger.AccessID, &ledger.DeletedAt)
		if errR != nil {
			return ledgerList, errR
		}
		ledgerList = append(ledgerList, ledger)
	}

	return ledgerList, nil
}

func (l *ledgerRepository) paginateLedgerSearch(
	ctx context.Context,
	sqlQuery *SearchSQLQuery,
	jobResult workerpool.JobResultPipe[[]*models.Ledger],
) error {
	for sqlQuery.canLoad() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		ledgerList, dbErr := l.searchLedgers(ctx, sqlQuery)
		if dbErr != nil {
			if data.ErrorIsNoRows(dbErr) {
				return jobResult.WriteError(ctx, apperrors.ErrLedgerNotFound)
			}
			return jobResult.WriteError(ctx, apperrors.ErrSystemFailure.Override(dbErr))
		}

		errR := jobResult.WriteResult(ctx, ledgerList)
		if errR != nil {
			return errR
		}

		if sqlQuery.stop(len(ledgerList)) {
			break
		}
	}
	return nil
}

func (l *ledgerRepository) SearchAsESQ(
	ctx context.Context,
	query string,
) (workerpool.JobResultPipe[[]*models.Ledger], error) {
	job := workerpool.NewJob(func(ctxI context.Context, jobResult workerpool.JobResultPipe[[]*models.Ledger]) error {
		rawQuery, err := NewSearchRawQuery(ctxI, query)
		if err != nil {
			return jobResult.WriteError(ctxI, err)
		}

		return l.paginateLedgerSearch(ctxI, rawQuery.ToQueryConditions(), jobResult)
	})

	err := workerpool.SubmitJob(ctx, l.WorkManager(), job)
	if err != nil {
		return nil, err
	}

	return job, nil
}
