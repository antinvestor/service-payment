import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/account_providers.dart';
import '../providers/ledger_providers.dart';
import '../widgets/account_balance_card.dart';
import '../widgets/ledger_type_badge.dart';

/// Screen showing details for a single ledger with Info and Accounts tabs.
class LedgerDetailScreen extends ConsumerStatefulWidget {
  const LedgerDetailScreen({
    super.key,
    required this.ledgerId,
  });

  final String ledgerId;

  @override
  ConsumerState<LedgerDetailScreen> createState() =>
      _LedgerDetailScreenState();
}

class _LedgerDetailScreenState extends ConsumerState<LedgerDetailScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final asyncLedgers = ref.watch(ledgerSearchProvider(widget.ledgerId));

    return asyncLedgers.when(
      loading: () => Scaffold(
        appBar: _buildAppBar(context, theme, 'Ledger'),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Scaffold(
        appBar: _buildAppBar(context, theme, 'Ledger'),
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
                    ref.invalidate(ledgerSearchProvider(widget.ledgerId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (ledgers) {
        final ledger = ledgers.isNotEmpty ? ledgers.first : null;
        if (ledger == null) {
          return Scaffold(
            appBar: _buildAppBar(context, theme, 'Ledger'),
            body: Center(
              child: Text(
                'Ledger not found',
                style: theme.textTheme.bodyLarge,
              ),
            ),
          );
        }
        return _buildDetail(context, theme, ledger);
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
            context.canPop() ? context.pop() : context.go('/ledger/ledgers'),
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
    ThemeData theme,
    Ledger ledger,
  ) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () =>
              context.canPop() ? context.pop() : context.go('/ledger/ledgers'),
        ),
        title: Text(
          'Ledger ${_truncate(ledger.id, 16)}',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () =>
                ref.invalidate(ledgerSearchProvider(widget.ledgerId)),
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Info'),
            Tab(text: 'Accounts'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _InfoTab(ledger: ledger),
          _AccountsTab(ledgerId: ledger.id),
        ],
      ),
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }
}

class _InfoTab extends StatelessWidget {
  const _InfoTab({required this.ledger});

  final Ledger ledger;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Type card
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
                    'Ledger Details',
                    style: theme.textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: theme.colorScheme.primary,
                    ),
                  ),
                  const SizedBox(height: 12),
                  _metadataRow(theme, 'ID', ledger.id,
                      copyable: true, context: context),
                  Row(
                    children: [
                      SizedBox(
                        width: 130,
                        child: Text(
                          'Type',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                      LedgerTypeBadge(type: ledger.type),
                    ],
                  ),
                  const SizedBox(height: 8),
                  if (ledger.parent.isNotEmpty) ...[
                    _parentRow(context, theme, ledger.parent),
                  ],
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Data / metadata
          if (ledger.data.fields.isNotEmpty) ...[
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
                      'Metadata',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 12),
                    ...ledger.data.fields.entries.map(
                      (e) => _metadataRow(theme, e.key, e.value.stringValue),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _parentRow(
    BuildContext context,
    ThemeData theme,
    String parentId,
  ) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 130,
            child: Text(
              'Parent Ledger',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          Expanded(
            child: InkWell(
              onTap: () => context.go('/ledger/ledgers/$parentId'),
              borderRadius: BorderRadius.circular(4),
              child: Text(
                parentId,
                style: theme.textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w500,
                  color: theme.colorScheme.primary,
                  decoration: TextDecoration.underline,
                ),
              ),
            ),
          ),
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
}

class _AccountsTab extends ConsumerWidget {
  const _AccountsTab({required this.ledgerId});

  final String ledgerId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncAccounts = ref.watch(accountSearchProvider(ledgerId));

    return asyncAccounts.when(
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
                  ref.invalidate(accountSearchProvider(ledgerId)),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
      data: (accounts) {
        if (accounts.isEmpty) {
          return Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 64,
                  height: 64,
                  decoration: BoxDecoration(
                    color: theme.colorScheme.surfaceContainerLow,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Icon(
                    Icons.account_balance_wallet,
                    size: 28,
                    color: theme.colorScheme.primary.withAlpha(160),
                  ),
                ),
                const SizedBox(height: 16),
                Text(
                  'No accounts in this ledger',
                  style: theme.textTheme.titleMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          );
        }

        return ListView.separated(
          padding: const EdgeInsets.all(24),
          itemCount: accounts.length,
          separatorBuilder: (_, _) => const SizedBox(height: 8),
          itemBuilder: (context, index) {
            final account = accounts[index];
            return AccountBalanceCard(
              account: account,
              onTap: () => context.go('/ledger/accounts/${account.id}'),
            );
          },
        );
      },
    );
  }
}
