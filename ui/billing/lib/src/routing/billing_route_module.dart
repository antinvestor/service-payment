import 'package:antinvestor_ui_core/navigation/nav_items.dart';
import 'package:antinvestor_ui_core/permissions/permission_manifest.dart';
import 'package:antinvestor_ui_core/routing/route_module.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../screens/catalog_list_screen.dart';
import '../screens/catalog_detail_screen.dart';
import '../screens/subscription_list_screen.dart';
import '../screens/subscription_detail_screen.dart';
import '../screens/start_subscription_screen.dart';
import '../screens/payment_return_screen.dart';
import '../screens/invoice_list_screen.dart';
import '../screens/invoice_detail_screen.dart';
import '../screens/usage_events_screen.dart';
import '../screens/billing_run_screen.dart';
import '../screens/credit_screen.dart';
import '../screens/discount_list_screen.dart';

class BillingRouteModule extends RouteModule {
  @override
  String get moduleId => 'billing';

  @override
  List<RouteBase> buildRoutes() => [
        GoRoute(
          path: '/billing/catalogs',
          builder: (context, state) => const CatalogListScreen(),
          routes: [
            GoRoute(
              path: ':catalogId',
              builder: (context, state) => CatalogDetailScreen(
                catalogId: state.pathParameters['catalogId']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: '/billing/subscriptions',
          builder: (context, state) => const SubscriptionListScreen(),
          routes: [
            GoRoute(
              path: 'start',
              builder: (context, state) => StartSubscriptionScreen(
                catalogVersionId: state.uri.queryParameters['catalogVersionId'],
                planId: state.uri.queryParameters['planId'],
                currency: state.uri.queryParameters['currency'],
              ),
            ),
            GoRoute(
              path: ':subscriptionId',
              builder: (context, state) => SubscriptionDetailScreen(
                subscriptionId: state.pathParameters['subscriptionId']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: '/billing/invoices',
          builder: (context, state) => const InvoiceListScreen(),
          routes: [
            GoRoute(
              path: ':invoiceId',
              builder: (context, state) => InvoiceDetailScreen(
                invoiceId: state.pathParameters['invoiceId']!,
              ),
            ),
          ],
        ),
        // Hosted-checkout return landing (session + optional status query).
        GoRoute(
          path: '/billing/payment/return',
          builder: (context, state) => PaymentReturnScreen(
            sessionRef: state.uri.queryParameters['session'] ?? '',
            statusHint: state.uri.queryParameters['status'],
          ),
        ),
        GoRoute(
          path: '/billing/usage',
          builder: (context, state) => const UsageEventsScreen(),
        ),
        GoRoute(
          path: '/billing/runs',
          builder: (context, state) => const BillingRunScreen(),
        ),
        GoRoute(
          path: '/billing/credits',
          builder: (context, state) => const CreditScreen(),
        ),
        GoRoute(
          path: '/billing/discounts',
          builder: (context, state) => const DiscountListScreen(),
        ),
      ];

  @override
  List<NavItem> buildNavItems() => [
        NavItem(
          id: 'billing',
          label: 'Billing',
          icon: Icons.receipt_long_outlined,
          activeIcon: Icons.receipt_long,
          requiredPermissions: const {'catalog_view', 'subscription_view'},
          children: [
            const NavItem(
              id: 'billing_catalogs',
              label: 'Catalogs',
              icon: Icons.menu_book_outlined,
              route: '/billing/catalogs',
              requiredPermissions: {'catalog_view'},
            ),
            const NavItem(
              id: 'billing_subscriptions',
              label: 'Subscriptions',
              icon: Icons.autorenew_outlined,
              route: '/billing/subscriptions',
              requiredPermissions: {'subscription_view'},
            ),
            const NavItem(
              id: 'billing_invoices',
              label: 'Invoices',
              icon: Icons.description_outlined,
              route: '/billing/invoices',
              requiredPermissions: {'invoice_view'},
            ),
            const NavItem(
              id: 'billing_usage',
              label: 'Usage Events',
              icon: Icons.bar_chart_outlined,
              route: '/billing/usage',
              requiredPermissions: {'usage_view'},
            ),
            const NavItem(
              id: 'billing_runs',
              label: 'Billing Runs',
              icon: Icons.play_circle_outline,
              route: '/billing/runs',
              requiredPermissions: {'billing_run_execute'},
            ),
            const NavItem(
              id: 'billing_credits',
              label: 'Credits',
              icon: Icons.account_balance_wallet_outlined,
              route: '/billing/credits',
              requiredPermissions: {'credit_manage'},
            ),
            const NavItem(
              id: 'billing_discounts',
              label: 'Discounts',
              icon: Icons.local_offer_outlined,
              route: '/billing/discounts',
              requiredPermissions: {'discount_view'},
            ),
          ],
        ),
      ];

  @override
  Map<String, Set<String>> get routePermissions => {
        '/billing/catalogs': {'catalog_view'},
        '/billing/subscriptions': {'subscription_view'},
        '/billing/subscriptions/start': {'subscription_manage'},
        '/billing/invoices': {'invoice_view'},
        '/billing/payment/return': {'payment_collect'},
        '/billing/usage': {'usage_view'},
        '/billing/runs': {'billing_run_execute'},
        '/billing/credits': {'credit_manage'},
        '/billing/discounts': {'discount_view'},
      };

  @override
  PermissionManifest get permissionManifest => const PermissionManifest(
        namespace: 'service_billing',
        permissions: [
          PermissionEntry(
            key: 'catalog_view',
            label: 'View Catalogs',
            scope: PermissionScope.feature,
          ),
          PermissionEntry(
            key: 'catalog_manage',
            label: 'Manage Catalogs',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'subscription_view',
            label: 'View Subscriptions',
            scope: PermissionScope.feature,
          ),
          PermissionEntry(
            key: 'subscription_manage',
            label: 'Manage Subscriptions',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'invoice_view',
            label: 'View Invoices',
            scope: PermissionScope.feature,
          ),
          PermissionEntry(
            key: 'invoice_manage',
            label: 'Manage Invoices',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'payment_collect',
            label: 'Collect Payments',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'usage_view',
            label: 'View Usage Events',
            scope: PermissionScope.feature,
          ),
          PermissionEntry(
            key: 'billing_run_execute',
            label: 'Execute Billing Runs',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'credit_manage',
            label: 'Manage Credits',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'discount_view',
            label: 'View Discounts',
            scope: PermissionScope.feature,
          ),
          PermissionEntry(
            key: 'discount_manage',
            label: 'Manage Discounts',
            scope: PermissionScope.action,
          ),
        ],
      );
}
