# Test credentials & config needed from you

To make **collection + Fincra card checkout + return to your app** work in the test environment, please provide the following. Copy this list and fill values (prefer a secure channel for secrets).

---

## 1. Fincra — required

From the [Fincra dashboard](https://docs.fincra.com/docs/getting-started) (sandbox):

| # | Item | Where to find | Example shape |
|---|------|---------------|---------------|
| 1 | **Secret key** (`api-key`) and **public key** (`x-pub-key`) | Settings → API keys (sandbox) | long strings |
| 2 | **Business ID** (`x-business-id`) | Same page / account profile | id string |
| 3 | **Webhook encryption key** | Settings → Webhooks | the key Fincra signs with (HMAC-SHA512) |
| 4 | **Webhook URL you will register** | We need a public HTTPS URL for the integrator | `https://…/webhook/fincra` |

Confirm:

- [ ] Sandbox mode is enabled for first integration
- [ ] Hosted checkout is enabled for card (and bank transfer / mobile money if wanted)
- [ ] Payouts enabled if you want to test **disbursements** (optional for collect-only)
- [ ] Our egress IPs are whitelisted (Fincra requires this in production)

---

## 2. Public URLs — required

| # | Item | Purpose | Example |
|---|------|---------|---------|
| 5 | **Checkout public base URL** | Hosted pay page origin | `https://pay.dev.example.com` |
| 6 | **App return URL** | After pay, browser lands here → `ConfirmPayment` | `https://admin.dev.example.com/billing/payment/return` |
| 7 | **Fincra integration public base** | Webhooks | `https://integrations.dev.example.com` |
| 8 | **Payment core service URL** | Integrator → `StatusUpdate` | cluster DNS / `host:port` |
| 9 | **Billing / collection API URL** | App → Collect/Start/Confirm | `https://api.dev.example.com/billing` |
| 10 | **Profile service URL** | Checkout profile prefill | cluster DNS |

---

## 3. Queue / messaging — required for E2E

| # | Item | Purpose | Example |
|---|------|---------|---------|
| 11 | **Prompt queue URI** | Core must publish `fincra` prompts here; Fincra must subscribe | `nats://user:pass@nats:4222/fincra.prompts.dequeue` or `mem://…` for local |
| 12 | **Payment (payout) queue URI** | Payment routes `tx` → Fincra | same bus, `…/fincra.payments.dequeue` |
| 13 | **NATS (or bus) credentials** | If not `mem://` | URL + user/pass or token |

**Critical:** the `fincra` entry in `INITIATE_PROMPT_ROUTE_URIS` on **payment core** and `QUEUE_FINCRA_PROMPT_URI` on the **Fincra service** must be the **same queue**.

---

## 4. Checkout signing — required

| # | Item | Purpose |
|---|------|---------|
| 14 | **`CHECKOUT_SIGNING_SECRET`** | CSRF + guest cookie HMAC (any strong random ≥ 32 chars) |

---

## 5. Auth / tenancy — required for real API calls

| # | Item | Purpose |
|---|------|---------|
| 15 | **Test tenant ID** | Tenancy header / claims |
| 16 | **Test partition ID** | Routes + RLS + method allowlists |
| 17 | **Service account / OIDC config** (if services run securely) | Workload identity paths or client credentials for service-to-service |
| 18 | **Test user access token** (or how to mint one) | Call Collection + Checkout as a real app |

---

## 6. Profile data for “no re-type at checkout” — required for that UX

Pick **one** test profile and ensure:

| # | Item | Why |
|---|------|-----|
| 19 | **Profile ID** | Passed as payer on checkout / subscription |
| 20 | **MSISDN contact** on that profile (E.164, e.g. `2547…`) | MoMo + prefill phone |
| 21 | **EMAIL contact** (or `properties.email`) | Fincra requires the customer email |
| 22 | **Display name** (`properties.name` or known) | Fincra customer name |

Without 20–21, checkout still works but may ask for phone or use a placeholder email.

---

## 7. Optional but recommended

| # | Item | Purpose |
|---|------|---------|
| 23 | **Catalog version ID + plan ID** | `StartSubscription` smoke test |
| 24 | **Issued invoice ID** | `CollectPayment` smoke test |
| 25 | **Ledger AR / Cash account IDs** | Cash posting on confirm (can leave empty) |
| 26 | **`CHECKOUT_PARTITION_METHODS`** JSON | Restrict methods per partition if needed |
| 27 | **DB access** for payment `routes` seed | `mode=tx`, `uri=<payment queue URI>` for Fincra disbursements |

---

## What we will set from the above (mapping)

```bash
# Fincra integrator
FINCRA_SECRET_KEY=<1 secret>
FINCRA_PUBLIC_KEY=<1 public>
FINCRA_BUSINESS_ID=<2>
FINCRA_WEBHOOK_SECRET=<3>
FINCRA_ENVIRONMENT=sandbox
FINCRA_WEBHOOK_BASE_URL=<7>
FINCRA_DEFAULT_REDIRECT_URL=<6>
PAYMENT_SERVICE_URI=<8>
QUEUE_FINCRA_PROMPT_URI=<11>
QUEUE_FINCRA_PAYMENT_URI=<12>

# Payment core
INITIATE_PROMPT_ROUTE_URIS={"fincra":"<11>"}

# Checkout
CHECKOUT_PUBLIC_BASE_URL=<5>
CHECKOUT_SIGNING_SECRET=<14>
PROFILE_SERVICE_URI=<10>
PAYMENT_SERVICE_URI=<8>

# Billing
CHECKOUT_SERVICE_URI=<checkout>
CHECKOUT_INVOICE_RETURN_URL=<6>
BILLING_SETTLEMENT_SWEEP_INTERVAL_SECONDS=60

# App (Flutter)
--dart-define=BILLING_URL=<9>
--dart-define=CHECKOUT_RETURN_URL=<6>
```

**Fincra dashboard webhook:** `POST <7>/webhook/fincra` signed with the encryption key (3).

---

## Minimum set to start *collect-only* testing

If you want the smallest possible handoff, send **only**:

1. Fincra **secret + public keys**  
2. Fincra **business ID**  
3. Fincra **webhook encryption key**  
4. Public base for checkout (**5**)  
5. App return URL (**6**)  
6. Public base for integrator webhooks (**7**) — or confirm we may use a tunnel  
7. How services talk to each other (**mem://** local vs **NATS URL**)  
8. One **profile ID** with phone + email  
9. **Tenant + partition** IDs  

Everything else can be generated or defaulted for sandbox.

---

## What “done” looks like after credentials

1. Start subscription / collect invoice → browser opens checkout.  
2. Profile name/phone/method prefilled (no re-entry if contacts exist).  
3. Pay by card on Fincra's hosted page, or by MoMo.  
4. Webhook marks prompt successful.  
5. Browser returns to app `…/payment/return?session=…`.  
6. `ConfirmPayment` shows invoice paid / subscription active.  
