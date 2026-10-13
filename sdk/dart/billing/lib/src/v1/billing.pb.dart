// This is a generated file - do not edit.
//
// Generated from v1/billing.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart' as $1;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

import '../common/v1/common.pb.dart' as $4;
import '../common/v1/money.pb.dart' as $2;
import '../google/type/interval.pb.dart' as $3;
import 'billing.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'billing.pbenum.dart';

/// CatalogVersion represents an immutable, versioned catalog of plans and pricing.
class CatalogVersion extends $pb.GeneratedMessage {
  factory CatalogVersion({
    $core.String? id,
    $core.String? catalogId,
    $core.int? version,
    $core.String? name,
    $core.String? currency,
    $0.Timestamp? publishedAt,
    $0.Timestamp? effectiveAt,
    $0.Timestamp? retiredAt,
    $1.Struct? data,
    $core.Iterable<Plan>? plans,
  }) {
    final result = CatalogVersion._();
    if (id != null) result.id = id;
    if (catalogId != null) result.catalogId = catalogId;
    if (version != null) result.version = version;
    if (name != null) result.name = name;
    if (currency != null) result.currency = currency;
    if (publishedAt != null) result.publishedAt = publishedAt;
    if (effectiveAt != null) result.effectiveAt = effectiveAt;
    if (retiredAt != null) result.retiredAt = retiredAt;
    if (data != null) result.data = data;
    if (plans != null) result.plans.addAll(plans);
    return result;
  }

  CatalogVersion._();

