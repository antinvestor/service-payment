package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"buf.build/gen/go/antinvestor/billing/connectrpc/go/v1/billingv1connect"
	billingv1 "buf.build/gen/go/antinvestor/billing/protocolbuffers/go/v1"
	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/apps/billing/service/repository"
	"github.com/antinvestor/service-payments/internal/apperrors"
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
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
		return connect.NewError(connect.CodeInternal, err)
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
		return connect.NewError(connect.CodeNotFound, appErr)

	case apperrors.ErrorCodeCatalogVersionRetired,
		apperrors.ErrorCodeCatalogNotPublished,
		apperrors.ErrorCodeSubscriptionNotActive,
		apperrors.ErrorCodeInvoiceAlreadyIssued,
		apperrors.ErrorCodeInvoiceNotIssuable,
		apperrors.ErrorCodeInvoiceNotVoidable,
		apperrors.ErrorCodeInvoiceNotPayable,
		apperrors.ErrorCodeCreditExpired:
		return connect.NewError(connect.CodeFailedPrecondition, appErr)

	case apperrors.ErrorCodeInvalidPricingModel,
		apperrors.ErrorCodeUsageEventInvalid,
		apperrors.ErrorCodeRatingFailed,
		apperrors.ErrorCodeDiscountInvalid:
		return connect.NewError(connect.CodeInvalidArgument, appErr)

	case apperrors.ErrorCodeUsageEventDuplicate,
		apperrors.ErrorCodeBillingRunAlreadyExists:
		return connect.NewError(connect.CodeAlreadyExists, appErr)

	case apperrors.ErrorCodeCreditInsufficientFunds:
		return connect.NewError(connect.CodeResourceExhausted, appErr)

	case apperrors.ErrorCodeBillingRunFailed:
		return connect.NewError(connect.CodeInternal, appErr)

	default:
		return connect.NewError(connect.CodeInternal, appErr)
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
	req *connect.Request[billingv1.CreateCatalogVersionRequest],
) (*connect.Response[billingv1.CreateCatalogVersionResponse], error) {
	cv := &models.CatalogVersion{
		CatalogID: req.Msg.GetCatalogId(),
		Name:      req.Msg.GetName(),
		Currency:  req.Msg.GetCurrency(),
		Data:      structToJSONMap(req.Msg.GetData()),
	}
	if req.Msg.GetId() != "" {
		cv.ID = req.Msg.GetId()
	}

	created, err := s.Catalog.CreateCatalogVersion(ctx, cv)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.CreateCatalogVersionResponse{
		Data: catalogVersionToProto(created),
	}), nil
}

func (s *BillingServer) GetCatalogVersion(
	ctx context.Context,
	req *connect.Request[billingv1.GetCatalogVersionRequest],
) (*connect.Response[billingv1.GetCatalogVersionResponse], error) {
	cv, err := s.Catalog.GetCatalogVersion(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.GetCatalogVersionResponse{
		Data: catalogVersionToProto(cv),
	}), nil
}

func (s *BillingServer) PublishCatalogVersion(
	ctx context.Context,
	req *connect.Request[billingv1.PublishCatalogVersionRequest],
) (*connect.Response[billingv1.PublishCatalogVersionResponse], error) {
	if req.Msg.GetEffectiveAt() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("effective_at is required"))
	}
	effectiveAt := req.Msg.GetEffectiveAt().AsTime()
	cv, err := s.Catalog.PublishCatalogVersion(ctx, req.Msg.GetId(), effectiveAt)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.PublishCatalogVersionResponse{
		Data: catalogVersionToProto(cv),
	}), nil
}

func (s *BillingServer) SearchCatalogVersions(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[billingv1.SearchCatalogVersionsResponse],
) error {
	results, err := s.CatalogSearch.SearchAsESQ(ctx, req.Msg.GetQuery())
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
	req *connect.Request[billingv1.CreatePlanRequest],
) (*connect.Response[billingv1.CreatePlanResponse], error) {
	plan := &models.Plan{
		CatalogVersionID: req.Msg.GetCatalogVersionId(),
		ExternalID:       req.Msg.GetExternalId(),
		Name:             req.Msg.GetName(),
		Description:      req.Msg.GetDescription(),
		Data:             structToJSONMap(req.Msg.GetData()),
	}
	if req.Msg.GetId() != "" {
		plan.ID = req.Msg.GetId()
	}

	created, err := s.Catalog.CreatePlan(ctx, plan)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.CreatePlanResponse{
		Data: planToProto(created),
	}), nil
}

