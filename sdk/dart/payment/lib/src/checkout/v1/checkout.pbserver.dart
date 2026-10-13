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

import 'checkout.pb.dart' as $1;
import 'checkout.pbjson.dart';

export 'checkout.pb.dart';

abstract class CheckoutServiceBase extends $pb.GeneratedService {
  $async.Future<$1.CreateCheckoutSessionResponse> createCheckoutSession(
      $pb.ServerContext ctx, $1.CreateCheckoutSessionRequest request);
  $async.Future<$1.GetCheckoutSessionResponse> getCheckoutSession(
      $pb.ServerContext ctx, $1.GetCheckoutSessionRequest request);
  $async.Future<$1.CreateCheckoutLinkResponse> createCheckoutLink(
      $pb.ServerContext ctx, $1.CreateCheckoutLinkRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateCheckoutSession':
        return $1.CreateCheckoutSessionRequest();
      case 'GetCheckoutSession':
        return $1.GetCheckoutSessionRequest();
      case 'CreateCheckoutLink':
        return $1.CreateCheckoutLinkRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateCheckoutSession':
        return createCheckoutSession(
            ctx, request as $1.CreateCheckoutSessionRequest);
      case 'GetCheckoutSession':
        return getCheckoutSession(ctx, request as $1.GetCheckoutSessionRequest);
      case 'CreateCheckoutLink':
        return createCheckoutLink(ctx, request as $1.CreateCheckoutLinkRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => CheckoutServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => CheckoutServiceBase$messageJson;
}
