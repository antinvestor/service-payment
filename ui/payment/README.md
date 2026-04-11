# antinvestor_ui_payment

Embeddable payment management UI for Antinvestor applications. Provides screens and widgets for searching, sending, receiving, and managing payments and payment links.

## Installation

```yaml
dependencies:
  antinvestor_ui_payment: ^0.1.0
```

## Features

- **Payment Search**: Paginated list with filters and status tracking
- **Payment Detail**: Full payment information with status timeline
- **Send/Receive**: Guided payment send and receive flows
- **Payment Links**: Create and manage shareable payment links
- **Embeddable Widgets**: `PaymentStatusBadge`, `PaymentCard`, `PaymentRouteSelector`, `AccountField`, `PaymentLinkCard`
- **Routing**: `PaymentRouteModule` with GoRouter integration

## Usage

```dart
import 'package:antinvestor_ui_payment/antinvestor_ui_payment.dart';

// Payment status indicator
PaymentStatusBadge(status: PaymentStatus.COMPLETED)

// Payment summary card
PaymentCard(payment: paymentObject)

// Register routes in your host app
final module = PaymentRouteModule();
ShellRoute(
  routes: [...ownRoutes, ...module.buildRoutes()],
);
```

## Routes

| Path | Screen |
|------|--------|
| `/payments` | Payment search list |
| `/payments/detail/:id` | Payment detail |
| `/payments/send` | Send a payment |
| `/payments/receive` | Receive a payment |
| `/payments/links` | Manage payment links |

## Embedding Widgets

```dart
// Route selector for payment channels
PaymentRouteSelector(onSelected: (route) => print(route))

// Account input field with validation
AccountField(controller: accountController)

// Payment link card with share action
PaymentLinkCard(link: linkObject)
```
