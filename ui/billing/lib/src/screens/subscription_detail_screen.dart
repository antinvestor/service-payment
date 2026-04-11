import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/subscription_providers.dart';
import '../providers/usage_providers.dart';
import '../widgets/subscription_state_badge.dart';
import '../widgets/usage_event_tile.dart';

/// Screen showing details for a single subscription: state, plan info,
/// period, and usage history.
class SubscriptionDetailScreen extends ConsumerWidget {
  const SubscriptionDetailScreen({
    super.key,
    required this.subscriptionId,
  });

  final String subscriptionId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncSubscription =
        ref.watch(subscriptionProvider(subscriptionId));

    return asyncSubscription.when(
      loading: () => Scaffold(
        appBar: _buildAppBar(context, theme, 'Subscription'),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Scaffold(
        appBar: _buildAppBar(context, theme, 'Subscription'),
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
                    ref.invalidate(subscriptionProvider(subscriptionId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (subscription) =>
          _buildDetail(context, ref, theme, subscription),
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
            : context.go('/billing/subscriptions'),
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
    Subscription subscription,
  ) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.canPop()
              ? context.pop()
              : context.go('/billing/subscriptions'),
        ),
        title: Text(
          'Subscription ${_truncate(subscription.id, 12)}',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () =>
                ref.invalidate(subscriptionProvider(subscriptionId)),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // State hero card
            Card(
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(color: theme.colorScheme.outlineVariant),
              ),
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    Container(
                      width: 56,
                      height: 56,
                      decoration: BoxDecoration(
                        color: Colors.green.withAlpha(25),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: const Icon(
                        Icons.subscriptions,
                        size: 28,
                        color: Colors.green,
                      ),
                    ),
                    const SizedBox(height: 16),
                    SubscriptionStateBadge(state: subscription.state),
                    const SizedBox(height: 16),
                    _metadataRow(theme, 'ID', subscription.id,
                        copyable: true, context: context),
                    _metadataRow(
                        theme, 'Plan', subscription.planId),
                    _metadataRow(
                        theme, 'Profile', subscription.profileId),
                    if (subscription.currency.isNotEmpty)
                      _metadataRow(
                          theme, 'Currency', subscription.currency),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Period card
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
                      'Billing Period',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 12),
                    if (subscription.hasStartAt())
                      _metadataRow(
                        theme,
                        'Start',
                        _formatTimestamp(subscription.startAt),
                      ),
                    if (subscription.hasEndAt())
                      _metadataRow(
                        theme,
                        'End',
                        _formatTimestamp(subscription.endAt),
                      ),
                    if (subscription.hasCancelledAt())
                      _metadataRow(
                        theme,
                        'Cancelled',
                        _formatTimestamp(subscription.cancelledAt),
                      ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Cancel action
            if (subscription.state == SubscriptionState.SUBSCRIPTION_ACTIVE)
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: () =>
                      _cancelSubscription(context, ref, subscription),
                  icon: const Icon(Icons.cancel, size: 18),
                  label: const Text('Cancel Subscription'),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: Colors.red,
                    side: const BorderSide(color: Colors.red),
                  ),
                ),
              ),
            const SizedBox(height: 24),

            // Usage history
            Text(
              'Usage History',
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w700,
                color: theme.colorScheme.primary,
              ),
            ),
            const SizedBox(height: 8),
            _UsageHistorySection(subscriptionId: subscription.id),
          ],
        ),
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

  String _formatTimestamp(dynamic ts) {
    try {
      final dt = ts.toDateTime();
      return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-'
          '${dt.day.toString().padLeft(2, '0')} '
          '${dt.hour.toString().padLeft(2, '0')}:'
          '${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return 'N/A';
    }
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  Future<void> _cancelSubscription(
    BuildContext context,
    WidgetRef ref,
    Subscription subscription,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Cancel Subscription'),
        content: Text(
          'Cancel subscription ${_truncate(subscription.id, 16)}?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Keep'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Cancel Subscription'),
          ),
        ],
      ),
    );

    if (confirmed != true || !context.mounted) return;

    try {
      final notifier = ref.read(subscriptionNotifierProvider.notifier);
      final request = CancelSubscriptionRequest()..id = subscription.id;
      await notifier.cancel(request);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Subscription cancelled'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        ref.invalidate(subscriptionProvider(subscription.id));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Cancel failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }
}

class _UsageHistorySection extends ConsumerWidget {
  const _UsageHistorySection({required this.subscriptionId});

  final String subscriptionId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncUsage =
        ref.watch(usageEventSearchProvider(subscriptionId));

    return asyncUsage.when(
      loading: () => const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 16),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(friendlyError(error)),
              const SizedBox(height: 8),
              FilledButton.tonal(
                onPressed: () => ref.invalidate(
                    usageEventSearchProvider(subscriptionId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (events) {
        if (events.isEmpty) {
          return Padding(
            padding: const EdgeInsets.symmetric(vertical: 24),
            child: Center(
              child: Text(
                'No usage events yet',
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            ),
          );
        }

        return Column(
          children: events.map((event) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: UsageEventTile(event: event),
            );
          }).toList(),
        );
      },
    );
  }
}
