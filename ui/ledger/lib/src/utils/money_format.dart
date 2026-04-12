import 'package:fixnum/fixnum.dart';

/// Formats a protobuf Money message for display, avoiding cross-package
/// type conflicts between antinvestor_api_common and antinvestor_api_ledger.
String fmtMoney(dynamic money) {
  if (money == null) return '\u2014';
  final int units = (money.units as dynamic).toInt();
  final int nanos = money.nanos as int;
  final String currency = money.currencyCode as String;
  if (units == 0 && nanos == 0) return '\u2014';
  final cents = nanos ~/ 10000000;
  final formatted = '$units.${cents.toString().padLeft(2, '0')}';
  if (currency.isEmpty) return formatted;
  return '$currency $formatted';
}

/// Sets Money fields on an existing protobuf Money instance (from any package).
void setMoneyFields(dynamic money, String amount, String currencyCode) {
  money.currencyCode = currencyCode;

  final cleaned = amount.trim();
  if (cleaned.isEmpty) return;

  final sanitized = cleaned.replaceAll(RegExp(r'[^\d.\-]'), '');
  if (sanitized.isEmpty) return;

  try {
    final parts = sanitized.split('.');
    money.units = Int64.parseInt(parts[0].isEmpty ? '0' : parts[0]);
    if (parts.length > 1) {
      final fracStr = parts[1].padRight(9, '0').substring(0, 9);
      money.nanos = int.parse(fracStr);
    }
  } catch (_) {
    // Leave zero-value on any parse failure
  }
}
