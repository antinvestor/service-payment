/// Ledger management UI library for Antinvestor.
///
/// Provides embeddable screens, widgets, and Riverpod providers for
/// managing ledgers, accounts, transactions, and transaction entries.
library;

// Providers
export 'src/providers/account_providers.dart';
export 'src/providers/book_providers.dart';
export 'src/providers/ledger_providers.dart';
export 'src/providers/ledger_transport_provider.dart';
export 'src/providers/report_providers.dart';
export 'src/providers/transaction_providers.dart';

// Widgets
export 'src/widgets/account_balance_card.dart';
export 'src/widgets/ledger_tree_view.dart';
export 'src/widgets/ledger_type_badge.dart';
export 'src/widgets/transaction_entry_row.dart';
export 'src/widgets/transaction_type_badge.dart';

// Screens
export 'src/screens/account_detail_screen.dart';
export 'src/screens/account_list_screen.dart';
export 'src/screens/account_statement_screen.dart';
export 'src/screens/book_detail_screen.dart';
export 'src/screens/book_form_screen.dart';
export 'src/screens/book_list_screen.dart';
export 'src/screens/ledger_detail_screen.dart';
export 'src/screens/ledger_list_screen.dart';
export 'src/screens/transaction_detail_screen.dart';
export 'src/screens/transaction_list_screen.dart';
export 'src/screens/trial_balance_screen.dart';

// Export utilities
export 'src/export/report_export.dart';

// Routing
export 'src/routing/ledger_route_module.dart';
