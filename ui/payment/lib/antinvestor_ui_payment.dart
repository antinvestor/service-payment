/// Payment management UI library for Antinvestor.
///
/// Provides embeddable screens, widgets, and Riverpod providers for
/// searching, sending, receiving, and managing payments and payment links.
library;

// Providers
export 'src/providers/payment_transport_provider.dart';
export 'src/providers/payment_providers.dart';
export 'src/providers/payment_link_providers.dart';

// Widgets
export 'src/widgets/payment_status_badge.dart';
export 'src/widgets/payment_card.dart';
export 'src/widgets/payment_route_selector.dart';
export 'src/widgets/account_field.dart';
export 'src/widgets/payment_link_card.dart';

// Screens
export 'src/screens/payment_search_screen.dart';
export 'src/screens/payment_detail_screen.dart';
export 'src/screens/payment_send_screen.dart';
export 'src/screens/payment_receive_screen.dart';
export 'src/screens/payment_links_screen.dart';

// Routing
export 'src/routing/payment_route_module.dart';
