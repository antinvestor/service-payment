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

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/datastore"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
)

// BookRepository persists Book aggregates — one per independent
// accounting scope (platform book, group book, merchant book, etc.).
// Embeds BaseRepository so standard Create/Get/Update/Delete flow
// through frame's TenancyPartition scope and soft-delete unchanged.
type BookRepository interface {
	datastore.BaseRepository[*models.Book]
	ListByType(ctx context.Context, bookType string) ([]*models.Book, error)
	// ListDescendantIDs returns the supplied book's id together with the
	// ids of every transitive child book (down the parent_id chain),
	// scoped by tenancy. Used by consolidated reports — e.g. "trial
	// balance for this organization and every group beneath it".
	ListDescendantIDs(ctx context.Context, bookID string) ([]string, error)
}

type bookRepository struct {
	datastore.BaseRepository[*models.Book]
}

// NewBookRepository constructs the book repository.
func NewBookRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) BookRepository {
	return &bookRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Book](
			ctx, dbPool, workMan, func() *models.Book { return &models.Book{} },
		),
	}
}

// ListByType returns active books of the given type within the caller's
// tenancy scope. Useful for listing all group books, all merchant books,
// etc. when an operator is building per-entity reports.
func (b *bookRepository) ListByType(
	ctx context.Context, bookType string,
) ([]*models.Book, error) {
	if bookType == "" {
		return nil, apperrors.ErrUnspecifiedID.Extend("book type is required")
	}
	var books []*models.Book
	err := b.Pool().DB(ctx, true).
		Where("type = ?", bookType).
		Order("created_at DESC").
		Find(&books).Error
	if err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return books, nil
}

// ListDescendantIDs walks the parent_id tree from the supplied book down
// to every leaf, returning all reachable book ids (root included). Uses
// a recursive CTE so we avoid N+1 lookups and let the database manage the
// traversal cost. Tenancy isolation continues to apply via frame's
// auto-scope on the pool; the CTE starts from the supplied id which has
// already been confirmed to belong to the caller's scope by GetByID.
func (b *bookRepository) ListDescendantIDs(
	ctx context.Context, bookID string,
) ([]string, error) {
	if bookID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	const query = `
WITH RECURSIVE descendants AS (
    SELECT id, parent_id
    FROM books
    WHERE id = ? AND deleted_at IS NULL
    UNION ALL
    SELECT child.id, child.parent_id
    FROM books child
    JOIN descendants d ON child.parent_id = d.id
    WHERE child.deleted_at IS NULL
)
SELECT id FROM descendants
`

	var ids []string
	if err := b.Pool().DB(ctx, true).Raw(query, bookID).Scan(&ids).Error; err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return ids, nil
}
