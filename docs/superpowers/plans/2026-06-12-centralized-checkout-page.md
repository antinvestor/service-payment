# Centralized Checkout Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A self-contained `apps/checkout` service that server-renders a Stripe-Link-style hosted payments page at `pay.stawi.org`, exposes merchant Connect RPCs (sessions + links), executes payments through the existing payment-service prompt rails, and remembers payer "clues" in the profile payload.

**Architecture:** New frame-based Go app with its own Postgres tables (`checkout_sessions`, `checkout_links`). Merchant apps create sessions via Connect RPC; payers hit `GET /c/{ref}` (html/template + one vanilla JS file); `POST /pay` calls service-payment `InitiatePrompt` with `session_ref`/`order_ref` seeded into extras; the browser polls `/c/{ref}/status`; on success checkout writes clues into the payer's profile properties via service-profile.

**Tech Stack:** Go, frame (`github.com/pitabwire/frame`), Connect RPC (locally buf-generated under `apps/checkout/gen`), GORM via `datastore.BaseRepository`, `html/template` + `embed.FS`, testify + httptest, frametests/testcontainers for the repository suite.

**Spec:** `docs/superpowers/specs/2026-06-12-centralized-checkout-page-design.md`

**Out of scope for this plan** (follow-up plan after merge): cluster manifests (`deployment.manifests`), tenancy client seed (`service-authentication`), gateway routes for `pay.stawi.org` — mirror the pawaPay rollout (PRs stawi-org/deployment.manifests#41, antinvestor/service-authentication#715).

---

## File structure

```
proto/payment/checkout/v1/checkout.proto          # new package inside the payment buf module
proto/buf.gen.checkout.yaml                       # local Go generation template
Makefile                                          # APP_DIRS += apps/checkout; proto-generate-checkout target
apps/checkout/
├── cmd/main.go                                   # wiring: datastore, clients, RPC+web mux, sweeper
├── config/config.go                              # env config incl. method registry JSON, secrets, limits
├── Dockerfile                                    # copy of integration Dockerfile, path adjusted
├── migrations/0001/.gitkeep                      # automigrate via models; dir reserved for SQL
├── gen/payment/checkout/v1/                      # buf-generated (checked in)
└── service/
    ├── models/models.go                          # CheckoutSession, CheckoutLink
    ├── repository/interfaces.go                  # SessionRepository, LinkRepository
    ├── repository/sessions.go
    ├── repository/links.go
    ├── repository/migrate.go
    ├── repository/repository_test.go             # frametests suite (testcontainers postgres)
    ├── business/amount.go + amount_test.go       # decimal string <-> commonv1.Money
    ├── business/methods.go + methods_test.go     # method registry + preselection
    ├── business/clues.go + clues_test.go         # profile payload clues + guest cookie
    ├── business/checkout.go + checkout_test.go   # session/link lifecycle, pay, status refresh
    ├── handlers/rpc.go + rpc_test.go             # Connect CheckoutService
    ├── handlers/render.go + render_test.go       # templates, masking, i18n strings, CSRF
    ├── handlers/web.go + web_test.go             # /c/{ref}, /pay, /status, /l/{ref}
    └── web/
        ├── embed.go
        ├── templates/{layout,pay,confirm,done,gone}.html
        └── static/checkout.css, checkout.js
```

Every Go file carries the repo's standard Apache 2.0 header (copy the 13-line comment block from `apps/integrations/pawapay/config/config.go`).

---

### Task 1: Proto definition + local Go generation

**Files:**
- Create: `proto/payment/checkout/v1/checkout.proto`
- Create: `proto/buf.gen.checkout.yaml`
- Modify: `Makefile` (proto target; APP_DIRS comes in Task 12)

- [ ] **Step 1: Write the proto**

`proto/payment/checkout/v1/checkout.proto`:

```protobuf
syntax = "proto3";

package checkout.v1;

import "buf/validate/validate.proto";
import "common/v1/common.proto";
import "common/v1/money.proto";
import "common/v1/permissions.proto";

enum AmountOption {
  AMOUNT_OPTION_FIXED_UNSPECIFIED = 0;
  AMOUNT_OPTION_VARIABLE = 1;
}

enum SessionStatus {
  SESSION_STATUS_PENDING_UNSPECIFIED = 0;
  SESSION_STATUS_PROCESSING = 1;
  SESSION_STATUS_COMPLETED = 2;
  SESSION_STATUS_FAILED = 3;
  SESSION_STATUS_EXPIRED = 4;
}

message PayerContact {
  string contact_id = 1;
  string msisdn = 2;
}

message PayerPrefill {
  string profile_id = 1;
  string display_name = 2;
  string language = 3;
  repeated PayerContact contacts = 4;
}

message CheckoutSession {
  string ref = 1;
  string name = 2;
  string description = 3;
  common.v1.Money amount = 4;
  AmountOption amount_option = 5;
  string order_ref = 6;
  map<string, string> metadata = 7;
  string return_url = 8;
  PayerPrefill payer = 9;
  SessionStatus status = 10;
  string prompt_id = 11;
  string page_url = 12;
  string expires_at = 13; // RFC3339
}

message CheckoutLink {
  string ref = 1;
  string name = 2;
  string description = 3;
  common.v1.Money amount = 4;
  AmountOption amount_option = 5;
  string order_ref = 6;
  map<string, string> metadata = 7;
  string return_url = 8;
  string page_url = 9;
  string expires_at = 10; // RFC3339, empty = no expiry
  bool active = 11;
}

message CreateCheckoutSessionRequest {
  string name = 1 [
    (buf.validate.field).string.min_len = 3,
    (buf.validate.field).string.max_len = 100
  ];
  string description = 2 [(buf.validate.field).string.max_len = 500];
  common.v1.Money amount = 3;
  AmountOption amount_option = 4;
  string order_ref = 5 [(buf.validate.field).string.max_len = 250];
  map<string, string> metadata = 6;
  string return_url = 7 [(buf.validate.field).string.uri = true];
  PayerPrefill payer = 8;
  repeated string methods = 9; // optional restriction; method registry keys
}

message CreateCheckoutSessionResponse {
  CheckoutSession data = 1;
}

message GetCheckoutSessionRequest {
  string ref = 1 [
    (buf.validate.field).string.min_len = 8,
    (buf.validate.field).string.max_len = 64
  ];
}

message GetCheckoutSessionResponse {
  CheckoutSession data = 1;
}

message CreateCheckoutLinkRequest {
  string name = 1 [
    (buf.validate.field).string.min_len = 3,
    (buf.validate.field).string.max_len = 100
  ];
  string description = 2 [(buf.validate.field).string.max_len = 500];
  common.v1.Money amount = 3;
  AmountOption amount_option = 4;
  string order_ref = 5 [(buf.validate.field).string.max_len = 250];
  map<string, string> metadata = 6;
  string return_url = 7 [(buf.validate.field).string.uri = true];
  string expires_at = 8; // RFC3339, empty = no expiry
}

message CreateCheckoutLinkResponse {
  CheckoutLink data = 1;
}

service CheckoutService {
  option (common.v1.service_permissions) = {
    namespace: "service_checkout"
    permissions: [
      "checkout_session_create",
      "checkout_session_view",
      "checkout_link_create"
    ]
    role_bindings: [
      {
        role: ROLE_OWNER
        permissions: [
          "checkout_session_create",
          "checkout_session_view",
          "checkout_link_create"
        ]
      }
    ]
  };

  rpc CreateCheckoutSession(CreateCheckoutSessionRequest) returns (CreateCheckoutSessionResponse) {
    option (common.v1.method_permissions) = {
      permissions: ["checkout_session_create"]
    };
  }

  rpc GetCheckoutSession(GetCheckoutSessionRequest) returns (GetCheckoutSessionResponse) {
    option idempotency_level = NO_SIDE_EFFECTS;
    option (common.v1.method_permissions) = {
      permissions: ["checkout_session_view"]
    };
  }

  rpc CreateCheckoutLink(CreateCheckoutLinkRequest) returns (CreateCheckoutLinkResponse) {
    option (common.v1.method_permissions) = {
      permissions: ["checkout_link_create"]
    };
  }
}
```

- [ ] **Step 2: Write the local generation template**

`proto/buf.gen.checkout.yaml`:

```yaml
version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/antinvestor/service-payments/apps/checkout/gen
    - file_option: go_package_prefix
      module: buf.build/antinvestor/common
      value: buf.build/gen/go/antinvestor/common/protocolbuffers/go
  disable:
    - file_option: go_package_prefix
      module: buf.build/bufbuild/protovalidate
    - file_option: go_package_prefix
      module: buf.build/googleapis/googleapis
    - file_option: go_package_prefix
      module: buf.build/gnostic/gnostic
plugins:
  - local: protoc-gen-go
    out: ../apps/checkout/gen
    opt: paths=source_relative
  - local: protoc-gen-connect-go
    out: ../apps/checkout/gen
    opt: paths=source_relative
```

- [ ] **Step 3: Add Makefile target**

In `Makefile`, after the `proto-generate-dart` target, add:

```makefile
.PHONY: proto-generate-checkout
proto-generate-checkout: $(BIN)/buf ## Regenerate checkout Go stubs locally
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	@(cd $(PROTO_DIR) && PATH="$$(go env GOPATH)/bin:$$PATH" buf generate --template buf.gen.checkout.yaml --path payment/checkout)

proto-generate: proto-generate-checkout
```

- [ ] **Step 4: Generate and verify**

Run: `make proto-generate-checkout && ls apps/checkout/gen/payment/checkout/v1/`
Expected files: `checkout.pb.go` and `checkoutv1connect/checkout.connect.go`.
Run: `(cd proto && buf lint --path payment/checkout)` — expected: no errors (fix naming complaints if any by following the error messages; the enum zero-values above already follow buf style).
Run: `go build ./apps/checkout/...` — expected: PASS (generated code compiles; imports resolve to existing `buf.build/gen/go/antinvestor/common/...` dependency).

- [ ] **Step 5: Commit**

```bash
git add proto/payment/checkout proto/buf.gen.checkout.yaml Makefile apps/checkout/gen
git commit -m "feat(checkout): checkout/v1 proto with locally generated Go stubs"
```

---

### Task 2: Config

**Files:**
- Create: `apps/checkout/config/config.go`

- [ ] **Step 1: Write the config**

```go
package config

import (
	"github.com/pitabwire/frame/config"
)

type CheckoutConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                   string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	ProfileServiceURI                   string `envDefault:"127.0.0.1:7003"                  env:"PROFILE_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	ProfileServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-profile"  env:"PROFILE_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// PublicBaseURL is the externally visible origin of the checkout page.
	PublicBaseURL string `envDefault:"http://localhost:8080" env:"CHECKOUT_PUBLIC_BASE_URL"`

	// SigningSecret signs CSRF tokens and the guest hint cookie.
	SigningSecret string `env:"CHECKOUT_SIGNING_SECRET"`

	SessionTTLMinutes      int `envDefault:"30" env:"CHECKOUT_SESSION_TTL_MINUTES"`
	MaxAttempts            int `envDefault:"3"  env:"CHECKOUT_MAX_ATTEMPTS"`
	AttemptCooldownSeconds int `envDefault:"20" env:"CHECKOUT_ATTEMPT_COOLDOWN_SECONDS"`
	LinkSpawnPerMinute     int `envDefault:"10" env:"CHECKOUT_LINK_SPAWN_PER_MINUTE"`
	SweepIntervalSeconds   int `envDefault:"60" env:"CHECKOUT_SWEEP_INTERVAL_SECONDS"`

	// MethodsJSON is the v1 method registry: which payment methods the page
	// renders and the prompt route hint each maps to.
	MethodsJSON string `env:"CHECKOUT_METHODS" envDefault:"[{\"key\":\"mpesa\",\"name\":\"M-PESA\",\"route\":\"mpesa\",\"prefixes\":[\"254\"],\"currencies\":[\"KES\"]},{\"key\":\"mtn_momo\",\"name\":\"MTN MoMo\",\"route\":\"mtn\",\"prefixes\":[\"256\",\"260\"],\"currencies\":[\"UGX\",\"ZMW\"]},{\"key\":\"airtel_money\",\"name\":\"Airtel Money\",\"route\":\"airtel\",\"prefixes\":[\"255\",\"256\"],\"currencies\":[\"TZS\",\"UGX\"]},{\"key\":\"pawapay\",\"name\":\"Mobile Money\",\"route\":\"pawapay\",\"prefixes\":[],\"currencies\":[]}]"`
}
```

- [ ] **Step 2: Build and commit**

Run: `go build ./apps/checkout/...` — expected PASS.

```bash
git add apps/checkout/config
git commit -m "feat(checkout): service configuration"
```

---

### Task 3: Models, repositories, migrate

**Files:**
- Create: `apps/checkout/service/models/models.go`
- Create: `apps/checkout/service/repository/interfaces.go`
- Create: `apps/checkout/service/repository/sessions.go`
- Create: `apps/checkout/service/repository/links.go`
- Create: `apps/checkout/service/repository/migrate.go`
- Create: `apps/checkout/migrations/0001/.gitkeep`
- Test: `apps/checkout/service/repository/repository_test.go`

- [ ] **Step 1: Write the models**

`apps/checkout/service/models/models.go`:

```go
package models

