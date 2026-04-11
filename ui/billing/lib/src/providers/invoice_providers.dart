import 'package:antinvestor_api_billing/antinvestor_api_billing.dart';
import 'package:antinvestor_api_common/antinvestor_api_common.dart'
    show SearchRequest;
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'billing_transport_provider.dart';

/// Search invoices by query string.
final invoiceSearchProvider =
    FutureProvider.family<List<Invoice>, String>((ref, query) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = SearchRequest()..query = query;
  final stream = client.searchInvoices(request);
  return collectStream<SearchInvoicesResponse, Invoice>(
    stream,
    extract: (r) => r.data,
  );
});

/// Get a single invoice by ID.
final invoiceProvider =
    FutureProvider.family<Invoice, String>((ref, id) async {
  final client = ref.watch(billingServiceClientProvider);
  final request = GetInvoiceRequest()..id = id;
  final response = await client.getInvoice(request);
  return response.data;
});

/// Notifier for invoice mutations (issue, void, recordPayment).
class InvoiceNotifier extends StateNotifier<AsyncValue<void>> {
  InvoiceNotifier(this._client) : super(const AsyncValue.data(null));
  final BillingServiceClient _client;

  Future<Invoice> issue(IssueInvoiceRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.issueInvoice(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Invoice> voidInvoice(VoidInvoiceRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.voidInvoice(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Invoice> recordPayment(RecordPaymentRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.recordPayment(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final invoiceNotifierProvider =
    StateNotifierProvider<InvoiceNotifier, AsyncValue<void>>((ref) {
  final client = ref.watch(billingServiceClientProvider);
  return InvoiceNotifier(client);
});
