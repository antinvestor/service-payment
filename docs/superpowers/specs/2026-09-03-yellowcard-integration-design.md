# Yellow Card integration — design

Date: 2026-09-03
Status: approved (design reviewed in chat)
Provider docs: https://docs.yellowcard.engineering/ (index: `/llms.txt`)

## 1. Goal

Add Yellow Card (https://yellowcard.io) as a payment provider so that:

1. A merchant can collect local fiat (mobile money or bank transfer) from a
   customer through the hosted checkout page by selecting a `yellowcard`
   checkout method, with no changes to product apps, exactly as `mpesa`,
   `flutterwave` or `stripe` are used today.
2. The platform can also disburse local fiat to a customer via Yellow Card
   ("Send"), routed through the standard payment `Route` mechanism.
3. Swapping a merchant from another provider to Yellow Card is a config
   change: the checkout method registry `route` and the prompt route URI.

Non-goals: stablecoin settlement, virtual accounts (Latin America), RFQ /
conversion, widget, deployment manifests in the external deployment repo.

## 2. Yellow Card API facts the design depends on

| Item | Value |
|------|-------|
| Base URL sandbox | `https://sandbox.api.yellowcard.io/business` |
| Base URL production | `https://api.yellowcard.io/business` |
| Request auth | `Authorization: YcHmacV1 {apiKey}:{signature}` + `X-YC-Timestamp` (ISO8601). Signature = base64(HMAC-SHA256(secret, timestamp + path + METHOD [+ base64(sha256(body)) for POST/PUT])). Path excludes host and includes the `/business` prefix. |
| Receive (collection) | `POST /receive` (quote, 10 min expiry) → `POST /receive/{id}/accept` or `forceAccept: true`. After acceptance the customer has 4 hours to pay. |
| Receive lookup | `GET /receive/{id}`, `GET /receive/sequence-id/{sequenceId}` |
| Receive cancel / refund | `POST /receive/{id}/cancel`, `POST /receive/{id}/refund` (Nigeria only) |
| Send (payout) | `POST /send` (`forceAccept` required field) → `POST /send/{id}/accept`; lookup `GET /send/sequence-id/{sequenceId}` (legacy `/payments/...`) |
| Catalog | `GET /channels?country=XX`, `GET /networks?country=XX`, `GET /rates?currency=XXX` |
| Amount | `localAmount` (integer, local currency) or `amount` (integer USD). We always send `localAmount`. Response carries `convertedAmount`, `rate`, fees. |
| Channel selection | `channelType: momo|bank` + `country` + `currency` lets Yellow Card pick an active channel; `channelId` is the explicit alternative. Momo receives require `source.networkId`. |
| KYC | `recipient` (receive) / `sender` (send) metadata: name, country, email always; phone, address, dob, idType, idNumber for full KYC (mandatory over 20 USD lifetime 200 USD, and always for NGN, ZAR, BWP; Nigeria adds BVN via `additionalIdType/Number`). `customerUID` enables reduced KYC. |
| Webhook | `POST` JSON `{id, sequenceId, status, apiKey, event, errorCode?, sessionId, executedAt}`. `X-YC-Signature` = base64(HMAC-SHA256(secret, rawBody)). Events `RECEIVE.*` / `SEND.*` (legacy `COLLECTION.*` / `PAYMENT.*`). |
| Statuses | created, pending_approval, process, processing, pending_liquidity, pending, complete, failed, expired, cancelled, pending_refund, refund_processing, refund_failed, refunded |
| Errors | `{code, message}`; e.g. `AuthenticationError` 401, `PaymentValidationError` 400, `PaymentExpired` 400, `PaymentInvalidState` 400, `CollectionNotFoundError`/`PaymentNotFound` 404. Transaction `errorCode`: EXPIRED, INVALID_RECIPIENT, VALIDATION_FAILED, INVALID_NETWORK, INVALID_CURRENCY, INSUFFICIENT_BALANCE, REFUSED, GATEWAY_TIMEOUT, PROVIDER_ERROR, POSSIBLE_DUPLICATE, NAME_MISMATCH, FRAUD_CHECK, OTHER_ERROR |
| Sandbox | momo `+{cc}1111111111` completes, `+{cc}0000000000` fails; bank `1111111111` / `0000000000` |
| Receive coverage | momo: BW, BJ, CM, CI, MW, RW, TG, UG, ZM. bank: BW, CG, GA, MW, NG, RW, ZA (EFT), TZ, UG, ZM. **No Kenya receive.** |
| Production | Static IP allowlist required for API calls; webhooks come from a static IP. |

## 3. Architecture

```
Checkout page (method key "yellowcard", route "yellowcard")
  → PaymentService.InitiatePrompt(route=yellowcard, extras)
  → apps/default publishes InitiatePromptRequest to publisher "prompt.yellowcard"
     (INITIATE_PROMPT_ROUTE_URIS {"yellowcard": QUEUE_YELLOWCARD_PROMPT_URI})
  → apps/integrations/yellowcard prompt worker
       ├─ resolve credentials (headers → settings connection → env)
       ├─ resolve country + currency from MSISDN / extras / defaults
       ├─ pick channel type (momo if active momo channel, else bank; extras override)
       ├─ momo: resolve networkId (extras → default → single active → first active)
       ├─ POST /receive {sequenceId = prompt id, localAmount, forceAccept: true, ...}
       └─ StatusUpdate IN_PROCESS with portable extras
            momo: payment_instruction "Approve the request on your phone…"
            bank: payment_instruction + bank_name / bank_account_number /
                  bank_account_name / payment_reference / payment_expires_at
                  (+ checkout_url when Yellow Card returns a payment link)
  → Yellow Card webhook POST /webhook/yellowcard/receives
       ├─ verify X-YC-Signature (HMAC over raw body)
       ├─ re-fetch GET /receive/sequence-id/{sequenceId} (authoritative)
       └─ PaymentService.StatusUpdate(SUCCESSFUL | FAILED | IN_PROCESS)
  → checkout session completed / failed → product return URL

Payment Route (mode send, URI = QUEUE_YELLOWCARD_PAYMENT_URI)
  → apps/integrations/yellowcard payment worker → POST /send (forceAccept)
  → webhook POST /webhook/yellowcard/sends → re-fetch → StatusUpdate
```

The integration owns no database. All state lives in the payment service
(prompts, payments, status rows) and in Yellow Card (receives, sends).

## 4. Components

### 4.1 `apps/integrations/yellowcard/config`

Header constants (queue headers and settings-JSON keys share names):

- `X-API_CONNECTION_CREDENTIALS` (shared platform key)
- `X-YELLOWCARD_API_KEY`, `X-YELLOWCARD_SECRET_KEY`, `X-YELLOWCARD_ENVIRONMENT`,
  `X-YELLOWCARD_COUNTRY`, `X-YELLOWCARD_CURRENCY`, `X-YELLOWCARD_NETWORK`,
  `X-YELLOWCARD_CHANNEL_TYPE`, `X-YELLOWCARD_CUSTOMER_TYPE`,
  `X-YELLOWCARD_BUSINESS_ID`, `X-YELLOWCARD_BUSINESS_NAME`,
  `X-YELLOWCARD_WEBHOOK_SECRET`

`YellowcardConfig` (embeds `config.ConfigurationDefault`):

| Field | Env | Default |
|-------|-----|---------|
| PaymentServiceURI / WorkloadAPITargetPath | `PAYMENT_SERVICE_URI`, `PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH` | `127.0.0.1:7006`, `/ns/payments/sa/service-payment` |
| SettingsServiceURI / WorkloadAPITargetPath | `SETTINGS_SERVICE_URI`, `SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH` | `127.0.0.1:7005`, `/ns/profile/sa/service-settings` |
| SettingsIntegrationName / ID | `SETTINGS_INTEGRATION_NAME`, `SETTINGS_INTEGRATION_ID` | `Yellowcard`, `Default` |
| APIKey, SecretKey | `YELLOWCARD_API_KEY`, `YELLOWCARD_SECRET_KEY` | |
| Environment | `YELLOWCARD_ENVIRONMENT` | `sandbox` |
| Country, Currency | `YELLOWCARD_COUNTRY`, `YELLOWCARD_CURRENCY` | (ISO alpha-2, ISO 4217) |
| Network | `YELLOWCARD_NETWORK` | default network id or name |
| ChannelType | `YELLOWCARD_CHANNEL_TYPE` | empty = auto |
| CustomerType | `YELLOWCARD_CUSTOMER_TYPE` | `retail` |
| BusinessID, BusinessName | `YELLOWCARD_BUSINESS_ID`, `YELLOWCARD_BUSINESS_NAME` | used when CustomerType = institution |
| WebhookSecret | `YELLOWCARD_WEBHOOK_SECRET` | empty = use SecretKey |
| DefaultRedirectURL | `YELLOWCARD_DEFAULT_REDIRECT_URL` | used for channels that need `redirectUrl` when no `redirect_url` extra |
| CatalogCacheSeconds | `YELLOWCARD_CATALOG_CACHE_SECONDS` | `300` |
| QueuePaymentName / URI | `QUEUE_YELLOWCARD_PAYMENT_NAME`, `QUEUE_YELLOWCARD_PAYMENT_URI` | `yellowcard.payments.dequeue`, `mem://yellowcard.payments.dequeue` |
| QueuePromptName / URI | `QUEUE_YELLOWCARD_PROMPT_NAME`, `QUEUE_YELLOWCARD_PROMPT_URI` | `yellowcard.prompts.dequeue`, `mem://yellowcard.prompts.dequeue` |

### 4.2 `service/client`

- `interfaces.go`: `YellowcardClient` with `SubmitReceive`, `AcceptReceive`,
  `DenyReceive`, `GetReceive`, `GetReceiveBySequenceID`, `CancelReceive`,
  `RefundReceive`, `SubmitSend`, `AcceptSend`, `DenySend`, `GetSend`,
  `GetSendBySequenceID`, `GetChannels`, `GetNetworks`, `GetRates`,
  `VerifyWebhookSignature`. Every call takes `creds *Credentials`.
- `models.go`: `Credentials{APIKey, SecretKey, Environment, BaseURL, Country,
  Currency, Network, ChannelType, CustomerType, BusinessID, BusinessName,
  WebhookSecret}` with `ResolveBaseURL()` and `ResolveWebhookSecret()`;
  request structs (`ReceiveRequest`, `SendRequest`, `Party` for KYC,
  `Source`, `Destination`); response structs (`Receive`, `Send`, `BankInfo`,
  `Channel`, `Network`, `Rate`); status and error-code constants; `APIError{
  HTTPStatus, Code, Message}` implementing `error` so callers can branch on
  `PaymentExpired`, `PaymentInvalidState`, not-found.
- `signer.go`: `signRequest(req, rawBody, apiKey, secret, now)` sets
  `X-YC-Timestamp` and `Authorization`; `webhookSignature(secret, body)`.
  Pure functions with test vectors.
- `client.go`: `doRequest` with 30 s client timeout, metrics via
  `integrationobs.NewMetrics("yellowcard")`, HMAC signing on every request,
  JSON decode, `APIError` on non-2xx. Idempotent GETs retry on 408/429/5xx
  and transport errors (3 attempts, 200 ms → 2 s backoff). Financial POSTs
  are not retried; queue redelivery reuses the same `sequenceId` and Yellow
  Card rejects the duplicate, which the worker treats by looking the receive
  up by sequence id instead of failing.
- `catalog.go`: per-client cache of channels and networks by country with TTL
  (`CatalogCacheSeconds`), with helpers `ActiveChannel(country, currency,
  channelType, rampType)` and `ResolveNetwork(country, hint, channelID)`.
  `Channel.RampType` distinguishes `deposit` (receive) from `withdraw` (send).

### 4.3 `service/credentials`

Same 3-level resolver as pawaPay: headers → settings connection → env
defaults. Missing API key or secret → `ErrMissingCredentials`.

### 4.4 `service/queue`

- `status.go`: `statusEmitter.emitStatus` (sets `STATE_INACTIVE` on
  `STATUS_FAILED`).
- `payload.go` (pure, unit-tested):
  - `moneyToLocalAmount(Money) int64` (round half up to whole units).
  - `resolveCountry(phone, extras, creds) (country, currency)` using a
    prefix table covering all Yellow Card receive and send countries
    (BW 267 BWP, BJ 229 XOF, CM 237 XAF, CI 225 XOF, MW 265 MWK, RW 250 RWF,
    TG 228 XOF, UG 256 UGX, ZM 260 ZMW, NG 234 NGN, ZA 27 ZAR, TZ 255 TZS,
    GA 241 XAF, CG 242 XAF, KE 254 KES, SN 221 XOF, ML 223 XOF, BF 226 XOF).
    Extras `country` / `currency` and the prompt currency win over the table.
  - `normalizeMSISDN(phone) string` → E.164 with leading `+`.
  - `buildParty(prompt extras, phone, country, creds) Party` mapping portable
    extras: `customer_name`/`display_name`, `customer_email`/`email`,
    `customer_address`, `customer_dob`, `customer_id_type`,
    `customer_id_number`, `customer_additional_id_type`,
    `customer_additional_id_number`; institution fields from creds.
  - `customerUID(headers, extras)`: `customer_id` extra → profile id →
    tenant id.
  - `momoInstruction(amount, currency, network)` and
    `bankInstruction(bankInfo, amount, currency, reference, expiresAt)` text.
  - `failureExtras(errorCode)` mapping Yellow Card error codes to
    `failure_code` / `failure_message`.
- `prompts.go` (collections):
  1. Unmarshal `InitiatePromptRequest`; failure → metric, return nil.
  2. Resolve credentials; failure → `STATUS_FAILED`.
  3. Phone from `recipient.contact_id` → `source.contact_id`.
  4. Country / currency / amount.
  5. Channel type: extras `channel_type` → `payment_method_type`
     (`bank_transfer` → bank, `mobile_money` → momo) → creds → auto
     (momo if an active momo deposit channel exists for the country, else
     bank).
  6. Momo: resolve network id; missing → `STATUS_FAILED` with
     `failure_code: INVALID_NETWORK`.
  7. Build `ReceiveRequest{SequenceID: promptID, LocalAmount, Country,
     Currency, ChannelType, Source{AccountType, AccountNumber(phone for
     momo), NetworkID}, Recipient: party, CustomerType, CustomerUID,
     ForceAccept: true, RedirectURL: redirect_url extra or default, Reason:
     extras reason or "other"}`. Submit.
  8. Duplicate / already exists (400 with duplicate message) → fetch by
     sequence id and continue with that record.
  9. Terminal failure in the response (`failed`, `expired`) →
     `STATUS_FAILED`; otherwise `STATUS_IN_PROCESS` with extras:
     `entity_type: prompt`, `provider: yellowcard`, `receive_id`,
     `sequence_id`, `status`, `channel_type`, `channel_id`, `network_id`,
     `country`, `currency`, `local_amount`, `converted_amount`, `rate`,
     `service_fee_local`, `partner_fee_local`, `expires_at`,
     `payment_instruction`, and for bank: `bank_name`,
     `bank_account_number`, `bank_account_name`, `payment_reference`,
     `payment_expires_at`, plus `checkout_url`/`auth_redirect_url` and
     `next_action: redirect_url` when `bankInfo` carries a payment link.
- `payments.go` (payouts): unmarshal `paymentv1.Payment`; recipient phone or
  `recipient_account` (bank account number + `bank_code` / `network` extra);
  channel type from account type; `SendRequest{SequenceID: paymentID,
  LocalAmount, Reason (extras `reason` or "other"), Sender: party from
  extras / institution creds, Destination{AccountNumber, AccountType,
  NetworkID, AccountName}, CustomerType, CustomerUID, ForceAccept: true}`;
  emit `STATUS_IN_PROCESS` / `STATUS_FAILED` with `entity_type: payment`.

### 4.5 `service/handlers`

`YellowcardWebhookServer.NewRouterV1()`:

- `POST /webhook/yellowcard/receives`
- `POST /webhook/yellowcard/sends`
- `POST /webhook/yellowcard` (all events; dispatch on `event` prefix)
- `GET /healthz`

Flow per request:

1. Read raw body (limit 1 MiB). Decode `{id, sequenceId, status, apiKey,
   event, errorCode, executedAt}`.
2. Resolve credentials: `connection` query param → settings; else default.
   Reject 503 when none.
3. Verify `X-YC-Signature` with `creds.ResolveWebhookSecret()`; mismatch →
   401 `verification_failed`. When the payload `apiKey` is present and
   differs from `creds.APIKey`, reject 401 as well.
4. Determine kind from `event` prefix (`RECEIVE.`/`COLLECTION.` → receive,
   `SEND.`/`PAYMENT.` → send) or route.
5. Re-fetch by sequence id (fallback: by id). 404 → `unknown_payment`; other
   error → 502.
6. Tenancy from `tenant_id` / `partition_id` query params on the configured
   webhook URL (Yellow Card carries no metadata); build claims context.
7. `StatusUpdate{Id: sequenceId, ExternalId: id, Status, State, Extras}` with
   extras `entity_type`, `provider`, `receive_id`/`send_id`, `status`,
   `event`, `error_code`, `failure_code`, `failure_message`, `country`,
   `currency`, `local_amount`, `converted_amount`, `rate`, `reference`,
   `executed_at`, `refund_status` for refund states.

Status map: `complete` → SUCCESSFUL; `failed`, `expired`, `cancelled` →
FAILED (STATE_INACTIVE); `refunded` → FAILED with `refund_status: refunded`
(collection is no longer held); every other status → IN_PROCESS.

Events are idempotent: each `StatusUpdate` writes a new status row and the
latest wins; a repeated `complete` is harmless.

### 4.6 Wiring

- `cmd/main.go` mirrors pawaPay: config, frame service, payment + settings
  clients, client, resolver, webhook server, prompt + payment workers,
  `events.NewPaymentStatusUpdate`.
- `Dockerfile` copied from pawaPay with paths and title changed.
- `Makefile` `APP_DIRS` += `apps/integrations/yellowcard`.
- `deploy/env.example`.

## 5. Checkout changes (drop-in method)

1. `pkg/collection/provider.go`: portable bank-transfer status extras
   `bank_name`, `bank_account_number`, `bank_account_name`,
   `payment_reference`, `payment_expires_at`, and `NextActionBankTransfer =
   "requires_bank_transfer"`. Portable KYC input extras `customer_address`,
   `customer_dob`, `customer_id_type`, `customer_id_number`,
   `customer_additional_id_type`, `customer_additional_id_number`,
   `customer_country`, `reason`, `channel_type`, `network`.
2. `apps/checkout/service/business/checkout.go::captureProviderExtras`
   stores the bank extras as `_bank_name`, `_bank_account_number`,
   `_bank_account_name`, `_payment_reference`, `_payment_expires_at`.
3. `render.go::PageData` gains `BankTransfer *BankTransferDetails`;
   `pageDataFor` fills it; `HandleStatus` returns a `bank_transfer` object
   when present (payload becomes `map[string]any`).
4. `confirm.html` renders a details card (bank, account number, account
   name, reference, expiry) when present; `checkout.js` updates or reveals
   that card from the poll payload so details that arrive after page load
   are shown without a reload.
5. `InferCountryFromPhone`: add BW, BJ, CM, CI, MW, TG, GA, CG, SN, ML and
   BF so locality preselection works in Yellow Card countries.
6. `CHECKOUT_METHODS` default gains
   `{"key":"yellowcard","name":"Mobile Money / Bank (Yellow Card)","route":"yellowcard","prefixes":["267","229","237","225","265","250","228","256","260","234","27","255","241","242"],"currencies":["BWP","XOF","XAF","MWK","RWF","UGX","ZMW","NGN","ZAR","TZS"],"countries":["BW","BJ","CM","CI","MW","RW","TG","UG","ZM","NG","ZA","TZ","GA","CG"]}`.
7. Flutter `knownPaymentMethods` gains `'yellowcard': 'Yellow Card'`.

Existing checkout behaviour for momo needs no change: the adapter emits
`payment_instruction` only (no `next_action`), so the confirm page shows the
banner and polls to a terminal state.

## 6. Error handling

| Situation | Behaviour |
|-----------|-----------|
| Credentials missing | `STATUS_FAILED`, `failure_code: CREDENTIALS` |
| Country unresolvable / unsupported | `STATUS_FAILED`, `failure_code: INVALID_COUNTRY` |
| No active channel | `STATUS_FAILED`, `failure_code: CHANNEL_UNAVAILABLE` |
| Momo network unresolvable | `STATUS_FAILED`, `failure_code: INVALID_NETWORK` |
| Yellow Card 4xx | `STATUS_FAILED`, `failure_code` = API `code`, `failure_message` = API `message` |
| Yellow Card 5xx / transport on submit | `STATUS_FAILED`, `failure_code: PROVIDER_ERROR` (checkout lets the payer retry) |
| Duplicate sequenceId | look up existing receive and report its status |
| Webhook bad signature | 401, metric `verification_failed`, no state change |
| Webhook unknown sequence | 404, metric `unknown_payment` |
| Webhook provider unreachable | 502 so Yellow Card retries |
| StatusUpdate failure | 500 so Yellow Card retries |

Queue handlers always return nil; failures are reported as status events.

## 7. Observability

`integrationobs.NewMetrics("yellowcard")` in client, workers and webhook
server. Provider call operations: `receive submit`, `receive lookup`,
`send submit`, `send lookup`, `channels`, `networks`, `rates`, ... Log fields
`type=yellowcard.prompt|payment|webhook.<kind>`, `prompt_id`, `receive_id`,
`sequence_id`. Secrets and full KYC payloads are never logged; response
bodies are logged at debug only.

## 8. Security

- HMAC request signing on every call; timestamp is generated per request.
- Webhooks verified with constant-time compare; payload `apiKey` must match
  the resolved credentials.
- Status is only ever taken from a re-fetched Yellow Card record.
- Tenant/partition come from the operator-configured webhook URL query
  string, never from the body.
- Production requires static egress IP allowlisting on Yellow Card's side
  (documented in the runbook).

## 9. Testing

- `client/signer_test.go`: signature vectors for GET and POST, webhook
  signature verify (valid, tampered, wrong secret).
- `client/client_test.go`: httptest server asserting headers, path,
  body, and JSON mapping for submit receive (momo, bank with bankInfo),
  lookup by sequence id, submit send, channels/networks/rates, `APIError`
  mapping, retry on 503 for GET only.
- `client/catalog_test.go`: cache TTL, channel selection, network
  resolution precedence.
- `queue/payload_test.go`: amount rounding, country/currency resolution,
  MSISDN normalisation, party building (tier 0 vs full KYC), instructions.
- `queue/prompts_test.go` and `payments_test.go`: fake client + fake events
  manager; momo happy path, bank happy path with bank extras, failure paths,
  duplicate handling.
- `handlers/webhooks_test.go`: signature accept/reject, apiKey mismatch,
  event prefix dispatch, re-fetch mapping to SUCCESSFUL / FAILED /
  IN_PROCESS, tenant injection from query, unknown sequence 404.
- Checkout: extend `methods_test.go` for the new prefixes/countries;
  `checkout_test.go` for bank extras capture; `web_test.go` for the
  `bank_transfer` status payload and confirm page rendering.
- `go build ./...`, `go vet`, `golangci-lint`, `make format` clean.

## 10. Documentation and operations

- `docs/yellowcard-integration.md` in the style of the Flutterwave doc:
  flow, auth, env, quick start, sandbox test numbers, webhook setup,
  production checklist (static IP, KYC, webhook version v2).
- `docs/collection-production.md`: Yellow Card env table and the two queue
  constraints (`QUEUE_YELLOWCARD_PROMPT_URI` in `INITIATE_PROMPT_ROUTE_URIS`,
  `QUEUE_YELLOWCARD_PAYMENT_URI` = payout `Route.URI`).
- Rollout: deploy the integration, set `INITIATE_PROMPT_ROUTE_URIS`, add
  the checkout method (or set the merchant's `route` to `yellowcard`),
  create Yellow Card webhooks for `RECEIVE.*` and `SEND.*` pointing at
  `/webhook/yellowcard/receives?tenant_id=…&partition_id=…` and
  `/webhook/yellowcard/sends?...`.
