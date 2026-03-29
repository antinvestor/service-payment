//
//  Generated code. Do not modify.
//  source: v1/billing.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../common/v1/common.pbjson.dart' as $9;
import '../google/protobuf/struct.pbjson.dart' as $6;
import '../google/protobuf/timestamp.pbjson.dart' as $2;
import '../google/type/interval.pbjson.dart' as $8;
import '../google/type/money.pbjson.dart' as $7;

@$core.Deprecated('Use pricingModelDescriptor instead')
const PricingModel$json = {
  '1': 'PricingModel',
  '2': [
    {'1': 'FLAT', '2': 0},
    {'1': 'PER_UNIT', '2': 1},
    {'1': 'TIERED', '2': 2},
    {'1': 'VOLUME', '2': 3},
    {'1': 'STAIRSTEP', '2': 4},
  ],
};

/// Descriptor for `PricingModel`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List pricingModelDescriptor = $convert.base64Decode(
    'CgxQcmljaW5nTW9kZWwSCAoERkxBVBAAEgwKCFBFUl9VTklUEAESCgoGVElFUkVEEAISCgoGVk'
    '9MVU1FEAMSDQoJU1RBSVJTVEVQEAQ=');

@$core.Deprecated('Use aggregationTypeDescriptor instead')
const AggregationType$json = {
  '1': 'AggregationType',
  '2': [
    {'1': 'SUM', '2': 0},
    {'1': 'COUNT', '2': 1},
    {'1': 'MAX', '2': 2},
    {'1': 'MIN', '2': 3},
    {'1': 'AVG', '2': 4},
    {'1': 'LAST', '2': 5},
  ],
};

/// Descriptor for `AggregationType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List aggregationTypeDescriptor = $convert.base64Decode(
    'Cg9BZ2dyZWdhdGlvblR5cGUSBwoDU1VNEAASCQoFQ09VTlQQARIHCgNNQVgQAhIHCgNNSU4QAx'
    'IHCgNBVkcQBBIICgRMQVNUEAU=');

@$core.Deprecated('Use subscriptionStateDescriptor instead')
const SubscriptionState$json = {
  '1': 'SubscriptionState',
  '2': [
    {'1': 'SUBSCRIPTION_ACTIVE', '2': 0},
    {'1': 'SUBSCRIPTION_CANCELLED', '2': 1},
    {'1': 'SUBSCRIPTION_EXPIRED', '2': 2},
    {'1': 'SUBSCRIPTION_PENDING', '2': 3},
  ],
};

/// Descriptor for `SubscriptionState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List subscriptionStateDescriptor = $convert.base64Decode(
    'ChFTdWJzY3JpcHRpb25TdGF0ZRIXChNTVUJTQ1JJUFRJT05fQUNUSVZFEAASGgoWU1VCU0NSSV'
    'BUSU9OX0NBTkNFTExFRBABEhgKFFNVQlNDUklQVElPTl9FWFBJUkVEEAISGAoUU1VCU0NSSVBU'
    'SU9OX1BFTkRJTkcQAw==');

@$core.Deprecated('Use invoiceStateDescriptor instead')
const InvoiceState$json = {
  '1': 'InvoiceState',
  '2': [
    {'1': 'INVOICE_DRAFT', '2': 0},
    {'1': 'INVOICE_ISSUED', '2': 1},
    {'1': 'INVOICE_PAID', '2': 2},
    {'1': 'INVOICE_VOIDED', '2': 3},
    {'1': 'INVOICE_OVERDUE', '2': 4},
  ],
};

/// Descriptor for `InvoiceState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List invoiceStateDescriptor = $convert.base64Decode(
    'CgxJbnZvaWNlU3RhdGUSEQoNSU5WT0lDRV9EUkFGVBAAEhIKDklOVk9JQ0VfSVNTVUVEEAESEA'
    'oMSU5WT0lDRV9QQUlEEAISEgoOSU5WT0lDRV9WT0lERUQQAxITCg9JTlZPSUNFX09WRVJEVUUQ'
    'BA==');

@$core.Deprecated('Use billingRunStateDescriptor instead')
const BillingRunState$json = {
  '1': 'BillingRunState',
  '2': [
    {'1': 'BILLING_RUN_PENDING', '2': 0},
    {'1': 'BILLING_RUN_METERING', '2': 1},
    {'1': 'BILLING_RUN_RATING', '2': 2},
    {'1': 'BILLING_RUN_DISCOUNTING', '2': 3},
    {'1': 'BILLING_RUN_CREDITING', '2': 4},
    {'1': 'BILLING_RUN_INVOICING', '2': 5},
    {'1': 'BILLING_RUN_POSTING', '2': 6},
    {'1': 'BILLING_RUN_COMPLETED', '2': 7},
    {'1': 'BILLING_RUN_FAILED', '2': 8},
  ],
};

/// Descriptor for `BillingRunState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List billingRunStateDescriptor = $convert.base64Decode(
    'Cg9CaWxsaW5nUnVuU3RhdGUSFwoTQklMTElOR19SVU5fUEVORElORxAAEhgKFEJJTExJTkdfUl'
    'VOX01FVEVSSU5HEAESFgoSQklMTElOR19SVU5fUkFUSU5HEAISGwoXQklMTElOR19SVU5fRElT'
    'Q09VTlRJTkcQAxIZChVCSUxMSU5HX1JVTl9DUkVESVRJTkcQBBIZChVCSUxMSU5HX1JVTl9JTl'
    'ZPSUNJTkcQBRIXChNCSUxMSU5HX1JVTl9QT1NUSU5HEAYSGQoVQklMTElOR19SVU5fQ09NUExF'
    'VEVEEAcSFgoSQklMTElOR19SVU5fRkFJTEVEEAg=');

@$core.Deprecated('Use discountTypeDescriptor instead')
const DiscountType$json = {
  '1': 'DiscountType',
  '2': [
    {'1': 'DISCOUNT_PERCENTAGE', '2': 0},
    {'1': 'DISCOUNT_FIXED', '2': 1},
  ],
};

/// Descriptor for `DiscountType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List discountTypeDescriptor = $convert.base64Decode(
    'CgxEaXNjb3VudFR5cGUSFwoTRElTQ09VTlRfUEVSQ0VOVEFHRRAAEhIKDkRJU0NPVU5UX0ZJWE'
    'VEEAE=');

@$core.Deprecated('Use creditEntryTypeDescriptor instead')
const CreditEntryType$json = {
  '1': 'CreditEntryType',
  '2': [
    {'1': 'CREDIT_GRANT', '2': 0},
    {'1': 'CREDIT_CONSUME', '2': 1},
    {'1': 'CREDIT_EXPIRE', '2': 2},
    {'1': 'CREDIT_REFUND', '2': 3},
  ],
};

/// Descriptor for `CreditEntryType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List creditEntryTypeDescriptor = $convert.base64Decode(
    'Cg9DcmVkaXRFbnRyeVR5cGUSEAoMQ1JFRElUX0dSQU5UEAASEgoOQ1JFRElUX0NPTlNVTUUQAR'
    'IRCg1DUkVESVRfRVhQSVJFEAISEQoNQ1JFRElUX1JFRlVORBAD');

@$core.Deprecated('Use invoiceLineTypeDescriptor instead')
const InvoiceLineType$json = {
  '1': 'InvoiceLineType',
  '2': [
    {'1': 'LINE_USAGE', '2': 0},
    {'1': 'LINE_FLAT', '2': 1},
    {'1': 'LINE_DISCOUNT', '2': 2},
    {'1': 'LINE_CREDIT', '2': 3},
  ],
};

/// Descriptor for `InvoiceLineType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List invoiceLineTypeDescriptor = $convert.base64Decode(
    'Cg9JbnZvaWNlTGluZVR5cGUSDgoKTElORV9VU0FHRRAAEg0KCUxJTkVfRkxBVBABEhEKDUxJTk'
    'VfRElTQ09VTlQQAhIPCgtMSU5FX0NSRURJVBAD');

@$core.Deprecated('Use catalogVersionDescriptor instead')
const CatalogVersion$json = {
  '1': 'CatalogVersion',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'catalog_id', '3': 2, '4': 1, '5': 9, '10': 'catalogId'},
    {'1': 'version', '3': 3, '4': 1, '5': 5, '10': 'version'},
    {'1': 'name', '3': 4, '4': 1, '5': 9, '10': 'name'},
    {'1': 'currency', '3': 5, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'published_at', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'publishedAt'},
    {'1': 'effective_at', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'effectiveAt'},
    {'1': 'retired_at', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'retiredAt'},
    {'1': 'data', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'plans', '3': 10, '4': 3, '5': 11, '6': '.billing.v1.Plan', '10': 'plans'},
  ],
};

/// Descriptor for `CatalogVersion`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List catalogVersionDescriptor = $convert.base64Decode(
    'Cg5DYXRhbG9nVmVyc2lvbhIOCgJpZBgBIAEoCVICaWQSHQoKY2F0YWxvZ19pZBgCIAEoCVIJY2'
    'F0YWxvZ0lkEhgKB3ZlcnNpb24YAyABKAVSB3ZlcnNpb24SEgoEbmFtZRgEIAEoCVIEbmFtZRIa'
    'CghjdXJyZW5jeRgFIAEoCVIIY3VycmVuY3kSPQoMcHVibGlzaGVkX2F0GAYgASgLMhouZ29vZ2'
    'xlLnByb3RvYnVmLlRpbWVzdGFtcFILcHVibGlzaGVkQXQSPQoMZWZmZWN0aXZlX2F0GAcgASgL'
    'MhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFILZWZmZWN0aXZlQXQSOQoKcmV0aXJlZF9hdB'
    'gIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXJldGlyZWRBdBIrCgRkYXRhGAkg'
    'ASgLMhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdFIEZGF0YRImCgVwbGFucxgKIAMoCzIQLmJpbG'
    'xpbmcudjEuUGxhblIFcGxhbnM=');

@$core.Deprecated('Use planDescriptor instead')
const Plan$json = {
  '1': 'Plan',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'catalog_version_id', '3': 2, '4': 1, '5': 9, '10': 'catalogVersionId'},
    {'1': 'external_id', '3': 3, '4': 1, '5': 9, '10': 'externalId'},
    {'1': 'name', '3': 4, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 5, '4': 1, '5': 9, '10': 'description'},
    {'1': 'data', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'components', '3': 7, '4': 3, '5': 11, '6': '.billing.v1.Component', '10': 'components'},
  ],
};

/// Descriptor for `Plan`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List planDescriptor = $convert.base64Decode(
    'CgRQbGFuEg4KAmlkGAEgASgJUgJpZBIsChJjYXRhbG9nX3ZlcnNpb25faWQYAiABKAlSEGNhdG'
    'Fsb2dWZXJzaW9uSWQSHwoLZXh0ZXJuYWxfaWQYAyABKAlSCmV4dGVybmFsSWQSEgoEbmFtZRgE'
    'IAEoCVIEbmFtZRIgCgtkZXNjcmlwdGlvbhgFIAEoCVILZGVzY3JpcHRpb24SKwoEZGF0YRgGIA'
    'EoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdGESNQoKY29tcG9uZW50cxgHIAMoCzIV'
    'LmJpbGxpbmcudjEuQ29tcG9uZW50Ugpjb21wb25lbnRz');

@$core.Deprecated('Use componentDescriptor instead')
const Component$json = {
  '1': 'Component',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'plan_id', '3': 2, '4': 1, '5': 9, '10': 'planId'},
    {'1': 'external_id', '3': 3, '4': 1, '5': 9, '10': 'externalId'},
    {'1': 'name', '3': 4, '4': 1, '5': 9, '10': 'name'},
    {'1': 'metric_key', '3': 5, '4': 1, '5': 9, '10': 'metricKey'},
    {'1': 'pricing_model', '3': 6, '4': 1, '5': 14, '6': '.billing.v1.PricingModel', '10': 'pricingModel'},
    {'1': 'aggregation_type', '3': 7, '4': 1, '5': 14, '6': '.billing.v1.AggregationType', '10': 'aggregationType'},
    {'1': 'unit_name', '3': 8, '4': 1, '5': 9, '10': 'unitName'},
    {'1': 'free_quantity', '3': 9, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'freeQuantity'},
    {'1': 'minimum_charge', '3': 10, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'minimumCharge'},
    {'1': 'data', '3': 11, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'tiers', '3': 12, '4': 3, '5': 11, '6': '.billing.v1.Tier', '10': 'tiers'},
  ],
};

