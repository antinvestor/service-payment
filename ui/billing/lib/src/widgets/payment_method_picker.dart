import 'package:flutter/material.dart';

import '../providers/collection_providers.dart';

/// Multi-select chips for optional payment method restriction.
/// Empty selection means "all methods" on the hosted checkout page.
class PaymentMethodPicker extends StatelessWidget {
  const PaymentMethodPicker({
    super.key,
    required this.selected,
    required this.onChanged,
  });

  final Set<String> selected;
  final ValueChanged<Set<String>> onChanged;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Payment methods',
          style: theme.textTheme.titleSmall?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          selected.isEmpty
              ? 'All methods available for the currency will be shown.'
              : 'Only selected methods will appear on checkout.',
          style: theme.textTheme.bodySmall?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: knownPaymentMethods.entries.map((e) {
            final isSelected = selected.contains(e.key);
            return FilterChip(
              label: Text(e.value),
              selected: isSelected,
              onSelected: (value) {
                final next = Set<String>.from(selected);
                if (value) {
                  next.add(e.key);
                } else {
                  next.remove(e.key);
                }
                onChanged(next);
              },
            );
          }).toList(),
        ),
      ],
    );
  }
}
