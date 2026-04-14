package queue

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type paymentHandler struct {
	eventsMan frameEvents.Manager
}

// NewPaymentHandler creates a queue worker for the payment queue.
// Polar.sh does not support disbursements, so this handler returns a failure status.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
) queue.SubscribeWorker {
	return &paymentHandler{
		eventsMan: eventsMan,
	}
}

func (h *paymentHandler) Handle(ctx context.Context, _ map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "polar.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		return nil
	}

	paymentID := payment.GetId()
	logger.WithField("payment_id", paymentID).Warn("Polar does not support disbursements")

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
