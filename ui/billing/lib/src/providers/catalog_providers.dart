import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'billing_transport_provider.dart';

/// Search catalog versions by query string.
final catalogVersionSearchProvider =
    FutureProvider.family<List<CatalogVersion>, String>((ref, query) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = SearchRequest()..query = query;
  final stream = client.searchCatalogVersions(request);
  return collectStream<SearchCatalogVersionsResponse, CatalogVersion>(
    stream,
    extract: (r) => r.data,
  );
});

/// Get a single catalog version by ID.
final catalogVersionProvider =
    FutureProvider.family<CatalogVersion, String>((ref, id) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = GetCatalogVersionRequest()..id = id;
  final response = await client.getCatalogVersion(request);
  return response.data;
});

/// Notifier for catalog mutations (create, publish, plans, components, tiers).
class CatalogNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  BillingServiceClient get _client =>
      ref.read(billingServiceClientProvider);

  Future<CatalogVersion> create(CreateCatalogVersionRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createCatalogVersion(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<CatalogVersion> publish(PublishCatalogVersionRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.publishCatalogVersion(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Plan> createPlan(CreatePlanRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createPlan(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Component> createComponent(CreateComponentRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createComponent(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Tier> createTier(CreateTierRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createTier(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final catalogNotifierProvider =
    NotifierProvider<CatalogNotifier, AsyncValue<void>>(CatalogNotifier.new);
