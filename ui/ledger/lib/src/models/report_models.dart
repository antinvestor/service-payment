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

// ignore_for_file: implementation_imports
import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_api_ledger/src/google/type/money.pb.dart';
import 'package:fixnum/fixnum.dart';

/// Report models derived client-side.
///
/// `antinvestor_api_ledger` 1.53.0 removed the server-side
/// `GetAccountStatement` and `GetTrialBalance` RPCs (and their
/// `StatementEntry` / `TrialBalanceLine` / `TrialBalanceTotal` messages).
/// The same reports are now derived in the client from the generic search
/// RPCs: statements from `SearchTransactionEntries` (each entry carries the
/// post-entry running balance in `acc_balance`) and trial balances from
/// `SearchAccounts` + `SearchLedgers`.

/// One line of an account statement (mirrors the removed `StatementEntry`).
class StatementLine {
  const StatementLine({
    required this.transactedAt,
    required this.transactionId,
    required this.credit,
    required this.amount,
    required this.runningBalance,
  });

  /// RFC3339 timestamp the entry was transacted at.
  final String transactedAt;

  /// Id of the owning transaction.
  final String transactionId;

  /// True for credit entries, false for debits.
  final bool credit;

  /// Entry amount.
  final Money amount;

  /// Account balance after this entry was applied.
  final Money runningBalance;
}

/// Account statement report (mirrors the removed
/// `GetAccountStatementResponse`).
class AccountStatementReport {
  const AccountStatementReport({
    required this.accountId,
    required this.currency,
    required this.openingBalance,
    required this.closingBalance,
    required this.totalDebits,
    required this.totalCredits,
    required this.entries,
  });

  final String accountId;
  final String currency;

  /// Balance carried in from before the period.
  final Money openingBalance;

  /// Balance after the last entry in the period.
  final Money closingBalance;

  /// Sum of debit amounts within the period.
  final Money totalDebits;

  /// Sum of credit amounts within the period.
  final Money totalCredits;

  /// Chronological entries within the period.
  final List<StatementLine> entries;
}

/// One per-account line of a trial balance (mirrors the removed
/// `TrialBalanceLine`).
class TrialBalanceLine {
  const TrialBalanceLine({
    required this.accountId,
    required this.ledgerId,
    required this.ledgerType,
    required this.currency,
    required this.debit,
    required this.credit,
    required this.netBalance,
  });

  final String accountId;
  final String ledgerId;

  /// Type of the owning ledger, when known. Null if the ledger could not
  /// be resolved (e.g. the account references a pruned ledger).
  final LedgerType? ledgerType;

  final String currency;

  /// Amount shown in the debit column (net balance when positive).
  final Money debit;

  /// Amount shown in the credit column (absolute net balance when negative).
  final Money credit;

  /// Signed net balance of the account.
  final Money netBalance;
}

/// Per-currency totals row of a trial balance (mirrors the removed
/// `TrialBalanceTotal`).
class TrialBalanceTotal {
  const TrialBalanceTotal({
    required this.currency,
    required this.totalDebits,
    required this.totalCredits,
    required this.isBalanced,
  });

  final String currency;
  final Money totalDebits;
  final Money totalCredits;
  final bool isBalanced;
}

/// Trial balance report (mirrors the removed `GetTrialBalanceResponse`).
class TrialBalanceReport {
  const TrialBalanceReport({required this.lines, required this.totals});

  final List<TrialBalanceLine> lines;
  final List<TrialBalanceTotal> totals;
}

const int _nanosPerUnit = 1000000000;

Int64 _toNanos(Money m) => m.units * _nanosPerUnit + Int64(m.nanos);

// `Int64.%` is euclidean (always non-negative); `remainder` keeps the
// dividend's sign, which is what proto `Money` requires (units and nanos
// must agree in sign).
Money _fromNanos(Int64 totalNanos, String currencyCode) => Money()
  ..currencyCode = currencyCode
  ..units = totalNanos ~/ _nanosPerUnit
  ..nanos = totalNanos.remainder(_nanosPerUnit).toInt();

/// Returns a zero [Money] in [currencyCode].
Money moneyZero(String currencyCode) => Money()..currencyCode = currencyCode;

/// Adds two money values, carrying nanos. The currency of [a] wins when set.
Money moneyAdd(Money a, Money b) => _fromNanos(
  _toNanos(a) + _toNanos(b),
  a.currencyCode.isNotEmpty ? a.currencyCode : b.currencyCode,
);

/// Absolute value of a money amount.
Money moneyAbs(Money m) =>
    moneyIsNegative(m) ? _fromNanos(-_toNanos(m), m.currencyCode) : m;

/// True when the amount is strictly negative.
bool moneyIsNegative(Money m) => _toNanos(m) < 0;

/// True when both amounts represent the same value.
bool moneyEquals(Money a, Money b) => _toNanos(a) == _toNanos(b);

DateTime? _parseTime(String rfc3339) =>
    rfc3339.isEmpty ? null : DateTime.tryParse(rfc3339);

