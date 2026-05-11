# Concepts

Read this once. Every other doc assumes you understand it.

## 1. Book — independent accounting scope

A **Book** is one entity's complete set of financial records. Each book
has its own chart of accounts, its own trial balance, its own balance
sheet. Entries posted in one book must never cross-contaminate another.

Conventional types (any string is allowed; these are platform defaults):

| Type | Use |
|---|---|
| `platform` | The company's own books — fee revenue, infra cost, retained earnings |
| `group` | A savings group / chama / SACCO / branch (one per group) |
| `customer` | An individual member's personal book |
| `merchant` | A merchant's settlement book |
| `agent` | An agent's float book |
| `branch` | A physical branch's books |

### Hierarchy

A book can have a `parent_id` pointing at another book. This is for
**consolidated reporting** (organisation → groups → members), not for
posting. Cross-book postings are rejected regardless of hierarchy —
settlements that genuinely cross books are modeled as two separate
transactions linked by `external_ref`.

```
Stawi Organization (platform)
  ├─ Group A (group)
  │    ├─ Member Jane (customer)
  │    └─ Member John (customer)
  └─ Group B (group)
       └─ Member Mary (customer)
```

A trial balance scoped to "Stawi Organization" with descendants expansion
rolls up activity from Group A + Group B + Jane + John + Mary into one
report. Scoped to just "Group A" returns only Group A + its two members.

## 2. Ledger — chart-of-accounts node

A **Ledger** is a classification bucket. Five fundamental types:

| Type | Normal balance | Examples |
|---|---|---|
| `ASSET` | debit | Cash, Bank, Mobile Money, Loan Principal, Interest Receivable |
| `LIABILITY` | credit | Member Savings, Vendor Payables, Customer Wallet |
| `INCOME` | credit | Loan Interest Income, Fee Income, Penalty Income |
| `EXPENSE` | debit | SMS Costs, Cloud Infra, Bad Debt, Provider Fees |
| `CAPITAL` | credit | Share Capital, Retained Earnings, Current Year Earnings |

Ledgers can have a `parent_id` for hierarchy:

```
Assets (ASSET)
  ├─ Cash and Bank (ASSET)
  │    ├─ Cash on Hand
  │    └─ MTN Mobile Money Float
  ├─ Loans Receivable (ASSET)
  │    └─ Loan Principal Receivable
```

A **Ledger** belongs to a **Book** (via `book_id`).

## 3. Account — the posting bucket

An **Account** is what entries actually post against. It carries:

| Field | Meaning |
|---|---|
| `id` | Stable reference (e.g. `member-123-savings`) |
| `ledger_id` | Parent chart-of-accounts node |
| `book_id` | Denormalised from parent ledger for cross-book validation |
| `account_type` | Finer than `LedgerType`. Includes contra/clearing/suspense/memo. |
| `normal_balance` | `debit` / `credit` / `none`. Drives DEADCLIC. |
| `currency` | ISO 4217. Every account is single-currency. |

### Extended account types (beyond the five LedgerTypes)

| account_type | normal_balance | Meaning |
|---|---|---|
| `contra_asset` | credit | Allowance for doubtful debts, accumulated depreciation |
| `contra_liability` | debit | Reduction of a liability (rare) |
| `contra_income` | debit | Sales returns, discounts |
| `contra_expense` | credit | Expense recoveries |
| `clearing` | debit | Settlement clearing, MTN clearing |
| `suspense` | credit | Unallocated funds awaiting reconciliation |
| `memo` | none | Statistical accounts (member count etc.) — no balance contribution |

`AccountType` and `NormalBalance` default from the parent ledger's type
when the account is created. Override them explicitly for contra/
clearing/suspense/memo accounts.

## 4. Transaction — the journal voucher

A **Transaction** is a single accounting event: a list of debit and
credit lines that together must balance. Always has at least two entries
(one debit + one credit) for NORMAL and REVERSAL types.

```
Transaction: deposit-001  (NORMAL, posted)
  Entry 1:  DEBIT   cash-on-hand        UGX 100,000
  Entry 2:  CREDIT  member-savings      UGX 100,000
```

Transaction types:

| Type | When |
|---|---|
| `NORMAL` | Standard postings. Must zero-sum per currency. |
| `REVERSAL` | Posted to offset a previous NORMAL. Carries `reversed_transaction_id`. |
| `RESERVATION` | Temporary hold (e.g. pending authorisation). Single entry, contributes to `reserved_balance` only. |

### Idempotency

