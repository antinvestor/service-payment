import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/transaction_providers.dart';
import '../widgets/transaction_type_badge.dart';

/// Screen that lists transactions with search and type badges.
class TransactionListScreen extends ConsumerStatefulWidget {
  const TransactionListScreen({super.key});

  @override
  ConsumerState<TransactionListScreen> createState() =>
      _TransactionListScreenState();
}

class _TransactionListScreenState
    extends ConsumerState<TransactionListScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncTransactions =
        ref.watch(transactionSearchProvider(_searchQuery));

    return asyncTransactions.when(
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (transactions) => _buildShell(items: transactions),
    );
  }

  Widget _buildShell({
    required List<Transaction> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Transaction>(
      title: 'Transactions',
      icon: Icons.receipt_long,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () =>
          ref.invalidate(transactionSearchProvider(_searchQuery)),
      searchHint: 'Search transactions...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      itemBuilder: (context, transaction) {
        return _TransactionListTile(
          transaction: transaction,
          onTap: () =>
              context.go('/ledger/transactions/${transaction.id}'),
        );
      },
    );
  }
}

class _TransactionListTile extends StatelessWidget {
  const _TransactionListTile({
    required this.transaction,
    this.onTap,
  });

  final Transaction transaction;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: theme.colorScheme.outlineVariant),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Row(
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: theme.colorScheme.primaryContainer.withAlpha(80),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(
                  Icons.receipt_long,
                  size: 20,
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      _truncate(transaction.id, 24),
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        if (transaction.currencyCode.isNotEmpty) ...[
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.surfaceContainerLow,
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: Text(
                              transaction.currencyCode,
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                            ),
                          ),
                          const SizedBox(width: 8),
                        ],
                        Text(
                          '${transaction.entries.length} '
                          'entr${transaction.entries.length == 1 ? 'y' : 'ies'}',
                          style: theme.textTheme.bodySmall?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: theme.colorScheme.primary,
                          ),
                        ),
                        const Spacer(),
                        if (transaction.transactedAt.isNotEmpty)
                          Text(
                            _formatDateString(transaction.transactedAt),
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  TransactionTypeBadge(type: transaction.type),
                  const SizedBox(height: 4),
                  Icon(
                    Icons.chevron_right,
                    size: 20,
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  String _formatDateString(String rfc3339) {
    try {
      final dt = DateTime.parse(rfc3339);
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')}';
    } catch (_) {
      return rfc3339.length > 10 ? rfc3339.substring(0, 10) : rfc3339;
    }
  }
}
