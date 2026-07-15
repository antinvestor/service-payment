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
	// InitiatePromptRouteURIs is optional JSON map of route key → NATS/mem URI.
	// When set, InitiatePrompt publishes to the matching route publisher instead of
	// (or in addition to configuring) the single default INITIATE_PROMPT_TOPIC_URI.
	// Example: {"mpesa":"nats://…mpesa.prompts","flutterwave":"nats://…flutterwave.prompts"}
	// Publisher reference becomes "prompt.<route>". Unlisted routes fall back to default.
	InitiatePromptRouteURIs string `env:"INITIATE_PROMPT_ROUTE_URIS"`
	PaymentLinkTopicName    string `envDefault:"create_payment_link"       env:"PAYMENT_LINK_TOPIC_NAME"    required:"true"`
	PaymentLinkTopicURI     string `envDefault:"mem://create_payment_link" env:"PAYMENT_LINK_TOPIC_URI"     required:"true"`
}
