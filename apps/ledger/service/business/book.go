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

package business

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/data"
)

// Errors for book operations.
var (
	ErrBookNameRequired = errors.New("book name is required")
	ErrBookTypeRequired = errors.New("book type is required")
)

// BookBusiness exposes the application-level operations on Book aggregates.
// A Book is an independent accounting scope — a platform book, a savings-
// group book, a merchant book, etc. — whose entries balance independently
// from every other book in the system. CRUD only; cross-book integrity
// rules live with the transaction posting flow.
type BookBusiness interface {
	CreateBook(
		ctx context.Context,
		name, bookType, currency string,
		parentID *string,
		dataMap data.JSONMap,
	) (*models.Book, error)
	GetBook(ctx context.Context, id string) (*models.Book, error)
	ListBooksByType(ctx context.Context, bookType string) ([]*models.Book, error)
	// ListDescendantIDs returns the supplied book's id and every
	// transitive child for consolidated reporting (organization → groups
	// → individual members).
	ListDescendantIDs(ctx context.Context, bookID string) ([]string, error)
}

type bookBusiness struct {
	bookRepo repository.BookRepository
}

// NewBookBusiness constructs the book business layer.
func NewBookBusiness(bookRepo repository.BookRepository) BookBusiness {
	return &bookBusiness{bookRepo: bookRepo}
}

func (b *bookBusiness) CreateBook(
	ctx context.Context,
	name, bookType, currency string,
	parentID *string,
	dataMap data.JSONMap,
) (*models.Book, error) {
	if name == "" {
		return nil, ErrBookNameRequired
	}
	if bookType == "" {
		return nil, ErrBookTypeRequired
	}

	// If a parent is supplied, confirm it exists and belongs to the
	// caller's tenancy scope (GetByID applies the auto-scope). This
	// rejects forging a hierarchy across tenants.
	if parentID != nil && *parentID != "" {
		if _, err := b.bookRepo.GetByID(ctx, *parentID); err != nil {
			return nil, err
		}
	}

	book := &models.Book{
		ParentID: parentID,
		Name:     name,
		Type:     bookType,
		Currency: currency,
		Data:     dataMap,
	}
	book.GenID(ctx)

	if err := b.bookRepo.Create(ctx, book); err != nil {
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return book, nil
}

func (b *bookBusiness) ListDescendantIDs(
	ctx context.Context, bookID string,
) ([]string, error) {
	return b.bookRepo.ListDescendantIDs(ctx, bookID)
}

func (b *bookBusiness) GetBook(ctx context.Context, id string) (*models.Book, error) {
	if id == "" {
		return nil, apperrors.ErrUnspecifiedID
	}
	book, err := b.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (b *bookBusiness) ListBooksByType(
	ctx context.Context, bookType string,
) ([]*models.Book, error) {
	return b.bookRepo.ListByType(ctx, bookType)
}
