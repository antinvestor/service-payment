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

	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

// failurePayoutUnsupported is reported for every payout: the Jenga client
// only offers tills pay (a payment to a merchant till), which is not a
// payout to a person's mobile money or bank account.
const (
	failurePayoutUnsupported        = "PAYOUT_NOT_SUPPORTED"
	failurePayoutUnsupportedMessage = "Payouts are not supported on the Jenga route; only STK collection prompts are"
)

type paymentHandler struct {
	emitter *events.StatusEmitter
	metrics *integrationobs.Metrics
}

// NewPaymentHandler creates the queue worker for payouts. Jenga's tills pay
// endpoint pays a merchant till, not a recipient contact, so until a proper
// send-money endpoint is wired every payout is failed with a clear reason
// rather than paying a till by mistake.
func NewPaymentHandler(eventsMan frameEvents.Manager) queue.SubscribeWorker {
	return &paymentHandler{
		emitter: events.NewStatusEmitter(eventsMan),
		metrics: integrationobs.NewMetrics(providerName),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, _ map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("handler", "jenga.payment")
	defer logger.Release()

	payment := &paymentv1.Payment{}
	if err := proto.Unmarshal(payload, payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil // non-retriable
	}
	paymentID := payment.GetId()
	logger.WithField("payment_id", paymentID).Error("payout requested on the Jenga route, which has no payout endpoint")

	h.metrics.QueueFailed(ctx, "payment", "rejected")
	return h.emitter.Fail(ctx, events.EntityTypePayment, paymentID, failurePayoutUnsupportedMessage, map[string]any{
		collection.ExtraProvider: providerName,
		"failure_code":           failurePayoutUnsupported,
		"failure_message":        failurePayoutUnsupportedMessage,
	})
}
