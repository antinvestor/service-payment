// This is a generated file - do not edit.
//
// Generated from v1/billing.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// PricingModel defines how a component is priced.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class PricingModel extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const PricingModel FLAT =
      PricingModel._(0, _omitEnumNames ? '' : 'FLAT');
  static const PricingModel PER_UNIT =
      PricingModel._(1, _omitEnumNames ? '' : 'PER_UNIT');
  static const PricingModel TIERED =
      PricingModel._(2, _omitEnumNames ? '' : 'TIERED');
  static const PricingModel VOLUME =
      PricingModel._(3, _omitEnumNames ? '' : 'VOLUME');
  static const PricingModel STAIRSTEP =
      PricingModel._(4, _omitEnumNames ? '' : 'STAIRSTEP');

  static const $core.List<PricingModel> values = <PricingModel>[
    FLAT,
    PER_UNIT,
    TIERED,
    VOLUME,
    STAIRSTEP,
  ];

  static final $core.List<PricingModel?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static PricingModel? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const PricingModel._(super.value, super.name);
}

/// AggregationType defines how usage events are aggregated within a window.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class AggregationType extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const AggregationType SUM =
      AggregationType._(0, _omitEnumNames ? '' : 'SUM');
  static const AggregationType COUNT =
      AggregationType._(1, _omitEnumNames ? '' : 'COUNT');
  static const AggregationType MAX =
      AggregationType._(2, _omitEnumNames ? '' : 'MAX');
  static const AggregationType MIN =
      AggregationType._(3, _omitEnumNames ? '' : 'MIN');
  static const AggregationType AVG =
      AggregationType._(4, _omitEnumNames ? '' : 'AVG');
  static const AggregationType LAST =
      AggregationType._(5, _omitEnumNames ? '' : 'LAST');

  static const $core.List<AggregationType> values = <AggregationType>[
    SUM,
    COUNT,
    MAX,
    MIN,
    AVG,
    LAST,
  ];

  static final $core.List<AggregationType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static AggregationType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AggregationType._(super.value, super.name);
}

/// SubscriptionState defines the lifecycle state of a subscription.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class SubscriptionState extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const SubscriptionState SUBSCRIPTION_ACTIVE =
      SubscriptionState._(0, _omitEnumNames ? '' : 'SUBSCRIPTION_ACTIVE');
  static const SubscriptionState SUBSCRIPTION_CANCELLED =
      SubscriptionState._(1, _omitEnumNames ? '' : 'SUBSCRIPTION_CANCELLED');
  static const SubscriptionState SUBSCRIPTION_EXPIRED =
      SubscriptionState._(2, _omitEnumNames ? '' : 'SUBSCRIPTION_EXPIRED');
  static const SubscriptionState SUBSCRIPTION_PENDING =
      SubscriptionState._(3, _omitEnumNames ? '' : 'SUBSCRIPTION_PENDING');

  static const $core.List<SubscriptionState> values = <SubscriptionState>[
    SUBSCRIPTION_ACTIVE,
    SUBSCRIPTION_CANCELLED,
    SUBSCRIPTION_EXPIRED,
    SUBSCRIPTION_PENDING,
  ];

  static final $core.List<SubscriptionState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static SubscriptionState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SubscriptionState._(super.value, super.name);
}

/// InvoiceState defines the lifecycle state of an invoice.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class InvoiceState extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const InvoiceState INVOICE_DRAFT =
      InvoiceState._(0, _omitEnumNames ? '' : 'INVOICE_DRAFT');
  static const InvoiceState INVOICE_ISSUED =
      InvoiceState._(1, _omitEnumNames ? '' : 'INVOICE_ISSUED');
  static const InvoiceState INVOICE_PAID =
      InvoiceState._(2, _omitEnumNames ? '' : 'INVOICE_PAID');
  static const InvoiceState INVOICE_VOIDED =
      InvoiceState._(3, _omitEnumNames ? '' : 'INVOICE_VOIDED');
  static const InvoiceState INVOICE_OVERDUE =
      InvoiceState._(4, _omitEnumNames ? '' : 'INVOICE_OVERDUE');

  static const $core.List<InvoiceState> values = <InvoiceState>[
    INVOICE_DRAFT,
    INVOICE_ISSUED,
    INVOICE_PAID,
    INVOICE_VOIDED,
    INVOICE_OVERDUE,
  ];

  static final $core.List<InvoiceState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static InvoiceState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const InvoiceState._(super.value, super.name);
}

