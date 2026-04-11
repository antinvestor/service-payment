import 'package:antinvestor_api_common/antinvestor_api_common.dart'
    show SearchRequest, StatusResponse;
import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'payment_transport_provider.dart';

/// Search payments by query string (payment links are a subset of payments).
/// SearchResponse has `repeated Payment data`, not a separate links field.
final paymentLinkSearchProvider =
    FutureProvider.family<List<Payment>, String>((ref, query) async {
  final client = ref.watch(paymentServiceClientProvider);
  final request = SearchRequest()..query = query;
  final stream = client.search(request);
  return collectStream<SearchResponse, Payment>(
    stream,
    extract: (r) => r.data,
  );
});

/// Notifier for payment link mutations.
class PaymentLinkNotifier extends StateNotifier<AsyncValue<void>> {
  PaymentLinkNotifier(this._client) : super(const AsyncValue.data(null));
  final PaymentServiceClient _client;

  /// CreatePaymentLink returns StatusResponse (via CreatePaymentLinkResponse.data).
  Future<StatusResponse> create(CreatePaymentLinkRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createPaymentLink(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final paymentLinkNotifierProvider =
    StateNotifierProvider<PaymentLinkNotifier, AsyncValue<void>>((ref) {
  final client = ref.watch(paymentServiceClientProvider);
  return PaymentLinkNotifier(client);
});
