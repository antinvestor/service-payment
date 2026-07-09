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
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"
	HeaderAPIKey                = "X-API_KEY"
	HeaderWebhookSecret         = "X-API_WEBHOOK_SECRET" //nolint:gosec // header name, not a credential
)

type StripeConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7010"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration lookup
	SettingsIntegrationName string `envDefault:"stripe" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"stripe" env:"SETTINGS_INTEGRATION_ID"`

	// Stripe API credentials (optional defaults, can be overridden via headers)
	APIKey        string `env:"STRIPE_API_KEY"`
	WebhookSecret string `env:"STRIPE_WEBHOOK_SECRET"`

	// Queue configuration - payment queue for Payouts
	QueuePaymentName string `envDefault:"stripe.payments.dequeue"       env:"QUEUE_STRIPE_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://stripe.payments.dequeue" env:"QUEUE_STRIPE_PAYMENT_URI"`

	// Queue configuration - prompt queue for PaymentIntent creation
	QueuePromptName string `envDefault:"stripe.prompts.dequeue"       env:"QUEUE_STRIPE_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://stripe.prompts.dequeue" env:"QUEUE_STRIPE_PROMPT_URI"`
}
