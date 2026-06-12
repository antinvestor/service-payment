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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/config"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/credentials"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	statusEmitter
	credsResolver *credentials.Resolver
	pawapayCli    client.PawapayClient
	metrics       *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for handling payout disbursements.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	pawapayCli client.PawapayClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.PawapayConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		statusEmitter: statusEmitter{eventsMan: eventsMan},
		credsResolver: credentials.NewResolver(settingsCli, cfg),
		pawapayCli:    pawapayCli,
		metrics:       integrationobs.NewMetrics("pawapay"),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "pawapay.payment")
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

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
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
	extraProvider := extraString(payment.GetExtra(), "provider")

	provider, phoneNumber, err := resolveProvider(ctx, h.pawapayCli, creds, extraProvider, phoneNumber)
	if err != nil {
		logger.WithError(err).Error("failed to resolve provider")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	payoutID := paymentUUID(paymentID)

	payoutReq := &client.PayoutRequest{
		PayoutID:          payoutID,
		Amount:            formatMoneyAmount(payment.GetAmount()),
		Currency:          payment.GetAmount().GetCurrencyCode(),
		PhoneNumber:       phoneNumber,
		Provider:          provider,
		ClientReferenceID: paymentID,
		CustomerMessage:   sanitizeCustomerMessage(extraString(payment.GetExtra(), "customer_message")),
		Metadata:          paymentMetadata("payment", paymentID, headers),
	}

	resp, err := h.pawapayCli.InitiatePayout(ctx, creds, payoutReq)
	if err != nil {
		logger.WithError(err).Error("payout initiation failed")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	if resp.Status == client.InitiationStatusRejected {
		logger.WithField("failure_reason", resp.FailureReason).Error("payout rejected")
		h.metrics.QueueFailed(ctx, "payment", "rejected")
		h.emitStatus(ctx, paymentID, payoutID, commonv1.STATUS_FAILED, map[string]any{
			"payout_id":       payoutID,
			"failure_code":    failureCode(resp.FailureReason),
			"failure_message": failureMessage(resp.FailureReason),
			"entity_type":     "payment",
		})
		return nil
	}

	logger.WithField("payout_id", payoutID).WithField("status", resp.Status).Debug("payout initiated")

	h.metrics.QueueProcessed(ctx, "payment")
	h.emitStatus(ctx, paymentID, payoutID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"payout_id":   payoutID,
		"provider":    provider,
		"status":      resp.Status,
		"created":     resp.Created,
		"entity_type": "payment",
	})

	return nil
}
