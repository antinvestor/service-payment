import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';

/// Displays a colored badge for billing run state.
class BillingRunStateBadge extends StatelessWidget {
  const BillingRunStateBadge({super.key, required this.state});

  final BillingRunState state;

  @override
  Widget build(BuildContext context) {
    final (label, color, icon) = _stateInfo(state);
    return StatusBadge(label: label, color: color, icon: icon);
  }

  static (String, Color, IconData?) _stateInfo(BillingRunState state) {
    return switch (state) {
      BillingRunState.BILLING_RUN_PENDING =>
        ('Pending', Colors.grey, Icons.schedule),
      BillingRunState.BILLING_RUN_METERING =>
        ('Metering', Colors.blue, Icons.speed),
      BillingRunState.BILLING_RUN_RATING =>
        ('Rating', Colors.indigo, Icons.calculate),
      BillingRunState.BILLING_RUN_DISCOUNTING =>
        ('Discounting', Colors.purple, Icons.discount),
      BillingRunState.BILLING_RUN_CREDITING =>
        ('Crediting', Colors.teal, Icons.account_balance_wallet),
      BillingRunState.BILLING_RUN_INVOICING =>
        ('Invoicing', Colors.orange, Icons.receipt_long),
      BillingRunState.BILLING_RUN_POSTING =>
        ('Posting', Colors.amber, Icons.publish),
      BillingRunState.BILLING_RUN_COMPLETED =>
        ('Completed', Colors.green, Icons.check_circle),
      BillingRunState.BILLING_RUN_FAILED =>
        ('Failed', Colors.red, Icons.error),
      _ => ('Unknown', Colors.grey, null),
    };
  }
}

/// Returns a human-readable label for billing run state.
String billingRunStateLabel(BillingRunState state) {
  return switch (state) {
    BillingRunState.BILLING_RUN_PENDING => 'Pending',
    BillingRunState.BILLING_RUN_METERING => 'Metering',
    BillingRunState.BILLING_RUN_RATING => 'Rating',
    BillingRunState.BILLING_RUN_DISCOUNTING => 'Discounting',
    BillingRunState.BILLING_RUN_CREDITING => 'Crediting',
    BillingRunState.BILLING_RUN_INVOICING => 'Invoicing',
    BillingRunState.BILLING_RUN_POSTING => 'Posting',
    BillingRunState.BILLING_RUN_COMPLETED => 'Completed',
    BillingRunState.BILLING_RUN_FAILED => 'Failed',
    _ => 'Unknown',
  };
}

/// The ordered pipeline stages for display.
const billingRunPipelineStages = [
  BillingRunState.BILLING_RUN_PENDING,
  BillingRunState.BILLING_RUN_METERING,
  BillingRunState.BILLING_RUN_RATING,
  BillingRunState.BILLING_RUN_DISCOUNTING,
  BillingRunState.BILLING_RUN_CREDITING,
  BillingRunState.BILLING_RUN_INVOICING,
  BillingRunState.BILLING_RUN_POSTING,
  BillingRunState.BILLING_RUN_COMPLETED,
];
