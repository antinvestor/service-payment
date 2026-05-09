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

import "errors"

// Business validation errors.
var (
	// Ledger errors.
	ErrLedgerReferenceRequired = errors.New("ledger reference is required")
	ErrLedgerIDRequired        = errors.New("ledger ID is required")
	ErrInvalidLedgerType       = errors.New("invalid ledger type returned from repository")

	// Account errors.
	ErrAccountReferenceRequired = errors.New("account reference is required")
	ErrAccountIDRequired        = errors.New("account ID is required")
	ErrAccountLedgerIDRequired  = errors.New("account ledger ID is required")
	ErrAccountCurrencyInvalid   = errors.New("account currency is invalid")
	ErrAccountNotFound          = errors.New("account not found")
	ErrInvalidAccountType       = errors.New("invalid account type returned from repository")

	// Transaction errors.
	ErrTransactionReferenceRequired      = errors.New("transaction reference is required")
	ErrTransactionIDRequired             = errors.New("transaction ID is required")
	ErrTransactionAccountIDRequired      = errors.New("transaction account ID is required")
	ErrTransactionCurrencyRequired       = errors.New("transaction currency is required")
	ErrTransactionNonZeroSum             = errors.New("transaction has non-zero sum")
	ErrTransactionInvalidDrCrEntry       = errors.New("transaction has invalid debit/credit entry")
	ErrTransactionEntriesNotFound        = errors.New("transaction entries not found")
	ErrTransactionEntryZeroAmount        = errors.New("transaction entry has zero amount")
	ErrTransactionAccountNotFound        = errors.New("transaction account not found")
	ErrTransactionAccountsDifferCurrency = errors.New("transaction accounts have different currencies")
	ErrInvalidTransactionType            = errors.New("invalid transaction type returned from repository")

	// General errors.
	ErrInvalidSearchResult = errors.New("invalid search result type from repository")
)
