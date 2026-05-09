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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type PaymentOutQueue struct {
	qMan        queue.Manager
	eventMan    events.Manager
	paymentRepo repository.PaymentRepository
	statusRepo  repository.StatusRepository
}

// NewPaymentOutQueue creates a new PaymentOutQueue event handler with the required dependencies.
func NewPaymentOutQueue(
	qMan queue.Manager,
	eventMan events.Manager,
	paymentRepo repository.PaymentRepository,
	statusRepo repository.StatusRepository,
) *PaymentOutQueue {
	return &PaymentOutQueue{
		qMan:        qMan,
		eventMan:    eventMan,
		paymentRepo: paymentRepo,
		statusRepo:  statusRepo,
	}
}

func (event *PaymentOutQueue) Name() string {
	return EventNamePaymentOutQueue
}

func (event *PaymentOutQueue) PayloadType() any {
	pType := ""
	return &pType
}

func (event *PaymentOutQueue) Validate(_ context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New("payload is not of type string")
	}
	return nil
}

func (event *PaymentOutQueue) Execute(ctx context.Context, payload any) error {
	paymentIDPtr, ok := payload.(*string)
	if !ok {
		return errors.New("payload is not of type *string")
	}
	if paymentIDPtr == nil {
		return errors.New("payload is nil")
	}
	paymentID := *paymentIDPtr

	logger := util.Log(ctx).WithFields(map[string]any{"payment_id": paymentID, "type": event.Name()})
	logger.Debug("handling payment event")

	// Fetch payment record by ID
	payment, err := event.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Fetch payment status
	status, err := event.statusRepo.GetByEntity(ctx, payment.ID, "payment")
	if err != nil {
		logger.WithError(err).Warn("could not get payment status")
		return err
	}

	apiPayment := payment.ToAPI(status, nil)

	binaryProto, err := proto.Marshal(apiPayment)
	if err != nil {
		return err
	}

	// Publish the payment message for further processing
	err = event.qMan.Publish(ctx, payment.RouteID, binaryProto)
	if err != nil {
		return err
	}

	logger.WithField("route_id", payment.RouteID).Debug("payment message successfully queued")

	// Update payment status using unified Status
	status = &models.Status{
		EntityID:   payment.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_ACTIVE),
		Status:     int32(commonv1.STATUS_IN_PROCESS),
		Extra:      make(map[string]interface{}),
	}
	status.GenID(ctx)

	// Emit status event
	err = event.eventMan.Emit(ctx, EventNameStatusSave, status)
	if err != nil {
		return err
	}

	return nil
}
