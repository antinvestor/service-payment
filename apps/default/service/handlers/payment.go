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
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v2/v1/ledgerv1connect"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	"buf.build/gen/go/antinvestor/tenancy/connectrpc/go/v2/tenancy/v1/tenancyv1connect"
	"github.com/antinvestor/service-payments/apps/default/service/business"
	"github.com/antinvestor/service-payments/pkg/connecterr"
)

var _ paymentv1connect.PaymentServiceHandler = (*PaymentServer)(nil)

type PaymentServer struct {
	PaymentBusiness business.PaymentBusiness
	ProfileCli      profilev1connect.ProfileServiceClient
	LedgerCli       ledgerv1connect.LedgerServiceClient
	TenancyCli      tenancyv1connect.TenancyServiceClient
}

// NewPaymentServer creates a new PaymentServer with the required dependencies.
func NewPaymentServer(
	paymentBusiness business.PaymentBusiness,
	profileCli profilev1connect.ProfileServiceClient,
	ledgerCli ledgerv1connect.LedgerServiceClient,
	tenancyCli tenancyv1connect.TenancyServiceClient,
) *PaymentServer {
	return &PaymentServer{
		PaymentBusiness: paymentBusiness,
		ProfileCli:      profileCli,
		TenancyCli:      tenancyCli,
		LedgerCli:       ledgerCli,
	}
}

func (ps *PaymentServer) Send(
	ctx context.Context,
	req *paymentv1.SendRequest,
) (*paymentv1.SendResponse, error) {
	response, err := ps.PaymentBusiness.Send(ctx, req.GetData())
	if err != nil {
		return nil, connecterr.From(err)
	}
	return &paymentv1.SendResponse{Data: response}, nil
}

func (ps *PaymentServer) Status(
	ctx context.Context,
	req *commonv1.StatusRequest,
) (*commonv1.StatusResponse, error) {
	response, err := ps.PaymentBusiness.Status(ctx, req)
	if err != nil {
		return nil, connecterr.From(err)
	}
	return response, nil
}

// StatusUpdate request to allow continuation of payment processing.
func (ps *PaymentServer) StatusUpdate(
	ctx context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusUpdateResponse, error) {
	response, err := ps.PaymentBusiness.StatusUpdate(ctx, req)
	if err != nil {
		return nil, connecterr.From(err)
	}

	return &commonv1.StatusUpdateResponse{Data: response}, nil
}

// Release method for releasing queued payments and returns if payment status if released.
func (ps *PaymentServer) Release(
	ctx context.Context,
	req *paymentv1.ReleaseRequest,
) (*paymentv1.ReleaseResponse, error) {
	response, err := ps.PaymentBusiness.Release(ctx, req)

	if err != nil {
		return nil, connecterr.From(err)
	}

	return &paymentv1.ReleaseResponse{Data: response}, nil
}

// Receive method is for client request for particular Payment responses from system.
func (ps *PaymentServer) Receive(
	ctx context.Context,
	req *paymentv1.ReceiveRequest,
) (*paymentv1.ReceiveResponse, error) {
	response, err := ps.PaymentBusiness.Receive(ctx, req.GetData())

	if err != nil {
		return nil, connecterr.From(err)
	}

	return &paymentv1.ReceiveResponse{Data: response}, nil
}

// InitiatePrompt method for client request for particular Prompt responses from system.
func (ps *PaymentServer) InitiatePrompt(
	ctx context.Context,
	req *paymentv1.InitiatePromptRequest,
) (*paymentv1.InitiatePromptResponse, error) {
	response, err := ps.PaymentBusiness.InitiatePrompt(ctx, req)

	if err != nil {
		return nil, connecterr.From(err)
	}

	return &paymentv1.InitiatePromptResponse{Data: response}, nil
}

// CreatePaymentLink method for client request to create a payment link.
func (ps *PaymentServer) CreatePaymentLink(
	ctx context.Context,
	req *paymentv1.CreatePaymentLinkRequest,
) (*paymentv1.CreatePaymentLinkResponse, error) {
	response, err := ps.PaymentBusiness.CreatePaymentLink(ctx, req)

	if err != nil {
		return nil, connecterr.From(err)
	}

	return &paymentv1.CreatePaymentLinkResponse{Data: response}, nil
}

// Search method for searching payments based on query criteria.
func (ps *PaymentServer) Search(
	ctx context.Context,
	req *commonv1.SearchRequest,
	stream paymentv1connect.PaymentServiceSearchServerStream,
) error {
	resultPipe, err := ps.PaymentBusiness.Search(ctx, req)
	if err != nil {
		return connecterr.From(err)
	}

	// Read from the result pipe and stream back to client
	for {
		result, ok := resultPipe.ReadResult(ctx)
		if !ok {
			return nil
		}

		if result.IsError() {
			return result.Error()
		}

		// Stream each payment back to the client
		err = stream.Send(&paymentv1.SearchResponse{Data: result.Item()})
		if err != nil {
			return connecterr.From(err)
		}
	}
}

func (ps *PaymentServer) Reconcile(
	ctx context.Context,
	req *paymentv1.ReconcileRequest,
) (*paymentv1.ReconcileResponse, error) {
	response, err := ps.PaymentBusiness.Reconcile(ctx, req)

	if err != nil {
		return nil, connecterr.From(err)
	}

	return response, nil
}
