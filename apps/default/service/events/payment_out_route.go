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
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
)

type PaymentOutRoute struct {
	qMan        queue.Manager
	eventMan    events.Manager
	paymentRepo repository.PaymentRepository
	routeRepo   repository.RouteRepository
	statusRepo  repository.StatusRepository
}

// NewPaymentOutRoute creates a new PaymentOutRoute event handler with the required dependencies.
func NewPaymentOutRoute(
	qMan queue.Manager,
	eventMan events.Manager,
	paymentRepo repository.PaymentRepository,
	routeRepo repository.RouteRepository,
	statusRepo repository.StatusRepository,
) *PaymentOutRoute {
	return &PaymentOutRoute{
		qMan:        qMan,
		eventMan:    eventMan,
		paymentRepo: paymentRepo,
		routeRepo:   routeRepo,
		statusRepo:  statusRepo,
	}
}

func (event *PaymentOutRoute) Name() string {
	return EventNamePaymentOutRoute
}

func (event *PaymentOutRoute) PayloadType() any {
	pType := ""
	return &pType
}

func (event *PaymentOutRoute) Validate(_ context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New(" payload is not of type string")
	}

	return nil
}

func (event *PaymentOutRoute) Execute(ctx context.Context, payload any) error {
	paymentPtr, ok := payload.(*string)
	if !ok {
		return errors.New("payload is not of type *string")
	}
	if paymentPtr == nil {
		return errors.New("payload is nil")
	}
	paymentID := *paymentPtr

	logger := util.Log(ctx).WithFields(map[string]any{"payment_id": paymentID, "type": event.Name()})
	logger.Debug("handling event")

	p, err := event.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		logger.WithError(err).Warn("could not get payment from db")
		return err
	}

	route, err := routePayment(ctx, event.routeRepo, models.RouteModeTransmit, p)
	if err != nil {
		logger.WithError(err).Error("could not route payment")

		if strings.Contains(err.Error(), "no routes matched for payment") {
			status := models.Status{
				EntityID:   p.GetID(),
				EntityType: "payment",
				State:      int32(commonv1.STATE_INACTIVE),
				Status:     int32(commonv1.STATUS_FAILED),
				Extra:      data.JSONMap{"error": err.Error()},
			}
			status.GenID(ctx)

			err = event.eventMan.Emit(ctx, EventNameStatusSave, &status)
			if err != nil {
				logger.WithError(err).Warn("could not emit status for save")
				return err
			}
			return nil
		}

		return err
	}

	p.RouteID = route.ID
	_, err = event.paymentRepo.Update(ctx, p, "route_id")
	if err != nil {
		logger.WithError(err).Warn("could not save routed payment to db")
		return err
	}

	evt := PaymentOutQueue{}
	err = event.eventMan.Emit(ctx, evt.Name(), p.GetID())
	if err != nil {
		logger.WithError(err).Warn("could not queue out payment")
		return err
	}

	status := models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_ACTIVE),
		Status:     int32(commonv1.STATUS_QUEUED),
		Extra:      make(map[string]interface{}),
	}
	status.GenID(ctx)

	err = event.eventMan.Emit(ctx, EventNameStatusSave, &status)
	if err != nil {
		logger.WithError(err).Warn("could not emit status for save")
		return err
	}

	return nil
}
