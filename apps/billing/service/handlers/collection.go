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
	"errors"

	"connectrpc.com/connect"
	collectionv1 "github.com/antinvestor/service-payments/apps/billing/gen/collection/v1"
	"github.com/antinvestor/service-payments/apps/billing/gen/collection/v1/collectionv1connect"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
)

// CollectionServer implements the simplified CollectionService Connect RPC handler.
type CollectionServer struct {
	Collection business.CollectionBusiness
}

// NewCollectionServer creates a CollectionService handler.
func NewCollectionServer(collection business.CollectionBusiness) collectionv1connect.CollectionServiceHandler {
	return &CollectionServer{Collection: collection}
}

func (s *CollectionServer) CollectPayment(
	ctx context.Context,
	req *connect.Request[collectionv1.CollectPaymentRequest],
) (*connect.Response[collectionv1.CollectPaymentResponse], error) {
	result, err := s.Collection.CollectPayment(ctx, business.CollectPaymentInput{
		InvoiceID: req.Msg.GetInvoiceId(),
		ReturnURL: req.Msg.GetReturnUrl(),
		Methods:   req.Msg.GetMethods(),
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&collectionv1.CollectPaymentResponse{
		Data: collectionResultToProto(result),
	}), nil
}

func (s *CollectionServer) StartSubscription(
	ctx context.Context,
	req *connect.Request[collectionv1.StartSubscriptionRequest],
) (*connect.Response[collectionv1.StartSubscriptionResponse], error) {
	result, err := s.Collection.StartSubscription(ctx, business.StartSubscriptionInput{
		ProfileID:          req.Msg.GetProfileId(),
		PlanID:             req.Msg.GetPlanId(),
		CatalogVersionID:   req.Msg.GetCatalogVersionId(),
		Currency:           req.Msg.GetCurrency(),
		ReturnURL:          req.Msg.GetReturnUrl(),
		PayerDisplayName:   req.Msg.GetPayerDisplayName(),
		Methods:            req.Msg.GetMethods(),
		ExternalEntityID:   req.Msg.GetExternalEntityId(),
		ExternalEntityType: req.Msg.GetExternalEntityType(),
		IntegrationRouteID: req.Msg.GetIntegrationRouteId(),
		Metadata:           req.Msg.GetMetadata(),
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&collectionv1.StartSubscriptionResponse{
		Data: collectionResultToProto(result),
	}), nil
}

func (s *CollectionServer) ConfirmPayment(
	ctx context.Context,
	req *connect.Request[collectionv1.ConfirmPaymentRequest],
) (*connect.Response[collectionv1.ConfirmPaymentResponse], error) {
	result, err := s.Collection.ConfirmPayment(ctx, req.Msg.GetSessionRef())
	if err != nil {
		if errors.Is(err, business.ErrCheckoutNotCompleted) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&collectionv1.ConfirmPaymentResponse{
		InvoiceId:         result.InvoiceID,
		InvoiceState:      result.InvoiceState,
		SubscriptionId:    result.SubscriptionID,
		SubscriptionState: result.SubscriptionState,
		Paid:              result.Paid,
	}), nil
}

func (s *CollectionServer) CancelSubscription(
	ctx context.Context,
	req *connect.Request[collectionv1.CancelSubscriptionRequest],
) (*connect.Response[collectionv1.CancelSubscriptionResponse], error) {
	result, err := s.Collection.CancelSubscription(ctx, req.Msg.GetSubscriptionId())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&collectionv1.CancelSubscriptionResponse{
		SubscriptionId:    result.SubscriptionID,
		SubscriptionState: result.SubscriptionState,
		VoidedInvoiceId:   result.VoidedInvoiceID,
	}), nil
}

func collectionResultToProto(r *business.CollectionResult) *collectionv1.CollectionResult {
	if r == nil {
		return nil
	}
	return &collectionv1.CollectionResult{
		PageUrl:         r.PageURL,
		SessionRef:      r.SessionRef,
		InvoiceId:       r.InvoiceID,
		SubscriptionId:  r.SubscriptionID,
		AlreadyComplete: r.AlreadyComplete,
	}
}
