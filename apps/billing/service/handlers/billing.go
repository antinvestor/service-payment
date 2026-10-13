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
	"strconv"
	"strings"

	"buf.build/gen/go/antinvestor/billing/connectrpc/go/v2/v1/billingv1connect"
	billingv1 "buf.build/gen/go/antinvestor/billing/protocolbuffers/go/v1"
	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"connectrpc.com/connect/v2"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/util/decimalx"
)

// toConnectError translates application errors into appropriate ConnectRPC error codes.
func toConnectError(err error) error {
	if err == nil {
		return nil
	}

	var appErr apperrors.ApplicationError
	if !errors.As(err, &appErr) {
		// Business validation errors (plain Go errors) map to InvalidArgument
		if isBizValidationError(err) {
			return connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
		}
		return connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	code := appErr.ErrorCode()

	switch code - apperrors.DefaultCodeOffset {
	case apperrors.ErrorCodeCatalogVersionNotFound,
		apperrors.ErrorCodePlanNotFound,
		apperrors.ErrorCodeComponentNotFound,
		apperrors.ErrorCodeSubscriptionNotFound,
		apperrors.ErrorCodeInvoiceNotFound,
		apperrors.ErrorCodeCreditGrantNotFound,
		apperrors.ErrorCodeBillingRunNotFound:
		return connect.NewError(connect.CodeNotFound, appErr.Error()).WithCause(appErr)

	case apperrors.ErrorCodeCatalogVersionRetired,
		apperrors.ErrorCodeCatalogNotPublished,
		apperrors.ErrorCodeSubscriptionNotActive,
		apperrors.ErrorCodeInvoiceAlreadyIssued,
		apperrors.ErrorCodeInvoiceNotIssuable,
		apperrors.ErrorCodeInvoiceNotVoidable,
		apperrors.ErrorCodeInvoiceNotPayable,
		apperrors.ErrorCodeCreditExpired:
		return connect.NewError(connect.CodeFailedPrecondition, appErr.Error()).WithCause(appErr)

	case apperrors.ErrorCodeInvalidPricingModel,
		apperrors.ErrorCodeUsageEventInvalid,
		apperrors.ErrorCodeRatingFailed,
		apperrors.ErrorCodeDiscountInvalid:
		return connect.NewError(connect.CodeInvalidArgument, appErr.Error()).WithCause(appErr)

	case apperrors.ErrorCodeUsageEventDuplicate,
		apperrors.ErrorCodeBillingRunAlreadyExists:
		return connect.NewError(connect.CodeAlreadyExists, appErr.Error()).WithCause(appErr)

	case apperrors.ErrorCodeCreditInsufficientFunds:
		return connect.NewError(connect.CodeResourceExhausted, appErr.Error()).WithCause(appErr)

	case apperrors.ErrorCodeBillingRunFailed:
		return connect.NewError(connect.CodeInternal, appErr.Error()).WithCause(appErr)

	default:
		return connect.NewError(connect.CodeInternal, appErr.Error()).WithCause(appErr)
	}
}

// isBizValidationError checks if the error is a business validation error (plain Go error
// with "required" or "cannot" in the message, indicating a missing field or invalid state).
func isBizValidationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "required") || strings.Contains(msg, "cannot") || strings.Contains(msg, "invalid")
}

// BillingServer implements the billing ConnectRPC handler.
type BillingServer struct {
	Catalog        business.CatalogBusiness
	Subscription   business.SubscriptionBusiness
	Usage          business.UsageIngestionBusiness
	Metering       business.MeteringEngine
	Pricing        *business.PricingEngine
	Discount       business.DiscountEngine
	Credit         business.CreditEngine
	Invoice        business.InvoiceEngine
	Workflow       business.BillingWorkflow
	Ledger         business.LedgerIntegration
	CatalogSearch  repository.CatalogVersionRepository
	UsageSearch    repository.UsageEventRepository
	InvoiceSearch  repository.InvoiceRepository
	DiscountSearch repository.DiscountRepository
	SubSearch      repository.SubscriptionRepository
}

