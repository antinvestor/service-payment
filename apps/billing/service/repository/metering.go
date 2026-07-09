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

//nolint:dupl // Structurally identical to rating.go but type-specific; Go generics cannot abstract this.
package repository

import (
	"context"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

// MeteredUsageRepository provides operations for metered usage records.
type MeteredUsageRepository interface {
	datastore.BaseRepository[*models.MeteredUsage]
	ListByBillingRunID(ctx context.Context, billingRunID string) ([]*models.MeteredUsage, error)
}

type meteredUsageRepository struct {
	datastore.BaseRepository[*models.MeteredUsage]
}

func NewMeteredUsageRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) MeteredUsageRepository {
	return &meteredUsageRepository{
		BaseRepository: datastore.NewBaseRepository[*models.MeteredUsage](
			ctx, dbPool, workMan, func() *models.MeteredUsage { return &models.MeteredUsage{} },
		),
	}
}

func (r *meteredUsageRepository) ListByBillingRunID(
	ctx context.Context,
	billingRunID string,
) ([]*models.MeteredUsage, error) {
	if billingRunID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.MeteredUsage
	result := r.Pool().DB(ctx, true).Where("billing_run_id = ?", billingRunID).Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}
