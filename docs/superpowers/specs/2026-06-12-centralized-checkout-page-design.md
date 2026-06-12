# Centralized Checkout Page — Design

Date: 2026-06-12
Status: Approved (brainstorm with repo owner)

## Purpose

A centralized, hosted payments page — "Stripe Link"-simple — that any platform
app can redirect a user to. It prefills who is paying (name, language,
locality, currency, contacts) from identity established at session-creation
time, preselects the best payment method from locality and past behaviour,
lets the payer switch methods with one tap, executes the payment over the
existing provider rails, and routes the successful result back to the
originating service with full tenancy and order context.

## Decisions (agreed during brainstorm)

1. **Entry model**: checkout *sessions* (merchant-created, server-side) and
   shareable *links*, both from day one. Links are owned by checkout itself —
   the existing `PaymentLink` concept in service-payment is **not** reused.
2. **Payer auth**: no login wall. Sessions created by merchant backends carry
   the payer's identity (resolved from the user's JWT by the merchant);
   visitors arriving on bare links pay as guests with just a phone number.
3. **Methods**: render every method available for the session's tenant;
   preselect the best option from (a) profile clue, (b) locality/phone-prefix
   inference, (c) tenant default — switching is one tap.
4. **Hosting**: server-rendered Go (html/template + minimal vanilla JS) — no
   Flutter bundle. Fast first paint.
5. **Structure**: self-contained `apps/checkout` app in this repo. Owns its
   own tables (`checkout_sessions`, `checkout_links`); treats service-payment
   and service-profile purely as RPC clients, like the integrations do.
6. **Memory ("clues")**: when a linked contact fulfils a payment, capture the
   hints into the payer's **profile details payload** (profile service
   properties). Guests get a signed device cookie with the same hints.
7. **Capture**: the browser redirect to `return_url` is UX-only. Source of
   truth is the session (pollable via `GetCheckoutSession`) plus the payment
   service's existing status events, which carry `session_ref`/`order_ref`
   because checkout seeds them into the prompt's extras.

## Architecture

```
Browser ──GET pay.stawi.org/c/{ref}──────► apps/checkout  (HTML page, vanilla JS)
Merchant ──CreateCheckoutSession RPC─────► apps/checkout  (api.stawi.org/checkout)
apps/checkout ──InitiatePrompt/Status────► service-payment (existing rails)
apps/checkout ──read prefill/write clues─► service-profile
```

`apps/checkout` is a frame-based service mirroring the integrations layout:

```
apps/checkout/
├── cmd/main.go
├── config/config.go
├── migrations/0001/…            # checkout_sessions, checkout_links
├── Dockerfile
└── service/
    ├── models/                  # CheckoutSession, CheckoutLink (data.BaseModel)
    ├── repository/              # datastore.BaseRepository implementations
    ├── business/                # session lifecycle, preselection, clue write-back
    ├── handlers/
    │   ├── rpc.go               # CheckoutService connect handlers
    │   └── web.go               # GET /c/{ref}, POST /c/{ref}/pay,
    │                            # GET /c/{ref}/status, GET /l/{ref}
    └── web/                     # embed.FS: templates/, static/ (css, one js file)
```

## API surface

### Connect RPC — `proto/checkout/v1/checkout.proto`

- `CreateCheckoutSession(CreateCheckoutSessionRequest) → CheckoutSession`
  - amount, currency, `amount_option` (FIXED | VARIABLE), name, description,
    `order_ref`, metadata (map), `return_url`, optional payer block
    (profile_id, display name, contact ids, language), optional method
    restriction. Tenancy from the caller's claims.
  - Returns `session_ref` and the full page URL.
- `GetCheckoutSession(ref) → CheckoutSession` — status + payment ids +
  order/tenancy echo; merchants poll this as one source of truth.
- `CreateCheckoutLink(CreateCheckoutLinkRequest) → CheckoutLink` — reusable
  session template (same fields, plus expiry/active); returns
  `pay.stawi.org/l/{link_ref}`.

### Public HTTP (the page)

| Route | Action |
|---|---|
| `GET /c/{ref}` | Render the session page (state machine below) |
| `POST /c/{ref}/pay` | Validate + initiate payment (CSRF-protected form) |
| `GET /c/{ref}/status` | JSON status for browser polling |
| `GET /l/{ref}` | Stamp a fresh single-use session from the link, redirect to `/c/{ref}` |

## Data model

`checkout_sessions` (data.BaseModel ⇒ tenant/partition included):

| Field | Notes |
|---|---|
| `ref` | 32-char random token; the only identifier ever in a URL |
| `link_id` | set when spawned from a link |
| `name`, `description` | display |
| `amount`, `currency`, `amount_option` | money facts, server-side only |
| `order_ref`, `metadata` (JSONMap) | merchant correlation, echoed downstream |
| `return_url` | post-payment redirect |
| `payer_profile_id` | bound at creation when known; empty for guests |
| `prefill` (JSONMap) | snapshot: display name, language, contacts (masked on render) |
| `prompt_id`, `payment_id` | set when an attempt starts |
| `attempts` | prompt attempt counter (rate limiting) |
| `status` | PENDING → PROCESSING → COMPLETED / FAILED / EXPIRED; FAILED may retry → PENDING |
| `expires_at` | default 30 min |