// NewBillingServer creates a new BillingServer with injected dependencies.
func NewBillingServer(
	catalog business.CatalogBusiness,
	subscription business.SubscriptionBusiness,
	usage business.UsageIngestionBusiness,
	metering business.MeteringEngine,
	pricing *business.PricingEngine,
	discount business.DiscountEngine,
	credit business.CreditEngine,
	invoice business.InvoiceEngine,
	workflow business.BillingWorkflow,
	ledger business.LedgerIntegration,
	catalogSearch repository.CatalogVersionRepository,
	usageSearch repository.UsageEventRepository,
	invoiceSearch repository.InvoiceRepository,
	discountSearch repository.DiscountRepository,
	subSearch repository.SubscriptionRepository,
) billingv1connect.BillingServiceHandler {
	return &BillingServer{
		Catalog:        catalog,
		Subscription:   subscription,
		Usage:          usage,
		Metering:       metering,
		Pricing:        pricing,
		Discount:       discount,
		Credit:         credit,
		Invoice:        invoice,
		Workflow:       workflow,
		Ledger:         ledger,
		CatalogSearch:  catalogSearch,
		UsageSearch:    usageSearch,
		InvoiceSearch:  invoiceSearch,
		DiscountSearch: discountSearch,
		SubSearch:      subSearch,
	}
}

// -- Catalog RPCs --

func (s *BillingServer) CreateCatalogVersion(
	ctx context.Context,
	req *billingv1.CreateCatalogVersionRequest,
) (*billingv1.CreateCatalogVersionResponse, error) {
	cv := &models.CatalogVersion{
		CatalogID: req.GetCatalogId(),
		Name:      req.GetName(),
		Currency:  req.GetCurrency(),
		Data:      structToJSONMap(req.GetData()),
	}
	if req.GetId() != "" {
		cv.ID = req.GetId()
	}

	created, err := s.Catalog.CreateCatalogVersion(ctx, cv)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.CreateCatalogVersionResponse{
		Data: catalogVersionToProto(created),
	}, nil
}

func (s *BillingServer) GetCatalogVersion(
	ctx context.Context,
	req *billingv1.GetCatalogVersionRequest,
) (*billingv1.GetCatalogVersionResponse, error) {
	cv, err := s.Catalog.GetCatalogVersion(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.GetCatalogVersionResponse{
		Data: catalogVersionToProto(cv),
	}, nil
}

func (s *BillingServer) PublishCatalogVersion(
	ctx context.Context,
	req *billingv1.PublishCatalogVersionRequest,
) (*billingv1.PublishCatalogVersionResponse, error) {
	if req.GetEffectiveAt() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, "effective_at is required")
	}
	effectiveAt := req.GetEffectiveAt().AsTime()
	cv, err := s.Catalog.PublishCatalogVersion(ctx, req.GetId(), effectiveAt)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.PublishCatalogVersionResponse{
		Data: catalogVersionToProto(cv),
	}, nil
}

func (s *BillingServer) SearchCatalogVersions(
	ctx context.Context,
	req *commonv1.SearchRequest,
	stream billingv1connect.BillingServiceSearchCatalogVersionsServerStream,
) error {
	results, err := s.CatalogSearch.SearchAsESQ(ctx, req.GetQuery())
	if err != nil {
		return toConnectError(err)
	}

	for {
		res, ok := results.ReadResult(ctx)
		if !ok {
			break
		}
		if res.IsError() {
			return toConnectError(res.Error())
		}

		batch := res.Item()
		protoItems := make([]*billingv1.CatalogVersion, len(batch))
		for i, cv := range batch {
			protoItems[i] = catalogVersionToProto(cv)
		}
		if sendErr := stream.Send(&billingv1.SearchCatalogVersionsResponse{
			Data: protoItems,
		}); sendErr != nil {
			return sendErr
		}
	}
	return nil
}

func (s *BillingServer) CreatePlan(
	ctx context.Context,
	req *billingv1.CreatePlanRequest,
) (*billingv1.CreatePlanResponse, error) {
	plan := &models.Plan{
		CatalogVersionID: req.GetCatalogVersionId(),
		ExternalID:       req.GetExternalId(),
		Name:             req.GetName(),
		Description:      req.GetDescription(),
		Data:             structToJSONMap(req.GetData()),
	}
	if req.GetId() != "" {
		plan.ID = req.GetId()
	}

	created, err := s.Catalog.CreatePlan(ctx, plan)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.CreatePlanResponse{
		Data: planToProto(created),
	}, nil
}

