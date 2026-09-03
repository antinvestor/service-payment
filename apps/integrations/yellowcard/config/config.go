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

const (
	// HeaderConnectionCredentials is the settings key header for credential lookup.
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"

	// HeaderAPIKey is the header for the Yellow Card API key.
	HeaderAPIKey = "X-YELLOWCARD_API_KEY" //nolint:gosec // header name, not a credential
	// HeaderSecretKey is the header for the Yellow Card API secret used for HMAC signing.
	HeaderSecretKey = "X-YELLOWCARD_SECRET_KEY" //nolint:gosec // header name, not a credential
	// HeaderEnvironment is the header for the Yellow Card environment (sandbox or production).
	HeaderEnvironment = "X-YELLOWCARD_ENVIRONMENT"
	// HeaderCountry is the header for the default ISO 3166-1 alpha-2 country code.
	HeaderCountry = "X-YELLOWCARD_COUNTRY"
	// HeaderCurrency is the header for the default ISO 4217 currency code.
	HeaderCurrency = "X-YELLOWCARD_CURRENCY"
	// HeaderNetwork is the header for the default mobile money / bank network (id, code or name).
	HeaderNetwork = "X-YELLOWCARD_NETWORK"
	// HeaderChannelType is the header for the default channel type (momo or bank).
	HeaderChannelType = "X-YELLOWCARD_CHANNEL_TYPE"
	// HeaderCustomerType is the header for the KYC customer type (retail or institution).
	HeaderCustomerType = "X-YELLOWCARD_CUSTOMER_TYPE"
	// HeaderBusinessID is the header for the institution business id.
	HeaderBusinessID = "X-YELLOWCARD_BUSINESS_ID"
	// HeaderBusinessName is the header for the institution business name.
	HeaderBusinessName = "X-YELLOWCARD_BUSINESS_NAME"
	// HeaderWebhookSecret is the header for the webhook signing secret override.
	HeaderWebhookSecret = "X-YELLOWCARD_WEBHOOK_SECRET" //nolint:gosec // header name, not a credential
)

type YellowcardConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration identifiers for credential lookup
	SettingsIntegrationName string `envDefault:"Yellowcard" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default"    env:"SETTINGS_INTEGRATION_ID"`

	// Yellow Card API credentials (defaults, overridden by headers or settings)
	APIKey    string `env:"YELLOWCARD_API_KEY"`
	SecretKey string `env:"YELLOWCARD_SECRET_KEY"`

	// Environment: sandbox or production
	Environment string `envDefault:"sandbox" env:"YELLOWCARD_ENVIRONMENT"`

	// Country is the default ISO 3166-1 alpha-2 country used when it cannot
	// be derived from the customer's phone number or the payment extras.
	Country string `env:"YELLOWCARD_COUNTRY"`
	// Currency is the default ISO 4217 currency code.
	Currency string `env:"YELLOWCARD_CURRENCY"`
	// Network is the default network (id, code or name) for mobile money.
	Network string `env:"YELLOWCARD_NETWORK"`
	// ChannelType forces momo or bank; empty selects automatically.
	ChannelType string `env:"YELLOWCARD_CHANNEL_TYPE"`

	// CustomerType is the KYC validation mode: retail or institution.
	CustomerType string `envDefault:"retail" env:"YELLOWCARD_CUSTOMER_TYPE"`
	BusinessID   string `env:"YELLOWCARD_BUSINESS_ID"`
	BusinessName string `env:"YELLOWCARD_BUSINESS_NAME"`

	// WebhookSecret overrides the secret used to verify X-YC-Signature.
	// When empty the API secret key is used, as Yellow Card documents.
	WebhookSecret string `env:"YELLOWCARD_WEBHOOK_SECRET"`

	// DefaultRedirectURL is sent as redirectUrl for channels that require a
	// browser redirect (payment links) when the prompt carries none.
	DefaultRedirectURL string `env:"YELLOWCARD_DEFAULT_REDIRECT_URL"`

	// CatalogCacheSeconds bounds how long channels and networks are cached.
	CatalogCacheSeconds int `envDefault:"300" env:"YELLOWCARD_CATALOG_CACHE_SECONDS"`

	// Queue configuration - payment queue for sends (disbursements)
	QueuePaymentName string `envDefault:"yellowcard.payments.dequeue"       env:"QUEUE_YELLOWCARD_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://yellowcard.payments.dequeue" env:"QUEUE_YELLOWCARD_PAYMENT_URI"`

	// Queue configuration - prompt queue for receives (collections)
	QueuePromptName string `envDefault:"yellowcard.prompts.dequeue"       env:"QUEUE_YELLOWCARD_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://yellowcard.prompts.dequeue" env:"QUEUE_YELLOWCARD_PROMPT_URI"`
}
