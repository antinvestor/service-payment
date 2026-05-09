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

// SearchAccounts finds accounts matching specified criteria.
// Supports filtering by ledger, balance range, and custom properties.
func (ledgerSrv *LedgerServer) SearchAccounts(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[ledgerv1.SearchAccountsResponse],
) error {
	return ToConnectError(
		ledgerSrv.Account.SearchAccounts(ctx, req.Msg, func(_ context.Context, batch []*ledgerv1.Account) error {
			return stream.Send(&ledgerv1.SearchAccountsResponse{
				Data: batch,
			})
		}),
	)
}

// CreateAccount creates a new account within a ledger.
// Each account tracks balances and transaction history.
func (ledgerSrv *LedgerServer) CreateAccount(
	ctx context.Context,
	req *connect.Request[ledgerv1.CreateAccountRequest],
) (*connect.Response[ledgerv1.CreateAccountResponse], error) {
	// Create the account using business layer
	createdAccount, err := ledgerSrv.Account.CreateAccount(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}

	response := &ledgerv1.CreateAccountResponse{
		Data: createdAccount,
	}

	return connect.NewResponse(response), nil
}

// UpdateAccount updates an existing account's metadata.
// Balances are updated through transactions, not directly.
func (ledgerSrv *LedgerServer) UpdateAccount(
	ctx context.Context,
	req *connect.Request[ledgerv1.UpdateAccountRequest],
) (*connect.Response[ledgerv1.UpdateAccountResponse], error) {
	// Update the account using business layer
	updatedAccount, err := ledgerSrv.Account.UpdateAccount(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}

	response := &ledgerv1.UpdateAccountResponse{
		Data: updatedAccount,
	}

	return connect.NewResponse(response), nil
}
