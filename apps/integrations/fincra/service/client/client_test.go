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

package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/fincratest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckoutRoundTrip(t *testing.T) {
	srv := fincratest.New(t)
	cli := client.NewClient()
	ctx := context.Background()

	link, err := cli.CreateCheckout(ctx, srv.Credentials(), &client.CheckoutRequest{
		Amount: 350, Currency: "NGN", Reference: "prompt-1", FeeBearer: client.FeeBearerBusiness,
		Customer: client.Customer{Name: "Ada", Email: "ada@example.com"},
		Metadata: map[string]string{"tenant_id": "t1"},
	})
	require.NoError(t, err)
	assert.Contains(t, link.Link, "/pay/"+link.PayCode)

	cp, err := cli.GetCheckout(ctx, srv.Credentials(), "prompt-1")
	require.NoError(t, err)
	assert.Equal(t, "prompt-1", cp.MerchantReference)
	assert.Equal(t, link.PayCode, cp.Reference)
	assert.InDelta(t, 350, cp.Amount.Float(), 0)
	assert.Equal(t, "t1", cp.Metadata["tenant_id"])

	_, err = cli.CreateCheckout(ctx, srv.Credentials(), &client.CheckoutRequest{Amount: 1, Currency: "NGN", Reference: "prompt-1"})
	require.Error(t, err)
	assert.True(t, client.IsDuplicate(err), "a reused reference is a duplicate: %v", err)
}

func TestLookupNotFound(t *testing.T) {
	srv := fincratest.New(t)
	_, err := client.NewClient().GetCharge(context.Background(), srv.Credentials(), "missing")
	require.Error(t, err)
	assert.True(t, client.IsNotFound(err))
	assert.False(t, client.IsDuplicate(err))
}

func TestWrongSecretKeyIsRejected(t *testing.T) {
	srv := fincratest.New(t)
	creds := srv.Credentials()
	creds.SecretKey = "wrong"
	_, err := client.NewClient().GetPayout(context.Background(), creds, "p-1")
	var apiErr *client.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.HTTPStatus)
	assert.Equal(t, "Invalid authentication credentials", apiErr.Message)
}

func TestReadsAreRetriedOnServerErrors(t *testing.T) {
	srv := fincratest.New(t)
	srv.PutPayout(&client.Payout{CustomerReference: "p-1", Status: client.StatusSuccessful})
	srv.FailNext["/disbursements/payouts/customer-reference/"] = []int{http.StatusBadGateway, http.StatusServiceUnavailable}

	p, err := client.NewClient().GetPayout(context.Background(), srv.Credentials(), "p-1")
	require.NoError(t, err)
	assert.Equal(t, client.StatusSuccessful, p.Status)
}

func TestWritesAreNotRetried(t *testing.T) {
	srv := fincratest.New(t)
	srv.FailNext["POST /disbursements/payouts"] = []int{http.StatusInternalServerError}

	_, err := client.NewClient().CreatePayout(context.Background(), srv.Credentials(), &client.PayoutRequest{
		Business: fincratest.BusinessID, CustomerReference: "p-1", Amount: "10",
	})
	require.Error(t, err)
	assert.Nil(t, srv.Payout("p-1"), "a failed payout POST must not be sent again")
	assert.Equal(t, []string{"POST /disbursements/payouts"}, srv.Calls)
}

func TestFalseEnvelopeIsAnError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":false,"message":"Charge could not be created"}`))
	}))
	defer ts.Close()

	_, err := client.NewClient().InitiateCharge(context.Background(), &client.Credentials{BaseURL: ts.URL}, &client.ChargeRequest{})
	var apiErr *client.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "Charge could not be created", apiErr.Message)
}

