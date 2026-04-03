//
//  Generated code. Do not modify.
//  source: v1/billing.proto
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
import 'billing.pb.dart' as $10;
import 'billing.pbjson.dart';

export 'billing.pb.dart';

abstract class BillingServiceBase extends $pb.GeneratedService {
  $async.Future<$10.CreateCatalogVersionResponse> createCatalogVersion($pb.ServerContext ctx, $10.CreateCatalogVersionRequest request);
  $async.Future<$10.GetCatalogVersionResponse> getCatalogVersion($pb.ServerContext ctx, $10.GetCatalogVersionRequest request);
  $async.Future<$10.PublishCatalogVersionResponse> publishCatalogVersion($pb.ServerContext ctx, $10.PublishCatalogVersionRequest request);
  $async.Future<$10.SearchCatalogVersionsResponse> searchCatalogVersions($pb.ServerContext ctx, $9.SearchRequest request);
  $async.Future<$10.CreatePlanResponse> createPlan($pb.ServerContext ctx, $10.CreatePlanRequest request);
  $async.Future<$10.CreateComponentResponse> createComponent($pb.ServerContext ctx, $10.CreateComponentRequest request);
  $async.Future<$10.CreateTierResponse> createTier($pb.ServerContext ctx, $10.CreateTierRequest request);
  $async.Future<$10.CreateSubscriptionResponse> createSubscription($pb.ServerContext ctx, $10.CreateSubscriptionRequest request);
  $async.Future<$10.GetSubscriptionResponse> getSubscription($pb.ServerContext ctx, $10.GetSubscriptionRequest request);
  $async.Future<$10.CancelSubscriptionResponse> cancelSubscription($pb.ServerContext ctx, $10.CancelSubscriptionRequest request);
  $async.Future<$10.ListSubscriptionsResponse> listSubscriptions($pb.ServerContext ctx, $10.ListSubscriptionsRequest request);
  $async.Future<$10.IngestUsageEventResponse> ingestUsageEvent($pb.ServerContext ctx, $10.IngestUsageEventRequest request);
  $async.Future<$10.SearchUsageEventsResponse> searchUsageEvents($pb.ServerContext ctx, $9.SearchRequest request);
  $async.Future<$10.RunBillingResponse> runBilling($pb.ServerContext ctx, $10.RunBillingRequest request);
  $async.Future<$10.GetBillingRunResponse> getBillingRun($pb.ServerContext ctx, $10.GetBillingRunRequest request);
  $async.Future<$10.GetInvoiceResponse> getInvoice($pb.ServerContext ctx, $10.GetInvoiceRequest request);
  $async.Future<$10.IssueInvoiceResponse> issueInvoice($pb.ServerContext ctx, $10.IssueInvoiceRequest request);
  $async.Future<$10.VoidInvoiceResponse> voidInvoice($pb.ServerContext ctx, $10.VoidInvoiceRequest request);
  $async.Future<$10.RecordPaymentResponse> recordPayment($pb.ServerContext ctx, $10.RecordPaymentRequest request);
  $async.Future<$10.SearchInvoicesResponse> searchInvoices($pb.ServerContext ctx, $9.SearchRequest request);
  $async.Future<$10.GrantCreditResponse> grantCredit($pb.ServerContext ctx, $10.GrantCreditRequest request);
  $async.Future<$10.GetCreditBalanceResponse> getCreditBalance($pb.ServerContext ctx, $10.GetCreditBalanceRequest request);
  $async.Future<$10.CreateDiscountResponse> createDiscount($pb.ServerContext ctx, $10.CreateDiscountRequest request);
  $async.Future<$10.SearchDiscountsResponse> searchDiscounts($pb.ServerContext ctx, $9.SearchRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateCatalogVersion': return $10.CreateCatalogVersionRequest();
      case 'GetCatalogVersion': return $10.GetCatalogVersionRequest();
      case 'PublishCatalogVersion': return $10.PublishCatalogVersionRequest();
      case 'SearchCatalogVersions': return $9.SearchRequest();
      case 'CreatePlan': return $10.CreatePlanRequest();
      case 'CreateComponent': return $10.CreateComponentRequest();
      case 'CreateTier': return $10.CreateTierRequest();
      case 'CreateSubscription': return $10.CreateSubscriptionRequest();
      case 'GetSubscription': return $10.GetSubscriptionRequest();
      case 'CancelSubscription': return $10.CancelSubscriptionRequest();
      case 'ListSubscriptions': return $10.ListSubscriptionsRequest();
      case 'IngestUsageEvent': return $10.IngestUsageEventRequest();
      case 'SearchUsageEvents': return $9.SearchRequest();
      case 'RunBilling': return $10.RunBillingRequest();
      case 'GetBillingRun': return $10.GetBillingRunRequest();
      case 'GetInvoice': return $10.GetInvoiceRequest();
      case 'IssueInvoice': return $10.IssueInvoiceRequest();
      case 'VoidInvoice': return $10.VoidInvoiceRequest();
      case 'RecordPayment': return $10.RecordPaymentRequest();
      case 'SearchInvoices': return $9.SearchRequest();
      case 'GrantCredit': return $10.GrantCreditRequest();
      case 'GetCreditBalance': return $10.GetCreditBalanceRequest();
      case 'CreateDiscount': return $10.CreateDiscountRequest();
      case 'SearchDiscounts': return $9.SearchRequest();
      default: throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx, $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateCatalogVersion': return this.createCatalogVersion(ctx, request as $10.CreateCatalogVersionRequest);
      case 'GetCatalogVersion': return this.getCatalogVersion(ctx, request as $10.GetCatalogVersionRequest);
      case 'PublishCatalogVersion': return this.publishCatalogVersion(ctx, request as $10.PublishCatalogVersionRequest);
      case 'SearchCatalogVersions': return this.searchCatalogVersions(ctx, request as $9.SearchRequest);
      case 'CreatePlan': return this.createPlan(ctx, request as $10.CreatePlanRequest);
      case 'CreateComponent': return this.createComponent(ctx, request as $10.CreateComponentRequest);
      case 'CreateTier': return this.createTier(ctx, request as $10.CreateTierRequest);
      case 'CreateSubscription': return this.createSubscription(ctx, request as $10.CreateSubscriptionRequest);
      case 'GetSubscription': return this.getSubscription(ctx, request as $10.GetSubscriptionRequest);
      case 'CancelSubscription': return this.cancelSubscription(ctx, request as $10.CancelSubscriptionRequest);
      case 'ListSubscriptions': return this.listSubscriptions(ctx, request as $10.ListSubscriptionsRequest);
      case 'IngestUsageEvent': return this.ingestUsageEvent(ctx, request as $10.IngestUsageEventRequest);
      case 'SearchUsageEvents': return this.searchUsageEvents(ctx, request as $9.SearchRequest);
      case 'RunBilling': return this.runBilling(ctx, request as $10.RunBillingRequest);
      case 'GetBillingRun': return this.getBillingRun(ctx, request as $10.GetBillingRunRequest);
      case 'GetInvoice': return this.getInvoice(ctx, request as $10.GetInvoiceRequest);
      case 'IssueInvoice': return this.issueInvoice(ctx, request as $10.IssueInvoiceRequest);
      case 'VoidInvoice': return this.voidInvoice(ctx, request as $10.VoidInvoiceRequest);
      case 'RecordPayment': return this.recordPayment(ctx, request as $10.RecordPaymentRequest);
      case 'SearchInvoices': return this.searchInvoices(ctx, request as $9.SearchRequest);
      case 'GrantCredit': return this.grantCredit(ctx, request as $10.GrantCreditRequest);
      case 'GetCreditBalance': return this.getCreditBalance(ctx, request as $10.GetCreditBalanceRequest);
      case 'CreateDiscount': return this.createDiscount(ctx, request as $10.CreateDiscountRequest);
      case 'SearchDiscounts': return this.searchDiscounts(ctx, request as $9.SearchRequest);
      default: throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => BillingServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>> get $messageJson => BillingServiceBase$messageJson;
}

