package business

import (
	"context"
	"fmt"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/internal/apperrors"
	"github.com/pitabwire/frame/datastore/pool"
	"github.com/pitabwire/frame/workerpool"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const defaultPaymentTermDays = 30

// InvoiceEngine generates and manages invoices.
type InvoiceEngine interface {
	GenerateInvoice(ctx context.Context, billingRun *models.BillingRun, ratedLines []*models.RatedLine,
		discountedLines []*models.DiscountedLine, creditAmount decimal.Decimal) (*models.Invoice, error)
	UpdateInvoiceTotals(ctx context.Context, invoice *models.Invoice, creditAmount decimal.Decimal) error
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
	creditAmount decimal.Decimal,
) (*models.Invoice, error) {
	if len(ratedLines) == 0 {
		return nil, ErrInvoiceNoRatedLines
	}

	// Calculate subtotal
	subtotal := decimal.Zero
	for _, rl := range ratedLines {
		if rl.Amount.Valid {
			subtotal = subtotal.Add(rl.Amount.Decimal)
		}
	}

	// Calculate total discounts
	discountTotal := decimal.Zero
	for _, dl := range discountedLines {
		if dl.Amount.Valid {
			discountTotal = discountTotal.Add(dl.Amount.Decimal)
		}
	}

	total := subtotal.Sub(discountTotal).Sub(creditAmount)
	if total.IsNegative() {
		total = decimal.Zero
	}

	invoice := &models.Invoice{
		BillingRunID:   billingRun.GetID(),
		ProfileID:      billingRun.ProfileID,
		SubscriptionID: billingRun.SubscriptionID,
		InvoiceNumber:  fmt.Sprintf("INV-%s", billingRun.GetID()),
		State:          models.InvoiceStateDraft,
		Currency:       ratedLines[0].Currency,
		SubtotalAmount: decimal.NewNullDecimal(subtotal),
		DiscountAmount: decimal.NewNullDecimal(discountTotal),
		CreditAmount:   decimal.NewNullDecimal(creditAmount),
		TotalAmount:    decimal.NewNullDecimal(total),
		PeriodStart:    billingRun.PeriodStart,
		PeriodEnd:      billingRun.PeriodEnd,
	}
	invoice.GenID(ctx)

	// Build discount lookup by rated line ID
	discountByRatedLine := make(map[string]decimal.Decimal)
	for _, dl := range discountedLines {
		existing := discountByRatedLine[dl.RatedLineID]
		discountByRatedLine[dl.RatedLineID] = existing.Add(dl.Amount.Decimal)
	}

	// Wrap invoice + all lines in a transaction
	err := e.dbPool.DB(ctx, false).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(invoice).Error; err != nil {
			return err
		}

		for _, rl := range ratedLines {
			lineDiscount := discountByRatedLine[rl.GetID()]
			netAmount := rl.Amount.Decimal.Sub(lineDiscount)
			if netAmount.IsNegative() {
				netAmount = decimal.Zero
			}

			il := &models.InvoiceLine{
				InvoiceID:      invoice.GetID(),
				RatedLineID:    rl.GetID(),
				ComponentID:    rl.ComponentID,
				Description:    rl.Description,
				Quantity:       rl.Quantity,
				UnitPrice:      rl.UnitPrice,
				Amount:         rl.Amount,
				DiscountAmount: decimal.NewNullDecimal(lineDiscount),
				CreditAmount:   decimal.NewNullDecimal(decimal.Zero),
				NetAmount:      decimal.NewNullDecimal(netAmount),
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
	creditAmount decimal.Decimal,
) error {
	newTotal := invoice.SubtotalAmount.Decimal.Sub(invoice.DiscountAmount.Decimal).Sub(creditAmount)
	if newTotal.IsNegative() {
		newTotal = decimal.Zero
	}
	invoice.CreditAmount = decimal.NewNullDecimal(creditAmount)
	invoice.TotalAmount = decimal.NewNullDecimal(newTotal)

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