func TestErrorShapes(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantType string
		wantMsg  string
		dup      bool
		notFound bool
	}{
		{
			name: "ip not whitelisted", status: http.StatusForbidden,
			body:     `{"success":false,"error":"Your IP address is not allowed to access this service","errorType":"ACCESS_DENIED"}`,
			wantType: "ACCESS_DENIED", wantMsg: "Your IP address is not allowed to access this service",
		},
		{
			name: "duplicate customer reference", status: http.StatusUnprocessableEntity,
			body:     `{"success":false,"error":"Cannot continue, Duplicate Customer Reference Passed","errorType":"DUPLICATE_CUSTOMER_REFERENCE"}`,
			wantType: "DUPLICATE_CUSTOMER_REFERENCE", wantMsg: "Cannot continue, Duplicate Customer Reference Passed", dup: true,
		},
		{
			name: "payout not found", status: http.StatusNotFound,
			body:     `{"success":false,"error":"Payout not found","errorType":"RESOURCE_NOT_FOUND"}`,
			wantType: "RESOURCE_NOT_FOUND", wantMsg: "Payout not found", notFound: true,
		},
		{
			name: "validation messages", status: http.StatusBadRequest,
			body:     `{"message":["amount must be a number","currency is required"]}`,
			wantType: "BAD_REQUEST", wantMsg: "amount must be a number; currency is required",
		},
		{
			name: "plain text", status: http.StatusBadGateway, body: `upstream down`,
			wantType: "BAD_GATEWAY", wantMsg: "upstream down",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer ts.Close()

			_, err := client.NewClient().CreatePayout(context.Background(), &client.Credentials{BaseURL: ts.URL}, &client.PayoutRequest{})
			var apiErr *client.APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tt.wantType, apiErr.Code())
			assert.Equal(t, tt.wantMsg, apiErr.Message)
			assert.Equal(t, tt.dup, client.IsDuplicate(err))
			assert.Equal(t, tt.notFound, client.IsNotFound(err))
		})
	}
}

func TestRequestHeaders(t *testing.T) {
	var got http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer ts.Close()

	creds := &client.Credentials{BaseURL: ts.URL, SecretKey: "sk", PublicKey: "pk", BusinessID: "biz"}
	_, err := client.NewClient().ListProviders(context.Background(), creds, "KE", "KES", client.PaymentDestinationMobileMoney)
	require.NoError(t, err)
	assert.Equal(t, "sk", got.Get("api-key"))
	assert.Equal(t, "pk", got.Get("x-pub-key"))
	assert.Equal(t, "biz", got.Get("x-business-id"))
}

func TestFlexibleDecoding(t *testing.T) {
	var v struct {
		ID     client.FlexString `json:"id"`
		Ref    client.FlexString `json:"ref"`
		Null   client.FlexString `json:"null"`
		Amount client.FlexFloat  `json:"amount"`
		Text   client.FlexFloat  `json:"text"`
		Junk   client.FlexFloat  `json:"junk"`
		None   client.FlexFloat  `json:"none"`
	}
	raw := `{"id":17740,"ref":"abc","null":null,"amount":101.51,"text":"5914.000000","junk":"n/a","none":null}`
	require.NoError(t, json.Unmarshal([]byte(raw), &v))
	assert.Equal(t, "17740", v.ID.String())
	assert.Equal(t, "abc", v.Ref.String())
	assert.Empty(t, v.Null.String())
	assert.InDelta(t, 101.51, v.Amount.Float(), 1e-9)
	assert.InDelta(t, 5914, v.Text.Float(), 1e-9)
	assert.Zero(t, v.Junk.Float())
	assert.Zero(t, v.None.Float())
}

func TestCredentials(t *testing.T) {
	sandbox := &client.Credentials{Environment: "sandbox"}
	live := &client.Credentials{Environment: "production"}
	assert.Equal(t, client.SandboxBaseURL, sandbox.ResolveBaseURL())
	assert.Equal(t, client.ProductionBaseURL, live.ResolveBaseURL())
	assert.Equal(t, "https://sandbox-checkout.fincra.com/pay/fcr-p-1", sandbox.CheckoutLink("fcr-p-1"))
	assert.Equal(t, "https://checkout.fincra.com/pay/fcr-p-1", live.CheckoutLink("fcr-p-1"))
	assert.Equal(t, client.CollectionModeCheckout, sandbox.ResolveCollectionMode())
	assert.Equal(t, client.CollectionModeDirect, (&client.Credentials{CollectionMode: "DIRECT"}).ResolveCollectionMode())
	assert.Equal(t, client.FeeBearerBusiness, sandbox.ResolveFeeBearer())
	assert.Equal(t, client.FeeBearerCustomer, (&client.Credentials{FeeBearer: "customer"}).ResolveFeeBearer())
}

func TestIsDuplicateIgnoresOtherErrors(t *testing.T) {
	assert.False(t, client.IsDuplicate(errors.New("duplicate")), "only Fincra API errors count")
	assert.False(t, client.IsNotFound(nil))
}
