import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/transaction_providers.dart';
import '../widgets/transaction_entry_row.dart';
import '../widgets/transaction_type_badge.dart';

/// Screen showing details for a single transaction including type badge,
/// currency, timestamp, and entry list.
class TransactionDetailScreen extends ConsumerWidget {
  const TransactionDetailScreen({super.key, required this.transactionId});

  final String transactionId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncTransactions = ref.watch(
      transactionSearchProvider(transactionId),
    );

    return asyncTransactions.when(
      loading: () => Scaffold(
        appBar: _buildAppBar(context, theme, 'Transaction'),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Scaffold(
        appBar: _buildAppBar(context, theme, 'Transaction'),
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.error_outline,
                size: 48,
                color: theme.colorScheme.error,
              ),
              const SizedBox(height: 16),
              Text(friendlyError(error)),
              const SizedBox(height: 16),
              FilledButton.tonal(
                onPressed: () =>
                    ref.invalidate(transactionSearchProvider(transactionId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (transactions) {
        final transaction = transactions.isNotEmpty ? transactions.first : null;
        if (transaction == null) {
          return Scaffold(
            appBar: _buildAppBar(context, theme, 'Transaction'),
            body: Center(
              child: Text(
                'Transaction not found',
                style: theme.textTheme.bodyLarge,
              ),
            ),
          );
        }
        return _buildDetail(context, ref, theme, transaction);
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
        onPressed: () => context.canPop()
            ? context.pop()
            : context.go('/ledger/transactions'),
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
    Transaction transaction,
  ) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.canPop()
              ? context.pop()
              : context.go('/ledger/transactions'),
        ),
        title: Text(
          'Transaction ${_truncate(transaction.id, 12)}',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        actions: [
          _LifecycleActions(transaction: transaction),
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () {
              ref.invalidate(transactionSearchProvider(transactionId));
              ref.invalidate(transactionEntrySearchProvider(transactionId));
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // Transaction header card
          Padding(
            padding: const EdgeInsets.all(24),
            child: Card(
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(color: theme.colorScheme.outlineVariant),
              ),
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    // Entry count (Transaction has no top-level amount)
                    Text(
                      '${transaction.entries.length} '
                      'entr${transaction.entries.length == 1 ? 'y' : 'ies'}',
                      style: theme.textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 12),

                    // Type badge and currency
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        TransactionTypeBadge(type: transaction.type),
                        if (transaction.currencyCode.isNotEmpty) ...[
                          const SizedBox(width: 8),
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 10,
                              vertical: 4,
                            ),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.surfaceContainerLow,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Text(
                              transaction.currencyCode,
                              style: theme.textTheme.labelMedium?.copyWith(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                    const SizedBox(height: 16),

                    // Metadata rows
                    _metadataRow(
                      theme,
                      'ID',
                      transaction.id,
                      copyable: true,
                      context: context,
                    ),
                    if (transaction.transactedAt.isNotEmpty)
                      _metadataRow(
                        theme,
                        'Transacted',
                        _formatDateString(transaction.transactedAt),
                      ),
                  ],
                ),
              ),
            ),
          ),

          // Entries header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Row(
              children: [
                Text(
                  'Entries',
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: theme.colorScheme.primary,
                  ),
                ),
                const Spacer(),
                if (transaction.entries.isNotEmpty)
                  Text(
                    '${transaction.entries.length} '
                    'entr${transaction.entries.length == 1 ? 'y' : 'ies'}',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 8),

          // Entries list
          Expanded(child: _EntriesList(transactionId: transaction.id)),
        ],
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

  String _formatDateString(String rfc3339) {
    try {
      final dt = DateTime.parse(rfc3339);
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')} '
          '${dt.hour.toString().padLeft(2, '0')}:'
          '${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return rfc3339.length > 16 ? rfc3339.substring(0, 16) : rfc3339;
    }
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }
}

class _EntriesList extends ConsumerWidget {
  const _EntriesList({required this.transactionId});

  final String transactionId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncEntries = ref.watch(
      transactionEntrySearchProvider(transactionId),
    );

    return asyncEntries.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (error, _) => Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, size: 48, color: theme.colorScheme.error),
            const SizedBox(height: 16),
            Text(friendlyError(error)),
            const SizedBox(height: 16),
            FilledButton.tonal(
              onPressed: () =>
                  ref.invalidate(transactionEntrySearchProvider(transactionId)),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
      data: (entries) {
        if (entries.isEmpty) {
          return Center(
            child: Text(
              'No entries found',
              style: theme.textTheme.bodyLarge?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          );
        }

        return ListView.builder(
          padding: const EdgeInsets.symmetric(horizontal: 8),
          itemCount: entries.length,
          itemBuilder: (context, index) {
            return TransactionEntryRow(
              entry: entries[index],
              onTap: () =>
                  context.go('/ledger/accounts/${entries[index].accountId}'),
            );
          },
        );
      },
    );
  }
}

/// Lifecycle action buttons shown in the transaction app bar. The 1.53
/// API exposes a single lifecycle transition — reverse — so the action is
/// offered for any posting that is not itself a reversal. The backend
/// enforces the legality of the transition atomically, so a stale UI
/// cannot move a row through an illegal transition even under concurrency.
class _LifecycleActions extends ConsumerWidget {
  const _LifecycleActions({required this.transaction});

  final Transaction transaction;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final canReverse = transaction.type != TransactionType.REVERSAL;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (canReverse)
          IconButton(
            tooltip: 'Reverse this posting',
            icon: const Icon(Icons.undo),
            onPressed: () => _runWithConfirm(
              context,
              ref,
              title: 'Reverse transaction?',
              body:
                  'Creates an offsetting REVERSAL transaction. The original remains in the journal.',
              confirmLabel: 'Reverse',
              action: () => ref
                  .read(transactionNotifierProvider.notifier)
                  .reverse(ReverseTransactionRequest()..id = transaction.id),
            ),
          ),
      ],
    );
  }

  Future<void> _runWithConfirm(
    BuildContext context,
    WidgetRef ref, {
    required String title,
    required String body,
    required String confirmLabel,
    required Future<Transaction> Function() action,
  }) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(title),
        content: Text(body),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: Text(confirmLabel),
          ),
        ],
      ),
    );
    if (ok != true) return;
    try {
      await action();
      if (context.mounted) {
        ref.invalidate(transactionSearchProvider(transaction.id));
        ref.invalidate(transactionEntrySearchProvider(transaction.id));
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('$confirmLabel succeeded')));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('$confirmLabel failed: $e')));
      }
    }
  }
}
