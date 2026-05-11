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

package business_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/business"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"
)

// ReportsSuite exercises the trial-balance and account-statement reports
// against a real Postgres instance. Fixtures install both UGX and USD
// chart-of-accounts so we can verify per-currency aggregation.
type ReportsSuite struct {
	tests.BaseTestSuite
}

type acctFixture struct {
	id         string
	ledgerID   string
	ledgerType string
	currency   string
}

func (rs *ReportsSuite) setupFixtures(ctx context.Context, res *tests.ServiceResources) {
	createLedger := func(id string, lt ledgerv1.LedgerType) string {
		_, err := res.LedgerBusiness.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{
			Id:   id,
			Type: lt,
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"name": {Kind: &structpb.Value_StringValue{StringValue: id}},
				},
			},
		})
		rs.Require().NoError(err)
		return id
	}

	ledgerAssetUGX := createLedger("rpt-asset-ugx", ledgerv1.LedgerType_ASSET)
	ledgerIncomeUGX := createLedger("rpt-income-ugx", ledgerv1.LedgerType_INCOME)
	ledgerAssetUSD := createLedger("rpt-asset-usd", ledgerv1.LedgerType_ASSET)
	ledgerIncomeUSD := createLedger("rpt-income-usd", ledgerv1.LedgerType_INCOME)

	accounts := []acctFixture{
		{"rpt-cash-ugx", ledgerAssetUGX, models.LedgerTypeAsset, "UGX"},
		{"rpt-rev-ugx", ledgerIncomeUGX, models.LedgerTypeIncome, "UGX"},
		{"rpt-cash-usd", ledgerAssetUSD, models.LedgerTypeAsset, "USD"},
		{"rpt-rev-usd", ledgerIncomeUSD, models.LedgerTypeIncome, "USD"},
	}

	for _, acc := range accounts {
		_, err := res.AccountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
			Id:       acc.id,
			LedgerId: acc.ledgerID,
			Currency: acc.currency,
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"name": {Kind: &structpb.Value_StringValue{StringValue: acc.id}},
				},
			},
		})
		rs.Require().NoError(err)
	}
}

// postBalanced is a tiny helper that posts a cleared NORMAL balanced
// transaction with a debit on debitAcc and a credit on creditAcc for the
// given amount. Returns the resulting transaction id.
func (rs *ReportsSuite) postBalanced(
	ctx context.Context, res *tests.ServiceResources,
	txnID, currency, debitAcc, creditAcc string, amount int64, transactedAt time.Time,
) {
	txn := &models.Transaction{
		BaseModel:       data.BaseModel{ID: txnID},
		Currency:        currency,
		TransactionType: ledgerv1.TransactionType_NORMAL.String(),
		TransactedAt:    transactedAt,
		ClearedAt:       transactedAt,
		Entries: []*models.TransactionEntry{
			{AccountID: debitAcc, Credit: false, Amount: decimalx.NewFromInt64(amount).Ptr()},
			{AccountID: creditAcc, Credit: true, Amount: decimalx.NewFromInt64(amount).Ptr()},
		},
	}
	_, err := res.TransactionBusiness.Transact(ctx, txn)
	rs.Require().NoError(err)
}

// TestTrialBalanceIsBalancedPerCurrency posts a mix of UGX and USD
// transactions and proves that the textbook accounting invariant —
// total_debits = total_credits within each currency — holds across the
// posted set. Also verifies that cross-currency totals do not contaminate
// each other.
func (rs *ReportsSuite) TestTrialBalanceIsBalancedPerCurrency() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)
		rs.setupFixtures(ctx, res)

		now := time.Now().UTC()
		rs.postBalanced(ctx, res, "tb_ugx_1", "UGX", "rpt-cash-ugx", "rpt-rev-ugx", 100, now)
		rs.postBalanced(ctx, res, "tb_ugx_2", "UGX", "rpt-cash-ugx", "rpt-rev-ugx", 250, now)
		rs.postBalanced(ctx, res, "tb_usd_1", "USD", "rpt-cash-usd", "rpt-rev-usd", 70, now)

		report, err := res.ReportBusiness.TrialBalance(ctx, models.TrialBalanceParams{})
		require.NoError(t, err)
		require.NotNil(t, report)
		require.Len(t, report.Totals, 2, "expected one total per currency")

		totalsByCurrency := map[string]*business.TrialBalanceTotal{}
		for _, total := range report.Totals {
			totalsByCurrency[total.Currency] = total
		}

		ugx := totalsByCurrency["UGX"]
		require.NotNil(t, ugx, "UGX total must be present")
		require.True(t, ugx.IsBalanced, "UGX trial balance must balance")
		assertDecEqual(t, decimalx.NewFromInt64(350), ugx.TotalDebits)
		assertDecEqual(t, decimalx.NewFromInt64(350), ugx.TotalCredits)

		usd := totalsByCurrency["USD"]
		require.NotNil(t, usd, "USD total must be present")
		require.True(t, usd.IsBalanced, "USD trial balance must balance")
		assertDecEqual(t, decimalx.NewFromInt64(70), usd.TotalDebits)
		assertDecEqual(t, decimalx.NewFromInt64(70), usd.TotalCredits)
	})
}

