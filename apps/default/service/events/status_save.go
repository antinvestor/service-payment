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

	"github.com/antinvestor/service-payments/apps/default/service/metrics"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util"
)

const entityTypePayment = "payment"

type StatusSave struct {
	statusRepo     repository.StatusRepository
	paymentRepo    repository.PaymentRepository
	paymentMetrics *metrics.PaymentMetrics
}

// NewStatusSave creates a new StatusSave event handler with the required dependencies.
func NewStatusSave(
	statusRepo repository.StatusRepository,
	paymentRepo repository.PaymentRepository,
) *StatusSave {
	return &StatusSave{
		statusRepo:     statusRepo,
		paymentRepo:    paymentRepo,
		paymentMetrics: metrics.NewPaymentMetrics(),
	}
}

func (e *StatusSave) Name() string {
	return EventNameStatusSave
}

func (e *StatusSave) PayloadType() any {
	return &models.Status{}
}

func (e *StatusSave) Validate(_ context.Context, payload any) error {
	status, ok := payload.(*models.Status)
	if !ok {
		return errors.New("payload is not of type models.Status")
	}
	if status.GetID() == "" {
		return errors.New("status Id should already have been set")
	}
	return nil
}

func (e *StatusSave) Execute(ctx context.Context, payload any) error {
	status, ok := payload.(*models.Status)
	if !ok {
		return errors.New("payload is not of type models.Status")
	}

	logger := util.Log(ctx).WithFields(map[string]any{
		"entity_id":   status.EntityID,
		"entity_type": status.EntityType,
		"type":        e.Name(),
	})
	logger.Debug("handling event")

	err := e.statusRepo.Create(ctx, status)
	if err != nil {
		if data.ErrorIsDuplicateKey(err) {
			logger.Debug("record already exists, skipping duplicate")
			return nil
		}
		logger.WithError(err).Warn("could not save status to db")
		return err
	}
	logger.Debug("successfully saved record to db")

	e.recordPaymentMetrics(ctx, status)

	return nil
}

// recordPaymentMetrics emits business metrics for payments reaching a
// terminal status. All payment status transitions funnel through this event
// handler, so terminal outcomes from every path are captured here.
func (e *StatusSave) recordPaymentMetrics(ctx context.Context, status *models.Status) {
	if status.EntityType != entityTypePayment || !metrics.IsTerminalStatus(status.Status) {
		return
	}

	p, err := e.paymentRepo.GetByID(ctx, status.EntityID)
	if err != nil {
		util.Log(ctx).WithError(err).
			WithField("entity_id", status.EntityID).
			Warn("could not load payment for terminal status metrics")
		return
	}

	e.paymentMetrics.RecordTerminal(ctx, p, status)
}
