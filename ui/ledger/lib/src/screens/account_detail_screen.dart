import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/account_providers.dart';
import '../providers/transaction_providers.dart';
import '../widgets/account_balance_card.dart';
import '../widgets/transaction_entry_row.dart';

/// Screen showing details for a single account with balance card and
/// transaction history.
class AccountDetailScreen extends ConsumerWidget {
  const AccountDetailScreen({
    super.key,
    required this.accountId,
  });

  final String accountId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncAccounts = ref.watch(accountSearchProvider(accountId));

    return asyncAccounts.when(
      loading: () => Scaffold(
        appBar: _buildAppBar(context, theme, 'Account'),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Scaffold(
        appBar: _buildAppBar(context, theme, 'Account'),
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
                    ref.invalidate(accountSearchProvider(accountId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (accounts) {
        final account = accounts.isNotEmpty ? accounts.first : null;
        if (account == null) {
          return Scaffold(
            appBar: _buildAppBar(context, theme, 'Account'),
            body: Center(
              child: Text(
                'Account not found',
                style: theme.textTheme.bodyLarge,
              ),
            ),
          );
        }
        return _buildDetail(context, ref, theme, account);
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
            context.canPop() ? context.pop() : context.go('/ledger/accounts'),
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
    Account account,
  ) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.canPop()
              ? context.pop()
              : context.go('/ledger/accounts'),
        ),
        title: Text(
          'Account ${_truncate(account.id, 16)}',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () {
              ref.invalidate(accountSearchProvider(accountId));
              ref.invalidate(transactionEntrySearchProvider(accountId));
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // Balance card at top
          Padding(
            padding: const EdgeInsets.fromLTRB(24, 24, 24, 0),
            child: AccountBalanceCard(account: account),
          ),
          const SizedBox(height: 16),

          // Ledger link
          if (account.ledger.isNotEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24),
              child: Align(
                alignment: Alignment.centerLeft,
                child: InkWell(
                  onTap: () =>
                      context.go('/ledger/ledgers/${account.ledger}'),
                  borderRadius: BorderRadius.circular(4),
                  child: Text(
                    'View Ledger: ${account.ledger}',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.primary,
                      decoration: TextDecoration.underline,
                    ),
                  ),
                ),
              ),
            ),
          const SizedBox(height: 16),

          // Transaction history header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Align(
              alignment: Alignment.centerLeft,
              child: Text(
                'Transaction History',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: theme.colorScheme.primary,
                ),
              ),
            ),
          ),
          const SizedBox(height: 8),

          // Transaction entries list
          Expanded(
            child: _TransactionHistoryList(accountId: account.id),
          ),
        ],
      ),
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }
}

class _TransactionHistoryList extends ConsumerWidget {
  const _TransactionHistoryList({required this.accountId});

  final String accountId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncEntries = ref.watch(transactionEntrySearchProvider(accountId));

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
                  ref.invalidate(transactionEntrySearchProvider(accountId)),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
      data: (entries) {
        if (entries.isEmpty) {
          return Center(
            child: Text(
              'No transactions yet',
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
            final entry = entries[index];
            return TransactionEntryRow(
              entry: entry,
              onTap: () => context.go(
                '/ledger/transactions/${entry.transactionId}',
              ),
            );
          },
        );
      },
    );
  }
}
