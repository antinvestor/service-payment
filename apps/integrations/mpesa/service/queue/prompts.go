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
	"encoding/base64"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type promptHandler struct {
	credentialResolver
	statusEmitter
	mpesaCli client.MpesaClient
	metrics  *integrationobs.Metrics
}

// NewPromptHandler creates a queue worker for handling STK Push prompt requests.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	mpesaCli client.MpesaClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MpesaConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		mpesaCli:           mpesaCli,
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

	amount := formatMoneyAmount(prompt.GetAmount())

	timestamp := time.Now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString(
		[]byte(creds.Shortcode + creds.Passkey + timestamp),
	)

	accountRef := promptID
	if prompt.GetExtra() != nil {
		if ref, ok := prompt.GetExtra().GetFields()["account_reference"]; ok {
			accountRef = ref.GetStringValue()
		}
	}

	callbackURL := appendTenantParams(creds.CallbackURL+"/webhook/mpesa/stk", headers)

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

	h.emitStatus(ctx, promptID, resp.CheckoutRequestID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"merchant_request_id": resp.MerchantRequestID,
		"checkout_request_id": resp.CheckoutRequestID,
		"response_code":       resp.ResponseCode,
		"customer_message":    resp.CustomerMessage,
		"entity_type":         "prompt",
	})

	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}
