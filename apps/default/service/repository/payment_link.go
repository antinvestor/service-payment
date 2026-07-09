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
	"strings"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

type paymentLinkRepository struct {
	datastore.BaseRepository[*models.PaymentLink]
}

func NewPaymentLinkRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) PaymentLinkRepository {
	repo := paymentLinkRepository{
		BaseRepository: datastore.NewBaseRepository[*models.PaymentLink](
			ctx, dbPool, workMan, func() *models.PaymentLink { return &models.PaymentLink{} },
		),
	}
	return &repo
}

func (plr *paymentLinkRepository) GetByPartitionAndID(
	ctx context.Context,
	partitionID string,
	id string,
) (*models.PaymentLink, error) {
	link := &models.PaymentLink{}
	err := plr.Pool().DB(ctx, true).First(link, "partition_id = ? AND id = ?", partitionID, id).Error
	return link, err
}

func (plr *paymentLinkRepository) GetByProfileID(_ context.Context, _ string) ([]*models.PaymentLink, error) {
	// PaymentLink has no direct profile association — return empty rather than scanning entire table
	return nil, nil
}

// Legacy method for backward compatibility.
func (plr *paymentLinkRepository) SearchLegacy(ctx context.Context, query string) ([]*models.PaymentLink, error) {
	var links []*models.PaymentLink
	err := plr.Pool().DB(ctx, true).Where("name ILIKE ?", "%"+strings.ToLower(query)+"%").Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}
