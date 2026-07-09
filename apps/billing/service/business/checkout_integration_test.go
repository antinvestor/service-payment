// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package business_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	billingTests "github.com/antinvestor/service-payments/apps/billing/tests"
	checkoutConfig "github.com/antinvestor/service-payments/apps/checkout/config"
	"github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1/checkoutv1connect"
	checkoutBusiness "github.com/antinvestor/service-payments/apps/checkout/service/business"
	checkoutHandlers "github.com/antinvestor/service-payments/apps/checkout/service/handlers"
	checkoutModels "github.com/antinvestor/service-payments/apps/checkout/service/models"
	checkoutRepo "github.com/antinvestor/service-payments/apps/checkout/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Fake checkout-side repos (re-declared — they're package-local in checkout tests)
// ---------------------------------------------------------------------------

type fakeCheckoutSessionRepo struct {
	checkoutRepo.SessionRepository
	sessions  map[string]*checkoutModels.CheckoutSession
	createErr error
	updateErr error
}

func newFakeCheckoutSessionRepo() *fakeCheckoutSessionRepo {
	return &fakeCheckoutSessionRepo{
		sessions: make(map[string]*checkoutModels.CheckoutSession),
	}
}

func (f *fakeCheckoutSessionRepo) Create(_ context.Context, s *checkoutModels.CheckoutSession) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.sessions[s.Ref] = s
	return nil
}

func (f *fakeCheckoutSessionRepo) Update(
	_ context.Context,
	s *checkoutModels.CheckoutSession,
	_ ...string,
) (int64, error) {
	if f.updateErr != nil {
		return 0, f.updateErr
	}
	f.sessions[s.Ref] = s
	return 1, nil
}

func (f *fakeCheckoutSessionRepo) GetByRef(_ context.Context, ref string) (*checkoutModels.CheckoutSession, error) {
	s, ok := f.sessions[ref]
	if !ok {
		return nil, fmt.Errorf("session %q: %w", ref, gorm.ErrRecordNotFound)
	}
	return s, nil
}

func (f *fakeCheckoutSessionRepo) ListByStatus(
	_ context.Context,
	status string,
	limit int,
) ([]*checkoutModels.CheckoutSession, error) {
	var list []*checkoutModels.CheckoutSession
	for _, s := range f.sessions {
		if s.Status == status {
			list = append(list, s)
		}
		if len(list) >= limit {
			break
		}
	}
	return list, nil
}

// ---------------------------------------------------------------------------

type fakeCheckoutLinkRepo struct {
	checkoutRepo.LinkRepository
	links map[string]*checkoutModels.CheckoutLink
}

func newFakeCheckoutLinkRepo() *fakeCheckoutLinkRepo {
	return &fakeCheckoutLinkRepo{links: make(map[string]*checkoutModels.CheckoutLink)}
}

func (f *fakeCheckoutLinkRepo) Create(_ context.Context, l *checkoutModels.CheckoutLink) error {
	f.links[l.Ref] = l
	return nil
}

func (f *fakeCheckoutLinkRepo) Update(
	_ context.Context,
	l *checkoutModels.CheckoutLink,
	_ ...string,
) (int64, error) {
	f.links[l.Ref] = l
	return 1, nil
}

func (f *fakeCheckoutLinkRepo) GetByRef(_ context.Context, ref string) (*checkoutModels.CheckoutLink, error) {
	l, ok := f.links[ref]
	if !ok {
		return nil, fmt.Errorf("link %q: %w", ref, gorm.ErrRecordNotFound)
	}
	return l, nil
}

// ---------------------------------------------------------------------------

type fakeCheckoutPaymentClient struct {
	paymentv1connect.PaymentServiceClient
	promptResp *connect.Response[paymentv1.InitiatePromptResponse]
	statusResp *connect.Response[commonv1.StatusResponse]
}

func (f *fakeCheckoutPaymentClient) InitiatePrompt(
	_ context.Context,
	_ *connect.Request[paymentv1.InitiatePromptRequest],
) (*connect.Response[paymentv1.InitiatePromptResponse], error) {
	if f.promptResp != nil {
		return f.promptResp, nil
	}
	return connect.NewResponse(&paymentv1.InitiatePromptResponse{
		Data: &commonv1.StatusResponse{Id: "prompt-test-id"},
	}), nil
}

