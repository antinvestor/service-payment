# Yellow Card Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `apps/integrations/yellowcard` so checkout can collect mobile money and bank transfers through Yellow Card (and the payment service can disburse via Yellow Card Sends), as a drop-in checkout method.

**Architecture:** A frame-based integration service mirroring `apps/integrations/pawapay`: HMAC-signed HTTP client, credentials resolver, prompt (receive) and payment (send) queue workers emitting `StatusUpdate` events, and a webhook server that verifies `X-YC-Signature`, re-fetches the record from Yellow Card and calls `StatusUpdate`. Checkout gains portable bank-transfer status extras rendered on the confirm page.

**Tech Stack:** Go 1.26, `github.com/pitabwire/frame/v2`, Connect RPC (`buf.build/gen/go/antinvestor/...`), `testify`, `net/http/httptest`, `pkg/integrationobs`, `pkg/events`, `pkg/collection`.

**Spec:** `docs/superpowers/specs/2026-09-03-yellowcard-integration-design.md`

## Global Constraints

- Every `.go` file starts with the 13-line Apache-2.0 header copied from `apps/integrations/pawapay/config/config.go`.
- Module path `github.com/antinvestor/service-payments`; no new third-party dependencies.
- All logging via `util.Log(ctx)`; metrics via `integrationobs.NewMetrics("yellowcard")`.
- Provider HTTP client timeout 30 s at client level; no per-call deadlines.
- Queue handlers always return `nil`; failures are emitted as status events with `entity_type`.
- Every `StatusUpdate` extras map must contain `entity_type` (`prompt` or `payment`).
- Yellow Card base URLs: sandbox `https://sandbox.api.yellowcard.io/business`, production `https://api.yellowcard.io/business`.
- Run `make format` (gofmt + goimports) before every commit; pre-commit hook rejects unformatted code.
- Commit messages end with `Claude-Session: https://claude.ai/code/session_01VYBW6m9dsLKfhyrMEyDm4A`.

---

## File structure

```
apps/integrations/yellowcard/
├── Dockerfile
├── deploy/env.example
├── cmd/main.go                      bootstrap (frame service, clients, workers, webhook mux)
├── config/config.go                 env config + header constants
└── service/
    ├── client/
    │   ├── interfaces.go            YellowcardClient interface
    │   ├── models.go                Credentials, request/response DTOs, constants, APIError
    │   ├── signer.go                request signing + webhook signature
    │   ├── client.go                HTTP client, doRequest, endpoint wrappers
    │   ├── catalog.go               channels/networks cache + selection helpers
    │   ├── signer_test.go
    │   ├── client_test.go
    │   └── catalog_test.go
    ├── credentials/resolver.go      headers → settings → env
    ├── queue/
    │   ├── status.go                statusEmitter
    │   ├── payload.go               pure helpers (amount, country, party, instructions)
    │   ├── prompts.go               receive (collection) worker
    │   ├── payments.go              send (payout) worker
    │   ├── payload_test.go
    │   ├── prompts_test.go
    │   └── payments_test.go
    └── handlers/
        ├── webhooks.go              webhook server
        └── webhooks_test.go
pkg/collection/provider.go           new portable extras
apps/checkout/service/business/checkout.go   captureProviderExtras bank extras
apps/checkout/service/business/methods.go    InferCountryFromPhone prefixes
apps/checkout/service/handlers/render.go     PageData.BankTransfer
apps/checkout/service/handlers/web.go        pageDataFor + HandleStatus bank_transfer
apps/checkout/service/web/templates/confirm.html
apps/checkout/service/web/static/checkout.js
apps/checkout/config/config.go               CHECKOUT_METHODS default
ui/billing/lib/src/providers/collection_providers.dart
Makefile                                     APP_DIRS
docs/yellowcard-integration.md
docs/collection-production.md
```

---

### Task 1: Config, credentials model and signer

**Files:**
- Create: `apps/integrations/yellowcard/config/config.go`
- Create: `apps/integrations/yellowcard/service/client/models.go` (Credentials + constants only in this task)
- Create: `apps/integrations/yellowcard/service/client/signer.go`
- Test: `apps/integrations/yellowcard/service/client/signer_test.go`

**Interfaces:**
- Produces: `config.YellowcardConfig`, header constants `config.HeaderAPIKey` etc.
- Produces: `client.Credentials{APIKey, SecretKey, Environment, BaseURL, Country, Currency, Network, ChannelType, CustomerType, BusinessID, BusinessName, WebhookSecret}` with `ResolveBaseURL() string` and `ResolveWebhookSecret() string`.
- Produces: `client.SignRequest(req *http.Request, rawBody []byte, apiKey, secret string, now time.Time)`, `client.SignatureMessage(timestamp, path, method string, body []byte) string`, `client.WebhookSignature(secret string, body []byte) string`, `client.VerifyWebhookSignature(body []byte, header, secret string) bool`.

- [ ] **Step 1: Write config.go**

