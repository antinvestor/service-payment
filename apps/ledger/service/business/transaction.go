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
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util/decimalx"
)

// TransactionBusiness defines the business interface for transaction operations.
type TransactionBusiness interface {
	CreateTransaction(ctx context.Context, req *ledgerv1.CreateTransactionRequest) (*ledgerv1.Transaction, error)
	SearchTransactions(ctx context.Context, req *commonv1.SearchRequest,
		consumer func(ctx context.Context, batch []*ledgerv1.Transaction) error) error
	GetTransaction(ctx context.Context, id string) (*ledgerv1.Transaction, error)
	UpdateTransaction(ctx context.Context, req *ledgerv1.UpdateTransactionRequest) (*ledgerv1.Transaction, error)
	ReverseTransaction(ctx context.Context, req *ledgerv1.ReverseTransactionRequest) (*ledgerv1.Transaction, error)
	DeleteTransaction(ctx context.Context, id string) error
	SearchEntries(
		ctx context.Context,
		req *commonv1.SearchRequest,
		consumer func(ctx context.Context, batch []*ledgerv1.TransactionEntry) error,
	) error

	IsConflict(
		ctx context.Context, transaction2 *models.Transaction) (bool, error)
	Transact(
		ctx context.Context, transaction *models.Transaction) (*models.Transaction, error)
	// VoidTransaction marks a not-yet-posted transaction as voided. Only
	// draft and pending transactions are voidable — posted activity must
	// be reversed instead so the books carry the audit trail.
	VoidTransaction(ctx context.Context, id string) (*models.Transaction, error)
	// MarkTransactionFailed transitions a pending transaction to failed.
	// Used by webhook handlers when the upstream provider rejects a
	// posting; the row stays in the journal for audit but no longer
	// contributes to any balance.
	MarkTransactionFailed(ctx context.Context, id string) (*models.Transaction, error)
}

// transactionBusiness implements the TransactionBusiness interface.
type transactionBusiness struct {
	workMan         workerpool.Manager
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
}

