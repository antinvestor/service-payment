// This is a generated file - do not edit.
//
// Generated from collection/v1/collection.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../common/v1/money.pbjson.dart' as $0;

@$core.Deprecated('Use collectionResultDescriptor instead')
const CollectionResult$json = {
  '1': 'CollectionResult',
  '2': [
    {'1': 'page_url', '3': 1, '4': 1, '5': 9, '10': 'pageUrl'},
    {'1': 'session_ref', '3': 2, '4': 1, '5': 9, '10': 'sessionRef'},
    {'1': 'invoice_id', '3': 3, '4': 1, '5': 9, '10': 'invoiceId'},
    {'1': 'subscription_id', '3': 4, '4': 1, '5': 9, '10': 'subscriptionId'},
    {'1': 'already_complete', '3': 5, '4': 1, '5': 8, '10': 'alreadyComplete'},
  ],
};

/// Descriptor for `CollectionResult`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List collectionResultDescriptor = $convert.base64Decode(
    'ChBDb2xsZWN0aW9uUmVzdWx0EhkKCHBhZ2VfdXJsGAEgASgJUgdwYWdlVXJsEh8KC3Nlc3Npb2'
    '5fcmVmGAIgASgJUgpzZXNzaW9uUmVmEh0KCmludm9pY2VfaWQYAyABKAlSCWludm9pY2VJZBIn'
    'Cg9zdWJzY3JpcHRpb25faWQYBCABKAlSDnN1YnNjcmlwdGlvbklkEikKEGFscmVhZHlfY29tcG'
    'xldGUYBSABKAhSD2FscmVhZHlDb21wbGV0ZQ==');

@$core.Deprecated('Use collectPaymentRequestDescriptor instead')
const CollectPaymentRequest$json = {
  '1': 'CollectPaymentRequest',
  '2': [
    {'1': 'invoice_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'invoiceId'},
    {'1': 'return_url', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'returnUrl'},
    {'1': 'methods', '3': 3, '4': 3, '5': 9, '10': 'methods'},
  ],
};

/// Descriptor for `CollectPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List collectPaymentRequestDescriptor = $convert.base64Decode(
    'ChVDb2xsZWN0UGF5bWVudFJlcXVlc3QSKAoKaW52b2ljZV9pZBgBIAEoCUIJukgGcgQQAxgoUg'
    'lpbnZvaWNlSWQSKgoKcmV0dXJuX3VybBgCIAEoCUILukgI2AEBcgOIAQFSCXJldHVyblVybBIY'
    'CgdtZXRob2RzGAMgAygJUgdtZXRob2Rz');

@$core.Deprecated('Use collectPaymentResponseDescriptor instead')
const CollectPaymentResponse$json = {
  '1': 'CollectPaymentResponse',
  '2': [
    {
      '1': 'data',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.collection.v1.CollectionResult',
      '10': 'data'
    },
  ],
};

/// Descriptor for `CollectPaymentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List collectPaymentResponseDescriptor =
    $convert.base64Decode(
        'ChZDb2xsZWN0UGF5bWVudFJlc3BvbnNlEjMKBGRhdGEYASABKAsyHy5jb2xsZWN0aW9uLnYxLk'
        'NvbGxlY3Rpb25SZXN1bHRSBGRhdGE=');

