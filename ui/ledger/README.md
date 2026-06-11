# antinvestor_ui_ledger

Embeddable Flutter screens, widgets, and Riverpod providers for the
Antinvestor double-entry ledger service. Drops into any host app via
`LedgerRouteModule` — registers the routes, nav items and permission
manifest with no host glue required.

## Backend it talks to

Configured via the compile-time env variable `LEDGER_URL`:

```bash
flutter run --dart-define LEDGER_URL=https://ledger.your.cluster
```

Auth follows the host app's `authTokenProviderProvider` from
`antinvestor_ui_core`.

## Screen index

### `/ledger/ledgers` — Chart of accounts

The ledger hierarchy. Each row shows id, type, parent ledger, and a
3-key data preview. Type filter narrows by Asset / Liability / Income /
Expense / Capital.

**Action**: "New ledger" opens a modal. Required: name, type. Optional:
parent ledger id (for hierarchical nesting), default currency — name and
currency are carried in the ledger's free-form `data` struct.

Permissions: `ledger_view`. Manage permission is `ledger_create`.

### `/ledger/ledgers/:id` — Ledger detail

Two tabs:
- **Info**: identity, type badge, parent reference, metadata.
- **Accounts**: scrollable list of accounts in this ledger with their
  balance / uncleared / reserved tiles.

### `/ledger/accounts` — Account list

Free-text search + standard list table. Columns: id, ledger, balance,
uncleared, reserved.

### `/ledger/accounts/:id` — Account detail

Balance card on top, transaction history below. Running balance is
shown per entry. From here, navigate into:

### `/ledger/accounts/:id/statement` — Account statement

The customer-facing statement screen. Filter by date range (RFC3339),
view opening + closing balances, per-side totals, and the chronological
entry list with running balance per row.

**Export** in the app bar: CSV / Excel / PDF. Filename pattern:
`statement_<accountID>_<UTC stamp>.<ext>`.

Permissions: `report_view`.

### `/ledger/transactions` — Transaction list

Most-recently-transacted first by default. Columns: id, currency, type
badge, transacted-at, entry count, cleared indicator.

### `/ledger/transactions/:id` — Transaction detail

Entry count headline, type + currency badges, ID and timestamp.
Below: list of `TransactionEntryRow` items, each tappable to navigate
to its account.

**Lifecycle actions** in the app bar:

| Action | Visible when | What it does |
|---|---|---|
| Reverse | type is not REVERSAL | Posts an offsetting REVERSAL transaction. |

A confirmation dialog precedes the mutation. Snackbar feedback on
success / failure. The transaction's view is invalidated so the result
surfaces immediately.

Permissions: `transaction_view` to read, `transaction_manage` to act.

### `/ledger/reports/trial-balance` — Trial balance

Filter bar: currency, ledger type, ledger id.

The body has:

1. **Per-currency totals card** — one row per currency, with a green
   `BALANCED` or red `UNBALANCED` chip from the `is_balanced` flag.
   This is the textbook integrity check: total_debits == total_credits.
2. **Per-account lines** — DataTable of Account / Ledger / Type /
   Currency / Debits / Credits / Net.

**Export** in the app bar: CSV / Excel / PDF.

Reachable from:
- The main nav ("Trial balance").
- Pre-scoped to one ledger via `?ledgerId=…`.

Permissions: `report_view`.

## Permissions reference

| Permission | Grants |
|---|---|
| `ledger_view` | Read chart-of-accounts |
| `ledger_create` | Create / update ledgers |
| `account_view` | Read accounts and balances |
| `account_create` | Create / update accounts |
| `transaction_view` | Read transactions and entries |
| `transaction_create` | Create transactions and trigger Reverse |
| `report_view` | Run trial balance and account statement reports + exports |

## Providers

| Provider | Returns |
|---|---|
| `ledgerSearchProvider(query)` | `List<Ledger>` |
| `accountSearchProvider(query)` | `List<Account>` |
| `transactionSearchProvider(query)` | `List<Transaction>` |
| `transactionEntrySearchProvider(query)` | `List<TransactionEntry>` |
| `trialBalanceProvider(query)` | `TrialBalanceReport` (derived client-side) |
| `accountStatementProvider(query)` | `AccountStatementReport` (derived client-side) |
| `transactionNotifierProvider` | mutations: create / reverse / update |
| `ledgerNotifierProvider` | mutations: create / update |
| `accountNotifierProvider` | mutations: create / update |

Each `Notifier` exposes an `AsyncValue<void>` for loading / error state
plus a method that returns the resulting domain object on success.

## Exporting reports — what the user sees

1. Open a report screen (trial balance or account statement).
2. Apply filters; the screen re-fetches.
3. Tap the download icon in the app bar.
4. Pick the format: CSV / Excel / PDF.

What happens by platform:

- **Web**: the file downloads via an anchor.
- **iOS / Android**: the system share sheet opens — save to Files,
  AirDrop, send via WhatsApp, etc.
- **Desktop**: a Save-As dialog opens.

PDFs always route through the OS print preview (so AirPrint / Generic
PDF Printer / Save as PDF are all available).

## Embedding into a host app

```dart
import 'package:antinvestor_ui_ledger/antinvestor_ui_ledger.dart';
import 'package:antinvestor_ui_core/routing/route_module.dart';

final modules = <RouteModule>[
  LedgerRouteModule(),
  // …other modules
];

// Build the GoRouter with all module routes.
final router = GoRouter(
  routes: modules.expand((m) => m.buildRoutes()).toList(),
);

// Build the nav drawer / rail with all module nav items.
final navItems = modules.expand((m) => m.buildNavItems()).toList();

// Aggregate permissions across modules for role-driven UI.
final manifests = modules.map((m) => m.permissionManifest).toList();
```

## Local development

```bash
cd ui/ledger
flutter pub get
flutter analyze
```

All dependencies resolve from pub.dev (`antinvestor_api_ledger`
^1.53.0, `antinvestor_api_common` ^1.53.1, `antinvestor_ui_core`
^0.5.0). To regenerate the Dart SDK after editing
`proto/ledger/v1/ledger.proto`:

```bash
make proto-generate-dart
```

To publish the proto and the Go gen:

```bash
make proto-push
```

Then bump the Go module:

```bash
GOPROXY=direct go get \
  buf.build/gen/go/antinvestor/ledger/protocolbuffers/go@latest \
  buf.build/gen/go/antinvestor/ledger/connectrpc/go@latest
```