// NewTransactionBusiness creates a new transaction business instance.
func NewTransactionBusiness(
	workMan workerpool.Manager,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
) TransactionBusiness {
	return &transactionBusiness{
		workMan:         workMan,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

// CreateTransaction creates a new transaction with business validation.
func (b *transactionBusiness) CreateTransaction(
	ctx context.Context,
	req *ledgerv1.CreateTransactionRequest,
) (*ledgerv1.Transaction, error) {
	// Business logic validation
	if req.GetId() == "" {
		return nil, ErrTransactionReferenceRequired
	}

	if req.GetCurrency() == "" {
		return nil, ErrTransactionCurrencyRequired
	}

	// Convert API request to model
	transactionModel := models.TransactionFromAPI(ctx, &ledgerv1.Transaction{
		Id:           req.GetId(),
		CurrencyCode: req.GetCurrency(),
		TransactedAt: req.GetTransactedAt(),
		Data:         req.GetData(),
		Entries:      req.GetEntries(),
		Cleared:      req.GetCleared(),
		Type:         req.GetType(),
	})

	// All validation (structure, entries, accounts, currency) is performed
	// inside Transact → Validate, so no separate validation is needed here.

	// Create the transaction through repository
	result, err := b.Transact(ctx, transactionModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Convert back to API type
	return result.ToAPI(), nil
}

// SearchTransactions searches for transactions based on query.
func (b *transactionBusiness) SearchTransactions(ctx context.Context, req *commonv1.SearchRequest,
	consumer func(ctx context.Context, batch []*ledgerv1.Transaction) error) error {
	// Business logic for search validation
	query := req.GetQuery()
	if query == "" {
		query = "{}" // Default empty query
	}

	// Search through repository
	result, err := b.transactionRepo.SearchAsESQ(ctx, query)
	if err != nil {
		return err
	}

	for {
		res, ok := result.ReadResult(ctx)
		if !ok {
			return nil
		}

		if res.IsError() {
			return res.Error()
		}

		var apiResults []*ledgerv1.Transaction
		for _, transaction := range res.Item() {
			apiResults = append(apiResults, transaction.ToAPI())
		}

		jobErr := consumer(ctx, apiResults)
		if jobErr != nil {
			return jobErr
		}
	}
}

// GetTransaction retrieves a transaction by ID.
func (b *transactionBusiness) GetTransaction(ctx context.Context, id string) (*ledgerv1.Transaction, error) {
	if id == "" {
		return nil, ErrTransactionIDRequired
	}

	transaction, err := b.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert to API type
	return transaction.ToAPI(), nil
}

// UpdateTransaction updates an existing transaction.
func (b *transactionBusiness) UpdateTransaction(
	ctx context.Context,
	req *ledgerv1.UpdateTransactionRequest,
) (*ledgerv1.Transaction, error) {
	// Business logic validation
	if req.GetId() == "" {
		return nil, ErrTransactionIDRequired
	}

	// Convert API request to model - need to get existing transaction first
	existingTransaction, err := b.transactionRepo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	if existingTransaction == nil {
		return nil, apperrors.ErrTransactionNotFound
	}

	// Update fields from request
	if req.GetData() != nil {
		dataMap := data.JSONMap{}
		dataMap = dataMap.FromProtoStruct(req.GetData())

		for key, value := range dataMap {
			if value != "" && value != existingTransaction.Data[key] {
				existingTransaction.Data[key] = value
			}
		}
	}

	// Only pending transactions can be advanced to posted. Already-posted
	// rows are immutable in that regard; terminal statuses (reversed/voided/
	// failed) cannot be reopened.
	if existingTransaction.Status == models.TransactionStatusPending {
		err = b.processClearanceUpdate(ctx, req, existingTransaction)
		if err != nil {
			return nil, err
		}
	}

	// Update through repository
	_, err = b.transactionRepo.Update(ctx, existingTransaction)
	if err != nil {
		return nil, err
	}

	// Convert to API type
	return existingTransaction.ToAPI(), nil
}

// ReverseTransaction posts a REVERSAL transaction whose entries cancel the
// original's balance impact, and atomically marks the original as 'reversed'.
// Only posted NORMAL transactions can be reversed; the atomic guard in
// transactionRepo.CreateReversal prevents double-reversal under concurrency.
func (b *transactionBusiness) ReverseTransaction(
	ctx context.Context,
	req *ledgerv1.ReverseTransactionRequest,
) (*ledgerv1.Transaction, error) {
	if req.GetId() == "" {
		return nil, ErrTransactionIDRequired
	}

	originalTxn, err := b.transactionRepo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}

	if originalTxn.TransactionType != ledgerv1.TransactionType_NORMAL.String() {
		return nil, apperrors.ErrTransactionTypeNotReversible.Extend(
			fmt.Sprintf("transaction (type=%s) is not reversible", originalTxn.TransactionType),
		)
	}
	if originalTxn.Status != models.TransactionStatusPosted {
		return nil, apperrors.ErrTransactionTypeNotReversible.Extend(
			fmt.Sprintf("transaction (status=%s) is not reversible; only posted transactions can be reversed",
				originalTxn.Status),
		)
	}

	now := time.Now()
	originalID := originalTxn.ID
	reversalTxn := &models.Transaction{
		Currency:              originalTxn.Currency,
		TransactionType:       ledgerv1.TransactionType_REVERSAL.String(),
		Status:                models.TransactionStatusPosted,
		ReversedTransactionID: &originalID,
		TransactedAt:          now,
		ClearedAt:             now,
		PostedAt:              &now,
		Data:                  originalTxn.Data,
	}
	reversalTxn.GenID(ctx)
	reversalTxn.ID = fmt.Sprintf("%s_REVERSAL", originalTxn.ID)

	// Build reversal entries by flipping the Credit flag while keeping the
	// original stored amount. preProcessTransactionEntries applies DEADCLIC
	// sign rules based on the Credit flag and account type. Since the credit
	// flag is flipped, preProcess negates the amount, producing the
	// offsetting entry that zeroes out the original's balance impact.
	//
	// Do NOT also negate the amount here — that would double-negate (once
	// explicit, once via preProcess), restoring the original positive value
	// and making the reversal ineffective.
	for _, entry := range originalTxn.Entries {
		reversalTxn.Entries = append(reversalTxn.Entries, &models.TransactionEntry{
			BaseModel: data.BaseModel{ID: fmt.Sprintf("%s_REVERSAL", entry.ID)},
			AccountID: entry.AccountID,
			Amount:    entry.Amount,
			Credit:    !entry.Credit,
		})
	}

	reversedTxn, err := b.transactAsReversal(ctx, reversalTxn, originalTxn.ID)
	if err != nil {
		return nil, err
	}

	return reversedTxn.ToAPI(), nil
}

// VoidTransaction transitions a draft or pending transaction to 'voided'
// and stamps voided_at. Once a transaction posts, voiding is no longer
// permitted — reverse instead so the books carry the offset and the
// audit trail.
func (b *transactionBusiness) VoidTransaction(
	ctx context.Context, id string,
) (*models.Transaction, error) {
	if id == "" {
		return nil, ErrTransactionIDRequired
	}
	now := time.Now()
	err := b.transactionRepo.TransitionStatus(
		ctx, id,
		[]string{models.TransactionStatusDraft, models.TransactionStatusPending},
		models.TransactionStatusVoided,
		"voided_at", &now,
	)
	if err != nil {
		return nil, err
	}
	return b.transactionRepo.GetByID(ctx, id)
}

// MarkTransactionFailed transitions a pending transaction to 'failed' —
// e.g. the upstream payment provider rejected the request. Draft and
// terminal states are not affected so callers cannot retroactively
// fail something that already posted or that was never submitted.
func (b *transactionBusiness) MarkTransactionFailed(
	ctx context.Context, id string,
) (*models.Transaction, error) {
	if id == "" {
		return nil, ErrTransactionIDRequired
	}
	err := b.transactionRepo.TransitionStatus(
		ctx, id,
		[]string{models.TransactionStatusPending},
		models.TransactionStatusFailed,
		"", nil,
	)
	if err != nil {
		return nil, err
	}
	return b.transactionRepo.GetByID(ctx, id)
}

// transactAsReversal mirrors Transact for the reversal path but routes the
// final write through CreateReversal so the original is flipped to 'reversed'
// in the same DB transaction.
func (b *transactionBusiness) transactAsReversal(
	ctx context.Context, transaction *models.Transaction, originalID string,
) (*models.Transaction, error) {
	if transaction.TransactedAt.IsZero() {
		transaction.TransactedAt = time.Now()
	}

	accountsMap, aerr := b.Validate(ctx, transaction)
	if aerr != nil {
		var appErr apperrors.ApplicationError
		if errors.As(aerr, &appErr) {
			return nil, appErr
		}
		return nil, apperrors.ErrSystemFailure.Override(aerr)
	}

	b.preProcessTransactionEntries(transaction, accountsMap)

	sort.Slice(transaction.Entries, func(i, j int) bool {
		ei, ej := transaction.Entries[i], transaction.Entries[j]
		if ei.AccountID != ej.AccountID {
			return ei.AccountID < ej.AccountID
		}
		if ei.Credit != ej.Credit {
			return !ei.Credit
		}
		absI := decimalx.DerefOr(ei.Amount, decimalx.Zero()).Abs()
		absJ := decimalx.DerefOr(ej.Amount, decimalx.Zero()).Abs()
		return absI.LessThan(absJ)
	})

	for i, entry := range transaction.Entries {
		if entry.ID == "" {
			entry.ID = fmt.Sprintf("%s_%s_%d", transaction.GetID(), entry.AccountID, i)
		}
		entry.TransactionID = transaction.GetID()
	}

	err := b.transactionRepo.CreateReversal(ctx, transaction, originalID)
	if err == nil {
		return transaction, nil
	}
	if !data.ErrorIsDuplicateKey(err) {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return b.resolveDuplicate(ctx, transaction)
}

// DeleteTransaction is not supported — transactions are immutable audit records.
// Use ReverseTransaction to create an offsetting reversal instead.
func (b *transactionBusiness) DeleteTransaction(_ context.Context, id string) error {
	if id == "" {
		return ErrTransactionIDRequired
	}

	return apperrors.ErrTransactionTypeNotReversible.Extend(
		"transactions cannot be deleted; use ReverseTransaction to create an offsetting reversal")
}

// SearchEntries searches for transaction entries based on query.
func (b *transactionBusiness) SearchEntries(
	ctx context.Context,
	req *commonv1.SearchRequest,
	consumer func(ctx context.Context, batch []*ledgerv1.TransactionEntry) error,
) error {
	// Business logic for search validation
	query := req.GetQuery()
	if query == "" {
		query = "{}" // Default empty query
	}

	// Search through repository
	result, err := b.transactionRepo.SearchEntries(ctx, query)
	if err != nil {
		return err
	}

	for {
		res, ok := result.ReadResult(ctx)
		if !ok {
			return nil
		}

		if res.IsError() {
			return res.Error()
		}

		var apiResults []*ledgerv1.TransactionEntry
		for _, txEntry := range res.Item() {
			apiResults = append(apiResults, txEntry.ToAPI())
		}

		jobErr := consumer(ctx, apiResults)
		if jobErr != nil {
			return jobErr
		}
	}
}

// Validate checks all issues around transaction are satisfied.
func (b *transactionBusiness) Validate(
	ctx context.Context,
	txn *models.Transaction,
) (map[string]*models.Account, error) {
	if err := validateTransactionShape(txn); err != nil {
		return nil, err
	}
	if len(txn.Entries) == 0 {
		return nil, apperrors.ErrTransactionEntriesNotFound
	}

	accountIDSet := map[string]bool{}
	for _, entry := range txn.Entries {
		accountIDSet[entry.AccountID] = true
	}

	accountIDs := make([]string, 0, len(accountIDSet))
	for accountID := range accountIDSet {
		accountIDs = append(accountIDs, accountID)
	}

	// Posting only needs account metadata (LedgerType, Currency, BookID).
	// Calling the balance-aware ListByID here would run a LATERAL subquery
	// per account on every Create and dominates posting latency under
	// concurrent load.
	accountsMap, errAcc := b.accountRepo.ListMetaByID(ctx, accountIDs...)
	if errAcc != nil {
		return nil, errAcc
	}

	if err := validateBookScope(txn, accountsMap); err != nil {
		return nil, err
	}

	for _, entry := range txn.Entries {
		entryAmount := decimalx.DerefOr(entry.Amount, decimalx.Zero())
		if entryAmount.IsZero() {
			return nil, apperrors.ErrTransactionEntryHasZeroAmount.Extend(
				fmt.Sprintf("entry [id=%s, account_id=%s] amount is zero", entry.ID, entry.AccountID),
			)
		}

		account, ok := accountsMap[entry.AccountID]
		if !ok {
			// // Accounts have to be predefined hence check all references exist.
			return nil, apperrors.ErrAccountNotFound.Extend(
				fmt.Sprintf("Account %s was not found in the system", entry.AccountID),
			)
		}

		if !strings.EqualFold(txn.Currency, account.Currency) {
			return nil, apperrors.ErrTransactionAccountsDifferCurrency.Extend(
				fmt.Sprintf(
					"entry [id=%s, account_id=%s] currency [%s] != [%s]",
					entry.ID,
					entry.AccountID,
					account.Currency,
					txn.Currency,
				),
			)
		}
	}

	return accountsMap, nil
}

// IsConflict says whether a transaction conflicts with an existing transaction.
func (b *transactionBusiness) IsConflict(
	ctx context.Context, transaction2 *models.Transaction) (bool, error) {
	transaction1, err := b.transactionRepo.GetByID(ctx, transaction2.ID)
	if err != nil {
		return false, apperrors.ErrSystemFailure.Override(err)
	}

	// CompareMoney new and existing transaction Entries
	return !containsSameElements(transaction1.Entries, transaction2.Entries), nil
}

// Transact creates the input transaction in the DB with idempotent duplicate
// handling and conflict detection.
//
// Flow:
//  1. Validate entries and accounts.
//  2. Apply DEADCLIC sign rules and generate deterministic entry IDs.
//  3. Attempt insert. On success, return immediately.
//  4. On duplicate key error, fetch the existing transaction and compare
//     entries — identical means idempotent retry, different means conflict.
//
// Note: we intentionally skip a pre-insert existence check. GetByID uses
// Preload(Associations) which runs separate queries for parent and children.
// Under concurrent inserts this can observe the transaction record before its
// entries are committed, producing a false conflict. The duplicate-key error
// path is safe because PostgreSQL only returns 23505 after the competing
// transaction has fully committed, guaranteeing the subsequent GetByID sees
// complete data.
func (b *transactionBusiness) Transact(
	ctx context.Context, transaction *models.Transaction,
) (*models.Transaction, error) {
	// Set transaction time early to ensure consistency
	if transaction.TransactedAt.IsZero() {
		transaction.TransactedAt = time.Now()
	}

	// Default Status from any legacy ClearedAt the caller set directly (the
	// model-layer conversion handles the proto path, but the repository/test
	// surface constructs Transaction values directly). Pending if uncleared.
	if transaction.Status == "" {
		if !transaction.ClearedAt.IsZero() {
			transaction.Status = models.TransactionStatusPosted
			if transaction.PostedAt == nil {
				posted := transaction.ClearedAt
				transaction.PostedAt = &posted
			}
		} else {
			transaction.Status = models.TransactionStatusPending
		}
	}

	// Pre-validate accounts before any database operations to fail fast
	accountsMap, aerr := b.Validate(ctx, transaction)
	if aerr != nil {
		var appErr apperrors.ApplicationError
		if errors.As(aerr, &appErr) {
			return nil, appErr
		}
		return nil, apperrors.ErrSystemFailure.Override(aerr)
	}

	// Process transaction entries with account balances and signage
	b.preProcessTransactionEntries(transaction, accountsMap)

	// Sort entries by (AccountID, Credit, |Amount|) so that deterministic
	// ID generation produces the same IDs regardless of input order.
	sort.Slice(transaction.Entries, func(i, j int) bool {
		ei, ej := transaction.Entries[i], transaction.Entries[j]
		if ei.AccountID != ej.AccountID {
			return ei.AccountID < ej.AccountID
		}
		if ei.Credit != ej.Credit {
			return !ei.Credit // debit before credit
		}
		absI := decimalx.DerefOr(ei.Amount, decimalx.Zero()).Abs()
		absJ := decimalx.DerefOr(ej.Amount, decimalx.Zero()).Abs()
		return absI.LessThan(absJ)
	})

	// Generate deterministic entry IDs and set TransactionID.
	// The index is included to handle multiple entries for the same account
	// (e.g., split transactions).
	for i, entry := range transaction.Entries {
		if entry.ID == "" {
			entry.ID = fmt.Sprintf("%s_%s_%d", transaction.GetID(), entry.AccountID, i)
		}
		entry.TransactionID = transaction.GetID()
	}

	// Attempt to create the transaction.
	err := b.transactionRepo.Create(ctx, transaction)
	if err == nil {
		return transaction, nil
	}
	if !data.ErrorIsDuplicateKey(err) {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return b.resolveDuplicate(ctx, transaction)
}

// resolveDuplicate is invoked when the optimistic Create hits a unique
// constraint violation. The conflict could be on the primary key
// (same Transaction.ID re-submitted) or on the idempotency_key
// partial unique index (different Transaction.ID, same client-supplied
// dedup token). Try the idempotency_key path first since it carries the
// explicit caller intent, then fall back to PK-based reconciliation.
func (b *transactionBusiness) resolveDuplicate(
	ctx context.Context, transaction *models.Transaction,
) (*models.Transaction, error) {
	if transaction.IdempotencyKey != "" {
		existing, lerr := b.transactionRepo.GetByIdempotencyKey(ctx, transaction.IdempotencyKey)
		if lerr == nil && existing != nil {
			// Entry IDs differ across retries (derived from distinct
			// Transaction.IDs), so dedup uses content multiset equivalence.
			if !entriesEquivalent(existing.Entries, transaction.Entries) {
				return nil, apperrors.ErrTransactionIsConflicting.Extend(
					"idempotency_key reused with different entries")
			}
			return existing, nil
		}
		// Lookup failed or no match — fall through to PK-based resolution.
	}

	storedTxn, verifyErr := b.transactionRepo.GetByID(ctx, transaction.GetID())
	if verifyErr != nil {
		return nil, apperrors.ErrSystemFailure.Override(verifyErr)
	}
	if !containsSameElements(storedTxn.Entries, transaction.Entries) {
		return nil, apperrors.ErrTransactionIsConflicting
	}
	return storedTxn, nil
}

// validateTransactionShape applies the type-specific structural rules:
// NORMAL and REVERSAL must zero-sum per currency and contain at least one
// debit and one credit per currency. RESERVATION carries exactly one entry.
// All other types are accepted without shape-level checks.
func validateTransactionShape(txn *models.Transaction) error {
	switch txn.TransactionType {
	case ledgerv1.TransactionType_NORMAL.String(),
		ledgerv1.TransactionType_REVERSAL.String():
		if !txn.IsZeroSum() {
			return apperrors.ErrTransactionHasNonZeroSum
		}
		if !txn.IsTrueDrCr() {
			return apperrors.ErrTransactionHasInvalidDrCrEntry
		}
	case ledgerv1.TransactionType_RESERVATION.String():
		if len(txn.Entries) != 1 {
			return apperrors.ErrTransactionHasInvalidDrCrEntry
		}
	}
	return nil
}

// validateBookScope enforces cross-book integrity: when a transaction is
// scoped to a book, every entry's account must belong to that same book.
// Settlements that cross book boundaries must be modeled as two separate
// transactions linked by external_ref — never as a single entry list
// spanning books. Backward compatibility: if the transaction does not
// carry a BookID, the check is skipped entirely.
func validateBookScope(
	txn *models.Transaction, accountsMap map[string]*models.Account,
) error {
	if txn.BookID == nil || *txn.BookID == "" {
		return nil
	}
	for _, entry := range txn.Entries {
		acc := accountsMap[entry.AccountID]
		if acc == nil {
			continue
		}
		if acc.BookID == nil || *acc.BookID != *txn.BookID {
			return apperrors.ErrAccountNotFound.Extend(
				fmt.Sprintf("account %s does not belong to book %s",
					entry.AccountID, *txn.BookID))
		}
	}
	return nil
}

// preProcessTransactionEntries applies DEADCLIC sign rules and stamps the
// entry-level currency from the resolved account so it is persisted and
// available for currency-aware integrity checks at any point post-write.
//
// Sign rule, generalised: store +amount when the entry's side equals the
// account's normal balance side, -amount when it is the opposite. Memo
// accounts (normal_balance='none') opt out of normalisation and store
// amounts as supplied. The function falls back to LedgerType-based
// inference for accounts that have not yet had NormalBalance populated
// (in-memory paths constructing models directly).
func (b *transactionBusiness) preProcessTransactionEntries(
	transaction *models.Transaction,
	accountsMap map[string]*models.Account,
) {
	for _, line := range transaction.Entries {
		account := accountsMap[line.AccountID]

		// Validate already asserts entry account currency matches the
		// transaction currency, so either is correct here. Use the account's
		// to keep entries tied to their posting destination explicitly.
		line.Currency = account.Currency

		normalSide := account.NormalBalance
		if normalSide == "" {
			normalSide = models.NormalBalanceFromLedgerType(account.LedgerType)
		}
		if normalSide == "" || normalSide == models.NormalBalanceNone {
			// Memo or unclassified — store as supplied, no sign flip.
			continue
		}

		entrySide := models.NormalBalanceDebit
		if line.Credit {
			entrySide = models.NormalBalanceCredit
		}
		if entrySide != normalSide {
			neg := decimalx.DerefOr(line.Amount, decimalx.Zero()).Neg()
			line.Amount = &neg
		}
	}
}

// processClearanceUpdate drives the pending → posted transition: sets
// ClearedAt + PostedAt and flips Status. Posted entries themselves are
// immutable; only the parent transaction's status timestamps move.
func (b *transactionBusiness) processClearanceUpdate(
	ctx context.Context,
	req *ledgerv1.UpdateTransactionRequest,
	existingTransaction *models.Transaction,
) error {
	if req.GetClearedAt() == "" {
		return nil
	}

	clearanceTime, parseErr := time.Parse(DefaultTimestampLayout, req.GetClearedAt())
	if parseErr != nil {
		return parseErr
	}

	if _, validationErr := b.Validate(ctx, existingTransaction); validationErr != nil {
		return validationErr
	}

	existingTransaction.ClearedAt = clearanceTime
	existingTransaction.Status = models.TransactionStatusPosted
	existingTransaction.PostedAt = &clearanceTime
	return nil
}
