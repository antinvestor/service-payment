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

	// SettlementSweepIntervalSeconds is how often billing auto-confirms completed
	// checkout sessions for open invoices (abandoned browser recovery). 0 disables.
	SettlementSweepIntervalSeconds int `envDefault:"60" env:"BILLING_SETTLEMENT_SWEEP_INTERVAL_SECONDS"`
	// SettlementSweepBatchSize limits invoices processed per sweep tick.
	SettlementSweepBatchSize int `envDefault:"50" env:"BILLING_SETTLEMENT_SWEEP_BATCH_SIZE"`

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
}
