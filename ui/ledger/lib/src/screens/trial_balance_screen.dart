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
import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../export/report_export.dart';
import '../providers/report_providers.dart';

/// Trial balance report screen.
///
/// Surfaces the per-account debit / credit totals plus per-currency grand
/// totals with the textbook integrity check (debits == credits). Filters
/// for currency, ledger-type, book scope and as-of date narrow the view.
///
/// Export buttons in the app bar produce CSV / Excel / PDF downloads of
/// the currently displayed report — same data the operator sees.
class TrialBalanceScreen extends ConsumerStatefulWidget {
  const TrialBalanceScreen({super.key, this.initialBookId = ''});

  /// Book to scope to when the screen is reached from a Book detail page.
  /// Empty means "no book filter" — typically used by platform-level
  /// operators who want the full picture across all books.
  final String initialBookId;

  @override
  ConsumerState<TrialBalanceScreen> createState() => _TrialBalanceScreenState();
}

const _kLedgerTypes = <String>['', 'ASSET', 'LIABILITY', 'INCOME', 'EXPENSE', 'CAPITAL'];

class _TrialBalanceScreenState extends ConsumerState<TrialBalanceScreen> {
  final _currencyCtl = TextEditingController();
  final _asOfCtl = TextEditingController();
  String _ledgerType = '';
  late String _bookId = widget.initialBookId;

  @override
  void dispose() {
    _currencyCtl.dispose();
    _asOfCtl.dispose();
    super.dispose();
  }

  TrialBalanceQuery get _query => TrialBalanceQuery(
        currency: _currencyCtl.text.trim().toUpperCase(),
        ledgerType: _ledgerType,
        bookIds: _bookId.isEmpty ? const [] : [_bookId],
        asOf: _asOfCtl.text.trim(),
      );

  @override
  Widget build(BuildContext context) {
    final asyncReport = ref.watch(trialBalanceProvider(_query));
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Trial balance'),
        actions: [
          asyncReport.maybeWhen(
            data: (report) => _ExportMenu(report: report),
            orElse: () => const SizedBox.shrink(),
          ),
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(trialBalanceProvider(_query)),
          ),
        ],
      ),
      body: Column(
        children: [
          _FilterBar(
            currencyCtl: _currencyCtl,
            asOfCtl: _asOfCtl,
            ledgerType: _ledgerType,
            bookId: _bookId,
            onLedgerTypeChanged: (v) => setState(() => _ledgerType = v),
            onBookIdChanged: (v) => setState(() => _bookId = v),
            onApply: () => setState(() {}),
          ),
          Expanded(
            child: asyncReport.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(child: Text('Error: $e', style: TextStyle(color: theme.colorScheme.error))),
              data: (report) => _TrialBalanceBody(report: report),
            ),
          ),
        ],
      ),
    );
  }
}

class _FilterBar extends StatelessWidget {
  const _FilterBar({
    required this.currencyCtl,
    required this.asOfCtl,
    required this.ledgerType,
    required this.bookId,
    required this.onLedgerTypeChanged,
    required this.onBookIdChanged,
    required this.onApply,
  });

  final TextEditingController currencyCtl;
  final TextEditingController asOfCtl;
  final String ledgerType;
  final String bookId;
  final ValueChanged<String> onLedgerTypeChanged;
  final ValueChanged<String> onBookIdChanged;
  final VoidCallback onApply;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.all(16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Wrap(
          spacing: 16,
          runSpacing: 12,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            SizedBox(
              width: 140,
              child: TextField(
                controller: currencyCtl,
                decoration: const InputDecoration(
                  labelText: 'Currency',
                  hintText: 'UGX / USD …',
                ),
                textCapitalization: TextCapitalization.characters,
              ),
            ),
            SizedBox(
              width: 180,
              child: DropdownButtonFormField<String>(
                initialValue: ledgerType,
                decoration: const InputDecoration(labelText: 'Ledger type'),
                items: _kLedgerTypes
                    .map((t) => DropdownMenuItem(
                          value: t,
                          child: Text(t.isEmpty ? 'All types' : t),
                        ))
                    .toList(),
                onChanged: (v) => onLedgerTypeChanged(v ?? ''),
              ),
            ),
            SizedBox(
              width: 240,
              child: TextField(
                controller: TextEditingController(text: bookId),
                decoration: const InputDecoration(
                  labelText: 'Book id (optional)',
                  hintText: 'Scope to one book',
                ),
                onSubmitted: onBookIdChanged,
              ),
            ),
            SizedBox(
              width: 220,
              child: TextField(
                controller: asOfCtl,
                decoration: const InputDecoration(
                  labelText: 'As-of (RFC3339)',
                  hintText: '2026-12-31T23:59:59Z',
                ),
              ),
            ),
            FilledButton.icon(
              onPressed: onApply,
              icon: const Icon(Icons.filter_list),
              label: const Text('Apply'),
            ),
          ],
        ),
      ),
    );
  }
}

