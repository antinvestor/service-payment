import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/payment_link_providers.dart';

/// Screen that lists payment links with search functionality using DataTable.
class PaymentLinksScreen extends ConsumerStatefulWidget {
  const PaymentLinksScreen({super.key});

  @override
  ConsumerState<PaymentLinksScreen> createState() => _PaymentLinksScreenState();
}

class _PaymentLinksScreenState extends ConsumerState<PaymentLinksScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncLinks = ref.watch(paymentLinkSearchProvider(_searchQuery));

    return asyncLinks.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (links) => _buildTable(items: links),
    );
  }

  Widget _buildTable({required List<Payment> items}) {
    return AdminEntityListPage<Payment>(
      title: 'Payment Links',
      breadcrumbs: const ['Payments', 'Links'],
      searchHint: 'Search payment links...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      onAdd: () => _showCreateDialog(context),
      addLabel: 'Create Link',
      columns: const [
        DataColumn(label: Text('Name')),
        DataColumn(label: Text('Amount')),
        DataColumn(label: Text('Currency')),
        DataColumn(label: Text('Type')),
        DataColumn(label: Text('Expiry')),
        DataColumn(label: Text('Ref')),
      ],
      items: items,
      rowBuilder: (link, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(link.source.profileName.isNotEmpty
              ? link.source.profileName
              : link.id)),
          DataCell(Text(formatMoney(link.amount))),
          DataCell(Text(link.amount.currencyCode)),
          DataCell(Text(link.route.isNotEmpty ? link.route : '')),
          DataCell(Text(link.dateCreated.isNotEmpty
              ? _formatDate(link.dateCreated)
              : '')),
          DataCell(Text(link.referenceId)),
        ],
      ),
      exportRow: (link) => [
        link.source.profileName.isNotEmpty
            ? link.source.profileName
            : link.id,
        formatMoney(link.amount),
        link.amount.currencyCode,
        link.route,
        link.dateCreated,
        link.referenceId,
      ],
      onExport: (format, rowCount) {
        debugPrint(
            'AUDIT: Exported $rowCount payment links as $format');
      },
    );
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

  Future<void> _showCreateDialog(BuildContext context) async {
    final referenceController = TextEditingController();
    final amountController = TextEditingController();

    await showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Create Payment Link'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: referenceController,
              decoration: const InputDecoration(
                labelText: 'Reference',
                hintText: 'Payment reference...',
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: amountController,
              keyboardType:
                  const TextInputType.numberWithOptions(decimal: true),
              decoration: const InputDecoration(
                labelText: 'Amount',
                hintText: '0.00',
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () async {
              try {
                final paymentLink = PaymentLink()
                  ..externalRef = referenceController.text.trim();
                final request = CreatePaymentLinkRequest()
                  ..paymentLink = paymentLink;
                final notifier =
                    ref.read(paymentLinkNotifierProvider.notifier);
                await notifier.create(request);
                if (ctx.mounted) Navigator.pop(ctx);
                if (mounted) {
                  ref.invalidate(paymentLinkSearchProvider(_searchQuery));
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Payment link created'),
                      behavior: SnackBarBehavior.floating,
                    ),
                  );
                }
              } catch (e) {
                if (ctx.mounted) {
                  ScaffoldMessenger.of(ctx).showSnackBar(
                    SnackBar(
                      content: Text(friendlyError(e)),
                      behavior: SnackBarBehavior.floating,
                    ),
                  );
                }
              }
            },
            child: const Text('Create'),
          ),
        ],
      ),
    );

    referenceController.dispose();
    amountController.dispose();
  }
}
