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

	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/pitabwire/util/decimalx"
)

// TrialBalanceReport is the full output of a trial-balance run: the
// per-account lines plus per-currency grand totals. IsBalanced on each
// total proves the fundamental double-entry invariant — total debits
// equals total credits within a currency — for the scope of the run.
type TrialBalanceReport struct {
	Lines  []*models.TrialBalanceLine
	Totals []*TrialBalanceTotal
}

// TrialBalanceTotal aggregates one currency's debits and credits across
// every line in the report. IsBalanced is true iff TotalDebits and
// TotalCredits are equal — the prove-the-books check.
type TrialBalanceTotal struct {
	Currency     string
	TotalDebits  decimalx.Decimal
	TotalCredits decimalx.Decimal
	IsBalanced   bool
}

// AccountStatement is a customer-facing account ledger for a period: the
// opening balance carried in from before the period, the entries posted
// within the period with running balances applied row by row, the closing
// balance, and the period's raw debit and credit totals.
type AccountStatement struct {
	AccountID      string
	Currency       string
	OpeningBalance decimalx.Decimal
	ClosingBalance decimalx.Decimal
	TotalDebits    decimalx.Decimal
	TotalCredits   decimalx.Decimal
	Entries        []*models.StatementEntryRow
}

// ReportBusiness exposes derived accounting reports. Implementations
// assemble repository aggregations into report types the API layer can
// surface to operators and customers.
type ReportBusiness interface {
	TrialBalance(
		ctx context.Context, params models.TrialBalanceParams,
	) (*TrialBalanceReport, error)
	AccountStatement(
		ctx context.Context, params models.StatementParams,
	) (*AccountStatement, error)
}

type reportBusiness struct {
	reportRepo repository.ReportRepository
}

// NewReportBusiness constructs the report business layer.
func NewReportBusiness(reportRepo repository.ReportRepository) ReportBusiness {
	return &reportBusiness{reportRepo: reportRepo}
}

func (b *reportBusiness) TrialBalance(
	ctx context.Context, params models.TrialBalanceParams,
) (*TrialBalanceReport, error) {
	lines, err := b.reportRepo.AggregateTrialBalance(ctx, params)
	if err != nil {
		return nil, err
	}

	// Group debits and credits per currency to compute the integrity check.
	totalsByCurrency := make(map[string]*TrialBalanceTotal)
	for _, line := range lines {
		t, ok := totalsByCurrency[line.Currency]
		if !ok {
			t = &TrialBalanceTotal{Currency: line.Currency}
			totalsByCurrency[line.Currency] = t
		}
		t.TotalDebits = t.TotalDebits.Add(line.TotalDebits)
		t.TotalCredits = t.TotalCredits.Add(line.TotalCredits)
	}

	totals := make([]*TrialBalanceTotal, 0, len(totalsByCurrency))
	for _, t := range totalsByCurrency {
		t.IsBalanced = t.TotalDebits.Equal(t.TotalCredits)
		totals = append(totals, t)
	}

	return &TrialBalanceReport{Lines: lines, Totals: totals}, nil
}

func (b *reportBusiness) AccountStatement(
	ctx context.Context, params models.StatementParams,
) (*AccountStatement, error) {
	opening, err := b.reportRepo.StatementOpeningBalance(ctx, params.AccountID, params.From)
	if err != nil {
		return nil, err
	}

	entries, err := b.reportRepo.AccountStatementEntries(ctx, params)
	if err != nil {
		return nil, err
	}

	stmt := &AccountStatement{
		AccountID:      params.AccountID,
		OpeningBalance: opening,
		ClosingBalance: opening,
		Entries:        entries,
	}

	for _, entry := range entries {
		// Capture the currency from the first hydrated entry. All entries
		// for one account share a currency (account currency is single-
		// valued), so this is stable across the page.
		if stmt.Currency == "" {
			stmt.Currency = entry.Currency
		}
		stmt.ClosingBalance = stmt.ClosingBalance.Add(entry.Amount)
		entry.RunningBalance = stmt.ClosingBalance

		abs := entry.Amount.Abs()
		if entry.Credit {
			stmt.TotalCredits = stmt.TotalCredits.Add(abs)
		} else {
			stmt.TotalDebits = stmt.TotalDebits.Add(abs)
		}
	}

	return stmt, nil
}