/// BillingRunState defines the state machine for billing run execution.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class BillingRunState extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const BillingRunState BILLING_RUN_PENDING =
      BillingRunState._(0, _omitEnumNames ? '' : 'BILLING_RUN_PENDING');
  static const BillingRunState BILLING_RUN_METERING =
      BillingRunState._(1, _omitEnumNames ? '' : 'BILLING_RUN_METERING');
  static const BillingRunState BILLING_RUN_RATING =
      BillingRunState._(2, _omitEnumNames ? '' : 'BILLING_RUN_RATING');
  static const BillingRunState BILLING_RUN_DISCOUNTING =
      BillingRunState._(3, _omitEnumNames ? '' : 'BILLING_RUN_DISCOUNTING');
  static const BillingRunState BILLING_RUN_CREDITING =
      BillingRunState._(4, _omitEnumNames ? '' : 'BILLING_RUN_CREDITING');
  static const BillingRunState BILLING_RUN_INVOICING =
      BillingRunState._(5, _omitEnumNames ? '' : 'BILLING_RUN_INVOICING');
  static const BillingRunState BILLING_RUN_POSTING =
      BillingRunState._(6, _omitEnumNames ? '' : 'BILLING_RUN_POSTING');
  static const BillingRunState BILLING_RUN_COMPLETED =
      BillingRunState._(7, _omitEnumNames ? '' : 'BILLING_RUN_COMPLETED');
  static const BillingRunState BILLING_RUN_FAILED =
      BillingRunState._(8, _omitEnumNames ? '' : 'BILLING_RUN_FAILED');

  static const $core.List<BillingRunState> values = <BillingRunState>[
    BILLING_RUN_PENDING,
    BILLING_RUN_METERING,
    BILLING_RUN_RATING,
    BILLING_RUN_DISCOUNTING,
    BILLING_RUN_CREDITING,
    BILLING_RUN_INVOICING,
    BILLING_RUN_POSTING,
    BILLING_RUN_COMPLETED,
    BILLING_RUN_FAILED,
  ];

  static final $core.List<BillingRunState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 8);
  static BillingRunState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const BillingRunState._(super.value, super.name);
}

/// DiscountType defines the type of discount applied.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class DiscountType extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const DiscountType DISCOUNT_PERCENTAGE =
      DiscountType._(0, _omitEnumNames ? '' : 'DISCOUNT_PERCENTAGE');
  static const DiscountType DISCOUNT_FIXED =
      DiscountType._(1, _omitEnumNames ? '' : 'DISCOUNT_FIXED');

  static const $core.List<DiscountType> values = <DiscountType>[
    DISCOUNT_PERCENTAGE,
    DISCOUNT_FIXED,
  ];

  static final $core.List<DiscountType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static DiscountType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DiscountType._(super.value, super.name);
}

/// CreditEntryType defines the type of credit ledger entry.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class CreditEntryType extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const CreditEntryType CREDIT_GRANT =
      CreditEntryType._(0, _omitEnumNames ? '' : 'CREDIT_GRANT');
  static const CreditEntryType CREDIT_CONSUME =
      CreditEntryType._(1, _omitEnumNames ? '' : 'CREDIT_CONSUME');
  static const CreditEntryType CREDIT_EXPIRE =
      CreditEntryType._(2, _omitEnumNames ? '' : 'CREDIT_EXPIRE');
  static const CreditEntryType CREDIT_REFUND =
      CreditEntryType._(3, _omitEnumNames ? '' : 'CREDIT_REFUND');

  static const $core.List<CreditEntryType> values = <CreditEntryType>[
    CREDIT_GRANT,
    CREDIT_CONSUME,
    CREDIT_EXPIRE,
    CREDIT_REFUND,
  ];

  static final $core.List<CreditEntryType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static CreditEntryType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CreditEntryType._(super.value, super.name);
}

/// InvoiceLineType defines the type of invoice line item.
/// buf:lint:ignore ENUM_VALUE_PREFIX
class InvoiceLineType extends $pb.ProtobufEnum {
  /// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
  static const InvoiceLineType LINE_USAGE =
      InvoiceLineType._(0, _omitEnumNames ? '' : 'LINE_USAGE');
  static const InvoiceLineType LINE_FLAT =
      InvoiceLineType._(1, _omitEnumNames ? '' : 'LINE_FLAT');
  static const InvoiceLineType LINE_DISCOUNT =
      InvoiceLineType._(2, _omitEnumNames ? '' : 'LINE_DISCOUNT');
  static const InvoiceLineType LINE_CREDIT =
      InvoiceLineType._(3, _omitEnumNames ? '' : 'LINE_CREDIT');

  static const $core.List<InvoiceLineType> values = <InvoiceLineType>[
    LINE_USAGE,
    LINE_FLAT,
    LINE_DISCOUNT,
    LINE_CREDIT,
  ];

  static final $core.List<InvoiceLineType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static InvoiceLineType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const InvoiceLineType._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
