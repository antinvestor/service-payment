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
	"regexp"

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/service/observability"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/data"
)

// Errors for book operations.
var (
	ErrBookNameRequired = errors.New("book name is required")
	ErrBookTypeRequired = errors.New("book type is required")
)

// bookIDPattern is CreateBookRequest.id's rule from the proto, anchored so
// the whole id must match rather than any substring of it.
var bookIDPattern = regexp.MustCompile(`^[0-9a-z_-]{3,40}$`)

// BookBusiness exposes the application-level operations on Book aggregates.
// A Book is an independent accounting scope — a platform book, a savings-
// group book, a merchant book, etc. — whose entries balance independently
// from every other book in the system. CRUD only; cross-book integrity
// rules live with the transaction posting flow.
type BookBusiness interface {
	// CreateBook opens a book. A non-empty id names it, otherwise one is
	// generated. Creating a book whose id already exists returns that book
	// when name, type, parent and currency match, and ErrBookIDTaken when
	// they do not.
	CreateBook(
		ctx context.Context,
		id, name, bookType, currency string,
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
	metrics  *observability.Metrics
}

// NewBookBusiness constructs the book business layer.
func NewBookBusiness(bookRepo repository.BookRepository) BookBusiness {
	return &bookBusiness{bookRepo: bookRepo, metrics: observability.NewMetrics()}
}

func (b *bookBusiness) CreateBook(
	ctx context.Context,
	id, name, bookType, currency string,
	parentID *string,
	dataMap data.JSONMap,
) (*models.Book, error) {
	ctx, span := b.metrics.StartSpan(ctx, "CreateBook")
	var err error
	defer func() { b.metrics.EndSpan(ctx, span, err) }()

	if name == "" {
		err = ErrBookNameRequired
		return nil, err
	}
	if bookType == "" {
		err = ErrBookTypeRequired
		return nil, err
	}
	if id != "" && !bookIDPattern.MatchString(id) {
		err = apperrors.ErrBookIDInvalid
		return nil, err
	}
	if parentID != nil && *parentID == "" {
		parentID = nil
	}

	book := &models.Book{
		BaseModel: data.BaseModel{ID: id},
		ParentID:  parentID,
		Name:      name,
		Type:      bookType,
		Currency:  currency,
		Data:      dataMap,
	}

	if id != "" {
		var existing *models.Book
		existing, err = b.findBook(ctx, id)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			existing, err = sameBook(existing, book)
			return existing, err
		}
	}

	// If a parent is supplied, confirm it exists and belongs to the
	// caller's tenancy scope (GetByID applies the auto-scope). This
	// rejects forging a hierarchy across tenants.
	if parentID != nil {
		if _, err = b.GetBook(ctx, *parentID); err != nil {
			return nil, err
		}
	}

	book.GenID(ctx)

	if err = b.bookRepo.Create(ctx, book); err != nil {
		if id == "" || !data.ErrorIsDuplicateKey(err) {
			err = apperrors.ErrSystemFailure.Override(err)
			return nil, err
		}
		// A concurrent create won the id, or the id belongs to a book
		// outside the caller's scope (which reads as taken).
		var existing *models.Book
		existing, err = b.findBook(ctx, id)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			err = apperrors.ErrBookIDTaken
			return nil, err
		}
		existing, err = sameBook(existing, book)
		return existing, err
	}

	b.metrics.RecordBookCreated(ctx)
	return book, nil
}

// findBook returns the book with id in the caller's scope, or nil when there
// is none.
func (b *bookBusiness) findBook(ctx context.Context, id string) (*models.Book, error) {
	book, err := b.bookRepo.GetByID(ctx, id)
	if err != nil {
		if data.ErrorIsNoRows(err) {
			return nil, nil
		}
		return nil, apperrors.ErrSystemFailure.Override(err)
	}
	return book, nil
}

// sameBook returns existing when it carries the attributes a create asked
// for, so a repeated create is idempotent, and ErrBookIDTaken otherwise.
// Data is metadata and is not compared.
func sameBook(existing, requested *models.Book) (*models.Book, error) {
	if existing.Name != requested.Name ||
		existing.Type != requested.Type ||
		existing.Currency != requested.Currency ||
		deref(existing.ParentID) != deref(requested.ParentID) {
		return nil, apperrors.ErrBookIDTaken
	}
	return existing, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
	book, err := b.findBook(ctx, id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, apperrors.ErrBookNotFound
	}
	return book, nil
}

func (b *bookBusiness) ListBooksByType(
	ctx context.Context, bookType string,
) ([]*models.Book, error) {
	return b.bookRepo.ListByType(ctx, bookType)
}
