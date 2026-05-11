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
	"errors"
	"fmt"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepository interface {
	datastore.BaseRepository[*models.Transaction]
	SearchAsESQ(ctx context.Context, query string,
	) (workerpool.JobResultPipe[[]*models.Transaction], error)
	SearchEntries(ctx context.Context, query string,
	) (workerpool.JobResultPipe[[]*models.TransactionEntry], error)
	// GetByIdempotencyKey looks up a transaction by its idempotency key.
	// Returns gorm.ErrRecordNotFound when no row matches so callers can
	// distinguish "no prior posting" from a system failure.
	GetByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error)
}

// transactionRepository is the interface to all transaction operations.
type transactionRepository struct {
	accountRepo AccountRepository
	datastore.BaseRepository[*models.Transaction]
}

// NewTransactionRepository returns a new instance of `transactionRepository`.
func NewTransactionRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
	accountRepo AccountRepository,
) TransactionRepository {
	return &transactionRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Transaction](
			ctx, dbPool, workMan, func() *models.Transaction { return &models.Transaction{} },
		),
		accountRepo: accountRepo,
	}
}

// GetByIdempotencyKey returns the transaction (with its entries preloaded)
// whose idempotency_key column matches the provided value. Uses the partial
// UNIQUE index so the lookup is O(1) on the indexed column.
func (t *transactionRepository) GetByIdempotencyKey(
	ctx context.Context, key string,
) (*models.Transaction, error) {
	if key == "" {
		return nil, apperrors.ErrUnspecifiedID
	}
	var txn models.Transaction
	err := t.Pool().DB(ctx, true).
		Preload(clause.Associations).
		Where("idempotency_key = ?", key).
		First(&txn).Error
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

// Create overrides BaseRepository.Create to guarantee that the transaction
// header and all its entries are inserted atomically.
//
// The frame pool is configured with SkipDefaultTransaction=true, so a plain
// gorm.Create that walks the association tree would issue the parent INSERT
// and each child INSERT as independent statements. A process crash, network
// blip or query cancellation between those statements leaves an orphan
// transactions row with zero entries — invisible to the invalid_transactions
// view (SUM over empty set is 0) and permanently blocking idempotent retry
// because containsSameElements would reject the legitimate replay.
//
// Wrapping in an explicit transaction commits both the header and the entries
// or neither.
func (t *transactionRepository) Create(ctx context.Context, txn *models.Transaction) error {
	if txn.GetVersion() > 0 {
		return errors.New("entity version is more than 0, consider using Update instead of Create")
	}
	return t.Pool().DB(ctx, false).Transaction(func(tx *gorm.DB) error {
		return tx.Create(txn).Error
	})
}

func (t *transactionRepository) searchTransactions(
	ctx context.Context,
	sqlQuery *SearchSQLQuery,
) ([]*models.Transaction, error) {
	var transactionList []*models.Transaction

	// Tenancy filtering is applied automatically by frame's pool, which wraps
	// every .DB(ctx,_) result with scopes.TenancyPartition. GORM's query builder
	// (.Where + .Find) honours that scope, so we do not inject tenant_id/
	// partition_id manually here — see repository/accounts.go for the raw-SQL
	// exception where the scope cannot reach.
	result := t.Pool().DB(ctx, true).Where(sqlQuery.sql, sqlQuery.args...).Offset(sqlQuery.offset).
		Limit(sqlQuery.batchSize).Find(&transactionList)
	err1 := result.Error
	if err1 != nil {
		return transactionList, err1
	}

	if len(transactionList) > 0 {
		var transactionIDs []string
		for _, transaction := range transactionList {
			transactionIDs = append(transactionIDs, transaction.GetID())
		}

		entriesMap, err2 := t.SearchEntriesByTransactionID(ctx, transactionIDs...)
		if err2 != nil {
			return transactionList, err2
		}

		for _, transaction := range transactionList {
			entries, ok := entriesMap[transaction.GetID()]
			if ok {
				transaction.Entries = entries
			}
		}
	}

	return transactionList, nil
}

func (t *transactionRepository) SearchAsESQ(
	ctx context.Context, queryStr string,
) (workerpool.JobResultPipe[[]*models.Transaction], error) {
	job := workerpool.NewJob(
		func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.Transaction]) error {
			rawQuery, err := NewSearchRawQuery(ctx, queryStr)
			if err != nil {
				return jobResult.WriteError(ctx, err)
			}

			sqlQuery := rawQuery.ToQueryConditions()

			for sqlQuery.canLoad() {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				transactionList, dbErr := t.searchTransactions(ctx, sqlQuery)
				if dbErr != nil {
					return jobResult.WriteError(ctx, apperrors.ErrSystemFailure.Override(dbErr))
				}
				dbErr = jobResult.WriteResult(ctx, transactionList)
				if dbErr != nil {
					return dbErr
				}

				if sqlQuery.stop(len(transactionList)) {
					break
				}
			}
			return nil
		},
	)

	err := workerpool.SubmitJob(ctx, t.WorkManager(), job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (t *transactionRepository) SearchEntriesByTransactionID(
	ctx context.Context,
	transactionIDs ...string,
) (map[string][]*models.TransactionEntry, error) {
	entriesMap := make(map[string][]*models.TransactionEntry)

	queryMap := map[string]any{
		"query": map[string]any{
			"must": map[string]any{
				"fields": []map[string]any{
					{
						"transaction_id": map[string][]string{
							"in": transactionIDs,
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

	logger := util.Log(ctx)

	query := string(queryBytes)

	logger.WithField("query", query).Debug("searching entries by transaction ID")

	jobResult, err := t.SearchEntries(ctx, query)
	if err != nil {
		logger.WithError(err).Warn("could not query for entries")

		return nil, apperrors.ErrSystemFailure.Override(err).Extend(fmt.Sprintf("db query error [%s]", query))
	}

	for {
		result, ok := jobResult.ReadResult(ctx)

		if !ok {
			return entriesMap, nil
		}
		if result.IsError() {
			logger.WithError(result.Error()).Warn("could not read entry results")
			return nil, apperrors.ErrSystemFailure.Override(result.Error())
		}

		for _, entry := range result.Item() {
			entries, ok0 := entriesMap[entry.TransactionID]
			if !ok0 {
				entries = make([]*models.TransactionEntry, 0)
			}

			entriesMap[entry.TransactionID] = append(entries, entry)
		}
	}
}

func (t *transactionRepository) SearchEntries(
	ctx context.Context,
	query string,
) (workerpool.JobResultPipe[[]*models.TransactionEntry], error) {
	job := workerpool.NewJob(
		func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.TransactionEntry]) error {
			rawQuery, err := NewSearchRawQuery(ctx, query)
			if err != nil {
				return jobResult.WriteError(ctx, err)
			}

			sqlQuery := rawQuery.ToQueryConditions()

			for sqlQuery.canLoad() {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				var transactionEntriesList []*models.TransactionEntry
				// Tenancy filtering applied automatically by frame's pool scope —
				// see searchTransactions for the rationale.
				result := t.Pool().DB(ctx, true).Offset(sqlQuery.offset).Limit(sqlQuery.batchSize).
					Where(sqlQuery.sql, sqlQuery.args...).Find(&transactionEntriesList)

				err1 := result.Error
				if err1 != nil {
					return jobResult.WriteError(ctx, apperrors.ErrSystemFailure.Override(err1))
				}

				err1 = jobResult.WriteResult(ctx, transactionEntriesList)
				if err1 != nil {
					return err1
				}

				if sqlQuery.stop(len(transactionEntriesList)) {
					break
				}
			}

			return nil
		},
	)

	err := workerpool.SubmitJob(ctx, t.WorkManager(), job)
	if err != nil {
		return nil, err
	}

	return job, nil
}
