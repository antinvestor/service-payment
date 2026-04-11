import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:antinvestor_ui_core/api/api_base.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const _ledgerUrl = String.fromEnvironment(
  'LEDGER_URL',
  defaultValue: 'https://api.antinvestor.com/ledger',
);

final ledgerTransportProvider = Provider<Transport>((ref) {
  final tokenProvider = ref.watch(authTokenProviderProvider);
  return createTransport(tokenProvider, baseUrl: _ledgerUrl);
});

final ledgerServiceClientProvider = Provider<LedgerServiceClient>((ref) {
  final transport = ref.watch(ledgerTransportProvider);
  return LedgerServiceClient(transport);
});
