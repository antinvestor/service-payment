# Fincra integration

`apps/integrations/fincra` connects the payment service to
[Fincra](https://docs.fincra.com/docs/getting-started) for collections
(checkout method `fincra`) and mobile money payouts.

| | Currencies (Fincra docs) |
|---|---|
| Hosted checkout | NGN (card, bank transfer, PayAttitude, OPay), GHS, KES, UGX, TZS, XAF, XOF (mobile money), ZMW, ZAR, USD (card) |
| Direct charge | mobile money: KES, GHS, UGX, ZMW, TZS, XOF, XAF; bank transfer: NGN, KES, GHS, UGX, ZAR |
| Payouts (this adapter) | mobile money wallets: KES, GHS, UGX, ZMW, TZS, XOF, XAF, RWF, EGP |

## Flow

```
Checkout / Billing
  → PaymentService.InitiatePrompt (route "fincra")
  → INITIATE_PROMPT_ROUTE_URIS["fincra"] == QUEUE_FINCRA_PROMPT_URI
  → prompt worker (reference = prompt id; a redelivery finds the existing record)
       checkout mode: POST /checkout/payments → checkout_url + next_action=redirect_url
       direct mode:   POST /checkout/charges
                        mobile_money → phone prompt (payment_instruction);
                                       OTP operators → next_action=requires_otp
                        bank_transfer → account details (bank_* extras)
                        card → hosted checkout (direct card needs PCI-DSS)
  → StatusUpdate(IN_PROCESS | FAILED)
  → webhook (HMAC-SHA512 "signature") → re-read from Fincra by our reference
  → StatusUpdate(SUCCESSFUL | FAILED)

PaymentService.Send (payee by profile + contact)
  → PAYMENT_ROUTES tx route (partition, route_type msisdn) uri = QUEUE_FINCRA_PAYMENT_URI
  → payment worker → POST /disbursements/payouts (customerReference = payment id)
  → webhook payout.* → GET /disbursements/payouts/customer-reference/{id} → StatusUpdate
```

A create call that fails without a definitive answer (timeout, 5xx, 408, 425,
429) may still have been accepted, so it is never reported FAILED on a guess:
the worker looks its reference up and reports what Fincra holds, or returns
the error so the queue redelivers. Only a 4xx refusal fails the payment.

The integration owns no database. Prompt and payment ids are Fincra
references, which Fincra refuses to reuse, so queue redelivery never charges or
pays twice: each worker looks its reference up before creating anything.

## Authentication

Every request carries `api-key: <secret key>`, `x-pub-key: <public key>` and
`x-business-id: <business id>` (the `_id` from `GET /profile/business/me`).
Live access requires completed KYC and **IP whitelisting** (at most two IPs;
non-whitelisted calls get `403 ACCESS_DENIED`).

| Environment | API | Hosted pages |
|---|---|---|
| sandbox | `https://sandboxapi.fincra.com` | `https://sandbox-checkout.fincra.com/pay/{payCode}` |
| production | `https://api.fincra.com` | `https://checkout.fincra.com/pay/{payCode}` |

## Amounts

- Collections: Fincra's `amount` is an integer, so a fractional amount fails
  the prompt (`INVALID_AMOUNT`) instead of being rounded, and checkout never
  offers the `fincra` route for a fractional amount.
- Payouts: up to two decimal places, whole units for UGX and Airtel (Fincra's
  rules).
- `FINCRA_FEE_BEARER=business` (default) makes the payer pay exactly the
  amount asked. With `customer`, Fincra adds its fee to what the payer pays.

## Payment verification

A webhook body is never trusted for the outcome. The handler verifies the
`signature` header (hex HMAC-SHA512 with the webhook encryption key, over the
body as received or as `JSON.stringify` would produce it), then reads the
record back from Fincra by our reference:

| Event | Lookup |
|---|---|
| `charge.*`, `collection.*` | `GET /checkout/payments/merchant-reference/{ref}`, else `GET /checkout/charges/merchant-reference/{ref}`; `ref` is `metadata.prompt_id`, `merchantReference` or `reference` (Fincra's own `fcr-…` codes are skipped) |
| `payout.*` | `GET /disbursements/payouts/customer-reference/{customerReference}` |

| Fincra status | Platform status |
|---|---|
| `success` / `successful` | SUCCESSFUL, with `amount` + `currency` for checkout's paid-amount check |
| `success` with `varianceType: underpayment` | FAILED (`AMOUNT_MISMATCH`); Fincra returns the money to the payer by default |
| `success` with `varianceType: overpayment` | SUCCESSFUL; Fincra settles the expected amount and returns the excess |
| `failed`, `expired` | FAILED |
| anything else | the webhook is answered `503` so Fincra redelivers it later |

Responses: `401` bad signature, `503` no webhook secret / still in flight,
`404` record unknown to Fincra, `502` Fincra unreachable, `500` payment
service error (Fincra retries), `200` recorded or not ours (conversions,
permanent virtual accounts).

Fincra sends checkout webhooks for successes only and never for expiry, so a
hosted checkout that is abandoned stays IN_PROCESS until checkout's own
session expiry.

## Tenancy and webhooks

Set one webhook URL per Fincra environment on the dashboard:

```
POST https://<integration>/webhook/fincra?tenant_id=…&partition_id=…[&connection=…]
```

- Hosted checkouts store `tenant_id` / `partition_id` / `prompt_id` in Fincra
  `metadata`; the verified record's metadata wins over the URL.
- Direct charges send a per-request `webhookUrl` built from
  `FINCRA_WEBHOOK_BASE_URL` with the prompt's tenancy and connection.
- Payouts use the dashboard URL's query parameters: Fincra's payout API takes
  neither a per-request `webhookUrl` nor `metadata`. A Fincra account's payout
  results are therefore recorded under the tenancy (and `connection`) on that
  account's dashboard URL. Give each tenant/partition that pays out its own
  Fincra account (or sub-account) and dashboard URL; sharing one account across
  partitions works for collections only.
- `connection` selects a settings-service credential set (keys
  `X-FINCRA_SECRET_KEY`, `X-FINCRA_PUBLIC_KEY`, `X-FINCRA_BUSINESS_ID`,
  `X-FINCRA_WEBHOOK_SECRET`, `X-FINCRA_ENVIRONMENT`, `X-FINCRA_COLLECTION_MODE`,
  `X-FINCRA_FEE_BEARER`, `X-FINCRA_COUNTRY`).

## Mobile money operators

Fincra's docs name operators inconsistently (`MPESA` and `SAFARICOM`), so the
adapter never hard-codes them:

- Payouts: `FINCRA_PAYOUT_PROVIDERS` (by network or currency), else the
  provider from `GET /core/banks?paymentDestination=mobile_money_wallet`
  matching the network of the payee's number, else the corridor's only
  provider. No match fails the payout (`UNKNOWN_OPERATOR`).
- Direct charges: the prompt's `network` extra, `FINCRA_CHARGE_OPERATORS`, or
  the network of the payer's number (`MPESA`, `AIRTEL`, `MTN`, `VODAFONE`,
  `AIRTELTIGO`, `TIGO`, `ZAMTEL`, `EQUITEL`). Verify the codes against the
  sandbox and pin them in `FINCRA_CHARGE_OPERATORS` before going live.

## Payout destinations

`PaymentService.Send` publishes only the payee's verified mobile number
(caller extras are not forwarded), so Fincra payouts go to mobile money
wallets. The beneficiary name comes from the payee's profile (`au_name`, …),
else `FINCRA_DEFAULT_BENEFICIARY_NAME`.

## Portable extras

Read from prompts: `customer_name`, `customer_email` (required by Fincra; else
`FINCRA_FALLBACK_CUSTOMER_EMAIL`), `success_url` / `redirect_url`,
`session_ref`, `payment_method_type` / `channel_type` (`mobile_money`,
`bank_transfer`, `card`), `network`, and for OTP continuation `action=authorize`,
`charge_id`, `otp`.

Emitted: `checkout_url` + `next_action=redirect_url` (hosted checkout);
`payment_instruction`; `charge_id` + `next_action=requires_otp`;
`bank_name`, `bank_account_number`, `bank_account_name`, `payment_expires_at`
or `payment_reference` (pool accounts); `failure_code`, `failure_message`.

> The OTP step relies on checkout's `Authorize`, which re-sends the prompt
> with the same id. Since #311 the payment service answers a repeated prompt
> id from its store without publishing it, so the OTP does not reach any
> provider until that is fixed in the payment service. Hosted checkout is
> unaffected.

## Quick start

```bash
cp apps/integrations/fincra/deploy/env.example .env   # fill in sandbox keys
# payment service:
export INITIATE_PROMPT_ROUTE_URIS='{"fincra":"mem://fincra.prompts.dequeue"}'
go run ./apps/integrations/fincra/cmd
```

Add the checkout method:

```json
{"key":"fincra","name":"Card / Bank / Mobile Money","route":"fincra",
 "currencies":["NGN","GHS","KES","UGX","TZS","ZMW","ZAR","XAF","XOF"],"redirect":true}
```

### Sandbox test data

- Mobile money collections: `+254700000000` (KES), `+233240000000` (GHS),
  `+256770000000` (UGX), `+260970000000` (ZMW), `+255780000000` (TZS).
- Payouts: see *Test Accounts & Mobile Wallets* in the Fincra docs (for
  example `254731234567` Airtel KES passes, `254712345678` M-Pesa fails).
- Fund the sandbox balance before testing payouts.

## Production checklist

- [ ] KYC complete; live secret, public and webhook keys issued
- [ ] Egress IP whitelisted (no `0.0.0.0` wildcard)
- [ ] `FINCRA_ENVIRONMENT=production`, `FINCRA_BUSINESS_ID` set
- [ ] Live dashboard webhook URL set with `tenant_id` / `partition_id`
- [ ] `FINCRA_WEBHOOK_BASE_URL` set if `FINCRA_COLLECTION_MODE=direct`
      (direct charge access granted by Fincra)
- [ ] `INITIATE_PROMPT_ROUTE_URIS` includes `fincra`
- [ ] `PAYMENT_ROUTES` declares the partition's `tx` route with `uri` = `QUEUE_FINCRA_PAYMENT_URI`
- [ ] Operator codes verified in sandbox; pinned in `FINCRA_CHARGE_OPERATORS` /
      `FINCRA_PAYOUT_PROVIDERS` where needed
- [ ] Wallet funded for payouts
