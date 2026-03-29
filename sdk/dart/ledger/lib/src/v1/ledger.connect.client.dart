//
//  Generated code. Do not modify.
//  source: v1/ledger.proto
//

import "package:connectrpc/connect.dart" as connect;
import "ledger.pb.dart" as v1ledger;
import "../common/v1/common.pb.dart" as commonv1common;
import "ledger.connect.spec.dart" as specs;

/// LedgerService provides double-entry bookkeeping and financial accounting.
/// All RPCs require authentication via Bearer token.
extension type LedgerServiceClient (connect.Transport _transport) {
  /// SearchLedgers finds ledgers in the chart of accounts.
  /// Supports filtering by type, parent, and custom properties.
  Stream<v1ledger.SearchLedgersResponse> searchLedgers(
    commonv1common.SearchRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).server(
      specs.LedgerService.searchLedgers,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateLedger creates a new ledger in the chart of accounts.
  /// Ledgers can be hierarchical with parent-child relationships.
  Future<v1ledger.CreateLedgerResponse> createLedger(
    v1ledger.CreateLedgerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LedgerService.createLedger,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateLedger updates an existing ledger's metadata.
  /// The ledger type and reference cannot be changed.
  Future<v1ledger.UpdateLedgerResponse> updateLedger(
    v1ledger.UpdateLedgerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LedgerService.updateLedger,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SearchAccounts finds accounts matching specified criteria.
  /// Supports filtering by ledger, balance range, and custom properties.
  Stream<v1ledger.SearchAccountsResponse> searchAccounts(
    commonv1common.SearchRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).server(
      specs.LedgerService.searchAccounts,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateAccount creates a new account within a ledger.
  /// Each account tracks balances and transaction history.
  Future<v1ledger.CreateAccountResponse> createAccount(
    v1ledger.CreateAccountRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LedgerService.createAccount,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateAccount updates an existing account's metadata.
  /// Balances are updated through transactions, not directly.
  Future<v1ledger.UpdateAccountResponse> updateAccount(
    v1ledger.UpdateAccountRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LedgerService.updateAccount,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SearchTransactions finds transactions matching specified criteria.
  /// Supports filtering by date range, account, currency, and status.
  Stream<v1ledger.SearchTransactionsResponse> searchTransactions(
    commonv1common.SearchRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).server(
      specs.LedgerService.searchTransactions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateTransaction creates a new double-entry transaction.
  /// All entries must be balanced (sum of debits = sum of credits).
  Future<v1ledger.CreateTransactionResponse> createTransaction(
    v1ledger.CreateTransactionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LedgerService.createTransaction,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ReverseTransaction reverses a transaction by creating offsetting entries.
  /// Creates a new REVERSAL transaction that negates the original.
  Future<v1ledger.ReverseTransactionResponse> reverseTransaction(
    v1ledger.ReverseTransactionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LedgerService.reverseTransaction,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateTransaction updates a transaction's metadata.
  /// Entries and amounts cannot be changed after creation.
  Future<v1ledger.UpdateTransactionResponse> updateTransaction(
    v1ledger.UpdateTransactionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LedgerService.updateTransaction,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SearchTransactionEntries finds individual transaction entries.
  /// Useful for account statement generation and reconciliation.
  Stream<v1ledger.SearchTransactionEntriesResponse> searchTransactionEntries(
    commonv1common.SearchRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).server(
      specs.LedgerService.searchTransactionEntries,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
