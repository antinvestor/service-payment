# Test credentials & config needed from you

To make **collection + Flutterwave v4 checkout + return to your app** work in the test environment, please provide the following. Copy this list and fill values (prefer a secure channel for secrets).

---

## 1. Flutterwave (v4) — required

From the [Flutterwave developer dashboard](https://developer.flutterwave.com/docs/getting-started) (sandbox / test mode):

| # | Item | Where to find | Example shape |
|---|------|---------------|---------------|
| 1 | **Client ID** | Settings → API Keys → **v4** (sandbox) | UUID-like string |
| 2 | **Client Secret** | Same page | long secret string |
| 3 | **Webhook secret hash** | Settings → Webhooks → secret hash | arbitrary strong string you choose |
| 4 | **Webhook URL you will register** | We need a public HTTPS URL for the integrator | `https://…/webhook/flutterwave` |

Confirm:

- [ ] Account is on **v4** API keys (not v3 public/secret/encryption keys)
- [ ] Sandbox/test mode is enabled for first integration
- [ ] Transfers enabled if you want to test **disbursements** (optional for collect-only)
- [ ] Egress IPs whitelisted for transfers (only if testing Send/payouts)

---

## 2. Public URLs — required

| # | Item | Purpose | Example |
|---|------|---------|---------|
| 5 | **Checkout public base URL** | Hosted pay page origin | `https://pay.dev.example.com` |
| 6 | **App return URL** | After pay, browser lands here → `ConfirmPayment` | `https://admin.dev.example.com/billing/payment/return` |
| 7 | **Flutterwave integration public base** | Webhooks + optional FW return | `https://integrations.dev.example.com` |
| 8 | **Payment core service URL** | Integrator → `StatusUpdate` | cluster DNS / `host:port` |
| 9 | **Billing / collection API URL** | App → Collect/Start/Confirm | `https://api.dev.example.com/billing` |
| 10 | **Profile service URL** | Checkout profile prefill | cluster DNS |

---

## 3. Queue / messaging — required for E2E

| # | Item | Purpose | Example |
|---|------|---------|---------|
| 11 | **Prompt queue URI** | Core `INITIATE_PROMPT` must publish here; Flutterwave must subscribe | `nats://user:pass@nats:4222/flutterwave.prompts.dequeue` or `mem://…` for local |
| 12 | **Payment (payout) queue URI** | Payment routes `tx` → Flutterwave | same bus, `…/flutterwave.payments.dequeue` |
| 13 | **NATS (or bus) credentials** | If not `mem://` | URL + user/pass or token |

**Critical:** `INITIATE_PROMPT_TOPIC_URI` on **payment core** and `QUEUE_FLUTTERWAVE_PROMPT_URI` on the **Flutterwave service** must be the **same queue**.

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
| 21 | **EMAIL contact** (or `properties.email`) | Flutterwave v4 requires email |
| 22 | **Display name** (`properties.name` or known) | Flutterwave customer name |

Without 20–21, checkout still works but may ask for phone or use a placeholder email.

---

## 7. Optional but recommended

| # | Item | Purpose |
|---|------|---------|
| 23 | **Catalog version ID + plan ID** | `StartSubscription` smoke test |
| 24 | **Issued invoice ID** | `CollectPayment` smoke test |
| 25 | **Ledger AR / Cash account IDs** | Cash posting on confirm (can leave empty) |
| 26 | **`CHECKOUT_PARTITION_METHODS`** JSON | Restrict methods per partition if needed |
| 27 | **DB access** for payment `routes` seed | `mode=tx`, `uri=<payment queue URI>` for Flutterwave disbursements |

---

## What we will set from the above (mapping)

```bash
# Flutterwave integrator
FLUTTERWAVE_CLIENT_ID=<1>
FLUTTERWAVE_CLIENT_SECRET=<2>
FLUTTERWAVE_WEBHOOK_SECRET=<3>
FLUTTERWAVE_ENVIRONMENT=sandbox
FLUTTERWAVE_PUBLIC_WEBHOOK_BASE=<7>
FLUTTERWAVE_DEFAULT_REDIRECT_URL=<7>/webhook/flutterwave/return
PAYMENT_SERVICE_URI=<8>
QUEUE_FLUTTERWAVE_PROMPT_URI=<11>
QUEUE_FLUTTERWAVE_PAYMENT_URI=<12>

# Payment core
INITIATE_PROMPT_TOPIC_NAME=flutterwave.prompts.dequeue   # or your name
INITIATE_PROMPT_TOPIC_URI=<11>

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

**Flutterwave dashboard webhook:** `POST <7>/webhook/flutterwave` with secret hash = (3).

---

## Minimum set to start *collect-only* testing

If you want the smallest possible handoff, send **only**:

1. Flutterwave **Client ID**  
2. Flutterwave **Client Secret**  
3. Flutterwave **webhook secret hash**  
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
3. Pay with Flutterwave (card or MoMo).  
4. Webhook marks prompt successful.  
5. Browser returns to app `…/payment/return?session=…`.  
6. `ConfirmPayment` shows invoice paid / subscription active.  
