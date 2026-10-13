// This is a generated file - do not edit.
//
// Generated from checkout/v1/checkout.proto.

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
import 'checkout.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'checkout.pbenum.dart';

class PayerContact extends $pb.GeneratedMessage {
  factory PayerContact({
    $core.String? contactId,
    $core.String? msisdn,
  }) {
    final result = PayerContact._();
    if (contactId != null) result.contactId = contactId;
    if (msisdn != null) result.msisdn = msisdn;
    return result;
  }

  PayerContact._();

  factory PayerContact.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PayerContact()..mergeFromBuffer(data, registry);
  factory PayerContact.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PayerContact()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PayerContact',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: PayerContact.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'contactId')
    ..aOS(2, _omitFieldNames ? '' : 'msisdn')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PayerContact clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PayerContact copyWith(void Function(PayerContact) updates) =>
      super.copyWith((message) => updates(message as PayerContact))
          as PayerContact;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PayerContact() / PayerContact.new instead')
  static PayerContact create() => PayerContact._();
  static $pb.GeneratedMessage $_createMessage() => PayerContact._();
  @$core.override
  PayerContact createEmptyInstance() => PayerContact._();
  @$core.pragma('dart2js:noInline')
  static PayerContact getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PayerContact>(
          PayerContact.$_createMessage);
  static PayerContact? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get contactId => $_getSZ(0);
  @$pb.TagNumber(1)
  set contactId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasContactId() => $_has(0);
  @$pb.TagNumber(1)
  void clearContactId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get msisdn => $_getSZ(1);
  @$pb.TagNumber(2)
  set msisdn($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMsisdn() => $_has(1);
  @$pb.TagNumber(2)
  void clearMsisdn() => $_clearField(2);
}

class PayerPrefill extends $pb.GeneratedMessage {
  factory PayerPrefill({
    $core.String? profileId,
    $core.String? displayName,
    $core.String? language,
    $core.Iterable<PayerContact>? contacts,
  }) {
    final result = PayerPrefill._();
    if (profileId != null) result.profileId = profileId;
    if (displayName != null) result.displayName = displayName;
    if (language != null) result.language = language;
    if (contacts != null) result.contacts.addAll(contacts);
    return result;
  }

  PayerPrefill._();

