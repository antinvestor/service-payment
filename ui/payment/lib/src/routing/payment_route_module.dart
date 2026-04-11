import 'package:antinvestor_api_payment/antinvestor_api_payment.dart';
import 'package:antinvestor_ui_core/navigation/nav_items.dart';
import 'package:antinvestor_ui_core/permissions/permission_manifest.dart';
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
        requiredPermissions: {'payment_search'},
        children: [
          NavItem(
            id: 'payments-list',
            label: 'All Payments',
            icon: Icons.list,
            route: '/payments',
            requiredPermissions: {'payment_search'},
          ),
          NavItem(
            id: 'payments-send',
            label: 'Send',
            icon: Icons.send,
            route: '/payments/send',
            requiredPermissions: {'payment_send'},
          ),
          NavItem(
            id: 'payments-receive',
            label: 'Receive',
            icon: Icons.call_received,
            route: '/payments/receive',
            requiredPermissions: {'payment_receive'},
          ),
          NavItem(
            id: 'payments-links',
            label: 'Payment Links',
            icon: Icons.link,
            route: '/payments/links',
            requiredPermissions: {'payment_link_create'},
          ),
        ],
      ),
    ];
  }

  @override
  Map<String, Set<String>> get routePermissions => {
        '/payments': {'payment_search'},
        '/payments/detail': {'payment_search'},
        '/payments/send': {'payment_send'},
        '/payments/receive': {'payment_receive'},
        '/payments/links': {'payment_link_create'},
      };

  @override
  PermissionManifest get permissionManifest => const PermissionManifest(
        namespace: 'service_payment',
        permissions: [
          PermissionEntry(
            key: 'payment_search',
            label: 'Search Payments',
            scope: PermissionScope.service,
          ),
          PermissionEntry(
            key: 'payment_send',
            label: 'Send Payments',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'payment_receive',
            label: 'Receive Payments',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'payment_link_create',
            label: 'Create Payment Links',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'payment_reconcile',
            label: 'Reconcile Payments',
            scope: PermissionScope.action,
          ),
        ],
      );
}
