package handlers

import (
	"time"

	billingv1 "buf.build/gen/go/antinvestor/billing/protocolbuffers/go/billing/v1"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/internal/utility"
	"github.com/pitabwire/frame/data"
	"github.com/shopspring/decimal"
	money "google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// -- Struct/JSONMap conversions --

func structToJSONMap(s *structpb.Struct) data.JSONMap {
	if s == nil {
		return nil
	}
	result := make(data.JSONMap)
	for k, v := range s.GetFields() {
		result[k] = v.AsInterface()
	}
	return result
}

func jsonMapToStruct(m data.JSONMap) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}

// -- Money/Decimal conversions --

func moneyToNullDecimal(m *money.Money) decimal.NullDecimal {
	if m == nil {
		return decimal.NullDecimal{}
	}
	return decimal.NewNullDecimal(utility.FromMoney(m))
}

func nullDecimalToMoney(nd decimal.NullDecimal, currency string) *money.Money {
	if !nd.Valid {
		return nil
	}
	m := utility.ToMoney(currency, nd.Decimal)
	return &m
}

func decimalToMoney(d decimal.Decimal, currency string) *money.Money {
	m := utility.ToMoney(currency, d)
	return &m
}

// -- Timestamp helpers --

func timeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func timePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// -- Enum conversions --

func pricingModelFromProto(pm billingv1.PricingModel) string {
	switch pm {
	case billingv1.PricingModel_FLAT:
		return models.PricingModelFlat
	case billingv1.PricingModel_PER_UNIT:
		return models.PricingModelPerUnit
	case billingv1.PricingModel_TIERED:
		return models.PricingModelTiered
	case billingv1.PricingModel_VOLUME:
		return models.PricingModelVolume
	case billingv1.PricingModel_STAIRSTEP:
		return models.PricingModelStairstep
	default:
		return models.PricingModelFlat
	}
}

func pricingModelToProto(pm string) billingv1.PricingModel {
	switch pm {
	case models.PricingModelFlat:
		return billingv1.PricingModel_FLAT
	case models.PricingModelPerUnit:
		return billingv1.PricingModel_PER_UNIT
	case models.PricingModelTiered:
		return billingv1.PricingModel_TIERED
	case models.PricingModelVolume:
		return billingv1.PricingModel_VOLUME
	case models.PricingModelStairstep:
		return billingv1.PricingModel_STAIRSTEP
	default:
		return billingv1.PricingModel_FLAT
	}
}

func aggregationTypeFromProto(at billingv1.AggregationType) string {
	switch at {
	case billingv1.AggregationType_SUM:
		return models.AggregationTypeSum
	case billingv1.AggregationType_COUNT:
		return models.AggregationTypeCount
	case billingv1.AggregationType_MAX:
		return models.AggregationTypeMax
	case billingv1.AggregationType_MIN:
		return models.AggregationTypeMin
	case billingv1.AggregationType_AVG:
		return models.AggregationTypeAvg
	case billingv1.AggregationType_LAST:
		return models.AggregationTypeLast
	default:
		return models.AggregationTypeSum
	}
}

func subscriptionStateToProto(state string) billingv1.SubscriptionState {
	switch state {
	case models.SubscriptionStateActive:
		return billingv1.SubscriptionState_SUBSCRIPTION_ACTIVE
	case models.SubscriptionStateCancelled:
		return billingv1.SubscriptionState_SUBSCRIPTION_CANCELLED
	case models.SubscriptionStateExpired:
		return billingv1.SubscriptionState_SUBSCRIPTION_EXPIRED
	case models.SubscriptionStatePending:
		return billingv1.SubscriptionState_SUBSCRIPTION_PENDING
	default:
		return billingv1.SubscriptionState_SUBSCRIPTION_PENDING
	}
}

func invoiceStateToProto(state string) billingv1.InvoiceState {
	switch state {
	case models.InvoiceStateDraft:
		return billingv1.InvoiceState_INVOICE_DRAFT
	case models.InvoiceStateIssued:
		return billingv1.InvoiceState_INVOICE_ISSUED
	case models.InvoiceStatePaid:
		return billingv1.InvoiceState_INVOICE_PAID
	case models.InvoiceStateVoided:
		return billingv1.InvoiceState_INVOICE_VOIDED
	case models.InvoiceStateOverdue:
		return billingv1.InvoiceState_INVOICE_OVERDUE
	default:
		return billingv1.InvoiceState_INVOICE_DRAFT
	}
}

