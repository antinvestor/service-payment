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

package client

// Credentials holds Flutterwave credentials for a tenant/request.
//
// Two modes are supported:
//   - v4 OAuth: ClientID + ClientSecret (UUID-style, from "Switch to v4 API keys")
//   - v3 secret key: SecretKey (FLWSECK_*) + optional PublicKey (FLWPUBK_*) + EncryptionKey
//
// Dashboard test keys often look like FLWPUBK_TEST-… / FLWSECK_TEST-… — those are v3.
// The client auto-detects via IsV3Credentials.
type Credentials struct {
	// v4 OAuth
	ClientID     string
	ClientSecret string
	// v3 classic
	PublicKey     string
	SecretKey     string
	EncryptionKey string
	// Shared
	WebhookSecret string
	Environment   string // sandbox | production
	// Optional overrides
	APIBaseURL    string
	OAuthTokenURL string
}

// --- Customers ---

type CustomerName struct {
	First  string `json:"first,omitempty"`
	Middle string `json:"middle,omitempty"`
	Last   string `json:"last,omitempty"`
}

type CustomerPhone struct {
	CountryCode string `json:"country_code"`
	Number      string `json:"number"`
}

type CustomerAddress struct {
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	State      string `json:"state,omitempty"`
}

// CreateCustomerRequest is POST /customers.
type CreateCustomerRequest struct {
	Email   string            `json:"email"`
	Name    *CustomerName     `json:"name,omitempty"`
	Phone   *CustomerPhone    `json:"phone,omitempty"`
	Address *CustomerAddress  `json:"address,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// --- Payment methods ---

type MobileMoneyDetails struct {
	CountryCode string `json:"country_code"`
	Network     string `json:"network"`
	PhoneNumber string `json:"phone_number"`
}

type BankTransferDetails struct {
	AccountType        string `json:"account_type"` // static | dynamic
	AccountExpiresIn   int    `json:"account_expires_in,omitempty"`
	AccountDisplayName string `json:"account_display_name,omitempty"`
}

type USSDDetails struct {
	AccountBank string `json:"account_bank,omitempty"`
}

// CardCOF enables Credential-on-File when Flutterwave has approved the feature.
type CardCOF struct {
	Enabled     bool   `json:"enabled"`
	AgreementID string `json:"agreement_id,omitempty"`
}

// CardDetails holds AES-256-GCM encrypted card fields for v4.
// Docs: https://developer.flutterwave.com/docs/encryption
// Docs: https://developer.flutterwave.com/docs/card
type CardDetails struct {
	EncryptedCardNumber  string   `json:"encrypted_card_number"`
	EncryptedExpiryMonth string   `json:"encrypted_expiry_month"`
	EncryptedExpiryYear  string   `json:"encrypted_expiry_year"`
	EncryptedCVV         string   `json:"encrypted_cvv"`
	Nonce                string   `json:"nonce"` // exactly 12 alphanumeric chars
	COF                  *CardCOF `json:"cof,omitempty"`
}

// PaymentMethodInput is embedded in orchestrator or POST /payment-methods.
type PaymentMethodInput struct {
	Type         string               `json:"type"` // mobile_money | bank_transfer | opay | ussd | card
	MobileMoney  *MobileMoneyDetails  `json:"mobile_money,omitempty"`
	BankTransfer *BankTransferDetails `json:"bank_transfer,omitempty"`
	USSD         *USSDDetails         `json:"ussd,omitempty"`
	Card         *CardDetails         `json:"card,omitempty"`
	CustomerID   string               `json:"customer_id,omitempty"`
	Meta         map[string]string    `json:"meta,omitempty"`
}

// PaymentMethod is the response object from POST /payment-methods.
type PaymentMethod struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Card      map[string]any `json:"card,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	CreatedAt string         `json:"created_datetime"`
}

// Customer is the response from POST /customers.
type Customer struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	Name      *CustomerName  `json:"name,omitempty"`
	Phone     *CustomerPhone `json:"phone,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	CreatedAt string         `json:"created_datetime"`
}

// ChargeRequest is POST /charges (general flow with existing customer + method).
type ChargeRequest struct {
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	Reference       string            `json:"reference"`
	CustomerID      string            `json:"customer_id"`
	PaymentMethodID string            `json:"payment_method_id"`
	RedirectURL     string            `json:"redirect_url,omitempty"`
	Recurring       bool              `json:"recurring,omitempty"`
	Meta            map[string]string `json:"meta,omitempty"`
}

// ChargeAuthorization updates a pending charge (PIN / OTP / AVS).
// Docs: https://developer.flutterwave.com/docs/card#authorising-payments-auth-models
type ChargeAuthorization struct {
	Type string   `json:"type"` // pin | otp | avs
	PIN  *PINAuth `json:"pin,omitempty"`
	OTP  *OTPAuth `json:"otp,omitempty"`
	AVS  *AVSAuth `json:"avs,omitempty"`
}

// PINAuth is encrypted card PIN authorization.
type PINAuth struct {
	Nonce        string `json:"nonce"`
	EncryptedPIN string `json:"encrypted_pin"`
}

// OTPAuth is OTP authorization.
type OTPAuth struct {
	Code string `json:"code"`
}

// AVSAuth is address verification authorization.
type AVSAuth struct {
	Address *CustomerAddress `json:"address"`
}

// UpdateChargeRequest is PUT /charges/{id}.
type UpdateChargeRequest struct {
	Authorization ChargeAuthorization `json:"authorization"`
}

// --- Orchestrator charge (collections) ---

// OrchestratorChargeRequest is POST /orchestration/direct-charges.
// Creates customer + payment method + charge in one call.
// Docs: https://developer.flutterwave.com/docs/payment-orchestrator-flow
type OrchestratorChargeRequest struct {
	Amount        float64            `json:"amount"`
	Currency      string             `json:"currency"`
	Reference     string             `json:"reference"`
	RedirectURL   string             `json:"redirect_url,omitempty"`
	Customer      CustomerInput      `json:"customer"`
	PaymentMethod PaymentMethodInput `json:"payment_method"`
	Meta          map[string]string  `json:"meta,omitempty"`
}

// CustomerInput is the nested customer on orchestrator requests.
type CustomerInput struct {
	Email   string            `json:"email"`
	Name    *CustomerName     `json:"name,omitempty"`
	Phone   *CustomerPhone    `json:"phone,omitempty"`
	Address *CustomerAddress  `json:"address,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// Charge is the data object returned from charges / orchestrator.
