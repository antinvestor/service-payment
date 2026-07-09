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
	"context"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/moneyx"
	"gorm.io/gorm"
)

// Book represents an independent accounting scope: one entity's complete
// set of financial records. Each book has its own chart of accounts, its
// own trial balance and its own balance sheet — entries posted in one book
// must never cross into another. Type follows the Stawi-style convention
// (platform/group/customer/merchant/agent/branch) but is an open string so
// product domains can grow new entity classifications without a migration.
//
// ParentID supports hierarchy: an organization holds many group books
// (one per chama / SACCO / branch), each group holds many individual
// member books. Hierarchy is a read-side concern — consolidated reports
// roll a parent's descendants into one trial balance — while POSTING
// stays strictly per-book. Settlements that cross book boundaries are
// modeled as two separate transactions linked by external_ref, not as
// cross-book entries.
type Book struct {
	data.BaseModel
	ParentID *string      `gorm:"type:varchar(50)"                     json:"parent_id"`
	Name     string       `gorm:"type:varchar(100);not null"           json:"name"`
	Type     string       `gorm:"type:varchar(50);not null"            json:"type"`
	Currency string       `gorm:"type:varchar(10)"                     json:"currency"`
	Data     data.JSONMap `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}

// Ledger represents the hierarchy for organising ledgers with information such as type, and JSON data.
type Ledger struct {
	data.BaseModel
	BookID   *string      `gorm:"type:varchar(50)"                     json:"book_id"`
	Type     string       `gorm:"type:varchar(50)"                     json:"type"`
	ParentID string       `gorm:"type:varchar(50)"                     json:"parent_id"`
	Data     data.JSONMap `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
}

func FromLedgerType(raw ledgerv1.LedgerType) string {
	return ledgerv1.LedgerType_name[int32(raw)]
}

func ToLedgerType(model string) ledgerv1.LedgerType {
	ledgerType := ledgerv1.LedgerType_value[model]
	return ledgerv1.LedgerType(ledgerType)
}

// transactionStatusByModel maps stored Status strings (lowercase, matching
// the DB CHECK constraint) onto the generated proto enum. Symmetric helper
// covers the reverse direction.
//
//nolint:gochecknoglobals // immutable proto enum lookup
var transactionStatusByModel = map[string]ledgerv1.TransactionStatus{
	TransactionStatusPending:  ledgerv1.TransactionStatus_PENDING,
	TransactionStatusPosted:   ledgerv1.TransactionStatus_POSTED,
	TransactionStatusReversed: ledgerv1.TransactionStatus_REVERSED,
	TransactionStatusVoided:   ledgerv1.TransactionStatus_VOIDED,
	TransactionStatusFailed:   ledgerv1.TransactionStatus_FAILED,
	TransactionStatusDraft:    ledgerv1.TransactionStatus_DRAFT,
}

//nolint:gochecknoglobals // immutable proto enum lookup
var transactionStatusByProto = map[ledgerv1.TransactionStatus]string{
	ledgerv1.TransactionStatus_PENDING:  TransactionStatusPending,
	ledgerv1.TransactionStatus_POSTED:   TransactionStatusPosted,
	ledgerv1.TransactionStatus_REVERSED: TransactionStatusReversed,
	ledgerv1.TransactionStatus_VOIDED:   TransactionStatusVoided,
	ledgerv1.TransactionStatus_FAILED:   TransactionStatusFailed,
	ledgerv1.TransactionStatus_DRAFT:    TransactionStatusDraft,
}

// ToTransactionStatusProto returns the proto enum value for a stored
// status string, defaulting to PENDING for unknown inputs.
func ToTransactionStatusProto(s string) ledgerv1.TransactionStatus {
	if v, ok := transactionStatusByModel[s]; ok {
		return v
	}
	return ledgerv1.TransactionStatus_PENDING
}

// FromTransactionStatusProto converts a proto enum back to the stored
// status string used by the model + DB CHECK constraint.
func FromTransactionStatusProto(s ledgerv1.TransactionStatus) string {
	if v, ok := transactionStatusByProto[s]; ok {
		return v
	}
	return TransactionStatusPending
}