@$core.Deprecated('Use startSubscriptionRequestDescriptor instead')
const StartSubscriptionRequest$json = {
  '1': 'StartSubscriptionRequest',
  '2': [
    {'1': 'profile_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'profileId'},
    {'1': 'plan_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'planId'},
    {
      '1': 'catalog_version_id',
      '3': 3,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'catalogVersionId'
    },
    {'1': 'currency', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'currency'},
    {'1': 'return_url', '3': 5, '4': 1, '5': 9, '8': {}, '10': 'returnUrl'},
    {
      '1': 'payer_display_name',
      '3': 6,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'payerDisplayName'
    },
    {'1': 'methods', '3': 7, '4': 3, '5': 9, '10': 'methods'},
    {
      '1': 'external_entity_id',
      '3': 8,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'externalEntityId'
    },
    {
      '1': 'external_entity_type',
      '3': 9,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'externalEntityType'
    },
    {
      '1': 'integration_route_id',
      '3': 10,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'integrationRouteId'
    },
    {
      '1': 'metadata',
      '3': 11,
      '4': 3,
      '5': 11,
      '6': '.collection.v1.StartSubscriptionRequest.MetadataEntry',
      '10': 'metadata'
    },
  ],
  '3': [StartSubscriptionRequest_MetadataEntry$json],
};

@$core.Deprecated('Use startSubscriptionRequestDescriptor instead')
const StartSubscriptionRequest_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `StartSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startSubscriptionRequestDescriptor = $convert.base64Decode(
    'ChhTdGFydFN1YnNjcmlwdGlvblJlcXVlc3QSJgoKcHJvZmlsZV9pZBgBIAEoCUIHukgEcgIQAV'
    'IJcHJvZmlsZUlkEiAKB3BsYW5faWQYAiABKAlCB7pIBHICEAFSBnBsYW5JZBI1ChJjYXRhbG9n'
    'X3ZlcnNpb25faWQYAyABKAlCB7pIBHICEAFSEGNhdGFsb2dWZXJzaW9uSWQSJAoIY3VycmVuY3'
    'kYBCABKAlCCLpIBXIDmAEDUghjdXJyZW5jeRIqCgpyZXR1cm5fdXJsGAUgASgJQgu6SAjYAQFy'
    'A4gBAVIJcmV0dXJuVXJsEjUKEnBheWVyX2Rpc3BsYXlfbmFtZRgGIAEoCUIHukgEcgIYeFIQcG'
    'F5ZXJEaXNwbGF5TmFtZRIYCgdtZXRob2RzGAcgAygJUgdtZXRob2RzEjUKEmV4dGVybmFsX2Vu'
    'dGl0eV9pZBgIIAEoCUIHukgEcgIYeFIQZXh0ZXJuYWxFbnRpdHlJZBI5ChRleHRlcm5hbF9lbn'
    'RpdHlfdHlwZRgJIAEoCUIHukgEcgIYQFISZXh0ZXJuYWxFbnRpdHlUeXBlEjkKFGludGVncmF0'
    'aW9uX3JvdXRlX2lkGAogASgJQge6SARyAhgyUhJpbnRlZ3JhdGlvblJvdXRlSWQSUQoIbWV0YW'
    'RhdGEYCyADKAsyNS5jb2xsZWN0aW9uLnYxLlN0YXJ0U3Vic2NyaXB0aW9uUmVxdWVzdC5NZXRh'
    'ZGF0YUVudHJ5UghtZXRhZGF0YRo7Cg1NZXRhZGF0YUVudHJ5EhAKA2tleRgBIAEoCVIDa2V5Eh'
    'QKBXZhbHVlGAIgASgJUgV2YWx1ZToCOAE=');

@$core.Deprecated('Use startSubscriptionResponseDescriptor instead')
const StartSubscriptionResponse$json = {
  '1': 'StartSubscriptionResponse',
  '2': [
    {
      '1': 'data',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.collection.v1.CollectionResult',
      '10': 'data'
    },
  ],
};

/// Descriptor for `StartSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startSubscriptionResponseDescriptor =
    $convert.base64Decode(
        'ChlTdGFydFN1YnNjcmlwdGlvblJlc3BvbnNlEjMKBGRhdGEYASABKAsyHy5jb2xsZWN0aW9uLn'
        'YxLkNvbGxlY3Rpb25SZXN1bHRSBGRhdGE=');

@$core.Deprecated('Use confirmPaymentRequestDescriptor instead')
const ConfirmPaymentRequest$json = {
  '1': 'ConfirmPaymentRequest',
  '2': [
    {'1': 'session_ref', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'sessionRef'},
  ],
};

/// Descriptor for `ConfirmPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List confirmPaymentRequestDescriptor = $convert.base64Decode(
    'ChVDb25maXJtUGF5bWVudFJlcXVlc3QSKgoLc2Vzc2lvbl9yZWYYASABKAlCCbpIBnIEEAgYQF'
    'IKc2Vzc2lvblJlZg==');

@$core.Deprecated('Use confirmPaymentResponseDescriptor instead')
const ConfirmPaymentResponse$json = {
  '1': 'ConfirmPaymentResponse',
  '2': [
    {'1': 'invoice_id', '3': 1, '4': 1, '5': 9, '10': 'invoiceId'},
    {'1': 'invoice_state', '3': 2, '4': 1, '5': 9, '10': 'invoiceState'},
    {'1': 'subscription_id', '3': 3, '4': 1, '5': 9, '10': 'subscriptionId'},
    {
      '1': 'subscription_state',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'subscriptionState'
    },
    {'1': 'paid', '3': 5, '4': 1, '5': 8, '10': 'paid'},
  ],
};

/// Descriptor for `ConfirmPaymentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List confirmPaymentResponseDescriptor = $convert.base64Decode(
    'ChZDb25maXJtUGF5bWVudFJlc3BvbnNlEh0KCmludm9pY2VfaWQYASABKAlSCWludm9pY2VJZB'
    'IjCg1pbnZvaWNlX3N0YXRlGAIgASgJUgxpbnZvaWNlU3RhdGUSJwoPc3Vic2NyaXB0aW9uX2lk'
    'GAMgASgJUg5zdWJzY3JpcHRpb25JZBItChJzdWJzY3JpcHRpb25fc3RhdGUYBCABKAlSEXN1Yn'
    'NjcmlwdGlvblN0YXRlEhIKBHBhaWQYBSABKAhSBHBhaWQ=');

