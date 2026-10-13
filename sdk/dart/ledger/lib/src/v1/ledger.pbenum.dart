// This is a generated file - do not edit.
//
// Generated from v1/ledger.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// LedgerType defines the fundamental accounting categories.
/// Based on standard accounting equation: Assets = Liabilities + Capital + (Income - Expenses)
/// buf:lint:ignore ENUM_VALUE_PREFIX
class LedgerType extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const LedgerType ASSET =
      LedgerType._(0, _omitEnumNames ? '' : 'ASSET');
  static const LedgerType LIABILITY =
      LedgerType._(1, _omitEnumNames ? '' : 'LIABILITY');
  static const LedgerType INCOME =
      LedgerType._(2, _omitEnumNames ? '' : 'INCOME');
  static const LedgerType EXPENSE =
      LedgerType._(3, _omitEnumNames ? '' : 'EXPENSE');
  static const LedgerType CAPITAL =
      LedgerType._(4, _omitEnumNames ? '' : 'CAPITAL');

  static const $core.List<LedgerType> values = <LedgerType>[
    ASSET,
    LIABILITY,
    INCOME,
    EXPENSE,
    CAPITAL,
  ];

  static final $core.List<LedgerType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static LedgerType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LedgerType._(super.value, super.name);
}

/// TransactionType defines the nature of a transaction.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class TransactionType extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const TransactionType NORMAL =
      TransactionType._(0, _omitEnumNames ? '' : 'NORMAL');
  static const TransactionType REVERSAL =
      TransactionType._(1, _omitEnumNames ? '' : 'REVERSAL');
  static const TransactionType RESERVATION =
      TransactionType._(2, _omitEnumNames ? '' : 'RESERVATION');

  static const $core.List<TransactionType> values = <TransactionType>[
    NORMAL,
    REVERSAL,
    RESERVATION,
  ];

  static final $core.List<TransactionType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static TransactionType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TransactionType._(super.value, super.name);
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
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const TransactionStatus PENDING =
      TransactionStatus._(0, _omitEnumNames ? '' : 'PENDING');
  static const TransactionStatus POSTED =
      TransactionStatus._(1, _omitEnumNames ? '' : 'POSTED');
  static const TransactionStatus REVERSED =
      TransactionStatus._(2, _omitEnumNames ? '' : 'REVERSED');
  static const TransactionStatus VOIDED =
      TransactionStatus._(3, _omitEnumNames ? '' : 'VOIDED');
  static const TransactionStatus FAILED =
      TransactionStatus._(4, _omitEnumNames ? '' : 'FAILED');
  static const TransactionStatus DRAFT =
      TransactionStatus._(5, _omitEnumNames ? '' : 'DRAFT');

  static const $core.List<TransactionStatus> values = <TransactionStatus>[
    PENDING,
    POSTED,
    REVERSED,
    VOIDED,
    FAILED,
    DRAFT,
  ];

  static final $core.List<TransactionStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static TransactionStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TransactionStatus._(super.value, super.name);
}

/// AccountType is the per-account classification, finer-grained than
/// LedgerType so contra, clearing, suspense and memo accounts can live
/// under their natural parent ledgers without overloading the type.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class AccountType extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const AccountType ACCOUNT_ASSET =
      AccountType._(0, _omitEnumNames ? '' : 'ACCOUNT_ASSET');
  static const AccountType ACCOUNT_LIABILITY =
      AccountType._(1, _omitEnumNames ? '' : 'ACCOUNT_LIABILITY');
  static const AccountType ACCOUNT_EQUITY =
      AccountType._(2, _omitEnumNames ? '' : 'ACCOUNT_EQUITY');
  static const AccountType ACCOUNT_INCOME =
      AccountType._(3, _omitEnumNames ? '' : 'ACCOUNT_INCOME');
  static const AccountType ACCOUNT_EXPENSE =
      AccountType._(4, _omitEnumNames ? '' : 'ACCOUNT_EXPENSE');
  static const AccountType ACCOUNT_CONTRA_ASSET =
      AccountType._(5, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_ASSET');
  static const AccountType ACCOUNT_CONTRA_LIABILITY =
      AccountType._(6, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_LIABILITY');
  static const AccountType ACCOUNT_CONTRA_INCOME =
      AccountType._(7, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_INCOME');
  static const AccountType ACCOUNT_CONTRA_EXPENSE =
      AccountType._(8, _omitEnumNames ? '' : 'ACCOUNT_CONTRA_EXPENSE');
  static const AccountType ACCOUNT_CLEARING =
      AccountType._(9, _omitEnumNames ? '' : 'ACCOUNT_CLEARING');
  static const AccountType ACCOUNT_SUSPENSE =
      AccountType._(10, _omitEnumNames ? '' : 'ACCOUNT_SUSPENSE');
  static const AccountType ACCOUNT_MEMO =
      AccountType._(11, _omitEnumNames ? '' : 'ACCOUNT_MEMO');

  static const $core.List<AccountType> values = <AccountType>[
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

  static final $core.List<AccountType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 11);
  static AccountType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AccountType._(super.value, super.name);
}

/// NormalBalance is the side an account's balance accumulates on. DEADCLIC
/// stores +amount when an entry's side equals the normal balance, -amount
/// when it is the opposite. NONE opts a memo account out of normalisation.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class NormalBalance extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const NormalBalance DEBIT =
      NormalBalance._(0, _omitEnumNames ? '' : 'DEBIT');
  static const NormalBalance CREDIT =
      NormalBalance._(1, _omitEnumNames ? '' : 'CREDIT');
  static const NormalBalance NONE =
      NormalBalance._(2, _omitEnumNames ? '' : 'NONE');

  static const $core.List<NormalBalance> values = <NormalBalance>[
    DEBIT,
    CREDIT,
    NONE,
  ];

  static final $core.List<NormalBalance?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static NormalBalance? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const NormalBalance._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
