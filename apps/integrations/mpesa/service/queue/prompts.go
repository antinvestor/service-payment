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
	"net/url"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type promptHandler struct {
	credentialResolver
	statusEmitter
	mpesaCli     client.MpesaClient
	statusReader PromptStatusReader
	metrics      *integrationobs.Metrics
}

// PromptStatusReader reads a prompt's latest status from the payment service
// (satisfied by paymentv1connect.PaymentServiceClient).
type PromptStatusReader interface {
	Status(
		ctx context.Context,
		req *connect.Request[commonv1.StatusRequest],
	) (*connect.Response[commonv1.StatusResponse], error)
}

// NewPromptHandler creates a queue worker for handling STK Push prompt requests.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	mpesaCli client.MpesaClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MpesaConfig,
	statusReader PromptStatusReader,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		mpesaCli:           mpesaCli,
		statusReader:       statusReader,
		metrics:            integrationobs.NewMetrics("mpesa"),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mpesa.prompt")
	defer logger.Release()
	logger.Debug("queue handler started")

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
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
		h.metrics.QueueProcessed(ctx, "prompt")
		return nil
	}

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "prompt", "credentials_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}

	// Daraja only charges KES; pushing a USD amount as KES would let a
	// session complete for the wrong currency.
	if cur := strings.ToUpper(strings.TrimSpace(prompt.GetAmount().GetCurrencyCode())); cur != "" && cur != mpesaCurrency {
		logger.WithField("currency", cur).Error("unsupported prompt currency for M-Pesa")
		h.metrics.QueueFailed(ctx, "prompt", "unsupported_currency")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       "M-Pesa only supports " + mpesaCurrency + ", got " + cur,
			"entity_type": "prompt",
		})
		return nil
	}

	amount := formatMoneyAmount(prompt.GetAmount())

	timestamp := time.Now().Format("20060102150405")
	password := client.STKPassword(creds.Shortcode, creds.Passkey, timestamp)

	accountRef := promptID
	if prompt.GetExtra() != nil {
		if ref, ok := prompt.GetExtra().GetFields()["account_reference"]; ok {
			accountRef = ref.GetStringValue()
		}
	}

	callbackURL := stkCallbackURL(creds.CallbackURL, headers, promptID)

	stkReq := &client.STKPushRequest{
		BusinessShortCode: creds.Shortcode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            phoneNumber,
		PartyB:            creds.Shortcode,
		PhoneNumber:       phoneNumber,
		CallBackURL:       callbackURL,
		AccountReference:  accountRef,
		TransactionDesc:   "Payment prompt",
	}

	resp, err := h.mpesaCli.STKPush(ctx, creds, stkReq)
	if err != nil {
		logger.WithError(err).Error("STK push failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	logger.WithField("checkout_request_id", resp.CheckoutRequestID).Debug("STK push initiated")

	// checkout_request_id binds this prompt to the Daraja request so the STK
	// callback can prove it belongs here; requested_amount and the credential
	// connection let it verify amount and re-query Daraja.
	inProcessExtras := map[string]any{
		"merchant_request_id":         resp.MerchantRequestID,
		client.ExtraCheckoutRequestID: resp.CheckoutRequestID,
		client.ExtraRequestedAmount:   amount,
		"response_code":               resp.ResponseCode,
		"customer_message":            resp.CustomerMessage,
		"entity_type":                 "prompt",
	}
	if connection := headers[config.HeaderConnectionCredentials]; connection != "" {
		inProcessExtras[client.ExtraCredentialsConnection] = connection
	}
	h.emitStatus(ctx, promptID, resp.CheckoutRequestID, commonv1.STATUS_IN_PROCESS, inProcessExtras)

	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

// mpesaCurrency is the only currency Daraja STK Push collects.
const mpesaCurrency = "KES"

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
	extras, _ := structpb.NewStruct(map[string]any{"entity_type": "prompt"})
	resp, err := h.statusReader.Status(ctx, connect.NewRequest(&commonv1.StatusRequest{
		Id:     promptID,
		Extras: extras,
	}))
	if err != nil {
		util.Log(ctx).WithError(err).WithField("prompt_id", promptID).
			Debug("no prior prompt status found before STK push")
		return ""
	}
	if f, ok := resp.Msg.GetExtras().GetFields()[client.ExtraCheckoutRequestID]; ok {
		return f.GetStringValue()
	}
	return ""
}
