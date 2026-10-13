//
//  Generated code. Do not modify.
//  source: collection/v1/collection.proto
//

import "package:connectrpc/connect.dart" as connect;
import "collection.pb.dart" as collectionv1collection;

/// CollectionService is the merchant-facing simplified payment collection API.
abstract final class CollectionService {
  /// Fully-qualified name of the CollectionService service.
  static const name = 'collection.v1.CollectionService';

  /// CollectPayment opens hosted checkout for an issued invoice.
  /// Redirect the end user to data.page_url, then call ConfirmPayment with data.session_ref.
  static const collectPayment = connect.Spec(
    '/$name/CollectPayment',
    connect.StreamType.unary,
    collectionv1collection.CollectPaymentRequest.new,
    collectionv1collection.CollectPaymentResponse.new,
  );

  /// StartSubscription creates a subscription and opens checkout for any upfront fee.
  /// Redirect the end user to data.page_url when present; otherwise the subscription is already active.
  static const startSubscription = connect.Spec(
    '/$name/StartSubscription',
    connect.StreamType.unary,
    collectionv1collection.StartSubscriptionRequest.new,
    collectionv1collection.StartSubscriptionResponse.new,
  );

  /// ConfirmPayment captures a completed checkout session into billing.
  /// Source of truth is the checkout session status — never trust browser redirects alone.
  static const confirmPayment = connect.Spec(
    '/$name/ConfirmPayment',
    connect.StreamType.unary,
    collectionv1collection.ConfirmPaymentRequest.new,
    collectionv1collection.ConfirmPaymentResponse.new,
  );

  /// CancelSubscription cancels a subscription (and voids unpaid first invoices).
  static const cancelSubscription = connect.Spec(
    '/$name/CancelSubscription',
    connect.StreamType.unary,
    collectionv1collection.CancelSubscriptionRequest.new,
    collectionv1collection.CancelSubscriptionResponse.new,
  );
}
