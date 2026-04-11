import 'package:antinvestor_ui_core/navigation/nav_items.dart';
import 'package:antinvestor_ui_core/routing/route_module.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../screens/catalog_list_screen.dart';
import '../screens/catalog_detail_screen.dart';
import '../screens/subscription_list_screen.dart';
import '../screens/subscription_detail_screen.dart';
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
          children: [
            const NavItem(
              id: 'billing_catalogs',
              label: 'Catalogs',
              icon: Icons.menu_book_outlined,
              route: '/billing/catalogs',
            ),
            const NavItem(
              id: 'billing_subscriptions',
              label: 'Subscriptions',
              icon: Icons.autorenew_outlined,
              route: '/billing/subscriptions',
            ),
            const NavItem(
              id: 'billing_invoices',
              label: 'Invoices',
              icon: Icons.description_outlined,
              route: '/billing/invoices',
            ),
            const NavItem(
              id: 'billing_usage',
              label: 'Usage Events',
              icon: Icons.bar_chart_outlined,
              route: '/billing/usage',
            ),
            const NavItem(
              id: 'billing_runs',
              label: 'Billing Runs',
              icon: Icons.play_circle_outline,
              route: '/billing/runs',
            ),
            const NavItem(
              id: 'billing_credits',
              label: 'Credits',
              icon: Icons.account_balance_wallet_outlined,
              route: '/billing/credits',
            ),
            const NavItem(
              id: 'billing_discounts',
              label: 'Discounts',
              icon: Icons.local_offer_outlined,
              route: '/billing/discounts',
            ),
          ],
        ),
      ];

  @override
  Map<String, Set<String>> get routePermissions => {
        '/billing/catalogs': {'billing:read', 'admin'},
        '/billing/subscriptions': {'billing:read', 'admin'},
        '/billing/invoices': {'billing:read', 'admin'},
        '/billing/usage': {'billing:read', 'admin'},
        '/billing/runs': {'billing:write', 'admin'},
        '/billing/credits': {'billing:write', 'admin'},
        '/billing/discounts': {'billing:read', 'admin'},
      };
}
