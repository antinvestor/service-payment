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

import 'collection.pb.dart' as $1;
import 'collection.pbjson.dart';

export 'collection.pb.dart';

abstract class CollectionServiceBase extends $pb.GeneratedService {
  $async.Future<$1.CollectPaymentResponse> collectPayment(
      $pb.ServerContext ctx, $1.CollectPaymentRequest request);
  $async.Future<$1.StartSubscriptionResponse> startSubscription(
      $pb.ServerContext ctx, $1.StartSubscriptionRequest request);
  $async.Future<$1.ConfirmPaymentResponse> confirmPayment(
      $pb.ServerContext ctx, $1.ConfirmPaymentRequest request);
  $async.Future<$1.SyncProviderSubscriptionResponse> syncProviderSubscription(
      $pb.ServerContext ctx, $1.SyncProviderSubscriptionRequest request);
  $async.Future<$1.CancelSubscriptionResponse> cancelSubscription(
      $pb.ServerContext ctx, $1.CancelSubscriptionRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CollectPayment':
        return $1.CollectPaymentRequest();
      case 'StartSubscription':
        return $1.StartSubscriptionRequest();
      case 'ConfirmPayment':
        return $1.ConfirmPaymentRequest();
      case 'SyncProviderSubscription':
        return $1.SyncProviderSubscriptionRequest();
      case 'CancelSubscription':
        return $1.CancelSubscriptionRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CollectPayment':
        return collectPayment(ctx, request as $1.CollectPaymentRequest);
      case 'StartSubscription':
        return startSubscription(ctx, request as $1.StartSubscriptionRequest);
      case 'ConfirmPayment':
        return confirmPayment(ctx, request as $1.ConfirmPaymentRequest);
      case 'SyncProviderSubscription':
        return syncProviderSubscription(
            ctx, request as $1.SyncProviderSubscriptionRequest);
      case 'CancelSubscription':
        return cancelSubscription(ctx, request as $1.CancelSubscriptionRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      CollectionServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => CollectionServiceBase$messageJson;
}
