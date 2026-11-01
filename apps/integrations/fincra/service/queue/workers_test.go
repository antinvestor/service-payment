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

package queue_test

import (
	"context"
	"errors"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/config"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/credentials"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/fincratest"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/queue"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// fakeEvents captures emitted status updates.
type fakeEvents struct {
	frameEvents.Manager
	emitted []*commonv1.StatusUpdateRequest
	err     error
}

func (f *fakeEvents) Emit(_ context.Context, name string, payload any) error {
	if f.err != nil {
		return f.err
	}
	if name != events.PaymentStatusUpdateEvent {
		return errors.New("unexpected event " + name)
	}
	req, ok := payload.(*commonv1.StatusUpdateRequest)
	if !ok {
		return errors.New("unexpected payload")
	}
	f.emitted = append(f.emitted, req)
	return nil
}

func (f *fakeEvents) last(t *testing.T) (*commonv1.StatusUpdateRequest, map[string]any) {
	t.Helper()
	require.NotEmpty(t, f.emitted, "expected a status update")
	st := f.emitted[len(f.emitted)-1]
	return st, st.GetExtras().AsMap()
}

// fakeProfiles answers the payee's profile name.
type fakeProfiles struct {
	profilev1connect.ProfileServiceClient
	name string
}

func (f *fakeProfiles) GetById(_ context.Context, req *profilev1.GetByIdRequest) (*profilev1.GetByIdResponse, error) {
	props, _ := structpb.NewStruct(map[string]any{"au_name": f.name})
	return &profilev1.GetByIdResponse{Data: &profilev1.ProfileObject{Id: req.GetId(), Properties: props}}, nil
}

type harness struct {
	srv    *fincratest.Server
	events *fakeEvents
	cfg    *config.FincraConfig
}

func newHarness(t *testing.T, mode string) *harness {
	t.Helper()
	srv := fincratest.New(t)
	return &harness{
		srv:    srv,
		events: &fakeEvents{},
		cfg: &config.FincraConfig{
			SecretKey: fincratest.SecretKey, PublicKey: fincratest.PublicKey, BusinessID: fincratest.BusinessID,
			Environment: "sandbox", BaseURL: srv.URL, CollectionMode: mode, FeeBearer: "business",
			DefaultRedirectURL: "https://pay.example.com/return", WebhookBaseURL: "https://hooks.example.com",
			DefaultBeneficiaryName: "Mobile Money Customer",
		},
	}
}

func (h *harness) prompt(t *testing.T, msg *paymentv1.InitiatePromptRequest, headers map[string]string) {
	t.Helper()
	require.NoError(t, h.promptErr(t, msg, headers))
}

func (h *harness) promptErr(t *testing.T, msg *paymentv1.InitiatePromptRequest, headers map[string]string) error {
	t.Helper()
	worker := queue.NewPromptHandler(h.events, client.NewClient(), credentials.NewResolver(nil, h.cfg), h.cfg)
	payload, err := proto.Marshal(msg)
	require.NoError(t, err)
	claims := &security.AuthenticationClaims{TenantID: "tenant-1", PartitionID: "partition-1"}
	return worker.Handle(claims.ClaimsToContext(context.Background()), headers, payload)
}

func (h *harness) pay(t *testing.T, msg *paymentv1.Payment, profiles profilev1connect.ProfileServiceClient) {
	t.Helper()
	require.NoError(t, h.payErr(t, msg, profiles))
}

// payErr runs the payout worker and returns its result: an error means the
// queue redelivers the message.
func (h *harness) payErr(t *testing.T, msg *paymentv1.Payment, profiles profilev1connect.ProfileServiceClient) error {
	t.Helper()
	cli := client.NewClient()
	worker := queue.NewPaymentHandler(h.events, cli, client.NewCatalog(cli, 0), profiles, credentials.NewResolver(nil, h.cfg), h.cfg)
	payload, err := proto.Marshal(msg)
	require.NoError(t, err)
	return worker.Handle(context.Background(), map[string]string{}, payload)
}

func extras(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	require.NoError(t, err)
	return s
}

func checkoutPrompt(t *testing.T, units int64, nanos int32) *paymentv1.InitiatePromptRequest {
	t.Helper()
	return &paymentv1.InitiatePromptRequest{
		Id:     "prompt-1",
		Source: &commonv1.ContactLink{ProfileId: "payer-1", ContactId: "+2348189299860"},
		Amount: &commonv1.Money{CurrencyCode: "NGN", Units: units, Nanos: nanos},
		Route:  "fincra",
		Extra: extras(t, map[string]any{
			"customer_name": "Ada Obi", "customer_email": "ada@example.com",
			"success_url": "https://shop.example.com/done", "session_ref": "sess-9",
		}),
	}
}

// --- collections: hosted checkout ---

func TestPrompt_HostedCheckout(t *testing.T) {
	h := newHarness(t, "checkout")
	h.prompt(t, checkoutPrompt(t, 9758, 0), map[string]string{})

	req := h.srv.LastCheckout
	require.NotNil(t, req)
	assert.Equal(t, int64(9758), req.Amount)
	assert.Equal(t, "NGN", req.Currency)
	assert.Equal(t, "prompt-1", req.Reference)
	assert.Equal(t, "business", req.FeeBearer)
	assert.Equal(t, "https://shop.example.com/done", req.RedirectURL)
	assert.Equal(t, client.Customer{Name: "Ada Obi", Email: "ada@example.com", PhoneNumber: "2348189299860"}, req.Customer)
	assert.Equal(t, map[string]string{
		"prompt_id": "prompt-1", "tenant_id": "tenant-1", "partition_id": "partition-1", "session_ref": "sess-9",
	}, req.Metadata, "tenancy travels with the checkout so the webhook can be attributed")
	assert.Empty(t, req.PaymentMethods, "no method asked: Fincra offers every method enabled")

	st, ex := h.events.last(t)
	assert.Equal(t, "prompt-1", st.GetId())
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	payCode := h.srv.Checkout("prompt-1").Reference
	assert.Equal(t, payCode, st.GetExternalId())
	assert.Equal(t, "https://sandbox-checkout.fincra.com/pay/"+payCode, ex["checkout_url"])
	assert.Equal(t, "redirect_url", ex["next_action"])
	assert.Equal(t, "fincra", ex["provider"])
	assert.Equal(t, "9758", ex["requested_amount"])
}

func TestPrompt_HostedCheckoutNarrowsToRequestedMethod(t *testing.T) {
	h := newHarness(t, "checkout")
	msg := checkoutPrompt(t, 500, 0)
	msg.Extra.Fields["payment_method_type"] = structpb.NewStringValue("bank_transfer")
	h.prompt(t, msg, map[string]string{})

	assert.Equal(t, []string{"bank_transfer"}, h.srv.LastCheckout.PaymentMethods)
	assert.Equal(t, "bank_transfer", h.srv.LastCheckout.DefaultPaymentMethod)
}

func TestPrompt_RedeliveryReusesTheCheckout(t *testing.T) {
	h := newHarness(t, "checkout")
	h.prompt(t, checkoutPrompt(t, 500, 0), map[string]string{})
	first := h.srv.Checkout("prompt-1").Reference
	h.prompt(t, checkoutPrompt(t, 500, 0), map[string]string{})

	assert.Equal(t, first, h.srv.Checkout("prompt-1").Reference)
	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, "https://sandbox-checkout.fincra.com/pay/"+first, ex["checkout_url"])
	assert.Equal(t, 1, countCalls(h.srv.Calls, "POST /checkout/payments"), "one checkout per prompt")
}

