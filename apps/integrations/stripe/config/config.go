package config

import (
	"github.com/pitabwire/frame/config"
)

const (
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"
	HeaderAPIKey                = "X-API_KEY"
	HeaderWebhookSecret         = "X-API_WEBHOOK_SECRET" //nolint:gosec // header name, not a credential
)

type StripeConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI  string `envDefault:"127.0.0.1:7006" env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI string `envDefault:"127.0.0.1:7010" env:"SETTINGS_SERVICE_URI"`

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
