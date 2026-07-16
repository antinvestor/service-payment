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

import "github.com/pitabwire/frame/v2/data"

// Integration route modes — mirror payment route modes for billing lifecycle.
const (
	// IntegrationRouteModeLifecycle delivers subscription lifecycle events
	// (created / activated / cancelled / billed) to external product systems.
	IntegrationRouteModeLifecycle = "lifecycle"
	// IntegrationRouteModeAny matches any mode (wildcard).
	IntegrationRouteModeAny = "any"
)

// Integration route types — filter which events a route receives.
const (
	// IntegrationRouteTypeAny matches all lifecycle events.
	IntegrationRouteTypeAny = "any"
)

// Subscription lifecycle event names published to integration routes.
const (
	SubscriptionEventCreated   = "subscription.created"
	SubscriptionEventActivated = "subscription.activated"
	SubscriptionEventCancelled = "subscription.cancelled"
	// SubscriptionEventBilled is emitted when an invoice linked to a
	// subscription is paid (signup or recurring).
	SubscriptionEventBilled = "subscription.billed"
)

// Well-known Subscription.Data keys for external-entity integration.
const (
	// SubDataExternalEntityID is the product-side entity this subscription entitles.
	SubDataExternalEntityID = "externalEntityId"
	// SubDataExternalEntityType classifies the entity (workspace, membership, seat, …).
	SubDataExternalEntityType = "externalEntityType"
	// SubDataIntegrationRouteID optionally pins delivery to one IntegrationRoute row.
	SubDataIntegrationRouteID = "integrationRouteId"
	// SubDataSignupInvoiceID links the first-charge invoice (existing).
	SubDataSignupInvoiceID = "signupInvoiceId"
	// Card-on-file instrument (Flutterwave v4 pmd_* / cus_*) for silent renewals.
	SubDataPaymentMethodID    = "paymentMethodId"
	SubDataProviderCustomerID = "providerCustomerId"
	SubDataPaymentProvider    = "paymentProvider"
	// CancelAtPeriodEnd mirrors product soft-cancel: skip rebill, expire after EndAt.
	SubDataCancelAtPeriodEnd = "cancelAtPeriodEnd"
	// Renewal bookkeeping (JSON-friendly strings / numbers).
	SubDataCurrentPeriodEnd   = "currentPeriodEnd"   // RFC3339
	SubDataLastRenewAttemptAt = "lastRenewAttemptAt" // RFC3339
	SubDataRenewAttemptCount  = "renewAttemptCount"  // int
	SubDataLastRenewInvoiceID = "lastRenewInvoiceID"
	SubDataRenewPeriodKey     = "renewPeriodKey" // idempotency: subscription:periodStart
	// Next Trustage reminder fire time (RFC3339). Derived from period + dunning.
	SubDataNextRenewAt = "nextRenewAt"
	// Trustage workflow name for this subscription's renew reminder.
	SubDataTrustageRenewWorkflow = "trustageRenewWorkflow"
	// Lifecycle event for failed rebill (product past_due UX).
	SubscriptionEventPaymentFailed = "subscription.payment_failed"
	SubscriptionEventPastDue       = "subscription.past_due"
)

// IntegrationRoute routes billing lifecycle events to external entity queues,
// analogous to payment Route (tx/rx) for Send/Receive.
//
// Selection (same shape as payment routes):
//   - partition_id from tenancy / subscription
//   - mode = lifecycle (or any)
//   - route_type = event name or "any"
//
// URI is a queue URL that product integrators subscribe to
// (e.g. nats://…, mem://product.subscription.lifecycle).
type IntegrationRoute struct {
	data.BaseModel

	Name        string `gorm:"type:varchar(100);not null"  json:"name"`
	Description string `gorm:"type:text"                   json:"description"`
	// RouteType is the lifecycle event filter: subscription.activated, any, …
	RouteType string `gorm:"type:varchar(64);not null;index" json:"route_type"`
	// Mode is typically "lifecycle".
	Mode string `gorm:"type:varchar(32);not null;index" json:"mode"`
	// URI is the queue destination integrators consume.
	URI string `gorm:"type:varchar(512);not null" json:"uri"`
}
