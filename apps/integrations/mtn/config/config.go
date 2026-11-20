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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pitabwire/frame/v3/config"
	"github.com/pitabwire/util"
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
	// HeaderTargetEnvironment overrides the X-Target-Environment market code
	// (mtnuganda, mtnghana, …) MTN routes production calls by.
	HeaderTargetEnvironment = "X-MTN_TARGET_ENVIRONMENT"
)

const (
	// EnvironmentProduction selects MTN's live API.
	EnvironmentProduction = "production"
	// EnvironmentSandbox selects MTN's sandbox API.
	EnvironmentSandbox = "sandbox"
	// SandboxCurrency is the only currency MTN's sandbox accepts.
	SandboxCurrency = "EUR"
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

	// TargetEnvironment is the X-Target-Environment market code MTN expects
	// in production (mtnuganda, mtnghana, mtnzambia, …). It defaults to
	// "sandbox" in the sandbox environment and is required in production.
	TargetEnvironment string `env:"MTN_TARGET_ENVIRONMENT"`

	// Queue configuration - payment queue for disbursements
	QueuePaymentName string `envDefault:"mtn.payments.dequeue"       env:"QUEUE_MTN_PAYMENT_NAME"`
	QueuePaymentURI  string `envDefault:"mem://mtn.payments.dequeue" env:"QUEUE_MTN_PAYMENT_URI"`

	// Queue configuration - prompt queue for requestToPay
	QueuePromptName string `envDefault:"mtn.prompts.dequeue"       env:"QUEUE_MTN_PROMPT_NAME"`
	QueuePromptURI  string `envDefault:"mem://mtn.prompts.dequeue" env:"QUEUE_MTN_PROMPT_URI"`
}

// BaseURL returns the MTN MoMo API base URL based on environment.
func (c *MtnConfig) BaseURL() string {
	if c.Environment == EnvironmentProduction {
		return "https://proxy.momoapi.mtn.com"
	}
	return "https://sandbox.momodeveloper.mtn.com"
}

// ResolveTargetEnvironment returns the X-Target-Environment value for an
// environment: the explicit value, else "sandbox" outside production. In
// production a market code is mandatory, since MTN rejects (or misroutes)
// calls sent with "production" as the target.
func ResolveTargetEnvironment(environment, target string) (string, error) {
	target = strings.TrimSpace(target)
	if target != "" {
		return target, nil
	}
	if environment != EnvironmentProduction {
		return EnvironmentSandbox, nil
	}
	return "", errors.New("MTN target environment (market code such as mtnuganda) is required in production")
}

// Validate checks the startup configuration and warns about sandbox-only
// defaults left in place for production.
func (c *MtnConfig) Validate(ctx context.Context) error {
	if _, err := ResolveTargetEnvironment(c.Environment, c.TargetEnvironment); err != nil {
		return fmt.Errorf("MTN_TARGET_ENVIRONMENT: %w", err)
	}
	if c.Environment == EnvironmentProduction && strings.EqualFold(c.Currency, SandboxCurrency) {
		util.Log(ctx).WithField("currency", c.Currency).
			Warn("MTN_CURRENCY is the sandbox default EUR while MTN_ENVIRONMENT is production; set the market currency")
	}
	return nil
}