```go
package config

import "github.com/pitabwire/frame/v2/config"

const (
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"
	HeaderAPIKey                = "X-YELLOWCARD_API_KEY" //nolint:gosec // header name
	HeaderSecretKey             = "X-YELLOWCARD_SECRET_KEY" //nolint:gosec // header name
	HeaderEnvironment           = "X-YELLOWCARD_ENVIRONMENT"
	HeaderCountry               = "X-YELLOWCARD_COUNTRY"
	HeaderCurrency              = "X-YELLOWCARD_CURRENCY"
	HeaderNetwork               = "X-YELLOWCARD_NETWORK"
	HeaderChannelType           = "X-YELLOWCARD_CHANNEL_TYPE"
	HeaderCustomerType          = "X-YELLOWCARD_CUSTOMER_TYPE"
	HeaderBusinessID            = "X-YELLOWCARD_BUSINESS_ID"
	HeaderBusinessName          = "X-YELLOWCARD_BUSINESS_NAME"
	HeaderWebhookSecret         = "X-YELLOWCARD_WEBHOOK_SECRET" //nolint:gosec // header name
)

type YellowcardConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	SettingsIntegrationName string `envDefault:"Yellowcard" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default"    env:"SETTINGS_INTEGRATION_ID"`

	APIKey    string `env:"YELLOWCARD_API_KEY"`
	SecretKey string `env:"YELLOWCARD_SECRET_KEY"`
	Environment string `envDefault:"sandbox" env:"YELLOWCARD_ENVIRONMENT"`
	Country      string `env:"YELLOWCARD_COUNTRY"`
	Currency     string `env:"YELLOWCARD_CURRENCY"`
	Network      string `env:"YELLOWCARD_NETWORK"`
	ChannelType  string `env:"YELLOWCARD_CHANNEL_TYPE"`
	CustomerType string `envDefault:"retail" env:"YELLOWCARD_CUSTOMER_TYPE"`
	BusinessID   string `env:"YELLOWCARD_BUSINESS_ID"`
	BusinessName string `env:"YELLOWCARD_BUSINESS_NAME"`
	WebhookSecret string `env:"YELLOWCARD_WEBHOOK_SECRET"`
	DefaultRedirectURL string `env:"YELLOWCARD_DEFAULT_REDIRECT_URL"`
	CatalogCacheSeconds int `envDefault:"300" env:"YELLOWCARD_CATALOG_CACHE_SECONDS"`

	QueuePaymentName string `envDefault:"yellowcard.payments.dequeue"       env:"QUEUE_YELLOWCARD_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://yellowcard.payments.dequeue" env:"QUEUE_YELLOWCARD_PAYMENT_URI"`
	QueuePromptName  string `envDefault:"yellowcard.prompts.dequeue"        env:"QUEUE_YELLOWCARD_PROMPT_NAME"`
	QueuePromptURI   string `envDefault:"mem://yellowcard.prompts.dequeue"  env:"QUEUE_YELLOWCARD_PROMPT_URI"`
}
```

- [ ] **Step 2: Write the failing signer test**

```go
package client_test

func TestSignatureMessage(t *testing.T) {
	body := []byte(`{"a":1}`)
	sum := sha256.Sum256(body)
	want := "2022-01-11T15:48:37.424Z" + "/business/receive" + "POST" + base64.StdEncoding.EncodeToString(sum[:])
	got := client.SignatureMessage("2022-01-11T15:48:37.424Z", "/business/receive", http.MethodPost, body)
	assert.Equal(t, want, got)
	assert.Equal(t, "2022-01-11T15:48:37.424Z/business/channelsGET",
		client.SignatureMessage("2022-01-11T15:48:37.424Z", "/business/channels", http.MethodGet, nil))
}

func TestSignRequest(t *testing.T) {
	now := time.Date(2022, 1, 11, 15, 48, 37, 424_000_000, time.UTC)
	body := []byte(`{"a":1}`)
	req, _ := http.NewRequest(http.MethodPost, "https://sandbox.api.yellowcard.io/business/receive?x=1", bytes.NewReader(body))
	client.SignRequest(req, body, "key", "secret", now)
	assert.Equal(t, "2022-01-11T15:48:37.424Z", req.Header.Get("X-YC-Timestamp"))
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte(client.SignatureMessage("2022-01-11T15:48:37.424Z", "/business/receive", "POST", body)))
	assert.Equal(t, "YcHmacV1 key:"+base64.StdEncoding.EncodeToString(mac.Sum(nil)), req.Header.Get("Authorization"))
}

func TestWebhookSignature(t *testing.T) {
	body := []byte(`{"id":"1"}`)
	sig := client.WebhookSignature("secret", body)
	assert.True(t, client.VerifyWebhookSignature(body, sig, "secret"))
	assert.False(t, client.VerifyWebhookSignature([]byte(`{"id":"2"}`), sig, "secret"))
	assert.False(t, client.VerifyWebhookSignature(body, sig, "other"))
	assert.False(t, client.VerifyWebhookSignature(body, "", "secret"))
}
```

- [ ] **Step 3: Run `go test ./apps/integrations/yellowcard/service/client/` — expect compile failure.**

- [ ] **Step 4: Implement models.go (Credentials, constants) and signer.go**

```go
// signer.go
const (
	timestampLayout = "2006-01-02T15:04:05.000Z"
	authScheme      = "YcHmacV1"
	HeaderTimestamp = "X-YC-Timestamp"
	HeaderWebhookSignature = "X-YC-Signature"
)

