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
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

type paymentRepository struct {
	datastore.BaseRepository[*models.Payment]
}

func NewPaymentRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) PaymentRepository {
	repo := paymentRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Payment](
			ctx, dbPool, workMan, func() *models.Payment { return &models.Payment{} },
		),
	}
	return &repo
}

func (pr *paymentRepository) GetByPartitionAndID(
	ctx context.Context,
	partitionID string,
	id string,
) (*models.Payment, error) {
	payment := models.Payment{}
	err := pr.Pool().DB(ctx, true).First(&payment, "partition_id = ? AND id = ?", partitionID, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (pr *paymentRepository) GetByProfileID(ctx context.Context, profileID string) ([]*models.Payment, error) {
	var payments []*models.Payment
	err := pr.Pool().DB(ctx, true).
		Where("sender_profile_id = ? OR recipient_profile_id = ?", profileID, profileID).
		Find(&payments).Error
	return payments, err
}