@$core.Deprecated('Use cancelSubscriptionRequestDescriptor instead')
const CancelSubscriptionRequest$json = {
  '1': 'CancelSubscriptionRequest',
  '2': [
    {
      '1': 'subscription_id',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'subscriptionId'
    },
  ],
};

/// Descriptor for `CancelSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelSubscriptionRequestDescriptor =
    $convert.base64Decode(
        'ChlDYW5jZWxTdWJzY3JpcHRpb25SZXF1ZXN0EjIKD3N1YnNjcmlwdGlvbl9pZBgBIAEoCUIJuk'
        'gGcgQQAxgoUg5zdWJzY3JpcHRpb25JZA==');

@$core.Deprecated('Use cancelSubscriptionResponseDescriptor instead')
const CancelSubscriptionResponse$json = {
  '1': 'CancelSubscriptionResponse',
  '2': [
    {'1': 'subscription_id', '3': 1, '4': 1, '5': 9, '10': 'subscriptionId'},
    {
      '1': 'subscription_state',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'subscriptionState'
    },
    {'1': 'voided_invoice_id', '3': 3, '4': 1, '5': 9, '10': 'voidedInvoiceId'},
  ],
};

/// Descriptor for `CancelSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelSubscriptionResponseDescriptor =
    $convert.base64Decode(
        'ChpDYW5jZWxTdWJzY3JpcHRpb25SZXNwb25zZRInCg9zdWJzY3JpcHRpb25faWQYASABKAlSDn'
        'N1YnNjcmlwdGlvbklkEi0KEnN1YnNjcmlwdGlvbl9zdGF0ZRgCIAEoCVIRc3Vic2NyaXB0aW9u'
        'U3RhdGUSKgoRdm9pZGVkX2ludm9pY2VfaWQYAyABKAlSD3ZvaWRlZEludm9pY2VJZA==');

@$core.Deprecated('Use syncProviderSubscriptionRequestDescriptor instead')
const SyncProviderSubscriptionRequest$json = {
  '1': 'SyncProviderSubscriptionRequest',
  '2': [
    {'1': 'provider', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'provider'},
    {'1': 'event', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'event'},
    {
      '1': 'external_subscription_id',
      '3': 3,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'externalSubscriptionId'
    },
    {
      '1': 'external_customer_id',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'externalCustomerId'
    },
    {
      '1': 'external_product_id',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'externalProductId'
    },
    {'1': 'status', '3': 6, '4': 1, '5': 9, '10': 'status'},
    {
      '1': 'cancel_at_period_end',
      '3': 7,
      '4': 1,
      '5': 8,
      '10': 'cancelAtPeriodEnd'
    },
    {
      '1': 'current_period_start',
      '3': 8,
      '4': 1,
      '5': 9,
      '10': 'currentPeriodStart'
    },
    {
      '1': 'current_period_end',
      '3': 9,
      '4': 1,
      '5': 9,
      '10': 'currentPeriodEnd'
    },
    {'1': 'subscription_id', '3': 10, '4': 1, '5': 9, '10': 'subscriptionId'},
    {'1': 'profile_id', '3': 11, '4': 1, '5': 9, '10': 'profileId'},
    {'1': 'plan_id', '3': 12, '4': 1, '5': 9, '10': 'planId'},
    {
      '1': 'external_order_id',
      '3': 13,
      '4': 1,
      '5': 9,
      '10': 'externalOrderId'
    },
    {
      '1': 'amount',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.common.v1.Money',
      '10': 'amount'
    },
    {'1': 'occurred_at', '3': 15, '4': 1, '5': 9, '10': 'occurredAt'},
    {
      '1': 'metadata',
      '3': 16,
      '4': 3,
      '5': 11,
      '6': '.collection.v1.SyncProviderSubscriptionRequest.MetadataEntry',
      '10': 'metadata'
    },
  ],
  '3': [SyncProviderSubscriptionRequest_MetadataEntry$json],
};