func SignatureMessage(timestamp, path, method string, body []byte) string {
	msg := timestamp + path + strings.ToUpper(method)
	if len(body) > 0 && (method == http.MethodPost || method == http.MethodPut) {
		sum := sha256.Sum256(body)
		msg += base64.StdEncoding.EncodeToString(sum[:])
	}
	return msg
}

func SignRequest(req *http.Request, rawBody []byte, apiKey, secret string, now time.Time) {
	ts := now.UTC().Format(timestampLayout)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(SignatureMessage(ts, req.URL.Path, req.Method, rawBody)))
	req.Header.Set(HeaderTimestamp, ts)
	req.Header.Set("Authorization", authScheme+" "+apiKey+":"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}

func WebhookSignature(secret string, body []byte) string { /* base64 HMAC-SHA256 */ }
func VerifyWebhookSignature(body []byte, header, secret string) bool { /* hmac.Equal */ }
```

`Credentials.ResolveBaseURL`: `BaseURL` if set; `production` → `https://api.yellowcard.io/business`; else sandbox. `ResolveWebhookSecret`: `WebhookSecret` if set else `SecretKey`.

- [ ] **Step 5: Run tests — PASS. `make format`. Commit `feat(yellowcard): config, credentials and request signing`.**

---

### Task 2: Client models, HTTP client and catalog

**Files:**
- Modify: `apps/integrations/yellowcard/service/client/models.go`
- Create: `apps/integrations/yellowcard/service/client/interfaces.go`
- Create: `apps/integrations/yellowcard/service/client/client.go`
- Create: `apps/integrations/yellowcard/service/client/catalog.go`
- Test: `client_test.go`, `catalog_test.go`

**Interfaces (Produces):**

```go
const (
	StatusCreated = "created"; StatusPendingApproval = "pending_approval"; StatusProcess = "process"
	StatusProcessing = "processing"; StatusPendingLiquidity = "pending_liquidity"; StatusPending = "pending"
	StatusComplete = "complete"; StatusFailed = "failed"; StatusExpired = "expired"; StatusCancelled = "cancelled"
	StatusPendingRefund = "pending_refund"; StatusRefundProcessing = "refund_processing"
	StatusRefundFailed = "refund_failed"; StatusRefunded = "refunded"
	ChannelTypeMomo = "momo"; ChannelTypeBank = "bank"
	RampTypeDeposit = "deposit"; RampTypeWithdraw = "withdraw"
	CustomerTypeRetail = "retail"; CustomerTypeInstitution = "institution"
	ErrCodeNotFound... (CollectionNotFoundError, PaymentNotFound)
)

type APIError struct{ HTTPStatus int; Code, Message string }  // Error(); IsNotFound(); IsDuplicate()
func IsNotFound(err error) bool

type Party struct { Name, Country, Address, DOB, Email, IDNumber, IDType, AdditionalIDType, AdditionalIDNumber, Phone, BusinessID, BusinessName string } // json tags camelCase, omitempty
type Source struct { AccountType, AccountNumber, NetworkID string }
type Destination struct { AccountNumber, AccountType, NetworkID, AccountName, AccountBank, PhoneNumber, Country string }
type BankInfo struct { Name, AccountNumber, AccountName, PaymentLink string; Extra map[string]any } // custom UnmarshalJSON: known keys + any of "paymentLink","url","link","redirectUrl" → PaymentLink
type ReceiveRequest struct { ChannelID, ChannelType, SequenceID, Country, Currency string; LocalAmount int64; Source Source; Recipient Party; CustomerType, CustomerUID, RedirectURL, Reason string; ForceAccept bool }
type Receive struct { ID, SequenceID, Status, ChannelID, Country, Currency string; Amount, LocalAmount, ConvertedAmount, Rate, ServiceFeeAmountLocal, PartnerFeeAmountLocal float64; Source Source; Recipient Party; BankInfo *BankInfo; Reference, ErrorCode, SessionID, ExpiresAt, CreatedAt, UpdatedAt string }
type SendRequest struct { ChannelID, ChannelType, SequenceID, Country, Currency, Reason string; LocalAmount int64; Sender Party; Destination Destination; CustomerType, CustomerUID string; ForceAccept bool }
type Send struct { ID, SequenceID, Status, ChannelID, Country, Currency string; Amount, LocalAmount, ConvertedAmount, Rate float64; Destination Destination; Sender Party; Reference, ErrorCode, ExpiresAt, CreatedAt, UpdatedAt string }
type Channel struct { ID, Country, Currency, Status, APIStatus, ChannelType, RampType, SettlementType, VendorID string; Min, Max, FeeLocal, FeeUSD float64; EstimatedSettlementTime int }
type Network struct { ID, Code, Name, Country, Status, AccountNumberType string; ChannelIDs []string }
type Rate struct { Code, Locale, RateID string; Buy, Sell float64; UpdatedAt string }

type YellowcardClient interface {
	SubmitReceive(ctx, creds, *ReceiveRequest) (*Receive, error)
	AcceptReceive / DenyReceive / CancelReceive / RefundReceive(ctx, creds, id) (*Receive, error)
	GetReceive(ctx, creds, id) (*Receive, error)
	GetReceiveBySequenceID(ctx, creds, sequenceID) (*Receive, error)
	SubmitSend(ctx, creds, *SendRequest) (*Send, error)
	AcceptSend / DenySend(ctx, creds, id) (*Send, error)
	GetSend(ctx, creds, id) (*Send, error)
	GetSendBySequenceID(ctx, creds, sequenceID) (*Send, error)
	GetChannels(ctx, creds, country string) ([]Channel, error)
	GetNetworks(ctx, creds, country string) ([]Network, error)
	GetRates(ctx, creds, currency string) ([]Rate, error)
}

type Catalog struct{...}  // NewCatalog(cli YellowcardClient, ttl time.Duration) *Catalog
func (c *Catalog) Channels(ctx, creds, country) ([]Channel, error)   // cached per creds.APIKey+country
func (c *Catalog) Networks(ctx, creds, country) ([]Network, error)
func SelectChannel(channels []Channel, currency, channelType, rampType string) (*Channel, bool)  // status active && apiStatus != inactive; match currency (if given), channelType, rampType
func HasActiveChannel(channels []Channel, channelType, rampType string) bool
func ResolveNetwork(networks []Network, hint string, channel *Channel, accountType string) (*Network, bool)  // hint matches id, code or name (fold); else filter status active + accountNumberType==accountType (+ channel.ID in ChannelIDs when channel given); single → it; else first
```