//nolint:gochecknoglobals // immutable proto enum lookup
var accountTypeByModel = map[string]ledgerv1.AccountType{
	AccountTypeAsset:           ledgerv1.AccountType_ACCOUNT_ASSET,
	AccountTypeLiability:       ledgerv1.AccountType_ACCOUNT_LIABILITY,
	AccountTypeEquity:          ledgerv1.AccountType_ACCOUNT_EQUITY,
	AccountTypeIncome:          ledgerv1.AccountType_ACCOUNT_INCOME,
	AccountTypeExpense:         ledgerv1.AccountType_ACCOUNT_EXPENSE,
	AccountTypeContraAsset:     ledgerv1.AccountType_ACCOUNT_CONTRA_ASSET,
	AccountTypeContraLiability: ledgerv1.AccountType_ACCOUNT_CONTRA_LIABILITY,
	AccountTypeContraIncome:    ledgerv1.AccountType_ACCOUNT_CONTRA_INCOME,
	AccountTypeContraExpense:   ledgerv1.AccountType_ACCOUNT_CONTRA_EXPENSE,
	AccountTypeClearing:        ledgerv1.AccountType_ACCOUNT_CLEARING,
	AccountTypeSuspense:        ledgerv1.AccountType_ACCOUNT_SUSPENSE,
	AccountTypeMemo:            ledgerv1.AccountType_ACCOUNT_MEMO,
}

// ToAccountTypeProto returns the proto enum value for a stored account
// type string, defaulting to ACCOUNT_ASSET for unknown inputs (matches
// the conventional zero value).
func ToAccountTypeProto(s string) ledgerv1.AccountType {
	if v, ok := accountTypeByModel[s]; ok {
		return v
	}
	return ledgerv1.AccountType_ACCOUNT_ASSET
}

//nolint:gochecknoglobals // immutable proto enum lookup
var normalBalanceByModel = map[string]ledgerv1.NormalBalance{
	NormalBalanceDebit:  ledgerv1.NormalBalance_DEBIT,
	NormalBalanceCredit: ledgerv1.NormalBalance_CREDIT,
	NormalBalanceNone:   ledgerv1.NormalBalance_NONE,
}

// ToNormalBalanceProto returns the proto enum value for a stored normal
// balance string, defaulting to DEBIT.
func ToNormalBalanceProto(s string) ledgerv1.NormalBalance {
	if v, ok := normalBalanceByModel[s]; ok {
		return v
	}
	return ledgerv1.NormalBalance_DEBIT
}

func (lg *Ledger) ToAPI() *ledgerv1.Ledger {
	bookID := ""
	if lg.BookID != nil {
		bookID = *lg.BookID
	}
	return &ledgerv1.Ledger{
		Id:     lg.ID,
		Type:   ToLedgerType(lg.Type),
		Parent: lg.ParentID,
		Data:   lg.Data.ToProtoStruct(),
		BookId: bookID,
	}
}

// ToAPI converts a Book domain model to the wire representation. ParentID
// and Currency surface as empty strings when unset to match proto3
// "no value" semantics callers expect.
func (b *Book) ToAPI() *ledgerv1.Book {
	parentID := ""
	if b.ParentID != nil {
		parentID = *b.ParentID
	}
	return &ledgerv1.Book{
		Id:       b.ID,
		Name:     b.Name,
		Type:     b.Type,
		ParentId: parentID,
		Currency: b.Currency,
		Data:     b.Data.ToProtoStruct(),
	}
}

// Account represents the ledger account with information such as Reference, balance and JSON data.
//
// AccountType and NormalBalance carry the per-account classification used
// for balance signage and report grouping. LedgerType is retained alongside
// for backward compatibility — both are derived from the parent ledger at
// creation time and stay in sync. BookID is denormalised from the parent
// Ledger so cross-book validation in posting can be done in a single
// account lookup without a second JOIN to the ledger table.
type Account struct {
	data.BaseModel
	BookID           *string           `gorm:"type:varchar(50)"                     json:"book_id"`
	Currency         string            `gorm:"type:varchar(10)"                     json:"currency"`
	Balance          *decimalx.Decimal `gorm:"-"                                    json:"balance"`
	UnClearedBalance *decimalx.Decimal `gorm:"-"                                    json:"un_cleared_balance"`
	ReservedBalance  *decimalx.Decimal `gorm:"-"                                    json:"reserved_balance"`
	LedgerID         string            `gorm:"type:varchar(50)"                     json:"ledger_id"`
	Data             data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
	LedgerType       string            `gorm:"type:varchar(50)"                     json:"ledger_type"`
	AccountType      string            `gorm:"type:varchar(50);not null"            json:"account_type"`
	NormalBalance    string            `gorm:"type:varchar(10);not null"            json:"normal_balance"`
}

