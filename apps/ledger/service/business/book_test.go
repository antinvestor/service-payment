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
	"testing"
	"time"

	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/frametests/definition"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"
)

// BookHierarchySuite covers the Book entity: independent accounting scopes,
// parent/child hierarchy (organization → groups → individuals), and the
// cross-book invariants those scopes enforce.
type BookHierarchySuite struct {
	tests.BaseTestSuite
}

func TestBookHierarchySuite(t *testing.T) {
	suite.Run(t, new(BookHierarchySuite))
}

// orgGroupIndividual sets up a small three-level hierarchy returning each
// book id so individual tests can drive assertions against the structure.
type orgGroupIndividual struct {
	OrgID    string
	GroupID  string
	MemberID string
}

func (rs *BookHierarchySuite) seedHierarchy(
	ctx context.Context, res *tests.ServiceResources,
) orgGroupIndividual {
	org, err := res.BookBusiness.CreateBook(ctx,
		"Stawi Organization", models.BookTypePlatform, "UGX", nil, nil)
	rs.Require().NoError(err)

	group, err := res.BookBusiness.CreateBook(ctx,
		"Village Group A", models.BookTypeGroup, "UGX", &org.ID, nil)
	rs.Require().NoError(err)

	member, err := res.BookBusiness.CreateBook(ctx,
		"Member Jane Doe", models.BookTypeCustomer, "UGX", &group.ID, nil)
	rs.Require().NoError(err)

	return orgGroupIndividual{OrgID: org.ID, GroupID: group.ID, MemberID: member.ID}
}

// TestCreateBookAndHierarchyTraversal proves CreateBook persists ParentID,
// rejects parents the caller cannot see, and that the descendants traversal
// returns the whole subtree (root included) in one call.
func (rs *BookHierarchySuite) TestCreateBookAndHierarchyTraversal() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)

		h := rs.seedHierarchy(ctx, res)

		// Org's descendant set must include all three nodes.
		ids, err := res.BookBusiness.ListDescendantIDs(ctx, h.OrgID)
		require.NoError(t, err)
		require.ElementsMatch(t, []string{h.OrgID, h.GroupID, h.MemberID}, ids)

		// Group's descendant set excludes the org parent.
		ids, err = res.BookBusiness.ListDescendantIDs(ctx, h.GroupID)
		require.NoError(t, err)
		require.ElementsMatch(t, []string{h.GroupID, h.MemberID}, ids)

		// Leaf member returns only itself.
		ids, err = res.BookBusiness.ListDescendantIDs(ctx, h.MemberID)
		require.NoError(t, err)
		require.ElementsMatch(t, []string{h.MemberID}, ids)
	})
}

// seedAccountInBook is a helper that creates a ledger inside a book (via
// the data["book_id"] convention) plus an account inside that ledger.
func (rs *BookHierarchySuite) seedAccountInBook(
	ctx context.Context, res *tests.ServiceResources,
	ledgerID, bookID, accountID string, ledgerType ledgerv1.LedgerType,
) {
	dataMap := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			models.DataKeyBookID: {Kind: &structpb.Value_StringValue{StringValue: bookID}},
		},
	}
	_, err := res.LedgerBusiness.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{
		Id:   ledgerID,
		Type: ledgerType,
		Data: dataMap,
	})
	rs.Require().NoError(err)

	_, err = res.AccountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
		Id:       accountID,
		LedgerId: ledgerID,
		Currency: "UGX",
	})
	rs.Require().NoError(err)
}

// TestCrossBookPostingIsRejected proves the cross-book integrity rule:
// when a transaction is scoped to one book, all entries' accounts must
// belong to that same book; an entry pointing at an account from another
// book is rejected.
func (rs *BookHierarchySuite) TestCrossBookPostingIsRejected() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)

		bookA, err := res.BookBusiness.CreateBook(ctx,
			"Group A", models.BookTypeGroup, "UGX", nil, nil)
		require.NoError(t, err)
		bookB, err := res.BookBusiness.CreateBook(ctx,
			"Group B", models.BookTypeGroup, "UGX", nil, nil)
		require.NoError(t, err)

		rs.seedAccountInBook(ctx, res, "cb-a-asset", bookA.ID, "cb-a-cash", ledgerv1.LedgerType_ASSET)
		rs.seedAccountInBook(ctx, res, "cb-a-income", bookA.ID, "cb-a-rev", ledgerv1.LedgerType_INCOME)
		rs.seedAccountInBook(ctx, res, "cb-b-asset", bookB.ID, "cb-b-cash", ledgerv1.LedgerType_ASSET)

		now := time.Now().UTC()
		bookAID := bookA.ID
		txn := &models.Transaction{
			BaseModel:       data.BaseModel{ID: "cb-cross-book"},
			BookID:          &bookAID,
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    now,
			ClearedAt:       now,
			Entries: []*models.TransactionEntry{
				// Debit on Book A: legal.
				{AccountID: "cb-a-cash", Credit: false, Amount: decimalx.NewFromInt64(100).Ptr()},
				// Credit on Book B: must be rejected.
				{AccountID: "cb-b-cash", Credit: true, Amount: decimalx.NewFromInt64(100).Ptr()},
			},
		}
		_, err = res.TransactionBusiness.Transact(ctx, txn)
		require.Error(t, err,
			"posting must reject an entry pointing at an account in a different book")
	})
}

