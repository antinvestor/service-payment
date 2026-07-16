import 'dart:convert';

import 'package:connectrpc/connect.dart';
import 'package:http/http.dart' as http;

/// Result of CollectPayment / StartSubscription.
class CollectionResult {
  const CollectionResult({
    this.pageUrl = '',
    this.sessionRef = '',
    this.invoiceId = '',
    this.subscriptionId = '',
    this.alreadyComplete = false,
  });

  final String pageUrl;
  final String sessionRef;
  final String invoiceId;
  final String subscriptionId;
  final bool alreadyComplete;

  factory CollectionResult.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>? ?? json;
    return CollectionResult(
      pageUrl: data['pageUrl'] as String? ?? data['page_url'] as String? ?? '',
      sessionRef:
          data['sessionRef'] as String? ?? data['session_ref'] as String? ?? '',
      invoiceId:
          data['invoiceId'] as String? ?? data['invoice_id'] as String? ?? '',
      subscriptionId: data['subscriptionId'] as String? ??
          data['subscription_id'] as String? ??
          '',
      alreadyComplete: data['alreadyComplete'] as bool? ??
          data['already_complete'] as bool? ??
          false,
    );
  }
}

/// Result of ConfirmPayment.
class ConfirmPaymentResult {
  const ConfirmPaymentResult({
    this.invoiceId = '',
    this.invoiceState = '',
    this.subscriptionId = '',
    this.subscriptionState = '',
    this.paid = false,
  });

  final String invoiceId;
  final String invoiceState;
  final String subscriptionId;
  final String subscriptionState;
  final bool paid;

  factory ConfirmPaymentResult.fromJson(Map<String, dynamic> json) {
    return ConfirmPaymentResult(
      invoiceId:
          json['invoiceId'] as String? ?? json['invoice_id'] as String? ?? '',
      invoiceState: json['invoiceState'] as String? ??
          json['invoice_state'] as String? ??
          '',
      subscriptionId: json['subscriptionId'] as String? ??
          json['subscription_id'] as String? ??
          '',
      subscriptionState: json['subscriptionState'] as String? ??
          json['subscription_state'] as String? ??
          '',
      paid: json['paid'] as bool? ?? false,
    );
  }
}

/// Result of CancelSubscription.
class CancelCollectionResult {
  const CancelCollectionResult({
    this.subscriptionId = '',
    this.subscriptionState = '',
    this.voidedInvoiceId = '',
  });

  final String subscriptionId;
  final String subscriptionState;
  final String voidedInvoiceId;

  factory CancelCollectionResult.fromJson(Map<String, dynamic> json) {
    return CancelCollectionResult(
      subscriptionId: json['subscriptionId'] as String? ??
          json['subscription_id'] as String? ??
          '',
      subscriptionState: json['subscriptionState'] as String? ??
          json['subscription_state'] as String? ??
          '',
      voidedInvoiceId: json['voidedInvoiceId'] as String? ??
          json['voided_invoice_id'] as String? ??
          '',
    );
  }
}

/// Lightweight Connect-JSON client for collection.v1.CollectionService.
///
/// Uses the same billing base URL and bearer token as [BillingServiceClient].
/// Protocol is Connect unary over HTTP with application/json bodies so we do
/// not depend on published protobuf stubs for this service yet.
class CollectionClient {
  CollectionClient({
    required this.baseUrl,
    required this.tokenProvider,
    http.Client? httpClient,
  }) : _http = httpClient ?? http.Client();

  final String baseUrl;
  final Future<String?> Function() tokenProvider;
  final http.Client _http;

  static const _service = 'collection.v1.CollectionService';

  Future<CollectionResult> collectPayment({
    required String invoiceId,
    String returnUrl = '',
    List<String> methods = const [],
  }) async {
    final body = <String, dynamic>{
      'invoiceId': invoiceId,
      if (returnUrl.isNotEmpty) 'returnUrl': returnUrl,
      if (methods.isNotEmpty) 'methods': methods,
    };
    final json = await _unary('CollectPayment', body);
    return CollectionResult.fromJson(json);
  }

  Future<CollectionResult> startSubscription({
    required String profileId,
    required String planId,
    required String catalogVersionId,
    required String currency,
    String returnUrl = '',
    String payerDisplayName = '',
    List<String> methods = const [],
  }) async {
    final body = <String, dynamic>{
      'profileId': profileId,
      'planId': planId,
      'catalogVersionId': catalogVersionId,
      'currency': currency,
      if (returnUrl.isNotEmpty) 'returnUrl': returnUrl,
      if (payerDisplayName.isNotEmpty) 'payerDisplayName': payerDisplayName,
      if (methods.isNotEmpty) 'methods': methods,
    };
    final json = await _unary('StartSubscription', body);
    return CollectionResult.fromJson(json);
  }

  Future<ConfirmPaymentResult> confirmPayment({
    required String sessionRef,
  }) async {
    final json = await _unary('ConfirmPayment', {'sessionRef': sessionRef});
    return ConfirmPaymentResult.fromJson(json);
  }

  Future<CancelCollectionResult> cancelSubscription({
    required String subscriptionId,
  }) async {
    final json = await _unary('CancelSubscription', {
      'subscriptionId': subscriptionId,
    });
    return CancelCollectionResult.fromJson(json);
  }

  Future<Map<String, dynamic>> _unary(
    String method,
    Map<String, dynamic> body,
  ) async {
    final uri = Uri.parse(
      '${baseUrl.replaceAll(RegExp(r'/$'), '')}/$_service/$method',
    );
    final token = await tokenProvider();
    final headers = <String, String>{
      'content-type': 'application/json',
      'connect-protocol-version': '1',
      if (token != null && token.isNotEmpty) 'authorization': 'Bearer $token',
    };

    final response = await _http.post(
      uri,
      headers: headers,
      body: jsonEncode(body),
    );

    final decoded = jsonDecode(response.body);
    if (decoded is! Map<String, dynamic>) {
      throw ConnectException(
        Code.unknown,
        'Unexpected collection response shape',
      );
    }

    if (response.statusCode >= 400) {
      final err = decoded['error'] as Map<String, dynamic>?;
      final code = _parseCode(err?['code'] as String?);
      final message = err?['message'] as String? ??
          decoded['message'] as String? ??
          'Collection request failed (${response.statusCode})';
      throw ConnectException(code, message);
    }

    return decoded;
  }

  Code _parseCode(String? raw) {
    if (raw == null || raw.isEmpty) return Code.unknown;
    return switch (raw) {
      'invalid_argument' => Code.invalidArgument,
      'not_found' => Code.notFound,
      'already_exists' => Code.alreadyExists,
      'permission_denied' => Code.permissionDenied,
      'failed_precondition' => Code.failedPrecondition,
      'unauthenticated' => Code.unauthenticated,
      'resource_exhausted' => Code.resourceExhausted,
      'unavailable' => Code.unavailable,
      'internal' => Code.internal,
      _ => Code.unknown,
    };
  }
}