func TestPrompt_RedeliveryAfterPaymentReportsSuccess(t *testing.T) {
	h := newHarness(t, "checkout")
	h.prompt(t, checkoutPrompt(t, 500, 0), map[string]string{})
	h.srv.SetCheckout("prompt-1", "success", "", 500)
	h.prompt(t, checkoutPrompt(t, 500, 0), map[string]string{})

	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, "500", ex["amount"])
	assert.Equal(t, "NGN", ex["currency"])
	assert.NotContains(t, ex, "checkout_url")
}

func TestPrompt_FractionalAmountIsRefused(t *testing.T) {
	h := newHarness(t, "checkout")
	h.prompt(t, checkoutPrompt(t, 100, 400_000_000), map[string]string{})

	assert.Nil(t, h.srv.LastCheckout, "nothing is sent to Fincra")
	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "INVALID_AMOUNT", ex["failure_code"])
}

func TestPrompt_EmailIsRequired(t *testing.T) {
	h := newHarness(t, "checkout")
	msg := checkoutPrompt(t, 500, 0)
	msg.Extra = extras(t, map[string]any{"customer_name": "Ada"})
	h.prompt(t, msg, map[string]string{})

	_, ex := h.events.last(t)
	assert.Equal(t, "MISSING_CUSTOMER_EMAIL", ex["failure_code"])

	h.cfg.FallbackCustomerEmail = "payments@example.com"
	h.prompt(t, msg, map[string]string{})
	assert.Equal(t, "payments@example.com", h.srv.LastCheckout.Customer.Email)
}

