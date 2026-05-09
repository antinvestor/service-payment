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

package models

// PricingModel defines how a component is priced.
const (
	PricingModelFlat      = "FLAT"
	PricingModelPerUnit   = "PER_UNIT"
	PricingModelTiered    = "TIERED"
	PricingModelVolume    = "VOLUME"
	PricingModelStairstep = "STAIRSTEP"
)

// AggregationType defines how usage events are aggregated.
const (
	AggregationTypeSum   = "SUM"
	AggregationTypeCount = "COUNT"
	AggregationTypeMax   = "MAX"
	AggregationTypeMin   = "MIN"
	AggregationTypeAvg   = "AVG"
	AggregationTypeLast  = "LAST"
)

// WindowGranularity defines the time window for metering.
const (
	WindowGranularityHour  = "HOUR"
	WindowGranularityDay   = "DAY"
	WindowGranularityMonth = "MONTH"
)

// SubscriptionState defines the lifecycle state of a subscription.
const (
	SubscriptionStateActive    = "ACTIVE"
	SubscriptionStateCancelled = "CANCELLED"
	SubscriptionStateExpired   = "EXPIRED"
	SubscriptionStatePending   = "PENDING"
)

// DiscountType defines the type of discount.
const (
	DiscountTypePercentage = "PERCENTAGE"
	DiscountTypeFixed      = "FIXED"
)

// CreditEntryType defines the type of credit ledger entry.
const (
	CreditEntryTypeGrant   = "GRANT"
	CreditEntryTypeConsume = "CONSUME"
	CreditEntryTypeExpire  = "EXPIRE"
	CreditEntryTypeRefund  = "REFUND"
)

// InvoiceState defines the lifecycle state of an invoice.
const (
	InvoiceStateDraft   = "DRAFT"
	InvoiceStateIssued  = "ISSUED"
	InvoiceStatePaid    = "PAID"
	InvoiceStateVoided  = "VOIDED"
	InvoiceStateOverdue = "OVERDUE"
)

// InvoiceLineType defines the type of invoice line item.
const (
	InvoiceLineTypeUsage    = "USAGE"
	InvoiceLineTypeFlat     = "FLAT"
	InvoiceLineTypeDiscount = "DISCOUNT"
	InvoiceLineTypeCredit   = "CREDIT"
)

// BillingRunState defines the state machine for billing runs.
const (
	BillingRunStatePending     = "PENDING"
	BillingRunStateMetering    = "METERING"
	BillingRunStateRating      = "RATING"
	BillingRunStateDiscounting = "DISCOUNTING"
	BillingRunStateCrediting   = "CREDITING"
	BillingRunStateInvoicing   = "INVOICING"
	BillingRunStatePosting     = "POSTING"
	BillingRunStateCompleted   = "COMPLETED"
	BillingRunStateFailed      = "FAILED"
)