/// Descriptor for `Component`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List componentDescriptor = $convert.base64Decode(
    'CglDb21wb25lbnQSDgoCaWQYASABKAlSAmlkEhcKB3BsYW5faWQYAiABKAlSBnBsYW5JZBIfCg'
    'tleHRlcm5hbF9pZBgDIAEoCVIKZXh0ZXJuYWxJZBISCgRuYW1lGAQgASgJUgRuYW1lEh0KCm1l'
    'dHJpY19rZXkYBSABKAlSCW1ldHJpY0tleRI9Cg1wcmljaW5nX21vZGVsGAYgASgOMhguYmlsbG'
    'luZy52MS5QcmljaW5nTW9kZWxSDHByaWNpbmdNb2RlbBJGChBhZ2dyZWdhdGlvbl90eXBlGAcg'
    'ASgOMhsuYmlsbGluZy52MS5BZ2dyZWdhdGlvblR5cGVSD2FnZ3JlZ2F0aW9uVHlwZRIbCgl1bm'
    'l0X25hbWUYCCABKAlSCHVuaXROYW1lEjcKDWZyZWVfcXVhbnRpdHkYCSABKAsyEi5nb29nbGUu'
    'dHlwZS5Nb25leVIMZnJlZVF1YW50aXR5EjkKDm1pbmltdW1fY2hhcmdlGAogASgLMhIuZ29vZ2'
    'xlLnR5cGUuTW9uZXlSDW1pbmltdW1DaGFyZ2USKwoEZGF0YRgLIAEoCzIXLmdvb2dsZS5wcm90'
    'b2J1Zi5TdHJ1Y3RSBGRhdGESJgoFdGllcnMYDCADKAsyEC5iaWxsaW5nLnYxLlRpZXJSBXRpZX'
    'Jz');

@$core.Deprecated('Use tierDescriptor instead')
const Tier$json = {
  '1': 'Tier',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'component_id', '3': 2, '4': 1, '5': 9, '10': 'componentId'},
    {'1': 'sort_order', '3': 3, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'lower_bound', '3': 4, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'lowerBound'},
    {'1': 'upper_bound', '3': 5, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'upperBound'},
    {'1': 'unit_price', '3': 6, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'unitPrice'},
    {'1': 'flat_fee', '3': 7, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'flatFee'},
  ],
};

/// Descriptor for `Tier`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List tierDescriptor = $convert.base64Decode(
    'CgRUaWVyEg4KAmlkGAEgASgJUgJpZBIhCgxjb21wb25lbnRfaWQYAiABKAlSC2NvbXBvbmVudE'
    'lkEh0KCnNvcnRfb3JkZXIYAyABKAVSCXNvcnRPcmRlchIzCgtsb3dlcl9ib3VuZBgEIAEoCzIS'
    'Lmdvb2dsZS50eXBlLk1vbmV5Ugpsb3dlckJvdW5kEjMKC3VwcGVyX2JvdW5kGAUgASgLMhIuZ2'
    '9vZ2xlLnR5cGUuTW9uZXlSCnVwcGVyQm91bmQSMQoKdW5pdF9wcmljZRgGIAEoCzISLmdvb2ds'
    'ZS50eXBlLk1vbmV5Ugl1bml0UHJpY2USLQoIZmxhdF9mZWUYByABKAsyEi5nb29nbGUudHlwZS'
    '5Nb25leVIHZmxhdEZlZQ==');

@$core.Deprecated('Use subscriptionDescriptor instead')
const Subscription$json = {
  '1': 'Subscription',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'profile_id', '3': 2, '4': 1, '5': 9, '10': 'profileId'},
    {'1': 'catalog_version_id', '3': 3, '4': 1, '5': 9, '10': 'catalogVersionId'},
    {'1': 'plan_id', '3': 4, '4': 1, '5': 9, '10': 'planId'},
    {'1': 'state', '3': 5, '4': 1, '5': 14, '6': '.billing.v1.SubscriptionState', '10': 'state'},
    {'1': 'start_at', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'startAt'},
    {'1': 'end_at', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'endAt'},
    {'1': 'cancelled_at', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'cancelledAt'},
    {'1': 'billing_anchor', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'billingAnchor'},
    {'1': 'currency', '3': 10, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'data', '3': 11, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `Subscription`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List subscriptionDescriptor = $convert.base64Decode(
    'CgxTdWJzY3JpcHRpb24SDgoCaWQYASABKAlSAmlkEh0KCnByb2ZpbGVfaWQYAiABKAlSCXByb2'
    'ZpbGVJZBIsChJjYXRhbG9nX3ZlcnNpb25faWQYAyABKAlSEGNhdGFsb2dWZXJzaW9uSWQSFwoH'
    'cGxhbl9pZBgEIAEoCVIGcGxhbklkEjMKBXN0YXRlGAUgASgOMh0uYmlsbGluZy52MS5TdWJzY3'
    'JpcHRpb25TdGF0ZVIFc3RhdGUSNQoIc3RhcnRfYXQYBiABKAsyGi5nb29nbGUucHJvdG9idWYu'
    'VGltZXN0YW1wUgdzdGFydEF0EjEKBmVuZF9hdBgHIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW'
    '1lc3RhbXBSBWVuZEF0Ej0KDGNhbmNlbGxlZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5U'
    'aW1lc3RhbXBSC2NhbmNlbGxlZEF0EkEKDmJpbGxpbmdfYW5jaG9yGAkgASgLMhouZ29vZ2xlLn'
    'Byb3RvYnVmLlRpbWVzdGFtcFINYmlsbGluZ0FuY2hvchIaCghjdXJyZW5jeRgKIAEoCVIIY3Vy'
    'cmVuY3kSKwoEZGF0YRgLIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdGE=');

@$core.Deprecated('Use usageEventDescriptor instead')
const UsageEvent$json = {
  '1': 'UsageEvent',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'subscription_id', '3': 3, '4': 1, '5': 9, '10': 'subscriptionId'},
    {'1': 'metric_key', '3': 5, '4': 1, '5': 9, '10': 'metricKey'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    {'1': 'quantity', '3': 6, '4': 1, '5': 1, '10': 'quantity'},
    {'1': 'unit', '3': 12, '4': 1, '5': 9, '10': 'unit'},
    {'1': 'timestamp', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
    {'1': 'interval', '3': 11, '4': 1, '5': 11, '6': '.google.type.Interval', '10': 'interval'},
    {'1': 'amount', '3': 13, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'amount'},
    {'1': 'properties', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'properties'},
    {'1': 'group_id', '3': 14, '4': 1, '5': 9, '10': 'groupId'},
  ],
};

/// Descriptor for `UsageEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List usageEventDescriptor = $convert.base64Decode(
    'CgpVc2FnZUV2ZW50Eg4KAmlkGAEgASgJUgJpZBInCg9zdWJzY3JpcHRpb25faWQYAyABKAlSDn'
    'N1YnNjcmlwdGlvbklkEh0KCm1ldHJpY19rZXkYBSABKAlSCW1ldHJpY0tleRIgCgtkZXNjcmlw'
    'dGlvbhgCIAEoCVILZGVzY3JpcHRpb24SGgoIcXVhbnRpdHkYBiABKAFSCHF1YW50aXR5EhIKBH'
    'VuaXQYDCABKAlSBHVuaXQSOAoJdGltZXN0YW1wGAcgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRp'
    'bWVzdGFtcFIJdGltZXN0YW1wEjEKCGludGVydmFsGAsgASgLMhUuZ29vZ2xlLnR5cGUuSW50ZX'
    'J2YWxSCGludGVydmFsEioKBmFtb3VudBgNIAEoCzISLmdvb2dsZS50eXBlLk1vbmV5UgZhbW91'
    'bnQSNwoKcHJvcGVydGllcxgIIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSCnByb3Blcn'
    'RpZXMSGQoIZ3JvdXBfaWQYDiABKAlSB2dyb3VwSWQ=');

@$core.Deprecated('Use invoiceDescriptor instead')
const Invoice$json = {
  '1': 'Invoice',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'billing_run_id', '3': 2, '4': 1, '5': 9, '10': 'billingRunId'},
    {'1': 'profile_id', '3': 3, '4': 1, '5': 9, '10': 'profileId'},
    {'1': 'subscription_id', '3': 4, '4': 1, '5': 9, '10': 'subscriptionId'},
    {'1': 'invoice_number', '3': 5, '4': 1, '5': 9, '10': 'invoiceNumber'},
    {'1': 'state', '3': 6, '4': 1, '5': 14, '6': '.billing.v1.InvoiceState', '10': 'state'},
    {'1': 'currency', '3': 7, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'subtotal_amount', '3': 8, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'subtotalAmount'},
    {'1': 'discount_amount', '3': 9, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'discountAmount'},
    {'1': 'credit_amount', '3': 10, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'creditAmount'},
    {'1': 'total_amount', '3': 11, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'totalAmount'},
    {'1': 'period_start', '3': 12, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'periodStart'},
    {'1': 'period_end', '3': 13, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'periodEnd'},
    {'1': 'issued_at', '3': 14, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'issuedAt'},
    {'1': 'due_at', '3': 15, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'dueAt'},
    {'1': 'paid_at', '3': 16, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'paidAt'},
    {'1': 'ledger_txn_id', '3': 17, '4': 1, '5': 9, '10': 'ledgerTxnId'},
    {'1': 'data', '3': 18, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'lines', '3': 19, '4': 3, '5': 11, '6': '.billing.v1.InvoiceLine', '10': 'lines'},
  ],
};

/// Descriptor for `Invoice`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List invoiceDescriptor = $convert.base64Decode(
    'CgdJbnZvaWNlEg4KAmlkGAEgASgJUgJpZBIkCg5iaWxsaW5nX3J1bl9pZBgCIAEoCVIMYmlsbG'
    'luZ1J1bklkEh0KCnByb2ZpbGVfaWQYAyABKAlSCXByb2ZpbGVJZBInCg9zdWJzY3JpcHRpb25f'
    'aWQYBCABKAlSDnN1YnNjcmlwdGlvbklkEiUKDmludm9pY2VfbnVtYmVyGAUgASgJUg1pbnZvaW'
    'NlTnVtYmVyEi4KBXN0YXRlGAYgASgOMhguYmlsbGluZy52MS5JbnZvaWNlU3RhdGVSBXN0YXRl'
    'EhoKCGN1cnJlbmN5GAcgASgJUghjdXJyZW5jeRI7Cg9zdWJ0b3RhbF9hbW91bnQYCCABKAsyEi'
    '5nb29nbGUudHlwZS5Nb25leVIOc3VidG90YWxBbW91bnQSOwoPZGlzY291bnRfYW1vdW50GAkg'
    'ASgLMhIuZ29vZ2xlLnR5cGUuTW9uZXlSDmRpc2NvdW50QW1vdW50EjcKDWNyZWRpdF9hbW91bn'
    'QYCiABKAsyEi5nb29nbGUudHlwZS5Nb25leVIMY3JlZGl0QW1vdW50EjUKDHRvdGFsX2Ftb3Vu'
    'dBgLIAEoCzISLmdvb2dsZS50eXBlLk1vbmV5Ugt0b3RhbEFtb3VudBI9CgxwZXJpb2Rfc3Rhcn'
    'QYDCABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgtwZXJpb2RTdGFydBI5CgpwZXJp'
    'b2RfZW5kGA0gASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJcGVyaW9kRW5kEjcKCW'
    'lzc3VlZF9hdBgOIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCGlzc3VlZEF0EjEK'
    'BmR1ZV9hdBgPIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSBWR1ZUF0EjMKB3BhaW'
    'RfYXQYECABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgZwYWlkQXQSIgoNbGVkZ2Vy'
    'X3R4bl9pZBgRIAEoCVILbGVkZ2VyVHhuSWQSKwoEZGF0YRgSIAEoCzIXLmdvb2dsZS5wcm90b2'
    'J1Zi5TdHJ1Y3RSBGRhdGESLQoFbGluZXMYEyADKAsyFy5iaWxsaW5nLnYxLkludm9pY2VMaW5l'
    'UgVsaW5lcw==');