func (f *fakeCheckoutPaymentClient) Status(
	_ context.Context,
	_ *connect.Request[commonv1.StatusRequest],
) (*connect.Response[commonv1.StatusResponse], error) {
	if f.statusResp != nil {
		return f.statusResp, nil
	}
	return connect.NewResponse(&commonv1.StatusResponse{Status: commonv1.STATUS_IN_PROCESS}), nil
}

// ---------------------------------------------------------------------------

type fakeCheckoutProfileClient struct {
	profilev1connect.ProfileServiceClient
}

//nolint:revive,staticcheck // GetById matches generated interface name.
func (f *fakeCheckoutProfileClient) GetById(
	_ context.Context,
	_ *connect.Request[profilev1.GetByIdRequest],
) (*connect.Response[profilev1.GetByIdResponse], error) {
	return connect.NewResponse(&profilev1.GetByIdResponse{}), nil
}

func (f *fakeCheckoutProfileClient) Update(
	_ context.Context,
	_ *connect.Request[profilev1.UpdateRequest],
) (*connect.Response[profilev1.UpdateResponse], error) {
	return connect.NewResponse(&profilev1.UpdateResponse{}), nil
}

// ---------------------------------------------------------------------------
// checkoutTestHarness: real checkout service on an httptest server
// ---------------------------------------------------------------------------

type checkoutTestHarness struct {
	sessionRepo *fakeCheckoutSessionRepo
	server      *httptest.Server
	client      checkoutv1connect.CheckoutServiceClient
}

func newCheckoutTestHarness(t *testing.T) *checkoutTestHarness {
	t.Helper()

	sessionRepo := newFakeCheckoutSessionRepo()
	linkRepo := newFakeCheckoutLinkRepo()
	payCli := &fakeCheckoutPaymentClient{}
	profCli := &fakeCheckoutProfileClient{}

	cfg := &checkoutConfig.CheckoutConfig{
		SessionTTLMinutes:      30,
		MaxAttempts:            3,
		AttemptCooldownSeconds: 20,
		MethodsJSON:            `[{"key":"mpesa","name":"M-PESA","route":"mpesa","prefixes":["254"],"currencies":["KES"]}]`,
		PublicBaseURL:          "http://localhost:8080",
	}

	reg, err := checkoutBusiness.ParseMethodRegistry(cfg.MethodsJSON)
	require.NoError(t, err)

	biz := checkoutBusiness.NewCheckoutBusiness(cfg, reg, sessionRepo, linkRepo, payCli, profCli, nil)

	rpcServer := checkoutHandlers.NewCheckoutServer(biz, cfg)
	mux := http.NewServeMux()
	path, h := checkoutv1connect.NewCheckoutServiceHandler(rpcServer)
	mux.Handle(path, h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := checkoutv1connect.NewCheckoutServiceClient(srv.Client(), srv.URL)

	return &checkoutTestHarness{
		sessionRepo: sessionRepo,
		server:      srv,
		client:      cli,
	}
}

// ---------------------------------------------------------------------------
// CheckoutIntegrationSuite
// ---------------------------------------------------------------------------

type CheckoutIntegrationSuite struct {
	billingTests.BaseTestSuite
}

func TestCheckoutIntegrationSuite(t *testing.T) {
	suite.Run(t, new(CheckoutIntegrationSuite))
}

// makeIssuedInvoice creates and issues an invoice via the real billing engine.
// All test invoices use "KES" as currency.
//
//nolint:revive // t first is idiomatic for test helpers; ctx follows
func makeIssuedInvoice(
	t *testing.T,
	ctx context.Context,
	resources *billingTests.ServiceResources,
	profileID string,
	amount string,
) *models.Invoice {
	t.Helper()

	const currency = "KES"

	invoiceEng := resources.InvoiceEngine
	billingRunRepo := resources.BillingRunRepo

	// Create a billing run to associate the invoice with.
	runID := util.IDString()
	now := time.Now()
	run := &models.BillingRun{
		BaseModel:        data.BaseModel{ID: runID},
		SubscriptionID:   util.IDString(),
		ProfileID:        profileID,
		CatalogVersionID: util.IDString(),
		State:            models.BillingRunStatePending,
		PeriodStart:      now.AddDate(0, -1, 0),
		PeriodEnd:        now,
		Idempotency:      runID,
	}
	err := billingRunRepo.Create(ctx, run)
	require.NoError(t, err)

	amtDecimal, err := decimalx.NewFromString(amount)
	require.NoError(t, err)

	rl := &models.RatedLine{
		BaseModel:      data.BaseModel{ID: util.IDString()},
		BillingRunID:   runID,
		SubscriptionID: run.SubscriptionID,
		ComponentID:    util.IDString(),
		Description:    "test line",
		Amount:         amtDecimal.Ptr(),
		Currency:       currency,
	}
	rl.GenID(ctx)

	invoice, err := invoiceEng.GenerateInvoice(ctx, run, []*models.RatedLine{rl}, nil, decimalx.Zero())
	require.NoError(t, err)

	invoice, err = invoiceEng.IssueInvoice(ctx, invoice.GetID())
	require.NoError(t, err)
	require.Equal(t, models.InvoiceStateIssued, invoice.State)

	return invoice
}

// ---------------------------------------------------------------------------
// Test cases
// ---------------------------------------------------------------------------

// 1. CreateInvoiceCheckout: session created, order_ref == invoice id, money + metadata + payer correct.
func (ts *CheckoutIntegrationSuite) TestCreateInvoiceCheckout_SessionCreated() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "100.00")

		integ := business.NewCheckoutIntegration(ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "")
		result, err := integ.CreateInvoiceCheckout(ctx, invoice.GetID())

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.SessionRef)
		assert.Contains(t, result.PageURL, "/c/", "page URL should contain /c/")

		// Verify the stored session has correct order_ref, money, metadata, and payer.
		stored := ch.sessionRepo.sessions[result.SessionRef]
		require.NotNil(t, stored, "session must be stored in fake repo")

		assert.Equal(t, invoice.GetID(), stored.OrderRef, "order_ref must be invoice ID")
		assert.Equal(t, "100", stored.Amount, "amount must match invoice total")
		assert.Equal(t, "KES", stored.Currency)

		if invoiceMeta, ok := stored.Metadata["invoiceId"]; ok {
			assert.Equal(t, invoice.GetID(), invoiceMeta)
		}
		if invNum, ok := stored.Metadata["invoiceNumber"]; ok {
			assert.Equal(t, invoice.InvoiceNumber, invNum)
		}

		assert.Equal(t, profileID, stored.PayerProfileID, "payer profile id should be set")
	})
}