func (s *BillingServer) CreateComponent(
	ctx context.Context,
	req *connect.Request[billingv1.CreateComponentRequest],
) (*connect.Response[billingv1.CreateComponentResponse], error) {
	comp := &models.Component{
		PlanID:          req.Msg.GetPlanId(),
		ExternalID:      req.Msg.GetExternalId(),
		Name:            req.Msg.GetName(),
		MetricKey:       req.Msg.GetMetricKey(),
		PricingModel:    pricingModelFromProto(req.Msg.GetPricingModel()),
		AggregationType: aggregationTypeFromProto(req.Msg.GetAggregationType()),
		UnitName:        req.Msg.GetUnitName(),
		FreeQuantity:    moneyToDecPtr(req.Msg.GetFreeQuantity()),
		MinimumCharge:   moneyToDecPtr(req.Msg.GetMinimumCharge()),
		Data:            structToJSONMap(req.Msg.GetData()),
	}
	if req.Msg.GetId() != "" {
		comp.ID = req.Msg.GetId()
	}

	created, err := s.Catalog.CreateComponent(ctx, comp)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.CreateComponentResponse{
		Data: componentToProto(created),
	}), nil
}

func (s *BillingServer) CreateTier(
	ctx context.Context,
	req *connect.Request[billingv1.CreateTierRequest],
) (*connect.Response[billingv1.CreateTierResponse], error) {
	tier := &models.Tier{
		ComponentID: req.Msg.GetComponentId(),
		SortOrder:   int(req.Msg.GetSortOrder()),
		LowerBound:  moneyToDecPtr(req.Msg.GetLowerBound()),
		UpperBound:  moneyToDecPtr(req.Msg.GetUpperBound()),
		UnitPrice:   moneyToDecPtr(req.Msg.GetUnitPrice()),
		FlatFee:     moneyToDecPtr(req.Msg.GetFlatFee()),
	}
	if req.Msg.GetId() != "" {
		tier.ID = req.Msg.GetId()
	}

	created, err := s.Catalog.CreateTier(ctx, tier)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.CreateTierResponse{
		Data: tierToProto(created),
	}), nil
}

// -- Subscription RPCs --

func (s *BillingServer) CreateSubscription(
	ctx context.Context,
	req *connect.Request[billingv1.CreateSubscriptionRequest],
) (*connect.Response[billingv1.CreateSubscriptionResponse], error) {
	sub := &models.Subscription{
		ProfileID:        req.Msg.GetProfileId(),
		CatalogVersionID: req.Msg.GetCatalogVersionId(),
		PlanID:           req.Msg.GetPlanId(),
		State:            models.SubscriptionStateActive,
		Currency:         req.Msg.GetCurrency(),
		Data:             structToJSONMap(req.Msg.GetData()),
	}
	if req.Msg.GetStartAt() != nil {
		sub.StartAt = req.Msg.GetStartAt().AsTime()
	}
	if req.Msg.GetBillingAnchor() != nil {
		sub.BillingAnchor = req.Msg.GetBillingAnchor().AsTime()
	}
	if req.Msg.GetId() != "" {
		sub.ID = req.Msg.GetId()
	}

	created, err := s.Subscription.CreateSubscription(ctx, sub)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.CreateSubscriptionResponse{
		Data: subscriptionToProto(created),
	}), nil
}

func (s *BillingServer) GetSubscription(
	ctx context.Context,
	req *connect.Request[billingv1.GetSubscriptionRequest],
) (*connect.Response[billingv1.GetSubscriptionResponse], error) {
	sub, err := s.Subscription.GetSubscription(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.GetSubscriptionResponse{
		Data: subscriptionToProto(sub),
	}), nil
}