func TestPrompt_MissingCredentials(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.SecretKey = ""
	h.prompt(t, checkoutPrompt(t, 500, 0), map[string]string{})

	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "CREDENTIALS_ERROR", ex["failure_code"])
}

func TestPrompt_HeaderCredentialsOverrideConfig(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.SecretKey = "wrong"
	h.prompt(t, checkoutPrompt(t, 500, 0), map[string]string{config.HeaderSecretKey: fincratest.SecretKey})

	st, _ := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
}

// --- collections: direct charge ---

func momoPrompt(t *testing.T, phone, currency string) *paymentv1.InitiatePromptRequest {
	t.Helper()
	return &paymentv1.InitiatePromptRequest{
		Id:     "prompt-2",
		Source: &commonv1.ContactLink{ProfileId: "payer-1", ContactId: phone},
		Amount: &commonv1.Money{CurrencyCode: currency, Units: 600},
		Extra:  extras(t, map[string]any{"customer_email": "kweku@example.com"}),
	}
}

func TestPrompt_DirectMobileMoney(t *testing.T) {
	h := newHarness(t, "direct")
	h.prompt(t, momoPrompt(t, "+254700000000", "KES"), map[string]string{config.HeaderConnectionCredentials: ""})

	req := h.srv.LastCharge
	require.NotNil(t, req)
	assert.Equal(t, "mobile_money", req.Type)
	assert.Equal(t, "+254700000000", req.Phone)
	assert.Equal(t, "MPESA", req.Operator, "operator detected from the Safaricom prefix")
	assert.Equal(t, "prompt-2", req.Reference)
	assert.Equal(t, "https://hooks.example.com/webhook/fincra?partition_id=partition-1&tenant_id=tenant-1", req.WebhookURL)

	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, h.srv.Charge("prompt-2").ID.String(), st.GetExternalId())
	assert.Equal(t, "You will receive a prompt on mobile number +254700000000.\nEnter your PIN to authorize your payment.",
		ex["payment_instruction"])
	assert.NotContains(t, ex, "next_action")
}

func TestPrompt_DirectMobileMoneyOperatorFromConfig(t *testing.T) {
	h := newHarness(t, "direct")
	h.cfg.ChargeOperators = `{"MPESA":"SAFARICOM"}`
	h.prompt(t, momoPrompt(t, "+254700000000", "KES"), map[string]string{})
	assert.Equal(t, "SAFARICOM", h.srv.LastCharge.Operator)
}

func TestPrompt_DirectMobileMoneyUnknownOperator(t *testing.T) {
	h := newHarness(t, "direct")
	h.prompt(t, momoPrompt(t, "+2250700000000", "XOF"), map[string]string{})

	assert.Nil(t, h.srv.LastCharge)
	_, ex := h.events.last(t)
	assert.Equal(t, "UNKNOWN_OPERATOR", ex["failure_code"])
}

