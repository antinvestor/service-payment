// This is a generated file - do not edit.
//
// Generated from checkout/v1/checkout.proto.

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

@$core.Deprecated('Use amountOptionDescriptor instead')
const AmountOption$json = {
  '1': 'AmountOption',
  '2': [
    {'1': 'AMOUNT_OPTION_FIXED_UNSPECIFIED', '2': 0},
    {'1': 'AMOUNT_OPTION_VARIABLE', '2': 1},
  ],
};

/// Descriptor for `AmountOption`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List amountOptionDescriptor = $convert.base64Decode(
    'CgxBbW91bnRPcHRpb24SIwofQU1PVU5UX09QVElPTl9GSVhFRF9VTlNQRUNJRklFRBAAEhoKFk'
    'FNT1VOVF9PUFRJT05fVkFSSUFCTEUQAQ==');

@$core.Deprecated('Use sessionStatusDescriptor instead')
const SessionStatus$json = {
  '1': 'SessionStatus',
  '2': [
    {'1': 'SESSION_STATUS_PENDING_UNSPECIFIED', '2': 0},
    {'1': 'SESSION_STATUS_PROCESSING', '2': 1},
    {'1': 'SESSION_STATUS_COMPLETED', '2': 2},
    {'1': 'SESSION_STATUS_FAILED', '2': 3},
    {'1': 'SESSION_STATUS_EXPIRED', '2': 4},
  ],
};

/// Descriptor for `SessionStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List sessionStatusDescriptor = $convert.base64Decode(
    'Cg1TZXNzaW9uU3RhdHVzEiYKIlNFU1NJT05fU1RBVFVTX1BFTkRJTkdfVU5TUEVDSUZJRUQQAB'
    'IdChlTRVNTSU9OX1NUQVRVU19QUk9DRVNTSU5HEAESHAoYU0VTU0lPTl9TVEFUVVNfQ09NUExF'
    'VEVEEAISGQoVU0VTU0lPTl9TVEFUVVNfRkFJTEVEEAMSGgoWU0VTU0lPTl9TVEFUVVNfRVhQSV'
    'JFRBAE');

@$core.Deprecated('Use payerContactDescriptor instead')
const PayerContact$json = {
  '1': 'PayerContact',
  '2': [
    {'1': 'contact_id', '3': 1, '4': 1, '5': 9, '10': 'contactId'},
    {'1': 'msisdn', '3': 2, '4': 1, '5': 9, '10': 'msisdn'},
  ],
};

/// Descriptor for `PayerContact`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List payerContactDescriptor = $convert.base64Decode(
    'CgxQYXllckNvbnRhY3QSHQoKY29udGFjdF9pZBgBIAEoCVIJY29udGFjdElkEhYKBm1zaXNkbh'
    'gCIAEoCVIGbXNpc2Ru');

@$core.Deprecated('Use payerPrefillDescriptor instead')
const PayerPrefill$json = {
  '1': 'PayerPrefill',
  '2': [
    {'1': 'profile_id', '3': 1, '4': 1, '5': 9, '10': 'profileId'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'language', '3': 3, '4': 1, '5': 9, '10': 'language'},
    {
      '1': 'contacts',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.checkout.v1.PayerContact',
      '10': 'contacts'
    },
  ],
};

/// Descriptor for `PayerPrefill`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List payerPrefillDescriptor = $convert.base64Decode(
    'CgxQYXllclByZWZpbGwSHQoKcHJvZmlsZV9pZBgBIAEoCVIJcHJvZmlsZUlkEiEKDGRpc3BsYX'
    'lfbmFtZRgCIAEoCVILZGlzcGxheU5hbWUSGgoIbGFuZ3VhZ2UYAyABKAlSCGxhbmd1YWdlEjUK'
    'CGNvbnRhY3RzGAQgAygLMhkuY2hlY2tvdXQudjEuUGF5ZXJDb250YWN0Ughjb250YWN0cw==');

