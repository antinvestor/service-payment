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

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/datastore"
	"github.com/pitabwire/frame/v2/datastore/pool"
	"github.com/pitabwire/frame/v2/workerpool"
)

// InvoiceRepository provides operations for invoices.
type InvoiceRepository interface {
	datastore.BaseRepository[*models.Invoice]
	SearchAsESQ(ctx context.Context, query string) (workerpool.JobResultPipe[[]*models.Invoice], error)
	GetByBillingRunID(ctx context.Context, billingRunID string) (*models.Invoice, error)
	// ListOpenWithCheckoutSession returns ISSUED invoices that have a
	// checkoutSessionRef stored in Data (candidates for auto-settle).
	ListOpenWithCheckoutSession(ctx context.Context, limit int) ([]*models.Invoice, error)
}

type invoiceRepository struct {
	datastore.BaseRepository[*models.Invoice]
}

func NewInvoiceRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) InvoiceRepository {
	return &invoiceRepository{
		BaseRepository: datastore.NewBaseRepository[*models.Invoice](
			ctx, dbPool, workMan, func() *models.Invoice { return &models.Invoice{} },
		),
	}
}

func (r *invoiceRepository) GetByBillingRunID(ctx context.Context, billingRunID string) (*models.Invoice, error) {
	if billingRunID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var invoice models.Invoice
	result := r.Pool().DB(ctx, true).Where("billing_run_id = ?", billingRunID).Preload("Lines").First(&invoice)
	if result.Error != nil {
		if data.ErrorIsNoRows(result.Error) {
			return nil, apperrors.ErrInvoiceNotFound
		}
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return &invoice, nil
}

// ListOpenWithCheckoutSession finds ISSUED invoices that started hosted checkout.
// Uses a JSONB path filter on data->>'checkoutSessionRef'.
func (r *invoiceRepository) ListOpenWithCheckoutSession(
	ctx context.Context,
	limit int,
) ([]*models.Invoice, error) {
	if limit <= 0 {
		limit = 50
	}
	var list []*models.Invoice
	result := r.Pool().DB(ctx, true).
		Where("state = ?", models.InvoiceStateIssued).
		Where("data->>'checkoutSessionRef' IS NOT NULL").
		Where("data->>'checkoutSessionRef' <> ''").
		Order("modified_at ASC").
		Limit(limit).
		Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}
	return list, nil
}

//nolint:dupl // SearchAsESQ follows a generic pattern that Go cannot DRY with generics + worker pools.
func (r *invoiceRepository) SearchAsESQ(
	ctx context.Context, query string,
) (workerpool.JobResultPipe[[]*models.Invoice], error) {
	job := workerpool.NewJob(func(ctx context.Context, jobResult workerpool.JobResultPipe[[]*models.Invoice]) error {
		rawQuery, err := NewSearchRawQuery(ctx, query)
		if err != nil {
			return jobResult.WriteError(ctx, err)
		}

		sqlQuery := rawQuery.ToQueryConditions()
		for sqlQuery.canLoad() {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			var list []*models.Invoice
			result := r.Pool().DB(ctx, true).Where(sqlQuery.sql, sqlQuery.args...).
				Offset(sqlQuery.offset).Limit(sqlQuery.batchSize).Find(&list)
			if result.Error != nil {
				return jobResult.WriteError(ctx, apperrors.ErrSystemFailure.Override(result.Error))
			}

			if err := jobResult.WriteResult(ctx, list); err != nil {
				return err
			}

			if sqlQuery.stop(len(list)) {
				break
			}
		}
		return nil
	})

	if err := workerpool.SubmitJob(ctx, r.WorkManager(), job); err != nil {
		return nil, err
	}
	return job, nil
}

// InvoiceLineRepository provides operations for invoice lines.
type InvoiceLineRepository interface {
	datastore.BaseRepository[*models.InvoiceLine]
	ListByInvoiceID(ctx context.Context, invoiceID string) ([]*models.InvoiceLine, error)
}

type invoiceLineRepository struct {
	datastore.BaseRepository[*models.InvoiceLine]
}

func NewInvoiceLineRepository(ctx context.Context, dbPool pool.Pool, workMan workerpool.Manager) InvoiceLineRepository {
	return &invoiceLineRepository{
		BaseRepository: datastore.NewBaseRepository[*models.InvoiceLine](
			ctx, dbPool, workMan, func() *models.InvoiceLine { return &models.InvoiceLine{} },
		),
	}
}

func (r *invoiceLineRepository) ListByInvoiceID(ctx context.Context, invoiceID string) ([]*models.InvoiceLine, error) {
	if invoiceID == "" {
		return nil, apperrors.ErrUnspecifiedID
	}

	var list []*models.InvoiceLine
	result := r.Pool().DB(ctx, true).Where("invoice_id = ?", invoiceID).Find(&list)
	if result.Error != nil {
		return nil, apperrors.ErrSystemFailure.Override(result.Error)
	}

	return list, nil
}
