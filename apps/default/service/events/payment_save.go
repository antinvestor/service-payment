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
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/util"
)

type PaymentSave struct {
	paymentRepo repository.PaymentRepository
	eventMan    events.Manager
}

// NewPaymentSave creates a new PaymentSave event handler with the required dependencies.
func NewPaymentSave(paymentRepo repository.PaymentRepository, eventMan events.Manager) *PaymentSave {
	return &PaymentSave{
		paymentRepo: paymentRepo,
		eventMan:    eventMan,
	}
}

func (event *PaymentSave) Name() string {
	return EventNamePaymentSave
}

func (event *PaymentSave) PayloadType() any {
	return &models.Payment{}
}

func (event *PaymentSave) Validate(_ context.Context, payload any) error {
	payment, ok := payload.(*models.Payment)
	if !ok {
		return errors.New(" payload is not of type models.Payment")
	}

	if payment.GetID() == "" {
		return errors.New(" payment Id should already have been set ")
	}

	return nil
}

func (event *PaymentSave) Execute(ctx context.Context, payload any) error {
	payment, ok := payload.(*models.Payment)
	if !ok {
		return errors.New("payload is not of type models.Payment")
	}

	logger := util.Log(ctx).WithFields(map[string]any{"payment_id": payment.GetID(), "type": event.Name()})
	logger.Debug("handling event")

	err := event.paymentRepo.Create(ctx, payment)
	if err != nil {
		if data.ErrorIsDuplicateKey(err) {
			logger.Debug("record already exists, skipping duplicate")
			return nil
		}
		logger.WithError(err).Warn("could not save to db")
		return err
	}
	logger.Debug("successfully saved record to db")

	if !payment.OutBound {
		// Emit payment in route event
		err = event.eventMan.Emit(ctx, EventNamePaymentInRoute, payment.GetID())
		if err != nil {
			return err
		}

		return nil
	}

	if payment.IsReleased() {
		// Emit payment out route event
		err = event.eventMan.Emit(ctx, EventNamePaymentOutRoute, payment.GetID())
		if err != nil {
			logger.WithError(err).Warn("could not emit for queue out")
			return err
		}
	} else {
		status := &models.Status{
			EntityID:   payment.GetID(),
			EntityType: "payment",
			State:      int32(commonv1.STATE_CHECKED.Number()),
			Status:     int32(commonv1.STATUS_QUEUED.Number()),
			Extra:      make(map[string]interface{}),
		}
		status.GenID(ctx)
		// Emit status save event
		err = event.eventMan.Emit(ctx, EventNameStatusSave, status)
		if err != nil {
			logger.WithError(err).Warn("could not emit status")
			return err
		}
	}
	return nil
}
