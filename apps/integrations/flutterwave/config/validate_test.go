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

package config_test

import (
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/stretchr/testify/require"
)

func production() config.FlutterwaveConfig {
	return config.FlutterwaveConfig{
		ClientID:           "7f0c1d2e-client",
		ClientSecret:       "oauth-secret",
		WebhookSecret:      "whsec",
		Environment:        "production",
		DefaultRedirectURL: "https://pay.stawi.org/return",
		QueuePromptURI:     "nats://nats:4222/flutterwave.prompts",
		QueuePaymentURI:    "nats://nats:4222/flutterwave.payments",
	}
}

func TestProductionConfiguration(t *testing.T) {
	ok := production()
	require.NoError(t, ok.Validate())

	for name, mutate := range map[string]func(*config.FlutterwaveConfig){
		"no webhook secret":   func(c *config.FlutterwaveConfig) { c.WebhookSecret = "" },
		"v3 keys":             func(c *config.FlutterwaveConfig) { c.ClientSecret = "FLWSECK-abc" },
		"v3 secret alongside": func(c *config.FlutterwaveConfig) { c.SecretKey = "FLWSECK-abc" },
		"in-memory queue":     func(c *config.FlutterwaveConfig) { c.QueuePromptURI = "mem://flutterwave.prompts" },
		"localhost redirect":  func(c *config.FlutterwaveConfig) { c.DefaultRedirectURL = "http://localhost:8080/return" },
		"no client id":        func(c *config.FlutterwaveConfig) { c.ClientID = "" },
	} {
		t.Run(name, func(t *testing.T) {
			c := production()
			mutate(&c)
			require.Error(t, c.Validate())
		})
	}

	sandbox := config.FlutterwaveConfig{Environment: "sandbox"}
	require.NoError(t, sandbox.Validate(), "a sandbox may be incomplete")
	typo := config.FlutterwaveConfig{Environment: "prod"}
	require.Error(t, typo.Validate())
}
