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

import '../providers/ledger_providers.dart';
import '../widgets/ledger_type_badge.dart';

/// Modal form for creating a new ledger. Returns the created [Ledger] when
/// the request succeeds so the caller can navigate to its detail page.
///
/// Field semantics:
/// - type: accounting classification ([LedgerType]).
/// - parentId: optional FK to another ledger for hierarchical charts.
/// - name / currency: human-readable label and default ISO 4217 currency,
///   carried in the ledger's free-form `data` struct (the 1.53 API has no
///   dedicated columns for them).
class LedgerFormScreen extends ConsumerStatefulWidget {
  const LedgerFormScreen({super.key, this.initialType = LedgerType.ASSET});

  final LedgerType initialType;

  @override
  ConsumerState<LedgerFormScreen> createState() => _LedgerFormScreenState();
}

class _LedgerFormScreenState extends ConsumerState<LedgerFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameCtl = TextEditingController();
  final _parentCtl = TextEditingController();
  final _currencyCtl = TextEditingController();
  late LedgerType _type = widget.initialType;
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
      final data = Struct();
      final name = _nameCtl.text.trim();
      final currency = _currencyCtl.text.trim();
      if (name.isNotEmpty) {
        data.fields['name'] = Value()..stringValue = name;
      }
      if (currency.isNotEmpty) {
        data.fields['currency'] = Value()..stringValue = currency;
      }
      final request = CreateLedgerRequest()
        ..type = _type
        ..parentId = _parentCtl.text.trim()
        ..data = data;
      final ledger = await ref
          .read(ledgerNotifierProvider.notifier)
          .create(request);
      if (mounted) Navigator.of(context).pop<Ledger>(ledger);
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
      title: const Text('New ledger'),
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
                    hintText: 'e.g. Customer wallets',
                  ),
                  validator: (v) =>
                      v == null || v.trim().isEmpty ? 'Name is required' : null,
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<LedgerType>(
                  initialValue: _type,
                  decoration: const InputDecoration(labelText: 'Type'),
                  items: LedgerType.values
                      .map(
                        (t) => DropdownMenuItem(
                          value: t,
                          child: Text(ledgerTypeLabel(t)),
                        ),
                      )
                      .toList(),
                  onChanged: (v) =>
                      setState(() => _type = v ?? widget.initialType),
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _parentCtl,
                  decoration: const InputDecoration(
                    labelText: 'Parent ledger id (optional)',
                    hintText: 'Nests this ledger in the chart of accounts',
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
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Text('Create'),
        ),
      ],
    );
  }
}
