# antinvestor_ui_billing

Embeddable billing management UI for Antinvestor applications. Provides screens and widgets for catalogs, subscriptions, usage-based billing, invoicing, credits, and discounts.

## Installation

```yaml
dependencies:
  antinvestor_ui_billing: ^0.1.0
```

## Features

- **Catalog Management**: Product/service catalogs with pricing models
- **Subscriptions**: Lifecycle management with state tracking
- **Invoicing**: Invoice generation, listing, and detail views
- **Usage-Based Billing**: Usage event tracking and metered billing
- **Billing Runs**: Automated billing cycle execution
- **Credits & Discounts**: Credit management and discount configuration
- **Embeddable Widgets**: `SubscriptionStateBadge`, `InvoiceStateBadge`, `BillingRunStateBadge`, `InvoiceCard`, `SubscriptionCard`, `PricingModelBadge`, `InvoiceLineTile`, `UsageEventTile`
- **Routing**: `BillingRouteModule` with GoRouter integration

## Usage

```dart
import 'package:antinvestor_ui_billing/antinvestor_ui_billing.dart';

// Subscription state indicator
SubscriptionStateBadge(state: SubscriptionState.ACTIVE)

// Invoice summary card
InvoiceCard(invoice: invoiceObject)

// Register routes in your host app
final module = BillingRouteModule();
ShellRoute(
  routes: [...ownRoutes, ...module.buildRoutes()],
);
```

## Routes

| Path | Screen |
|------|--------|
| `/billing/catalogs` | Catalog list |
| `/billing/catalogs/:catalogId` | Catalog detail |
| `/billing/subscriptions` | Subscription list |
| `/billing/subscriptions/:subscriptionId` | Subscription detail |
| `/billing/invoices` | Invoice list |
| `/billing/invoices/:invoiceId` | Invoice detail |
| `/billing/usage` | Usage events |
| `/billing/runs` | Billing runs |
| `/billing/credits` | Credit management |
| `/billing/discounts` | Discount list |

## Embedding Widgets

```dart
// Pricing model indicator
PricingModelBadge(model: pricingModel)

// Invoice line item row
InvoiceLineTile(line: lineItem)

// Usage event row
UsageEventTile(event: usageEvent)
```
