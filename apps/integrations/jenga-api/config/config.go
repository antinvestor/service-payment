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

// Header constants for credential resolution via queue message headers.
const (
	// HeaderConnectionCredentials is the settings key header for credential lookup.
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"

	HeaderMerchantCode   = "X-JENGA_MERCHANT_CODE"
	HeaderConsumerSecret = "X-JENGA_CONSUMER_SECRET"
	HeaderAPIKey         = "X-JENGA_API_KEY" //nolint:gosec // header name, not a credential
	HeaderCallbackURL    = "X-JENGA_CALLBACK_URL"
	HeaderEnvironment    = "X-JENGA_ENVIRONMENT"
	HeaderBaseURL        = "X-JENGA_BASE_URL"
	HeaderPrivateKeyPath = "X-JENGA_PRIVATE_KEY_PATH"
	HeaderPrivateKeyPEM  = "X-JENGA_PRIVATE_KEY_PEM" //nolint:gosec // settings key name, not a credential
)

type JengaConfig struct {
	config.ConfigurationDefault

	// Service dependencies
	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"                       required:"true"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                  env:"SETTINGS_SERVICE_URI"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration identifiers for per-tenant credential lookup
	SettingsIntegrationName string `envDefault:"Jenga"   env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default" env:"SETTINGS_INTEGRATION_ID"`

	// Jenga API credentials (defaults, overridden by per-tenant settings or headers).
	// The RSA signing key comes from a secret: either the PEM itself
	// (JENGA_PRIVATE_KEY_PEM) or a path to a mounted file (JENGA_PRIVATE_KEY_PATH).
	JengaPrivateKeyPEM string `env:"JENGA_PRIVATE_KEY_PEM"`
	JengaPrivateKey    string `env:"JENGA_PRIVATE_KEY_PATH"`
	//nolint:revive // ApiKey follows external API naming convention
	ApiKey         string `env:"JENGA_API_KEY"         required:"false"` //nolint:staticcheck // API field name
	ConsumerSecret string `env:"JENGA_CONSUMER_SECRET" required:"false"`
	MerchantCode   string `env:"JENGA_MERCHANT_CODE"   required:"false"`

	// Jenga environment and callback. JengaCallbackURL is the public base URL
	// of this service; the STK callback path and binding parameters are added
	// by the prompt worker. Environment (sandbox|production) selects the API
	// host unless JENGA_BASE_URL overrides it.
	JengaCallbackURL string `envDefault:"http://localhost" env:"JENGA_CALLBACK_URL" required:"true"`
	Environment      string `envDefault:"sandbox"          env:"JENGA_ENV"`
	BaseURL          string `env:"JENGA_BASE_URL"`
	// BaseURLAllowlist adds hosts (comma separated, host[:port]) to the known
	// Jenga hosts that a base URL may point at; meant for test servers.
	BaseURLAllowlist string `env:"JENGA_BASE_URL_ALLOWLIST"`

	// Queue configuration - payment queue for disbursements (tills pay, B2B)
	QueuePaymentName string `envDefault:"jenga.payments.dequeue"       env:"QUEUE_JENGA_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://jenga.payments.dequeue" env:"QUEUE_JENGA_PAYMENT_URI"`

	// Queue configuration - prompt queue for STK/USSD push
	QueuePromptName string `envDefault:"jenga.prompts.dequeue"       env:"QUEUE_JENGA_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://jenga.prompts.dequeue" env:"QUEUE_JENGA_PROMPT_URI"`
}
