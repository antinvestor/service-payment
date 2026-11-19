// This is a generated file - do not edit.
//
// Generated from collection/v1/collection.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../common/v1/money.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// CollectionResult is returned when starting collection for an invoice or subscription.
class CollectionResult extends $pb.GeneratedMessage {
  factory CollectionResult({
    $core.String? pageUrl,
    $core.String? sessionRef,
    $core.String? invoiceId,
    $core.String? subscriptionId,
    $core.bool? alreadyComplete,
  }) {
    final result = CollectionResult._();
    if (pageUrl != null) result.pageUrl = pageUrl;
    if (sessionRef != null) result.sessionRef = sessionRef;
    if (invoiceId != null) result.invoiceId = invoiceId;
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (alreadyComplete != null) result.alreadyComplete = alreadyComplete;
    return result;
  }

  CollectionResult._();

  factory CollectionResult.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CollectionResult()..mergeFromBuffer(data, registry);
  factory CollectionResult.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CollectionResult()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CollectionResult',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: CollectionResult.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'pageUrl')
    ..aOS(2, _omitFieldNames ? '' : 'sessionRef')
    ..aOS(3, _omitFieldNames ? '' : 'invoiceId')
    ..aOS(4, _omitFieldNames ? '' : 'subscriptionId')
    ..aOB(5, _omitFieldNames ? '' : 'alreadyComplete')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CollectionResult clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CollectionResult copyWith(void Function(CollectionResult) updates) =>
      super.copyWith((message) => updates(message as CollectionResult))
          as CollectionResult;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CollectionResult() / CollectionResult.new instead')
  static CollectionResult create() => CollectionResult._();
  static $pb.GeneratedMessage $_createMessage() => CollectionResult._();
  @$core.override
  CollectionResult createEmptyInstance() => CollectionResult._();
  @$core.pragma('dart2js:noInline')
  static CollectionResult getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CollectionResult>(
          CollectionResult.$_createMessage);
  static CollectionResult? _defaultInstance;

  /// Hosted checkout page URL. Empty when no payment is required (free / zero total).
  @$pb.TagNumber(1)
  $core.String get pageUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set pageUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPageUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearPageUrl() => $_clearField(1);

