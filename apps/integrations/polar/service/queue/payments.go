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
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type paymentHandler struct {
	eventsMan frameEvents.Manager
	metrics   *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for the payment queue.
// Polar.sh does not support disbursements, so this handler returns a failure status.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
) queue.SubscribeWorker {
	return &paymentHandler{
		eventsMan: eventsMan,
		metrics:   integrationobs.NewMetrics("polar"),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, _ map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "polar.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil
	}

	paymentID := payment.GetId()
	logger.WithField("payment_id", paymentID).Warn("Polar does not support disbursements")

	h.metrics.QueueFailed(ctx, "payment", "rejected")

	// Polar doesn't support disbursements — report failure
	extra, _ := structpb.NewStruct(map[string]any{
		"error":       "Polar.sh does not support disbursement payments",
		"entity_type": "payment",
	})
	err := h.eventsMan.Emit(ctx, events.PaymentStatusUpdateEvent, &commonv1.StatusUpdateRequest{
		Id:     paymentID,
		State:  commonv1.STATE_INACTIVE,
		Status: commonv1.STATUS_FAILED,
		Extras: extra,
	})
	if err != nil {
		logger.WithError(err).Warn("could not emit status update")
	}

	return nil
}
