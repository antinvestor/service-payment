import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/api/api_base.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const _billingUrl = String.fromEnvironment(
  'BILLING_URL',
  defaultValue: 'https://api.antinvestor.com/billing',
);

final billingTransportProvider = Provider<Transport>((ref) {
  final tokenProvider = ref.watch(authTokenProviderProvider);
  return createTransport(tokenProvider, baseUrl: _billingUrl);
});

final billingServiceClientProvider = Provider<BillingServiceClient>((ref) {
  final transport = ref.watch(billingTransportProvider);
  return BillingServiceClient(transport);
});
