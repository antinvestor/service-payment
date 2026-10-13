// This is a generated file - do not edit.
//
// Generated from v1/ledger.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart' as $0;

import '../common/v1/common.pb.dart' as $2;
import '../common/v1/money.pb.dart' as $1;
import 'ledger.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'ledger.pbenum.dart';

/// Ledger represents a category in the chart of accounts.
/// Ledgers can be hierarchical with parent-child relationships.
class Ledger extends $pb.GeneratedMessage {
  factory Ledger({
    $core.String? id,
    LedgerType? type,
    $core.String? parent,
    $0.Struct? data,
    $core.String? bookId,
  }) {
    final result = Ledger._();
    if (id != null) result.id = id;
    if (type != null) result.type = type;
    if (parent != null) result.parent = parent;
    if (data != null) result.data = data;
    if (bookId != null) result.bookId = bookId;
    return result;
  }

  Ledger._();

  factory Ledger.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Ledger()..mergeFromBuffer(data, registry);
  factory Ledger.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Ledger()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Ledger',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: Ledger.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aE<LedgerType>(2, _omitFieldNames ? '' : 'type',
        enumValues: LedgerType.values)
    ..aOS(3, _omitFieldNames ? '' : 'parent')
    ..aOM<$0.Struct>(4, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'bookId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Ledger clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Ledger copyWith(void Function(Ledger) updates) =>
      super.copyWith((message) => updates(message as Ledger)) as Ledger;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Ledger() / Ledger.new instead')
  static Ledger create() => Ledger._();
  static $pb.GeneratedMessage $_createMessage() => Ledger._();
  @$core.override
  Ledger createEmptyInstance() => Ledger._();
  @$core.pragma('dart2js:noInline')
  static Ledger getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Ledger>(Ledger.$_createMessage);
  static Ledger? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  LedgerType get type => $_getN(1);
  @$pb.TagNumber(2)
  set type(LedgerType value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get parent => $_getSZ(2);
  @$pb.TagNumber(3)
  set parent($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasParent() => $_has(2);
  @$pb.TagNumber(3)
  void clearParent() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($0.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get bookId => $_getSZ(4);
  @$pb.TagNumber(5)
  set bookId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasBookId() => $_has(4);
  @$pb.TagNumber(5)
  void clearBookId() => $_clearField(5);
}

/// Account represents a specific account within a ledger.
/// Tracks balances and supports multi-currency operations.
class Account extends $pb.GeneratedMessage {
  factory Account({
    $core.String? id,
    $core.String? ledger,
    $1.Money? balance,
    $0.Struct? data,
    $1.Money? unclearedBalance,
    $1.Money? reservedBalance,
    AccountType? accountType,
    NormalBalance? normalBalance,
    $core.String? bookId,
  }) {
    final result = Account._();
    if (id != null) result.id = id;
    if (ledger != null) result.ledger = ledger;
    if (balance != null) result.balance = balance;
    if (data != null) result.data = data;
    if (unclearedBalance != null) result.unclearedBalance = unclearedBalance;
    if (reservedBalance != null) result.reservedBalance = reservedBalance;
    if (accountType != null) result.accountType = accountType;
    if (normalBalance != null) result.normalBalance = normalBalance;
    if (bookId != null) result.bookId = bookId;
    return result;
  }

  Account._();

  factory Account.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Account()..mergeFromBuffer(data, registry);
  factory Account.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Account()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Account',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: Account.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'ledger')
    ..aOM<$1.Money>(4, _omitFieldNames ? '' : 'balance',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$0.Struct>(5, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..aOM<$1.Money>(6, _omitFieldNames ? '' : 'unclearedBalance',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$1.Money>(7, _omitFieldNames ? '' : 'reservedBalance',
        subBuilder: $1.Money.$_createMessage)
    ..aE<AccountType>(8, _omitFieldNames ? '' : 'accountType',
        enumValues: AccountType.values)
    ..aE<NormalBalance>(9, _omitFieldNames ? '' : 'normalBalance',
        enumValues: NormalBalance.values)
    ..aOS(10, _omitFieldNames ? '' : 'bookId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Account clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Account copyWith(void Function(Account) updates) =>
      super.copyWith((message) => updates(message as Account)) as Account;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Account() / Account.new instead')
  static Account create() => Account._();
  static $pb.GeneratedMessage $_createMessage() => Account._();
  @$core.override
  Account createEmptyInstance() => Account._();
  @$core.pragma('dart2js:noInline')
  static Account getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Account>(Account.$_createMessage);
  static Account? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(3)
  $core.String get ledger => $_getSZ(1);
  @$pb.TagNumber(3)
  set ledger($core.String value) => $_setString(1, value);
  @$pb.TagNumber(3)
  $core.bool hasLedger() => $_has(1);
  @$pb.TagNumber(3)
  void clearLedger() => $_clearField(3);

  @$pb.TagNumber(4)
  $1.Money get balance => $_getN(2);
  @$pb.TagNumber(4)
  set balance($1.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasBalance() => $_has(2);
  @$pb.TagNumber(4)
  void clearBalance() => $_clearField(4);
  @$pb.TagNumber(4)
  $1.Money ensureBalance() => $_ensure(2);

  @$pb.TagNumber(5)
  $0.Struct get data => $_getN(3);
  @$pb.TagNumber(5)
  set data($0.Struct value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(5)
  void clearData() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(6)
  $1.Money get unclearedBalance => $_getN(4);
  @$pb.TagNumber(6)
  set unclearedBalance($1.Money value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUnclearedBalance() => $_has(4);
  @$pb.TagNumber(6)
  void clearUnclearedBalance() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Money ensureUnclearedBalance() => $_ensure(4);

  @$pb.TagNumber(7)
  $1.Money get reservedBalance => $_getN(5);
  @$pb.TagNumber(7)
  set reservedBalance($1.Money value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasReservedBalance() => $_has(5);
  @$pb.TagNumber(7)
  void clearReservedBalance() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Money ensureReservedBalance() => $_ensure(5);

  @$pb.TagNumber(8)
  AccountType get accountType => $_getN(6);
  @$pb.TagNumber(8)
  set accountType(AccountType value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasAccountType() => $_has(6);
  @$pb.TagNumber(8)
  void clearAccountType() => $_clearField(8);

  @$pb.TagNumber(9)
  NormalBalance get normalBalance => $_getN(7);
  @$pb.TagNumber(9)
  set normalBalance(NormalBalance value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasNormalBalance() => $_has(7);
  @$pb.TagNumber(9)
  void clearNormalBalance() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get bookId => $_getSZ(8);
  @$pb.TagNumber(10)
  set bookId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(10)
  $core.bool hasBookId() => $_has(8);
  @$pb.TagNumber(10)
  void clearBookId() => $_clearField(10);
}

/// TransactionEntry represents one side of a double-entry transaction.
/// Each transaction must have at least two entries with balanced debits and credits.
class TransactionEntry extends $pb.GeneratedMessage {
  factory TransactionEntry({
    $core.String? id,
    $core.String? accountId,
    $core.String? transactionId,
    $core.String? transactedAt,
    $1.Money? amount,
    $core.bool? credit,
    $1.Money? accBalance,
    $core.String? clearedAt,
  }) {
    final result = TransactionEntry._();
    if (id != null) result.id = id;
    if (accountId != null) result.accountId = accountId;
    if (transactionId != null) result.transactionId = transactionId;
    if (transactedAt != null) result.transactedAt = transactedAt;
    if (amount != null) result.amount = amount;
    if (credit != null) result.credit = credit;
    if (accBalance != null) result.accBalance = accBalance;
    if (clearedAt != null) result.clearedAt = clearedAt;
    return result;
  }

  TransactionEntry._();

  factory TransactionEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TransactionEntry()..mergeFromBuffer(data, registry);
  factory TransactionEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TransactionEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TransactionEntry',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: TransactionEntry.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'accountId')
    ..aOS(4, _omitFieldNames ? '' : 'transactionId')
    ..aOS(5, _omitFieldNames ? '' : 'transactedAt')
    ..aOM<$1.Money>(7, _omitFieldNames ? '' : 'amount',
        subBuilder: $1.Money.$_createMessage)
    ..aOB(8, _omitFieldNames ? '' : 'credit')
    ..aOM<$1.Money>(9, _omitFieldNames ? '' : 'accBalance',
        subBuilder: $1.Money.$_createMessage)
    ..aOS(11, _omitFieldNames ? '' : 'clearedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TransactionEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TransactionEntry copyWith(void Function(TransactionEntry) updates) =>
      super.copyWith((message) => updates(message as TransactionEntry))
          as TransactionEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TransactionEntry() / TransactionEntry.new instead')
  static TransactionEntry create() => TransactionEntry._();
  static $pb.GeneratedMessage $_createMessage() => TransactionEntry._();
  @$core.override
  TransactionEntry createEmptyInstance() => TransactionEntry._();
  @$core.pragma('dart2js:noInline')
  static TransactionEntry getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TransactionEntry>(
          TransactionEntry.$_createMessage);
  static TransactionEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(3)
  $core.String get accountId => $_getSZ(1);
  @$pb.TagNumber(3)
  set accountId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(3)
  $core.bool hasAccountId() => $_has(1);
  @$pb.TagNumber(3)
  void clearAccountId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get transactionId => $_getSZ(2);
  @$pb.TagNumber(4)
  set transactionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(4)
  $core.bool hasTransactionId() => $_has(2);
  @$pb.TagNumber(4)
  void clearTransactionId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get transactedAt => $_getSZ(3);
  @$pb.TagNumber(5)
  set transactedAt($core.String value) => $_setString(3, value);
  @$pb.TagNumber(5)
  $core.bool hasTransactedAt() => $_has(3);
  @$pb.TagNumber(5)
  void clearTransactedAt() => $_clearField(5);

  /// Amount for this entry. On create, supply a positive amount; the side is
  /// carried by `credit`. On read, the amount is signed by its effect on the
  /// account's balance: positive when the entry's side matches the account's
  /// normal balance (debit on ASSET/EXPENSE, credit on LIABILITY/INCOME/
  /// CAPITAL), negative when it opposes it. So a credit to an asset account
  /// reads back with credit = true and a negative amount.
  @$pb.TagNumber(7)
  $1.Money get amount => $_getN(4);
  @$pb.TagNumber(7)
  set amount($1.Money value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasAmount() => $_has(4);
  @$pb.TagNumber(7)
  void clearAmount() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Money ensureAmount() => $_ensure(4);

  @$pb.TagNumber(8)
  $core.bool get credit => $_getBF(5);
  @$pb.TagNumber(8)
  set credit($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(8)
  $core.bool hasCredit() => $_has(5);
  @$pb.TagNumber(8)
  void clearCredit() => $_clearField(8);

  @$pb.TagNumber(9)
  $1.Money get accBalance => $_getN(6);
  @$pb.TagNumber(9)
  set accBalance($1.Money value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasAccBalance() => $_has(6);
  @$pb.TagNumber(9)
  void clearAccBalance() => $_clearField(9);
  @$pb.TagNumber(9)
  $1.Money ensureAccBalance() => $_ensure(6);

  @$pb.TagNumber(11)
  $core.String get clearedAt => $_getSZ(7);
  @$pb.TagNumber(11)
  set clearedAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(11)
  $core.bool hasClearedAt() => $_has(7);
  @$pb.TagNumber(11)
  void clearClearedAt() => $_clearField(11);
}

/// Transaction represents a complete double-entry transaction.
/// Must contain balanced entries (sum of debits = sum of credits).
class Transaction extends $pb.GeneratedMessage {
  factory Transaction({
    $core.String? id,
    $core.String? currencyCode,
    $core.String? transactedAt,
    $0.Struct? data,
    $core.Iterable<TransactionEntry>? entries,
    $core.bool? cleared,
    TransactionType? type,
    TransactionStatus? status,
    $core.String? idempotencyKey,
    $core.String? externalRef,
    $core.String? source,
    $core.String? bookId,
    $core.String? postedAt,
    $core.String? voidedAt,
    $core.String? reversedTransactionId,
  }) {
    final result = Transaction._();
    if (id != null) result.id = id;
    if (currencyCode != null) result.currencyCode = currencyCode;
    if (transactedAt != null) result.transactedAt = transactedAt;
    if (data != null) result.data = data;
    if (entries != null) result.entries.addAll(entries);
    if (cleared != null) result.cleared = cleared;
    if (type != null) result.type = type;
    if (status != null) result.status = status;
    if (idempotencyKey != null) result.idempotencyKey = idempotencyKey;
    if (externalRef != null) result.externalRef = externalRef;
    if (source != null) result.source = source;
    if (bookId != null) result.bookId = bookId;
    if (postedAt != null) result.postedAt = postedAt;
    if (voidedAt != null) result.voidedAt = voidedAt;
    if (reversedTransactionId != null)
      result.reversedTransactionId = reversedTransactionId;
    return result;
  }

  Transaction._();

  factory Transaction.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Transaction()..mergeFromBuffer(data, registry);
  factory Transaction.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Transaction()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Transaction',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: Transaction.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'currencyCode')
    ..aOS(3, _omitFieldNames ? '' : 'transactedAt')
    ..aOM<$0.Struct>(4, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..pPM<TransactionEntry>(5, _omitFieldNames ? '' : 'entries',
        subBuilder: TransactionEntry.$_createMessage)
    ..aOB(6, _omitFieldNames ? '' : 'cleared')
    ..aE<TransactionType>(7, _omitFieldNames ? '' : 'type',
        enumValues: TransactionType.values)
    ..aE<TransactionStatus>(8, _omitFieldNames ? '' : 'status',
        enumValues: TransactionStatus.values)
    ..aOS(9, _omitFieldNames ? '' : 'idempotencyKey')
    ..aOS(10, _omitFieldNames ? '' : 'externalRef')
    ..aOS(11, _omitFieldNames ? '' : 'source')
    ..aOS(12, _omitFieldNames ? '' : 'bookId')
    ..aOS(13, _omitFieldNames ? '' : 'postedAt')
    ..aOS(14, _omitFieldNames ? '' : 'voidedAt')
    ..aOS(15, _omitFieldNames ? '' : 'reversedTransactionId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Transaction clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Transaction copyWith(void Function(Transaction) updates) =>
      super.copyWith((message) => updates(message as Transaction))
          as Transaction;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Transaction() / Transaction.new instead')
  static Transaction create() => Transaction._();
  static $pb.GeneratedMessage $_createMessage() => Transaction._();
  @$core.override
  Transaction createEmptyInstance() => Transaction._();
  @$core.pragma('dart2js:noInline')
  static Transaction getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Transaction>(
          Transaction.$_createMessage);
  static Transaction? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Asset code shared by every entry and their accounts: an ISO 4217 code
  /// (USD, UGX) or a token symbol (USDC), 3-10 uppercase letters or digits.
  @$pb.TagNumber(2)
  $core.String get currencyCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set currencyCode($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCurrencyCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrencyCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get transactedAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set transactedAt($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTransactedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransactedAt() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($0.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(5)
  $pb.PbList<TransactionEntry> get entries => $_getList(4);

  @$pb.TagNumber(6)
  $core.bool get cleared => $_getBF(5);
  @$pb.TagNumber(6)
  set cleared($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCleared() => $_has(5);
  @$pb.TagNumber(6)
  void clearCleared() => $_clearField(6);

  @$pb.TagNumber(7)
  TransactionType get type => $_getN(6);
  @$pb.TagNumber(7)
  set type(TransactionType value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasType() => $_has(6);
  @$pb.TagNumber(7)
  void clearType() => $_clearField(7);

  @$pb.TagNumber(8)
  TransactionStatus get status => $_getN(7);
  @$pb.TagNumber(8)
  set status(TransactionStatus value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasStatus() => $_has(7);
  @$pb.TagNumber(8)
  void clearStatus() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get idempotencyKey => $_getSZ(8);
  @$pb.TagNumber(9)
  set idempotencyKey($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasIdempotencyKey() => $_has(8);
  @$pb.TagNumber(9)
  void clearIdempotencyKey() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get externalRef => $_getSZ(9);
  @$pb.TagNumber(10)
  set externalRef($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasExternalRef() => $_has(9);
  @$pb.TagNumber(10)
  void clearExternalRef() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get source => $_getSZ(10);
  @$pb.TagNumber(11)
  set source($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasSource() => $_has(10);
  @$pb.TagNumber(11)
  void clearSource() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get bookId => $_getSZ(11);
  @$pb.TagNumber(12)
  set bookId($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasBookId() => $_has(11);
  @$pb.TagNumber(12)
  void clearBookId() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.String get postedAt => $_getSZ(12);
  @$pb.TagNumber(13)
  set postedAt($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasPostedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearPostedAt() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.String get voidedAt => $_getSZ(13);
  @$pb.TagNumber(14)
  set voidedAt($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasVoidedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearVoidedAt() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.String get reversedTransactionId => $_getSZ(14);
  @$pb.TagNumber(15)
  set reversedTransactionId($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasReversedTransactionId() => $_has(14);
  @$pb.TagNumber(15)
  void clearReversedTransactionId() => $_clearField(15);
}

/// Book is an independent accounting scope — one entity's complete set of
/// financial records. Examples by convention: "platform" (the company's own
/// book), "group" (a savings group / chama / SACCO / branch), "customer"
/// (an individual member), "merchant", "agent". Type is a free-form string
/// so product domains can grow new entity classifications without a
/// migration. ParentID supports hierarchy (organisation → groups → members)
/// for consolidated reporting; cross-book posting remains strict regardless.
class Book extends $pb.GeneratedMessage {
  factory Book({
    $core.String? id,
    $core.String? name,
    $core.String? type,
    $core.String? parentId,
    $core.String? currency,
    $0.Struct? data,
  }) {
    final result = Book._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (type != null) result.type = type;
    if (parentId != null) result.parentId = parentId;
    if (currency != null) result.currency = currency;
    if (data != null) result.data = data;
    return result;
  }

  Book._();

  factory Book.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Book()..mergeFromBuffer(data, registry);
  factory Book.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Book()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Book',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: Book.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOS(4, _omitFieldNames ? '' : 'parentId')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$0.Struct>(6, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Book clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Book copyWith(void Function(Book) updates) =>
      super.copyWith((message) => updates(message as Book)) as Book;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Book() / Book.new instead')
  static Book create() => Book._();
  static $pb.GeneratedMessage $_createMessage() => Book._();
  @$core.override
  Book createEmptyInstance() => Book._();
  @$core.pragma('dart2js:noInline')
  static Book getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Book>(Book.$_createMessage);
  static Book? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get parentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set parentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasParentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearParentId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Struct get data => $_getN(5);
  @$pb.TagNumber(6)
  set data($0.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasData() => $_has(5);
  @$pb.TagNumber(6)
  void clearData() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Struct ensureData() => $_ensure(5);
}

/// TrialBalanceLine is one row of the trial balance — a single account's
/// raw debit and credit totals plus the DEADCLIC-signed net balance.
class TrialBalanceLine extends $pb.GeneratedMessage {
  factory TrialBalanceLine({
    $core.String? accountId,
    $core.String? ledgerId,
    LedgerType? ledgerType,
    $core.String? currency,
    $1.Money? totalDebits,
    $1.Money? totalCredits,
    $1.Money? netBalance,
  }) {
    final result = TrialBalanceLine._();
    if (accountId != null) result.accountId = accountId;
    if (ledgerId != null) result.ledgerId = ledgerId;
    if (ledgerType != null) result.ledgerType = ledgerType;
    if (currency != null) result.currency = currency;
    if (totalDebits != null) result.totalDebits = totalDebits;
    if (totalCredits != null) result.totalCredits = totalCredits;
    if (netBalance != null) result.netBalance = netBalance;
    return result;
  }

  TrialBalanceLine._();

  factory TrialBalanceLine.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TrialBalanceLine()..mergeFromBuffer(data, registry);
  factory TrialBalanceLine.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TrialBalanceLine()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TrialBalanceLine',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: TrialBalanceLine.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'ledgerId')
    ..aE<LedgerType>(3, _omitFieldNames ? '' : 'ledgerType',
        enumValues: LedgerType.values)
    ..aOS(4, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Money>(5, _omitFieldNames ? '' : 'totalDebits',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$1.Money>(6, _omitFieldNames ? '' : 'totalCredits',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$1.Money>(7, _omitFieldNames ? '' : 'netBalance',
        subBuilder: $1.Money.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TrialBalanceLine clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TrialBalanceLine copyWith(void Function(TrialBalanceLine) updates) =>
      super.copyWith((message) => updates(message as TrialBalanceLine))
          as TrialBalanceLine;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TrialBalanceLine() / TrialBalanceLine.new instead')
  static TrialBalanceLine create() => TrialBalanceLine._();
  static $pb.GeneratedMessage $_createMessage() => TrialBalanceLine._();
  @$core.override
  TrialBalanceLine createEmptyInstance() => TrialBalanceLine._();
  @$core.pragma('dart2js:noInline')
  static TrialBalanceLine getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TrialBalanceLine>(
          TrialBalanceLine.$_createMessage);
  static TrialBalanceLine? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get ledgerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set ledgerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearLedgerId() => $_clearField(2);

  @$pb.TagNumber(3)
  LedgerType get ledgerType => $_getN(2);
  @$pb.TagNumber(3)
  set ledgerType(LedgerType value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasLedgerType() => $_has(2);
  @$pb.TagNumber(3)
  void clearLedgerType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(3);
  @$pb.TagNumber(4)
  set currency($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(3);
  @$pb.TagNumber(4)
  void clearCurrency() => $_clearField(4);

  @$pb.TagNumber(5)
  $1.Money get totalDebits => $_getN(4);
  @$pb.TagNumber(5)
  set totalDebits($1.Money value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasTotalDebits() => $_has(4);
  @$pb.TagNumber(5)
  void clearTotalDebits() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Money ensureTotalDebits() => $_ensure(4);

  @$pb.TagNumber(6)
  $1.Money get totalCredits => $_getN(5);
  @$pb.TagNumber(6)
  set totalCredits($1.Money value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasTotalCredits() => $_has(5);
  @$pb.TagNumber(6)
  void clearTotalCredits() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Money ensureTotalCredits() => $_ensure(5);

  @$pb.TagNumber(7)
  $1.Money get netBalance => $_getN(6);
  @$pb.TagNumber(7)
  set netBalance($1.Money value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasNetBalance() => $_has(6);
  @$pb.TagNumber(7)
  void clearNetBalance() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Money ensureNetBalance() => $_ensure(6);
}

/// TrialBalanceTotal aggregates one currency's debits and credits across
/// every line in the report. is_balanced is the textbook integrity check.
class TrialBalanceTotal extends $pb.GeneratedMessage {
  factory TrialBalanceTotal({
    $core.String? currency,
    $1.Money? totalDebits,
    $1.Money? totalCredits,
    $core.bool? isBalanced,
  }) {
    final result = TrialBalanceTotal._();
    if (currency != null) result.currency = currency;
    if (totalDebits != null) result.totalDebits = totalDebits;
    if (totalCredits != null) result.totalCredits = totalCredits;
    if (isBalanced != null) result.isBalanced = isBalanced;
    return result;
  }

  TrialBalanceTotal._();

  factory TrialBalanceTotal.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TrialBalanceTotal()..mergeFromBuffer(data, registry);
  factory TrialBalanceTotal.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TrialBalanceTotal()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TrialBalanceTotal',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: TrialBalanceTotal.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Money>(2, _omitFieldNames ? '' : 'totalDebits',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$1.Money>(3, _omitFieldNames ? '' : 'totalCredits',
        subBuilder: $1.Money.$_createMessage)
    ..aOB(4, _omitFieldNames ? '' : 'isBalanced')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TrialBalanceTotal clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TrialBalanceTotal copyWith(void Function(TrialBalanceTotal) updates) =>
      super.copyWith((message) => updates(message as TrialBalanceTotal))
          as TrialBalanceTotal;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TrialBalanceTotal() / TrialBalanceTotal.new instead')
  static TrialBalanceTotal create() => TrialBalanceTotal._();
  static $pb.GeneratedMessage $_createMessage() => TrialBalanceTotal._();
  @$core.override
  TrialBalanceTotal createEmptyInstance() => TrialBalanceTotal._();
  @$core.pragma('dart2js:noInline')
  static TrialBalanceTotal getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TrialBalanceTotal>(
          TrialBalanceTotal.$_createMessage);
  static TrialBalanceTotal? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get currency => $_getSZ(0);
  @$pb.TagNumber(1)
  set currency($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCurrency() => $_has(0);
  @$pb.TagNumber(1)
  void clearCurrency() => $_clearField(1);

  @$pb.TagNumber(2)
  $1.Money get totalDebits => $_getN(1);
  @$pb.TagNumber(2)
  set totalDebits($1.Money value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasTotalDebits() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotalDebits() => $_clearField(2);
  @$pb.TagNumber(2)
  $1.Money ensureTotalDebits() => $_ensure(1);

  @$pb.TagNumber(3)
  $1.Money get totalCredits => $_getN(2);
  @$pb.TagNumber(3)
  set totalCredits($1.Money value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasTotalCredits() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotalCredits() => $_clearField(3);
  @$pb.TagNumber(3)
  $1.Money ensureTotalCredits() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.bool get isBalanced => $_getBF(3);
  @$pb.TagNumber(4)
  set isBalanced($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIsBalanced() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsBalanced() => $_clearField(4);
}

/// StatementEntry is one row on an account statement, hydrated with its
/// parent transaction's metadata and the running balance after the entry.
class StatementEntry extends $pb.GeneratedMessage {
  factory StatementEntry({
    $core.String? entryId,
    $core.String? transactionId,
    $core.String? transactedAt,
    TransactionType? transactionType,
    $1.Money? amount,
    $core.bool? credit,
    $1.Money? runningBalance,
    $0.Struct? transactionData,
    $core.String? clearedAt,
  }) {
    final result = StatementEntry._();
    if (entryId != null) result.entryId = entryId;
    if (transactionId != null) result.transactionId = transactionId;
    if (transactedAt != null) result.transactedAt = transactedAt;
    if (transactionType != null) result.transactionType = transactionType;
    if (amount != null) result.amount = amount;
    if (credit != null) result.credit = credit;
    if (runningBalance != null) result.runningBalance = runningBalance;
    if (transactionData != null) result.transactionData = transactionData;
    if (clearedAt != null) result.clearedAt = clearedAt;
    return result;
  }

  StatementEntry._();

  factory StatementEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StatementEntry()..mergeFromBuffer(data, registry);
  factory StatementEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StatementEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StatementEntry',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: StatementEntry.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'entryId')
    ..aOS(2, _omitFieldNames ? '' : 'transactionId')
    ..aOS(3, _omitFieldNames ? '' : 'transactedAt')
    ..aE<TransactionType>(4, _omitFieldNames ? '' : 'transactionType',
        enumValues: TransactionType.values)
    ..aOM<$1.Money>(5, _omitFieldNames ? '' : 'amount',
        subBuilder: $1.Money.$_createMessage)
    ..aOB(6, _omitFieldNames ? '' : 'credit')
    ..aOM<$1.Money>(7, _omitFieldNames ? '' : 'runningBalance',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$0.Struct>(8, _omitFieldNames ? '' : 'transactionData',
        subBuilder: $0.Struct.$_createMessage)
    ..aOS(9, _omitFieldNames ? '' : 'clearedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StatementEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StatementEntry copyWith(void Function(StatementEntry) updates) =>
      super.copyWith((message) => updates(message as StatementEntry))
          as StatementEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use StatementEntry() / StatementEntry.new instead')
  static StatementEntry create() => StatementEntry._();
  static $pb.GeneratedMessage $_createMessage() => StatementEntry._();
  @$core.override
  StatementEntry createEmptyInstance() => StatementEntry._();
  @$core.pragma('dart2js:noInline')
  static StatementEntry getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StatementEntry>(
          StatementEntry.$_createMessage);
  static StatementEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get entryId => $_getSZ(0);
  @$pb.TagNumber(1)
  set entryId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEntryId() => $_has(0);
  @$pb.TagNumber(1)
  void clearEntryId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get transactionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set transactionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTransactionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearTransactionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get transactedAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set transactedAt($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTransactedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransactedAt() => $_clearField(3);

  @$pb.TagNumber(4)
  TransactionType get transactionType => $_getN(3);
  @$pb.TagNumber(4)
  set transactionType(TransactionType value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasTransactionType() => $_has(3);
  @$pb.TagNumber(4)
  void clearTransactionType() => $_clearField(4);

  /// Signed effect of the entry on the account's balance: positive when the
  /// entry's side matches the account's normal balance, negative when it
  /// opposes it. On a debit-normal account (ASSET, EXPENSE) a credit carries
  /// credit = true and a negative amount; on a credit-normal account
  /// (LIABILITY, INCOME, CAPITAL) a credit carries a positive amount and a
  /// debit a negative one. running_balance is the opening balance plus the
  /// sum of these amounts; use credit for the side and |amount| for the size.
  @$pb.TagNumber(5)
  $1.Money get amount => $_getN(4);
  @$pb.TagNumber(5)
  set amount($1.Money value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearAmount() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Money ensureAmount() => $_ensure(4);

  @$pb.TagNumber(6)
  $core.bool get credit => $_getBF(5);
  @$pb.TagNumber(6)
  set credit($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCredit() => $_has(5);
  @$pb.TagNumber(6)
  void clearCredit() => $_clearField(6);

  @$pb.TagNumber(7)
  $1.Money get runningBalance => $_getN(6);
  @$pb.TagNumber(7)
  set runningBalance($1.Money value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasRunningBalance() => $_has(6);
  @$pb.TagNumber(7)
  void clearRunningBalance() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Money ensureRunningBalance() => $_ensure(6);

  @$pb.TagNumber(8)
  $0.Struct get transactionData => $_getN(7);
  @$pb.TagNumber(8)
  set transactionData($0.Struct value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasTransactionData() => $_has(7);
  @$pb.TagNumber(8)
  void clearTransactionData() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Struct ensureTransactionData() => $_ensure(7);

  @$pb.TagNumber(9)
  $core.String get clearedAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set clearedAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasClearedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearClearedAt() => $_clearField(9);
}

/// SearchLedgersResponse returns ledgers matching search criteria.
class SearchLedgersResponse extends $pb.GeneratedMessage {
  factory SearchLedgersResponse({
    $core.Iterable<Ledger>? data,
  }) {
    final result = SearchLedgersResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchLedgersResponse._();

  factory SearchLedgersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchLedgersResponse()..mergeFromBuffer(data, registry);
  factory SearchLedgersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchLedgersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchLedgersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: SearchLedgersResponse.$_createMessage)
    ..pPM<Ledger>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Ledger.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchLedgersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchLedgersResponse copyWith(
          void Function(SearchLedgersResponse) updates) =>
      super.copyWith((message) => updates(message as SearchLedgersResponse))
          as SearchLedgersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchLedgersResponse() / SearchLedgersResponse.new instead')
  static SearchLedgersResponse create() => SearchLedgersResponse._();
  static $pb.GeneratedMessage $_createMessage() => SearchLedgersResponse._();
  @$core.override
  SearchLedgersResponse createEmptyInstance() => SearchLedgersResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchLedgersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchLedgersResponse>(
          SearchLedgersResponse.$_createMessage);
  static SearchLedgersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Ledger> get data => $_getList(0);
}

/// CreateLedgerRequest creates a new ledger in the chart of accounts.
class CreateLedgerRequest extends $pb.GeneratedMessage {
  factory CreateLedgerRequest({
    $core.String? id,
    LedgerType? type,
    $core.String? parentId,
    $0.Struct? data,
  }) {
    final result = CreateLedgerRequest._();
    if (id != null) result.id = id;
    if (type != null) result.type = type;
    if (parentId != null) result.parentId = parentId;
    if (data != null) result.data = data;
    return result;
  }

  CreateLedgerRequest._();

  factory CreateLedgerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLedgerRequest()..mergeFromBuffer(data, registry);
  factory CreateLedgerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLedgerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateLedgerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateLedgerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aE<LedgerType>(2, _omitFieldNames ? '' : 'type',
        enumValues: LedgerType.values)
    ..aOS(3, _omitFieldNames ? '' : 'parentId')
    ..aOM<$0.Struct>(4, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLedgerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLedgerRequest copyWith(void Function(CreateLedgerRequest) updates) =>
      super.copyWith((message) => updates(message as CreateLedgerRequest))
          as CreateLedgerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CreateLedgerRequest() / CreateLedgerRequest.new instead')
  static CreateLedgerRequest create() => CreateLedgerRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateLedgerRequest._();
  @$core.override
  CreateLedgerRequest createEmptyInstance() => CreateLedgerRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateLedgerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateLedgerRequest>(
          CreateLedgerRequest.$_createMessage);
  static CreateLedgerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  LedgerType get type => $_getN(1);
  @$pb.TagNumber(2)
  set type(LedgerType value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get parentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set parentId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasParentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearParentId() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($0.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Struct ensureData() => $_ensure(3);
}

/// CreateLedgerResponse returns the newly created ledger.
class CreateLedgerResponse extends $pb.GeneratedMessage {
  factory CreateLedgerResponse({
    Ledger? data,
  }) {
    final result = CreateLedgerResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateLedgerResponse._();

  factory CreateLedgerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLedgerResponse()..mergeFromBuffer(data, registry);
  factory CreateLedgerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLedgerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateLedgerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateLedgerResponse.$_createMessage)
    ..aOM<Ledger>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Ledger.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLedgerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLedgerResponse copyWith(void Function(CreateLedgerResponse) updates) =>
      super.copyWith((message) => updates(message as CreateLedgerResponse))
          as CreateLedgerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateLedgerResponse() / CreateLedgerResponse.new instead')
  static CreateLedgerResponse create() => CreateLedgerResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateLedgerResponse._();
  @$core.override
  CreateLedgerResponse createEmptyInstance() => CreateLedgerResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateLedgerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateLedgerResponse>(
          CreateLedgerResponse.$_createMessage);
  static CreateLedgerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Ledger get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Ledger value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Ledger ensureData() => $_ensure(0);
}

/// UpdateLedgerRequest updates an existing ledger's metadata.
class UpdateLedgerRequest extends $pb.GeneratedMessage {
  factory UpdateLedgerRequest({
    $core.String? id,
    $0.Struct? data,
  }) {
    final result = UpdateLedgerRequest._();
    if (id != null) result.id = id;
    if (data != null) result.data = data;
    return result;
  }

  UpdateLedgerRequest._();

  factory UpdateLedgerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLedgerRequest()..mergeFromBuffer(data, registry);
  factory UpdateLedgerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLedgerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateLedgerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: UpdateLedgerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<$0.Struct>(4, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLedgerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLedgerRequest copyWith(void Function(UpdateLedgerRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateLedgerRequest))
          as UpdateLedgerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use UpdateLedgerRequest() / UpdateLedgerRequest.new instead')
  static UpdateLedgerRequest create() => UpdateLedgerRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateLedgerRequest._();
  @$core.override
  UpdateLedgerRequest createEmptyInstance() => UpdateLedgerRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateLedgerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateLedgerRequest>(
          UpdateLedgerRequest.$_createMessage);
  static UpdateLedgerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(4)
  $0.Struct get data => $_getN(1);
  @$pb.TagNumber(4)
  set data($0.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(4)
  void clearData() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Struct ensureData() => $_ensure(1);
}

/// UpdateLedgerResponse returns the updated ledger.
class UpdateLedgerResponse extends $pb.GeneratedMessage {
  factory UpdateLedgerResponse({
    Ledger? data,
  }) {
    final result = UpdateLedgerResponse._();
    if (data != null) result.data = data;
    return result;
  }

  UpdateLedgerResponse._();

  factory UpdateLedgerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLedgerResponse()..mergeFromBuffer(data, registry);
  factory UpdateLedgerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLedgerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateLedgerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: UpdateLedgerResponse.$_createMessage)
    ..aOM<Ledger>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Ledger.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLedgerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLedgerResponse copyWith(void Function(UpdateLedgerResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateLedgerResponse))
          as UpdateLedgerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateLedgerResponse() / UpdateLedgerResponse.new instead')
  static UpdateLedgerResponse create() => UpdateLedgerResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateLedgerResponse._();
  @$core.override
  UpdateLedgerResponse createEmptyInstance() => UpdateLedgerResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateLedgerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateLedgerResponse>(
          UpdateLedgerResponse.$_createMessage);
  static UpdateLedgerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Ledger get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Ledger value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Ledger ensureData() => $_ensure(0);
}

/// SearchAccountsResponse returns accounts matching search criteria.
class SearchAccountsResponse extends $pb.GeneratedMessage {
  factory SearchAccountsResponse({
    $core.Iterable<Account>? data,
  }) {
    final result = SearchAccountsResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchAccountsResponse._();

  factory SearchAccountsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchAccountsResponse()..mergeFromBuffer(data, registry);
  factory SearchAccountsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchAccountsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchAccountsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: SearchAccountsResponse.$_createMessage)
    ..pPM<Account>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Account.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchAccountsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchAccountsResponse copyWith(
          void Function(SearchAccountsResponse) updates) =>
      super.copyWith((message) => updates(message as SearchAccountsResponse))
          as SearchAccountsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchAccountsResponse() / SearchAccountsResponse.new instead')
  static SearchAccountsResponse create() => SearchAccountsResponse._();
  static $pb.GeneratedMessage $_createMessage() => SearchAccountsResponse._();
  @$core.override
  SearchAccountsResponse createEmptyInstance() => SearchAccountsResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchAccountsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchAccountsResponse>(
          SearchAccountsResponse.$_createMessage);
  static SearchAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Account> get data => $_getList(0);
}

/// CreateAccountRequest creates a new account within a ledger.
class CreateAccountRequest extends $pb.GeneratedMessage {
  factory CreateAccountRequest({
    $core.String? id,
    $core.String? ledgerId,
    $core.String? currency,
    $0.Struct? data,
  }) {
    final result = CreateAccountRequest._();
    if (id != null) result.id = id;
    if (ledgerId != null) result.ledgerId = ledgerId;
    if (currency != null) result.currency = currency;
    if (data != null) result.data = data;
    return result;
  }

  CreateAccountRequest._();

  factory CreateAccountRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAccountRequest()..mergeFromBuffer(data, registry);
  factory CreateAccountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAccountRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAccountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateAccountRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'ledgerId')
    ..aOS(4, _omitFieldNames ? '' : 'currency')
    ..aOM<$0.Struct>(5, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAccountRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAccountRequest copyWith(void Function(CreateAccountRequest) updates) =>
      super.copyWith((message) => updates(message as CreateAccountRequest))
          as CreateAccountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAccountRequest() / CreateAccountRequest.new instead')
  static CreateAccountRequest create() => CreateAccountRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateAccountRequest._();
  @$core.override
  CreateAccountRequest createEmptyInstance() => CreateAccountRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateAccountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAccountRequest>(
          CreateAccountRequest.$_createMessage);
  static CreateAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(3)
  $core.String get ledgerId => $_getSZ(1);
  @$pb.TagNumber(3)
  set ledgerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(3)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(3)
  void clearLedgerId() => $_clearField(3);

  /// Asset the account holds: an ISO 4217 code (USD, UGX) or a token symbol
  /// (USDC, USDT, EURC), 3-10 letters or digits, stored upper-cased. The
  /// ledger keeps 9 fractional digits for every asset and never derives minor
  /// units from the code.
  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(2);
  @$pb.TagNumber(4)
  set currency($core.String value) => $_setString(2, value);
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(2);
  @$pb.TagNumber(4)
  void clearCurrency() => $_clearField(4);

  @$pb.TagNumber(5)
  $0.Struct get data => $_getN(3);
  @$pb.TagNumber(5)
  set data($0.Struct value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(5)
  void clearData() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Struct ensureData() => $_ensure(3);
}

/// CreateAccountResponse returns the newly created account.
class CreateAccountResponse extends $pb.GeneratedMessage {
  factory CreateAccountResponse({
    Account? data,
  }) {
    final result = CreateAccountResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateAccountResponse._();

  factory CreateAccountResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAccountResponse()..mergeFromBuffer(data, registry);
  factory CreateAccountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAccountResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAccountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateAccountResponse.$_createMessage)
    ..aOM<Account>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Account.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAccountResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAccountResponse copyWith(
          void Function(CreateAccountResponse) updates) =>
      super.copyWith((message) => updates(message as CreateAccountResponse))
          as CreateAccountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAccountResponse() / CreateAccountResponse.new instead')
  static CreateAccountResponse create() => CreateAccountResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateAccountResponse._();
  @$core.override
  CreateAccountResponse createEmptyInstance() => CreateAccountResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateAccountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAccountResponse>(
          CreateAccountResponse.$_createMessage);
  static CreateAccountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Account get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Account value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Account ensureData() => $_ensure(0);
}

/// UpdateAccountRequest updates an existing account's metadata.
class UpdateAccountRequest extends $pb.GeneratedMessage {
  factory UpdateAccountRequest({
    $core.String? id,
    $0.Struct? data,
  }) {
    final result = UpdateAccountRequest._();
    if (id != null) result.id = id;
    if (data != null) result.data = data;
    return result;
  }

  UpdateAccountRequest._();

  factory UpdateAccountRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAccountRequest()..mergeFromBuffer(data, registry);
  factory UpdateAccountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAccountRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAccountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: UpdateAccountRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<$0.Struct>(4, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAccountRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAccountRequest copyWith(void Function(UpdateAccountRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateAccountRequest))
          as UpdateAccountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAccountRequest() / UpdateAccountRequest.new instead')
  static UpdateAccountRequest create() => UpdateAccountRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAccountRequest._();
  @$core.override
  UpdateAccountRequest createEmptyInstance() => UpdateAccountRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateAccountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAccountRequest>(
          UpdateAccountRequest.$_createMessage);
  static UpdateAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(4)
  $0.Struct get data => $_getN(1);
  @$pb.TagNumber(4)
  set data($0.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(4)
  void clearData() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Struct ensureData() => $_ensure(1);
}

/// UpdateAccountResponse returns the updated account.
class UpdateAccountResponse extends $pb.GeneratedMessage {
  factory UpdateAccountResponse({
    Account? data,
  }) {
    final result = UpdateAccountResponse._();
    if (data != null) result.data = data;
    return result;
  }

  UpdateAccountResponse._();

  factory UpdateAccountResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAccountResponse()..mergeFromBuffer(data, registry);
  factory UpdateAccountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAccountResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAccountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: UpdateAccountResponse.$_createMessage)
    ..aOM<Account>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Account.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAccountResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAccountResponse copyWith(
          void Function(UpdateAccountResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateAccountResponse))
          as UpdateAccountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAccountResponse() / UpdateAccountResponse.new instead')
  static UpdateAccountResponse create() => UpdateAccountResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAccountResponse._();
  @$core.override
  UpdateAccountResponse createEmptyInstance() => UpdateAccountResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateAccountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAccountResponse>(
          UpdateAccountResponse.$_createMessage);
  static UpdateAccountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Account get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Account value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Account ensureData() => $_ensure(0);
}

/// SearchTransactionsResponse returns transactions matching search criteria.
class SearchTransactionsResponse extends $pb.GeneratedMessage {
  factory SearchTransactionsResponse({
    $core.Iterable<Transaction>? data,
  }) {
    final result = SearchTransactionsResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchTransactionsResponse._();

  factory SearchTransactionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchTransactionsResponse()..mergeFromBuffer(data, registry);
  factory SearchTransactionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchTransactionsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchTransactionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: SearchTransactionsResponse.$_createMessage)
    ..pPM<Transaction>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Transaction.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchTransactionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchTransactionsResponse copyWith(
          void Function(SearchTransactionsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SearchTransactionsResponse))
          as SearchTransactionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchTransactionsResponse() / SearchTransactionsResponse.new instead')
  static SearchTransactionsResponse create() => SearchTransactionsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SearchTransactionsResponse._();
  @$core.override
  SearchTransactionsResponse createEmptyInstance() =>
      SearchTransactionsResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchTransactionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchTransactionsResponse>(
          SearchTransactionsResponse.$_createMessage);
  static SearchTransactionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Transaction> get data => $_getList(0);
}

/// CreateTransactionRequest creates a new double-entry transaction.
class CreateTransactionRequest extends $pb.GeneratedMessage {
  factory CreateTransactionRequest({
    $core.String? id,
    $core.String? currency,
    $core.String? transactedAt,
    $0.Struct? data,
    $core.Iterable<TransactionEntry>? entries,
    $core.bool? cleared,
    TransactionType? type,
  }) {
    final result = CreateTransactionRequest._();
    if (id != null) result.id = id;
    if (currency != null) result.currency = currency;
    if (transactedAt != null) result.transactedAt = transactedAt;
    if (data != null) result.data = data;
    if (entries != null) result.entries.addAll(entries);
    if (cleared != null) result.cleared = cleared;
    if (type != null) result.type = type;
    return result;
  }

  CreateTransactionRequest._();

  factory CreateTransactionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTransactionRequest()..mergeFromBuffer(data, registry);
  factory CreateTransactionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTransactionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTransactionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateTransactionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'currency')
    ..aOS(3, _omitFieldNames ? '' : 'transactedAt')
    ..aOM<$0.Struct>(4, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..pPM<TransactionEntry>(5, _omitFieldNames ? '' : 'entries',
        subBuilder: TransactionEntry.$_createMessage)
    ..aOB(6, _omitFieldNames ? '' : 'cleared')
    ..aE<TransactionType>(7, _omitFieldNames ? '' : 'type',
        enumValues: TransactionType.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTransactionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTransactionRequest copyWith(
          void Function(CreateTransactionRequest) updates) =>
      super.copyWith((message) => updates(message as CreateTransactionRequest))
          as CreateTransactionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateTransactionRequest() / CreateTransactionRequest.new instead')
  static CreateTransactionRequest create() => CreateTransactionRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateTransactionRequest._();
  @$core.override
  CreateTransactionRequest createEmptyInstance() =>
      CreateTransactionRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateTransactionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTransactionRequest>(
          CreateTransactionRequest.$_createMessage);
  static CreateTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get currency => $_getSZ(1);
  @$pb.TagNumber(2)
  set currency($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCurrency() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrency() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get transactedAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set transactedAt($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTransactedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransactedAt() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($0.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(5)
  $pb.PbList<TransactionEntry> get entries => $_getList(4);

  @$pb.TagNumber(6)
  $core.bool get cleared => $_getBF(5);
  @$pb.TagNumber(6)
  set cleared($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCleared() => $_has(5);
  @$pb.TagNumber(6)
  void clearCleared() => $_clearField(6);

  @$pb.TagNumber(7)
  TransactionType get type => $_getN(6);
  @$pb.TagNumber(7)
  set type(TransactionType value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasType() => $_has(6);
  @$pb.TagNumber(7)
  void clearType() => $_clearField(7);
}

/// CreateTransactionResponse returns the newly created transaction.
class CreateTransactionResponse extends $pb.GeneratedMessage {
  factory CreateTransactionResponse({
    Transaction? data,
  }) {
    final result = CreateTransactionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateTransactionResponse._();

  factory CreateTransactionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTransactionResponse()..mergeFromBuffer(data, registry);
  factory CreateTransactionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTransactionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTransactionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateTransactionResponse.$_createMessage)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Transaction.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTransactionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTransactionResponse copyWith(
          void Function(CreateTransactionResponse) updates) =>
      super.copyWith((message) => updates(message as CreateTransactionResponse))
          as CreateTransactionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateTransactionResponse() / CreateTransactionResponse.new instead')
  static CreateTransactionResponse create() => CreateTransactionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateTransactionResponse._();
  @$core.override
  CreateTransactionResponse createEmptyInstance() =>
      CreateTransactionResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateTransactionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTransactionResponse>(
          CreateTransactionResponse.$_createMessage);
  static CreateTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

/// ReverseTransactionRequest reverses a transaction by creating offsetting entries.
class ReverseTransactionRequest extends $pb.GeneratedMessage {
  factory ReverseTransactionRequest({
    $core.String? id,
  }) {
    final result = ReverseTransactionRequest._();
    if (id != null) result.id = id;
    return result;
  }

  ReverseTransactionRequest._();

  factory ReverseTransactionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReverseTransactionRequest()..mergeFromBuffer(data, registry);
  factory ReverseTransactionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReverseTransactionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReverseTransactionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: ReverseTransactionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseTransactionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseTransactionRequest copyWith(
          void Function(ReverseTransactionRequest) updates) =>
      super.copyWith((message) => updates(message as ReverseTransactionRequest))
          as ReverseTransactionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReverseTransactionRequest() / ReverseTransactionRequest.new instead')
  static ReverseTransactionRequest create() => ReverseTransactionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ReverseTransactionRequest._();
  @$core.override
  ReverseTransactionRequest createEmptyInstance() =>
      ReverseTransactionRequest._();
  @$core.pragma('dart2js:noInline')
  static ReverseTransactionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReverseTransactionRequest>(
          ReverseTransactionRequest.$_createMessage);
  static ReverseTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// ReverseTransactionResponse returns the reversal transaction.
class ReverseTransactionResponse extends $pb.GeneratedMessage {
  factory ReverseTransactionResponse({
    Transaction? data,
  }) {
    final result = ReverseTransactionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  ReverseTransactionResponse._();

  factory ReverseTransactionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReverseTransactionResponse()..mergeFromBuffer(data, registry);
  factory ReverseTransactionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReverseTransactionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReverseTransactionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: ReverseTransactionResponse.$_createMessage)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Transaction.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseTransactionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReverseTransactionResponse copyWith(
          void Function(ReverseTransactionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ReverseTransactionResponse))
          as ReverseTransactionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReverseTransactionResponse() / ReverseTransactionResponse.new instead')
  static ReverseTransactionResponse create() => ReverseTransactionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ReverseTransactionResponse._();
  @$core.override
  ReverseTransactionResponse createEmptyInstance() =>
      ReverseTransactionResponse._();
  @$core.pragma('dart2js:noInline')
  static ReverseTransactionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReverseTransactionResponse>(
          ReverseTransactionResponse.$_createMessage);
  static ReverseTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

/// UpdateTransactionRequest updates a transaction's metadata.
class UpdateTransactionRequest extends $pb.GeneratedMessage {
  factory UpdateTransactionRequest({
    $core.String? id,
    $core.String? clearedAt,
    $0.Struct? data,
  }) {
    final result = UpdateTransactionRequest._();
    if (id != null) result.id = id;
    if (clearedAt != null) result.clearedAt = clearedAt;
    if (data != null) result.data = data;
    return result;
  }

  UpdateTransactionRequest._();

  factory UpdateTransactionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTransactionRequest()..mergeFromBuffer(data, registry);
  factory UpdateTransactionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTransactionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTransactionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: UpdateTransactionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'clearedAt')
    ..aOM<$0.Struct>(4, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTransactionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTransactionRequest copyWith(
          void Function(UpdateTransactionRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateTransactionRequest))
          as UpdateTransactionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTransactionRequest() / UpdateTransactionRequest.new instead')
  static UpdateTransactionRequest create() => UpdateTransactionRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateTransactionRequest._();
  @$core.override
  UpdateTransactionRequest createEmptyInstance() =>
      UpdateTransactionRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateTransactionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTransactionRequest>(
          UpdateTransactionRequest.$_createMessage);
  static UpdateTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get clearedAt => $_getSZ(1);
  @$pb.TagNumber(2)
  set clearedAt($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasClearedAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearClearedAt() => $_clearField(2);

  @$pb.TagNumber(4)
  $0.Struct get data => $_getN(2);
  @$pb.TagNumber(4)
  set data($0.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(2);
  @$pb.TagNumber(4)
  void clearData() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Struct ensureData() => $_ensure(2);
}

/// UpdateTransactionResponse returns the updated transaction.
class UpdateTransactionResponse extends $pb.GeneratedMessage {
  factory UpdateTransactionResponse({
    Transaction? data,
  }) {
    final result = UpdateTransactionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  UpdateTransactionResponse._();

  factory UpdateTransactionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTransactionResponse()..mergeFromBuffer(data, registry);
  factory UpdateTransactionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTransactionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTransactionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: UpdateTransactionResponse.$_createMessage)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Transaction.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTransactionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTransactionResponse copyWith(
          void Function(UpdateTransactionResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateTransactionResponse))
          as UpdateTransactionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTransactionResponse() / UpdateTransactionResponse.new instead')
  static UpdateTransactionResponse create() => UpdateTransactionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateTransactionResponse._();
  @$core.override
  UpdateTransactionResponse createEmptyInstance() =>
      UpdateTransactionResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateTransactionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTransactionResponse>(
          UpdateTransactionResponse.$_createMessage);
  static UpdateTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

/// SearchTransactionEntriesResponse returns transaction entries matching search criteria.
class SearchTransactionEntriesResponse extends $pb.GeneratedMessage {
  factory SearchTransactionEntriesResponse({
    $core.Iterable<TransactionEntry>? data,
  }) {
    final result = SearchTransactionEntriesResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchTransactionEntriesResponse._();

  factory SearchTransactionEntriesResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchTransactionEntriesResponse()..mergeFromBuffer(data, registry);
  factory SearchTransactionEntriesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchTransactionEntriesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchTransactionEntriesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: SearchTransactionEntriesResponse.$_createMessage)
    ..pPM<TransactionEntry>(1, _omitFieldNames ? '' : 'data',
        subBuilder: TransactionEntry.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchTransactionEntriesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchTransactionEntriesResponse copyWith(
          void Function(SearchTransactionEntriesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SearchTransactionEntriesResponse))
          as SearchTransactionEntriesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchTransactionEntriesResponse() / SearchTransactionEntriesResponse.new instead')
  static SearchTransactionEntriesResponse create() =>
      SearchTransactionEntriesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SearchTransactionEntriesResponse._();
  @$core.override
  SearchTransactionEntriesResponse createEmptyInstance() =>
      SearchTransactionEntriesResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchTransactionEntriesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchTransactionEntriesResponse>(
          SearchTransactionEntriesResponse.$_createMessage);
  static SearchTransactionEntriesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<TransactionEntry> get data => $_getList(0);
}

class CreateBookRequest extends $pb.GeneratedMessage {
  factory CreateBookRequest({
    $core.String? id,
    $core.String? name,
    $core.String? type,
    $core.String? parentId,
    $core.String? currency,
    $0.Struct? data,
  }) {
    final result = CreateBookRequest._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (type != null) result.type = type;
    if (parentId != null) result.parentId = parentId;
    if (currency != null) result.currency = currency;
    if (data != null) result.data = data;
    return result;
  }

  CreateBookRequest._();

  factory CreateBookRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateBookRequest()..mergeFromBuffer(data, registry);
  factory CreateBookRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateBookRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateBookRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateBookRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOS(4, _omitFieldNames ? '' : 'parentId')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$0.Struct>(6, _omitFieldNames ? '' : 'data',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateBookRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateBookRequest copyWith(void Function(CreateBookRequest) updates) =>
      super.copyWith((message) => updates(message as CreateBookRequest))
          as CreateBookRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateBookRequest() / CreateBookRequest.new instead')
  static CreateBookRequest create() => CreateBookRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateBookRequest._();
  @$core.override
  CreateBookRequest createEmptyInstance() => CreateBookRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateBookRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateBookRequest>(
          CreateBookRequest.$_createMessage);
  static CreateBookRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get parentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set parentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasParentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearParentId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Struct get data => $_getN(5);
  @$pb.TagNumber(6)
  set data($0.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasData() => $_has(5);
  @$pb.TagNumber(6)
  void clearData() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Struct ensureData() => $_ensure(5);
}

class CreateBookResponse extends $pb.GeneratedMessage {
  factory CreateBookResponse({
    Book? data,
  }) {
    final result = CreateBookResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateBookResponse._();

  factory CreateBookResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateBookResponse()..mergeFromBuffer(data, registry);
  factory CreateBookResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateBookResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateBookResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: CreateBookResponse.$_createMessage)
    ..aOM<Book>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Book.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateBookResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateBookResponse copyWith(void Function(CreateBookResponse) updates) =>
      super.copyWith((message) => updates(message as CreateBookResponse))
          as CreateBookResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateBookResponse() / CreateBookResponse.new instead')
  static CreateBookResponse create() => CreateBookResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateBookResponse._();
  @$core.override
  CreateBookResponse createEmptyInstance() => CreateBookResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateBookResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateBookResponse>(
          CreateBookResponse.$_createMessage);
  static CreateBookResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Book get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Book value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Book ensureData() => $_ensure(0);
}

class GetBookRequest extends $pb.GeneratedMessage {
  factory GetBookRequest({
    $core.String? id,
  }) {
    final result = GetBookRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetBookRequest._();

  factory GetBookRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBookRequest()..mergeFromBuffer(data, registry);
  factory GetBookRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBookRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBookRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: GetBookRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBookRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBookRequest copyWith(void Function(GetBookRequest) updates) =>
      super.copyWith((message) => updates(message as GetBookRequest))
          as GetBookRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetBookRequest() / GetBookRequest.new instead')
  static GetBookRequest create() => GetBookRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetBookRequest._();
  @$core.override
  GetBookRequest createEmptyInstance() => GetBookRequest._();
  @$core.pragma('dart2js:noInline')
  static GetBookRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetBookRequest>(
          GetBookRequest.$_createMessage);
  static GetBookRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetBookResponse extends $pb.GeneratedMessage {
  factory GetBookResponse({
    Book? data,
  }) {
    final result = GetBookResponse._();
    if (data != null) result.data = data;
    return result;
  }

  GetBookResponse._();

  factory GetBookResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBookResponse()..mergeFromBuffer(data, registry);
  factory GetBookResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBookResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBookResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: GetBookResponse.$_createMessage)
    ..aOM<Book>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Book.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBookResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBookResponse copyWith(void Function(GetBookResponse) updates) =>
      super.copyWith((message) => updates(message as GetBookResponse))
          as GetBookResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetBookResponse() / GetBookResponse.new instead')
  static GetBookResponse create() => GetBookResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetBookResponse._();
  @$core.override
  GetBookResponse createEmptyInstance() => GetBookResponse._();
  @$core.pragma('dart2js:noInline')
  static GetBookResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetBookResponse>(
          GetBookResponse.$_createMessage);
  static GetBookResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Book get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Book value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Book ensureData() => $_ensure(0);
}

class ListBooksByTypeRequest extends $pb.GeneratedMessage {
  factory ListBooksByTypeRequest({
    $core.String? type,
  }) {
    final result = ListBooksByTypeRequest._();
    if (type != null) result.type = type;
    return result;
  }

  ListBooksByTypeRequest._();

  factory ListBooksByTypeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListBooksByTypeRequest()..mergeFromBuffer(data, registry);
  factory ListBooksByTypeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListBooksByTypeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListBooksByTypeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: ListBooksByTypeRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'type')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBooksByTypeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBooksByTypeRequest copyWith(
          void Function(ListBooksByTypeRequest) updates) =>
      super.copyWith((message) => updates(message as ListBooksByTypeRequest))
          as ListBooksByTypeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListBooksByTypeRequest() / ListBooksByTypeRequest.new instead')
  static ListBooksByTypeRequest create() => ListBooksByTypeRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListBooksByTypeRequest._();
  @$core.override
  ListBooksByTypeRequest createEmptyInstance() => ListBooksByTypeRequest._();
  @$core.pragma('dart2js:noInline')
  static ListBooksByTypeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListBooksByTypeRequest>(
          ListBooksByTypeRequest.$_createMessage);
  static ListBooksByTypeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);
}

class ListBooksByTypeResponse extends $pb.GeneratedMessage {
  factory ListBooksByTypeResponse({
    $core.Iterable<Book>? data,
  }) {
    final result = ListBooksByTypeResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  ListBooksByTypeResponse._();

  factory ListBooksByTypeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListBooksByTypeResponse()..mergeFromBuffer(data, registry);
  factory ListBooksByTypeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListBooksByTypeResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListBooksByTypeResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: ListBooksByTypeResponse.$_createMessage)
    ..pPM<Book>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Book.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBooksByTypeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBooksByTypeResponse copyWith(
          void Function(ListBooksByTypeResponse) updates) =>
      super.copyWith((message) => updates(message as ListBooksByTypeResponse))
          as ListBooksByTypeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListBooksByTypeResponse() / ListBooksByTypeResponse.new instead')
  static ListBooksByTypeResponse create() => ListBooksByTypeResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListBooksByTypeResponse._();
  @$core.override
  ListBooksByTypeResponse createEmptyInstance() => ListBooksByTypeResponse._();
  @$core.pragma('dart2js:noInline')
  static ListBooksByTypeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListBooksByTypeResponse>(
          ListBooksByTypeResponse.$_createMessage);
  static ListBooksByTypeResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Book> get data => $_getList(0);
}

class GetTrialBalanceRequest extends $pb.GeneratedMessage {
  factory GetTrialBalanceRequest({
    $core.String? currency,
    $core.String? ledgerId,
    $core.String? ledgerType,
    $core.Iterable<$core.String>? bookIds,
    $core.String? asOf,
  }) {
    final result = GetTrialBalanceRequest._();
    if (currency != null) result.currency = currency;
    if (ledgerId != null) result.ledgerId = ledgerId;
    if (ledgerType != null) result.ledgerType = ledgerType;
    if (bookIds != null) result.bookIds.addAll(bookIds);
    if (asOf != null) result.asOf = asOf;
    return result;
  }

  GetTrialBalanceRequest._();

  factory GetTrialBalanceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTrialBalanceRequest()..mergeFromBuffer(data, registry);
  factory GetTrialBalanceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTrialBalanceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTrialBalanceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: GetTrialBalanceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'currency')
    ..aOS(2, _omitFieldNames ? '' : 'ledgerId')
    ..aOS(3, _omitFieldNames ? '' : 'ledgerType')
    ..pPS(4, _omitFieldNames ? '' : 'bookIds')
    ..aOS(5, _omitFieldNames ? '' : 'asOf')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTrialBalanceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTrialBalanceRequest copyWith(
          void Function(GetTrialBalanceRequest) updates) =>
      super.copyWith((message) => updates(message as GetTrialBalanceRequest))
          as GetTrialBalanceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTrialBalanceRequest() / GetTrialBalanceRequest.new instead')
  static GetTrialBalanceRequest create() => GetTrialBalanceRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetTrialBalanceRequest._();
  @$core.override
  GetTrialBalanceRequest createEmptyInstance() => GetTrialBalanceRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTrialBalanceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTrialBalanceRequest>(
          GetTrialBalanceRequest.$_createMessage);
  static GetTrialBalanceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get currency => $_getSZ(0);
  @$pb.TagNumber(1)
  set currency($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCurrency() => $_has(0);
  @$pb.TagNumber(1)
  void clearCurrency() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get ledgerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set ledgerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearLedgerId() => $_clearField(2);

  /// ledger_type is a case-sensitive string filter ("ASSET", "LIABILITY",
  /// "INCOME", "EXPENSE", "CAPITAL"). Empty means "no filter". A string is
  /// used instead of the LedgerType enum so the caller can omit the filter
  /// entirely — proto3 enums treat the zero value as both "unset" and the
  /// first enum constant (ASSET in this codebase), which is ambiguous.
  @$pb.TagNumber(3)
  $core.String get ledgerType => $_getSZ(2);
  @$pb.TagNumber(3)
  set ledgerType($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLedgerType() => $_has(2);
  @$pb.TagNumber(3)
  void clearLedgerType() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get bookIds => $_getList(3);

  @$pb.TagNumber(5)
  $core.String get asOf => $_getSZ(4);
  @$pb.TagNumber(5)
  set asOf($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAsOf() => $_has(4);
  @$pb.TagNumber(5)
  void clearAsOf() => $_clearField(5);
}

class GetTrialBalanceResponse extends $pb.GeneratedMessage {
  factory GetTrialBalanceResponse({
    $core.Iterable<TrialBalanceLine>? lines,
    $core.Iterable<TrialBalanceTotal>? totals,
  }) {
    final result = GetTrialBalanceResponse._();
    if (lines != null) result.lines.addAll(lines);
    if (totals != null) result.totals.addAll(totals);
    return result;
  }

  GetTrialBalanceResponse._();

  factory GetTrialBalanceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTrialBalanceResponse()..mergeFromBuffer(data, registry);
  factory GetTrialBalanceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTrialBalanceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTrialBalanceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: GetTrialBalanceResponse.$_createMessage)
    ..pPM<TrialBalanceLine>(1, _omitFieldNames ? '' : 'lines',
        subBuilder: TrialBalanceLine.$_createMessage)
    ..pPM<TrialBalanceTotal>(2, _omitFieldNames ? '' : 'totals',
        subBuilder: TrialBalanceTotal.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTrialBalanceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTrialBalanceResponse copyWith(
          void Function(GetTrialBalanceResponse) updates) =>
      super.copyWith((message) => updates(message as GetTrialBalanceResponse))
          as GetTrialBalanceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTrialBalanceResponse() / GetTrialBalanceResponse.new instead')
  static GetTrialBalanceResponse create() => GetTrialBalanceResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetTrialBalanceResponse._();
  @$core.override
  GetTrialBalanceResponse createEmptyInstance() => GetTrialBalanceResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTrialBalanceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTrialBalanceResponse>(
          GetTrialBalanceResponse.$_createMessage);
  static GetTrialBalanceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<TrialBalanceLine> get lines => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<TrialBalanceTotal> get totals => $_getList(1);
}

class GetAccountStatementRequest extends $pb.GeneratedMessage {
  factory GetAccountStatementRequest({
    $core.String? accountId,
    $core.String? from,
    $core.String? to,
    $core.int? limit,
    $core.int? offset,
  }) {
    final result = GetAccountStatementRequest._();
    if (accountId != null) result.accountId = accountId;
    if (from != null) result.from = from;
    if (to != null) result.to = to;
    if (limit != null) result.limit = limit;
    if (offset != null) result.offset = offset;
    return result;
  }

  GetAccountStatementRequest._();

  factory GetAccountStatementRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAccountStatementRequest()..mergeFromBuffer(data, registry);
  factory GetAccountStatementRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAccountStatementRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAccountStatementRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: GetAccountStatementRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'from')
    ..aOS(3, _omitFieldNames ? '' : 'to')
    ..aI(4, _omitFieldNames ? '' : 'limit')
    ..aI(5, _omitFieldNames ? '' : 'offset')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAccountStatementRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAccountStatementRequest copyWith(
          void Function(GetAccountStatementRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetAccountStatementRequest))
          as GetAccountStatementRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAccountStatementRequest() / GetAccountStatementRequest.new instead')
  static GetAccountStatementRequest create() => GetAccountStatementRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAccountStatementRequest._();
  @$core.override
  GetAccountStatementRequest createEmptyInstance() =>
      GetAccountStatementRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAccountStatementRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAccountStatementRequest>(
          GetAccountStatementRequest.$_createMessage);
  static GetAccountStatementRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get from => $_getSZ(1);
  @$pb.TagNumber(2)
  set from($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFrom() => $_has(1);
  @$pb.TagNumber(2)
  void clearFrom() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get to => $_getSZ(2);
  @$pb.TagNumber(3)
  set to($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTo() => $_has(2);
  @$pb.TagNumber(3)
  void clearTo() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get limit => $_getIZ(3);
  @$pb.TagNumber(4)
  set limit($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLimit() => $_has(3);
  @$pb.TagNumber(4)
  void clearLimit() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get offset => $_getIZ(4);
  @$pb.TagNumber(5)
  set offset($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOffset() => $_has(4);
  @$pb.TagNumber(5)
  void clearOffset() => $_clearField(5);
}

class GetAccountStatementResponse extends $pb.GeneratedMessage {
  factory GetAccountStatementResponse({
    $core.String? accountId,
    $core.String? currency,
    $1.Money? openingBalance,
    $1.Money? closingBalance,
    $1.Money? totalDebits,
    $1.Money? totalCredits,
    $core.Iterable<StatementEntry>? entries,
  }) {
    final result = GetAccountStatementResponse._();
    if (accountId != null) result.accountId = accountId;
    if (currency != null) result.currency = currency;
    if (openingBalance != null) result.openingBalance = openingBalance;
    if (closingBalance != null) result.closingBalance = closingBalance;
    if (totalDebits != null) result.totalDebits = totalDebits;
    if (totalCredits != null) result.totalCredits = totalCredits;
    if (entries != null) result.entries.addAll(entries);
    return result;
  }

  GetAccountStatementResponse._();

  factory GetAccountStatementResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAccountStatementResponse()..mergeFromBuffer(data, registry);
  factory GetAccountStatementResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAccountStatementResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAccountStatementResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: GetAccountStatementResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Money>(3, _omitFieldNames ? '' : 'openingBalance',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$1.Money>(4, _omitFieldNames ? '' : 'closingBalance',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$1.Money>(5, _omitFieldNames ? '' : 'totalDebits',
        subBuilder: $1.Money.$_createMessage)
    ..aOM<$1.Money>(6, _omitFieldNames ? '' : 'totalCredits',
        subBuilder: $1.Money.$_createMessage)
    ..pPM<StatementEntry>(7, _omitFieldNames ? '' : 'entries',
        subBuilder: StatementEntry.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAccountStatementResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAccountStatementResponse copyWith(
          void Function(GetAccountStatementResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetAccountStatementResponse))
          as GetAccountStatementResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAccountStatementResponse() / GetAccountStatementResponse.new instead')
  static GetAccountStatementResponse create() =>
      GetAccountStatementResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAccountStatementResponse._();
  @$core.override
  GetAccountStatementResponse createEmptyInstance() =>
      GetAccountStatementResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAccountStatementResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAccountStatementResponse>(
          GetAccountStatementResponse.$_createMessage);
  static GetAccountStatementResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get currency => $_getSZ(1);
  @$pb.TagNumber(2)
  set currency($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCurrency() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrency() => $_clearField(2);

  @$pb.TagNumber(3)
  $1.Money get openingBalance => $_getN(2);
  @$pb.TagNumber(3)
  set openingBalance($1.Money value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasOpeningBalance() => $_has(2);
  @$pb.TagNumber(3)
  void clearOpeningBalance() => $_clearField(3);
  @$pb.TagNumber(3)
  $1.Money ensureOpeningBalance() => $_ensure(2);

  @$pb.TagNumber(4)
  $1.Money get closingBalance => $_getN(3);
  @$pb.TagNumber(4)
  set closingBalance($1.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasClosingBalance() => $_has(3);
  @$pb.TagNumber(4)
  void clearClosingBalance() => $_clearField(4);
  @$pb.TagNumber(4)
  $1.Money ensureClosingBalance() => $_ensure(3);

  @$pb.TagNumber(5)
  $1.Money get totalDebits => $_getN(4);
  @$pb.TagNumber(5)
  set totalDebits($1.Money value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasTotalDebits() => $_has(4);
  @$pb.TagNumber(5)
  void clearTotalDebits() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Money ensureTotalDebits() => $_ensure(4);

  @$pb.TagNumber(6)
  $1.Money get totalCredits => $_getN(5);
  @$pb.TagNumber(6)
  set totalCredits($1.Money value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasTotalCredits() => $_has(5);
  @$pb.TagNumber(6)
  void clearTotalCredits() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Money ensureTotalCredits() => $_ensure(5);

  @$pb.TagNumber(7)
  $pb.PbList<StatementEntry> get entries => $_getList(6);
}

class VoidTransactionRequest extends $pb.GeneratedMessage {
  factory VoidTransactionRequest({
    $core.String? id,
  }) {
    final result = VoidTransactionRequest._();
    if (id != null) result.id = id;
    return result;
  }

  VoidTransactionRequest._();

  factory VoidTransactionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidTransactionRequest()..mergeFromBuffer(data, registry);
  factory VoidTransactionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidTransactionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VoidTransactionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: VoidTransactionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidTransactionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidTransactionRequest copyWith(
          void Function(VoidTransactionRequest) updates) =>
      super.copyWith((message) => updates(message as VoidTransactionRequest))
          as VoidTransactionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use VoidTransactionRequest() / VoidTransactionRequest.new instead')
  static VoidTransactionRequest create() => VoidTransactionRequest._();
  static $pb.GeneratedMessage $_createMessage() => VoidTransactionRequest._();
  @$core.override
  VoidTransactionRequest createEmptyInstance() => VoidTransactionRequest._();
  @$core.pragma('dart2js:noInline')
  static VoidTransactionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VoidTransactionRequest>(
          VoidTransactionRequest.$_createMessage);
  static VoidTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class VoidTransactionResponse extends $pb.GeneratedMessage {
  factory VoidTransactionResponse({
    Transaction? data,
  }) {
    final result = VoidTransactionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  VoidTransactionResponse._();

  factory VoidTransactionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidTransactionResponse()..mergeFromBuffer(data, registry);
  factory VoidTransactionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidTransactionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VoidTransactionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: VoidTransactionResponse.$_createMessage)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Transaction.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidTransactionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidTransactionResponse copyWith(
          void Function(VoidTransactionResponse) updates) =>
      super.copyWith((message) => updates(message as VoidTransactionResponse))
          as VoidTransactionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use VoidTransactionResponse() / VoidTransactionResponse.new instead')
  static VoidTransactionResponse create() => VoidTransactionResponse._();
  static $pb.GeneratedMessage $_createMessage() => VoidTransactionResponse._();
  @$core.override
  VoidTransactionResponse createEmptyInstance() => VoidTransactionResponse._();
  @$core.pragma('dart2js:noInline')
  static VoidTransactionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VoidTransactionResponse>(
          VoidTransactionResponse.$_createMessage);
  static VoidTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

class MarkTransactionFailedRequest extends $pb.GeneratedMessage {
  factory MarkTransactionFailedRequest({
    $core.String? id,
  }) {
    final result = MarkTransactionFailedRequest._();
    if (id != null) result.id = id;
    return result;
  }

  MarkTransactionFailedRequest._();

  factory MarkTransactionFailedRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkTransactionFailedRequest()..mergeFromBuffer(data, registry);
  factory MarkTransactionFailedRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkTransactionFailedRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MarkTransactionFailedRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: MarkTransactionFailedRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkTransactionFailedRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkTransactionFailedRequest copyWith(
          void Function(MarkTransactionFailedRequest) updates) =>
      super.copyWith(
              (message) => updates(message as MarkTransactionFailedRequest))
          as MarkTransactionFailedRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MarkTransactionFailedRequest() / MarkTransactionFailedRequest.new instead')
  static MarkTransactionFailedRequest create() =>
      MarkTransactionFailedRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      MarkTransactionFailedRequest._();
  @$core.override
  MarkTransactionFailedRequest createEmptyInstance() =>
      MarkTransactionFailedRequest._();
  @$core.pragma('dart2js:noInline')
  static MarkTransactionFailedRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MarkTransactionFailedRequest>(
          MarkTransactionFailedRequest.$_createMessage);
  static MarkTransactionFailedRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class MarkTransactionFailedResponse extends $pb.GeneratedMessage {
  factory MarkTransactionFailedResponse({
    Transaction? data,
  }) {
    final result = MarkTransactionFailedResponse._();
    if (data != null) result.data = data;
    return result;
  }

  MarkTransactionFailedResponse._();

  factory MarkTransactionFailedResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkTransactionFailedResponse()..mergeFromBuffer(data, registry);
  factory MarkTransactionFailedResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkTransactionFailedResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MarkTransactionFailedResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'),
      createEmptyInstance: MarkTransactionFailedResponse.$_createMessage)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Transaction.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkTransactionFailedResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkTransactionFailedResponse copyWith(
          void Function(MarkTransactionFailedResponse) updates) =>
      super.copyWith(
              (message) => updates(message as MarkTransactionFailedResponse))
          as MarkTransactionFailedResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MarkTransactionFailedResponse() / MarkTransactionFailedResponse.new instead')
  static MarkTransactionFailedResponse create() =>
      MarkTransactionFailedResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      MarkTransactionFailedResponse._();
  @$core.override
  MarkTransactionFailedResponse createEmptyInstance() =>
      MarkTransactionFailedResponse._();
  @$core.pragma('dart2js:noInline')
  static MarkTransactionFailedResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MarkTransactionFailedResponse>(
          MarkTransactionFailedResponse.$_createMessage);
  static MarkTransactionFailedResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

/// LedgerService provides double-entry bookkeeping and financial accounting.
/// All RPCs require authentication via Bearer token.
class LedgerServiceApi {
  final $pb.RpcClient _client;

  LedgerServiceApi(this._client);

  /// SearchLedgers finds ledgers in the chart of accounts.
  /// Supports filtering by type, parent, and custom properties.
  $async.Future<SearchLedgersResponse> searchLedgers(
          $pb.ClientContext? ctx, $2.SearchRequest request) =>
      _client.invoke<SearchLedgersResponse>(ctx, 'LedgerService',
          'SearchLedgers', request, SearchLedgersResponse());

  /// CreateLedger creates a new ledger in the chart of accounts.
  /// Ledgers can be hierarchical with parent-child relationships.
  $async.Future<CreateLedgerResponse> createLedger(
          $pb.ClientContext? ctx, CreateLedgerRequest request) =>
      _client.invoke<CreateLedgerResponse>(ctx, 'LedgerService', 'CreateLedger',
          request, CreateLedgerResponse());

  /// UpdateLedger updates an existing ledger's metadata.
  /// The ledger type and reference cannot be changed.
  $async.Future<UpdateLedgerResponse> updateLedger(
          $pb.ClientContext? ctx, UpdateLedgerRequest request) =>
      _client.invoke<UpdateLedgerResponse>(ctx, 'LedgerService', 'UpdateLedger',
          request, UpdateLedgerResponse());

  /// SearchAccounts finds accounts matching specified criteria.
  /// Supports filtering by ledger, balance range, and custom properties.
  $async.Future<SearchAccountsResponse> searchAccounts(
          $pb.ClientContext? ctx, $2.SearchRequest request) =>
      _client.invoke<SearchAccountsResponse>(ctx, 'LedgerService',
          'SearchAccounts', request, SearchAccountsResponse());

  /// CreateAccount creates a new account within a ledger.
  /// Each account tracks balances and transaction history.
  $async.Future<CreateAccountResponse> createAccount(
          $pb.ClientContext? ctx, CreateAccountRequest request) =>
      _client.invoke<CreateAccountResponse>(ctx, 'LedgerService',
          'CreateAccount', request, CreateAccountResponse());

  /// UpdateAccount updates an existing account's metadata.
  /// Balances are updated through transactions, not directly.
  $async.Future<UpdateAccountResponse> updateAccount(
          $pb.ClientContext? ctx, UpdateAccountRequest request) =>
      _client.invoke<UpdateAccountResponse>(ctx, 'LedgerService',
          'UpdateAccount', request, UpdateAccountResponse());

  /// SearchTransactions finds transactions matching specified criteria.
  /// Supports filtering by date range, account, currency, and status.
  $async.Future<SearchTransactionsResponse> searchTransactions(
          $pb.ClientContext? ctx, $2.SearchRequest request) =>
      _client.invoke<SearchTransactionsResponse>(ctx, 'LedgerService',
          'SearchTransactions', request, SearchTransactionsResponse());

  /// CreateTransaction creates a new double-entry transaction.
  /// All entries must be balanced (sum of debits = sum of credits).
  $async.Future<CreateTransactionResponse> createTransaction(
          $pb.ClientContext? ctx, CreateTransactionRequest request) =>
      _client.invoke<CreateTransactionResponse>(ctx, 'LedgerService',
          'CreateTransaction', request, CreateTransactionResponse());

  /// ReverseTransaction reverses a transaction by creating offsetting entries.
  /// Creates a new REVERSAL transaction that negates the original.
  $async.Future<ReverseTransactionResponse> reverseTransaction(
          $pb.ClientContext? ctx, ReverseTransactionRequest request) =>
      _client.invoke<ReverseTransactionResponse>(ctx, 'LedgerService',
          'ReverseTransaction', request, ReverseTransactionResponse());

  /// UpdateTransaction updates a transaction's metadata.
  /// Entries and amounts cannot be changed after creation.
  $async.Future<UpdateTransactionResponse> updateTransaction(
          $pb.ClientContext? ctx, UpdateTransactionRequest request) =>
      _client.invoke<UpdateTransactionResponse>(ctx, 'LedgerService',
          'UpdateTransaction', request, UpdateTransactionResponse());

  /// SearchTransactionEntries finds individual transaction entries.
  /// Useful for account statement generation and reconciliation.
  $async.Future<SearchTransactionEntriesResponse> searchTransactionEntries(
          $pb.ClientContext? ctx, $2.SearchRequest request) =>
      _client.invoke<SearchTransactionEntriesResponse>(
          ctx,
          'LedgerService',
          'SearchTransactionEntries',
          request,
          SearchTransactionEntriesResponse());

  /// VoidTransaction transitions a draft or pending transaction to VOIDED.
  /// Posted activity cannot be voided — reverse it instead so the books
  /// carry the offset and the audit trail.
  $async.Future<VoidTransactionResponse> voidTransaction(
          $pb.ClientContext? ctx, VoidTransactionRequest request) =>
      _client.invoke<VoidTransactionResponse>(ctx, 'LedgerService',
          'VoidTransaction', request, VoidTransactionResponse());

  /// MarkTransactionFailed transitions a pending transaction to FAILED for
  /// upstream rejections (e.g. payment provider declined the webhook).
  $async.Future<MarkTransactionFailedResponse> markTransactionFailed(
          $pb.ClientContext? ctx, MarkTransactionFailedRequest request) =>
      _client.invoke<MarkTransactionFailedResponse>(ctx, 'LedgerService',
          'MarkTransactionFailed', request, MarkTransactionFailedResponse());

  /// CreateBook creates a new accounting scope. Each book balances
  /// independently; cross-book postings are rejected.
  $async.Future<CreateBookResponse> createBook(
          $pb.ClientContext? ctx, CreateBookRequest request) =>
      _client.invoke<CreateBookResponse>(
          ctx, 'LedgerService', 'CreateBook', request, CreateBookResponse());

  /// GetBook fetches a single book by id within the caller's tenancy.
  $async.Future<GetBookResponse> getBook(
          $pb.ClientContext? ctx, GetBookRequest request) =>
      _client.invoke<GetBookResponse>(
          ctx, 'LedgerService', 'GetBook', request, GetBookResponse());

  /// ListBooksByType returns all books of a given conventional type.
  $async.Future<ListBooksByTypeResponse> listBooksByType(
          $pb.ClientContext? ctx, ListBooksByTypeRequest request) =>
      _client.invoke<ListBooksByTypeResponse>(ctx, 'LedgerService',
          'ListBooksByType', request, ListBooksByTypeResponse());

  /// GetTrialBalance returns the per-account debit/credit totals plus per-
  /// currency grand totals with the textbook integrity check (is_balanced).
  $async.Future<GetTrialBalanceResponse> getTrialBalance(
          $pb.ClientContext? ctx, GetTrialBalanceRequest request) =>
      _client.invoke<GetTrialBalanceResponse>(ctx, 'LedgerService',
          'GetTrialBalance', request, GetTrialBalanceResponse());

  /// GetAccountStatement returns a customer-facing account ledger for a
  /// period with opening balance, running balance per entry, and totals.
  $async.Future<GetAccountStatementResponse> getAccountStatement(
          $pb.ClientContext? ctx, GetAccountStatementRequest request) =>
      _client.invoke<GetAccountStatementResponse>(ctx, 'LedgerService',
          'GetAccountStatement', request, GetAccountStatementResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