@$core.Deprecated('Use invoiceLineDescriptor instead')
const InvoiceLine$json = {
  '1': 'InvoiceLine',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'invoice_id', '3': 2, '4': 1, '5': 9, '10': 'invoiceId'},
    {'1': 'component_id', '3': 3, '4': 1, '5': 9, '10': 'componentId'},
    {'1': 'description', '3': 4, '4': 1, '5': 9, '10': 'description'},
    {'1': 'quantity', '3': 5, '4': 1, '5': 1, '10': 'quantity'},
    {'1': 'unit_price', '3': 6, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'unitPrice'},
    {'1': 'amount', '3': 7, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'amount'},
    {'1': 'discount_amount', '3': 8, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'discountAmount'},
    {'1': 'credit_amount', '3': 9, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'creditAmount'},
    {'1': 'net_amount', '3': 10, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'netAmount'},
    {'1': 'currency', '3': 11, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'line_type', '3': 12, '4': 1, '5': 14, '6': '.billing.v1.InvoiceLineType', '10': 'lineType'},
    {'1': 'data', '3': 13, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `InvoiceLine`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List invoiceLineDescriptor = $convert.base64Decode(
    'CgtJbnZvaWNlTGluZRIOCgJpZBgBIAEoCVICaWQSHQoKaW52b2ljZV9pZBgCIAEoCVIJaW52b2'
    'ljZUlkEiEKDGNvbXBvbmVudF9pZBgDIAEoCVILY29tcG9uZW50SWQSIAoLZGVzY3JpcHRpb24Y'
    'BCABKAlSC2Rlc2NyaXB0aW9uEhoKCHF1YW50aXR5GAUgASgBUghxdWFudGl0eRIxCgp1bml0X3'
    'ByaWNlGAYgASgLMhIuZ29vZ2xlLnR5cGUuTW9uZXlSCXVuaXRQcmljZRIqCgZhbW91bnQYByAB'
    'KAsyEi5nb29nbGUudHlwZS5Nb25leVIGYW1vdW50EjsKD2Rpc2NvdW50X2Ftb3VudBgIIAEoCz'
    'ISLmdvb2dsZS50eXBlLk1vbmV5Ug5kaXNjb3VudEFtb3VudBI3Cg1jcmVkaXRfYW1vdW50GAkg'
    'ASgLMhIuZ29vZ2xlLnR5cGUuTW9uZXlSDGNyZWRpdEFtb3VudBIxCgpuZXRfYW1vdW50GAogAS'
    'gLMhIuZ29vZ2xlLnR5cGUuTW9uZXlSCW5ldEFtb3VudBIaCghjdXJyZW5jeRgLIAEoCVIIY3Vy'
    'cmVuY3kSOAoJbGluZV90eXBlGAwgASgOMhsuYmlsbGluZy52MS5JbnZvaWNlTGluZVR5cGVSCG'
    'xpbmVUeXBlEisKBGRhdGEYDSABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgRkYXRh');

@$core.Deprecated('Use creditGrantDescriptor instead')
const CreditGrant$json = {
  '1': 'CreditGrant',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'profile_id', '3': 2, '4': 1, '5': 9, '10': 'profileId'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'original_amount', '3': 4, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'originalAmount'},
    {'1': 'remaining_amount', '3': 5, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'remainingAmount'},
    {'1': 'currency', '3': 6, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'expires_at', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'expiresAt'},
    {'1': 'priority', '3': 8, '4': 1, '5': 5, '10': 'priority'},
    {'1': 'data', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreditGrant`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List creditGrantDescriptor = $convert.base64Decode(
    'CgtDcmVkaXRHcmFudBIOCgJpZBgBIAEoCVICaWQSHQoKcHJvZmlsZV9pZBgCIAEoCVIJcHJvZm'
    'lsZUlkEhIKBG5hbWUYAyABKAlSBG5hbWUSOwoPb3JpZ2luYWxfYW1vdW50GAQgASgLMhIuZ29v'
    'Z2xlLnR5cGUuTW9uZXlSDm9yaWdpbmFsQW1vdW50Ej0KEHJlbWFpbmluZ19hbW91bnQYBSABKA'
    'syEi5nb29nbGUudHlwZS5Nb25leVIPcmVtYWluaW5nQW1vdW50EhoKCGN1cnJlbmN5GAYgASgJ'
    'UghjdXJyZW5jeRI5CgpleHBpcmVzX2F0GAcgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdG'
    'FtcFIJZXhwaXJlc0F0EhoKCHByaW9yaXR5GAggASgFUghwcmlvcml0eRIrCgRkYXRhGAkgASgL'
    'MhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdFIEZGF0YQ==');

@$core.Deprecated('Use billingRunDescriptor instead')
const BillingRun$json = {
  '1': 'BillingRun',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'subscription_id', '3': 2, '4': 1, '5': 9, '10': 'subscriptionId'},
    {'1': 'profile_id', '3': 3, '4': 1, '5': 9, '10': 'profileId'},
    {'1': 'catalog_version_id', '3': 4, '4': 1, '5': 9, '10': 'catalogVersionId'},
    {'1': 'state', '3': 5, '4': 1, '5': 14, '6': '.billing.v1.BillingRunState', '10': 'state'},
    {'1': 'period_start', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'periodStart'},
    {'1': 'period_end', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'periodEnd'},
    {'1': 'started_at', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'startedAt'},
    {'1': 'completed_at', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'completedAt'},
    {'1': 'failed_at', '3': 10, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'failedAt'},
    {'1': 'error_message', '3': 11, '4': 1, '5': 9, '10': 'errorMessage'},
    {'1': 'invoice_id', '3': 12, '4': 1, '5': 9, '10': 'invoiceId'},
    {'1': 'data', '3': 13, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `BillingRun`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List billingRunDescriptor = $convert.base64Decode(
    'CgpCaWxsaW5nUnVuEg4KAmlkGAEgASgJUgJpZBInCg9zdWJzY3JpcHRpb25faWQYAiABKAlSDn'
    'N1YnNjcmlwdGlvbklkEh0KCnByb2ZpbGVfaWQYAyABKAlSCXByb2ZpbGVJZBIsChJjYXRhbG9n'
    'X3ZlcnNpb25faWQYBCABKAlSEGNhdGFsb2dWZXJzaW9uSWQSMQoFc3RhdGUYBSABKA4yGy5iaW'
    'xsaW5nLnYxLkJpbGxpbmdSdW5TdGF0ZVIFc3RhdGUSPQoMcGVyaW9kX3N0YXJ0GAYgASgLMhou'
    'Z29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFILcGVyaW9kU3RhcnQSOQoKcGVyaW9kX2VuZBgHIA'
    'EoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXBlcmlvZEVuZBI5CgpzdGFydGVkX2F0'
    'GAggASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJc3RhcnRlZEF0Ej0KDGNvbXBsZX'
    'RlZF9hdBgJIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSC2NvbXBsZXRlZEF0EjcK'
    'CWZhaWxlZF9hdBgKIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCGZhaWxlZEF0Ei'
    'MKDWVycm9yX21lc3NhZ2UYCyABKAlSDGVycm9yTWVzc2FnZRIdCgppbnZvaWNlX2lkGAwgASgJ'
    'UglpbnZvaWNlSWQSKwoEZGF0YRgNIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdG'
    'E=');

@$core.Deprecated('Use discountDescriptor instead')
const Discount$json = {
  '1': 'Discount',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'discount_type', '3': 3, '4': 1, '5': 14, '6': '.billing.v1.DiscountType', '10': 'discountType'},
    {'1': 'value', '3': 4, '4': 1, '5': 1, '10': 'value'},
    {'1': 'currency', '3': 5, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'applicable_to', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'applicableTo'},
    {'1': 'start_at', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'startAt'},
    {'1': 'end_at', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'endAt'},
    {'1': 'max_applications', '3': 9, '4': 1, '5': 5, '10': 'maxApplications'},
    {'1': 'data', '3': 10, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `Discount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List discountDescriptor = $convert.base64Decode(
    'CghEaXNjb3VudBIOCgJpZBgBIAEoCVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZRI9Cg1kaXNjb3'
    'VudF90eXBlGAMgASgOMhguYmlsbGluZy52MS5EaXNjb3VudFR5cGVSDGRpc2NvdW50VHlwZRIU'
    'CgV2YWx1ZRgEIAEoAVIFdmFsdWUSGgoIY3VycmVuY3kYBSABKAlSCGN1cnJlbmN5EjwKDWFwcG'
    'xpY2FibGVfdG8YBiABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgxhcHBsaWNhYmxlVG8S'
    'NQoIc3RhcnRfYXQYByABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgdzdGFydEF0Ej'
    'EKBmVuZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSBWVuZEF0EikKEG1h'
    'eF9hcHBsaWNhdGlvbnMYCSABKAVSD21heEFwcGxpY2F0aW9ucxIrCgRkYXRhGAogASgLMhcuZ2'
    '9vZ2xlLnByb3RvYnVmLlN0cnVjdFIEZGF0YQ==');

@$core.Deprecated('Use createCatalogVersionRequestDescriptor instead')
const CreateCatalogVersionRequest$json = {
  '1': 'CreateCatalogVersionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'catalog_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'catalogId'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'currency', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'currency'},
    {'1': 'data', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreateCatalogVersionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCatalogVersionRequestDescriptor = $convert.base64Decode(
    'ChtDcmVhdGVDYXRhbG9nVmVyc2lvblJlcXVlc3QSLgoCaWQYASABKAlCHrpIG9gBAXIWEAMYKD'
    'IQWzAtOWEtel8tXXszLDQwfVICaWQSJgoKY2F0YWxvZ19pZBgCIAEoCUIHukgEcgIQAVIJY2F0'
    'YWxvZ0lkEhsKBG5hbWUYAyABKAlCB7pIBHICEAFSBG5hbWUSIwoIY3VycmVuY3kYBCABKAlCB7'
    'pIBHICEANSCGN1cnJlbmN5EisKBGRhdGEYBSABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0'
    'UgRkYXRh');

@$core.Deprecated('Use createCatalogVersionResponseDescriptor instead')
const CreateCatalogVersionResponse$json = {
  '1': 'CreateCatalogVersionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.CatalogVersion', '10': 'data'},
  ],
};

/// Descriptor for `CreateCatalogVersionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCatalogVersionResponseDescriptor = $convert.base64Decode(
    'ChxDcmVhdGVDYXRhbG9nVmVyc2lvblJlc3BvbnNlEi4KBGRhdGEYASABKAsyGi5iaWxsaW5nLn'
    'YxLkNhdGFsb2dWZXJzaW9uUgRkYXRh');

@$core.Deprecated('Use getCatalogVersionRequestDescriptor instead')
const GetCatalogVersionRequest$json = {
  '1': 'GetCatalogVersionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetCatalogVersionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCatalogVersionRequestDescriptor = $convert.base64Decode(
    'ChhHZXRDYXRhbG9nVmVyc2lvblJlcXVlc3QSKwoCaWQYASABKAlCG7pIGHIWEAMYKDIQWzAtOW'
    'Etel8tXXszLDQwfVICaWQ=');

@$core.Deprecated('Use getCatalogVersionResponseDescriptor instead')
const GetCatalogVersionResponse$json = {
  '1': 'GetCatalogVersionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.CatalogVersion', '10': 'data'},
  ],
};

/// Descriptor for `GetCatalogVersionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCatalogVersionResponseDescriptor = $convert.base64Decode(
    'ChlHZXRDYXRhbG9nVmVyc2lvblJlc3BvbnNlEi4KBGRhdGEYASABKAsyGi5iaWxsaW5nLnYxLk'
    'NhdGFsb2dWZXJzaW9uUgRkYXRh');

@$core.Deprecated('Use publishCatalogVersionRequestDescriptor instead')
const PublishCatalogVersionRequest$json = {
  '1': 'PublishCatalogVersionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'effective_at', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'effectiveAt'},
  ],
};

/// Descriptor for `PublishCatalogVersionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List publishCatalogVersionRequestDescriptor = $convert.base64Decode(
    'ChxQdWJsaXNoQ2F0YWxvZ1ZlcnNpb25SZXF1ZXN0EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEF'
    'swLTlhLXpfLV17Myw0MH1SAmlkEj0KDGVmZmVjdGl2ZV9hdBgCIAEoCzIaLmdvb2dsZS5wcm90'
    'b2J1Zi5UaW1lc3RhbXBSC2VmZmVjdGl2ZUF0');

@$core.Deprecated('Use publishCatalogVersionResponseDescriptor instead')
const PublishCatalogVersionResponse$json = {
  '1': 'PublishCatalogVersionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.CatalogVersion', '10': 'data'},
  ],
};

/// Descriptor for `PublishCatalogVersionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List publishCatalogVersionResponseDescriptor = $convert.base64Decode(
    'Ch1QdWJsaXNoQ2F0YWxvZ1ZlcnNpb25SZXNwb25zZRIuCgRkYXRhGAEgASgLMhouYmlsbGluZy'
    '52MS5DYXRhbG9nVmVyc2lvblIEZGF0YQ==');

@$core.Deprecated('Use searchCatalogVersionsResponseDescriptor instead')
const SearchCatalogVersionsResponse$json = {
  '1': 'SearchCatalogVersionsResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.billing.v1.CatalogVersion', '10': 'data'},
  ],
};

/// Descriptor for `SearchCatalogVersionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchCatalogVersionsResponseDescriptor = $convert.base64Decode(
    'Ch1TZWFyY2hDYXRhbG9nVmVyc2lvbnNSZXNwb25zZRIuCgRkYXRhGAEgAygLMhouYmlsbGluZy'
    '52MS5DYXRhbG9nVmVyc2lvblIEZGF0YQ==');

