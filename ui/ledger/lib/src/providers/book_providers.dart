// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'ledger_transport_provider.dart';

/// Books-by-type provider. Uses the typed `ListBooksByType` RPC so the
/// returned list is already scoped to the requested entity classification.
final booksByTypeProvider =
    FutureProvider.family<List<Book>, String>((ref, bookType) async {
  if (bookType.isEmpty) return <Book>[];
  final client = ref.watch(ledgerServiceClientProvider);
  final response = await client.listBooksByType(
    ListBooksByTypeRequest()..type = bookType,
  );
  return response.data;
});

/// Single-book lookup by id. Used by the detail screen.
final bookByIdProvider =
    FutureProvider.family<Book, String>((ref, id) async {
  final client = ref.watch(ledgerServiceClientProvider);
  final response = await client.getBook(GetBookRequest()..id = id);
  return response.data;
});

/// Notifier for book mutations (create). The returned [Book] carries the
/// server-assigned id so the caller can navigate straight to the detail
/// screen without an extra lookup.
class BookNotifier extends Notifier<AsyncValue<void>> {
  @override
  AsyncValue<void> build() => const AsyncValue.data(null);

  LedgerServiceClient get _client => ref.read(ledgerServiceClientProvider);

  /// Create a new book. [parentId] may be empty for top-level scopes
  /// (platform / agent root) or set to nest a group under an organisation.
  Future<Book> create({
    required String name,
    required String type,
    String currency = '',
    String parentId = '',
  }) async {
    state = const AsyncValue.loading();
    try {
      final request = CreateBookRequest()
        ..name = name
        ..type = type
        ..currency = currency
        ..parentId = parentId;
      final response = await _client.createBook(request);
      // Invalidate the by-type cache so the new book shows up immediately.
      ref.invalidate(booksByTypeProvider(type));
      state = const AsyncValue.data(null);
      return response.data;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }
}

final bookNotifierProvider =
    NotifierProvider<BookNotifier, AsyncValue<void>>(BookNotifier.new);
