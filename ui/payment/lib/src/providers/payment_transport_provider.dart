import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/api/api_base.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const _paymentUrl = String.fromEnvironment(
  'PAYMENT_URL',
  defaultValue: 'https://api.antinvestor.com/payment',
);

final paymentTransportProvider = Provider<Transport>((ref) {
  final tokenProvider = ref.watch(authTokenProviderProvider);
  return createTransport(tokenProvider, baseUrl: _paymentUrl);
});

final paymentServiceClientProvider = Provider<PaymentServiceClient>((ref) {
  final transport = ref.watch(paymentTransportProvider);
  return PaymentServiceClient(transport);
});
