import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';

/// Displays a colored badge for payment status.
class PaymentStatusBadge extends StatelessWidget {
  const PaymentStatusBadge({super.key, required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final normalized = status.toLowerCase().trim();
    final (label, color, icon) = _statusInfo(normalized);
    return StatusBadge(label: label, color: color, icon: icon);
  }

  static (String, Color, IconData?) _statusInfo(String status) {
    return switch (status) {
      'pending' => ('Pending', Colors.orange, Icons.schedule),
      'processing' => ('Processing', Colors.blue, Icons.sync),
      'completed' || 'success' => ('Completed', Colors.green, Icons.check_circle),
      'failed' || 'error' => ('Failed', Colors.red, Icons.error),
      'cancelled' || 'canceled' => ('Cancelled', Colors.grey, Icons.cancel),
      'reversed' => ('Reversed', Colors.purple, Icons.undo),
      'held' => ('Held', Colors.amber, Icons.pause_circle),
      _ => (status.isNotEmpty ? status : 'Unknown', Colors.grey, null),
    };
  }
}

/// Displays a colored badge for payment state.
class PaymentStateBadge extends StatelessWidget {
  const PaymentStateBadge({super.key, required this.state});

  final String state;

  @override
  Widget build(BuildContext context) {
    final normalized = this.state.toLowerCase().trim();
    final (label, color, icon) = _stateInfo(normalized);
    return StatusBadge(label: label, color: color, icon: icon);
  }

  static (String, Color, IconData?) _stateInfo(String state) {
    return switch (state) {
      'created' => ('Created', Colors.blue, null),
      'active' => ('Active', Colors.green, null),
      'inactive' => ('Inactive', Colors.grey, null),
      'deleted' => ('Deleted', Colors.red, null),
      _ => (state.isNotEmpty ? state : 'Unknown', Colors.grey, null),
    };
  }
}
