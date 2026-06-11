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

import 'dart:convert';
import 'dart:typed_data';

import 'package:antinvestor_ui_core/widgets/money_helpers.dart';
import 'package:csv/csv.dart';
import 'package:excel/excel.dart';
import 'package:file_saver/file_saver.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';

import '../models/report_models.dart';

/// Export formats supported for the trial balance report.
enum TrialBalanceExportFormat { csv, excel, pdf }

/// Export formats supported for the account statement report.
enum StatementExportFormat { csv, excel, pdf }

/// Unified export entry point so screens don't need to know how each
/// format is generated. All exports go through [FileSaver] on web /
/// desktop / mobile, and [Printing.sharePdf] for PDFs so the OS share
/// sheet handles the final destination (Save / AirPrint / WhatsApp / …).
class ReportExport {
  ReportExport._();

  /// Export the trial-balance report in the requested format.
  ///
  /// Filename pattern: `trial_balance_<UTCdateTime>.<ext>`.
  static Future<void> trialBalance(
    TrialBalanceReport report,
    TrialBalanceExportFormat format,
  ) async {
    final stamp = _stamp();
    final base = 'trial_balance_$stamp';
    switch (format) {
      case TrialBalanceExportFormat.csv:
        await _saveBytes(
          name: '$base.csv',
          mimeType: MimeType.csv,
          fileExtension: 'csv',
          bytes: Uint8List.fromList(utf8.encode(_trialBalanceCsv(report))),
        );
        return;
      case TrialBalanceExportFormat.excel:
        await _saveBytes(
          name: '$base.xlsx',
          mimeType: MimeType.microsoftExcel,
          fileExtension: 'xlsx',
          bytes: Uint8List.fromList(_trialBalanceExcel(report)),
        );
        return;
      case TrialBalanceExportFormat.pdf:
        final bytes = await _trialBalancePdf(report);
        await Printing.sharePdf(bytes: bytes, filename: '$base.pdf');
        return;
    }
  }

  /// Export the account-statement report in the requested format.
  static Future<void> accountStatement(
    AccountStatementReport report,
    StatementExportFormat format,
  ) async {
    final stamp = _stamp();
    final base = 'statement_${report.accountId}_$stamp';
    switch (format) {
      case StatementExportFormat.csv:
        await _saveBytes(
          name: '$base.csv',
          mimeType: MimeType.csv,
          fileExtension: 'csv',
          bytes: Uint8List.fromList(utf8.encode(_statementCsv(report))),
        );
        return;
      case StatementExportFormat.excel:
        await _saveBytes(
          name: '$base.xlsx',
          mimeType: MimeType.microsoftExcel,
          fileExtension: 'xlsx',
          bytes: Uint8List.fromList(_statementExcel(report)),
        );
        return;
      case StatementExportFormat.pdf:
        final bytes = await _statementPdf(report);
        await Printing.sharePdf(bytes: bytes, filename: '$base.pdf');
        return;
    }
  }

  // ---------- CSV ----------

  static String _trialBalanceCsv(TrialBalanceReport report) {
    final rows = <List<dynamic>>[
      [
        'Account',
        'Ledger',
        'LedgerType',
        'Currency',
        'TotalDebits',
        'TotalCredits',
        'NetBalance',
      ],
      ...report.lines.map(
        (l) => [
          l.accountId,
          l.ledgerId,
          l.ledgerType?.name ?? '',
          l.currency,
          formatMoney(l.debit),
          formatMoney(l.credit),
          formatMoney(l.netBalance),
        ],
      ),
      [],
      ['Per-currency totals', '', '', '', '', '', ''],
      ['Currency', 'TotalDebits', 'TotalCredits', 'IsBalanced', '', '', ''],
      ...report.totals.map(
        (t) => [
          t.currency,
          formatMoney(t.totalDebits),
          formatMoney(t.totalCredits),
          t.isBalanced ? 'BALANCED' : 'UNBALANCED',
          '',
          '',
          '',
        ],
      ),
    ];
    return const CsvEncoder().convert(rows);
  }