// BeforeCreate fills in conventional defaults for AccountType,
// NormalBalance and BookID, then delegates to the embedded BaseModel
// hook which sets ID, CreatedAt and Version. Centralising the defaults on
// the model means every caller — business layer, direct repository writes,
// admin tools — gets a row that satisfies the NOT NULL + CHECK constraints
// without having to know the chart-of-accounts convention.
//
// BookID is denormalised from the parent Ledger at creation time so the
// hot posting path can validate cross-book consistency from one account
// lookup. The business layer is expected to pre-populate BookID before
// calling Create; if the parent Ledger has a BookID and the account does
// not, callers should propagate it explicitly.
func (acc *Account) BeforeCreate(db *gorm.DB) error {
	if acc.AccountType == "" {
		acc.AccountType = AccountTypeFromLedgerType(acc.LedgerType)
	}
	if acc.NormalBalance == "" {
		acc.NormalBalance = NormalBalanceForAccountType(acc.AccountType)
	}
	return acc.BaseModel.BeforeCreate(db)
}

func (acc *Account) ToAPI() *ledgerv1.Account {
	accountBalance := decimalx.DerefOr(acc.Balance, decimalx.Zero())
	balance := utilmoney.ToMoney(acc.Currency, accountBalance)

	reservedBalanceAmt := decimalx.DerefOr(acc.ReservedBalance, decimalx.Zero())
	reservedBalance := utilmoney.ToMoney(acc.Currency, reservedBalanceAmt)

	unClearedBalanceAmt := decimalx.DerefOr(acc.UnClearedBalance, decimalx.Zero())
	unClearedBalance := utilmoney.ToMoney(acc.Currency, unClearedBalanceAmt)

	bookID := ""
	if acc.BookID != nil {
		bookID = *acc.BookID
	}
	return &ledgerv1.Account{
		Id:               acc.ID,
		Ledger:           acc.LedgerID,
		Balance:          balance,
		ReservedBalance:  reservedBalance,
		UnclearedBalance: unClearedBalance,
		Data:             acc.Data.ToProtoStruct(),
		AccountType:      ToAccountTypeProto(acc.AccountType),
		NormalBalance:    ToNormalBalanceProto(acc.NormalBalance),
		BookId:           bookID,
	}
}

// Reserved Data keys lifted into typed columns. Callers may continue to
// supply these via Transaction.Data (or Ledger.Data for BookID) to avoid
// a proto change; the values are extracted into indexed columns during
// TransactionFromAPI / CreateLedger and remain in the JSONB blob for
// backwards compatibility.
const (
	DataKeyIdempotencyKey = "idempotency_key"
	DataKeyExternalRef    = "external_ref"
	DataKeySource         = "source"
	DataKeyBookID         = "book_id"
)

func TransactionFromAPI(ctx context.Context, aTxn *ledgerv1.Transaction) *Transaction {
	dataMap := &data.JSONMap{}
	transaction := &Transaction{
		Currency:        aTxn.GetCurrencyCode(),
		TransactionType: aTxn.GetType().String(),
		Data:            dataMap.FromProtoStruct(aTxn.GetData()),
	}

	transaction.IdempotencyKey = StringFromJSON(transaction.Data, DataKeyIdempotencyKey)
	transaction.ExternalRef = StringFromJSON(transaction.Data, DataKeyExternalRef)
	transaction.Source = StringFromJSON(transaction.Data, DataKeySource)
	if bookID := StringFromJSON(transaction.Data, DataKeyBookID); bookID != "" {
		transaction.BookID = &bookID
	}

	transaction.GenID(ctx)
	transaction.ID = aTxn.GetId()

	// Parse transacted_at timestamp
	if aTxn.GetTransactedAt() != "" {
		if transactedAt, err := time.Parse(time.RFC3339, aTxn.GetTransactedAt()); err == nil {
			transaction.TransactedAt = transactedAt
		}
	}

	// Map the legacy `cleared` boolean onto the explicit Status + PostedAt
	// lifecycle. Callers that already use cleared=true expect immediate
	// settlement (posted, contributing to balance); cleared=false expects
	// the row to sit in pending until a clearance update transitions it.
	now := time.Now()
	if aTxn.GetCleared() {
		transaction.Status = TransactionStatusPosted
		transaction.ClearedAt = now
		transaction.PostedAt = &now
	} else {
		transaction.Status = TransactionStatusPending
	}

	// Convert entries
	if len(aTxn.GetEntries()) > 0 {
		transaction.Entries = make([]*TransactionEntry, len(aTxn.GetEntries()))
		for index, aEntry := range aTxn.GetEntries() {
			transaction.Entries[index] = TransactionEntryFromAPI(aEntry)
		}
	}

	return transaction
}

