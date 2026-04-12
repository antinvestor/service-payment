// ignore_for_file: implementation_imports
import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_api_ledger/src/common/v1/common.pb.dart'
    as ledger_common;
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'ledger_transport_provider.dart';

/// Search ledgers by query string.
final ledgerSearchProvider =
    FutureProvider.family<List<Ledger>, String>((ref, query) async {
  final client = ref.watch(ledgerServiceClientProvider);
  final request = ledger_common.SearchRequest()..query = query;
  final stream = client.searchLedgers(request);
  return collectStream<SearchLedgersResponse, Ledger>(
    stream,
    extract: (r) => r.data,
  );
});

/// Notifier for ledger mutations (create, update).
class LedgerNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  LedgerServiceClient get _client =>
      ref.read(ledgerServiceClientProvider);

  Future<Ledger> create(CreateLedgerRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createLedger(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Ledger> update(UpdateLedgerRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.updateLedger(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final ledgerNotifierProvider =
    NotifierProvider<LedgerNotifier, AsyncValue<void>>(LedgerNotifier.new);
