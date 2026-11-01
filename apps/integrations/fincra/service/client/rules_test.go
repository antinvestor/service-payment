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
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/fincratest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookSignature(t *testing.T) {
	const secret = "6d1d475adef242e59a648335288ee80b"
	pretty := []byte("{\n  \"event\": \"payout.successful\",\n  \"data\": {\"id\": 40614, \"status\": \"successful\"}\n}")
	compact := []byte(`{"event":"payout.successful","data":{"id":40614,"status":"successful"}}`)
	withBusiness := []byte(`{"event":"collection.failed","data":{"status":"failed"},"business":"b1"}`)

	tests := []struct {
		name      string
		body      []byte
		signature string
		secret    string
		want      bool
	}{
		{"raw body", compact, client.WebhookSignature(secret, compact), secret, true},
		{"upper-case hex", compact, "  " + upper(client.WebhookSignature(secret, compact)), secret, true},
		{"re-serialised body (JSON.stringify)", pretty, client.WebhookSignature(secret, compact), secret, true},
		{"event and data only", withBusiness,
			client.WebhookSignature(secret, []byte(`{"event":"collection.failed","data":{"status":"failed"}}`)), secret, true},
		{"wrong secret", compact, client.WebhookSignature("other", compact), secret, false},
		{"tampered body", []byte(`{"event":"payout.successful","data":{"id":40615}}`), client.WebhookSignature(secret, compact), secret, false},
		{"no secret configured", compact, client.WebhookSignature("", compact), "", false},
		{"no signature", compact, "", secret, false},
		{"not hex", compact, "zz", secret, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, client.VerifyWebhookSignature(tt.body, tt.signature, tt.secret))
		})
	}
}

func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'f' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}

func TestCollectionStatus(t *testing.T) {
	tests := []struct {
		status, variance string
		want             commonv1.STATUS
	}{
		{"success", "", commonv1.STATUS_SUCCESSFUL},
		{"successful", "", commonv1.STATUS_SUCCESSFUL},
		{"success", "overpayment", commonv1.STATUS_SUCCESSFUL},
		{"success", "underpayment", commonv1.STATUS_FAILED},
		{"failed", "", commonv1.STATUS_FAILED},
		{"expired", "", commonv1.STATUS_FAILED},
		{"pending", "", commonv1.STATUS_IN_PROCESS},
		{"initiated", "", commonv1.STATUS_IN_PROCESS},
		{"", "", commonv1.STATUS_IN_PROCESS},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, client.CollectionStatus(tt.status, tt.variance), "%s/%s", tt.status, tt.variance)
	}
}

func TestPayoutStatus(t *testing.T) {
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, client.PayoutStatus("successful"))
	assert.Equal(t, commonv1.STATUS_FAILED, client.PayoutStatus("failed"))
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, client.PayoutStatus("processing"))
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, client.PayoutStatus("pending"))
	assert.Equal(t, commonv1.STATE_INACTIVE, client.StateFor(commonv1.STATUS_FAILED))
	assert.Equal(t, commonv1.STATE_ACTIVE, client.StateFor(commonv1.STATUS_SUCCESSFUL))
}

func TestChargePaidAmount(t *testing.T) {
	ch := &client.Charge{Amount: 101.51, AmountExpected: 100}
	assert.InDelta(t, 100, ch.PaidAmount(client.FeeBearerBusiness), 0)
	assert.Zero(t, ch.PaidAmount(client.FeeBearerCustomer), "customer-borne fees make Fincra's amount differ from what was asked")
	assert.InDelta(t, 50, (&client.Charge{Amount: 50}).PaidAmount(client.FeeBearerBusiness), 0)
}

func TestMatchProvider(t *testing.T) {
	kenya := []client.Provider{{Code: "MPESA", Name: "Safaricom M-Pesa"}, {Code: "AIRTEL", Name: "Airtel Money"}}
	ghana := []client.Provider{
		{Code: "AIRTEL_TIGO", Name: "AirtelTigo Money"}, {Code: "MTN", Name: "MTN Mobile Money"},
		{Code: "VODAFONE", Name: "Vodafone Cash"},
	}
	tests := []struct {
		name      string
		providers []client.Provider
		hint      string
		want      string
		found     bool
	}{
		{"code", kenya, "AIRTEL", "AIRTEL", true},
		{"m-pesa alias", kenya, "SAFARICOM", "MPESA", true},
		{"airteltigo by alias", ghana, "AIRTELTIGO", "AIRTEL_TIGO", true},
		{"airtel never resolves to airteltigo", ghana, "AIRTEL", "", false},
		{"telecel is vodafone", ghana, "TELECEL", "VODAFONE", true},
		{"name contains", []client.Provider{{Code: "X1", Name: "MTN MoMo Zambia"}}, "MTN", "X1", true},
		{"no hint, one provider", []client.Provider{{Code: "ONLY"}}, "", "ONLY", true},
		{"no hint, many providers", kenya, "", "", false},
		{"unknown", kenya, "ZAMTEL", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := client.MatchProvider(tt.providers, tt.hint)
			assert.Equal(t, tt.found, ok)
			assert.Equal(t, tt.want, p.Code)
		})
	}
}

func TestCatalogCachesProviders(t *testing.T) {
	srv := fincratest.New(t)
	srv.Providers["KES|mobile_money_wallet"] = []client.Provider{{Code: "MPESA"}}
	catalog := client.NewCatalog(client.NewClient(), time.Minute)
	ctx := context.Background()

	for range 3 {
		providers, err := catalog.Providers(ctx, srv.Credentials(), "KE", "KES", client.PaymentDestinationMobileMoney)
		require.NoError(t, err)
		require.Len(t, providers, 1)
	}
	assert.Len(t, srv.Calls, 1, "providers are fetched once per TTL")
}

func TestIsRejected(t *testing.T) {
	rejected := []int{400, 401, 403, 404, 409, 422}
	unknown := []int{408, 425, 429, 500, 502, 503, 504}
	for _, code := range rejected {
		assert.True(t, client.IsRejected(&client.APIError{HTTPStatus: code}), "%d is a definitive refusal", code)
	}
	for _, code := range unknown {
		assert.False(t, client.IsRejected(&client.APIError{HTTPStatus: code}), "%d leaves the outcome unknown", code)
	}
	assert.False(t, client.IsRejected(context.DeadlineExceeded), "a transport error leaves the outcome unknown")
}
