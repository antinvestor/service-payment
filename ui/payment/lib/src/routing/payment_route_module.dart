import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/navigation/nav_items.dart';
import 'package:antinvestor_ui_core/routing/route_module.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../screens/payment_detail_screen.dart';
import '../screens/payment_links_screen.dart';
import '../screens/payment_receive_screen.dart';
import '../screens/payment_search_screen.dart';
import '../screens/payment_send_screen.dart';

/// Route module for payment management.
///
/// Registers the following routes:
/// - `/payments` - payment search and list
/// - `/payments/detail/:id` - view payment details
/// - `/payments/send` - send a payment
/// - `/payments/receive` - receive a payment
/// - `/payments/links` - manage payment links
class PaymentRouteModule extends RouteModule {
  @override
  String get moduleId => 'payment';

  @override
  List<RouteBase> buildRoutes() {
    return [
      GoRoute(
        path: '/payments',
        builder: (context, state) => const PaymentSearchScreen(),
        routes: [
          GoRoute(
            path: 'detail/:id',
            builder: (context, state) {
              final id = state.pathParameters['id'] ?? '';
              final extra = state.extra;
              final payment = extra is Payment ? extra : null;
              return PaymentDetailScreen(
                paymentId: id,
                initialPayment: payment,
              );
            },
          ),
          GoRoute(
            path: 'send',
            builder: (context, state) => const PaymentSendScreen(),
          ),
          GoRoute(
            path: 'receive',
            builder: (context, state) => const PaymentReceiveScreen(),
          ),
          GoRoute(
            path: 'links',
            builder: (context, state) => const PaymentLinksScreen(),
          ),
        ],
      ),
    ];
  }

  @override
  List<NavItem> buildNavItems() {
    return [
      const NavItem(
        id: 'payments',
        label: 'Payments',
        icon: Icons.payments_outlined,
        activeIcon: Icons.payments,
        route: '/payments',
        children: [
          NavItem(
            id: 'payments-list',
            label: 'All Payments',
            icon: Icons.list,
            route: '/payments',
          ),
          NavItem(
            id: 'payments-send',
            label: 'Send',
            icon: Icons.send,
            route: '/payments/send',
          ),
          NavItem(
            id: 'payments-receive',
            label: 'Receive',
            icon: Icons.call_received,
            route: '/payments/receive',
          ),
          NavItem(
            id: 'payments-links',
            label: 'Payment Links',
            icon: Icons.link,
            route: '/payments/links',
          ),
        ],
      ),
    ];
  }

  @override
  Map<String, Set<String>> get routePermissions => {
        '/payments': {'payment:read', 'admin'},
        '/payments/detail': {'payment:read', 'admin'},
        '/payments/send': {'payment:write', 'admin'},
        '/payments/receive': {'payment:write', 'admin'},
        '/payments/links': {'payment:read', 'admin'},
      };
}
