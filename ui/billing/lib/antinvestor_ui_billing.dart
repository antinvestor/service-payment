/// Billing management UI library for Antinvestor.
///
/// Provides embeddable screens, widgets, and Riverpod providers for
/// catalogs, subscriptions, usage-based billing, invoicing, credits,
/// and discounts.
library;

// Providers
export 'src/providers/billing_transport_provider.dart';
export 'src/providers/catalog_providers.dart';
export 'src/providers/subscription_providers.dart';
export 'src/providers/usage_providers.dart';
export 'src/providers/invoice_providers.dart';
export 'src/providers/billing_run_providers.dart';
export 'src/providers/credit_providers.dart';
export 'src/providers/discount_providers.dart';
export 'src/providers/collection_providers.dart';

// API
export 'src/api/collection_client.dart';

// Widgets
export 'src/widgets/subscription_state_badge.dart';
export 'src/widgets/invoice_state_badge.dart';
export 'src/widgets/billing_run_state_badge.dart';
export 'src/widgets/invoice_card.dart';
export 'src/widgets/subscription_card.dart';
export 'src/widgets/pricing_model_badge.dart';
export 'src/widgets/invoice_line_tile.dart';
export 'src/widgets/usage_event_tile.dart';
export 'src/widgets/payment_method_picker.dart';

// Screens
export 'src/screens/catalog_list_screen.dart';
export 'src/screens/catalog_detail_screen.dart';
export 'src/screens/subscription_list_screen.dart';
export 'src/screens/subscription_detail_screen.dart';
export 'src/screens/start_subscription_screen.dart';
export 'src/screens/payment_return_screen.dart';
export 'src/screens/invoice_list_screen.dart';
export 'src/screens/invoice_detail_screen.dart';
export 'src/screens/usage_events_screen.dart';
export 'src/screens/billing_run_screen.dart';
export 'src/screens/credit_screen.dart';
export 'src/screens/discount_list_screen.dart';

// Routing
export 'src/routing/billing_route_module.dart';
