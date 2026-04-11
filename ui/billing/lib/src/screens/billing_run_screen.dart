import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:antinvestor_ui_core/widgets/form_field_card.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/billing_run_providers.dart';
import '../widgets/billing_run_state_badge.dart';

/// Screen for running billing and viewing the billing run state pipeline.
class BillingRunScreen extends ConsumerStatefulWidget {
  const BillingRunScreen({super.key});

  @override
  ConsumerState<BillingRunScreen> createState() => _BillingRunScreenState();
}

class _BillingRunScreenState extends ConsumerState<BillingRunScreen> {
  final _billingRunIdController = TextEditingController();
  final _subscriptionIdController = TextEditingController();
  String? _activeBillingRunId;

  @override
  void dispose() {
    _billingRunIdController.dispose();
    _subscriptionIdController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Billing Runs',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Run billing section
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
                      'Run Billing',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 16),
                    FormFieldCard(
                      label: 'Subscription ID',
                      isRequired: true,
                      description:
                          'The subscription to bill.',
                      child: TextField(
                        controller: _subscriptionIdController,
                        decoration: const InputDecoration(
                          hintText: 'Enter subscription ID',
                          prefixIcon:
                              Icon(Icons.subscriptions, size: 20),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        onPressed: _runBilling,
                        icon: const Icon(Icons.play_arrow, size: 20),
                        label: const Text('Start Billing Run'),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // View billing run section
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
                      'View Billing Run',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 16),
                    FormFieldCard(
                      label: 'Billing Run ID',
                      description: 'Enter a billing run ID to view its state.',
                      child: Row(
                        children: [
                          Expanded(
                            child: TextField(
                              controller: _billingRunIdController,
                              decoration: const InputDecoration(
                                hintText: 'Enter billing run ID',
                              ),
                            ),
                          ),
                          const SizedBox(width: 12),
                          FilledButton.tonal(
                            onPressed: () {
                              final id =
                                  _billingRunIdController.text.trim();
                              if (id.isNotEmpty) {
                                setState(() => _activeBillingRunId = id);
                              }
                            },
                            child: const Text('Load'),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Pipeline view
            if (_activeBillingRunId != null)
              _BillingRunPipeline(billingRunId: _activeBillingRunId!),
          ],
        ),
      ),
    );
  }

  Future<void> _runBilling() async {
    final subscriptionId = _subscriptionIdController.text.trim();
    if (subscriptionId.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please enter a subscription ID'),
          behavior: SnackBarBehavior.floating,
        ),
      );
      return;
    }

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Start Billing Run'),
        content: Text(
          'This will start a new billing run for subscription $subscriptionId. Continue?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Start'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    try {
      final notifier = ref.read(billingRunNotifierProvider.notifier);
      final request = RunBillingRequest()
        ..subscriptionId = subscriptionId;
      final billingRun = await notifier.runBilling(request);
      if (mounted) {
        setState(() => _activeBillingRunId = billingRun.id);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Billing run started'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to start: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }
}

class _BillingRunPipeline extends ConsumerWidget {
  const _BillingRunPipeline({required this.billingRunId});

  final String billingRunId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncRun = ref.watch(billingRunProvider(billingRunId));

    return asyncRun.when(
      loading: () => const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: CircularProgressIndicator(),
        ),
      ),
      error: (error, _) => Card(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: BorderSide(color: theme.colorScheme.error),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              Text(friendlyError(error)),
              const SizedBox(height: 8),
              FilledButton.tonal(
                onPressed: () =>
                    ref.invalidate(billingRunProvider(billingRunId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (billingRun) => _buildPipeline(context, ref, theme, billingRun),
    );
  }

  Widget _buildPipeline(
    BuildContext context,
    WidgetRef ref,
    ThemeData theme,
    BillingRun billingRun,
  ) {
    final currentIndex =
        billingRunPipelineStages.indexOf(billingRun.state);
    final isFailed = billingRun.state == BillingRunState.BILLING_RUN_FAILED;

    return Card(
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
            Row(
              children: [
                Text(
                  'Pipeline Status',
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: theme.colorScheme.primary,
                  ),
                ),
                const Spacer(),
                BillingRunStateBadge(state: billingRun.state),
                const SizedBox(width: 8),
                IconButton(
                  icon: const Icon(Icons.refresh, size: 20),
                  tooltip: 'Refresh',
                  onPressed: () =>
                      ref.invalidate(billingRunProvider(billingRunId)),
                ),
              ],
            ),
            const SizedBox(height: 16),

            // Pipeline stages
            ...List.generate(billingRunPipelineStages.length, (index) {
              final stage = billingRunPipelineStages[index];
              final isCompleted = currentIndex > index;
              final isCurrent = currentIndex == index;
              final isPending = currentIndex < index;

              Color stageColor;
              IconData stageIcon;
              if (isFailed && isCurrent) {
                stageColor = Colors.red;
                stageIcon = Icons.error;
              } else if (isCompleted) {
                stageColor = Colors.green;
                stageIcon = Icons.check_circle;
              } else if (isCurrent) {
                stageColor = Colors.blue;
                stageIcon = Icons.radio_button_checked;
              } else {
                stageColor = theme.colorScheme.outlineVariant;
                stageIcon = Icons.radio_button_unchecked;
              }

              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Row(
                  children: [
                    Icon(stageIcon, size: 20, color: stageColor),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        billingRunStateLabel(stage),
                        style: theme.textTheme.bodyMedium?.copyWith(
                          fontWeight:
                              isCurrent ? FontWeight.w700 : FontWeight.w500,
                          color: isPending
                              ? theme.colorScheme.onSurfaceVariant
                              : null,
                        ),
                      ),
                    ),
                    if (isCompleted)
                      Icon(
                        Icons.done,
                        size: 16,
                        color: Colors.green,
                      ),
                  ],
                ),
              );
            }),
          ],
        ),
      ),
    );
  }
}