func billingRunStateToProto(state string) billingv1.BillingRunState {
	switch state {
	case models.BillingRunStatePending:
		return billingv1.BillingRunState_BILLING_RUN_PENDING
	case models.BillingRunStateMetering:
		return billingv1.BillingRunState_BILLING_RUN_METERING
	case models.BillingRunStateRating:
		return billingv1.BillingRunState_BILLING_RUN_RATING
	case models.BillingRunStateDiscounting:
		return billingv1.BillingRunState_BILLING_RUN_DISCOUNTING
	case models.BillingRunStateCrediting:
		return billingv1.BillingRunState_BILLING_RUN_CREDITING
	case models.BillingRunStateInvoicing:
		return billingv1.BillingRunState_BILLING_RUN_INVOICING
	case models.BillingRunStatePosting:
		return billingv1.BillingRunState_BILLING_RUN_POSTING
	case models.BillingRunStateCompleted:
		return billingv1.BillingRunState_BILLING_RUN_COMPLETED
	case models.BillingRunStateFailed:
		return billingv1.BillingRunState_BILLING_RUN_FAILED
	default:
		return billingv1.BillingRunState_BILLING_RUN_PENDING
	}
}

func invoiceLineTypeToProto(lt string) billingv1.InvoiceLineType {
	switch lt {
	case models.InvoiceLineTypeUsage:
		return billingv1.InvoiceLineType_LINE_USAGE
	case models.InvoiceLineTypeFlat:
		return billingv1.InvoiceLineType_LINE_FLAT
	case models.InvoiceLineTypeDiscount:
		return billingv1.InvoiceLineType_LINE_DISCOUNT
	case models.InvoiceLineTypeCredit:
		return billingv1.InvoiceLineType_LINE_CREDIT
	default:
		return billingv1.InvoiceLineType_LINE_USAGE
	}
}

// -- Model to Proto conversions --

func catalogVersionToProto(cv *models.CatalogVersion) *billingv1.CatalogVersion {
	if cv == nil {
		return nil
	}

	proto := &billingv1.CatalogVersion{
		Id:          cv.GetID(),
		CatalogId:   cv.CatalogID,
		Version:     int32(cv.Version), //nolint:gosec // G115: version is always a small positive integer
		Name:        cv.Name,
		Currency:    cv.Currency,
		PublishedAt: timePtrToTimestamp(cv.PublishedAt),
		EffectiveAt: timePtrToTimestamp(cv.EffectiveAt),
		RetiredAt:   timePtrToTimestamp(cv.RetiredAt),
		Data:        jsonMapToStruct(cv.Data),
	}

	if cv.Plans != nil {
		proto.Plans = make([]*billingv1.Plan, len(cv.Plans))
		for i, p := range cv.Plans {
			proto.Plans[i] = planToProto(p)
		}
	}

	return proto
}

func planToProto(p *models.Plan) *billingv1.Plan {
	if p == nil {
		return nil
	}

	proto := &billingv1.Plan{
		Id:               p.GetID(),
		CatalogVersionId: p.CatalogVersionID,
		ExternalId:       p.ExternalID,
		Name:             p.Name,
		Description:      p.Description,
		Data:             jsonMapToStruct(p.Data),
	}

	if p.Components != nil {
		proto.Components = make([]*billingv1.Component, len(p.Components))
		for i, c := range p.Components {
			proto.Components[i] = componentToProto(c)
		}
	}

	return proto
}

