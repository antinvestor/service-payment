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
	"fmt"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util/decimalx"
	"gorm.io/gorm"
)

const defaultPaymentTermDays = 30

// InvoiceEngine generates and manages invoices.
type InvoiceEngine interface {
	GenerateInvoice(ctx context.Context, billingRun *models.BillingRun, ratedLines []*models.RatedLine,
		discountedLines []*models.DiscountedLine, creditAmount decimalx.Decimal) (*models.Invoice, error)
	UpdateInvoiceTotals(ctx context.Context, invoice *models.Invoice, creditAmount decimalx.Decimal) error
	GetInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error)
	IssueInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error)
	VoidInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error)
	RecordPayment(ctx context.Context, invoiceID string) (*models.Invoice, error)
}

type invoiceEngine struct {
	workMan     workerpool.Manager
	dbPool      pool.Pool
	invoiceRepo repository.InvoiceRepository
	lineRepo    repository.InvoiceLineRepository
}

func NewInvoiceEngine(
	workMan workerpool.Manager,
	dbPool pool.Pool,
	invoiceRepo repository.InvoiceRepository,
	lineRepo repository.InvoiceLineRepository,
) InvoiceEngine {
	return &invoiceEngine{
		workMan:     workMan,
		dbPool:      dbPool,
		invoiceRepo: invoiceRepo,
		lineRepo:    lineRepo,
	}
}

func (e *invoiceEngine) GenerateInvoice(
	ctx context.Context,
	billingRun *models.BillingRun,
	ratedLines []*models.RatedLine,
	discountedLines []*models.DiscountedLine,
	creditAmount decimalx.Decimal,
) (*models.Invoice, error) {
	if len(ratedLines) == 0 {
		return nil, ErrInvoiceNoRatedLines
	}

	// Calculate subtotal
	subtotal := decimalx.Zero()
	for _, rl := range ratedLines {
		if rl.Amount != nil {
			subtotal = subtotal.Add(*rl.Amount)
		}
	}

	// Calculate total discounts
	discountTotal := decimalx.Zero()
	for _, dl := range discountedLines {
		if dl.Amount != nil {
			discountTotal = discountTotal.Add(*dl.Amount)
		}
	}

	total := subtotal.Sub(discountTotal).Sub(creditAmount)
	if total.IsNegative() {
		total = decimalx.Zero()
	}

	invoice := &models.Invoice{
		BillingRunID:   billingRun.GetID(),
		ProfileID:      billingRun.ProfileID,
		SubscriptionID: billingRun.SubscriptionID,
		InvoiceNumber:  fmt.Sprintf("INV-%s", billingRun.GetID()),
		State:          models.InvoiceStateDraft,
		Currency:       ratedLines[0].Currency,
		SubtotalAmount: subtotal.Ptr(),
		DiscountAmount: discountTotal.Ptr(),
		CreditAmount:   creditAmount.Ptr(),
		TotalAmount:    total.Ptr(),
		PeriodStart:    billingRun.PeriodStart,
		PeriodEnd:      billingRun.PeriodEnd,
	}
	invoice.GenID(ctx)

	// Build discount lookup by rated line ID
	discountByRatedLine := make(map[string]decimalx.Decimal)
	for _, dl := range discountedLines {
		existing := discountByRatedLine[dl.RatedLineID]
		discountByRatedLine[dl.RatedLineID] = existing.Add(*dl.Amount)
	}

	// Wrap invoice + all lines in a transaction
	err := e.dbPool.DB(ctx, false).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(invoice).Error; err != nil {
			return err
		}

		for _, rl := range ratedLines {
			lineDiscount := discountByRatedLine[rl.GetID()]
			rlAmount := decimalx.DerefOr(rl.Amount, decimalx.Zero())
			netAmount := rlAmount.Sub(lineDiscount)
			if netAmount.IsNegative() {
				netAmount = decimalx.Zero()
			}

			il := &models.InvoiceLine{
				InvoiceID:      invoice.GetID(),
				RatedLineID:    rl.GetID(),
				ComponentID:    rl.ComponentID,
				Description:    rl.Description,
				Quantity:       rl.Quantity,
				UnitPrice:      rl.UnitPrice,
				Amount:         rl.Amount,
				DiscountAmount: lineDiscount.Ptr(),
				CreditAmount:   decimalx.Zero().Ptr(),
				NetAmount:      netAmount.Ptr(),
				Currency:       rl.Currency,
				LineType:       models.InvoiceLineTypeUsage,
			}
			il.GenID(ctx)
			il.ID = fmt.Sprintf("%s_%s", invoice.GetID(), rl.GetID())

			if err := tx.Create(il).Error; err != nil {
				return err
			}

			invoice.Lines = append(invoice.Lines, il)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

func (e *invoiceEngine) UpdateInvoiceTotals(
	ctx context.Context,
	invoice *models.Invoice,
	creditAmount decimalx.Decimal,
) error {
	subtotal := decimalx.DerefOr(invoice.SubtotalAmount, decimalx.Zero())
	discount := decimalx.DerefOr(invoice.DiscountAmount, decimalx.Zero())
	newTotal := subtotal.Sub(discount).Sub(creditAmount)
	if newTotal.IsNegative() {
		newTotal = decimalx.Zero()
	}
	invoice.CreditAmount = creditAmount.Ptr()
	invoice.TotalAmount = newTotal.Ptr()

	_, err := e.invoiceRepo.Update(ctx, invoice)
	return err
}

func (e *invoiceEngine) GetInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	if invoiceID == "" {
		return nil, ErrInvoiceIDRequired
	}

	invoice, err := e.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	// Load invoice lines
	lines, err := e.lineRepo.ListByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	invoice.Lines = lines

	return invoice, nil
}

func (e *invoiceEngine) IssueInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	if invoiceID == "" {
		return nil, ErrInvoiceIDRequired
	}

	invoice, err := e.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	if invoice.State != models.InvoiceStateDraft {
		return nil, apperrors.ErrInvoiceNotIssuable
	}

	now := time.Now()
	dueAt := now.AddDate(0, 0, defaultPaymentTermDays)
	invoice.State = models.InvoiceStateIssued
	invoice.IssuedAt = &now
	invoice.DueAt = &dueAt

	_, err = e.invoiceRepo.Update(ctx, invoice)
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

func (e *invoiceEngine) VoidInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	if invoiceID == "" {
		return nil, ErrInvoiceIDRequired
	}

	invoice, err := e.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	if invoice.State == models.InvoiceStatePaid {
		return nil, apperrors.ErrInvoiceNotVoidable.Extend("Cannot void a paid invoice")
	}

	invoice.State = models.InvoiceStateVoided

	_, err = e.invoiceRepo.Update(ctx, invoice)
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

func (e *invoiceEngine) RecordPayment(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	if invoiceID == "" {
		return nil, ErrInvoiceIDRequired
	}

	invoice, err := e.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	if invoice.State != models.InvoiceStateIssued {
		return nil, apperrors.ErrInvoiceNotPayable
	}

	now := time.Now()
	invoice.State = models.InvoiceStatePaid
	invoice.PaidAt = &now

	_, err = e.invoiceRepo.Update(ctx, invoice)
	if err != nil {
		return nil, err
	}

	return invoice, nil
}
