import 'package:antinvestor_ui_core/navigation/nav_items.dart';
import 'package:antinvestor_ui_core/permissions/permission_manifest.dart';
import 'package:antinvestor_ui_core/routing/route_module.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../screens/account_detail_screen.dart';
import '../screens/account_list_screen.dart';
import '../screens/ledger_detail_screen.dart';
import '../screens/ledger_list_screen.dart';
import '../screens/transaction_detail_screen.dart';
import '../screens/transaction_list_screen.dart';

/// Route module for ledger management.
///
/// Registers the following routes:
/// - `/ledger/ledgers` - ledger list
/// - `/ledger/ledgers/:id` - ledger details
/// - `/ledger/accounts` - account list
/// - `/ledger/accounts/:id` - account details
/// - `/ledger/transactions` - transaction list
/// - `/ledger/transactions/:id` - transaction details
class LedgerRouteModule extends RouteModule {
  @override
  String get moduleId => 'ledger';

  @override
  List<RouteBase> buildRoutes() {
    return [
      GoRoute(
        path: '/ledger/ledgers',
        builder: (context, state) => const LedgerListScreen(),
        routes: [
          GoRoute(
            path: ':id',
            builder: (context, state) {
              final id = state.pathParameters['id'] ?? '';
              return LedgerDetailScreen(ledgerId: id);
            },
          ),
        ],
      ),
      GoRoute(
        path: '/ledger/accounts',
        builder: (context, state) => const AccountListScreen(),
        routes: [
          GoRoute(
            path: ':id',
            builder: (context, state) {
              final id = state.pathParameters['id'] ?? '';
              return AccountDetailScreen(accountId: id);
            },
          ),
        ],
      ),
      GoRoute(
        path: '/ledger/transactions',
        builder: (context, state) => const TransactionListScreen(),
        routes: [
          GoRoute(
            path: ':id',
            builder: (context, state) {
              final id = state.pathParameters['id'] ?? '';
              return TransactionDetailScreen(transactionId: id);
            },
          ),
        ],
      ),
    ];
  }

  @override
  List<NavItem> buildNavItems() {
    return [
      const NavItem(
        id: 'ledger',
        label: 'Ledger',
        icon: Icons.account_tree_outlined,
        activeIcon: Icons.account_tree,
        route: '/ledger/ledgers',
        requiredPermissions: {'ledger_view'},
        children: [
          NavItem(
            id: 'ledger-ledgers',
            label: 'Ledgers',
            icon: Icons.account_tree,
            route: '/ledger/ledgers',
            requiredPermissions: {'ledger_view'},
          ),
          NavItem(
            id: 'ledger-accounts',
            label: 'Accounts',
            icon: Icons.account_balance_wallet,
            route: '/ledger/accounts',
            requiredPermissions: {'account_view'},
          ),
          NavItem(
            id: 'ledger-transactions',
            label: 'Transactions',
            icon: Icons.receipt_long,
            route: '/ledger/transactions',
            requiredPermissions: {'transaction_view'},
          ),
        ],
      ),
    ];
  }

  @override
  Map<String, Set<String>> get routePermissions => {
        '/ledger/ledgers': {'ledger_view'},
        '/ledger/ledgers/': {'ledger_view'},
        '/ledger/accounts': {'account_view'},
        '/ledger/accounts/': {'account_view'},
        '/ledger/transactions': {'transaction_view'},
        '/ledger/transactions/': {'transaction_view'},
      };

  @override
  PermissionManifest get permissionManifest => const PermissionManifest(
        namespace: 'service_ledger',
        permissions: [
          PermissionEntry(
            key: 'ledger_view',
            label: 'View Ledgers',
            scope: PermissionScope.service,
          ),
          PermissionEntry(
            key: 'ledger_create',
            label: 'Create Ledgers',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'account_view',
            label: 'View Accounts',
            scope: PermissionScope.feature,
          ),
          PermissionEntry(
            key: 'account_create',
            label: 'Create Accounts',
            scope: PermissionScope.action,
          ),
          PermissionEntry(
            key: 'transaction_view',
            label: 'View Transactions',
            scope: PermissionScope.feature,
          ),
          PermissionEntry(
            key: 'transaction_create',
            label: 'Create Transactions',
            scope: PermissionScope.action,
          ),
        ],
      );
}
