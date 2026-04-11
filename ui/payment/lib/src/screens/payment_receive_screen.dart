import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/error_helpers.dart';
import 'package:antinvestor_ui_core/widgets/form_field_card.dart';
import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/payment_providers.dart';
import '../widgets/account_field.dart';
import '../widgets/payment_route_selector.dart';

/// Screen for receiving (requesting) a payment.
class PaymentReceiveScreen extends ConsumerStatefulWidget {
  const PaymentReceiveScreen({super.key});

  @override
  ConsumerState<PaymentReceiveScreen> createState() =>
      _PaymentReceiveScreenState();
}

class _PaymentReceiveScreenState extends ConsumerState<PaymentReceiveScreen> {
  final _formKey = GlobalKey<FormState>();
  bool _submitting = false;
  String? _error;

  // Form controllers
  final _amountController = TextEditingController();
  final _currencyController = TextEditingController(text: 'KES');
  String? _selectedRoute;

  // Source account (payer)
  final _sourceAccountController = TextEditingController();
  final _sourceCountryController = TextEditingController(text: 'KE');
  final _sourceNameController = TextEditingController();

  // Recipient account (payee / us)
  final _recipientAccountController = TextEditingController();
  final _recipientCountryController = TextEditingController(text: 'KE');
  final _recipientNameController = TextEditingController();

  final _referenceController = TextEditingController();

  @override
  void dispose() {
    _amountController.dispose();
    _currencyController.dispose();
    _sourceAccountController.dispose();
    _sourceCountryController.dispose();
    _sourceNameController.dispose();
    _recipientAccountController.dispose();
    _recipientCountryController.dispose();
    _recipientNameController.dispose();
    _referenceController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final mutationState = ref.watch(paymentNotifierProvider);

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () =>
              context.canPop() ? context.pop() : context.go('/payments'),
        ),
        title: Text(
          'Receive Payment',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Amount section
              FormSection(
                title: 'Amount',
                description: 'Enter the amount to receive.',
                children: [
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        flex: 3,
                        child: FormFieldCard(
                          label: 'Amount',
                          isRequired: true,
                          child: TextFormField(
                            controller: _amountController,
                            keyboardType: const TextInputType.numberWithOptions(
                              decimal: true,
                            ),
                            decoration: const InputDecoration(
                              hintText: '0.00',
                              prefixIcon:
                                  Icon(Icons.attach_money, size: 20),
                            ),
                            validator: validateAmount,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        flex: 1,
                        child: FormFieldCard(
                          label: 'Currency',
                          isRequired: true,
                          child: TextFormField(
                            controller: _currencyController,
                            textCapitalization: TextCapitalization.characters,
                            maxLength: 3,
                            decoration: const InputDecoration(
                              hintText: 'KES',
                              counterText: '',
                            ),
                            validator: validateCurrency,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 8),

              // Route section
              FormSection(
                title: 'Route',
                description: 'Select the payment channel.',
                children: [
                  FormFieldCard(
                    label: 'Payment Route',
                    isRequired: true,
                    child: PaymentRouteSelector(
                      value: _selectedRoute,
                      onChanged: (value) =>
                          setState(() => _selectedRoute = value),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),

              // Source account (payer)
              FormSection(
                title: 'Payer Account',
                description: 'The account that will send the payment.',
                children: [
                  AccountField(
                    label: 'Payer',
                    accountNumberController: _sourceAccountController,
                    countryCodeController: _sourceCountryController,
                    nameController: _sourceNameController,
                  ),
                ],
              ),
              const SizedBox(height: 8),

              // Recipient account (payee)
              FormSection(
                title: 'Payee Account',
                description: 'The account that will receive the payment.',
                children: [
                  AccountField(
                    label: 'Payee',
                    accountNumberController: _recipientAccountController,
                    countryCodeController: _recipientCountryController,
                    nameController: _recipientNameController,
                  ),
                ],
              ),
              const SizedBox(height: 8),

              // Reference
              FormFieldCard(
                label: 'Reference',
                description: 'Optional reference for this payment.',
                child: TextFormField(
                  controller: _referenceController,
                  decoration: const InputDecoration(
                    hintText: 'Payment reference...',
                    prefixIcon: Icon(Icons.tag, size: 20),
                  ),
                ),
              ),

              // Error display
              if (_error != null) ...[
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.errorContainer,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        Icons.error_outline,
                        size: 20,
                        color: theme.colorScheme.onErrorContainer,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _error!,
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onErrorContainer,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],

              const SizedBox(height: 24),

              // Submit button
              FilledButton.icon(
                onPressed:
                    _submitting || mutationState.isLoading ? null : _receive,
                icon: _submitting
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.call_received, size: 18),
                label: Text(_submitting ? 'Processing...' : 'Receive Payment'),
                style: FilledButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _receive() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _submitting = true;
      _error = null;
    });

    try {
      final amount = moneyFromString(
        _amountController.text,
        _currencyController.text.trim().toUpperCase(),
      );

      final source = ContactLink()
        ..detail = _sourceAccountController.text.trim()
        ..profileName = _sourceNameController.text.trim();

      final recipient = ContactLink()
        ..detail = _recipientAccountController.text.trim()
        ..profileName = _recipientNameController.text.trim();

      final payment = Payment()
        ..route = _selectedRoute ?? ''
        ..source = source
        ..recipient = recipient
        ..amount = amount
        ..referenceId = _referenceController.text.trim();

      final request = ReceiveRequest()..data = payment;

      final notifier = ref.read(paymentNotifierProvider.notifier);
      final statusResponse = await notifier.receive(request);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Payment received successfully'),
            behavior: SnackBarBehavior.floating,
          ),
        );
        context.go('/payments');
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _error = friendlyError(e);
        });
      }
    }
  }
}