import (
	"time"

	"github.com/pitabwire/frame/data"
)

// Session statuses.
const (
	SessionStatusPending    = "pending"
	SessionStatusProcessing = "processing"
	SessionStatusCompleted  = "completed"
	SessionStatusFailed     = "failed"
	SessionStatusExpired    = "expired"
)

// Amount options.
const (
	AmountOptionFixed    = "fixed"
	AmountOptionVariable = "variable"
)

// CheckoutSession is a single-payer, single-use payment session.
type CheckoutSession struct {
	data.BaseModel

	Ref            string       `gorm:"type:varchar(64);uniqueIndex;not null"     json:"ref"`
	LinkID         string       `gorm:"type:varchar(50);index"                    json:"link_id"`
	Name           string       `gorm:"type:varchar(100);not null"                json:"name"`
	Description    string       `gorm:"type:varchar(500)"                         json:"description"`
	Amount         string       `gorm:"type:varchar(40)"                          json:"amount"` // decimal string
	Currency       string       `gorm:"type:varchar(10);not null"                 json:"currency"`
	AmountOption   string       `gorm:"type:varchar(20);not null;default:'fixed'" json:"amount_option"`
	OrderRef       string       `gorm:"type:varchar(250);index"                   json:"order_ref"`
	Metadata       data.JSONMap `gorm:"type:jsonb"                                json:"metadata"`
	ReturnURL      string       `gorm:"type:varchar(500)"                         json:"return_url"`
	PayerProfileID string       `gorm:"type:varchar(50);index"                    json:"payer_profile_id"`
	Prefill        data.JSONMap `gorm:"type:jsonb"                                json:"prefill"`
	Methods        data.JSONMap `gorm:"type:jsonb"                                json:"methods"` // {"keys": [...]} restriction
	PromptID       string       `gorm:"type:varchar(50);index"                    json:"prompt_id"`
	PaymentID      string       `gorm:"type:varchar(50)"                          json:"payment_id"`
	Attempts       int          `gorm:"not null;default:0"                        json:"attempts"`
	LastAttemptAt  *time.Time   `                                                 json:"last_attempt_at"`
	Status         string       `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	ExpiresAt      time.Time    `gorm:"index"                                     json:"expires_at"`
}

func (CheckoutSession) TableName() string { return "checkout_sessions" }

// IsTerminal reports whether the session can no longer accept payment.
func (s *CheckoutSession) IsTerminal() bool {
	return s.Status == SessionStatusCompleted || s.Status == SessionStatusExpired
}

// CheckoutLink is a reusable template that spawns CheckoutSessions.
type CheckoutLink struct {
	data.BaseModel

	Ref          string       `gorm:"type:varchar(64);uniqueIndex;not null"     json:"ref"`
	Name         string       `gorm:"type:varchar(100);not null"                json:"name"`
	Description  string       `gorm:"type:varchar(500)"                         json:"description"`
	Amount       string       `gorm:"type:varchar(40)"                          json:"amount"`
	Currency     string       `gorm:"type:varchar(10);not null"                 json:"currency"`
	AmountOption string       `gorm:"type:varchar(20);not null;default:'fixed'" json:"amount_option"`
	OrderRef     string       `gorm:"type:varchar(250)"                         json:"order_ref"`
	Metadata     data.JSONMap `gorm:"type:jsonb"                                json:"metadata"`
	ReturnURL    string       `gorm:"type:varchar(500)"                         json:"return_url"`
	ExpiresAt    *time.Time   `                                                 json:"expires_at"`
	Active       bool         `gorm:"not null;default:true"                     json:"active"`
}

func (CheckoutLink) TableName() string { return "checkout_links" }

// IsUsable reports whether new sessions may be spawned from this link.
func (l *CheckoutLink) IsUsable(now time.Time) bool {
	if !l.Active {
		return false
	}
	if l.ExpiresAt != nil && now.After(*l.ExpiresAt) {
		return false
	}
	return true
}
```

- [ ] **Step 2: Write repository interfaces + implementations**

`apps/checkout/service/repository/interfaces.go`:

```go
package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/datastore"
)

type SessionRepository interface {
	datastore.BaseRepository[*models.CheckoutSession]
	GetByRef(ctx context.Context, ref string) (*models.CheckoutSession, error)
	ListByStatus(ctx context.Context, status string, limit int) ([]*models.CheckoutSession, error)
}

type LinkRepository interface {
	datastore.BaseRepository[*models.CheckoutLink]
	GetByRef(ctx context.Context, ref string) (*models.CheckoutLink, error)
}
```

`apps/checkout/service/repository/sessions.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

type sessionRepository struct {
	datastore.BaseRepository[*models.CheckoutSession]
}

func NewSessionRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) SessionRepository {
	return &sessionRepository{
		BaseRepository: datastore.NewBaseRepository[*models.CheckoutSession](
			ctx, dbPool, workMan, func() *models.CheckoutSession { return &models.CheckoutSession{} },
		),
	}
}

func (r *sessionRepository) GetByRef(ctx context.Context, ref string) (*models.CheckoutSession, error) {
	var session models.CheckoutSession
	if err := r.Pool().DB(ctx, false).Where("ref = ?", ref).First(&session).Error; err != nil {
		return nil, fmt.Errorf("get checkout session by ref: %w", err)
	}
	return &session, nil
}

func (r *sessionRepository) ListByStatus(
	ctx context.Context, status string, limit int,
) ([]*models.CheckoutSession, error) {
	var sessions []*models.CheckoutSession
	err := r.Pool().DB(ctx, true).
		Where("status = ?", status).
		Order("modified_at asc").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("list checkout sessions by status: %w", err)
	}
	return sessions, nil
}
```

Note for the implementer: confirm the accessor for the underlying pool/DB on `datastore.BaseRepository` by reading an existing custom query in this repo first — `apps/ledger/service/repository/*.go` uses the same pattern (e.g. its `GetByMode` style queries). Match whatever it does exactly (`r.Pool().DB(ctx, readOnly)` vs a direct pool field); the snippets above show intent and must compile against the same API the ledger repos use.

`apps/checkout/service/repository/links.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

type linkRepository struct {
	datastore.BaseRepository[*models.CheckoutLink]
}

func NewLinkRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) LinkRepository {
	return &linkRepository{
		BaseRepository: datastore.NewBaseRepository[*models.CheckoutLink](
			ctx, dbPool, workMan, func() *models.CheckoutLink { return &models.CheckoutLink{} },
		),
	}
}

func (r *linkRepository) GetByRef(ctx context.Context, ref string) (*models.CheckoutLink, error) {
	var link models.CheckoutLink
	if err := r.Pool().DB(ctx, false).Where("ref = ?", ref).First(&link).Error; err != nil {
		return nil, fmt.Errorf("get checkout link by ref: %w", err)
	}
	return &link, nil
}
```

`apps/checkout/service/repository/migrate.go`:

```go
package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/datastore"
)

// Migrate applies SQL migrations and auto-migrates the checkout models.
func Migrate(ctx context.Context, dbManager datastore.Manager, migrationPath string) error {
	dbPool := dbManager.GetPool(ctx, datastore.DefaultMigrationPoolName)
	return dbManager.Migrate(ctx, dbPool, migrationPath,
		&models.CheckoutSession{}, &models.CheckoutLink{})
}
```

- [ ] **Step 3: Write the repository suite test**

`apps/checkout/service/repository/repository_test.go` — follow the frametests pattern used in this repo (see `apps/ledger`'s test setup for the exact import paths of `frametests`, `definition`, and `testpostgres`; mirror them):

```go
package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/antinvestor/service-payments/apps/checkout/service/repository"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/frametests"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/frame/frametests/deps/testpostgres"
	"github.com/pitabwire/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RepositorySuite struct {
	frametests.FrameBaseTestSuite
}

func initResources(_ context.Context) []definition.TestResource {
	return []definition.TestResource{
		testpostgres.NewWithOpts("checkout_test", definition.WithUserName("test")),
	}
}

func (s *RepositorySuite) SetupSuite() {
	s.InitResourceFunc = initResources
	s.FrameBaseTestSuite.SetupSuite()
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

func (s *RepositorySuite) TestSessionSaveAndGetByRef() {
	ctx, svc := s.CreateService(s.T(), initResources) // mirror the ledger suite's service bootstrap exactly
	dbPool := svc.DatastoreManager().GetPool(ctx, datastore.DefaultPoolName)
	require.NoError(s.T(), repository.Migrate(ctx, svc.DatastoreManager(), "../../migrations/0001"))

	repo := repository.NewSessionRepository(ctx, dbPool, svc.WorkManager())

	session := &models.CheckoutSession{
		Ref:      util.RandomAlphaNumericString(32),
		Name:     "Test order",
		Amount:   "150",
		Currency: "KES",
		Status:   models.SessionStatusPending,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(s.T(), repo.Create(ctx, session))

	got, err := repo.GetByRef(ctx, session.Ref)
	require.NoError(s.T(), err)
	require.Equal(s.T(), session.GetID(), got.GetID())
	require.Equal(s.T(), "KES", got.Currency)
}

func (s *RepositorySuite) TestLinkSaveAndGetByRef() {
	ctx, svc := s.CreateService(s.T(), initResources)
	dbPool := svc.DatastoreManager().GetPool(ctx, datastore.DefaultPoolName)
	require.NoError(s.T(), repository.Migrate(ctx, svc.DatastoreManager(), "../../migrations/0001"))

	repo := repository.NewLinkRepository(ctx, dbPool, svc.WorkManager())

	link := &models.CheckoutLink{
		Ref:      util.RandomAlphaNumericString(12),
		Name:     "Donations",
		Currency: "KES",
		AmountOption: models.AmountOptionVariable,
		Active:   true,
	}
	require.NoError(s.T(), repo.Create(ctx, link))

	got, err := repo.GetByRef(ctx, link.Ref)
	require.NoError(s.T(), err)
	require.True(s.T(), got.IsUsable(time.Now()))
}
```

Before finalizing, open one existing suite test in this repo (search: `grep -rl "FrameBaseTestSuite" apps/`) and align the bootstrap calls (`CreateService` signature, migrate invocation) exactly to that file.

- [ ] **Step 4: Run tests**

Run: `go test ./apps/checkout/service/repository/ -run TestRepositorySuite -v` (requires Docker).
Expected: PASS (2 subtests).

- [ ] **Step 5: Commit**

```bash
git add apps/checkout/service/models apps/checkout/service/repository apps/checkout/migrations
git commit -m "feat(checkout): session and link models with repositories"
```

---

### Task 4: Amount helpers (TDD)

**Files:**
- Create: `apps/checkout/service/business/amount.go`
- Test: `apps/checkout/service/business/amount_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package business

import (
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		in        string
		units     int64
		nanos     int32
		expectErr bool
	}{
		{in: "150", units: 150},
		{in: "123.45", units: 123, nanos: 450000000},
		{in: "0.5", units: 0, nanos: 500000000},
		{in: "0", expectErr: true},        // zero not payable
		{in: "-5", expectErr: true},       // negative
		{in: "12.345", expectErr: true},   // > 2dp
		{in: "abc", expectErr: true},
		{in: "", expectErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			units, nanos, err := ParseAmount(tt.in)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.units, units)
			assert.Equal(t, tt.nanos, nanos)
		})
	}
}

func TestMoneyFromAmount(t *testing.T) {
	m, err := MoneyFromAmount("123.45", "KES")
	require.NoError(t, err)
	assert.Equal(t, "KES", m.GetCurrencyCode())
	assert.Equal(t, int64(123), m.GetUnits())
	assert.Equal(t, int32(450000000), m.GetNanos())
}

func TestFormatMoney(t *testing.T) {
	assert.Equal(t, "KES 123.45",
		FormatMoney(&commonv1.Money{CurrencyCode: "KES", Units: 123, Nanos: 450000000}))
	assert.Equal(t, "KES 150.00",
		FormatMoney(&commonv1.Money{CurrencyCode: "KES", Units: 150}))
}

func TestAmountString(t *testing.T) {
	assert.Equal(t, "123.45", AmountString(&commonv1.Money{Units: 123, Nanos: 450000000}))
	assert.Equal(t, "150", AmountString(&commonv1.Money{Units: 150}))
}
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./apps/checkout/service/business/ -run "TestParseAmount|TestMoneyFromAmount|TestFormatMoney|TestAmountString" -v`
Expected: FAIL (undefined functions).

- [ ] **Step 3: Implement**

`apps/checkout/service/business/amount.go`:

```go
package business

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
)

const (
	nanosPerCent  = 10_000_000
	centsPerUnit  = 100
	maxDecimalDigits = 2
)

// ParseAmount parses a positive decimal string with at most two decimal
// places into Money units and nanos.
func ParseAmount(amount string) (int64, int32, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return 0, 0, errors.New("amount is required")
	}
	wholePart, fracPart, _ := strings.Cut(amount, ".")
	if len(fracPart) > maxDecimalDigits {
		return 0, 0, fmt.Errorf("amount %q has more than %d decimal places", amount, maxDecimalDigits)
	}
	units, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid amount %q", amount)
	}
	var cents int64
	if fracPart != "" {
		// pad "5" -> "50"
		for len(fracPart) < maxDecimalDigits {
			fracPart += "0"
		}
		cents, err = strconv.ParseInt(fracPart, 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid amount %q", amount)
		}
	}
	if units < 0 || (units == 0 && cents == 0) || strings.HasPrefix(amount, "-") {
		return 0, 0, fmt.Errorf("amount must be positive, got %q", amount)
	}
	return units, int32(cents * nanosPerCent), nil
}

// MoneyFromAmount builds a commonv1.Money from a decimal string and currency.
func MoneyFromAmount(amount, currency string) (*commonv1.Money, error) {
	units, nanos, err := ParseAmount(amount)
	if err != nil {
		return nil, err
	}
	return &commonv1.Money{CurrencyCode: currency, Units: units, Nanos: nanos}, nil
}

// FormatMoney renders Money for display, always with two decimal places.
func FormatMoney(m *commonv1.Money) string {
	cents := (int64(m.GetNanos()) + nanosPerCent/2) / nanosPerCent
	return fmt.Sprintf("%s %d.%02d", m.GetCurrencyCode(), m.GetUnits(), cents)
}

// AmountString renders Money as a bare decimal string without trailing zeros.
func AmountString(m *commonv1.Money) string {
	cents := (int64(m.GetNanos()) + nanosPerCent/2) / nanosPerCent
	if cents == 0 {
		return strconv.FormatInt(m.GetUnits(), 10)
	}
	s := fmt.Sprintf("%d.%02d", m.GetUnits(), cents)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}
```

- [ ] **Step 4: Run tests — expected PASS, then commit**

```bash
git add apps/checkout/service/business/amount.go apps/checkout/service/business/amount_test.go
git commit -m "feat(checkout): amount parsing and money formatting helpers"
```

---

### Task 5: Method registry + preselection (TDD)

**Files:**
- Create: `apps/checkout/service/business/methods.go`
- Test: `apps/checkout/service/business/methods_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package business

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMethodsJSON = `[
  {"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]},
  {"key":"mtn_momo","name":"MTN MoMo","route":"mtn","prefixes":["256","260"],"currencies":["UGX","ZMW"]},
  {"key":"pawapay","name":"Mobile Money","route":"pawapay","prefixes":[],"currencies":[]}
]`

func TestParseMethodRegistry(t *testing.T) {
	reg, err := ParseMethodRegistry(testMethodsJSON)
	require.NoError(t, err)
	require.Len(t, reg.Methods, 3)
	assert.Equal(t, "mpesa", reg.Methods[0].Key)
	assert.Equal(t, "mtn", reg.Methods[1].Route)

	_, err = ParseMethodRegistry("not json")
	require.Error(t, err)

	_, err = ParseMethodRegistry("[]")
	require.Error(t, err, "empty registry must be rejected")
}

func TestAvailableMethods(t *testing.T) {
	reg, _ := ParseMethodRegistry(testMethodsJSON)

	all := reg.Available(nil)
	assert.Len(t, all, 3)

	restricted := reg.Available([]string{"mpesa", "unknown"})
	require.Len(t, restricted, 1)
	assert.Equal(t, "mpesa", restricted[0].Key)
}

func TestPreselect(t *testing.T) {
	reg, _ := ParseMethodRegistry(testMethodsJSON)
	methods := reg.Available(nil)

	t.Run("clue wins", func(t *testing.T) {
		m := Preselect(methods, "mtn_momo", "254712345678")
		assert.Equal(t, "mtn_momo", m.Key)
	})
	t.Run("phone prefix when no clue", func(t *testing.T) {
		m := Preselect(methods, "", "254712345678")
		assert.Equal(t, "mpesa", m.Key)
	})
	t.Run("first method as fallback", func(t *testing.T) {
		m := Preselect(methods, "", "")
		assert.Equal(t, "mpesa", m.Key)
	})
	t.Run("unknown clue falls through to prefix", func(t *testing.T) {
		m := Preselect(methods, "card", "260763456789")
		assert.Equal(t, "mtn_momo", m.Key)
	})
}
```

- [ ] **Step 2: Run — expected FAIL (undefined). Step 3: Implement**

`apps/checkout/service/business/methods.go`:

```go
package business

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Method describes one renderable payment method and its prompt route hint.
type Method struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Route      string   `json:"route"`
	Prefixes   []string `json:"prefixes"`
	Currencies []string `json:"currencies"`
}

// MethodRegistry is the config-defined list of supported methods.
type MethodRegistry struct {
	Methods []Method
}

// ParseMethodRegistry parses the CHECKOUT_METHODS config JSON.
func ParseMethodRegistry(raw string) (*MethodRegistry, error) {
	var methods []Method
	if err := json.Unmarshal([]byte(raw), &methods); err != nil {
		return nil, fmt.Errorf("parse method registry: %w", err)
	}
	if len(methods) == 0 {
		return nil, errors.New("method registry is empty")
	}
	return &MethodRegistry{Methods: methods}, nil
}

// Available returns methods filtered by an optional restriction key list.
func (r *MethodRegistry) Available(restriction []string) []Method {
	if len(restriction) == 0 {
		return r.Methods
	}
	allowed := make(map[string]bool, len(restriction))
	for _, k := range restriction {
		allowed[k] = true
	}
	var out []Method
	for _, m := range r.Methods {
		if allowed[m.Key] {
			out = append(out, m)
		}
	}
	return out
}

// Get returns the method for a key, or false.
func (r *MethodRegistry) Get(key string) (Method, bool) {
	for _, m := range r.Methods {
		if m.Key == key {
			return m, true
		}
	}
	return Method{}, false
}

// Preselect picks the default method: profile clue -> phone prefix -> first.
func Preselect(methods []Method, clueKey, phoneNumber string) Method {
	for _, m := range methods {
		if clueKey != "" && m.Key == clueKey {
			return m
		}
	}
	phone := strings.TrimPrefix(strings.TrimSpace(phoneNumber), "+")
	if phone != "" {
		for _, m := range methods {
			for _, p := range m.Prefixes {
				if strings.HasPrefix(phone, p) {
					return m
				}
			}
		}
	}
	return methods[0]
}
```

- [ ] **Step 4: Run tests — PASS. Step 5: Commit**

```bash
git add apps/checkout/service/business/methods.go apps/checkout/service/business/methods_test.go
git commit -m "feat(checkout): method registry with clue/prefix preselection"
```

---

### Task 6: Clues — profile payload + guest cookie (TDD)

**Files:**
- Create: `apps/checkout/service/business/clues.go`
- Test: `apps/checkout/service/business/clues_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package business

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCluesFromProfileProperties(t *testing.T) {
	props, _ := structpb.NewStruct(map[string]any{
		"checkout": map[string]any{
			"lastMethod":    "mobile_money",
			"lastProvider":  "mpesa",
			"lastContactId": "contact-1",
			"lastCurrency":  "KES",
		},
	})
	clues := CluesFromProperties(props)
	assert.Equal(t, "mpesa", clues.LastProvider)
	assert.Equal(t, "contact-1", clues.LastContactID)

	assert.Equal(t, Clues{}, CluesFromProperties(nil))
	empty, _ := structpb.NewStruct(map[string]any{"other": "x"})
	assert.Equal(t, Clues{}, CluesFromProperties(empty))
}

func TestCluesToProperties(t *testing.T) {
	c := Clues{LastMethod: "mobile_money", LastProvider: "mpesa", LastContactID: "c1", LastCurrency: "KES", LastPaidAt: "2026-06-12T00:00:00Z"}
	props := c.ToProperties()
	checkout := props.GetFields()["checkout"].GetStructValue()
	require.NotNil(t, checkout)
	assert.Equal(t, "mpesa", checkout.GetFields()["lastProvider"].GetStringValue())
}

func TestGuestCookieRoundTrip(t *testing.T) {
	secret := []byte("test-secret")
	val := EncodeGuestHints(secret, GuestHints{Phone: "254712345678", Method: "mpesa"})

	hints, ok := DecodeGuestHints(secret, val)
	require.True(t, ok)
	assert.Equal(t, "254712345678", hints.Phone)
	assert.Equal(t, "mpesa", hints.Method)
}

func TestGuestCookieTamperRejected(t *testing.T) {
	secret := []byte("test-secret")
	val := EncodeGuestHints(secret, GuestHints{Phone: "254712345678", Method: "mpesa"})

	_, ok := DecodeGuestHints(secret, val+"x")
	assert.False(t, ok)
	_, ok = DecodeGuestHints([]byte("other-secret"), val)
	assert.False(t, ok)
	_, ok = DecodeGuestHints(secret, "garbage")
	assert.False(t, ok)
}
```

- [ ] **Step 2: Run — FAIL. Step 3: Implement**

`apps/checkout/service/business/clues.go`:

```go
package business

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

// Clues are the quick-repeat hints stored under the "checkout" key of a
// profile's properties payload.
type Clues struct {
	LastMethod    string `json:"lastMethod"`
	LastProvider  string `json:"lastProvider"`
	LastContactID string `json:"lastContactId"`
	LastCurrency  string `json:"lastCurrency"`
	LastPaidAt    string `json:"lastPaidAt"`
}

// CluesFromProperties extracts checkout clues from profile properties.
func CluesFromProperties(props *structpb.Struct) Clues {
	if props == nil {
		return Clues{}
	}
	checkout := props.GetFields()["checkout"].GetStructValue()
	if checkout == nil {
		return Clues{}
	}
	f := checkout.GetFields()
	return Clues{
		LastMethod:    f["lastMethod"].GetStringValue(),
		LastProvider:  f["lastProvider"].GetStringValue(),
		LastContactID: f["lastContactId"].GetStringValue(),
		LastCurrency:  f["lastCurrency"].GetStringValue(),
		LastPaidAt:    f["lastPaidAt"].GetStringValue(),
	}
}

// ToProperties renders the clues as a properties patch for profile Update.
func (c Clues) ToProperties() *structpb.Struct {
	props, _ := structpb.NewStruct(map[string]any{
		"checkout": map[string]any{
			"lastMethod":    c.LastMethod,
			"lastProvider":  c.LastProvider,
			"lastContactId": c.LastContactID,
			"lastCurrency":  c.LastCurrency,
			"lastPaidAt":    c.LastPaidAt,
		},
	})
	return props
}

// GuestHints are device-local hints for unauthenticated payers.
type GuestHints struct {
	Phone  string `json:"phone"`
	Method string `json:"method"`
}

// EncodeGuestHints signs hints into a cookie value: v1.<b64(json)>.<hmac-hex>.
func EncodeGuestHints(secret []byte, hints GuestHints) string {
	payload, _ := json.Marshal(hints)
	body := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(body))
	return "v1." + body + "." + hex.EncodeToString(mac.Sum(nil))
}

// DecodeGuestHints verifies and decodes a cookie value.
func DecodeGuestHints(secret []byte, value string) (GuestHints, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 || parts[0] != "v1" {
		return GuestHints{}, false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[1]))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return GuestHints{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return GuestHints{}, false
	}
	var hints GuestHints
	if err = json.Unmarshal(payload, &hints); err != nil {
		return GuestHints{}, false
	}
	return hints, true
}
```

- [ ] **Step 4: Run tests — PASS. Step 5: Commit**

```bash
git add apps/checkout/service/business/clues.go apps/checkout/service/business/clues_test.go
git commit -m "feat(checkout): profile clue payload and signed guest hints"
```

---

### Task 7: Checkout business — sessions, links, pay, status (TDD with fakes)

**Files:**
- Create: `apps/checkout/service/business/checkout.go`
- Test: `apps/checkout/service/business/checkout_test.go`

The business struct and its dependencies:

```go
// CheckoutBusiness orchestrates session/link lifecycle and payment execution.
type CheckoutBusiness struct {
	cfg         *config.CheckoutConfig
	registry    *MethodRegistry
	sessionRepo repository.SessionRepository
	linkRepo    repository.LinkRepository
	paymentCli  paymentv1connect.PaymentServiceClient
	profileCli  profilev1connect.ProfileServiceClient
	now         func() time.Time // injectable clock for tests
}

func NewCheckoutBusiness(
	cfg *config.CheckoutConfig,
	registry *MethodRegistry,
	sessionRepo repository.SessionRepository,
	linkRepo repository.LinkRepository,
	paymentCli paymentv1connect.PaymentServiceClient,
	profileCli profilev1connect.ProfileServiceClient,
) *CheckoutBusiness {
	return &CheckoutBusiness{
		cfg: cfg, registry: registry,
		sessionRepo: sessionRepo, linkRepo: linkRepo,
		paymentCli: paymentCli, profileCli: profileCli,
		now: time.Now,
	}
}
```

Public methods (each TDD'd below):

| Method | Behaviour |
|---|---|
| `CreateSession(ctx, in *CreateSessionInput) (*models.CheckoutSession, error)` | validates amount/currency, hydrates clues from profile when `payer_profile_id` set, generates `Ref = util.RandomAlphaNumericString(32)`, persists with TTL |
| `GetSessionByRef(ctx, ref)` | loads; lazily flips PENDING/PROCESSING past `ExpiresAt` to EXPIRED (persisting) |
| `CreateLink(ctx, in *CreateLinkInput) (*models.CheckoutLink, error)` | validates; `Ref = util.RandomAlphaNumericString(12)` |
| `SpawnSession(ctx, linkRef) (*models.CheckoutSession, error)` | rejects unusable links; stamps a session copying link fields, `LinkID` set |
| `Pay(ctx, ref string, in PayInput) (*models.CheckoutSession, error)` | guards: terminal/expired → `ErrSessionGone`; `Attempts >= cfg.MaxAttempts` → `ErrTooManyAttempts`; cooldown via `LastAttemptAt` → `ErrCooldown`; VARIABLE amount comes from `in.Amount`; resolves method via registry (clue/prefix preselect already happened client-side — `in.MethodKey` must exist in registry else `ErrUnknownMethod`); calls `InitiatePrompt`; stores `PromptID`, bumps `Attempts`, sets PROCESSING |
| `RefreshStatus(ctx, session) (*models.CheckoutSession, error)` | no-op unless PROCESSING; calls payment `Status` RPC; maps SUCCESSFUL→completed (then async clue write-back), FAILED→failed, else unchanged |
| `SweepProcessing(ctx) error` | `ListByStatus(processing, 50)` → `RefreshStatus` each; expire overdue PENDING |

`PayInput`:

```go
type PayInput struct {
	MethodKey   string
	PhoneNumber string // guests; for recognized payers resolved server-side from prefill ContactID
	ContactID   string // recognized payers: which prefilled contact to charge
	Amount      string // VARIABLE sessions only
}
```

Sentinel errors:

```go
var (
	ErrSessionGone     = errors.New("checkout session is no longer payable")
	ErrTooManyAttempts = errors.New("too many payment attempts for this session")
	ErrCooldown        = errors.New("please wait before retrying")
	ErrUnknownMethod   = errors.New("unknown payment method")
	ErrLinkUnusable    = errors.New("checkout link is not usable")
	ErrAmountRequired  = errors.New("an amount is required for this payment")
)
```

The `InitiatePrompt` call built in `Pay` (exact shape — field names from `proto/payment/v1/payment.proto`):

```go
extra, _ := structpb.NewStruct(map[string]any{
	"session_ref": session.Ref,
	"order_ref":   session.OrderRef,
	"provider":    method.Key,
})
for k, v := range session.Metadata {
	if s, ok := v.(string); ok {
		extra.Fields["meta_"+k] = structpb.NewStringValue(s)
	}
}

amount, err := MoneyFromAmount(session.Amount, session.Currency)
if err != nil {
	return nil, err
}

promptReq := &paymentv1.InitiatePromptRequest{
	Source: &commonv1.ContactLink{
		ProfileId: session.PayerProfileID,
		ContactId: contactRef, // contact id for recognized payers, raw msisdn for guests
	},
	Amount: amount,
	Route:  method.Route,
	Extra:  extra,
}

resp, err := b.paymentCli.InitiatePrompt(ctx, connect.NewRequest(promptReq))
if err != nil {
	return nil, fmt.Errorf("initiate prompt: %w", err)
}
session.PromptID = resp.Msg.GetData().GetId()
```

Implementer note: open `apps/default/service/handlers/payment.go` to confirm the exact response accessor for `InitiatePrompt` (`resp.Msg.GetData()` vs direct fields on the response message) and align.

Status mapping in `RefreshStatus`:

```go
statusResp, err := b.paymentCli.Status(ctx, connect.NewRequest(&paymentv1.StatusRequest{Id: session.PromptID}))
if err != nil {
	return session, nil // transient: stay PROCESSING, sweeper retries
}
switch statusResp.Msg.GetData().GetStatus() {
case commonv1.STATUS_SUCCESSFUL:
	session.Status = models.SessionStatusCompleted
	session.PaymentID = statusResp.Msg.GetData().GetExternalId()
	b.writeCluesAsync(ctx, session) // best-effort profile Update
case commonv1.STATUS_FAILED:
	session.Status = models.SessionStatusFailed
}
```

`writeCluesAsync` submits to the frame worker pool (never blocks): reads the profile, merges `Clues{...}.ToProperties()`, calls `profileCli.Update(ctx, connect.NewRequest(&profilev1.UpdateRequest{Id: session.PayerProfileID, Properties: props}))`. Skip entirely when `PayerProfileID == ""` or the paying `ContactID` is not among the prefill contacts.

- [ ] **Step 1: Write failing tests** — `checkout_test.go` with hand-rolled fakes (same embedding trick as the pawapay tests):

```go
package business_test

// fakes: embed the generated interfaces, override only what's called.
type fakeSessionRepo struct {
	repository.SessionRepository
	byRef map[string]*models.CheckoutSession
	saved []*models.CheckoutSession
}
func (f *fakeSessionRepo) Create(_ context.Context, s *models.CheckoutSession) error {
	f.byRef[s.Ref] = s; f.saved = append(f.saved, s); return nil
}
func (f *fakeSessionRepo) Update(_ context.Context, s *models.CheckoutSession) error {
	f.byRef[s.Ref] = s; return nil
}
func (f *fakeSessionRepo) GetByRef(_ context.Context, ref string) (*models.CheckoutSession, error) {
	s, ok := f.byRef[ref]
	if !ok { return nil, gorm.ErrRecordNotFound }
	return s, nil
}

type fakePaymentClient struct {
	paymentv1connect.PaymentServiceClient
	promptReq  *paymentv1.InitiatePromptRequest
	promptErr  error
	statusResp commonv1.STATUS
}
func (f *fakePaymentClient) InitiatePrompt(_ context.Context, req *connect.Request[paymentv1.InitiatePromptRequest]) (*connect.Response[paymentv1.InitiatePromptResponse], error) { ... }
func (f *fakePaymentClient) Status(_ context.Context, req *connect.Request[paymentv1.StatusRequest]) (*connect.Response[paymentv1.StatusResponse], error) { ... }
```

(Implementer: check the exact request/response generic types in `buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect` and fill the two fake bodies to record the request and return canned responses.)

Test cases (each a `t.Run`):
1. `CreateSession` persists with 32-char ref, pending status, TTL = cfg minutes, tenancy-agnostic fields copied.
2. `CreateSession` with payer profile fetches profile (fake profileCli returns properties with clues) and stores `prefill` containing displayName/contacts/clue provider.
3. `GetSessionByRef` flips an expired pending session to `expired` and persists.
4. `CreateLink` + `SpawnSession` copies fields, sets `LinkID`, fresh ref; inactive/expired link → `ErrLinkUnusable`.
5. `Pay` happy path: prompt called with correct Money/route/extras (assert `session_ref` + `order_ref` + `meta_*` in extras), session → processing, attempts=1, promptID stored.
6. `Pay` on completed session → `ErrSessionGone`; attempts at max → `ErrTooManyAttempts`; immediate retry inside cooldown → `ErrCooldown` (use injected clock); unknown method → `ErrUnknownMethod`; VARIABLE session without amount → `ErrAmountRequired`, with amount uses it.
7. `RefreshStatus`: SUCCESSFUL → completed + clue write-back invoked (fake profileCli records Update with `checkout.lastProvider`); FAILED → failed; transient status error → stays processing.
8. `SweepProcessing` expires overdue pending sessions and refreshes processing ones.

- [ ] **Step 2: Run — FAIL. Step 3: Implement `checkout.go`** per the method table above. Keep each method under ~50 lines; pull guards into small helpers (`guardPayable`, `resolveContact`).
- [ ] **Step 4: Run — PASS: `go test ./apps/checkout/service/business/ -v`**
- [ ] **Step 5: Commit**

```bash
git add apps/checkout/service/business
git commit -m "feat(checkout): session/link lifecycle, payment execution and status refresh"
```

---

### Task 8: Render layer — templates, masking, i18n, CSRF (TDD)

**Files:**
- Create: `apps/checkout/service/web/embed.go`
- Create: `apps/checkout/service/web/templates/layout.html`, `pay.html`, `confirm.html`, `done.html`, `gone.html`
- Create: `apps/checkout/service/web/static/checkout.css`, `checkout.js`
- Create: `apps/checkout/service/handlers/render.go`
- Test: `apps/checkout/service/handlers/render_test.go`

- [ ] **Step 1: Failing tests**

```go
package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaskMsisdn(t *testing.T) {
	assert.Equal(t, "+254 7•• ••789", MaskMsisdn("254712345789"))
	assert.Equal(t, "+254 7•• ••789", MaskMsisdn("+254712345789"))
	assert.Equal(t, "••••", MaskMsisdn("12345"))   // too short: fully masked
	assert.Equal(t, "••••", MaskMsisdn(""))
	// property: no more than the first 4 and last 3 digits survive
	masked := MaskMsisdn("254712345789")
	digits := 0
	for _, r := range masked {
		if r >= '0' && r <= '9' { digits++ }
	}
	assert.LessOrEqual(t, digits, 7)
}

func TestCSRFTokenRoundTrip(t *testing.T) {
	secret := []byte("s")
	tok := CSRFToken(secret, "session-ref-1")
	assert.True(t, VerifyCSRF(secret, "session-ref-1", tok))
	assert.False(t, VerifyCSRF(secret, "session-ref-2", tok))
	assert.False(t, VerifyCSRF(secret, "session-ref-1", tok+"x"))
}

func TestTranslate(t *testing.T) {
	assert.Equal(t, "Pay", T("en", "pay_button"))
	assert.Equal(t, "Payer", T("fr", "pay_button"))
	assert.Equal(t, "Pay", T("de", "pay_button"), "unknown language falls back to en")
	assert.Equal(t, "pay_button", T("en", "missing_key"), "missing key returns the key")
}

func TestAllTemplatesRenderInBothLanguages(t *testing.T) {
	r := NewRenderer([]byte("secret"))
	for _, lang := range []string{"en", "fr"} {
		for _, page := range []string{"pay", "confirm", "done", "gone"} {
			var sb strings.Builder
			err := r.Render(&sb, page, samplePageData(lang, page))
			require.NoError(t, err, "%s/%s", page, lang)
			assert.NotEmpty(t, sb.String())
			assert.NotContains(t, sb.String(), "254712345789", "raw msisdn must never render")
		}
	}
}
```

`samplePageData` is a test helper in the same file producing a fully populated `PageData` for each page name.

- [ ] **Step 2: Run — FAIL. Step 3: Implement**

`render.go` essentials:

```go
// PageData is everything a template can see. No raw contact details.
type PageData struct {
	Lang          string
	SessionRef    string
	MerchantName  string
	Description   string
	AmountDisplay string   // formatted, or "" for VARIABLE
	Variable      bool
	Currency      string
	PayerName     string   // first name only, "" for guests
	MaskedPhone   string
	Contacts      []ContactChoice // {ContactID, Masked}
	Methods       []MethodChoice  // {Key, Name, Selected}
	CSRF          string
	Status        string
	FailureReason string
	ReturnURL     string
	PollURL       string
}

func MaskMsisdn(msisdn string) string {
	digits := strings.TrimPrefix(strings.TrimSpace(msisdn), "+")
	if len(digits) < 9 {
		return "••••"
	}
	return "+" + digits[:3] + " " + digits[3:4] + "•• ••" + digits[len(digits)-3:]
}

func CSRFToken(secret []byte, ref string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("csrf:" + ref))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyCSRF(secret []byte, ref, token string) bool {
	return hmac.Equal([]byte(CSRFToken(secret, ref)), []byte(token))
}

var translations = map[string]map[string]string{
	"en": {
		"pay_button": "Pay", "paying_as": "Paying as", "change": "Change",
		"phone_label": "Mobile money number", "confirm_title": "Confirm on your phone",
		"confirm_hint": "Enter your PIN on your phone to approve the payment.",
		"retry": "Didn't get the prompt? Try again", "done_title": "Payment received",
		"redirecting": "Taking you back…", "failed_title": "Payment failed",
		"try_again": "Try again", "gone_title": "This payment is no longer available",
		"amount_label": "Amount",
	},
	"fr": {
		"pay_button": "Payer", "paying_as": "Payer en tant que", "change": "Modifier",
		"phone_label": "Numéro mobile money", "confirm_title": "Confirmez sur votre téléphone",
		"confirm_hint": "Saisissez votre code PIN sur votre téléphone pour approuver le paiement.",
		"retry": "Pas d'invite ? Réessayer", "done_title": "Paiement reçu",
		"redirecting": "Redirection…", "failed_title": "Échec du paiement",
		"try_again": "Réessayer", "gone_title": "Ce paiement n'est plus disponible",
		"amount_label": "Montant",
	},
}

func T(lang, key string) string {
	if m, ok := translations[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := translations["en"][key]; ok {
		return v
	}
	return key
}

type Renderer struct {
	tmpl   *template.Template
	secret []byte
}

func NewRenderer(secret []byte) *Renderer {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{"t": T}).
		ParseFS(web.Templates, "templates/*.html"))
	return &Renderer{tmpl: tmpl, secret: secret}
}

func (r *Renderer) Render(w io.Writer, page string, data PageData) error {
	return r.tmpl.ExecuteTemplate(w, page+".html", data)
}
```

`web/embed.go`:

```go
// Package web embeds the checkout page assets.
package web

import "embed"

//go:embed templates/*.html
var Templates embed.FS

//go:embed static/*
var Static embed.FS
```

Templates (complete, minimal-but-real; `layout.html` defines blocks, others fill them). `pay.html` core:

```html
{{define "pay.html"}}{{template "layout_top" .}}
<form method="POST" action="/c/{{.SessionRef}}/pay" class="card">
  <h1>{{.MerchantName}}</h1>
  {{if .Description}}<p class="desc">{{.Description}}</p>{{end}}
  {{if .Variable}}
    <label>{{t .Lang "amount_label"}} ({{.Currency}})
      <input name="amount" inputmode="decimal" required autofocus>
    </label>
  {{else}}<div class="amount">{{.AmountDisplay}}</div>{{end}}

  {{if .PayerName}}
    <div class="payer">{{t .Lang "paying_as"}} <strong>{{.PayerName}}</strong> · {{.MaskedPhone}}
      <button type="button" id="change" class="link">{{t .Lang "change"}}</button></div>
    <div id="contact-edit" hidden>
      {{range .Contacts}}<label><input type="radio" name="contact_id" value="{{.ContactID}}"> {{.Masked}}</label>{{end}}
    </div>
  {{else}}
    <label>{{t .Lang "phone_label"}}<input name="phone" inputmode="tel" required autofocus></label>
  {{end}}

  <div class="methods">
    {{range .Methods}}
      <label class="chip"><input type="radio" name="method" value="{{.Key}}" {{if .Selected}}checked{{end}}>{{.Name}}</label>
    {{end}}
  </div>

  <input type="hidden" name="csrf" value="{{.CSRF}}">
  <button type="submit" class="pay">{{t .Lang "pay_button"}}{{if .AmountDisplay}} {{.AmountDisplay}}{{end}}</button>
</form>
{{template "layout_bottom" .}}{{end}}
```

`confirm.html` shows the spinner + `t .Lang "confirm_hint"` and embeds `<script>window.__poll="{{.PollURL}}";window.__return="{{.ReturnURL}}";</script>`; `checkout.js` polls `__poll` every 2s, on `completed` redirects to `__return`, on `failed` reloads to show state 1 with the failure banner. `done.html` renders ✓ + meta-refresh fallback to `.ReturnURL`. `gone.html` renders `gone_title` only — no payer or amount details.

- [ ] **Step 4: Run `go test ./apps/checkout/service/handlers/ -run "TestMask|TestCSRF|TestTranslate|TestAllTemplates" -v` — PASS. Step 5: Commit**

```bash
git add apps/checkout/service/web apps/checkout/service/handlers/render.go apps/checkout/service/handlers/render_test.go
git commit -m "feat(checkout): server-rendered page templates with masking and i18n"
```

---

### Task 9: Web handlers (TDD)

**Files:**
- Create: `apps/checkout/service/handlers/web.go`
- Test: `apps/checkout/service/handlers/web_test.go`

`WebServer` struct: `{business *business.CheckoutBusiness, renderer *Renderer, registry *business.MethodRegistry, cfg *config.CheckoutConfig, spawnLimiter *rateLimiter}` with router:

```go
func (s *WebServer) NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(web.Static))
	mux.HandleFunc("GET /c/{ref}", s.HandlePage)
	mux.HandleFunc("POST /c/{ref}/pay", s.HandlePay)
	mux.HandleFunc("GET /c/{ref}/status", s.HandleStatus)
	mux.HandleFunc("GET /l/{ref}", s.HandleLink)
	return mux
}
```

Handler behaviour (implement exactly):
- `HandlePage`: load session via business (`GetSessionByRef`); not found → 404 plain; terminal completed → `done` page; expired → `gone`; processing → `confirm`; pending → `pay` with: methods from registry filtered by session restriction, `Preselect` fed by prefill clue + guest cookie hints, language pick (`?lang` → prefill language → `Accept-Language` first tag), masked contacts, CSRF token.
- `HandlePay`: parse form; verify CSRF else 403; map business sentinel errors → HTTP: `ErrSessionGone` → redirect to `/c/{ref}` (page shows terminal state), `ErrTooManyAttempts`/`ErrCooldown` → re-render `pay` with failure banner + 429, `ErrUnknownMethod`/`ErrAmountRequired` → 400 re-render with banner; success → set guest-hints cookie when guest, redirect 303 to `/c/{ref}` (renders `confirm`).
- `HandleStatus`: business `RefreshStatus`; respond `{"status":"processing|completed|failed","failure_reason":"..."}` JSON; 404 unknown ref.
- `HandleLink`: per-IP token bucket (`spawnLimiter.Allow(ip)` else 429); `SpawnSession`; unusable link → `gone` page 410; success → 303 redirect to `/c/{ref}`.

The rate limiter (same file, ~25 lines): `map[string]*bucket` + mutex, refill `cfg.LinkSpawnPerMinute` per minute, capacity = same.

- [ ] **Step 1: Failing tests** — httptest against `NewRouter()` with a real business wired to fake repos/clients (reuse the fakes from Task 7 by exporting them into a shared `business_test`-style helper or redeclaring locally):
1. `GET /c/{unknown}` → 404.
2. Pending recognized session renders: payer first name, masked phone, preselected method checked, CSRF hidden field present, raw msisdn absent from body.
3. Pending guest session renders phone input.
4. Expired session → gone page (no amount, no name in body).
5. `POST /pay` without CSRF → 403, with wrong CSRF → 403.
6. `POST /pay` happy path → 303 to `/c/{ref}`, session now processing.
7. `POST /pay` 4th attempt → 429.
8. `GET /status` for completed session → `{"status":"completed"}`.
9. `GET /l/{ref}` active link → 303 to a fresh `/c/...`; inactive link → 410; 11th hit in a minute from one IP → 429.

- [ ] **Step 2: Run — FAIL. Step 3: Implement `web.go`. Step 4: Run `go test ./apps/checkout/service/handlers/ -v` — PASS. Step 5: Commit**

```bash
git add apps/checkout/service/handlers
git commit -m "feat(checkout): public checkout page handlers"
```

---

### Task 10: Connect RPC handlers (TDD)

**Files:**
- Create: `apps/checkout/service/handlers/rpc.go`
- Test: `apps/checkout/service/handlers/rpc_test.go`

`CheckoutServer` implements `checkoutv1connect.CheckoutServiceHandler` (from `apps/checkout/gen/payment/checkout/v1/checkoutv1connect`), delegating to the business and converting model ↔ proto:

```go
type CheckoutServer struct {
	business *business.CheckoutBusiness
	cfg      *config.CheckoutConfig
}

func (s *CheckoutServer) CreateCheckoutSession(
	ctx context.Context,
	req *connect.Request[checkoutv1.CreateCheckoutSessionRequest],
) (*connect.Response[checkoutv1.CreateCheckoutSessionResponse], error) {
	session, err := s.business.CreateSession(ctx, business.CreateSessionInputFromProto(req.Msg))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&checkoutv1.CreateCheckoutSessionResponse{
		Data: toProtoSession(session, s.cfg.PublicBaseURL),
	}), nil
}
```

`toProtoSession` sets `PageUrl = publicBaseURL + "/c/" + session.Ref` and maps the status string → `checkoutv1.SessionStatus` enum. `GetCheckoutSession` maps `gorm.ErrRecordNotFound` → `connect.NewError(connect.CodeNotFound, err)` and refreshes status first (so merchant polling sees live truth). `CreateCheckoutLink` mirrors create-session; `PageUrl = base + "/l/" + ref`.

- [ ] **Step 1: Failing tests**: create→get round trip via the connect handlers mounted on httptest + `checkoutv1connect.NewCheckoutServiceClient`; invalid amount → CodeInvalidArgument; unknown ref → CodeNotFound; link create returns `/l/` URL.
- [ ] **Step 2: FAIL → Step 3 implement → Step 4 `go test ./apps/checkout/... -v` PASS → Step 5 commit**

```bash
git add apps/checkout/service/handlers/rpc.go apps/checkout/service/handlers/rpc_test.go
git commit -m "feat(checkout): merchant connect RPC surface"
```

---

### Task 11: main.go wiring + sweeper

**Files:**
- Create: `apps/checkout/cmd/main.go`

- [ ] **Step 1: Write main** — combine the ledger datastore/interceptor wiring with the integration client-setup pattern:

```go
func main() {
	ctx := context.Background()

	cfg, err := config.LoadWithOIDC[aconfig.CheckoutConfig](ctx)
	if err != nil { util.Log(ctx).With("err", err).Error("could not process configs"); return }
	if cfg.Name() == "" { cfg.ServiceName = "service_payment_checkout" }

	ctx, svc := frame.NewServiceWithContext(ctx,
		frame.WithConfig(&cfg), frame.WithDatastore())
	defer svc.Stop(ctx)
	logger := svc.Log(ctx)

	// migrate-and-exit branch — copy the exact arg handling from apps/ledger/cmd/main.go
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err = repository.Migrate(ctx, svc.DatastoreManager(), "./apps/checkout/migrations/0001"); err != nil {
			logger.WithError(err).Fatal("migration failed")
		}
		return
	}

	registry, err := business.ParseMethodRegistry(cfg.MethodsJSON)
	if err != nil { logger.WithError(err).Fatal("invalid CHECKOUT_METHODS") }
	if cfg.SigningSecret == "" { logger.Fatal("CHECKOUT_SIGNING_SECRET is required") }

	dbPool := svc.DatastoreManager().GetPool(ctx, datastore.DefaultPoolName)
	sessionRepo := repository.NewSessionRepository(ctx, dbPool, svc.WorkManager())
	linkRepo := repository.NewLinkRepository(ctx, dbPool, svc.WorkManager())

	paymentCli, err := setupPaymentClient(ctx, cfg)   // same connection.NewServiceClient pattern as integrations
	if err != nil { logger.WithError(err).Fatal("could not setup payment client") }
	profileCli, err := setupProfileClient(ctx, cfg)
	if err != nil { logger.WithError(err).Fatal("could not setup profile client") }

	checkoutBiz := business.NewCheckoutBusiness(&cfg, registry, sessionRepo, linkRepo, paymentCli, profileCli)

	// Connect RPC with the standard interceptor stack (tenancy + function access)
	rpcHandler := setupConnectServer(ctx, svc.SecurityManager(), checkoutBiz, &cfg)

	// Public page (no auth interceptors)
	webServer := handlers.NewWebServer(checkoutBiz, handlers.NewRenderer([]byte(cfg.SigningSecret)), registry, &cfg)

	mux := http.NewServeMux()
	mux.Handle("/", webServer.NewRouter())
	mux.Handle(checkoutv1connect.CheckoutServiceName+"/", rpcHandler) // connect procedures path-prefixed

	sd := checkoutv1.File_payment_checkout_v1_checkout_proto.Services().ByName("CheckoutService")
	svc.Init(ctx, frame.WithHTTPHandler(mux), frame.WithPermissionRegistration(sd))

	go runSweeper(ctx, &cfg, checkoutBiz)

	logger.Info("Initiating checkout server operations")
	if err = svc.Run(ctx, ""); err != nil { logger.WithError(err).Error("could not run Server") }
}

func runSweeper(ctx context.Context, cfg *aconfig.CheckoutConfig, biz *business.CheckoutBusiness) {
	ticker := time.NewTicker(time.Duration(cfg.SweepIntervalSeconds) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := biz.SweepProcessing(ctx); err != nil {
				util.Log(ctx).WithError(err).Warn("checkout sweep failed")
			}
		}
	}
}
```

`setupConnectServer` mirrors `apps/ledger/cmd/main.go:128-162` exactly, substituting `checkoutv1`/`checkoutv1connect` and `"CheckoutService"`. Note `checkoutv1connect.CheckoutServiceName` is generated as `/checkout.v1.CheckoutService` — mounting the connect handler under that prefix keeps `/c/`, `/l/`, `/static/` free for the page.

- [ ] **Step 2: Build + full test sweep**

Run: `go build ./... && go test -race ./apps/checkout/...` — expected PASS.

- [ ] **Step 3: Commit**

```bash
git add apps/checkout/cmd
git commit -m "feat(checkout): service wiring with sweeper and dual RPC/web mux"
```

---

### Task 12: Dockerfile + Makefile registration + gates

**Files:**
- Create: `apps/checkout/Dockerfile` (copy `apps/integrations/pawapay/Dockerfile`; replace both `apps/integrations/pawapay` path references with `apps/checkout`, set title label `"Payments Checkout"`; **add** `COPY ./apps/checkout/migrations ./apps/checkout/migrations` is unnecessary — migrations are embedded in the build context already via the app copy; verify the final image runs `./apps/checkout/cmd/`)
- Modify: `Makefile` — `APP_DIRS` gains `apps/checkout` (alphabetically after `apps/billing`)

- [ ] **Step 1: Make both changes.**
- [ ] **Step 2: Gates**

Run, expect all clean:
```bash
go build ./...
go test -race ./apps/checkout/...
golangci-lint fmt ./apps/checkout/... && golangci-lint run ./apps/checkout/...
```
Generated code under `apps/checkout/gen` will be flagged by some linters — if so, add to `.golangci.yml` issues exclusion: `- path: apps/checkout/gen/` (check how the repo handles other generated code first; prefer the existing exclusion mechanism).

- [ ] **Step 3: Commit**

```bash
git add apps/checkout/Dockerfile Makefile .golangci.yml
git commit -m "feat(checkout): containerize and register checkout app"
```

---

## Self-review checklist (run after Task 12)

1. **Spec coverage**: sessions ✓(T3,7,10) links ✓(T3,7,9,10) page states ✓(T8,9) preselection ✓(T5) clues+guest cookie ✓(T6,7) prompt extras with session/order refs ✓(T7) status polling + sweeper ✓(T7,9,11) security (CSRF, masking, rate limits, capability refs) ✓(T8,9) deployment → follow-up plan ✓(header note).
2. **Placeholder scan**: Task 7 fakes and Task 9 tests reference patterns proven in `apps/integrations/pawapay/service/handlers/webhooks_test.go` — implementer fills only mechanical bodies with explicit instructions to verify generated signatures first.
3. **Type consistency**: `PageData`/`Renderer` (T8) match usage in T9; `CheckoutBusiness` methods (T7) match T9/T10/T11 call sites; sentinel errors named identically across T7/T9.
