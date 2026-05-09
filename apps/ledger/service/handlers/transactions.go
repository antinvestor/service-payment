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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
)

// SearchTransactions finds transactions matching specified criteria.
// Supports filtering by date range, account, currency, and status.
func (ledgerSrv *LedgerServer) SearchTransactions(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[ledgerv1.SearchTransactionsResponse],
) error {
	return ToConnectError(ledgerSrv.Transaction.SearchTransactions(
		ctx,
		req.Msg,
		func(_ context.Context, batch []*ledgerv1.Transaction) error {
			return stream.Send(&ledgerv1.SearchTransactionsResponse{
				Data: batch,
			})
		},
	))
}

// CreateTransaction creates a new double-entry transaction.
// All entries must be balanced (sum of debits = sum of credits).
func (ledgerSrv *LedgerServer) CreateTransaction(
	ctx context.Context,
	req *connect.Request[ledgerv1.CreateTransactionRequest],
) (*connect.Response[ledgerv1.CreateTransactionResponse], error) {
	createdTransaction, err := ledgerSrv.Transaction.CreateTransaction(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}

	response := &ledgerv1.CreateTransactionResponse{
		Data: createdTransaction,
	}

	return connect.NewResponse(response), nil
}

// ReverseTransaction reverses a transaction by creating offsetting entries.
// Creates a new REVERSAL transaction that negates the original.
func (ledgerSrv *LedgerServer) ReverseTransaction(
	ctx context.Context,
	req *connect.Request[ledgerv1.ReverseTransactionRequest],
) (*connect.Response[ledgerv1.ReverseTransactionResponse], error) {
	reversedTransaction, err := ledgerSrv.Transaction.ReverseTransaction(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}

	response := &ledgerv1.ReverseTransactionResponse{
		Data: reversedTransaction,
	}

	return connect.NewResponse(response), nil
}

// UpdateTransaction updates a transaction's metadata.
// Entries and amounts cannot be changed after creation.
func (ledgerSrv *LedgerServer) UpdateTransaction(
	ctx context.Context,
	req *connect.Request[ledgerv1.UpdateTransactionRequest],
) (*connect.Response[ledgerv1.UpdateTransactionResponse], error) {
	updatedTransaction, err := ledgerSrv.Transaction.UpdateTransaction(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}

	response := &ledgerv1.UpdateTransactionResponse{
		Data: updatedTransaction,
	}

	return connect.NewResponse(response), nil
}