func (tx *Transaction) ToAPI() *ledgerv1.Transaction {
	apiEntries := make([]*ledgerv1.TransactionEntry, len(tx.Entries))
	for index, mEntry := range tx.Entries {
		apiEntries[index] = mEntry.ToAPI()
	}

	bookID := ""
	if tx.BookID != nil {
		bookID = *tx.BookID
	}
	reversedID := ""
	if tx.ReversedTransactionID != nil {
		reversedID = *tx.ReversedTransactionID
	}

	trx := &ledgerv1.Transaction{
		Id:                    tx.ID,
		CurrencyCode:          tx.Currency,
		Cleared:               tx.Status == TransactionStatusPosted,
		Data:                  tx.Data.ToProtoStruct(),
		Entries:               apiEntries,
		Status:                ToTransactionStatusProto(tx.Status),
		IdempotencyKey:        tx.IdempotencyKey,
		ExternalRef:           tx.ExternalRef,
		Source:                tx.Source,
		BookId:                bookID,
		ReversedTransactionId: reversedID,
	}

	// Convert transaction type.
	if txnType, ok := ledgerv1.TransactionType_value[tx.TransactionType]; ok {
		trx.Type = ledgerv1.TransactionType(txnType)
	}

	// Format timestamps in RFC3339 — proto carries them as strings to keep
	// the API serializer-agnostic.
	if !tx.TransactedAt.IsZero() {
		trx.TransactedAt = tx.TransactedAt.Format(time.RFC3339)
	}
	if tx.PostedAt != nil && !tx.PostedAt.IsZero() {
		trx.PostedAt = tx.PostedAt.Format(time.RFC3339)
	}
	if tx.VoidedAt != nil && !tx.VoidedAt.IsZero() {
		trx.VoidedAt = tx.VoidedAt.Format(time.RFC3339)
	}

	return trx
}

func TransactionEntryFromAPI(aEntry *ledgerv1.TransactionEntry) *TransactionEntry {
	amt := utilmoney.FromMoney(aEntry.GetAmount())
	return &TransactionEntry{
		AccountID: aEntry.GetAccountId(),
		Currency:  aEntry.GetAmount().GetCurrencyCode(),
		Amount:    &amt,
		Credit:    aEntry.GetCredit(),
	}
}

func (te *TransactionEntry) ToAPI() *ledgerv1.TransactionEntry {
	var amount *commonv1.Money
	if te.Amount != nil {
		amount = utilmoney.ToMoney(te.Currency, *te.Amount)
	}

	return &ledgerv1.TransactionEntry{
		Id:            te.ID,
		AccountId:     te.AccountID,
		TransactionId: te.TransactionID,
		Amount:        amount,
		Credit:        te.Credit,
	}
}

