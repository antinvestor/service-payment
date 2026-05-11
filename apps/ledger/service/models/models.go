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
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/moneyx"
)

// Ledger represents the hierarchy for organising ledgers with information such as type, and JSON data.
type Ledger struct {
	data.BaseModel
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

func (lg *Ledger) ToAPI() *ledgerv1.Ledger {
	return &ledgerv1.Ledger{Id: lg.ID, Type: ToLedgerType(lg.Type),
		Parent: lg.ParentID, Data: lg.Data.ToProtoStruct()}
}

// Account represents the ledger account with information such as Reference, balance and JSON data.
type Account struct {
	data.BaseModel
	Currency         string            `gorm:"type:varchar(10)"                     json:"currency"`
	Balance          *decimalx.Decimal `gorm:"-"                                    json:"balance"`
	UnClearedBalance *decimalx.Decimal `gorm:"-"                                    json:"un_cleared_balance"`
	ReservedBalance  *decimalx.Decimal `gorm:"-"                                    json:"reserved_balance"`
	LedgerID         string            `gorm:"type:varchar(50)"                     json:"ledger_id"`
	Data             data.JSONMap      `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
	LedgerType       string            `gorm:"type:varchar(50)"                     json:"ledger_type"`
}

func (acc *Account) ToAPI() *ledgerv1.Account {
	accountBalance := decimalx.DerefOr(acc.Balance, decimalx.Zero())
	balance := utilmoney.ToMoney(acc.Currency, accountBalance)

	reservedBalanceAmt := decimalx.DerefOr(acc.ReservedBalance, decimalx.Zero())
	reservedBalance := utilmoney.ToMoney(acc.Currency, reservedBalanceAmt)

	unClearedBalanceAmt := decimalx.DerefOr(acc.UnClearedBalance, decimalx.Zero())
	unClearedBalance := utilmoney.ToMoney(acc.Currency, unClearedBalanceAmt)

	return &ledgerv1.Account{
		Id: acc.ID, Ledger: acc.LedgerID,
		Balance: balance, ReservedBalance: reservedBalance, UnclearedBalance: unClearedBalance,
		Data: acc.Data.ToProtoStruct()}
}

// Reserved Data keys lifted into typed columns. Callers may continue to
// supply these via Transaction.Data to avoid a proto change; the values are
// extracted into indexed columns during TransactionFromAPI and remain in the
// JSONB blob for backwards compatibility.
const (
	dataKeyIdempotencyKey = "idempotency_key"
	dataKeyExternalRef    = "external_ref"
	dataKeySource         = "source"
)

func TransactionFromAPI(ctx context.Context, aTxn *ledgerv1.Transaction) *Transaction {
	dataMap := &data.JSONMap{}
	transaction := &Transaction{
		Currency:        aTxn.GetCurrencyCode(),
		TransactionType: aTxn.GetType().String(),
		Data:            dataMap.FromProtoStruct(aTxn.GetData()),
	}

	transaction.IdempotencyKey = stringFromJSON(transaction.Data, dataKeyIdempotencyKey)
	transaction.ExternalRef = stringFromJSON(transaction.Data, dataKeyExternalRef)
	transaction.Source = stringFromJSON(transaction.Data, dataKeySource)

	transaction.GenID(ctx)
	transaction.ID = aTxn.GetId()

	// Parse transacted_at timestamp
	if aTxn.GetTransactedAt() != "" {
		if transactedAt, err := time.Parse(time.RFC3339, aTxn.GetTransactedAt()); err == nil {
			transaction.TransactedAt = transactedAt
		}
	}

	// Set cleared_at if transaction is cleared
	if aTxn.GetCleared() {
		transaction.ClearedAt = time.Now()
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

	trx := &ledgerv1.Transaction{
		Id:           tx.ID,
		CurrencyCode: tx.Currency,
		Cleared:      !tx.ClearedAt.IsZero(),
		Data:         tx.Data.ToProtoStruct(),
		Entries:      apiEntries,
	}

	// Convert transaction type
	if txnType, ok := ledgerv1.TransactionType_value[tx.TransactionType]; ok {
		trx.Type = ledgerv1.TransactionType(txnType)
	}

	// Format transacted_at timestamp
	if !tx.TransactedAt.IsZero() {
		trx.TransactedAt = tx.TransactedAt.Format(time.RFC3339)
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
type Transaction struct {
	data.BaseModel
	Currency        string              `gorm:"type:varchar(10);not null"            json:"currency"`
	TransactionType string              `gorm:"type:varchar(50)"                     json:"transaction_type"`
	IdempotencyKey  string              `gorm:"type:varchar(120)"                    json:"idempotency_key"`
	ExternalRef     string              `gorm:"type:varchar(120)"                    json:"external_ref"`
	Source          string              `gorm:"type:varchar(50)"                     json:"source"`
	Data            data.JSONMap        `gorm:"type:jsonb;index:,gin:jsonb_path_ops" json:"data"`
	ClearedAt       time.Time           `gorm:"type:timestamp"                       json:"cleared_at"`
	TransactedAt    time.Time           `gorm:"type:timestamp"                       json:"transacted_at"`
	Entries         []*TransactionEntry `gorm:"foreignKey:TransactionID"             json:"entries"`
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

// stringFromJSON safely extracts a string value from a JSONMap. Returns ""
// when the key is missing or the value is not a string.
func stringFromJSON(m data.JSONMap, key string) string {
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
