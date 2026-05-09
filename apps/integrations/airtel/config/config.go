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
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"
	HeaderClientID              = "X-API_CLIENT_ID"
	HeaderClientSecret          = "X-API_CLIENT_SECRET" //nolint:gosec // header name, not a credential
	HeaderCallbackURL           = "X-API_CALLBACK_URL"
	HeaderCountryCode           = "X-API_COUNTRY_CODE"
	HeaderCurrency              = "X-API_CURRENCY"
	HeaderEnvironment           = "X-API_ENVIRONMENT"
)

type AirtelConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7010"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration lookup
	SettingsIntegrationName string `envDefault:"airtel" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"airtel" env:"SETTINGS_INTEGRATION_ID"`

	// Airtel Money API credentials (optional defaults, can be overridden via headers)
	ClientID     string `env:"AIRTEL_CLIENT_ID"`
	ClientSecret string `env:"AIRTEL_CLIENT_SECRET"`
	CallbackURL  string `env:"AIRTEL_CALLBACK_URL"`
	CountryCode  string `env:"AIRTEL_COUNTRY_CODE"  envDefault:"UG"`
	Currency     string `env:"AIRTEL_CURRENCY"      envDefault:"UGX"`

	// Environment: sandbox or production
	Environment string `envDefault:"sandbox" env:"AIRTEL_ENVIRONMENT"`

	// Queue configuration - payment queue for disbursements
	QueuePaymentName string `envDefault:"airtel.payments.dequeue"       env:"QUEUE_AIRTEL_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://airtel.payments.dequeue" env:"QUEUE_AIRTEL_PAYMENT_URI"`

	// Queue configuration - prompt queue for USSD Push
	QueuePromptName string `envDefault:"airtel.prompts.dequeue"       env:"QUEUE_AIRTEL_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://airtel.prompts.dequeue" env:"QUEUE_AIRTEL_PROMPT_URI"`
}
