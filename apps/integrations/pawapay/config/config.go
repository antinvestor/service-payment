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
	"github.com/pitabwire/frame/config"
)

const (
	// HeaderConnectionCredentials is the settings key header for credential lookup.
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"

	// HeaderAPIToken is the header for the pawaPay API bearer token.
	HeaderAPIToken = "X-PAWAPAY_API_TOKEN" //nolint:gosec // header name, not a credential
	// HeaderEnvironment is the header for the pawaPay environment (sandbox or production).
	HeaderEnvironment = "X-PAWAPAY_ENVIRONMENT"
	// HeaderProvider is the header for the default mobile money provider code.
	HeaderProvider = "X-PAWAPAY_PROVIDER"
	// HeaderCountry is the header for the default ISO 3166-1 alpha-3 country code.
	HeaderCountry = "X-PAWAPAY_COUNTRY"
)

type PawapayConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration identifiers for credential lookup
	SettingsIntegrationName string `envDefault:"Pawapay" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default" env:"SETTINGS_INTEGRATION_ID"`

	// pawaPay API credentials (defaults, overridden by headers or settings)
	APIToken string `env:"PAWAPAY_API_TOKEN"`

	// Default provider code (e.g. MTN_MOMO_ZMB). When empty, the provider is
	// predicted from the phone number via the predict-provider endpoint.
	Provider string `env:"PAWAPAY_PROVIDER"`

	// Default ISO 3166-1 alpha-3 country code.
	Country string `env:"PAWAPAY_COUNTRY"`

	// Environment: sandbox or production
	Environment string `envDefault:"sandbox" env:"PAWAPAY_ENVIRONMENT"`

	// Queue configuration - payment queue for payouts (disbursements)
	QueuePaymentName string `envDefault:"pawapay.payments.dequeue"       env:"QUEUE_PAWAPAY_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://pawapay.payments.dequeue" env:"QUEUE_PAWAPAY_PAYMENT_URI"`

	// Queue configuration - prompt queue for deposits (collections)
	QueuePromptName string `envDefault:"pawapay.prompts.dequeue"       env:"QUEUE_PAWAPAY_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://pawapay.prompts.dequeue" env:"QUEUE_PAWAPAY_PROMPT_URI"`
}
