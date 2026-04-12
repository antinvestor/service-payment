// ignore_for_file: implementation_imports
import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_api_ledger/src/common/v1/common.pb.dart'
    as ledger_common;
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'ledger_transport_provider.dart';

/// Search transactions by query string.
final transactionSearchProvider =
    FutureProvider.family<List<Transaction>, String>((ref, query) async {
  final client = ref.watch(ledgerServiceClientProvider);
  final request = ledger_common.SearchRequest()..query = query;
  final stream = client.searchTransactions(request);
  return collectStream<SearchTransactionsResponse, Transaction>(
    stream,
    extract: (r) => r.data,
  );
});

/// Search transaction entries by query string.
final transactionEntrySearchProvider =
    FutureProvider.family<List<TransactionEntry>, String>((ref, query) async {
  final client = ref.watch(ledgerServiceClientProvider);
  final request = ledger_common.SearchRequest()..query = query;
  final stream = client.searchTransactionEntries(request);
  return collectStream<SearchTransactionEntriesResponse, TransactionEntry>(
    stream,
    extract: (r) => r.data,
  );
});

/// Notifier for transaction mutations (create, reverse, update).
class TransactionNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  LedgerServiceClient get _client =>
      ref.read(ledgerServiceClientProvider);

  Future<Transaction> create(CreateTransactionRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createTransaction(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Transaction> reverse(ReverseTransactionRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.reverseTransaction(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Transaction> update(UpdateTransactionRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.updateTransaction(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final transactionNotifierProvider =
    NotifierProvider<TransactionNotifier, AsyncValue<void>>(
        TransactionNotifier.new);
