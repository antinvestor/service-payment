## 0.3.0

- BREAKING: migrate to `antinvestor_api_ledger` ^1.53.0 (and
  `antinvestor_api_common` ^1.53.1). The 1.53 API renamed the Book domain to
  Ledger and removed several RPCs; the package surface follows:
  - Removed `BookListScreen`, `BookDetailScreen`, `BookFormScreen` and
    `booksByTypeProvider` / `bookByIdProvider` / `bookNotifierProvider` —
    the existing `Ledger*` screens/providers cover the renamed entity, and
    the new `LedgerFormScreen` (create dialog, reachable from
    `LedgerListScreen`) replaces `BookFormScreen`.
  - `/ledger/books` routes, the Books nav item and the `book_view` /
    `book_manage` permissions are gone from `LedgerRouteModule`.
  - `TrialBalanceScreen.initialBookId` is now `initialLedgerId` (query
    parameter `ledgerId`).
  - Account statement and trial balance reports are now derived client-side
    (`AccountStatementReport`, `TrialBalanceReport` in
    `src/models/report_models.dart`) from `SearchTransactionEntries` /
    `SearchAccounts` + `SearchLedgers`, since the 1.53 API removed the
    `GetAccountStatement` and `GetTrialBalance` RPCs. The statement query
    lost `limit` / `offset`; the trial balance query lost `bookIds` and
    `asOf`.
  - `TransactionNotifier.voidTransaction` / `markFailed` removed (RPCs gone
    in 1.53); transaction lifecycle actions reduce to Reverse.

## 0.1.1

- Migrate providers to Riverpod 3.x Notifier, fix lint warnings and deprecations

## 0.1.0

- Initial release
- Ledger/accounting UI with accounts, transactions, balance cards, tree view
