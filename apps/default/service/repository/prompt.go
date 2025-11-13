package repository

import (
	"context"
	"strings"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
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

// Legacy method for backward compatibility
func (pr *promptRepository) SearchLegacy(ctx context.Context, query string) ([]*models.Prompt, error) {
	var prompts []*models.Prompt
	err := pr.Pool().DB(ctx, true).Where("name ILIKE ?", "%"+strings.ToLower(query)+"%").Find(&prompts).Error
	if err != nil {
		return nil, err
	}
	return prompts, nil
}
