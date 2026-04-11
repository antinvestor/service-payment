import 'package:antinvestor_api_common/antinvestor_api_common.dart'
    show SearchRequest;
import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/api/stream_helpers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'ledger_transport_provider.dart';

/// Search accounts by query string.
final accountSearchProvider =
    FutureProvider.family<List<Account>, String>((ref, query) async {
  final client = ref.watch(ledgerServiceClientProvider);
  final request = SearchRequest()..query = query;
  final stream = client.searchAccounts(request);
  return collectStream<SearchAccountsResponse, Account>(
    stream,
    extract: (r) => r.data,
  );
});

/// Notifier for account mutations (create, update).
class AccountNotifier extends StateNotifier<AsyncValue<void>> {
  AccountNotifier(this._client) : super(const AsyncValue.data(null));
  final LedgerServiceClient _client;

  Future<Account> create(CreateAccountRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.createAccount(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<Account> update(UpdateAccountRequest request) async {
    state = const AsyncValue.loading();
    try {
      final response = await _client.updateAccount(request);
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final accountNotifierProvider =
    StateNotifierProvider<AccountNotifier, AsyncValue<void>>((ref) {
  final client = ref.watch(ledgerServiceClientProvider);
  return AccountNotifier(client);
});
