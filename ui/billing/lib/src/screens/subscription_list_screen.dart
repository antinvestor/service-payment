import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/subscription_providers.dart';
import '../widgets/subscription_card.dart';
import '../widgets/subscription_state_badge.dart';

/// Screen that lists subscriptions with search and state filter.
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
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (subscriptions) {
        final filtered = _stateFilter == null
            ? subscriptions
            : subscriptions.where((s) => s.state == _stateFilter).toList();
        return _buildShell(items: filtered);
      },
    );
  }

  Widget _buildShell({
    required List<Subscription> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<Subscription>(
      title: 'Subscriptions',
      icon: Icons.subscriptions,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () =>
          ref.invalidate(subscriptionListProvider(_searchQuery)),
      searchHint: 'Search by profile ID...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      filterWidget: _buildStateFilter(),
      itemBuilder: (context, subscription) {
        return SubscriptionCard(
          subscription: subscription,
          onTap: () =>
              context.go('/billing/subscriptions/${subscription.id}'),
        );
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
}
