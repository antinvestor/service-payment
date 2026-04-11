import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:flutter/material.dart';

import 'payment_status_badge.dart';

/// A card widget displaying a payment summary: amount, route, source to
/// recipient, status, and direction (outbound/inbound).
class PaymentCard extends StatelessWidget {
  const PaymentCard({
    super.key,
    required this.payment,
    this.onTap,
  });

  final Payment payment;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isOutbound = payment.outbound;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: theme.colorScheme.outlineVariant),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Row(
            children: [
              // Direction icon
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: isOutbound
                      ? Colors.red.withAlpha(25)
                      : Colors.green.withAlpha(25),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(
                  isOutbound ? Icons.arrow_upward : Icons.arrow_downward,
                  size: 20,
                  color: isOutbound ? Colors.red : Colors.green,
                ),
              ),
              const SizedBox(width: 12),

              // Payment details
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Amount and route
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            formatMoney(payment.amount),
                            style: theme.textTheme.titleSmall?.copyWith(
                              fontWeight: FontWeight.w700,
                              color: isOutbound ? Colors.red : Colors.green,
                            ),
                          ),
                        ),
                        if (payment.route.isNotEmpty)
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.surfaceContainerLow,
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: Text(
                              payment.route,
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                            ),
                          ),
                      ],
                    ),
                    const SizedBox(height: 4),

                    // Source -> Recipient
                    Row(
                      children: [
                        Flexible(
                          child: Text(
                            _accountLabel(payment.source),
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 4),
                          child: Icon(
                            Icons.arrow_forward,
                            size: 12,
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                        Flexible(
                          child: Text(
                            _accountLabel(payment.recipient),
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),

                    // Cost and reference
                    Row(
                      children: [
                        if (payment.hasCost())
                          Text(
                            'Cost: ${formatMoney(payment.cost)}',
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                        const Spacer(),
                        if (payment.referenceId.isNotEmpty)
                          Text(
                            'Ref: ${_truncate(payment.referenceId, 12)}',
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),

              // Status badge and chevron
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  PaymentStatusBadge(status: payment.status),
                  const SizedBox(height: 4),
                  Icon(
                    Icons.chevron_right,
                    size: 20,
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _contactLabel(ContactLink contact) {
    if (contact.profileName.isNotEmpty) return contact.profileName;
    if (contact.detail.isNotEmpty) return contact.detail;
    if (contact.profileId.isNotEmpty) return contact.profileId;
    return 'Unknown';
  }

  String _truncate(String value, int maxLength) {
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }
}