@$core.Deprecated('Use createPlanRequestDescriptor instead')
const CreatePlanRequest$json = {
  '1': 'CreatePlanRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'catalog_version_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'catalogVersionId'},
    {'1': 'external_id', '3': 3, '4': 1, '5': 9, '10': 'externalId'},
    {'1': 'name', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'description', '3': 5, '4': 1, '5': 9, '10': 'description'},
    {'1': 'data', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreatePlanRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createPlanRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVQbGFuUmVxdWVzdBIuCgJpZBgBIAEoCUIeukgb2AEBchYQAxgoMhBbMC05YS16Xy'
    '1dezMsNDB9UgJpZBI1ChJjYXRhbG9nX3ZlcnNpb25faWQYAiABKAlCB7pIBHICEAFSEGNhdGFs'
    'b2dWZXJzaW9uSWQSHwoLZXh0ZXJuYWxfaWQYAyABKAlSCmV4dGVybmFsSWQSGwoEbmFtZRgEIA'
    'EoCUIHukgEcgIQAVIEbmFtZRIgCgtkZXNjcmlwdGlvbhgFIAEoCVILZGVzY3JpcHRpb24SKwoE'
    'ZGF0YRgGIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdGE=');

@$core.Deprecated('Use createPlanResponseDescriptor instead')
const CreatePlanResponse$json = {
  '1': 'CreatePlanResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Plan', '10': 'data'},
  ],
};

/// Descriptor for `CreatePlanResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createPlanResponseDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVQbGFuUmVzcG9uc2USJAoEZGF0YRgBIAEoCzIQLmJpbGxpbmcudjEuUGxhblIEZG'
    'F0YQ==');

@$core.Deprecated('Use createComponentRequestDescriptor instead')
const CreateComponentRequest$json = {
  '1': 'CreateComponentRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'plan_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'planId'},
    {'1': 'external_id', '3': 3, '4': 1, '5': 9, '10': 'externalId'},
    {'1': 'name', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'metric_key', '3': 5, '4': 1, '5': 9, '8': {}, '10': 'metricKey'},
    {'1': 'pricing_model', '3': 6, '4': 1, '5': 14, '6': '.billing.v1.PricingModel', '10': 'pricingModel'},
    {'1': 'aggregation_type', '3': 7, '4': 1, '5': 14, '6': '.billing.v1.AggregationType', '10': 'aggregationType'},
    {'1': 'unit_name', '3': 8, '4': 1, '5': 9, '10': 'unitName'},
    {'1': 'free_quantity', '3': 9, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'freeQuantity'},
    {'1': 'minimum_charge', '3': 10, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'minimumCharge'},
    {'1': 'data', '3': 11, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreateComponentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createComponentRequestDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVDb21wb25lbnRSZXF1ZXN0Ei4KAmlkGAEgASgJQh66SBvYAQFyFhADGCgyEFswLT'
    'lhLXpfLV17Myw0MH1SAmlkEiAKB3BsYW5faWQYAiABKAlCB7pIBHICEAFSBnBsYW5JZBIfCgtl'
    'eHRlcm5hbF9pZBgDIAEoCVIKZXh0ZXJuYWxJZBIbCgRuYW1lGAQgASgJQge6SARyAhABUgRuYW'
    '1lEiYKCm1ldHJpY19rZXkYBSABKAlCB7pIBHICEAFSCW1ldHJpY0tleRI9Cg1wcmljaW5nX21v'
    'ZGVsGAYgASgOMhguYmlsbGluZy52MS5QcmljaW5nTW9kZWxSDHByaWNpbmdNb2RlbBJGChBhZ2'
    'dyZWdhdGlvbl90eXBlGAcgASgOMhsuYmlsbGluZy52MS5BZ2dyZWdhdGlvblR5cGVSD2FnZ3Jl'
    'Z2F0aW9uVHlwZRIbCgl1bml0X25hbWUYCCABKAlSCHVuaXROYW1lEjcKDWZyZWVfcXVhbnRpdH'
    'kYCSABKAsyEi5nb29nbGUudHlwZS5Nb25leVIMZnJlZVF1YW50aXR5EjkKDm1pbmltdW1fY2hh'
    'cmdlGAogASgLMhIuZ29vZ2xlLnR5cGUuTW9uZXlSDW1pbmltdW1DaGFyZ2USKwoEZGF0YRgLIA'
    'EoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdGE=');

@$core.Deprecated('Use createComponentResponseDescriptor instead')
const CreateComponentResponse$json = {
  '1': 'CreateComponentResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Component', '10': 'data'},
  ],
};

/// Descriptor for `CreateComponentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createComponentResponseDescriptor = $convert.base64Decode(
    'ChdDcmVhdGVDb21wb25lbnRSZXNwb25zZRIpCgRkYXRhGAEgASgLMhUuYmlsbGluZy52MS5Db2'
    '1wb25lbnRSBGRhdGE=');

@$core.Deprecated('Use createTierRequestDescriptor instead')
const CreateTierRequest$json = {
  '1': 'CreateTierRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'component_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'componentId'},
    {'1': 'sort_order', '3': 3, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'lower_bound', '3': 4, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'lowerBound'},
    {'1': 'upper_bound', '3': 5, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'upperBound'},
    {'1': 'unit_price', '3': 6, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'unitPrice'},
    {'1': 'flat_fee', '3': 7, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'flatFee'},
  ],
};

/// Descriptor for `CreateTierRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTierRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVUaWVyUmVxdWVzdBIuCgJpZBgBIAEoCUIeukgb2AEBchYQAxgoMhBbMC05YS16Xy'
    '1dezMsNDB9UgJpZBIqCgxjb21wb25lbnRfaWQYAiABKAlCB7pIBHICEAFSC2NvbXBvbmVudElk'
    'Eh0KCnNvcnRfb3JkZXIYAyABKAVSCXNvcnRPcmRlchIzCgtsb3dlcl9ib3VuZBgEIAEoCzISLm'
    'dvb2dsZS50eXBlLk1vbmV5Ugpsb3dlckJvdW5kEjMKC3VwcGVyX2JvdW5kGAUgASgLMhIuZ29v'
    'Z2xlLnR5cGUuTW9uZXlSCnVwcGVyQm91bmQSMQoKdW5pdF9wcmljZRgGIAEoCzISLmdvb2dsZS'
    '50eXBlLk1vbmV5Ugl1bml0UHJpY2USLQoIZmxhdF9mZWUYByABKAsyEi5nb29nbGUudHlwZS5N'
    'b25leVIHZmxhdEZlZQ==');

@$core.Deprecated('Use createTierResponseDescriptor instead')
const CreateTierResponse$json = {
  '1': 'CreateTierResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Tier', '10': 'data'},
  ],
};

/// Descriptor for `CreateTierResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTierResponseDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVUaWVyUmVzcG9uc2USJAoEZGF0YRgBIAEoCzIQLmJpbGxpbmcudjEuVGllclIEZG'
    'F0YQ==');

@$core.Deprecated('Use createSubscriptionRequestDescriptor instead')
const CreateSubscriptionRequest$json = {
  '1': 'CreateSubscriptionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'profile_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'profileId'},
    {'1': 'catalog_version_id', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'catalogVersionId'},
    {'1': 'plan_id', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'planId'},
    {'1': 'start_at', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'startAt'},
    {'1': 'billing_anchor', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'billingAnchor'},
    {'1': 'currency', '3': 7, '4': 1, '5': 9, '8': {}, '10': 'currency'},
    {'1': 'data', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreateSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSubscriptionRequestDescriptor = $convert.base64Decode(
    'ChlDcmVhdGVTdWJzY3JpcHRpb25SZXF1ZXN0Ei4KAmlkGAEgASgJQh66SBvYAQFyFhADGCgyEF'
    'swLTlhLXpfLV17Myw0MH1SAmlkEiYKCnByb2ZpbGVfaWQYAiABKAlCB7pIBHICEAFSCXByb2Zp'
    'bGVJZBI1ChJjYXRhbG9nX3ZlcnNpb25faWQYAyABKAlCB7pIBHICEAFSEGNhdGFsb2dWZXJzaW'
    '9uSWQSIAoHcGxhbl9pZBgEIAEoCUIHukgEcgIQAVIGcGxhbklkEjUKCHN0YXJ0X2F0GAUgASgL'
    'MhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIHc3RhcnRBdBJBCg5iaWxsaW5nX2FuY2hvch'
    'gGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSDWJpbGxpbmdBbmNob3ISIwoIY3Vy'
    'cmVuY3kYByABKAlCB7pIBHICEANSCGN1cnJlbmN5EisKBGRhdGEYCCABKAsyFy5nb29nbGUucH'
    'JvdG9idWYuU3RydWN0UgRkYXRh');

@$core.Deprecated('Use createSubscriptionResponseDescriptor instead')
const CreateSubscriptionResponse$json = {
  '1': 'CreateSubscriptionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Subscription', '10': 'data'},
  ],
};

/// Descriptor for `CreateSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSubscriptionResponseDescriptor = $convert.base64Decode(
    'ChpDcmVhdGVTdWJzY3JpcHRpb25SZXNwb25zZRIsCgRkYXRhGAEgASgLMhguYmlsbGluZy52MS'
    '5TdWJzY3JpcHRpb25SBGRhdGE=');

@$core.Deprecated('Use getSubscriptionRequestDescriptor instead')
const GetSubscriptionRequest$json = {
  '1': 'GetSubscriptionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSubscriptionRequestDescriptor = $convert.base64Decode(
    'ChZHZXRTdWJzY3JpcHRpb25SZXF1ZXN0EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEFswLTlhLX'
    'pfLV17Myw0MH1SAmlk');

@$core.Deprecated('Use getSubscriptionResponseDescriptor instead')
const GetSubscriptionResponse$json = {
  '1': 'GetSubscriptionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Subscription', '10': 'data'},
  ],
};

/// Descriptor for `GetSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSubscriptionResponseDescriptor = $convert.base64Decode(
    'ChdHZXRTdWJzY3JpcHRpb25SZXNwb25zZRIsCgRkYXRhGAEgASgLMhguYmlsbGluZy52MS5TdW'
    'JzY3JpcHRpb25SBGRhdGE=');

@$core.Deprecated('Use cancelSubscriptionRequestDescriptor instead')
const CancelSubscriptionRequest$json = {
  '1': 'CancelSubscriptionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `CancelSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelSubscriptionRequestDescriptor = $convert.base64Decode(
    'ChlDYW5jZWxTdWJzY3JpcHRpb25SZXF1ZXN0EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEFswLT'
    'lhLXpfLV17Myw0MH1SAmlk');

@$core.Deprecated('Use cancelSubscriptionResponseDescriptor instead')
const CancelSubscriptionResponse$json = {
  '1': 'CancelSubscriptionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Subscription', '10': 'data'},
  ],
};

/// Descriptor for `CancelSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelSubscriptionResponseDescriptor = $convert.base64Decode(
    'ChpDYW5jZWxTdWJzY3JpcHRpb25SZXNwb25zZRIsCgRkYXRhGAEgASgLMhguYmlsbGluZy52MS'
    '5TdWJzY3JpcHRpb25SBGRhdGE=');

@$core.Deprecated('Use listSubscriptionsRequestDescriptor instead')
const ListSubscriptionsRequest$json = {
  '1': 'ListSubscriptionsRequest',
  '2': [
    {'1': 'profile_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'profileId'},
  ],
};

/// Descriptor for `ListSubscriptionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSubscriptionsRequestDescriptor = $convert.base64Decode(
    'ChhMaXN0U3Vic2NyaXB0aW9uc1JlcXVlc3QSJgoKcHJvZmlsZV9pZBgBIAEoCUIHukgEcgIQAV'
    'IJcHJvZmlsZUlk');

@$core.Deprecated('Use listSubscriptionsResponseDescriptor instead')
const ListSubscriptionsResponse$json = {
  '1': 'ListSubscriptionsResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.billing.v1.Subscription', '10': 'data'},
  ],
};

/// Descriptor for `ListSubscriptionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSubscriptionsResponseDescriptor = $convert.base64Decode(
    'ChlMaXN0U3Vic2NyaXB0aW9uc1Jlc3BvbnNlEiwKBGRhdGEYASADKAsyGC5iaWxsaW5nLnYxLl'
    'N1YnNjcmlwdGlvblIEZGF0YQ==');

@$core.Deprecated('Use ingestUsageEventRequestDescriptor instead')
const IngestUsageEventRequest$json = {
  '1': 'IngestUsageEventRequest',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.billing.v1.UsageEvent', '10': 'data'},
  ],
};

/// Descriptor for `IngestUsageEventRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ingestUsageEventRequestDescriptor = $convert.base64Decode(
    'ChdJbmdlc3RVc2FnZUV2ZW50UmVxdWVzdBIqCgRkYXRhGAEgAygLMhYuYmlsbGluZy52MS5Vc2'
    'FnZUV2ZW50UgRkYXRh');

@$core.Deprecated('Use ingestUsageEventResponseDescriptor instead')
const IngestUsageEventResponse$json = {
  '1': 'IngestUsageEventResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 9, '10': 'data'},
  ],
};

