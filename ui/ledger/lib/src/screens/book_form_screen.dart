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
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/book_providers.dart';

const _kBookTypes = <String>[
  'platform',
  'group',
  'customer',
  'merchant',
  'agent',
  'branch',
];

/// Modal form for creating a new book. Returns the created [Book] when
/// the request succeeds so the caller can navigate to its detail page.
///
/// Field semantics:
/// - name: human-readable label (e.g. "Village Group A").
/// - type: conventional classification (platform/group/.../branch).
/// - parentId: optional FK to another book for hierarchical scopes.
/// - currency: default currency for the book (ISO 4217); empty is OK.
class BookFormScreen extends ConsumerStatefulWidget {
  const BookFormScreen({super.key, this.initialType = 'platform'});

  final String initialType;

  @override
  ConsumerState<BookFormScreen> createState() => _BookFormScreenState();
}

class _BookFormScreenState extends ConsumerState<BookFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameCtl = TextEditingController();
  final _parentCtl = TextEditingController();
  final _currencyCtl = TextEditingController();
  late String _type = widget.initialType;
  bool _saving = false;
  String? _error;

  @override
  void dispose() {
    _nameCtl.dispose();
    _parentCtl.dispose();
    _currencyCtl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      final book = await ref.read(bookNotifierProvider.notifier).create(
            name: _nameCtl.text.trim(),
            type: _type,
            currency: _currencyCtl.text.trim(),
            parentId: _parentCtl.text.trim(),
          );
      if (mounted) Navigator.of(context).pop<Book>(book);
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _saving = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('New book'),
      content: SizedBox(
        width: 420,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextFormField(
                  controller: _nameCtl,
                  decoration: const InputDecoration(
                    labelText: 'Name',
                    hintText: 'e.g. Village Group A',
                  ),
                  validator: (v) =>
                      v == null || v.trim().isEmpty ? 'Name is required' : null,
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: _type,
                  decoration: const InputDecoration(labelText: 'Type'),
                  items: _kBookTypes
                      .map((t) => DropdownMenuItem(value: t, child: Text(t)))
                      .toList(),
                  onChanged: (v) => setState(() => _type = v ?? widget.initialType),
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _parentCtl,
                  decoration: const InputDecoration(
                    labelText: 'Parent book id (optional)',
                    hintText: 'Nests this book under an organisation',
                  ),
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _currencyCtl,
                  decoration: const InputDecoration(
                    labelText: 'Currency (ISO 4217, optional)',
                    hintText: 'e.g. UGX',
                  ),
                ),
                if (_error != null) ...[
                  const SizedBox(height: 12),
                  Text(_error!, style: const TextStyle(color: Colors.red)),
                ],
              ],
            ),
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: _saving ? null : () => Navigator.of(context).pop(),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _saving ? null : _submit,
          child: _saving
              ? const SizedBox(
                  width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
              : const Text('Create'),
        ),
      ],
    );
  }
}
