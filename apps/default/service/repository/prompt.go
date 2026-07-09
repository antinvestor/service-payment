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

type promptRepository struct {
	datastore.BaseRepository[*models.Prompt]
}

func NewPromptRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) PromptRepository {
	repo := promptRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Prompt](
			ctx, dbPool, workMan, func() *models.Prompt { return &models.Prompt{} },
		),
	}
	return &repo
}

func (pr *promptRepository) GetByPartitionAndID(
	ctx context.Context,
	partitionID string,
	id string,
) (*models.Prompt, error) {
	prompt := &models.Prompt{}
	err := pr.Pool().DB(ctx, true).First(prompt, "partition_id = ? AND id = ?", partitionID, id).Error
	return prompt, err
}

func (pr *promptRepository) GetByProfileID(ctx context.Context, profileID string) ([]*models.Prompt, error) {
	var prompts []*models.Prompt
	err := pr.Pool().DB(ctx, true).
		Where("source_id = ? OR recipient_id = ?", profileID, profileID).
		Find(&prompts).Error
	return prompts, err
}

// Legacy method for backward compatibility.
func (pr *promptRepository) SearchLegacy(ctx context.Context, query string) ([]*models.Prompt, error) {
	var prompts []*models.Prompt
	err := pr.Pool().DB(ctx, true).Where("name ILIKE ?", "%"+strings.ToLower(query)+"%").Find(&prompts).Error
	if err != nil {
		return nil, err
	}
	return prompts, nil
}