// 2. CreateInvoiceCheckout on DRAFT invoice → not-payable error.
func (ts *CheckoutIntegrationSuite) TestCreateInvoiceCheckout_DraftInvoice_NotPayable() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		// Create a DRAFT invoice (not issued).
		billingRunRepo := resources.BillingRunRepo
		invoiceEng := resources.InvoiceEngine

		runID := util.IDString()
		now := time.Now()
		run := &models.BillingRun{
			BaseModel:        data.BaseModel{ID: runID},
			SubscriptionID:   util.IDString(),
			ProfileID:        profileID,
			CatalogVersionID: util.IDString(),
			State:            models.BillingRunStatePending,
			PeriodStart:      now.AddDate(0, -1, 0),
			PeriodEnd:        now,
			Idempotency:      runID,
		}
		err := billingRunRepo.Create(ctx, run)
		require.NoError(t, err)

		amt := decimalx.New(5000, -2) // 50.00
		rl := &models.RatedLine{
			BaseModel:      data.BaseModel{ID: util.IDString()},
			BillingRunID:   runID,
			SubscriptionID: run.SubscriptionID,
			ComponentID:    util.IDString(),
			Description:    "test",
			Amount:         amt.Ptr(),
			Currency:       "KES",
		}
		rl.GenID(ctx)

		invoice, err := invoiceEng.GenerateInvoice(ctx, run, []*models.RatedLine{rl}, nil, decimalx.Zero())
		require.NoError(t, err)
		require.Equal(t, models.InvoiceStateDraft, invoice.State)

		integ := business.NewCheckoutIntegration(ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "")
		_, checkoutErr := integ.CreateInvoiceCheckout(ctx, invoice.GetID())

		require.Error(t, checkoutErr)
		require.ErrorIs(t, checkoutErr, apperrors.ErrInvoiceNotPayable,
			"draft invoice must return not-payable error")
		assert.Empty(t, ch.sessionRepo.sessions, "no session must be created")
	})
}

