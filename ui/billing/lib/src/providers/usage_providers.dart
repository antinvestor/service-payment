import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_api_common/antinvestor_api_common.dart'
    show SearchRequest;
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'billing_transport_provider.dart';

/// Search usage events by query string.
final usageEventSearchProvider =
    FutureProvider.family<List<UsageEvent>, String>((ref, query) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = SearchRequest()..query = query;
  final stream = client.searchUsageEvents(request);
  return collectStream<SearchUsageEventsResponse, UsageEvent>(
    stream,
    extract: (r) => r.data,
  );
});

/// Notifier for usage event mutations (ingest).
class UsageEventNotifier extends StateNotifier<AsyncValue<void>> {
  UsageEventNotifier(this._client) : super(const AsyncValue.data(null));
  final BillingServiceClient _client;

  /// Ingest returns a list of string IDs for the ingested events.
  Future<List<String>> ingest(IngestUsageEventRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.ingestUsageEvent(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final usageEventNotifierProvider =
    StateNotifierProvider<UsageEventNotifier, AsyncValue<void>>((ref) {
  final client = ref.watch(billingServiceClientProvider);
  return UsageEventNotifier(client);
});
