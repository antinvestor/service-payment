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
import 'package:antinvestor_ui_core/widgets/admin_entity_list_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/book_providers.dart';
import 'book_form_screen.dart';

/// Conventional book types surfaced in the UI. The backend accepts any
/// string but we curate this list for the type filter so operators see
/// recognised entity classifications first.
const _kBookTypes = <String>[
  'platform',
  'group',
  'customer',
  'merchant',
  'agent',
  'branch',
];

/// Screen that lists books of a given type. ListBooksByType requires a
/// type filter — there is no "list everything" RPC by design because
/// books for different scopes (platform vs. group vs. merchant) tend to
/// be operationally distinct concerns.
class BookListScreen extends ConsumerStatefulWidget {
  const BookListScreen({super.key});

  @override
  ConsumerState<BookListScreen> createState() => _BookListScreenState();
}

class _BookListScreenState extends ConsumerState<BookListScreen> {
  String _typeFilter = 'platform';

  @override
  Widget build(BuildContext context) {
    final asyncBooks = ref.watch(booksByTypeProvider(_typeFilter));

    return asyncBooks.when(
      loading: () => _buildTable(items: const []),
      error: (_, _) => _buildTable(items: const []),
      data: (books) => _buildTable(items: books),
    );
  }

  Widget _buildTable({required List<Book> items}) {
    return AdminEntityListPage<Book>(
      title: 'Books',
      breadcrumbs: const ['Ledger', 'Books'],
      searchHint: 'Filter books in this type',
      onSearch: (_) {},
      actions: [
        _buildTypeFilter(),
        const SizedBox(width: 12),
        FilledButton.icon(
          onPressed: _openCreateBook,
          icon: const Icon(Icons.add),
          label: const Text('New book'),
        ),
      ],
      columns: const [
        DataColumn(label: Text('ID')),
        DataColumn(label: Text('Name')),
        DataColumn(label: Text('Type')),
        DataColumn(label: Text('Parent')),
        DataColumn(label: Text('Currency')),
      ],
      items: items,
      rowBuilder: (book, selected, onSelect) => DataRow(
        selected: selected,
        onSelectChanged: (_) => onSelect(),
        cells: [
          DataCell(Text(_truncate(book.id, 16))),
          DataCell(Text(book.name)),
          DataCell(Text(book.type)),
          DataCell(Text(_truncate(book.parentId, 12))),
          DataCell(Text(book.currency)),
        ],
      ),
      onRowNavigate: (book) => context.go('/ledger/books/${book.id}'),
      exportRow: (book) => [
        book.id,
        book.name,
        book.type,
        book.parentId,
        book.currency,
      ],
      onExport: (format, rowCount) {
        debugPrint('AUDIT: Exported $rowCount books as $format');
      },
    );
  }

  Widget _buildTypeFilter() {
    return DropdownButton<String>(
      value: _typeFilter,
      underline: const SizedBox.shrink(),
      borderRadius: BorderRadius.circular(8),
      items: _kBookTypes
          .map(
            (type) => DropdownMenuItem<String>(
              value: type,
              child: Text(_humaniseType(type)),
            ),
          )
          .toList(),
      onChanged: (value) {
        if (value != null) setState(() => _typeFilter = value);
      },
    );
  }

  void _openCreateBook() async {
    final created = await showDialog<Book?>(
      context: context,
      builder: (_) => BookFormScreen(initialType: _typeFilter),
    );
    if (created != null && mounted) {
      ref.invalidate(booksByTypeProvider(created.type));
    }
  }

  String _truncate(String value, int maxLength) {
    if (value.isEmpty) return '';
    if (value.length <= maxLength) return value;
    return '${value.substring(0, maxLength - 1)}...';
  }

  String _humaniseType(String t) =>
      t.isEmpty ? '' : '${t[0].toUpperCase()}${t.substring(1)}';
}