func TestPrompt_DirectMobileMoneyWithOTP(t *testing.T) {
	h := newHarness(t, "direct")
	h.prompt(t, momoPrompt(t, "+233240000000", "GHS"), map[string]string{})
	assert.Equal(t, "MTN", h.srv.LastCharge.Operator)

	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, "requires_otp", ex["next_action"])
	assert.Equal(t, "requires_otp", ex["next_action_type"])
	chargeID := ex["charge_id"].(string)
	require.NotEmpty(t, chargeID)

	// Checkout sends the OTP back on the same prompt (action=authorize).
	authorize := func(otp string) {
		h.prompt(t, &paymentv1.InitiatePromptRequest{
			Id: "prompt-2", Amount: &commonv1.Money{CurrencyCode: "GHS", Units: 600},
			Extra: extras(t, map[string]any{"action": "authorize", "charge_id": chargeID, "otp": otp}),
		}, map[string]string{})
	}

	authorize("000000")
	st, ex = h.events.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "Invalid OTP supplied", ex["failure_message"])

	authorize(fincratest.ValidOTP)
	assert.Equal(t, fincratest.ValidOTP, h.srv.LastAuthorize.OTP)
	st, ex = h.events.last(t)
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
	assert.Equal(t, "600", ex["amount"])
	assert.Equal(t, "GHS", ex["currency"])
}

func TestPrompt_AuthorizeWithoutCharge(t *testing.T) {
	h := newHarness(t, "direct")
	h.prompt(t, &paymentv1.InitiatePromptRequest{
		Id: "prompt-3", Extra: extras(t, map[string]any{"action": "authorize", "otp": "1"}),
	}, map[string]string{})
	_, ex := h.events.last(t)
	assert.Equal(t, "MISSING_CHARGE_ID", ex["failure_code"])
}

func TestPrompt_DirectBankTransfer(t *testing.T) {
	h := newHarness(t, "direct")
	msg := checkoutPrompt(t, 1200, 0)
	msg.Extra.Fields["channel_type"] = structpb.NewStringValue("bank")
	h.prompt(t, msg, map[string]string{})

	assert.Equal(t, "bank_transfer", h.srv.LastCharge.Type)
	assert.Empty(t, h.srv.LastCharge.Operator)

	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, "Test Bank", ex["bank_name"])
	assert.Equal(t, "9600414830", ex["bank_account_number"])
	assert.Equal(t, "Fincra Checkout", ex["bank_account_name"])
	assert.Equal(t, "2030-01-01T16:17:56.180Z", ex["payment_expires_at"])
	assert.Equal(t, "requires_bank_transfer", ex["next_action"])
	assert.Equal(t,
		"Transfer exactly NGN 1200 to Test Bank account 9600414830 (Fincra Checkout) before 2030-01-01T16:17:56.180Z",
		ex["payment_instruction"])
}

func TestPrompt_DirectCardFallsBackToHostedCheckout(t *testing.T) {
	h := newHarness(t, "direct")
	msg := checkoutPrompt(t, 500, 0)
	msg.Extra.Fields["payment_method_type"] = structpb.NewStringValue("card")
	h.prompt(t, msg, map[string]string{})

	assert.Nil(t, h.srv.LastCharge, "direct card charges need PCI-DSS; cards go through the hosted page")
	require.NotNil(t, h.srv.LastCheckout)
	assert.Equal(t, []string{"card"}, h.srv.LastCheckout.PaymentMethods)
}

func TestPrompt_DirectChargeRedeliveryReusesTheCharge(t *testing.T) {
	h := newHarness(t, "direct")
	h.prompt(t, momoPrompt(t, "+254700000000", "KES"), map[string]string{})
	h.prompt(t, momoPrompt(t, "+254700000000", "KES"), map[string]string{})
	assert.Equal(t, 1, countCalls(h.srv.Calls, "POST /checkout/charges"))
	st, _ := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
}

func TestPrompt_UnderpaidChargeFails(t *testing.T) {
	h := newHarness(t, "direct")
	h.srv.PutCharge(&client.Charge{
		ID: "77", Reference: "prompt-2", Status: "success", VarianceType: "underpayment",
		Amount: 600, AmountExpected: 600, AmountReceived: 100, Currency: "KES", Type: "bank_transfer",
	})
	h.prompt(t, momoPrompt(t, "+254700000000", "KES"), map[string]string{})

	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "AMOUNT_MISMATCH", ex["failure_code"])
}

// --- payouts ---