// Transaction represents a transaction in a ledger.
//
// IdempotencyKey, ExternalRef, and Source are persisted as indexed columns so
// retries from webhooks, queue replays, and at-least-once delivery paths
// resolve to a single posting. IdempotencyKey enforces uniqueness at the DB
// layer (partial UNIQUE INDEX where the value is non-empty); ExternalRef and
// Source are non-unique search keys for reconciliation.
//
// Status drives the lifecycle. PostedAt is set when status transitions to
// 'posted'; VoidedAt when it transitions to 'voided'. ReversedTransactionID
// points back at the original NORMAL transaction this row offsets (set only
// on REVERSAL rows). ClearedAt is retained alongside Status for backward
// compatibility — both move together when a transaction is posted.
type Transaction struct {
	data.BaseModel
	BookID                *string             `gorm:"type:varchar(50)"                     json:"book_id"`
	Currency              string              `gorm:"type:varchar(10);not null"            json:"currency"`
	TransactionType       string              `gorm:"type:varchar(50)"                     json:"transaction_type"`
	Status                string              `gorm:"type:varchar(20);not null"            json:"status"`
	IdempotencyKey        string              `gorm:"type:varchar(120)"                    json:"idempotency_key"`
	ExternalRef           string              `gorm:"type:varchar(120)"                    json:"external_ref"`
	Source                string              `gorm:"type:varchar(50)"                     json:"source"`
	ReversedTransactionID *string             `gorm:"type:varchar(50)"                     json:"reversed_transaction_id"`
	Data                  data.JSONMap        `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
	ClearedAt             time.Time           `gorm:"type:timestamp"                       json:"cleared_at"`
	PostedAt              *time.Time          `gorm:"type:timestamp"                       json:"posted_at"`
	VoidedAt              *time.Time          `gorm:"type:timestamp"                       json:"voided_at"`
	TransactedAt          time.Time           `gorm:"type:timestamp"                       json:"transacted_at"`
	Entries               []*TransactionEntry `gorm:"foreignKey:TransactionID"             json:"entries"`
}

// TransactionEntry represents a transaction line in a ledger.
type TransactionEntry struct {
	data.BaseModel
	AccountID     string            `gorm:"type:varchar(50);not null;index" json:"account_id"`
	TransactionID string            `gorm:"type:varchar(50);not null;index" json:"transaction_id"`
	Currency      string            `gorm:"type:varchar(10);not null;index" json:"currency"`
	Amount        *decimalx.Decimal `gorm:"type:numeric(29,9)"              json:"amount"`
	Credit        bool              `                                       json:"credit"`
	ClearedAt     time.Time         `gorm:"-"                               json:"cleared_at"`
	TransactedAt  time.Time         `gorm:"-"                               json:"transacted_at"`
}

func (te *TransactionEntry) Equal(ot TransactionEntry) bool {
	return te.AccountID == ot.AccountID && te.Credit == ot.Credit &&
		te.Amount != nil && ot.Amount != nil &&
		te.Amount.Equal(*ot.Amount)
}

// StringFromJSON safely extracts a string value from a JSONMap. Returns ""
// when the key is missing or the value is not a string.
func StringFromJSON(m data.JSONMap, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// entryCurrency returns the currency to attribute an entry to for per-currency
// balance checks. Falls back to the parent transaction currency when the entry
// itself carries none (e.g. legacy rows or in-flight constructions before the
// entry currency is set).
func (tx *Transaction) entryCurrency(entry *TransactionEntry) string {
	if entry != nil && entry.Currency != "" {
		return entry.Currency
	}
	return tx.Currency
}

// IsZeroSum validates that entries net to zero independently within each
// currency. A multi-currency transaction must balance per-currency: USD
// debits = USD credits, UGX debits = UGX credits, etc. Summing across
// currencies would mask a genuinely unbalanced transaction.
func (tx *Transaction) IsZeroSum() bool {
	sums := map[string]decimalx.Decimal{}
	for _, entry := range tx.Entries {
		currency := tx.entryCurrency(entry)
		amount := decimalx.DerefOr(entry.Amount, decimalx.Zero())
		sum := sums[currency]
		if entry.Credit {
			sum = sum.Add(amount)
		} else {
			sum = sum.Sub(amount)
		}
		sums[currency] = sum
	}
	for _, sum := range sums {
		if !sum.IsZero() {
			return false
		}
	}
	return true
}

// IsTrueDrCr validates that each currency present in the entries has at
// least one debit and at least one credit. A single-side currency group
// cannot be a valid double-entry posting even if the overall sum is zero.
func (tx *Transaction) IsTrueDrCr() bool {
	type sides struct{ debits, credits int }
	perCurrency := map[string]*sides{}
	for _, entry := range tx.Entries {
		currency := tx.entryCurrency(entry)
		s, ok := perCurrency[currency]
		if !ok {
			s = &sides{}
			perCurrency[currency] = s
		}
		if entry.Credit {
			s.credits++
		} else {
			s.debits++
		}
	}
	if len(perCurrency) == 0 {
		return false
	}
	for _, s := range perCurrency {
		if s.debits < 1 || s.credits < 1 {
			return false
		}
	}
	return true
}
