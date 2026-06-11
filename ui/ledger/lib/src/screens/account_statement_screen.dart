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

import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../export/report_export.dart';
import '../models/report_models.dart';
import '../providers/report_providers.dart';

/// Account statement screen.
///
/// Walks the consumer through one account's activity over a period:
/// opening balance carried in from before the period, the chronological
/// entries within with running balances, the closing balance and the
/// per-side totals. Exports the same data as CSV / Excel / PDF.
class AccountStatementScreen extends ConsumerStatefulWidget {
  const AccountStatementScreen({super.key, required this.accountId});

  final String accountId;

  @override
  ConsumerState<AccountStatementScreen> createState() =>
      _AccountStatementScreenState();
}

class _AccountStatementScreenState
    extends ConsumerState<AccountStatementScreen> {
  final _fromCtl = TextEditingController();
  final _toCtl = TextEditingController();

  @override
  void dispose() {
    _fromCtl.dispose();
    _toCtl.dispose();
    super.dispose();
  }

  AccountStatementQuery get _query => AccountStatementQuery(
    accountId: widget.accountId,
    from: _fromCtl.text.trim(),
    to: _toCtl.text.trim(),
  );

  @override
  Widget build(BuildContext context) {
    final asyncReport = ref.watch(accountStatementProvider(_query));
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text('Statement — ${widget.accountId}'),
        actions: [
          asyncReport.maybeWhen(
            data: (r) => _ExportMenu(report: r),
            orElse: () => const SizedBox.shrink(),
          ),
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(accountStatementProvider(_query)),
          ),
        ],
      ),
      body: Column(
        children: [
          _FilterBar(
            fromCtl: _fromCtl,
            toCtl: _toCtl,
            onApply: () => setState(() {}),
          ),
          Expanded(
            child: asyncReport.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(
                child: Text(
                  'Error: $e',
                  style: TextStyle(color: theme.colorScheme.error),
                ),
              ),
              data: (report) => _Body(report: report),
            ),
          ),
        ],
      ),
    );
  }
}

class _FilterBar extends StatelessWidget {
  const _FilterBar({
    required this.fromCtl,
    required this.toCtl,
    required this.onApply,
  });
  final TextEditingController fromCtl;
  final TextEditingController toCtl;
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
              width: 240,
              child: TextField(
                controller: fromCtl,
                decoration: const InputDecoration(
                  labelText: 'From (RFC3339)',
                  hintText: '2026-01-01T00:00:00Z',
                ),
              ),
            ),
            SizedBox(
              width: 240,
              child: TextField(
                controller: toCtl,
                decoration: const InputDecoration(
                  labelText: 'To (RFC3339)',
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

class _Body extends StatelessWidget {
  const _Body({required this.report});
  final AccountStatementReport report;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Wrap(
              spacing: 24,
              runSpacing: 12,
              children: [
                _summary(
                  theme,
                  'Opening balance',
                  formatMoney(report.openingBalance),
                ),
                _summary(
                  theme,
                  'Closing balance',
                  formatMoney(report.closingBalance),
                ),
                _summary(
                  theme,
                  'Period debits',
                  formatMoney(report.totalDebits),
                ),
                _summary(
                  theme,
                  'Period credits',
                  formatMoney(report.totalCredits),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: DataTable(
                headingTextStyle: theme.textTheme.labelLarge,
                columns: const [
                  DataColumn(label: Text('Transacted')),
                  DataColumn(label: Text('Transaction')),
                  DataColumn(label: Text('DR/CR')),
                  DataColumn(label: Text('Amount'), numeric: true),
                  DataColumn(label: Text('Running'), numeric: true),
                ],
                rows: report.entries
                    .map(
                      (e) => DataRow(
                        cells: [
                          DataCell(Text(e.transactedAt)),
                          DataCell(Text(e.transactionId)),
                          DataCell(Text(e.credit ? 'CR' : 'DR')),
                          DataCell(Text(formatMoney(e.amount))),
                          DataCell(Text(formatMoney(e.runningBalance))),
                        ],
                      ),
                    )
                    .toList(),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _summary(ThemeData theme, String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: theme.textTheme.labelMedium?.copyWith(
            color: theme.colorScheme.outline,
          ),
        ),
        const SizedBox(height: 4),
        Text(value, style: theme.textTheme.titleMedium),
      ],
    );
  }
}

class _ExportMenu extends StatelessWidget {
  const _ExportMenu({required this.report});
  final AccountStatementReport report;

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<StatementExportFormat>(
      icon: const Icon(Icons.download),
      tooltip: 'Export',
      onSelected: (format) async {
        try {
          await ReportExport.accountStatement(report, format);
        } catch (e) {
          if (context.mounted) {
            ScaffoldMessenger.of(
              context,
            ).showSnackBar(SnackBar(content: Text('Export failed: $e')));
          }
        }
      },
      itemBuilder: (_) => const [
        PopupMenuItem(
          value: StatementExportFormat.csv,
          child: Text('Export CSV'),
        ),
        PopupMenuItem(
          value: StatementExportFormat.excel,
          child: Text('Export Excel'),
        ),
        PopupMenuItem(
          value: StatementExportFormat.pdf,
          child: Text('Export PDF'),
        ),
      ],
    );
  }
}