func componentToProto(c *models.Component) *billingv1.Component {
	if c == nil {
		return nil
	}

	proto := &billingv1.Component{
		Id:            c.GetID(),
		PlanId:        c.PlanID,
		ExternalId:    c.ExternalID,
		Name:          c.Name,
		MetricKey:     c.MetricKey,
		PricingModel:  pricingModelToProto(c.PricingModel),
		UnitName:      c.UnitName,
		FreeQuantity:  nullDecimalToMoney(c.FreeQuantity, ""),
		MinimumCharge: nullDecimalToMoney(c.MinimumCharge, ""),
		Data:          jsonMapToStruct(c.Data),
	}

	if c.Tiers != nil {
		proto.Tiers = make([]*billingv1.Tier, len(c.Tiers))
		for i, t := range c.Tiers {
			proto.Tiers[i] = tierToProto(t)
		}
	}

	return proto
}

func tierToProto(t *models.Tier) *billingv1.Tier {
	if t == nil {
		return nil
	}
	return &billingv1.Tier{
		Id:          t.GetID(),
		ComponentId: t.ComponentID,
		SortOrder:   int32(t.SortOrder), //nolint:gosec // G115: sort_order is always a small positive integer
		LowerBound:  nullDecimalToMoney(t.LowerBound, ""),
		UpperBound:  nullDecimalToMoney(t.UpperBound, ""),
		UnitPrice:   nullDecimalToMoney(t.UnitPrice, ""),
		FlatFee:     nullDecimalToMoney(t.FlatFee, ""),
	}
}

func subscriptionToProto(s *models.Subscription) *billingv1.Subscription {
	if s == nil {
		return nil
	}
	return &billingv1.Subscription{
		Id:               s.GetID(),
		CustomerId:       s.ProfileID,
		CatalogVersionId: s.CatalogVersionID,
		PlanId:           s.PlanID,
		State:            subscriptionStateToProto(s.State),
		StartAt:          timeToTimestamp(s.StartAt),
		EndAt:            timePtrToTimestamp(s.EndAt),
		CancelledAt:      timePtrToTimestamp(s.CancelledAt),
		BillingAnchor:    timeToTimestamp(s.BillingAnchor),
		Currency:         s.Currency,
		Data:             jsonMapToStruct(s.Data),
	}
}

func usageEventToProto(e *models.UsageEvent) *billingv1.UsageEvent {
	if e == nil {
		return nil
	}
	qty, _ := e.Quantity.Decimal.Float64()
	return &billingv1.UsageEvent{
		Id:             e.GetID(),
		EventId:        e.EventID,
		SubscriptionId: e.SubscriptionID,
		CustomerId:     e.ProfileID,
		MetricKey:      e.MetricKey,
		Quantity:       qty,
		Timestamp:      timeToTimestamp(e.Timestamp),
		Properties:     jsonMapToStruct(e.Properties),
	}
}

func invoiceToProto(inv *models.Invoice) *billingv1.Invoice {
	if inv == nil {
		return nil
	}

	proto := &billingv1.Invoice{
		Id:             inv.GetID(),
		BillingRunId:   inv.BillingRunID,
		CustomerId:     inv.ProfileID,
		SubscriptionId: inv.SubscriptionID,
		InvoiceNumber:  inv.InvoiceNumber,
		State:          invoiceStateToProto(inv.State),
		Currency:       inv.Currency,
		SubtotalAmount: nullDecimalToMoney(inv.SubtotalAmount, inv.Currency),
		DiscountAmount: nullDecimalToMoney(inv.DiscountAmount, inv.Currency),
		CreditAmount:   nullDecimalToMoney(inv.CreditAmount, inv.Currency),
		TotalAmount:    nullDecimalToMoney(inv.TotalAmount, inv.Currency),
		PeriodStart:    timeToTimestamp(inv.PeriodStart),
		PeriodEnd:      timeToTimestamp(inv.PeriodEnd),
		IssuedAt:       timePtrToTimestamp(inv.IssuedAt),
		DueAt:          timePtrToTimestamp(inv.DueAt),
		PaidAt:         timePtrToTimestamp(inv.PaidAt),
		LedgerTxnId:    inv.LedgerTxnID,
		Data:           jsonMapToStruct(inv.Data),
	}

	if inv.Lines != nil {
		proto.Lines = make([]*billingv1.InvoiceLine, len(inv.Lines))
		for i, l := range inv.Lines {
			proto.Lines[i] = invoiceLineToProto(l, inv.Currency)
		}
	}

	return proto
}