class _TrialBalanceBody extends StatelessWidget {
  const _TrialBalanceBody({required this.report});
  final GetTrialBalanceResponse report;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        _TotalsCard(report: report),
        const SizedBox(height: 16),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: DataTable(
                headingTextStyle: theme.textTheme.labelLarge,
                columns: const [
                  DataColumn(label: Text('Account')),
                  DataColumn(label: Text('Ledger')),
                  DataColumn(label: Text('Type')),
                  DataColumn(label: Text('Currency')),
                  DataColumn(label: Text('Debits'), numeric: true),
                  DataColumn(label: Text('Credits'), numeric: true),
                  DataColumn(label: Text('Net'), numeric: true),
                ],
                rows: report.lines
                    .map(
                      (l) => DataRow(cells: [
                        DataCell(Text(l.accountId)),
                        DataCell(Text(l.ledgerId)),
                        DataCell(Text(l.ledgerType.name)),
                        DataCell(Text(l.currency)),
                        DataCell(Text(formatMoney(l.totalDebits))),
                        DataCell(Text(formatMoney(l.totalCredits))),
                        DataCell(Text(formatMoney(l.netBalance))),
                      ]),
                    )
                    .toList(),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _TotalsCard extends StatelessWidget {
  const _TotalsCard({required this.report});
  final GetTrialBalanceResponse report;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    if (report.totals.isEmpty) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Text(
            'No activity in the selected scope.',
            style: theme.textTheme.bodyMedium,
          ),
        ),
      );
    }
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Per-currency totals',
                style: theme.textTheme.titleMedium),
            const SizedBox(height: 12),
            ...report.totals.map((t) => Padding(
                  padding: const EdgeInsets.symmetric(vertical: 4),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(
                          color: t.isBalanced
                              ? Colors.green.withValues(alpha: 0.15)
                              : Colors.red.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          t.isBalanced ? 'BALANCED' : 'UNBALANCED',
                          style: TextStyle(
                            fontWeight: FontWeight.w600,
                            color: t.isBalanced
                                ? Colors.green.shade800
                                : Colors.red.shade800,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Text(t.currency,
                          style: theme.textTheme.titleSmall),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Wrap(
                          spacing: 16,
                          children: [
                            Text('Debits ${formatMoney(t.totalDebits)}'),
                            Text('Credits ${formatMoney(t.totalCredits)}'),
                          ],
                        ),
                      ),
                    ],
                  ),
                )),
          ],
        ),
      ),
    );
  }
}

class _ExportMenu extends StatelessWidget {
  const _ExportMenu({required this.report});
  final GetTrialBalanceResponse report;

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<TrialBalanceExportFormat>(
      icon: const Icon(Icons.download),
      tooltip: 'Export',
      onSelected: (format) async {
        try {
          await ReportExport.trialBalance(report, format);
        } catch (e) {
          if (context.mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text('Export failed: $e')),
            );
          }
        }
      },
      itemBuilder: (_) => const [
        PopupMenuItem(
          value: TrialBalanceExportFormat.csv,
          child: Text('Export CSV'),
        ),
        PopupMenuItem(
          value: TrialBalanceExportFormat.excel,
          child: Text('Export Excel'),
        ),
        PopupMenuItem(
          value: TrialBalanceExportFormat.pdf,
          child: Text('Export PDF'),
        ),
      ],
    );
  }
}