/// Descriptor for `IngestUsageEventResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ingestUsageEventResponseDescriptor = $convert.base64Decode(
    'ChhJbmdlc3RVc2FnZUV2ZW50UmVzcG9uc2USEgoEZGF0YRgBIAMoCVIEZGF0YQ==');

@$core.Deprecated('Use searchUsageEventsResponseDescriptor instead')
const SearchUsageEventsResponse$json = {
  '1': 'SearchUsageEventsResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.billing.v1.UsageEvent', '10': 'data'},
  ],
};

/// Descriptor for `SearchUsageEventsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchUsageEventsResponseDescriptor = $convert.base64Decode(
    'ChlTZWFyY2hVc2FnZUV2ZW50c1Jlc3BvbnNlEioKBGRhdGEYASADKAsyFi5iaWxsaW5nLnYxLl'
    'VzYWdlRXZlbnRSBGRhdGE=');

@$core.Deprecated('Use runBillingRequestDescriptor instead')
const RunBillingRequest$json = {
  '1': 'RunBillingRequest',
  '2': [
    {'1': 'subscription_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'subscriptionId'},
    {'1': 'period_start', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'periodStart'},
    {'1': 'period_end', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'periodEnd'},
  ],
};

/// Descriptor for `RunBillingRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List runBillingRequestDescriptor = $convert.base64Decode(
    'ChFSdW5CaWxsaW5nUmVxdWVzdBIwCg9zdWJzY3JpcHRpb25faWQYASABKAlCB7pIBHICEAFSDn'
    'N1YnNjcmlwdGlvbklkEj0KDHBlcmlvZF9zdGFydBgCIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5U'
    'aW1lc3RhbXBSC3BlcmlvZFN0YXJ0EjkKCnBlcmlvZF9lbmQYAyABKAsyGi5nb29nbGUucHJvdG'
    '9idWYuVGltZXN0YW1wUglwZXJpb2RFbmQ=');

@$core.Deprecated('Use runBillingResponseDescriptor instead')
const RunBillingResponse$json = {
  '1': 'RunBillingResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.BillingRun', '10': 'data'},
  ],
};

/// Descriptor for `RunBillingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List runBillingResponseDescriptor = $convert.base64Decode(
    'ChJSdW5CaWxsaW5nUmVzcG9uc2USKgoEZGF0YRgBIAEoCzIWLmJpbGxpbmcudjEuQmlsbGluZ1'
    'J1blIEZGF0YQ==');

@$core.Deprecated('Use getBillingRunRequestDescriptor instead')
const GetBillingRunRequest$json = {
  '1': 'GetBillingRunRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetBillingRunRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBillingRunRequestDescriptor = $convert.base64Decode(
    'ChRHZXRCaWxsaW5nUnVuUmVxdWVzdBIrCgJpZBgBIAEoCUIbukgYchYQAxgoMhBbMC05YS16Xy'
    '1dezMsNDB9UgJpZA==');

@$core.Deprecated('Use getBillingRunResponseDescriptor instead')
const GetBillingRunResponse$json = {
  '1': 'GetBillingRunResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.BillingRun', '10': 'data'},
  ],
};

/// Descriptor for `GetBillingRunResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBillingRunResponseDescriptor = $convert.base64Decode(
    'ChVHZXRCaWxsaW5nUnVuUmVzcG9uc2USKgoEZGF0YRgBIAEoCzIWLmJpbGxpbmcudjEuQmlsbG'
    'luZ1J1blIEZGF0YQ==');

@$core.Deprecated('Use getInvoiceRequestDescriptor instead')
const GetInvoiceRequest$json = {
  '1': 'GetInvoiceRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetInvoiceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getInvoiceRequestDescriptor = $convert.base64Decode(
    'ChFHZXRJbnZvaWNlUmVxdWVzdBIrCgJpZBgBIAEoCUIbukgYchYQAxgoMhBbMC05YS16Xy1dez'
    'MsNDB9UgJpZA==');

@$core.Deprecated('Use getInvoiceResponseDescriptor instead')
const GetInvoiceResponse$json = {
  '1': 'GetInvoiceResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Invoice', '10': 'data'},
  ],
};

/// Descriptor for `GetInvoiceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getInvoiceResponseDescriptor = $convert.base64Decode(
    'ChJHZXRJbnZvaWNlUmVzcG9uc2USJwoEZGF0YRgBIAEoCzITLmJpbGxpbmcudjEuSW52b2ljZV'
    'IEZGF0YQ==');

@$core.Deprecated('Use issueInvoiceRequestDescriptor instead')
const IssueInvoiceRequest$json = {
  '1': 'IssueInvoiceRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `IssueInvoiceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List issueInvoiceRequestDescriptor = $convert.base64Decode(
    'ChNJc3N1ZUludm9pY2VSZXF1ZXN0EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEFswLTlhLXpfLV'
    '17Myw0MH1SAmlk');

@$core.Deprecated('Use issueInvoiceResponseDescriptor instead')
const IssueInvoiceResponse$json = {
  '1': 'IssueInvoiceResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Invoice', '10': 'data'},
  ],
};

/// Descriptor for `IssueInvoiceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List issueInvoiceResponseDescriptor = $convert.base64Decode(
    'ChRJc3N1ZUludm9pY2VSZXNwb25zZRInCgRkYXRhGAEgASgLMhMuYmlsbGluZy52MS5JbnZvaW'
    'NlUgRkYXRh');

@$core.Deprecated('Use voidInvoiceRequestDescriptor instead')
const VoidInvoiceRequest$json = {
  '1': 'VoidInvoiceRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `VoidInvoiceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidInvoiceRequestDescriptor = $convert.base64Decode(
    'ChJWb2lkSW52b2ljZVJlcXVlc3QSKwoCaWQYASABKAlCG7pIGHIWEAMYKDIQWzAtOWEtel8tXX'
    'szLDQwfVICaWQ=');

@$core.Deprecated('Use voidInvoiceResponseDescriptor instead')
const VoidInvoiceResponse$json = {
  '1': 'VoidInvoiceResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Invoice', '10': 'data'},
  ],
};

/// Descriptor for `VoidInvoiceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidInvoiceResponseDescriptor = $convert.base64Decode(
    'ChNWb2lkSW52b2ljZVJlc3BvbnNlEicKBGRhdGEYASABKAsyEy5iaWxsaW5nLnYxLkludm9pY2'
    'VSBGRhdGE=');

@$core.Deprecated('Use recordPaymentRequestDescriptor instead')
const RecordPaymentRequest$json = {
  '1': 'RecordPaymentRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `RecordPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List recordPaymentRequestDescriptor = $convert.base64Decode(
    'ChRSZWNvcmRQYXltZW50UmVxdWVzdBIrCgJpZBgBIAEoCUIbukgYchYQAxgoMhBbMC05YS16Xy'
    '1dezMsNDB9UgJpZA==');

@$core.Deprecated('Use recordPaymentResponseDescriptor instead')
const RecordPaymentResponse$json = {
  '1': 'RecordPaymentResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Invoice', '10': 'data'},
  ],
};

/// Descriptor for `RecordPaymentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List recordPaymentResponseDescriptor = $convert.base64Decode(
    'ChVSZWNvcmRQYXltZW50UmVzcG9uc2USJwoEZGF0YRgBIAEoCzITLmJpbGxpbmcudjEuSW52b2'
    'ljZVIEZGF0YQ==');

@$core.Deprecated('Use searchInvoicesResponseDescriptor instead')
const SearchInvoicesResponse$json = {
  '1': 'SearchInvoicesResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.billing.v1.Invoice', '10': 'data'},
  ],
};

/// Descriptor for `SearchInvoicesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchInvoicesResponseDescriptor = $convert.base64Decode(
    'ChZTZWFyY2hJbnZvaWNlc1Jlc3BvbnNlEicKBGRhdGEYASADKAsyEy5iaWxsaW5nLnYxLkludm'
    '9pY2VSBGRhdGE=');

@$core.Deprecated('Use grantCreditRequestDescriptor instead')
const GrantCreditRequest$json = {
  '1': 'GrantCreditRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'profile_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'profileId'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'amount', '3': 4, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'amount'},
    {'1': 'currency', '3': 5, '4': 1, '5': 9, '8': {}, '10': 'currency'},
    {'1': 'expires_at', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'expiresAt'},
    {'1': 'priority', '3': 7, '4': 1, '5': 5, '10': 'priority'},
    {'1': 'data', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `GrantCreditRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List grantCreditRequestDescriptor = $convert.base64Decode(
    'ChJHcmFudENyZWRpdFJlcXVlc3QSLgoCaWQYASABKAlCHrpIG9gBAXIWEAMYKDIQWzAtOWEtel'
    '8tXXszLDQwfVICaWQSJgoKcHJvZmlsZV9pZBgCIAEoCUIHukgEcgIQAVIJcHJvZmlsZUlkEhsK'
    'BG5hbWUYAyABKAlCB7pIBHICEAFSBG5hbWUSKgoGYW1vdW50GAQgASgLMhIuZ29vZ2xlLnR5cG'
    'UuTW9uZXlSBmFtb3VudBIjCghjdXJyZW5jeRgFIAEoCUIHukgEcgIQA1IIY3VycmVuY3kSOQoK'
    'ZXhwaXJlc19hdBgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWV4cGlyZXNBdB'
    'IaCghwcmlvcml0eRgHIAEoBVIIcHJpb3JpdHkSKwoEZGF0YRgIIAEoCzIXLmdvb2dsZS5wcm90'
    'b2J1Zi5TdHJ1Y3RSBGRhdGE=');

@$core.Deprecated('Use grantCreditResponseDescriptor instead')
const GrantCreditResponse$json = {
  '1': 'GrantCreditResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.CreditGrant', '10': 'data'},
  ],
};

/// Descriptor for `GrantCreditResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List grantCreditResponseDescriptor = $convert.base64Decode(
    'ChNHcmFudENyZWRpdFJlc3BvbnNlEisKBGRhdGEYASABKAsyFy5iaWxsaW5nLnYxLkNyZWRpdE'
    'dyYW50UgRkYXRh');

@$core.Deprecated('Use getCreditBalanceRequestDescriptor instead')
const GetCreditBalanceRequest$json = {
  '1': 'GetCreditBalanceRequest',
  '2': [
    {'1': 'profile_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'profileId'},
    {'1': 'currency', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'currency'},
  ],
};

/// Descriptor for `GetCreditBalanceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCreditBalanceRequestDescriptor = $convert.base64Decode(
    'ChdHZXRDcmVkaXRCYWxhbmNlUmVxdWVzdBImCgpwcm9maWxlX2lkGAEgASgJQge6SARyAhABUg'
    'lwcm9maWxlSWQSIwoIY3VycmVuY3kYAiABKAlCB7pIBHICEANSCGN1cnJlbmN5');

@$core.Deprecated('Use getCreditBalanceResponseDescriptor instead')
const GetCreditBalanceResponse$json = {
  '1': 'GetCreditBalanceResponse',
  '2': [
    {'1': 'balance', '3': 1, '4': 1, '5': 11, '6': '.google.type.Money', '10': 'balance'},
  ],
};

/// Descriptor for `GetCreditBalanceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCreditBalanceResponseDescriptor = $convert.base64Decode(
    'ChhHZXRDcmVkaXRCYWxhbmNlUmVzcG9uc2USLAoHYmFsYW5jZRgBIAEoCzISLmdvb2dsZS50eX'
    'BlLk1vbmV5UgdiYWxhbmNl');

@$core.Deprecated('Use createDiscountRequestDescriptor instead')
const CreateDiscountRequest$json = {
  '1': 'CreateDiscountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'discount_type', '3': 3, '4': 1, '5': 14, '6': '.billing.v1.DiscountType', '10': 'discountType'},
    {'1': 'value', '3': 4, '4': 1, '5': 1, '10': 'value'},
    {'1': 'currency', '3': 5, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'applicable_to', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'applicableTo'},
    {'1': 'start_at', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'startAt'},
    {'1': 'end_at', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'endAt'},
    {'1': 'max_applications', '3': 9, '4': 1, '5': 5, '10': 'maxApplications'},
    {'1': 'data', '3': 10, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreateDiscountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDiscountRequestDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVEaXNjb3VudFJlcXVlc3QSLgoCaWQYASABKAlCHrpIG9gBAXIWEAMYKDIQWzAtOW'
    'Etel8tXXszLDQwfVICaWQSGwoEbmFtZRgCIAEoCUIHukgEcgIQAVIEbmFtZRI9Cg1kaXNjb3Vu'
    'dF90eXBlGAMgASgOMhguYmlsbGluZy52MS5EaXNjb3VudFR5cGVSDGRpc2NvdW50VHlwZRIUCg'
    'V2YWx1ZRgEIAEoAVIFdmFsdWUSGgoIY3VycmVuY3kYBSABKAlSCGN1cnJlbmN5EjwKDWFwcGxp'
    'Y2FibGVfdG8YBiABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgxhcHBsaWNhYmxlVG8SNQ'
    'oIc3RhcnRfYXQYByABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgdzdGFydEF0EjEK'
    'BmVuZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSBWVuZEF0EikKEG1heF'
    '9hcHBsaWNhdGlvbnMYCSABKAVSD21heEFwcGxpY2F0aW9ucxIrCgRkYXRhGAogASgLMhcuZ29v'
    'Z2xlLnByb3RvYnVmLlN0cnVjdFIEZGF0YQ==');

@$core.Deprecated('Use createDiscountResponseDescriptor instead')
const CreateDiscountResponse$json = {
  '1': 'CreateDiscountResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.billing.v1.Discount', '10': 'data'},
  ],
};

