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

const (
	providerName = "mpesa"
	kindPayment  = "payment"
	kindPrompt   = "prompt"
)

type paymentHandler struct {
	credentialResolver
	emitter      *events.StatusEmitter
	mpesaCli     client.MpesaClient
	statusReader StatusReader
	metrics      *integrationobs.Metrics
}

// NewPaymentHandler creates the queue worker for B2C payouts. statusReader
// (optional, the payment service client) lets a redelivered payout find the
// one already sent: Daraja has no synchronous payout lookup, so the worker
// records the binding (OriginatorConversationID, amount) before it sends.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	mpesaCli client.MpesaClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MpesaConfig,
	statusReader StatusReader,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		mpesaCli:           mpesaCli,
		statusReader:       statusReader,
		metrics:            integrationobs.NewMetrics(providerName),
	}
}

func (h *paymentHandler) fail(ctx context.Context, paymentID, reason, message string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, kindPayment, reason)
	return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID, message, extras)
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mpesa.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, kindPayment, "unmarshal_error")
		return nil // non-retriable
	}
	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	// A redelivered payout must not pay twice: the binding is recorded
	// before the B2C call, so its presence means a request was (possibly)
	// sent. Daraja's result callback settles it.
	if sent, _ := h.storedBinding(ctx, paymentID); sent != "" {
		logger.WithField(client.ExtraOriginatorConversationID, sent).
			Warn("B2C payout already issued for payment; awaiting Daraja result, not resending")
		h.metrics.QueueProcessed(ctx, kindPayment)
		return nil
	}

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, paymentID, "credentials_error", err.Error(), nil)
	}
	if !creds.HasInitiator() {
		logger.Error("B2C credentials lack initiator name or security credential")
		return h.fail(ctx, paymentID, "credentials_error",
			"M-Pesa B2C needs an initiator name and security credential", nil)
	}

	// Daraja only pays KES, in whole shillings.
	currency := strings.ToUpper(strings.TrimSpace(payment.GetAmount().GetCurrencyCode()))
	if currency != "" && currency != client.Currency {
		logger.WithField("currency", currency).Error("unsupported payout currency for M-Pesa")
		return h.fail(ctx, paymentID, "rejected", "M-Pesa only pays "+client.Currency+", got "+currency, nil)
	}
	units, err := collection.WholeUnits(payment.GetAmount())
	if err != nil {
		logger.WithError(err).Error("payout amount cannot be sent to M-Pesa")
		return h.fail(ctx, paymentID, "rejected", "invalid amount: "+err.Error(), nil)
	}
	amount := strconv.FormatInt(units, 10)

	commandID := "BusinessPayment"
	if cmd, ok := payment.GetExtra().GetFields()["command_id"]; ok && cmd.GetStringValue() != "" {
		commandID = cmd.GetStringValue()
	}

	extras := map[string]any{
		client.ExtraOriginatorConversationID: paymentID,
		client.ExtraRequestedAmount:          amount,
		client.ExtraRequestedCurrency:        client.Currency,
		"command_id":                         commandID,
	}
	if connection := headers[config.HeaderConnectionCredentials]; connection != "" {
		extras[client.ExtraCredentialsConnection] = connection
	}

	// Record the binding first: if the process dies or the answer is lost
	// after Daraja accepted the request, a redelivery finds it and does not
	// pay again. A failed emit is returned so the message is redelivered
	// before anything is sent.
	if err = h.emitter.Emit(ctx, events.EntityTypePayment, paymentID, "", commonv1.STATUS_IN_PROCESS,
		copyExtras(extras, map[string]any{"verification": "sending"})); err != nil {
		return err
	}

	b2cReq := &client.B2CRequest{
		OriginatorConversationID: paymentID,
		InitiatorName:            creds.InitiatorName,
		SecurityCredential:       creds.SecurityCredential,
		CommandID:                commandID,
		Amount:                   amount,
		PartyA:                   creds.Shortcode,
		PartyB:                   payment.GetRecipient().GetContactId(),
		Remarks:                  "Payment disbursement",
		QueueTimeOutURL:          b2cCallbackURL(creds.CallbackURL, "/webhook/mpesa/b2c/timeout", headers, paymentID),
		ResultURL:                b2cCallbackURL(creds.CallbackURL, "/webhook/mpesa/b2c", headers, paymentID),
		Occasion:                 paymentID,
	}

	resp, err := h.mpesaCli.B2CPayment(ctx, creds, b2cReq)
	if err != nil {
		if client.IsRejected(err) {
			logger.WithError(err).Error("B2C payout refused by Daraja")
			return h.fail(ctx, paymentID, "provider_error", err.Error(), extras)
		}
		// Unknown outcome (timeout, 5xx): Daraja may have queued the payout
		// under this OriginatorConversationID. The binding recorded above
		// keeps a redelivery from paying again; the result callback (or a
		// reconciliation against Daraja's Transaction Status API) settles it.
		logger.WithError(err).Warn("B2C payout outcome unknown; redelivering")
		h.metrics.QueueFailed(ctx, kindPayment, "provider_error")
		return fmt.Errorf("b2c payout %s outcome unknown: %w", paymentID, err)
	}

	logger.WithField("conversation_id", resp.ConversationID).Debug("B2C payout accepted")
	extras["conversation_id"] = resp.ConversationID
	extras["response_code"] = resp.ResponseCode
	extras["verification"] = "accepted"
	if err = h.emitter.Emit(ctx, events.EntityTypePayment, paymentID, resp.ConversationID, commonv1.STATUS_IN_PROCESS, extras); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, kindPayment)
	return nil
}

// storedBinding returns the OriginatorConversationID already recorded on the
// payment's status, or "" when none (or when no status reader is configured).
func (h *paymentHandler) storedBinding(ctx context.Context, paymentID string) (string, error) {
	if h.statusReader == nil || paymentID == "" {
		return "", nil
	}
	reqExtras, _ := structpb.NewStruct(map[string]any{events.ExtraEntityType: events.EntityTypePayment})
	resp, err := h.statusReader.Status(ctx, &commonv1.StatusRequest{Id: paymentID, Extras: reqExtras})
	if err != nil {
		util.Log(ctx).WithError(err).WithField("payment_id", paymentID).Debug("no prior payment status before B2C payout")
		return "", err
	}
	if f, ok := resp.GetExtras().GetFields()[client.ExtraOriginatorConversationID]; ok {
		return f.GetStringValue(), nil
	}
	return "", nil
}

func copyExtras(base map[string]any, add map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(add))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range add {
		out[k] = v
	}
	return out
}

// b2cCallbackURL builds a B2C result/timeout URL carrying the tenancy and
// the payment id, so the callback can be bound to the payout it answers.
func b2cCallbackURL(base, path string, headers map[string]string, paymentID string) string {
	callbackURL := appendTenantParams(base+path, headers)
	if paymentID == "" {
		return callbackURL
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return callbackURL
	}
	q := u.Query()
	q.Set(client.CallbackParamPaymentID, paymentID)
	u.RawQuery = q.Encode()
	return u.String()
}

// appendTenantParams appends tenant_id and partition_id query params to a callback URL.
func appendTenantParams(callbackURL string, headers map[string]string) string {
	tenantID := headers["tenant_id"]
	partitionID := headers["partition_id"]
	if tenantID == "" && partitionID == "" {
		return callbackURL
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return callbackURL
	}
	q := u.Query()
	if tenantID != "" {
		q.Set("tenant_id", tenantID)
	}
	if partitionID != "" {
		q.Set("partition_id", partitionID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