type Charge struct {
	ID            string         `json:"id"`
	Amount        float64        `json:"amount"`
	Currency      string         `json:"currency"`
	CustomerID    string         `json:"customer_id"`
	Reference     string         `json:"reference"`
	Status        string         `json:"status"` // succeeded | pending | failed | voided
	RedirectURL   string         `json:"redirect_url"`
	NextAction    map[string]any `json:"next_action"`
	Meta          map[string]any `json:"meta"`
	PaymentMethod map[string]any `json:"payment_method_details"`
	CreatedAt     string         `json:"created_datetime"`
	ProcessorResp map[string]any `json:"processor_response"`
}

// APIEnvelope is the standard {status,message,data} wrapper.
type APIEnvelope[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Error   *struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// --- Transfers ---

// BankRecipientDetails is bank payout destination.
type BankRecipientDetails struct {
	AccountNumber string `json:"account_number"`
	Code          string `json:"code"`
}

// MobileMoneyRecipientDetails is MoMo payout destination.
type MobileMoneyRecipientDetails struct {
	Network string `json:"network"`
	Country string `json:"country"`
	MSISDN  string `json:"msisdn"`
}

// TransferRecipientRequest is POST /transfers/recipients.
type TransferRecipientRequest struct {
	Type        string                       `json:"type"` // bank_ngn, bank_kes, mobile_money, …
	Bank        *BankRecipientDetails        `json:"bank,omitempty"`
	MobileMoney *MobileMoneyRecipientDetails `json:"mobile_money,omitempty"`
	Name        *CustomerName                `json:"name,omitempty"`
}

// DirectTransferRequest is POST /direct-transfers (orchestrator payout).
type DirectTransferRequest struct {
	Action             string             `json:"action"` // instant
	Type               string             `json:"type"`   // bank | mobile_money
	Reference          string             `json:"reference"`
	Narration          string             `json:"narration,omitempty"`
	CallbackURL        string             `json:"callback_url,omitempty"`
	Meta               map[string]string  `json:"meta,omitempty"`
	PaymentInstruction PaymentInstruction `json:"payment_instruction"`
}

// PaymentInstruction for transfers.
type PaymentInstruction struct {
	SourceCurrency      string         `json:"source_currency"`
	DestinationCurrency string         `json:"destination_currency"`
	Amount              TransferAmount `json:"amount"`
	Recipient           any            `json:"recipient,omitempty"` // inline recipient for orchestrator
	RecipientID         string         `json:"recipient_id,omitempty"`
}

// TransferAmount is the v4 amount object.
type TransferAmount struct {
	Value     float64 `json:"value"`
	AppliesTo string  `json:"applies_to"` // destination_currency | source_currency
}

// Transfer is the transfer data object.
type Transfer struct {
	ID                  string         `json:"id"`
	Type                string         `json:"type"`
	Reference           string         `json:"reference"`
	Status              string         `json:"status"` // NEW | PENDING | SUCCESSFUL | FAILED | …
	SourceCurrency      string         `json:"source_currency"`
	DestinationCurrency string         `json:"destination_currency"`
	Amount              map[string]any `json:"amount"`
	Meta                map[string]any `json:"meta"`
	CreatedAt           string         `json:"created_datetime"`
}

// --- OAuth ---

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// --- Webhooks ---

// WebhookEvent is the v4 webhook body (also accepts classic v3 shape).
// v4: type / webhook_id / data
// v3: event / data  (event ≈ charge.completed | transfer.completed)
// https://developer.flutterwave.com/docs/webhooks
type WebhookEvent struct {
	WebhookID string         `json:"webhook_id"`
	Timestamp int64          `json:"timestamp"`
	Type      string         `json:"type"`  // charge.completed | transfer.disburse | …
	Event     string         `json:"event"` // v3 alias for Type
	Data      map[string]any `json:"data"`
}

// EventType returns the webhook event name from v4 or v3 fields.
func (e *WebhookEvent) EventType() string {
	if e == nil {
		return ""
	}
	if e.Type != "" {
		return e.Type
	}
	return e.Event
}
