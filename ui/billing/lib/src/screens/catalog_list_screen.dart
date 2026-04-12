import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/catalog_providers.dart';

/// Screen that lists catalog versions with search using DataTable.
class CatalogListScreen extends ConsumerStatefulWidget {
  const CatalogListScreen({super.key});

  @override
  ConsumerState<CatalogListScreen> createState() => _CatalogListScreenState();
}

class _CatalogListScreenState extends ConsumerState<CatalogListScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncCatalogs =
        ref.watch(catalogVersionSearchProvider(_searchQuery));

    return asyncCatalogs.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (catalogs) => _buildTable(items: catalogs),
    );
  }

  Widget _buildTable({required List<CatalogVersion> items}) {
    return AdminEntityListPage<CatalogVersion>(
      title: 'Catalogs',
      breadcrumbs: const ['Billing', 'Catalogs'],
      searchHint: 'Search catalog versions...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      columns: const [
        DataColumn(label: Text('Name')),
        DataColumn(label: Text('Version'), numeric: true),
        DataColumn(label: Text('Currency')),
        DataColumn(label: Text('Published')),
        DataColumn(label: Text('Effective')),
        DataColumn(label: Text('Plans'), numeric: true),
      ],
      items: items,
      rowBuilder: (catalog, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(catalog.name.isNotEmpty
              ? catalog.name
              : _truncate(catalog.id, 16))),
          DataCell(Text('v${catalog.version}')),
          DataCell(Text(catalog.currency)),
          DataCell(_publishedBadge(catalog)),
          DataCell(Text(_formatTimestamp(catalog.effectiveAt))),
          DataCell(Text('${catalog.plans.length}')),
        ],
      ),
      onRowNavigate: (catalog) =>
          context.go('/billing/catalogs/${catalog.id}'),
      exportRow: (catalog) => [
        catalog.name.isNotEmpty ? catalog.name : catalog.id,
        '${catalog.version}',
        catalog.currency,
        catalog.hasPublishedAt() ? 'Published' : 'Draft',
        _formatTimestamp(catalog.effectiveAt),
        '${catalog.plans.length}',
      ],
      onExport: (format, rowCount) {
        debugPrint(
            'AUDIT: Exported $rowCount catalog versions as $format');
      },
    );
  }

  Widget _publishedBadge(CatalogVersion catalog) {
    final published = catalog.hasPublishedAt();
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: published
            ? Colors.green.withAlpha(25)
            : Colors.amber.withAlpha(25),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        published ? 'Published' : 'Draft',
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w600,
          color: published ? Colors.green : Colors.amber,
        ),
      ),
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
          '${dt.day.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }
}