  static String _statementCsv(AccountStatementReport report) {
    final rows = <List<dynamic>>[
      ['Statement for account', report.accountId, '', '', ''],
      ['Currency', report.currency, '', '', ''],
      ['Opening balance', formatMoney(report.openingBalance), '', '', ''],
      ['Closing balance', formatMoney(report.closingBalance), '', '', ''],
      ['Period debits', formatMoney(report.totalDebits), '', '', ''],
      ['Period credits', formatMoney(report.totalCredits), '', '', ''],
      [],
      ['Transacted', 'Transaction', 'DR/CR', 'Amount', 'Running'],
      ...report.entries.map(
        (e) => [
          e.transactedAt,
          e.transactionId,
          e.credit ? 'CR' : 'DR',
          formatMoney(e.amount),
          formatMoney(e.runningBalance),
        ],
      ),
    ];
    return const CsvEncoder().convert(rows);
  }

  // ---------- Excel ----------

  static List<int> _trialBalanceExcel(TrialBalanceReport report) {
    final excel = Excel.createExcel();
    final sheet = excel['Trial Balance'];
    excel.delete('Sheet1');

    _appendRow(sheet, [
      'Account',
      'Ledger',
      'LedgerType',
      'Currency',
      'TotalDebits',
      'TotalCredits',
      'NetBalance',
    ], bold: true);
    for (final l in report.lines) {
      _appendRow(sheet, [
        l.accountId,
        l.ledgerId,
        l.ledgerType?.name ?? '',
        l.currency,
        formatMoney(l.debit),
        formatMoney(l.credit),
        formatMoney(l.netBalance),
      ]);
    }
    _appendRow(sheet, []);
    _appendRow(sheet, ['Per-currency totals'], bold: true);
    _appendRow(sheet, [
      'Currency',
      'TotalDebits',
      'TotalCredits',
      'IsBalanced',
    ], bold: true);
    for (final t in report.totals) {
      _appendRow(sheet, [
        t.currency,
        formatMoney(t.totalDebits),
        formatMoney(t.totalCredits),
        t.isBalanced ? 'BALANCED' : 'UNBALANCED',
      ]);
    }

    final bytes = excel.encode();
    if (bytes == null) {
      throw StateError('failed to encode excel workbook');
    }
    return bytes;
  }

  static List<int> _statementExcel(AccountStatementReport report) {
    final excel = Excel.createExcel();
    final sheet = excel['Statement'];
    excel.delete('Sheet1');

    _appendRow(sheet, ['Account', report.accountId], bold: true);
    _appendRow(sheet, ['Currency', report.currency]);
    _appendRow(sheet, ['Opening balance', formatMoney(report.openingBalance)]);
    _appendRow(sheet, ['Closing balance', formatMoney(report.closingBalance)]);
    _appendRow(sheet, ['Period debits', formatMoney(report.totalDebits)]);
    _appendRow(sheet, ['Period credits', formatMoney(report.totalCredits)]);
    _appendRow(sheet, []);

    _appendRow(sheet, [
      'Transacted',
      'Transaction',
      'DR/CR',
      'Amount',
      'Running',
    ], bold: true);
    for (final e in report.entries) {
      _appendRow(sheet, [
        e.transactedAt,
        e.transactionId,
        e.credit ? 'CR' : 'DR',
        formatMoney(e.amount),
        formatMoney(e.runningBalance),
      ]);
    }

    final bytes = excel.encode();
    if (bytes == null) {
      throw StateError('failed to encode excel workbook');
    }
    return bytes;
  }

  static void _appendRow(
    Sheet sheet,
    List<dynamic> values, {
    bool bold = false,
  }) {
    final cells = values.map<CellValue?>((v) {
      if (v == null) return null;
      if (v is num) return DoubleCellValue(v.toDouble());
      return TextCellValue(v.toString());
    }).toList();
    sheet.appendRow(cells);
    if (bold) {
      final rowIndex = sheet.maxRows - 1;
      for (var col = 0; col < cells.length; col++) {
        sheet
            .cell(
              CellIndex.indexByColumnRow(columnIndex: col, rowIndex: rowIndex),
            )
            .cellStyle = CellStyle(
          bold: true,
        );
      }
    }
  }

