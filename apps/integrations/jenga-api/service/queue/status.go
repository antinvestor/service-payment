package queue

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

// statusEmitter provides shared status emission logic for queue handlers.
type statusEmitter struct {
	eventsMan frameEvents.Manager
}

func (e *statusEmitter) emitStatus(
	ctx context.Context,
	id, externalID string,
	status commonv1.STATUS,
	extras map[string]any,
) {
	extra, _ := structpb.NewStruct(extras)
	err := e.eventsMan.Emit(ctx, events.PaymentStatusUpdateEvent, &commonv1.StatusUpdateRequest{
		Id:         id,
		State:      commonv1.STATE_ACTIVE,
		Status:     status,
		ExternalId: externalID,
		Extras:     extra,
	})
	if err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit status update")
	}
}
