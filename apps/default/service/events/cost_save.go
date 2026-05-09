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

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util"
)

type CostSave struct {
	costRepo repository.CostRepository
}

// NewCostSave creates a new CostSave event handler with the required dependencies.
func NewCostSave(costRepo repository.CostRepository) *CostSave {
	return &CostSave{
		costRepo: costRepo,
	}
}

func (e *CostSave) Name() string {
	return EventNameCostSave
}

func (e *CostSave) PayloadType() any {
	return &models.Cost{}
}

func (e *CostSave) Validate(_ context.Context, payload any) error {
	cost, ok := payload.(*models.Cost)
	if !ok {
		return errors.New("payload is not of type models.Cost")
	}

	if cost.ID == "" {
		return errors.New("cost ID should already have been set")
	}

	return nil
}

func (e *CostSave) Execute(ctx context.Context, payload any) error {
	cost, ok := payload.(*models.Cost)
	if !ok {
		return errors.New("payload is not of type models.Cost")
	}

	logger := util.Log(ctx).WithFields(map[string]any{"cost_id": cost.ID, "type": e.Name()})
	logger.Debug("handling event")

	err := e.costRepo.Create(ctx, cost)
	if err != nil {
		if data.ErrorIsDuplicateKey(err) {
			logger.Debug("record already exists, skipping duplicate")
			return nil
		}
		logger.WithError(err).Error("could not save cost to db")
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
