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
	HeaderOrganizationID        = "X-API_ORGANIZATION_ID"
	HeaderEnvironment           = "X-API_ENVIRONMENT"
)

type PolarConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7010"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration lookup
	SettingsIntegrationName string `envDefault:"polar" env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"polar" env:"SETTINGS_INTEGRATION_ID"`

	// Polar API credentials (optional defaults, can be overridden via headers)
	APIKey         string `env:"POLAR_API_KEY"`
	WebhookSecret  string `env:"POLAR_WEBHOOK_SECRET"`
	OrganizationID string `env:"POLAR_ORGANIZATION_ID"`

	// Environment: sandbox or production
	Environment string `envDefault:"sandbox" env:"POLAR_ENVIRONMENT"`

	// Queue configuration - payment queue (no-op for Polar, no disbursements)
	QueuePaymentName string `envDefault:"polar.payments.dequeue"       env:"QUEUE_POLAR_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://polar.payments.dequeue" env:"QUEUE_POLAR_PAYMENT_URI"`

	// Queue configuration - prompt queue for checkout session creation
	QueuePromptName string `envDefault:"polar.prompts.dequeue"       env:"QUEUE_POLAR_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://polar.prompts.dequeue" env:"QUEUE_POLAR_PROMPT_URI"`
}
