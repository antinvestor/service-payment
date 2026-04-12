import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';

import '../utils/money_format.dart';
import 'package:flutter/material.dart';

/// A list tile displaying an invoice line: description, quantity, unit price, amount.
class InvoiceLineTile extends StatelessWidget {
  const InvoiceLineTile({super.key, required this.line});

  final InvoiceLine line;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Description and line type
          Expanded(
            flex: 3,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  line.description.isNotEmpty ? line.description : 'Line item',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w500,
                  ),
                ),
                if (line.hasLineType())
                  Text(
                    line.lineType.name,
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 8),

          // Quantity
          SizedBox(
            width: 60,
            child: Text(
              '${line.quantity}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          const SizedBox(width: 8),

          // Unit price
          SizedBox(
            width: 80,
            child: Text(
              fmtMoney(line.unitPrice),
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          const SizedBox(width: 8),

          // Net amount
          SizedBox(
            width: 90,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  fmtMoney(line.netAmount),
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
                  textAlign: TextAlign.right,
                ),
                if (line.hasDiscountAmount() || line.hasCreditAmount())
                  Text(
                    _adjustmentsLabel(),
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: Colors.green,
                    ),
                    textAlign: TextAlign.right,
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _adjustmentsLabel() {
    final parts = <String>[];
    if (line.hasDiscountAmount()) {
      parts.add('-${fmtMoney(line.discountAmount)} disc');
    }
    if (line.hasCreditAmount()) {
      parts.add('-${fmtMoney(line.creditAmount)} credit');
    }
    return parts.join(', ');
  }
}

/// Header row for invoice line tiles.
class InvoiceLineHeader extends StatelessWidget {
  const InvoiceLineHeader({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final style = theme.textTheme.labelSmall?.copyWith(
      fontWeight: FontWeight.w600,
      color: theme.colorScheme.onSurfaceVariant,
    );

    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        children: [
          Expanded(flex: 3, child: Text('Description', style: style)),
          const SizedBox(width: 8),
          SizedBox(
            width: 60,
            child: Text('Qty', style: style, textAlign: TextAlign.right),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 80,
            child: Text('Unit Price', style: style, textAlign: TextAlign.right),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 90,
            child: Text('Amount', style: style, textAlign: TextAlign.right),
          ),
        ],
      ),
    );
  }
}
