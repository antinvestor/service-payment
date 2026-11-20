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
	"errors"
	"strings"

	"github.com/pitabwire/frame/v3/config"
)

const (
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"
	HeaderAPIKey                = "X-POLAR_API_KEY"
	HeaderWebhookSecret         = "X-POLAR_WEBHOOK_SECRET" //nolint:gosec // header name, not a credential
	HeaderOrganizationID        = "X-POLAR_ORGANIZATION_ID"
	HeaderEnvironment           = "X-POLAR_ENVIRONMENT"

	EnvironmentSandbox    = "sandbox"
	EnvironmentProduction = "production"
)

type PolarConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                    string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                  env:"SETTINGS_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings" env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`
	// Billing receives subscription and order webhooks (SyncProviderSubscription).
	// Empty disables the mirror: webhooks then only settle checkout prompts.
	BillingServiceURI                   string `env:"BILLING_SERVICE_URI"`
	BillingServiceWorkloadAPITargetPath string `envDefault:"/ns/payments/sa/service-billing" env:"BILLING_SERVICE_WORKLOAD_API_TARGET_PATH"`

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

	// Queue configuration - subscription commands from billing
	// (BILLING_PROVIDER_COMMAND_URIS["polar"] must be this queue).
	QueueSubscriptionName string `envDefault:"polar.subscriptions.dequeue"       env:"QUEUE_POLAR_SUBSCRIPTION_NAME"`
	QueueSubscriptionURI  string `envDefault:"mem://polar.subscriptions.dequeue" env:"QUEUE_POLAR_SUBSCRIPTION_URI"`
}

// Validate refuses a configuration that would accept forged webhooks or
// silently talk to the sandbox in production.
func (c *PolarConfig) Validate() error {
	env := strings.ToLower(strings.TrimSpace(c.Environment))
	if env != EnvironmentSandbox && env != EnvironmentProduction {
		return errors.New("POLAR_ENVIRONMENT must be sandbox or production")
	}
	c.Environment = env
	if strings.TrimSpace(c.WebhookSecret) == "" {
		return errors.New("POLAR_WEBHOOK_SECRET is required: unsigned webhooks are never accepted")
	}
	return nil
}
