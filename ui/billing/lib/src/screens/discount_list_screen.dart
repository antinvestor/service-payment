import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/discount_providers.dart';

/// Screen that lists discounts with search using DataTable.
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
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (discounts) => _buildTable(items: discounts),
    );
  }

  Widget _buildTable({required List<Discount> items}) {
    return AdminEntityListPage<Discount>(
      title: 'Discounts',
      breadcrumbs: const ['Billing', 'Discounts'],
      searchHint: 'Search discounts...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      columns: const [
        DataColumn(label: Text('Name')),
        DataColumn(label: Text('Type')),
        DataColumn(label: Text('Value')),
        DataColumn(label: Text('Start')),
        DataColumn(label: Text('End')),
      ],
      items: items,
      rowBuilder: (discount, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(discount.name.isNotEmpty
              ? discount.name
              : discount.id)),
          DataCell(Text(_discountTypeLabel(discount.discountType))),
          DataCell(Text(_discountValueLabel(discount))),
          DataCell(Text(_formatTimestamp(discount.startAt))),
          DataCell(Text(
              discount.hasEndAt() ? _formatTimestamp(discount.endAt) : 'Open')),
        ],
      ),
      exportRow: (discount) => [
        discount.name.isNotEmpty ? discount.name : discount.id,
        _discountTypeLabel(discount.discountType),
        _discountValueLabel(discount),
        _formatTimestamp(discount.startAt),
        discount.hasEndAt() ? _formatTimestamp(discount.endAt) : 'Open',
      ],
      onExport: (format, rowCount) {
        debugPrint('AUDIT: Exported $rowCount discounts as $format');
      },
    );
  }

  String _discountTypeLabel(DiscountType type) {
    switch (type) {
      case DiscountType.DISCOUNT_PERCENTAGE:
        return 'Percentage';
      case DiscountType.DISCOUNT_FIXED:
        return 'Fixed';
      default:
        return type.name;
    }
  }

  String _discountValueLabel(Discount discount) {
    if (discount.discountType == DiscountType.DISCOUNT_PERCENTAGE) {
      return '${discount.value}%';
    }
    return '${discount.currency} ${discount.value}';
  }

  String _formatTimestamp(dynamic ts) {
    try {
      final seconds = ts.seconds.toInt();
      if (seconds == 0) return '';
      final dt = DateTime.fromMillisecondsSinceEpoch(seconds * 1000);
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }
}
