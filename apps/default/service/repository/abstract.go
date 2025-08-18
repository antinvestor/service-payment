package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/pitabwire/frame"
)

type abstractRepository struct {
	service *frame.Service
}

func (ar *abstractRepository) readDB(ctx context.Context) *gorm.DB {
	return ar.service.DB(ctx, true)
}

func (ar *abstractRepository) writeDB(ctx context.Context) *gorm.DB {
	return ar.service.DB(ctx, true)
}
