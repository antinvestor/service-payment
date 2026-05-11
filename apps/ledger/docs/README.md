# Ledger Service

Production-grade double-entry accounting service for Antinvestor. Records
the financial truth of every transaction across many independent
accounting books (platform, group, customer, merchant, agent, branch),
with strict integrity rules and a complete state machine for every
posting.

This directory is the operator and integrator handbook. Read in order:

1. [concepts.md](concepts.md) — Books / Ledgers / Accounts / Transactions /
   Entries. The state machine. DEADCLIC and how amounts are stored.
   What "balanced" means.
2. [rpcs.md](rpcs.md) — every RPC, what it does, when to use it, and
   complete request / response examples.
3. [integration.md](integration.md) — webhook integration patterns:
   idempotency, retries, settlement timing, reversal, voiding.
4. [reports.md](reports.md) — trial balance, account statement, exports
   (CSV / Excel / PDF) and the consolidated-hierarchy report pattern.
5. UI: [ui/ledger/README.md](../../../ui/ledger/README.md) — what every
   screen does, what permissions it needs, and what action it produces.

## At a glance

```
Tenant / Partition
  └── Book(s)                          (independent accounting scope)
       └── Ledger(s) [chart-of-accounts root]
            └── Ledger(s) [sub-categories]
                 └── Account(s)         (the actual posting bucket)
                      └── TransactionEntry(s)   (one debit or credit line)
                           ↑
                           └── Transaction      (groups entries; lives in one Book)
```

| Concept | Role | Owns | Lives in |
|---|---|---|---|
| **Book** | Independent accounting scope | Many Ledgers | Tenant / Partition |
| **Ledger** | Chart-of-accounts node (asset / liability / income / …) | Many Accounts; many child Ledgers | Book |
| **Account** | Posting bucket (cash, member savings, loan principal, …) | Many TransactionEntries | Ledger (and through it, Book) |
| **Transaction** | Journal voucher (debit + credit lines) | Many TransactionEntries | Book |
| **TransactionEntry** | A single debit or credit line | n/a | Transaction → Account |

## Invariants the system enforces

For every NORMAL or REVERSAL transaction:

1. `sum(debits) == sum(credits)` **per currency** (no cross-currency
   cancellation can hide an unbalanced posting).
2. At least one debit and at least one credit per currency.
3. Parent + entries commit atomically (no orphan transactions even on
   crash mid-write).
4. Cross-book postings are rejected — when `transaction.book_id` is set,
   every entry's account must belong to that same book.
5. Idempotency: a unique `idempotency_key` ensures a retry — even with
   a fresh `Transaction.ID` — resolves to a single stored row.
6. Once posted, the transaction is immutable except for the auto-
   transition to `reversed` when an offsetting REVERSAL is posted.

## Status lifecycle

```
                       Create
   ┌──────────┐  cleared=false   ┌─────────┐  clearance    ┌────────┐
   │  caller  │ ───────────────▶ │ PENDING │ ────────────▶ │ POSTED │
   └──────────┘                   └─────────┘   update     └────────┘
        │                              │                       │
        │   cleared=true               │                       │  Reverse
        └──────────────────────────────┘                       │  (creates
                                       │                       │   REVERSAL)
                                       │                       ▼
                                   void RPC               ┌──────────┐
                                       │                  │ REVERSED │
                                       ▼                  └──────────┘
                                  ┌────────┐
                                  │ VOIDED │   markFailed RPC
                                  └────────┘   from PENDING
                                                       │
                                                       ▼
                                                  ┌────────┐
                                                  │ FAILED │
                                                  └────────┘
```

- **DRAFT**: created but not submitted; mutable; no balance impact.
- **PENDING**: submitted, awaiting settlement; contributes to
  `un_cleared_balance`. Can transition to POSTED via a clearance
  update, VOIDED via VoidTransaction, or FAILED via
  MarkTransactionFailed.
- **POSTED**: confirmed; contributes to `balance`; immutable except for
  the auto-flip to REVERSED when a REVERSAL is posted against it.
- **REVERSED**: was POSTED, then offset by a REVERSAL whose entries
  cancel its balance impact. Stays in the journal for the audit trail.
- **VOIDED**: cancelled before posting; no balance impact; terminal.
- **FAILED**: rejected by upstream; no balance impact; terminal.

## Quick start

```go
// 1. Create a book for an organisation.
org, _ := client.CreateBook(ctx, &ledgerv1.CreateBookRequest{
    Name: "Stawi Organization", Type: "platform", Currency: "UGX",
})

// 2. Create a ledger (chart-of-accounts node) inside the book.
client.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{
    Id: "1000-cash", Type: ledgerv1.LedgerType_ASSET,
    Data: structFromMap(map[string]string{"book_id": org.GetData().GetId()}),
})

// 3. Create an account under that ledger.
client.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
    Id: "cash-on-hand", LedgerId: "1000-cash", Currency: "UGX",
})

// 4. Post a balanced transaction.
client.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
    Id: "deposit-001", Currency: "UGX",
    Type: ledgerv1.TransactionType_NORMAL, Cleared: true,
    Data: structFromMap(map[string]string{
        "idempotency_key": "mtn-webhook-abc123",
        "book_id":         org.GetData().GetId(),
    }),
    Entries: []*ledgerv1.TransactionEntry{
        {AccountId: "cash-on-hand", Credit: false, Amount: ugx(100_000)},
        {AccountId: "member-savings", Credit: true, Amount: ugx(100_000)},
    },
})

// 5. Verify the books.
tb, _ := client.GetTrialBalance(ctx, &ledgerv1.GetTrialBalanceRequest{
    Currency: "UGX",
})
// tb.Totals[0].IsBalanced == true if debits = credits per currency.
```

## Where to read code

| Layer | Path |
|---|---|
| Proto | `proto/ledger/v1/ledger.proto` |
| Handlers (Connect/gRPC) | `apps/ledger/service/handlers/` |
| Business (rules + state machine) | `apps/ledger/service/business/` |
| Repository (SQL + atomicity) | `apps/ledger/service/repository/` |
| Models (Go types + ToAPI) | `apps/ledger/service/models/` |
| Migrations | `apps/ledger/migrations/0001/` |
| UI (Flutter) | `ui/ledger/` |
