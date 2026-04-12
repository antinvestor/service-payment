import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/payment_providers.dart';
import '../utils/money_format.dart';
import '../widgets/payment_status_badge.dart';

/// Screen for viewing full payment details with actions.
class PaymentDetailScreen extends ConsumerWidget {
  const PaymentDetailScreen({
    super.key,
    required this.paymentId,
    this.initialPayment,
  });

  final String paymentId;
  final Payment? initialPayment;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);

    // If we have an initial payment, show it; otherwise search by ID.
    if (initialPayment != null) {
      return _buildDetail(context, ref, theme, initialPayment!);
    }

    final asyncPayments = ref.watch(paymentSearchProvider(paymentId));
    return asyncPayments.when(
      loading: () => Scaffold(
        appBar: _buildAppBar(context, theme, 'Payment'),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Scaffold(
        appBar: _buildAppBar(context, theme, 'Payment'),
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.error_outline, size: 48, color: theme.colorScheme.error),
              const SizedBox(height: 16),
              Text(friendlyError(error)),
              const SizedBox(height: 16),
              FilledButton.tonal(
                onPressed: () => ref.invalidate(paymentSearchProvider(paymentId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (payments) {
        final payment = payments.isNotEmpty ? payments.first : null;
        if (payment == null) {
          return Scaffold(
            appBar: _buildAppBar(context, theme, 'Payment'),
            body: Center(
              child: Text(
                'Payment not found',
                style: theme.textTheme.bodyLarge,
              ),
            ),
          );
        }
        return _buildDetail(context, ref, theme, payment);
      },
    );
  }

  PreferredSizeWidget _buildAppBar(
    BuildContext context,
    ThemeData theme,
    String title,
  ) {
    return AppBar(
      leading: IconButton(
        icon: const Icon(Icons.arrow_back),
        onPressed: () =>
            context.canPop() ? context.pop() : context.go('/payments'),
      ),
      title: Text(
        title,
        style: theme.textTheme.titleMedium?.copyWith(
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  Widget _buildDetail(
    BuildContext context,
    WidgetRef ref,
    ThemeData theme,
    Payment payment,
  ) {
    final isOutbound = payment.outbound;

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () =>
              context.canPop() ? context.pop() : context.go('/payments'),
        ),
        title: Text(
          'Payment ${_truncate(payment.id, 12)}',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () =>
                ref.invalidate(paymentSearchProvider(payment.id)),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Amount hero section
            Card(
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(color: theme.colorScheme.outlineVariant),
              ),
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    // Direction icon
                    Container(
                      width: 56,
                      height: 56,
                      decoration: BoxDecoration(
                        color: isOutbound
                            ? Colors.red.withAlpha(25)
                            : Colors.green.withAlpha(25),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Icon(
                        isOutbound ? Icons.arrow_upward : Icons.arrow_downward,
                        size: 28,
                        color: isOutbound ? Colors.red : Colors.green,
                      ),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      fmtMoney(payment.amount),
                      style: theme.textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: isOutbound ? Colors.red : Colors.green,
                      ),
                    ),
                    if (payment.hasCost()) ...[
                      const SizedBox(height: 4),
                      Text(
                        'Cost: ${fmtMoney(payment.cost)}',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        PaymentStatusBadge(status: payment.status.name),
                        const SizedBox(width: 8),
                        PaymentStateBadge(state: payment.state.name),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Details section
            _sectionCard(
              theme,
              'Payment Details',
              [
                _metadataRow(theme, 'ID', payment.id, copyable: true, context: context),
                _metadataRow(theme, 'Transaction ID', payment.transactionId,
                    copyable: true, context: context),
                _metadataRow(theme, 'Reference ID', payment.referenceId,
                    copyable: true, context: context),
                _metadataRow(theme, 'Route', payment.route),
                _metadataRow(
                  theme,
                  'Direction',
                  isOutbound ? 'Outbound' : 'Inbound',
                ),
                if (payment.hasDateCreated())
                  _metadataRow(
                    theme,
                    'Date Created',
                    _formatTimestamp(payment.dateCreated),
                  ),
              ],
            ),
            const SizedBox(height: 16),

            // Source account
            _contactCard(theme, 'Source', payment.source),
            const SizedBox(height: 16),

            // Recipient account
            _contactCard(theme, 'Recipient', payment.recipient),
            const SizedBox(height: 16),

            // Extra data
            if (payment.extra.fields.isNotEmpty) ...[
              _sectionCard(
                theme,
                'Extra Data',
                payment.extra.fields.entries.map((e) {
                  return _metadataRow(theme, e.key, e.value.stringValue);
                }).toList(),
              ),
              const SizedBox(height: 16),
            ],

            // Action buttons
            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => _release(context, ref, payment),
                    icon: const Icon(Icons.lock_open, size: 18),
                    label: const Text('Release'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => _reconcile(context, ref, payment),
                    icon: const Icon(Icons.sync, size: 18),
                    label: const Text('Reconcile'),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _sectionCard(
    ThemeData theme,
    String title,
    List<Widget> children,
  ) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: theme.colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w600,
                color: theme.colorScheme.primary,
              ),
            ),
            const SizedBox(height: 12),
            ...children,
          ],
        ),
      ),
    );
  }

  Widget _contactCard(ThemeData theme, String label, ContactLink contact) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: theme.colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  label == 'Source' ? Icons.output : Icons.input,
                  size: 18,
                  color: theme.colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Text(
                  label,
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: theme.colorScheme.primary,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (contact.detail.isNotEmpty)
              _metadataRow(theme, 'Detail', contact.detail),
            if (contact.profileName.isNotEmpty)
              _metadataRow(theme, 'Name', contact.profileName),
            if (contact.profileId.isNotEmpty)
              _metadataRow(theme, 'Profile ID', contact.profileId),
          ],
        ),
      ),
    );
  }

  Widget _metadataRow(
    ThemeData theme,
    String label,
    String value, {
    bool copyable = false,
    BuildContext? context,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 130,
            child: Text(
              label,
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value.isNotEmpty ? value : '\u2014',
              style: theme.textTheme.bodyMedium?.copyWith(
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          if (copyable && value.isNotEmpty && context != null)
            IconButton(
              icon: Icon(
                Icons.copy,
                size: 14,
                color: theme.colorScheme.onSurfaceVariant,
              ),
              constraints: const BoxConstraints(minWidth: 24, minHeight: 24),
              padding: EdgeInsets.zero,
              tooltip: 'Copy',
              onPressed: () {
                Clipboard.setData(ClipboardData(text: value));
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('$label copied'),
                    behavior: SnackBarBehavior.floating,
                    duration: const Duration(seconds: 2),
                  ),
                );
              },
            ),
        ],
      ),
    );
  }

  String _formatTimestamp(dynamic ts) {
    try {
      final dt = ts.toDateTime();
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')} '
          '${dt.hour.toString().padLeft(2, '0')}:'
          '${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return 'N/A';
    }
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  Future<void> _release(
    BuildContext context,
    WidgetRef ref,
    Payment payment,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Release Payment'),
        content: Text('Release payment ${_truncate(payment.id, 16)}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Release'),
          ),
        ],
      ),
    );

    if (confirmed != true || !context.mounted) return;

    try {
      final notifier = ref.read(paymentNotifierProvider.notifier);
      final request = ReleaseRequest()..id = payment.id;
      await notifier.release(request);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Payment released successfully'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        ref.invalidate(paymentSearchProvider(payment.id));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Release failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }

  Future<void> _reconcile(
    BuildContext context,
    WidgetRef ref,
    Payment payment,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Reconcile Payment'),
        content: Text('Reconcile payment ${_truncate(payment.id, 16)}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Reconcile'),
          ),
        ],
      ),
    );

    if (confirmed != true || !context.mounted) return;

    try {
      final notifier = ref.read(paymentNotifierProvider.notifier);
      final request = ReconcileRequest()
        ..externalTransactionId = payment.externalTransactionId
        ..route = payment.route
        ..outbound = payment.outbound
        ..amount = payment.amount
        ..owner = payment.source.profileId
        ..countryCode = '';
      await notifier.reconcile(request);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Payment reconciled successfully'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        ref.invalidate(paymentSearchProvider(payment.id));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Reconciliation failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }
}
