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

package queue

import (
	"context"
	"net/http"
	"sync"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/fwtest"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// statuses records the status updates a worker emits.
type statuses struct {
	frameEvents.Manager

	mu      sync.Mutex
	updates []*commonv1.StatusUpdateRequest
}

func (s *statuses) Emit(_ context.Context, _ string, payload any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, payload.(*commonv1.StatusUpdateRequest))
	return nil
}

func (s *statuses) last(t *testing.T) *commonv1.StatusUpdateRequest {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	require.NotEmpty(t, s.updates, "no status was reported")
	return s.updates[len(s.updates)-1]
}

func (s *statuses) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.updates)
}

func extrasOf(u *commonv1.StatusUpdateRequest) map[string]any {
	return u.GetExtras().AsMap()
}

func testConfig(fw *fwtest.Server) *config.FlutterwaveConfig {
	return &config.FlutterwaveConfig{
		ClientID:           "client-id",
		ClientSecret:       "client-secret",
		Environment:        "sandbox",
		SandboxAPIBaseURL:  fw.URL,
		OAuthTokenURL:      fw.URL + "/token",
		DefaultRedirectURL: "https://pay.example.com/return",
	}
}

type harness struct {
	fw       *fwtest.Server
	statuses *statuses
	prompts  *promptHandler
	payments *paymentHandler
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	fw := fwtest.New(t)
	st := &statuses{}
	cfg := testConfig(fw)
	cli := client.NewClient()
	return &harness{
		fw:       fw,
		statuses: st,
		prompts:  NewPromptHandler(st, cli, nil, cfg).(*promptHandler),
		payments: NewPaymentHandler(st, cli, nil, cfg).(*paymentHandler),
	}
}

func cardPrompt(t *testing.T, id string, money *commonv1.Money, extra map[string]any) []byte {
	t.Helper()
	base := map[string]any{
		"encrypted_card_number": "enc-pan", "encrypted_expiry_month": "enc-mm",
		"encrypted_expiry_year": "enc-yy", "encrypted_cvv": "enc-cvv", "card_nonce": "abcdefghijkl",
		"customer_email": "payer@example.com", "tenant_id": "tenant-1", "partition_id": "partition-1",
	}
	for k, v := range extra {
		base[k] = v
	}
	return promptPayload(t, id, money, "", base)
}

func promptPayload(t *testing.T, id string, money *commonv1.Money, contact string, extra map[string]any) []byte {
	t.Helper()
	ex, err := structpb.NewStruct(extra)
	require.NoError(t, err)
	raw, err := proto.Marshal(&paymentv1.InitiatePromptRequest{
		Id:     id,
		Source: &commonv1.ContactLink{ProfileId: "payer-1", ContactId: contact},
		Amount: money,
		Route:  "flutterwave",
		Extra:  ex,
	})
	require.NoError(t, err)
	return raw
}

var kes1500 = &commonv1.Money{CurrencyCode: "KES", Units: 1500}

func TestCardChargeIsMadeOnceAcrossRedeliveries(t *testing.T) {
	h := newHarness(t)
	payload := cardPrompt(t, "prompt-card-1", kes1500, nil)

	require.NoError(t, h.prompts.Handle(context.Background(), nil, payload))
	require.NoError(t, h.prompts.Handle(context.Background(), nil, payload))

	require.Equal(t, 1, h.fw.ChargeCount(), "a redelivered prompt never charges again")
	require.Equal(t, 1, h.fw.Count("POST /orchestration/direct-charges"))
	got := h.statuses.last(t)
	require.Equal(t, "prompt-card-1", got.GetId())
	require.Equal(t, commonv1.STATUS_IN_PROCESS, got.GetStatus())
	ex := extrasOf(got)
	require.Equal(t, "requires_pin", ex["next_action"])
	require.Equal(t, "1500.00", ex["amount"])
	require.Equal(t, "KES", ex["currency"])

	var charge *client.Charge
	for _, ch := range h.fw.Charges {
		charge = ch
	}
	require.Equal(t, "prompt-card-1", charge.Meta[MetaPromptID])
	require.Equal(t, "tenant-1", charge.Meta[MetaTenantID])
	require.Equal(t, "1500.00", charge.Meta[MetaExpectedAmount])
}

