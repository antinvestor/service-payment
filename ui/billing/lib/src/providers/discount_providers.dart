import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
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
class DiscountNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  BillingServiceClient get _client =>
      ref.read(billingServiceClientProvider);

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
    NotifierProvider<DiscountNotifier, AsyncValue<void>>(
        DiscountNotifier.new);
