import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/usage_providers.dart';

/// Screen that lists usage events with search using DataTable.
class UsageEventsScreen extends ConsumerStatefulWidget {
  const UsageEventsScreen({super.key});

  @override
  ConsumerState<UsageEventsScreen> createState() => _UsageEventsScreenState();
}

class _UsageEventsScreenState extends ConsumerState<UsageEventsScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncEvents =
        ref.watch(usageEventSearchProvider(_searchQuery));

    return asyncEvents.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (events) => _buildTable(items: events),
    );
  }

  Widget _buildTable({required List<UsageEvent> items}) {
    return AdminEntityListPage<UsageEvent>(
      title: 'Usage Events',
      breadcrumbs: const ['Billing', 'Usage Events'],
      searchHint: 'Search usage events...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      columns: const [
        DataColumn(label: Text('Metric')),
        DataColumn(label: Text('Quantity'), numeric: true),
        DataColumn(label: Text('Unit')),
        DataColumn(label: Text('Timestamp')),
        DataColumn(label: Text('Subscription')),
      ],
      items: items,
      rowBuilder: (event, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(event.metricKey)),
          DataCell(Text(event.quantity.toStringAsFixed(2))),
          DataCell(Text(event.unit)),
          DataCell(Text(_formatTimestamp(event.timestamp))),
          DataCell(Text(_truncate(event.subscriptionId, 12))),
        ],
      ),
      exportRow: (event) => [
        event.metricKey,
        event.quantity.toStringAsFixed(2),
        event.unit,
        _formatTimestamp(event.timestamp),
        event.subscriptionId,
      ],
      onExport: (format, rowCount) {
        debugPrint(
            'AUDIT: Exported $rowCount usage events as $format');
      },
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  String _formatTimestamp(dynamic ts) {
    try {
      final seconds = ts.seconds.toInt();
      if (seconds == 0) return '';
      final dt = DateTime.fromMillisecondsSinceEpoch(seconds * 1000);
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')} '
          '${dt.hour.toString().padLeft(2, '0')}:'
          '${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }
}