  factory CatalogVersion.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CatalogVersion()..mergeFromBuffer(data, registry);
  factory CatalogVersion.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CatalogVersion()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CatalogVersion',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CatalogVersion.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'catalogId')
    ..aI(3, _omitFieldNames ? '' : 'version')
    ..aOS(4, _omitFieldNames ? '' : 'name')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'publishedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'effectiveAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'retiredAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$1.Struct>(9, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..pPM<Plan>(10, _omitFieldNames ? '' : 'plans',
        subBuilder: Plan.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CatalogVersion clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CatalogVersion copyWith(void Function(CatalogVersion) updates) =>
      super.copyWith((message) => updates(message as CatalogVersion))
          as CatalogVersion;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CatalogVersion() / CatalogVersion.new instead')
  static CatalogVersion create() => CatalogVersion._();
  static $pb.GeneratedMessage $_createMessage() => CatalogVersion._();
  @$core.override
  CatalogVersion createEmptyInstance() => CatalogVersion._();
  @$core.pragma('dart2js:noInline')
  static CatalogVersion getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CatalogVersion>(
          CatalogVersion.$_createMessage);
  static CatalogVersion? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get catalogId => $_getSZ(1);
  @$pb.TagNumber(2)
  set catalogId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCatalogId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCatalogId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get version => $_getIZ(2);
  @$pb.TagNumber(3)
  set version($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearVersion() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get name => $_getSZ(3);
  @$pb.TagNumber(4)
  set name($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Timestamp get publishedAt => $_getN(5);
  @$pb.TagNumber(6)
  set publishedAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasPublishedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearPublishedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensurePublishedAt() => $_ensure(5);

  @$pb.TagNumber(7)
  $0.Timestamp get effectiveAt => $_getN(6);
  @$pb.TagNumber(7)
  set effectiveAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasEffectiveAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearEffectiveAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureEffectiveAt() => $_ensure(6);

  @$pb.TagNumber(8)
  $0.Timestamp get retiredAt => $_getN(7);
  @$pb.TagNumber(8)
  set retiredAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasRetiredAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearRetiredAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureRetiredAt() => $_ensure(7);

  @$pb.TagNumber(9)
  $1.Struct get data => $_getN(8);
  @$pb.TagNumber(9)
  set data($1.Struct value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasData() => $_has(8);
  @$pb.TagNumber(9)
  void clearData() => $_clearField(9);
  @$pb.TagNumber(9)
  $1.Struct ensureData() => $_ensure(8);

  @$pb.TagNumber(10)
  $pb.PbList<Plan> get plans => $_getList(9);
}

/// Plan represents a billing plan within a catalog version.
class Plan extends $pb.GeneratedMessage {
  factory Plan({
    $core.String? id,
    $core.String? catalogVersionId,
    $core.String? externalId,
    $core.String? name,
    $core.String? description,
    $1.Struct? data,
    $core.Iterable<Component>? components,
  }) {
    final result = Plan._();
    if (id != null) result.id = id;
    if (catalogVersionId != null) result.catalogVersionId = catalogVersionId;
    if (externalId != null) result.externalId = externalId;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (data != null) result.data = data;
    if (components != null) result.components.addAll(components);
    return result;
  }

  Plan._();

  factory Plan.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Plan()..mergeFromBuffer(data, registry);
  factory Plan.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Plan()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Plan',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: Plan.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'catalogVersionId')
    ..aOS(3, _omitFieldNames ? '' : 'externalId')
    ..aOS(4, _omitFieldNames ? '' : 'name')
    ..aOS(5, _omitFieldNames ? '' : 'description')
    ..aOM<$1.Struct>(6, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..pPM<Component>(7, _omitFieldNames ? '' : 'components',
        subBuilder: Component.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Plan clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Plan copyWith(void Function(Plan) updates) =>
      super.copyWith((message) => updates(message as Plan)) as Plan;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Plan() / Plan.new instead')
  static Plan create() => Plan._();
  static $pb.GeneratedMessage $_createMessage() => Plan._();
  @$core.override
  Plan createEmptyInstance() => Plan._();
  @$core.pragma('dart2js:noInline')
  static Plan getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Plan>(Plan.$_createMessage);
  static Plan? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get catalogVersionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set catalogVersionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCatalogVersionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCatalogVersionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get externalId => $_getSZ(2);
  @$pb.TagNumber(3)
  set externalId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExternalId() => $_has(2);
  @$pb.TagNumber(3)
  void clearExternalId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get name => $_getSZ(3);
  @$pb.TagNumber(4)
  set name($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get description => $_getSZ(4);
  @$pb.TagNumber(5)
  set description($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDescription() => $_has(4);
  @$pb.TagNumber(5)
  void clearDescription() => $_clearField(5);

  @$pb.TagNumber(6)
  $1.Struct get data => $_getN(5);
  @$pb.TagNumber(6)
  set data($1.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasData() => $_has(5);
  @$pb.TagNumber(6)
  void clearData() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Struct ensureData() => $_ensure(5);

  @$pb.TagNumber(7)
  $pb.PbList<Component> get components => $_getList(6);
}

/// Component represents a billable component within a plan.
class Component extends $pb.GeneratedMessage {
  factory Component({
    $core.String? id,
    $core.String? planId,
    $core.String? externalId,
    $core.String? name,
    $core.String? metricKey,
    PricingModel? pricingModel,
    AggregationType? aggregationType,
    $core.String? unitName,
    $2.Money? freeQuantity,
    $2.Money? minimumCharge,
    $1.Struct? data,
    $core.Iterable<Tier>? tiers,
  }) {
    final result = Component._();
    if (id != null) result.id = id;
    if (planId != null) result.planId = planId;
    if (externalId != null) result.externalId = externalId;
    if (name != null) result.name = name;
    if (metricKey != null) result.metricKey = metricKey;
    if (pricingModel != null) result.pricingModel = pricingModel;
    if (aggregationType != null) result.aggregationType = aggregationType;
    if (unitName != null) result.unitName = unitName;
    if (freeQuantity != null) result.freeQuantity = freeQuantity;
    if (minimumCharge != null) result.minimumCharge = minimumCharge;
    if (data != null) result.data = data;
    if (tiers != null) result.tiers.addAll(tiers);
    return result;
  }

  Component._();

  factory Component.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Component()..mergeFromBuffer(data, registry);
  factory Component.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Component()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Component',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: Component.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'planId')
    ..aOS(3, _omitFieldNames ? '' : 'externalId')
    ..aOS(4, _omitFieldNames ? '' : 'name')
    ..aOS(5, _omitFieldNames ? '' : 'metricKey')
    ..aE<PricingModel>(6, _omitFieldNames ? '' : 'pricingModel',
        enumValues: PricingModel.values)
    ..aE<AggregationType>(7, _omitFieldNames ? '' : 'aggregationType',
        enumValues: AggregationType.values)
    ..aOS(8, _omitFieldNames ? '' : 'unitName')
    ..aOM<$2.Money>(9, _omitFieldNames ? '' : 'freeQuantity',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(10, _omitFieldNames ? '' : 'minimumCharge',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$1.Struct>(11, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..pPM<Tier>(12, _omitFieldNames ? '' : 'tiers',
        subBuilder: Tier.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Component clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Component copyWith(void Function(Component) updates) =>
      super.copyWith((message) => updates(message as Component)) as Component;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Component() / Component.new instead')
  static Component create() => Component._();
  static $pb.GeneratedMessage $_createMessage() => Component._();
  @$core.override
  Component createEmptyInstance() => Component._();
  @$core.pragma('dart2js:noInline')
  static Component getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Component>(Component.$_createMessage);
  static Component? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get planId => $_getSZ(1);
  @$pb.TagNumber(2)
  set planId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPlanId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlanId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get externalId => $_getSZ(2);
  @$pb.TagNumber(3)
  set externalId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExternalId() => $_has(2);
  @$pb.TagNumber(3)
  void clearExternalId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get name => $_getSZ(3);
  @$pb.TagNumber(4)
  set name($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get metricKey => $_getSZ(4);
  @$pb.TagNumber(5)
  set metricKey($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasMetricKey() => $_has(4);
  @$pb.TagNumber(5)
  void clearMetricKey() => $_clearField(5);

  @$pb.TagNumber(6)
  PricingModel get pricingModel => $_getN(5);
  @$pb.TagNumber(6)
  set pricingModel(PricingModel value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasPricingModel() => $_has(5);
  @$pb.TagNumber(6)
  void clearPricingModel() => $_clearField(6);

  @$pb.TagNumber(7)
  AggregationType get aggregationType => $_getN(6);
  @$pb.TagNumber(7)
  set aggregationType(AggregationType value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasAggregationType() => $_has(6);
  @$pb.TagNumber(7)
  void clearAggregationType() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get unitName => $_getSZ(7);
  @$pb.TagNumber(8)
  set unitName($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasUnitName() => $_has(7);
  @$pb.TagNumber(8)
  void clearUnitName() => $_clearField(8);

  @$pb.TagNumber(9)
  $2.Money get freeQuantity => $_getN(8);
  @$pb.TagNumber(9)
  set freeQuantity($2.Money value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasFreeQuantity() => $_has(8);
  @$pb.TagNumber(9)
  void clearFreeQuantity() => $_clearField(9);
  @$pb.TagNumber(9)
  $2.Money ensureFreeQuantity() => $_ensure(8);

  @$pb.TagNumber(10)
  $2.Money get minimumCharge => $_getN(9);
  @$pb.TagNumber(10)
  set minimumCharge($2.Money value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasMinimumCharge() => $_has(9);
  @$pb.TagNumber(10)
  void clearMinimumCharge() => $_clearField(10);
  @$pb.TagNumber(10)
  $2.Money ensureMinimumCharge() => $_ensure(9);

  @$pb.TagNumber(11)
  $1.Struct get data => $_getN(10);
  @$pb.TagNumber(11)
  set data($1.Struct value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasData() => $_has(10);
  @$pb.TagNumber(11)
  void clearData() => $_clearField(11);
  @$pb.TagNumber(11)
  $1.Struct ensureData() => $_ensure(10);

  @$pb.TagNumber(12)
  $pb.PbList<Tier> get tiers => $_getList(11);
}

/// Tier represents a pricing tier within a component.
class Tier extends $pb.GeneratedMessage {
  factory Tier({
    $core.String? id,
    $core.String? componentId,
    $core.int? sortOrder,
    $2.Money? lowerBound,
    $2.Money? upperBound,
    $2.Money? unitPrice,
    $2.Money? flatFee,
  }) {
    final result = Tier._();
    if (id != null) result.id = id;
    if (componentId != null) result.componentId = componentId;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (lowerBound != null) result.lowerBound = lowerBound;
    if (upperBound != null) result.upperBound = upperBound;
    if (unitPrice != null) result.unitPrice = unitPrice;
    if (flatFee != null) result.flatFee = flatFee;
    return result;
  }

  Tier._();

  factory Tier.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Tier()..mergeFromBuffer(data, registry);
  factory Tier.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Tier()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Tier',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: Tier.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'componentId')
    ..aI(3, _omitFieldNames ? '' : 'sortOrder')
    ..aOM<$2.Money>(4, _omitFieldNames ? '' : 'lowerBound',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(5, _omitFieldNames ? '' : 'upperBound',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(6, _omitFieldNames ? '' : 'unitPrice',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(7, _omitFieldNames ? '' : 'flatFee',
        subBuilder: $2.Money.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Tier clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Tier copyWith(void Function(Tier) updates) =>
      super.copyWith((message) => updates(message as Tier)) as Tier;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Tier() / Tier.new instead')
  static Tier create() => Tier._();
  static $pb.GeneratedMessage $_createMessage() => Tier._();
  @$core.override
  Tier createEmptyInstance() => Tier._();
  @$core.pragma('dart2js:noInline')
  static Tier getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Tier>(Tier.$_createMessage);
  static Tier? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get componentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set componentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasComponentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearComponentId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get sortOrder => $_getIZ(2);
  @$pb.TagNumber(3)
  set sortOrder($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSortOrder() => $_has(2);
  @$pb.TagNumber(3)
  void clearSortOrder() => $_clearField(3);

  @$pb.TagNumber(4)
  $2.Money get lowerBound => $_getN(3);
  @$pb.TagNumber(4)
  set lowerBound($2.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasLowerBound() => $_has(3);
  @$pb.TagNumber(4)
  void clearLowerBound() => $_clearField(4);
  @$pb.TagNumber(4)
  $2.Money ensureLowerBound() => $_ensure(3);

  @$pb.TagNumber(5)
  $2.Money get upperBound => $_getN(4);
  @$pb.TagNumber(5)
  set upperBound($2.Money value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasUpperBound() => $_has(4);
  @$pb.TagNumber(5)
  void clearUpperBound() => $_clearField(5);
  @$pb.TagNumber(5)
  $2.Money ensureUpperBound() => $_ensure(4);

  @$pb.TagNumber(6)
  $2.Money get unitPrice => $_getN(5);
  @$pb.TagNumber(6)
  set unitPrice($2.Money value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUnitPrice() => $_has(5);
  @$pb.TagNumber(6)
  void clearUnitPrice() => $_clearField(6);
  @$pb.TagNumber(6)
  $2.Money ensureUnitPrice() => $_ensure(5);

  @$pb.TagNumber(7)
  $2.Money get flatFee => $_getN(6);
  @$pb.TagNumber(7)
  set flatFee($2.Money value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasFlatFee() => $_has(6);
  @$pb.TagNumber(7)
  void clearFlatFee() => $_clearField(7);
  @$pb.TagNumber(7)
  $2.Money ensureFlatFee() => $_ensure(6);
}

/// Subscription represents a customer's subscription to a plan.
class Subscription extends $pb.GeneratedMessage {
  factory Subscription({
    $core.String? id,
    $core.String? profileId,
    $core.String? catalogVersionId,
    $core.String? planId,
    SubscriptionState? state,
    $0.Timestamp? startAt,
    $0.Timestamp? endAt,
    $0.Timestamp? cancelledAt,
    $0.Timestamp? billingAnchor,
    $core.String? currency,
    $1.Struct? data,
  }) {
    final result = Subscription._();
    if (id != null) result.id = id;
    if (profileId != null) result.profileId = profileId;
    if (catalogVersionId != null) result.catalogVersionId = catalogVersionId;
    if (planId != null) result.planId = planId;
    if (state != null) result.state = state;
    if (startAt != null) result.startAt = startAt;
    if (endAt != null) result.endAt = endAt;
    if (cancelledAt != null) result.cancelledAt = cancelledAt;
    if (billingAnchor != null) result.billingAnchor = billingAnchor;
    if (currency != null) result.currency = currency;
    if (data != null) result.data = data;
    return result;
  }

  Subscription._();

  factory Subscription.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Subscription()..mergeFromBuffer(data, registry);
  factory Subscription.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Subscription()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Subscription',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: Subscription.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'profileId')
    ..aOS(3, _omitFieldNames ? '' : 'catalogVersionId')
    ..aOS(4, _omitFieldNames ? '' : 'planId')
    ..aE<SubscriptionState>(5, _omitFieldNames ? '' : 'state',
        enumValues: SubscriptionState.values)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'startAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'endAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'cancelledAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(9, _omitFieldNames ? '' : 'billingAnchor',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(10, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Struct>(11, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Subscription clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Subscription copyWith(void Function(Subscription) updates) =>
      super.copyWith((message) => updates(message as Subscription))
          as Subscription;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Subscription() / Subscription.new instead')
  static Subscription create() => Subscription._();
  static $pb.GeneratedMessage $_createMessage() => Subscription._();
  @$core.override
  Subscription createEmptyInstance() => Subscription._();
  @$core.pragma('dart2js:noInline')
  static Subscription getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Subscription>(
          Subscription.$_createMessage);
  static Subscription? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get profileId => $_getSZ(1);
  @$pb.TagNumber(2)
  set profileId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProfileId() => $_has(1);
  @$pb.TagNumber(2)
  void clearProfileId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get catalogVersionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set catalogVersionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCatalogVersionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCatalogVersionId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get planId => $_getSZ(3);
  @$pb.TagNumber(4)
  set planId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPlanId() => $_has(3);
  @$pb.TagNumber(4)
  void clearPlanId() => $_clearField(4);

  @$pb.TagNumber(5)
  SubscriptionState get state => $_getN(4);
  @$pb.TagNumber(5)
  set state(SubscriptionState value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasState() => $_has(4);
  @$pb.TagNumber(5)
  void clearState() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Timestamp get startAt => $_getN(5);
  @$pb.TagNumber(6)
  set startAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasStartAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearStartAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureStartAt() => $_ensure(5);

  @$pb.TagNumber(7)
  $0.Timestamp get endAt => $_getN(6);
  @$pb.TagNumber(7)
  set endAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasEndAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearEndAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureEndAt() => $_ensure(6);

  @$pb.TagNumber(8)
  $0.Timestamp get cancelledAt => $_getN(7);
  @$pb.TagNumber(8)
  set cancelledAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasCancelledAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearCancelledAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureCancelledAt() => $_ensure(7);

  @$pb.TagNumber(9)
  $0.Timestamp get billingAnchor => $_getN(8);
  @$pb.TagNumber(9)
  set billingAnchor($0.Timestamp value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasBillingAnchor() => $_has(8);
  @$pb.TagNumber(9)
  void clearBillingAnchor() => $_clearField(9);
  @$pb.TagNumber(9)
  $0.Timestamp ensureBillingAnchor() => $_ensure(8);

  @$pb.TagNumber(10)
  $core.String get currency => $_getSZ(9);
  @$pb.TagNumber(10)
  set currency($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasCurrency() => $_has(9);
  @$pb.TagNumber(10)
  void clearCurrency() => $_clearField(10);

  @$pb.TagNumber(11)
  $1.Struct get data => $_getN(10);
  @$pb.TagNumber(11)
  set data($1.Struct value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasData() => $_has(10);
  @$pb.TagNumber(11)
  void clearData() => $_clearField(11);
  @$pb.TagNumber(11)
  $1.Struct ensureData() => $_ensure(10);
}

/// UsageEvent represents a usage event for metering.
/// Fields 1-8 support simple metering (metric + quantity).
/// Fields 9-14 support resource spend context (resource identity, intervals, pre-calculated cost).
class UsageEvent extends $pb.GeneratedMessage {
  factory UsageEvent({
    $core.String? id,
    $core.String? description,
    $core.String? subscriptionId,
    $core.String? metricKey,
    $core.double? quantity,
    $0.Timestamp? timestamp,
    $1.Struct? properties,
    $3.Interval? interval,
    $core.String? unit,
    $2.Money? amount,
    $core.String? groupId,
  }) {
    final result = UsageEvent._();
    if (id != null) result.id = id;
    if (description != null) result.description = description;
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (metricKey != null) result.metricKey = metricKey;
    if (quantity != null) result.quantity = quantity;
    if (timestamp != null) result.timestamp = timestamp;
    if (properties != null) result.properties = properties;
    if (interval != null) result.interval = interval;
    if (unit != null) result.unit = unit;
    if (amount != null) result.amount = amount;
    if (groupId != null) result.groupId = groupId;
    return result;
  }

  UsageEvent._();

  factory UsageEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UsageEvent()..mergeFromBuffer(data, registry);
  factory UsageEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UsageEvent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UsageEvent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: UsageEvent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..aOS(3, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(5, _omitFieldNames ? '' : 'metricKey')
    ..aD(6, _omitFieldNames ? '' : 'quantity')
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'timestamp',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$1.Struct>(8, _omitFieldNames ? '' : 'properties',
        subBuilder: $1.Struct.$_createMessage)
    ..aOM<$3.Interval>(11, _omitFieldNames ? '' : 'interval',
        subBuilder: $3.Interval.$_createMessage)
    ..aOS(12, _omitFieldNames ? '' : 'unit')
    ..aOM<$2.Money>(13, _omitFieldNames ? '' : 'amount',
        subBuilder: $2.Money.$_createMessage)
    ..aOS(14, _omitFieldNames ? '' : 'groupId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UsageEvent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UsageEvent copyWith(void Function(UsageEvent) updates) =>
      super.copyWith((message) => updates(message as UsageEvent)) as UsageEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UsageEvent() / UsageEvent.new instead')
  static UsageEvent create() => UsageEvent._();
  static $pb.GeneratedMessage $_createMessage() => UsageEvent._();
  @$core.override
  UsageEvent createEmptyInstance() => UsageEvent._();
  @$core.pragma('dart2js:noInline')
  static UsageEvent getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UsageEvent>(UsageEvent.$_createMessage);
  static UsageEvent? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get subscriptionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set subscriptionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSubscriptionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSubscriptionId() => $_clearField(3);

  @$pb.TagNumber(5)
  $core.String get metricKey => $_getSZ(3);
  @$pb.TagNumber(5)
  set metricKey($core.String value) => $_setString(3, value);
  @$pb.TagNumber(5)
  $core.bool hasMetricKey() => $_has(3);
  @$pb.TagNumber(5)
  void clearMetricKey() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.double get quantity => $_getN(4);
  @$pb.TagNumber(6)
  set quantity($core.double value) => $_setDouble(4, value);
  @$pb.TagNumber(6)
  $core.bool hasQuantity() => $_has(4);
  @$pb.TagNumber(6)
  void clearQuantity() => $_clearField(6);

  @$pb.TagNumber(7)
  $0.Timestamp get timestamp => $_getN(5);
  @$pb.TagNumber(7)
  set timestamp($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasTimestamp() => $_has(5);
  @$pb.TagNumber(7)
  void clearTimestamp() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureTimestamp() => $_ensure(5);

  @$pb.TagNumber(8)
  $1.Struct get properties => $_getN(6);
  @$pb.TagNumber(8)
  set properties($1.Struct value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasProperties() => $_has(6);
  @$pb.TagNumber(8)
  void clearProperties() => $_clearField(8);
  @$pb.TagNumber(8)
  $1.Struct ensureProperties() => $_ensure(6);

  @$pb.TagNumber(11)
  $3.Interval get interval => $_getN(7);
  @$pb.TagNumber(11)
  set interval($3.Interval value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasInterval() => $_has(7);
  @$pb.TagNumber(11)
  void clearInterval() => $_clearField(11);
  @$pb.TagNumber(11)
  $3.Interval ensureInterval() => $_ensure(7);

  @$pb.TagNumber(12)
  $core.String get unit => $_getSZ(8);
  @$pb.TagNumber(12)
  set unit($core.String value) => $_setString(8, value);
  @$pb.TagNumber(12)
  $core.bool hasUnit() => $_has(8);
  @$pb.TagNumber(12)
  void clearUnit() => $_clearField(12);

  @$pb.TagNumber(13)
  $2.Money get amount => $_getN(9);
  @$pb.TagNumber(13)
  set amount($2.Money value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasAmount() => $_has(9);
  @$pb.TagNumber(13)
  void clearAmount() => $_clearField(13);
  @$pb.TagNumber(13)
  $2.Money ensureAmount() => $_ensure(9);

  @$pb.TagNumber(14)
  $core.String get groupId => $_getSZ(10);
  @$pb.TagNumber(14)
  set groupId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(14)
  $core.bool hasGroupId() => $_has(10);
  @$pb.TagNumber(14)
  void clearGroupId() => $_clearField(14);
}

/// Invoice represents a billing invoice.
class Invoice extends $pb.GeneratedMessage {
  factory Invoice({
    $core.String? id,
    $core.String? billingRunId,
    $core.String? profileId,
    $core.String? subscriptionId,
    $core.String? invoiceNumber,
    InvoiceState? state,
    $core.String? currency,
    $2.Money? subtotalAmount,
    $2.Money? discountAmount,
    $2.Money? creditAmount,
    $2.Money? totalAmount,
    $0.Timestamp? periodStart,
    $0.Timestamp? periodEnd,
    $0.Timestamp? issuedAt,
    $0.Timestamp? dueAt,
    $0.Timestamp? paidAt,
    $core.String? ledgerTxnId,
    $1.Struct? data,
    $core.Iterable<InvoiceLine>? lines,
  }) {
    final result = Invoice._();
    if (id != null) result.id = id;
    if (billingRunId != null) result.billingRunId = billingRunId;
    if (profileId != null) result.profileId = profileId;
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (invoiceNumber != null) result.invoiceNumber = invoiceNumber;
    if (state != null) result.state = state;
    if (currency != null) result.currency = currency;
    if (subtotalAmount != null) result.subtotalAmount = subtotalAmount;
    if (discountAmount != null) result.discountAmount = discountAmount;
    if (creditAmount != null) result.creditAmount = creditAmount;
    if (totalAmount != null) result.totalAmount = totalAmount;
    if (periodStart != null) result.periodStart = periodStart;
    if (periodEnd != null) result.periodEnd = periodEnd;
    if (issuedAt != null) result.issuedAt = issuedAt;
    if (dueAt != null) result.dueAt = dueAt;
    if (paidAt != null) result.paidAt = paidAt;
    if (ledgerTxnId != null) result.ledgerTxnId = ledgerTxnId;
    if (data != null) result.data = data;
    if (lines != null) result.lines.addAll(lines);
    return result;
  }

  Invoice._();

  factory Invoice.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Invoice()..mergeFromBuffer(data, registry);
  factory Invoice.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Invoice()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Invoice',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: Invoice.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'billingRunId')
    ..aOS(3, _omitFieldNames ? '' : 'profileId')
    ..aOS(4, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(5, _omitFieldNames ? '' : 'invoiceNumber')
    ..aE<InvoiceState>(6, _omitFieldNames ? '' : 'state',
        enumValues: InvoiceState.values)
    ..aOS(7, _omitFieldNames ? '' : 'currency')
    ..aOM<$2.Money>(8, _omitFieldNames ? '' : 'subtotalAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(9, _omitFieldNames ? '' : 'discountAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(10, _omitFieldNames ? '' : 'creditAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(11, _omitFieldNames ? '' : 'totalAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$0.Timestamp>(12, _omitFieldNames ? '' : 'periodStart',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(13, _omitFieldNames ? '' : 'periodEnd',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(14, _omitFieldNames ? '' : 'issuedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(15, _omitFieldNames ? '' : 'dueAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(16, _omitFieldNames ? '' : 'paidAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(17, _omitFieldNames ? '' : 'ledgerTxnId')
    ..aOM<$1.Struct>(18, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..pPM<InvoiceLine>(19, _omitFieldNames ? '' : 'lines',
        subBuilder: InvoiceLine.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Invoice clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Invoice copyWith(void Function(Invoice) updates) =>
      super.copyWith((message) => updates(message as Invoice)) as Invoice;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Invoice() / Invoice.new instead')
  static Invoice create() => Invoice._();
  static $pb.GeneratedMessage $_createMessage() => Invoice._();
  @$core.override
  Invoice createEmptyInstance() => Invoice._();
  @$core.pragma('dart2js:noInline')
  static Invoice getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Invoice>(Invoice.$_createMessage);
  static Invoice? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get billingRunId => $_getSZ(1);
  @$pb.TagNumber(2)
  set billingRunId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBillingRunId() => $_has(1);
  @$pb.TagNumber(2)
  void clearBillingRunId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get profileId => $_getSZ(2);
  @$pb.TagNumber(3)
  set profileId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProfileId() => $_has(2);
  @$pb.TagNumber(3)
  void clearProfileId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get subscriptionId => $_getSZ(3);
  @$pb.TagNumber(4)
  set subscriptionId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSubscriptionId() => $_has(3);
  @$pb.TagNumber(4)
  void clearSubscriptionId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get invoiceNumber => $_getSZ(4);
  @$pb.TagNumber(5)
  set invoiceNumber($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasInvoiceNumber() => $_has(4);
  @$pb.TagNumber(5)
  void clearInvoiceNumber() => $_clearField(5);

  @$pb.TagNumber(6)
  InvoiceState get state => $_getN(5);
  @$pb.TagNumber(6)
  set state(InvoiceState value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasState() => $_has(5);
  @$pb.TagNumber(6)
  void clearState() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get currency => $_getSZ(6);
  @$pb.TagNumber(7)
  set currency($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCurrency() => $_has(6);
  @$pb.TagNumber(7)
  void clearCurrency() => $_clearField(7);

  @$pb.TagNumber(8)
  $2.Money get subtotalAmount => $_getN(7);
  @$pb.TagNumber(8)
  set subtotalAmount($2.Money value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasSubtotalAmount() => $_has(7);
  @$pb.TagNumber(8)
  void clearSubtotalAmount() => $_clearField(8);
  @$pb.TagNumber(8)
  $2.Money ensureSubtotalAmount() => $_ensure(7);

  @$pb.TagNumber(9)
  $2.Money get discountAmount => $_getN(8);
  @$pb.TagNumber(9)
  set discountAmount($2.Money value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasDiscountAmount() => $_has(8);
  @$pb.TagNumber(9)
  void clearDiscountAmount() => $_clearField(9);
  @$pb.TagNumber(9)
  $2.Money ensureDiscountAmount() => $_ensure(8);

  @$pb.TagNumber(10)
  $2.Money get creditAmount => $_getN(9);
  @$pb.TagNumber(10)
  set creditAmount($2.Money value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasCreditAmount() => $_has(9);
  @$pb.TagNumber(10)
  void clearCreditAmount() => $_clearField(10);
  @$pb.TagNumber(10)
  $2.Money ensureCreditAmount() => $_ensure(9);

  @$pb.TagNumber(11)
  $2.Money get totalAmount => $_getN(10);
  @$pb.TagNumber(11)
  set totalAmount($2.Money value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasTotalAmount() => $_has(10);
  @$pb.TagNumber(11)
  void clearTotalAmount() => $_clearField(11);
  @$pb.TagNumber(11)
  $2.Money ensureTotalAmount() => $_ensure(10);

  @$pb.TagNumber(12)
  $0.Timestamp get periodStart => $_getN(11);
  @$pb.TagNumber(12)
  set periodStart($0.Timestamp value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasPeriodStart() => $_has(11);
  @$pb.TagNumber(12)
  void clearPeriodStart() => $_clearField(12);
  @$pb.TagNumber(12)
  $0.Timestamp ensurePeriodStart() => $_ensure(11);

  @$pb.TagNumber(13)
  $0.Timestamp get periodEnd => $_getN(12);
  @$pb.TagNumber(13)
  set periodEnd($0.Timestamp value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasPeriodEnd() => $_has(12);
  @$pb.TagNumber(13)
  void clearPeriodEnd() => $_clearField(13);
  @$pb.TagNumber(13)
  $0.Timestamp ensurePeriodEnd() => $_ensure(12);

  @$pb.TagNumber(14)
  $0.Timestamp get issuedAt => $_getN(13);
  @$pb.TagNumber(14)
  set issuedAt($0.Timestamp value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasIssuedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearIssuedAt() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Timestamp ensureIssuedAt() => $_ensure(13);

  @$pb.TagNumber(15)
  $0.Timestamp get dueAt => $_getN(14);
  @$pb.TagNumber(15)
  set dueAt($0.Timestamp value) => $_setField(15, value);
  @$pb.TagNumber(15)
  $core.bool hasDueAt() => $_has(14);
  @$pb.TagNumber(15)
  void clearDueAt() => $_clearField(15);
  @$pb.TagNumber(15)
  $0.Timestamp ensureDueAt() => $_ensure(14);

  @$pb.TagNumber(16)
  $0.Timestamp get paidAt => $_getN(15);
  @$pb.TagNumber(16)
  set paidAt($0.Timestamp value) => $_setField(16, value);
  @$pb.TagNumber(16)
  $core.bool hasPaidAt() => $_has(15);
  @$pb.TagNumber(16)
  void clearPaidAt() => $_clearField(16);
  @$pb.TagNumber(16)
  $0.Timestamp ensurePaidAt() => $_ensure(15);

  @$pb.TagNumber(17)
  $core.String get ledgerTxnId => $_getSZ(16);
  @$pb.TagNumber(17)
  set ledgerTxnId($core.String value) => $_setString(16, value);
  @$pb.TagNumber(17)
  $core.bool hasLedgerTxnId() => $_has(16);
  @$pb.TagNumber(17)
  void clearLedgerTxnId() => $_clearField(17);

  @$pb.TagNumber(18)
  $1.Struct get data => $_getN(17);
  @$pb.TagNumber(18)
  set data($1.Struct value) => $_setField(18, value);
  @$pb.TagNumber(18)
  $core.bool hasData() => $_has(17);
  @$pb.TagNumber(18)
  void clearData() => $_clearField(18);
  @$pb.TagNumber(18)
  $1.Struct ensureData() => $_ensure(17);

  @$pb.TagNumber(19)
  $pb.PbList<InvoiceLine> get lines => $_getList(18);
}

/// InvoiceLine represents a line item on an invoice.
class InvoiceLine extends $pb.GeneratedMessage {
  factory InvoiceLine({
    $core.String? id,
    $core.String? invoiceId,
    $core.String? componentId,
    $core.String? description,
    $core.double? quantity,
    $2.Money? unitPrice,
    $2.Money? amount,
    $2.Money? discountAmount,
    $2.Money? creditAmount,
    $2.Money? netAmount,
    $core.String? currency,
    InvoiceLineType? lineType,
    $1.Struct? data,
  }) {
    final result = InvoiceLine._();
    if (id != null) result.id = id;
    if (invoiceId != null) result.invoiceId = invoiceId;
    if (componentId != null) result.componentId = componentId;
    if (description != null) result.description = description;
    if (quantity != null) result.quantity = quantity;
    if (unitPrice != null) result.unitPrice = unitPrice;
    if (amount != null) result.amount = amount;
    if (discountAmount != null) result.discountAmount = discountAmount;
    if (creditAmount != null) result.creditAmount = creditAmount;
    if (netAmount != null) result.netAmount = netAmount;
    if (currency != null) result.currency = currency;
    if (lineType != null) result.lineType = lineType;
    if (data != null) result.data = data;
    return result;
  }

  InvoiceLine._();

  factory InvoiceLine.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InvoiceLine()..mergeFromBuffer(data, registry);
  factory InvoiceLine.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InvoiceLine()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InvoiceLine',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: InvoiceLine.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'invoiceId')
    ..aOS(3, _omitFieldNames ? '' : 'componentId')
    ..aOS(4, _omitFieldNames ? '' : 'description')
    ..aD(5, _omitFieldNames ? '' : 'quantity')
    ..aOM<$2.Money>(6, _omitFieldNames ? '' : 'unitPrice',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(7, _omitFieldNames ? '' : 'amount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(8, _omitFieldNames ? '' : 'discountAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(9, _omitFieldNames ? '' : 'creditAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(10, _omitFieldNames ? '' : 'netAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOS(11, _omitFieldNames ? '' : 'currency')
    ..aE<InvoiceLineType>(12, _omitFieldNames ? '' : 'lineType',
        enumValues: InvoiceLineType.values)
    ..aOM<$1.Struct>(13, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvoiceLine clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvoiceLine copyWith(void Function(InvoiceLine) updates) =>
      super.copyWith((message) => updates(message as InvoiceLine))
          as InvoiceLine;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use InvoiceLine() / InvoiceLine.new instead')
  static InvoiceLine create() => InvoiceLine._();
  static $pb.GeneratedMessage $_createMessage() => InvoiceLine._();
  @$core.override
  InvoiceLine createEmptyInstance() => InvoiceLine._();
  @$core.pragma('dart2js:noInline')
  static InvoiceLine getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<InvoiceLine>(
          InvoiceLine.$_createMessage);
  static InvoiceLine? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get invoiceId => $_getSZ(1);
  @$pb.TagNumber(2)
  set invoiceId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInvoiceId() => $_has(1);
  @$pb.TagNumber(2)
  void clearInvoiceId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get componentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set componentId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasComponentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearComponentId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get description => $_getSZ(3);
  @$pb.TagNumber(4)
  set description($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDescription() => $_has(3);
  @$pb.TagNumber(4)
  void clearDescription() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.double get quantity => $_getN(4);
  @$pb.TagNumber(5)
  set quantity($core.double value) => $_setDouble(4, value);
  @$pb.TagNumber(5)
  $core.bool hasQuantity() => $_has(4);
  @$pb.TagNumber(5)
  void clearQuantity() => $_clearField(5);

  @$pb.TagNumber(6)
  $2.Money get unitPrice => $_getN(5);
  @$pb.TagNumber(6)
  set unitPrice($2.Money value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUnitPrice() => $_has(5);
  @$pb.TagNumber(6)
  void clearUnitPrice() => $_clearField(6);
  @$pb.TagNumber(6)
  $2.Money ensureUnitPrice() => $_ensure(5);

  @$pb.TagNumber(7)
  $2.Money get amount => $_getN(6);
  @$pb.TagNumber(7)
  set amount($2.Money value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasAmount() => $_has(6);
  @$pb.TagNumber(7)
  void clearAmount() => $_clearField(7);
  @$pb.TagNumber(7)
  $2.Money ensureAmount() => $_ensure(6);

  @$pb.TagNumber(8)
  $2.Money get discountAmount => $_getN(7);
  @$pb.TagNumber(8)
  set discountAmount($2.Money value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasDiscountAmount() => $_has(7);
  @$pb.TagNumber(8)
  void clearDiscountAmount() => $_clearField(8);
  @$pb.TagNumber(8)
  $2.Money ensureDiscountAmount() => $_ensure(7);

  @$pb.TagNumber(9)
  $2.Money get creditAmount => $_getN(8);
  @$pb.TagNumber(9)
  set creditAmount($2.Money value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasCreditAmount() => $_has(8);
  @$pb.TagNumber(9)
  void clearCreditAmount() => $_clearField(9);
  @$pb.TagNumber(9)
  $2.Money ensureCreditAmount() => $_ensure(8);

  @$pb.TagNumber(10)
  $2.Money get netAmount => $_getN(9);
  @$pb.TagNumber(10)
  set netAmount($2.Money value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasNetAmount() => $_has(9);
  @$pb.TagNumber(10)
  void clearNetAmount() => $_clearField(10);
  @$pb.TagNumber(10)
  $2.Money ensureNetAmount() => $_ensure(9);

  @$pb.TagNumber(11)
  $core.String get currency => $_getSZ(10);
  @$pb.TagNumber(11)
  set currency($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasCurrency() => $_has(10);
  @$pb.TagNumber(11)
  void clearCurrency() => $_clearField(11);

  @$pb.TagNumber(12)
  InvoiceLineType get lineType => $_getN(11);
  @$pb.TagNumber(12)
  set lineType(InvoiceLineType value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasLineType() => $_has(11);
  @$pb.TagNumber(12)
  void clearLineType() => $_clearField(12);

  @$pb.TagNumber(13)
  $1.Struct get data => $_getN(12);
  @$pb.TagNumber(13)
  set data($1.Struct value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasData() => $_has(12);
  @$pb.TagNumber(13)
  void clearData() => $_clearField(13);
  @$pb.TagNumber(13)
  $1.Struct ensureData() => $_ensure(12);
}

/// CreditGrant represents a prepaid credit grant to a customer.
class CreditGrant extends $pb.GeneratedMessage {
  factory CreditGrant({
    $core.String? id,
    $core.String? profileId,
    $core.String? name,
    $2.Money? originalAmount,
    $2.Money? remainingAmount,
    $core.String? currency,
    $0.Timestamp? expiresAt,
    $core.int? priority,
    $1.Struct? data,
  }) {
    final result = CreditGrant._();
    if (id != null) result.id = id;
    if (profileId != null) result.profileId = profileId;
    if (name != null) result.name = name;
    if (originalAmount != null) result.originalAmount = originalAmount;
    if (remainingAmount != null) result.remainingAmount = remainingAmount;
    if (currency != null) result.currency = currency;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (priority != null) result.priority = priority;
    if (data != null) result.data = data;
    return result;
  }

  CreditGrant._();

  factory CreditGrant.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreditGrant()..mergeFromBuffer(data, registry);
  factory CreditGrant.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreditGrant()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreditGrant',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreditGrant.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'profileId')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOM<$2.Money>(4, _omitFieldNames ? '' : 'originalAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(5, _omitFieldNames ? '' : 'remainingAmount',
        subBuilder: $2.Money.$_createMessage)
    ..aOS(6, _omitFieldNames ? '' : 'currency')
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aI(8, _omitFieldNames ? '' : 'priority')
    ..aOM<$1.Struct>(9, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreditGrant clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreditGrant copyWith(void Function(CreditGrant) updates) =>
      super.copyWith((message) => updates(message as CreditGrant))
          as CreditGrant;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreditGrant() / CreditGrant.new instead')
  static CreditGrant create() => CreditGrant._();
  static $pb.GeneratedMessage $_createMessage() => CreditGrant._();
  @$core.override
  CreditGrant createEmptyInstance() => CreditGrant._();
  @$core.pragma('dart2js:noInline')
  static CreditGrant getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreditGrant>(
          CreditGrant.$_createMessage);
  static CreditGrant? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get profileId => $_getSZ(1);
  @$pb.TagNumber(2)
  set profileId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProfileId() => $_has(1);
  @$pb.TagNumber(2)
  void clearProfileId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $2.Money get originalAmount => $_getN(3);
  @$pb.TagNumber(4)
  set originalAmount($2.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasOriginalAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearOriginalAmount() => $_clearField(4);
  @$pb.TagNumber(4)
  $2.Money ensureOriginalAmount() => $_ensure(3);

  @$pb.TagNumber(5)
  $2.Money get remainingAmount => $_getN(4);
  @$pb.TagNumber(5)
  set remainingAmount($2.Money value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasRemainingAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearRemainingAmount() => $_clearField(5);
  @$pb.TagNumber(5)
  $2.Money ensureRemainingAmount() => $_ensure(4);

  @$pb.TagNumber(6)
  $core.String get currency => $_getSZ(5);
  @$pb.TagNumber(6)
  set currency($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCurrency() => $_has(5);
  @$pb.TagNumber(6)
  void clearCurrency() => $_clearField(6);

  @$pb.TagNumber(7)
  $0.Timestamp get expiresAt => $_getN(6);
  @$pb.TagNumber(7)
  set expiresAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasExpiresAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearExpiresAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureExpiresAt() => $_ensure(6);

  @$pb.TagNumber(8)
  $core.int get priority => $_getIZ(7);
  @$pb.TagNumber(8)
  set priority($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasPriority() => $_has(7);
  @$pb.TagNumber(8)
  void clearPriority() => $_clearField(8);

  @$pb.TagNumber(9)
  $1.Struct get data => $_getN(8);
  @$pb.TagNumber(9)
  set data($1.Struct value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasData() => $_has(8);
  @$pb.TagNumber(9)
  void clearData() => $_clearField(9);
  @$pb.TagNumber(9)
  $1.Struct ensureData() => $_ensure(8);
}

/// BillingRun represents a billing workflow execution.
class BillingRun extends $pb.GeneratedMessage {
  factory BillingRun({
    $core.String? id,
    $core.String? subscriptionId,
    $core.String? profileId,
    $core.String? catalogVersionId,
    BillingRunState? state,
    $0.Timestamp? periodStart,
    $0.Timestamp? periodEnd,
    $0.Timestamp? startedAt,
    $0.Timestamp? completedAt,
    $0.Timestamp? failedAt,
    $core.String? errorMessage,
    $core.String? invoiceId,
    $1.Struct? data,
  }) {
    final result = BillingRun._();
    if (id != null) result.id = id;
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (profileId != null) result.profileId = profileId;
    if (catalogVersionId != null) result.catalogVersionId = catalogVersionId;
    if (state != null) result.state = state;
    if (periodStart != null) result.periodStart = periodStart;
    if (periodEnd != null) result.periodEnd = periodEnd;
    if (startedAt != null) result.startedAt = startedAt;
    if (completedAt != null) result.completedAt = completedAt;
    if (failedAt != null) result.failedAt = failedAt;
    if (errorMessage != null) result.errorMessage = errorMessage;
    if (invoiceId != null) result.invoiceId = invoiceId;
    if (data != null) result.data = data;
    return result;
  }

  BillingRun._();

  factory BillingRun.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BillingRun()..mergeFromBuffer(data, registry);
  factory BillingRun.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BillingRun()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BillingRun',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: BillingRun.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(3, _omitFieldNames ? '' : 'profileId')
    ..aOS(4, _omitFieldNames ? '' : 'catalogVersionId')
    ..aE<BillingRunState>(5, _omitFieldNames ? '' : 'state',
        enumValues: BillingRunState.values)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'periodStart',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'periodEnd',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'startedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(9, _omitFieldNames ? '' : 'completedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(10, _omitFieldNames ? '' : 'failedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(11, _omitFieldNames ? '' : 'errorMessage')
    ..aOS(12, _omitFieldNames ? '' : 'invoiceId')
    ..aOM<$1.Struct>(13, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BillingRun clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BillingRun copyWith(void Function(BillingRun) updates) =>
      super.copyWith((message) => updates(message as BillingRun)) as BillingRun;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BillingRun() / BillingRun.new instead')
  static BillingRun create() => BillingRun._();
  static $pb.GeneratedMessage $_createMessage() => BillingRun._();
  @$core.override
  BillingRun createEmptyInstance() => BillingRun._();
  @$core.pragma('dart2js:noInline')
  static BillingRun getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BillingRun>(BillingRun.$_createMessage);
  static BillingRun? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get subscriptionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set subscriptionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSubscriptionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSubscriptionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get profileId => $_getSZ(2);
  @$pb.TagNumber(3)
  set profileId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProfileId() => $_has(2);
  @$pb.TagNumber(3)
  void clearProfileId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get catalogVersionId => $_getSZ(3);
  @$pb.TagNumber(4)
  set catalogVersionId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCatalogVersionId() => $_has(3);
  @$pb.TagNumber(4)
  void clearCatalogVersionId() => $_clearField(4);

  @$pb.TagNumber(5)
  BillingRunState get state => $_getN(4);
  @$pb.TagNumber(5)
  set state(BillingRunState value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasState() => $_has(4);
  @$pb.TagNumber(5)
  void clearState() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Timestamp get periodStart => $_getN(5);
  @$pb.TagNumber(6)
  set periodStart($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasPeriodStart() => $_has(5);
  @$pb.TagNumber(6)
  void clearPeriodStart() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensurePeriodStart() => $_ensure(5);

  @$pb.TagNumber(7)
  $0.Timestamp get periodEnd => $_getN(6);
  @$pb.TagNumber(7)
  set periodEnd($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasPeriodEnd() => $_has(6);
  @$pb.TagNumber(7)
  void clearPeriodEnd() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensurePeriodEnd() => $_ensure(6);

  @$pb.TagNumber(8)
  $0.Timestamp get startedAt => $_getN(7);
  @$pb.TagNumber(8)
  set startedAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasStartedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearStartedAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureStartedAt() => $_ensure(7);

  @$pb.TagNumber(9)
  $0.Timestamp get completedAt => $_getN(8);
  @$pb.TagNumber(9)
  set completedAt($0.Timestamp value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasCompletedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearCompletedAt() => $_clearField(9);
  @$pb.TagNumber(9)
  $0.Timestamp ensureCompletedAt() => $_ensure(8);

  @$pb.TagNumber(10)
  $0.Timestamp get failedAt => $_getN(9);
  @$pb.TagNumber(10)
  set failedAt($0.Timestamp value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasFailedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearFailedAt() => $_clearField(10);
  @$pb.TagNumber(10)
  $0.Timestamp ensureFailedAt() => $_ensure(9);

  @$pb.TagNumber(11)
  $core.String get errorMessage => $_getSZ(10);
  @$pb.TagNumber(11)
  set errorMessage($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasErrorMessage() => $_has(10);
  @$pb.TagNumber(11)
  void clearErrorMessage() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get invoiceId => $_getSZ(11);
  @$pb.TagNumber(12)
  set invoiceId($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasInvoiceId() => $_has(11);
  @$pb.TagNumber(12)
  void clearInvoiceId() => $_clearField(12);

  @$pb.TagNumber(13)
  $1.Struct get data => $_getN(12);
  @$pb.TagNumber(13)
  set data($1.Struct value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasData() => $_has(12);
  @$pb.TagNumber(13)
  void clearData() => $_clearField(13);
  @$pb.TagNumber(13)
  $1.Struct ensureData() => $_ensure(12);
}

/// Discount represents a discount rule.
class Discount extends $pb.GeneratedMessage {
  factory Discount({
    $core.String? id,
    $core.String? name,
    DiscountType? discountType,
    $core.double? value,
    $core.String? currency,
    $1.Struct? applicableTo,
    $0.Timestamp? startAt,
    $0.Timestamp? endAt,
    $core.int? maxApplications,
    $1.Struct? data,
  }) {
    final result = Discount._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (discountType != null) result.discountType = discountType;
    if (value != null) result.value = value;
    if (currency != null) result.currency = currency;
    if (applicableTo != null) result.applicableTo = applicableTo;
    if (startAt != null) result.startAt = startAt;
    if (endAt != null) result.endAt = endAt;
    if (maxApplications != null) result.maxApplications = maxApplications;
    if (data != null) result.data = data;
    return result;
  }

  Discount._();

  factory Discount.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Discount()..mergeFromBuffer(data, registry);
  factory Discount.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Discount()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Discount',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: Discount.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aE<DiscountType>(3, _omitFieldNames ? '' : 'discountType',
        enumValues: DiscountType.values)
    ..aD(4, _omitFieldNames ? '' : 'value')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Struct>(6, _omitFieldNames ? '' : 'applicableTo',
        subBuilder: $1.Struct.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'startAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'endAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aI(9, _omitFieldNames ? '' : 'maxApplications')
    ..aOM<$1.Struct>(10, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Discount clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Discount copyWith(void Function(Discount) updates) =>
      super.copyWith((message) => updates(message as Discount)) as Discount;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Discount() / Discount.new instead')
  static Discount create() => Discount._();
  static $pb.GeneratedMessage $_createMessage() => Discount._();
  @$core.override
  Discount createEmptyInstance() => Discount._();
  @$core.pragma('dart2js:noInline')
  static Discount getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Discount>(Discount.$_createMessage);
  static Discount? _defaultInstance;

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
  DiscountType get discountType => $_getN(2);
  @$pb.TagNumber(3)
  set discountType(DiscountType value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDiscountType() => $_has(2);
  @$pb.TagNumber(3)
  void clearDiscountType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.double get value => $_getN(3);
  @$pb.TagNumber(4)
  set value($core.double value) => $_setDouble(3, value);
  @$pb.TagNumber(4)
  $core.bool hasValue() => $_has(3);
  @$pb.TagNumber(4)
  void clearValue() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => $_clearField(5);

  @$pb.TagNumber(6)
  $1.Struct get applicableTo => $_getN(5);
  @$pb.TagNumber(6)
  set applicableTo($1.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasApplicableTo() => $_has(5);
  @$pb.TagNumber(6)
  void clearApplicableTo() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Struct ensureApplicableTo() => $_ensure(5);

  @$pb.TagNumber(7)
  $0.Timestamp get startAt => $_getN(6);
  @$pb.TagNumber(7)
  set startAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasStartAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearStartAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureStartAt() => $_ensure(6);

  @$pb.TagNumber(8)
  $0.Timestamp get endAt => $_getN(7);
  @$pb.TagNumber(8)
  set endAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasEndAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearEndAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureEndAt() => $_ensure(7);

  @$pb.TagNumber(9)
  $core.int get maxApplications => $_getIZ(8);
  @$pb.TagNumber(9)
  set maxApplications($core.int value) => $_setSignedInt32(8, value);
  @$pb.TagNumber(9)
  $core.bool hasMaxApplications() => $_has(8);
  @$pb.TagNumber(9)
  void clearMaxApplications() => $_clearField(9);

  @$pb.TagNumber(10)
  $1.Struct get data => $_getN(9);
  @$pb.TagNumber(10)
  set data($1.Struct value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasData() => $_has(9);
  @$pb.TagNumber(10)
  void clearData() => $_clearField(10);
  @$pb.TagNumber(10)
  $1.Struct ensureData() => $_ensure(9);
}

/// CreateCatalogVersionRequest creates a new catalog version.
class CreateCatalogVersionRequest extends $pb.GeneratedMessage {
  factory CreateCatalogVersionRequest({
    $core.String? id,
    $core.String? catalogId,
    $core.String? name,
    $core.String? currency,
    $1.Struct? data,
  }) {
    final result = CreateCatalogVersionRequest._();
    if (id != null) result.id = id;
    if (catalogId != null) result.catalogId = catalogId;
    if (name != null) result.name = name;
    if (currency != null) result.currency = currency;
    if (data != null) result.data = data;
    return result;
  }

  CreateCatalogVersionRequest._();

  factory CreateCatalogVersionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCatalogVersionRequest()..mergeFromBuffer(data, registry);
  factory CreateCatalogVersionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCatalogVersionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCatalogVersionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateCatalogVersionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'catalogId')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Struct>(5, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCatalogVersionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCatalogVersionRequest copyWith(
          void Function(CreateCatalogVersionRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateCatalogVersionRequest))
          as CreateCatalogVersionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCatalogVersionRequest() / CreateCatalogVersionRequest.new instead')
  static CreateCatalogVersionRequest create() =>
      CreateCatalogVersionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCatalogVersionRequest._();
  @$core.override
  CreateCatalogVersionRequest createEmptyInstance() =>
      CreateCatalogVersionRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateCatalogVersionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCatalogVersionRequest>(
          CreateCatalogVersionRequest.$_createMessage);
  static CreateCatalogVersionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get catalogId => $_getSZ(1);
  @$pb.TagNumber(2)
  set catalogId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCatalogId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCatalogId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(3);
  @$pb.TagNumber(4)
  set currency($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(3);
  @$pb.TagNumber(4)
  void clearCurrency() => $_clearField(4);

  @$pb.TagNumber(5)
  $1.Struct get data => $_getN(4);
  @$pb.TagNumber(5)
  set data($1.Struct value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasData() => $_has(4);
  @$pb.TagNumber(5)
  void clearData() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Struct ensureData() => $_ensure(4);
}

/// CreateCatalogVersionResponse returns the created catalog version.
class CreateCatalogVersionResponse extends $pb.GeneratedMessage {
  factory CreateCatalogVersionResponse({
    CatalogVersion? data,
  }) {
    final result = CreateCatalogVersionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateCatalogVersionResponse._();

  factory CreateCatalogVersionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCatalogVersionResponse()..mergeFromBuffer(data, registry);
  factory CreateCatalogVersionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCatalogVersionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCatalogVersionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateCatalogVersionResponse.$_createMessage)
    ..aOM<CatalogVersion>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CatalogVersion.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCatalogVersionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCatalogVersionResponse copyWith(
          void Function(CreateCatalogVersionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateCatalogVersionResponse))
          as CreateCatalogVersionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCatalogVersionResponse() / CreateCatalogVersionResponse.new instead')
  static CreateCatalogVersionResponse create() =>
      CreateCatalogVersionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCatalogVersionResponse._();
  @$core.override
  CreateCatalogVersionResponse createEmptyInstance() =>
      CreateCatalogVersionResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateCatalogVersionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCatalogVersionResponse>(
          CreateCatalogVersionResponse.$_createMessage);
  static CreateCatalogVersionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CatalogVersion get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CatalogVersion value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CatalogVersion ensureData() => $_ensure(0);
}

/// GetCatalogVersionRequest retrieves a catalog version by ID.
class GetCatalogVersionRequest extends $pb.GeneratedMessage {
  factory GetCatalogVersionRequest({
    $core.String? id,
  }) {
    final result = GetCatalogVersionRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetCatalogVersionRequest._();

  factory GetCatalogVersionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCatalogVersionRequest()..mergeFromBuffer(data, registry);
  factory GetCatalogVersionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCatalogVersionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCatalogVersionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetCatalogVersionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCatalogVersionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCatalogVersionRequest copyWith(
          void Function(GetCatalogVersionRequest) updates) =>
      super.copyWith((message) => updates(message as GetCatalogVersionRequest))
          as GetCatalogVersionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCatalogVersionRequest() / GetCatalogVersionRequest.new instead')
  static GetCatalogVersionRequest create() => GetCatalogVersionRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetCatalogVersionRequest._();
  @$core.override
  GetCatalogVersionRequest createEmptyInstance() =>
      GetCatalogVersionRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCatalogVersionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCatalogVersionRequest>(
          GetCatalogVersionRequest.$_createMessage);
  static GetCatalogVersionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// GetCatalogVersionResponse returns the requested catalog version.
class GetCatalogVersionResponse extends $pb.GeneratedMessage {
  factory GetCatalogVersionResponse({
    CatalogVersion? data,
  }) {
    final result = GetCatalogVersionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  GetCatalogVersionResponse._();

  factory GetCatalogVersionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCatalogVersionResponse()..mergeFromBuffer(data, registry);
  factory GetCatalogVersionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCatalogVersionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCatalogVersionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetCatalogVersionResponse.$_createMessage)
    ..aOM<CatalogVersion>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CatalogVersion.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCatalogVersionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCatalogVersionResponse copyWith(
          void Function(GetCatalogVersionResponse) updates) =>
      super.copyWith((message) => updates(message as GetCatalogVersionResponse))
          as GetCatalogVersionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCatalogVersionResponse() / GetCatalogVersionResponse.new instead')
  static GetCatalogVersionResponse create() => GetCatalogVersionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetCatalogVersionResponse._();
  @$core.override
  GetCatalogVersionResponse createEmptyInstance() =>
      GetCatalogVersionResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCatalogVersionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCatalogVersionResponse>(
          GetCatalogVersionResponse.$_createMessage);
  static GetCatalogVersionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CatalogVersion get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CatalogVersion value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CatalogVersion ensureData() => $_ensure(0);
}

/// PublishCatalogVersionRequest publishes a catalog version, making it effective.
class PublishCatalogVersionRequest extends $pb.GeneratedMessage {
  factory PublishCatalogVersionRequest({
    $core.String? id,
    $0.Timestamp? effectiveAt,
  }) {
    final result = PublishCatalogVersionRequest._();
    if (id != null) result.id = id;
    if (effectiveAt != null) result.effectiveAt = effectiveAt;
    return result;
  }

  PublishCatalogVersionRequest._();

  factory PublishCatalogVersionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishCatalogVersionRequest()..mergeFromBuffer(data, registry);
  factory PublishCatalogVersionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishCatalogVersionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PublishCatalogVersionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: PublishCatalogVersionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'effectiveAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishCatalogVersionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishCatalogVersionRequest copyWith(
          void Function(PublishCatalogVersionRequest) updates) =>
      super.copyWith(
              (message) => updates(message as PublishCatalogVersionRequest))
          as PublishCatalogVersionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PublishCatalogVersionRequest() / PublishCatalogVersionRequest.new instead')
  static PublishCatalogVersionRequest create() =>
      PublishCatalogVersionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      PublishCatalogVersionRequest._();
  @$core.override
  PublishCatalogVersionRequest createEmptyInstance() =>
      PublishCatalogVersionRequest._();
  @$core.pragma('dart2js:noInline')
  static PublishCatalogVersionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PublishCatalogVersionRequest>(
          PublishCatalogVersionRequest.$_createMessage);
  static PublishCatalogVersionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $0.Timestamp get effectiveAt => $_getN(1);
  @$pb.TagNumber(2)
  set effectiveAt($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasEffectiveAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearEffectiveAt() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureEffectiveAt() => $_ensure(1);
}

/// PublishCatalogVersionResponse returns the published catalog version.
class PublishCatalogVersionResponse extends $pb.GeneratedMessage {
  factory PublishCatalogVersionResponse({
    CatalogVersion? data,
  }) {
    final result = PublishCatalogVersionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  PublishCatalogVersionResponse._();

  factory PublishCatalogVersionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishCatalogVersionResponse()..mergeFromBuffer(data, registry);
  factory PublishCatalogVersionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishCatalogVersionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PublishCatalogVersionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: PublishCatalogVersionResponse.$_createMessage)
    ..aOM<CatalogVersion>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CatalogVersion.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishCatalogVersionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishCatalogVersionResponse copyWith(
          void Function(PublishCatalogVersionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as PublishCatalogVersionResponse))
          as PublishCatalogVersionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PublishCatalogVersionResponse() / PublishCatalogVersionResponse.new instead')
  static PublishCatalogVersionResponse create() =>
      PublishCatalogVersionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      PublishCatalogVersionResponse._();
  @$core.override
  PublishCatalogVersionResponse createEmptyInstance() =>
      PublishCatalogVersionResponse._();
  @$core.pragma('dart2js:noInline')
  static PublishCatalogVersionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PublishCatalogVersionResponse>(
          PublishCatalogVersionResponse.$_createMessage);
  static PublishCatalogVersionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CatalogVersion get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CatalogVersion value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CatalogVersion ensureData() => $_ensure(0);
}

/// SearchCatalogVersionsResponse returns catalog versions matching search criteria.
class SearchCatalogVersionsResponse extends $pb.GeneratedMessage {
  factory SearchCatalogVersionsResponse({
    $core.Iterable<CatalogVersion>? data,
  }) {
    final result = SearchCatalogVersionsResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchCatalogVersionsResponse._();

  factory SearchCatalogVersionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchCatalogVersionsResponse()..mergeFromBuffer(data, registry);
  factory SearchCatalogVersionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchCatalogVersionsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchCatalogVersionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: SearchCatalogVersionsResponse.$_createMessage)
    ..pPM<CatalogVersion>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CatalogVersion.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchCatalogVersionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchCatalogVersionsResponse copyWith(
          void Function(SearchCatalogVersionsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SearchCatalogVersionsResponse))
          as SearchCatalogVersionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchCatalogVersionsResponse() / SearchCatalogVersionsResponse.new instead')
  static SearchCatalogVersionsResponse create() =>
      SearchCatalogVersionsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SearchCatalogVersionsResponse._();
  @$core.override
  SearchCatalogVersionsResponse createEmptyInstance() =>
      SearchCatalogVersionsResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchCatalogVersionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchCatalogVersionsResponse>(
          SearchCatalogVersionsResponse.$_createMessage);
  static SearchCatalogVersionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CatalogVersion> get data => $_getList(0);
}

/// CreatePlanRequest creates a new plan within a catalog version.
class CreatePlanRequest extends $pb.GeneratedMessage {
  factory CreatePlanRequest({
    $core.String? id,
    $core.String? catalogVersionId,
    $core.String? externalId,
    $core.String? name,
    $core.String? description,
    $1.Struct? data,
  }) {
    final result = CreatePlanRequest._();
    if (id != null) result.id = id;
    if (catalogVersionId != null) result.catalogVersionId = catalogVersionId;
    if (externalId != null) result.externalId = externalId;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (data != null) result.data = data;
    return result;
  }

  CreatePlanRequest._();

  factory CreatePlanRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreatePlanRequest()..mergeFromBuffer(data, registry);
  factory CreatePlanRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreatePlanRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreatePlanRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreatePlanRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'catalogVersionId')
    ..aOS(3, _omitFieldNames ? '' : 'externalId')
    ..aOS(4, _omitFieldNames ? '' : 'name')
    ..aOS(5, _omitFieldNames ? '' : 'description')
    ..aOM<$1.Struct>(6, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreatePlanRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreatePlanRequest copyWith(void Function(CreatePlanRequest) updates) =>
      super.copyWith((message) => updates(message as CreatePlanRequest))
          as CreatePlanRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreatePlanRequest() / CreatePlanRequest.new instead')
  static CreatePlanRequest create() => CreatePlanRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreatePlanRequest._();
  @$core.override
  CreatePlanRequest createEmptyInstance() => CreatePlanRequest._();
  @$core.pragma('dart2js:noInline')
  static CreatePlanRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreatePlanRequest>(
          CreatePlanRequest.$_createMessage);
  static CreatePlanRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get catalogVersionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set catalogVersionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCatalogVersionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCatalogVersionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get externalId => $_getSZ(2);
  @$pb.TagNumber(3)
  set externalId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExternalId() => $_has(2);
  @$pb.TagNumber(3)
  void clearExternalId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get name => $_getSZ(3);
  @$pb.TagNumber(4)
  set name($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get description => $_getSZ(4);
  @$pb.TagNumber(5)
  set description($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDescription() => $_has(4);
  @$pb.TagNumber(5)
  void clearDescription() => $_clearField(5);

  @$pb.TagNumber(6)
  $1.Struct get data => $_getN(5);
  @$pb.TagNumber(6)
  set data($1.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasData() => $_has(5);
  @$pb.TagNumber(6)
  void clearData() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Struct ensureData() => $_ensure(5);
}

/// CreatePlanResponse returns the created plan.
class CreatePlanResponse extends $pb.GeneratedMessage {
  factory CreatePlanResponse({
    Plan? data,
  }) {
    final result = CreatePlanResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreatePlanResponse._();

  factory CreatePlanResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreatePlanResponse()..mergeFromBuffer(data, registry);
  factory CreatePlanResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreatePlanResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreatePlanResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreatePlanResponse.$_createMessage)
    ..aOM<Plan>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Plan.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreatePlanResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreatePlanResponse copyWith(void Function(CreatePlanResponse) updates) =>
      super.copyWith((message) => updates(message as CreatePlanResponse))
          as CreatePlanResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreatePlanResponse() / CreatePlanResponse.new instead')
  static CreatePlanResponse create() => CreatePlanResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreatePlanResponse._();
  @$core.override
  CreatePlanResponse createEmptyInstance() => CreatePlanResponse._();
  @$core.pragma('dart2js:noInline')
  static CreatePlanResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreatePlanResponse>(
          CreatePlanResponse.$_createMessage);
  static CreatePlanResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Plan get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Plan value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Plan ensureData() => $_ensure(0);
}

/// CreateComponentRequest creates a new billable component within a plan.
class CreateComponentRequest extends $pb.GeneratedMessage {
  factory CreateComponentRequest({
    $core.String? id,
    $core.String? planId,
    $core.String? externalId,
    $core.String? name,
    $core.String? metricKey,
    PricingModel? pricingModel,
    AggregationType? aggregationType,
    $core.String? unitName,
    $2.Money? freeQuantity,
    $2.Money? minimumCharge,
    $1.Struct? data,
  }) {
    final result = CreateComponentRequest._();
    if (id != null) result.id = id;
    if (planId != null) result.planId = planId;
    if (externalId != null) result.externalId = externalId;
    if (name != null) result.name = name;
    if (metricKey != null) result.metricKey = metricKey;
    if (pricingModel != null) result.pricingModel = pricingModel;
    if (aggregationType != null) result.aggregationType = aggregationType;
    if (unitName != null) result.unitName = unitName;
    if (freeQuantity != null) result.freeQuantity = freeQuantity;
    if (minimumCharge != null) result.minimumCharge = minimumCharge;
    if (data != null) result.data = data;
    return result;
  }

  CreateComponentRequest._();

  factory CreateComponentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateComponentRequest()..mergeFromBuffer(data, registry);
  factory CreateComponentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateComponentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateComponentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateComponentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'planId')
    ..aOS(3, _omitFieldNames ? '' : 'externalId')
    ..aOS(4, _omitFieldNames ? '' : 'name')
    ..aOS(5, _omitFieldNames ? '' : 'metricKey')
    ..aE<PricingModel>(6, _omitFieldNames ? '' : 'pricingModel',
        enumValues: PricingModel.values)
    ..aE<AggregationType>(7, _omitFieldNames ? '' : 'aggregationType',
        enumValues: AggregationType.values)
    ..aOS(8, _omitFieldNames ? '' : 'unitName')
    ..aOM<$2.Money>(9, _omitFieldNames ? '' : 'freeQuantity',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(10, _omitFieldNames ? '' : 'minimumCharge',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$1.Struct>(11, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateComponentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateComponentRequest copyWith(
          void Function(CreateComponentRequest) updates) =>
      super.copyWith((message) => updates(message as CreateComponentRequest))
          as CreateComponentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateComponentRequest() / CreateComponentRequest.new instead')
  static CreateComponentRequest create() => CreateComponentRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateComponentRequest._();
  @$core.override
  CreateComponentRequest createEmptyInstance() => CreateComponentRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateComponentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateComponentRequest>(
          CreateComponentRequest.$_createMessage);
  static CreateComponentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get planId => $_getSZ(1);
  @$pb.TagNumber(2)
  set planId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPlanId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlanId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get externalId => $_getSZ(2);
  @$pb.TagNumber(3)
  set externalId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExternalId() => $_has(2);
  @$pb.TagNumber(3)
  void clearExternalId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get name => $_getSZ(3);
  @$pb.TagNumber(4)
  set name($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get metricKey => $_getSZ(4);
  @$pb.TagNumber(5)
  set metricKey($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasMetricKey() => $_has(4);
  @$pb.TagNumber(5)
  void clearMetricKey() => $_clearField(5);

  @$pb.TagNumber(6)
  PricingModel get pricingModel => $_getN(5);
  @$pb.TagNumber(6)
  set pricingModel(PricingModel value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasPricingModel() => $_has(5);
  @$pb.TagNumber(6)
  void clearPricingModel() => $_clearField(6);

  @$pb.TagNumber(7)
  AggregationType get aggregationType => $_getN(6);
  @$pb.TagNumber(7)
  set aggregationType(AggregationType value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasAggregationType() => $_has(6);
  @$pb.TagNumber(7)
  void clearAggregationType() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get unitName => $_getSZ(7);
  @$pb.TagNumber(8)
  set unitName($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasUnitName() => $_has(7);
  @$pb.TagNumber(8)
  void clearUnitName() => $_clearField(8);

  @$pb.TagNumber(9)
  $2.Money get freeQuantity => $_getN(8);
  @$pb.TagNumber(9)
  set freeQuantity($2.Money value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasFreeQuantity() => $_has(8);
  @$pb.TagNumber(9)
  void clearFreeQuantity() => $_clearField(9);
  @$pb.TagNumber(9)
  $2.Money ensureFreeQuantity() => $_ensure(8);

  @$pb.TagNumber(10)
  $2.Money get minimumCharge => $_getN(9);
  @$pb.TagNumber(10)
  set minimumCharge($2.Money value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasMinimumCharge() => $_has(9);
  @$pb.TagNumber(10)
  void clearMinimumCharge() => $_clearField(10);
  @$pb.TagNumber(10)
  $2.Money ensureMinimumCharge() => $_ensure(9);

  @$pb.TagNumber(11)
  $1.Struct get data => $_getN(10);
  @$pb.TagNumber(11)
  set data($1.Struct value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasData() => $_has(10);
  @$pb.TagNumber(11)
  void clearData() => $_clearField(11);
  @$pb.TagNumber(11)
  $1.Struct ensureData() => $_ensure(10);
}

/// CreateComponentResponse returns the created component.
class CreateComponentResponse extends $pb.GeneratedMessage {
  factory CreateComponentResponse({
    Component? data,
  }) {
    final result = CreateComponentResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateComponentResponse._();

  factory CreateComponentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateComponentResponse()..mergeFromBuffer(data, registry);
  factory CreateComponentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateComponentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateComponentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateComponentResponse.$_createMessage)
    ..aOM<Component>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Component.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateComponentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateComponentResponse copyWith(
          void Function(CreateComponentResponse) updates) =>
      super.copyWith((message) => updates(message as CreateComponentResponse))
          as CreateComponentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateComponentResponse() / CreateComponentResponse.new instead')
  static CreateComponentResponse create() => CreateComponentResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateComponentResponse._();
  @$core.override
  CreateComponentResponse createEmptyInstance() => CreateComponentResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateComponentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateComponentResponse>(
          CreateComponentResponse.$_createMessage);
  static CreateComponentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Component get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Component value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Component ensureData() => $_ensure(0);
}

/// CreateTierRequest creates a new pricing tier within a component.
class CreateTierRequest extends $pb.GeneratedMessage {
  factory CreateTierRequest({
    $core.String? id,
    $core.String? componentId,
    $core.int? sortOrder,
    $2.Money? lowerBound,
    $2.Money? upperBound,
    $2.Money? unitPrice,
    $2.Money? flatFee,
  }) {
    final result = CreateTierRequest._();
    if (id != null) result.id = id;
    if (componentId != null) result.componentId = componentId;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (lowerBound != null) result.lowerBound = lowerBound;
    if (upperBound != null) result.upperBound = upperBound;
    if (unitPrice != null) result.unitPrice = unitPrice;
    if (flatFee != null) result.flatFee = flatFee;
    return result;
  }

  CreateTierRequest._();

  factory CreateTierRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTierRequest()..mergeFromBuffer(data, registry);
  factory CreateTierRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTierRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTierRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateTierRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'componentId')
    ..aI(3, _omitFieldNames ? '' : 'sortOrder')
    ..aOM<$2.Money>(4, _omitFieldNames ? '' : 'lowerBound',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(5, _omitFieldNames ? '' : 'upperBound',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(6, _omitFieldNames ? '' : 'unitPrice',
        subBuilder: $2.Money.$_createMessage)
    ..aOM<$2.Money>(7, _omitFieldNames ? '' : 'flatFee',
        subBuilder: $2.Money.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTierRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTierRequest copyWith(void Function(CreateTierRequest) updates) =>
      super.copyWith((message) => updates(message as CreateTierRequest))
          as CreateTierRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateTierRequest() / CreateTierRequest.new instead')
  static CreateTierRequest create() => CreateTierRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateTierRequest._();
  @$core.override
  CreateTierRequest createEmptyInstance() => CreateTierRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateTierRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateTierRequest>(
          CreateTierRequest.$_createMessage);
  static CreateTierRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get componentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set componentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasComponentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearComponentId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get sortOrder => $_getIZ(2);
  @$pb.TagNumber(3)
  set sortOrder($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSortOrder() => $_has(2);
  @$pb.TagNumber(3)
  void clearSortOrder() => $_clearField(3);

  @$pb.TagNumber(4)
  $2.Money get lowerBound => $_getN(3);
  @$pb.TagNumber(4)
  set lowerBound($2.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasLowerBound() => $_has(3);
  @$pb.TagNumber(4)
  void clearLowerBound() => $_clearField(4);
  @$pb.TagNumber(4)
  $2.Money ensureLowerBound() => $_ensure(3);

  @$pb.TagNumber(5)
  $2.Money get upperBound => $_getN(4);
  @$pb.TagNumber(5)
  set upperBound($2.Money value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasUpperBound() => $_has(4);
  @$pb.TagNumber(5)
  void clearUpperBound() => $_clearField(5);
  @$pb.TagNumber(5)
  $2.Money ensureUpperBound() => $_ensure(4);

  @$pb.TagNumber(6)
  $2.Money get unitPrice => $_getN(5);
  @$pb.TagNumber(6)
  set unitPrice($2.Money value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUnitPrice() => $_has(5);
  @$pb.TagNumber(6)
  void clearUnitPrice() => $_clearField(6);
  @$pb.TagNumber(6)
  $2.Money ensureUnitPrice() => $_ensure(5);

  @$pb.TagNumber(7)
  $2.Money get flatFee => $_getN(6);
  @$pb.TagNumber(7)
  set flatFee($2.Money value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasFlatFee() => $_has(6);
  @$pb.TagNumber(7)
  void clearFlatFee() => $_clearField(7);
  @$pb.TagNumber(7)
  $2.Money ensureFlatFee() => $_ensure(6);
}

/// CreateTierResponse returns the created tier.
class CreateTierResponse extends $pb.GeneratedMessage {
  factory CreateTierResponse({
    Tier? data,
  }) {
    final result = CreateTierResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateTierResponse._();

  factory CreateTierResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTierResponse()..mergeFromBuffer(data, registry);
  factory CreateTierResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTierResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTierResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateTierResponse.$_createMessage)
    ..aOM<Tier>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Tier.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTierResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTierResponse copyWith(void Function(CreateTierResponse) updates) =>
      super.copyWith((message) => updates(message as CreateTierResponse))
          as CreateTierResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateTierResponse() / CreateTierResponse.new instead')
  static CreateTierResponse create() => CreateTierResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateTierResponse._();
  @$core.override
  CreateTierResponse createEmptyInstance() => CreateTierResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateTierResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTierResponse>(
          CreateTierResponse.$_createMessage);
  static CreateTierResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Tier get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Tier value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Tier ensureData() => $_ensure(0);
}

/// CreateSubscriptionRequest creates a new subscription.
class CreateSubscriptionRequest extends $pb.GeneratedMessage {
  factory CreateSubscriptionRequest({
    $core.String? id,
    $core.String? profileId,
    $core.String? catalogVersionId,
    $core.String? planId,
    $0.Timestamp? startAt,
    $0.Timestamp? billingAnchor,
    $core.String? currency,
    $1.Struct? data,
  }) {
    final result = CreateSubscriptionRequest._();
    if (id != null) result.id = id;
    if (profileId != null) result.profileId = profileId;
    if (catalogVersionId != null) result.catalogVersionId = catalogVersionId;
    if (planId != null) result.planId = planId;
    if (startAt != null) result.startAt = startAt;
    if (billingAnchor != null) result.billingAnchor = billingAnchor;
    if (currency != null) result.currency = currency;
    if (data != null) result.data = data;
    return result;
  }

  CreateSubscriptionRequest._();

  factory CreateSubscriptionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionRequest()..mergeFromBuffer(data, registry);
  factory CreateSubscriptionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSubscriptionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'profileId')
    ..aOS(3, _omitFieldNames ? '' : 'catalogVersionId')
    ..aOS(4, _omitFieldNames ? '' : 'planId')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'startAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'billingAnchor',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(7, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Struct>(8, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionRequest copyWith(
          void Function(CreateSubscriptionRequest) updates) =>
      super.copyWith((message) => updates(message as CreateSubscriptionRequest))
          as CreateSubscriptionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSubscriptionRequest() / CreateSubscriptionRequest.new instead')
  static CreateSubscriptionRequest create() => CreateSubscriptionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateSubscriptionRequest._();
  @$core.override
  CreateSubscriptionRequest createEmptyInstance() =>
      CreateSubscriptionRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateSubscriptionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSubscriptionRequest>(
          CreateSubscriptionRequest.$_createMessage);
  static CreateSubscriptionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get profileId => $_getSZ(1);
  @$pb.TagNumber(2)
  set profileId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProfileId() => $_has(1);
  @$pb.TagNumber(2)
  void clearProfileId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get catalogVersionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set catalogVersionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCatalogVersionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCatalogVersionId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get planId => $_getSZ(3);
  @$pb.TagNumber(4)
  set planId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPlanId() => $_has(3);
  @$pb.TagNumber(4)
  void clearPlanId() => $_clearField(4);

  @$pb.TagNumber(5)
  $0.Timestamp get startAt => $_getN(4);
  @$pb.TagNumber(5)
  set startAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasStartAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearStartAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureStartAt() => $_ensure(4);

  @$pb.TagNumber(6)
  $0.Timestamp get billingAnchor => $_getN(5);
  @$pb.TagNumber(6)
  set billingAnchor($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasBillingAnchor() => $_has(5);
  @$pb.TagNumber(6)
  void clearBillingAnchor() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureBillingAnchor() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.String get currency => $_getSZ(6);
  @$pb.TagNumber(7)
  set currency($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCurrency() => $_has(6);
  @$pb.TagNumber(7)
  void clearCurrency() => $_clearField(7);

  @$pb.TagNumber(8)
  $1.Struct get data => $_getN(7);
  @$pb.TagNumber(8)
  set data($1.Struct value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasData() => $_has(7);
  @$pb.TagNumber(8)
  void clearData() => $_clearField(8);
  @$pb.TagNumber(8)
  $1.Struct ensureData() => $_ensure(7);
}

/// CreateSubscriptionResponse returns the created subscription.
class CreateSubscriptionResponse extends $pb.GeneratedMessage {
  factory CreateSubscriptionResponse({
    Subscription? data,
  }) {
    final result = CreateSubscriptionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateSubscriptionResponse._();

  factory CreateSubscriptionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionResponse()..mergeFromBuffer(data, registry);
  factory CreateSubscriptionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSubscriptionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateSubscriptionResponse.$_createMessage)
    ..aOM<Subscription>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Subscription.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionResponse copyWith(
          void Function(CreateSubscriptionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateSubscriptionResponse))
          as CreateSubscriptionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSubscriptionResponse() / CreateSubscriptionResponse.new instead')
  static CreateSubscriptionResponse create() => CreateSubscriptionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateSubscriptionResponse._();
  @$core.override
  CreateSubscriptionResponse createEmptyInstance() =>
      CreateSubscriptionResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateSubscriptionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSubscriptionResponse>(
          CreateSubscriptionResponse.$_createMessage);
  static CreateSubscriptionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Subscription get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Subscription value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Subscription ensureData() => $_ensure(0);
}

/// GetSubscriptionRequest retrieves a subscription by ID.
class GetSubscriptionRequest extends $pb.GeneratedMessage {
  factory GetSubscriptionRequest({
    $core.String? id,
  }) {
    final result = GetSubscriptionRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetSubscriptionRequest._();

  factory GetSubscriptionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSubscriptionRequest()..mergeFromBuffer(data, registry);
  factory GetSubscriptionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSubscriptionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSubscriptionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSubscriptionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSubscriptionRequest copyWith(
          void Function(GetSubscriptionRequest) updates) =>
      super.copyWith((message) => updates(message as GetSubscriptionRequest))
          as GetSubscriptionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetSubscriptionRequest() / GetSubscriptionRequest.new instead')
  static GetSubscriptionRequest create() => GetSubscriptionRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetSubscriptionRequest._();
  @$core.override
  GetSubscriptionRequest createEmptyInstance() => GetSubscriptionRequest._();
  @$core.pragma('dart2js:noInline')
  static GetSubscriptionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetSubscriptionRequest>(
          GetSubscriptionRequest.$_createMessage);
  static GetSubscriptionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// GetSubscriptionResponse returns the requested subscription.
class GetSubscriptionResponse extends $pb.GeneratedMessage {
  factory GetSubscriptionResponse({
    Subscription? data,
  }) {
    final result = GetSubscriptionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  GetSubscriptionResponse._();

  factory GetSubscriptionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSubscriptionResponse()..mergeFromBuffer(data, registry);
  factory GetSubscriptionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSubscriptionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSubscriptionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetSubscriptionResponse.$_createMessage)
    ..aOM<Subscription>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Subscription.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSubscriptionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSubscriptionResponse copyWith(
          void Function(GetSubscriptionResponse) updates) =>
      super.copyWith((message) => updates(message as GetSubscriptionResponse))
          as GetSubscriptionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetSubscriptionResponse() / GetSubscriptionResponse.new instead')
  static GetSubscriptionResponse create() => GetSubscriptionResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetSubscriptionResponse._();
  @$core.override
  GetSubscriptionResponse createEmptyInstance() => GetSubscriptionResponse._();
  @$core.pragma('dart2js:noInline')
  static GetSubscriptionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetSubscriptionResponse>(
          GetSubscriptionResponse.$_createMessage);
  static GetSubscriptionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Subscription get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Subscription value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Subscription ensureData() => $_ensure(0);
}

/// CancelSubscriptionRequest cancels a subscription.
class CancelSubscriptionRequest extends $pb.GeneratedMessage {
  factory CancelSubscriptionRequest({
    $core.String? id,
  }) {
    final result = CancelSubscriptionRequest._();
    if (id != null) result.id = id;
    return result;
  }

  CancelSubscriptionRequest._();

  factory CancelSubscriptionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionRequest()..mergeFromBuffer(data, registry);
  factory CancelSubscriptionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelSubscriptionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CancelSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionRequest copyWith(
          void Function(CancelSubscriptionRequest) updates) =>
      super.copyWith((message) => updates(message as CancelSubscriptionRequest))
          as CancelSubscriptionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelSubscriptionRequest() / CancelSubscriptionRequest.new instead')
  static CancelSubscriptionRequest create() => CancelSubscriptionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CancelSubscriptionRequest._();
  @$core.override
  CancelSubscriptionRequest createEmptyInstance() =>
      CancelSubscriptionRequest._();
  @$core.pragma('dart2js:noInline')
  static CancelSubscriptionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelSubscriptionRequest>(
          CancelSubscriptionRequest.$_createMessage);
  static CancelSubscriptionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// CancelSubscriptionResponse returns the cancelled subscription.
class CancelSubscriptionResponse extends $pb.GeneratedMessage {
  factory CancelSubscriptionResponse({
    Subscription? data,
  }) {
    final result = CancelSubscriptionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CancelSubscriptionResponse._();

  factory CancelSubscriptionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionResponse()..mergeFromBuffer(data, registry);
  factory CancelSubscriptionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelSubscriptionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CancelSubscriptionResponse.$_createMessage)
    ..aOM<Subscription>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Subscription.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionResponse copyWith(
          void Function(CancelSubscriptionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CancelSubscriptionResponse))
          as CancelSubscriptionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelSubscriptionResponse() / CancelSubscriptionResponse.new instead')
  static CancelSubscriptionResponse create() => CancelSubscriptionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CancelSubscriptionResponse._();
  @$core.override
  CancelSubscriptionResponse createEmptyInstance() =>
      CancelSubscriptionResponse._();
  @$core.pragma('dart2js:noInline')
  static CancelSubscriptionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelSubscriptionResponse>(
          CancelSubscriptionResponse.$_createMessage);
  static CancelSubscriptionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Subscription get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Subscription value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Subscription ensureData() => $_ensure(0);
}

/// ListSubscriptionsRequest lists active subscriptions for a customer.
class ListSubscriptionsRequest extends $pb.GeneratedMessage {
  factory ListSubscriptionsRequest({
    $core.String? profileId,
  }) {
    final result = ListSubscriptionsRequest._();
    if (profileId != null) result.profileId = profileId;
    return result;
  }

  ListSubscriptionsRequest._();

  factory ListSubscriptionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionsRequest()..mergeFromBuffer(data, registry);
  factory ListSubscriptionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSubscriptionsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: ListSubscriptionsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'profileId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionsRequest copyWith(
          void Function(ListSubscriptionsRequest) updates) =>
      super.copyWith((message) => updates(message as ListSubscriptionsRequest))
          as ListSubscriptionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListSubscriptionsRequest() / ListSubscriptionsRequest.new instead')
  static ListSubscriptionsRequest create() => ListSubscriptionsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListSubscriptionsRequest._();
  @$core.override
  ListSubscriptionsRequest createEmptyInstance() =>
      ListSubscriptionsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListSubscriptionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSubscriptionsRequest>(
          ListSubscriptionsRequest.$_createMessage);
  static ListSubscriptionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get profileId => $_getSZ(0);
  @$pb.TagNumber(1)
  set profileId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProfileId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProfileId() => $_clearField(1);
}

/// ListSubscriptionsResponse returns the matching subscriptions.
class ListSubscriptionsResponse extends $pb.GeneratedMessage {
  factory ListSubscriptionsResponse({
    $core.Iterable<Subscription>? data,
  }) {
    final result = ListSubscriptionsResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  ListSubscriptionsResponse._();

  factory ListSubscriptionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionsResponse()..mergeFromBuffer(data, registry);
  factory ListSubscriptionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSubscriptionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: ListSubscriptionsResponse.$_createMessage)
    ..pPM<Subscription>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Subscription.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionsResponse copyWith(
          void Function(ListSubscriptionsResponse) updates) =>
      super.copyWith((message) => updates(message as ListSubscriptionsResponse))
          as ListSubscriptionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListSubscriptionsResponse() / ListSubscriptionsResponse.new instead')
  static ListSubscriptionsResponse create() => ListSubscriptionsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListSubscriptionsResponse._();
  @$core.override
  ListSubscriptionsResponse createEmptyInstance() =>
      ListSubscriptionsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListSubscriptionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSubscriptionsResponse>(
          ListSubscriptionsResponse.$_createMessage);
  static ListSubscriptionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Subscription> get data => $_getList(0);
}

/// IngestUsageEventRequest ingests a single usage event.
class IngestUsageEventRequest extends $pb.GeneratedMessage {
  factory IngestUsageEventRequest({
    $core.Iterable<UsageEvent>? data,
  }) {
    final result = IngestUsageEventRequest._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  IngestUsageEventRequest._();

  factory IngestUsageEventRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IngestUsageEventRequest()..mergeFromBuffer(data, registry);
  factory IngestUsageEventRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IngestUsageEventRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IngestUsageEventRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: IngestUsageEventRequest.$_createMessage)
    ..pPM<UsageEvent>(1, _omitFieldNames ? '' : 'data',
        subBuilder: UsageEvent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IngestUsageEventRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IngestUsageEventRequest copyWith(
          void Function(IngestUsageEventRequest) updates) =>
      super.copyWith((message) => updates(message as IngestUsageEventRequest))
          as IngestUsageEventRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use IngestUsageEventRequest() / IngestUsageEventRequest.new instead')
  static IngestUsageEventRequest create() => IngestUsageEventRequest._();
  static $pb.GeneratedMessage $_createMessage() => IngestUsageEventRequest._();
  @$core.override
  IngestUsageEventRequest createEmptyInstance() => IngestUsageEventRequest._();
  @$core.pragma('dart2js:noInline')
  static IngestUsageEventRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IngestUsageEventRequest>(
          IngestUsageEventRequest.$_createMessage);
  static IngestUsageEventRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<UsageEvent> get data => $_getList(0);
}

/// IngestUsageEventResponse returns the ingested usage event.
class IngestUsageEventResponse extends $pb.GeneratedMessage {
  factory IngestUsageEventResponse({
    $core.Iterable<$core.String>? data,
  }) {
    final result = IngestUsageEventResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  IngestUsageEventResponse._();

  factory IngestUsageEventResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IngestUsageEventResponse()..mergeFromBuffer(data, registry);
  factory IngestUsageEventResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IngestUsageEventResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IngestUsageEventResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: IngestUsageEventResponse.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'data')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IngestUsageEventResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IngestUsageEventResponse copyWith(
          void Function(IngestUsageEventResponse) updates) =>
      super.copyWith((message) => updates(message as IngestUsageEventResponse))
          as IngestUsageEventResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use IngestUsageEventResponse() / IngestUsageEventResponse.new instead')
  static IngestUsageEventResponse create() => IngestUsageEventResponse._();
  static $pb.GeneratedMessage $_createMessage() => IngestUsageEventResponse._();
  @$core.override
  IngestUsageEventResponse createEmptyInstance() =>
      IngestUsageEventResponse._();
  @$core.pragma('dart2js:noInline')
  static IngestUsageEventResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IngestUsageEventResponse>(
          IngestUsageEventResponse.$_createMessage);
  static IngestUsageEventResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get data => $_getList(0);
}

/// SearchUsageEventsResponse returns usage events matching search criteria.
class SearchUsageEventsResponse extends $pb.GeneratedMessage {
  factory SearchUsageEventsResponse({
    $core.Iterable<UsageEvent>? data,
  }) {
    final result = SearchUsageEventsResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchUsageEventsResponse._();

  factory SearchUsageEventsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchUsageEventsResponse()..mergeFromBuffer(data, registry);
  factory SearchUsageEventsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchUsageEventsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchUsageEventsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: SearchUsageEventsResponse.$_createMessage)
    ..pPM<UsageEvent>(1, _omitFieldNames ? '' : 'data',
        subBuilder: UsageEvent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchUsageEventsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchUsageEventsResponse copyWith(
          void Function(SearchUsageEventsResponse) updates) =>
      super.copyWith((message) => updates(message as SearchUsageEventsResponse))
          as SearchUsageEventsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchUsageEventsResponse() / SearchUsageEventsResponse.new instead')
  static SearchUsageEventsResponse create() => SearchUsageEventsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SearchUsageEventsResponse._();
  @$core.override
  SearchUsageEventsResponse createEmptyInstance() =>
      SearchUsageEventsResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchUsageEventsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchUsageEventsResponse>(
          SearchUsageEventsResponse.$_createMessage);
  static SearchUsageEventsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<UsageEvent> get data => $_getList(0);
}

/// -- Billing Run --
/// RunBillingRequest starts a billing run for a subscription and period.
class RunBillingRequest extends $pb.GeneratedMessage {
  factory RunBillingRequest({
    $core.String? subscriptionId,
    $0.Timestamp? periodStart,
    $0.Timestamp? periodEnd,
  }) {
    final result = RunBillingRequest._();
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (periodStart != null) result.periodStart = periodStart;
    if (periodEnd != null) result.periodEnd = periodEnd;
    return result;
  }

  RunBillingRequest._();

  factory RunBillingRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunBillingRequest()..mergeFromBuffer(data, registry);
  factory RunBillingRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunBillingRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RunBillingRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: RunBillingRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'subscriptionId')
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'periodStart',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(3, _omitFieldNames ? '' : 'periodEnd',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunBillingRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunBillingRequest copyWith(void Function(RunBillingRequest) updates) =>
      super.copyWith((message) => updates(message as RunBillingRequest))
          as RunBillingRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RunBillingRequest() / RunBillingRequest.new instead')
  static RunBillingRequest create() => RunBillingRequest._();
  static $pb.GeneratedMessage $_createMessage() => RunBillingRequest._();
  @$core.override
  RunBillingRequest createEmptyInstance() => RunBillingRequest._();
  @$core.pragma('dart2js:noInline')
  static RunBillingRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RunBillingRequest>(
          RunBillingRequest.$_createMessage);
  static RunBillingRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get subscriptionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set subscriptionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSubscriptionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSubscriptionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $0.Timestamp get periodStart => $_getN(1);
  @$pb.TagNumber(2)
  set periodStart($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPeriodStart() => $_has(1);
  @$pb.TagNumber(2)
  void clearPeriodStart() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensurePeriodStart() => $_ensure(1);

  @$pb.TagNumber(3)
  $0.Timestamp get periodEnd => $_getN(2);
  @$pb.TagNumber(3)
  set periodEnd($0.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasPeriodEnd() => $_has(2);
  @$pb.TagNumber(3)
  void clearPeriodEnd() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Timestamp ensurePeriodEnd() => $_ensure(2);
}

/// RunBillingResponse returns the billing run result.
class RunBillingResponse extends $pb.GeneratedMessage {
  factory RunBillingResponse({
    BillingRun? data,
  }) {
    final result = RunBillingResponse._();
    if (data != null) result.data = data;
    return result;
  }

  RunBillingResponse._();

  factory RunBillingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunBillingResponse()..mergeFromBuffer(data, registry);
  factory RunBillingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunBillingResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RunBillingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: RunBillingResponse.$_createMessage)
    ..aOM<BillingRun>(1, _omitFieldNames ? '' : 'data',
        subBuilder: BillingRun.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunBillingResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunBillingResponse copyWith(void Function(RunBillingResponse) updates) =>
      super.copyWith((message) => updates(message as RunBillingResponse))
          as RunBillingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RunBillingResponse() / RunBillingResponse.new instead')
  static RunBillingResponse create() => RunBillingResponse._();
  static $pb.GeneratedMessage $_createMessage() => RunBillingResponse._();
  @$core.override
  RunBillingResponse createEmptyInstance() => RunBillingResponse._();
  @$core.pragma('dart2js:noInline')
  static RunBillingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RunBillingResponse>(
          RunBillingResponse.$_createMessage);
  static RunBillingResponse? _defaultInstance;

  @$pb.TagNumber(1)
  BillingRun get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(BillingRun value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  BillingRun ensureData() => $_ensure(0);
}

/// GetBillingRunRequest retrieves a billing run by ID.
class GetBillingRunRequest extends $pb.GeneratedMessage {
  factory GetBillingRunRequest({
    $core.String? id,
  }) {
    final result = GetBillingRunRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetBillingRunRequest._();

  factory GetBillingRunRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingRunRequest()..mergeFromBuffer(data, registry);
  factory GetBillingRunRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingRunRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBillingRunRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetBillingRunRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingRunRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingRunRequest copyWith(void Function(GetBillingRunRequest) updates) =>
      super.copyWith((message) => updates(message as GetBillingRunRequest))
          as GetBillingRunRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetBillingRunRequest() / GetBillingRunRequest.new instead')
  static GetBillingRunRequest create() => GetBillingRunRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetBillingRunRequest._();
  @$core.override
  GetBillingRunRequest createEmptyInstance() => GetBillingRunRequest._();
  @$core.pragma('dart2js:noInline')
  static GetBillingRunRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetBillingRunRequest>(
          GetBillingRunRequest.$_createMessage);
  static GetBillingRunRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// GetBillingRunResponse returns the requested billing run.
class GetBillingRunResponse extends $pb.GeneratedMessage {
  factory GetBillingRunResponse({
    BillingRun? data,
  }) {
    final result = GetBillingRunResponse._();
    if (data != null) result.data = data;
    return result;
  }

  GetBillingRunResponse._();

  factory GetBillingRunResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingRunResponse()..mergeFromBuffer(data, registry);
  factory GetBillingRunResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingRunResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBillingRunResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetBillingRunResponse.$_createMessage)
    ..aOM<BillingRun>(1, _omitFieldNames ? '' : 'data',
        subBuilder: BillingRun.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingRunResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingRunResponse copyWith(
          void Function(GetBillingRunResponse) updates) =>
      super.copyWith((message) => updates(message as GetBillingRunResponse))
          as GetBillingRunResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetBillingRunResponse() / GetBillingRunResponse.new instead')
  static GetBillingRunResponse create() => GetBillingRunResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetBillingRunResponse._();
  @$core.override
  GetBillingRunResponse createEmptyInstance() => GetBillingRunResponse._();
  @$core.pragma('dart2js:noInline')
  static GetBillingRunResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetBillingRunResponse>(
          GetBillingRunResponse.$_createMessage);
  static GetBillingRunResponse? _defaultInstance;

  @$pb.TagNumber(1)
  BillingRun get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(BillingRun value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  BillingRun ensureData() => $_ensure(0);
}

/// GetInvoiceRequest retrieves an invoice by ID.
class GetInvoiceRequest extends $pb.GeneratedMessage {
  factory GetInvoiceRequest({
    $core.String? id,
  }) {
    final result = GetInvoiceRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetInvoiceRequest._();

  factory GetInvoiceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetInvoiceRequest()..mergeFromBuffer(data, registry);
  factory GetInvoiceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetInvoiceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetInvoiceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetInvoiceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInvoiceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInvoiceRequest copyWith(void Function(GetInvoiceRequest) updates) =>
      super.copyWith((message) => updates(message as GetInvoiceRequest))
          as GetInvoiceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetInvoiceRequest() / GetInvoiceRequest.new instead')
  static GetInvoiceRequest create() => GetInvoiceRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetInvoiceRequest._();
  @$core.override
  GetInvoiceRequest createEmptyInstance() => GetInvoiceRequest._();
  @$core.pragma('dart2js:noInline')
  static GetInvoiceRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetInvoiceRequest>(
          GetInvoiceRequest.$_createMessage);
  static GetInvoiceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// GetInvoiceResponse returns the requested invoice with lines.
class GetInvoiceResponse extends $pb.GeneratedMessage {
  factory GetInvoiceResponse({
    Invoice? data,
  }) {
    final result = GetInvoiceResponse._();
    if (data != null) result.data = data;
    return result;
  }

  GetInvoiceResponse._();

  factory GetInvoiceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetInvoiceResponse()..mergeFromBuffer(data, registry);
  factory GetInvoiceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetInvoiceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetInvoiceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetInvoiceResponse.$_createMessage)
    ..aOM<Invoice>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Invoice.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInvoiceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInvoiceResponse copyWith(void Function(GetInvoiceResponse) updates) =>
      super.copyWith((message) => updates(message as GetInvoiceResponse))
          as GetInvoiceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetInvoiceResponse() / GetInvoiceResponse.new instead')
  static GetInvoiceResponse create() => GetInvoiceResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetInvoiceResponse._();
  @$core.override
  GetInvoiceResponse createEmptyInstance() => GetInvoiceResponse._();
  @$core.pragma('dart2js:noInline')
  static GetInvoiceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetInvoiceResponse>(
          GetInvoiceResponse.$_createMessage);
  static GetInvoiceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Invoice get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Invoice value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Invoice ensureData() => $_ensure(0);
}

/// IssueInvoiceRequest issues a draft invoice.
class IssueInvoiceRequest extends $pb.GeneratedMessage {
  factory IssueInvoiceRequest({
    $core.String? id,
  }) {
    final result = IssueInvoiceRequest._();
    if (id != null) result.id = id;
    return result;
  }

  IssueInvoiceRequest._();

  factory IssueInvoiceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IssueInvoiceRequest()..mergeFromBuffer(data, registry);
  factory IssueInvoiceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IssueInvoiceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IssueInvoiceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: IssueInvoiceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IssueInvoiceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IssueInvoiceRequest copyWith(void Function(IssueInvoiceRequest) updates) =>
      super.copyWith((message) => updates(message as IssueInvoiceRequest))
          as IssueInvoiceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use IssueInvoiceRequest() / IssueInvoiceRequest.new instead')
  static IssueInvoiceRequest create() => IssueInvoiceRequest._();
  static $pb.GeneratedMessage $_createMessage() => IssueInvoiceRequest._();
  @$core.override
  IssueInvoiceRequest createEmptyInstance() => IssueInvoiceRequest._();
  @$core.pragma('dart2js:noInline')
  static IssueInvoiceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IssueInvoiceRequest>(
          IssueInvoiceRequest.$_createMessage);
  static IssueInvoiceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// IssueInvoiceResponse returns the issued invoice.
class IssueInvoiceResponse extends $pb.GeneratedMessage {
  factory IssueInvoiceResponse({
    Invoice? data,
  }) {
    final result = IssueInvoiceResponse._();
    if (data != null) result.data = data;
    return result;
  }

  IssueInvoiceResponse._();

  factory IssueInvoiceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IssueInvoiceResponse()..mergeFromBuffer(data, registry);
  factory IssueInvoiceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IssueInvoiceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IssueInvoiceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: IssueInvoiceResponse.$_createMessage)
    ..aOM<Invoice>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Invoice.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IssueInvoiceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IssueInvoiceResponse copyWith(void Function(IssueInvoiceResponse) updates) =>
      super.copyWith((message) => updates(message as IssueInvoiceResponse))
          as IssueInvoiceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use IssueInvoiceResponse() / IssueInvoiceResponse.new instead')
  static IssueInvoiceResponse create() => IssueInvoiceResponse._();
  static $pb.GeneratedMessage $_createMessage() => IssueInvoiceResponse._();
  @$core.override
  IssueInvoiceResponse createEmptyInstance() => IssueInvoiceResponse._();
  @$core.pragma('dart2js:noInline')
  static IssueInvoiceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IssueInvoiceResponse>(
          IssueInvoiceResponse.$_createMessage);
  static IssueInvoiceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Invoice get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Invoice value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Invoice ensureData() => $_ensure(0);
}

/// VoidInvoiceRequest voids an invoice.
class VoidInvoiceRequest extends $pb.GeneratedMessage {
  factory VoidInvoiceRequest({
    $core.String? id,
  }) {
    final result = VoidInvoiceRequest._();
    if (id != null) result.id = id;
    return result;
  }

  VoidInvoiceRequest._();

  factory VoidInvoiceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidInvoiceRequest()..mergeFromBuffer(data, registry);
  factory VoidInvoiceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidInvoiceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VoidInvoiceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: VoidInvoiceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidInvoiceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidInvoiceRequest copyWith(void Function(VoidInvoiceRequest) updates) =>
      super.copyWith((message) => updates(message as VoidInvoiceRequest))
          as VoidInvoiceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use VoidInvoiceRequest() / VoidInvoiceRequest.new instead')
  static VoidInvoiceRequest create() => VoidInvoiceRequest._();
  static $pb.GeneratedMessage $_createMessage() => VoidInvoiceRequest._();
  @$core.override
  VoidInvoiceRequest createEmptyInstance() => VoidInvoiceRequest._();
  @$core.pragma('dart2js:noInline')
  static VoidInvoiceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VoidInvoiceRequest>(
          VoidInvoiceRequest.$_createMessage);
  static VoidInvoiceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// VoidInvoiceResponse returns the voided invoice.
class VoidInvoiceResponse extends $pb.GeneratedMessage {
  factory VoidInvoiceResponse({
    Invoice? data,
  }) {
    final result = VoidInvoiceResponse._();
    if (data != null) result.data = data;
    return result;
  }

  VoidInvoiceResponse._();

  factory VoidInvoiceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidInvoiceResponse()..mergeFromBuffer(data, registry);
  factory VoidInvoiceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidInvoiceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VoidInvoiceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: VoidInvoiceResponse.$_createMessage)
    ..aOM<Invoice>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Invoice.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidInvoiceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidInvoiceResponse copyWith(void Function(VoidInvoiceResponse) updates) =>
      super.copyWith((message) => updates(message as VoidInvoiceResponse))
          as VoidInvoiceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use VoidInvoiceResponse() / VoidInvoiceResponse.new instead')
  static VoidInvoiceResponse create() => VoidInvoiceResponse._();
  static $pb.GeneratedMessage $_createMessage() => VoidInvoiceResponse._();
  @$core.override
  VoidInvoiceResponse createEmptyInstance() => VoidInvoiceResponse._();
  @$core.pragma('dart2js:noInline')
  static VoidInvoiceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VoidInvoiceResponse>(
          VoidInvoiceResponse.$_createMessage);
  static VoidInvoiceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Invoice get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Invoice value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Invoice ensureData() => $_ensure(0);
}

/// RecordPaymentRequest records payment on an invoice.
class RecordPaymentRequest extends $pb.GeneratedMessage {
  factory RecordPaymentRequest({
    $core.String? id,
  }) {
    final result = RecordPaymentRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RecordPaymentRequest._();

  factory RecordPaymentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentRequest()..mergeFromBuffer(data, registry);
  factory RecordPaymentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RecordPaymentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: RecordPaymentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentRequest copyWith(void Function(RecordPaymentRequest) updates) =>
      super.copyWith((message) => updates(message as RecordPaymentRequest))
          as RecordPaymentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RecordPaymentRequest() / RecordPaymentRequest.new instead')
  static RecordPaymentRequest create() => RecordPaymentRequest._();
  static $pb.GeneratedMessage $_createMessage() => RecordPaymentRequest._();
  @$core.override
  RecordPaymentRequest createEmptyInstance() => RecordPaymentRequest._();
  @$core.pragma('dart2js:noInline')
  static RecordPaymentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RecordPaymentRequest>(
          RecordPaymentRequest.$_createMessage);
  static RecordPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// RecordPaymentResponse returns the paid invoice.
class RecordPaymentResponse extends $pb.GeneratedMessage {
  factory RecordPaymentResponse({
    Invoice? data,
  }) {
    final result = RecordPaymentResponse._();
    if (data != null) result.data = data;
    return result;
  }

  RecordPaymentResponse._();

  factory RecordPaymentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentResponse()..mergeFromBuffer(data, registry);
  factory RecordPaymentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RecordPaymentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: RecordPaymentResponse.$_createMessage)
    ..aOM<Invoice>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Invoice.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentResponse copyWith(
          void Function(RecordPaymentResponse) updates) =>
      super.copyWith((message) => updates(message as RecordPaymentResponse))
          as RecordPaymentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RecordPaymentResponse() / RecordPaymentResponse.new instead')
  static RecordPaymentResponse create() => RecordPaymentResponse._();
  static $pb.GeneratedMessage $_createMessage() => RecordPaymentResponse._();
  @$core.override
  RecordPaymentResponse createEmptyInstance() => RecordPaymentResponse._();
  @$core.pragma('dart2js:noInline')
  static RecordPaymentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RecordPaymentResponse>(
          RecordPaymentResponse.$_createMessage);
  static RecordPaymentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Invoice get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Invoice value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Invoice ensureData() => $_ensure(0);
}

/// SearchInvoicesResponse returns invoices matching search criteria.
class SearchInvoicesResponse extends $pb.GeneratedMessage {
  factory SearchInvoicesResponse({
    $core.Iterable<Invoice>? data,
  }) {
    final result = SearchInvoicesResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchInvoicesResponse._();

  factory SearchInvoicesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchInvoicesResponse()..mergeFromBuffer(data, registry);
  factory SearchInvoicesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchInvoicesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchInvoicesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: SearchInvoicesResponse.$_createMessage)
    ..pPM<Invoice>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Invoice.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchInvoicesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchInvoicesResponse copyWith(
          void Function(SearchInvoicesResponse) updates) =>
      super.copyWith((message) => updates(message as SearchInvoicesResponse))
          as SearchInvoicesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchInvoicesResponse() / SearchInvoicesResponse.new instead')
  static SearchInvoicesResponse create() => SearchInvoicesResponse._();
  static $pb.GeneratedMessage $_createMessage() => SearchInvoicesResponse._();
  @$core.override
  SearchInvoicesResponse createEmptyInstance() => SearchInvoicesResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchInvoicesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchInvoicesResponse>(
          SearchInvoicesResponse.$_createMessage);
  static SearchInvoicesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Invoice> get data => $_getList(0);
}

/// GrantCreditRequest creates a prepaid credit grant.
class GrantCreditRequest extends $pb.GeneratedMessage {
  factory GrantCreditRequest({
    $core.String? id,
    $core.String? profileId,
    $core.String? name,
    $2.Money? amount,
    $core.String? currency,
    $0.Timestamp? expiresAt,
    $core.int? priority,
    $1.Struct? data,
  }) {
    final result = GrantCreditRequest._();
    if (id != null) result.id = id;
    if (profileId != null) result.profileId = profileId;
    if (name != null) result.name = name;
    if (amount != null) result.amount = amount;
    if (currency != null) result.currency = currency;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (priority != null) result.priority = priority;
    if (data != null) result.data = data;
    return result;
  }

  GrantCreditRequest._();

  factory GrantCreditRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GrantCreditRequest()..mergeFromBuffer(data, registry);
  factory GrantCreditRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GrantCreditRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GrantCreditRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GrantCreditRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'profileId')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOM<$2.Money>(4, _omitFieldNames ? '' : 'amount',
        subBuilder: $2.Money.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aI(7, _omitFieldNames ? '' : 'priority')
    ..aOM<$1.Struct>(8, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GrantCreditRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GrantCreditRequest copyWith(void Function(GrantCreditRequest) updates) =>
      super.copyWith((message) => updates(message as GrantCreditRequest))
          as GrantCreditRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GrantCreditRequest() / GrantCreditRequest.new instead')
  static GrantCreditRequest create() => GrantCreditRequest._();
  static $pb.GeneratedMessage $_createMessage() => GrantCreditRequest._();
  @$core.override
  GrantCreditRequest createEmptyInstance() => GrantCreditRequest._();
  @$core.pragma('dart2js:noInline')
  static GrantCreditRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GrantCreditRequest>(
          GrantCreditRequest.$_createMessage);
  static GrantCreditRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get profileId => $_getSZ(1);
  @$pb.TagNumber(2)
  set profileId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProfileId() => $_has(1);
  @$pb.TagNumber(2)
  void clearProfileId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $2.Money get amount => $_getN(3);
  @$pb.TagNumber(4)
  set amount($2.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmount() => $_clearField(4);
  @$pb.TagNumber(4)
  $2.Money ensureAmount() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Timestamp get expiresAt => $_getN(5);
  @$pb.TagNumber(6)
  set expiresAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasExpiresAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearExpiresAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureExpiresAt() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.int get priority => $_getIZ(6);
  @$pb.TagNumber(7)
  set priority($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPriority() => $_has(6);
  @$pb.TagNumber(7)
  void clearPriority() => $_clearField(7);

  @$pb.TagNumber(8)
  $1.Struct get data => $_getN(7);
  @$pb.TagNumber(8)
  set data($1.Struct value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasData() => $_has(7);
  @$pb.TagNumber(8)
  void clearData() => $_clearField(8);
  @$pb.TagNumber(8)
  $1.Struct ensureData() => $_ensure(7);
}

/// GrantCreditResponse returns the created credit grant.
class GrantCreditResponse extends $pb.GeneratedMessage {
  factory GrantCreditResponse({
    CreditGrant? data,
  }) {
    final result = GrantCreditResponse._();
    if (data != null) result.data = data;
    return result;
  }

  GrantCreditResponse._();

  factory GrantCreditResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GrantCreditResponse()..mergeFromBuffer(data, registry);
  factory GrantCreditResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GrantCreditResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GrantCreditResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GrantCreditResponse.$_createMessage)
    ..aOM<CreditGrant>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CreditGrant.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GrantCreditResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GrantCreditResponse copyWith(void Function(GrantCreditResponse) updates) =>
      super.copyWith((message) => updates(message as GrantCreditResponse))
          as GrantCreditResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GrantCreditResponse() / GrantCreditResponse.new instead')
  static GrantCreditResponse create() => GrantCreditResponse._();
  static $pb.GeneratedMessage $_createMessage() => GrantCreditResponse._();
  @$core.override
  GrantCreditResponse createEmptyInstance() => GrantCreditResponse._();
  @$core.pragma('dart2js:noInline')
  static GrantCreditResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GrantCreditResponse>(
          GrantCreditResponse.$_createMessage);
  static GrantCreditResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CreditGrant get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CreditGrant value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CreditGrant ensureData() => $_ensure(0);
}

/// GetCreditBalanceRequest retrieves a customer's credit balance.
class GetCreditBalanceRequest extends $pb.GeneratedMessage {
  factory GetCreditBalanceRequest({
    $core.String? profileId,
    $core.String? currency,
  }) {
    final result = GetCreditBalanceRequest._();
    if (profileId != null) result.profileId = profileId;
    if (currency != null) result.currency = currency;
    return result;
  }

  GetCreditBalanceRequest._();

  factory GetCreditBalanceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCreditBalanceRequest()..mergeFromBuffer(data, registry);
  factory GetCreditBalanceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCreditBalanceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCreditBalanceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetCreditBalanceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'profileId')
    ..aOS(2, _omitFieldNames ? '' : 'currency')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCreditBalanceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCreditBalanceRequest copyWith(
          void Function(GetCreditBalanceRequest) updates) =>
      super.copyWith((message) => updates(message as GetCreditBalanceRequest))
          as GetCreditBalanceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCreditBalanceRequest() / GetCreditBalanceRequest.new instead')
  static GetCreditBalanceRequest create() => GetCreditBalanceRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetCreditBalanceRequest._();
  @$core.override
  GetCreditBalanceRequest createEmptyInstance() => GetCreditBalanceRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCreditBalanceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCreditBalanceRequest>(
          GetCreditBalanceRequest.$_createMessage);
  static GetCreditBalanceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get profileId => $_getSZ(0);
  @$pb.TagNumber(1)
  set profileId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProfileId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProfileId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get currency => $_getSZ(1);
  @$pb.TagNumber(2)
  set currency($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCurrency() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrency() => $_clearField(2);
}

/// GetCreditBalanceResponse returns the credit balance.
class GetCreditBalanceResponse extends $pb.GeneratedMessage {
  factory GetCreditBalanceResponse({
    $2.Money? balance,
  }) {
    final result = GetCreditBalanceResponse._();
    if (balance != null) result.balance = balance;
    return result;
  }

  GetCreditBalanceResponse._();

  factory GetCreditBalanceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCreditBalanceResponse()..mergeFromBuffer(data, registry);
  factory GetCreditBalanceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCreditBalanceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCreditBalanceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: GetCreditBalanceResponse.$_createMessage)
    ..aOM<$2.Money>(1, _omitFieldNames ? '' : 'balance',
        subBuilder: $2.Money.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCreditBalanceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCreditBalanceResponse copyWith(
          void Function(GetCreditBalanceResponse) updates) =>
      super.copyWith((message) => updates(message as GetCreditBalanceResponse))
          as GetCreditBalanceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCreditBalanceResponse() / GetCreditBalanceResponse.new instead')
  static GetCreditBalanceResponse create() => GetCreditBalanceResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetCreditBalanceResponse._();
  @$core.override
  GetCreditBalanceResponse createEmptyInstance() =>
      GetCreditBalanceResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCreditBalanceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCreditBalanceResponse>(
          GetCreditBalanceResponse.$_createMessage);
  static GetCreditBalanceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $2.Money get balance => $_getN(0);
  @$pb.TagNumber(1)
  set balance($2.Money value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBalance() => $_has(0);
  @$pb.TagNumber(1)
  void clearBalance() => $_clearField(1);
  @$pb.TagNumber(1)
  $2.Money ensureBalance() => $_ensure(0);
}

/// CreateDiscountRequest creates a new discount rule.
class CreateDiscountRequest extends $pb.GeneratedMessage {
  factory CreateDiscountRequest({
    $core.String? id,
    $core.String? name,
    DiscountType? discountType,
    $core.double? value,
    $core.String? currency,
    $1.Struct? applicableTo,
    $0.Timestamp? startAt,
    $0.Timestamp? endAt,
    $core.int? maxApplications,
    $1.Struct? data,
  }) {
    final result = CreateDiscountRequest._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (discountType != null) result.discountType = discountType;
    if (value != null) result.value = value;
    if (currency != null) result.currency = currency;
    if (applicableTo != null) result.applicableTo = applicableTo;
    if (startAt != null) result.startAt = startAt;
    if (endAt != null) result.endAt = endAt;
    if (maxApplications != null) result.maxApplications = maxApplications;
    if (data != null) result.data = data;
    return result;
  }

  CreateDiscountRequest._();

  factory CreateDiscountRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDiscountRequest()..mergeFromBuffer(data, registry);
  factory CreateDiscountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDiscountRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDiscountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateDiscountRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aE<DiscountType>(3, _omitFieldNames ? '' : 'discountType',
        enumValues: DiscountType.values)
    ..aD(4, _omitFieldNames ? '' : 'value')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOM<$1.Struct>(6, _omitFieldNames ? '' : 'applicableTo',
        subBuilder: $1.Struct.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'startAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'endAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aI(9, _omitFieldNames ? '' : 'maxApplications')
    ..aOM<$1.Struct>(10, _omitFieldNames ? '' : 'data',
        subBuilder: $1.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDiscountRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDiscountRequest copyWith(
          void Function(CreateDiscountRequest) updates) =>
      super.copyWith((message) => updates(message as CreateDiscountRequest))
          as CreateDiscountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDiscountRequest() / CreateDiscountRequest.new instead')
  static CreateDiscountRequest create() => CreateDiscountRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateDiscountRequest._();
  @$core.override
  CreateDiscountRequest createEmptyInstance() => CreateDiscountRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateDiscountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDiscountRequest>(
          CreateDiscountRequest.$_createMessage);
  static CreateDiscountRequest? _defaultInstance;

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
  DiscountType get discountType => $_getN(2);
  @$pb.TagNumber(3)
  set discountType(DiscountType value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDiscountType() => $_has(2);
  @$pb.TagNumber(3)
  void clearDiscountType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.double get value => $_getN(3);
  @$pb.TagNumber(4)
  set value($core.double value) => $_setDouble(3, value);
  @$pb.TagNumber(4)
  $core.bool hasValue() => $_has(3);
  @$pb.TagNumber(4)
  void clearValue() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => $_clearField(5);

  @$pb.TagNumber(6)
  $1.Struct get applicableTo => $_getN(5);
  @$pb.TagNumber(6)
  set applicableTo($1.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasApplicableTo() => $_has(5);
  @$pb.TagNumber(6)
  void clearApplicableTo() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Struct ensureApplicableTo() => $_ensure(5);

  @$pb.TagNumber(7)
  $0.Timestamp get startAt => $_getN(6);
  @$pb.TagNumber(7)
  set startAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasStartAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearStartAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureStartAt() => $_ensure(6);

  @$pb.TagNumber(8)
  $0.Timestamp get endAt => $_getN(7);
  @$pb.TagNumber(8)
  set endAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasEndAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearEndAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureEndAt() => $_ensure(7);

  @$pb.TagNumber(9)
  $core.int get maxApplications => $_getIZ(8);
  @$pb.TagNumber(9)
  set maxApplications($core.int value) => $_setSignedInt32(8, value);
  @$pb.TagNumber(9)
  $core.bool hasMaxApplications() => $_has(8);
  @$pb.TagNumber(9)
  void clearMaxApplications() => $_clearField(9);

  @$pb.TagNumber(10)
  $1.Struct get data => $_getN(9);
  @$pb.TagNumber(10)
  set data($1.Struct value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasData() => $_has(9);
  @$pb.TagNumber(10)
  void clearData() => $_clearField(10);
  @$pb.TagNumber(10)
  $1.Struct ensureData() => $_ensure(9);
}

/// CreateDiscountResponse returns the created discount.
class CreateDiscountResponse extends $pb.GeneratedMessage {
  factory CreateDiscountResponse({
    Discount? data,
  }) {
    final result = CreateDiscountResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateDiscountResponse._();

  factory CreateDiscountResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDiscountResponse()..mergeFromBuffer(data, registry);
  factory CreateDiscountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDiscountResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDiscountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: CreateDiscountResponse.$_createMessage)
    ..aOM<Discount>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Discount.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDiscountResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDiscountResponse copyWith(
          void Function(CreateDiscountResponse) updates) =>
      super.copyWith((message) => updates(message as CreateDiscountResponse))
          as CreateDiscountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDiscountResponse() / CreateDiscountResponse.new instead')
  static CreateDiscountResponse create() => CreateDiscountResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateDiscountResponse._();
  @$core.override
  CreateDiscountResponse createEmptyInstance() => CreateDiscountResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateDiscountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDiscountResponse>(
          CreateDiscountResponse.$_createMessage);
  static CreateDiscountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Discount get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(Discount value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  Discount ensureData() => $_ensure(0);
}

/// SearchDiscountsResponse returns discounts matching search criteria.
class SearchDiscountsResponse extends $pb.GeneratedMessage {
  factory SearchDiscountsResponse({
    $core.Iterable<Discount>? data,
  }) {
    final result = SearchDiscountsResponse._();
    if (data != null) result.data.addAll(data);
    return result;
  }

  SearchDiscountsResponse._();

  factory SearchDiscountsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchDiscountsResponse()..mergeFromBuffer(data, registry);
  factory SearchDiscountsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchDiscountsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchDiscountsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'billing.v1'),
      createEmptyInstance: SearchDiscountsResponse.$_createMessage)
    ..pPM<Discount>(1, _omitFieldNames ? '' : 'data',
        subBuilder: Discount.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchDiscountsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchDiscountsResponse copyWith(
          void Function(SearchDiscountsResponse) updates) =>
      super.copyWith((message) => updates(message as SearchDiscountsResponse))
          as SearchDiscountsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchDiscountsResponse() / SearchDiscountsResponse.new instead')
  static SearchDiscountsResponse create() => SearchDiscountsResponse._();
  static $pb.GeneratedMessage $_createMessage() => SearchDiscountsResponse._();
  @$core.override
  SearchDiscountsResponse createEmptyInstance() => SearchDiscountsResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchDiscountsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchDiscountsResponse>(
          SearchDiscountsResponse.$_createMessage);
  static SearchDiscountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Discount> get data => $_getList(0);
}

/// BillingService provides usage-based billing, subscription management, and invoicing.
/// All RPCs require authentication via Bearer token.
class BillingServiceApi {
  final $pb.RpcClient _client;

  BillingServiceApi(this._client);

  /// CreateCatalogVersion creates a new catalog version with plans and pricing.
  $async.Future<CreateCatalogVersionResponse> createCatalogVersion(
          $pb.ClientContext? ctx, CreateCatalogVersionRequest request) =>
      _client.invoke<CreateCatalogVersionResponse>(ctx, 'BillingService',
          'CreateCatalogVersion', request, CreateCatalogVersionResponse());

  /// GetCatalogVersion retrieves a specific catalog version by ID.
  $async.Future<GetCatalogVersionResponse> getCatalogVersion(
          $pb.ClientContext? ctx, GetCatalogVersionRequest request) =>
      _client.invoke<GetCatalogVersionResponse>(ctx, 'BillingService',
          'GetCatalogVersion', request, GetCatalogVersionResponse());

  /// PublishCatalogVersion publishes a catalog version, making it available for subscriptions.
  $async.Future<PublishCatalogVersionResponse> publishCatalogVersion(
          $pb.ClientContext? ctx, PublishCatalogVersionRequest request) =>
      _client.invoke<PublishCatalogVersionResponse>(ctx, 'BillingService',
          'PublishCatalogVersion', request, PublishCatalogVersionResponse());

  /// SearchCatalogVersions searches for catalog versions.
  $async.Future<SearchCatalogVersionsResponse> searchCatalogVersions(
          $pb.ClientContext? ctx, $4.SearchRequest request) =>
      _client.invoke<SearchCatalogVersionsResponse>(ctx, 'BillingService',
          'SearchCatalogVersions', request, SearchCatalogVersionsResponse());

  /// CreatePlan creates a new plan within a catalog version.
  $async.Future<CreatePlanResponse> createPlan(
          $pb.ClientContext? ctx, CreatePlanRequest request) =>
      _client.invoke<CreatePlanResponse>(
          ctx, 'BillingService', 'CreatePlan', request, CreatePlanResponse());

  /// CreateComponent creates a new billable component within a plan.
  $async.Future<CreateComponentResponse> createComponent(
          $pb.ClientContext? ctx, CreateComponentRequest request) =>
      _client.invoke<CreateComponentResponse>(ctx, 'BillingService',
          'CreateComponent', request, CreateComponentResponse());

  /// CreateTier creates a new pricing tier within a component.
  $async.Future<CreateTierResponse> createTier(
          $pb.ClientContext? ctx, CreateTierRequest request) =>
      _client.invoke<CreateTierResponse>(
          ctx, 'BillingService', 'CreateTier', request, CreateTierResponse());

  /// CreateSubscription creates a new subscription for a customer.
  $async.Future<CreateSubscriptionResponse> createSubscription(
          $pb.ClientContext? ctx, CreateSubscriptionRequest request) =>
      _client.invoke<CreateSubscriptionResponse>(ctx, 'BillingService',
          'CreateSubscription', request, CreateSubscriptionResponse());

  /// GetSubscription retrieves a subscription by ID.
  $async.Future<GetSubscriptionResponse> getSubscription(
          $pb.ClientContext? ctx, GetSubscriptionRequest request) =>
      _client.invoke<GetSubscriptionResponse>(ctx, 'BillingService',
          'GetSubscription', request, GetSubscriptionResponse());

  /// CancelSubscription cancels an active subscription.
  $async.Future<CancelSubscriptionResponse> cancelSubscription(
          $pb.ClientContext? ctx, CancelSubscriptionRequest request) =>
      _client.invoke<CancelSubscriptionResponse>(ctx, 'BillingService',
          'CancelSubscription', request, CancelSubscriptionResponse());

  /// ListSubscriptions lists active subscriptions for a customer.
  $async.Future<ListSubscriptionsResponse> listSubscriptions(
          $pb.ClientContext? ctx, ListSubscriptionsRequest request) =>
      _client.invoke<ListSubscriptionsResponse>(ctx, 'BillingService',
          'ListSubscriptions', request, ListSubscriptionsResponse());

  /// IngestUsageEvent ingests a single usage event for metering.
  $async.Future<IngestUsageEventResponse> ingestUsageEvent(
          $pb.ClientContext? ctx, IngestUsageEventRequest request) =>
      _client.invoke<IngestUsageEventResponse>(ctx, 'BillingService',
          'IngestUsageEvent', request, IngestUsageEventResponse());

  /// SearchUsageEvents searches for usage events.
  $async.Future<SearchUsageEventsResponse> searchUsageEvents(
          $pb.ClientContext? ctx, $4.SearchRequest request) =>
      _client.invoke<SearchUsageEventsResponse>(ctx, 'BillingService',
          'SearchUsageEvents', request, SearchUsageEventsResponse());

  /// RunBilling executes the billing pipeline for a subscription and period.
  $async.Future<RunBillingResponse> runBilling(
          $pb.ClientContext? ctx, RunBillingRequest request) =>
      _client.invoke<RunBillingResponse>(
          ctx, 'BillingService', 'RunBilling', request, RunBillingResponse());

  /// GetBillingRun retrieves a billing run by ID.
  $async.Future<GetBillingRunResponse> getBillingRun(
          $pb.ClientContext? ctx, GetBillingRunRequest request) =>
      _client.invoke<GetBillingRunResponse>(ctx, 'BillingService',
          'GetBillingRun', request, GetBillingRunResponse());

  /// GetInvoice retrieves an invoice by ID with all line items.
  $async.Future<GetInvoiceResponse> getInvoice(
          $pb.ClientContext? ctx, GetInvoiceRequest request) =>
      _client.invoke<GetInvoiceResponse>(
          ctx, 'BillingService', 'GetInvoice', request, GetInvoiceResponse());

  /// IssueInvoice issues a draft invoice to the customer.
  $async.Future<IssueInvoiceResponse> issueInvoice(
          $pb.ClientContext? ctx, IssueInvoiceRequest request) =>
      _client.invoke<IssueInvoiceResponse>(ctx, 'BillingService',
          'IssueInvoice', request, IssueInvoiceResponse());

  /// VoidInvoice voids an invoice.
  $async.Future<VoidInvoiceResponse> voidInvoice(
          $pb.ClientContext? ctx, VoidInvoiceRequest request) =>
      _client.invoke<VoidInvoiceResponse>(
          ctx, 'BillingService', 'VoidInvoice', request, VoidInvoiceResponse());

  /// RecordPayment records payment against an invoice.
  $async.Future<RecordPaymentResponse> recordPayment(
          $pb.ClientContext? ctx, RecordPaymentRequest request) =>
      _client.invoke<RecordPaymentResponse>(ctx, 'BillingService',
          'RecordPayment', request, RecordPaymentResponse());

  /// SearchInvoices searches for invoices.
  $async.Future<SearchInvoicesResponse> searchInvoices(
          $pb.ClientContext? ctx, $4.SearchRequest request) =>
      _client.invoke<SearchInvoicesResponse>(ctx, 'BillingService',
          'SearchInvoices', request, SearchInvoicesResponse());

  /// GrantCredit creates a prepaid credit grant for a customer.
  $async.Future<GrantCreditResponse> grantCredit(
          $pb.ClientContext? ctx, GrantCreditRequest request) =>
      _client.invoke<GrantCreditResponse>(
          ctx, 'BillingService', 'GrantCredit', request, GrantCreditResponse());

  /// GetCreditBalance retrieves a customer's credit balance.
  $async.Future<GetCreditBalanceResponse> getCreditBalance(
          $pb.ClientContext? ctx, GetCreditBalanceRequest request) =>
      _client.invoke<GetCreditBalanceResponse>(ctx, 'BillingService',
          'GetCreditBalance', request, GetCreditBalanceResponse());

  /// CreateDiscount creates a new discount rule.
  $async.Future<CreateDiscountResponse> createDiscount(
          $pb.ClientContext? ctx, CreateDiscountRequest request) =>
      _client.invoke<CreateDiscountResponse>(ctx, 'BillingService',
          'CreateDiscount', request, CreateDiscountResponse());

  /// SearchDiscounts searches for discount rules.
  $async.Future<SearchDiscountsResponse> searchDiscounts(
          $pb.ClientContext? ctx, $4.SearchRequest request) =>
      _client.invoke<SearchDiscountsResponse>(ctx, 'BillingService',
          'SearchDiscounts', request, SearchDiscountsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
