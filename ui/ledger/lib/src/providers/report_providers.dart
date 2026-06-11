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

import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/report_models.dart';
import 'ledger_transport_provider.dart';

/// Parameter bundle for the trial balance provider so Riverpod's family
/// cache key respects all filter dimensions.
@immutable
class TrialBalanceQuery {
  const TrialBalanceQuery({
    this.currency = '',
    this.ledgerId = '',
    this.ledgerType = '',
  });

  /// ISO 4217 currency code. Empty = no filter.
  final String currency;

  /// Optional ledger id to scope the report to one ledger's accounts.
  final String ledgerId;

  /// Optional case-sensitive ledger-type filter ("ASSET" / "LIABILITY" /
  /// "INCOME" / "EXPENSE" / "CAPITAL"). Empty = no filter.
  final String ledgerType;

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is TrialBalanceQuery &&
        other.currency == currency &&
        other.ledgerId == ledgerId &&
        other.ledgerType == ledgerType;
  }

  @override
  int get hashCode => Object.hash(currency, ledgerId, ledgerType);
}

/// Trial balance — per-account debit/credit presentation of current
/// balances plus per-currency integrity totals.
///
/// Derived client-side from `SearchAccounts` + `SearchLedgers` since
/// api_ledger 1.53 removed the dedicated `GetTrialBalance` RPC.
final trialBalanceProvider =
    FutureProvider.family<TrialBalanceReport, TrialBalanceQuery>((
      ref,
      query,
    ) async {
      final client = ref.watch(ledgerServiceClientProvider);
      final accounts = await collectStream<SearchAccountsResponse, Account>(
        client.searchAccounts(SearchRequest()..query = ''),
        extract: (r) => r.data,
      );
      final ledgers = await collectStream<SearchLedgersResponse, Ledger>(
        client.searchLedgers(SearchRequest()..query = ''),
        extract: (r) => r.data,
      );
      return buildTrialBalance(
        accounts: accounts,
        ledgers: ledgers,
        currency: query.currency,
        ledgerId: query.ledgerId,
        ledgerType: query.ledgerType,
      );
    });

/// Parameter bundle for the account-statement provider.
@immutable
class AccountStatementQuery {
  const AccountStatementQuery({
    required this.accountId,
    this.from = '',
    this.to = '',
  });

  /// Required account id.
  final String accountId;

  /// Optional RFC3339 lower bound (inclusive). Empty = no bound.
  final String from;

  /// Optional RFC3339 upper bound (inclusive). Empty = no bound.
  final String to;

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is AccountStatementQuery &&
        other.accountId == accountId &&
        other.from == from &&
        other.to == to;
  }

  @override
  int get hashCode => Object.hash(accountId, from, to);
}

/// Account statement — chronological entries with running balance,
/// opening + closing balances and the period's totals.
///
/// Derived client-side from `SearchTransactionEntries` (each entry carries
/// its post-entry running balance) since api_ledger 1.53 removed the
/// dedicated `GetAccountStatement` RPC.
final accountStatementProvider =
    FutureProvider.family<AccountStatementReport, AccountStatementQuery>((
      ref,
      query,
    ) async {
      final client = ref.watch(ledgerServiceClientProvider);
      final entries =
          await collectStream<
            SearchTransactionEntriesResponse,
            TransactionEntry
          >(
            client.searchTransactionEntries(
              SearchRequest()..query = query.accountId,
            ),
            extract: (r) => r.data,
          );
      return buildAccountStatement(
        accountId: query.accountId,
        entries: entries,
        from: query.from,
        to: query.to,
      );
    });