`checkout_links`: `ref`, `name`, `description`, `amount`, `currency`,
`amount_option`, `order_ref`, `metadata`, `return_url`, `expires_at`,
`active`.

## Page UX (three states, one template set)

1. **Pay** — merchant name/description + amount (input only when VARIABLE).
   Payer block:
   - *Recognized*: “Paying as **{first name}** · {masked number}”, preselected
     method chip, single Pay button. “Change” expands editing.
   - *Guest*: one phone-number field.
   Method chips render from checkout's configured **method registry** (v1: a
   config-defined list of methods/providers with display names, logos and the
   prompt route hint each maps to), filtered by any per-session/link method
   restriction; one-tap switching. (A payment-service RPC for live
   route/availability discovery can replace the static registry later.)
2. **Confirm on phone** — provider-specific PIN-prompt instructions, live
   polling, “didn't get the prompt?” retry after timeout.
3. **Done** — success ✓ then auto-redirect to
   `return_url?session={ref}&status=completed` (~2s); failure shows a
   human-readable reason with retry back to state 1.

**Preselection order**: profile clue (`checkout` payload) → provider inferred
from phone prefix/locality → tenant default route. **Language**: `?lang` →
profile preference → `Accept-Language`. **Currency**: fixed by the session.

## Payment execution & capture

- `POST /pay` → business layer calls service-payment `InitiatePrompt` under
  the **session's tenancy** (claim-enrichment headers), with `session_ref`,
  `order_ref` and merchant metadata in the prompt `Extra`. Existing routing
  (Route table → provider queues → integrations) is unchanged.
- Browser polling: `/c/{ref}/status` → checkout fetches prompt status from
  service-payment, persists the session transition, returns JSON.
- Merchant capture: `GetCheckoutSession` polling + existing payment status
  events (now carrying `session_ref`/`order_ref` in extras). The redirect is
  never trusted as proof of payment.
- A background sweeper reconciles sessions stuck in PROCESSING (abandoned
  browsers): polls payment status until terminal or session expiry.

## Clues (quick-repeat memory)

On COMPLETED with a `payer_profile_id` whose linked contact made the payment,
write (async, best-effort, never blocking the redirect) into the profile
properties payload:

```json
"checkout": {
  "lastMethod": "mobile_money",
  "lastProvider": "MPESA_KEN",
  "lastContactId": "…",
  "lastCurrency": "KES",
  "lastPaidAt": "2026-06-12T…Z"
}
```

Session creation and rendering read this payload for prefill/preselection.
Guests: HMAC-signed httpOnly cookie `{phone, provider}` for device-local
recognition.

## Security

- Session refs are unguessable capability tokens; terminal sessions render no
  payer details.
- Rendered HTML contains only first name + masked number; full contact data
  never round-trips through the browser for recognized payers.
- Client input is limited to: phone (guests), method choice, amount (VARIABLE
  only, validated against link bounds), CSRF token. All other facts are
  server-side.
- CSRF token bound to session; max 3 prompt attempts per session with
  cooldown; per-IP rate limit on `/l/` session minting; 30-min expiry.
- Guest cookie HMAC-signed, httpOnly, Secure, SameSite=Lax — hints only.
- Checkout runs as service account `service-payment-checkout` (tenancy seed
  migration, same pattern as the integrations).

## Deployment

- `apps/checkout` added to Makefile `APP_DIRS`; Dockerfile per repo pattern →
  CI publishes `ghcr.io/antinvestor/service-payment-checkout`.
- Colony HelmRelease `namespaces/finance/checkout/` with DB migration enabled.
- Gateway: public page on hostname `pay.stawi.org` (+ `.dev`) via HTTPRoute
  (same-origin, no CORS); merchant RPCs path-routed at
  `api.stawi.{org,dev}/checkout` with the standard CORS header set;
  per-service DNS disabled; unified-DNS entry for `pay.stawi.org`.
- Tenancy client seed in service-authentication.

## Testing

- Handler tests (httptest + fake payment/profile clients): all render states,
  pay happy path, forged/expired refs → 404, CSRF rejection, rate limits,
  polling transitions, link → session spawning.
- Business/repo: lifecycle transitions incl. retry + sweeper, preselection
  ordering, clue payload write-back.
- Template smoke tests: every state × language renders; masked-number
  formatter leaks no digits.
- Gates: `go test -race`, `golangci-lint`, as repo-standard.

## Out of scope (v1)

- Card payments (Stripe) and bank rails — the method-chip model accommodates
  them later.
- Silent SSO on bare-link visits (guests stay guests in v1).
- Per-tenant page theming beyond merchant name/description.
- Outbound merchant webhooks (events + polling cover capture).
