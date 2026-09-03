# Yellow Card integration

Provider service: `apps/integrations/yellowcard`  
Official docs: [Overview](https://docs.yellowcard.engineering/) ·
[Authentication](https://docs.yellowcard.engineering/docs/authentication-api) ·
[Environments](https://docs.yellowcard.engineering/docs/environments-api) ·
[Making a Receive](https://docs.yellowcard.engineering/docs/making-a-collection) ·
[Making a Send](https://docs.yellowcard.engineering/docs/disbursement-user-journey) ·
[Channels](https://docs.yellowcard.engineering/docs/channels-api) ·
[Webhooks](https://docs.yellowcard.engineering/docs/webhooks-api) ·
[Events](https://docs.yellowcard.engineering/docs/events-api) ·
[Sandbox testing](https://docs.yellowcard.engineering/docs/sandbox-testing-api)

Yellow Card collects (**Receive**) and disburses (**Send**) local fiat through
mobile money and bank transfer in 20+ African countries. This integration
exposes it as the checkout method `yellowcard` and as a payout route.

## Product experience

```
Product (opportunities / billing)
  → CheckoutService.CreateCheckoutSession
  → Browser: https://pay.stawi.org/c/{session_ref}
       • payer picks "Mobile Money / Bank (Yellow Card)" (preselected for
         matching phone prefixes / countries)
  → POST /c/{ref}/pay → InitiatePrompt(route=yellowcard)
  → Yellow Card POST /receive (forceAccept)
       • momo  → USSD approval prompt on the payer's phone
                 confirm page shows "Approve the UGX 5,000 request…"
       • bank  → account details (or payment link) returned
                 confirm page shows bank, account number, reference, expiry
  → Webhook RECEIVE.COMPLETE → GET /receive/sequence-id/{id} verify
  → StatusUpdate SUCCESSFUL → checkout session completed → product return_url
```

## Architecture

```
Checkout / Billing collection
  → PaymentService.InitiatePrompt (route "yellowcard")
  → INITIATE_PROMPT_ROUTE_URIS["yellowcard"] == QUEUE_YELLOWCARD_PROMPT_URI
  → Yellow Card integration prompt worker
       ├─ credentials: headers → settings connection → env
       ├─ corridor: extras country/currency → MSISDN prefix → env default
       ├─ channel: extras channel_type / payment_method_type → env → auto
       │           (momo when an active momo deposit channel exists, else bank)
       ├─ momo network: extras network → env → first active on the channel
       └─ POST /receive {sequenceId = prompt id, localAmount, forceAccept}
  → StatusUpdate(IN_PROCESS | FAILED, portable extras)
  → Webhook (X-YC-Signature verified) + re-fetch → StatusUpdate

Payment Route (mode send, URI = QUEUE_YELLOWCARD_PAYMENT_URI)
  → payment worker → POST /send (forceAccept) → webhook SEND.* → StatusUpdate
```

The integration owns no database. Prompt and payment ids are used as the
Yellow Card `sequenceId`, so queue redelivery is idempotent: a duplicate
submission is answered by looking the existing record up.

## Authentication

Every request carries:

- `X-YC-Timestamp`: ISO8601 (`2006-01-02T15:04:05.000Z`)
- `Authorization: YcHmacV1 {apiKey}:{signature}`

`signature = base64(HMAC-SHA256(secretKey, timestamp + path + METHOD [+ base64(sha256(body))]))`
where `path` excludes the host and includes `/business`.

| Environment | API base URL |
|-------------|--------------|
| sandbox | `https://sandbox.api.yellowcard.io/business` |
| production | `https://api.yellowcard.io/business` |

Production additionally requires your egress IP to be allow-listed by Yellow
Card, and webhooks arrive from a static Yellow Card IP.

## Portable extras

Checkout already sends `customer_name`, `customer_email`, `success_url`,
`redirect_url`, `session_ref`, `customer_id`. The adapter also understands:

| Extra | Purpose |
|-------|---------|
| `channel_type` | `momo` or `bank`; `payment_method_type` `mobile_money` / `bank_transfer` also works |
| `network` | mobile money operator or bank (id, code or name) |
| `country`, `currency` | override the corridor derived from the MSISDN |
| `reason` | regulatory reason (`gift`, `bills`, … `other`) |
| `customer_address`, `customer_dob`, `customer_id_type`, `customer_id_number`, `customer_additional_id_type`, `customer_additional_id_number` | full KYC when the product holds it |
| `account_number`, `account_name`, `bank_code` | bank destinations for **sends** |

Status extras emitted back (consumed by checkout):

| Extra | When |
|-------|------|
| `payment_instruction` | always while in process |
| `bank_name`, `bank_account_number`, `bank_account_name`, `payment_reference`, `payment_expires_at` | bank receives |
| `checkout_url` / `auth_redirect_url` + `next_action=redirect_url` | channels that return a payment link (South Africa EFT, Ivory Coast) |
| `failure_code`, `failure_message` | failures (Yellow Card `errorCode` or API `code`) |
| `receive_id` / `send_id`, `rate`, `converted_amount`, `local_amount`, `expires_at` | always |

## KYC

Yellow Card applies reduced KYC (name, country, email + `customerUID`) for
retail transactions under 20 USD with a 200 USD lifetime limit per customer,
except in NGN, ZAR and BWP where full KYC is always required. Pass the
`customer_*` extras above for full KYC; Nigeria also needs BVN via
`customer_additional_id_type=BVN`. Set `YELLOWCARD_CUSTOMER_TYPE=institution`
with `YELLOWCARD_BUSINESS_ID` / `YELLOWCARD_BUSINESS_NAME` when the partner
itself is the KYC subject.

## Coverage (receives)

| Country | Mobile money | Bank transfer |
|---------|--------------|---------------|
| Botswana, Malawi, Rwanda, Uganda, Zambia | yes | yes |
| Benin, Cameroon, Ivory Coast, Togo | yes | – |
| Nigeria, South Africa (EFT), Tanzania, Gabon, Congo Brazzaville | – | yes |

Kenya has **no receive channel** (sends only). Keep `mpesa` for KES.

## Quick start

```bash
export YELLOWCARD_API_KEY=…
export YELLOWCARD_SECRET_KEY=…
export YELLOWCARD_ENVIRONMENT=sandbox
export YELLOWCARD_COUNTRY=UG
export YELLOWCARD_DEFAULT_REDIRECT_URL=https://pay.example.com
export QUEUE_YELLOWCARD_PROMPT_URI=mem://yellowcard.prompts.dequeue
# payment service:
export INITIATE_PROMPT_ROUTE_URIS='{"yellowcard":"mem://yellowcard.prompts.dequeue"}'

go run ./apps/integrations/yellowcard/cmd
```

See `apps/integrations/yellowcard/deploy/env.example`.

### Sandbox test numbers

| Outcome | Mobile money | Bank |
|---------|--------------|------|
| complete | `+{cc}1111111111` (e.g. `+2561111111111`) | `1111111111` |
| failed | `+{cc}0000000000` | `0000000000` |

## Webhooks

Create webhooks in the Treasury Portal (version **v2**) pointing at:

- `POST https://<integration>/webhook/yellowcard/receives?tenant_id=…&partition_id=…`
- `POST https://<integration>/webhook/yellowcard/sends?tenant_id=…&partition_id=…`
- or one catch-all `POST https://<integration>/webhook/yellowcard?…` (dispatches on the `event` prefix)

Add `&connection=<name>` to select a settings-service credential set for
multi-tenant deployments. The handler:

1. verifies `X-YC-Signature` (base64 HMAC-SHA256 of the raw body with the
   secret key, or `YELLOWCARD_WEBHOOK_SECRET`),
2. rejects payloads whose `apiKey` differs from the verifying credentials,
3. re-fetches the record by `sequenceId` and only then calls `StatusUpdate`.

Both v2 (`RECEIVE.*`, `SEND.*`) and legacy (`COLLECTION.*`, `PAYMENT.*`)
event names are accepted.

| Yellow Card status | Platform status |
|--------------------|-----------------|
| `complete` | SUCCESSFUL |
| `failed`, `expired`, `cancelled`, `refunded` | FAILED (state inactive) |
| everything else | IN_PROCESS (`refund_status` extra for refund states) |

## Switching providers

1. Set the checkout method `route` to `yellowcard` (or add the `yellowcard`
   method to `CHECKOUT_METHODS`).
2. Add `"yellowcard": "<prompt queue uri>"` to `INITIATE_PROMPT_ROUTE_URIS`.
3. Nothing else changes: checkout UI and product gateways are provider-agnostic.

## Production checklist

- [ ] Production API key + secret with **API write** permission
- [ ] Egress IP allow-listed with Yellow Card
- [ ] Webhooks created (v2) with HTTPS URLs and tenant query params
- [ ] `YELLOWCARD_ENVIRONMENT=production`
- [ ] `INITIATE_PROMPT_ROUTE_URIS` includes `yellowcard`
- [ ] Payout `Route.URI` equals `QUEUE_YELLOWCARD_PAYMENT_URI` (sends)
- [ ] USD float funded in the Treasury Portal for sends
- [ ] KYC extras supplied for NGN / ZAR / BWP and amounts over 20 USD
