# Production collection runbook

This document covers the streamlined payment collection path (Stripe Link /
Flutterwave style) and what must be configured in each environment.

## End-to-end flow

```
Product / Admin UI  (or opportunities SPA)
  CollectPayment / StartSubscription / CreateCheckoutSession
        │  (session pre-created with profile prefill — ready when user commits)
        ▼
  Hosted checkout  https://pay.stawi.org/c/{session_ref}
  Stripe Link style:
    • Show name / email / phone already stored
    • Prefer Card (embedded AES-GCM form)
    • Optional saved card one-click
    • MoMo chips only when locality matches
        │
        ▼
  Payment service route (default flutterwave) → v4 orchestrator / charge
        │  PIN/OTP on our page · 3DS bank page only if required
        ▼
  ConfirmPayment(session_ref)   ← return page OR settlement sweeper
        │
        ▼
  Invoice PAID + optional subscription ACTIVE + ledger cash post
  Profile clues updated (last method + payment_method_id for renewals)
```

**Never trust the browser alone.** Capture is always `ConfirmPayment` / activator.

**No provider multipay redirect** when card encryption + v4 OAuth are configured.

## Required configuration

### Checkout (`apps/checkout`)

| Variable | Purpose | Example |
|----------|---------|---------|
| `CHECKOUT_PUBLIC_BASE_URL` | Public origin for page URLs | `https://pay.stawi.org` |
| `CHECKOUT_SIGNING_SECRET` | CSRF + guest cookie HMAC | strong random |
| `CHECKOUT_METHODS` | Method registry JSON | default includes MoMo + card |
| `PAYMENT_SERVICE_URI` | Core payment service | cluster DNS |
| `PROFILE_SERVICE_URI` | Profile prefill/clues | cluster DNS |

### Billing (`apps/billing`)

| Variable | Purpose | Example |
|----------|---------|---------|
| `CHECKOUT_SERVICE_URI` | Checkout Connect target | cluster DNS |
| `CHECKOUT_INVOICE_RETURN_URL` | Browser landing after pay | `https://admin.stawi.org/billing/payment/return` |
| `BILLING_SETTLEMENT_SWEEP_INTERVAL_SECONDS` | Auto-confirm open sessions | `60` (0 disables) |
| `BILLING_SETTLEMENT_SWEEP_BATCH_SIZE` | Invoices per tick | `50` |
| `BILLING_LEDGER_AR_ACCOUNT_ID` | AR account for cash post | ledger account id |
| `BILLING_LEDGER_CASH_ACCOUNT_ID` | Cash account for capture | ledger account id |
| `BILLING_LEDGER_REVENUE_ACCOUNT_ID` | Revenue on invoice issue | ledger account id |
| `BILLING_SUBSCRIPTION_LIFECYCLE_TOPIC_NAME` | Default lifecycle queue ref | `subscription.lifecycle` |
| `BILLING_SUBSCRIPTION_LIFECYCLE_TOPIC_URI` | Queue URI product apps subscribe to | `nats://…` / `mem://…` |

### Flutterwave v4 (`apps/integrations/flutterwave`)

