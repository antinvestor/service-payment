import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';

/// Displays a colored badge for transaction type (NORMAL, REVERSAL, RESERVATION).
class TransactionTypeBadge extends StatelessWidget {
  const TransactionTypeBadge({super.key, required this.type});

  final TransactionType type;

  @override
  Widget build(BuildContext context) {
    return StatusBadge.fromEnum(value: type, mapper: _transactionTypeInfo);
  }

  static (String, Color, IconData?) _transactionTypeInfo(TransactionType type) {
    return switch (type) {
      TransactionType.NORMAL => ('Normal', Colors.blue, Icons.receipt_long),
      TransactionType.REVERSAL => ('Reversal', Colors.orange, Icons.undo),
      TransactionType.RESERVATION => (
        'Reservation',
        Colors.purple,
        Icons.lock_clock,
      ),
      _ => ('Unknown', Colors.grey, null),
    };
  }
}

/// Returns the display label for a TransactionType.
String transactionTypeLabel(TransactionType type) {
  return switch (type) {
    TransactionType.NORMAL => 'Normal',
    TransactionType.REVERSAL => 'Reversal',
    TransactionType.RESERVATION => 'Reservation',
    _ => 'Unknown',
  };
}
