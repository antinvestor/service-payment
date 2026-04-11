import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';

/// Displays a colored badge for subscription state.
class SubscriptionStateBadge extends StatelessWidget {
  const SubscriptionStateBadge({super.key, required this.state});

  final SubscriptionState state;

  @override
  Widget build(BuildContext context) {
    final (label, color, icon) = _stateInfo(state);
    return StatusBadge(label: label, color: color, icon: icon);
  }

  static (String, Color, IconData?) _stateInfo(SubscriptionState state) {
    return switch (state) {
      SubscriptionState.SUBSCRIPTION_ACTIVE =>
        ('Active', Colors.green, Icons.check_circle),
      SubscriptionState.SUBSCRIPTION_CANCELLED =>
        ('Cancelled', Colors.grey, Icons.cancel),
      SubscriptionState.SUBSCRIPTION_EXPIRED =>
        ('Expired', Colors.red, Icons.timer_off),
      SubscriptionState.SUBSCRIPTION_PENDING =>
        ('Pending', Colors.amber, Icons.schedule),
      _ => ('Unknown', Colors.grey, null),
    };
  }
}

/// Returns a human-readable label for subscription state.
String subscriptionStateLabel(SubscriptionState state) {
  return switch (state) {
    SubscriptionState.SUBSCRIPTION_ACTIVE => 'Active',
    SubscriptionState.SUBSCRIPTION_CANCELLED => 'Cancelled',
    SubscriptionState.SUBSCRIPTION_EXPIRED => 'Expired',
    SubscriptionState.SUBSCRIPTION_PENDING => 'Pending',
    _ => 'Unknown',
  };
}
