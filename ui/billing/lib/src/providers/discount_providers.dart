import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_api_common/antinvestor_api_common.dart'
    show SearchRequest;
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'billing_transport_provider.dart';

/// Search discounts by query string.
final discountSearchProvider =
    FutureProvider.family<List<Discount>, String>((ref, query) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = SearchRequest()..query = query;
  final stream = client.searchDiscounts(request);
  return collectStream<SearchDiscountsResponse, Discount>(
    stream,
    extract: (r) => r.data,
  );
});

/// Notifier for discount mutations (create).
class DiscountNotifier extends StateNotifier<AsyncValue<void>> {
  DiscountNotifier(this._client) : super(const AsyncValue.data(null));
  final BillingServiceClient _client;

  Future<Discount> create(CreateDiscountRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createDiscount(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final discountNotifierProvider =
    StateNotifierProvider<DiscountNotifier, AsyncValue<void>>((ref) {
  final client = ref.watch(billingServiceClientProvider);
  return DiscountNotifier(client);
});
