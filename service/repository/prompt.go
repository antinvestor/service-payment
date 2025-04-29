package repository

import (
	"context"
	"strings"
	"github.com/pitabwire/frame"
	"github.com/antinvestor/service-payments/service/models"
)

type PromptRepository interface {
	GetByID(ctx context.Context, id string) (*models.Prompt, error)
	GetByPartitionAndID(ctx context.Context, partitionID string, id string) (*models.Prompt, error)
	Search(ctx context.Context, query string) ([]*models.Prompt, error)
	Save(ctx context.Context, prompt *models.Prompt) error
}

 type promptRepository struct {
     abstractRepository
 }

 func NewPromptRepository(ctx context.Context, service *frame.Service) PromptRepository {
     return &promptRepository{abstractRepository{service: service}}
 }

 func (repo *promptRepository) GetByID(ctx context.Context, id string) (*models.Prompt, error) {
     prompt := models.Prompt{}
     err := repo.readDb(ctx).First(&prompt, "id = ?", id).Error
     if err != nil {
         return nil, err
     }
     return &prompt, nil
 }

 func (repo *promptRepository) GetByPartitionAndID(ctx context.Context, partitionID string, id string) (*models.Prompt, error) {
     prompt := models.Prompt{}
     err := repo.readDb(ctx).First(&prompt, "partition_id = ? AND id = ?", partitionID, id).Error
     if err != nil {
         return nil, err
     }
     return &prompt, nil
 }

 func (repo *promptRepository) Search(ctx context.Context, query string) ([]*models.Prompt, error) {
     var prompts []*models.Prompt
     err := repo.readDb(ctx).Where("name ILIKE ?", "%"+strings.ToLower(query)+"%").Find(&prompts).Error
     if err != nil {
         return nil, err
     }
     return prompts, nil
 }

 func (repo *promptRepository) Save(ctx context.Context, prompt *models.Prompt) error {
     return repo.writeDb(ctx).Save(prompt).Error
 }