  /// Checkout session reference. Empty when no payment is required.
  @$pb.TagNumber(2)
  $core.String get sessionRef => $_getSZ(1);
  @$pb.TagNumber(2)
  set sessionRef($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSessionRef() => $_has(1);
  @$pb.TagNumber(2)
  void clearSessionRef() => $_clearField(2);

  /// Invoice being collected (if any).
  @$pb.TagNumber(3)
  $core.String get invoiceId => $_getSZ(2);
  @$pb.TagNumber(3)
  set invoiceId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasInvoiceId() => $_has(2);
  @$pb.TagNumber(3)
  void clearInvoiceId() => $_clearField(3);

  /// Subscription created or already linked (if any).
  @$pb.TagNumber(4)
  $core.String get subscriptionId => $_getSZ(3);
  @$pb.TagNumber(4)
  set subscriptionId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSubscriptionId() => $_has(3);
  @$pb.TagNumber(4)
  void clearSubscriptionId() => $_clearField(4);

  /// True when payment was not needed and the resource is already settled/active.
  @$pb.TagNumber(5)
  $core.bool get alreadyComplete => $_getBF(4);
  @$pb.TagNumber(5)
  set alreadyComplete($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAlreadyComplete() => $_has(4);
  @$pb.TagNumber(5)
  void clearAlreadyComplete() => $_clearField(5);
}

/// CollectPaymentRequest starts hosted checkout for an issued invoice.
class CollectPaymentRequest extends $pb.GeneratedMessage {
  factory CollectPaymentRequest({
    $core.String? invoiceId,
    $core.String? returnUrl,
    $core.Iterable<$core.String>? methods,
  }) {
    final result = CollectPaymentRequest._();
    if (invoiceId != null) result.invoiceId = invoiceId;
    if (returnUrl != null) result.returnUrl = returnUrl;
    if (methods != null) result.methods.addAll(methods);
    return result;
  }

  CollectPaymentRequest._();

  factory CollectPaymentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CollectPaymentRequest()..mergeFromBuffer(data, registry);
  factory CollectPaymentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CollectPaymentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CollectPaymentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: CollectPaymentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invoiceId')
    ..aOS(2, _omitFieldNames ? '' : 'returnUrl')
    ..pPS(3, _omitFieldNames ? '' : 'methods')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CollectPaymentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CollectPaymentRequest copyWith(
          void Function(CollectPaymentRequest) updates) =>
      super.copyWith((message) => updates(message as CollectPaymentRequest))
          as CollectPaymentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CollectPaymentRequest() / CollectPaymentRequest.new instead')
  static CollectPaymentRequest create() => CollectPaymentRequest._();
  static $pb.GeneratedMessage $_createMessage() => CollectPaymentRequest._();
  @$core.override
  CollectPaymentRequest createEmptyInstance() => CollectPaymentRequest._();
  @$core.pragma('dart2js:noInline')
  static CollectPaymentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CollectPaymentRequest>(
          CollectPaymentRequest.$_createMessage);
  static CollectPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get invoiceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invoiceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvoiceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvoiceId() => $_clearField(1);

  /// Optional override for the post-payment browser return URL.
  @$pb.TagNumber(2)
  $core.String get returnUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set returnUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReturnUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearReturnUrl() => $_clearField(2);

  /// Optional payment method keys (e.g. mpesa, mtn_momo). Empty = all methods
  /// available for the invoice currency on the hosted checkout page.
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get methods => $_getList(2);
}

class CollectPaymentResponse extends $pb.GeneratedMessage {
  factory CollectPaymentResponse({
    CollectionResult? data,
  }) {
    final result = CollectPaymentResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CollectPaymentResponse._();

  factory CollectPaymentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CollectPaymentResponse()..mergeFromBuffer(data, registry);
  factory CollectPaymentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CollectPaymentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CollectPaymentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: CollectPaymentResponse.$_createMessage)
    ..aOM<CollectionResult>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CollectionResult.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CollectPaymentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CollectPaymentResponse copyWith(
          void Function(CollectPaymentResponse) updates) =>
      super.copyWith((message) => updates(message as CollectPaymentResponse))
          as CollectPaymentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CollectPaymentResponse() / CollectPaymentResponse.new instead')
  static CollectPaymentResponse create() => CollectPaymentResponse._();
  static $pb.GeneratedMessage $_createMessage() => CollectPaymentResponse._();
  @$core.override
  CollectPaymentResponse createEmptyInstance() => CollectPaymentResponse._();
  @$core.pragma('dart2js:noInline')
  static CollectPaymentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CollectPaymentResponse>(
          CollectPaymentResponse.$_createMessage);
  static CollectPaymentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CollectionResult get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CollectionResult value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CollectionResult ensureData() => $_ensure(0);
}

/// StartSubscriptionRequest creates a subscription and, when the plan has an
/// upfront flat fee, a checkout session for first payment.
class StartSubscriptionRequest extends $pb.GeneratedMessage {
  factory StartSubscriptionRequest({
    $core.String? profileId,
    $core.String? planId,
    $core.String? catalogVersionId,
    $core.String? currency,
    $core.String? returnUrl,
    $core.String? payerDisplayName,
    $core.Iterable<$core.String>? methods,
    $core.String? externalEntityId,
    $core.String? externalEntityType,
    $core.String? integrationRouteId,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
  }) {
    final result = StartSubscriptionRequest._();
    if (profileId != null) result.profileId = profileId;
    if (planId != null) result.planId = planId;
    if (catalogVersionId != null) result.catalogVersionId = catalogVersionId;
    if (currency != null) result.currency = currency;
    if (returnUrl != null) result.returnUrl = returnUrl;
    if (payerDisplayName != null) result.payerDisplayName = payerDisplayName;
    if (methods != null) result.methods.addAll(methods);
    if (externalEntityId != null) result.externalEntityId = externalEntityId;
    if (externalEntityType != null)
      result.externalEntityType = externalEntityType;
    if (integrationRouteId != null)
      result.integrationRouteId = integrationRouteId;
    if (metadata != null) result.metadata.addEntries(metadata);
    return result;
  }

  StartSubscriptionRequest._();

  factory StartSubscriptionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartSubscriptionRequest()..mergeFromBuffer(data, registry);
  factory StartSubscriptionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartSubscriptionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartSubscriptionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: StartSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'profileId')
    ..aOS(2, _omitFieldNames ? '' : 'planId')
    ..aOS(3, _omitFieldNames ? '' : 'catalogVersionId')
    ..aOS(4, _omitFieldNames ? '' : 'currency')
    ..aOS(5, _omitFieldNames ? '' : 'returnUrl')
    ..aOS(6, _omitFieldNames ? '' : 'payerDisplayName')
    ..pPS(7, _omitFieldNames ? '' : 'methods')
    ..aOS(8, _omitFieldNames ? '' : 'externalEntityId')
    ..aOS(9, _omitFieldNames ? '' : 'externalEntityType')
    ..aOS(10, _omitFieldNames ? '' : 'integrationRouteId')
    ..m<$core.String, $core.String>(11, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'StartSubscriptionRequest.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('collection.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartSubscriptionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartSubscriptionRequest copyWith(
          void Function(StartSubscriptionRequest) updates) =>
      super.copyWith((message) => updates(message as StartSubscriptionRequest))
          as StartSubscriptionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use StartSubscriptionRequest() / StartSubscriptionRequest.new instead')
  static StartSubscriptionRequest create() => StartSubscriptionRequest._();
  static $pb.GeneratedMessage $_createMessage() => StartSubscriptionRequest._();
  @$core.override
  StartSubscriptionRequest createEmptyInstance() =>
      StartSubscriptionRequest._();
  @$core.pragma('dart2js:noInline')
  static StartSubscriptionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartSubscriptionRequest>(
          StartSubscriptionRequest.$_createMessage);
  static StartSubscriptionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get profileId => $_getSZ(0);
  @$pb.TagNumber(1)
  set profileId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProfileId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProfileId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get planId => $_getSZ(1);
  @$pb.TagNumber(2)
  set planId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPlanId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlanId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get catalogVersionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set catalogVersionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCatalogVersionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCatalogVersionId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(3);
  @$pb.TagNumber(4)
  set currency($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(3);
  @$pb.TagNumber(4)
  void clearCurrency() => $_clearField(4);

  /// Optional override for the post-payment browser return URL.
  @$pb.TagNumber(5)
  $core.String get returnUrl => $_getSZ(4);
  @$pb.TagNumber(5)
  set returnUrl($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasReturnUrl() => $_has(4);
  @$pb.TagNumber(5)
  void clearReturnUrl() => $_clearField(5);

  /// Optional payer display name used to prefill the checkout page.
  @$pb.TagNumber(6)
  $core.String get payerDisplayName => $_getSZ(5);
  @$pb.TagNumber(6)
  set payerDisplayName($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPayerDisplayName() => $_has(5);
  @$pb.TagNumber(6)
  void clearPayerDisplayName() => $_clearField(6);

  /// Optional payment method keys shown on checkout (empty = all for currency).
  @$pb.TagNumber(7)
  $pb.PbList<$core.String> get methods => $_getList(6);

  /// External-entity integration (product apps / entitlement systems).
  /// Mirrored onto Subscription.data and included in lifecycle queue events —
  /// same role as payment route destinations for Send/Receive.
  ///
  /// external_entity_id identifies the product resource this subscription
  /// entitles (workspace, membership, seat, device, …).
  @$pb.TagNumber(8)
  $core.String get externalEntityId => $_getSZ(7);
  @$pb.TagNumber(8)
  set externalEntityId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasExternalEntityId() => $_has(7);
  @$pb.TagNumber(8)
  void clearExternalEntityId() => $_clearField(8);

  /// external_entity_type classifies the entity (e.g. "workspace", "membership").
  @$pb.TagNumber(9)
  $core.String get externalEntityType => $_getSZ(8);
  @$pb.TagNumber(9)
  set externalEntityType($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasExternalEntityType() => $_has(8);
  @$pb.TagNumber(9)
  void clearExternalEntityType() => $_clearField(9);

  /// integration_route_id optionally pins lifecycle delivery to one
  /// IntegrationRoute row (otherwise partition routes + default topic apply).
  @$pb.TagNumber(10)
  $core.String get integrationRouteId => $_getSZ(9);
  @$pb.TagNumber(10)
  set integrationRouteId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasIntegrationRouteId() => $_has(9);
  @$pb.TagNumber(10)
  void clearIntegrationRouteId() => $_clearField(10);

  /// metadata is free-form product data stored on Subscription.data and
  /// forwarded on lifecycle events (keys reserved by billing are ignored).
  @$pb.TagNumber(11)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(10);
}

class StartSubscriptionResponse extends $pb.GeneratedMessage {
  factory StartSubscriptionResponse({
    CollectionResult? data,
  }) {
    final result = StartSubscriptionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  StartSubscriptionResponse._();

  factory StartSubscriptionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartSubscriptionResponse()..mergeFromBuffer(data, registry);
  factory StartSubscriptionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartSubscriptionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartSubscriptionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: StartSubscriptionResponse.$_createMessage)
    ..aOM<CollectionResult>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CollectionResult.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartSubscriptionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartSubscriptionResponse copyWith(
          void Function(StartSubscriptionResponse) updates) =>
      super.copyWith((message) => updates(message as StartSubscriptionResponse))
          as StartSubscriptionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use StartSubscriptionResponse() / StartSubscriptionResponse.new instead')
  static StartSubscriptionResponse create() => StartSubscriptionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      StartSubscriptionResponse._();
  @$core.override
  StartSubscriptionResponse createEmptyInstance() =>
      StartSubscriptionResponse._();
  @$core.pragma('dart2js:noInline')
  static StartSubscriptionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartSubscriptionResponse>(
          StartSubscriptionResponse.$_createMessage);
  static StartSubscriptionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CollectionResult get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CollectionResult value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CollectionResult ensureData() => $_ensure(0);
}

/// ConfirmPaymentRequest settles an invoice (and activates a pending subscription)
/// from a completed checkout session. Safe to call repeatedly (idempotent).
class ConfirmPaymentRequest extends $pb.GeneratedMessage {
  factory ConfirmPaymentRequest({
    $core.String? sessionRef,
  }) {
    final result = ConfirmPaymentRequest._();
    if (sessionRef != null) result.sessionRef = sessionRef;
    return result;
  }

  ConfirmPaymentRequest._();

  factory ConfirmPaymentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmPaymentRequest()..mergeFromBuffer(data, registry);
  factory ConfirmPaymentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmPaymentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfirmPaymentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: ConfirmPaymentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionRef')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmPaymentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmPaymentRequest copyWith(
          void Function(ConfirmPaymentRequest) updates) =>
      super.copyWith((message) => updates(message as ConfirmPaymentRequest))
          as ConfirmPaymentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ConfirmPaymentRequest() / ConfirmPaymentRequest.new instead')
  static ConfirmPaymentRequest create() => ConfirmPaymentRequest._();
  static $pb.GeneratedMessage $_createMessage() => ConfirmPaymentRequest._();
  @$core.override
  ConfirmPaymentRequest createEmptyInstance() => ConfirmPaymentRequest._();
  @$core.pragma('dart2js:noInline')
  static ConfirmPaymentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConfirmPaymentRequest>(
          ConfirmPaymentRequest.$_createMessage);
  static ConfirmPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionRef => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionRef($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionRef() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionRef() => $_clearField(1);
}

/// ConfirmPaymentResponse returns the settled invoice and optional subscription.
class ConfirmPaymentResponse extends $pb.GeneratedMessage {
  factory ConfirmPaymentResponse({
    $core.String? invoiceId,
    $core.String? invoiceState,
    $core.String? subscriptionId,
    $core.String? subscriptionState,
    $core.bool? paid,
  }) {
    final result = ConfirmPaymentResponse._();
    if (invoiceId != null) result.invoiceId = invoiceId;
    if (invoiceState != null) result.invoiceState = invoiceState;
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (subscriptionState != null) result.subscriptionState = subscriptionState;
    if (paid != null) result.paid = paid;
    return result;
  }

  ConfirmPaymentResponse._();

  factory ConfirmPaymentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmPaymentResponse()..mergeFromBuffer(data, registry);
  factory ConfirmPaymentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmPaymentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfirmPaymentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: ConfirmPaymentResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invoiceId')
    ..aOS(2, _omitFieldNames ? '' : 'invoiceState')
    ..aOS(3, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(4, _omitFieldNames ? '' : 'subscriptionState')
    ..aOB(5, _omitFieldNames ? '' : 'paid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmPaymentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmPaymentResponse copyWith(
          void Function(ConfirmPaymentResponse) updates) =>
      super.copyWith((message) => updates(message as ConfirmPaymentResponse))
          as ConfirmPaymentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ConfirmPaymentResponse() / ConfirmPaymentResponse.new instead')
  static ConfirmPaymentResponse create() => ConfirmPaymentResponse._();
  static $pb.GeneratedMessage $_createMessage() => ConfirmPaymentResponse._();
  @$core.override
  ConfirmPaymentResponse createEmptyInstance() => ConfirmPaymentResponse._();
  @$core.pragma('dart2js:noInline')
  static ConfirmPaymentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConfirmPaymentResponse>(
          ConfirmPaymentResponse.$_createMessage);
  static ConfirmPaymentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get invoiceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invoiceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvoiceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvoiceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get invoiceState => $_getSZ(1);
  @$pb.TagNumber(2)
  set invoiceState($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInvoiceState() => $_has(1);
  @$pb.TagNumber(2)
  void clearInvoiceState() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get subscriptionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set subscriptionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSubscriptionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSubscriptionId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get subscriptionState => $_getSZ(3);
  @$pb.TagNumber(4)
  set subscriptionState($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSubscriptionState() => $_has(3);
  @$pb.TagNumber(4)
  void clearSubscriptionState() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get paid => $_getBF(4);
  @$pb.TagNumber(5)
  set paid($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPaid() => $_has(4);
  @$pb.TagNumber(5)
  void clearPaid() => $_clearField(5);
}

/// CancelSubscriptionRequest cancels an active or pending subscription.
/// For PENDING subscriptions still awaiting first payment, any open signup
/// invoice is voided automatically.
class CancelSubscriptionRequest extends $pb.GeneratedMessage {
  factory CancelSubscriptionRequest({
    $core.String? subscriptionId,
  }) {
    final result = CancelSubscriptionRequest._();
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
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
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: CancelSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'subscriptionId')
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
  $core.String get subscriptionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set subscriptionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSubscriptionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSubscriptionId() => $_clearField(1);
}

class CancelSubscriptionResponse extends $pb.GeneratedMessage {
  factory CancelSubscriptionResponse({
    $core.String? subscriptionId,
    $core.String? subscriptionState,
    $core.String? voidedInvoiceId,
  }) {
    final result = CancelSubscriptionResponse._();
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (subscriptionState != null) result.subscriptionState = subscriptionState;
    if (voidedInvoiceId != null) result.voidedInvoiceId = voidedInvoiceId;
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
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: CancelSubscriptionResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(2, _omitFieldNames ? '' : 'subscriptionState')
    ..aOS(3, _omitFieldNames ? '' : 'voidedInvoiceId')
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
  $core.String get subscriptionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set subscriptionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSubscriptionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSubscriptionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get subscriptionState => $_getSZ(1);
  @$pb.TagNumber(2)
  set subscriptionState($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSubscriptionState() => $_has(1);
  @$pb.TagNumber(2)
  void clearSubscriptionState() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get voidedInvoiceId => $_getSZ(2);
  @$pb.TagNumber(3)
  set voidedInvoiceId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasVoidedInvoiceId() => $_has(2);
  @$pb.TagNumber(3)
  void clearVoidedInvoiceId() => $_clearField(3);
}

/// CollectionService is the merchant-facing simplified payment collection API.
/// SyncProviderSubscriptionRequest mirrors a provider-managed subscription
/// (Polar, Stripe) into billing. Integration workers call it from provider
/// webhooks; billing keeps the subscription row, lifecycle fan-out and
/// renewal bookkeeping in step with the provider.
class SyncProviderSubscriptionRequest extends $pb.GeneratedMessage {
  factory SyncProviderSubscriptionRequest({
    $core.String? provider,
    $core.String? event,
    $core.String? externalSubscriptionId,
    $core.String? externalCustomerId,
    $core.String? externalProductId,
    $core.String? status,
    $core.bool? cancelAtPeriodEnd,
    $core.String? currentPeriodStart,
    $core.String? currentPeriodEnd,
    $core.String? subscriptionId,
    $core.String? profileId,
    $core.String? planId,
    $core.String? externalOrderId,
    $0.Money? amount,
    $core.String? occurredAt,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
  }) {
    final result = SyncProviderSubscriptionRequest._();
    if (provider != null) result.provider = provider;
    if (event != null) result.event = event;
    if (externalSubscriptionId != null)
      result.externalSubscriptionId = externalSubscriptionId;
    if (externalCustomerId != null)
      result.externalCustomerId = externalCustomerId;
    if (externalProductId != null) result.externalProductId = externalProductId;
    if (status != null) result.status = status;
    if (cancelAtPeriodEnd != null) result.cancelAtPeriodEnd = cancelAtPeriodEnd;
    if (currentPeriodStart != null)
      result.currentPeriodStart = currentPeriodStart;
    if (currentPeriodEnd != null) result.currentPeriodEnd = currentPeriodEnd;
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (profileId != null) result.profileId = profileId;
    if (planId != null) result.planId = planId;
    if (externalOrderId != null) result.externalOrderId = externalOrderId;
    if (amount != null) result.amount = amount;
    if (occurredAt != null) result.occurredAt = occurredAt;
    if (metadata != null) result.metadata.addEntries(metadata);
    return result;
  }

  SyncProviderSubscriptionRequest._();

  factory SyncProviderSubscriptionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncProviderSubscriptionRequest()..mergeFromBuffer(data, registry);
  factory SyncProviderSubscriptionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncProviderSubscriptionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SyncProviderSubscriptionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: SyncProviderSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'provider')
    ..aOS(2, _omitFieldNames ? '' : 'event')
    ..aOS(3, _omitFieldNames ? '' : 'externalSubscriptionId')
    ..aOS(4, _omitFieldNames ? '' : 'externalCustomerId')
    ..aOS(5, _omitFieldNames ? '' : 'externalProductId')
    ..aOS(6, _omitFieldNames ? '' : 'status')
    ..aOB(7, _omitFieldNames ? '' : 'cancelAtPeriodEnd')
    ..aOS(8, _omitFieldNames ? '' : 'currentPeriodStart')
    ..aOS(9, _omitFieldNames ? '' : 'currentPeriodEnd')
    ..aOS(10, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(11, _omitFieldNames ? '' : 'profileId')
    ..aOS(12, _omitFieldNames ? '' : 'planId')
    ..aOS(13, _omitFieldNames ? '' : 'externalOrderId')
    ..aOM<$0.Money>(14, _omitFieldNames ? '' : 'amount',
        subBuilder: $0.Money.$_createMessage)
    ..aOS(15, _omitFieldNames ? '' : 'occurredAt')
    ..m<$core.String, $core.String>(16, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'SyncProviderSubscriptionRequest.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('collection.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncProviderSubscriptionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncProviderSubscriptionRequest copyWith(
          void Function(SyncProviderSubscriptionRequest) updates) =>
      super.copyWith(
              (message) => updates(message as SyncProviderSubscriptionRequest))
          as SyncProviderSubscriptionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SyncProviderSubscriptionRequest() / SyncProviderSubscriptionRequest.new instead')
  static SyncProviderSubscriptionRequest create() =>
      SyncProviderSubscriptionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      SyncProviderSubscriptionRequest._();
  @$core.override
  SyncProviderSubscriptionRequest createEmptyInstance() =>
      SyncProviderSubscriptionRequest._();
  @$core.pragma('dart2js:noInline')
  static SyncProviderSubscriptionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SyncProviderSubscriptionRequest>(
          SyncProviderSubscriptionRequest.$_createMessage);
  static SyncProviderSubscriptionRequest? _defaultInstance;

  /// Payment route of the provider (polar, stripe).
  @$pb.TagNumber(1)
  $core.String get provider => $_getSZ(0);
  @$pb.TagNumber(1)
  set provider($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearProvider() => $_clearField(1);

  /// Normalized event: subscription.updated, subscription.billed,
  /// subscription.payment_failed.
  @$pb.TagNumber(2)
  $core.String get event => $_getSZ(1);
  @$pb.TagNumber(2)
  set event($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEvent() => $_has(1);
  @$pb.TagNumber(2)
  void clearEvent() => $_clearField(2);

  /// Provider ids.
  @$pb.TagNumber(3)
  $core.String get externalSubscriptionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set externalSubscriptionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExternalSubscriptionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearExternalSubscriptionId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get externalCustomerId => $_getSZ(3);
  @$pb.TagNumber(4)
  set externalCustomerId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasExternalCustomerId() => $_has(3);
  @$pb.TagNumber(4)
  void clearExternalCustomerId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get externalProductId => $_getSZ(4);
  @$pb.TagNumber(5)
  set externalProductId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasExternalProductId() => $_has(4);
  @$pb.TagNumber(5)
  void clearExternalProductId() => $_clearField(5);

  /// Normalized status: active, trialing, past_due, incomplete,
  /// incomplete_expired, canceled, revoked, unpaid, paused.
  @$pb.TagNumber(6)
  $core.String get status => $_getSZ(5);
  @$pb.TagNumber(6)
  set status($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get cancelAtPeriodEnd => $_getBF(6);
  @$pb.TagNumber(7)
  set cancelAtPeriodEnd($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCancelAtPeriodEnd() => $_has(6);
  @$pb.TagNumber(7)
  void clearCancelAtPeriodEnd() => $_clearField(7);

  /// RFC3339 period bounds when the provider reports them.
  @$pb.TagNumber(8)
  $core.String get currentPeriodStart => $_getSZ(7);
  @$pb.TagNumber(8)
  set currentPeriodStart($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasCurrentPeriodStart() => $_has(7);
  @$pb.TagNumber(8)
  void clearCurrentPeriodStart() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get currentPeriodEnd => $_getSZ(8);
  @$pb.TagNumber(9)
  set currentPeriodEnd($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasCurrentPeriodEnd() => $_has(8);
  @$pb.TagNumber(9)
  void clearCurrentPeriodEnd() => $_clearField(9);

  /// Billing subscription id, when the provider metadata carries it
  /// (checkout sessions set subscription_id). Preferred lookup key.
  @$pb.TagNumber(10)
  $core.String get subscriptionId => $_getSZ(9);
  @$pb.TagNumber(10)
  set subscriptionId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasSubscriptionId() => $_has(9);
  @$pb.TagNumber(10)
  void clearSubscriptionId() => $_clearField(10);

  /// Fallbacks when subscription_id is unknown.
  @$pb.TagNumber(11)
  $core.String get profileId => $_getSZ(10);
  @$pb.TagNumber(11)
  set profileId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasProfileId() => $_has(10);
  @$pb.TagNumber(11)
  void clearProfileId() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get planId => $_getSZ(11);
  @$pb.TagNumber(12)
  set planId($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasPlanId() => $_has(11);
  @$pb.TagNumber(12)
  void clearPlanId() => $_clearField(12);

  /// Billed events: provider order / invoice id and amount charged.
  @$pb.TagNumber(13)
  $core.String get externalOrderId => $_getSZ(12);
  @$pb.TagNumber(13)
  set externalOrderId($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasExternalOrderId() => $_has(12);
  @$pb.TagNumber(13)
  void clearExternalOrderId() => $_clearField(13);

  @$pb.TagNumber(14)
  $0.Money get amount => $_getN(13);
  @$pb.TagNumber(14)
  set amount($0.Money value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasAmount() => $_has(13);
  @$pb.TagNumber(14)
  void clearAmount() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Money ensureAmount() => $_ensure(13);

  /// RFC3339 provider timestamp; older events never overwrite newer state.
  @$pb.TagNumber(15)
  $core.String get occurredAt => $_getSZ(14);
  @$pb.TagNumber(15)
  set occurredAt($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasOccurredAt() => $_has(14);
  @$pb.TagNumber(15)
  void clearOccurredAt() => $_clearField(15);

  /// Free-form provider metadata (never used for lookup).
  @$pb.TagNumber(16)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(15);
}

class SyncProviderSubscriptionResponse extends $pb.GeneratedMessage {
  factory SyncProviderSubscriptionResponse({
    $core.String? subscriptionId,
    $core.String? subscriptionState,
    $core.bool? applied,
  }) {
    final result = SyncProviderSubscriptionResponse._();
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (subscriptionState != null) result.subscriptionState = subscriptionState;
    if (applied != null) result.applied = applied;
    return result;
  }

  SyncProviderSubscriptionResponse._();

  factory SyncProviderSubscriptionResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncProviderSubscriptionResponse()..mergeFromBuffer(data, registry);
  factory SyncProviderSubscriptionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncProviderSubscriptionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SyncProviderSubscriptionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'collection.v1'),
      createEmptyInstance: SyncProviderSubscriptionResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(2, _omitFieldNames ? '' : 'subscriptionState')
    ..aOB(3, _omitFieldNames ? '' : 'applied')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncProviderSubscriptionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncProviderSubscriptionResponse copyWith(
          void Function(SyncProviderSubscriptionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SyncProviderSubscriptionResponse))
          as SyncProviderSubscriptionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SyncProviderSubscriptionResponse() / SyncProviderSubscriptionResponse.new instead')
  static SyncProviderSubscriptionResponse create() =>
      SyncProviderSubscriptionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SyncProviderSubscriptionResponse._();
  @$core.override
  SyncProviderSubscriptionResponse createEmptyInstance() =>
      SyncProviderSubscriptionResponse._();
  @$core.pragma('dart2js:noInline')
  static SyncProviderSubscriptionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SyncProviderSubscriptionResponse>(
          SyncProviderSubscriptionResponse.$_createMessage);
  static SyncProviderSubscriptionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get subscriptionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set subscriptionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSubscriptionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSubscriptionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get subscriptionState => $_getSZ(1);
  @$pb.TagNumber(2)
  set subscriptionState($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSubscriptionState() => $_has(1);
  @$pb.TagNumber(2)
  void clearSubscriptionState() => $_clearField(2);

  /// False when the event was ignored (stale, unknown subscription).
  @$pb.TagNumber(3)
  $core.bool get applied => $_getBF(2);
  @$pb.TagNumber(3)
  set applied($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasApplied() => $_has(2);
  @$pb.TagNumber(3)
  void clearApplied() => $_clearField(3);
}

class CollectionServiceApi {
  final $pb.RpcClient _client;

  CollectionServiceApi(this._client);

  /// CollectPayment opens hosted checkout for an issued invoice.
  /// Redirect the end user to data.page_url, then call ConfirmPayment with data.session_ref.
  $async.Future<CollectPaymentResponse> collectPayment(
          $pb.ClientContext? ctx, CollectPaymentRequest request) =>
      _client.invoke<CollectPaymentResponse>(ctx, 'CollectionService',
          'CollectPayment', request, CollectPaymentResponse());

  /// StartSubscription creates a subscription and opens checkout for any upfront fee.
  /// Redirect the end user to data.page_url when present; otherwise the subscription is already active.
  $async.Future<StartSubscriptionResponse> startSubscription(
          $pb.ClientContext? ctx, StartSubscriptionRequest request) =>
      _client.invoke<StartSubscriptionResponse>(ctx, 'CollectionService',
          'StartSubscription', request, StartSubscriptionResponse());

  /// ConfirmPayment captures a completed checkout session into billing.
  /// Source of truth is the checkout session status — never trust browser redirects alone.
  $async.Future<ConfirmPaymentResponse> confirmPayment(
          $pb.ClientContext? ctx, ConfirmPaymentRequest request) =>
      _client.invoke<ConfirmPaymentResponse>(ctx, 'CollectionService',
          'ConfirmPayment', request, ConfirmPaymentResponse());

  /// SyncProviderSubscription mirrors provider-managed subscription state
  /// (webhooks from Polar, Stripe) into billing. Grant provider_subscription_sync
  /// to integration service accounts only.
  $async.Future<SyncProviderSubscriptionResponse> syncProviderSubscription(
          $pb.ClientContext? ctx, SyncProviderSubscriptionRequest request) =>
      _client.invoke<SyncProviderSubscriptionResponse>(
          ctx,
          'CollectionService',
          'SyncProviderSubscription',
          request,
          SyncProviderSubscriptionResponse());

  /// CancelSubscription cancels a subscription (and voids unpaid first invoices).
  $async.Future<CancelSubscriptionResponse> cancelSubscription(
          $pb.ClientContext? ctx, CancelSubscriptionRequest request) =>
      _client.invoke<CancelSubscriptionResponse>(ctx, 'CollectionService',
          'CancelSubscription', request, CancelSubscriptionResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
