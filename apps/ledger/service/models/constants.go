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

const (
	LedgerTypeAsset     = "ASSET"
	LedgerTypeExpense   = "EXPENSE"
	LedgerTypeLiability = "LIABILITY"
	LedgerTypeIncome    = "INCOME"
	LedgerTypeCapital   = "CAPITAL"
)

// Book classifies who the accounting scope belongs to. Conventional values
// — open-ended because the platform model may grow new entity types.
const (
	BookTypePlatform = "platform"
	BookTypeGroup    = "group"
	BookTypeCustomer = "customer"
	BookTypeMerchant = "merchant"
	BookTypeAgent    = "agent"
	BookTypeBranch   = "branch"
)

// Per-account classification, more granular than LedgerType so contra,
// clearing, suspense and memo accounts can sit under their natural parent
// ledgers without overloading the parent's type. Values map 1:1 to the
// accounts_account_type_check CHECK constraint.
const (
	AccountTypeAsset           = "asset"
	AccountTypeLiability       = "liability"
	AccountTypeEquity          = "equity"
	AccountTypeIncome          = "income"
	AccountTypeExpense         = "expense"
	AccountTypeContraAsset     = "contra_asset"
	AccountTypeContraLiability = "contra_liability"
	AccountTypeContraIncome    = "contra_income"
	AccountTypeContraExpense   = "contra_expense"
	AccountTypeClearing        = "clearing"
	AccountTypeSuspense        = "suspense"
	AccountTypeMemo            = "memo"
)

// Normal balance side. Maps 1:1 to the accounts_normal_balance_check
// CHECK constraint. DEADCLIC stores +amount when the entry side equals
// the account's normal balance, -amount otherwise. Memo accounts use
// 'none' to opt out of sign normalization entirely.
const (
	NormalBalanceDebit  = "debit"
	NormalBalanceCredit = "credit"
	NormalBalanceNone   = "none"
)

// AccountTypeFromLedgerType returns the conventional account_type for a
// given parent ledger_type. Returns empty string for unknown inputs so
// callers can decide whether to require explicit classification.
func AccountTypeFromLedgerType(lt string) string {
	switch lt {
	case LedgerTypeAsset:
		return AccountTypeAsset
	case LedgerTypeLiability:
		return AccountTypeLiability
	case LedgerTypeIncome:
		return AccountTypeIncome
	case LedgerTypeExpense:
		return AccountTypeExpense
	case LedgerTypeCapital:
		return AccountTypeEquity
	}
	return ""
}

// NormalBalanceForAccountType returns the conventional normal balance side
// for a given account_type. Returns empty string for unknown inputs.
func NormalBalanceForAccountType(at string) string {
	switch at {
	case AccountTypeAsset, AccountTypeExpense,
		AccountTypeContraLiability, AccountTypeContraIncome,
		AccountTypeClearing:
		return NormalBalanceDebit
	case AccountTypeLiability, AccountTypeEquity, AccountTypeIncome,
		AccountTypeContraAsset, AccountTypeContraExpense,
		AccountTypeSuspense:
		return NormalBalanceCredit
	case AccountTypeMemo:
		return NormalBalanceNone
	}
	return ""
}

// NormalBalanceFromLedgerType is a convenience for legacy paths that have
// LedgerType but not AccountType. Same conventions as the
// AccountTypeFromLedgerType + NormalBalanceForAccountType pipeline.
func NormalBalanceFromLedgerType(lt string) string {
	return NormalBalanceForAccountType(AccountTypeFromLedgerType(lt))
}

// Transaction lifecycle statuses. Values map 1:1 to the
// transactions_status_check CHECK constraint in the database.
//
//	draft    → created but not submitted; mutable; no balance impact
//	pending  → submitted, awaiting settlement; contributes to
//	           un_cleared_balance; can transition to posted | failed | voided
//	posted   → confirmed/settled; contributes to balance; immutable except
//	           for the auto-transition to reversed when a REVERSAL is posted
//	reversed → was posted, then offset by a REVERSAL transaction whose
//	           entries cancel this one's balance impact; immutable
//	voided   → was draft or pending, then cancelled administratively;
//	           no balance impact; terminal
//	failed   → posting attempt rejected by external system; no balance
//	           impact; terminal
const (
	TransactionStatusDraft    = "draft"
	TransactionStatusPending  = "pending"
	TransactionStatusPosted   = "posted"
	TransactionStatusReversed = "reversed"
	TransactionStatusVoided   = "voided"
	TransactionStatusFailed   = "failed"
)
