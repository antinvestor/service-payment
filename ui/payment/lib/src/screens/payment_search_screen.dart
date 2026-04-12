import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/payment_providers.dart';
import '../utils/money_format.dart';

/// Screen that lists payments with search functionality using DataTable.
class PaymentSearchScreen extends ConsumerStatefulWidget {
  const PaymentSearchScreen({super.key});

  @override
  ConsumerState<PaymentSearchScreen> createState() =>
      _PaymentSearchScreenState();
}

class _PaymentSearchScreenState extends ConsumerState<PaymentSearchScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncPayments = ref.watch(paymentSearchProvider(_searchQuery));

    return asyncPayments.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (payments) => _buildTable(items: payments),
    );
  }

  Widget _buildTable({required List<Payment> items}) {
    return AdminEntityListPage<Payment>(
      title: 'Payments',
      breadcrumbs: const ['Payments', 'Search'],
      searchHint: 'Search by reference, account, route...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      onAdd: () => context.go('/payments/send'),
      addLabel: 'Send Payment',
      columns: const [
        DataColumn(label: Text('ID')),
        DataColumn(label: Text('Route')),
        DataColumn(label: Text('Source')),
        DataColumn(label: Text('Recipient')),
        DataColumn(label: Text('Amount')),
        DataColumn(label: Text('Status')),
        DataColumn(label: Text('Date')),
      ],
      items: items,
      rowBuilder: (payment, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(_truncate(payment.id, 12))),
          DataCell(Text(payment.route)),
          DataCell(Text(payment.source.detail.isNotEmpty
              ? payment.source.detail
              : payment.source.profileName)),
          DataCell(Text(payment.recipient.detail.isNotEmpty
              ? payment.recipient.detail
              : payment.recipient.profileName)),
          DataCell(Text(fmtMoney(payment.amount))),
          DataCell(Text(payment.status.name)),
          DataCell(Text(payment.dateCreated.isNotEmpty
              ? _formatDate(payment.dateCreated)
              : '')),
        ],
      ),
      onRowNavigate: (payment) => context.go(
        '/payments/detail/${payment.id}',
        extra: payment,
      ),
      exportRow: (payment) => [
        payment.id,
        payment.route,
        payment.source.detail.isNotEmpty
            ? payment.source.detail
            : payment.source.profileName,
        payment.recipient.detail.isNotEmpty
            ? payment.recipient.detail
            : payment.recipient.profileName,
        fmtMoney(payment.amount),
        payment.status.name,
        payment.dateCreated,
      ],
      onExport: (format, rowCount) {
        debugPrint(
            'AUDIT: Exported $rowCount payments as $format');
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