@$core.Deprecated('Use syncProviderSubscriptionRequestDescriptor instead')
const SyncProviderSubscriptionRequest_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `SyncProviderSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List syncProviderSubscriptionRequestDescriptor = $convert.base64Decode(
    'Ch9TeW5jUHJvdmlkZXJTdWJzY3JpcHRpb25SZXF1ZXN0EiMKCHByb3ZpZGVyGAEgASgJQge6SA'
    'RyAhABUghwcm92aWRlchIdCgVldmVudBgCIAEoCUIHukgEcgIQAVIFZXZlbnQSQQoYZXh0ZXJu'
    'YWxfc3Vic2NyaXB0aW9uX2lkGAMgASgJQge6SARyAhABUhZleHRlcm5hbFN1YnNjcmlwdGlvbk'
    'lkEjAKFGV4dGVybmFsX2N1c3RvbWVyX2lkGAQgASgJUhJleHRlcm5hbEN1c3RvbWVySWQSLgoT'
    'ZXh0ZXJuYWxfcHJvZHVjdF9pZBgFIAEoCVIRZXh0ZXJuYWxQcm9kdWN0SWQSFgoGc3RhdHVzGA'
    'YgASgJUgZzdGF0dXMSLwoUY2FuY2VsX2F0X3BlcmlvZF9lbmQYByABKAhSEWNhbmNlbEF0UGVy'
    'aW9kRW5kEjAKFGN1cnJlbnRfcGVyaW9kX3N0YXJ0GAggASgJUhJjdXJyZW50UGVyaW9kU3Rhcn'
    'QSLAoSY3VycmVudF9wZXJpb2RfZW5kGAkgASgJUhBjdXJyZW50UGVyaW9kRW5kEicKD3N1YnNj'
    'cmlwdGlvbl9pZBgKIAEoCVIOc3Vic2NyaXB0aW9uSWQSHQoKcHJvZmlsZV9pZBgLIAEoCVIJcH'
    'JvZmlsZUlkEhcKB3BsYW5faWQYDCABKAlSBnBsYW5JZBIqChFleHRlcm5hbF9vcmRlcl9pZBgN'
    'IAEoCVIPZXh0ZXJuYWxPcmRlcklkEigKBmFtb3VudBgOIAEoCzIQLmNvbW1vbi52MS5Nb25leV'
    'IGYW1vdW50Eh8KC29jY3VycmVkX2F0GA8gASgJUgpvY2N1cnJlZEF0ElgKCG1ldGFkYXRhGBAg'
    'AygLMjwuY29sbGVjdGlvbi52MS5TeW5jUHJvdmlkZXJTdWJzY3JpcHRpb25SZXF1ZXN0Lk1ldG'
    'FkYXRhRW50cnlSCG1ldGFkYXRhGjsKDU1ldGFkYXRhRW50cnkSEAoDa2V5GAEgASgJUgNrZXkS'
    'FAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4AQ==');

@$core.Deprecated('Use syncProviderSubscriptionResponseDescriptor instead')
const SyncProviderSubscriptionResponse$json = {
  '1': 'SyncProviderSubscriptionResponse',
  '2': [
    {'1': 'subscription_id', '3': 1, '4': 1, '5': 9, '10': 'subscriptionId'},
    {
      '1': 'subscription_state',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'subscriptionState'
    },
    {'1': 'applied', '3': 3, '4': 1, '5': 8, '10': 'applied'},
  ],
};

/// Descriptor for `SyncProviderSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List syncProviderSubscriptionResponseDescriptor =
    $convert.base64Decode(
        'CiBTeW5jUHJvdmlkZXJTdWJzY3JpcHRpb25SZXNwb25zZRInCg9zdWJzY3JpcHRpb25faWQYAS'
        'ABKAlSDnN1YnNjcmlwdGlvbklkEi0KEnN1YnNjcmlwdGlvbl9zdGF0ZRgCIAEoCVIRc3Vic2Ny'
        'aXB0aW9uU3RhdGUSGAoHYXBwbGllZBgDIAEoCFIHYXBwbGllZA==');

