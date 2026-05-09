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

package events

import (
	"context"
	"errors"
	"log/slog"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
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

	logger := util.Log(ctx).WithFields(map[string]any{"type": e.Name(), "payment_id": statusUpdateRequest.GetId()})
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
