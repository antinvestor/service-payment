//
//  Generated code. Do not modify.
//  source: checkout/v1/checkout.proto
//

import "package:connectrpc/connect.dart" as connect;
import "checkout.pb.dart" as checkoutv1checkout;

abstract final class CheckoutService {
  /// Fully-qualified name of the CheckoutService service.
  static const name = 'checkout.v1.CheckoutService';

  /// CreateCheckoutSession creates a single-use hosted payment session and
  /// returns the page URL the payer should be redirected to.
  static const createCheckoutSession = connect.Spec(
    '/$name/CreateCheckoutSession',
    connect.StreamType.unary,
    checkoutv1checkout.CreateCheckoutSessionRequest.new,
    checkoutv1checkout.CreateCheckoutSessionResponse.new,
  );

  /// GetCheckoutSession returns the current state of a checkout session;
  /// merchants poll this as one source of payment truth.
  static const getCheckoutSession = connect.Spec(
    '/$name/GetCheckoutSession',
    connect.StreamType.unary,
    checkoutv1checkout.GetCheckoutSessionRequest.new,
    checkoutv1checkout.GetCheckoutSessionResponse.new,
    idempotency: connect.Idempotency.noSideEffects,
  );

  /// CreateCheckoutLink creates a reusable, shareable link that spawns a
  /// fresh checkout session per visitor.
  static const createCheckoutLink = connect.Spec(
    '/$name/CreateCheckoutLink',
    connect.StreamType.unary,
    checkoutv1checkout.CreateCheckoutLinkRequest.new,
    checkoutv1checkout.CreateCheckoutLinkResponse.new,
  );
}