Every Transaction can carry these typed control fields (extracted from
`Data` for backward compat until clients adopt the typed proto fields):

| Field | Meaning |
|---|---|
| `idempotency_key` | Caller's dedup token (e.g. `webhook:mtn:abc123`). Partial UNIQUE at DB. |
| `external_ref` | Provider/external reference (mobile money txn id, bank stmt id). |
| `source` | Origin (`mtn`, `airtel`, `manual_admin`, `internal_billing`, …). |

Concurrent retries with the same `idempotency_key` (even with different
`Transaction.ID`) all converge on the single stored row. Mismatched
entries with the same key are rejected with `ErrTransactionIsConflicting`.

## 5. TransactionEntry — one debit or credit line

```
TransactionEntry
  ├─ account_id        (where the entry posts)
  ├─ transaction_id    (parent voucher)
  ├─ amount            (always stored DEADCLIC-signed; see below)
  ├─ credit            (true = credit side, false = debit side)
  ├─ currency          (persisted at the entry level; defends multi-currency)
```

## 6. DEADCLIC and stored signs

DEADCLIC is the rule that determines whether a debit increases or
decreases a given account:

```
   ┌────────────────────────────┐  ┌────────────────────────────┐
   │       Debit increases      │  │      Credit increases       │
   ├────────────────────────────┤  ├────────────────────────────┤
   │ • Expenses                 │  │ • Liabilities              │
   │ • Assets                   │  │ • Income                   │
   │ • Drawings                 │  │ • Capital / Equity         │
   └────────────────────────────┘  └────────────────────────────┘
```

The ledger stores `amount` already sign-adjusted: positive when the
entry is on the account's **normal balance** side, negative when on the
opposite side. This means a simple `SUM(amount)` over an account's
entries returns the account's natural balance.

Example: a cash account (ASSET, normal_balance=debit).

```
DEBIT  100  →  stored as +100   (entry side matches normal → positive)
CREDIT  30  →  stored as  -30   (entry side opposite normal → negative)
                                 ────────
SUM(amount) = +70 = cash balance
```

Reversal of a `DEBIT 100` flips the Credit flag (now true) and
DEADCLIC stores -100. The original's +100 + the reversal's -100 nets
to 0. Both rows stay in the books.

## 7. Status lifecycle (state machine)

| State | Balance impact | Mutable | Transitions to |
|---|---|---|---|
| DRAFT | none | yes | PENDING (submit), VOIDED |
| PENDING | un_cleared_balance | metadata only | POSTED (clearance), VOIDED, FAILED |
| POSTED | balance | no | REVERSED (auto, via REVERSAL) |
| REVERSED | balance (still counted, offset by REVERSAL) | no | terminal |
| VOIDED | none | no | terminal |
| FAILED | none | no | terminal |

### Backward compatibility

The legacy `Transaction.cleared` boolean maps to status:
- `cleared=true` at create time ⇒ `status = POSTED`, `posted_at = now`.
- `cleared=false` at create time ⇒ `status = PENDING`.
- API consumers reading the legacy field get `cleared = (status == POSTED)`.

## 8. Multi-currency

Each account is single-currency. A transaction can carry entries in
multiple currencies — `IsZeroSum` then balances **per currency**, not
across the whole transaction. The system rejects:

- A transaction with USD debits and UGX credits that "cancel" numerically
  (they wouldn't actually balance the books).
- A multi-currency transaction where one currency's debits don't equal
  its credits.

In practice most transactions are single-currency. The per-currency
check is a safety net against subtle multi-currency bugs.

## 9. Tenancy

Every persisted row carries `tenant_id`, `partition_id`, `access_id`
inherited from the auth claims in the request context. The framework's
`TenancyPartition` GORM scope automatically appends
`WHERE tenant_id = ? AND partition_id = ?` to every query — operators
of tenant A never see tenant B's books.

Cross-tenant reporting (e.g. system-wide reconciliation) is supported
by services that explicitly opt out of the tenancy scope.

## 10. What the system does NOT do

By design:

- **No silent void of a posted transaction.** Posted activity must be
  reversed so the audit trail records the offset.
- **No cross-book posting.** A transaction is in exactly one book;
  cross-book settlements are modeled as two linked transactions.
- **No mutable entry amounts.** Once written, an entry is immutable.
  Corrections happen via reversal + repost.
- **No business-intent abstractions in the API.** The system records
  accounting truth; the calling service (billing, payments) composes
  business intents (deposit / withdraw / charge fee) into the
  appropriate debit / credit entries.
