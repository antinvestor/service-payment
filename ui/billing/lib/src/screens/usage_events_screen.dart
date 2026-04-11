import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/entity_list_page.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/usage_providers.dart';
import '../widgets/usage_event_tile.dart';

/// Screen that lists usage events with search.
class UsageEventsScreen extends ConsumerStatefulWidget {
  const UsageEventsScreen({super.key});

  @override
  ConsumerState<UsageEventsScreen> createState() => _UsageEventsScreenState();
}

class _UsageEventsScreenState extends ConsumerState<UsageEventsScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final asyncEvents =
        ref.watch(usageEventSearchProvider(_searchQuery));

    return asyncEvents.when(
      loading: () => _buildShell(isLoading: true, items: const []),
      error: (error, _) => _buildShell(
        error: friendlyError(error),
        items: const [],
      ),
      data: (events) => _buildShell(items: events),
    );
  }

  Widget _buildShell({
    required List<UsageEvent> items,
    bool isLoading = false,
    String? error,
  }) {
    return EntityListPage<UsageEvent>(
      title: 'Usage Events',
      icon: Icons.data_usage,
      items: items,
      isLoading: isLoading,
      error: error,
      onRetry: () =>
          ref.invalidate(usageEventSearchProvider(_searchQuery)),
      searchHint: 'Search usage events...',
      onSearchChanged: (query) {
        setState(() => _searchQuery = query.trim());
      },
      itemBuilder: (context, event) {
        return UsageEventTile(event: event);
      },
    );
  }
}
