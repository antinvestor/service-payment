import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'billing_transport_provider.dart';

/// Get credit balance by profile ID.
final creditBalanceProvider =
    FutureProvider.family<GetCreditBalanceResponse, String>(
        (ref, profileId) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = GetCreditBalanceRequest()..profileId = profileId;
  return await client.getCreditBalance(request);
});

/// Notifier for credit mutations (grantCredit).
class CreditNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  BillingServiceClient get _client =>
      ref.read(billingServiceClientProvider);

  Future<GrantCreditResponse> grantCredit(GrantCreditRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.grantCredit(request);
      state = const AsyncValue.data(null);
      return response;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final creditNotifierProvider =
    NotifierProvider<CreditNotifier, AsyncValue<void>>(CreditNotifier.new);
