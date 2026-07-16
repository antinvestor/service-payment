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

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	ledgerBusiness "github.com/antinvestor/service-payments/apps/ledger/service/business"
	ledgerModels "github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/util/decimalx"
)

// LedgerIntegration posts billing events to the ledger.
type LedgerIntegration interface {
	PostInvoiceToLedger(
		ctx context.Context,
		invoice *models.Invoice,
		arAccountID, revenueAccountID string,
	) (string, error)
	// PostPaymentToLedger records cash collection against AR:
	// Debit Cash (ASSET), Credit AR (ASSET).
	PostPaymentToLedger(
		ctx context.Context,
		invoice *models.Invoice,
		cashAccountID, arAccountID string,
	) (string, error)
	PostCreditGrantToLedger(
		ctx context.Context,
		grant *models.CreditGrant,
		creditLiabilityAccountID, sourceAccountID string,
	) (string, error)
	PostCreditConsumeToLedger(
		ctx context.Context,
		entry *models.CreditEntry,
		revenueAccountID, creditLiabilityAccountID string,
	) (string, error)
}

type ledgerIntegration struct {
	txnBusiness ledgerBusiness.TransactionBusiness
}

func NewLedgerIntegration(txnBusiness ledgerBusiness.TransactionBusiness) LedgerIntegration {
	return &ledgerIntegration{
		txnBusiness: txnBusiness,
	}
}

// PostInvoiceToLedger creates a double-entry transaction: Debit AR (ASSET), Credit Revenue (INCOME).
func (l *ledgerIntegration) PostInvoiceToLedger(
	ctx context.Context,
	invoice *models.Invoice,
	arAccountID, revenueAccountID string,
) (string, error) {
	txnID := fmt.Sprintf("billing_inv_%s", invoice.GetID())
	amount := decimalx.DerefOr(invoice.TotalAmount, decimalx.Zero())

	if amount.IsZero() {
		return "", nil
	}

	txn := &ledgerModels.Transaction{
		BaseModel:       data.BaseModel{ID: txnID},
		Currency:        invoice.Currency,
		TransactionType: "NORMAL",
		Entries: []*ledgerModels.TransactionEntry{
			{
				AccountID: arAccountID,
				Amount:    amount.Ptr(),
				Credit:    false, // Debit AR
			},
			{
				AccountID: revenueAccountID,
				Amount:    amount.Ptr(),
				Credit:    true, // Credit Revenue
			},
		},
	}

	result, err := l.txnBusiness.Transact(ctx, txn)
	if err != nil {
		return "", fmt.Errorf("failed to post invoice to ledger: %w", err)
	}

	return result.GetID(), nil
}

// PostPaymentToLedger creates a double-entry transaction: Debit Cash, Credit AR.
// Idempotent by transaction ID derived from the invoice ID.
func (l *ledgerIntegration) PostPaymentToLedger(
	ctx context.Context,
	invoice *models.Invoice,
	cashAccountID, arAccountID string,
) (string, error) {
	if cashAccountID == "" || arAccountID == "" {
		return "", nil // ledger accounts not configured — skip silently
	}
	txnID := fmt.Sprintf("billing_pay_%s", invoice.GetID())
	amount := decimalx.DerefOr(invoice.TotalAmount, decimalx.Zero())

	if amount.IsZero() {
		return "", nil
	}

	txn := &ledgerModels.Transaction{
		BaseModel:       data.BaseModel{ID: txnID},
		Currency:        invoice.Currency,
		TransactionType: "NORMAL",
		Entries: []*ledgerModels.TransactionEntry{
			{
				AccountID: cashAccountID,
				Amount:    amount.Ptr(),
				Credit:    false, // Debit Cash
			},
			{
				AccountID: arAccountID,
				Amount:    amount.Ptr(),
				Credit:    true, // Credit AR
			},
		},
	}

	result, err := l.txnBusiness.Transact(ctx, txn)
	if err != nil {
		return "", fmt.Errorf("failed to post payment to ledger: %w", err)
	}

	return result.GetID(), nil
}

// PostCreditGrantToLedger creates a transaction: Debit Credit Liability, Credit source account.
func (l *ledgerIntegration) PostCreditGrantToLedger(
	ctx context.Context,
	grant *models.CreditGrant,
	creditLiabilityAccountID, sourceAccountID string,
) (string, error) {
	txnID := fmt.Sprintf("billing_credit_grant_%s", grant.GetID())
	amount := decimalx.DerefOr(grant.OriginalAmount, decimalx.Zero())

	txn := &ledgerModels.Transaction{
		BaseModel:       data.BaseModel{ID: txnID},
		Currency:        grant.Currency,
		TransactionType: "NORMAL",
		Entries: []*ledgerModels.TransactionEntry{
			{
				AccountID: creditLiabilityAccountID,
				Amount:    amount.Ptr(),
				Credit:    false, // Debit Credit Liability
			},
			{
				AccountID: sourceAccountID,
				Amount:    amount.Ptr(),
				Credit:    true, // Credit source
			},
		},
	}

	result, err := l.txnBusiness.Transact(ctx, txn)
	if err != nil {
		return "", fmt.Errorf("failed to post credit grant to ledger: %w", err)
	}

	return result.GetID(), nil
}

// PostCreditConsumeToLedger creates a transaction: Debit Revenue, Credit Credit Liability.
func (l *ledgerIntegration) PostCreditConsumeToLedger(
	ctx context.Context,
	entry *models.CreditEntry,
	revenueAccountID, creditLiabilityAccountID string,
) (string, error) {
	txnID := fmt.Sprintf("billing_credit_consume_%s", entry.GetID())
	amount := decimalx.DerefOr(entry.Amount, decimalx.Zero())

	txn := &ledgerModels.Transaction{
		BaseModel:       data.BaseModel{ID: txnID},
		Currency:        entry.Currency,
		TransactionType: "NORMAL",
		Entries: []*ledgerModels.TransactionEntry{
			{
				AccountID: revenueAccountID,
				Amount:    amount.Ptr(),
				Credit:    false, // Debit Revenue
			},
			{
				AccountID: creditLiabilityAccountID,
				Amount:    amount.Ptr(),
				Credit:    true, // Credit Credit Liability
			},
		},
	}

	result, err := l.txnBusiness.Transact(ctx, txn)
	if err != nil {
		return "", fmt.Errorf("failed to post credit consume to ledger: %w", err)
	}

	return result.GetID(), nil
}
