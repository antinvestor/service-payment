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
	"fmt"

	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

type linkRepository struct {
	datastore.BaseRepository[*models.CheckoutLink]
}

// NewLinkRepository returns a LinkRepository backed by the given pool.
func NewLinkRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) LinkRepository {
	return &linkRepository{
		BaseRepository: datastore.NewBaseRepository[*models.CheckoutLink](
			ctx, dbPool, workMan, func() *models.CheckoutLink { return &models.CheckoutLink{} },
		),
	}
}

// GetByRef fetches a CheckoutLink by its unique ref using the primary DB
// (read-modify-write; replica lag can cause silent NotFound).
func (r *linkRepository) GetByRef(ctx context.Context, ref string) (*models.CheckoutLink, error) {
	var link models.CheckoutLink
	err := r.Pool().DB(ctx, false).
		Where("ref = ? AND deleted_at IS NULL", ref).
		First(&link).Error
	if err != nil {
		return nil, fmt.Errorf("get checkout link by ref: %w", err)
	}
	return &link, nil
}
