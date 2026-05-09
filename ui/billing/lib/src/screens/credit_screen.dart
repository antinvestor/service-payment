import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:antinvestor_ui_core/widgets/form_field_card.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/credit_providers.dart';

/// Screen for granting credit and viewing credit balance.
class CreditScreen extends ConsumerStatefulWidget {
  const CreditScreen({super.key});

  @override
  ConsumerState<CreditScreen> createState() => _CreditScreenState();
}

class _CreditScreenState extends ConsumerState<CreditScreen> {
  final _balanceProfileController = TextEditingController();
  final _profileIdController = TextEditingController();
  final _amountController = TextEditingController();
  final _currencyController = TextEditingController(text: 'KES');
  final _descriptionController = TextEditingController();

  String? _balanceProfileId;

  @override
  void dispose() {
    _balanceProfileController.dispose();
    _profileIdController.dispose();
    _amountController.dispose();
    _currencyController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              children: [
                Icon(Icons.account_balance_wallet_outlined,
                    size: 28, color: theme.colorScheme.primary),
                const SizedBox(width: 12),
                Text(
                  'Credits',
                  style: theme.textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                    letterSpacing: -0.3,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 24),

            // Check balance section
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
                      'Check Balance',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _balanceProfileController,
                            decoration: const InputDecoration(
                              hintText: 'Profile ID',
                              prefixIcon:
                                  Icon(Icons.person_outline, size: 20),
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        FilledButton.tonal(
                          onPressed: () {
                            final id =
                                _balanceProfileController.text.trim();
                            if (id.isNotEmpty) {
                              setState(() => _balanceProfileId = id);
                            }
                          },
                          child: const Text('Check'),
                        ),
                      ],
                    ),
                    if (_balanceProfileId != null &&
                        _balanceProfileId!.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      _CreditBalanceDisplay(profileId: _balanceProfileId!),
                    ],
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Grant credit section
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
                      'Grant Credit',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 16),
                    FormFieldCard(
                      label: 'Profile ID',
                      isRequired: true,
                      child: TextField(
                        controller: _profileIdController,
                        decoration: const InputDecoration(
                          hintText: 'Enter profile ID',
                          prefixIcon: Icon(Icons.person_outline, size: 20),
                        ),
                      ),
                    ),
                    FormFieldCard(
                      label: 'Amount',
                      isRequired: true,
                      child: TextField(
                        controller: _amountController,
                        keyboardType: const TextInputType.numberWithOptions(
                            decimal: true),
                        decoration: const InputDecoration(
                          hintText: '0.00',
                        ),
                      ),
                    ),
                    FormFieldCard(
                      label: 'Currency',
                      isRequired: true,
                      child: TextField(
                        controller: _currencyController,
                        decoration: const InputDecoration(
                          hintText: 'KES',
                        ),
                      ),
                    ),
                    FormFieldCard(
                      label: 'Description',
                      child: TextField(
                        controller: _descriptionController,
                        maxLines: 2,
                        decoration: const InputDecoration(
                          hintText: 'Reason for credit grant',
                        ),
                      ),
                    ),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        onPressed: _grantCredit,
                        icon: const Icon(Icons.add, size: 18),
                        label: const Text('Grant Credit'),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _grantCredit() async {
    final profileId = _profileIdController.text.trim();
    final amountText = _amountController.text.trim();
    final currency = _currencyController.text.trim();

    if (profileId.isEmpty || amountText.isEmpty || currency.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please fill in all required fields'),
          behavior: SnackBarBehavior.floating,
        ),
      );
      return;
    }

    final amount = double.tryParse(amountText);
    if (amount == null || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please enter a valid positive amount'),
          behavior: SnackBarBehavior.floating,
        ),
      );
      return;
    }

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Grant Credit'),
        content: Text(
          'Grant $currency $amountText credit to profile $profileId?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Grant'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    try {
      final notifier = ref.read(creditNotifierProvider.notifier);
      final request = GrantCreditRequest()
        ..profileId = profileId
        ..currency = currency
        ..name = _descriptionController.text.trim();
      setMoneyFields(
        request.ensureAmount(),
        amountText,
        currency.toUpperCase(),
      );
      await notifier.grantCredit(request);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Credit granted successfully'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        _amountController.clear();
        _descriptionController.clear();
        // Refresh balance if viewing this profile
        if (_balanceProfileId == profileId) {
          ref.invalidate(creditBalanceProvider(profileId));
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Grant failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }
}

class _CreditBalanceDisplay extends ConsumerWidget {
  const _CreditBalanceDisplay({required this.profileId});

  final String profileId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final asyncBalance = ref.watch(creditBalanceProvider(profileId));

    return asyncBalance.when(
      loading: () => const Padding(
        padding: EdgeInsets.symmetric(vertical: 16),
        child: Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              friendlyError(error),
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.error,
              ),
            ),
            const SizedBox(height: 8),
            FilledButton.tonal(
              onPressed: () =>
                  ref.invalidate(creditBalanceProvider(profileId)),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
      data: (response) {
        return Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.green.withAlpha(15),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            children: [
              Text(
                'Available Credit',
                style: theme.textTheme.labelMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                formatMoney(response.balance),
                style: theme.textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: Colors.green,
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
