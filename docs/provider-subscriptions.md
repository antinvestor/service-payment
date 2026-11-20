# Provider-managed subscriptions (Polar, Stripe)

Billing can hand recurring collection to a provider that runs its own
subscription engine. The provider charges the payer on its schedule and
reports every change by webhook; billing mirrors that state, fans lifecycle
events out to product systems, and never raises its own renewal charge for
such subscriptions.

```
Product / Admin UI
  StartSubscription(plan)                 plan.data.paymentProvider = polar
        │                                 plan.data.providerProductId = prod_…
        ▼
  billing: PENDING subscription (collectionMode=provider, paymentProvider, providerProductId)
        │  CreateCheckoutSession(methods=[polar], metadata: product_id, subscriptionId)
        ▼
  hosted checkout  https://pay.…/c/{ref}   → InitiatePrompt(route=polar, extras: product_id …)
        │                                   → polar worker opens a Polar checkout (checkout_url)
        ▼
  payer pays on Polar's page, returns to /c/{ref}
        │  checkout.updated(succeeded) → prompt SUCCESSFUL → session COMPLETED
        │  subscription.active        → SyncProviderSubscription → billing ACTIVE
        ▼
  ConfirmPayment(session_ref)  (return page or settlement sweeper) → ACTIVE
        │
  order.paid (subscription_cycle) → SyncProviderSubscription(billed) → subscription.billed
  subscription.canceled/revoked   → SyncProviderSubscription          → cancelled / expired
  CancelSubscription(id)          → ProviderCommand on the polar command queue → Polar cancels
```

## Plan configuration

Set on `Plan.data`:

| Key | Value |
|-----|-------|
| `paymentProvider` | payment route of the provider: `polar` or `stripe` |
| `providerProductId` | the provider's product (Polar product id, Stripe price id) |
| `polarProductId` | legacy form, still read as `paymentProvider=polar` |

The plan must also carry a flat-fee component: the hosted page shows that
amount while the provider charges its product price.

## Billing

| Variable | Purpose |
|----------|---------|
| `BILLING_PROVIDER_COMMAND_URIS` | JSON `{"polar":"nats://…/polar.subscriptions.dequeue"}`: the queue each provider worker consumes commands (cancel) from |

RPC: `collection.v1.CollectionService/SyncProviderSubscription`, permission
`provider_subscription_sync` in namespace `service_billing`. Grant it to the
integration service accounts only (`granted_provider_subscription_sync`);
no role implies it. The caller's tenancy headers scope the lookup.

Lookup order for an event: billing `subscription_id` from the provider
metadata → provider subscription id → newest PENDING row for profile+plan.
A candidate must already be provider-managed by that provider, with the same
product and (once pinned) the same provider subscription id; anything else is
refused, so webhook metadata can never re-point another subscription.

Status mapping (normalized set in `pkg/collection`):

| Provider status | Billing |
|-----------------|---------|
| `active`, `trialing` | ACTIVE; `cancel_at_period_end` keeps ACTIVE with the cancel flag and `end_at` = period end |
| `past_due`, `paused` | ACTIVE, `providerStatus` recorded, `subscription.past_due` emitted once |
| `incomplete` | PENDING |
| `canceled` | ACTIVE with cancel flag while the paid period runs, else CANCELLED |
| `incomplete_expired` | CANCELLED |
| `revoked`, `unpaid` | EXPIRED |

Events older than the last applied one (`occurred_at`) are ignored. A billed
event (`order.paid`) activates a pending row, advances `currentPeriodEnd` and
emits `subscription.billed` with the provider order id as invoice id; the
same order is never recorded twice.

`ConfirmPayment` on a provider session only activates when the subscription
row is provider-managed, opened exactly that session, and the session was
paid on the provider's rail.

## Polar integration

| Variable | Purpose |
|----------|---------|
| `POLAR_API_KEY` | Organization access token |
| `POLAR_WEBHOOK_SECRET` | Webhook secret from the Polar dashboard (required; the raw value is the HMAC key, `whsec_…` also accepted) |
| `POLAR_ENVIRONMENT` | `sandbox` or `production` |
| `BILLING_SERVICE_URI` | Billing Connect target for `SyncProviderSubscription` (empty = subscription webhooks acknowledged, not mirrored) |
| `QUEUE_POLAR_PROMPT_URI` | Must be the `polar` entry in the payment service's `INITIATE_PROMPT_ROUTE_URIS` |
| `QUEUE_POLAR_SUBSCRIPTION_URI` | Must be the `polar` entry in billing's `BILLING_PROVIDER_COMMAND_URIS` |

Webhook: `POST /webhook/polar`. Subscribe to `checkout.created`,
`checkout.updated`, `subscription.*` and `order.*`. The checkout metadata
carries `prompt_id`, `session_ref`, `subscription_id`, `tenant_id` and
`partition_id`; Polar copies it onto the subscription and its orders.

Prompts need `product_id` (checkout forwards the session's `product_id`,
`productId` or `polarProductId` metadata). A redelivered prompt finds the
checkout it already opened; outcome-unknown answers from Polar are
redelivered rather than failed. Checkout `confirmed` is still processing;
only `succeeded` settles the prompt.

## Stripe integration

See `apps/integrations/stripe` (hosted Checkout Sessions, `customer.subscription.*`
and `invoice.*` webhooks mirrored the same way). `providerProductId` is the
Stripe Price id.

## Adding a provider

1. Add a `collection.Provider` descriptor in `pkg/collection/providers.go`
   with `Subscriptions: true` (and `RequiresProductID` when it charges a
   catalog product).
2. Prompt worker: open the provider's hosted session from the portable
   extras and emit `checkout_url` + `next_action=redirect_url`.
3. Webhooks: map the provider's statuses onto the normalized set and call
   `SyncProviderSubscription`; settle the prompt from the checkout event.
4. Command worker: consume `collection.ProviderCommand` and cancel at the
   provider.
5. Register the route in `CHECKOUT_ROUTES`, `INITIATE_PROMPT_ROUTE_URIS` and
   `BILLING_PROVIDER_COMMAND_URIS`; grant `provider_subscription_sync` to the
   worker's service account.