func (s *BillingServer) CancelSubscription(
	ctx context.Context,
	req *connect.Request[billingv1.CancelSubscriptionRequest],
) (*connect.Response[billingv1.CancelSubscriptionResponse], error) {
	sub, err := s.Subscription.CancelSubscription(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.CancelSubscriptionResponse{
		Data: subscriptionToProto(sub),
	}), nil
}

func (s *BillingServer) ListSubscriptions(
	ctx context.Context,
	req *connect.Request[billingv1.ListSubscriptionsRequest],
) (*connect.Response[billingv1.ListSubscriptionsResponse], error) {
	subs, err := s.Subscription.ListActiveByProfile(ctx, req.Msg.GetProfileId())
	if err != nil {
		return nil, toConnectError(err)
	}

	protoSubs := make([]*billingv1.Subscription, len(subs))
	for i, sub := range subs {
		protoSubs[i] = subscriptionToProto(sub)
	}

	return connect.NewResponse(&billingv1.ListSubscriptionsResponse{
		Data: protoSubs,
	}), nil
}

// -- Usage RPCs --

func (s *BillingServer) IngestUsageEvent(
	ctx context.Context,
	req *connect.Request[billingv1.IngestUsageEventRequest],
) (*connect.Response[billingv1.IngestUsageEventResponse], error) {
	events := make([]*models.UsageEvent, len(req.Msg.GetData()))
	for i, e := range req.Msg.GetData() {
		if e.GetTimestamp() == nil {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				fmt.Errorf("event at index %d: timestamp is required", i))
		}
		qty, _ := decimalx.NewFromString(fmt.Sprintf("%g", e.GetQuantity()))
		events[i] = &models.UsageEvent{
			SubscriptionID: e.GetSubscriptionId(),
			MetricKey:      e.GetMetricKey(),
			Quantity:       qty.Ptr(),
			Timestamp:      e.GetTimestamp().AsTime(),
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

	return connect.NewResponse(&billingv1.IngestUsageEventResponse{
		Data: ids,
	}), nil
}

func (s *BillingServer) SearchUsageEvents(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[billingv1.SearchUsageEventsResponse],
) error {
	results, err := s.UsageSearch.SearchAsESQ(ctx, req.Msg.GetQuery())
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
	req *connect.Request[billingv1.RunBillingRequest],
) (*connect.Response[billingv1.RunBillingResponse], error) {
	if req.Msg.GetPeriodStart() == nil || req.Msg.GetPeriodEnd() == nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("period_start and period_end are required"),
		)
	}
	run, err := s.Workflow.RunBilling(ctx,
		req.Msg.GetSubscriptionId(),
		req.Msg.GetPeriodStart().AsTime(),
		req.Msg.GetPeriodEnd().AsTime())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.RunBillingResponse{
		Data: billingRunToProto(run),
	}), nil
}

func (s *BillingServer) GetBillingRun(
	ctx context.Context,
	req *connect.Request[billingv1.GetBillingRunRequest],
) (*connect.Response[billingv1.GetBillingRunResponse], error) {
	run, err := s.Workflow.GetBillingRun(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.GetBillingRunResponse{
		Data: billingRunToProto(run),
	}), nil
}

// -- Invoice RPCs --

func (s *BillingServer) GetInvoice(
	ctx context.Context,
	req *connect.Request[billingv1.GetInvoiceRequest],
) (*connect.Response[billingv1.GetInvoiceResponse], error) {
	inv, err := s.Invoice.GetInvoice(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.GetInvoiceResponse{
		Data: invoiceToProto(inv),
	}), nil
}

func (s *BillingServer) IssueInvoice(
	ctx context.Context,
	req *connect.Request[billingv1.IssueInvoiceRequest],
) (*connect.Response[billingv1.IssueInvoiceResponse], error) {
	inv, err := s.Invoice.IssueInvoice(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.IssueInvoiceResponse{
		Data: invoiceToProto(inv),
	}), nil
}

