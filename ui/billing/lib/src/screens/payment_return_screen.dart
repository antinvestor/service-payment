import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../api/collection_client.dart';
import '../providers/collection_providers.dart';

/// Landing page after hosted checkout redirect.
///
/// Query params: `session` (required), `status` (optional UX hint only).
/// Always confirms payment server-side via ConfirmPayment.
class PaymentReturnScreen extends ConsumerStatefulWidget {
  const PaymentReturnScreen({
    super.key,
    required this.sessionRef,
    this.statusHint,
  });

  final String sessionRef;
  final String? statusHint;

  @override
  ConsumerState<PaymentReturnScreen> createState() =>
      _PaymentReturnScreenState();
}

class _PaymentReturnScreenState extends ConsumerState<PaymentReturnScreen> {
  bool _loading = true;
  String? _error;
  ConfirmPaymentResult? _result;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _confirm());
  }

  Future<void> _confirm() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final result = await ref
          .read(collectionNotifierProvider.notifier)
          .confirmPayment(widget.sessionRef);
      if (!mounted) return;
      setState(() {
        _result = result;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = friendlyError(e);
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Payment status',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 480),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: _buildBody(theme),
          ),
        ),
      ),
    );
  }

  Widget _buildBody(ThemeData theme) {
    if (_loading) {
      return const Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          CircularProgressIndicator(),
          SizedBox(height: 16),
          Text('Confirming payment…'),
        ],
      );
    }

    if (_error != null) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.hourglass_top, size: 48, color: theme.colorScheme.primary),
          const SizedBox(height: 16),
          Text(
            _error!,
            textAlign: TextAlign.center,
            style: theme.textTheme.bodyLarge,
          ),
          const SizedBox(height: 8),
          Text(
            'If the customer just finished paying, wait a moment and retry.',
            textAlign: TextAlign.center,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: _confirm,
            child: const Text('Retry confirm'),
          ),
          const SizedBox(height: 12),
          TextButton(
            onPressed: () => context.go('/billing/invoices'),
            child: const Text('Back to invoices'),
          ),
        ],
      );
    }

    final result = _result!;
    final paid = result.paid;

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(
          paid ? Icons.check_circle : Icons.info_outline,
          size: 56,
          color: paid ? Colors.green : theme.colorScheme.primary,
        ),
        const SizedBox(height: 16),
        Text(
          paid ? 'Payment confirmed' : 'Payment not complete',
          style: theme.textTheme.headlineSmall?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 12),
        if (result.invoiceId.isNotEmpty)
          Text('Invoice: ${result.invoiceId}'),
        if (result.invoiceState.isNotEmpty)
          Text('Invoice state: ${result.invoiceState}'),
        if (result.subscriptionId.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text('Subscription: ${result.subscriptionId}'),
          Text('Subscription state: ${result.subscriptionState}'),
        ],
        const SizedBox(height: 24),
        if (result.subscriptionId.isNotEmpty)
          FilledButton(
            onPressed: () =>
                context.go('/billing/subscriptions/${result.subscriptionId}'),
            child: const Text('View subscription'),
          )
        else if (result.invoiceId.isNotEmpty)
          FilledButton(
            onPressed: () => context.go('/billing/invoices/${result.invoiceId}'),
            child: const Text('View invoice'),
          ),
        const SizedBox(height: 8),
        TextButton(
          onPressed: () => context.go('/billing/invoices'),
          child: const Text('All invoices'),
        ),
      ],
    );
  }
}
