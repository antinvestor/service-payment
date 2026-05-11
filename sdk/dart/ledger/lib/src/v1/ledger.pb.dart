//
//  Generated code. Do not modify.
//  source: v1/ledger.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../common/v1/common.pb.dart' as $8;
import '../common/v1/money.pb.dart' as $7;
import '../google/protobuf/struct.pb.dart' as $6;
import 'ledger.pbenum.dart';

export 'ledger.pbenum.dart';

/// Ledger represents a category in the chart of accounts.
/// Ledgers can be hierarchical with parent-child relationships.
class Ledger extends $pb.GeneratedMessage {
  factory Ledger({
    $core.String? id,
    LedgerType? type,
    $core.String? parent,
    $6.Struct? data,
    $core.String? bookId,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (type != null) {
      $result.type = type;
    }
    if (parent != null) {
      $result.parent = parent;
    }
    if (data != null) {
      $result.data = data;
    }
    if (bookId != null) {
      $result.bookId = bookId;
    }
    return $result;
  }
  Ledger._() : super();
  factory Ledger.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Ledger.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Ledger', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..e<LedgerType>(2, _omitFieldNames ? '' : 'type', $pb.PbFieldType.OE, defaultOrMaker: LedgerType.ASSET, valueOf: LedgerType.valueOf, enumValues: LedgerType.values)
    ..aOS(3, _omitFieldNames ? '' : 'parent')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..aOS(5, _omitFieldNames ? '' : 'bookId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Ledger clone() => Ledger()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Ledger copyWith(void Function(Ledger) updates) => super.copyWith((message) => updates(message as Ledger)) as Ledger;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Ledger create() => Ledger._();
  Ledger createEmptyInstance() => create();
  static $pb.PbList<Ledger> createRepeated() => $pb.PbList<Ledger>();
  @$core.pragma('dart2js:noInline')
  static Ledger getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Ledger>(create);
  static Ledger? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  LedgerType get type => $_getN(1);
  @$pb.TagNumber(2)
  set type(LedgerType v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get parent => $_getSZ(2);
  @$pb.TagNumber(3)
  set parent($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasParent() => $_has(2);
  @$pb.TagNumber(3)
  void clearParent() => clearField(3);

  @$pb.TagNumber(4)
  $6.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($6.Struct v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get bookId => $_getSZ(4);
  @$pb.TagNumber(5)
  set bookId($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasBookId() => $_has(4);
  @$pb.TagNumber(5)
  void clearBookId() => clearField(5);
}

/// Account represents a specific account within a ledger.
/// Tracks balances and supports multi-currency operations.
class Account extends $pb.GeneratedMessage {
  factory Account({
    $core.String? id,
    $core.String? ledger,
    $7.Money? balance,
    $6.Struct? data,
    $7.Money? unclearedBalance,
    $7.Money? reservedBalance,
    AccountType? accountType,
    NormalBalance? normalBalance,
    $core.String? bookId,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (ledger != null) {
      $result.ledger = ledger;
    }
    if (balance != null) {
      $result.balance = balance;
    }
    if (data != null) {
      $result.data = data;
    }
    if (unclearedBalance != null) {
      $result.unclearedBalance = unclearedBalance;
    }
    if (reservedBalance != null) {
      $result.reservedBalance = reservedBalance;
    }
    if (accountType != null) {
      $result.accountType = accountType;
    }
    if (normalBalance != null) {
      $result.normalBalance = normalBalance;
    }
    if (bookId != null) {
      $result.bookId = bookId;
    }
    return $result;
  }
  Account._() : super();
  factory Account.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Account.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Account', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'ledger')
    ..aOM<$7.Money>(4, _omitFieldNames ? '' : 'balance', subBuilder: $7.Money.create)
    ..aOM<$6.Struct>(5, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..aOM<$7.Money>(6, _omitFieldNames ? '' : 'unclearedBalance', subBuilder: $7.Money.create)
    ..aOM<$7.Money>(7, _omitFieldNames ? '' : 'reservedBalance', subBuilder: $7.Money.create)
    ..e<AccountType>(8, _omitFieldNames ? '' : 'accountType', $pb.PbFieldType.OE, defaultOrMaker: AccountType.ACCOUNT_ASSET, valueOf: AccountType.valueOf, enumValues: AccountType.values)
    ..e<NormalBalance>(9, _omitFieldNames ? '' : 'normalBalance', $pb.PbFieldType.OE, defaultOrMaker: NormalBalance.DEBIT, valueOf: NormalBalance.valueOf, enumValues: NormalBalance.values)
    ..aOS(10, _omitFieldNames ? '' : 'bookId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Account clone() => Account()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Account copyWith(void Function(Account) updates) => super.copyWith((message) => updates(message as Account)) as Account;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Account create() => Account._();
  Account createEmptyInstance() => create();
  static $pb.PbList<Account> createRepeated() => $pb.PbList<Account>();
  @$core.pragma('dart2js:noInline')
  static Account getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Account>(create);
  static Account? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(3)
  $core.String get ledger => $_getSZ(1);
  @$pb.TagNumber(3)
  set ledger($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasLedger() => $_has(1);
  @$pb.TagNumber(3)
  void clearLedger() => clearField(3);

  @$pb.TagNumber(4)
  $7.Money get balance => $_getN(2);
  @$pb.TagNumber(4)
  set balance($7.Money v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasBalance() => $_has(2);
  @$pb.TagNumber(4)
  void clearBalance() => clearField(4);
  @$pb.TagNumber(4)
  $7.Money ensureBalance() => $_ensure(2);

  @$pb.TagNumber(5)
  $6.Struct get data => $_getN(3);
  @$pb.TagNumber(5)
  set data($6.Struct v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(5)
  void clearData() => clearField(5);
  @$pb.TagNumber(5)
  $6.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(6)
  $7.Money get unclearedBalance => $_getN(4);
  @$pb.TagNumber(6)
  set unclearedBalance($7.Money v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasUnclearedBalance() => $_has(4);
  @$pb.TagNumber(6)
  void clearUnclearedBalance() => clearField(6);
  @$pb.TagNumber(6)
  $7.Money ensureUnclearedBalance() => $_ensure(4);

  @$pb.TagNumber(7)
  $7.Money get reservedBalance => $_getN(5);
  @$pb.TagNumber(7)
  set reservedBalance($7.Money v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasReservedBalance() => $_has(5);
  @$pb.TagNumber(7)
  void clearReservedBalance() => clearField(7);
  @$pb.TagNumber(7)
  $7.Money ensureReservedBalance() => $_ensure(5);

  @$pb.TagNumber(8)
  AccountType get accountType => $_getN(6);
  @$pb.TagNumber(8)
  set accountType(AccountType v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasAccountType() => $_has(6);
  @$pb.TagNumber(8)
  void clearAccountType() => clearField(8);

  @$pb.TagNumber(9)
  NormalBalance get normalBalance => $_getN(7);
  @$pb.TagNumber(9)
  set normalBalance(NormalBalance v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasNormalBalance() => $_has(7);
  @$pb.TagNumber(9)
  void clearNormalBalance() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get bookId => $_getSZ(8);
  @$pb.TagNumber(10)
  set bookId($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(10)
  $core.bool hasBookId() => $_has(8);
  @$pb.TagNumber(10)
  void clearBookId() => clearField(10);
}

/// TransactionEntry represents one side of a double-entry transaction.
/// Each transaction must have at least two entries with balanced debits and credits.
class TransactionEntry extends $pb.GeneratedMessage {
  factory TransactionEntry({
    $core.String? id,
    $core.String? accountId,
    $core.String? transactionId,
    $core.String? transactedAt,
    $7.Money? amount,
    $core.bool? credit,
    $7.Money? accBalance,
    $core.String? clearedAt,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (accountId != null) {
      $result.accountId = accountId;
    }
    if (transactionId != null) {
      $result.transactionId = transactionId;
    }
    if (transactedAt != null) {
      $result.transactedAt = transactedAt;
    }
    if (amount != null) {
      $result.amount = amount;
    }
    if (credit != null) {
      $result.credit = credit;
    }
    if (accBalance != null) {
      $result.accBalance = accBalance;
    }
    if (clearedAt != null) {
      $result.clearedAt = clearedAt;
    }
    return $result;
  }
  TransactionEntry._() : super();
  factory TransactionEntry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TransactionEntry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'TransactionEntry', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'accountId')
    ..aOS(4, _omitFieldNames ? '' : 'transactionId')
    ..aOS(5, _omitFieldNames ? '' : 'transactedAt')
    ..aOM<$7.Money>(7, _omitFieldNames ? '' : 'amount', subBuilder: $7.Money.create)
    ..aOB(8, _omitFieldNames ? '' : 'credit')
    ..aOM<$7.Money>(9, _omitFieldNames ? '' : 'accBalance', subBuilder: $7.Money.create)
    ..aOS(11, _omitFieldNames ? '' : 'clearedAt')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TransactionEntry clone() => TransactionEntry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TransactionEntry copyWith(void Function(TransactionEntry) updates) => super.copyWith((message) => updates(message as TransactionEntry)) as TransactionEntry;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TransactionEntry create() => TransactionEntry._();
  TransactionEntry createEmptyInstance() => create();
  static $pb.PbList<TransactionEntry> createRepeated() => $pb.PbList<TransactionEntry>();
  @$core.pragma('dart2js:noInline')
  static TransactionEntry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TransactionEntry>(create);
  static TransactionEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(3)
  $core.String get accountId => $_getSZ(1);
  @$pb.TagNumber(3)
  set accountId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasAccountId() => $_has(1);
  @$pb.TagNumber(3)
  void clearAccountId() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get transactionId => $_getSZ(2);
  @$pb.TagNumber(4)
  set transactionId($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(4)
  $core.bool hasTransactionId() => $_has(2);
  @$pb.TagNumber(4)
  void clearTransactionId() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get transactedAt => $_getSZ(3);
  @$pb.TagNumber(5)
  set transactedAt($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(5)
  $core.bool hasTransactedAt() => $_has(3);
  @$pb.TagNumber(5)
  void clearTransactedAt() => clearField(5);

  @$pb.TagNumber(7)
  $7.Money get amount => $_getN(4);
  @$pb.TagNumber(7)
  set amount($7.Money v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasAmount() => $_has(4);
  @$pb.TagNumber(7)
  void clearAmount() => clearField(7);
  @$pb.TagNumber(7)
  $7.Money ensureAmount() => $_ensure(4);

  @$pb.TagNumber(8)
  $core.bool get credit => $_getBF(5);
  @$pb.TagNumber(8)
  set credit($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(8)
  $core.bool hasCredit() => $_has(5);
  @$pb.TagNumber(8)
  void clearCredit() => clearField(8);

  @$pb.TagNumber(9)
  $7.Money get accBalance => $_getN(6);
  @$pb.TagNumber(9)
  set accBalance($7.Money v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasAccBalance() => $_has(6);
  @$pb.TagNumber(9)
  void clearAccBalance() => clearField(9);
  @$pb.TagNumber(9)
  $7.Money ensureAccBalance() => $_ensure(6);

  @$pb.TagNumber(11)
  $core.String get clearedAt => $_getSZ(7);
  @$pb.TagNumber(11)
  set clearedAt($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(11)
  $core.bool hasClearedAt() => $_has(7);
  @$pb.TagNumber(11)
  void clearClearedAt() => clearField(11);
}

/// Transaction represents a complete double-entry transaction.
/// Must contain balanced entries (sum of debits = sum of credits).
class Transaction extends $pb.GeneratedMessage {
  factory Transaction({
    $core.String? id,
    $core.String? currencyCode,
    $core.String? transactedAt,
    $6.Struct? data,
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
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (currencyCode != null) {
      $result.currencyCode = currencyCode;
    }
    if (transactedAt != null) {
      $result.transactedAt = transactedAt;
    }
    if (data != null) {
      $result.data = data;
    }
    if (entries != null) {
      $result.entries.addAll(entries);
    }
    if (cleared != null) {
      $result.cleared = cleared;
    }
    if (type != null) {
      $result.type = type;
    }
    if (status != null) {
      $result.status = status;
    }
    if (idempotencyKey != null) {
      $result.idempotencyKey = idempotencyKey;
    }
    if (externalRef != null) {
      $result.externalRef = externalRef;
    }
    if (source != null) {
      $result.source = source;
    }
    if (bookId != null) {
      $result.bookId = bookId;
    }
    if (postedAt != null) {
      $result.postedAt = postedAt;
    }
    if (voidedAt != null) {
      $result.voidedAt = voidedAt;
    }
    if (reversedTransactionId != null) {
      $result.reversedTransactionId = reversedTransactionId;
    }
    return $result;
  }
  Transaction._() : super();
  factory Transaction.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transaction.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Transaction', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'currencyCode')
    ..aOS(3, _omitFieldNames ? '' : 'transactedAt')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..pc<TransactionEntry>(5, _omitFieldNames ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: TransactionEntry.create)
    ..aOB(6, _omitFieldNames ? '' : 'cleared')
    ..e<TransactionType>(7, _omitFieldNames ? '' : 'type', $pb.PbFieldType.OE, defaultOrMaker: TransactionType.NORMAL, valueOf: TransactionType.valueOf, enumValues: TransactionType.values)
    ..e<TransactionStatus>(8, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE, defaultOrMaker: TransactionStatus.PENDING, valueOf: TransactionStatus.valueOf, enumValues: TransactionStatus.values)
    ..aOS(9, _omitFieldNames ? '' : 'idempotencyKey')
    ..aOS(10, _omitFieldNames ? '' : 'externalRef')
    ..aOS(11, _omitFieldNames ? '' : 'source')
    ..aOS(12, _omitFieldNames ? '' : 'bookId')
    ..aOS(13, _omitFieldNames ? '' : 'postedAt')
    ..aOS(14, _omitFieldNames ? '' : 'voidedAt')
    ..aOS(15, _omitFieldNames ? '' : 'reversedTransactionId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Transaction clone() => Transaction()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Transaction copyWith(void Function(Transaction) updates) => super.copyWith((message) => updates(message as Transaction)) as Transaction;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Transaction create() => Transaction._();
  Transaction createEmptyInstance() => create();
  static $pb.PbList<Transaction> createRepeated() => $pb.PbList<Transaction>();
  @$core.pragma('dart2js:noInline')
  static Transaction getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Transaction>(create);
  static Transaction? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  /// The three-letter currency code defined in ISO 4217.
  @$pb.TagNumber(2)
  $core.String get currencyCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set currencyCode($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCurrencyCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrencyCode() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get transactedAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set transactedAt($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTransactedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransactedAt() => clearField(3);

  @$pb.TagNumber(4)
  $6.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($6.Struct v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.List<TransactionEntry> get entries => $_getList(4);

  @$pb.TagNumber(6)
  $core.bool get cleared => $_getBF(5);
  @$pb.TagNumber(6)
  set cleared($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasCleared() => $_has(5);
  @$pb.TagNumber(6)
  void clearCleared() => clearField(6);

  @$pb.TagNumber(7)
  TransactionType get type => $_getN(6);
  @$pb.TagNumber(7)
  set type(TransactionType v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasType() => $_has(6);
  @$pb.TagNumber(7)
  void clearType() => clearField(7);

  @$pb.TagNumber(8)
  TransactionStatus get status => $_getN(7);
  @$pb.TagNumber(8)
  set status(TransactionStatus v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasStatus() => $_has(7);
  @$pb.TagNumber(8)
  void clearStatus() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get idempotencyKey => $_getSZ(8);
  @$pb.TagNumber(9)
  set idempotencyKey($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasIdempotencyKey() => $_has(8);
  @$pb.TagNumber(9)
  void clearIdempotencyKey() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get externalRef => $_getSZ(9);
  @$pb.TagNumber(10)
  set externalRef($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasExternalRef() => $_has(9);
  @$pb.TagNumber(10)
  void clearExternalRef() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get source => $_getSZ(10);
  @$pb.TagNumber(11)
  set source($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasSource() => $_has(10);
  @$pb.TagNumber(11)
  void clearSource() => clearField(11);

  @$pb.TagNumber(12)
  $core.String get bookId => $_getSZ(11);
  @$pb.TagNumber(12)
  set bookId($core.String v) { $_setString(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasBookId() => $_has(11);
  @$pb.TagNumber(12)
  void clearBookId() => clearField(12);

  @$pb.TagNumber(13)
  $core.String get postedAt => $_getSZ(12);
  @$pb.TagNumber(13)
  set postedAt($core.String v) { $_setString(12, v); }
  @$pb.TagNumber(13)
  $core.bool hasPostedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearPostedAt() => clearField(13);

  @$pb.TagNumber(14)
  $core.String get voidedAt => $_getSZ(13);
  @$pb.TagNumber(14)
  set voidedAt($core.String v) { $_setString(13, v); }
  @$pb.TagNumber(14)
  $core.bool hasVoidedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearVoidedAt() => clearField(14);

  @$pb.TagNumber(15)
  $core.String get reversedTransactionId => $_getSZ(14);
  @$pb.TagNumber(15)
  set reversedTransactionId($core.String v) { $_setString(14, v); }
  @$pb.TagNumber(15)
  $core.bool hasReversedTransactionId() => $_has(14);
  @$pb.TagNumber(15)
  void clearReversedTransactionId() => clearField(15);
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
    $6.Struct? data,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (name != null) {
      $result.name = name;
    }
    if (type != null) {
      $result.type = type;
    }
    if (parentId != null) {
      $result.parentId = parentId;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  Book._() : super();
  factory Book.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Book.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Book', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOS(4, _omitFieldNames ? '' : 'parentId')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$6.Struct>(6, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Book clone() => Book()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Book copyWith(void Function(Book) updates) => super.copyWith((message) => updates(message as Book)) as Book;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Book create() => Book._();
  Book createEmptyInstance() => create();
  static $pb.PbList<Book> createRepeated() => $pb.PbList<Book>();
  @$core.pragma('dart2js:noInline')
  static Book getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Book>(create);
  static Book? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get parentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set parentId($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasParentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearParentId() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => clearField(5);

  @$pb.TagNumber(6)
  $6.Struct get data => $_getN(5);
  @$pb.TagNumber(6)
  set data($6.Struct v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasData() => $_has(5);
  @$pb.TagNumber(6)
  void clearData() => clearField(6);
  @$pb.TagNumber(6)
  $6.Struct ensureData() => $_ensure(5);
}

/// TrialBalanceLine is one row of the trial balance — a single account's
/// raw debit and credit totals plus the DEADCLIC-signed net balance.
class TrialBalanceLine extends $pb.GeneratedMessage {
  factory TrialBalanceLine({
    $core.String? accountId,
    $core.String? ledgerId,
    LedgerType? ledgerType,
    $core.String? currency,
    $7.Money? totalDebits,
    $7.Money? totalCredits,
    $7.Money? netBalance,
  }) {
    final $result = create();
    if (accountId != null) {
      $result.accountId = accountId;
    }
    if (ledgerId != null) {
      $result.ledgerId = ledgerId;
    }
    if (ledgerType != null) {
      $result.ledgerType = ledgerType;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (totalDebits != null) {
      $result.totalDebits = totalDebits;
    }
    if (totalCredits != null) {
      $result.totalCredits = totalCredits;
    }
    if (netBalance != null) {
      $result.netBalance = netBalance;
    }
    return $result;
  }
  TrialBalanceLine._() : super();
  factory TrialBalanceLine.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TrialBalanceLine.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'TrialBalanceLine', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'ledgerId')
    ..e<LedgerType>(3, _omitFieldNames ? '' : 'ledgerType', $pb.PbFieldType.OE, defaultOrMaker: LedgerType.ASSET, valueOf: LedgerType.valueOf, enumValues: LedgerType.values)
    ..aOS(4, _omitFieldNames ? '' : 'currency')
    ..aOM<$7.Money>(5, _omitFieldNames ? '' : 'totalDebits', subBuilder: $7.Money.create)
    ..aOM<$7.Money>(6, _omitFieldNames ? '' : 'totalCredits', subBuilder: $7.Money.create)
    ..aOM<$7.Money>(7, _omitFieldNames ? '' : 'netBalance', subBuilder: $7.Money.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TrialBalanceLine clone() => TrialBalanceLine()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TrialBalanceLine copyWith(void Function(TrialBalanceLine) updates) => super.copyWith((message) => updates(message as TrialBalanceLine)) as TrialBalanceLine;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TrialBalanceLine create() => TrialBalanceLine._();
  TrialBalanceLine createEmptyInstance() => create();
  static $pb.PbList<TrialBalanceLine> createRepeated() => $pb.PbList<TrialBalanceLine>();
  @$core.pragma('dart2js:noInline')
  static TrialBalanceLine getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TrialBalanceLine>(create);
  static TrialBalanceLine? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get ledgerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set ledgerId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearLedgerId() => clearField(2);

  @$pb.TagNumber(3)
  LedgerType get ledgerType => $_getN(2);
  @$pb.TagNumber(3)
  set ledgerType(LedgerType v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasLedgerType() => $_has(2);
  @$pb.TagNumber(3)
  void clearLedgerType() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(3);
  @$pb.TagNumber(4)
  set currency($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(3);
  @$pb.TagNumber(4)
  void clearCurrency() => clearField(4);

  @$pb.TagNumber(5)
  $7.Money get totalDebits => $_getN(4);
  @$pb.TagNumber(5)
  set totalDebits($7.Money v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasTotalDebits() => $_has(4);
  @$pb.TagNumber(5)
  void clearTotalDebits() => clearField(5);
  @$pb.TagNumber(5)
  $7.Money ensureTotalDebits() => $_ensure(4);

  @$pb.TagNumber(6)
  $7.Money get totalCredits => $_getN(5);
  @$pb.TagNumber(6)
  set totalCredits($7.Money v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasTotalCredits() => $_has(5);
  @$pb.TagNumber(6)
  void clearTotalCredits() => clearField(6);
  @$pb.TagNumber(6)
  $7.Money ensureTotalCredits() => $_ensure(5);

  @$pb.TagNumber(7)
  $7.Money get netBalance => $_getN(6);
  @$pb.TagNumber(7)
  set netBalance($7.Money v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasNetBalance() => $_has(6);
  @$pb.TagNumber(7)
  void clearNetBalance() => clearField(7);
  @$pb.TagNumber(7)
  $7.Money ensureNetBalance() => $_ensure(6);
}

/// TrialBalanceTotal aggregates one currency's debits and credits across
/// every line in the report. is_balanced is the textbook integrity check.
class TrialBalanceTotal extends $pb.GeneratedMessage {
  factory TrialBalanceTotal({
    $core.String? currency,
    $7.Money? totalDebits,
    $7.Money? totalCredits,
    $core.bool? isBalanced,
  }) {
    final $result = create();
    if (currency != null) {
      $result.currency = currency;
    }
    if (totalDebits != null) {
      $result.totalDebits = totalDebits;
    }
    if (totalCredits != null) {
      $result.totalCredits = totalCredits;
    }
    if (isBalanced != null) {
      $result.isBalanced = isBalanced;
    }
    return $result;
  }
  TrialBalanceTotal._() : super();
  factory TrialBalanceTotal.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TrialBalanceTotal.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'TrialBalanceTotal', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'currency')
    ..aOM<$7.Money>(2, _omitFieldNames ? '' : 'totalDebits', subBuilder: $7.Money.create)
    ..aOM<$7.Money>(3, _omitFieldNames ? '' : 'totalCredits', subBuilder: $7.Money.create)
    ..aOB(4, _omitFieldNames ? '' : 'isBalanced')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TrialBalanceTotal clone() => TrialBalanceTotal()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TrialBalanceTotal copyWith(void Function(TrialBalanceTotal) updates) => super.copyWith((message) => updates(message as TrialBalanceTotal)) as TrialBalanceTotal;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TrialBalanceTotal create() => TrialBalanceTotal._();
  TrialBalanceTotal createEmptyInstance() => create();
  static $pb.PbList<TrialBalanceTotal> createRepeated() => $pb.PbList<TrialBalanceTotal>();
  @$core.pragma('dart2js:noInline')
  static TrialBalanceTotal getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TrialBalanceTotal>(create);
  static TrialBalanceTotal? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get currency => $_getSZ(0);
  @$pb.TagNumber(1)
  set currency($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCurrency() => $_has(0);
  @$pb.TagNumber(1)
  void clearCurrency() => clearField(1);

  @$pb.TagNumber(2)
  $7.Money get totalDebits => $_getN(1);
  @$pb.TagNumber(2)
  set totalDebits($7.Money v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasTotalDebits() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotalDebits() => clearField(2);
  @$pb.TagNumber(2)
  $7.Money ensureTotalDebits() => $_ensure(1);

  @$pb.TagNumber(3)
  $7.Money get totalCredits => $_getN(2);
  @$pb.TagNumber(3)
  set totalCredits($7.Money v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasTotalCredits() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotalCredits() => clearField(3);
  @$pb.TagNumber(3)
  $7.Money ensureTotalCredits() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.bool get isBalanced => $_getBF(3);
  @$pb.TagNumber(4)
  set isBalanced($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasIsBalanced() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsBalanced() => clearField(4);
}

/// StatementEntry is one row on an account statement, hydrated with its
/// parent transaction's metadata and the running balance after the entry.
class StatementEntry extends $pb.GeneratedMessage {
  factory StatementEntry({
    $core.String? entryId,
    $core.String? transactionId,
    $core.String? transactedAt,
    TransactionType? transactionType,
    $7.Money? amount,
    $core.bool? credit,
    $7.Money? runningBalance,
    $6.Struct? transactionData,
    $core.String? clearedAt,
  }) {
    final $result = create();
    if (entryId != null) {
      $result.entryId = entryId;
    }
    if (transactionId != null) {
      $result.transactionId = transactionId;
    }
    if (transactedAt != null) {
      $result.transactedAt = transactedAt;
    }
    if (transactionType != null) {
      $result.transactionType = transactionType;
    }
    if (amount != null) {
      $result.amount = amount;
    }
    if (credit != null) {
      $result.credit = credit;
    }
    if (runningBalance != null) {
      $result.runningBalance = runningBalance;
    }
    if (transactionData != null) {
      $result.transactionData = transactionData;
    }
    if (clearedAt != null) {
      $result.clearedAt = clearedAt;
    }
    return $result;
  }
  StatementEntry._() : super();
  factory StatementEntry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory StatementEntry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'StatementEntry', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'entryId')
    ..aOS(2, _omitFieldNames ? '' : 'transactionId')
    ..aOS(3, _omitFieldNames ? '' : 'transactedAt')
    ..e<TransactionType>(4, _omitFieldNames ? '' : 'transactionType', $pb.PbFieldType.OE, defaultOrMaker: TransactionType.NORMAL, valueOf: TransactionType.valueOf, enumValues: TransactionType.values)
    ..aOM<$7.Money>(5, _omitFieldNames ? '' : 'amount', subBuilder: $7.Money.create)
    ..aOB(6, _omitFieldNames ? '' : 'credit')
    ..aOM<$7.Money>(7, _omitFieldNames ? '' : 'runningBalance', subBuilder: $7.Money.create)
    ..aOM<$6.Struct>(8, _omitFieldNames ? '' : 'transactionData', subBuilder: $6.Struct.create)
    ..aOS(9, _omitFieldNames ? '' : 'clearedAt')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  StatementEntry clone() => StatementEntry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  StatementEntry copyWith(void Function(StatementEntry) updates) => super.copyWith((message) => updates(message as StatementEntry)) as StatementEntry;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StatementEntry create() => StatementEntry._();
  StatementEntry createEmptyInstance() => create();
  static $pb.PbList<StatementEntry> createRepeated() => $pb.PbList<StatementEntry>();
  @$core.pragma('dart2js:noInline')
  static StatementEntry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StatementEntry>(create);
  static StatementEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get entryId => $_getSZ(0);
  @$pb.TagNumber(1)
  set entryId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasEntryId() => $_has(0);
  @$pb.TagNumber(1)
  void clearEntryId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get transactionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set transactionId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasTransactionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearTransactionId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get transactedAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set transactedAt($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTransactedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransactedAt() => clearField(3);

  @$pb.TagNumber(4)
  TransactionType get transactionType => $_getN(3);
  @$pb.TagNumber(4)
  set transactionType(TransactionType v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasTransactionType() => $_has(3);
  @$pb.TagNumber(4)
  void clearTransactionType() => clearField(4);

  @$pb.TagNumber(5)
  $7.Money get amount => $_getN(4);
  @$pb.TagNumber(5)
  set amount($7.Money v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearAmount() => clearField(5);
  @$pb.TagNumber(5)
  $7.Money ensureAmount() => $_ensure(4);

  @$pb.TagNumber(6)
  $core.bool get credit => $_getBF(5);
  @$pb.TagNumber(6)
  set credit($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasCredit() => $_has(5);
  @$pb.TagNumber(6)
  void clearCredit() => clearField(6);

  @$pb.TagNumber(7)
  $7.Money get runningBalance => $_getN(6);
  @$pb.TagNumber(7)
  set runningBalance($7.Money v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasRunningBalance() => $_has(6);
  @$pb.TagNumber(7)
  void clearRunningBalance() => clearField(7);
  @$pb.TagNumber(7)
  $7.Money ensureRunningBalance() => $_ensure(6);

  @$pb.TagNumber(8)
  $6.Struct get transactionData => $_getN(7);
  @$pb.TagNumber(8)
  set transactionData($6.Struct v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasTransactionData() => $_has(7);
  @$pb.TagNumber(8)
  void clearTransactionData() => clearField(8);
  @$pb.TagNumber(8)
  $6.Struct ensureTransactionData() => $_ensure(7);

  @$pb.TagNumber(9)
  $core.String get clearedAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set clearedAt($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasClearedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearClearedAt() => clearField(9);
}

/// SearchLedgersResponse returns ledgers matching search criteria.
class SearchLedgersResponse extends $pb.GeneratedMessage {
  factory SearchLedgersResponse({
    $core.Iterable<Ledger>? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data.addAll(data);
    }
    return $result;
  }
  SearchLedgersResponse._() : super();
  factory SearchLedgersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SearchLedgersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SearchLedgersResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..pc<Ledger>(1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.PM, subBuilder: Ledger.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SearchLedgersResponse clone() => SearchLedgersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SearchLedgersResponse copyWith(void Function(SearchLedgersResponse) updates) => super.copyWith((message) => updates(message as SearchLedgersResponse)) as SearchLedgersResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SearchLedgersResponse create() => SearchLedgersResponse._();
  SearchLedgersResponse createEmptyInstance() => create();
  static $pb.PbList<SearchLedgersResponse> createRepeated() => $pb.PbList<SearchLedgersResponse>();
  @$core.pragma('dart2js:noInline')
  static SearchLedgersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SearchLedgersResponse>(create);
  static SearchLedgersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Ledger> get data => $_getList(0);
}

/// CreateLedgerRequest creates a new ledger in the chart of accounts.
class CreateLedgerRequest extends $pb.GeneratedMessage {
  factory CreateLedgerRequest({
    $core.String? id,
    LedgerType? type,
    $core.String? parentId,
    $6.Struct? data,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (type != null) {
      $result.type = type;
    }
    if (parentId != null) {
      $result.parentId = parentId;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  CreateLedgerRequest._() : super();
  factory CreateLedgerRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateLedgerRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateLedgerRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..e<LedgerType>(2, _omitFieldNames ? '' : 'type', $pb.PbFieldType.OE, defaultOrMaker: LedgerType.ASSET, valueOf: LedgerType.valueOf, enumValues: LedgerType.values)
    ..aOS(3, _omitFieldNames ? '' : 'parentId')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateLedgerRequest clone() => CreateLedgerRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateLedgerRequest copyWith(void Function(CreateLedgerRequest) updates) => super.copyWith((message) => updates(message as CreateLedgerRequest)) as CreateLedgerRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateLedgerRequest create() => CreateLedgerRequest._();
  CreateLedgerRequest createEmptyInstance() => create();
  static $pb.PbList<CreateLedgerRequest> createRepeated() => $pb.PbList<CreateLedgerRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateLedgerRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateLedgerRequest>(create);
  static CreateLedgerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  LedgerType get type => $_getN(1);
  @$pb.TagNumber(2)
  set type(LedgerType v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get parentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set parentId($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasParentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearParentId() => clearField(3);

  @$pb.TagNumber(4)
  $6.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($6.Struct v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureData() => $_ensure(3);
}

/// CreateLedgerResponse returns the newly created ledger.
class CreateLedgerResponse extends $pb.GeneratedMessage {
  factory CreateLedgerResponse({
    Ledger? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  CreateLedgerResponse._() : super();
  factory CreateLedgerResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateLedgerResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateLedgerResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Ledger>(1, _omitFieldNames ? '' : 'data', subBuilder: Ledger.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateLedgerResponse clone() => CreateLedgerResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateLedgerResponse copyWith(void Function(CreateLedgerResponse) updates) => super.copyWith((message) => updates(message as CreateLedgerResponse)) as CreateLedgerResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateLedgerResponse create() => CreateLedgerResponse._();
  CreateLedgerResponse createEmptyInstance() => create();
  static $pb.PbList<CreateLedgerResponse> createRepeated() => $pb.PbList<CreateLedgerResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateLedgerResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateLedgerResponse>(create);
  static CreateLedgerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Ledger get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Ledger v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Ledger ensureData() => $_ensure(0);
}

/// UpdateLedgerRequest updates an existing ledger's metadata.
class UpdateLedgerRequest extends $pb.GeneratedMessage {
  factory UpdateLedgerRequest({
    $core.String? id,
    $6.Struct? data,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  UpdateLedgerRequest._() : super();
  factory UpdateLedgerRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateLedgerRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateLedgerRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateLedgerRequest clone() => UpdateLedgerRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateLedgerRequest copyWith(void Function(UpdateLedgerRequest) updates) => super.copyWith((message) => updates(message as UpdateLedgerRequest)) as UpdateLedgerRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateLedgerRequest create() => UpdateLedgerRequest._();
  UpdateLedgerRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateLedgerRequest> createRepeated() => $pb.PbList<UpdateLedgerRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateLedgerRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateLedgerRequest>(create);
  static UpdateLedgerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(4)
  $6.Struct get data => $_getN(1);
  @$pb.TagNumber(4)
  set data($6.Struct v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureData() => $_ensure(1);
}

/// UpdateLedgerResponse returns the updated ledger.
class UpdateLedgerResponse extends $pb.GeneratedMessage {
  factory UpdateLedgerResponse({
    Ledger? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  UpdateLedgerResponse._() : super();
  factory UpdateLedgerResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateLedgerResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateLedgerResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Ledger>(1, _omitFieldNames ? '' : 'data', subBuilder: Ledger.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateLedgerResponse clone() => UpdateLedgerResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateLedgerResponse copyWith(void Function(UpdateLedgerResponse) updates) => super.copyWith((message) => updates(message as UpdateLedgerResponse)) as UpdateLedgerResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateLedgerResponse create() => UpdateLedgerResponse._();
  UpdateLedgerResponse createEmptyInstance() => create();
  static $pb.PbList<UpdateLedgerResponse> createRepeated() => $pb.PbList<UpdateLedgerResponse>();
  @$core.pragma('dart2js:noInline')
  static UpdateLedgerResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateLedgerResponse>(create);
  static UpdateLedgerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Ledger get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Ledger v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Ledger ensureData() => $_ensure(0);
}

/// SearchAccountsResponse returns accounts matching search criteria.
class SearchAccountsResponse extends $pb.GeneratedMessage {
  factory SearchAccountsResponse({
    $core.Iterable<Account>? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data.addAll(data);
    }
    return $result;
  }
  SearchAccountsResponse._() : super();
  factory SearchAccountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SearchAccountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SearchAccountsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..pc<Account>(1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.PM, subBuilder: Account.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SearchAccountsResponse clone() => SearchAccountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SearchAccountsResponse copyWith(void Function(SearchAccountsResponse) updates) => super.copyWith((message) => updates(message as SearchAccountsResponse)) as SearchAccountsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SearchAccountsResponse create() => SearchAccountsResponse._();
  SearchAccountsResponse createEmptyInstance() => create();
  static $pb.PbList<SearchAccountsResponse> createRepeated() => $pb.PbList<SearchAccountsResponse>();
  @$core.pragma('dart2js:noInline')
  static SearchAccountsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SearchAccountsResponse>(create);
  static SearchAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Account> get data => $_getList(0);
}

/// CreateAccountRequest creates a new account within a ledger.
class CreateAccountRequest extends $pb.GeneratedMessage {
  factory CreateAccountRequest({
    $core.String? id,
    $core.String? ledgerId,
    $core.String? currency,
    $6.Struct? data,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (ledgerId != null) {
      $result.ledgerId = ledgerId;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  CreateAccountRequest._() : super();
  factory CreateAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'ledgerId')
    ..aOS(4, _omitFieldNames ? '' : 'currency')
    ..aOM<$6.Struct>(5, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateAccountRequest clone() => CreateAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateAccountRequest copyWith(void Function(CreateAccountRequest) updates) => super.copyWith((message) => updates(message as CreateAccountRequest)) as CreateAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateAccountRequest create() => CreateAccountRequest._();
  CreateAccountRequest createEmptyInstance() => create();
  static $pb.PbList<CreateAccountRequest> createRepeated() => $pb.PbList<CreateAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateAccountRequest>(create);
  static CreateAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(3)
  $core.String get ledgerId => $_getSZ(1);
  @$pb.TagNumber(3)
  set ledgerId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(3)
  void clearLedgerId() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(2);
  @$pb.TagNumber(4)
  set currency($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(2);
  @$pb.TagNumber(4)
  void clearCurrency() => clearField(4);

  @$pb.TagNumber(5)
  $6.Struct get data => $_getN(3);
  @$pb.TagNumber(5)
  set data($6.Struct v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(5)
  void clearData() => clearField(5);
  @$pb.TagNumber(5)
  $6.Struct ensureData() => $_ensure(3);
}

/// CreateAccountResponse returns the newly created account.
class CreateAccountResponse extends $pb.GeneratedMessage {
  factory CreateAccountResponse({
    Account? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  CreateAccountResponse._() : super();
  factory CreateAccountResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateAccountResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateAccountResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Account>(1, _omitFieldNames ? '' : 'data', subBuilder: Account.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateAccountResponse clone() => CreateAccountResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateAccountResponse copyWith(void Function(CreateAccountResponse) updates) => super.copyWith((message) => updates(message as CreateAccountResponse)) as CreateAccountResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateAccountResponse create() => CreateAccountResponse._();
  CreateAccountResponse createEmptyInstance() => create();
  static $pb.PbList<CreateAccountResponse> createRepeated() => $pb.PbList<CreateAccountResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateAccountResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateAccountResponse>(create);
  static CreateAccountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Account get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Account v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Account ensureData() => $_ensure(0);
}

/// UpdateAccountRequest updates an existing account's metadata.
class UpdateAccountRequest extends $pb.GeneratedMessage {
  factory UpdateAccountRequest({
    $core.String? id,
    $6.Struct? data,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  UpdateAccountRequest._() : super();
  factory UpdateAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateAccountRequest clone() => UpdateAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateAccountRequest copyWith(void Function(UpdateAccountRequest) updates) => super.copyWith((message) => updates(message as UpdateAccountRequest)) as UpdateAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateAccountRequest create() => UpdateAccountRequest._();
  UpdateAccountRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateAccountRequest> createRepeated() => $pb.PbList<UpdateAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateAccountRequest>(create);
  static UpdateAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(4)
  $6.Struct get data => $_getN(1);
  @$pb.TagNumber(4)
  set data($6.Struct v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureData() => $_ensure(1);
}

/// UpdateAccountResponse returns the updated account.
class UpdateAccountResponse extends $pb.GeneratedMessage {
  factory UpdateAccountResponse({
    Account? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  UpdateAccountResponse._() : super();
  factory UpdateAccountResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateAccountResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateAccountResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Account>(1, _omitFieldNames ? '' : 'data', subBuilder: Account.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateAccountResponse clone() => UpdateAccountResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateAccountResponse copyWith(void Function(UpdateAccountResponse) updates) => super.copyWith((message) => updates(message as UpdateAccountResponse)) as UpdateAccountResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateAccountResponse create() => UpdateAccountResponse._();
  UpdateAccountResponse createEmptyInstance() => create();
  static $pb.PbList<UpdateAccountResponse> createRepeated() => $pb.PbList<UpdateAccountResponse>();
  @$core.pragma('dart2js:noInline')
  static UpdateAccountResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateAccountResponse>(create);
  static UpdateAccountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Account get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Account v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Account ensureData() => $_ensure(0);
}

/// SearchTransactionsResponse returns transactions matching search criteria.
class SearchTransactionsResponse extends $pb.GeneratedMessage {
  factory SearchTransactionsResponse({
    $core.Iterable<Transaction>? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data.addAll(data);
    }
    return $result;
  }
  SearchTransactionsResponse._() : super();
  factory SearchTransactionsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SearchTransactionsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SearchTransactionsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..pc<Transaction>(1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.PM, subBuilder: Transaction.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SearchTransactionsResponse clone() => SearchTransactionsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SearchTransactionsResponse copyWith(void Function(SearchTransactionsResponse) updates) => super.copyWith((message) => updates(message as SearchTransactionsResponse)) as SearchTransactionsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SearchTransactionsResponse create() => SearchTransactionsResponse._();
  SearchTransactionsResponse createEmptyInstance() => create();
  static $pb.PbList<SearchTransactionsResponse> createRepeated() => $pb.PbList<SearchTransactionsResponse>();
  @$core.pragma('dart2js:noInline')
  static SearchTransactionsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SearchTransactionsResponse>(create);
  static SearchTransactionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Transaction> get data => $_getList(0);
}

/// CreateTransactionRequest creates a new double-entry transaction.
class CreateTransactionRequest extends $pb.GeneratedMessage {
  factory CreateTransactionRequest({
    $core.String? id,
    $core.String? currency,
    $core.String? transactedAt,
    $6.Struct? data,
    $core.Iterable<TransactionEntry>? entries,
    $core.bool? cleared,
    TransactionType? type,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (transactedAt != null) {
      $result.transactedAt = transactedAt;
    }
    if (data != null) {
      $result.data = data;
    }
    if (entries != null) {
      $result.entries.addAll(entries);
    }
    if (cleared != null) {
      $result.cleared = cleared;
    }
    if (type != null) {
      $result.type = type;
    }
    return $result;
  }
  CreateTransactionRequest._() : super();
  factory CreateTransactionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateTransactionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateTransactionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'currency')
    ..aOS(3, _omitFieldNames ? '' : 'transactedAt')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..pc<TransactionEntry>(5, _omitFieldNames ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: TransactionEntry.create)
    ..aOB(6, _omitFieldNames ? '' : 'cleared')
    ..e<TransactionType>(7, _omitFieldNames ? '' : 'type', $pb.PbFieldType.OE, defaultOrMaker: TransactionType.NORMAL, valueOf: TransactionType.valueOf, enumValues: TransactionType.values)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateTransactionRequest clone() => CreateTransactionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateTransactionRequest copyWith(void Function(CreateTransactionRequest) updates) => super.copyWith((message) => updates(message as CreateTransactionRequest)) as CreateTransactionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateTransactionRequest create() => CreateTransactionRequest._();
  CreateTransactionRequest createEmptyInstance() => create();
  static $pb.PbList<CreateTransactionRequest> createRepeated() => $pb.PbList<CreateTransactionRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateTransactionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateTransactionRequest>(create);
  static CreateTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get currency => $_getSZ(1);
  @$pb.TagNumber(2)
  set currency($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCurrency() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrency() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get transactedAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set transactedAt($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTransactedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransactedAt() => clearField(3);

  @$pb.TagNumber(4)
  $6.Struct get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($6.Struct v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureData() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.List<TransactionEntry> get entries => $_getList(4);

  @$pb.TagNumber(6)
  $core.bool get cleared => $_getBF(5);
  @$pb.TagNumber(6)
  set cleared($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasCleared() => $_has(5);
  @$pb.TagNumber(6)
  void clearCleared() => clearField(6);

  @$pb.TagNumber(7)
  TransactionType get type => $_getN(6);
  @$pb.TagNumber(7)
  set type(TransactionType v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasType() => $_has(6);
  @$pb.TagNumber(7)
  void clearType() => clearField(7);
}

/// CreateTransactionResponse returns the newly created transaction.
class CreateTransactionResponse extends $pb.GeneratedMessage {
  factory CreateTransactionResponse({
    Transaction? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  CreateTransactionResponse._() : super();
  factory CreateTransactionResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateTransactionResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateTransactionResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data', subBuilder: Transaction.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateTransactionResponse clone() => CreateTransactionResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateTransactionResponse copyWith(void Function(CreateTransactionResponse) updates) => super.copyWith((message) => updates(message as CreateTransactionResponse)) as CreateTransactionResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateTransactionResponse create() => CreateTransactionResponse._();
  CreateTransactionResponse createEmptyInstance() => create();
  static $pb.PbList<CreateTransactionResponse> createRepeated() => $pb.PbList<CreateTransactionResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateTransactionResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateTransactionResponse>(create);
  static CreateTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

/// ReverseTransactionRequest reverses a transaction by creating offsetting entries.
class ReverseTransactionRequest extends $pb.GeneratedMessage {
  factory ReverseTransactionRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  ReverseTransactionRequest._() : super();
  factory ReverseTransactionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ReverseTransactionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ReverseTransactionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ReverseTransactionRequest clone() => ReverseTransactionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ReverseTransactionRequest copyWith(void Function(ReverseTransactionRequest) updates) => super.copyWith((message) => updates(message as ReverseTransactionRequest)) as ReverseTransactionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ReverseTransactionRequest create() => ReverseTransactionRequest._();
  ReverseTransactionRequest createEmptyInstance() => create();
  static $pb.PbList<ReverseTransactionRequest> createRepeated() => $pb.PbList<ReverseTransactionRequest>();
  @$core.pragma('dart2js:noInline')
  static ReverseTransactionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ReverseTransactionRequest>(create);
  static ReverseTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

/// ReverseTransactionResponse returns the reversal transaction.
class ReverseTransactionResponse extends $pb.GeneratedMessage {
  factory ReverseTransactionResponse({
    Transaction? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  ReverseTransactionResponse._() : super();
  factory ReverseTransactionResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ReverseTransactionResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ReverseTransactionResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data', subBuilder: Transaction.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ReverseTransactionResponse clone() => ReverseTransactionResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ReverseTransactionResponse copyWith(void Function(ReverseTransactionResponse) updates) => super.copyWith((message) => updates(message as ReverseTransactionResponse)) as ReverseTransactionResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ReverseTransactionResponse create() => ReverseTransactionResponse._();
  ReverseTransactionResponse createEmptyInstance() => create();
  static $pb.PbList<ReverseTransactionResponse> createRepeated() => $pb.PbList<ReverseTransactionResponse>();
  @$core.pragma('dart2js:noInline')
  static ReverseTransactionResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ReverseTransactionResponse>(create);
  static ReverseTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

/// UpdateTransactionRequest updates a transaction's metadata.
class UpdateTransactionRequest extends $pb.GeneratedMessage {
  factory UpdateTransactionRequest({
    $core.String? id,
    $core.String? clearedAt,
    $6.Struct? data,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (clearedAt != null) {
      $result.clearedAt = clearedAt;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  UpdateTransactionRequest._() : super();
  factory UpdateTransactionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateTransactionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateTransactionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'clearedAt')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateTransactionRequest clone() => UpdateTransactionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateTransactionRequest copyWith(void Function(UpdateTransactionRequest) updates) => super.copyWith((message) => updates(message as UpdateTransactionRequest)) as UpdateTransactionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateTransactionRequest create() => UpdateTransactionRequest._();
  UpdateTransactionRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateTransactionRequest> createRepeated() => $pb.PbList<UpdateTransactionRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateTransactionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateTransactionRequest>(create);
  static UpdateTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get clearedAt => $_getSZ(1);
  @$pb.TagNumber(2)
  set clearedAt($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasClearedAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearClearedAt() => clearField(2);

  @$pb.TagNumber(4)
  $6.Struct get data => $_getN(2);
  @$pb.TagNumber(4)
  set data($6.Struct v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(2);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureData() => $_ensure(2);
}

/// UpdateTransactionResponse returns the updated transaction.
class UpdateTransactionResponse extends $pb.GeneratedMessage {
  factory UpdateTransactionResponse({
    Transaction? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  UpdateTransactionResponse._() : super();
  factory UpdateTransactionResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateTransactionResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateTransactionResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data', subBuilder: Transaction.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateTransactionResponse clone() => UpdateTransactionResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateTransactionResponse copyWith(void Function(UpdateTransactionResponse) updates) => super.copyWith((message) => updates(message as UpdateTransactionResponse)) as UpdateTransactionResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateTransactionResponse create() => UpdateTransactionResponse._();
  UpdateTransactionResponse createEmptyInstance() => create();
  static $pb.PbList<UpdateTransactionResponse> createRepeated() => $pb.PbList<UpdateTransactionResponse>();
  @$core.pragma('dart2js:noInline')
  static UpdateTransactionResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateTransactionResponse>(create);
  static UpdateTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

/// SearchTransactionEntriesResponse returns transaction entries matching search criteria.
class SearchTransactionEntriesResponse extends $pb.GeneratedMessage {
  factory SearchTransactionEntriesResponse({
    $core.Iterable<TransactionEntry>? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data.addAll(data);
    }
    return $result;
  }
  SearchTransactionEntriesResponse._() : super();
  factory SearchTransactionEntriesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SearchTransactionEntriesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SearchTransactionEntriesResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..pc<TransactionEntry>(1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.PM, subBuilder: TransactionEntry.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SearchTransactionEntriesResponse clone() => SearchTransactionEntriesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SearchTransactionEntriesResponse copyWith(void Function(SearchTransactionEntriesResponse) updates) => super.copyWith((message) => updates(message as SearchTransactionEntriesResponse)) as SearchTransactionEntriesResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SearchTransactionEntriesResponse create() => SearchTransactionEntriesResponse._();
  SearchTransactionEntriesResponse createEmptyInstance() => create();
  static $pb.PbList<SearchTransactionEntriesResponse> createRepeated() => $pb.PbList<SearchTransactionEntriesResponse>();
  @$core.pragma('dart2js:noInline')
  static SearchTransactionEntriesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SearchTransactionEntriesResponse>(create);
  static SearchTransactionEntriesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<TransactionEntry> get data => $_getList(0);
}

class CreateBookRequest extends $pb.GeneratedMessage {
  factory CreateBookRequest({
    $core.String? id,
    $core.String? name,
    $core.String? type,
    $core.String? parentId,
    $core.String? currency,
    $6.Struct? data,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (name != null) {
      $result.name = name;
    }
    if (type != null) {
      $result.type = type;
    }
    if (parentId != null) {
      $result.parentId = parentId;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  CreateBookRequest._() : super();
  factory CreateBookRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateBookRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateBookRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOS(4, _omitFieldNames ? '' : 'parentId')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$6.Struct>(6, _omitFieldNames ? '' : 'data', subBuilder: $6.Struct.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateBookRequest clone() => CreateBookRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateBookRequest copyWith(void Function(CreateBookRequest) updates) => super.copyWith((message) => updates(message as CreateBookRequest)) as CreateBookRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateBookRequest create() => CreateBookRequest._();
  CreateBookRequest createEmptyInstance() => create();
  static $pb.PbList<CreateBookRequest> createRepeated() => $pb.PbList<CreateBookRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateBookRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateBookRequest>(create);
  static CreateBookRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get parentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set parentId($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasParentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearParentId() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => clearField(5);

  @$pb.TagNumber(6)
  $6.Struct get data => $_getN(5);
  @$pb.TagNumber(6)
  set data($6.Struct v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasData() => $_has(5);
  @$pb.TagNumber(6)
  void clearData() => clearField(6);
  @$pb.TagNumber(6)
  $6.Struct ensureData() => $_ensure(5);
}

class CreateBookResponse extends $pb.GeneratedMessage {
  factory CreateBookResponse({
    Book? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  CreateBookResponse._() : super();
  factory CreateBookResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateBookResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateBookResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Book>(1, _omitFieldNames ? '' : 'data', subBuilder: Book.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateBookResponse clone() => CreateBookResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateBookResponse copyWith(void Function(CreateBookResponse) updates) => super.copyWith((message) => updates(message as CreateBookResponse)) as CreateBookResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateBookResponse create() => CreateBookResponse._();
  CreateBookResponse createEmptyInstance() => create();
  static $pb.PbList<CreateBookResponse> createRepeated() => $pb.PbList<CreateBookResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateBookResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateBookResponse>(create);
  static CreateBookResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Book get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Book v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Book ensureData() => $_ensure(0);
}

class GetBookRequest extends $pb.GeneratedMessage {
  factory GetBookRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetBookRequest._() : super();
  factory GetBookRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetBookRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetBookRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetBookRequest clone() => GetBookRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetBookRequest copyWith(void Function(GetBookRequest) updates) => super.copyWith((message) => updates(message as GetBookRequest)) as GetBookRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetBookRequest create() => GetBookRequest._();
  GetBookRequest createEmptyInstance() => create();
  static $pb.PbList<GetBookRequest> createRepeated() => $pb.PbList<GetBookRequest>();
  @$core.pragma('dart2js:noInline')
  static GetBookRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetBookRequest>(create);
  static GetBookRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class GetBookResponse extends $pb.GeneratedMessage {
  factory GetBookResponse({
    Book? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  GetBookResponse._() : super();
  factory GetBookResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetBookResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetBookResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Book>(1, _omitFieldNames ? '' : 'data', subBuilder: Book.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetBookResponse clone() => GetBookResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetBookResponse copyWith(void Function(GetBookResponse) updates) => super.copyWith((message) => updates(message as GetBookResponse)) as GetBookResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetBookResponse create() => GetBookResponse._();
  GetBookResponse createEmptyInstance() => create();
  static $pb.PbList<GetBookResponse> createRepeated() => $pb.PbList<GetBookResponse>();
  @$core.pragma('dart2js:noInline')
  static GetBookResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetBookResponse>(create);
  static GetBookResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Book get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Book v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Book ensureData() => $_ensure(0);
}

class ListBooksByTypeRequest extends $pb.GeneratedMessage {
  factory ListBooksByTypeRequest({
    $core.String? type,
  }) {
    final $result = create();
    if (type != null) {
      $result.type = type;
    }
    return $result;
  }
  ListBooksByTypeRequest._() : super();
  factory ListBooksByTypeRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListBooksByTypeRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListBooksByTypeRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'type')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListBooksByTypeRequest clone() => ListBooksByTypeRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListBooksByTypeRequest copyWith(void Function(ListBooksByTypeRequest) updates) => super.copyWith((message) => updates(message as ListBooksByTypeRequest)) as ListBooksByTypeRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListBooksByTypeRequest create() => ListBooksByTypeRequest._();
  ListBooksByTypeRequest createEmptyInstance() => create();
  static $pb.PbList<ListBooksByTypeRequest> createRepeated() => $pb.PbList<ListBooksByTypeRequest>();
  @$core.pragma('dart2js:noInline')
  static ListBooksByTypeRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListBooksByTypeRequest>(create);
  static ListBooksByTypeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => clearField(1);
}

class ListBooksByTypeResponse extends $pb.GeneratedMessage {
  factory ListBooksByTypeResponse({
    $core.Iterable<Book>? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data.addAll(data);
    }
    return $result;
  }
  ListBooksByTypeResponse._() : super();
  factory ListBooksByTypeResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListBooksByTypeResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListBooksByTypeResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..pc<Book>(1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.PM, subBuilder: Book.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListBooksByTypeResponse clone() => ListBooksByTypeResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListBooksByTypeResponse copyWith(void Function(ListBooksByTypeResponse) updates) => super.copyWith((message) => updates(message as ListBooksByTypeResponse)) as ListBooksByTypeResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListBooksByTypeResponse create() => ListBooksByTypeResponse._();
  ListBooksByTypeResponse createEmptyInstance() => create();
  static $pb.PbList<ListBooksByTypeResponse> createRepeated() => $pb.PbList<ListBooksByTypeResponse>();
  @$core.pragma('dart2js:noInline')
  static ListBooksByTypeResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListBooksByTypeResponse>(create);
  static ListBooksByTypeResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Book> get data => $_getList(0);
}

class GetTrialBalanceRequest extends $pb.GeneratedMessage {
  factory GetTrialBalanceRequest({
    $core.String? currency,
    $core.String? ledgerId,
    $core.String? ledgerType,
    $core.Iterable<$core.String>? bookIds,
    $core.String? asOf,
  }) {
    final $result = create();
    if (currency != null) {
      $result.currency = currency;
    }
    if (ledgerId != null) {
      $result.ledgerId = ledgerId;
    }
    if (ledgerType != null) {
      $result.ledgerType = ledgerType;
    }
    if (bookIds != null) {
      $result.bookIds.addAll(bookIds);
    }
    if (asOf != null) {
      $result.asOf = asOf;
    }
    return $result;
  }
  GetTrialBalanceRequest._() : super();
  factory GetTrialBalanceRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetTrialBalanceRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetTrialBalanceRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'currency')
    ..aOS(2, _omitFieldNames ? '' : 'ledgerId')
    ..aOS(3, _omitFieldNames ? '' : 'ledgerType')
    ..pPS(4, _omitFieldNames ? '' : 'bookIds')
    ..aOS(5, _omitFieldNames ? '' : 'asOf')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetTrialBalanceRequest clone() => GetTrialBalanceRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetTrialBalanceRequest copyWith(void Function(GetTrialBalanceRequest) updates) => super.copyWith((message) => updates(message as GetTrialBalanceRequest)) as GetTrialBalanceRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTrialBalanceRequest create() => GetTrialBalanceRequest._();
  GetTrialBalanceRequest createEmptyInstance() => create();
  static $pb.PbList<GetTrialBalanceRequest> createRepeated() => $pb.PbList<GetTrialBalanceRequest>();
  @$core.pragma('dart2js:noInline')
  static GetTrialBalanceRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTrialBalanceRequest>(create);
  static GetTrialBalanceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get currency => $_getSZ(0);
  @$pb.TagNumber(1)
  set currency($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCurrency() => $_has(0);
  @$pb.TagNumber(1)
  void clearCurrency() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get ledgerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set ledgerId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearLedgerId() => clearField(2);

  /// ledger_type is a case-sensitive string filter ("ASSET", "LIABILITY",
  /// "INCOME", "EXPENSE", "CAPITAL"). Empty means "no filter". A string is
  /// used instead of the LedgerType enum so the caller can omit the filter
  /// entirely — proto3 enums treat the zero value as both "unset" and the
  /// first enum constant (ASSET in this codebase), which is ambiguous.
  @$pb.TagNumber(3)
  $core.String get ledgerType => $_getSZ(2);
  @$pb.TagNumber(3)
  set ledgerType($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasLedgerType() => $_has(2);
  @$pb.TagNumber(3)
  void clearLedgerType() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$core.String> get bookIds => $_getList(3);

  @$pb.TagNumber(5)
  $core.String get asOf => $_getSZ(4);
  @$pb.TagNumber(5)
  set asOf($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasAsOf() => $_has(4);
  @$pb.TagNumber(5)
  void clearAsOf() => clearField(5);
}

class GetTrialBalanceResponse extends $pb.GeneratedMessage {
  factory GetTrialBalanceResponse({
    $core.Iterable<TrialBalanceLine>? lines,
    $core.Iterable<TrialBalanceTotal>? totals,
  }) {
    final $result = create();
    if (lines != null) {
      $result.lines.addAll(lines);
    }
    if (totals != null) {
      $result.totals.addAll(totals);
    }
    return $result;
  }
  GetTrialBalanceResponse._() : super();
  factory GetTrialBalanceResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetTrialBalanceResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetTrialBalanceResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..pc<TrialBalanceLine>(1, _omitFieldNames ? '' : 'lines', $pb.PbFieldType.PM, subBuilder: TrialBalanceLine.create)
    ..pc<TrialBalanceTotal>(2, _omitFieldNames ? '' : 'totals', $pb.PbFieldType.PM, subBuilder: TrialBalanceTotal.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetTrialBalanceResponse clone() => GetTrialBalanceResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetTrialBalanceResponse copyWith(void Function(GetTrialBalanceResponse) updates) => super.copyWith((message) => updates(message as GetTrialBalanceResponse)) as GetTrialBalanceResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTrialBalanceResponse create() => GetTrialBalanceResponse._();
  GetTrialBalanceResponse createEmptyInstance() => create();
  static $pb.PbList<GetTrialBalanceResponse> createRepeated() => $pb.PbList<GetTrialBalanceResponse>();
  @$core.pragma('dart2js:noInline')
  static GetTrialBalanceResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTrialBalanceResponse>(create);
  static GetTrialBalanceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<TrialBalanceLine> get lines => $_getList(0);

  @$pb.TagNumber(2)
  $core.List<TrialBalanceTotal> get totals => $_getList(1);
}

class GetAccountStatementRequest extends $pb.GeneratedMessage {
  factory GetAccountStatementRequest({
    $core.String? accountId,
    $core.String? from,
    $core.String? to,
    $core.int? limit,
    $core.int? offset,
  }) {
    final $result = create();
    if (accountId != null) {
      $result.accountId = accountId;
    }
    if (from != null) {
      $result.from = from;
    }
    if (to != null) {
      $result.to = to;
    }
    if (limit != null) {
      $result.limit = limit;
    }
    if (offset != null) {
      $result.offset = offset;
    }
    return $result;
  }
  GetAccountStatementRequest._() : super();
  factory GetAccountStatementRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetAccountStatementRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetAccountStatementRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'from')
    ..aOS(3, _omitFieldNames ? '' : 'to')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'limit', $pb.PbFieldType.O3)
    ..a<$core.int>(5, _omitFieldNames ? '' : 'offset', $pb.PbFieldType.O3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetAccountStatementRequest clone() => GetAccountStatementRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetAccountStatementRequest copyWith(void Function(GetAccountStatementRequest) updates) => super.copyWith((message) => updates(message as GetAccountStatementRequest)) as GetAccountStatementRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetAccountStatementRequest create() => GetAccountStatementRequest._();
  GetAccountStatementRequest createEmptyInstance() => create();
  static $pb.PbList<GetAccountStatementRequest> createRepeated() => $pb.PbList<GetAccountStatementRequest>();
  @$core.pragma('dart2js:noInline')
  static GetAccountStatementRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetAccountStatementRequest>(create);
  static GetAccountStatementRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get from => $_getSZ(1);
  @$pb.TagNumber(2)
  set from($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasFrom() => $_has(1);
  @$pb.TagNumber(2)
  void clearFrom() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get to => $_getSZ(2);
  @$pb.TagNumber(3)
  set to($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTo() => $_has(2);
  @$pb.TagNumber(3)
  void clearTo() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get limit => $_getIZ(3);
  @$pb.TagNumber(4)
  set limit($core.int v) { $_setSignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasLimit() => $_has(3);
  @$pb.TagNumber(4)
  void clearLimit() => clearField(4);

  @$pb.TagNumber(5)
  $core.int get offset => $_getIZ(4);
  @$pb.TagNumber(5)
  set offset($core.int v) { $_setSignedInt32(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasOffset() => $_has(4);
  @$pb.TagNumber(5)
  void clearOffset() => clearField(5);
}

class GetAccountStatementResponse extends $pb.GeneratedMessage {
  factory GetAccountStatementResponse({
    $core.String? accountId,
    $core.String? currency,
    $7.Money? openingBalance,
    $7.Money? closingBalance,
    $7.Money? totalDebits,
    $7.Money? totalCredits,
    $core.Iterable<StatementEntry>? entries,
  }) {
    final $result = create();
    if (accountId != null) {
      $result.accountId = accountId;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (openingBalance != null) {
      $result.openingBalance = openingBalance;
    }
    if (closingBalance != null) {
      $result.closingBalance = closingBalance;
    }
    if (totalDebits != null) {
      $result.totalDebits = totalDebits;
    }
    if (totalCredits != null) {
      $result.totalCredits = totalCredits;
    }
    if (entries != null) {
      $result.entries.addAll(entries);
    }
    return $result;
  }
  GetAccountStatementResponse._() : super();
  factory GetAccountStatementResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetAccountStatementResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetAccountStatementResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'currency')
    ..aOM<$7.Money>(3, _omitFieldNames ? '' : 'openingBalance', subBuilder: $7.Money.create)
    ..aOM<$7.Money>(4, _omitFieldNames ? '' : 'closingBalance', subBuilder: $7.Money.create)
    ..aOM<$7.Money>(5, _omitFieldNames ? '' : 'totalDebits', subBuilder: $7.Money.create)
    ..aOM<$7.Money>(6, _omitFieldNames ? '' : 'totalCredits', subBuilder: $7.Money.create)
    ..pc<StatementEntry>(7, _omitFieldNames ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: StatementEntry.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetAccountStatementResponse clone() => GetAccountStatementResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetAccountStatementResponse copyWith(void Function(GetAccountStatementResponse) updates) => super.copyWith((message) => updates(message as GetAccountStatementResponse)) as GetAccountStatementResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetAccountStatementResponse create() => GetAccountStatementResponse._();
  GetAccountStatementResponse createEmptyInstance() => create();
  static $pb.PbList<GetAccountStatementResponse> createRepeated() => $pb.PbList<GetAccountStatementResponse>();
  @$core.pragma('dart2js:noInline')
  static GetAccountStatementResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetAccountStatementResponse>(create);
  static GetAccountStatementResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get currency => $_getSZ(1);
  @$pb.TagNumber(2)
  set currency($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCurrency() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrency() => clearField(2);

  @$pb.TagNumber(3)
  $7.Money get openingBalance => $_getN(2);
  @$pb.TagNumber(3)
  set openingBalance($7.Money v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasOpeningBalance() => $_has(2);
  @$pb.TagNumber(3)
  void clearOpeningBalance() => clearField(3);
  @$pb.TagNumber(3)
  $7.Money ensureOpeningBalance() => $_ensure(2);

  @$pb.TagNumber(4)
  $7.Money get closingBalance => $_getN(3);
  @$pb.TagNumber(4)
  set closingBalance($7.Money v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasClosingBalance() => $_has(3);
  @$pb.TagNumber(4)
  void clearClosingBalance() => clearField(4);
  @$pb.TagNumber(4)
  $7.Money ensureClosingBalance() => $_ensure(3);

  @$pb.TagNumber(5)
  $7.Money get totalDebits => $_getN(4);
  @$pb.TagNumber(5)
  set totalDebits($7.Money v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasTotalDebits() => $_has(4);
  @$pb.TagNumber(5)
  void clearTotalDebits() => clearField(5);
  @$pb.TagNumber(5)
  $7.Money ensureTotalDebits() => $_ensure(4);

  @$pb.TagNumber(6)
  $7.Money get totalCredits => $_getN(5);
  @$pb.TagNumber(6)
  set totalCredits($7.Money v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasTotalCredits() => $_has(5);
  @$pb.TagNumber(6)
  void clearTotalCredits() => clearField(6);
  @$pb.TagNumber(6)
  $7.Money ensureTotalCredits() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.List<StatementEntry> get entries => $_getList(6);
}

class VoidTransactionRequest extends $pb.GeneratedMessage {
  factory VoidTransactionRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  VoidTransactionRequest._() : super();
  factory VoidTransactionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VoidTransactionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'VoidTransactionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VoidTransactionRequest clone() => VoidTransactionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VoidTransactionRequest copyWith(void Function(VoidTransactionRequest) updates) => super.copyWith((message) => updates(message as VoidTransactionRequest)) as VoidTransactionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VoidTransactionRequest create() => VoidTransactionRequest._();
  VoidTransactionRequest createEmptyInstance() => create();
  static $pb.PbList<VoidTransactionRequest> createRepeated() => $pb.PbList<VoidTransactionRequest>();
  @$core.pragma('dart2js:noInline')
  static VoidTransactionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VoidTransactionRequest>(create);
  static VoidTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class VoidTransactionResponse extends $pb.GeneratedMessage {
  factory VoidTransactionResponse({
    Transaction? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  VoidTransactionResponse._() : super();
  factory VoidTransactionResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VoidTransactionResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'VoidTransactionResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data', subBuilder: Transaction.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VoidTransactionResponse clone() => VoidTransactionResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VoidTransactionResponse copyWith(void Function(VoidTransactionResponse) updates) => super.copyWith((message) => updates(message as VoidTransactionResponse)) as VoidTransactionResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VoidTransactionResponse create() => VoidTransactionResponse._();
  VoidTransactionResponse createEmptyInstance() => create();
  static $pb.PbList<VoidTransactionResponse> createRepeated() => $pb.PbList<VoidTransactionResponse>();
  @$core.pragma('dart2js:noInline')
  static VoidTransactionResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VoidTransactionResponse>(create);
  static VoidTransactionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

class MarkTransactionFailedRequest extends $pb.GeneratedMessage {
  factory MarkTransactionFailedRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  MarkTransactionFailedRequest._() : super();
  factory MarkTransactionFailedRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MarkTransactionFailedRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'MarkTransactionFailedRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MarkTransactionFailedRequest clone() => MarkTransactionFailedRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MarkTransactionFailedRequest copyWith(void Function(MarkTransactionFailedRequest) updates) => super.copyWith((message) => updates(message as MarkTransactionFailedRequest)) as MarkTransactionFailedRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MarkTransactionFailedRequest create() => MarkTransactionFailedRequest._();
  MarkTransactionFailedRequest createEmptyInstance() => create();
  static $pb.PbList<MarkTransactionFailedRequest> createRepeated() => $pb.PbList<MarkTransactionFailedRequest>();
  @$core.pragma('dart2js:noInline')
  static MarkTransactionFailedRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MarkTransactionFailedRequest>(create);
  static MarkTransactionFailedRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class MarkTransactionFailedResponse extends $pb.GeneratedMessage {
  factory MarkTransactionFailedResponse({
    Transaction? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  MarkTransactionFailedResponse._() : super();
  factory MarkTransactionFailedResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MarkTransactionFailedResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'MarkTransactionFailedResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'ledger.v1'), createEmptyInstance: create)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'data', subBuilder: Transaction.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MarkTransactionFailedResponse clone() => MarkTransactionFailedResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MarkTransactionFailedResponse copyWith(void Function(MarkTransactionFailedResponse) updates) => super.copyWith((message) => updates(message as MarkTransactionFailedResponse)) as MarkTransactionFailedResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MarkTransactionFailedResponse create() => MarkTransactionFailedResponse._();
  MarkTransactionFailedResponse createEmptyInstance() => create();
  static $pb.PbList<MarkTransactionFailedResponse> createRepeated() => $pb.PbList<MarkTransactionFailedResponse>();
  @$core.pragma('dart2js:noInline')
  static MarkTransactionFailedResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MarkTransactionFailedResponse>(create);
  static MarkTransactionFailedResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Transaction v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureData() => $_ensure(0);
}

class LedgerServiceApi {
  $pb.RpcClient _client;
  LedgerServiceApi(this._client);

  $async.Future<SearchLedgersResponse> searchLedgers($pb.ClientContext? ctx, $8.SearchRequest request) =>
    _client.invoke<SearchLedgersResponse>(ctx, 'LedgerService', 'SearchLedgers', request, SearchLedgersResponse())
  ;
  $async.Future<CreateLedgerResponse> createLedger($pb.ClientContext? ctx, CreateLedgerRequest request) =>
    _client.invoke<CreateLedgerResponse>(ctx, 'LedgerService', 'CreateLedger', request, CreateLedgerResponse())
  ;
  $async.Future<UpdateLedgerResponse> updateLedger($pb.ClientContext? ctx, UpdateLedgerRequest request) =>
    _client.invoke<UpdateLedgerResponse>(ctx, 'LedgerService', 'UpdateLedger', request, UpdateLedgerResponse())
  ;
  $async.Future<SearchAccountsResponse> searchAccounts($pb.ClientContext? ctx, $8.SearchRequest request) =>
    _client.invoke<SearchAccountsResponse>(ctx, 'LedgerService', 'SearchAccounts', request, SearchAccountsResponse())
  ;
  $async.Future<CreateAccountResponse> createAccount($pb.ClientContext? ctx, CreateAccountRequest request) =>
    _client.invoke<CreateAccountResponse>(ctx, 'LedgerService', 'CreateAccount', request, CreateAccountResponse())
  ;
  $async.Future<UpdateAccountResponse> updateAccount($pb.ClientContext? ctx, UpdateAccountRequest request) =>
    _client.invoke<UpdateAccountResponse>(ctx, 'LedgerService', 'UpdateAccount', request, UpdateAccountResponse())
  ;
  $async.Future<SearchTransactionsResponse> searchTransactions($pb.ClientContext? ctx, $8.SearchRequest request) =>
    _client.invoke<SearchTransactionsResponse>(ctx, 'LedgerService', 'SearchTransactions', request, SearchTransactionsResponse())
  ;
  $async.Future<CreateTransactionResponse> createTransaction($pb.ClientContext? ctx, CreateTransactionRequest request) =>
    _client.invoke<CreateTransactionResponse>(ctx, 'LedgerService', 'CreateTransaction', request, CreateTransactionResponse())
  ;
  $async.Future<ReverseTransactionResponse> reverseTransaction($pb.ClientContext? ctx, ReverseTransactionRequest request) =>
    _client.invoke<ReverseTransactionResponse>(ctx, 'LedgerService', 'ReverseTransaction', request, ReverseTransactionResponse())
  ;
  $async.Future<UpdateTransactionResponse> updateTransaction($pb.ClientContext? ctx, UpdateTransactionRequest request) =>
    _client.invoke<UpdateTransactionResponse>(ctx, 'LedgerService', 'UpdateTransaction', request, UpdateTransactionResponse())
  ;
  $async.Future<SearchTransactionEntriesResponse> searchTransactionEntries($pb.ClientContext? ctx, $8.SearchRequest request) =>
    _client.invoke<SearchTransactionEntriesResponse>(ctx, 'LedgerService', 'SearchTransactionEntries', request, SearchTransactionEntriesResponse())
  ;
  $async.Future<VoidTransactionResponse> voidTransaction($pb.ClientContext? ctx, VoidTransactionRequest request) =>
    _client.invoke<VoidTransactionResponse>(ctx, 'LedgerService', 'VoidTransaction', request, VoidTransactionResponse())
  ;
  $async.Future<MarkTransactionFailedResponse> markTransactionFailed($pb.ClientContext? ctx, MarkTransactionFailedRequest request) =>
    _client.invoke<MarkTransactionFailedResponse>(ctx, 'LedgerService', 'MarkTransactionFailed', request, MarkTransactionFailedResponse())
  ;
  $async.Future<CreateBookResponse> createBook($pb.ClientContext? ctx, CreateBookRequest request) =>
    _client.invoke<CreateBookResponse>(ctx, 'LedgerService', 'CreateBook', request, CreateBookResponse())
  ;
  $async.Future<GetBookResponse> getBook($pb.ClientContext? ctx, GetBookRequest request) =>
    _client.invoke<GetBookResponse>(ctx, 'LedgerService', 'GetBook', request, GetBookResponse())
  ;
  $async.Future<ListBooksByTypeResponse> listBooksByType($pb.ClientContext? ctx, ListBooksByTypeRequest request) =>
    _client.invoke<ListBooksByTypeResponse>(ctx, 'LedgerService', 'ListBooksByType', request, ListBooksByTypeResponse())
  ;
  $async.Future<GetTrialBalanceResponse> getTrialBalance($pb.ClientContext? ctx, GetTrialBalanceRequest request) =>
    _client.invoke<GetTrialBalanceResponse>(ctx, 'LedgerService', 'GetTrialBalance', request, GetTrialBalanceResponse())
  ;
  $async.Future<GetAccountStatementResponse> getAccountStatement($pb.ClientContext? ctx, GetAccountStatementRequest request) =>
    _client.invoke<GetAccountStatementResponse>(ctx, 'LedgerService', 'GetAccountStatement', request, GetAccountStatementResponse())
  ;
}


const _omitFieldNames = $core.bool.fromEnvironment('protobuf.omit_field_names');
const _omitMessageNames = $core.bool.fromEnvironment('protobuf.omit_message_names');
