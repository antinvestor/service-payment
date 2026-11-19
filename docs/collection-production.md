# Production collection runbook

This document covers the streamlined payment collection path (Stripe Link
style) and what must be configured in each environment.

## End-to-end flow

```
Product / Admin UI  (or opportunities SPA)
  CollectPayment / StartSubscription / CreateCheckoutSession
        │  (session pre-created with profile prefill — ready when user commits)
        ▼
  Hosted checkout  https://pay.stawi.org/c/{session_ref}
  Stripe Link style:
    • Show name / email / phone already stored
    • Card first (Fincra's hosted payment page; whole amounts only)
    • MoMo chips only when locality matches
        │
        ▼
  Payment service route (card → fincra hosted payment page; MoMo → mpesa/mtn/airtel)
        │  card details are entered on Fincra's page, never on ours
        ▼
  ConfirmPayment(session_ref)   ← return page OR settlement sweeper
        │
        ▼
  Invoice PAID + optional subscription ACTIVE + ledger cash post
  Profile clues updated (last method + payment_method_id for renewals)
```

**Never trust the browser alone.** Capture is always `ConfirmPayment` / activator.

## Required configuration

### Checkout (`apps/checkout`)

| Variable | Purpose | Example |
|----------|---------|---------|
| `CHECKOUT_PUBLIC_BASE_URL` | Public origin for page URLs | `https://pay.stawi.org` |
| `CHECKOUT_SIGNING_SECRET` | CSRF + guest cookie HMAC | strong random |
| `CHECKOUT_ROUTES` | Payment routes offered (provider descriptors) | `fincra,mpesa,mtn,airtel` |
| `CHECKOUT_METHODS` | Optional per-key override of the derived catalog | empty |
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
| `BILLING_PROVIDER_COMMAND_URIS` | Provider command queues (cancel at Polar/Stripe) | `{"polar":"nats://…/polar.subscriptions.dequeue"}` |

### Yellow Card (`apps/integrations/yellowcard`) — dormant

Not built or offered yet; see the status note in the guide.

See [yellowcard-integration.md](./yellowcard-integration.md) for full setup.
Docs: [docs.yellowcard.engineering](https://docs.yellowcard.engineering/).

| Variable | Purpose |
|----------|---------|
| `YELLOWCARD_API_KEY` | Partner API key (Treasury Portal) |
| `YELLOWCARD_SECRET_KEY` | HMAC secret for `YcHmacV1` request signing and webhooks |
| `YELLOWCARD_WEBHOOK_SECRET` | Optional override for `X-YC-Signature` verification |
| `YELLOWCARD_ENVIRONMENT` | `sandbox` or `production` |
| `YELLOWCARD_COUNTRY` / `YELLOWCARD_CURRENCY` | Defaults when the MSISDN gives no corridor |
| `YELLOWCARD_CHANNEL_TYPE` | Force `momo` or `bank`; empty = auto |
| `YELLOWCARD_DEFAULT_REDIRECT_URL` | Return URL for payment-link channels (ZA, CI) |
| `QUEUE_YELLOWCARD_PROMPT_URI` | Must be the `yellowcard` entry in `INITIATE_PROMPT_ROUTE_URIS` |
| `QUEUE_YELLOWCARD_PAYMENT_URI` | Must match payment `Route.URI` for mode `tx` (sends) |

Checkout method key `yellowcard` (route `yellowcard`). Webhooks:
`POST /webhook/yellowcard/receives`, `POST /webhook/yellowcard/sends` with
`?tenant_id=…&partition_id=…`. Production needs Yellow Card IP allow-listing.

### Polar (`apps/integrations/polar`) — provider-managed subscriptions

See [provider-subscriptions.md](./provider-subscriptions.md) for the flow,
plan configuration, billing's `SyncProviderSubscription` RPC and the
`provider_subscription_sync` grant.

### Fincra (`apps/integrations/fincra`) — card provider

Checkout's default `card` method routes here (Fincra's hosted payment page).
Run the sandbox checklist in the guide before production traffic.

