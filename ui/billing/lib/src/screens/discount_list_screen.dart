import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/discount_providers.dart';

/// Screen that lists discounts with search.
class DiscountListScreen extends ConsumerStatefulWidget {
  const DiscountListScreen({super.key});

  @override
  ConsumerState<DiscountListScreen> createState() =>
      _DiscountListScreenState();
}

class _DiscountListScreenState extends ConsumerState<DiscountListScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncDiscounts =
        ref.watch(discountSearchProvider(_searchQuery));

    return asyncDiscounts.when(
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (discounts) => _buildShell(items: discounts),
    );
  }

  Widget _buildShell({
    required List<Discount> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Discount>(
      title: 'Discounts',
      icon: Icons.local_offer,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () =>
          ref.invalidate(discountSearchProvider(_searchQuery)),
      searchHint: 'Search discounts...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      itemBuilder: (context, discount) {
        return _DiscountTile(discount: discount);
      },
    );
  }
}

class _DiscountTile extends StatelessWidget {
  const _DiscountTile({required this.discount});

  final Discount discount;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: theme.colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: theme.colorScheme.tertiaryContainer.withAlpha(80),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                Icons.local_offer,
                size: 20,
                color: theme.colorScheme.onTertiaryContainer,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    discount.name.isNotEmpty ? discount.name : discount.id,
                    style: theme.textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    discount.discountType == DiscountType.DISCOUNT_PERCENTAGE
                        ? '${discount.value}%'
                        : '${discount.currency} ${discount.value}',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            StatusBadge(
              label: discount.hasEndAt() ? 'Limited' : 'Open',
              color: discount.hasEndAt() ? Colors.amber : Colors.green,
            ),
          ],
        ),
      ),
    );
  }
}
