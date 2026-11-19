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

	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/credentials"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	reporter
	credsResolver *credentials.Resolver
	pawapayCli    client.PawapayClient
}

// NewPaymentHandler creates the queue worker for payout disbursements. The
// payout id is derived from the payment id, so a redelivered payout finds
// the one pawaPay already holds instead of paying twice.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	pawapayCli client.PawapayClient,
	credsResolver *credentials.Resolver,
) queue.SubscribeWorker {
	return &paymentHandler{
		reporter: reporter{
			emitter:    events.NewStatusEmitter(eventsMan),
			metrics:    integrationobs.NewMetrics(providerName),
			entityType: events.EntityTypePayment,
			kind:       "payout",
		},
		credsResolver: credsResolver,
		pawapayCli:    pawapayCli,
	}
}

// lookup reports the payout pawaPay holds, or found=false when it has none.
func (h *paymentHandler) lookup(
	ctx context.Context,
	creds *client.Credentials,
	paymentID, payoutID string,
) (bool, error) {
	result, err := h.pawapayCli.GetPayout(ctx, creds, payoutID)
	if err != nil {
		return false, err
	}
	if result.Status != client.SearchStatusFound || result.Data == nil {
		return false, nil
	}
	p := result.Data
	provider := ""
	if p.Recipient != nil {
		provider = p.Recipient.AccountDetails.Provider
	}
	return true, h.record(ctx, paymentID, p.PayoutID, p.Status, provider, p.Created, p.FailureReason)
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "pawapay.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, h.entityType, "unmarshal_error")
		return nil // non-retriable
	}

	paymentID := payment.GetId()
	payoutID := paymentUUID(paymentID)
	logger = logger.WithField("payment_id", paymentID).WithField("payout_id", payoutID)

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, paymentID, "", "credentials_error", map[string]any{"error": err.Error()})
	}

	// A redelivered payout answers with the one already sent.
	if found, lookErr := h.lookup(ctx, creds, paymentID, payoutID); lookErr != nil {
		logger.WithError(lookErr).Warn("payout lookup failed, sending")
	} else if found {
		logger.Debug("payout already sent")
		return nil
	}

	phoneNumber := payment.GetRecipient().GetContactId()
	extraProvider := extraString(payment.GetExtra(), "provider")
	provider, phoneNumber, err := resolveProvider(ctx, h.pawapayCli, creds, extraProvider, phoneNumber)
	if err != nil {
		if collection.IsUncertain(err) {
			return fmt.Errorf("payout %s provider prediction: %w", paymentID, err)
		}
		logger.WithError(err).Error("failed to resolve provider")
		return h.fail(ctx, paymentID, "", "provider_error", map[string]any{"error": err.Error()})
	}

	payoutReq := &client.PayoutRequest{
		PayoutID:          payoutID,
		Amount:            formatMoneyAmount(payment.GetAmount()),
		Currency:          payment.GetAmount().GetCurrencyCode(),
		PhoneNumber:       phoneNumber,
		Provider:          provider,
		ClientReferenceID: paymentID,
		CustomerMessage:   sanitizeCustomerMessage(extraString(payment.GetExtra(), "customer_message")),
		Metadata:          paymentMetadata(events.EntityTypePayment, paymentID, payment.GetAmount(), headers),
	}

	resp, err := h.pawapayCli.InitiatePayout(ctx, creds, payoutReq)
	switch {
	case err != nil && !collection.IsUncertain(err):
		logger.WithError(err).Error("payout refused by pawaPay")
		return h.fail(ctx, paymentID, "", "provider_error", map[string]any{"error": err.Error()})
	case err != nil || resp.Status == client.InitiationStatusDuplicateIgnored:
		// Unknown outcome or already held: report what pawaPay holds, or
		// redeliver until it can be found or is refused.
		found, lookErr := h.lookup(ctx, creds, paymentID, payoutID)
		if lookErr != nil || !found {
			if err == nil {
				err = fmt.Errorf("payout %s duplicate but not found", payoutID)
			}
			logger.WithError(err).Warn("payout outcome unknown; redelivering")
			return fmt.Errorf("payout %s outcome unknown: %w", paymentID, err)
		}
		return nil
	case resp.Status == client.InitiationStatusRejected:
		logger.WithField("failure_reason", resp.FailureReason).Error("payout rejected")
		return h.fail(ctx, paymentID, payoutID, "rejected", map[string]any{
			"failure_code":    failureCode(resp.FailureReason),
			"failure_message": failureMessage(resp.FailureReason),
		})
	}

	logger.WithField("status", resp.Status).Debug("payout initiated")
	return h.record(ctx, paymentID, payoutID, resp.Status, provider, resp.Created, nil)
}
