import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/payment_link_providers.dart';
import '../widgets/payment_link_card.dart';

/// Screen that lists payment links with search functionality.
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
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (links) => _buildShell(items: links),
    );
  }

  Widget _buildShell({
    required List<Payment> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Payment>(
      title: 'Payment Links',
      icon: Icons.link,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () => ref.invalidate(paymentLinkSearchProvider(_searchQuery)),
      searchHint: 'Search payment links...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      actionLabel: 'Create Link',
      onAction: () => _showCreateDialog(context),
      itemBuilder: (context, payment) {
        return PaymentLinkCard(
          payment: payment,
        );
      },
    );
  }

  Future<void> _showCreateDialog(BuildContext context) async {
    final theme = Theme.of(context);
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