See [flutterwave-integration.md](./flutterwave-integration.md) for full setup.
Docs: [developer.flutterwave.com](https://developer.flutterwave.com/docs/getting-started).

| Variable | Purpose |
|----------|---------|
| `FLUTTERWAVE_CLIENT_ID` | v4 OAuth client id |
| `FLUTTERWAVE_CLIENT_SECRET` | v4 OAuth client secret |
| `FLUTTERWAVE_WEBHOOK_SECRET` | Secret hash for `flutterwave-signature` (HMAC-SHA256) |
| `FLUTTERWAVE_ENVIRONMENT` | `sandbox` or `production` |
| `QUEUE_FLUTTERWAVE_PROMPT_URI` | Must match `INITIATE_PROMPT_TOPIC_URI` (or router) |
| `QUEUE_FLUTTERWAVE_PAYMENT_URI` | Must match payment `Route.URI` for mode `tx` |

Checkout default card route is **`flutterwave`**. Webhook: `POST /webhook/flutterwave`.

Flutter UI uses the same return path via `--dart-define=CHECKOUT_RETURN_URL=…`
and `--dart-define=BILLING_URL=…`.

## Subscription external-entity integration

Subscriptions fan out lifecycle events the same way payment **Send/Receive**
route messages to provider integrations.

### Events

| Event | When |
|-------|------|
| `subscription.created` | Subscription row created (PENDING or ACTIVE) |
| `subscription.activated` | Free plan create, or PENDING→ACTIVE after pay |
| `subscription.cancelled` | Active or pending cancel |
| `subscription.billed` | Linked invoice marked PAID (signup or renew) |

### Binding a product entity

`StartSubscription` accepts:

- `external_entity_id` / `external_entity_type` — product resource (workspace, membership, …)
- `integration_route_id` — optional pin to one `IntegrationRoute` row
- `metadata` — free-form keys stored on `Subscription.data` and forwarded on events

### Delivery

1. **Partition `IntegrationRoute` rows** (`mode=lifecycle`, `route_type=any` or a specific event, `uri=queue URL`)
2. **Explicit `integration_route_id`** on the subscription
3. **Default topic** from `BILLING_SUBSCRIPTION_LIFECYCLE_TOPIC_*`

Payload JSON includes subscription id, profile, plan, state, external entity, invoice id (when billed), partition/tenant, and full `data`.

Product services subscribe to the queue (same pattern as M-Pesa/MTN payment workers) and grant or revoke entitlements. Delivery is best-effort and never rolls back billing state.

## Gateway exposure (Colony / Gateway API)

Follow the antinvestor service-exposure standard:

1. **Checkout public page** — hostname `pay.stawi.{org,dev}`  
   - Path `/` (or `/c`, `/l`, `/static`) to the checkout service  
   - Same-origin, no CORS required for the HTML page  
   - `externalDNS.enabled: false` if using unified DNS for `pay.*`

2. **Checkout merchant RPC** — path on `api.stawi.{org,dev}`  
   - Prefix `/checkout.v1.CheckoutService` (or `/checkout`)  
   - Standard admin CORS + Connect headers

3. **Billing + Collection RPC** — path on `api.stawi.{org,dev}`  
   - Prefix for `billing.v1.BillingService`  
   - Prefix for `collection.v1.CollectionService`  
   - CORS allowOrigins: admin hosts + localhost  
   - Required `allowHeaders` include Authorization, Connect-*, X-Tenant-Id, X-Partition-Id, X-Access-Id, Traceparent

Example rule sketch (values only — apply in deployments HelmRelease):

```yaml
gateway:
  enabled: true
  type: http
  hostnames:
    - api.stawi.org
    - api.stawi.dev
  httpRoute:
    rules:
      - matches:
          - path:
              type: PathPrefix
              value: /collection.v1.CollectionService
        backendRefs:
          - name: service-billing
            port: 80
      - matches:
          - path:
              type: PathPrefix
              value: /billing.v1.BillingService
        backendRefs:
          - name: service-billing
            port: 80
```

```yaml
# checkout public host
gateway:
  enabled: true
  type: http
  hostnames:
    - pay.stawi.org
    - pay.stawi.dev
  httpRoute:
    rules:
      - matches:
          - path:
              type: PathPrefix
              value: /
        backendRefs:
          - name: service-payment-checkout
            port: 80
```

## Authorization (Keto / OPL)

Namespace: `service_billing`

| Permission | Used by |
|------------|---------|
| `payment_collect` | CollectPayment, ConfirmPayment |
| `subscription_manage` | StartSubscription, CancelSubscription |
| `subscription_view` | List/Get subscription |
| `invoice_view` / `invoice_manage` | Invoice admin |

Grant `payment_collect` to product and admin service accounts that initiate collection.

## Settlement recovery

1. **Primary:** return URL → Flutter `PaymentReturnScreen` → ConfirmPayment  
2. **Secondary:** billing settlement sweeper every N seconds for ISSUED invoices with `data.checkoutSessionRef`  
3. **Manual:** admin “Confirm” snackbar action / ConfirmPayment RPC  

## Payment methods

- Hosted page shows methods from `CHECKOUT_METHODS`, filtered by currency, locality, partition, and session restriction.
- **Card is first and embedded** (`embed: true`, `redirect: false`, route `flutterwave`).
- MoMo methods require phone when selected; card only needs email (prefilled when known).
- Merchant may pass `methods: ["card"]` or `["mpesa","card"]` on CollectPayment / StartSubscription / product CreateSession.
- Required env for embedded card: `CHECKOUT_CARD_ENCRYPTION_KEY` (or `FLUTTERWAVE_ENCRYPTION_KEY`).

## Health checklist

- [ ] `CHECKOUT_PUBLIC_BASE_URL` is public HTTPS  
- [ ] Return URL hits ConfirmPayment path  
- [ ] Settlement sweeper enabled in production  
- [ ] Ledger cash + AR IDs set if financial posting required  
- [ ] Gateway routes for pay.* and collection/billing APIs  
- [ ] Keto grants for `payment_collect`  
- [ ] Provider integrations subscribed to prompt queues  
