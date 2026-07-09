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
	"math"
	"net/url"
	"strconv"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	credentialResolver
	statusEmitter
	mpesaCli client.MpesaClient
	metrics  *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for handling B2C disbursement payments.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	mpesaCli client.MpesaClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MpesaConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		mpesaCli:           mpesaCli,
		metrics:            integrationobs.NewMetrics("mpesa"),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mpesa.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil // non-retriable
	}

	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "payment", "credentials_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	phoneNumber := payment.GetRecipient().GetContactId()
	amount := formatMoneyAmount(payment.GetAmount())

	commandID := "BusinessPayment"
	if payment.GetExtra() != nil {
		if cmd, ok := payment.GetExtra().GetFields()["command_id"]; ok {
			commandID = cmd.GetStringValue()
		}
	}

	b2cResultURL := appendTenantParams(creds.CallbackURL+"/webhook/mpesa/b2c", headers)
	b2cTimeoutURL := appendTenantParams(creds.CallbackURL+"/webhook/mpesa/b2c/timeout", headers)

	b2cReq := &client.B2CRequest{
		OriginatorConversationID: paymentID,
		InitiatorName:            creds.InitiatorName,
		SecurityCredential:       creds.SecurityCredential,
		CommandID:                commandID,
		Amount:                   amount,
		PartyA:                   creds.Shortcode,
		PartyB:                   phoneNumber,
		Remarks:                  "Payment disbursement",
		QueueTimeOutURL:          b2cTimeoutURL,
		ResultURL:                b2cResultURL,
		Occasion:                 paymentID,
	}

	resp, err := h.mpesaCli.B2CPayment(ctx, creds, b2cReq)
	if err != nil {
		logger.WithError(err).Error("B2C payment failed")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	logger.WithField("conversation_id", resp.ConversationID).Debug("B2C payment initiated")

	h.emitStatus(ctx, paymentID, resp.ConversationID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"conversation_id":            resp.ConversationID,
		"originator_conversation_id": resp.OriginatorConversationID,
		"response_code":              resp.ResponseCode,
		"entity_type":                "payment",
	})

	h.metrics.QueueProcessed(ctx, "payment")
	return nil
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

// formatMoneyAmount converts a protobuf Money value to a string suitable for M-Pesa (whole KES).
func formatMoneyAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
}) string {
	if amount == nil {
		return "0"
	}
	units := amount.GetUnits()
	nanos := amount.GetNanos()
	total := float64(units) + float64(nanos)/1e9 //nolint:mnd // standard nanos-to-unit conversion
	return strconv.FormatInt(int64(math.Round(total)), 10)
}