/// Descriptor for `CreateDiscountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDiscountResponseDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVEaXNjb3VudFJlc3BvbnNlEigKBGRhdGEYASABKAsyFC5iaWxsaW5nLnYxLkRpc2'
    'NvdW50UgRkYXRh');

@$core.Deprecated('Use searchDiscountsResponseDescriptor instead')
const SearchDiscountsResponse$json = {
  '1': 'SearchDiscountsResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.billing.v1.Discount', '10': 'data'},
  ],
};

/// Descriptor for `SearchDiscountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchDiscountsResponseDescriptor = $convert.base64Decode(
    'ChdTZWFyY2hEaXNjb3VudHNSZXNwb25zZRIoCgRkYXRhGAEgAygLMhQuYmlsbGluZy52MS5EaX'
    'Njb3VudFIEZGF0YQ==');

const $core.Map<$core.String, $core.dynamic> BillingServiceBase$json = {
  '1': 'BillingService',
  '2': [
    {'1': 'CreateCatalogVersion', '2': '.billing.v1.CreateCatalogVersionRequest', '3': '.billing.v1.CreateCatalogVersionResponse', '4': {}},
    {
      '1': 'GetCatalogVersion',
      '2': '.billing.v1.GetCatalogVersionRequest',
      '3': '.billing.v1.GetCatalogVersionResponse',
      '4': {'34': 1},
    },
    {'1': 'PublishCatalogVersion', '2': '.billing.v1.PublishCatalogVersionRequest', '3': '.billing.v1.PublishCatalogVersionResponse', '4': {}},
    {
      '1': 'SearchCatalogVersions',
      '2': '.common.v1.SearchRequest',
      '3': '.billing.v1.SearchCatalogVersionsResponse',
      '4': {'34': 1},
      '6': true,
    },
    {'1': 'CreatePlan', '2': '.billing.v1.CreatePlanRequest', '3': '.billing.v1.CreatePlanResponse', '4': {}},
    {'1': 'CreateComponent', '2': '.billing.v1.CreateComponentRequest', '3': '.billing.v1.CreateComponentResponse', '4': {}},
    {'1': 'CreateTier', '2': '.billing.v1.CreateTierRequest', '3': '.billing.v1.CreateTierResponse', '4': {}},
    {'1': 'CreateSubscription', '2': '.billing.v1.CreateSubscriptionRequest', '3': '.billing.v1.CreateSubscriptionResponse', '4': {}},
    {
      '1': 'GetSubscription',
      '2': '.billing.v1.GetSubscriptionRequest',
      '3': '.billing.v1.GetSubscriptionResponse',
      '4': {'34': 1},
    },
    {'1': 'CancelSubscription', '2': '.billing.v1.CancelSubscriptionRequest', '3': '.billing.v1.CancelSubscriptionResponse', '4': {}},
    {
      '1': 'ListSubscriptions',
      '2': '.billing.v1.ListSubscriptionsRequest',
      '3': '.billing.v1.ListSubscriptionsResponse',
      '4': {'34': 1},
    },
    {'1': 'IngestUsageEvent', '2': '.billing.v1.IngestUsageEventRequest', '3': '.billing.v1.IngestUsageEventResponse', '4': {}},
    {
      '1': 'SearchUsageEvents',
      '2': '.common.v1.SearchRequest',
      '3': '.billing.v1.SearchUsageEventsResponse',
      '4': {'34': 1},
      '6': true,
    },
    {'1': 'RunBilling', '2': '.billing.v1.RunBillingRequest', '3': '.billing.v1.RunBillingResponse', '4': {}},
    {
      '1': 'GetBillingRun',
      '2': '.billing.v1.GetBillingRunRequest',
      '3': '.billing.v1.GetBillingRunResponse',
      '4': {'34': 1},
    },
    {
      '1': 'GetInvoice',
      '2': '.billing.v1.GetInvoiceRequest',
      '3': '.billing.v1.GetInvoiceResponse',
      '4': {'34': 1},
    },
    {'1': 'IssueInvoice', '2': '.billing.v1.IssueInvoiceRequest', '3': '.billing.v1.IssueInvoiceResponse', '4': {}},
    {'1': 'VoidInvoice', '2': '.billing.v1.VoidInvoiceRequest', '3': '.billing.v1.VoidInvoiceResponse', '4': {}},
    {'1': 'RecordPayment', '2': '.billing.v1.RecordPaymentRequest', '3': '.billing.v1.RecordPaymentResponse', '4': {}},
    {
      '1': 'SearchInvoices',
      '2': '.common.v1.SearchRequest',
      '3': '.billing.v1.SearchInvoicesResponse',
      '4': {'34': 1},
      '6': true,
    },
    {'1': 'GrantCredit', '2': '.billing.v1.GrantCreditRequest', '3': '.billing.v1.GrantCreditResponse', '4': {}},
    {
      '1': 'GetCreditBalance',
      '2': '.billing.v1.GetCreditBalanceRequest',
      '3': '.billing.v1.GetCreditBalanceResponse',
      '4': {'34': 1},
    },
    {'1': 'CreateDiscount', '2': '.billing.v1.CreateDiscountRequest', '3': '.billing.v1.CreateDiscountResponse', '4': {}},
    {
      '1': 'SearchDiscounts',
      '2': '.common.v1.SearchRequest',
      '3': '.billing.v1.SearchDiscountsResponse',
      '4': {'34': 1},
      '6': true,
    },
  ],
  '3': {},
};

@$core.Deprecated('Use billingServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>> BillingServiceBase$messageJson = {
  '.billing.v1.CreateCatalogVersionRequest': CreateCatalogVersionRequest$json,
  '.google.protobuf.Struct': $6.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $6.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $6.Value$json,
  '.google.protobuf.ListValue': $6.ListValue$json,
  '.billing.v1.CreateCatalogVersionResponse': CreateCatalogVersionResponse$json,
  '.billing.v1.CatalogVersion': CatalogVersion$json,
  '.google.protobuf.Timestamp': $2.Timestamp$json,
  '.billing.v1.Plan': Plan$json,
  '.billing.v1.Component': Component$json,
  '.google.type.Money': $7.Money$json,
  '.billing.v1.Tier': Tier$json,
  '.billing.v1.GetCatalogVersionRequest': GetCatalogVersionRequest$json,
  '.billing.v1.GetCatalogVersionResponse': GetCatalogVersionResponse$json,
  '.billing.v1.PublishCatalogVersionRequest': PublishCatalogVersionRequest$json,
  '.billing.v1.PublishCatalogVersionResponse': PublishCatalogVersionResponse$json,
  '.common.v1.SearchRequest': $9.SearchRequest$json,
  '.common.v1.PageCursor': $9.PageCursor$json,
  '.billing.v1.SearchCatalogVersionsResponse': SearchCatalogVersionsResponse$json,
  '.billing.v1.CreatePlanRequest': CreatePlanRequest$json,
  '.billing.v1.CreatePlanResponse': CreatePlanResponse$json,
  '.billing.v1.CreateComponentRequest': CreateComponentRequest$json,
  '.billing.v1.CreateComponentResponse': CreateComponentResponse$json,
  '.billing.v1.CreateTierRequest': CreateTierRequest$json,
  '.billing.v1.CreateTierResponse': CreateTierResponse$json,
  '.billing.v1.CreateSubscriptionRequest': CreateSubscriptionRequest$json,
  '.billing.v1.CreateSubscriptionResponse': CreateSubscriptionResponse$json,
  '.billing.v1.Subscription': Subscription$json,
  '.billing.v1.GetSubscriptionRequest': GetSubscriptionRequest$json,
  '.billing.v1.GetSubscriptionResponse': GetSubscriptionResponse$json,
  '.billing.v1.CancelSubscriptionRequest': CancelSubscriptionRequest$json,
  '.billing.v1.CancelSubscriptionResponse': CancelSubscriptionResponse$json,
  '.billing.v1.ListSubscriptionsRequest': ListSubscriptionsRequest$json,
  '.billing.v1.ListSubscriptionsResponse': ListSubscriptionsResponse$json,
  '.billing.v1.IngestUsageEventRequest': IngestUsageEventRequest$json,
  '.billing.v1.UsageEvent': UsageEvent$json,
  '.google.type.Interval': $8.Interval$json,
  '.billing.v1.IngestUsageEventResponse': IngestUsageEventResponse$json,
  '.billing.v1.SearchUsageEventsResponse': SearchUsageEventsResponse$json,
  '.billing.v1.RunBillingRequest': RunBillingRequest$json,
  '.billing.v1.RunBillingResponse': RunBillingResponse$json,
  '.billing.v1.BillingRun': BillingRun$json,
  '.billing.v1.GetBillingRunRequest': GetBillingRunRequest$json,
  '.billing.v1.GetBillingRunResponse': GetBillingRunResponse$json,
  '.billing.v1.GetInvoiceRequest': GetInvoiceRequest$json,
  '.billing.v1.GetInvoiceResponse': GetInvoiceResponse$json,
  '.billing.v1.Invoice': Invoice$json,
  '.billing.v1.InvoiceLine': InvoiceLine$json,
  '.billing.v1.IssueInvoiceRequest': IssueInvoiceRequest$json,
  '.billing.v1.IssueInvoiceResponse': IssueInvoiceResponse$json,
  '.billing.v1.VoidInvoiceRequest': VoidInvoiceRequest$json,
  '.billing.v1.VoidInvoiceResponse': VoidInvoiceResponse$json,
  '.billing.v1.RecordPaymentRequest': RecordPaymentRequest$json,
  '.billing.v1.RecordPaymentResponse': RecordPaymentResponse$json,
  '.billing.v1.SearchInvoicesResponse': SearchInvoicesResponse$json,
  '.billing.v1.GrantCreditRequest': GrantCreditRequest$json,
  '.billing.v1.GrantCreditResponse': GrantCreditResponse$json,
  '.billing.v1.CreditGrant': CreditGrant$json,
  '.billing.v1.GetCreditBalanceRequest': GetCreditBalanceRequest$json,
  '.billing.v1.GetCreditBalanceResponse': GetCreditBalanceResponse$json,
  '.billing.v1.CreateDiscountRequest': CreateDiscountRequest$json,
  '.billing.v1.CreateDiscountResponse': CreateDiscountResponse$json,
  '.billing.v1.Discount': Discount$json,
  '.billing.v1.SearchDiscountsResponse': SearchDiscountsResponse$json,
};

