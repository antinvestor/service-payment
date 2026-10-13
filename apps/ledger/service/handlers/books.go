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
	"github.com/pitabwire/frame/v3/data"
)

// CreateBook creates a new independent accounting scope (book).
// Books represent platform/group/customer/merchant/agent/branch entities
// whose entries balance independently. Optional parent_id supports
// organisation → groups → members hierarchies for consolidated reporting.
// A caller-supplied id names the book (an id is generated when omitted), and
// repeating the same create returns the existing book.
func (ledgerSrv *LedgerServer) CreateBook(
	ctx context.Context,
	req *ledgerv1.CreateBookRequest,
) (*ledgerv1.CreateBookResponse, error) {
	dataMap := data.JSONMap{}
	if req.GetData() != nil {
		dataMap = dataMap.FromProtoStruct(req.GetData())
	}

	var parentID *string
	if req.GetParentId() != "" {
		p := req.GetParentId()
		parentID = &p
	}

	created, err := ledgerSrv.Book.CreateBook(ctx,
		req.GetId(),
		req.GetName(),
		req.GetType(),
		req.GetCurrency(),
		parentID,
		dataMap,
	)
	if err != nil {
		return nil, ToConnectError(err)
	}

	return &ledgerv1.CreateBookResponse{Data: created.ToAPI()}, nil
}

// GetBook fetches a single book by id within the caller's tenancy scope.
func (ledgerSrv *LedgerServer) GetBook(
	ctx context.Context,
	req *ledgerv1.GetBookRequest,
) (*ledgerv1.GetBookResponse, error) {
	book, err := ledgerSrv.Book.GetBook(ctx, req.GetId())
	if err != nil {
		return nil, ToConnectError(err)
	}
	return &ledgerv1.GetBookResponse{Data: book.ToAPI()}, nil
}

// ListBooksByType returns all active books of a conventional type in
// the caller's tenancy. Sort order is most-recently-created first to
// match the operator-facing UI pattern.
func (ledgerSrv *LedgerServer) ListBooksByType(
	ctx context.Context,
	req *ledgerv1.ListBooksByTypeRequest,
) (*ledgerv1.ListBooksByTypeResponse, error) {
	books, err := ledgerSrv.Book.ListBooksByType(ctx, req.GetType())
	if err != nil {
		return nil, ToConnectError(err)
	}
	apiBooks := make([]*ledgerv1.Book, len(books))
	for i, b := range books {
		apiBooks[i] = b.ToAPI()
	}
	return &ledgerv1.ListBooksByTypeResponse{Data: apiBooks}, nil
}
