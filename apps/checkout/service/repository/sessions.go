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

type sessionRepository struct {
	datastore.BaseRepository[*models.CheckoutSession]
}

// NewSessionRepository returns a SessionRepository backed by the given pool.
func NewSessionRepository(
	ctx context.Context,
	dbPool pool.Pool,
	workMan workerpool.Manager,
) SessionRepository {
	return &sessionRepository{
		BaseRepository: datastore.NewBaseRepository[*models.CheckoutSession](
			ctx, dbPool, workMan, func() *models.CheckoutSession { return &models.CheckoutSession{} },
		),
	}
}

// GetByRef fetches a CheckoutSession by its unique ref using the primary DB
// (read-modify-write; replica lag can cause silent NotFound).
func (r *sessionRepository) GetByRef(ctx context.Context, ref string) (*models.CheckoutSession, error) {
	var session models.CheckoutSession
	err := r.Pool().DB(ctx, false).
		Where("ref = ? AND deleted_at IS NULL", ref).
		First(&session).Error
	if err != nil {
		return nil, fmt.Errorf("get checkout session by ref: %w", err)
	}
	return &session, nil
}

// GetByOrderRef fetches the most recent CheckoutSession for a product order_ref.
func (r *sessionRepository) GetByOrderRef(ctx context.Context, orderRef string) (*models.CheckoutSession, error) {
	var session models.CheckoutSession
	err := r.Pool().DB(ctx, false).
		Where("order_ref = ? AND deleted_at IS NULL", orderRef).
		Order("created_at DESC").
		First(&session).Error
	if err != nil {
		return nil, fmt.Errorf("get checkout session by order_ref: %w", err)
	}
	return &session, nil
}

// ListByStatus returns sessions matching the given status, ordered by
// modified_at ascending (oldest first), using a replica read.
func (r *sessionRepository) ListByStatus(
	ctx context.Context, status string, limit int,
) ([]*models.CheckoutSession, error) {
	var sessions []*models.CheckoutSession
	err := r.Pool().DB(ctx, true).
		Where("status = ? AND deleted_at IS NULL", status).
		Order("modified_at ASC").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("list checkout sessions by status: %w", err)
	}
	return sessions, nil
}
