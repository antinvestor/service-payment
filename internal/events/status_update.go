package events

import (
	"context"
	"errors"
	"log/slog"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/pitabwire/util"
)

type PaymentStatusUpdate struct {
	PaymentCli paymentv1connect.PaymentServiceClient
}

// NewPaymentStatusUpdate creates a new PaymentStatusUpdate event handler.
func NewPaymentStatusUpdate(
	_ context.Context,
	paymentCli paymentv1connect.PaymentServiceClient,
) *PaymentStatusUpdate {
	return &PaymentStatusUpdate{
		PaymentCli: paymentCli,
	}
}

func (e *PaymentStatusUpdate) Name() string {
	return PaymentStatusUpdateEvent
}

func (e *PaymentStatusUpdate) PayloadType() any {
	return &commonv1.StatusUpdateRequest{}
}

func (e *PaymentStatusUpdate) Validate(_ context.Context, payload any) error {
	statusUpdateRequest, ok := payload.(*commonv1.StatusUpdateRequest)
	if !ok {
		return errors.New("payload is not of type *commonv1.StatusUpdateRequest")
	}

	if statusUpdateRequest.GetId() == "" {
		return errors.New("statusUpdateRequest Id should already have been set")
	}

	return nil
}

func (e *PaymentStatusUpdate) Execute(ctx context.Context, payload any) error {
	statusUpdateRequest, _ := payload.(*commonv1.StatusUpdateRequest)

	logger := util.Log(ctx).WithField("type", e.Name()).WithField("payment_id", statusUpdateRequest.GetId())
	defer logger.Release()

	logger.Debug("event handler started")

	if logger.Enabled(ctx, slog.LevelDebug) {
		logger.WithField("payload", statusUpdateRequest).Debug("processing status update request")
	}

	_, err := e.PaymentCli.StatusUpdate(ctx, connect.NewRequest(statusUpdateRequest))
	if err != nil {
		logger.WithError(err).Warn("could not update status")
		return nil
	}

	logger.Debug("event handler completed successfully")
	return nil
}
