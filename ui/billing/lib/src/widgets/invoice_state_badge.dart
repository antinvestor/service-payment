import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';

/// Displays a colored badge for invoice state.
class InvoiceStateBadge extends StatelessWidget {
  const InvoiceStateBadge({super.key, required this.state});

  final InvoiceState state;

  @override
  Widget build(BuildContext context) {
    final (label, color, icon) = _stateInfo(state);
    return StatusBadge(label: label, color: color, icon: icon);
  }

  static (String, Color, IconData?) _stateInfo(InvoiceState state) {
    return switch (state) {
      InvoiceState.INVOICE_DRAFT =>
        ('Draft', Colors.blue, Icons.edit_note),
      InvoiceState.INVOICE_ISSUED =>
        ('Issued', Colors.amber, Icons.send),
      InvoiceState.INVOICE_PAID =>
        ('Paid', Colors.green, Icons.check_circle),
      InvoiceState.INVOICE_VOIDED =>
        ('Voided', Colors.grey, Icons.block),
      InvoiceState.INVOICE_OVERDUE =>
        ('Overdue', Colors.red, Icons.warning),
      _ => ('Unknown', Colors.grey, null),
    };
  }
}

/// Returns a human-readable label for invoice state.
String invoiceStateLabel(InvoiceState state) {
  return switch (state) {
    InvoiceState.INVOICE_DRAFT => 'Draft',
    InvoiceState.INVOICE_ISSUED => 'Issued',
    InvoiceState.INVOICE_PAID => 'Paid',
    InvoiceState.INVOICE_VOIDED => 'Voided',
    InvoiceState.INVOICE_OVERDUE => 'Overdue',
    _ => 'Unknown',
  };
}
