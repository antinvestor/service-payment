import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/ledger_providers.dart';
import '../widgets/ledger_type_badge.dart';

/// Screen that lists ledgers with search and LedgerType filter.
class LedgerListScreen extends ConsumerStatefulWidget {
  const LedgerListScreen({super.key});

  @override
  ConsumerState<LedgerListScreen> createState() => _LedgerListScreenState();
}

class _LedgerListScreenState extends ConsumerState<LedgerListScreen> {
  String _searchQuery = '';
  LedgerType? _typeFilter;

  @override
  Widget build(BuildContext context) {
    final asyncLedgers = ref.watch(ledgerSearchProvider(_searchQuery));

    return asyncLedgers.when(
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (ledgers) {
        final filtered = _typeFilter == null
            ? ledgers
            : ledgers.where((l) => l.type == _typeFilter).toList();
        return _buildShell(items: filtered);
      },
    );
  }

  Widget _buildShell({
    required List<Ledger> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Ledger>(
      title: 'Ledgers',
      icon: Icons.account_tree,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () => ref.invalidate(ledgerSearchProvider(_searchQuery)),
      searchHint: 'Search ledgers...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      filterWidget: _buildTypeFilter(),
      itemBuilder: (context, ledger) {
        return _LedgerListTile(
          ledger: ledger,
          onTap: () => context.go('/ledger/ledgers/${ledger.id}'),
        );
      },
    );
  }

  Widget _buildTypeFilter() {
    return DropdownButton<LedgerType?>(
      value: _typeFilter,
      hint: const Text('All Types'),
      underline: const SizedBox.shrink(),
      borderRadius: BorderRadius.circular(8),
      items: [
        const DropdownMenuItem<LedgerType?>(
          value: null,
          child: Text('All Types'),
        ),
        ...LedgerType.values
            .map((type) => DropdownMenuItem<LedgerType?>(
                  value: type,
                  child: Text(ledgerTypeLabel(type)),
                )),
      ],
      onChanged: (value) {
        setState(() => _typeFilter = value);
      },
    );
  }
}

class _LedgerListTile extends StatelessWidget {
  const _LedgerListTile({
    required this.ledger,
    this.onTap,
  });

  final Ledger ledger;
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
                  color: ledgerTypeColor(ledger.type).withAlpha(25),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(
                  Icons.account_tree,
                  size: 20,
                  color: ledgerTypeColor(ledger.type),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      ledger.id,
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    if (ledger.parent.isNotEmpty)
                      Text(
                        'Parent: ${ledger.parent}',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              LedgerTypeBadge(type: ledger.type),
              const SizedBox(width: 4),
              Icon(
                Icons.chevron_right,
                size: 20,
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
