// ignore_for_file: implementation_imports
import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_api_ledger/src/google/type/money.pb.dart';
import 'package:antinvestor_ui_ledger/antinvestor_ui_ledger.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('LedgerTypeBadge and TransactionTypeBadge render labels', (
    tester,
  ) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              LedgerTypeBadge(type: LedgerType.ASSET),
              TransactionTypeBadge(type: TransactionType.REVERSAL),
            ],
          ),
        ),
      ),
    );
    expect(find.textContaining('Asset', findRichText: true), findsOneWidget);
    expect(find.textContaining('Reversal', findRichText: true), findsOneWidget);
  });

  testWidgets('LedgerTreeView nests children under their parent ledger', (
    tester,
  ) async {
    final root = Ledger()
      ..id = 'root'
      ..type = LedgerType.ASSET;
    final child = Ledger()
      ..id = 'child'
      ..type = LedgerType.LIABILITY
      ..parent = 'root';

    Ledger? tapped;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: LedgerTreeView(
            ledgers: [root, child],
            onLedgerTap: (l) => tapped = l,
          ),
        ),
      ),
    );

    expect(find.text('root'), findsOneWidget);
    expect(find.text('child'), findsOneWidget);

    await tester.tap(find.text('child'));
    await tester.pump();
    expect(tapped?.id, 'child');
  });

  testWidgets('TransactionEntryRow shows amount and DR/CR side', (
    tester,
  ) async {
    final entry = TransactionEntry()
      ..id = 'ent1'
      ..accountId = 'acc1'
      ..transactionId = 'txn1'
      ..transactedAt = '2026-01-10T00:00:00Z'
      ..credit = true
      ..amount = (Money()
        ..currencyCode = 'UGX'
        ..units = Int64(150))
      ..accBalance = (Money()
        ..currencyCode = 'UGX'
        ..units = Int64(150));

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(body: TransactionEntryRow(entry: entry)),
      ),
    );

    expect(find.textContaining('UGX'), findsWidgets);
  });
}
