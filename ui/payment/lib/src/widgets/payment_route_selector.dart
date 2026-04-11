import 'package:flutter/material.dart';

/// A dropdown selector for payment routes (e.g. MPESA, BANK, CARD).
class PaymentRouteSelector extends StatelessWidget {
  const PaymentRouteSelector({
    super.key,
    required this.value,
    required this.onChanged,
    this.routes = _defaultRoutes,
    this.label = 'Payment Route',
  });

  final String? value;
  final ValueChanged<String?> onChanged;
  final List<String> routes;
  final String label;

  static const _defaultRoutes = [
    'MPESA',
    'BANK',
    'CARD',
    'AIRTEL',
    'PAYPAL',
    'INTERNAL',
  ];

  IconData _routeIcon(String route) {
    return switch (route.toUpperCase()) {
      'MPESA' || 'AIRTEL' => Icons.phone_android,
      'BANK' => Icons.account_balance,
      'CARD' => Icons.credit_card,
      'PAYPAL' => Icons.payments,
      'INTERNAL' => Icons.swap_horiz,
      _ => Icons.route,
    };
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return DropdownButtonFormField<String>(
      value: value,
      decoration: InputDecoration(
        labelText: label,
        prefixIcon: Icon(
          value != null ? _routeIcon(value!) : Icons.route,
          size: 20,
        ),
      ),
      items: routes.map((route) {
        return DropdownMenuItem<String>(
          value: route,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(_routeIcon(route), size: 18),
              const SizedBox(width: 8),
              Text(route),
            ],
          ),
        );
      }).toList(),
      onChanged: onChanged,
      validator: (value) {
        if (value == null || value.isEmpty) return 'Route is required';
        return null;
      },
    );
  }
}
