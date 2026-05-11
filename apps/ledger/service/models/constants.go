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
