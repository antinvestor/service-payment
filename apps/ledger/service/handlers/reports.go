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

package handlers

import (
	"context"
	"time"

	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	utilmoney "github.com/pitabwire/util/moneyx"
)

// GetTrialBalance returns per-account debit/credit totals plus per-currency
// grand totals across cleared NORMAL + REVERSAL postings. The is_balanced
// flag is the textbook integrity check: total_debits == total_credits
// within a currency.
//
// Filters compose: currency / ledger_id / ledger_type narrow the lines,
// book_ids scopes by book (supply ListBookDescendants output for
// consolidated reports across an organisation's groups + members),
// as_of is an upper bound on transacted_at.
func (ledgerSrv *LedgerServer) GetTrialBalance(
	ctx context.Context,
	req *connect.Request[ledgerv1.GetTrialBalanceRequest],
) (*connect.Response[ledgerv1.GetTrialBalanceResponse], error) {
	params := models.TrialBalanceParams{
		Currency:   req.Msg.GetCurrency(),
		LedgerID:   req.Msg.GetLedgerId(),
		LedgerType: req.Msg.GetLedgerType(),
		BookIDs:    req.Msg.GetBookIds(),
	}
	if asOf := req.Msg.GetAsOf(); asOf != "" {
		if parsed, err := time.Parse(time.RFC3339, asOf); err == nil {
			params.AsOf = &parsed
		}
	}

	report, err := ledgerSrv.Report.TrialBalance(ctx, params)
	if err != nil {
		return nil, ToConnectError(err)
	}

	apiLines := make([]*ledgerv1.TrialBalanceLine, len(report.Lines))
	for i, line := range report.Lines {
		apiLines[i] = &ledgerv1.TrialBalanceLine{
			AccountId:    line.AccountID,
			LedgerId:     line.LedgerID,
			LedgerType:   models.ToLedgerType(line.LedgerType),
			Currency:     line.Currency,
			TotalDebits:  utilmoney.ToMoney(line.Currency, line.TotalDebits),
			TotalCredits: utilmoney.ToMoney(line.Currency, line.TotalCredits),
			NetBalance:   utilmoney.ToMoney(line.Currency, line.NetBalance),
		}
	}

	apiTotals := make([]*ledgerv1.TrialBalanceTotal, len(report.Totals))
	for i, total := range report.Totals {
		apiTotals[i] = &ledgerv1.TrialBalanceTotal{
			Currency:     total.Currency,
			TotalDebits:  utilmoney.ToMoney(total.Currency, total.TotalDebits),
			TotalCredits: utilmoney.ToMoney(total.Currency, total.TotalCredits),
			IsBalanced:   total.IsBalanced,
		}
	}

	return connect.NewResponse(&ledgerv1.GetTrialBalanceResponse{
		Lines:  apiLines,
		Totals: apiTotals,
	}), nil
}

// GetAccountStatement returns the account's entries in chronological order
// for the supplied period, hydrated with running balance per entry,
// opening + closing balances and per-side period totals.
//
// Pagination is limit+offset; running balance is computed in the business
// layer against the opening balance carried in from before the period,
// so paging forward yields a continuous statement.
func (ledgerSrv *LedgerServer) GetAccountStatement(
	ctx context.Context,
	req *connect.Request[ledgerv1.GetAccountStatementRequest],
) (*connect.Response[ledgerv1.GetAccountStatementResponse], error) {
	params := models.StatementParams{
		AccountID: req.Msg.GetAccountId(),
		Limit:     int(req.Msg.GetLimit()),
		Offset:    int(req.Msg.GetOffset()),
	}
	if from := req.Msg.GetFrom(); from != "" {
		if parsed, err := time.Parse(time.RFC3339, from); err == nil {
			params.From = &parsed
		}
	}
	if to := req.Msg.GetTo(); to != "" {
		if parsed, err := time.Parse(time.RFC3339, to); err == nil {
			params.To = &parsed
		}
	}

	stmt, err := ledgerSrv.Report.AccountStatement(ctx, params)
	if err != nil {
		return nil, ToConnectError(err)
	}

	apiEntries := make([]*ledgerv1.StatementEntry, len(stmt.Entries))
	for i, e := range stmt.Entries {
		apiEntries[i] = statementEntryToAPI(stmt.Currency, e)
	}

	return connect.NewResponse(&ledgerv1.GetAccountStatementResponse{
		AccountId:      stmt.AccountID,
		Currency:       stmt.Currency,
		OpeningBalance: utilmoney.ToMoney(stmt.Currency, stmt.OpeningBalance),
		ClosingBalance: utilmoney.ToMoney(stmt.Currency, stmt.ClosingBalance),
		TotalDebits:    utilmoney.ToMoney(stmt.Currency, stmt.TotalDebits),
		TotalCredits:   utilmoney.ToMoney(stmt.Currency, stmt.TotalCredits),
		Entries:        apiEntries,
	}), nil
}

func statementEntryToAPI(
	currency string, e *models.StatementEntryRow,
) *ledgerv1.StatementEntry {
	apiTxnType := ledgerv1.TransactionType_NORMAL
	if v, ok := ledgerv1.TransactionType_value[e.TransactionType]; ok {
		apiTxnType = ledgerv1.TransactionType(v)
	}

	out := &ledgerv1.StatementEntry{
		EntryId:         e.EntryID,
		TransactionId:   e.TransactionID,
		TransactedAt:    e.TransactedAt.Format(time.RFC3339),
		TransactionType: apiTxnType,
		Amount:          utilmoney.ToMoney(e.Currency, e.Amount),
		Credit:          e.Credit,
		RunningBalance:  utilmoney.ToMoney(currency, e.RunningBalance),
		TransactionData: e.TransactionData.ToProtoStruct(),
	}
	if !e.ClearedAt.IsZero() {
		out.ClearedAt = e.ClearedAt.Format(time.RFC3339)
	}
	return out
}
