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
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'ledger_transport_provider.dart';

/// Parameter bundle for the trial balance provider so Riverpod's family
/// cache key respects all filter dimensions.
@immutable
class TrialBalanceQuery {
  const TrialBalanceQuery({
    this.currency = '',
    this.ledgerId = '',
    this.ledgerType = '',
    this.bookIds = const <String>[],
    this.asOf = '',
  });

  /// ISO 4217 currency code. Empty = no filter.
  final String currency;

  /// Optional ledger id to scope the report to one chart-of-accounts subtree.
  final String ledgerId;

  /// Optional case-sensitive ledger-type filter ("ASSET" / "LIABILITY" /
  /// "INCOME" / "EXPENSE" / "CAPITAL"). Empty = no filter.
  final String ledgerType;

  /// Optional list of book ids to scope to. For a consolidated report
  /// across an organisation's groups + members, expand the root via
  /// the descendants helper (added with hierarchy support).
  final List<String> bookIds;

  /// Optional RFC3339 upper bound on transacted_at. Empty = no bound.
  final String asOf;

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is TrialBalanceQuery &&
        other.currency == currency &&
        other.ledgerId == ledgerId &&
        other.ledgerType == ledgerType &&
        other.asOf == asOf &&
        listEquals(other.bookIds, bookIds);
  }

  @override
  int get hashCode => Object.hash(
        currency,
        ledgerId,
        ledgerType,
        asOf,
        Object.hashAll(bookIds),
      );
}

/// Trial balance — per-account debit/credit totals plus per-currency
/// integrity totals.
final trialBalanceProvider = FutureProvider.family<GetTrialBalanceResponse,
    TrialBalanceQuery>((ref, query) async {
  final client = ref.watch(ledgerServiceClientProvider);
  final request = GetTrialBalanceRequest()
    ..currency = query.currency
    ..ledgerId = query.ledgerId
    ..ledgerType = query.ledgerType
    ..bookIds.addAll(query.bookIds)
    ..asOf = query.asOf;
  return client.getTrialBalance(request);
});

/// Parameter bundle for the account-statement provider.
@immutable
class AccountStatementQuery {
  const AccountStatementQuery({
    required this.accountId,
    this.from = '',
    this.to = '',
    this.limit = 100,
    this.offset = 0,
  });

  /// Required account id.
  final String accountId;

  /// Optional RFC3339 lower bound (inclusive). Empty = no bound.
  final String from;

  /// Optional RFC3339 upper bound (inclusive). Empty = no bound.
  final String to;

  /// Page size (server caps at 1000, defaults to 100 when <=0).
  final int limit;

  /// Page offset.
  final int offset;

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is AccountStatementQuery &&
        other.accountId == accountId &&
        other.from == from &&
        other.to == to &&
        other.limit == limit &&
        other.offset == offset;
  }

  @override
  int get hashCode => Object.hash(accountId, from, to, limit, offset);
}

/// Account statement — chronological entries with running balance,
/// opening + closing balances and the period's totals.
final accountStatementProvider = FutureProvider.family<
    GetAccountStatementResponse, AccountStatementQuery>((ref, query) async {
  final client = ref.watch(ledgerServiceClientProvider);
  final request = GetAccountStatementRequest()
    ..accountId = query.accountId
    ..from = query.from
    ..to = query.to
    ..limit = query.limit
    ..offset = query.offset;
  return client.getAccountStatement(request);
});
