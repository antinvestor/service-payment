import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'billing_transport_provider.dart';

/// List subscriptions by profile ID.
final subscriptionListProvider =
    FutureProvider.family<List<Subscription>, String>((ref, profileId) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = ListSubscriptionsRequest()..profileId = profileId;
  final response = await client.listSubscriptions(request);
  return response.data;
});

/// Get a single subscription by ID.
final subscriptionProvider =
    FutureProvider.family<Subscription, String>((ref, id) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = GetSubscriptionRequest()..id = id;
  final response = await client.getSubscription(request);
  return response.data;
});

/// Notifier for subscription mutations (create, cancel).
class SubscriptionNotifier extends StateNotifier<AsyncValue<void>> {
  SubscriptionNotifier(this._client) : super(const AsyncValue.data(null));
  final BillingServiceClient _client;

  Future<Subscription> create(CreateSubscriptionRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createSubscription(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Subscription> cancel(CancelSubscriptionRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.cancelSubscription(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final subscriptionNotifierProvider =
    StateNotifierProvider<SubscriptionNotifier, AsyncValue<void>>((ref) {
  final client = ref.watch(billingServiceClientProvider);
  return SubscriptionNotifier(client);
});
