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

type CheckoutConfig struct {
	config.ConfigurationDefault

	PaymentServiceURI                   string `envDefault:"127.0.0.1:7006"                  env:"PAYMENT_SERVICE_URI"`
	ProfileServiceURI                   string `envDefault:"127.0.0.1:7003"                  env:"PROFILE_SERVICE_URI"`
	PaymentServiceWorkloadAPITargetPath string `envDefault:"/ns/payments/sa/service-payment" env:"PAYMENT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	ProfileServiceWorkloadAPITargetPath string `envDefault:"/ns/profile/sa/service-profile"  env:"PROFILE_SERVICE_WORKLOAD_API_TARGET_PATH"`

	// PublicBaseURL is the externally visible origin of the checkout page.
	PublicBaseURL string `envDefault:"http://localhost:8080" env:"CHECKOUT_PUBLIC_BASE_URL"`

	// SigningSecret signs CSRF tokens and the guest hint cookie.
	SigningSecret string `env:"CHECKOUT_SIGNING_SECRET"`

	SessionTTLMinutes      int `envDefault:"30" env:"CHECKOUT_SESSION_TTL_MINUTES"`
	MaxAttempts            int `envDefault:"3"  env:"CHECKOUT_MAX_ATTEMPTS"`
	AttemptCooldownSeconds int `envDefault:"20" env:"CHECKOUT_ATTEMPT_COOLDOWN_SECONDS"`
	LinkSpawnPerMinute     int `envDefault:"10" env:"CHECKOUT_LINK_SPAWN_PER_MINUTE"`
	SweepIntervalSeconds   int `envDefault:"60" env:"CHECKOUT_SWEEP_INTERVAL_SECONDS"`

	// MethodsJSON is the v1 method registry: which payment methods the page
	// renders and the prompt route hint each maps to.
	MethodsJSON string `env:"CHECKOUT_METHODS" envDefault:"[{\"key\":\"mpesa\",\"name\":\"M-PESA\",\"route\":\"mpesa\",\"prefixes\":[\"254\"],\"currencies\":[\"KES\"]},{\"key\":\"mtn_momo\",\"name\":\"MTN MoMo\",\"route\":\"mtn\",\"prefixes\":[\"256\",\"260\"],\"currencies\":[\"UGX\",\"ZMW\"]},{\"key\":\"airtel_money\",\"name\":\"Airtel Money\",\"route\":\"airtel\",\"prefixes\":[\"255\",\"256\"],\"currencies\":[\"TZS\",\"UGX\"]},{\"key\":\"pawapay\",\"name\":\"Mobile Money\",\"route\":\"pawapay\",\"prefixes\":[],\"currencies\":[]}]"`
}
