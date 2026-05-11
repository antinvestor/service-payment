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

	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/pitabwire/frame/data"
)

// CreateBook creates a new independent accounting scope (book).
// Books represent platform/group/customer/merchant/agent/branch entities
// whose entries balance independently. Optional parent_id supports
// organisation → groups → members hierarchies for consolidated reporting.
func (ledgerSrv *LedgerServer) CreateBook(
	ctx context.Context,
	req *connect.Request[ledgerv1.CreateBookRequest],
) (*connect.Response[ledgerv1.CreateBookResponse], error) {
	dataMap := data.JSONMap{}
	if req.Msg.GetData() != nil {
		dataMap = dataMap.FromProtoStruct(req.Msg.GetData())
	}

	var parentID *string
	if req.Msg.GetParentId() != "" {
		p := req.Msg.GetParentId()
		parentID = &p
	}

	created, err := ledgerSrv.Book.CreateBook(ctx,
		req.Msg.GetName(),
		req.Msg.GetType(),
		req.Msg.GetCurrency(),
		parentID,
		dataMap,
	)
	if err != nil {
		return nil, ToConnectError(err)
	}

	// Honour the caller-supplied ID when one was provided; the model
	// already minted an xid if not. We re-fetch to ensure the returned
	// payload mirrors what is persisted.
	return connect.NewResponse(&ledgerv1.CreateBookResponse{Data: created.ToAPI()}), nil
}

// GetBook fetches a single book by id within the caller's tenancy scope.
func (ledgerSrv *LedgerServer) GetBook(
	ctx context.Context,
	req *connect.Request[ledgerv1.GetBookRequest],
) (*connect.Response[ledgerv1.GetBookResponse], error) {
	book, err := ledgerSrv.Book.GetBook(ctx, req.Msg.GetId())
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(&ledgerv1.GetBookResponse{Data: book.ToAPI()}), nil
}

// ListBooksByType returns all active books of a conventional type in
// the caller's tenancy. Sort order is most-recently-created first to
// match the operator-facing UI pattern.
func (ledgerSrv *LedgerServer) ListBooksByType(
	ctx context.Context,
	req *connect.Request[ledgerv1.ListBooksByTypeRequest],
) (*connect.Response[ledgerv1.ListBooksByTypeResponse], error) {
	books, err := ledgerSrv.Book.ListBooksByType(ctx, req.Msg.GetType())
	if err != nil {
		return nil, ToConnectError(err)
	}
	apiBooks := make([]*ledgerv1.Book, len(books))
	for i, b := range books {
		apiBooks[i] = b.ToAPI()
	}
	return connect.NewResponse(&ledgerv1.ListBooksByTypeResponse{Data: apiBooks}), nil
}
