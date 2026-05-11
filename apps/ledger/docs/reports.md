# Reports

The ledger derives every classical accounting report from the same
underlying `transaction_entries` table — no separate "trial balance" or
"statement" storage exists. This makes the reports always-fresh and
prevents the consistency drift you'd get with a denormalised reporting
store.

## Trial balance

The fundamental integrity check. For every currency in the books:

```
sum(debits)  ==  sum(credits)
```

If this fails, the books are corrupt. Run trial balance on a schedule
(daily, or after every batch of postings) and alert when `is_balanced`
is false.

### Filters

| Filter | Effect |
|---|---|
| `currency` | One ISO 4217 code |
| `ledger_id` | Scope to one chart-of-accounts subtree |
| `ledger_type` | "ASSET" / "LIABILITY" / "INCOME" / "EXPENSE" / "CAPITAL" |
| `book_ids` | One or more books (for consolidated reporting see below) |
| `as_of` | RFC3339 upper bound on `transacted_at` |

### Consolidated reporting (hierarchy)

A book can have children (organisation → groups → members). To get a
consolidated trial balance across an organisation's entire subtree:

1. Resolve the descendants: `bookBusiness.ListDescendantIDs(ctx, rootID)`
   returns the root plus every transitive child book id.
2. Pass them all in `book_ids` to `GetTrialBalance`.

```go
// Consolidated trial balance for an organisation and all its groups
// + members.
descendants, _ := bookBusiness.ListDescendantIDs(ctx, orgBookID)
report, _ := reportBusiness.TrialBalance(ctx, models.TrialBalanceParams{
    Currency: "UGX",
    BookIDs:  descendants,
})
```

Each line in `report.Lines` is one account in the scope. Each entry in
`report.Totals` is the per-currency grand total with `IsBalanced`.

### What's included

Trial balance includes activity from transactions with status:

- `POSTED` — committed to balance.
- `REVERSED` — committed and offset by a REVERSAL transaction whose
  entries cancel the original. Both rows are visible; net impact is
  zero. Trial balance still balances because each row's entries
  themselves zero-sum.

Excluded: `DRAFT`, `PENDING`, `VOIDED`, `FAILED` (never contributed
to balance).

## Account statement

A customer-facing statement for one account over a period. Returns:

- **Opening balance**: sum of activity strictly before `from`.
- **Entries**: chronological, each row carrying `running_balance` =
  opening + cumulative entries up to that point.
- **Closing balance**: opening + total period activity.
- **Total debits / Total credits**: per-side magnitudes of the period.

### Example: monthly member statement

```go
stmt, _ := reportBusiness.AccountStatement(ctx, models.StatementParams{
    AccountID: "member-123-savings",
    From:      mustTime("2026-05-01T00:00:00Z"),
    To:        mustTime("2026-05-31T23:59:59Z"),
    Limit:     100,
    Offset:    0,
})

// stmt.OpeningBalance carried in from before May 1.
// stmt.Entries are May's activity, oldest first.
// stmt.ClosingBalance is the balance at month-end.
// Each entry has RunningBalance — what the balance was right after
// that entry posted.
```

### Pagination

`Limit` caps at 1000 (defaults to 100 when ≤0). Use `Offset` to walk
longer periods. Running balances stay consistent across pages because
the opening balance carries forward from before `from`.

## CSV / Excel / PDF export

The UI exposes the same data in three formats. Same column layout, same
totals row, same metadata.

### Trial balance export

CSV/Excel header columns: `Account | Ledger | LedgerType | Currency |
TotalDebits | TotalCredits | NetBalance`.

A blank row separates the per-account lines from a "Per-currency totals"
section with: `Currency | TotalDebits | TotalCredits | IsBalanced`.

PDF formats as an A4 portrait page with the same two tables.

Filename convention: `trial_balance_<UTCYYYYMMDD_HHMMSS>.<ext>`.

### Account statement export

CSV/Excel begins with a header block:

```
Account | <id>
Currency | <code>
Opening balance | <amount>
Closing balance | <amount>
Period debits | <amount>
Period credits | <amount>
```

followed by a blank row and the entries table: `Transacted |
Transaction | Type | DR/CR | Amount | Running`.

PDF formats the same content as a printable A4 page.

Filename: `statement_<accountID>_<UTCYYYYMMDD_HHMMSS>.<ext>`.

### Where the files go

Driven by `file_saver` and `printing`:

- **Web**: triggers an anchor download.
- **iOS / Android**: opens the system share sheet (Save to Files,
  AirDrop, WhatsApp, …).
- **Desktop**: opens a save-as file picker.

PDF always goes through `Printing.sharePdf` so the OS print dialog is
available too.

## Books you can derive (without new tables)

All of these are filtered views over the same entry stream — no schema
change required:

| Book | Filter |
|---|---|
| Cash book | entries on accounts whose `account_type = asset` and `data.category = cash` (or `bank` / `mobile_money_float`) |
| Loan book | entries on accounts whose ledger represents loan principal / interest receivable |
| Savings book | entries on accounts whose `account_type = liability` and `data.category = member_savings` |
| Wallet book | entries on accounts whose `account_type = liability` and `data.category = wallet` |
| Suspense book | entries on accounts whose `account_type = suspense` |
| Customer statement | entries on one customer's account(s) |
| Group savings book | entries on accounts owned by a specific group's book |
| Revenue book | entries on `LedgerType = INCOME` accounts |
| Expense book | entries on `LedgerType = EXPENSE` accounts |

The trial-balance RPC supports filtering by `ledger_type` and `book_ids`
to express most of these. For deeper category filters (e.g. "cash
accounts only, not all assets"), the convention is to encode the
category in `account.data["category"]` and query JSONB in a custom
report job.

## Performance notes

- Trial-balance and statement queries use Raw SQL with explicit table
  aliases (the same pattern as the LATERAL balance subquery in
  `accounts.go`). Tenancy is stitched in via `buildTenancyClause` to
  avoid the ambiguous-column issue you'd get if the auto-applied
  TenancyPartition scope tried to filter unprefixed `tenant_id` across
  joined tables.
- Indexes:
  - `transaction_entries (account_id, transaction_id)` covers the
    statement and LATERAL paths.
  - `transactions (status)` partial on `deleted_at IS NULL` for the
    Status-based balance filter.
  - `transactions (transacted_at DESC)` partial on `deleted_at IS NULL`
    for chronological scans.
  - `accounts (book_id)` partial WHERE not null for book-scoped reports.
  - `books (parent_id)` partial WHERE not null for the recursive CTE.

For high-volume tenants, balance computation scales with entries per
account; consider a periodic snapshot ledger as a future optimisation.
The current LATERAL-based read is correct under concurrency and avoids
the consistency drift that a denormalised balance column would suffer.
