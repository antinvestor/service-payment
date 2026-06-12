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

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"connectrpc.com/connect"
	checkoutv1 "github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1"
	"github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1/checkoutv1connect"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/handlers"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GetById is implemented here so fakeProfileClient does not panic when
// CreateSession calls applyPayer for a non-empty ProfileID.
// The embedded interface auto-generated method would dereference a nil pointer.
//
//nolint:revive,staticcheck // GetById matches the generated interface name; renaming to GetByID would break compilation.
func (f *fakeProfileClient) GetById(
	_ context.Context,
	_ *connect.Request[profilev1.GetByIdRequest],
) (*connect.Response[profilev1.GetByIdResponse], error) {
	// Return an empty profile — sufficient for tests that just need non-panic behaviour.
	return connect.NewResponse(&profilev1.GetByIdResponse{}), nil
}

// ---------------------------------------------------------------------------
// RPC test harness
// ---------------------------------------------------------------------------

type rpcHarness struct {
	sessionRepo *fakeSessionRepo
	linkRepo    *fakeLinkRepo
	payCli      *fakePaymentClient
	biz         *business.CheckoutBusiness
	cli         checkoutv1connect.CheckoutServiceClient
	srv         *httptest.Server
}

func newRPCHarness(t *testing.T) *rpcHarness {
	t.Helper()

	sessionRepo := newFakeSessionRepo()
	linkRepo := newFakeLinkRepo()
	payCli := &fakePaymentClient{}

	cfg := testConfig()
	reg := testRegistry()

	biz := business.NewCheckoutBusiness(
		cfg,
		reg,
		sessionRepo,
		linkRepo,
		payCli,
		&fakeProfileClient{},
	)
	biz = biz.WithClock(fixedNow)

	rpcServer := handlers.NewCheckoutServer(biz, cfg)

	mux := http.NewServeMux()
	path, h := checkoutv1connect.NewCheckoutServiceHandler(rpcServer)
	mux.Handle(path, h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := checkoutv1connect.NewCheckoutServiceClient(srv.Client(), srv.URL)

	return &rpcHarness{
		sessionRepo: sessionRepo,
		linkRepo:    linkRepo,
		payCli:      payCli,
		biz:         biz,
		cli:         cli,
		srv:         srv,
	}
}

// ---------------------------------------------------------------------------
// Test 1: Create→Get round-trip
// ---------------------------------------------------------------------------

func TestRPC_CreateGet_RoundTrip(t *testing.T) {
	h := newRPCHarness(t)

	createReq := connect.NewRequest(&checkoutv1.CreateCheckoutSessionRequest{
		Name:        "Acme Store",
		Description: "Order payment",
		Amount: &commonv1.Money{
			CurrencyCode: "KES",
			Units:        150,
			Nanos:        0,
		},
		Metadata: map[string]string{"orderId": "O1"},
		Payer: &checkoutv1.PayerPrefill{
			ProfileId:   "profile-001",
			DisplayName: "Alice",
			Language:    "en",
		},
		ReturnUrl: "https://example.com/done",
	})
	createResp, err := h.cli.CreateCheckoutSession(t.Context(), createReq)
	require.NoError(t, err)

	data := createResp.Msg.GetData()
	require.NotNil(t, data)

	// ref must be 32 chars
	assert.Len(t, data.GetRef(), 32, "ref must be 32 characters")
	// PageUrl must be baseURL + /c/ + ref
	assert.Equal(t, testConfig().PublicBaseURL+"/c/"+data.GetRef(), data.GetPageUrl())
	// Status should be pending
	assert.Equal(t, checkoutv1.SessionStatus_SESSION_STATUS_PENDING_UNSPECIFIED, data.GetStatus())
	// Amount units = 150
	require.NotNil(t, data.GetAmount())
	assert.Equal(t, int64(150), data.GetAmount().GetUnits())
	assert.Equal(t, "KES", data.GetAmount().GetCurrencyCode())

	// Now Get the same session
	getResp, err := h.cli.GetCheckoutSession(t.Context(), connect.NewRequest(&checkoutv1.GetCheckoutSessionRequest{
		Ref: data.GetRef(),
	}))
	require.NoError(t, err)

	got := getResp.Msg.GetData()
	require.NotNil(t, got)
	assert.Equal(t, data.GetRef(), got.GetRef())
	// metadata should contain orderId
	assert.Equal(t, "O1", got.GetMetadata()["orderId"])
}

// ---------------------------------------------------------------------------
// Test 2: Internal metadata never leaks
// ---------------------------------------------------------------------------

func TestRPC_InternalMetadata_DoesNotLeak(t *testing.T) {
	h := newRPCHarness(t)

	// Create a session normally
	leakReq := connect.NewRequest(&checkoutv1.CreateCheckoutSessionRequest{
		Name: "Merchant",
		Amount: &commonv1.Money{
			CurrencyCode: "KES",
			Units:        100,
		},
		ReturnUrl: "https://example.com/done",
	})
	createResp, err := h.cli.CreateCheckoutSession(t.Context(), leakReq)
	require.NoError(t, err)
	ref := createResp.Msg.GetData().GetRef()

	// Mutate the stored session to add internal metadata (simulating Pay() writing _method).
	stored := h.sessionRepo.sessions[ref]
	if stored.Metadata == nil {
		stored.Metadata = make(map[string]any)
	}
	stored.Metadata["_method"] = "mpesa"
	stored.Metadata["_contact_id"] = "contact-1"
	stored.Metadata["publicKey"] = "visible"

	// Get and verify
	getResp, err := h.cli.GetCheckoutSession(t.Context(), connect.NewRequest(&checkoutv1.GetCheckoutSessionRequest{
		Ref: ref,
	}))
	require.NoError(t, err)

	meta := getResp.Msg.GetData().GetMetadata()
	assert.NotContains(t, meta, "_method", "_method must not be exposed")
	assert.NotContains(t, meta, "_contact_id", "_contact_id must not be exposed")
	assert.Equal(t, "visible", meta["publicKey"], "non-internal keys must be passed through")
}

// ---------------------------------------------------------------------------
// Test 3a: Zero amount with fixed option → CodeInvalidArgument
// ---------------------------------------------------------------------------

func TestRPC_CreateSession_ZeroAmount_InvalidArgument(t *testing.T) {
	h := newRPCHarness(t)

	_, err := h.cli.CreateCheckoutSession(t.Context(), connect.NewRequest(&checkoutv1.CreateCheckoutSessionRequest{
		Name: "Merchant",
		Amount: &commonv1.Money{
			CurrencyCode: "KES",
			Units:        0,
			Nanos:        0,
		},
		// default AmountOption is FIXED_UNSPECIFIED (fixed)
	}))
	require.Error(t, err)

	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

// ---------------------------------------------------------------------------
// Test 3b: javascript: ReturnURL → CodeInvalidArgument
// ---------------------------------------------------------------------------

func TestRPC_CreateSession_JavascriptReturnURL_InvalidArgument(t *testing.T) {
	h := newRPCHarness(t)

	_, err := h.cli.CreateCheckoutSession(t.Context(), connect.NewRequest(&checkoutv1.CreateCheckoutSessionRequest{
		Name: "Merchant",
		Amount: &commonv1.Money{
			CurrencyCode: "KES",
			Units:        100,
		},
		ReturnUrl: "javascript:alert(1)",
	}))
	require.Error(t, err)

	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

// ---------------------------------------------------------------------------
// Test 4: Get unknown ref → CodeNotFound
// ---------------------------------------------------------------------------

func TestRPC_GetSession_UnknownRef_NotFound(t *testing.T) {
	h := newRPCHarness(t)

	_, err := h.cli.GetCheckoutSession(t.Context(), connect.NewRequest(&checkoutv1.GetCheckoutSessionRequest{
		Ref: "no-such-ref",
	}))
	require.Error(t, err)

	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
}

// ---------------------------------------------------------------------------
// Test 5a: CreateCheckoutLink → PageUrl contains /l/
// ---------------------------------------------------------------------------

func TestRPC_CreateLink_ValidRequest(t *testing.T) {
	h := newRPCHarness(t)

	resp, err := h.cli.CreateCheckoutLink(t.Context(), connect.NewRequest(&checkoutv1.CreateCheckoutLinkRequest{
		Name: "Product Link",
		Amount: &commonv1.Money{
			CurrencyCode: "KES",
			Units:        500,
		},
	}))
	require.NoError(t, err)

	data := resp.Msg.GetData()
	require.NotNil(t, data)
	assert.Contains(t, data.GetPageUrl(), "/l/", "PageUrl must contain /l/")
	assert.True(t, strings.HasPrefix(data.GetPageUrl(), testConfig().PublicBaseURL+"/l/"))
	assert.True(t, data.GetActive())
}

// ---------------------------------------------------------------------------
// Test 5b: CreateCheckoutLink bad expires_at → CodeInvalidArgument
// ---------------------------------------------------------------------------

func TestRPC_CreateLink_BadExpiresAt_InvalidArgument(t *testing.T) {
	h := newRPCHarness(t)

	_, err := h.cli.CreateCheckoutLink(t.Context(), connect.NewRequest(&checkoutv1.CreateCheckoutLinkRequest{
		Name: "Product Link",
		Amount: &commonv1.Money{
			CurrencyCode: "KES",
			Units:        500,
		},
		ExpiresAt: "not-a-date",
	}))
	require.Error(t, err)

	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

// ---------------------------------------------------------------------------
// Test 6: Get on processing session with successful payment status → COMPLETED
// ---------------------------------------------------------------------------

func TestRPC_GetSession_ProcessingRefreshedToCompleted(t *testing.T) {
	h := newRPCHarness(t)

	// Seed a processing session with a PromptID directly in the fake repo.
	h.sessionRepo.sessions["proc-ref"] = &models.CheckoutSession{
		Ref:          "proc-ref",
		Name:         "Merchant",
		Amount:       "100",
		Currency:     "KES",
		AmountOption: models.AmountOptionFixed,
		Status:       models.SessionStatusProcessing,
		PromptID:     "prompt-xyz",
		ExpiresAt:    fixedNow().Add(30 * time.Minute),
	}

	// Fake payment client returns SUCCESSFUL for any status poll.
	h.payCli.statusResp = connect.NewResponse(&commonv1.StatusResponse{
		Status: commonv1.STATUS_SUCCESSFUL,
		Id:     "payment-abc",
	})

	getResp, err := h.cli.GetCheckoutSession(t.Context(), connect.NewRequest(&checkoutv1.GetCheckoutSessionRequest{
		Ref: "proc-ref",
	}))
	require.NoError(t, err)

	assert.Equal(t, checkoutv1.SessionStatus_SESSION_STATUS_COMPLETED, getResp.Msg.GetData().GetStatus(),
		"RefreshStatus must transition processing→completed when payment is SUCCESSFUL")
}
