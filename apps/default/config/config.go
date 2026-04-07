package config

import (
	"github.com/pitabwire/frame/config"
)

type PaymentConfig struct {
	config.ConfigurationDefault
	ProfileServiceURI                   string `envDefault:"127.0.0.1:7005"                 env:"PROFILE_SERVICE_URI"`
	TenancyServiceURI                   string `envDefault:"127.0.0.1:7003"                 env:"TENANCY_SERVICE_URI"`
	LedgerServiceURI                    string `envDefault:"127.0.0.1:7004"                 env:"LEDGER_SERVICE_URI"`
	ProfileServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-profile" env:"PROFILE_SERVICE_WORKLOAD_API_TARGET_PATH"`
	TenancyServiceWorkloadAPITargetPath string `envDefault:"/ns/auth/sa/service-tenancy"    env:"TENANCY_SERVICE_WORKLOAD_API_TARGET_PATH"`
	LedgerServiceWorkloadAPITargetPath  string `envDefault:"/ns/payments/sa/service-ledger" env:"LEDGER_SERVICE_WORKLOAD_API_TARGET_PATH"`

	SecurelyRunService      bool   `envDefault:"true"                      env:"SECURELY_RUN_SERVICE"`
	InitiatePromptTopicName string `envDefault:"initiate_prompt"           env:"INITIATE_PROMPT_TOPIC_NAME" required:"true"`
	InitiatePromptTopicURI  string `envDefault:"mem://initiate_prompt"     env:"INITIATE_PROMPT_TOPIC_URI"  required:"true"`
	PaymentLinkTopicName    string `envDefault:"create_payment_link"       env:"PAYMENT_LINK_TOPIC_NAME"    required:"true"`
	PaymentLinkTopicURI     string `envDefault:"mem://create_payment_link" env:"PAYMENT_LINK_TOPIC_URI"     required:"true"`
}
