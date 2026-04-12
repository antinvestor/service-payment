import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/invoice_providers.dart';
import '../utils/money_format.dart';
import '../widgets/invoice_state_badge.dart';

/// Screen that lists invoices with search and InvoiceStateBadge using DataTable.
class InvoiceListScreen extends ConsumerStatefulWidget {
  const InvoiceListScreen({super.key});

  @override
  ConsumerState<InvoiceListScreen> createState() => _InvoiceListScreenState();
}

class _InvoiceListScreenState extends ConsumerState<InvoiceListScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncInvoices = ref.watch(invoiceSearchProvider(_searchQuery));

    return asyncInvoices.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (invoices) => _buildTable(items: invoices),
    );
  }

  Widget _buildTable({required List<Invoice> items}) {
    return AdminEntityListPage<Invoice>(
      title: 'Invoices',
      breadcrumbs: const ['Billing', 'Invoices'],
      searchHint: 'Search invoices...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      columns: const [
        DataColumn(label: Text('Number')),
        DataColumn(label: Text('Profile')),
        DataColumn(label: Text('State')),
        DataColumn(label: Text('Subtotal'), numeric: true),
        DataColumn(label: Text('Total'), numeric: true),
        DataColumn(label: Text('Due Date')),
        DataColumn(label: Text('Issued')),
      ],
      items: items,
      rowBuilder: (invoice, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(invoice.invoiceNumber.isNotEmpty
              ? invoice.invoiceNumber
              : _truncate(invoice.id, 12))),
          DataCell(Text(_truncate(invoice.profileId, 12))),
          DataCell(InvoiceStateBadge(state: invoice.state)),
          DataCell(Text(fmtMoney(invoice.subtotalAmount))),
          DataCell(Text(fmtMoney(invoice.totalAmount))),
          DataCell(Text(_formatTimestamp(invoice.dueAt))),
          DataCell(Text(_formatTimestamp(invoice.issuedAt))),
        ],
      ),
      onRowNavigate: (invoice) =>
          context.go('/billing/invoices/${invoice.id}'),
      exportRow: (invoice) => [
        invoice.invoiceNumber.isNotEmpty
            ? invoice.invoiceNumber
            : invoice.id,
        invoice.profileId,
        invoiceStateLabel(invoice.state),
        fmtMoney(invoice.subtotalAmount),
        fmtMoney(invoice.totalAmount),
        _formatTimestamp(invoice.dueAt),
        _formatTimestamp(invoice.issuedAt),
      ],
      onExport: (format, rowCount) {
        debugPrint('AUDIT: Exported $rowCount invoices as $format');
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
          '${dt.day.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }
}