const $core.Map<$core.String, $core.dynamic> CollectionServiceBase$json = {
  '1': 'CollectionService',
  '2': [
    {
      '1': 'CollectPayment',
      '2': '.collection.v1.CollectPaymentRequest',
      '3': '.collection.v1.CollectPaymentResponse',
      '4': {}
    },
    {
      '1': 'StartSubscription',
      '2': '.collection.v1.StartSubscriptionRequest',
      '3': '.collection.v1.StartSubscriptionResponse',
      '4': {}
    },
    {
      '1': 'ConfirmPayment',
      '2': '.collection.v1.ConfirmPaymentRequest',
      '3': '.collection.v1.ConfirmPaymentResponse',
      '4': {}
    },
    {
      '1': 'SyncProviderSubscription',
      '2': '.collection.v1.SyncProviderSubscriptionRequest',
      '3': '.collection.v1.SyncProviderSubscriptionResponse',
      '4': {}
    },
    {
      '1': 'CancelSubscription',
      '2': '.collection.v1.CancelSubscriptionRequest',
      '3': '.collection.v1.CancelSubscriptionResponse',
      '4': {}
    },
  ],
};

@$core.Deprecated('Use collectionServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    CollectionServiceBase$messageJson = {
  '.collection.v1.CollectPaymentRequest': CollectPaymentRequest$json,
  '.collection.v1.CollectPaymentResponse': CollectPaymentResponse$json,
  '.collection.v1.CollectionResult': CollectionResult$json,
  '.collection.v1.StartSubscriptionRequest': StartSubscriptionRequest$json,
  '.collection.v1.StartSubscriptionRequest.MetadataEntry':
      StartSubscriptionRequest_MetadataEntry$json,
  '.collection.v1.StartSubscriptionResponse': StartSubscriptionResponse$json,
  '.collection.v1.ConfirmPaymentRequest': ConfirmPaymentRequest$json,
  '.collection.v1.ConfirmPaymentResponse': ConfirmPaymentResponse$json,
  '.collection.v1.SyncProviderSubscriptionRequest':
      SyncProviderSubscriptionRequest$json,
  '.common.v1.Money': $0.Money$json,
  '.collection.v1.SyncProviderSubscriptionRequest.MetadataEntry':
      SyncProviderSubscriptionRequest_MetadataEntry$json,
  '.collection.v1.SyncProviderSubscriptionResponse':
      SyncProviderSubscriptionResponse$json,
  '.collection.v1.CancelSubscriptionRequest': CancelSubscriptionRequest$json,
  '.collection.v1.CancelSubscriptionResponse': CancelSubscriptionResponse$json,
};

/// Descriptor for `CollectionService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List collectionServiceDescriptor = $convert.base64Decode(
    'ChFDb2xsZWN0aW9uU2VydmljZRJ0Cg5Db2xsZWN0UGF5bWVudBIkLmNvbGxlY3Rpb24udjEuQ2'
    '9sbGVjdFBheW1lbnRSZXF1ZXN0GiUuY29sbGVjdGlvbi52MS5Db2xsZWN0UGF5bWVudFJlc3Bv'
    'bnNlIhWCtRgRCg9wYXltZW50X2NvbGxlY3QSgQEKEVN0YXJ0U3Vic2NyaXB0aW9uEicuY29sbG'
    'VjdGlvbi52MS5TdGFydFN1YnNjcmlwdGlvblJlcXVlc3QaKC5jb2xsZWN0aW9uLnYxLlN0YXJ0'
    'U3Vic2NyaXB0aW9uUmVzcG9uc2UiGYK1GBUKE3N1YnNjcmlwdGlvbl9tYW5hZ2USdAoOQ29uZm'
    'lybVBheW1lbnQSJC5jb2xsZWN0aW9uLnYxLkNvbmZpcm1QYXltZW50UmVxdWVzdBolLmNvbGxl'
    'Y3Rpb24udjEuQ29uZmlybVBheW1lbnRSZXNwb25zZSIVgrUYEQoPcGF5bWVudF9jb2xsZWN0Ep'
    '0BChhTeW5jUHJvdmlkZXJTdWJzY3JpcHRpb24SLi5jb2xsZWN0aW9uLnYxLlN5bmNQcm92aWRl'
    'clN1YnNjcmlwdGlvblJlcXVlc3QaLy5jb2xsZWN0aW9uLnYxLlN5bmNQcm92aWRlclN1YnNjcm'
    'lwdGlvblJlc3BvbnNlIiCCtRgcChpwcm92aWRlcl9zdWJzY3JpcHRpb25fc3luYxKEAQoSQ2Fu'
    'Y2VsU3Vic2NyaXB0aW9uEiguY29sbGVjdGlvbi52MS5DYW5jZWxTdWJzY3JpcHRpb25SZXF1ZX'
    'N0GikuY29sbGVjdGlvbi52MS5DYW5jZWxTdWJzY3JpcHRpb25SZXNwb25zZSIZgrUYFQoTc3Vi'
    'c2NyaXB0aW9uX21hbmFnZQ==');
