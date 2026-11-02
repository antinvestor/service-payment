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

// Package collection defines portable collection extras and concepts so
// product apps (checkout, billing, opportunities) stay provider-agnostic.
//
// Switching PSPs means changing the payment Route and the integration worker
// that understands these extras — not the checkout UI or session model.
package collection

// Portable prompt Extra keys used across checkout → payment → provider adapters.
// Providers must accept these names (map as needed inside the adapter only).
const (
	// Identity (reuse profile data; only ask when missing).
	ExtraCustomerEmail = "customer_email"
	ExtraCustomerName  = "customer_name"
	ExtraEmail         = "email"
	ExtraDisplayName   = "display_name"

	// Return URLs (our domain — never the PSP multipay homepage as the product return).
	ExtraSuccessURL  = "success_url"
	ExtraRedirectURL = "redirect_url"

	// Correlation.
	ExtraSessionRef     = "session_ref"
	ExtraInvoiceID      = "invoice_id"
	ExtraSubscriptionID = "subscription_id"
	ExtraTxRef          = "tx_ref"

	// Method selection (adapter maps to PSP-specific payment_method.type).
	ExtraPaymentMethodType = "payment_method_type"
	// ExtraChannelType selects the provider rail directly: "momo" or "bank".
	ExtraChannelType = "channel_type"
	// ExtraNetwork names the mobile money operator or bank (id, code or name).
	ExtraNetwork = "network"
	// ExtraCountry / ExtraCurrency override corridor detection from the MSISDN.
	ExtraCountry  = "country"
	ExtraCurrency = "currency"
	// ExtraReason is the regulatory payment reason (gift, bills, ... other).
	ExtraReason = "reason"

	// Customer KYC metadata (only what the product already holds; providers
	// that need full KYC read these and fall back to reduced KYC when absent).
	ExtraCustomerCountry            = "customer_country"
	ExtraCustomerAddress            = "customer_address"
	ExtraCustomerDOB                = "customer_dob" // mm/dd/yyyy
	ExtraCustomerIDType             = "customer_id_type"
	ExtraCustomerIDNumber           = "customer_id_number"
	ExtraCustomerAdditionalIDType   = "customer_additional_id_type"
	ExtraCustomerAdditionalIDNumber = "customer_additional_id_number"

	// Embedded card (AES-GCM ciphertext + nonce) — never clear PAN in extras long-term logs.
	ExtraEncryptedCardNumber  = "encrypted_card_number"
	ExtraEncryptedExpiryMonth = "encrypted_expiry_month"
	ExtraEncryptedExpiryYear  = "encrypted_expiry_year"
	ExtraEncryptedCVV         = "encrypted_cvv"
	ExtraCardNonce            = "card_nonce"
	ExtraNonce                = "nonce"

	// Tokenized instrument (subscriptions + Link-style returning payers).
	ExtraPaymentMethodID = "payment_method_id"
	ExtraCustomerID      = "customer_id"
	ExtraRecurring       = "recurring"

	// Authorization continuation (PIN / OTP / AVS).
	ExtraAction            = "action" // "authorize"
	ExtraAuthorizationType = "authorization_type"
	ExtraChargeID          = "charge_id"
	ExtraEncryptedPIN      = "encrypted_pin"
	ExtraPIN               = "pin"
	ExtraOTP               = "otp"

	// Status extras returned by adapters (checkout consumes these).
	ExtraCheckoutURL        = "checkout_url" // 3DS or legacy hosted
	ExtraAuthRedirectURL    = "auth_redirect_url"
	ExtraNextAction         = "next_action"
	ExtraNextActionType     = "next_action_type"
	ExtraPaymentInstruction = "payment_instruction"
	ExtraProvider           = "provider"
	ExtraAPIVersion         = "api_version"
	ExtraMode               = "mode"

	// Bank transfer instructions returned by adapters when the payer must
	// push money to an account (checkout renders these on the confirm page).
	ExtraBankName          = "bank_name"
	ExtraBankAccountNumber = "bank_account_number"
	ExtraBankAccountName   = "bank_account_name"
	ExtraPaymentReference  = "payment_reference"
	ExtraPaymentExpiresAt  = "payment_expires_at" // RFC3339
)

// Payment method types understood by adapters (ExtraPaymentMethodType).
const (
	PaymentMethodTypeCard         = "card"
	PaymentMethodTypeMobileMoney  = "mobile_money"
	PaymentMethodTypeBankTransfer = "bank_transfer"
)

// ActionAuthorize is the portable action value for charge authorization prompts.
const ActionAuthorize = "authorize"

// NextAction types (normalized across providers).
const (
	NextActionRedirectURL  = "redirect_url"
	NextActionRequiresPIN  = "requires_pin"
	NextActionRequiresOTP  = "requires_otp"
	NextActionRequiresAVS  = "requires_additional_fields"
	NextActionBankTransfer = "requires_bank_transfer"
)

// DefaultCardRoute is the payment service route name for card collection.
// Override per tenant/partition via checkout method registry Route field.
const DefaultCardRoute = "flutterwave"