func TestALostChargeResponseIsFoundNotRepeated(t *testing.T) {
	h := newHarness(t)
	// Flutterwave charges but the answer never arrives, on every retry.
	h.fw.Lose("POST /orchestration/direct-charges")
	h.fw.FailN("POST /orchestration/direct-charges", http.StatusBadGateway, 2)

	require.NoError(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-lost", kes1500, nil)))
	require.Equal(t, 1, h.fw.ChargeCount())
	require.Equal(t, commonv1.STATUS_IN_PROCESS, h.statuses.last(t).GetStatus(), "found by reference, not failed")
}

func TestAnUnknownChargeOutcomeIsRedeliveredNotFailed(t *testing.T) {
	h := newHarness(t)
	// Every attempt times out on Flutterwave's side, and no charge can be
	// found: the outcome is unknown.
	h.fw.FailN("POST /orchestration/direct-charges", http.StatusServiceUnavailable, 3)

	err := h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-unknown", kes1500, nil))
	require.Error(t, err, "the queue redelivers")
	require.Zero(t, h.statuses.count(), "nothing is reported on a guess")

	// Redelivered once Flutterwave answers again: charged exactly once.
	require.NoError(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-unknown", kes1500, nil)))
	require.Equal(t, 1, h.fw.ChargeCount())
}

func TestAnUnansweredLookupIsRedelivered(t *testing.T) {
	h := newHarness(t)
	h.fw.FailN("GET /charges", http.StatusServiceUnavailable, 3)
	require.Error(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-nolookup", kes1500, nil)))
	require.Zero(t, h.fw.ChargeCount(), "nothing is charged while a charge might exist")
	require.Zero(t, h.statuses.count())
}

func TestARefusedChargeFailsThePrompt(t *testing.T) {
	h := newHarness(t)
	h.fw.Fail("POST /orchestration/direct-charges", http.StatusBadRequest)

	require.NoError(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-refused", kes1500, nil)))
	got := h.statuses.last(t)
	require.Equal(t, commonv1.STATUS_FAILED, got.GetStatus())
	require.Equal(t, "provider_rejected", extrasOf(got)["error_code"])
}

func TestAChargeForAnotherAmountIsNotASuccess(t *testing.T) {
	h := newHarness(t)
	h.fw.ChargeStatus = "succeeded"
	h.fw.ChargedAmount = 15

	require.NoError(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-short", kes1500, nil)))
	got := h.statuses.last(t)
	require.Equal(t, commonv1.STATUS_FAILED, got.GetStatus())
	require.Equal(t, "amount_mismatch", extrasOf(got)["error_code"])
}

func TestAmountsAreSentExactlyOrRefused(t *testing.T) {
	h := newHarness(t)
	require.NoError(t, h.prompts.Handle(context.Background(), nil,
		cardPrompt(t, "prompt-cents", &commonv1.Money{CurrencyCode: "KES", Units: 10, Nanos: 500_000_000}, nil)))
	for _, ch := range h.fw.Charges {
		require.InDelta(t, 10.50, ch.Amount, 0.0001)
	}

	for name, money := range map[string]*commonv1.Money{
		"fraction of a shilling cent": {CurrencyCode: "KES", Units: 10, Nanos: 5_000_000},
		"fractional UGX":              {CurrencyCode: "UGX", Units: 1000, Nanos: 500_000_000},
		"no currency":                 {Units: 10},
		"zero":                        {CurrencyCode: "KES"},
	} {
		t.Run(name, func(t *testing.T) {
			before := h.fw.ChargeCount()
			require.NoError(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-"+name, money, nil)))
			require.Equal(t, commonv1.STATUS_FAILED, h.statuses.last(t).GetStatus())
			require.Equal(t, before, h.fw.ChargeCount())
		})
	}
}

func TestMobileMoneyNeverSwitchesCurrency(t *testing.T) {
	h := newHarness(t)
	// A USD prompt from a Kenyan number is not charged as KES.
	payload := promptPayload(t, "prompt-usd", &commonv1.Money{CurrencyCode: "USD", Units: 10}, "+254712345678",
		map[string]any{"payment_method_type": "mobile_money"})
	require.NoError(t, h.prompts.Handle(context.Background(), nil, payload))
	require.Equal(t, commonv1.STATUS_FAILED, h.statuses.last(t).GetStatus())
	require.Zero(t, h.fw.ChargeCount())

	payload = promptPayload(t, "prompt-kes", kes1500, "+254712345678", map[string]any{})
	require.NoError(t, h.prompts.Handle(context.Background(), nil, payload))
	require.Equal(t, 1, h.fw.ChargeCount())
	require.Equal(t, "mobile_money", extrasOf(h.statuses.last(t))["mode"])
}

func authorizePayload(t *testing.T, chargeID string, extra map[string]any) []byte {
	t.Helper()
	base := map[string]any{"action": "authorize", "charge_id": chargeID}
	for k, v := range extra {
		base[k] = v
	}
	return promptPayload(t, "prompt-auth", kes1500, "", base)
}

func chargedOnce(t *testing.T, h *harness) string {
	t.Helper()
	require.NoError(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-auth", kes1500, nil)))
	return extrasOf(h.statuses.last(t))["charge_id"].(string)
}

func TestAuthorizeCompletesTheCharge(t *testing.T) {
	h := newHarness(t)
	chargeID := chargedOnce(t, h)

	require.NoError(t, h.prompts.Handle(context.Background(), nil,
		authorizePayload(t, chargeID, map[string]any{"authorization_type": "pin", "encrypted_pin": "x", "nonce": "abcdefghijkl"})))
	require.Equal(t, "requires_otp", extrasOf(h.statuses.last(t))["next_action"])

	otp := authorizePayload(t, chargeID, map[string]any{"authorization_type": "otp", "otp": "123456"})
	require.NoError(t, h.prompts.Handle(context.Background(), nil, otp))
	require.Equal(t, commonv1.STATUS_SUCCESSFUL, h.statuses.last(t).GetStatus())

	// A redelivered step on a settled charge reports it without sending again.
	puts := h.fw.Count("PUT /charges/" + chargeID)
	require.NoError(t, h.prompts.Handle(context.Background(), nil, otp))
	require.Equal(t, puts, h.fw.Count("PUT /charges/"+chargeID))
	require.Equal(t, commonv1.STATUS_SUCCESSFUL, h.statuses.last(t).GetStatus())
}

func TestAWrongOTPLeavesTheChargeOpen(t *testing.T) {
	h := newHarness(t)
	chargeID := chargedOnce(t, h)

	require.NoError(t, h.prompts.Handle(context.Background(), nil,
		authorizePayload(t, chargeID, map[string]any{"authorization_type": "otp", "otp": "000000"})))
	got := h.statuses.last(t)
	require.Equal(t, commonv1.STATUS_IN_PROCESS, got.GetStatus(), "the payer can try again")
	require.Contains(t, extrasOf(got), "authorization_error")
}

func TestAuthorizeRefusesAChargeOfAnotherPrompt(t *testing.T) {
	h := newHarness(t)
	require.NoError(t, h.prompts.Handle(context.Background(), nil, cardPrompt(t, "prompt-other", kes1500, nil)))
	other := extrasOf(h.statuses.last(t))["charge_id"].(string)
	reported := h.statuses.count()

	require.NoError(t, h.prompts.Handle(context.Background(), nil,
		authorizePayload(t, other, map[string]any{"authorization_type": "otp", "otp": "123456"})))
	require.Equal(t, reported, h.statuses.count())
	require.Zero(t, h.fw.Count("PUT /charges/"+other))
}

func TestTokenRenewalIsChargedOnce(t *testing.T) {
	h := newHarness(t)
	payload := promptPayload(t, "prompt-renew", kes1500, "", map[string]any{
		"payment_method_id": "pmd_1", "customer_id": "cus_1", "tenant_id": "tenant-1",
	})
	h.fw.Lose("POST /charges")
	h.fw.FailN("POST /charges", http.StatusBadGateway, 2)
	require.NoError(t, h.prompts.Handle(context.Background(), nil, payload))
	require.NoError(t, h.prompts.Handle(context.Background(), nil, payload))
	require.Equal(t, 1, h.fw.ChargeCount(), "a saved card is never charged twice")
}

func payoutPayload(t *testing.T, id, contact string, money *commonv1.Money, extra map[string]any) []byte {
	t.Helper()
	ex, err := structpb.NewStruct(extra)
	require.NoError(t, err)
	raw, err := proto.Marshal(&paymentv1.Payment{
		Id:        id,
		Recipient: &commonv1.ContactLink{ProfileId: "payee-1", ContactId: contact},
		Amount:    money,
		Extra:     ex,
	})
	require.NoError(t, err)
	return raw
}

func TestPayoutIsSentOnceAcrossRedeliveries(t *testing.T) {
	h := newHarness(t)
	payload := payoutPayload(t, "payment-1", "+256772123456",
		&commonv1.Money{CurrencyCode: "UGX", Units: 50000}, map[string]any{"tenant_id": "tenant-1"})

	require.NoError(t, h.payments.Handle(context.Background(), nil, payload))
	require.NoError(t, h.payments.Handle(context.Background(), nil, payload))
	require.Equal(t, 1, h.fw.TransferCount())
	got := h.statuses.last(t)
	require.Equal(t, commonv1.STATUS_IN_PROCESS, got.GetStatus())
	for _, tr := range h.fw.Transfers {
		require.Equal(t, "payment-1", tr.Meta["payment_id"])
		require.Equal(t, "tenant-1", tr.Meta["tenant_id"])
		require.InDelta(t, 50000.0, tr.Amount["value"], 0.0001)
	}
}

func TestPayoutWithAnUnknownOutcomeIsRedeliveredNotFailed(t *testing.T) {
	h := newHarness(t)
	payload := payoutPayload(t, "payment-2", "+254712345678",
		&commonv1.Money{CurrencyCode: "KES", Units: 100, Nanos: 500_000_000}, nil)
	h.fw.Lose("POST /transfers")
	h.fw.FailN("POST /transfers", http.StatusServiceUnavailable, 2)

	err := h.payments.Handle(context.Background(), nil, payload)
	if err == nil {
		// The client's own retry got the original answer back.
		require.Equal(t, commonv1.STATUS_IN_PROCESS, h.statuses.last(t).GetStatus())
	} else {
		require.Zero(t, h.statuses.count(), "never failed on a guess")
		require.NoError(t, h.payments.Handle(context.Background(), nil, payload))
	}
	require.Equal(t, 1, h.fw.TransferCount(), "paid exactly once")
	for _, tr := range h.fw.Transfers {
		require.InDelta(t, 100.50, tr.Amount["value"], 0.0001, "cents are paid, not rounded")
	}
}

func TestPayoutRefusals(t *testing.T) {
	h := newHarness(t)
	tests := map[string][]byte{
		"not a mobile number":   payoutPayload(t, "p-a", "payee@example.com", kes1500, nil),
		"currency of no wallet": payoutPayload(t, "p-b", "+254712345678", &commonv1.Money{CurrencyCode: "USD", Units: 5}, nil),
		"fractional UGX":        payoutPayload(t, "p-c", "+256772123456", &commonv1.Money{CurrencyCode: "UGX", Units: 5, Nanos: 500_000_000}, nil),
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, h.payments.Handle(context.Background(), nil, payload))
			require.Equal(t, commonv1.STATUS_FAILED, h.statuses.last(t).GetStatus())
		})
	}
	require.Zero(t, h.fw.TransferCount())

	// Caller extras never choose the destination.
	payload := payoutPayload(t, "p-d", "+254712345678", kes1500, map[string]any{
		"account_number": "0000000000", "account_bank": "044",
	})
	require.NoError(t, h.payments.Handle(context.Background(), nil, payload))
	require.Equal(t, 1, h.fw.TransferCount())

	h.fw.Fail("POST /transfers", http.StatusBadRequest)
	require.NoError(t, h.payments.Handle(context.Background(), nil,
		payoutPayload(t, "p-e", "+254712345678", kes1500, nil)))
	got := h.statuses.last(t)
	require.Equal(t, commonv1.STATUS_FAILED, got.GetStatus())
	require.Equal(t, "provider_rejected", extrasOf(got)["error_code"])
}

func TestKeysFromQueueHeadersAreIgnored(t *testing.T) {
	h := newHarness(t)
	headers := map[string]string{config.HeaderClientID: "attacker", config.HeaderClientSecret: "attacker"}
	creds, err := h.prompts.extractCredentials(context.Background(), headers)
	require.NoError(t, err)
	require.Equal(t, "client-id", creds.ClientID)
}
