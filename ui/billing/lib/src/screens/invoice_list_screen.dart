import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/invoice_providers.dart';
import '../widgets/invoice_card.dart';

/// Screen that lists invoices with search and InvoiceStateBadge per item.
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
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (invoices) => _buildShell(items: invoices),
    );
  }

  Widget _buildShell({
    required List<Invoice> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Invoice>(
      title: 'Invoices',
      icon: Icons.receipt_long,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () => ref.invalidate(invoiceSearchProvider(_searchQuery)),
      searchHint: 'Search invoices...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      itemBuilder: (context, invoice) {
        return InvoiceCard(
          invoice: invoice,
          onTap: () => context.go('/billing/invoices/${invoice.id}'),
        );
      },
    );
  }
}
