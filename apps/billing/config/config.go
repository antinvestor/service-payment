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

type BillingConfig struct {
	config.ConfigurationDefault

	CheckoutServiceURI                   string `envDefault:"127.0.0.1:7010"                           env:"CHECKOUT_SERVICE_URI"`
	CheckoutServiceWorkloadAPITargetPath string `envDefault:"/ns/payments/sa/service-payment-checkout" env:"CHECKOUT_SERVICE_WORKLOAD_API_TARGET_PATH"`
	CheckoutInvoiceReturnURL             string `                                                      env:"CHECKOUT_INVOICE_RETURN_URL"`
}
