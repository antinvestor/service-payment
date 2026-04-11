import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/account_providers.dart';
import '../widgets/account_balance_card.dart';

/// Screen that lists accounts with search and balance display.
class AccountListScreen extends ConsumerStatefulWidget {
  const AccountListScreen({super.key});

  @override
  ConsumerState<AccountListScreen> createState() => _AccountListScreenState();
}

class _AccountListScreenState extends ConsumerState<AccountListScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncAccounts = ref.watch(accountSearchProvider(_searchQuery));

    return asyncAccounts.when(
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (accounts) => _buildShell(items: accounts),
    );
  }

  Widget _buildShell({
    required List<Account> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Account>(
      title: 'Accounts',
      icon: Icons.account_balance_wallet,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () => ref.invalidate(accountSearchProvider(_searchQuery)),
      searchHint: 'Search accounts...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      itemBuilder: (context, account) {
        return AccountBalanceCard(
          account: account,
          onTap: () => context.go('/ledger/accounts/${account.id}'),
        );
      },
    );
  }
}
