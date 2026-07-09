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

package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

type costRepository struct {
	datastore.BaseRepository[*models.Cost]
}

func NewCostRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) CostRepository {
	repo := costRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Cost](
			ctx, dbPool, workMan, func() *models.Cost { return &models.Cost{} },
		),
	}
	return &repo
}

func (cr *costRepository) GetByPaymentID(ctx context.Context, paymentID string) ([]*models.Cost, error) {
	var costs []*models.Cost
	err := cr.Pool().DB(ctx, true).Where("payment_id = ?", paymentID).Find(&costs).Error
	return costs, err
}
