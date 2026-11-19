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

package collection

import "time"

// Provider-managed subscription contracts shared by billing and the
// integration workers (Polar, Stripe).

// Normalized provider subscription statuses. Integrations map the
// provider's own vocabulary onto these before calling billing.
const (
	SubscriptionStatusActive            = "active"
	SubscriptionStatusTrialing          = "trialing"
	SubscriptionStatusPastDue           = "past_due"
	SubscriptionStatusIncomplete        = "incomplete"
	SubscriptionStatusIncompleteExpired = "incomplete_expired"
	SubscriptionStatusCanceled          = "canceled"
	SubscriptionStatusRevoked           = "revoked"
	SubscriptionStatusUnpaid            = "unpaid"
	SubscriptionStatusPaused            = "paused"
)

// Normalized provider events billing understands.
const (
	// SubscriptionEventUpdated carries a status change (create, activate,
	// cancel, revoke, past due, …): billing applies the status.
	SubscriptionEventUpdated = "subscription.updated"
	// SubscriptionEventBilled reports a successful charge for a period
	// (signup or renewal): billing advances the period and emits billed.
	SubscriptionEventBilled = "subscription.billed"
	// SubscriptionEventPaymentFailed reports a failed renewal charge.
	SubscriptionEventPaymentFailed = "subscription.payment_failed"
)

// Provider command names billing sends to integration workers.
const (
	CommandCancelSubscription = "subscription.cancel"
)

// ProviderCommand is the JSON message billing publishes on a provider's
// command queue (BILLING_PROVIDER_COMMAND_URIS) and the integration worker
// executes against the provider API.
type ProviderCommand struct {
	Provider string `json:"provider"`
	Command  string `json:"command"`
	// SubscriptionID is billing's subscription id.
	SubscriptionID string `json:"subscription_id"`
	// ExternalSubscriptionID is the provider's subscription id.
	ExternalSubscriptionID string `json:"external_subscription_id"`
	// AtPeriodEnd asks for a soft cancel; false revokes immediately.
	AtPeriodEnd bool      `json:"at_period_end"`
	TenantID    string    `json:"tenant_id,omitempty"`
	PartitionID string    `json:"partition_id,omitempty"`
	IssuedAt    time.Time `json:"issued_at"`
}

// CommandPublisherRef is the queue publisher reference billing registers for
// a provider's command queue.
func CommandPublisherRef(provider string) string {
	return "provider." + provider + ".commands"
}
