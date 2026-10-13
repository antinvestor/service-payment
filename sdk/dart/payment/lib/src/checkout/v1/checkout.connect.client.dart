//
//  Generated code. Do not modify.
//  source: checkout/v1/checkout.proto
//

import "package:connectrpc/connect.dart" as connect;
import "checkout.pb.dart" as checkoutv1checkout;
import "checkout.connect.spec.dart" as specs;

extension type CheckoutServiceClient (connect.Transport _transport) {
  /// CreateCheckoutSession creates a single-use hosted payment session and
  /// returns the page URL the payer should be redirected to.
  Future<checkoutv1checkout.CreateCheckoutSessionResponse> createCheckoutSession(
    checkoutv1checkout.CreateCheckoutSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CheckoutService.createCheckoutSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetCheckoutSession returns the current state of a checkout session;
  /// merchants poll this as one source of payment truth.
  Future<checkoutv1checkout.GetCheckoutSessionResponse> getCheckoutSession(
    checkoutv1checkout.GetCheckoutSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CheckoutService.getCheckoutSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateCheckoutLink creates a reusable, shareable link that spawns a
  /// fresh checkout session per visitor.
  Future<checkoutv1checkout.CreateCheckoutLinkResponse> createCheckoutLink(
    checkoutv1checkout.CreateCheckoutLinkRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CheckoutService.createCheckoutLink,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
