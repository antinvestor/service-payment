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

	// HeaderSubscriptionKey is the header for MTN subscription key.
	HeaderSubscriptionKey = "X-MTN_SUBSCRIPTION_KEY"
	HeaderAPIUser         = "X-MTN_API_USER"
	HeaderAPIKey          = "X-MTN_API_KEY" //nolint:gosec // header name, not a credential
	HeaderCallbackURL     = "X-MTN_CALLBACK_URL"
	HeaderCurrency        = "X-MTN_CURRENCY"
	HeaderEnvironment     = "X-MTN_ENVIRONMENT"
)

type MtnConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration identifiers for credential lookup
	SettingsIntegrationName string `envDefault:"MtnMomo" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default" env:"SETTINGS_INTEGRATION_ID"`

	// MTN MoMo API credentials (defaults, overridden by headers)
	SubscriptionKey string `env:"MTN_SUBSCRIPTION_KEY"`
	APIUser         string `env:"MTN_API_USER"`
	APIKey          string `env:"MTN_API_KEY"`
	CallbackURL     string `env:"MTN_CALLBACK_URL"`
	Currency        string `env:"MTN_CURRENCY"         envDefault:"EUR"`

	// Environment: sandbox or production
	Environment string `envDefault:"sandbox" env:"MTN_ENVIRONMENT"`

	// Queue configuration - payment queue for disbursements
	QueuePaymentName string `envDefault:"mtn.payments.dequeue"       env:"QUEUE_MTN_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://mtn.payments.dequeue" env:"QUEUE_MTN_PAYMENT_URI"`

	// Queue configuration - prompt queue for requestToPay
	QueuePromptName string `envDefault:"mtn.prompts.dequeue"       env:"QUEUE_MTN_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://mtn.prompts.dequeue" env:"QUEUE_MTN_PROMPT_URI"`
}

// BaseURL returns the MTN MoMo API base URL based on environment.
func (c *MtnConfig) BaseURL() string {
	if c.Environment == "production" {
		return "https://proxy.momoapi.mtn.com"
	}
	return "https://sandbox.momodeveloper.mtn.com"
}
