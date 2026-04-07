package config

import (
	"github.com/pitabwire/frame/config"
)

// Header constants for credential resolution via queue message headers.
const (
	// HeaderConnectionCredentials is the settings key header for credential lookup.
	HeaderConnectionCredentials = "X-API_CONNECTION_CREDENTIALS"

	HeaderMerchantCode   = "X-JENGA_MERCHANT_CODE"
	HeaderConsumerSecret = "X-JENGA_CONSUMER_SECRET"
	HeaderAPIKey         = "X-JENGA_API_KEY"
	HeaderCallbackURL    = "X-JENGA_CALLBACK_URL"
	HeaderEnvironment    = "X-JENGA_ENVIRONMENT"
	HeaderPrivateKeyPath = "X-JENGA_PRIVATE_KEY_PATH"
)

type JengaConfig struct {
	config.ConfigurationDefault

	// Service dependencies
	PaymentServiceURI                    string `envDefault:"localhost:50051"                                         env:"PAYMENT_SERVICE_URI"                      required:"true"`
	PaymentServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-payment"                         env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	SettingsServiceURI                   string `envDefault:"127.0.0.1:7005"                                         env:"SETTINGS_SERVICE_URI"`
	SettingsServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-settings"                         env:"SETTINGS_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// Settings integration identifiers for per-tenant credential lookup
	SettingsIntegrationName string `envDefault:"Jenga"   env:"SETTINGS_INTEGRATION_NAME"`
	SettingsIntegrationID   string `envDefault:"Default" env:"SETTINGS_INTEGRATION_ID"`

	// Jenga API credentials (defaults, overridden by per-tenant settings or headers)
	JengaPrivateKey string `envDefault:"/keys/privatekey.pem" env:"JENGA_PRIVATE_KEY_PATH"`
	//nolint:revive // ApiKey follows external API naming convention
	ApiKey         string `env:"JENGA_API_KEY"         required:"false"` //nolint:staticcheck // API field name
	ConsumerSecret string `env:"JENGA_CONSUMER_SECRET" required:"false"`
	MerchantCode   string `env:"JENGA_MERCHANT_CODE"   required:"false"`

	// Jenga environment and callback
	JengaCallbackURL string `envDefault:"http://localhost/receivepayments" env:"JENGA_CALLBACK_URL" required:"true"`
	Env              string `envDefault:"https://uat.finserve.africa"      env:"JENGA_ENV"`

	// Queue configuration - payment queue for disbursements (tills pay, B2B)
	QueuePaymentName string `envDefault:"jenga.payments.dequeue"       env:"QUEUE_JENGA_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://jenga.payments.dequeue" env:"QUEUE_JENGA_PAYMENT_URI"`

	// Queue configuration - prompt queue for STK/USSD push
	QueuePromptName string `envDefault:"jenga.prompts.dequeue"       env:"QUEUE_JENGA_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://jenga.prompts.dequeue" env:"QUEUE_JENGA_PROMPT_URI"`
}
