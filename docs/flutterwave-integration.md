# Flutterwave v4 integration

Provider service: `apps/integrations/flutterwave`  
Official docs: [Getting started](https://developer.flutterwave.com/docs/getting-started) ·
[Authentication](https://developer.flutterwave.com/docs/authentication) ·
[Orchestrator charges](https://developer.flutterwave.com/docs/payment-orchestrator-flow) ·
[Mobile money](https://developer.flutterwave.com/docs/mobile-money) ·
[Transfers](https://developer.flutterwave.com/docs/making-a-transfer) ·
[Webhooks](https://developer.flutterwave.com/docs/webhooks)

This integration targets **API v4 only** (OAuth 2.0, not v3 secret keys).

## Architecture

```
Checkout / Billing collection
  → PaymentService.InitiatePrompt (route=flutterwave)
  → QUEUE_FLUTTERWAVE_PROMPT_URI
  → Flutterwave integration
       ├─ phone present → POST /orchestration/direct-charges (mobile_money)
       └─ else          → POST /orchestration/direct-charges (bank_transfer | opay | ussd)
  → StatusUpdate(IN_PROCESS, extras.checkout_url | payment_instruction)
  → Webhook charge.completed + GET /charges/{id} verify
  → StatusUpdate(SUCCESSFUL|FAILED)

PaymentService.Send (released)
  → route mode=tx URI=flutterwave.payments.dequeue
  → POST /transfers/recipients → POST /transfers (action=instant)
  → Webhook transfer.disburse
  → StatusUpdate(SUCCESSFUL|FAILED)

Billing subscription lifecycle (optional)
  → subscription.lifecycle queue
  → correlation / logging (billing remains source of truth;
     v4 has no v3-style payment-plans API)
```

## Auth (v4 OAuth)

```http
POST https://idp.flutterwave.com/realms/flutterwave/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

client_id=…&client_secret=…&grant_type=client_credentials
```

Tokens expire in **10 minutes**; the client refreshes ≥60s before expiry and
retries once on HTTP 401.

| Environment | API base URL |
|-------------|--------------|
| sandbox | `https://developersandbox-api.flutterwave.com` |
| production | `https://f4bexperience.flutterwave.com` |

Every API call sends:

- `Authorization: Bearer {access_token}`
- `X-Trace-Id` (12–255 chars)
- `X-Idempotency-Key` (12–255 chars, unique per request)

## Quick start

### 1. Dashboard

1. Create / open a [Flutterwave developer account](https://developer.flutterwave.com/docs/getting-started).
2. Copy **v4** Client ID + Client Secret (sandbox or live).
3. **Settings → Webhooks** → URL `https://<host>/webhook/flutterwave`, set secret hash.
4. For disbursements: enable transfers + whitelist egress IPs.

### 2. Run

```bash
export FLUTTERWAVE_CLIENT_ID=…
export FLUTTERWAVE_CLIENT_SECRET=…
export FLUTTERWAVE_WEBHOOK_SECRET=…
export FLUTTERWAVE_ENVIRONMENT=sandbox
export PAYMENT_SERVICE_URI=payment-service:7006
export QUEUE_FLUTTERWAVE_PROMPT_URI=mem://flutterwave.prompts.dequeue
export QUEUE_FLUTTERWAVE_PAYMENT_URI=mem://flutterwave.payments.dequeue
# Align core prompt topic:
export INITIATE_PROMPT_TOPIC_URI=mem://flutterwave.prompts.dequeue
export INITIATE_PROMPT_TOPIC_NAME=flutterwave.prompts.dequeue

go run ./apps/integrations/flutterwave/cmd
```

See `apps/integrations/flutterwave/deploy/env.example`.

### 3. Wire routes

```sql
-- Disbursement route (URI must match QUEUE_FLUTTERWAVE_PAYMENT_URI)
INSERT INTO routes (…, mode, route_type, uri)
VALUES (…, 'tx', 'any', 'mem://flutterwave.payments.dequeue');
```

### 4. Checkout methods

Defaults use `route: "flutterwave"` for card / multi-currency pay.  
With a phone number, MoMo is preferred automatically.

## Collections detail

| Signal | Payment method | User experience |
|--------|----------------|-----------------|
| Phone + KE/UG/TZ/GH/… | `mobile_money` | Push PIN / optional redirect |
| No phone | `bank_transfer` (default) | Virtual account instructions |
| Extra `payment_method_type=opay` | `opay` | Redirect URL |
| Extra `payment_method_type=ussd` | `ussd` | USSD dial string |

Prompt extras of interest: `customer_email`, `customer_name`, `network`,
`payment_method_type`, `success_url`, `session_ref`, `invoice_id`, `subscription_id`.

Charge `meta` always includes `prompt_id` so webhooks map back to StatusUpdate.

## Webhooks (v4)

- Header: **`flutterwave-signature`** = Base64(HMAC-SHA256(secret_hash, raw_body))
- Body: `{ "webhook_id", "timestamp", "type", "data" }`
- Types handled: `charge.completed`, `transfer.disburse`, `transfer.reversal`
- Charge status values: `succeeded` | `pending` | `failed` | `voided`
- Transfer status values: `SUCCESSFUL` | `FAILED` | `PENDING` | …

After `charge.completed`, we re-query `GET /charges/{id}` before fulfilling (best practice).

## Subscriptions

Our **billing service** owns subscription state (`StartSubscription` → checkout → `ConfirmPayment`).  
Renewals are additional invoice collections through the same prompt path.  
The Flutterwave subscription queue worker is a lifecycle **observer / extension point** —
v4 does not mirror v3 payment-plan endpoints.

## Multi-tenant credentials

Priority: settings connection → message headers → process env.

```json
{
  "client_id": "…",
  "client_secret": "…",
  "webhook_secret": "…",
  "environment": "sandbox"
}
```

Headers: `X-API_CLIENT_ID`, `X-API_CLIENT_SECRET`, `X-API_WEBHOOK_SECRET`, `X-API_ENVIRONMENT`.

## Docker

```bash
docker build -f apps/integrations/flutterwave/Dockerfile -t ghcr.io/antinvestor/service-payment/flutterwave:dev .
```

## Production checklist

- [ ] v4 Client ID + Secret (live) in secret store  
- [ ] Webhook HTTPS + secret hash + signature verify  
- [ ] Prompt topic URI aligned with integrator queue  
- [ ] Route rows for `tx`  
- [ ] Transfer IP whitelist  
- [ ] Settlement sweeper on billing  
- [ ] `FLUTTERWAVE_ENVIRONMENT=production`  

## Extensibility

- New MoMo corridors: `service/queue/momo.go`  
- New collection methods: `defaultCollectionMethod` / prompt extras  
- Token cache is per `client_id:environment` — safe for multi-tenant  