func invoiceLineToProto(l *models.InvoiceLine, currency string) *billingv1.InvoiceLine {
	if l == nil {
		return nil
	}
	qty, _ := l.Quantity.Decimal.Float64()
	return &billingv1.InvoiceLine{
		Id:             l.GetID(),
		InvoiceId:      l.InvoiceID,
		ComponentId:    l.ComponentID,
		Description:    l.Description,
		Quantity:       qty,
		UnitPrice:      nullDecimalToMoney(l.UnitPrice, currency),
		Amount:         nullDecimalToMoney(l.Amount, currency),
		DiscountAmount: nullDecimalToMoney(l.DiscountAmount, currency),
		CreditAmount:   nullDecimalToMoney(l.CreditAmount, currency),
		NetAmount:      nullDecimalToMoney(l.NetAmount, currency),
		Currency:       l.Currency,
		LineType:       invoiceLineTypeToProto(l.LineType),
		Data:           jsonMapToStruct(l.Data),
	}
}

func creditGrantToProto(g *models.CreditGrant) *billingv1.CreditGrant {
	if g == nil {
		return nil
	}
	return &billingv1.CreditGrant{
		Id:              g.GetID(),
		CustomerId:      g.ProfileID,
		Name:            g.Name,
		OriginalAmount:  nullDecimalToMoney(g.OriginalAmount, g.Currency),
		RemainingAmount: nullDecimalToMoney(g.RemainingAmount, g.Currency),
		Currency:        g.Currency,
		ExpiresAt:       timePtrToTimestamp(g.ExpiresAt),
		Priority:        int32(g.Priority), //nolint:gosec // G115: priority is always a small positive integer
		Data:            jsonMapToStruct(g.Data),
	}
}

func discountToProto(d *models.Discount) *billingv1.Discount {
	if d == nil {
		return nil
	}
	val, _ := d.Value.Decimal.Float64()
	return &billingv1.Discount{
		Id:           d.GetID(),
		Name:         d.Name,
		DiscountType: discountTypeToProto(d.DiscountType),
		Value:        val,
		Currency:     d.Currency,
		ApplicableTo: jsonMapToStruct(d.ApplicableTo),
		StartAt:      timePtrToTimestamp(d.StartAt),
		EndAt:        timePtrToTimestamp(d.EndAt),
		//nolint:gosec // G115: max_applications is always a small positive integer
		MaxApplications: int32(d.MaxApplications),
		Data:            jsonMapToStruct(d.Data),
	}
}

func discountTypeToProto(dt string) billingv1.DiscountType {
	switch dt {
	case models.DiscountTypePercentage:
		return billingv1.DiscountType_DISCOUNT_PERCENTAGE
	case models.DiscountTypeFixed:
		return billingv1.DiscountType_DISCOUNT_FIXED
	default:
		return billingv1.DiscountType_DISCOUNT_PERCENTAGE
	}
}

func discountTypeFromProto(dt billingv1.DiscountType) string {
	switch dt {
	case billingv1.DiscountType_DISCOUNT_PERCENTAGE:
		return models.DiscountTypePercentage
	case billingv1.DiscountType_DISCOUNT_FIXED:
		return models.DiscountTypeFixed
	default:
		return models.DiscountTypePercentage
	}
}

func billingRunToProto(r *models.BillingRun) *billingv1.BillingRun {
	if r == nil {
		return nil
	}
	return &billingv1.BillingRun{
		Id:               r.GetID(),
		SubscriptionId:   r.SubscriptionID,
		CustomerId:       r.ProfileID,
		CatalogVersionId: r.CatalogVersionID,
		State:            billingRunStateToProto(r.State),
		PeriodStart:      timeToTimestamp(r.PeriodStart),
		PeriodEnd:        timeToTimestamp(r.PeriodEnd),
		StartedAt:        timePtrToTimestamp(r.StartedAt),
		CompletedAt:      timePtrToTimestamp(r.CompletedAt),
		FailedAt:         timePtrToTimestamp(r.FailedAt),
		ErrorMessage:     r.ErrorMessage,
		InvoiceId:        r.InvoiceID,
		Data:             jsonMapToStruct(r.Data),
	}
}