func (s *BillingServer) CreateComponent(
	ctx context.Context,
	req *billingv1.CreateComponentRequest,
) (*billingv1.CreateComponentResponse, error) {
	comp := &models.Component{
		PlanID:          req.GetPlanId(),
		ExternalID:      req.GetExternalId(),
		Name:            req.GetName(),
		MetricKey:       req.GetMetricKey(),
		PricingModel:    pricingModelFromProto(req.GetPricingModel()),
		AggregationType: aggregationTypeFromProto(req.GetAggregationType()),
		UnitName:        req.GetUnitName(),
		FreeQuantity:    moneyToDecPtr(req.GetFreeQuantity()),
		MinimumCharge:   moneyToDecPtr(req.GetMinimumCharge()),
		Data:            structToJSONMap(req.GetData()),
	}
	if req.GetId() != "" {
		comp.ID = req.GetId()
	}

	created, err := s.Catalog.CreateComponent(ctx, comp)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.CreateComponentResponse{
		Data: componentToProto(created),
	}, nil
}

func (s *BillingServer) CreateTier(
	ctx context.Context,
	req *billingv1.CreateTierRequest,
) (*billingv1.CreateTierResponse, error) {
	tier := &models.Tier{
		ComponentID: req.GetComponentId(),
		SortOrder:   int(req.GetSortOrder()),
		LowerBound:  moneyToDecPtr(req.GetLowerBound()),
		UpperBound:  moneyToDecPtr(req.GetUpperBound()),
		UnitPrice:   moneyToDecPtr(req.GetUnitPrice()),
		FlatFee:     moneyToDecPtr(req.GetFlatFee()),
	}
	if req.GetId() != "" {
		tier.ID = req.GetId()
	}

	created, err := s.Catalog.CreateTier(ctx, tier)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.CreateTierResponse{
		Data: tierToProto(created),
	}, nil
}

// -- Subscription RPCs --

func (s *BillingServer) CreateSubscription(
	ctx context.Context,
	req *billingv1.CreateSubscriptionRequest,
) (*billingv1.CreateSubscriptionResponse, error) {
	sub := &models.Subscription{
		ProfileID:        req.GetProfileId(),
		CatalogVersionID: req.GetCatalogVersionId(),
		PlanID:           req.GetPlanId(),
		State:            models.SubscriptionStateActive,
		Currency:         req.GetCurrency(),
		Data:             structToJSONMap(req.GetData()),
	}
	if req.GetStartAt() != nil {
		sub.StartAt = req.GetStartAt().AsTime()
	}
	if req.GetBillingAnchor() != nil {
		sub.BillingAnchor = req.GetBillingAnchor().AsTime()
	}
	if req.GetId() != "" {
		sub.ID = req.GetId()
	}

	created, err := s.Subscription.CreateSubscription(ctx, sub)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.CreateSubscriptionResponse{
		Data: subscriptionToProto(created),
	}, nil
}

