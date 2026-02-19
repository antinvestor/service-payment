package config

import (
	"github.com/pitabwire/frame/config"
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

	PaymentServiceURI  string `envDefault:"127.0.0.1:7006" env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI string `envDefault:"127.0.0.1:7010" env:"SETTINGS_SERVICE_URI"`

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
