package config

import (
	"github.com/pitabwire/frame/config"
)

const (
	// HeaderConnectionCredentials is the settings key header for credential lookup.
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"

	// HeaderConsumerKey is the header for M-Pesa consumer key.
	HeaderConsumerKey = "X-MPESA_CONSUMER_KEY"
	// HeaderConsumerSecret is the header for M-Pesa consumer secret.
	HeaderConsumerSecret = "X-MPESA_CONSUMER_SECRET" //nolint:gosec // header name, not a credential
	HeaderShortcode      = "X-MPESA_SHORTCODE"
	// HeaderPasskey is the header for M-Pesa passkey.
	HeaderPasskey            = "X-MPESA_PASSKEY" //nolint:gosec // header name, not a credential
	HeaderCallbackURL        = "X-MPESA_CALLBACK_URL"
	HeaderInitiatorName      = "X-MPESA_INITIATOR_NAME"
	HeaderSecurityCredential = "X-MPESA_SECURITY_CREDENTIAL" //nolint:gosec // header name, not a credential
	HeaderEnvironment        = "X-MPESA_ENVIRONMENT"
)

type MpesaConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI  string `envDefault:"127.0.0.1:7006" env:"PAYMENT_SERVICE_URI"`
	SettingsServiceURI string `envDefault:"127.0.0.1:7005" env:"SETTINGS_SERVICE_URI"`

	// Settings integration identifiers for credential lookup
	SettingsIntegrationName string `envDefault:"Mpesa"   env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default" env:"SETTINGS_INTEGRATION_ID"`

	// M-Pesa Daraja API credentials (defaults, overridden by headers)
	ConsumerKey    string `env:"MPESA_CONSUMER_KEY"`
	ConsumerSecret string `env:"MPESA_CONSUMER_SECRET"`
	Shortcode      string `env:"MPESA_SHORTCODE"`
	Passkey        string `env:"MPESA_PASSKEY"`
	CallbackURL    string `env:"MPESA_CALLBACK_URL"`

	// B2C specific credentials
	InitiatorName      string `env:"MPESA_INITIATOR_NAME"`
	SecurityCredential string `env:"MPESA_SECURITY_CREDENTIAL"`

	// Environment: sandbox or production
	Environment string `envDefault:"sandbox" env:"MPESA_ENVIRONMENT"`

	// Queue configuration - payment queue for B2C disbursements
	QueuePaymentName string `envDefault:"mpesa.payments.dequeue"       env:"QUEUE_MPESA_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://mpesa.payments.dequeue" env:"QUEUE_MPESA_PAYMENT_URI"`

	// Queue configuration - prompt queue for STK Push
	QueuePromptName string `envDefault:"mpesa.prompts.dequeue"       env:"QUEUE_MPESA_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://mpesa.prompts.dequeue" env:"QUEUE_MPESA_PROMPT_URI"`
}

// BaseURL returns the Daraja API base URL based on environment.
func (c *MpesaConfig) BaseURL() string {
	if c.Environment == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}
