package repository

import (
	"context"

	"github.com/pitabwire/frame"
	"gorm.io/gorm"
)

type abstractRepository struct {
	service *frame.Service
}

func (ar *abstractRepository) readDb(ctx context.Context) *gorm.DB {
	return ar.service.DB(ctx, true)
}

func (ar *abstractRepository) writeDb(ctx context.Context) *gorm.DB {
	return ar.service.DB(ctx, true)
}
