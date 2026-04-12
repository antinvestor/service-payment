import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/invoice_providers.dart';
import '../utils/money_format.dart';
import '../widgets/invoice_line_tile.dart';
import '../widgets/invoice_state_badge.dart';

/// Screen showing details for a single invoice with amounts, lines,
/// and actions (issue, void, record payment).
class InvoiceDetailScreen extends ConsumerWidget {
  const InvoiceDetailScreen({
    super.key,
    required this.invoiceId,
  });

  final String invoiceId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncInvoice = ref.watch(invoiceProvider(invoiceId));

    return asyncInvoice.when(
      loading: () => Scaffold(
        appBar: _buildAppBar(context, theme, 'Invoice'),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Scaffold(
        appBar: _buildAppBar(context, theme, 'Invoice'),
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.error_outline,
                  size: 48, color: theme.colorScheme.error),
              const SizedBox(height: 16),
              Text(friendlyError(error)),
              const SizedBox(height: 16),
              FilledButton.tonal(
                onPressed: () =>
                    ref.invalidate(invoiceProvider(invoiceId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (invoice) => _buildDetail(context, ref, theme, invoice),
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
        onPressed: () => context.canPop()
            ? context.pop()
            : context.go('/billing/invoices'),
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
    Invoice invoice,
  ) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.canPop()
              ? context.pop()
              : context.go('/billing/invoices'),
        ),
        title: Text(
          invoice.invoiceNumber.isNotEmpty
              ? invoice.invoiceNumber
              : 'Invoice ${_truncate(invoice.id, 12)}',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () =>
                ref.invalidate(invoiceProvider(invoiceId)),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Amount hero card
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
                    Text(
                      fmtMoney(invoice.totalAmount),
                      style: theme.textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    if (invoice.currency.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        invoice.currency,
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                    const SizedBox(height: 12),
                    InvoiceStateBadge(state: invoice.state),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Details card
            Card(
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
                      'Invoice Details',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 12),
                    _metadataRow(theme, 'ID', invoice.id,
                        copyable: true, context: context),
                    if (invoice.invoiceNumber.isNotEmpty)
                      _metadataRow(
                          theme, 'Number', invoice.invoiceNumber),
                    _metadataRow(theme, 'Subtotal',
                        fmtMoney(invoice.subtotalAmount)),
                    if (invoice.hasDiscountAmount())
                      _metadataRow(theme, 'Discount',
                          fmtMoney(invoice.discountAmount)),
                    if (invoice.hasCreditAmount())
                      _metadataRow(theme, 'Credit',
                          fmtMoney(invoice.creditAmount)),
                    _metadataRow(
                        theme, 'Total', fmtMoney(invoice.totalAmount)),
                    if (invoice.hasDueAt())
                      _metadataRow(
                        theme,
                        'Due Date',
                        _formatTimestamp(invoice.dueAt),
                      ),
                    if (invoice.hasPeriodStart() && invoice.hasPeriodEnd())
                      _metadataRow(
                        theme,
                        'Period',
                        '${_formatDate(invoice.periodStart.toDateTime())} - '
                            '${_formatDate(invoice.periodEnd.toDateTime())}',
                      ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Invoice lines
            if (invoice.lines.isNotEmpty) ...[
              Text(
                'Line Items',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(height: 8),
              Card(
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: BorderSide(color: theme.colorScheme.outlineVariant),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      const InvoiceLineHeader(),
                      const Divider(height: 16),
                      ...invoice.lines
                          .map((line) => InvoiceLineTile(line: line)),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 24),
            ],

            // Action buttons
            _buildActions(context, ref, theme, invoice),
          ],
        ),
      ),
    );
  }

  Widget _buildActions(
    BuildContext context,
    WidgetRef ref,
    ThemeData theme,
    Invoice invoice,
  ) {
    final actions = <Widget>[];

    if (invoice.state == InvoiceState.INVOICE_DRAFT) {
      actions.add(
        Expanded(
          child: FilledButton.icon(
            onPressed: () => _issueInvoice(context, ref, invoice),
            icon: const Icon(Icons.send, size: 18),
            label: const Text('Issue'),
          ),
        ),
      );
    }

    if (invoice.state == InvoiceState.INVOICE_ISSUED ||
        invoice.state == InvoiceState.INVOICE_OVERDUE) {
      actions.add(
        Expanded(
          child: FilledButton.icon(
            onPressed: () => _recordPayment(context, ref, invoice),
            icon: const Icon(Icons.payment, size: 18),
            label: const Text('Record Payment'),
          ),
        ),
      );
    }

    if (invoice.state != InvoiceState.INVOICE_VOIDED &&
        invoice.state != InvoiceState.INVOICE_PAID) {
      if (actions.isNotEmpty) {
        actions.add(const SizedBox(width: 12));
      }
      actions.add(
        Expanded(
          child: OutlinedButton.icon(
            onPressed: () => _voidInvoice(context, ref, invoice),
            icon: const Icon(Icons.block, size: 18),
            label: const Text('Void'),
            style: OutlinedButton.styleFrom(
              foregroundColor: Colors.red,
              side: const BorderSide(color: Colors.red),
            ),
          ),
        ),
      );
    }

    if (actions.isEmpty) return const SizedBox.shrink();

    return Row(children: actions);
  }

  Future<void> _issueInvoice(
    BuildContext context,
    WidgetRef ref,
    Invoice invoice,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Issue Invoice'),
        content: Text('Issue invoice ${invoice.invoiceNumber.isNotEmpty ? invoice.invoiceNumber : _truncate(invoice.id, 16)}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Issue'),
          ),
        ],
      ),
    );

    if (confirmed != true || !context.mounted) return;

    try {
      final notifier = ref.read(invoiceNotifierProvider.notifier);
      final request = IssueInvoiceRequest()..id = invoice.id;
      await notifier.issue(request);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Invoice issued successfully'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        ref.invalidate(invoiceProvider(invoice.id));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Issue failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }

  Future<void> _voidInvoice(
    BuildContext context,
    WidgetRef ref,
    Invoice invoice,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Void Invoice'),
        content: Text('Void invoice ${invoice.invoiceNumber.isNotEmpty ? invoice.invoiceNumber : _truncate(invoice.id, 16)}? This cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Void'),
          ),
        ],
      ),
    );

    if (confirmed != true || !context.mounted) return;

    try {
      final notifier = ref.read(invoiceNotifierProvider.notifier);
      final request = VoidInvoiceRequest()..id = invoice.id;
      await notifier.voidInvoice(request);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Invoice voided'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        ref.invalidate(invoiceProvider(invoice.id));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Void failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }

  Future<void> _recordPayment(
    BuildContext context,
    WidgetRef ref,
    Invoice invoice,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Record Payment'),
        content: Text(
          'Record full payment of ${fmtMoney(invoice.totalAmount)} '
          'for invoice ${invoice.invoiceNumber.isNotEmpty ? invoice.invoiceNumber : _truncate(invoice.id, 16)}?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Record Payment'),
          ),
        ],
      ),
    );

    if (confirmed != true || !context.mounted) return;

    try {
      final notifier = ref.read(invoiceNotifierProvider.notifier);
      final request = RecordPaymentRequest()
        ..id = invoice.id;
      await notifier.recordPayment(request);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Payment recorded successfully'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        ref.invalidate(invoiceProvider(invoice.id));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Record payment failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
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
            width: 100,
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

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }
}
