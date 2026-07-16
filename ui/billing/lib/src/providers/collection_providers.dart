import 'package:antinvestor_ui_core/api/api_base.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/collection_client.dart';

/// Base URL for collection RPCs — same host as billing by default.
const _collectionUrl = String.fromEnvironment(
  'BILLING_URL',
  defaultValue: 'https://api.antinvestor.com/billing',
);

/// Optional app return URL after hosted checkout completes.
const checkoutReturnUrl = String.fromEnvironment(
  'CHECKOUT_RETURN_URL',
  defaultValue: '',
);

/// Known payment method keys exposed to admin UI for optional restriction.
/// The hosted checkout page is the source of truth for what is actually offered.
const knownPaymentMethods = <String, String>{
  'mpesa': 'M-PESA',
  'mtn_momo': 'MTN MoMo',
  'airtel_money': 'Airtel Money',
  'pawapay': 'Mobile Money (pawaPay)',
  'flutterwave': 'Flutterwave',
  'card': 'Card (Flutterwave)',
};

final collectionClientProvider = Provider<CollectionClient>((ref) {
  final tokenProvider = ref.watch(authTokenProviderProvider);
  return CollectionClient(
    baseUrl: _collectionUrl,
    tokenProvider: () => tokenProvider.ensureValidAccessToken(),
  );
});

/// Mutation notifier for streamlined collection flows.
class CollectionNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  CollectionClient get _client => ref.read(collectionClientProvider);

  Future<CollectionResult> collectPayment({
    required String invoiceId,
    String returnUrl = '',
    List<String> methods = const [],
  }) async {
    state = const AsyncValue.loading();
    try {
      final result = await _client.collectPayment(
        invoiceId: invoiceId,
        returnUrl: returnUrl.isNotEmpty ? returnUrl : checkoutReturnUrl,
        methods: methods,
      );
      state = const AsyncValue.data(null);
      return result;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<CollectionResult> startSubscription({
    required String profileId,
    required String planId,
    required String catalogVersionId,
    required String currency,
    String returnUrl = '',
    String payerDisplayName = '',
    List<String> methods = const [],
  }) async {
    state = const AsyncValue.loading();
    try {
      final result = await _client.startSubscription(
        profileId: profileId,
        planId: planId,
        catalogVersionId: catalogVersionId,
        currency: currency,
        returnUrl: returnUrl.isNotEmpty ? returnUrl : checkoutReturnUrl,
        payerDisplayName: payerDisplayName,
        methods: methods,
      );
      state = const AsyncValue.data(null);
      return result;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<ConfirmPaymentResult> confirmPayment(String sessionRef) async {
    state = const AsyncValue.loading();
    try {
      final result = await _client.confirmPayment(sessionRef: sessionRef);
      state = const AsyncValue.data(null);
      return result;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<CancelCollectionResult> cancelSubscription(String subscriptionId) async {
    state = const AsyncValue.loading();
    try {
      final result =
          await _client.cancelSubscription(subscriptionId: subscriptionId);
      state = const AsyncValue.data(null);
      return result;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final collectionNotifierProvider =
    NotifierProvider<CollectionNotifier, AsyncValue<void>>(
  CollectionNotifier.new,
);
