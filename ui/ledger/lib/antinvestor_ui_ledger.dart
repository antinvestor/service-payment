/// Ledger management UI library for Antinvestor.
///
/// Provides embeddable screens, widgets, and Riverpod providers for
/// managing ledgers, accounts, transactions, and transaction entries.
library antinvestor_ui_ledger;

// Providers
export 'src/providers/ledger_transport_provider.dart';
export 'src/providers/ledger_providers.dart';
export 'src/providers/account_providers.dart';
export 'src/providers/transaction_providers.dart';

// Widgets
export 'src/widgets/ledger_type_badge.dart';
export 'src/widgets/account_balance_card.dart';
export 'src/widgets/transaction_entry_row.dart';
export 'src/widgets/transaction_type_badge.dart';
export 'src/widgets/ledger_tree_view.dart';

// Screens
export 'src/screens/ledger_list_screen.dart';
export 'src/screens/ledger_detail_screen.dart';
export 'src/screens/account_list_screen.dart';
export 'src/screens/account_detail_screen.dart';
export 'src/screens/transaction_list_screen.dart';
export 'src/screens/transaction_detail_screen.dart';

// Routing
export 'src/routing/ledger_route_module.dart';