func (s *BillingServer) VoidInvoice(
	ctx context.Context,
	req *connect.Request[billingv1.VoidInvoiceRequest],
) (*connect.Response[billingv1.VoidInvoiceResponse], error) {
	inv, err := s.Invoice.VoidInvoice(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.VoidInvoiceResponse{
		Data: invoiceToProto(inv),
	}), nil
}

func (s *BillingServer) RecordPayment(
	ctx context.Context,
	req *connect.Request[billingv1.RecordPaymentRequest],
) (*connect.Response[billingv1.RecordPaymentResponse], error) {
	inv, err := s.Invoice.RecordPayment(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.RecordPaymentResponse{
		Data: invoiceToProto(inv),
	}), nil
}

func (s *BillingServer) SearchInvoices(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[billingv1.SearchInvoicesResponse],
) error {
	results, err := s.InvoiceSearch.SearchAsESQ(ctx, req.Msg.GetQuery())
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
	req *connect.Request[billingv1.GrantCreditRequest],
) (*connect.Response[billingv1.GrantCreditResponse], error) {
	grant := &models.CreditGrant{
		ProfileID:       req.Msg.GetProfileId(),
		Name:            req.Msg.GetName(),
		OriginalAmount:  moneyToDecPtr(req.Msg.GetAmount()),
		RemainingAmount: moneyToDecPtr(req.Msg.GetAmount()),
		Currency:        req.Msg.GetCurrency(),
		Priority:        int(req.Msg.GetPriority()),
		Data:            structToJSONMap(req.Msg.GetData()),
	}
	if req.Msg.GetId() != "" {
		grant.ID = req.Msg.GetId()
	}
	if req.Msg.GetExpiresAt() != nil {
		expiresAt := req.Msg.GetExpiresAt().AsTime()
		grant.ExpiresAt = &expiresAt
	}

	created, err := s.Credit.GrantCredit(ctx, grant)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.GrantCreditResponse{
		Data: creditGrantToProto(created),
	}), nil
}

func (s *BillingServer) GetCreditBalance(
	ctx context.Context,
	req *connect.Request[billingv1.GetCreditBalanceRequest],
) (*connect.Response[billingv1.GetCreditBalanceResponse], error) {
	balance, err := s.Credit.GetBalance(ctx, req.Msg.GetProfileId(), req.Msg.GetCurrency())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.GetCreditBalanceResponse{
		Balance: decimalToMoney(balance, req.Msg.GetCurrency()),
	}), nil
}

// -- Discount RPCs --

func (s *BillingServer) CreateDiscount(
	ctx context.Context,
	req *connect.Request[billingv1.CreateDiscountRequest],
) (*connect.Response[billingv1.CreateDiscountResponse], error) {
	disc := &models.Discount{
		Name:         req.Msg.GetName(),
		DiscountType: discountTypeFromProto(req.Msg.GetDiscountType()),
		Value: func() *decimalx.Decimal {
			v, _ := decimalx.NewFromString(fmt.Sprintf("%g", req.Msg.GetValue()))
			return v.Ptr()
		}(),
		Currency:        req.Msg.GetCurrency(),
		ApplicableTo:    structToJSONMap(req.Msg.GetApplicableTo()),
		MaxApplications: int(req.Msg.GetMaxApplications()),
		Data:            structToJSONMap(req.Msg.GetData()),
	}
	if req.Msg.GetId() != "" {
		disc.ID = req.Msg.GetId()
	}
	if req.Msg.GetStartAt() != nil {
		startAt := req.Msg.GetStartAt().AsTime()
		disc.StartAt = &startAt
	}
	if req.Msg.GetEndAt() != nil {
		endAt := req.Msg.GetEndAt().AsTime()
		disc.EndAt = &endAt
	}

	created, err := s.Discount.CreateDiscount(ctx, disc)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&billingv1.CreateDiscountResponse{
		Data: discountToProto(created),
	}), nil
}

func (s *BillingServer) SearchDiscounts(
	ctx context.Context,
	req *connect.Request[commonv1.SearchRequest],
	stream *connect.ServerStream[billingv1.SearchDiscountsResponse],
) error {
	results, err := s.DiscountSearch.SearchAsESQ(ctx, req.Msg.GetQuery())
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
