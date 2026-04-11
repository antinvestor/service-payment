import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/widgets/status_badge.dart';
import 'package:flutter/material.dart';

/// Displays a colored badge for pricing model type.
class PricingModelBadge extends StatelessWidget {
  const PricingModelBadge({super.key, required this.model});

  final PricingModel model;

  @override
  Widget build(BuildContext context) {
    final (label, color, icon) = _modelInfo(model);
    return StatusBadge(label: label, color: color, icon: icon);
  }

  static (String, Color, IconData?) _modelInfo(PricingModel model) {
    return switch (model) {
      PricingModel.FLAT =>
        ('Flat', Colors.blue, Icons.horizontal_rule),
      PricingModel.PER_UNIT =>
        ('Per Unit', Colors.teal, Icons.straighten),
      PricingModel.TIERED =>
        ('Tiered', Colors.purple, Icons.stacked_bar_chart),
      PricingModel.VOLUME =>
        ('Volume', Colors.indigo, Icons.bar_chart),
      PricingModel.STAIRSTEP =>
        ('Stairstep', Colors.orange, Icons.stairs),
      _ => ('Unknown', Colors.grey, null),
    };
  }
}

/// Returns a human-readable label for pricing model.
String pricingModelLabel(PricingModel model) {
  return switch (model) {
    PricingModel.FLAT => 'Flat',
    PricingModel.PER_UNIT => 'Per Unit',
    PricingModel.TIERED => 'Tiered',
    PricingModel.VOLUME => 'Volume',
    PricingModel.STAIRSTEP => 'Stairstep',
    _ => 'Unknown',
  };
}
