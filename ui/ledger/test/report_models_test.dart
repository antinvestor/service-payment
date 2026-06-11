// ignore_for_file: implementation_imports
import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_api_ledger/src/google/type/money.pb.dart';
import 'package:antinvestor_ui_ledger/antinvestor_ui_ledger.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';

Money money(int units, {int nanos = 0, String currency = 'UGX'}) => Money()
  ..currencyCode = currency
  ..units = Int64(units)
  ..nanos = nanos;

TransactionEntry entry({
  required String account,
  required String txn,
  required String at,
  required bool credit,
  required Money amount,
  required Money accBalance,
}) => TransactionEntry()
  ..accountId = account
  ..transactionId = txn
  ..transactedAt = at
  ..credit = credit
  ..amount = amount
  ..accBalance = accBalance;

void main() {
  group('buildAccountStatement', () {
    final entries = [
      entry(
        account: 'acc1',
        txn: 'txn1',
        at: '2026-01-10T00:00:00Z',
        credit: true,
        amount: money(100),
        accBalance: money(100),
      ),
      entry(
        account: 'acc1',
        txn: 'txn2',
        at: '2026-02-10T00:00:00Z',
        credit: false,
        amount: money(30),
        accBalance: money(70),
      ),
      entry(
        account: 'acc1',
        txn: 'txn3',
        at: '2026-03-10T00:00:00Z',
        credit: true,
        amount: money(50),
        accBalance: money(120),
      ),
    ];

    test('without bounds covers all entries from a zero opening', () {
      final report = buildAccountStatement(accountId: 'acc1', entries: entries);
      expect(report.accountId, 'acc1');
      expect(report.currency, 'UGX');
      expect(report.entries, hasLength(3));
      expect(report.openingBalance.units, Int64(0));
      expect(report.closingBalance.units, Int64(120));
      expect(report.totalCredits.units, Int64(150));
      expect(report.totalDebits.units, Int64(30));
    });

    test('opening balance carries in from before the period', () {
      final report = buildAccountStatement(
        accountId: 'acc1',
        entries: entries,
        from: '2026-02-01T00:00:00Z',
      );
      expect(report.entries, hasLength(2));
      expect(report.openingBalance.units, Int64(100));
      expect(report.closingBalance.units, Int64(120));
      expect(report.totalDebits.units, Int64(30));
      expect(report.totalCredits.units, Int64(50));
    });

    test('upper bound excludes later entries', () {
      final report = buildAccountStatement(
        accountId: 'acc1',
        entries: entries,
        to: '2026-02-28T00:00:00Z',
      );
      expect(report.entries, hasLength(2));
      expect(report.closingBalance.units, Int64(70));
    });

    test('foreign-account entries are ignored', () {
      final mixed = [
        ...entries,
        entry(
          account: 'other',
          txn: 'txnX',
          at: '2026-01-15T00:00:00Z',
          credit: true,
          amount: money(999),
          accBalance: money(999),
        ),
      ];
      final report = buildAccountStatement(accountId: 'acc1', entries: mixed);
      expect(report.entries, hasLength(3));
      expect(report.totalCredits.units, Int64(150));
    });
  });

  group('buildTrialBalance', () {
    final ledgers = [
      Ledger()
        ..id = 'led-assets'
        ..type = LedgerType.ASSET,
      Ledger()
        ..id = 'led-liab'
        ..type = LedgerType.LIABILITY,
    ];
    final accounts = [
      Account()
        ..id = 'acc1'
        ..ledger = 'led-assets'
        ..balance = money(100),
      Account()
        ..id = 'acc2'
        ..ledger = 'led-liab'
        ..balance = money(-100),
      Account()
        ..id = 'acc3'
        ..ledger = 'led-assets'
        ..balance = money(25, currency: 'USD'),
    ];

    test('splits signed balances into debit/credit columns', () {
      final report = buildTrialBalance(accounts: accounts, ledgers: ledgers);
      expect(report.lines, hasLength(3));
      final line1 = report.lines.firstWhere((l) => l.accountId == 'acc1');
      expect(line1.debit.units, Int64(100));
      expect(line1.credit.units, Int64(0));
      expect(line1.ledgerType, LedgerType.ASSET);
      final line2 = report.lines.firstWhere((l) => l.accountId == 'acc2');
      expect(line2.debit.units, Int64(0));
      expect(line2.credit.units, Int64(100));
    });

    test('per-currency totals carry the integrity check', () {
      final report = buildTrialBalance(accounts: accounts, ledgers: ledgers);
      expect(report.totals, hasLength(2));
      final ugx = report.totals.firstWhere((t) => t.currency == 'UGX');
      expect(ugx.totalDebits.units, Int64(100));
      expect(ugx.totalCredits.units, Int64(100));
      expect(ugx.isBalanced, isTrue);
      final usd = report.totals.firstWhere((t) => t.currency == 'USD');
      expect(usd.isBalanced, isFalse);
    });

    test('currency, ledger and type filters narrow the lines', () {
      expect(
        buildTrialBalance(
          accounts: accounts,
          ledgers: ledgers,
          currency: 'USD',
        ).lines.map((l) => l.accountId),
        ['acc3'],
      );
      expect(
        buildTrialBalance(
          accounts: accounts,
          ledgers: ledgers,
          ledgerId: 'led-liab',
        ).lines.map((l) => l.accountId),
        ['acc2'],
      );
      expect(
        buildTrialBalance(
          accounts: accounts,
          ledgers: ledgers,
          ledgerType: 'ASSET',
        ).lines.map((l) => l.accountId),
        ['acc1', 'acc3'],
      );
    });
  });

  group('money helpers', () {
    test('moneyAdd carries nanos into units', () {
      final sum = moneyAdd(
        money(1, nanos: 600000000),
        money(2, nanos: 700000000),
      );
      expect(sum.units, Int64(4));
      expect(sum.nanos, 300000000);
      expect(sum.currencyCode, 'UGX');
    });

    test('moneyAbs flips negative amounts', () {
      final abs = moneyAbs(money(-3, nanos: -500000000));
      expect(abs.units, Int64(3));
      expect(abs.nanos, 500000000);
      expect(moneyIsNegative(abs), isFalse);
    });
  });
}
