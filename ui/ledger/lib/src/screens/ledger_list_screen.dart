import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/ledger_providers.dart';
import '../widgets/ledger_type_badge.dart';

/// Screen that lists ledgers with search and LedgerType filter using DataTable.
class LedgerListScreen extends ConsumerStatefulWidget {
  const LedgerListScreen({super.key});

  @override
  ConsumerState<LedgerListScreen> createState() => _LedgerListScreenState();
}

class _LedgerListScreenState extends ConsumerState<LedgerListScreen> {
  String _searchQuery = '';
  LedgerType? _typeFilter;

  @override
  Widget build(BuildContext context) {
    final asyncLedgers = ref.watch(ledgerSearchProvider(_searchQuery));

    return asyncLedgers.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (ledgers) {
        final filtered = _typeFilter == null
            ? ledgers
            : ledgers.where((l) => l.type == _typeFilter).toList();
        return _buildTable(items: filtered);
      },
    );
  }

  Widget _buildTable({required List<Ledger> items}) {
    return AdminEntityListPage<Ledger>(
      title: 'Ledgers',
      breadcrumbs: const ['Ledger', 'Ledgers'],
      searchHint: 'Search ledgers...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      actions: [_buildTypeFilter()],
      columns: const [
        DataColumn(label: Text('ID')),
        DataColumn(label: Text('Type')),
        DataColumn(label: Text('Parent')),
        DataColumn(label: Text('Data')),
      ],
      items: items,
      rowBuilder: (ledger, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(_truncate(ledger.id, 16))),
          DataCell(LedgerTypeBadge(type: ledger.type)),
          DataCell(Text(ledger.parent)),
          DataCell(Text(_dataPreview(ledger))),
        ],
      ),
      onRowNavigate: (ledger) =>
          context.go('/ledger/ledgers/${ledger.id}'),
      exportRow: (ledger) => [
        ledger.id,
        ledgerTypeLabel(ledger.type),
        ledger.parent,
        _dataPreview(ledger),
      ],
      onExport: (format, rowCount) {
        debugPrint('AUDIT: Exported $rowCount ledgers as $format');
      },
    );
  }

  Widget _buildTypeFilter() {
    return DropdownButton<LedgerType?>(
      value: _typeFilter,
      hint: const Text('All Types'),
      underline: const SizedBox.shrink(),
      borderRadius: BorderRadius.circular(8),
      items: [
        const DropdownMenuItem<LedgerType?>(
          value: null,
          child: Text('All Types'),
        ),
        ...LedgerType.values
            .map((type) => DropdownMenuItem<LedgerType?>(
                  value: type,
                  child: Text(ledgerTypeLabel(type)),
                )),
      ],
      onChanged: (value) {
        setState(() => _typeFilter = value);
      },
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  String _dataPreview(Ledger ledger) {
    try {
      final fields = ledger.data.fields;
      if (fields.isEmpty) return '';
      final keys = fields.keys.take(3).join(', ');
      return '{$keys${fields.length > 3 ? ', ...' : ''}}';
    } catch (_) {
      return '';
    }
  }
}