// TestTrialBalanceCurrencyFilter confirms the Currency filter scopes the
// result and totals correctly — no foreign currency entries leak.
func (rs *ReportsSuite) TestTrialBalanceCurrencyFilter() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)
		rs.setupFixtures(ctx, res)

		now := time.Now().UTC()
		rs.postBalanced(ctx, res, "cf_ugx_1", "UGX", "rpt-cash-ugx", "rpt-rev-ugx", 100, now)
		rs.postBalanced(ctx, res, "cf_usd_1", "USD", "rpt-cash-usd", "rpt-rev-usd", 200, now)

		report, err := res.ReportBusiness.TrialBalance(ctx, models.TrialBalanceParams{Currency: "USD"})
		require.NoError(t, err)
		require.Len(t, report.Totals, 1)
		assert.Equal(t, "USD", report.Totals[0].Currency)
		assertDecEqual(t, decimalx.NewFromInt64(200), report.Totals[0].TotalDebits)

		for _, line := range report.Lines {
			assert.Equal(t, "USD", line.Currency, "currency filter must exclude non-matching lines")
		}
	})
}

// TestTrialBalanceAsOfExcludesLaterPostings verifies the AsOf upper bound
// only includes transactions transacted at or before the supplied instant.
func (rs *ReportsSuite) TestTrialBalanceAsOfExcludesLaterPostings() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)
		rs.setupFixtures(ctx, res)

		early := time.Now().UTC().Add(-2 * time.Hour)
		late := time.Now().UTC()
		rs.postBalanced(ctx, res, "aso_1", "UGX", "rpt-cash-ugx", "rpt-rev-ugx", 100, early)
		rs.postBalanced(ctx, res, "aso_2", "UGX", "rpt-cash-ugx", "rpt-rev-ugx", 200, late)

		midpoint := early.Add(time.Hour)
		report, err := res.ReportBusiness.TrialBalance(ctx, models.TrialBalanceParams{
			Currency: "UGX",
			AsOf:     &midpoint,
		})
		require.NoError(t, err)
		require.Len(t, report.Totals, 1)
		assertDecEqual(t, decimalx.NewFromInt64(100), report.Totals[0].TotalDebits,
			"as-of must exclude the later 200 posting")
	})
}

// TestAccountStatementRunningBalance verifies the per-entry running balance,
// totals and closing balance match the expected double-entry effect of the
// posted entries.
func (rs *ReportsSuite) TestAccountStatementRunningBalance() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)
		rs.setupFixtures(ctx, res)

		base := time.Now().UTC().Add(-3 * time.Hour)
		// Three +100 debits on the asset account, evenly spaced.
		for i := range 3 {
			rs.postBalanced(ctx, res,
				fmt.Sprintf("stmt_%d", i),
				"UGX", "rpt-cash-ugx", "rpt-rev-ugx", 100,
				base.Add(time.Duration(i)*time.Hour))
		}

		stmt, err := res.ReportBusiness.AccountStatement(ctx, models.StatementParams{
			AccountID: "rpt-cash-ugx",
		})
		require.NoError(t, err)
		require.NotNil(t, stmt)
		require.Len(t, stmt.Entries, 3)
		assert.Equal(t, "UGX", stmt.Currency)
		assertDecEqual(t, decimalx.Zero(), stmt.OpeningBalance)
		assertDecEqual(t, decimalx.NewFromInt64(300), stmt.ClosingBalance)
		assertDecEqual(t, decimalx.NewFromInt64(300), stmt.TotalDebits)
		assertDecEqual(t, decimalx.Zero(), stmt.TotalCredits)

		for i, entry := range stmt.Entries {
			expected := decimalx.NewFromInt64(int64((i + 1) * 100))
			assertDecEqual(t, expected, entry.RunningBalance,
				"running balance after entry %d should be %s", i, expected)
		}
	})
}

// TestAccountStatementOpeningBalanceCarry verifies entries strictly before
// the From date contribute to OpeningBalance and are not duplicated in the
// returned entry list. Pre-period activity should surface only as opening
// balance, not as page rows.
func (rs *ReportsSuite) TestAccountStatementOpeningBalanceCarry() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)
		rs.setupFixtures(ctx, res)

		ancient := time.Now().UTC().Add(-48 * time.Hour)
		recent := time.Now().UTC().Add(-1 * time.Hour)
		rs.postBalanced(ctx, res, "ob_old", "UGX", "rpt-cash-ugx", "rpt-rev-ugx", 500, ancient)
		rs.postBalanced(ctx, res, "ob_new", "UGX", "rpt-cash-ugx", "rpt-rev-ugx", 70, recent)

		windowStart := time.Now().UTC().Add(-12 * time.Hour)
		stmt, err := res.ReportBusiness.AccountStatement(ctx, models.StatementParams{
			AccountID: "rpt-cash-ugx",
			From:      &windowStart,
		})
		require.NoError(t, err)
		require.Len(t, stmt.Entries, 1, "only the recent posting should appear in the page")
		assertDecEqual(t, decimalx.NewFromInt64(500), stmt.OpeningBalance,
			"the ancient posting should surface only as opening balance")
		assertDecEqual(t, decimalx.NewFromInt64(570), stmt.ClosingBalance)
		assertDecEqual(t, decimalx.NewFromInt64(570), stmt.Entries[0].RunningBalance)
	})
}

func TestReportsSuite(t *testing.T) {
	suite.Run(t, new(ReportsSuite))
}
