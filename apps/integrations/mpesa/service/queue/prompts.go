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
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type promptHandler struct {
	credentialResolver
	emitter      *events.StatusEmitter
	mpesaCli     client.MpesaClient
	statusReader StatusReader
	metrics      *integrationobs.Metrics
}

// StatusReader reads an entity's latest status from the payment service
// (satisfied by paymentv1connect.PaymentServiceClient).
type StatusReader interface {
	Status(
		ctx context.Context,
		req *commonv1.StatusRequest,
	) (*commonv1.StatusResponse, error)
}

// NewPromptHandler creates a queue worker for handling STK Push prompt requests.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	mpesaCli client.MpesaClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MpesaConfig,
	statusReader StatusReader,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		mpesaCli:           mpesaCli,
		statusReader:       statusReader,
		metrics:            integrationobs.NewMetrics(providerName),
	}
}

func (h *promptHandler) fail(ctx context.Context, promptID, reason, message string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, kindPrompt, reason)
	return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, message, extras)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mpesa.prompt")
	defer logger.Release()

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, kindPrompt, "unmarshal_error")
		return nil // non-retriable
	}
	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	// Redelivered message: an STK push already went out for this prompt.
	// Pushing again would overwrite its checkout_request_id binding and the
	// callback for the first push (the one the customer may pay) would be
	// rejected.
	if existing := h.pushedCheckoutRequestID(ctx, promptID); existing != "" {
		logger.WithField("checkout_request_id", existing).Info("STK push already issued for prompt, skipping")
		h.metrics.QueueProcessed(ctx, kindPrompt)
		return nil
	}

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, promptID, "credentials_error", err.Error(), nil)
	}

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}

	// Daraja only charges KES, in whole shillings; pushing a USD amount as
	// KES would let a session complete for the wrong currency, and a
	// rounded amount would charge something else than what was asked.
	if cur := strings.ToUpper(strings.TrimSpace(prompt.GetAmount().GetCurrencyCode())); cur != "" && cur != client.Currency {
		logger.WithField("currency", cur).Error("unsupported prompt currency for M-Pesa")
		return h.fail(ctx, promptID, "rejected", "M-Pesa only supports "+client.Currency+", got "+cur, nil)
	}
	units, err := collection.WholeUnits(prompt.GetAmount())
	if err != nil {
		logger.WithError(err).Error("prompt amount cannot be sent to M-Pesa")
		return h.fail(ctx, promptID, "rejected", "invalid amount: "+err.Error(), nil)
	}
	amount := strconv.FormatInt(units, 10)

	timestamp := time.Now().Format("20060102150405")
	password := client.STKPassword(creds.Shortcode, creds.Passkey, timestamp)

	accountRef := promptID
	if ref, ok := prompt.GetExtra().GetFields()["account_reference"]; ok && ref.GetStringValue() != "" {
		accountRef = ref.GetStringValue()
	}

	stkReq := &client.STKPushRequest{
		BusinessShortCode: creds.Shortcode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            phoneNumber,
		PartyB:            creds.Shortcode,
		PhoneNumber:       phoneNumber,
		CallBackURL:       stkCallbackURL(creds.CallbackURL, headers, promptID),
		AccountReference:  accountRef,
		TransactionDesc:   "Payment prompt",
	}

	resp, err := h.mpesaCli.STKPush(ctx, creds, stkReq)
	if err != nil {
		if client.IsRejected(err) {
			logger.WithError(err).Error("STK push refused by Daraja")
			return h.fail(ctx, promptID, "provider_error", err.Error(), nil)
		}
		// Unknown outcome (timeout, 5xx): redeliver. If the push did go out
		// the customer sees a second prompt; only the one whose
		// CheckoutRequestID is bound below can be recorded.
		logger.WithError(err).Warn("STK push outcome unknown; redelivering")
		h.metrics.QueueFailed(ctx, kindPrompt, "provider_error")
		return fmt.Errorf("stk push %s outcome unknown: %w", promptID, err)
	}

	logger.WithField("checkout_request_id", resp.CheckoutRequestID).Debug("STK push initiated")

	// checkout_request_id binds this prompt to the Daraja request so the STK
	// callback can prove it belongs here; requested_amount and the credential
	// connection let it verify amount and re-query Daraja.
	inProcessExtras := map[string]any{
		"merchant_request_id":         resp.MerchantRequestID,
		client.ExtraCheckoutRequestID: resp.CheckoutRequestID,
		client.ExtraRequestedAmount:   amount,
		client.ExtraRequestedCurrency: client.Currency,
		"response_code":               resp.ResponseCode,
		"customer_message":            resp.CustomerMessage,
	}
	if connection := headers[config.HeaderConnectionCredentials]; connection != "" {
		inProcessExtras[client.ExtraCredentialsConnection] = connection
	}
	if err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, resp.CheckoutRequestID, commonv1.STATUS_IN_PROCESS, inProcessExtras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPrompt)
	return nil
}

// stkCallbackURL builds the STK callback URL: tenant params plus the prompt id,
// so the callback can record the final status under the prompt the checkout
// polls rather than under Daraja's CheckoutRequestID.
func stkCallbackURL(base string, headers map[string]string, promptID string) string {
	callbackURL := appendTenantParams(base+"/webhook/mpesa/stk", headers)
	if promptID == "" {
		return callbackURL
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return callbackURL
	}
	q := u.Query()
	q.Set(client.CallbackParamPromptID, promptID)
	u.RawQuery = q.Encode()
	return u.String()
}

// pushedCheckoutRequestID returns the CheckoutRequestID already bound to the
// prompt, or "" when none is recorded. A lookup failure is treated as "none"
// (the payment service reports a missing status as an error, and the QUEUED
// status may not be persisted yet), so a transient error can still allow a
// duplicate push.
func (h *promptHandler) pushedCheckoutRequestID(ctx context.Context, promptID string) string {
	if h.statusReader == nil || promptID == "" {
		return ""
	}
	extras, _ := structpb.NewStruct(map[string]any{events.ExtraEntityType: events.EntityTypePrompt})
	resp, err := h.statusReader.Status(ctx, &commonv1.StatusRequest{
		Id:     promptID,
		Extras: extras,
	})
	if err != nil {
		util.Log(ctx).WithError(err).WithField("prompt_id", promptID).
			Debug("no prior prompt status found before STK push")
		return ""
	}
	if f, ok := resp.GetExtras().GetFields()[client.ExtraCheckoutRequestID]; ok {
		return f.GetStringValue()
	}
	return ""
}
