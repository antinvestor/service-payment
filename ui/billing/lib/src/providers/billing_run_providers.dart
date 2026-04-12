import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'billing_transport_provider.dart';

/// Get a billing run by ID.
final billingRunProvider =
    FutureProvider.family<BillingRun, String>((ref, id) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = GetBillingRunRequest()..id = id;
  final response = await client.getBillingRun(request);
  return response.data;
});

/// Notifier for billing run mutations (runBilling).
class BillingRunNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  BillingServiceClient get _client =>
      ref.read(billingServiceClientProvider);

  Future<BillingRun> runBilling(RunBillingRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.runBilling(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final billingRunNotifierProvider =
    NotifierProvider<BillingRunNotifier, AsyncValue<void>>(
        BillingRunNotifier.new);