Endpoint paths: `/receive`, `/receive/{id}/accept`, `/receive/{id}/deny`, `/receive/{id}/cancel`, `/receive/{id}/refund`, `/receive/{id}`, `/receive/sequence-id/{id}`, `/send`, `/send/{id}/accept`, `/send/{id}/deny`, `/send/{id}`, `/send/sequence-id/{id}`, `/channels?country=`, `/networks?country=`, `/rates?currency=`. `GetChannels` response is a bare JSON array; `GetRates` is `{"rates":[...]}`; `GetNetworks` a bare array.

`doRequest(ctx, creds, method, endpoint, payload, result, op)`: marshal, build URL (`ResolveBaseURL()+endpoint`), set `Content-Type` when body, `SignRequest(...)`, `httpClient.Do`; for GET with retryable failure (net error / 408 / 429 / 5xx) retry up to 3 attempts with 200 ms doubling to 2 s; on non-2xx decode `{code,message}` into `*APIError`; decode result.

- [ ] **Step 1: Write client_test.go** covering: `TestResolveBaseURL`; `TestSubmitReceive_Momo` (httptest asserts method POST, path `/business/receive`... note: `BaseURL` in tests is `srv.URL + "/business"` so the signed path includes `/business`; assert `Authorization` prefix `YcHmacV1 KEY:` and `X-YC-Timestamp` set; decode request body and assert `sequenceId`, `localAmount`, `forceAccept`, `channelType`, `source.accountType == "momo"`; respond with sample response; assert mapped fields incl. `ConvertedAmount`, `Rate`); `TestSubmitReceive_BankInfo` (response with `bankInfo` → `BankInfo.Name/AccountNumber/AccountName`); `TestSubmitReceive_APIError` (400 `{"code":"PaymentValidationError","message":"amount must be between 0 and 1000"}` → `*APIError` with Code); `TestGetReceiveBySequenceID_NotFound` (404 → `IsNotFound`); `TestGetChannels_RetriesOn503` (first 503 then 200 → success, 2 calls); `TestSubmitReceive_NoRetryOn503` (POST 503 → error, 1 call); `TestGetRates`; `TestGetNetworks`; `TestSubmitSend`.
- [ ] **Step 2: Run — compile failure.**
- [ ] **Step 3: Implement models.go DTOs, interfaces.go, client.go.**
- [ ] **Step 4: Run — PASS.**
- [ ] **Step 5: Write catalog_test.go**: `TestSelectChannel` table (active momo deposit for UGX chosen; inactive skipped; withdraw ramp ignored for deposit; currency mismatch skipped); `TestResolveNetwork` table (hint by name fold, hint by code, hint by id, single active momo, multiple → first active, none → false); `TestCatalogCaches` (fake client counts calls; two `Channels` calls within TTL → 1 provider call; after TTL 0 → 2).
- [ ] **Step 6: Implement catalog.go; run — PASS. `make format`. Commit `feat(yellowcard): API client and catalog`.**

---

### Task 3: Credentials resolver and status emitter

