# RPC Reference

Every LedgerService RPC, what it does, when to call it, complete request
and response shapes, and the failure modes. All examples use the
Go Connect client; the same shapes apply over HTTP/JSON via the
generated Connect handler.

## Books

### `CreateBook(CreateBookRequest) → CreateBookResponse`

Create a new accounting scope. Books with a `parent_id` participate in
the hierarchy used by consolidated reports.

```go
resp, err := client.CreateBook(ctx, &ledgerv1.CreateBookRequest{
    Name:     "Village Group A",
    Type:     "group",          // platform / group / customer / merchant / agent / branch
    ParentId: "<org book id>",   // optional; empty for top-level
    Currency: "UGX",             // optional default for postings
    Data: structFromMap(map[string]string{
        "registered_by": "agent-42",
    }),
})
// resp.Data.Id is the new book's identifier.
```

| Failure | Cause |
|---|---|
| `InvalidArgument` | name or type missing |
| `NotFound` | parent_id supplied but does not exist in the caller's tenancy |

### `GetBook(GetBookRequest) → GetBookResponse`

Fetch a single book by id. Returns NotFound when the id is outside the
caller's tenancy.

### `ListBooksByType(ListBooksByTypeRequest) → ListBooksByTypeResponse`

Returns all active books of the supplied conventional type within the
caller's tenancy, most-recently-created first.

```go
resp, _ := client.ListBooksByType(ctx, &ledgerv1.ListBooksByTypeRequest{
    Type: "group",
})
// resp.Data is []*Book
```

A type filter is required by design. Operators rarely want every book
across every classification at once — different scopes (platform vs.
group vs. merchant) are operationally distinct.

## Ledgers

### `CreateLedger(CreateLedgerRequest) → CreateLedgerResponse`

Add a node to a book's chart of accounts.

```go
client.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{
    Id:       "1000-cash",      // optional; xid auto-generated if empty
    Type:     ledgerv1.LedgerType_ASSET,
    ParentId: "",               // optional parent ledger for hierarchy
    Data: structFromMap(map[string]string{
        "book_id": "<book id>",  // book scope (until proto carries it natively)
        "name":    "Cash and Bank",
    }),
})
```

### `UpdateLedger / SearchLedgers`

Standard metadata update and free-text search.

## Accounts

### `CreateAccount(CreateAccountRequest) → CreateAccountResponse`

```go
client.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
    Id:       "member-123-savings",
    LedgerId: "2000-member-savings",
    Currency: "UGX",
    Data: structFromMap(map[string]string{
        "owner_id": "member-123",
    }),
})
```

`account_type` and `normal_balance` default from the parent ledger's
type. To create a contra/clearing/suspense/memo account, set them
explicitly via the future proto field (or via direct repository write
in admin tooling).

### `SearchAccounts / UpdateAccount`

Standard. Balances on the response are computed live from the LATERAL
balance query — `balance` (cleared activity), `uncleared_balance`
(pending activity), `reserved_balance` (RESERVATION transactions).

## Transactions

### `CreateTransaction(CreateTransactionRequest) → CreateTransactionResponse`

The core posting RPC. All accounting integrity checks happen here.

```go
client.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
    Id:       "deposit-mtn-abc123",
    Currency: "UGX",
    Type:     ledgerv1.TransactionType_NORMAL,
    Cleared:  true,             // immediate posting; status = POSTED
    Data: structFromMap(map[string]string{
        "idempotency_key": "webhook:mtn:abc123",  // dedup token
        "external_ref":    "MTN-TXN-72839",
        "source":          "mtn",
        "book_id":         "<book id>",
        "memo":            "Member deposit",
    }),
    Entries: []*ledgerv1.TransactionEntry{
        {AccountId: "mtn-float",       Credit: false, Amount: ugx(100_000)},
        {AccountId: "member-savings",  Credit: true,  Amount: ugx(100_000)},
    },
})
```

| Failure | Cause |
|---|---|
| `InvalidArgument` | Missing id or currency |
| `InvalidArgument` | Zero entry amount |
| `InvalidArgument` | Entries don't zero-sum per currency |
| `InvalidArgument` | No matching DR + CR per currency |
| `InvalidArgument` | Entry account currency mismatch with transaction |
| `NotFound` | Referenced account does not exist |
| `NotFound` | Cross-book posting (account belongs to different book than transaction) |
| `AlreadyExists` | Same Transaction.ID submitted previously |
| `AlreadyExists` | Same idempotency_key with different entries |

### `ReverseTransaction(ReverseTransactionRequest) → ReverseTransactionResponse`

