//
//  Generated code. Do not modify.
//  source: v1/ledger.proto
//

import "package:connectrpc/connect.dart" as connect;
import "../common/v1/common.pb.dart" as commonv1common;
import "ledger.pb.dart" as v1ledger;

/// LedgerService provides double-entry bookkeeping and financial accounting.
/// All RPCs require authentication via Bearer token.
abstract final class LedgerService {
  /// Fully-qualified name of the LedgerService service.
  static const name = 'ledger.v1.LedgerService';

  /// SearchLedgers finds ledgers in the chart of accounts.
  /// Supports filtering by type, parent, and custom properties.
  static const searchLedgers = connect.Spec(
    '/$name/SearchLedgers',
    connect.StreamType.server,
    commonv1common.SearchRequest.new,
    v1ledger.SearchLedgersResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// CreateLedger creates a new ledger in the chart of accounts.
  /// Ledgers can be hierarchical with parent-child relationships.
  static const createLedger = connect.Spec(
    '/$name/CreateLedger',
    connect.StreamType.unary,
    v1ledger.CreateLedgerRequest.new,
    v1ledger.CreateLedgerResponse.new,
  );

  /// UpdateLedger updates an existing ledger's metadata.
  /// The ledger type and reference cannot be changed.
  static const updateLedger = connect.Spec(
    '/$name/UpdateLedger',
    connect.StreamType.unary,
    v1ledger.UpdateLedgerRequest.new,
    v1ledger.UpdateLedgerResponse.new,
  );

  /// SearchAccounts finds accounts matching specified criteria.
  /// Supports filtering by ledger, balance range, and custom properties.
  static const searchAccounts = connect.Spec(
    '/$name/SearchAccounts',
    connect.StreamType.server,
    commonv1common.SearchRequest.new,
    v1ledger.SearchAccountsResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// CreateAccount creates a new account within a ledger.
  /// Each account tracks balances and transaction history.
  static const createAccount = connect.Spec(
    '/$name/CreateAccount',
    connect.StreamType.unary,
    v1ledger.CreateAccountRequest.new,
    v1ledger.CreateAccountResponse.new,
  );

  /// UpdateAccount updates an existing account's metadata.
  /// Balances are updated through transactions, not directly.
  static const updateAccount = connect.Spec(
    '/$name/UpdateAccount',
    connect.StreamType.unary,
    v1ledger.UpdateAccountRequest.new,
    v1ledger.UpdateAccountResponse.new,
  );

  /// SearchTransactions finds transactions matching specified criteria.
  /// Supports filtering by date range, account, currency, and status.
  static const searchTransactions = connect.Spec(
    '/$name/SearchTransactions',
    connect.StreamType.server,
    commonv1common.SearchRequest.new,
    v1ledger.SearchTransactionsResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// CreateTransaction creates a new double-entry transaction.
  /// All entries must be balanced (sum of debits = sum of credits).
  static const createTransaction = connect.Spec(
    '/$name/CreateTransaction',
    connect.StreamType.unary,
    v1ledger.CreateTransactionRequest.new,
    v1ledger.CreateTransactionResponse.new,
  );

  /// ReverseTransaction reverses a transaction by creating offsetting entries.
  /// Creates a new REVERSAL transaction that negates the original.
  static const reverseTransaction = connect.Spec(
    '/$name/ReverseTransaction',
    connect.StreamType.unary,
    v1ledger.ReverseTransactionRequest.new,
    v1ledger.ReverseTransactionResponse.new,
  );

  /// UpdateTransaction updates a transaction's metadata.
  /// Entries and amounts cannot be changed after creation.
  static const updateTransaction = connect.Spec(
    '/$name/UpdateTransaction',
    connect.StreamType.unary,
    v1ledger.UpdateTransactionRequest.new,
    v1ledger.UpdateTransactionResponse.new,
  );

  /// SearchTransactionEntries finds individual transaction entries.
  /// Useful for account statement generation and reconciliation.
  static const searchTransactionEntries = connect.Spec(
    '/$name/SearchTransactionEntries',
    connect.StreamType.server,
    commonv1common.SearchRequest.new,
    v1ledger.SearchTransactionEntriesResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// VoidTransaction transitions a draft or pending transaction to VOIDED.
  /// Posted activity cannot be voided — reverse it instead so the books
  /// carry the offset and the audit trail.
  static const voidTransaction = connect.Spec(
    '/$name/VoidTransaction',
    connect.StreamType.unary,
    v1ledger.VoidTransactionRequest.new,
    v1ledger.VoidTransactionResponse.new,
  );

  /// MarkTransactionFailed transitions a pending transaction to FAILED for
  /// upstream rejections (e.g. payment provider declined the webhook).
  static const markTransactionFailed = connect.Spec(
    '/$name/MarkTransactionFailed',
    connect.StreamType.unary,
    v1ledger.MarkTransactionFailedRequest.new,
    v1ledger.MarkTransactionFailedResponse.new,
  );

  /// CreateBook creates a new accounting scope. Each book balances
  /// independently; cross-book postings are rejected.
  static const createBook = connect.Spec(
    '/$name/CreateBook',
    connect.StreamType.unary,
    v1ledger.CreateBookRequest.new,
    v1ledger.CreateBookResponse.new,
  );

  /// GetBook fetches a single book by id within the caller's tenancy.
  static const getBook = connect.Spec(
    '/$name/GetBook',
    connect.StreamType.unary,
    v1ledger.GetBookRequest.new,
    v1ledger.GetBookResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// ListBooksByType returns all books of a given conventional type.
  static const listBooksByType = connect.Spec(
    '/$name/ListBooksByType',
    connect.StreamType.unary,
    v1ledger.ListBooksByTypeRequest.new,
    v1ledger.ListBooksByTypeResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// GetTrialBalance returns the per-account debit/credit totals plus per-
  /// currency grand totals with the textbook integrity check (is_balanced).
  static const getTrialBalance = connect.Spec(
    '/$name/GetTrialBalance',
    connect.StreamType.unary,
    v1ledger.GetTrialBalanceRequest.new,
    v1ledger.GetTrialBalanceResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// GetAccountStatement returns a customer-facing account ledger for a
  /// period with opening balance, running balance per entry, and totals.
  static const getAccountStatement = connect.Spec(
    '/$name/GetAccountStatement',
    connect.StreamType.unary,
    v1ledger.GetAccountStatementRequest.new,
    v1ledger.GetAccountStatementResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );
}