func (s *BillingServer) GetSubscription(
	ctx context.Context,
	req *billingv1.GetSubscriptionRequest,
) (*billingv1.GetSubscriptionResponse, error) {
	sub, err := s.Subscription.GetSubscription(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.GetSubscriptionResponse{
		Data: subscriptionToProto(sub),
	}, nil
}

func (s *BillingServer) CancelSubscription(
	ctx context.Context,
	req *billingv1.CancelSubscriptionRequest,
) (*billingv1.CancelSubscriptionResponse, error) {
	sub, err := s.Subscription.CancelSubscription(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.CancelSubscriptionResponse{
		Data: subscriptionToProto(sub),
	}, nil
}

func (s *BillingServer) ListSubscriptions(
	ctx context.Context,
	req *billingv1.ListSubscriptionsRequest,
) (*billingv1.ListSubscriptionsResponse, error) {
	subs, err := s.Subscription.ListActiveByProfile(ctx, req.GetProfileId())
	if err != nil {
		return nil, toConnectError(err)
	}

	protoSubs := make([]*billingv1.Subscription, len(subs))
	for i, sub := range subs {
		protoSubs[i] = subscriptionToProto(sub)
	}

	return &billingv1.ListSubscriptionsResponse{
		Data: protoSubs,
	}, nil
}

// -- Usage RPCs --

func (s *BillingServer) IngestUsageEvent(
	ctx context.Context,
	req *billingv1.IngestUsageEventRequest,
) (*billingv1.IngestUsageEventResponse, error) {
	events := make([]*models.UsageEvent, len(req.GetData()))
	for i, e := range req.GetData() {
		if e.GetTimestamp() == nil {
			return nil, connect.Errorf(connect.CodeInvalidArgument,
				"event at index %d: timestamp is required", i)
		}
		qty, _ := decimalx.NewFromString(strconv.FormatFloat(e.GetQuantity(), 'f', -1, 64))
		events[i] = &models.UsageEvent{
			SubscriptionID: e.GetSubscriptionId(),
			MetricKey:      e.GetMetricKey(),
			Quantity:       qty.Ptr(),
			TrueCreatedAt:  e.GetTimestamp().AsTime(),
			Properties:     structToJSONMap(e.GetProperties()),
		}
		if e.GetId() != "" {
			events[i].EventID = e.GetId()
		}
	}

	created, err := s.Usage.IngestBatch(ctx, events)
	if err != nil {
		return nil, toConnectError(err)
	}

	ids := make([]string, len(created))
	for i, e := range created {
		ids[i] = e.GetID()
	}

	return &billingv1.IngestUsageEventResponse{
		Data: ids,
	}, nil
}

func (s *BillingServer) SearchUsageEvents(
	ctx context.Context,
	req *commonv1.SearchRequest,
	stream billingv1connect.BillingServiceSearchUsageEventsServerStream,
) error {
	results, err := s.UsageSearch.SearchAsESQ(ctx, req.GetQuery())
	if err != nil {
		return toConnectError(err)
	}

	for {
		res, ok := results.ReadResult(ctx)
		if !ok {
			break
		}
		if res.IsError() {
			return toConnectError(res.Error())
		}

		batch := res.Item()
		protoItems := make([]*billingv1.UsageEvent, len(batch))
		for i, e := range batch {
			protoItems[i] = usageEventToProto(e)
		}
		if sendErr := stream.Send(&billingv1.SearchUsageEventsResponse{
			Data: protoItems,
		}); sendErr != nil {
			return sendErr
		}
	}
	return nil
}

// -- Billing Run RPCs --

func (s *BillingServer) RunBilling(
	ctx context.Context,
	req *billingv1.RunBillingRequest,
) (*billingv1.RunBillingResponse, error) {
	if req.GetPeriodStart() == nil || req.GetPeriodEnd() == nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			"period_start and period_end are required",
		)
	}
	run, err := s.Workflow.RunBilling(ctx,
		req.GetSubscriptionId(),
		req.GetPeriodStart().AsTime(),
		req.GetPeriodEnd().AsTime())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.RunBillingResponse{
		Data: billingRunToProto(run),
	}, nil
}

