# Integration Patterns

How upstream services (billing, payment, webhooks) should call the
ledger service safely.

## 1. Idempotent webhook posting

The webhook may redeliver. Your queue may replay. Network errors may
cause retries. Without idempotency, every retry duplicates the books.

**Pattern**: every posting carries an `idempotency_key` derived from
the upstream system's natural id:

```go
key := fmt.Sprintf("webhook:%s:%s", provider, providerTxnID)

client.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
    Id: util.IDString(),  // any unique id; the dedup is via the key
    Currency: "UGX",
    Type: ledgerv1.TransactionType_NORMAL,
    Cleared: true,
    Data: structFromMap(map[string]string{
        "idempotency_key": key,
        "external_ref":    providerTxnID,
        "source":          provider,
    }),
    Entries: entries,
})
```

What happens under concurrent delivery of the same webhook:

- The first request wins. Inserts the transaction with `idempotency_key = K`.
- The second sees a UNIQUE-constraint violation on `idempotency_key`.
- The handler catches it, fetches the existing transaction by key, and
  returns it to the caller as if the second call had been the original.
- Both callers see the same Transaction.ID and entries.

If two requests with the **same idempotency_key but different entries**
arrive (caller bug), the second is rejected with `AlreadyExists` and an
"idempotency_key reused with different entries" message.

## 2. Two-phase settlement (pending → posted)

For asynchronous settlement (mobile money, bank transfer), do not commit
to balance until the provider confirms. Use `cleared=false` initially,
then transition via `UpdateTransaction` with `cleared_at`:

```go
// 1. User initiates a deposit. Provider returns "pending".
client.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
    Id: depositID,
    Currency: "UGX",
    Type: ledgerv1.TransactionType_NORMAL,
    Cleared: false,  // → status=PENDING; contributes to un_cleared_balance
    Data: idempotency(...),
    Entries: entries,
})

// 2. Provider confirms via webhook.
client.UpdateTransaction(ctx, &ledgerv1.UpdateTransactionRequest{
    Id: depositID,
    ClearedAt: time.Now().Format("2006-01-02T15:04:05.999999999"),
})
// → status=POSTED, contributes to balance.

// 2'. Or provider rejects via webhook.
client.MarkTransactionFailed(ctx, &ledgerv1.MarkTransactionFailedRequest{
    Id: depositID,
})
// → status=FAILED, no balance impact.
```

## 3. Reversal (correcting a posted transaction)

You cannot delete or edit a posted transaction — accounting truth must
preserve the audit trail. To correct, post a REVERSAL:

```go
client.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
    Id: "deposit-001",
})
// Creates "deposit-001_REVERSAL" with entries flipped (debit ↔ credit).
// Original transitions to status=REVERSED in the same DB transaction.
```

Net balance impact is zero (original +X is cancelled by REVERSAL -X)
but both rows stay visible in the journal for the audit trail.

## 4. Cross-book settlements

A transaction lives in exactly one book. To move money from one book to
another (e.g. a member's customer book pays a fee to the platform book),
post **two separate transactions** linked by `external_ref`:

```go
settlementRef := "settlement-2026-05-11-batch-42"

// 1. Customer book: debit Customer Wallet, credit Settlement Out.
client.CreateTransaction(ctx, txnIn(customerBookID, settlementRef))

// 2. Platform book: debit Settlement In, credit Fee Income.
client.CreateTransaction(ctx, txnOut(platformBookID, settlementRef))
```

`SearchTransactions` on `external_ref` returns the linked pair.

## 5. Handling concurrent same-account postings

The ledger is designed for concurrent writes. Atomicity is enforced at
the DB layer:

- Parent transaction + child entries commit or roll back together.
- Status transitions (pending → posted, posted → reversed, etc.) use
  CAS-style WHERE clauses so racing updates lose cleanly without
  producing half-applied state.
- Balance reads are computed live from entries via a LATERAL subquery,
  so they reflect the latest committed state of every concurrent
  posting.

You do **not** need application-level locking for ledger operations.

## 6. Audit trail

Every row carries:

- `created_at` / `created_by` — when and by whom (from auth claims)
- `modified_at` / `modified_by` — last update
- `version` — optimistic-lock counter incremented on every save
- Status timestamps: `posted_at` / `voided_at` mark the lifecycle moves
- `data` JSONB — caller-supplied metadata (reconciliation tags, refs)

Combined with the always-on tenancy columns (`tenant_id` /
`partition_id` / `access_id`), this gives a complete audit trail of who
posted what when from where.

## 7. Backpressure and rate

The ledger is built on a worker-pool pattern for searches and a
connection-pooled GORM stack for writes. Realistic throughput on a
modest Postgres:

- Single-node writes: ~500–1,500 atomic postings/sec (each = 1 parent + N
  entries) depending on entry count and DB latency.
- Reads: trivially scaled via read replicas — `pool.DB(ctx, true)` routes
  to replicas when configured.

For higher write throughput, batch transactions on the caller side
where possible (one Transaction with many entries beats many one-entry
transactions).

## 8. Reconciliation

Three patterns:

### a) External-ref reconciliation

Provider sends a daily settlement file with their transaction ids.
For each row:

```go
existing, _ := client.SearchTransactions(ctx, &commonv1.SearchRequest{
    Query: fmt.Sprintf(`{"must":{"fields":[{"external_ref":{"eq":"%s"}}]}}`, providerRef),
})
// If absent, investigate the gap.
// If present, confirm the amount matches.
```

### b) Trial-balance integrity check

Run `GetTrialBalance(currency=...)` on a schedule. Every currency's
`IsBalanced` must be `true`. If it isn't, the books are corrupt — page
oncall.

### c) Balance-from-source check

For each cash/bank/mobile-money account, compare the ledger's computed
balance against the provider's reported balance. The diff is your
unposted activity (which should be small and matched against the
provider's pending queue).

## 9. Reservations

For pending authorisations (e.g. card holds), use `transaction_type =
RESERVATION`. These are single-entry transactions that contribute to
`reserved_balance` only — they don't affect `balance` or
`un_cleared_balance`. When the authorisation completes, post a normal
transaction and (optionally) reverse the reservation.
