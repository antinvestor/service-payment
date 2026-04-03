//
//  Generated code. Do not modify.
//  source: v1/payment.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../common/v1/common.pb.dart' as $9;
import 'payment.pb.dart' as $12;
import 'payment.pbjson.dart';

export 'payment.pb.dart';

abstract class PaymentServiceBase extends $pb.GeneratedService {
  $async.Future<$12.SendResponse> send($pb.ServerContext ctx, $12.SendRequest request);
  $async.Future<$12.ReceiveResponse> receive($pb.ServerContext ctx, $12.ReceiveRequest request);
  $async.Future<$12.InitiatePromptResponse> initiatePrompt($pb.ServerContext ctx, $12.InitiatePromptRequest request);
  $async.Future<$12.CreatePaymentLinkResponse> createPaymentLink($pb.ServerContext ctx, $12.CreatePaymentLinkRequest request);
  $async.Future<$9.StatusResponse> status($pb.ServerContext ctx, $9.StatusRequest request);
  $async.Future<$9.StatusUpdateResponse> statusUpdate($pb.ServerContext ctx, $9.StatusUpdateRequest request);
  $async.Future<$12.ReleaseResponse> release($pb.ServerContext ctx, $12.ReleaseRequest request);
  $async.Future<$12.SearchResponse> search($pb.ServerContext ctx, $9.SearchRequest request);
  $async.Future<$12.ReconcileResponse> reconcile($pb.ServerContext ctx, $12.ReconcileRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'Send': return $12.SendRequest();
      case 'Receive': return $12.ReceiveRequest();
      case 'InitiatePrompt': return $12.InitiatePromptRequest();
      case 'CreatePaymentLink': return $12.CreatePaymentLinkRequest();
      case 'Status': return $9.StatusRequest();
      case 'StatusUpdate': return $9.StatusUpdateRequest();
      case 'Release': return $12.ReleaseRequest();
      case 'Search': return $9.SearchRequest();
      case 'Reconcile': return $12.ReconcileRequest();
      default: throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx, $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'Send': return this.send(ctx, request as $12.SendRequest);
      case 'Receive': return this.receive(ctx, request as $12.ReceiveRequest);
      case 'InitiatePrompt': return this.initiatePrompt(ctx, request as $12.InitiatePromptRequest);
      case 'CreatePaymentLink': return this.createPaymentLink(ctx, request as $12.CreatePaymentLinkRequest);
      case 'Status': return this.status(ctx, request as $9.StatusRequest);
      case 'StatusUpdate': return this.statusUpdate(ctx, request as $9.StatusUpdateRequest);
      case 'Release': return this.release(ctx, request as $12.ReleaseRequest);
      case 'Search': return this.search(ctx, request as $9.SearchRequest);
      case 'Reconcile': return this.reconcile(ctx, request as $12.ReconcileRequest);
      default: throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => PaymentServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>> get $messageJson => PaymentServiceBase$messageJson;
}

