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
	"fmt"
	"net/url"
	"strings"
)

// IsProduction reports whether the service talks to live Flutterwave.
func (c *FlutterwaveConfig) IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(c.Environment), "production")
}

func isClassicKey(v string) bool {
	return strings.HasPrefix(v, "FLWSECK") || strings.HasPrefix(v, "FLWPUBK")
}

// Validate refuses a configuration that would lose or misreport money in
// production. A sandbox is allowed to be incomplete.
func (c *FlutterwaveConfig) Validate() error {
	env := strings.ToLower(strings.TrimSpace(c.Environment))
	if env != "sandbox" && env != "production" {
		return fmt.Errorf("FLUTTERWAVE_ENVIRONMENT must be sandbox or production, got %q", c.Environment)
	}
	if !c.IsProduction() {
		return nil
	}

	var problems []string
	if c.ClientID == "" || c.ClientSecret == "" || isClassicKey(c.ClientID) || isClassicKey(c.ClientSecret) {
		problems = append(problems,
			"FLUTTERWAVE_CLIENT_ID and FLUTTERWAVE_CLIENT_SECRET must be v4 OAuth credentials (v3 FLWSECK keys are sandbox only)")
	}
	if c.SecretKey != "" {
		problems = append(problems, "FLUTTERWAVE_SECRET_KEY (v3) must not be set in production")
	}
	if strings.TrimSpace(c.WebhookSecret) == "" {
		problems = append(problems, "FLUTTERWAVE_WEBHOOK_SECRET is required: without it no webhook can be authenticated")
	}
	for name, uri := range map[string]string{
		"QUEUE_FLUTTERWAVE_PROMPT_URI":  c.QueuePromptURI,
		"QUEUE_FLUTTERWAVE_PAYMENT_URI": c.QueuePaymentURI,
	} {
		if uri == "" || strings.HasPrefix(uri, "mem://") {
			problems = append(problems, name+" must be a shared queue (not mem://) so the payment service's messages arrive")
		}
	}
	if u, err := url.Parse(c.DefaultRedirectURL); err != nil || u.Scheme != "https" ||
		u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
		problems = append(problems, "FLUTTERWAVE_DEFAULT_REDIRECT_URL must be a public https URL")
	}
	if len(problems) > 0 {
		return errors.New("flutterwave production configuration: " + strings.Join(problems, "; "))
	}
	return nil
}