// 3. Full settle: create session, drive to completed, SettleFromCheckout → invoice PAID.
func (ts *CheckoutIntegrationSuite) TestSettleFromCheckout_FullFlow() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "75.00")

		integ := business.NewCheckoutIntegration(ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "")
		checkout, err := integ.CreateInvoiceCheckout(ctx, invoice.GetID())
		require.NoError(t, err)

		// Drive session to completed: mutate the stored session status directly.
		stored := ch.sessionRepo.sessions[checkout.SessionRef]
		require.NotNil(t, stored)
		stored.Status = checkoutModels.SessionStatusCompleted

		// Now settle.
		paid, err := integ.SettleFromCheckout(ctx, checkout.SessionRef)

		require.NoError(t, err)
		require.NotNil(t, paid)
		assert.Equal(t, models.InvoiceStatePaid, paid.State, "invoice must be PAID")
		assert.NotNil(t, paid.PaidAt, "PaidAt must be set")
	})
}

// 4. Idempotency: SettleFromCheckout twice → second call returns nil error, invoice still PAID.
func (ts *CheckoutIntegrationSuite) TestSettleFromCheckout_Idempotent() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "50.00")

		integ := business.NewCheckoutIntegration(ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "")
		checkout, err := integ.CreateInvoiceCheckout(ctx, invoice.GetID())
		require.NoError(t, err)

		stored := ch.sessionRepo.sessions[checkout.SessionRef]
		require.NotNil(t, stored)
		stored.Status = checkoutModels.SessionStatusCompleted

		// First call — should pay.
		paid1, err := integ.SettleFromCheckout(ctx, checkout.SessionRef)
		require.NoError(t, err)
		require.Equal(t, models.InvoiceStatePaid, paid1.State)

		// Second call — must NOT error and must return the already-paid invoice.
		paid2, err := integ.SettleFromCheckout(ctx, checkout.SessionRef)
		require.NoError(t, err, "second SettleFromCheckout must not error (idempotent)")
		require.NotNil(t, paid2)
		assert.Equal(t, models.InvoiceStatePaid, paid2.State)
	})
}

// 5. Pending session → ErrCheckoutNotCompleted, invoice still ISSUED.
func (ts *CheckoutIntegrationSuite) TestSettleFromCheckout_PendingSession_NotCompleted() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "30.00")

		integ := business.NewCheckoutIntegration(ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "")
		checkout, err := integ.CreateInvoiceCheckout(ctx, invoice.GetID())
		require.NoError(t, err)

		// Leave the session as pending (do NOT drive to completed).
		_, settleErr := integ.SettleFromCheckout(ctx, checkout.SessionRef)

		require.Error(t, settleErr)
		require.ErrorIs(t, settleErr, business.ErrCheckoutNotCompleted)

		// Invoice must still be ISSUED.
		still, err := resources.InvoiceEngine.GetInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.InvoiceStateIssued, still.State)
	})
}

// 6. Amount mismatch: complete session with different amount → "mismatch" error, invoice still ISSUED.
func (ts *CheckoutIntegrationSuite) TestSettleFromCheckout_AmountMismatch() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)

		ch := newCheckoutTestHarness(t)

		profileID := util.IDString()
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), profileID)

		invoice := makeIssuedInvoice(t, ctx, resources, profileID, "200.00")

		integ := business.NewCheckoutIntegration(ch.client, resources.InvoiceRepo, resources.InvoiceEngine, "")
		checkout, err := integ.CreateInvoiceCheckout(ctx, invoice.GetID())
		require.NoError(t, err)

		// Mutate the stored session to have a different amount AND mark as completed.
		stored := ch.sessionRepo.sessions[checkout.SessionRef]
		require.NotNil(t, stored)
		stored.Amount = "999" // tampered — differs from invoice 200.00
		stored.Status = checkoutModels.SessionStatusCompleted

		_, settleErr := integ.SettleFromCheckout(ctx, checkout.SessionRef)

		require.Error(t, settleErr)
		assert.Contains(t, settleErr.Error(), "mismatch",
			"error must mention 'mismatch' when amounts differ")

		still, err := resources.InvoiceEngine.GetInvoice(ctx, invoice.GetID())
		require.NoError(t, err)
		assert.Equal(t, models.InvoiceStateIssued, still.State, "invoice must remain ISSUED on mismatch")
	})
}

