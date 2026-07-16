# Flutterwave v4 integration

Provider service: `apps/integrations/flutterwave`  
Official docs: [Getting started](https://developer.flutterwave.com/docs/getting-started) ·
[Authentication](https://developer.flutterwave.com/docs/authentication) ·
[Orchestrator charges](https://developer.flutterwave.com/docs/payment-orchestrator-flow) ·
[Cards](https://developer.flutterwave.com/docs/card) ·
[Encryption](https://developer.flutterwave.com/docs/encryption) ·
[Mobile money](https://developer.flutterwave.com/docs/mobile-money) ·
[Transfers](https://developer.flutterwave.com/docs/making-a-transfer) ·
[Webhooks](https://developer.flutterwave.com/docs/webhooks)

This integration targets **API v4** (OAuth 2.0). Classic v3 secret keys remain
only as a **legacy fallback** for multipay redirects when encrypted card fields
are absent.

## Preferred product experience

```
Product (opportunities / billing)
  → CheckoutService.CreateCheckoutSession (payer prefill from profile)
  → Browser: https://pay.stawi.org/c/{session_ref}
       • Stripe Link style: show name/email/phone already on file
       • Prefer Card (embedded form)
       • Browser AES-256-GCM encrypts PAN (never hits our servers clear)
  → POST /c/{ref}/pay → InitiatePrompt(route=flutterwave, encrypted card extras)
  → Flutterwave v4 POST /orchestration/direct-charges
  → next_action:
       • requires_pin / requires_otp → stay on pay.stawi.org
       • redirect_url (3DS) → iframe when possible, else bank page
  → Webhook charge.completed + GET /charges/{id}
  → StatusUpdate SUCCESSFUL → checkout session completed → product return_url
```

**No Flutterwave multipay homepage** when encryption key + OAuth v4 are configured.
3DS bank challenges may still open a bank-controlled page (unavoidable).

## Architecture (collections)

```
Checkout / Billing collection
  → PaymentService.InitiatePrompt (route from method registry, default flutterwave)
  → QUEUE_FLUTTERWAVE_PROMPT_URI
  → Flutterwave integration
       ├─ action=authorize     → PUT /charges/{id} (PIN/OTP/AVS)
       ├─ payment_method_id    → POST /charges (recurring / saved card)
       ├─ encrypted card       → POST /orchestration/direct-charges (type=card)
       ├─ phone + corridor     → mobile_money
       └─ legacy FLWSECK only  → v3 Standard multipay (fallback)
  → StatusUpdate(IN_PROCESS|SUCCESSFUL|FAILED, portable extras)
  → Webhook charge.completed + GET /charges/{id} verify
```

Portable Extra keys live in `pkg/collection` so swapping providers only changes
the route + adapter, not the checkout UI.

## Auth (v4 OAuth)

```http
POST https://idp.flutterwave.com/realms/flutterwave/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

client_id=…&client_secret=…&grant_type=client_credentials
```

Tokens expire in **10 minutes**; the client refreshes ≥60s before expiry and
retries once on HTTP 401. Charge calls retry transient 5xx/429 with backoff.

| Environment | API base URL |
|-------------|--------------|
| sandbox | `https://developersandbox-api.flutterwave.com` |
| production | `https://f4bexperience.flutterwave.com` |

Every API call sends:

- `Authorization: Bearer {access_token}`
- `X-Trace-Id` (12–255 chars)
- `X-Idempotency-Key` (12–255 chars, unique per request)

## Card encryption (embedded)

1. Dashboard → API settings → **Encryption key** (base64 AES-256).
2. Checkout: `CHECKOUT_CARD_ENCRYPTION_KEY` or `FLUTTERWAVE_ENCRYPTION_KEY`.
3. Browser `GET /c/{ref}/crypto` → encrypts PAN/expiry/CVV with AES-GCM + 12-char nonce.
4. Prompt extras: `encrypted_card_number`, `encrypted_expiry_month`,
   `encrypted_expiry_year`, `encrypted_cvv`, `card_nonce`.

## Subscriptions

- **Billing service** owns subscription state (`StartSubscription` → checkout → `ConfirmPayment`).
- First payment stores `payment_method_id` + `customer_id` on profile checkout clues.
- Renewals: InitiatePrompt with `payment_method_id`, `customer_id`, `recurring=true`
  (no card re-entry, no browser hop).
- Flutterwave subscription lifecycle queue is optional correlation only — v4 has no
  v3-style payment-plans API as the system of record.

## Retries (win fully)

| Layer | Mechanism |
|-------|-----------|
| Provider HTTP | Exponential backoff on 408/429/5xx (3 attempts) |
| Checkout UI | MaxAttempts + cooldown; failed sessions stay payable |
| Settlement | Billing sweeper confirms open invoices; checkout SweepProcessing |
| Webhook + poll | charge.completed re-queries GET /charges/{id} before SUCCESSFUL |

## Quick start

```bash
export FLUTTERWAVE_CLIENT_ID=…
export FLUTTERWAVE_CLIENT_SECRET=…
export FLUTTERWAVE_ENCRYPTION_KEY=…   # same as CHECKOUT_CARD_ENCRYPTION_KEY
export FLUTTERWAVE_WEBHOOK_SECRET=…
export FLUTTERWAVE_ENVIRONMENT=sandbox
export FLUTTERWAVE_DEFAULT_COLLECTION_METHOD=card
export QUEUE_FLUTTERWAVE_PROMPT_URI=mem://flutterwave.prompts.dequeue
export INITIATE_PROMPT_TOPIC_URI=mem://flutterwave.prompts.dequeue
export INITIATE_PROMPT_TOPIC_NAME=flutterwave.prompts.dequeue

go run ./apps/integrations/flutterwave/cmd
```

See `apps/integrations/flutterwave/deploy/env.example`.

## Collections detail

| Signal | Payment method | User experience |
|--------|----------------|-----------------|
| Encrypted card extras | `card` (v4 orchestrator) | Embedded form on pay.* |
| Saved `payment_method_id` | POST /charges recurring | One-click / renewal |
| Phone + KE/UG/TZ/GH/… | `mobile_money` | Push PIN |
| Explicit `opay` / `ussd` | opay / ussd | Redirect / dial |
| No card + FLWSECK only | Standard multipay | Legacy external page |

## Webhooks (v4)

- Header: **`flutterwave-signature`** = Base64(HMAC-SHA256(secret_hash, raw_body))
- Types: `charge.completed`, `transfer.disburse`, `transfer.reversal`
- After `charge.completed`, re-query `GET /charges/{id}` before fulfilling.

## Switching providers

1. Change checkout method `route` (e.g. `stripe` / `flutterwave`).
2. Implement the same portable Extra contract in the new adapter
   (`pkg/collection` keys).
3. Keep checkout UI and product gateways unchanged.

## Production checklist

- [ ] v4 Client ID + Secret (live)
- [ ] Encryption key on checkout + flutterwave
- [ ] Webhook HTTPS + signature verify
- [ ] Prompt topic URI aligned with integrator queue
- [ ] `CHECKOUT_PUBLIC_BASE_URL=https://pay.stawi.org`
- [ ] Settlement sweeper on billing
- [ ] `FLUTTERWAVE_ENVIRONMENT=production`
