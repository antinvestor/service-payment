import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/transaction_providers.dart';
import '../widgets/transaction_type_badge.dart';

/// Screen that lists transactions with search and type badges using DataTable.
class TransactionListScreen extends ConsumerStatefulWidget {
  const TransactionListScreen({super.key});

  @override
  ConsumerState<TransactionListScreen> createState() =>
      _TransactionListScreenState();
}

class _TransactionListScreenState extends ConsumerState<TransactionListScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncTransactions = ref.watch(
      transactionSearchProvider(_searchQuery),
    );

    return asyncTransactions.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (transactions) => _buildTable(items: transactions),
    );
  }

  Widget _buildTable({required List<Transaction> items}) {
    return AdminEntityListPage<Transaction>(
      title: 'Transactions',
      breadcrumbs: const ['Ledger', 'Transactions'],
      searchHint: 'Search transactions...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      columns: const [
        DataColumn(label: Text('ID')),
        DataColumn(label: Text('Currency')),
        DataColumn(label: Text('Type')),
        DataColumn(label: Text('Transacted At')),
        DataColumn(label: Text('Entries'), numeric: true),
        DataColumn(label: Text('Cleared')),
      ],
      items: items,
      rowBuilder: (txn, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(_truncate(txn.id, 16))),
          DataCell(Text(txn.currencyCode)),
          DataCell(TransactionTypeBadge(type: txn.type)),
          DataCell(
            Text(
              txn.transactedAt.isNotEmpty ? _formatDate(txn.transactedAt) : '',
            ),
          ),
          DataCell(Text('${txn.entries.length}')),
          DataCell(
            Icon(
              txn.cleared ? Icons.check_circle : Icons.cancel,
              size: 18,
              color: txn.cleared ? Colors.green : Colors.grey,
            ),
          ),
        ],
      ),
      onRowNavigate: (txn) => context.go('/ledger/transactions/${txn.id}'),
      exportRow: (txn) => [
        txn.id,
        txn.currencyCode,
        transactionTypeLabel(txn.type),
        txn.transactedAt,
        '${txn.entries.length}',
        txn.cleared ? 'Yes' : 'No',
      ],
      onExport: (format, rowCount) {
        debugPrint('AUDIT: Exported $rowCount transactions as $format');
      },
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  String _formatDate(String rfc3339) {
    try {
      final dt = DateTime.parse(rfc3339);
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')}';
    } catch (_) {
      return rfc3339.length > 10 ? rfc3339.substring(0, 10) : rfc3339;
    }
  }
}
