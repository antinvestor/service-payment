import 'package:antinvestor_api_common/antinvestor_api_common.dart'
    show SearchRequest, StatusRequest, StatusResponse, StatusUpdateRequest,
    StatusUpdateResponse;
import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'payment_transport_provider.dart';

/// Search payments by query string.
final paymentSearchProvider =
    FutureProvider.family<List<Payment>, String>((ref, query) async {
  final client = ref.watch(paymentServiceClientProvider);
  final request = SearchRequest()..query = query;
  final stream = client.search(request);
  return collectStream<SearchResponse, Payment>(
    stream,
    extract: (r) => r.data,
  );
});

/// Get payment status by ID.
/// Status RPC returns common.v1.StatusResponse, not Payment.
final paymentStatusProvider =
    FutureProvider.family<StatusResponse, String>((ref, id) async {
  final client = ref.watch(paymentServiceClientProvider);
  final request = StatusRequest()..id = id;
  return await client.status(request);
});

/// Notifier for payment mutations (send, receive, release, reconcile, statusUpdate).
class PaymentNotifier extends StateNotifier<AsyncValue<void>> {
  PaymentNotifier(this._client) : super(const AsyncValue.data(null));
  final PaymentServiceClient _client;

  /// Send returns StatusResponse (via SendResponse.data).
  Future<StatusResponse> send(SendRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.send(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  /// Receive returns StatusResponse (via ReceiveResponse.data).
  Future<StatusResponse> receive(ReceiveRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.receive(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  /// Release returns StatusResponse (via ReleaseResponse.data).
  Future<StatusResponse> release(ReleaseRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.release(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  /// Reconcile returns ReconcileResponse (id, transaction_id, reference_id, status, description).
  Future<ReconcileResponse> reconcile(ReconcileRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.reconcile(request);
      state = const AsyncValue.data(null);
      return response;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<StatusUpdateResponse> statusUpdate(StatusUpdateRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.statusUpdate(request);
      state = const AsyncValue.data(null);
      return response;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final paymentNotifierProvider =
    StateNotifierProvider<PaymentNotifier, AsyncValue<void>>((ref) {
  final client = ref.watch(paymentServiceClientProvider);
  return PaymentNotifier(client);
});
