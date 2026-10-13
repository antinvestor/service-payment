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

import 'collection.pb.dart' as $0;
import 'collection.pbjson.dart';

export 'collection.pb.dart';

abstract class CollectionServiceBase extends $pb.GeneratedService {
  $async.Future<$0.CollectPaymentResponse> collectPayment(
      $pb.ServerContext ctx, $0.CollectPaymentRequest request);
  $async.Future<$0.StartSubscriptionResponse> startSubscription(
      $pb.ServerContext ctx, $0.StartSubscriptionRequest request);
  $async.Future<$0.ConfirmPaymentResponse> confirmPayment(
      $pb.ServerContext ctx, $0.ConfirmPaymentRequest request);
  $async.Future<$0.CancelSubscriptionResponse> cancelSubscription(
      $pb.ServerContext ctx, $0.CancelSubscriptionRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CollectPayment':
        return $0.CollectPaymentRequest();
      case 'StartSubscription':
        return $0.StartSubscriptionRequest();
      case 'ConfirmPayment':
        return $0.ConfirmPaymentRequest();
      case 'CancelSubscription':
        return $0.CancelSubscriptionRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CollectPayment':
        return collectPayment(ctx, request as $0.CollectPaymentRequest);
      case 'StartSubscription':
        return startSubscription(ctx, request as $0.StartSubscriptionRequest);
      case 'ConfirmPayment':
        return confirmPayment(ctx, request as $0.ConfirmPaymentRequest);
      case 'CancelSubscription':
        return cancelSubscription(ctx, request as $0.CancelSubscriptionRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      CollectionServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => CollectionServiceBase$messageJson;
}