/// Descriptor for `BillingService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List billingServiceDescriptor = $convert.base64Decode(
    'Cg5CaWxsaW5nU2VydmljZRLcAgoUQ3JlYXRlQ2F0YWxvZ1ZlcnNpb24SJy5iaWxsaW5nLnYxLk'
    'NyZWF0ZUNhdGFsb2dWZXJzaW9uUmVxdWVzdBooLmJpbGxpbmcudjEuQ3JlYXRlQ2F0YWxvZ1Zl'
    'cnNpb25SZXNwb25zZSLwAbpH2AEKB0NhdGFsb2cSHENyZWF0ZSBhIG5ldyBjYXRhbG9nIHZlcn'
    'Npb24amAFDcmVhdGVzIGEgbmV3IGltbXV0YWJsZSBjYXRhbG9nIHZlcnNpb24gY29udGFpbmlu'
    'ZyBwbGFucywgY29tcG9uZW50cywgYW5kIHByaWNpbmcgdGllcnMuIENhdGFsb2dzIG11c3QgYm'
    'UgcHVibGlzaGVkIGJlZm9yZSB0aGV5IGNhbiBiZSB1c2VkIGZvciBiaWxsaW5nLioUY3JlYXRl'
    'Q2F0YWxvZ1ZlcnNpb26CtRgQCg5jYXRhbG9nX21hbmFnZRKQAgoRR2V0Q2F0YWxvZ1ZlcnNpb2'
    '4SJC5iaWxsaW5nLnYxLkdldENhdGFsb2dWZXJzaW9uUmVxdWVzdBolLmJpbGxpbmcudjEuR2V0'
    'Q2F0YWxvZ1ZlcnNpb25SZXNwb25zZSKtAZACAbpHlAEKB0NhdGFsb2cSFUdldCBhIGNhdGFsb2'
    'cgdmVyc2lvbhpfUmV0cmlldmVzIGEgc3BlY2lmaWMgY2F0YWxvZyB2ZXJzaW9uIGJ5IElELCBp'
    'bmNsdWRpbmcgaXRzIHBsYW5zLCBjb21wb25lbnRzLCBhbmQgcHJpY2luZyB0aWVycy4qEWdldE'
    'NhdGFsb2dWZXJzaW9ugrUYDgoMY2F0YWxvZ192aWV3ErICChVQdWJsaXNoQ2F0YWxvZ1ZlcnNp'
    'b24SKC5iaWxsaW5nLnYxLlB1Ymxpc2hDYXRhbG9nVmVyc2lvblJlcXVlc3QaKS5iaWxsaW5nLn'
    'YxLlB1Ymxpc2hDYXRhbG9nVmVyc2lvblJlc3BvbnNlIsMBukerAQoHQ2F0YWxvZxIZUHVibGlz'
    'aCBhIGNhdGFsb2cgdmVyc2lvbhpuUHVibGlzaGVzIGEgY2F0YWxvZyB2ZXJzaW9uIHdpdGggYW'
    '4gZWZmZWN0aXZlIGRhdGUuIFB1Ymxpc2hlZCBjYXRhbG9ncyBiZWNvbWUgYXZhaWxhYmxlIGZv'
    'ciBuZXcgc3Vic2NyaXB0aW9ucy4qFXB1Ymxpc2hDYXRhbG9nVmVyc2lvboK1GBAKDmNhdGFsb2'
    'dfbWFuYWdlEp4CChVTZWFyY2hDYXRhbG9nVmVyc2lvbnMSGC5jb21tb24udjEuU2VhcmNoUmVx'
    'dWVzdBopLmJpbGxpbmcudjEuU2VhcmNoQ2F0YWxvZ1ZlcnNpb25zUmVzcG9uc2UivQGQAgG6R6'
    'QBCgdDYXRhbG9nEhdTZWFyY2ggY2F0YWxvZyB2ZXJzaW9ucxppU2VhcmNoZXMgZm9yIGNhdGFs'
    'b2cgdmVyc2lvbnMgbWF0Y2hpbmcgc3BlY2lmaWVkIGNyaXRlcmlhLiBSZXR1cm5zIGEgc3RyZW'
    'FtIG9mIG1hdGNoaW5nIGNhdGFsb2cgdmVyc2lvbnMuKhVzZWFyY2hDYXRhbG9nVmVyc2lvbnOC'
    'tRgOCgxjYXRhbG9nX3ZpZXcwARLxAQoKQ3JlYXRlUGxhbhIdLmJpbGxpbmcudjEuQ3JlYXRlUG'
    'xhblJlcXVlc3QaHi5iaWxsaW5nLnYxLkNyZWF0ZVBsYW5SZXNwb25zZSKjAbpHjgEKB0NhdGFs'
    'b2cSEUNyZWF0ZSBhIG5ldyBwbGFuGmRDcmVhdGVzIGEgbmV3IGJpbGxpbmcgcGxhbiB3aXRoaW'
    '4gYSBjYXRhbG9nIHZlcnNpb24uIFBsYW5zIGNvbnRhaW4gYmlsbGFibGUgY29tcG9uZW50cyB3'
    'aXRoIHByaWNpbmcuKgpjcmVhdGVQbGFugrUYDQoLcGxhbl9tYW5hZ2USuQIKD0NyZWF0ZUNvbX'
    'BvbmVudBIiLmJpbGxpbmcudjEuQ3JlYXRlQ29tcG9uZW50UmVxdWVzdBojLmJpbGxpbmcudjEu'
    'Q3JlYXRlQ29tcG9uZW50UmVzcG9uc2Ui3AG6R8IBCgdDYXRhbG9nEhZDcmVhdGUgYSBuZXcgY2'
    '9tcG9uZW50Go0BQ3JlYXRlcyBhIG5ldyBiaWxsYWJsZSBjb21wb25lbnQgd2l0aGluIGEgcGxh'
    'bi4gQ29tcG9uZW50cyBkZWZpbmUgdGhlIHByaWNpbmcgbW9kZWwsIGFnZ3JlZ2F0aW9uIHR5cG'
    'UsIGFuZCBmcmVlIHRpZXIvbWluaW11bSBjaGFyZ2Ugc2V0dGluZ3MuKg9jcmVhdGVDb21wb25l'
    'bnSCtRgSChBjb21wb25lbnRfbWFuYWdlEpECCgpDcmVhdGVUaWVyEh0uYmlsbGluZy52MS5Dcm'
    'VhdGVUaWVyUmVxdWVzdBoeLmJpbGxpbmcudjEuQ3JlYXRlVGllclJlc3BvbnNlIsMBukeuAQoH'
    'Q2F0YWxvZxIZQ3JlYXRlIGEgbmV3IHByaWNpbmcgdGllchp8Q3JlYXRlcyBhIG5ldyBwcmljaW'
    '5nIHRpZXIgd2l0aGluIGEgY29tcG9uZW50LiBUaWVycyBkZWZpbmUgcHJpY2UgYnJhY2tldHMg'
    'Zm9yIHRpZXJlZCwgdm9sdW1lLCBhbmQgc3RhaXJzdGVwIHByaWNpbmcgbW9kZWxzLioKY3JlYX'
    'RlVGllcoK1GA0KC3RpZXJfbWFuYWdlEp4CChJDcmVhdGVTdWJzY3JpcHRpb24SJS5iaWxsaW5n'
    'LnYxLkNyZWF0ZVN1YnNjcmlwdGlvblJlcXVlc3QaJi5iaWxsaW5nLnYxLkNyZWF0ZVN1YnNjcm'
    'lwdGlvblJlc3BvbnNlIrgBukebAQoNU3Vic2NyaXB0aW9ucxIZQ3JlYXRlIGEgbmV3IHN1YnNj'
    'cmlwdGlvbhpbQ3JlYXRlcyBhIG5ldyBzdWJzY3JpcHRpb24gbGlua2luZyBhIGN1c3RvbWVyIH'
    'RvIGEgcGxhbiB3aXRoaW4gYSBwdWJsaXNoZWQgY2F0YWxvZyB2ZXJzaW9uLioSY3JlYXRlU3Vi'
    'c2NyaXB0aW9ugrUYFQoTc3Vic2NyaXB0aW9uX21hbmFnZRL0AQoPR2V0U3Vic2NyaXB0aW9uEi'
    'IuYmlsbGluZy52MS5HZXRTdWJzY3JpcHRpb25SZXF1ZXN0GiMuYmlsbGluZy52MS5HZXRTdWJz'
    'Y3JpcHRpb25SZXNwb25zZSKXAZACAbpHegoNU3Vic2NyaXB0aW9ucxISR2V0IGEgc3Vic2NyaX'
    'B0aW9uGkRSZXRyaWV2ZXMgYSBzcGVjaWZpYyBzdWJzY3JpcHRpb24gYnkgSUQgaW5jbHVkaW5n'
    'IGl0cyBjdXJyZW50IHN0YXRlLioPZ2V0U3Vic2NyaXB0aW9ugrUYEwoRc3Vic2NyaXB0aW9uX3'
    'ZpZXcSpQIKEkNhbmNlbFN1YnNjcmlwdGlvbhIlLmJpbGxpbmcudjEuQ2FuY2VsU3Vic2NyaXB0'
    'aW9uUmVxdWVzdBomLmJpbGxpbmcudjEuQ2FuY2VsU3Vic2NyaXB0aW9uUmVzcG9uc2UivwG6R6'
    'IBCg1TdWJzY3JpcHRpb25zEhVDYW5jZWwgYSBzdWJzY3JpcHRpb24aZkNhbmNlbHMgYW4gYWN0'
    'aXZlIHN1YnNjcmlwdGlvbi4gVGhlIHN1YnNjcmlwdGlvbiByZW1haW5zIGFjdGl2ZSB1bnRpbC'
    'B0aGUgY3VycmVudCBiaWxsaW5nIHBlcmlvZCBlbmRzLioSY2FuY2VsU3Vic2NyaXB0aW9ugrUY'
    'FQoTc3Vic2NyaXB0aW9uX21hbmFnZRLvAQoRTGlzdFN1YnNjcmlwdGlvbnMSJC5iaWxsaW5nLn'
    'YxLkxpc3RTdWJzY3JpcHRpb25zUmVxdWVzdBolLmJpbGxpbmcudjEuTGlzdFN1YnNjcmlwdGlv'
    'bnNSZXNwb25zZSKMAZACAbpHbwoNU3Vic2NyaXB0aW9ucxIbTGlzdCBjdXN0b21lciBzdWJzY3'
    'JpcHRpb25zGi5MaXN0cyBhbGwgYWN0aXZlIHN1YnNjcmlwdGlvbnMgZm9yIGEgY3VzdG9tZXIu'
    'KhFsaXN0U3Vic2NyaXB0aW9uc4K1GBMKEXN1YnNjcmlwdGlvbl92aWV3EpgCChBJbmdlc3RVc2'
    'FnZUV2ZW50EiMuYmlsbGluZy52MS5Jbmdlc3RVc2FnZUV2ZW50UmVxdWVzdBokLmJpbGxpbmcu'
    'djEuSW5nZXN0VXNhZ2VFdmVudFJlc3BvbnNlIrgBukeiAQoFVXNhZ2USFEluZ2VzdCBhIHVzYW'
    'dlIGV2ZW50GnFJbmdlc3RzIHVzYWdlIGV2ZW50cyBmb3IgbWV0ZXJpbmcuIEV2ZW50cyBhcmUg'
    'aWRlbXBvdGVudCBiYXNlZCBvbiBldmVudF9pZCAtIGR1cGxpY2F0ZSBldmVudHMgYXJlIHNhZm'
    'VseSBpZ25vcmVkLioQaW5nZXN0VXNhZ2VFdmVudIK1GA4KDHVzYWdlX2luZ2VzdBKiAgoRU2Vh'
    'cmNoVXNhZ2VFdmVudHMSGC5jb21tb24udjEuU2VhcmNoUmVxdWVzdBolLmJpbGxpbmcudjEuU2'
    'VhcmNoVXNhZ2VFdmVudHNSZXNwb25zZSLJAZACAbpHsgEKBVVzYWdlEhNTZWFyY2ggdXNhZ2Ug'
    'ZXZlbnRzGoABU2VhcmNoZXMgZm9yIHVzYWdlIGV2ZW50cyBtYXRjaGluZyBzcGVjaWZpZWQgY3'
    'JpdGVyaWEuIFN1cHBvcnRzIGZpbHRlcmluZyBieSBzdWJzY3JpcHRpb24sIGN1c3RvbWVyLCBt'
    'ZXRyaWMga2V5LCBhbmQgdGltZSByYW5nZS4qEXNlYXJjaFVzYWdlRXZlbnRzgrUYDAoKdXNhZ2'
    'VfdmlldzABErICCgpSdW5CaWxsaW5nEh0uYmlsbGluZy52MS5SdW5CaWxsaW5nUmVxdWVzdBoe'
    'LmJpbGxpbmcudjEuUnVuQmlsbGluZ1Jlc3BvbnNlIuQBukfIAQoHQmlsbGluZxITRXhlY3V0ZS'
    'BiaWxsaW5nIHJ1bhqbAUV4ZWN1dGVzIHRoZSBmdWxsIGJpbGxpbmcgcGlwZWxpbmUgZm9yIGEg'
    'c3Vic2NyaXB0aW9uOiBtZXRlcmluZywgcmF0aW5nLCBkaXNjb3VudGluZywgY3JlZGl0aW5nLC'
    'BpbnZvaWNpbmcuIElkZW1wb3RlbnQgcGVyIHN1YnNjcmlwdGlvbitwZXJpb2QgY29tYmluYXRp'
    'b24uKgpydW5CaWxsaW5ngrUYFAoSYmlsbGluZ19ydW5fbWFuYWdlEooCCg1HZXRCaWxsaW5nUn'
    'VuEiAuYmlsbGluZy52MS5HZXRCaWxsaW5nUnVuUmVxdWVzdBohLmJpbGxpbmcudjEuR2V0Qmls'
    'bGluZ1J1blJlc3BvbnNlIrMBkAIBukeWAQoHQmlsbGluZxIRR2V0IGEgYmlsbGluZyBydW4aaV'
    'JldHJpZXZlcyBhIGJpbGxpbmcgcnVuIGJ5IElEIGluY2x1ZGluZyBpdHMgY3VycmVudCBzdGF0'
    'ZSwgYXNzb2NpYXRlZCBpbnZvaWNlLCBhbmQgYW55IGVycm9yIGluZm9ybWF0aW9uLioNZ2V0Qm'
    'lsbGluZ1J1boK1GBIKEGJpbGxpbmdfcnVuX3ZpZXcS3wEKCkdldEludm9pY2USHS5iaWxsaW5n'
    'LnYxLkdldEludm9pY2VSZXF1ZXN0Gh4uYmlsbGluZy52MS5HZXRJbnZvaWNlUmVzcG9uc2UikQ'
    'GQAgG6R3kKCEludm9pY2VzEg5HZXQgYW4gaW52b2ljZRpRUmV0cmlldmVzIGFuIGludm9pY2Ug'
    'YnkgSUQgaW5jbHVkaW5nIGFsbCBsaW5lIGl0ZW1zLCBhbW91bnRzLCBhbmQgcGF5bWVudCBzdG'
    'F0dXMuKgpnZXRJbnZvaWNlgrUYDgoMaW52b2ljZV92aWV3EvkBCgxJc3N1ZUludm9pY2USHy5i'
    'aWxsaW5nLnYxLklzc3VlSW52b2ljZVJlcXVlc3QaIC5iaWxsaW5nLnYxLklzc3VlSW52b2ljZV'
    'Jlc3BvbnNlIqUBukeNAQoISW52b2ljZXMSEElzc3VlIGFuIGludm9pY2UaYUlzc3VlcyBhIGRy'
    'YWZ0IGludm9pY2UsIG1ha2luZyBpdCB2aXNpYmxlIHRvIHRoZSBjdXN0b21lci4gSW52b2ljZX'
    'MgYXJlIGltbXV0YWJsZSBhZnRlciBpc3N1YW5jZS4qDGlzc3VlSW52b2ljZYK1GBAKDmludm9p'
    'Y2VfbWFuYWdlEusBCgtWb2lkSW52b2ljZRIeLmJpbGxpbmcudjEuVm9pZEludm9pY2VSZXF1ZX'
    'N0Gh8uYmlsbGluZy52MS5Wb2lkSW52b2ljZVJlc3BvbnNlIpoBukeCAQoISW52b2ljZXMSD1Zv'
    'aWQgYW4gaW52b2ljZRpYVm9pZHMgYW4gaW52b2ljZSwgY2FuY2VsbGluZyBhbnkgb3V0c3Rhbm'
    'RpbmcgY2hhcmdlcy4gQ2Fubm90IHZvaWQgYWxyZWFkeS1wYWlkIGludm9pY2VzLioLdm9pZElu'
    'dm9pY2WCtRgQCg5pbnZvaWNlX21hbmFnZRKHAgoNUmVjb3JkUGF5bWVudBIgLmJpbGxpbmcudj'
    'EuUmVjb3JkUGF5bWVudFJlcXVlc3QaIS5iaWxsaW5nLnYxLlJlY29yZFBheW1lbnRSZXNwb25z'
    'ZSKwAbpHmAEKCEludm9pY2VzEhZSZWNvcmQgaW52b2ljZSBwYXltZW50GmVSZWNvcmRzIHBheW'
    '1lbnQgYWdhaW5zdCBhbiBpc3N1ZWQgaW52b2ljZS4gVXBkYXRlcyB0aGUgaW52b2ljZSBzdGF0'
    'ZSB0byBwYWlkIGFuZCBwb3N0cyB0byB0aGUgbGVkZ2VyLioNcmVjb3JkUGF5bWVudIK1GBAKDn'
    'BheW1lbnRfcmVjb3JkEpgCCg5TZWFyY2hJbnZvaWNlcxIYLmNvbW1vbi52MS5TZWFyY2hSZXF1'
    'ZXN0GiIuYmlsbGluZy52MS5TZWFyY2hJbnZvaWNlc1Jlc3BvbnNlIsUBkAIBukesAQoISW52b2'
    'ljZXMSD1NlYXJjaCBpbnZvaWNlcxp/U2VhcmNoZXMgZm9yIGludm9pY2VzIG1hdGNoaW5nIHNw'
    'ZWNpZmllZCBjcml0ZXJpYS4gU3VwcG9ydHMgZmlsdGVyaW5nIGJ5IGN1c3RvbWVyLCBzdWJzY3'
    'JpcHRpb24sIHN0YXRlLCBkYXRlIHJhbmdlLCBhbmQgYW1vdW50LioOc2VhcmNoSW52b2ljZXOC'
    'tRgOCgxpbnZvaWNlX3ZpZXcwARKdAgoLR3JhbnRDcmVkaXQSHi5iaWxsaW5nLnYxLkdyYW50Q3'
    'JlZGl0UmVxdWVzdBofLmJpbGxpbmcudjEuR3JhbnRDcmVkaXRSZXNwb25zZSLMAbpHtQEKB0Ny'
    'ZWRpdHMSFEdyYW50IHByZXBhaWQgY3JlZGl0GoYBQ3JlYXRlcyBhIHByZXBhaWQgY3JlZGl0IG'
    'dyYW50IGZvciBhIGN1c3RvbWVyLiBDcmVkaXRzIGFyZSBhdXRvbWF0aWNhbGx5IGFwcGxpZWQg'
    'ZHVyaW5nIGJpbGxpbmcgcnVucyBiYXNlZCBvbiBwcmlvcml0eSBhbmQgZXhwaXJhdGlvbi4qC2'
    'dyYW50Q3JlZGl0grUYDwoNY3JlZGl0X21hbmFnZRL8AQoQR2V0Q3JlZGl0QmFsYW5jZRIjLmJp'
    'bGxpbmcudjEuR2V0Q3JlZGl0QmFsYW5jZVJlcXVlc3QaJC5iaWxsaW5nLnYxLkdldENyZWRpdE'
    'JhbGFuY2VSZXNwb25zZSKcAZACAbpHhAEKB0NyZWRpdHMSEkdldCBjcmVkaXQgYmFsYW5jZRpT'
    'UmV0cmlldmVzIHRoZSB0b3RhbCByZW1haW5pbmcgY3JlZGl0IGJhbGFuY2UgZm9yIGEgY3VzdG'
    '9tZXIgaW4gYSBzcGVjaWZpYyBjdXJyZW5jeS4qEGdldENyZWRpdEJhbGFuY2WCtRgNCgtjcmVk'
    'aXRfdmlldxKzAgoOQ3JlYXRlRGlzY291bnQSIS5iaWxsaW5nLnYxLkNyZWF0ZURpc2NvdW50Um'
    'VxdWVzdBoiLmJpbGxpbmcudjEuQ3JlYXRlRGlzY291bnRSZXNwb25zZSLZAbpHwAEKCURpc2Nv'
    'dW50cxIRQ3JlYXRlIGEgZGlzY291bnQajwFDcmVhdGVzIGEgbmV3IGRpc2NvdW50IHJ1bGUgdG'
    'hhdCBjYW4gYmUgYXBwbGllZCBkdXJpbmcgYmlsbGluZyBydW5zLiBTdXBwb3J0cyBwZXJjZW50'
    'YWdlIGFuZCBmaXhlZCBhbW91bnQgZGlzY291bnRzIHdpdGggdGltZS1ib3VuZGVkIHZhbGlkaX'
    'R5LioOY3JlYXRlRGlzY291bnSCtRgRCg9kaXNjb3VudF9tYW5hZ2US1gEKD1NlYXJjaERpc2Nv'
    'dW50cxIYLmNvbW1vbi52MS5TZWFyY2hSZXF1ZXN0GiMuYmlsbGluZy52MS5TZWFyY2hEaXNjb3'
    'VudHNSZXNwb25zZSKBAZACAbpHaAoJRGlzY291bnRzEhBTZWFyY2ggZGlzY291bnRzGjhTZWFy'
    'Y2hlcyBmb3IgZGlzY291bnQgcnVsZXMgbWF0Y2hpbmcgc3BlY2lmaWVkIGNyaXRlcmlhLioPc2'
    'VhcmNoRGlzY291bnRzgrUYDwoNZGlzY291bnRfdmlldzABGrALgrUYqwsKD3NlcnZpY2VfYmls'
    'bGluZxIMY2F0YWxvZ192aWV3Eg5jYXRhbG9nX21hbmFnZRILcGxhbl9tYW5hZ2USEGNvbXBvbm'
    'VudF9tYW5hZ2USC3RpZXJfbWFuYWdlEhFzdWJzY3JpcHRpb25fdmlldxITc3Vic2NyaXB0aW9u'
    'X21hbmFnZRIKdXNhZ2VfdmlldxIMdXNhZ2VfaW5nZXN0EhBiaWxsaW5nX3J1bl92aWV3EhJiaW'
    'xsaW5nX3J1bl9tYW5hZ2USDGludm9pY2VfdmlldxIOaW52b2ljZV9tYW5hZ2USDnBheW1lbnRf'
    'cmVjb3JkEgtjcmVkaXRfdmlldxINY3JlZGl0X21hbmFnZRINZGlzY291bnRfdmlldxIPZGlzY2'
    '91bnRfbWFuYWdlGp4CCAESDGNhdGFsb2dfdmlldxIOY2F0YWxvZ19tYW5hZ2USC3BsYW5fbWFu'
    'YWdlEhBjb21wb25lbnRfbWFuYWdlEgt0aWVyX21hbmFnZRIRc3Vic2NyaXB0aW9uX3ZpZXcSE3'
    'N1YnNjcmlwdGlvbl9tYW5hZ2USCnVzYWdlX3ZpZXcSDHVzYWdlX2luZ2VzdBIQYmlsbGluZ19y'
    'dW5fdmlldxISYmlsbGluZ19ydW5fbWFuYWdlEgxpbnZvaWNlX3ZpZXcSDmludm9pY2VfbWFuYW'
    'dlEg5wYXltZW50X3JlY29yZBILY3JlZGl0X3ZpZXcSDWNyZWRpdF9tYW5hZ2USDWRpc2NvdW50'
    'X3ZpZXcSD2Rpc2NvdW50X21hbmFnZRqeAggCEgxjYXRhbG9nX3ZpZXcSDmNhdGFsb2dfbWFuYW'
    'dlEgtwbGFuX21hbmFnZRIQY29tcG9uZW50X21hbmFnZRILdGllcl9tYW5hZ2USEXN1YnNjcmlw'
    'dGlvbl92aWV3EhNzdWJzY3JpcHRpb25fbWFuYWdlEgp1c2FnZV92aWV3Egx1c2FnZV9pbmdlc3'
    'QSEGJpbGxpbmdfcnVuX3ZpZXcSEmJpbGxpbmdfcnVuX21hbmFnZRIMaW52b2ljZV92aWV3Eg5p'
    'bnZvaWNlX21hbmFnZRIOcGF5bWVudF9yZWNvcmQSC2NyZWRpdF92aWV3Eg1jcmVkaXRfbWFuYW'
    'dlEg1kaXNjb3VudF92aWV3Eg9kaXNjb3VudF9tYW5hZ2UaeQgDEgxjYXRhbG9nX3ZpZXcSEXN1'
    'YnNjcmlwdGlvbl92aWV3Egp1c2FnZV92aWV3Egx1c2FnZV9pbmdlc3QSEGJpbGxpbmdfcnVuX3'
    'ZpZXcSDGludm9pY2VfdmlldxILY3JlZGl0X3ZpZXcSDWRpc2NvdW50X3ZpZXcaawgEEgxjYXRh'
    'bG9nX3ZpZXcSEXN1YnNjcmlwdGlvbl92aWV3Egp1c2FnZV92aWV3EhBiaWxsaW5nX3J1bl92aW'
    'V3EgxpbnZvaWNlX3ZpZXcSC2NyZWRpdF92aWV3Eg1kaXNjb3VudF92aWV3GjEIBRIMY2F0YWxv'
    'Z192aWV3EhFzdWJzY3JpcHRpb25fdmlldxIMaW52b2ljZV92aWV3Gp4CCAYSDGNhdGFsb2dfdm'
    'lldxIOY2F0YWxvZ19tYW5hZ2USC3BsYW5fbWFuYWdlEhBjb21wb25lbnRfbWFuYWdlEgt0aWVy'
    'X21hbmFnZRIRc3Vic2NyaXB0aW9uX3ZpZXcSE3N1YnNjcmlwdGlvbl9tYW5hZ2USCnVzYWdlX3'
    'ZpZXcSDHVzYWdlX2luZ2VzdBIQYmlsbGluZ19ydW5fdmlldxISYmlsbGluZ19ydW5fbWFuYWdl'
    'EgxpbnZvaWNlX3ZpZXcSDmludm9pY2VfbWFuYWdlEg5wYXltZW50X3JlY29yZBILY3JlZGl0X3'
    'ZpZXcSDWNyZWRpdF9tYW5hZ2USDWRpc2NvdW50X3ZpZXcSD2Rpc2NvdW50X21hbmFnZQ==');

