import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/account_providers.dart';

/// Screen that lists accounts with search and balance display using DataTable.
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
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (accounts) => _buildTable(items: accounts),
    );
  }

  Widget _buildTable({required List<Account> items}) {
    return AdminEntityListPage<Account>(
      title: 'Accounts',
      breadcrumbs: const ['Ledger', 'Accounts'],
      searchHint: 'Search accounts...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      columns: const [
        DataColumn(label: Text('ID')),
        DataColumn(label: Text('Ledger')),
        DataColumn(label: Text('Balance'), numeric: true),
        DataColumn(label: Text('Uncleared'), numeric: true),
        DataColumn(label: Text('Reserved'), numeric: true),
      ],
      items: items,
      rowBuilder: (account, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(_truncate(account.id, 16))),
          DataCell(Text(_truncate(account.ledger, 16))),
          DataCell(Text(formatMoney(account.balance))),
          DataCell(Text(formatMoney(account.unclearedBalance))),
          DataCell(Text(formatMoney(account.reservedBalance))),
        ],
      ),
      onRowNavigate: (account) => context.go('/ledger/accounts/${account.id}'),
      exportRow: (account) => [
        account.id,
        account.ledger,
        formatMoney(account.balance),
        formatMoney(account.unclearedBalance),
        formatMoney(account.reservedBalance),
      ],
      onExport: (format, rowCount) {
        debugPrint('AUDIT: Exported $rowCount accounts as $format');
      },
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }
}
