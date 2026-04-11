# antinvestor_ui_ledger

Embeddable ledger management UI for Antinvestor applications. Provides screens and widgets for managing ledgers, accounts, transactions, and transaction entries with double-entry bookkeeping support.

## Installation

```yaml
dependencies:
  antinvestor_ui_ledger: ^0.1.0
```

## Features

- **Ledger Management**: Hierarchical ledger list with tree view
- **Account Management**: List and detail views with balance tracking
- **Transaction Management**: Transaction list with entry-level detail
- **Embeddable Widgets**: `LedgerTypeBadge`, `AccountBalanceCard`, `TransactionEntryRow`, `TransactionTypeBadge`, `LedgerTreeView`
- **Routing**: `LedgerRouteModule` with GoRouter integration

## Usage

```dart
import 'package:antinvestor_ui_ledger/antinvestor_ui_ledger.dart';

// Account balance display card
AccountBalanceCard(account: accountObject)

// Ledger hierarchy tree
LedgerTreeView(rootLedgerId: 'root')

// Register routes in your host app
final module = LedgerRouteModule();
ShellRoute(
  routes: [...ownRoutes, ...module.buildRoutes()],
);
```

## Routes

| Path | Screen |
|------|--------|
| `/ledger/ledgers` | Ledger list |
| `/ledger/ledgers/:id` | Ledger detail |
| `/ledger/accounts` | Account list |
| `/ledger/accounts/:id` | Account detail with balance |
| `/ledger/transactions` | Transaction list |
| `/ledger/transactions/:id` | Transaction detail with entries |

## Embedding Widgets

```dart
// Ledger type indicator
LedgerTypeBadge(type: ledgerType)

// Transaction entry row with debit/credit
TransactionEntryRow(entry: entryObject)

// Transaction type indicator
TransactionTypeBadge(type: transactionType)
```
