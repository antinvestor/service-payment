// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0

import 'package:antinvestor_api_common/antinvestor_api_common.dart';
import 'package:connectrpc/connect.dart' show Interceptor;
import '../antinvestor_api_payment.dart';

const String defaultPaymentEndpoint = 'https://payment.antinvestor.com';

Future<ConnectClientBase<PaymentServiceClient>> newPaymentClient({
  required TransportFactory createTransport,
  String? endpoint,
  TokenManager? tokenManager,
  TokenRefreshCallback? onTokenRefresh,
  List<Interceptor>? additionalInterceptors,
  bool noAuth = false,
}) {
  return newClient<PaymentServiceClient>(
    defaultEndpoint: defaultPaymentEndpoint,
    createServiceClient: PaymentServiceClient.new,
    createTransport: createTransport,
    endpoint: endpoint,
    tokenManager: tokenManager,
    onTokenRefresh: onTokenRefresh,
    additionalInterceptors: additionalInterceptors,
    noAuth: noAuth,
  );
}

typedef PaymentClient = ConnectClientBase<PaymentServiceClient>;