@$core.Deprecated('Use checkoutSessionDescriptor instead')
const CheckoutSession$json = {
  '1': 'CheckoutSession',
  '2': [
    {'1': 'ref', '3': 1, '4': 1, '5': 9, '10': 'ref'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
    {
      '1': 'amount',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.common.v1.Money',
      '10': 'amount'
    },
    {
      '1': 'amount_option',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.checkout.v1.AmountOption',
      '10': 'amountOption'
    },
    {'1': 'order_ref', '3': 6, '4': 1, '5': 9, '10': 'orderRef'},
    {
      '1': 'metadata',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.checkout.v1.CheckoutSession.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'return_url', '3': 8, '4': 1, '5': 9, '10': 'returnUrl'},
    {
      '1': 'payer',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.checkout.v1.PayerPrefill',
      '10': 'payer'
    },
    {
      '1': 'status',
      '3': 10,
      '4': 1,
      '5': 14,
      '6': '.checkout.v1.SessionStatus',
      '10': 'status'
    },
    {'1': 'prompt_id', '3': 11, '4': 1, '5': 9, '10': 'promptId'},
    {'1': 'page_url', '3': 12, '4': 1, '5': 9, '10': 'pageUrl'},
    {'1': 'expires_at', '3': 13, '4': 1, '5': 9, '10': 'expiresAt'},
    {'1': 'payment_id', '3': 14, '4': 1, '5': 9, '10': 'paymentId'},
  ],
  '3': [CheckoutSession_MetadataEntry$json],
};

@$core.Deprecated('Use checkoutSessionDescriptor instead')
const CheckoutSession_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `CheckoutSession`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List checkoutSessionDescriptor = $convert.base64Decode(
    'Cg9DaGVja291dFNlc3Npb24SEAoDcmVmGAEgASgJUgNyZWYSEgoEbmFtZRgCIAEoCVIEbmFtZR'
    'IgCgtkZXNjcmlwdGlvbhgDIAEoCVILZGVzY3JpcHRpb24SKAoGYW1vdW50GAQgASgLMhAuY29t'
    'bW9uLnYxLk1vbmV5UgZhbW91bnQSPgoNYW1vdW50X29wdGlvbhgFIAEoDjIZLmNoZWNrb3V0Ln'
    'YxLkFtb3VudE9wdGlvblIMYW1vdW50T3B0aW9uEhsKCW9yZGVyX3JlZhgGIAEoCVIIb3JkZXJS'
    'ZWYSRgoIbWV0YWRhdGEYByADKAsyKi5jaGVja291dC52MS5DaGVja291dFNlc3Npb24uTWV0YW'
    'RhdGFFbnRyeVIIbWV0YWRhdGESHQoKcmV0dXJuX3VybBgIIAEoCVIJcmV0dXJuVXJsEi8KBXBh'
    'eWVyGAkgASgLMhkuY2hlY2tvdXQudjEuUGF5ZXJQcmVmaWxsUgVwYXllchIyCgZzdGF0dXMYCi'
    'ABKA4yGi5jaGVja291dC52MS5TZXNzaW9uU3RhdHVzUgZzdGF0dXMSGwoJcHJvbXB0X2lkGAsg'
    'ASgJUghwcm9tcHRJZBIZCghwYWdlX3VybBgMIAEoCVIHcGFnZVVybBIdCgpleHBpcmVzX2F0GA'
    '0gASgJUglleHBpcmVzQXQSHQoKcGF5bWVudF9pZBgOIAEoCVIJcGF5bWVudElkGjsKDU1ldGFk'
    'YXRhRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4AQ==');

@$core.Deprecated('Use checkoutLinkDescriptor instead')
const CheckoutLink$json = {
  '1': 'CheckoutLink',
  '2': [
    {'1': 'ref', '3': 1, '4': 1, '5': 9, '10': 'ref'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
    {
      '1': 'amount',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.common.v1.Money',
      '10': 'amount'
    },
    {
      '1': 'amount_option',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.checkout.v1.AmountOption',
      '10': 'amountOption'
    },
    {'1': 'order_ref', '3': 6, '4': 1, '5': 9, '10': 'orderRef'},
    {
      '1': 'metadata',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.checkout.v1.CheckoutLink.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'return_url', '3': 8, '4': 1, '5': 9, '10': 'returnUrl'},
    {'1': 'page_url', '3': 9, '4': 1, '5': 9, '10': 'pageUrl'},
    {'1': 'expires_at', '3': 10, '4': 1, '5': 9, '10': 'expiresAt'},
    {'1': 'active', '3': 11, '4': 1, '5': 8, '10': 'active'},
  ],
  '3': [CheckoutLink_MetadataEntry$json],
};

@$core.Deprecated('Use checkoutLinkDescriptor instead')
const CheckoutLink_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `CheckoutLink`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List checkoutLinkDescriptor = $convert.base64Decode(
    'CgxDaGVja291dExpbmsSEAoDcmVmGAEgASgJUgNyZWYSEgoEbmFtZRgCIAEoCVIEbmFtZRIgCg'
    'tkZXNjcmlwdGlvbhgDIAEoCVILZGVzY3JpcHRpb24SKAoGYW1vdW50GAQgASgLMhAuY29tbW9u'
    'LnYxLk1vbmV5UgZhbW91bnQSPgoNYW1vdW50X29wdGlvbhgFIAEoDjIZLmNoZWNrb3V0LnYxLk'
    'Ftb3VudE9wdGlvblIMYW1vdW50T3B0aW9uEhsKCW9yZGVyX3JlZhgGIAEoCVIIb3JkZXJSZWYS'
    'QwoIbWV0YWRhdGEYByADKAsyJy5jaGVja291dC52MS5DaGVja291dExpbmsuTWV0YWRhdGFFbn'
    'RyeVIIbWV0YWRhdGESHQoKcmV0dXJuX3VybBgIIAEoCVIJcmV0dXJuVXJsEhkKCHBhZ2VfdXJs'
    'GAkgASgJUgdwYWdlVXJsEh0KCmV4cGlyZXNfYXQYCiABKAlSCWV4cGlyZXNBdBIWCgZhY3Rpdm'
    'UYCyABKAhSBmFjdGl2ZRo7Cg1NZXRhZGF0YUVudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZh'
    'bHVlGAIgASgJUgV2YWx1ZToCOAE=');

@$core.Deprecated('Use createCheckoutSessionRequestDescriptor instead')
const CreateCheckoutSessionRequest$json = {
  '1': 'CreateCheckoutSessionRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'description'},
    {
      '1': 'amount',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.common.v1.Money',
      '10': 'amount'
    },
    {
      '1': 'amount_option',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.checkout.v1.AmountOption',
      '10': 'amountOption'
    },
    {'1': 'order_ref', '3': 5, '4': 1, '5': 9, '8': {}, '10': 'orderRef'},
    {
      '1': 'metadata',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.checkout.v1.CreateCheckoutSessionRequest.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'return_url', '3': 7, '4': 1, '5': 9, '8': {}, '10': 'returnUrl'},
    {
      '1': 'payer',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.checkout.v1.PayerPrefill',
      '10': 'payer'
    },
    {'1': 'methods', '3': 9, '4': 3, '5': 9, '10': 'methods'},
  ],
  '3': [CreateCheckoutSessionRequest_MetadataEntry$json],
};

@$core.Deprecated('Use createCheckoutSessionRequestDescriptor instead')
const CreateCheckoutSessionRequest_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `CreateCheckoutSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCheckoutSessionRequestDescriptor = $convert.base64Decode(
    'ChxDcmVhdGVDaGVja291dFNlc3Npb25SZXF1ZXN0Eh0KBG5hbWUYASABKAlCCbpIBnIEEAMYZF'
    'IEbmFtZRIqCgtkZXNjcmlwdGlvbhgCIAEoCUIIukgFcgMY9ANSC2Rlc2NyaXB0aW9uEigKBmFt'
    'b3VudBgDIAEoCzIQLmNvbW1vbi52MS5Nb25leVIGYW1vdW50Ej4KDWFtb3VudF9vcHRpb24YBC'
    'ABKA4yGS5jaGVja291dC52MS5BbW91bnRPcHRpb25SDGFtb3VudE9wdGlvbhIlCglvcmRlcl9y'
    'ZWYYBSABKAlCCLpIBXIDGPoBUghvcmRlclJlZhJTCghtZXRhZGF0YRgGIAMoCzI3LmNoZWNrb3'
    'V0LnYxLkNyZWF0ZUNoZWNrb3V0U2Vzc2lvblJlcXVlc3QuTWV0YWRhdGFFbnRyeVIIbWV0YWRh'
    'dGESJwoKcmV0dXJuX3VybBgHIAEoCUIIukgFcgOIAQFSCXJldHVyblVybBIvCgVwYXllchgIIA'
    'EoCzIZLmNoZWNrb3V0LnYxLlBheWVyUHJlZmlsbFIFcGF5ZXISGAoHbWV0aG9kcxgJIAMoCVIH'
    'bWV0aG9kcxo7Cg1NZXRhZGF0YUVudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZhbHVlGAIgAS'
    'gJUgV2YWx1ZToCOAE=');

@$core.Deprecated('Use createCheckoutSessionResponseDescriptor instead')
const CreateCheckoutSessionResponse$json = {
  '1': 'CreateCheckoutSessionResponse',
  '2': [
    {
      '1': 'data',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.checkout.v1.CheckoutSession',
      '10': 'data'
    },
  ],
};

/// Descriptor for `CreateCheckoutSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCheckoutSessionResponseDescriptor =
    $convert.base64Decode(
        'Ch1DcmVhdGVDaGVja291dFNlc3Npb25SZXNwb25zZRIwCgRkYXRhGAEgASgLMhwuY2hlY2tvdX'
        'QudjEuQ2hlY2tvdXRTZXNzaW9uUgRkYXRh');

@$core.Deprecated('Use getCheckoutSessionRequestDescriptor instead')
const GetCheckoutSessionRequest$json = {
  '1': 'GetCheckoutSessionRequest',
  '2': [
    {'1': 'ref', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'ref'},
  ],
};

/// Descriptor for `GetCheckoutSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCheckoutSessionRequestDescriptor =
    $convert.base64Decode(
        'ChlHZXRDaGVja291dFNlc3Npb25SZXF1ZXN0EhsKA3JlZhgBIAEoCUIJukgGcgQQCBhAUgNyZW'
        'Y=');

@$core.Deprecated('Use getCheckoutSessionResponseDescriptor instead')
const GetCheckoutSessionResponse$json = {
  '1': 'GetCheckoutSessionResponse',
  '2': [
    {
      '1': 'data',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.checkout.v1.CheckoutSession',
      '10': 'data'
    },
  ],
};

/// Descriptor for `GetCheckoutSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCheckoutSessionResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRDaGVja291dFNlc3Npb25SZXNwb25zZRIwCgRkYXRhGAEgASgLMhwuY2hlY2tvdXQudj'
        'EuQ2hlY2tvdXRTZXNzaW9uUgRkYXRh');

