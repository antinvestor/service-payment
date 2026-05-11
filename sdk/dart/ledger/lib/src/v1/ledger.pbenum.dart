//
//  Generated code. Do not modify.
//  source: v1/ledger.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// LedgerType defines the fundamental accounting categories.
/// Based on standard accounting equation: Assets = Liabilities + Capital + (Income - Expenses)
/// buf:lint:ignore ENUM_VALUE_PREFIX
class LedgerType extends $pb.ProtobufEnum {
  static const LedgerType ASSET = LedgerType._(0, _omitEnumNames ? '' : 'ASSET');
  static const LedgerType LIABILITY = LedgerType._(1, _omitEnumNames ? '' : 'LIABILITY');
  static const LedgerType INCOME = LedgerType._(2, _omitEnumNames ? '' : 'INCOME');
  static const LedgerType EXPENSE = LedgerType._(3, _omitEnumNames ? '' : 'EXPENSE');
  static const LedgerType CAPITAL = LedgerType._(4, _omitEnumNames ? '' : 'CAPITAL');

  static const $core.List<LedgerType> values = <LedgerType> [
    ASSET,
    LIABILITY,
    INCOME,
    EXPENSE,
    CAPITAL,
  ];

  static final $core.Map<$core.int, LedgerType> _byValue = $pb.ProtobufEnum.initByValue(values);
  static LedgerType? valueOf($core.int value) => _byValue[value];

  const LedgerType._($core.int v, $core.String n) : super(v, n);
}

/// TransactionType defines the nature of a transaction.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class TransactionType extends $pb.ProtobufEnum {
  static const TransactionType NORMAL = TransactionType._(0, _omitEnumNames ? '' : 'NORMAL');
  static const TransactionType REVERSAL = TransactionType._(1, _omitEnumNames ? '' : 'REVERSAL');
  static const TransactionType RESERVATION = TransactionType._(2, _omitEnumNames ? '' : 'RESERVATION');

  static const $core.List<TransactionType> values = <TransactionType> [
    NORMAL,
    REVERSAL,
    RESERVATION,
  ];

  static final $core.Map<$core.int, TransactionType> _byValue = $pb.ProtobufEnum.initByValue(values);
  static TransactionType? valueOf($core.int value) => _byValue[value];

  const TransactionType._($core.int v, $core.String n) : super(v, n);
}

/// TransactionStatus drives the transaction lifecycle.
///   PENDING  — submitted, awaiting settlement; contributes to un_cleared_balance
///   POSTED   — confirmed/settled; contributes to balance; immutable except for
///              the auto-transition to REVERSED when a REVERSAL is posted
///   REVERSED — was posted, then offset by a REVERSAL whose entries cancel
///              this one's balance impact
///   VOIDED   — was draft or pending, then cancelled administratively
///   FAILED   — posting rejected by upstream system; no balance impact
///   DRAFT    — created but not submitted; no balance impact
/// buf:lint:ignore ENUM_VALUE_PREFIX
class TransactionStatus extends $pb.ProtobufEnum {
  static const TransactionStatus PENDING = TransactionStatus._(0, _omitEnumNames ? '' : 'PENDING');
  static const TransactionStatus POSTED = TransactionStatus._(1, _omitEnumNames ? '' : 'POSTED');
  static const TransactionStatus REVERSED = TransactionStatus._(2, _omitEnumNames ? '' : 'REVERSED');
  static const TransactionStatus VOIDED = TransactionStatus._(3, _omitEnumNames ? '' : 'VOIDED');
  static const TransactionStatus FAILED = TransactionStatus._(4, _omitEnumNames ? '' : 'FAILED');
  static const TransactionStatus DRAFT = TransactionStatus._(5, _omitEnumNames ? '' : 'DRAFT');

  static const $core.List<TransactionStatus> values = <TransactionStatus> [
    PENDING,
    POSTED,
    REVERSED,
    VOIDED,
    FAILED,
    DRAFT,
  ];

