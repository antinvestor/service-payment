// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import 'package:antinvestor_ui_core/navigation/nav_items.dart';
import 'package:antinvestor_ui_core/permissions/permission_manifest.dart';
import 'package:antinvestor_ui_core/routing/route_module.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../screens/account_detail_screen.dart';
import '../screens/account_list_screen.dart';
import '../screens/account_statement_screen.dart';
import '../screens/ledger_detail_screen.dart';
import '../screens/ledger_list_screen.dart';
import '../screens/transaction_detail_screen.dart';
import '../screens/transaction_list_screen.dart';
import '../screens/trial_balance_screen.dart';

/// Route module for ledger management.
///
/// Registers the following routes:
/// - `/ledger/ledgers` - ledger (chart-of-accounts) list
/// - `/ledger/ledgers/:id` - ledger details
/// - `/ledger/accounts` - account list
/// - `/ledger/accounts/:id` - account details
/// - `/ledger/accounts/:id/statement` - account statement (period + export)
/// - `/ledger/transactions` - transaction list
/// - `/ledger/transactions/:id` - transaction details with lifecycle actions
/// - `/ledger/reports/trial-balance` - trial balance report (filters + export)
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
            routes: [
              GoRoute(
                path: 'statement',
                builder: (context, state) {
                  final id = state.pathParameters['id'] ?? '';
                  return AccountStatementScreen(accountId: id);
                },
              ),
            ],
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
      GoRoute(
        path: '/ledger/reports/trial-balance',
        builder: (context, state) {
          final ledgerId = state.uri.queryParameters['ledgerId'] ?? '';
          return TrialBalanceScreen(initialLedgerId: ledgerId);
        },
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
            label: 'Chart of accounts',
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
          NavItem(
            id: 'ledger-trial-balance',
            label: 'Trial balance',
            icon: Icons.balance,
            route: '/ledger/reports/trial-balance',
            requiredPermissions: {'report_view'},
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
    '/ledger/accounts//statement': {'report_view'},
    '/ledger/transactions': {'transaction_view'},
    '/ledger/transactions/': {'transaction_view'},
    '/ledger/reports/trial-balance': {'report_view'},
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
      PermissionEntry(
        key: 'report_view',
        label: 'View Reports',
        scope: PermissionScope.feature,
      ),
    ],
  );
}