// ---------------------------------------------------------------------------
// 7. moneyFromDecimal unit tests (table-driven)
// ---------------------------------------------------------------------------

func TestMoneyFromDecimal(t *testing.T) {
	tests := []struct {
		name      string
		amount    *string // nil = pass nil pointer
		currency  string
		wantErr   bool
		wantMsg   string
		wantUnits int64
		wantNanos int32
	}{
		{
			name:    "nil decimal",
			amount:  nil,
			wantErr: true,
			wantMsg: "nil",
		},
		{
			name:     "zero amount",
			amount:   strPtr("0"),
			currency: "KES",
			wantErr:  true,
			wantMsg:  "positive",
		},
		{
			name:     "negative amount",
			amount:   strPtr("-10.00"),
			currency: "KES",
			wantErr:  true,
			wantMsg:  "positive",
		},
		{
			name:     "too many decimal places",
			amount:   strPtr("10.123"),
			currency: "KES",
			wantErr:  true,
			wantMsg:  "decimal",
		},
		{
			name:     "empty currency",
			amount:   strPtr("10.00"),
			currency: "",
			wantErr:  true,
			wantMsg:  "currency",
		},
		{
			name:      "valid integer amount",
			amount:    strPtr("100"),
			currency:  "KES",
			wantErr:   false,
			wantUnits: 100,
			wantNanos: 0,
		},
		{
			name:      "valid 2dp amount",
			amount:    strPtr("50.75"),
			currency:  "USD",
			wantErr:   false,
			wantUnits: 50,
			wantNanos: 750_000_000,
		},
		{
			name:      "valid 1dp amount",
			amount:    strPtr("25.5"),
			currency:  "KES",
			wantErr:   false,
			wantUnits: 25,
			wantNanos: 500_000_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dp *decimalx.Decimal
			if tt.amount != nil {
				d, err := decimalx.NewFromString(*tt.amount)
				require.NoError(t, err, "test input must parse as decimal")
				dp = &d
			}

			// Call the exported wrapper for testing — since moneyFromDecimal is unexported,
			// we test it indirectly via NewCheckoutIntegration + CreateInvoiceCheckout,
			// but we can also expose it for direct testing via a test helper function.
			// Here we use the public helper MoneyFromDecimalForTest.
			money, err := business.MoneyFromDecimalForTest(dp, tt.currency)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantMsg != "" {
					assert.Contains(t, err.Error(), tt.wantMsg)
				}
				assert.Nil(t, money)
			} else {
				require.NoError(t, err)
				require.NotNil(t, money)
				assert.Equal(t, tt.currency, money.GetCurrencyCode())
				assert.Equal(t, tt.wantUnits, money.GetUnits())
				assert.Equal(t, tt.wantNanos, money.GetNanos())
			}
		})
	}
}

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// Verify BillingWorkflow delegates correctly when checkout is nil
// ---------------------------------------------------------------------------

func (ts *CheckoutIntegrationSuite) TestBillingWorkflow_CollectInvoice_NilCheckout() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), util.IDString())

		// BillingWorkflow is constructed with nil checkout in base_testsuite.
		workflow := resources.BillingWorkflow
		_, err := workflow.CollectInvoice(ctx, "any-invoice-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checkout integration not configured")
	})
}

// Verify SettleInvoiceFromCheckout returns error when checkout is nil.
func (ts *CheckoutIntegrationSuite) TestBillingWorkflow_SettleInvoiceFromCheckout_NilCheckout() {
	ts.WithTestDependencies(ts.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := ts.CreateService(t, dep)
		ctx = ts.WithAuthClaims(ctx, util.IDString(), util.IDString(), util.IDString())

		workflow := resources.BillingWorkflow
		_, err := workflow.SettleInvoiceFromCheckout(ctx, "any-session-ref")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checkout integration not configured")
	})
}

// Ensure InvoiceRepo is exposed in ServiceResources for the test.
// (If it's not already there, this compilation error tells us to add it.)
var _ = (*billingTests.ServiceResources)(nil)