func payout(phone string, units int64, nanos int32, currency string) *paymentv1.Payment {
	return &paymentv1.Payment{
		Id:          "payment-1",
		ReferenceId: "refund-42",
		Recipient:   &commonv1.ContactLink{ProfileId: "payee-1", ContactId: phone},
		Amount:      &commonv1.Money{CurrencyCode: currency, Units: units, Nanos: nanos},
	}
}

func TestPayout_MobileMoney(t *testing.T) {
	h := newHarness(t, "checkout")
	h.srv.Providers["KES|mobile_money_wallet"] = []client.Provider{
		{Code: "MPESA", Name: "M-Pesa"}, {Code: "AIRTEL", Name: "Airtel Money"},
	}
	h.pay(t, payout("+254731234567", 1000, 0, "KES"), &fakeProfiles{name: "Mary Wanjiru Atieno"})

	req := h.srv.LastPayout
	require.NotNil(t, req)
	assert.Equal(t, fincratest.BusinessID, req.Business)
	assert.Equal(t, "KES", req.SourceCurrency)
	assert.Equal(t, "KES", req.DestinationCurrency)
	assert.Equal(t, "1000", req.Amount.String())
	assert.Equal(t, "mobile_money_wallet", req.PaymentDestination)
	assert.Equal(t, "payment-1", req.CustomerReference)
	assert.Equal(t, "Payout refund-42", req.Description)
	assert.Equal(t, client.Beneficiary{
		FirstName: "Mary Wanjiru", LastName: "Atieno", AccountHolderName: "Mary Wanjiru Atieno",
		Type: "individual", Country: "KE", AccountNumber: "254731234567", MobileMoneyCode: "AIRTEL",
	}, req.Beneficiary)

	st, ex := h.events.last(t)
	assert.Equal(t, "payment-1", st.GetId())
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, h.srv.Payout("payment-1").Reference, st.GetExternalId())
	assert.Equal(t, "payment", ex["entity_type"])
}

func TestPayout_CentsWhereTheRailAllowsThem(t *testing.T) {
	h := newHarness(t, "checkout")
	h.srv.Providers["KES|mobile_money_wallet"] = []client.Provider{{Code: "MPESA", Name: "M-Pesa"}}
	h.pay(t, payout("+254700000000", 1000, 500_000_000, "KES"), nil)

	require.NotNil(t, h.srv.LastPayout)
	assert.Equal(t, "MPESA", h.srv.LastPayout.Beneficiary.MobileMoneyCode)
	assert.Equal(t, "1000.50", h.srv.LastPayout.Amount.String())
}

func TestPayout_ProviderFromConfig(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.PayoutProviders = `{"UGX":"MTN_UG"}`
	h.pay(t, payout("+256770000000", 5000, 0, "UGX"), nil)
	assert.Equal(t, "MTN_UG", h.srv.LastPayout.Beneficiary.MobileMoneyCode)
	assert.Equal(t, "Mobile Money Customer", h.srv.LastPayout.Beneficiary.AccountHolderName, "no profile client: default name")
}

func TestPayout_WholeUnitsForUgandaAndAirtel(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.PayoutProviders = `{"UGX":"MTN"}`
	h.pay(t, payout("+256770000000", 5000, 500_000_000, "UGX"), nil)

	assert.Nil(t, h.srv.LastPayout)
	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, "INVALID_AMOUNT", ex["failure_code"])
}

func TestPayout_UnknownOperator(t *testing.T) {
	h := newHarness(t, "checkout")
	h.srv.Providers["XOF|mobile_money_wallet"] = []client.Provider{{Code: "ORANGE"}, {Code: "MTN"}}
	h.pay(t, payout("+2250701234567", 100, 0, "XOF"), nil)

	assert.Nil(t, h.srv.LastPayout)
	_, ex := h.events.last(t)
	assert.Equal(t, "UNKNOWN_OPERATOR", ex["failure_code"])
}

func TestPayout_RecipientMustBeANumber(t *testing.T) {
	h := newHarness(t, "checkout")
	h.pay(t, payout("payee@example.com", 100, 0, "KES"), nil)
	_, ex := h.events.last(t)
	assert.Equal(t, "MISSING_PHONE_NUMBER", ex["failure_code"])
}

