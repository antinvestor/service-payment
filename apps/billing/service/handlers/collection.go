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

	"connectrpc.com/connect/v2"
	collectionv1 "github.com/antinvestor/service-payments/apps/billing/gen/collection/v1"
	"github.com/antinvestor/service-payments/apps/billing/gen/collection/v1/collectionv1connect"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
)

// CollectionServer implements the simplified CollectionService Connect RPC handler.
type CollectionServer struct {
	Collection business.CollectionBusiness
	Providers  business.PolarSubscriptionBusiness
}

// NewCollectionServer creates a CollectionService handler. providers may be
// omitted, in which case SyncProviderSubscription answers Unimplemented.
func NewCollectionServer(
	collection business.CollectionBusiness,
	providers ...business.PolarSubscriptionBusiness,
) collectionv1connect.CollectionServiceHandler {
	s := &CollectionServer{Collection: collection}
	for _, p := range providers {
		if p != nil {
			s.Providers = p
		}
	}
	return s
}

// SyncProviderSubscription mirrors provider-managed subscription state.
func (s *CollectionServer) SyncProviderSubscription(
	ctx context.Context,
	req *collectionv1.SyncProviderSubscriptionRequest,
) (*collectionv1.SyncProviderSubscriptionResponse, error) {
	if s.Providers == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, "provider subscriptions are not configured")
	}
	sub, err := s.Providers.MirrorSubscriptionState(ctx, business.MirrorInput{
		Provider:            req.GetProvider(),
		Event:               req.GetEvent(),
		SubscriptionID:      req.GetSubscriptionId(),
		PolarSubscriptionID: req.GetExternalSubscriptionId(),
		ExternalCustomerID:  req.GetExternalCustomerId(),
		ExternalProductID:   req.GetExternalProductId(),
		ProfileID:           req.GetProfileId(),
		PlanID:              req.GetPlanId(),
		PolarStatus:         req.GetStatus(),
		CancelAtPeriodEnd:   req.GetCancelAtPeriodEnd(),
		CurrentPeriodStart:  req.GetCurrentPeriodStart(),
		CurrentPeriodEnd:    req.GetCurrentPeriodEnd(),
		ExternalOrderID:     req.GetExternalOrderId(),
		OccurredAt:          req.GetOccurredAt(),
	})
	if err != nil {
		if errors.Is(err, business.ErrStaleProviderEvent) && sub != nil {
			return &collectionv1.SyncProviderSubscriptionResponse{
				SubscriptionId: sub.GetID(), SubscriptionState: sub.State, Applied: false,
			}, nil
		}
		if errors.Is(err, business.ErrPolarSubscriptionNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err.Error()).WithCause(err)
		}
		return nil, toConnectError(err)
	}
	return &collectionv1.SyncProviderSubscriptionResponse{
		SubscriptionId: sub.GetID(), SubscriptionState: sub.State, Applied: true,
	}, nil
}

func (s *CollectionServer) CollectPayment(
	ctx context.Context,
	req *collectionv1.CollectPaymentRequest,
) (*collectionv1.CollectPaymentResponse, error) {
	result, err := s.Collection.CollectPayment(ctx, business.CollectPaymentInput{
		InvoiceID: req.GetInvoiceId(),
		ReturnURL: req.GetReturnUrl(),
		Methods:   req.GetMethods(),
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return &collectionv1.CollectPaymentResponse{
		Data: collectionResultToProto(result),
	}, nil
}

func (s *CollectionServer) StartSubscription(
	ctx context.Context,
	req *collectionv1.StartSubscriptionRequest,
) (*collectionv1.StartSubscriptionResponse, error) {
	result, err := s.Collection.StartSubscription(ctx, business.StartSubscriptionInput{
		ProfileID:          req.GetProfileId(),
		PlanID:             req.GetPlanId(),
		CatalogVersionID:   req.GetCatalogVersionId(),
		Currency:           req.GetCurrency(),
		ReturnURL:          req.GetReturnUrl(),
		PayerDisplayName:   req.GetPayerDisplayName(),
		Methods:            req.GetMethods(),
		ExternalEntityID:   req.GetExternalEntityId(),
		ExternalEntityType: req.GetExternalEntityType(),
		IntegrationRouteID: req.GetIntegrationRouteId(),
		Metadata:           req.GetMetadata(),
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return &collectionv1.StartSubscriptionResponse{
		Data: collectionResultToProto(result),
	}, nil
}

func (s *CollectionServer) ConfirmPayment(
	ctx context.Context,
	req *collectionv1.ConfirmPaymentRequest,
) (*collectionv1.ConfirmPaymentResponse, error) {
	result, err := s.Collection.ConfirmPayment(ctx, req.GetSessionRef())
	if err != nil {
		if errors.Is(err, business.ErrCheckoutNotCompleted) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err.Error()).WithCause(err)
		}
		return nil, toConnectError(err)
	}
	return &collectionv1.ConfirmPaymentResponse{
		InvoiceId:         result.InvoiceID,
		InvoiceState:      result.InvoiceState,
		SubscriptionId:    result.SubscriptionID,
		SubscriptionState: result.SubscriptionState,
		Paid:              result.Paid,
	}, nil
}

func (s *CollectionServer) CancelSubscription(
	ctx context.Context,
	req *collectionv1.CancelSubscriptionRequest,
) (*collectionv1.CancelSubscriptionResponse, error) {
	result, err := s.Collection.CancelSubscription(ctx, req.GetSubscriptionId())
	if err != nil {
		return nil, toConnectError(err)
	}
	return &collectionv1.CancelSubscriptionResponse{
		SubscriptionId:    result.SubscriptionID,
		SubscriptionState: result.SubscriptionState,
		VoidedInvoiceId:   result.VoidedInvoiceID,
	}, nil
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
