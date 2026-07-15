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

// Credential / header keys for multi-tenant override (settings or queue headers).
const (
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"
	HeaderClientID              = "X-API_CLIENT_ID"
	HeaderClientSecret          = "X-API_CLIENT_SECRET"  //nolint:gosec // header name
	HeaderWebhookSecret         = "X-API_WEBHOOK_SECRET" //nolint:gosec // header name
	HeaderEnvironment           = "X-API_ENVIRONMENT"
	HeaderRedirectURL           = "X-API_REDIRECT_URL"
)

// FlutterwaveConfig holds service and Flutterwave v4 provider configuration.
//
// v4 uses OAuth 2.0 (client_id + client_secret → short-lived access token).
// Docs: https://developer.flutterwave.com/docs/getting-started
//
// Credentials can be set globally via env, overridden per message via headers,
// or loaded from the settings service (X-API_CONNECTION_CREDENTIALS → settings key).
type FlutterwaveConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7010"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	SettingsIntegrationName string `envDefault:"flutterwave" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"flutterwave" env:"SETTINGS_INTEGRATION_ID"`

	// Flutterwave v4 OAuth credentials (dashboard → Settings → API Keys → v4).
	// Sandbox: developer sandbox portal. Live: switch to v4 live API keys.
	ClientID     string `env:"FLUTTERWAVE_CLIENT_ID"`
	ClientSecret string `env:"FLUTTERWAVE_CLIENT_SECRET"`
	// WebhookSecret is the secret hash used for flutterwave-signature (HMAC-SHA256).
	WebhookSecret string `env:"FLUTTERWAVE_WEBHOOK_SECRET"`
	// Environment: sandbox | production. Selects API base URL.
	Environment string `envDefault:"sandbox" env:"FLUTTERWAVE_ENVIRONMENT"`

	// OAuth token endpoint (same for sandbox and production).
	OAuthTokenURL string `envDefault:"https://idp.flutterwave.com/realms/flutterwave/protocol/openid-connect/token" env:"FLUTTERWAVE_OAUTH_TOKEN_URL"`

	// Optional base URL overrides (defaults from Environment).
	SandboxAPIBaseURL    string `envDefault:"https://developersandbox-api.flutterwave.com" env:"FLUTTERWAVE_SANDBOX_API_BASE_URL"`
	ProductionAPIBaseURL string `envDefault:"https://f4bexperience.flutterwave.com"          env:"FLUTTERWAVE_PRODUCTION_API_BASE_URL"`

	// DefaultRedirectURL used when the prompt has no success/return URL.
	DefaultRedirectURL string `envDefault:"http://localhost:8080/webhook/flutterwave/return" env:"FLUTTERWAVE_DEFAULT_REDIRECT_URL"`

	// PublicWebhookBase for transfer callback_url if set.
	PublicWebhookBase string `env:"FLUTTERWAVE_PUBLIC_WEBHOOK_BASE"`

	// Default non-MoMo payment method type for redirect collection when no phone:
	// bank_transfer | opay | ussd (requires bank code in extras).
	DefaultCollectionMethod string `envDefault:"bank_transfer" env:"FLUTTERWAVE_DEFAULT_COLLECTION_METHOD"`

	// Queue: outbound disbursements (Payment Send → route → this queue).
	QueuePaymentName string `envDefault:"flutterwave.payments.dequeue"       env:"QUEUE_FLUTTERWAVE_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://flutterwave.payments.dequeue" env:"QUEUE_FLUTTERWAVE_PAYMENT_URI"`

	// Queue: collections (InitiatePrompt).
	QueuePromptName string `envDefault:"flutterwave.prompts.dequeue"       env:"QUEUE_FLUTTERWAVE_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://flutterwave.prompts.dequeue" env:"QUEUE_FLUTTERWAVE_PROMPT_URI"`

	// Queue: billing subscription lifecycle (optional; our billing is source of truth).
	QueueSubscriptionName string `envDefault:"subscription.lifecycle"       env:"QUEUE_FLUTTERWAVE_SUBSCRIPTION_NAME"`
	QueueSubscriptionURI  string `envDefault:"mem://subscription.lifecycle" env:"QUEUE_FLUTTERWAVE_SUBSCRIPTION_URI"`

	// EnableSubscriptionWorker turns on the lifecycle subscriber.
	// v4 has no payment-plans API like v3; this worker is for correlation/logging
	// and optional cancel hooks when flutterwave charge ids are stored.
	EnableSubscriptionWorker bool `envDefault:"true" env:"FLUTTERWAVE_ENABLE_SUBSCRIPTION_WORKER"`
}
