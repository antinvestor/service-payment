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
	"net/http"
	"net/http/httptest"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	checkoutv1 "github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPC_ListPaymentMethods(t *testing.T) {
	h := newRPCHarness(t)

	resp, err := h.cli.ListPaymentMethods(t.Context(), &checkoutv1.ListPaymentMethodsRequest{Currency: "KES"})
	require.NoError(t, err)
	require.Len(t, resp.GetData(), 1, "only the KES rail is offered")
	m := resp.GetData()[0]
	assert.Equal(t, "mpesa", m.GetKey())
	assert.Equal(t, "mpesa", m.GetRoute())
	assert.Equal(t, "mobile_money", m.GetMethodType())
	assert.Equal(t, "push", m.GetFlow())
	assert.True(t, m.GetWholeAmounts())
	assert.Equal(t, "mpesa", resp.GetSelected())

	resp, err = h.cli.ListPaymentMethods(t.Context(), &checkoutv1.ListPaymentMethodsRequest{Phones: []string{"256700000000"}})
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetData())
	assert.Equal(t, "mtn_momo", resp.GetSelected(), "a Ugandan number preselects MTN")

	_, err = h.cli.ListPaymentMethods(t.Context(), &checkoutv1.ListPaymentMethodsRequest{Methods: []string{"nope"}})
	require.Error(t, err)
}

func TestRPC_CreateSession_CancelURLRoundTrip(t *testing.T) {
	h := newRPCHarness(t)
	resp, err := h.cli.CreateCheckoutSession(t.Context(), &checkoutv1.CreateCheckoutSessionRequest{
		Name:      "Acme Store",
		Amount:    &commonv1.Money{CurrencyCode: "KES", Units: 150},
		ReturnUrl: "https://example.com/done",
		CancelUrl: "https://example.com/cancel",
		Payer:     &checkoutv1.PayerPrefill{DisplayName: "Alice", Email: "alice@example.com"},
	})
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/cancel", resp.GetData().GetCancelUrl())
	assert.Equal(t, "alice@example.com", resp.GetData().GetPayer().GetEmail())

	got, err := h.cli.GetCheckoutSession(t.Context(), &checkoutv1.GetCheckoutSessionRequest{Ref: resp.GetData().GetRef()})
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/cancel", got.GetData().GetCancelUrl())
}

func TestHandlePage_CancelLinkAndDoneLink(t *testing.T) {
	h := newHarness(t)
	s := h.addPendingSession("sess-cancel-00000000000000000000")
	s.ReturnURL = "https://app.example.com/return"
	s.CancelURL = "https://app.example.com/cancel"

	req := httptest.NewRequest(http.MethodGet, "/c/"+s.Ref+"?cancelled=1", nil)
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "https://app.example.com/cancel?session="+s.Ref+"&amp;status=cancelled")
	assert.Contains(t, body, "Payment cancelled")

	s.Status = "completed"
	req = httptest.NewRequest(http.MethodGet, "/c/"+s.Ref, nil)
	rec = httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `href="https://app.example.com/return?session=`+s.Ref+`&amp;status=completed"`,
		"the success page shows its return link, not only the meta refresh")
}
