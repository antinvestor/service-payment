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

package config

import (
	"github.com/pitabwire/frame/v2/config"
)

type BillingConfig struct {
	config.ConfigurationDefault

	CheckoutServiceURI                   string `envDefault:"127.0.0.1:7010"                           env:"CHECKOUT_SERVICE_URI"`
	CheckoutServiceWorkloadAPITargetPath string `envDefault:"/ns/payments/sa/service-payment-checkout" env:"CHECKOUT_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// CheckoutInvoiceReturnURL is the browser landing page after hosted checkout.
	// Should point at the admin/app payment-return route that calls ConfirmPayment.
	// Example: https://admin.stawi.org/billing/payment/return
	CheckoutInvoiceReturnURL string `envDefault:"http://localhost:5173/billing/payment/return" env:"CHECKOUT_INVOICE_RETURN_URL"`

	// Settlement is per-invoice Trustage one-shots (no bulk scan).
	// SettlementRetryDelaysMinutesCSV: minutes from checkout open for each poll
	// attempt (spread out). Default "2,5,15,30,60,120".
	SettlementRetryDelaysMinutesCSV string `envDefault:"2,5,15,30,60,120" env:"BILLING_SETTLEMENT_RETRY_DELAYS_MINUTES"`
	// SettlementMaxAttempts caps settle polls; then the Trustage reminder is archived.
	SettlementMaxAttempts int `envDefault:"6" env:"BILLING_SETTLEMENT_MAX_ATTEMPTS"`

	// Ledger account IDs used when posting invoice issue and payment capture.
	// Empty values skip the corresponding ledger posts (safe for local dev).
	LedgerARAccountID      string `env:"BILLING_LEDGER_AR_ACCOUNT_ID"`
	LedgerRevenueAccountID string `env:"BILLING_LEDGER_REVENUE_ACCOUNT_ID"`
	LedgerCashAccountID    string `env:"BILLING_LEDGER_CASH_ACCOUNT_ID"`

	// Subscription lifecycle fan-out (payment-style external integration).
	// Product services subscribe to this queue to grant/revoke entitlements.
	// Partition-specific IntegrationRoute rows override / augment this default.
	// Empty URI disables the global default publisher (DB routes still work).
	SubscriptionLifecycleTopicName string `envDefault:"subscription.lifecycle" env:"BILLING_SUBSCRIPTION_LIFECYCLE_TOPIC_NAME"`
	SubscriptionLifecycleTopicURI  string `env:"BILLING_SUBSCRIPTION_LIFECYCLE_TOPIC_URI"`

	// Payment service for server-side COF (card-on-file) renewals — Flutterwave v4
	// token charges only (payment_method_id + customer_id + recurring=true).
	PaymentServiceURI                   string `envDefault:"127.0.0.1:7000"                     env:"PAYMENT_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath string `envDefault:"/ns/payments/sa/service-payment"   env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Profile service — load checkout clues (saved payment_method_id / customer_id).
	ProfileServiceURI                   string `envDefault:"" env:"PROFILE_SERVICE_URI"`
	ProfileServiceWorkloadAPITargetPath string `envDefault:"/ns/identity/sa/service-profile" env:"PROFILE_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// RenewalLeadHours start collecting this many hours before period end.
	// Used when scheduling the per-subscription Trustage one-shot.
	RenewalLeadHours int `envDefault:"24" env:"BILLING_RENEWAL_LEAD_HOURS"`
	// RenewalRetryDelaysCSV hours after first due moment for each attempt
	// (spread out). Example "0,24,72,168" = immediate, +1d, +3d, +7d.
	// On failure Trustage is re-armed to the next slot for that subscription only.
	RenewalRetryDelaysCSV string `envDefault:"0,24,72,168" env:"BILLING_RENEWAL_RETRY_DELAYS_HOURS"`
	// RenewalMaxAttempts caps collection attempts per billing period (includes first).
	// After exhaustion the per-sub Trustage reminder is archived (past_due).
	RenewalMaxAttempts int `envDefault:"4" env:"BILLING_RENEWAL_MAX_ATTEMPTS"`
	// RenewalDefaultRoute is the payment route for COF charges (flutterwave v4 only).
	RenewalDefaultRoute string `envDefault:"flutterwave" env:"BILLING_RENEWAL_DEFAULT_ROUTE"`

	// InternalAdminToken authenticates Trustage callers of
	// POST /_internal/billing/subscriptions/{id}/renew and settlement.
	// Empty disables the HTTP trigger.
	InternalAdminToken string `env:"BILLING_INTERNAL_ADMIN_TOKEN"`

	// TrustageURL enables per-subscription renew reminders (Create/Activate/Archive workflows).
	// Empty → NoopRenewalScheduler (no automatic per-sub schedules).
	TrustageURL                   string `env:"TRUSTAGE_URL"`
	TrustageWorkloadAPITargetPath string `envDefault:"/ns/platform/sa/service-trustage" env:"TRUSTAGE_WORKLOAD_API_TARGET_PATH"`
	// BillingInternalBaseURL is the URL Trustage POSTs back to for each sub
	// (e.g. http://service-payment-billing.finance.svc:80). Required for schedules.
	BillingInternalBaseURL string `env:"BILLING_INTERNAL_BASE_URL"`
}