  factory PayerPrefill.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PayerPrefill()..mergeFromBuffer(data, registry);
  factory PayerPrefill.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PayerPrefill()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PayerPrefill',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: PayerPrefill.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'profileId')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'language')
    ..pPM<PayerContact>(4, _omitFieldNames ? '' : 'contacts',
        subBuilder: PayerContact.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PayerPrefill clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PayerPrefill copyWith(void Function(PayerPrefill) updates) =>
      super.copyWith((message) => updates(message as PayerPrefill))
          as PayerPrefill;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PayerPrefill() / PayerPrefill.new instead')
  static PayerPrefill create() => PayerPrefill._();
  static $pb.GeneratedMessage $_createMessage() => PayerPrefill._();
  @$core.override
  PayerPrefill createEmptyInstance() => PayerPrefill._();
  @$core.pragma('dart2js:noInline')
  static PayerPrefill getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PayerPrefill>(
          PayerPrefill.$_createMessage);
  static PayerPrefill? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get profileId => $_getSZ(0);
  @$pb.TagNumber(1)
  set profileId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProfileId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProfileId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get language => $_getSZ(2);
  @$pb.TagNumber(3)
  set language($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLanguage() => $_has(2);
  @$pb.TagNumber(3)
  void clearLanguage() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<PayerContact> get contacts => $_getList(3);
}

class CheckoutSession extends $pb.GeneratedMessage {
  factory CheckoutSession({
    $core.String? ref,
    $core.String? name,
    $core.String? description,
    $0.Money? amount,
    AmountOption? amountOption,
    $core.String? orderRef,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.String? returnUrl,
    PayerPrefill? payer,
    SessionStatus? status,
    $core.String? promptId,
    $core.String? pageUrl,
    $core.String? expiresAt,
    $core.String? paymentId,
  }) {
    final result = CheckoutSession._();
    if (ref != null) result.ref = ref;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (amount != null) result.amount = amount;
    if (amountOption != null) result.amountOption = amountOption;
    if (orderRef != null) result.orderRef = orderRef;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (returnUrl != null) result.returnUrl = returnUrl;
    if (payer != null) result.payer = payer;
    if (status != null) result.status = status;
    if (promptId != null) result.promptId = promptId;
    if (pageUrl != null) result.pageUrl = pageUrl;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (paymentId != null) result.paymentId = paymentId;
    return result;
  }

  CheckoutSession._();

  factory CheckoutSession.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CheckoutSession()..mergeFromBuffer(data, registry);
  factory CheckoutSession.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CheckoutSession()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CheckoutSession',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: CheckoutSession.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'ref')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..aOM<$0.Money>(4, _omitFieldNames ? '' : 'amount',
        subBuilder: $0.Money.$_createMessage)
    ..aE<AmountOption>(5, _omitFieldNames ? '' : 'amountOption',
        enumValues: AmountOption.values)
    ..aOS(6, _omitFieldNames ? '' : 'orderRef')
    ..m<$core.String, $core.String>(7, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'CheckoutSession.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('checkout.v1'))
    ..aOS(8, _omitFieldNames ? '' : 'returnUrl')
    ..aOM<PayerPrefill>(9, _omitFieldNames ? '' : 'payer',
        subBuilder: PayerPrefill.$_createMessage)
    ..aE<SessionStatus>(10, _omitFieldNames ? '' : 'status',
        enumValues: SessionStatus.values)
    ..aOS(11, _omitFieldNames ? '' : 'promptId')
    ..aOS(12, _omitFieldNames ? '' : 'pageUrl')
    ..aOS(13, _omitFieldNames ? '' : 'expiresAt')
    ..aOS(14, _omitFieldNames ? '' : 'paymentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CheckoutSession clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CheckoutSession copyWith(void Function(CheckoutSession) updates) =>
      super.copyWith((message) => updates(message as CheckoutSession))
          as CheckoutSession;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CheckoutSession() / CheckoutSession.new instead')
  static CheckoutSession create() => CheckoutSession._();
  static $pb.GeneratedMessage $_createMessage() => CheckoutSession._();
  @$core.override
  CheckoutSession createEmptyInstance() => CheckoutSession._();
  @$core.pragma('dart2js:noInline')
  static CheckoutSession getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CheckoutSession>(
          CheckoutSession.$_createMessage);
  static CheckoutSession? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get ref => $_getSZ(0);
  @$pb.TagNumber(1)
  set ref($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRef() => $_has(0);
  @$pb.TagNumber(1)
  void clearRef() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Money get amount => $_getN(3);
  @$pb.TagNumber(4)
  set amount($0.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmount() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Money ensureAmount() => $_ensure(3);

  @$pb.TagNumber(5)
  AmountOption get amountOption => $_getN(4);
  @$pb.TagNumber(5)
  set amountOption(AmountOption value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasAmountOption() => $_has(4);
  @$pb.TagNumber(5)
  void clearAmountOption() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get orderRef => $_getSZ(5);
  @$pb.TagNumber(6)
  set orderRef($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasOrderRef() => $_has(5);
  @$pb.TagNumber(6)
  void clearOrderRef() => $_clearField(6);

  @$pb.TagNumber(7)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(6);

  @$pb.TagNumber(8)
  $core.String get returnUrl => $_getSZ(7);
  @$pb.TagNumber(8)
  set returnUrl($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasReturnUrl() => $_has(7);
  @$pb.TagNumber(8)
  void clearReturnUrl() => $_clearField(8);

  @$pb.TagNumber(9)
  PayerPrefill get payer => $_getN(8);
  @$pb.TagNumber(9)
  set payer(PayerPrefill value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasPayer() => $_has(8);
  @$pb.TagNumber(9)
  void clearPayer() => $_clearField(9);
  @$pb.TagNumber(9)
  PayerPrefill ensurePayer() => $_ensure(8);

  @$pb.TagNumber(10)
  SessionStatus get status => $_getN(9);
  @$pb.TagNumber(10)
  set status(SessionStatus value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasStatus() => $_has(9);
  @$pb.TagNumber(10)
  void clearStatus() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get promptId => $_getSZ(10);
  @$pb.TagNumber(11)
  set promptId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasPromptId() => $_has(10);
  @$pb.TagNumber(11)
  void clearPromptId() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get pageUrl => $_getSZ(11);
  @$pb.TagNumber(12)
  set pageUrl($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasPageUrl() => $_has(11);
  @$pb.TagNumber(12)
  void clearPageUrl() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.String get expiresAt => $_getSZ(12);
  @$pb.TagNumber(13)
  set expiresAt($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasExpiresAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearExpiresAt() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.String get paymentId => $_getSZ(13);
  @$pb.TagNumber(14)
  set paymentId($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasPaymentId() => $_has(13);
  @$pb.TagNumber(14)
  void clearPaymentId() => $_clearField(14);
}

class CheckoutLink extends $pb.GeneratedMessage {
  factory CheckoutLink({
    $core.String? ref,
    $core.String? name,
    $core.String? description,
    $0.Money? amount,
    AmountOption? amountOption,
    $core.String? orderRef,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.String? returnUrl,
    $core.String? pageUrl,
    $core.String? expiresAt,
    $core.bool? active,
  }) {
    final result = CheckoutLink._();
    if (ref != null) result.ref = ref;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (amount != null) result.amount = amount;
    if (amountOption != null) result.amountOption = amountOption;
    if (orderRef != null) result.orderRef = orderRef;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (returnUrl != null) result.returnUrl = returnUrl;
    if (pageUrl != null) result.pageUrl = pageUrl;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (active != null) result.active = active;
    return result;
  }

  CheckoutLink._();

  factory CheckoutLink.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CheckoutLink()..mergeFromBuffer(data, registry);
  factory CheckoutLink.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CheckoutLink()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CheckoutLink',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: CheckoutLink.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'ref')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..aOM<$0.Money>(4, _omitFieldNames ? '' : 'amount',
        subBuilder: $0.Money.$_createMessage)
    ..aE<AmountOption>(5, _omitFieldNames ? '' : 'amountOption',
        enumValues: AmountOption.values)
    ..aOS(6, _omitFieldNames ? '' : 'orderRef')
    ..m<$core.String, $core.String>(7, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'CheckoutLink.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('checkout.v1'))
    ..aOS(8, _omitFieldNames ? '' : 'returnUrl')
    ..aOS(9, _omitFieldNames ? '' : 'pageUrl')
    ..aOS(10, _omitFieldNames ? '' : 'expiresAt')
    ..aOB(11, _omitFieldNames ? '' : 'active')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CheckoutLink clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CheckoutLink copyWith(void Function(CheckoutLink) updates) =>
      super.copyWith((message) => updates(message as CheckoutLink))
          as CheckoutLink;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CheckoutLink() / CheckoutLink.new instead')
  static CheckoutLink create() => CheckoutLink._();
  static $pb.GeneratedMessage $_createMessage() => CheckoutLink._();
  @$core.override
  CheckoutLink createEmptyInstance() => CheckoutLink._();
  @$core.pragma('dart2js:noInline')
  static CheckoutLink getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CheckoutLink>(
          CheckoutLink.$_createMessage);
  static CheckoutLink? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get ref => $_getSZ(0);
  @$pb.TagNumber(1)
  set ref($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRef() => $_has(0);
  @$pb.TagNumber(1)
  void clearRef() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Money get amount => $_getN(3);
  @$pb.TagNumber(4)
  set amount($0.Money value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmount() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Money ensureAmount() => $_ensure(3);

  @$pb.TagNumber(5)
  AmountOption get amountOption => $_getN(4);
  @$pb.TagNumber(5)
  set amountOption(AmountOption value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasAmountOption() => $_has(4);
  @$pb.TagNumber(5)
  void clearAmountOption() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get orderRef => $_getSZ(5);
  @$pb.TagNumber(6)
  set orderRef($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasOrderRef() => $_has(5);
  @$pb.TagNumber(6)
  void clearOrderRef() => $_clearField(6);

  @$pb.TagNumber(7)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(6);

  @$pb.TagNumber(8)
  $core.String get returnUrl => $_getSZ(7);
  @$pb.TagNumber(8)
  set returnUrl($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasReturnUrl() => $_has(7);
  @$pb.TagNumber(8)
  void clearReturnUrl() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get pageUrl => $_getSZ(8);
  @$pb.TagNumber(9)
  set pageUrl($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasPageUrl() => $_has(8);
  @$pb.TagNumber(9)
  void clearPageUrl() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get expiresAt => $_getSZ(9);
  @$pb.TagNumber(10)
  set expiresAt($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasExpiresAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearExpiresAt() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.bool get active => $_getBF(10);
  @$pb.TagNumber(11)
  set active($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasActive() => $_has(10);
  @$pb.TagNumber(11)
  void clearActive() => $_clearField(11);
}

class CreateCheckoutSessionRequest extends $pb.GeneratedMessage {
  factory CreateCheckoutSessionRequest({
    $core.String? name,
    $core.String? description,
    $0.Money? amount,
    AmountOption? amountOption,
    $core.String? orderRef,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.String? returnUrl,
    PayerPrefill? payer,
    $core.Iterable<$core.String>? methods,
  }) {
    final result = CreateCheckoutSessionRequest._();
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (amount != null) result.amount = amount;
    if (amountOption != null) result.amountOption = amountOption;
    if (orderRef != null) result.orderRef = orderRef;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (returnUrl != null) result.returnUrl = returnUrl;
    if (payer != null) result.payer = payer;
    if (methods != null) result.methods.addAll(methods);
    return result;
  }

  CreateCheckoutSessionRequest._();

  factory CreateCheckoutSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutSessionRequest()..mergeFromBuffer(data, registry);
  factory CreateCheckoutSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutSessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCheckoutSessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: CreateCheckoutSessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..aOM<$0.Money>(3, _omitFieldNames ? '' : 'amount',
        subBuilder: $0.Money.$_createMessage)
    ..aE<AmountOption>(4, _omitFieldNames ? '' : 'amountOption',
        enumValues: AmountOption.values)
    ..aOS(5, _omitFieldNames ? '' : 'orderRef')
    ..m<$core.String, $core.String>(6, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'CreateCheckoutSessionRequest.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('checkout.v1'))
    ..aOS(7, _omitFieldNames ? '' : 'returnUrl')
    ..aOM<PayerPrefill>(8, _omitFieldNames ? '' : 'payer',
        subBuilder: PayerPrefill.$_createMessage)
    ..pPS(9, _omitFieldNames ? '' : 'methods')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutSessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutSessionRequest copyWith(
          void Function(CreateCheckoutSessionRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateCheckoutSessionRequest))
          as CreateCheckoutSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCheckoutSessionRequest() / CreateCheckoutSessionRequest.new instead')
  static CreateCheckoutSessionRequest create() =>
      CreateCheckoutSessionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCheckoutSessionRequest._();
  @$core.override
  CreateCheckoutSessionRequest createEmptyInstance() =>
      CreateCheckoutSessionRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateCheckoutSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCheckoutSessionRequest>(
          CreateCheckoutSessionRequest.$_createMessage);
  static CreateCheckoutSessionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);

  @$pb.TagNumber(3)
  $0.Money get amount => $_getN(2);
  @$pb.TagNumber(3)
  set amount($0.Money value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearAmount() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Money ensureAmount() => $_ensure(2);

  @$pb.TagNumber(4)
  AmountOption get amountOption => $_getN(3);
  @$pb.TagNumber(4)
  set amountOption(AmountOption value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasAmountOption() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmountOption() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get orderRef => $_getSZ(4);
  @$pb.TagNumber(5)
  set orderRef($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOrderRef() => $_has(4);
  @$pb.TagNumber(5)
  void clearOrderRef() => $_clearField(5);

  @$pb.TagNumber(6)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(5);

  @$pb.TagNumber(7)
  $core.String get returnUrl => $_getSZ(6);
  @$pb.TagNumber(7)
  set returnUrl($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasReturnUrl() => $_has(6);
  @$pb.TagNumber(7)
  void clearReturnUrl() => $_clearField(7);

  @$pb.TagNumber(8)
  PayerPrefill get payer => $_getN(7);
  @$pb.TagNumber(8)
  set payer(PayerPrefill value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasPayer() => $_has(7);
  @$pb.TagNumber(8)
  void clearPayer() => $_clearField(8);
  @$pb.TagNumber(8)
  PayerPrefill ensurePayer() => $_ensure(7);

  @$pb.TagNumber(9)
  $pb.PbList<$core.String> get methods => $_getList(8);
}

class CreateCheckoutSessionResponse extends $pb.GeneratedMessage {
  factory CreateCheckoutSessionResponse({
    CheckoutSession? data,
  }) {
    final result = CreateCheckoutSessionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateCheckoutSessionResponse._();

  factory CreateCheckoutSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutSessionResponse()..mergeFromBuffer(data, registry);
  factory CreateCheckoutSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutSessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCheckoutSessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: CreateCheckoutSessionResponse.$_createMessage)
    ..aOM<CheckoutSession>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CheckoutSession.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutSessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutSessionResponse copyWith(
          void Function(CreateCheckoutSessionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateCheckoutSessionResponse))
          as CreateCheckoutSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCheckoutSessionResponse() / CreateCheckoutSessionResponse.new instead')
  static CreateCheckoutSessionResponse create() =>
      CreateCheckoutSessionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCheckoutSessionResponse._();
  @$core.override
  CreateCheckoutSessionResponse createEmptyInstance() =>
      CreateCheckoutSessionResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateCheckoutSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCheckoutSessionResponse>(
          CreateCheckoutSessionResponse.$_createMessage);
  static CreateCheckoutSessionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CheckoutSession get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CheckoutSession value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CheckoutSession ensureData() => $_ensure(0);
}

class GetCheckoutSessionRequest extends $pb.GeneratedMessage {
  factory GetCheckoutSessionRequest({
    $core.String? ref,
  }) {
    final result = GetCheckoutSessionRequest._();
    if (ref != null) result.ref = ref;
    return result;
  }

  GetCheckoutSessionRequest._();

  factory GetCheckoutSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCheckoutSessionRequest()..mergeFromBuffer(data, registry);
  factory GetCheckoutSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCheckoutSessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCheckoutSessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: GetCheckoutSessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'ref')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCheckoutSessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCheckoutSessionRequest copyWith(
          void Function(GetCheckoutSessionRequest) updates) =>
      super.copyWith((message) => updates(message as GetCheckoutSessionRequest))
          as GetCheckoutSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCheckoutSessionRequest() / GetCheckoutSessionRequest.new instead')
  static GetCheckoutSessionRequest create() => GetCheckoutSessionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetCheckoutSessionRequest._();
  @$core.override
  GetCheckoutSessionRequest createEmptyInstance() =>
      GetCheckoutSessionRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCheckoutSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCheckoutSessionRequest>(
          GetCheckoutSessionRequest.$_createMessage);
  static GetCheckoutSessionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get ref => $_getSZ(0);
  @$pb.TagNumber(1)
  set ref($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRef() => $_has(0);
  @$pb.TagNumber(1)
  void clearRef() => $_clearField(1);
}

class GetCheckoutSessionResponse extends $pb.GeneratedMessage {
  factory GetCheckoutSessionResponse({
    CheckoutSession? data,
  }) {
    final result = GetCheckoutSessionResponse._();
    if (data != null) result.data = data;
    return result;
  }

  GetCheckoutSessionResponse._();

  factory GetCheckoutSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCheckoutSessionResponse()..mergeFromBuffer(data, registry);
  factory GetCheckoutSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCheckoutSessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCheckoutSessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: GetCheckoutSessionResponse.$_createMessage)
    ..aOM<CheckoutSession>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CheckoutSession.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCheckoutSessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCheckoutSessionResponse copyWith(
          void Function(GetCheckoutSessionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetCheckoutSessionResponse))
          as GetCheckoutSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCheckoutSessionResponse() / GetCheckoutSessionResponse.new instead')
  static GetCheckoutSessionResponse create() => GetCheckoutSessionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetCheckoutSessionResponse._();
  @$core.override
  GetCheckoutSessionResponse createEmptyInstance() =>
      GetCheckoutSessionResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCheckoutSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCheckoutSessionResponse>(
          GetCheckoutSessionResponse.$_createMessage);
  static GetCheckoutSessionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CheckoutSession get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CheckoutSession value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CheckoutSession ensureData() => $_ensure(0);
}

class CreateCheckoutLinkRequest extends $pb.GeneratedMessage {
  factory CreateCheckoutLinkRequest({
    $core.String? name,
    $core.String? description,
    $0.Money? amount,
    AmountOption? amountOption,
    $core.String? orderRef,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.String? returnUrl,
    $core.String? expiresAt,
  }) {
    final result = CreateCheckoutLinkRequest._();
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (amount != null) result.amount = amount;
    if (amountOption != null) result.amountOption = amountOption;
    if (orderRef != null) result.orderRef = orderRef;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (returnUrl != null) result.returnUrl = returnUrl;
    if (expiresAt != null) result.expiresAt = expiresAt;
    return result;
  }

  CreateCheckoutLinkRequest._();

  factory CreateCheckoutLinkRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutLinkRequest()..mergeFromBuffer(data, registry);
  factory CreateCheckoutLinkRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutLinkRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCheckoutLinkRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: CreateCheckoutLinkRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..aOM<$0.Money>(3, _omitFieldNames ? '' : 'amount',
        subBuilder: $0.Money.$_createMessage)
    ..aE<AmountOption>(4, _omitFieldNames ? '' : 'amountOption',
        enumValues: AmountOption.values)
    ..aOS(5, _omitFieldNames ? '' : 'orderRef')
    ..m<$core.String, $core.String>(6, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'CreateCheckoutLinkRequest.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('checkout.v1'))
    ..aOS(7, _omitFieldNames ? '' : 'returnUrl')
    ..aOS(8, _omitFieldNames ? '' : 'expiresAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutLinkRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutLinkRequest copyWith(
          void Function(CreateCheckoutLinkRequest) updates) =>
      super.copyWith((message) => updates(message as CreateCheckoutLinkRequest))
          as CreateCheckoutLinkRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCheckoutLinkRequest() / CreateCheckoutLinkRequest.new instead')
  static CreateCheckoutLinkRequest create() => CreateCheckoutLinkRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCheckoutLinkRequest._();
  @$core.override
  CreateCheckoutLinkRequest createEmptyInstance() =>
      CreateCheckoutLinkRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateCheckoutLinkRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCheckoutLinkRequest>(
          CreateCheckoutLinkRequest.$_createMessage);
  static CreateCheckoutLinkRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);

  @$pb.TagNumber(3)
  $0.Money get amount => $_getN(2);
  @$pb.TagNumber(3)
  set amount($0.Money value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearAmount() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Money ensureAmount() => $_ensure(2);

  @$pb.TagNumber(4)
  AmountOption get amountOption => $_getN(3);
  @$pb.TagNumber(4)
  set amountOption(AmountOption value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasAmountOption() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmountOption() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get orderRef => $_getSZ(4);
  @$pb.TagNumber(5)
  set orderRef($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOrderRef() => $_has(4);
  @$pb.TagNumber(5)
  void clearOrderRef() => $_clearField(5);

  @$pb.TagNumber(6)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(5);

  @$pb.TagNumber(7)
  $core.String get returnUrl => $_getSZ(6);
  @$pb.TagNumber(7)
  set returnUrl($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasReturnUrl() => $_has(6);
  @$pb.TagNumber(7)
  void clearReturnUrl() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get expiresAt => $_getSZ(7);
  @$pb.TagNumber(8)
  set expiresAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasExpiresAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearExpiresAt() => $_clearField(8);
}

class CreateCheckoutLinkResponse extends $pb.GeneratedMessage {
  factory CreateCheckoutLinkResponse({
    CheckoutLink? data,
  }) {
    final result = CreateCheckoutLinkResponse._();
    if (data != null) result.data = data;
    return result;
  }

  CreateCheckoutLinkResponse._();

  factory CreateCheckoutLinkResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutLinkResponse()..mergeFromBuffer(data, registry);
  factory CreateCheckoutLinkResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCheckoutLinkResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCheckoutLinkResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'checkout.v1'),
      createEmptyInstance: CreateCheckoutLinkResponse.$_createMessage)
    ..aOM<CheckoutLink>(1, _omitFieldNames ? '' : 'data',
        subBuilder: CheckoutLink.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutLinkResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCheckoutLinkResponse copyWith(
          void Function(CreateCheckoutLinkResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateCheckoutLinkResponse))
          as CreateCheckoutLinkResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCheckoutLinkResponse() / CreateCheckoutLinkResponse.new instead')
  static CreateCheckoutLinkResponse create() => CreateCheckoutLinkResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCheckoutLinkResponse._();
  @$core.override
  CreateCheckoutLinkResponse createEmptyInstance() =>
      CreateCheckoutLinkResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateCheckoutLinkResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCheckoutLinkResponse>(
          CreateCheckoutLinkResponse.$_createMessage);
  static CreateCheckoutLinkResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CheckoutLink get data => $_getN(0);
  @$pb.TagNumber(1)
  set data(CheckoutLink value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
  @$pb.TagNumber(1)
  CheckoutLink ensureData() => $_ensure(0);
}

class CheckoutServiceApi {
  final $pb.RpcClient _client;

  CheckoutServiceApi(this._client);

  /// CreateCheckoutSession creates a single-use hosted payment session and
  /// returns the page URL the payer should be redirected to.
  $async.Future<CreateCheckoutSessionResponse> createCheckoutSession(
          $pb.ClientContext? ctx, CreateCheckoutSessionRequest request) =>
      _client.invoke<CreateCheckoutSessionResponse>(ctx, 'CheckoutService',
          'CreateCheckoutSession', request, CreateCheckoutSessionResponse());

  /// GetCheckoutSession returns the current state of a checkout session;
  /// merchants poll this as one source of payment truth.
  $async.Future<GetCheckoutSessionResponse> getCheckoutSession(
          $pb.ClientContext? ctx, GetCheckoutSessionRequest request) =>
      _client.invoke<GetCheckoutSessionResponse>(ctx, 'CheckoutService',
          'GetCheckoutSession', request, GetCheckoutSessionResponse());

  /// CreateCheckoutLink creates a reusable, shareable link that spawns a
  /// fresh checkout session per visitor.
  $async.Future<CreateCheckoutLinkResponse> createCheckoutLink(
          $pb.ClientContext? ctx, CreateCheckoutLinkRequest request) =>
      _client.invoke<CreateCheckoutLinkResponse>(ctx, 'CheckoutService',
          'CreateCheckoutLink', request, CreateCheckoutLinkResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