@$core.Deprecated('Use createCheckoutLinkRequestDescriptor instead')
const CreateCheckoutLinkRequest$json = {
  '1': 'CreateCheckoutLinkRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'description'},
    {
      '1': 'amount',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.common.v1.Money',
      '10': 'amount'
    },
    {
      '1': 'amount_option',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.checkout.v1.AmountOption',
      '10': 'amountOption'
    },
    {'1': 'order_ref', '3': 5, '4': 1, '5': 9, '8': {}, '10': 'orderRef'},
    {
      '1': 'metadata',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.checkout.v1.CreateCheckoutLinkRequest.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'return_url', '3': 7, '4': 1, '5': 9, '8': {}, '10': 'returnUrl'},
    {'1': 'expires_at', '3': 8, '4': 1, '5': 9, '10': 'expiresAt'},
  ],
  '3': [CreateCheckoutLinkRequest_MetadataEntry$json],
};

@$core.Deprecated('Use createCheckoutLinkRequestDescriptor instead')
const CreateCheckoutLinkRequest_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `CreateCheckoutLinkRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCheckoutLinkRequestDescriptor = $convert.base64Decode(
    'ChlDcmVhdGVDaGVja291dExpbmtSZXF1ZXN0Eh0KBG5hbWUYASABKAlCCbpIBnIEEAMYZFIEbm'
    'FtZRIqCgtkZXNjcmlwdGlvbhgCIAEoCUIIukgFcgMY9ANSC2Rlc2NyaXB0aW9uEigKBmFtb3Vu'
    'dBgDIAEoCzIQLmNvbW1vbi52MS5Nb25leVIGYW1vdW50Ej4KDWFtb3VudF9vcHRpb24YBCABKA'
    '4yGS5jaGVja291dC52MS5BbW91bnRPcHRpb25SDGFtb3VudE9wdGlvbhIlCglvcmRlcl9yZWYY'
    'BSABKAlCCLpIBXIDGPoBUghvcmRlclJlZhJQCghtZXRhZGF0YRgGIAMoCzI0LmNoZWNrb3V0Ln'
    'YxLkNyZWF0ZUNoZWNrb3V0TGlua1JlcXVlc3QuTWV0YWRhdGFFbnRyeVIIbWV0YWRhdGESJwoK'
    'cmV0dXJuX3VybBgHIAEoCUIIukgFcgOIAQFSCXJldHVyblVybBIdCgpleHBpcmVzX2F0GAggAS'
    'gJUglleHBpcmVzQXQaOwoNTWV0YWRhdGFFbnRyeRIQCgNrZXkYASABKAlSA2tleRIUCgV2YWx1'
    'ZRgCIAEoCVIFdmFsdWU6AjgB');

@$core.Deprecated('Use createCheckoutLinkResponseDescriptor instead')
const CreateCheckoutLinkResponse$json = {
  '1': 'CreateCheckoutLinkResponse',
  '2': [
    {
      '1': 'data',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.checkout.v1.CheckoutLink',
      '10': 'data'
    },
  ],
};

/// Descriptor for `CreateCheckoutLinkResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCheckoutLinkResponseDescriptor =
    $convert.base64Decode(
        'ChpDcmVhdGVDaGVja291dExpbmtSZXNwb25zZRItCgRkYXRhGAEgASgLMhkuY2hlY2tvdXQudj'
        'EuQ2hlY2tvdXRMaW5rUgRkYXRh');