See [fincra-integration.md](./fincra-integration.md) for full setup.
Docs: [docs.fincra.com](https://docs.fincra.com/docs/getting-started).

| Variable | Purpose |
|----------|---------|
| `FINCRA_SECRET_KEY` / `FINCRA_PUBLIC_KEY` | `api-key` / `x-pub-key` headers |
| `FINCRA_BUSINESS_ID` | `x-business-id` header and payout `business` |
| `FINCRA_WEBHOOK_SECRET` | Webhook encryption key (`signature` HMAC-SHA512); required |
| `FINCRA_ENVIRONMENT` | `sandbox` or `production` |
| `FINCRA_COLLECTION_MODE` | `checkout` (hosted page, default) or `direct` (direct charge) |
| `FINCRA_FEE_BEARER` | `business` (default) or `customer` |
| `FINCRA_WEBHOOK_BASE_URL` | Public URL for direct-charge `webhookUrl` |
| `QUEUE_FINCRA_PROMPT_URI` | Must be the `fincra` entry in `INITIATE_PROMPT_ROUTE_URIS` |
| `QUEUE_FINCRA_PAYMENT_URI` | Must be the `uri` of the partition's `tx` route in `PAYMENT_ROUTES` |

Checkout method key `fincra` (route `fincra`, whole amounts only). Webhook:
`POST /webhook/fincra?tenant_id=…&partition_id=…`. Production needs Fincra
IP whitelisting.

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

## Integration workers: shared behaviour and new settings

Every worker now follows one contract (`pkg/events.StatusEmitter`,
`pkg/collection`): a FAILED status carries `STATE_INACTIVE` and an `error`
extra; a provider answer whose outcome is unknown (timeout, 5xx) is never
recorded as FAILED, the message is redelivered and the worker looks the
record up first; fractional amounts on whole-unit rails fail with a reason
instead of being rounded; webhook bodies are capped at 1 MiB; an
unauthenticated callback can only update a record it is bound to (stored
reference, amount and currency must match).

| Integration | New / changed setting | Purpose |
|-------------|-----------------------|---------|
| mpesa | `?payment_id=` on B2C result/timeout URLs, `?prompt_id=` on STK URL (set by the workers) | Callbacks bind to the payout / prompt we sent; callbacks without them are refused |
| mtn | `MTN_TARGET_ENVIRONMENT` | Market code sent as `X-Target-Environment` (`sandbox` outside production; required in production, e.g. `mtnuganda`) |
| airtel | `AIRTEL_DISBURSEMENT_PIN`, `AIRTEL_PUBLIC_KEY` | Disbursements need the RSA-encrypted PIN; payouts fail with a clear reason when unset |
| jenga | `JENGA_ENV` (`sandbox`/`production`), `JENGA_BASE_URL`, `JENGA_PRIVATE_KEY_PEM`, `JENGA_CALLBACK_URL` (service base) | The private key no longer ships in the image; callbacks carry a per-prompt secret; payouts are unsupported (tills-pay only) |
| stripe | `STRIPE_WEBHOOK_SECRET` (required), `BILLING_SERVICE_URI`, `QUEUE_STRIPE_SUBSCRIPTION_URI`, headers `X-STRIPE_*` | Hosted Checkout Sessions, subscription mirror into billing; settings connections keyed `X-API_*` must be re-saved |
| polar | `POLAR_WEBHOOK_SECRET` (required), `BILLING_SERVICE_URI`, `QUEUE_POLAR_SUBSCRIPTION_URI`, headers `X-POLAR_*` | See provider-subscriptions.md |

Known limits: Daraja's Transaction Status API is asynchronous, so a B2C
result is accepted on binding + amount match (the query is available for
reconciliation); Jenga exposes no transaction query, so an STK success is
recorded on callback secret + amount match only.

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

Checkout derives its catalog from the deployed payment routes:

- `CHECKOUT_ROUTES` (default `fincra,mpesa,mtn,airtel`) lists the routes this
  deployment offers, in display order. Each is a provider descriptor in
  `pkg/collection/providers.go`: name, currencies, countries, phone prefixes,
  flow (push / redirect), whole-amount rule, subscription support.
- `CHECKOUT_METHODS` optionally overrides or extends entries by key
  (unset fields inherit from the descriptor).
- `CHECKOUT_PARTITION_METHODS` restricts keys per partition; disjoint
  session/partition allowlists offer nothing.
- Every route in `CHECKOUT_ROUTES` needs a prompt queue in the payment
  service's `INITIATE_PROMPT_ROUTE_URIS`; the payment service refuses a
  prompt on a route it has no queue for (FailedPrecondition) instead of
  queueing it where nobody listens.

On the page: card is first and opens the provider's hosted page; MoMo rails
appear when the payer's currency and locality match; product rails (Polar)
appear only on sessions carrying `product_id` metadata. The submitted method
is validated with the same resolution the page used, so a crafted POST cannot
pick an excluded rail. Product apps can ask `CheckoutService.ListPaymentMethods`
which rails a session would be offered. Sessions take a `cancel_url`;
`paid_method` / `paid_route` on a completed session say which rail paid.

Silent saved-card renewals are off (`BILLING_RENEWAL_DEFAULT_ROUTE` empty):
each renewal sends the payer to the hosted page. Provider-managed plans
(Polar, Stripe) never renew through billing: see
[provider-subscriptions.md](./provider-subscriptions.md).

## Health checklist

- [ ] `CHECKOUT_PUBLIC_BASE_URL` is public HTTPS  
- [ ] Return URL hits ConfirmPayment path  
- [ ] Settlement sweeper enabled in production  
- [ ] Ledger cash + AR IDs set if financial posting required  
- [ ] Gateway routes for pay.* and collection/billing APIs  
- [ ] Keto grants for `payment_collect`  
- [ ] Provider integrations subscribed to prompt queues  
- [ ] `CHECKOUT_ROUTES` ⊆ keys of `INITIATE_PROMPT_ROUTE_URIS`  
- [ ] Integration service accounts granted `provider_subscription_sync` (Polar/Stripe)  