Posts an offsetting REVERSAL whose entries cancel the original's
balance impact. The original auto-transitions to `REVERSED` in the same
DB transaction.

```go
client.ReverseTransaction(ctx, &ledgerv1.ReverseTransactionRequest{
    Id: "deposit-mtn-abc123",  // the original transaction's id
})
// Reversal's id is "{original-id}_REVERSAL".
// Reversal's reversed_transaction_id points back at the original.
```

| Failure | Cause |
|---|---|
| `FailedPrecondition` | Original is not type=NORMAL |
| `FailedPrecondition` | Original is not status=POSTED (pending → void instead) |
| `FailedPrecondition` | Concurrent reversal already won the CAS |

### `VoidTransaction(VoidTransactionRequest) → VoidTransactionResponse`

Move a DRAFT or PENDING transaction to VOIDED. No balance impact.
Stamps `voided_at`.

```go
client.VoidTransaction(ctx, &ledgerv1.VoidTransactionRequest{Id: "draft-001"})
```

Posted transactions cannot be voided — reverse them instead so the
books carry the audit trail.

### `MarkTransactionFailed(MarkTransactionFailedRequest) → MarkTransactionFailedResponse`

Move a PENDING transaction to FAILED. Use when the upstream payment
provider rejects the posting via webhook callback.

```go
client.MarkTransactionFailed(ctx, &ledgerv1.MarkTransactionFailedRequest{
    Id: "pending-mtn-001",
})
```

### `UpdateTransaction`

Update transaction metadata (Data fields) and trigger the pending →
posted transition via `cleared_at`. Posted transactions are immutable
except for free-form metadata (reconciliation tags etc.).

### `SearchTransactions / SearchEntries`

Free-text queries. Useful for reconciliation tooling. For statement
reports use `GetAccountStatement` instead — it carries running balance.

## Reports

### `GetTrialBalance(GetTrialBalanceRequest) → GetTrialBalanceResponse`

Per-account debit / credit totals plus per-currency grand totals with
the textbook `is_balanced` integrity flag.

```go
resp, _ := client.GetTrialBalance(ctx, &ledgerv1.GetTrialBalanceRequest{
    Currency:   "UGX",            // optional ISO 4217 filter
    LedgerId:   "",               // optional one-subtree scope
    LedgerType: "ASSET",          // optional case-sensitive filter (empty = all)
    BookIds:    []string{"<book id>"},
    AsOf:       "2026-12-31T23:59:59Z",
})

// resp.Lines is per-account rows.
// resp.Totals is per-currency totals — each carries IsBalanced.
```

For a consolidated report across an organisation's groups + members,
expand the root book id into its descendants client-side and pass them
all in `BookIds`. (The descendants helper is exposed on the Go business
layer; the UI does this in the trial-balance screen.)

### `GetAccountStatement(GetAccountStatementRequest) → GetAccountStatementResponse`

Customer-facing account ledger for a period.

```go
resp, _ := client.GetAccountStatement(ctx, &ledgerv1.GetAccountStatementRequest{
    AccountId: "member-123-savings",
    From:      "2026-01-01T00:00:00Z",
    To:        "2026-12-31T23:59:59Z",
    Limit:     100,
    Offset:    0,
})

// resp.OpeningBalance: sum before From.
// resp.ClosingBalance: opening + period activity.
// resp.Entries:  chronological; each carries RunningBalance.
// resp.TotalDebits / resp.TotalCredits: period magnitudes.
```

`Limit` defaults to 100 when ≤0 and caps at 1000. Page forward with
`Offset` to walk longer periods; the running balance stays consistent
because the opening balance carries in.

## Error handling

All RPCs return Connect errors with these codes:

| Connect code | Business reason |
|---|---|
| `InvalidArgument` | Missing required field, validation failure, invalid amount, currency mismatch |
| `NotFound` | Referenced ledger/account/transaction/book not in caller's tenancy |
| `AlreadyExists` | Duplicate Transaction.ID or idempotency_key conflict |
| `FailedPrecondition` | Illegal state transition (e.g. reverse a non-posted txn) |
| `Internal` | Database or system failure |

Inspect the `ApplicationError` payload for a stable error code that
clients can branch on (defined in `pkg/apperrors`).

## Permissions

Default role bindings declared by the proto service:

| Role | Permissions |
|---|---|
| Owner / Admin / Service | full read+manage on ledger, account, transaction, book, report |
| Operator | read-only ledger/account/transaction + transaction_manage + book_view + report_view |
| Viewer / Member | read-only across the board |