const $core.Map<$core.String, $core.dynamic> CheckoutServiceBase$json = {
  '1': 'CheckoutService',
  '2': [
    {
      '1': 'CreateCheckoutSession',
      '2': '.checkout.v1.CreateCheckoutSessionRequest',
      '3': '.checkout.v1.CreateCheckoutSessionResponse',
      '4': {}
    },
    {
      '1': 'GetCheckoutSession',
      '2': '.checkout.v1.GetCheckoutSessionRequest',
      '3': '.checkout.v1.GetCheckoutSessionResponse',
      '4': {'34': 1},
    },
    {
      '1': 'CreateCheckoutLink',
      '2': '.checkout.v1.CreateCheckoutLinkRequest',
      '3': '.checkout.v1.CreateCheckoutLinkResponse',
      '4': {}
    },
  ],
  '3': {},
};

@$core.Deprecated('Use checkoutServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    CheckoutServiceBase$messageJson = {
  '.checkout.v1.CreateCheckoutSessionRequest':
      CreateCheckoutSessionRequest$json,
  '.common.v1.Money': $0.Money$json,
  '.checkout.v1.CreateCheckoutSessionRequest.MetadataEntry':
      CreateCheckoutSessionRequest_MetadataEntry$json,
  '.checkout.v1.PayerPrefill': PayerPrefill$json,
  '.checkout.v1.PayerContact': PayerContact$json,
  '.checkout.v1.CreateCheckoutSessionResponse':
      CreateCheckoutSessionResponse$json,
  '.checkout.v1.CheckoutSession': CheckoutSession$json,
  '.checkout.v1.CheckoutSession.MetadataEntry':
      CheckoutSession_MetadataEntry$json,
  '.checkout.v1.GetCheckoutSessionRequest': GetCheckoutSessionRequest$json,
  '.checkout.v1.GetCheckoutSessionResponse': GetCheckoutSessionResponse$json,
  '.checkout.v1.CreateCheckoutLinkRequest': CreateCheckoutLinkRequest$json,
  '.checkout.v1.CreateCheckoutLinkRequest.MetadataEntry':
      CreateCheckoutLinkRequest_MetadataEntry$json,
  '.checkout.v1.CreateCheckoutLinkResponse': CreateCheckoutLinkResponse$json,
  '.checkout.v1.CheckoutLink': CheckoutLink$json,
  '.checkout.v1.CheckoutLink.MetadataEntry': CheckoutLink_MetadataEntry$json,
};