  // ---------- PDF ----------

  static Future<Uint8List> _trialBalancePdf(TrialBalanceReport report) async {
    final doc = pw.Document();
    doc.addPage(
      pw.MultiPage(
        pageFormat: PdfPageFormat.a4,
        build: (ctx) => [
          pw.Header(level: 0, child: pw.Text('Trial Balance')),
          pw.Paragraph(
            text:
                'Per-account debit / credit totals plus per-currency integrity check. '
                'Generated at ${DateTime.now().toUtc().toIso8601String()}.',
          ),
          pw.TableHelper.fromTextArray(
            headerStyle: pw.TextStyle(fontWeight: pw.FontWeight.bold),
            headers: const [
              'Account',
              'Ledger',
              'Type',
              'Currency',
              'Debits',
              'Credits',
              'Net',
            ],
            data: report.lines
                .map(
                  (l) => [
                    l.accountId,
                    l.ledgerId,
                    l.ledgerType?.name ?? '',
                    l.currency,
                    formatMoney(l.debit),
                    formatMoney(l.credit),
                    formatMoney(l.netBalance),
                  ],
                )
                .toList(),
          ),
          pw.SizedBox(height: 16),
          pw.Header(level: 1, child: pw.Text('Per-currency totals')),
          pw.TableHelper.fromTextArray(
            headerStyle: pw.TextStyle(fontWeight: pw.FontWeight.bold),
            headers: const [
              'Currency',
              'TotalDebits',
              'TotalCredits',
              'IsBalanced',
            ],
            data: report.totals
                .map(
                  (t) => [
                    t.currency,
                    formatMoney(t.totalDebits),
                    formatMoney(t.totalCredits),
                    t.isBalanced ? 'BALANCED' : 'UNBALANCED',
                  ],
                )
                .toList(),
          ),
        ],
      ),
    );
    return doc.save();
  }

  static Future<Uint8List> _statementPdf(AccountStatementReport report) async {
    final doc = pw.Document();
    doc.addPage(
      pw.MultiPage(
        pageFormat: PdfPageFormat.a4,
        build: (ctx) => [
          pw.Header(
            level: 0,
            child: pw.Text('Account Statement — ${report.accountId}'),
          ),
          pw.Paragraph(
            text:
                'Currency: ${report.currency} • '
                'Opening ${formatMoney(report.openingBalance)} • '
                'Closing ${formatMoney(report.closingBalance)} • '
                'Debits ${formatMoney(report.totalDebits)} • '
                'Credits ${formatMoney(report.totalCredits)}.',
          ),
          pw.TableHelper.fromTextArray(
            headerStyle: pw.TextStyle(fontWeight: pw.FontWeight.bold),
            headers: const [
              'Transacted',
              'Transaction',
              'DR/CR',
              'Amount',
              'Running',
            ],
            data: report.entries
                .map(
                  (e) => [
                    e.transactedAt,
                    e.transactionId,
                    e.credit ? 'CR' : 'DR',
                    formatMoney(e.amount),
                    formatMoney(e.runningBalance),
                  ],
                )
                .toList(),
          ),
        ],
      ),
    );
    return doc.save();
  }

  // ---------- helpers ----------

  static String _stamp() {
    final now = DateTime.now().toUtc();
    final y = now.year.toString().padLeft(4, '0');
    final m = now.month.toString().padLeft(2, '0');
    final d = now.day.toString().padLeft(2, '0');
    final hh = now.hour.toString().padLeft(2, '0');
    final mm = now.minute.toString().padLeft(2, '0');
    final ss = now.second.toString().padLeft(2, '0');
    return '$y$m${d}_$hh$mm$ss';
  }

  static Future<void> _saveBytes({
    required String name,
    required MimeType mimeType,
    required String fileExtension,
    required Uint8List bytes,
  }) async {
    // file_saver picks the right channel automatically: anchor download
    // on web, share sheet on iOS/Android, file picker on desktop.
    await FileSaver.instance.saveFile(
      name: name,
      bytes: bytes,
      mimeType: mimeType,
      fileExtension: fileExtension,
    );
  }
}
