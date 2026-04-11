import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/payment_providers.dart';
import '../widgets/payment_card.dart';

/// Screen that lists payments with search functionality.
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
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (payments) => _buildShell(items: payments),
    );
  }

  Widget _buildShell({
    required List<Payment> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Payment>(
      title: 'Payments',
      icon: Icons.payments,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () => ref.invalidate(paymentSearchProvider(_searchQuery)),
      searchHint: 'Search by reference, account, route...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      actionLabel: 'Send Payment',
      onAction: () => context.go('/payments/send'),
      itemBuilder: (context, payment) {
        return PaymentCard(
          payment: payment,
          onTap: () => context.go(
            '/payments/detail/${payment.id}',
            extra: payment,
          ),
        );
      },
    );
  }
}
