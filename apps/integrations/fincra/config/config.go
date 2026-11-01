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
	"github.com/pitabwire/frame/v3/config"
)

const (
	// HeaderConnectionCredentials is the settings key header for credential lookup.
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"

	// HeaderSecretKey is the Fincra secret key, sent as the api-key header.
	HeaderSecretKey = "X-FINCRA_SECRET_KEY" //nolint:gosec // header name, not a credential
	// HeaderPublicKey is the Fincra public key, sent as the x-pub-key header.
	HeaderPublicKey = "X-FINCRA_PUBLIC_KEY"
	// HeaderBusinessID is the Fincra parent business id (x-business-id).
	HeaderBusinessID = "X-FINCRA_BUSINESS_ID"
	// HeaderWebhookSecret is the Fincra webhook encryption key.
	HeaderWebhookSecret = "X-FINCRA_WEBHOOK_SECRET" //nolint:gosec // header name, not a credential
	// HeaderEnvironment is sandbox or production.
	HeaderEnvironment = "X-FINCRA_ENVIRONMENT"
	// HeaderCollectionMode is checkout (hosted page) or direct (direct charge).
	HeaderCollectionMode = "X-FINCRA_COLLECTION_MODE"
	// HeaderFeeBearer is business or customer.
	HeaderFeeBearer = "X-FINCRA_FEE_BEARER"
	// HeaderCountry is the default ISO 3166-1 alpha-2 payout country.
	HeaderCountry = "X-FINCRA_COUNTRY"
	// HeaderTenantID and HeaderPartitionID bind a settings connection to the
	// tenancy that owns it. Webhooks verified with a connection are recorded
	// only under its bound tenancy; a connection without one is refused.
	HeaderTenantID    = "X-FINCRA_TENANT_ID"
	HeaderPartitionID = "X-FINCRA_PARTITION_ID"
)

type FincraConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                  env:"SETTINGS_SERVICE_URI"`
	ProfileServiceURI                    string `envDefault:"127.0.0.1:7003"                  env:"PROFILE_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`
	ProfileServiceWorkloadAPITargetPath  string `envDefault:"/ns/profile/sa/service-profile"  env:"PROFILE_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration identifiers for credential lookup.
	SettingsIntegrationName string `envDefault:"Fincra"  env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default" env:"SETTINGS_INTEGRATION_ID"`

	// Fincra API credentials (defaults, overridden by headers or settings).
	SecretKey     string `env:"FINCRA_SECRET_KEY"`
	PublicKey     string `env:"FINCRA_PUBLIC_KEY"`
	BusinessID    string `env:"FINCRA_BUSINESS_ID"`
	WebhookSecret string `env:"FINCRA_WEBHOOK_SECRET"`

	// Environment: sandbox or production.
	Environment string `envDefault:"sandbox" env:"FINCRA_ENVIRONMENT"`
	// BaseURL overrides the environment's API host (tests, proxies).
	BaseURL string `env:"FINCRA_BASE_URL"`

	// CollectionMode is "checkout" (Fincra hosted page, no extra approval) or
	// "direct" (direct charge API, which Fincra enables per business).
	CollectionMode string `envDefault:"checkout" env:"FINCRA_COLLECTION_MODE"`
	// FeeBearer is "business" (the merchant pays fees, the payer pays exactly
	// the amount asked) or "customer".
	FeeBearer string `envDefault:"business" env:"FINCRA_FEE_BEARER"`
	// CheckoutPaymentMethods limits the hosted checkout's methods
	// (comma-separated: card, bank_transfer, mobile_money, payattitude, opay).
	// Empty lets Fincra offer every method enabled for the currency.
	CheckoutPaymentMethods string `env:"FINCRA_CHECKOUT_PAYMENT_METHODS"`
	// DefaultRedirectURL is where the hosted checkout returns the payer when
	// the prompt carries no success_url / redirect_url.
	DefaultRedirectURL string `env:"FINCRA_DEFAULT_REDIRECT_URL"`
	// FallbackCustomerEmail is sent when the payer has no email: Fincra
	// requires customer.email on every collection.
	FallbackCustomerEmail string `env:"FINCRA_FALLBACK_CUSTOMER_EMAIL"`

	// WebhookBaseURL is this service's public URL. Direct charges carry a
	// per-request webhookUrl built from it (with tenancy and connection);
	// hosted checkouts use the URL set on the Fincra dashboard.
	WebhookBaseURL string `env:"FINCRA_WEBHOOK_BASE_URL"`

	// Country is the default payout country when the payee's number does not
	// identify one.
	Country string `env:"FINCRA_COUNTRY"`
	// ChargeOperators maps a currency or mobile network to the operator code
	// sent on a mobile money direct charge, e.g. {"KES":"MPESA","GHS":"MTN"}.
	ChargeOperators string `env:"FINCRA_CHARGE_OPERATORS"`
	// PayoutProviders maps a currency or mobile network to the
	// beneficiary.mobileMoneyCode for payouts, e.g. {"KES":"MPESA"}.
	PayoutProviders string `env:"FINCRA_PAYOUT_PROVIDERS"`
	// DefaultBeneficiaryName is used when the payee's profile has no name.
	DefaultBeneficiaryName string `envDefault:"Mobile Money Customer" env:"FINCRA_DEFAULT_BENEFICIARY_NAME"`

	// CatalogCacheSeconds bounds how long payout providers are cached.
	CatalogCacheSeconds int `envDefault:"600" env:"FINCRA_CATALOG_CACHE_SECONDS"`

	// Queue configuration - payment queue for payouts (disbursements).
	QueuePaymentName string `envDefault:"fincra.payments.dequeue"       env:"QUEUE_FINCRA_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://fincra.payments.dequeue" env:"QUEUE_FINCRA_PAYMENT_URI"`

	// Queue configuration - prompt queue for collections.
	QueuePromptName string `envDefault:"fincra.prompts.dequeue"       env:"QUEUE_FINCRA_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://fincra.prompts.dequeue" env:"QUEUE_FINCRA_PROMPT_URI"`
}
