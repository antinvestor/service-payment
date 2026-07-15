import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/subscription_providers.dart';
import '../widgets/subscription_state_badge.dart';

/// Screen that lists subscriptions with search and state filter using DataTable.
class SubscriptionListScreen extends ConsumerStatefulWidget {
  const SubscriptionListScreen({super.key});

  @override
  ConsumerState<SubscriptionListScreen> createState() =>
      _SubscriptionListScreenState();
}

class _SubscriptionListScreenState
    extends ConsumerState<SubscriptionListScreen> {
  String _searchQuery = '';
  SubscriptionState? _stateFilter;

  @override
  Widget build(BuildContext context) {
    final asyncSubscriptions =
        ref.watch(subscriptionListProvider(_searchQuery));

    return asyncSubscriptions.when(
      loading: () => _buildTable(items: const []),
      error: (error, _) => _buildTable(items: const []),
      data: (subscriptions) {
        final filtered = _stateFilter == null
            ? subscriptions
            : subscriptions.where((s) => s.state == _stateFilter).toList();
        return _buildTable(items: filtered);
      },
    );
  }

  Widget _buildTable({required List<Subscription> items}) {
    return AdminEntityListPage<Subscription>(
      title: 'Subscriptions',
      breadcrumbs: const ['Billing', 'Subscriptions'],
      searchHint: 'Search by profile ID...',
      onSearch: (query) {
        setState(() => _searchQuery = query.trim());
      },
      actions: [
        FilledButton.icon(
          onPressed: () => context.go('/billing/subscriptions/start'),
          icon: const Icon(Icons.add, size: 18),
          label: const Text('Start subscription'),
        ),
        _buildStateFilter(),
      ],
      columns: const [
        DataColumn(label: Text('ID')),
        DataColumn(label: Text('Profile')),
        DataColumn(label: Text('Plan')),
        DataColumn(label: Text('State')),
        DataColumn(label: Text('Start')),
        DataColumn(label: Text('End')),
        DataColumn(label: Text('Currency')),
      ],
      items: items,
      rowBuilder: (sub, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(_truncate(sub.id, 12))),
          DataCell(Text(_truncate(sub.profileId, 12))),
          DataCell(Text(_truncate(sub.planId, 12))),
          DataCell(SubscriptionStateBadge(state: sub.state)),
          DataCell(Text(_formatTimestamp(sub.startAt))),
          DataCell(Text(_formatTimestamp(sub.endAt))),
          DataCell(Text(sub.currency)),
        ],
      ),
      onRowNavigate: (sub) =>
          context.go('/billing/subscriptions/${sub.id}'),
      exportRow: (sub) => [
        sub.id,
        sub.profileId,
        sub.planId,
        subscriptionStateLabel(sub.state),
        _formatTimestamp(sub.startAt),
        _formatTimestamp(sub.endAt),
        sub.currency,
      ],
      onExport: (format, rowCount) {
        debugPrint(
            'AUDIT: Exported $rowCount subscriptions as $format');
      },
    );
  }

  Widget _buildStateFilter() {
    return DropdownButton<SubscriptionState?>(
      value: _stateFilter,
      hint: const Text('All States'),
      underline: const SizedBox.shrink(),
      borderRadius: BorderRadius.circular(8),
      items: [
        const DropdownMenuItem<SubscriptionState?>(
          value: null,
          child: Text('All States'),
        ),
        ...SubscriptionState.values
            .map((state) => DropdownMenuItem<SubscriptionState?>(
                  value: state,
                  child: Text(subscriptionStateLabel(state)),
                )),
      ],
      onChanged: (value) {
        setState(() => _stateFilter = value);
      },
    );
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  String _formatTimestamp(dynamic ts) {
    try {
      final seconds = ts.seconds.toInt();
      if (seconds == 0) return '';
      final dt = DateTime.fromMillisecondsSinceEpoch(seconds * 1000);
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }
}
