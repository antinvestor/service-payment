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

package business

import "errors"

// Business validation errors.
var (
	// Catalog errors.
	ErrCatalogIDRequired       = errors.New("catalog ID is required")
	ErrCatalogVersionRequired  = errors.New("catalog version is required")
	ErrCatalogNameRequired     = errors.New("catalog name is required")
	ErrCatalogCurrencyRequired = errors.New("catalog currency is required")
	ErrPlanIDRequired          = errors.New("plan ID is required")
	ErrPlanNameRequired        = errors.New("plan name is required")
	ErrComponentIDRequired     = errors.New("component ID is required")
	ErrComponentNameRequired   = errors.New("component name is required")
	ErrMetricKeyRequired       = errors.New("metric key is required")

	// Subscription errors.
	ErrSubscriptionIDRequired       = errors.New("subscription ID is required")
	ErrSubscriptionProfileRequired  = errors.New("subscription profile ID is required")
	ErrSubscriptionPlanRequired     = errors.New("subscription plan ID is required")
	ErrSubscriptionCurrencyRequired = errors.New("subscription currency is required")

	// Usage errors.
	ErrUsageEventIDRequired        = errors.New("usage event ID is required")
	ErrUsageSubscriptionIDRequired = errors.New("usage subscription ID is required")
	ErrUsageMetricKeyRequired      = errors.New("usage metric key is required")
	ErrUsageQuantityRequired       = errors.New("usage quantity is required")

	// Invoice errors.
	ErrInvoiceIDRequired   = errors.New("invoice ID is required")
	ErrInvoiceNoRatedLines = errors.New("cannot generate invoice with no rated lines")

	// Credit errors.
	ErrCreditGrantIDRequired   = errors.New("credit grant ID is required")
	ErrCreditProfileIDRequired = errors.New("credit profile ID is required")
	ErrCreditAmountRequired    = errors.New("credit amount is required and must be positive")
	ErrCreditCurrencyRequired  = errors.New("credit currency is required")

	// Discount errors.
	ErrDiscountNameRequired         = errors.New("discount name is required")
	ErrDiscountTypeInvalid          = errors.New("discount type must be PERCENTAGE or FIXED")
	ErrDiscountValueRequired        = errors.New("discount value is required")
	ErrDiscountPercentageOutOfRange = errors.New("percentage discount value must be between 0 and 100")

	// Billing run errors.
	ErrBillingRunIDRequired = errors.New("billing run ID is required")

	// Polar subscription errors.
	ErrPlanNotPolarCollected          = errors.New("plan does not have a polar product ID configured")
	ErrPolarSubscriptionIDRequired    = errors.New("polar subscription ID is required")
	ErrPolarSubscriptionNotFound      = errors.New("no billing subscription found for the given polar subscription")
	ErrPolarMirrorProfileRequired     = errors.New("profile ID is required for polar subscription mirror")
	ErrPolarMirrorPlanRequired        = errors.New("plan ID is required for polar subscription mirror")
	ErrPolarMirrorPolarStatusRequired = errors.New("polar status is required for polar subscription mirror")
)