bool _within(DateTime? t, DateTime? from, DateTime? to) {
  if (t == null) return from == null && to == null;
  if (from != null && t.isBefore(from)) return false;
  if (to != null && t.isAfter(to)) return false;
  return true;
}

/// Builds an [AccountStatementReport] for [accountId] from raw transaction
/// [entries] (as returned by `SearchTransactionEntries`).
///
/// [from] / [to] are optional inclusive RFC3339 bounds. Entries outside the
/// period still inform the opening balance: the opening balance is the
/// running balance of the last entry before [from], so no debit/credit sign
/// convention needs to be assumed.
AccountStatementReport buildAccountStatement({
  required String accountId,
  required List<TransactionEntry> entries,
  String from = '',
  String to = '',
}) {
  final fromTime = _parseTime(from);
  final toTime = _parseTime(to);

  final own =
      entries
          .where((e) => e.accountId.isEmpty || e.accountId == accountId)
          .toList()
        ..sort((a, b) {
          final ta = _parseTime(a.transactedAt);
          final tb = _parseTime(b.transactedAt);
          if (ta != null && tb != null) return ta.compareTo(tb);
          return a.transactedAt.compareTo(b.transactedAt);
        });

  final currency = own.isNotEmpty ? own.first.amount.currencyCode : '';

  TransactionEntry? lastBefore;
  final inPeriod = <TransactionEntry>[];
  for (final e in own) {
    final t = _parseTime(e.transactedAt);
    if (fromTime != null && t != null && t.isBefore(fromTime)) {
      lastBefore = e;
      continue;
    }
    if (_within(t, null, toTime) || (t == null && toTime == null)) {
      inPeriod.add(e);
    }
  }

  final opening = lastBefore != null
      ? lastBefore.accBalance
      : moneyZero(currency);
  final closing = inPeriod.isNotEmpty ? inPeriod.last.accBalance : opening;

  var debits = moneyZero(currency);
  var credits = moneyZero(currency);
  for (final e in inPeriod) {
    if (e.credit) {
      credits = moneyAdd(credits, e.amount);
    } else {
      debits = moneyAdd(debits, e.amount);
    }
  }

  return AccountStatementReport(
    accountId: accountId,
    currency: currency,
    openingBalance: opening,
    closingBalance: closing,
    totalDebits: debits,
    totalCredits: credits,
    entries: inPeriod
        .map(
          (e) => StatementLine(
            transactedAt: e.transactedAt,
            transactionId: e.transactionId,
            credit: e.credit,
            amount: e.amount,
            runningBalance: e.accBalance,
          ),
        )
        .toList(),
  );
}

/// Builds a [TrialBalanceReport] from current [accounts] (as returned by
/// `SearchAccounts`) and the [ledgers] they belong to.
///
/// Presentation follows the signed-balance convention: a positive net
/// balance lands in the debit column, a negative one (as its absolute
/// value) in the credit column. Optional filters narrow the report by
/// [currency] (ISO 4217), owning [ledgerId] and [ledgerType] name
/// ("ASSET" / "LIABILITY" / ...).
TrialBalanceReport buildTrialBalance({
  required List<Account> accounts,
  List<Ledger> ledgers = const [],
  String currency = '',
  String ledgerId = '',
  String ledgerType = '',
}) {
  final typeByLedger = {for (final l in ledgers) l.id: l.type};

  final lines = <TrialBalanceLine>[];
  for (final account in accounts) {
    final accountCurrency = account.balance.currencyCode;
    if (currency.isNotEmpty && accountCurrency != currency) continue;
    if (ledgerId.isNotEmpty && account.ledger != ledgerId) continue;
    final type = typeByLedger[account.ledger];
    if (ledgerType.isNotEmpty && type?.name != ledgerType) continue;

    final net = account.balance;
    final negative = moneyIsNegative(net);
    lines.add(
      TrialBalanceLine(
        accountId: account.id,
        ledgerId: account.ledger,
        ledgerType: type,
        currency: accountCurrency,
        debit: negative ? moneyZero(accountCurrency) : net,
        credit: negative ? moneyAbs(net) : moneyZero(accountCurrency),
        netBalance: net,
      ),
    );
  }

  final debitsByCurrency = <String, Money>{};
  final creditsByCurrency = <String, Money>{};
  for (final line in lines) {
    debitsByCurrency[line.currency] = moneyAdd(
      debitsByCurrency[line.currency] ?? moneyZero(line.currency),
      line.debit,
    );
    creditsByCurrency[line.currency] = moneyAdd(
      creditsByCurrency[line.currency] ?? moneyZero(line.currency),
      line.credit,
    );
  }

  final totals =
      debitsByCurrency.keys
          .map(
            (c) => TrialBalanceTotal(
              currency: c,
              totalDebits: debitsByCurrency[c]!,
              totalCredits: creditsByCurrency[c]!,
              isBalanced: moneyEquals(
                debitsByCurrency[c]!,
                creditsByCurrency[c]!,
              ),
            ),
          )
          .toList()
        ..sort((a, b) => a.currency.compareTo(b.currency));

  return TrialBalanceReport(lines: lines, totals: totals);
}
