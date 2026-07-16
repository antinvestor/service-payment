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
	"fmt"

	"github.com/antinvestor/service-payments/apps/billing/service/observability"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/pitabwire/util"
)

// SettlementSweepResult summarises one sweep pass.
type SettlementSweepResult struct {
	Candidates int
	Settled    int
	Skipped    int
	Errors     int
}

// SettlementSweeper auto-confirms completed checkout sessions for open invoices
// when the payer abandons the browser return/confirm path.
type SettlementSweeper interface {
	Sweep(ctx context.Context) (*SettlementSweepResult, error)
}

type settlementSweeper struct {
	invoiceRepo repository.InvoiceRepository
	collection  CollectionBusiness
	batchSize   int
	obs         *observability.Metrics
}

// NewSettlementSweeper constructs a sweeper. batchSize defaults to 50 when <= 0.
func NewSettlementSweeper(
	invoiceRepo repository.InvoiceRepository,
	collection CollectionBusiness,
	batchSize int,
) SettlementSweeper {
	if batchSize <= 0 {
		batchSize = 50
	}
	return &settlementSweeper{
		invoiceRepo: invoiceRepo,
		collection:  collection,
		batchSize:   batchSize,
		obs:         observability.NewMetrics(),
	}
}

// Sweep looks up ISSUED invoices with a stored checkoutSessionRef and attempts
// ConfirmPayment. Incomplete sessions are skipped (ErrCheckoutNotCompleted).
//
//nolint:nonamedreturns // named err captured by deferred span-end closure
func (s *settlementSweeper) Sweep(ctx context.Context) (result *SettlementSweepResult, err error) {
	ctx, span := s.obs.StartSpan(ctx, "SettlementSweep")
	defer func() { s.obs.EndSpan(ctx, span, err) }()

	if s.collection == nil {
		return nil, errors.New("collection business not configured")
	}

	invoices, err := s.invoiceRepo.ListOpenWithCheckoutSession(ctx, s.batchSize)
	if err != nil {
		return nil, fmt.Errorf("list open invoices with checkout session: %w", err)
	}

	out := &SettlementSweepResult{Candidates: len(invoices)}
	for _, inv := range invoices {
		sessionRef, _ := inv.Data[InvoiceDataCheckoutSessionRef].(string)
		if sessionRef == "" {
			out.Skipped++
			continue
		}

		confirmed, confirmErr := s.collection.ConfirmPayment(ctx, sessionRef)
		if confirmErr != nil {
			if errors.Is(confirmErr, ErrCheckoutNotCompleted) {
				out.Skipped++
				continue
			}
			out.Errors++
			util.Log(ctx).
				WithError(confirmErr).
				WithField("invoice_id", inv.GetID()).
				WithField("session_ref", sessionRef).
				Warn("settlement sweep confirm failed")
			continue
		}
		if confirmed != nil && confirmed.Paid {
			out.Settled++
			util.Log(ctx).
				WithField("invoice_id", inv.GetID()).
				WithField("session_ref", sessionRef).
				Info("settlement sweep settled invoice")
		} else {
			out.Skipped++
		}
	}

	return out, nil
}