  static final $core.Map<$core.int, TransactionStatus> _byValue = $pb.ProtobufEnum.initByValue(values);
  static TransactionStatus? valueOf($core.int value) => _byValue[value];

  const TransactionStatus._($core.int v, $core.String n) : super(v, n);
}

/// AccountType is the per-account classification, finer-grained than
/// LedgerType so contra, clearing, suspense and memo accounts can live
/// under their natural parent ledgers without overloading the type.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class AccountType extends $pb.ProtobufEnum {
  static const AccountType ACCOUNT_ASSET = AccountType._(0, _omitEnumNames ? '' : 'ACCOUNT_ASSET');
  static const AccountType ACCOUNT_LIABILITY = AccountType._(1, _omitEnumNames ? '' : 'ACCOUNT_LIABILITY');
  static const AccountType ACCOUNT_EQUITY = AccountType._(2, _omitEnumNames ? '' : 'ACCOUNT_EQUITY');
  static const AccountType ACCOUNT_INCOME = AccountType._(3, _omitEnumNames ? '' : 'ACCOUNT_INCOME');
  static const AccountType ACCOUNT_EXPENSE = AccountType._(4, _omitEnumNames ? '' : 'ACCOUNT_EXPENSE');
  static const AccountType ACCOUNT_CONTRA_ASSET = AccountType._(5, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_ASSET');
  static const AccountType ACCOUNT_CONTRA_LIABILITY = AccountType._(6, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_LIABILITY');
  static const AccountType ACCOUNT_CONTRA_INCOME = AccountType._(7, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_INCOME');
  static const AccountType ACCOUNT_CONTRA_EXPENSE = AccountType._(8, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_EXPENSE');
  static const AccountType ACCOUNT_CLEARING = AccountType._(9, _omitEnumNames ? '' : 'ACCOUNT_CLEARING');
  static const AccountType ACCOUNT_SUSPENSE = AccountType._(10, _omitEnumNames ? '' : 'ACCOUNT_SUSPENSE');
  static const AccountType ACCOUNT_MEMO = AccountType._(11, _omitEnumNames ? '' : 'ACCOUNT_MEMO');

  static const $core.List<AccountType> values = <AccountType> [
    ACCOUNT_ASSET,
    ACCOUNT_LIABILITY,
    ACCOUNT_EQUITY,
    ACCOUNT_INCOME,
    ACCOUNT_EXPENSE,
    ACCOUNT_CONTRA_ASSET,
    ACCOUNT_CONTRA_LIABILITY,
    ACCOUNT_CONTRA_INCOME,
    ACCOUNT_CONTRA_EXPENSE,
    ACCOUNT_CLEARING,
    ACCOUNT_SUSPENSE,
    ACCOUNT_MEMO,
  ];

  static final $core.Map<$core.int, AccountType> _byValue = $pb.ProtobufEnum.initByValue(values);
  static AccountType? valueOf($core.int value) => _byValue[value];

  const AccountType._($core.int v, $core.String n) : super(v, n);
}

/// NormalBalance is the side an account's balance accumulates on. DEADCLIC
/// stores +amount when an entry's side equals the normal balance, -amount
/// when it is the opposite. NONE opts a memo account out of normalisation.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class NormalBalance extends $pb.ProtobufEnum {
  static const NormalBalance DEBIT = NormalBalance._(0, _omitEnumNames ? '' : 'DEBIT');
  static const NormalBalance CREDIT = NormalBalance._(1, _omitEnumNames ? '' : 'CREDIT');
  static const NormalBalance NONE = NormalBalance._(2, _omitEnumNames ? '' : 'NONE');

  static const $core.List<NormalBalance> values = <NormalBalance> [
    DEBIT,
    CREDIT,
    NONE,
  ];

  static final $core.Map<$core.int, NormalBalance> _byValue = $pb.ProtobufEnum.initByValue(values);
  static NormalBalance? valueOf($core.int value) => _byValue[value];

  const NormalBalance._($core.int v, $core.String n) : super(v, n);
}


const _omitEnumNames = $core.bool.fromEnvironment('protobuf.omit_enum_names');
