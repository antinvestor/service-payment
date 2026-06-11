import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';

/// Displays a colored badge for ledger type
/// (ASSET=blue, LIABILITY=red, INCOME=green, EXPENSE=orange, CAPITAL=purple).
class LedgerTypeBadge extends StatelessWidget {
  const LedgerTypeBadge({super.key, required this.type});

  final LedgerType type;

  @override
  Widget build(BuildContext context) {
    return StatusBadge.fromEnum(value: type, mapper: _ledgerTypeInfo);
  }

  static (String, Color, IconData?) _ledgerTypeInfo(LedgerType type) {
    return switch (type) {
      LedgerType.ASSET => ('Asset', Colors.blue, Icons.account_balance),
      LedgerType.LIABILITY => ('Liability', Colors.red, Icons.credit_card),
      LedgerType.INCOME => ('Income', Colors.green, Icons.trending_up),
      LedgerType.EXPENSE => ('Expense', Colors.orange, Icons.trending_down),
      LedgerType.CAPITAL => ('Capital', Colors.purple, Icons.savings),
      _ => ('Unknown', Colors.grey, null),
    };
  }
}

/// Returns the display label for a LedgerType.
String ledgerTypeLabel(LedgerType type) {
  return switch (type) {
    LedgerType.ASSET => 'Asset',
    LedgerType.LIABILITY => 'Liability',
    LedgerType.INCOME => 'Income',
    LedgerType.EXPENSE => 'Expense',
    LedgerType.CAPITAL => 'Capital',
    _ => 'Unknown',
  };
}

/// Returns the color for a LedgerType.
Color ledgerTypeColor(LedgerType type) {
  return switch (type) {
    LedgerType.ASSET => Colors.blue,
    LedgerType.LIABILITY => Colors.red,
    LedgerType.INCOME => Colors.green,
    LedgerType.EXPENSE => Colors.orange,
    LedgerType.CAPITAL => Colors.purple,
    _ => Colors.grey,
  };
}