**Files:**
- Create: `apps/integrations/yellowcard/service/credentials/resolver.go`
- Create: `apps/integrations/yellowcard/service/queue/status.go`
- Test: `apps/integrations/yellowcard/service/credentials/resolver_test.go`

**Interfaces (Produces):**
- `credentials.NewResolver(settingsCli settingsv1connect.SettingsServiceClient, cfg *config.YellowcardConfig) *Resolver`
- `(*Resolver).FromHeaders(ctx, headers map[string]string) (*client.Credentials, error)`
- `(*Resolver).FromConnection(ctx, connection string) (*client.Credentials, error)`
- `(*Resolver).Default() (*client.Credentials, error)`
- `credentials.ErrMissingCredentials`
- `queue.statusEmitter{eventsMan frameEvents.Manager}` with `emitStatus(ctx, id, externalID string, status commonv1.STATUS, extras map[string]any)` setting `STATE_INACTIVE` for `STATUS_FAILED` (copy of flutterwave's `status.go:31-57`).

- [ ] **Step 1: Test** `TestFromHeaders_Defaults` (cfg key/secret used), `TestFromHeaders_Override` (header wins), `TestFromHeaders_Missing` (no key → `ErrMissingCredentials`), `TestFromConnection_NoSettings` (nil settings client → error), `TestDefault`.
- [ ] **Step 2: Implement** by copying pawaPay resolver, mapping all `config.Header*` constants to `client.Credentials` fields. Missing `APIKey` or `SecretKey` → `ErrMissingCredentials`.
- [ ] **Step 3: Run — PASS. `make format`. Commit `feat(yellowcard): credentials resolver and status emitter`.**

---

### Task 4: Portable collection extras and queue payload helpers

**Files:**
- Modify: `pkg/collection/provider.go`
- Create: `apps/integrations/yellowcard/service/queue/payload.go`
- Test: `apps/integrations/yellowcard/service/queue/payload_test.go`

**Interfaces (Produces):**

```go
// pkg/collection
ExtraCustomerCountry = "customer_country"; ExtraCustomerAddress = "customer_address"; ExtraCustomerDOB = "customer_dob"
ExtraCustomerIDType = "customer_id_type"; ExtraCustomerIDNumber = "customer_id_number"
ExtraCustomerAdditionalIDType = "customer_additional_id_type"; ExtraCustomerAdditionalIDNumber = "customer_additional_id_number"
ExtraReason = "reason"; ExtraChannelType = "channel_type"; ExtraNetwork = "network"; ExtraCountry = "country"; ExtraCurrency = "currency"
ExtraBankName = "bank_name"; ExtraBankAccountNumber = "bank_account_number"; ExtraBankAccountName = "bank_account_name"
ExtraPaymentReference = "payment_reference"; ExtraPaymentExpiresAt = "payment_expires_at"
NextActionBankTransfer = "requires_bank_transfer"
PaymentMethodTypeMobileMoney = "mobile_money"; PaymentMethodTypeBankTransfer = "bank_transfer"

// queue/payload.go
type corridor struct{ Country, Currency string }
var phoneCorridors = []struct{ prefix string; c corridor }{ {"267",{"BW","BWP"}}, {"229",{"BJ","XOF"}}, {"237",{"CM","XAF"}}, {"225",{"CI","XOF"}}, {"265",{"MW","MWK"}}, {"250",{"RW","RWF"}}, {"228",{"TG","XOF"}}, {"256",{"UG","UGX"}}, {"260",{"ZM","ZMW"}}, {"234",{"NG","NGN"}}, {"255",{"TZ","TZS"}}, {"241",{"GA","XAF"}}, {"242",{"CG","XAF"}}, {"254",{"KE","KES"}}, {"221",{"SN","XOF"}}, {"223",{"ML","XOF"}}, {"226",{"BF","XOF"}}, {"27",{"ZA","ZAR"}} }  // longest prefix first when matching
func normalizeMSISDN(raw string) string                     // digits only → "+" + digits; "" when < 7 digits
func corridorForPhone(msisdn string) (corridor, bool)
func resolveCorridor(phone string, extra *structpb.Struct, promptCurrency string, creds *client.Credentials) (corridor, error)  // extra country/currency > phone table > creds; currency: extra > promptCurrency > table > creds; error when country empty
func moneyToLocalAmount(m interface{GetUnits() int64; GetNanos() int32}) int64  // round half up
func buildParty(extra *structpb.Struct, phone, country string, creds *client.Credentials) client.Party  // institution when creds.CustomerType==institution → BusinessID/Name; else name/email/country/phone + optional address/dob/id fields
func customerUID(headers map[string]string, extra *structpb.Struct, profileID string) string  // customer_id → profileID → headers["tenant_id"]
func resolveChannelType(extra *structpb.Struct, creds *client.Credentials) string  // channel_type → payment_method_type mapping → creds.ChannelType → ""
func momoInstruction(amount int64, currency, network string) string   // "Approve the UGX 5,000 payment request sent to your phone (MTN). Enter your mobile money PIN to complete."
func bankInstruction(info *client.BankInfo, amount int64, currency, reference, expiresAt string) string
func formatLocalAmount(amount int64, currency string) string          // "UGX 5,000"
func failureExtras(code, message string) map[string]any             // {"failure_code","failure_message"}; message defaulted from a code→text table
func extraString(extra *structpb.Struct, keys ...string) string      // first non-empty
func receiveExtras(r *client.Receive, channelType, networkName string) map[string]any  // shared IN_PROCESS extras incl. bank_* and checkout_url when PaymentLink
func isTerminalFailure(status string) bool  // failed, expired, cancelled
```

- [ ] **Step 1: Tests** (table-driven): `TestNormalizeMSISDN`, `TestCorridorForPhone` (+256..., 27..., unknown), `TestResolveCorridor` (extras win; creds fallback; error), `TestMoneyToLocalAmount` (10.49→10, 10.5→11, nanos overflow), `TestBuildParty_Retail_Tier0`, `TestBuildParty_Retail_Full`, `TestBuildParty_Institution`, `TestCustomerUID`, `TestResolveChannelType`, `TestInstructions`, `TestReceiveExtras_Bank` (bank_* + checkout_url + next_action), `TestReceiveExtras_Momo` (payment_instruction, no next_action).
- [ ] **Step 2: Implement; run — PASS. `make format`. Commit `feat(yellowcard): payload helpers and portable bank extras`.**

---

### Task 5: Prompt (receive) worker

**Files:**
- Create: `apps/integrations/yellowcard/service/queue/prompts.go`
- Test: `apps/integrations/yellowcard/service/queue/prompts_test.go`

**Interfaces:**
- Produces: `queue.NewPromptHandler(eventsMan frameEvents.Manager, cli client.YellowcardClient, catalog *client.Catalog, settingsCli settingsv1connect.SettingsServiceClient, cfg *config.YellowcardConfig) queue.SubscribeWorker`
- Test doubles (in `prompts_test.go`, reused by `payments_test.go`): `fakeClient` embedding `client.YellowcardClient` implementing `SubmitReceive`, `GetReceiveBySequenceID`, `SubmitSend`, `GetSendBySequenceID`, `GetChannels`, `GetNetworks` with recorded requests; `fakeEvents` implementing `frameEvents.Manager` (embed the interface; implement `Emit` capturing `*commonv1.StatusUpdateRequest`). Check `frameEvents.Manager` methods in `github.com/pitabwire/frame/v2/events` before writing the fake.

Handle flow (spec §4.4). Duplicate handling: when `SubmitReceive` returns `*APIError` whose message contains "sequenceId" or "duplicate" (case-insensitive) or `Code == "PaymentValidationError"` with that message, call `GetReceiveBySequenceID` and continue.

- [ ] **Step 1: Tests**: `TestPrompt_MomoHappyPath` (UG phone, channels list has active momo deposit UGX channel, networks MTN+Airtel with hint `network: "MTN"`; assert request fields and emitted IN_PROCESS extras incl. `payment_instruction`, `network_id`, `entity_type=prompt`); `TestPrompt_BankHappyPath` (NG, only bank channel, response has bankInfo; assert bank_* extras); `TestPrompt_ProviderError` (APIError 400 → FAILED with failure_code=PaymentValidationError); `TestPrompt_MissingCredentials`; `TestPrompt_NoChannel` (CHANNEL_UNAVAILABLE); `TestPrompt_UnknownCountry` (INVALID_COUNTRY); `TestPrompt_DuplicateSequence` (submit returns duplicate error, lookup returns pending → IN_PROCESS); `TestPrompt_BadPayload` (no emit, returns nil).
- [ ] **Step 2: Implement; run — PASS. `make format`. Commit `feat(yellowcard): receive prompt worker`.**

---

### Task 6: Payment (send) worker

**Files:**
- Create: `apps/integrations/yellowcard/service/queue/payments.go`
- Test: `apps/integrations/yellowcard/service/queue/payments_test.go`

**Interfaces:**
- Produces: `queue.NewPaymentHandler(eventsMan, cli, catalog, settingsCli, cfg) queue.SubscribeWorker`

Destination resolution: `payment.GetRecipientAccount()` (`paymentv1.Account`: check its fields in `buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1` — use account number and name when present) with extras `account_type` (`bank`|`momo`), `network`/`bank_code`; default momo with `recipient.contact_id` phone. Channel: extras `channel_type` → account type. Sender party from extras / institution creds. `Reason` from extras `reason` default `other`.

- [ ] **Step 1: Tests**: `TestPayment_MomoHappyPath`, `TestPayment_BankHappyPath` (account number + bank_code hint resolved via networks by code), `TestPayment_ProviderError`, `TestPayment_MissingCredentials`, `TestPayment_BadPayload`.
- [ ] **Step 2: Implement; run — PASS. `make format`. Commit `feat(yellowcard): send payout worker`.**

---

### Task 7: Webhook server

**Files:**
- Create: `apps/integrations/yellowcard/service/handlers/webhooks.go`
- Test: `apps/integrations/yellowcard/service/handlers/webhooks_test.go`

**Interfaces:**
- Produces: `handlers.NewYellowcardWebhookServer(paymentCli paymentv1connect.PaymentServiceClient, cli client.YellowcardClient, resolver *credentials.Resolver) *YellowcardWebhookServer`, `(*YellowcardWebhookServer).NewRouterV1() *http.ServeMux` with `POST /webhook/yellowcard/receives`, `POST /webhook/yellowcard/sends`, `POST /webhook/yellowcard`, `GET /healthz`.
- Exported for tests: `handlers.MapStatus(status string) (commonv1.STATUS, commonv1.STATE)`.

Body limit `http.MaxBytesReader(w, r.Body, 1<<20)`. Flow per spec §4.5. Kind detection: route-specific handlers force kind; the catch-all uses `event` prefix (`RECEIVE.`/`COLLECTION.` → receive; `SEND.`/`PAYMENT.` → send; else 400). Lookup: `GetReceiveBySequenceID` (fallback `GetReceive(id)` when sequence empty); same for send. Extras per spec. `Id` = verified `SequenceID`, `ExternalId` = verified `ID`.

- [ ] **Step 1: Tests** using `fakePaymentClient` and `fakeYellowcardClient` (embedding interfaces, pawaPay style): `TestReceiveWebhook_Complete` (signed body → StatusUpdate SUCCESSFUL, Id=sequenceId, extras); `TestReceiveWebhook_Failed` (FAILED, STATE_INACTIVE, failure_code from errorCode); `TestReceiveWebhook_Pending` (IN_PROCESS); `TestWebhook_BadSignature` (401, no StatusUpdate); `TestWebhook_APIKeyMismatch` (401); `TestWebhook_UnknownSequence` (404); `TestWebhook_ProviderDown` (502); `TestWebhook_TenantFromQuery` (claims in ctx: verify via `security.ClaimsFromContext` in fake StatusUpdate); `TestCatchAll_DispatchesSend` (event `SEND.COMPLETE` → GetSendBySequenceID); `TestCatchAll_LegacyEventNames` (`COLLECTION.COMPLETE`); `TestHealthz`; `TestMapStatus` table.
- [ ] **Step 2: Implement; run — PASS. `make format`. Commit `feat(yellowcard): webhook server`.**

---

### Task 8: Service bootstrap, Dockerfile, Makefile, env example

**Files:**
- Create: `apps/integrations/yellowcard/cmd/main.go` (copy of pawaPay main with yellowcard packages; `cfg.ServiceName = "integration_payment_yellowcard"`; `catalog := client.NewCatalog(ycCli, time.Duration(cfg.CatalogCacheSeconds)*time.Second)`)
- Create: `apps/integrations/yellowcard/Dockerfile` (pawaPay Dockerfile with paths and title `Payments Yellow Card Integration`)
- Create: `apps/integrations/yellowcard/deploy/env.example`
- Modify: `Makefile:3` append `apps/integrations/yellowcard`

- [ ] **Step 1: Write files.**
- [ ] **Step 2: `go build ./... && go vet ./apps/integrations/yellowcard/... && go test ./apps/integrations/yellowcard/...` — PASS.**
- [ ] **Step 3: `docker build -f apps/integrations/yellowcard/Dockerfile .` only if docker is available (skip otherwise, note in commit).**
- [ ] **Step 4: `make format`. Commit `feat(yellowcard): service bootstrap and build wiring`.**

---

### Task 9: Checkout — bank transfer extras and Yellow Card method

**Files:**
- Modify: `apps/checkout/service/business/checkout.go` (`captureProviderExtras` ~1370-1446)
- Modify: `apps/checkout/service/business/methods.go` (`InferCountryFromPhone` ~519-550)
- Modify: `apps/checkout/service/handlers/render.go` (`PageData`)
- Modify: `apps/checkout/service/handlers/web.go` (`pageDataFor` ~585-597, `HandleStatus` ~849-899)
- Modify: `apps/checkout/service/web/templates/confirm.html`
- Modify: `apps/checkout/service/web/static/checkout.js`
- Modify: `apps/checkout/config/config.go:57`
- Modify: `ui/billing/lib/src/providers/collection_providers.dart:20-27`
- Tests: `apps/checkout/service/business/checkout_test.go`, `methods_test.go`, `apps/checkout/service/handlers/web_test.go`

**Interfaces (Produces):**

```go
// render.go
type BankTransferDetails struct { BankName, AccountNumber, AccountName, Reference, ExpiresAt string }
// PageData gains: BankTransfer *BankTransferDetails
// business: session metadata keys _bank_name, _bank_account_number, _bank_account_name, _payment_reference, _payment_expires_at
// web.go: HandleStatus payload map[string]any; "bank_transfer": {"bank_name","account_number","account_name","reference","expires_at"}
```

- [ ] **Step 1: Tests**: in `checkout_test.go` add a test that calls `captureProviderExtras` (or the exported path used by existing tests — check how existing tests exercise it, e.g. via `RefreshStatus` with a fake payment client) with `bank_name` etc. and asserts the `_bank_*` metadata. In `methods_test.go` add `InferCountryFromPhone` cases for `256`, `267`, `229`, `237`, `225`, `265`, `228`, `241`, `242`, `221`, `223`, `226`. In `web_test.go` add a `HandleStatus` test with a processing session carrying `_bank_*` metadata asserting the `bank_transfer` object, and a confirm-page render test asserting the account number appears.
- [ ] **Step 2: Run — FAIL.**
- [ ] **Step 3: Implement**:
  - `captureProviderExtras`: extend the `from/to` loop with `{collection.ExtraBankName, "_bank_name"}`, `{ExtraBankAccountNumber, "_bank_account_number"}`, `{ExtraBankAccountName, "_bank_account_name"}`, `{ExtraPaymentReference, "_payment_reference"}`, `{ExtraPaymentExpiresAt, "_payment_expires_at"}`; `next_action == "requires_bank_transfer"` is stored as-is.
  - `InferCountryFromPhone`: add cases (check `27` after `267`/`260`... note `strings.HasPrefix(phone,"27")` would wrongly match `+27x` vs `+267` — order 3-digit prefixes before `27`, and `1`/`44` last).
  - `render.go` `PageData.BankTransfer`; `pageDataFor` fills from metadata; `HandleStatus` payload type → `map[string]any`, add `bank_transfer` when `_bank_account_number` set.
  - `confirm.html`: after the instruction banner add
    ```html
    <div id="bank-transfer" class="bank-transfer{{if not .BankTransfer}} hidden{{end}}">
      <dl>
        <div><dt>{{t .Lang "bank_name"}}</dt><dd id="bt-bank">{{with .BankTransfer}}{{.BankName}}{{end}}</dd></div>
        <div><dt>{{t .Lang "bank_account_number"}}</dt><dd id="bt-account">{{with .BankTransfer}}{{.AccountNumber}}{{end}}</dd></div>
        <div><dt>{{t .Lang "bank_account_name"}}</dt><dd id="bt-name">{{with .BankTransfer}}{{.AccountName}}{{end}}</dd></div>
        <div><dt>{{t .Lang "bank_reference"}}</dt><dd id="bt-ref">{{with .BankTransfer}}{{.Reference}}{{end}}</dd></div>
        <div><dt>{{t .Lang "bank_expires"}}</dt><dd id="bt-expires">{{with .BankTransfer}}{{.ExpiresAt}}{{end}}</dd></div>
      </dl>
    </div>
    ```
    Add the five i18n keys to the translation table used by `t` (find it in `apps/checkout/service/handlers/i18n*.go` or similar; add EN + any other languages present with English fallback) and a `.bank-transfer` style in `checkout.css`.
  - `checkout.js`: in the poll handler, when `data.bank_transfer` is present, fill the `bt-*` elements and remove `hidden` from `#bank-transfer`; when `data.payment_instruction` is present and no banner exists, insert one.
  - `CHECKOUT_METHODS` default: append the `yellowcard` entry from spec §5.6.
  - Flutter map: `'yellowcard': 'Yellow Card'`.
- [ ] **Step 4: `go test ./apps/checkout/... ./pkg/...` — PASS. `make format`. Commit `feat(checkout): bank transfer instructions and Yellow Card method`.**

---

### Task 10: Documentation

**Files:**
- Create: `docs/yellowcard-integration.md` (structure of `docs/flutterwave-integration.md`: provider path, doc links, product experience diagram, architecture, auth with header/signature description, environments table, env quick start, sandbox test numbers, KYC notes and extras, webhooks setup incl. v2 event names and query-string tenancy, coverage table incl. "no Kenya receive", production checklist incl. static IP allowlist)
- Modify: `docs/collection-production.md` add `### Yellow Card (apps/integrations/yellowcard)` after the Flutterwave section with the env table and the two queue constraints; add `yellowcard` to any method table.

- [ ] **Step 1: Write docs. Commit `docs: Yellow Card integration guide`.**

---

### Task 11: Final verification

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `golangci-lint run ./apps/integrations/yellowcard/... ./apps/checkout/... ./pkg/...` (or `make lint`)
- [ ] `go test ./apps/integrations/yellowcard/... ./apps/checkout/... ./pkg/... -race`
- [ ] Manual smoke: `YELLOWCARD_API_KEY=x YELLOWCARD_SECRET_KEY=y go run ./apps/integrations/yellowcard/cmd` starts and serves `GET /healthz`.
- [ ] `git status` clean except intended changes.

## Self-review notes

- Spec §4.2 refund/cancel/deny/accept endpoints: Task 2. §4.4 prompt/payment: Tasks 5/6. §4.5 webhook: Task 7. §5 checkout: Task 9. §9 tests: Tasks 1-9. §10 docs: Task 10.
- Names used across tasks: `client.Catalog`, `client.SelectChannel`, `client.ResolveNetwork`, `queue.NewPromptHandler`, `queue.NewPaymentHandler`, `handlers.NewYellowcardWebhookServer`, `credentials.NewResolver` — consistent.
