//
//  Generated code. Do not modify.
//  source: collection/v1/collection.proto
//

import "package:connectrpc/connect.dart" as connect;
import "collection.pb.dart" as collectionv1collection;
import "collection.connect.spec.dart" as specs;

extension type CollectionServiceClient (connect.Transport _transport) {
  /// CollectPayment opens hosted checkout for an issued invoice.
  /// Redirect the end user to data.page_url, then call ConfirmPayment with data.session_ref.
  Future<collectionv1collection.CollectPaymentResponse> collectPayment(
    collectionv1collection.CollectPaymentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CollectionService.collectPayment,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// StartSubscription creates a subscription and opens checkout for any upfront fee.
  /// Redirect the end user to data.page_url when present; otherwise the subscription is already active.
  Future<collectionv1collection.StartSubscriptionResponse> startSubscription(
    collectionv1collection.StartSubscriptionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CollectionService.startSubscription,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ConfirmPayment captures a completed checkout session into billing.
  /// Source of truth is the checkout session status — never trust browser redirects alone.
  Future<collectionv1collection.ConfirmPaymentResponse> confirmPayment(
    collectionv1collection.ConfirmPaymentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CollectionService.confirmPayment,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SyncProviderSubscription mirrors provider-managed subscription state
  /// (webhooks from Polar, Stripe) into billing. Grant provider_subscription_sync
  /// to integration service accounts only.
  Future<collectionv1collection.SyncProviderSubscriptionResponse> syncProviderSubscription(
    collectionv1collection.SyncProviderSubscriptionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CollectionService.syncProviderSubscription,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CancelSubscription cancels a subscription (and voids unpaid first invoices).
  Future<collectionv1collection.CancelSubscriptionResponse> cancelSubscription(
    collectionv1collection.CancelSubscriptionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CollectionService.cancelSubscription,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