func TestPayout_RedeliveryDoesNotPayTwice(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.PayoutProviders = `{"KES":"MPESA"}`
	h.pay(t, payout("+254700000000", 1000, 0, "KES"), nil)
	h.srv.SetPayoutStatus("payment-1", "successful", "")
	h.pay(t, payout("+254700000000", 1000, 0, "KES"), nil)

	assert.Equal(t, 1, countCalls(h.srv.Calls, "POST /disbursements/payouts"))
	st, _ := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, st.GetStatus())
}

func TestPayout_InsufficientBalance(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.PayoutProviders = `{"KES":"MPESA"}`
	h.srv.PayoutError = client.ErrTypeInsufficientFunds
	h.pay(t, payout("+254700000000", 1000, 0, "KES"), nil)

	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_FAILED, st.GetStatus())
	assert.Equal(t, commonv1.STATE_INACTIVE, st.GetState())
	assert.Equal(t, "NO_ENOUGH_MONEY_IN_WALLET", ex["failure_code"])
}

func countCalls(calls []string, want string) int {
	n := 0
	for _, c := range calls {
		if c == want {
			n++
		}
	}
	return n
}

// A payout whose create call fails without a definitive answer may have been
// accepted. It is never reported failed on a guess.
func TestPayout_UnknownOutcomeIsRedeliveredNotFailed(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.PayoutProviders = `{"KES":"MPESA"}`
	h.srv.FailNext["POST /disbursements/payouts"] = []int{503}

	err := h.payErr(t, payout("+254700000000", 1000, 0, "KES"), nil)
	require.Error(t, err, "the queue must redeliver")
	assert.Empty(t, h.events.emitted, "no status is guessed")

	h.pay(t, payout("+254700000000", 1000, 0, "KES"), nil)
	st, _ := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, 2, countCalls(h.srv.Calls, "POST /disbursements/payouts"), "the failed attempt and the redelivery")
	require.NotNil(t, h.srv.Payout("payment-1"))
}

func TestPayout_AcceptedButResponseLost(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.PayoutProviders = `{"KES":"MPESA"}`
	h.srv.LoseResponse["/disbursements/payouts"] = []int{504}

	h.pay(t, payout("+254700000000", 1000, 0, "KES"), nil)

	require.NotNil(t, h.srv.Payout("payment-1"))
	st, _ := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus(), "the payout Fincra holds is reported, not a failure")
	assert.Equal(t, h.srv.Payout("payment-1").Reference, st.GetExternalId())
}

func TestPayout_LostStatusIsRedelivered(t *testing.T) {
	h := newHarness(t, "checkout")
	h.cfg.PayoutProviders = `{"KES":"MPESA"}`
	h.events.err = errors.New("events down")
	require.Error(t, h.payErr(t, payout("+254700000000", 1000, 0, "KES"), nil))

	h.events.err = nil
	h.pay(t, payout("+254700000000", 1000, 0, "KES"), nil)
	assert.Equal(t, 1, countCalls(h.srv.Calls, "POST /disbursements/payouts"), "the redelivery reports, it does not pay again")
	st, _ := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
}

func TestPrompt_CheckoutUnknownOutcome(t *testing.T) {
	h := newHarness(t, "checkout")
	h.srv.FailNext["POST /checkout/payments"] = []int{500}
	require.Error(t, h.promptErr(t, checkoutPrompt(t, 500, 0), map[string]string{}))
	assert.Empty(t, h.events.emitted)

	h.srv.LoseResponse["/checkout/payments"] = []int{502}
	h.prompt(t, checkoutPrompt(t, 500, 0), map[string]string{})
	st, ex := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Contains(t, ex["checkout_url"], h.srv.Checkout("prompt-1").Reference)
}

func TestPrompt_ChargeUnknownOutcome(t *testing.T) {
	h := newHarness(t, "direct")
	h.srv.LoseResponse["/checkout/charges"] = []int{503}
	h.prompt(t, momoPrompt(t, "+254700000000", "KES"), map[string]string{})

	st, _ := h.events.last(t)
	assert.Equal(t, commonv1.STATUS_IN_PROCESS, st.GetStatus())
	assert.Equal(t, h.srv.Charge("prompt-2").ID.String(), st.GetExternalId())
}
