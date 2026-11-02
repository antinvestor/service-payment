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

## Money safety

These rules hold for every write to Flutterwave and every webhook.

| Situation | What happens |
|-----------|--------------|
| A prompt or payout message is delivered twice | The charge reference (`prompt-<id>`, `renew-<id>`, `pay-<id>`) and the `X-Idempotency-Key` come from our id. The worker first looks the charge up by reference, and Flutterwave answers a repeated key with its original response. It is charged or paid once. |
| A timeout, 5xx or lost response on a charge or transfer | Never reported failed. The worker looks the charge up by reference and reports it, or returns the message to the queue to be redelivered with the same key. |
| Flutterwave refuses a request (4xx, or a failed envelope) | The prompt or payout is reported `FAILED` with `error_code=provider_rejected`. |
| A duplicate transfer reference or transfer recipient | Reported `IN_PROCESS` with `needs_review=true`. The transfer webhook settles it; it is never sent again and never reported failed. Flutterwave cannot list transfers by reference, so a payout still flagged after an hour needs an operator to check the dashboard. A duplicate is recognised only from a 409 or Flutterwave's own error message. |
| A charge succeeds for a different amount or currency | Reported `FAILED` with `error_code=amount_mismatch`. Status extras always carry the charged `amount`/`currency` for checkout's paid-amount check. |
| An amount that cannot be sent exactly | Refused, never rounded: more than 2 decimal places, or a fraction of UGX, RWF, XOF, XAF and similar. |
| Mobile money in a currency the number's corridor does not use | Refused; it is never converted (e.g. USD from a +254 number). |
| A PIN / OTP / AVS step | The charge is re-read first. A redelivered step on a settled charge only reports it. A wrong OTP leaves the charge `IN_PROCESS` with `authorization_error` so the payer can retry. A step naming another prompt's charge is dropped. |
| A webhook | Requires `FLUTTERWAVE_WEBHOOK_SECRET` (503 without it) and a valid `flutterwave-signature` (`verif-hash` only on v3 keys). The charge or transfer is re-read from the API; only that record's status, amount and meta are used, including tenancy. If Flutterwave is unreachable we answer 502 so it resends. |
| `transfer.reversal` | Logged as an error for manual reconciliation; a payout is never re-marked. |
| The return page | Static text only; it never calls Flutterwave or claims a payment. |

Payouts go only to the payee's verified mobile money number, which the
payment service resolves into `recipient.contact_id`. Caller extras never
choose the destination, so bank payouts are not reachable through `Send`.

## Limits

- **One Flutterwave account.** Collections and payouts run on the platform's
  credentials. The payment service never names a settings connection on a
  prompt, and webhooks are verified against the platform account; a charge on
  any other account is ignored.
- **v4 OAuth in production.** v3 `FLWSECK` keys work in sandbox only
  (production refuses to start with them). With v3 keys, payouts go to M-Pesa
  Kenya only, and transfer webhooks cannot be verified.
- **Abandoned charges.** Flutterwave sends webhooks for completed charges; a
  charge the payer abandons stays pending until checkout's session expires.
- **Reversals** need manual reconciliation.

## Production checklist

The service refuses to start in production unless:
- `FLUTTERWAVE_CLIENT_ID` and `FLUTTERWAVE_CLIENT_SECRET` are v4 OAuth credentials, and `FLUTTERWAVE_SECRET_KEY` is unset;
- `FLUTTERWAVE_WEBHOOK_SECRET` is set;
- `QUEUE_FLUTTERWAVE_PROMPT_URI` and `QUEUE_FLUTTERWAVE_PAYMENT_URI` are shared queues, not `mem://`;
- `FLUTTERWAVE_DEFAULT_REDIRECT_URL` is a public https URL.

Before going live:
- [ ] `FLUTTERWAVE_ENVIRONMENT=production`
- [ ] The same secret hash is set as `FLUTTERWAVE_WEBHOOK_SECRET` and in the dashboard; the webhook URL is `https://<integrations host>/webhook/flutterwave`
- [ ] `CHECKOUT_CARD_ENCRYPTION_KEY` equals the dashboard encryption key
- [ ] The `flutterwave` entry in `INITIATE_PROMPT_ROUTE_URIS` equals `QUEUE_FLUTTERWAVE_PROMPT_URI`
- [ ] The partition's `tx` route in `PAYMENT_ROUTES` uses `QUEUE_FLUTTERWAVE_PAYMENT_URI`
- [ ] `CHECKOUT_PUBLIC_BASE_URL=https://pay.stawi.org`, and the billing settlement sweeper is on
- [ ] The sandbox checklist below has passed

### Sandbox checklist (run before the first live payment)

1. A card charge with PIN, then OTP: a wrong OTP and then the right one; the checkout session completes.
2. A card charge needing 3DS redirect; the webhook settles it.
3. M-Pesa (KE), MTN and Airtel (UG) mobile money collections; check the `network` codes Flutterwave accepts.
4. Send the same prompt twice: one charge, and the second request returns `X-Idempotency-Cache-Hit: true`.
5. `GET /charges?reference=` returns the charge by reference.
6. Make a duplicate charge or transfer reference with a new idempotency key, and record the status code Flutterwave returns (409 is assumed).
7. A mobile money payout (UGX whole, KES with cents); the `transfer.disburse` webhook carries `meta.payment_id` on the re-read transfer.
8. A renewal with a saved `payment_method_id` (card-on-file must be enabled on the account).
9. A webhook with a bad signature gets 401; a valid one is applied once.