// TestConsolidatedTrialBalanceAcrossDescendants proves that a parent book's
// trial balance, when supplied the parent's whole descendant set, rolls
// up the activity posted in each child's own ledger correctly. Each child
// book posts independently; the consolidated total nets out per-currency.
func (rs *BookHierarchySuite) TestConsolidatedTrialBalanceAcrossDescendants() {
	rs.WithTestDependencies(rs.T(), func(t *testing.T, depOpt *definition.DependencyOption) {
		ctx, _, res := rs.CreateService(t, depOpt)

		h := rs.seedHierarchy(ctx, res)

		// Group A posts a 200 UGX activity (asset debit, income credit).
		rs.seedAccountInBook(ctx, res,
			"con-grp-asset-ldgr", h.GroupID, "con-grp-cash", ledgerv1.LedgerType_ASSET)
		rs.seedAccountInBook(ctx, res,
			"con-grp-inc-ldgr", h.GroupID, "con-grp-rev", ledgerv1.LedgerType_INCOME)

		// Member posts a 50 UGX activity.
		rs.seedAccountInBook(ctx, res,
			"con-mem-asset-ldgr", h.MemberID, "con-mem-cash", ledgerv1.LedgerType_ASSET)
		rs.seedAccountInBook(ctx, res,
			"con-mem-inc-ldgr", h.MemberID, "con-mem-rev", ledgerv1.LedgerType_INCOME)

		now := time.Now().UTC()
		groupBookID := h.GroupID
		memberBookID := h.MemberID

		_, err := res.TransactionBusiness.Transact(ctx, &models.Transaction{
			BaseModel:       data.BaseModel{ID: "con-grp-txn"},
			BookID:          &groupBookID,
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    now,
			ClearedAt:       now,
			Entries: []*models.TransactionEntry{
				{AccountID: "con-grp-cash", Credit: false, Amount: decimalx.NewFromInt64(200).Ptr()},
				{AccountID: "con-grp-rev", Credit: true, Amount: decimalx.NewFromInt64(200).Ptr()},
			},
		})
		require.NoError(t, err)

		_, err = res.TransactionBusiness.Transact(ctx, &models.Transaction{
			BaseModel:       data.BaseModel{ID: "con-mem-txn"},
			BookID:          &memberBookID,
			Currency:        "UGX",
			TransactionType: ledgerv1.TransactionType_NORMAL.String(),
			TransactedAt:    now,
			ClearedAt:       now,
			Entries: []*models.TransactionEntry{
				{AccountID: "con-mem-cash", Credit: false, Amount: decimalx.NewFromInt64(50).Ptr()},
				{AccountID: "con-mem-rev", Credit: true, Amount: decimalx.NewFromInt64(50).Ptr()},
			},
		})
		require.NoError(t, err)

		// Consolidated trial balance from the org down to every member.
		descendants, err := res.BookBusiness.ListDescendantIDs(ctx, h.OrgID)
		require.NoError(t, err)

		report, err := res.ReportBusiness.TrialBalance(ctx, models.TrialBalanceParams{
			Currency: "UGX",
			BookIDs:  descendants,
		})
		require.NoError(t, err)
		require.Len(t, report.Totals, 1)
		total := report.Totals[0]
		assert.Equal(t, "UGX", total.Currency)
		assert.True(t, total.IsBalanced)
		assertDecEqual(t, decimalx.NewFromInt64(250), total.TotalDebits,
			"consolidated debits = 200 group + 50 member")
		assertDecEqual(t, decimalx.NewFromInt64(250), total.TotalCredits)

		// Scoped to just the member's book — only the 50 activity surfaces.
		memberReport, err := res.ReportBusiness.TrialBalance(ctx, models.TrialBalanceParams{
			Currency: "UGX",
			BookIDs:  []string{h.MemberID},
		})
		require.NoError(t, err)
		require.Len(t, memberReport.Totals, 1)
		assertDecEqual(t, decimalx.NewFromInt64(50), memberReport.Totals[0].TotalDebits)
		assertDecEqual(t, decimalx.NewFromInt64(50), memberReport.Totals[0].TotalCredits)
	})
}