func (s *BillingServer) GetBillingRun(
	ctx context.Context,
	req *billingv1.GetBillingRunRequest,
) (*billingv1.GetBillingRunResponse, error) {
	run, err := s.Workflow.GetBillingRun(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.GetBillingRunResponse{
		Data: billingRunToProto(run),
	}, nil
}

// -- Invoice RPCs --

func (s *BillingServer) GetInvoice(
	ctx context.Context,
	req *billingv1.GetInvoiceRequest,
) (*billingv1.GetInvoiceResponse, error) {
	inv, err := s.Invoice.GetInvoice(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.GetInvoiceResponse{
		Data: invoiceToProto(inv),
	}, nil
}

func (s *BillingServer) IssueInvoice(
	ctx context.Context,
	req *billingv1.IssueInvoiceRequest,
) (*billingv1.IssueInvoiceResponse, error) {
	inv, err := s.Invoice.IssueInvoice(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.IssueInvoiceResponse{
		Data: invoiceToProto(inv),
	}, nil
}

func (s *BillingServer) VoidInvoice(
	ctx context.Context,
	req *billingv1.VoidInvoiceRequest,
) (*billingv1.VoidInvoiceResponse, error) {
	inv, err := s.Invoice.VoidInvoice(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.VoidInvoiceResponse{
		Data: invoiceToProto(inv),
	}, nil
}

func (s *BillingServer) RecordPayment(
	ctx context.Context,
	req *billingv1.RecordPaymentRequest,
) (*billingv1.RecordPaymentResponse, error) {
	inv, err := s.Invoice.RecordPayment(ctx, req.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.RecordPaymentResponse{
		Data: invoiceToProto(inv),
	}, nil
}

func (s *BillingServer) SearchInvoices(
	ctx context.Context,
	req *commonv1.SearchRequest,
	stream billingv1connect.BillingServiceSearchInvoicesServerStream,
) error {
	results, err := s.InvoiceSearch.SearchAsESQ(ctx, req.GetQuery())
	if err != nil {
		return toConnectError(err)
	}

	for {
		res, ok := results.ReadResult(ctx)
		if !ok {
			break
		}
		if res.IsError() {
			return toConnectError(res.Error())
		}

		batch := res.Item()
		protoItems := make([]*billingv1.Invoice, len(batch))
		for i, inv := range batch {
			protoItems[i] = invoiceToProto(inv)
		}
		if sendErr := stream.Send(&billingv1.SearchInvoicesResponse{
			Data: protoItems,
		}); sendErr != nil {
			return sendErr
		}
	}
	return nil
}

// -- Credit RPCs --

func (s *BillingServer) GrantCredit(
	ctx context.Context,
	req *billingv1.GrantCreditRequest,
) (*billingv1.GrantCreditResponse, error) {
	grant := &models.CreditGrant{
		ProfileID:       req.GetProfileId(),
		Name:            req.GetName(),
		OriginalAmount:  moneyToDecPtr(req.GetAmount()),
		RemainingAmount: moneyToDecPtr(req.GetAmount()),
		Currency:        req.GetCurrency(),
		Priority:        int(req.GetPriority()),
		Data:            structToJSONMap(req.GetData()),
	}
	if req.GetId() != "" {
		grant.ID = req.GetId()
	}
	if req.GetExpiresAt() != nil {
		expiresAt := req.GetExpiresAt().AsTime()
		grant.ExpiresAt = &expiresAt
	}

	created, err := s.Credit.GrantCredit(ctx, grant)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.GrantCreditResponse{
		Data: creditGrantToProto(created),
	}, nil
}

func (s *BillingServer) GetCreditBalance(
	ctx context.Context,
	req *billingv1.GetCreditBalanceRequest,
) (*billingv1.GetCreditBalanceResponse, error) {
	balance, err := s.Credit.GetBalance(ctx, req.GetProfileId(), req.GetCurrency())
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.GetCreditBalanceResponse{
		Balance: decimalToMoney(balance, req.GetCurrency()),
	}, nil
}

// -- Discount RPCs --

func (s *BillingServer) CreateDiscount(
	ctx context.Context,
	req *billingv1.CreateDiscountRequest,
) (*billingv1.CreateDiscountResponse, error) {
	disc := &models.Discount{
		Name:         req.GetName(),
		DiscountType: discountTypeFromProto(req.GetDiscountType()),
		Value: func() *decimalx.Decimal {
			v, _ := decimalx.NewFromString(strconv.FormatFloat(req.GetValue(), 'f', -1, 64))
			return v.Ptr()
		}(),
		Currency:        req.GetCurrency(),
		ApplicableTo:    structToJSONMap(req.GetApplicableTo()),
		MaxApplications: int(req.GetMaxApplications()),
		Data:            structToJSONMap(req.GetData()),
	}
	if req.GetId() != "" {
		disc.ID = req.GetId()
	}
	if req.GetStartAt() != nil {
		startAt := req.GetStartAt().AsTime()
		disc.StartAt = &startAt
	}
	if req.GetEndAt() != nil {
		endAt := req.GetEndAt().AsTime()
		disc.EndAt = &endAt
	}

	created, err := s.Discount.CreateDiscount(ctx, disc)
	if err != nil {
		return nil, toConnectError(err)
	}

	return &billingv1.CreateDiscountResponse{
		Data: discountToProto(created),
	}, nil
}

func (s *BillingServer) SearchDiscounts(
	ctx context.Context,
	req *commonv1.SearchRequest,
	stream billingv1connect.BillingServiceSearchDiscountsServerStream,
) error {
	results, err := s.DiscountSearch.SearchAsESQ(ctx, req.GetQuery())
	if err != nil {
		return toConnectError(err)
	}

	for {
		res, ok := results.ReadResult(ctx)
		if !ok {
			break
		}
		if res.IsError() {
			return toConnectError(res.Error())
		}

		batch := res.Item()
		protoItems := make([]*billingv1.Discount, len(batch))
		for i, d := range batch {
			protoItems[i] = discountToProto(d)
		}
		if sendErr := stream.Send(&billingv1.SearchDiscountsResponse{
			Data: protoItems,
		}); sendErr != nil {
			return sendErr
		}
	}
	return nil
}