/// Descriptor for `CheckoutService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List checkoutServiceDescriptor = $convert.base64Decode(
    'Cg9DaGVja291dFNlcnZpY2USjQEKFUNyZWF0ZUNoZWNrb3V0U2Vzc2lvbhIpLmNoZWNrb3V0Ln'
    'YxLkNyZWF0ZUNoZWNrb3V0U2Vzc2lvblJlcXVlc3QaKi5jaGVja291dC52MS5DcmVhdGVDaGVj'
    'a291dFNlc3Npb25SZXNwb25zZSIdgrUYGQoXY2hlY2tvdXRfc2Vzc2lvbl9jcmVhdGUShQEKEk'
    'dldENoZWNrb3V0U2Vzc2lvbhImLmNoZWNrb3V0LnYxLkdldENoZWNrb3V0U2Vzc2lvblJlcXVl'
    'c3QaJy5jaGVja291dC52MS5HZXRDaGVja291dFNlc3Npb25SZXNwb25zZSIekAIBgrUYFwoVY2'
    'hlY2tvdXRfc2Vzc2lvbl92aWV3EoEBChJDcmVhdGVDaGVja291dExpbmsSJi5jaGVja291dC52'
    'MS5DcmVhdGVDaGVja291dExpbmtSZXF1ZXN0GicuY2hlY2tvdXQudjEuQ3JlYXRlQ2hlY2tvdX'
    'RMaW5rUmVzcG9uc2UiGoK1GBYKFGNoZWNrb3V0X2xpbmtfY3JlYXRlGqcBgrUYogEKEHNlcnZp'
    'Y2VfY2hlY2tvdXQSF2NoZWNrb3V0X3Nlc3Npb25fY3JlYXRlEhVjaGVja291dF9zZXNzaW9uX3'
    'ZpZXcSFGNoZWNrb3V0X2xpbmtfY3JlYXRlGkgIARIXY2hlY2tvdXRfc2Vzc2lvbl9jcmVhdGUS'
    'FWNoZWNrb3V0X3Nlc3Npb25fdmlldxIUY2hlY2tvdXRfbGlua19jcmVhdGU=');
