import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:url_launcher/url_launcher.dart';

import '../providers/catalog_providers.dart';
import '../providers/collection_providers.dart';
import '../widgets/payment_method_picker.dart';

/// Streamlined "subscribe + pay" screen (Stripe/Flutterwave style entry).
///
/// Creates the subscription and opens hosted checkout when an upfront fee exists.
class StartSubscriptionScreen extends ConsumerStatefulWidget {
  const StartSubscriptionScreen({
    super.key,
    this.catalogVersionId,
    this.planId,
    this.currency,
  });

  final String? catalogVersionId;
  final String? planId;
  final String? currency;

  @override
  ConsumerState<StartSubscriptionScreen> createState() =>
      _StartSubscriptionScreenState();
}

class _StartSubscriptionScreenState
    extends ConsumerState<StartSubscriptionScreen> {
  final _profileController = TextEditingController();
  final _planController = TextEditingController();
  final _catalogController = TextEditingController();
  final _currencyController = TextEditingController(text: 'KES');
  final _nameController = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  Set<String> _methods = {};
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    if (widget.planId != null) {
      _planController.text = widget.planId!;
    }
    if (widget.catalogVersionId != null) {
      _catalogController.text = widget.catalogVersionId!;
    }
    if (widget.currency != null && widget.currency!.isNotEmpty) {
      _currencyController.text = widget.currency!;
    }
  }

  @override
  void dispose() {
    _profileController.dispose();
    _planController.dispose();
    _catalogController.dispose();
    _currencyController.dispose();
    _nameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.canPop()
              ? context.pop()
              : context.go('/billing/subscriptions'),
        ),
        title: Text(
          'Start subscription',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520),
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    'Customer pays on the hosted checkout page. '
                    'You will be redirected when a payment is required.',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 24),
                  TextFormField(
                    controller: _profileController,
                    decoration: const InputDecoration(
                      labelText: 'Customer profile ID',
                      border: OutlineInputBorder(),
                    ),
                    validator: (v) =>
                        (v == null || v.trim().isEmpty) ? 'Required' : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _catalogController,
                    decoration: const InputDecoration(
                      labelText: 'Catalog version ID',
                      border: OutlineInputBorder(),
                    ),
                    validator: (v) =>
                        (v == null || v.trim().isEmpty) ? 'Required' : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _planController,
                    decoration: const InputDecoration(
                      labelText: 'Plan ID',
                      border: OutlineInputBorder(),
                    ),
                    validator: (v) =>
                        (v == null || v.trim().isEmpty) ? 'Required' : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _currencyController,
                    decoration: const InputDecoration(
                      labelText: 'Currency (ISO 4217)',
                      border: OutlineInputBorder(),
                    ),
                    textCapitalization: TextCapitalization.characters,
                    validator: (v) {
                      if (v == null || v.trim().length != 3) {
                        return '3-letter currency code';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _nameController,
                    decoration: const InputDecoration(
                      labelText: 'Payer display name (optional)',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 24),
                  PaymentMethodPicker(
                    selected: _methods,
                    onChanged: (next) => setState(() => _methods = next),
                  ),
                  const SizedBox(height: 32),
                  FilledButton.icon(
                    onPressed: _submitting ? null : _submit,
                    icon: _submitting
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.payment),
                    label: Text(
                      _submitting ? 'Starting…' : 'Start & open checkout',
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _submitting = true);
    try {
      final notifier = ref.read(collectionNotifierProvider.notifier);
      final result = await notifier.startSubscription(
        profileId: _profileController.text.trim(),
        planId: _planController.text.trim(),
        catalogVersionId: _catalogController.text.trim(),
        currency: _currencyController.text.trim().toUpperCase(),
        payerDisplayName: _nameController.text.trim(),
        methods: _methods.toList(),
      );

      if (!mounted) return;

      if (result.alreadyComplete) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Subscription ${result.subscriptionId} is active (no payment required).',
            ),
            behavior: SnackBarBehavior.floating,
          ),
        );
        context.go('/billing/subscriptions/${result.subscriptionId}');
        return;
      }

      final uri = Uri.tryParse(result.pageUrl);
      if (uri == null || result.pageUrl.isEmpty) {
        throw Exception('Checkout page URL missing');
      }
      final launched = await launchUrl(
        uri,
        mode: LaunchMode.externalApplication,
      );
      if (!launched && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Open checkout: ${result.pageUrl}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Checkout opened. Session ${result.sessionRef}. '
              'Confirm payment after the customer finishes.',
            ),
            behavior: SnackBarBehavior.floating,
            action: SnackBarAction(
              label: 'Confirm',
              onPressed: () => _confirm(result.sessionRef),
            ),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Start failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  Future<void> _confirm(String sessionRef) async {
    try {
      final result = await ref
          .read(collectionNotifierProvider.notifier)
          .confirmPayment(sessionRef);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            result.paid
                ? 'Payment confirmed. Subscription ${result.subscriptionState}.'
                : 'Payment not complete yet.',
          ),
          behavior: SnackBarBehavior.floating,
        ),
      );
      if (result.subscriptionId.isNotEmpty) {
        context.go('/billing/subscriptions/${result.subscriptionId}');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Confirm failed: ${friendlyError(e)}'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }
}

/// Optional helper: seed plan fields from a loaded catalog when available.
void prefillFromCatalog(
  WidgetRef ref,
  String catalogId,
  void Function(CatalogVersion catalog) onCatalog,
) {
  ref.read(catalogVersionProvider(catalogId).future).then(onCatalog);
}
