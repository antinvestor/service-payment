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

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Fincra statuses. Collections report success/successful, payouts
// successful; pending/processing/initiated are in flight.
const (
	StatusInitiated  = "initiated"
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusSuccess    = "success"
	StatusSuccessful = "successful"
	StatusFailed     = "failed"
	StatusExpired    = "expired"
)

// Variance types Fincra reports when the amount received differs from the
// amount expected. An underpayment still reads as status success.
const (
	VarianceUnderpayment = "underpayment"
	VarianceOverpayment  = "overpayment"
)

// Charge types (direct charge "type", checkout "paymentMethods").
const (
	ChargeTypeMobileMoney  = "mobile_money"
	ChargeTypeBankTransfer = "bank_transfer"
	ChargeTypeCard         = "card"
)

// Auth models a pending direct charge may require.
const (
	AuthModelOTP = "OTP"
)

// Payout destinations.
const (
	PaymentDestinationMobileMoney = "mobile_money_wallet"
	PaymentDestinationBankAccount = "bank_account"
)

// Beneficiary types.
const (
	BeneficiaryIndividual = "individual"
)

// Fincra error types (errorType) the adapter acts on.
const (
	ErrTypeDuplicateReference = "DUPLICATE_CUSTOMER_REFERENCE"
	ErrTypeNotFound           = "RESOURCE_NOT_FOUND"
	ErrTypeInsufficientFunds  = "NO_ENOUGH_MONEY_IN_WALLET"
	ErrTypeValidationFailed   = "VALIDATION_FAILED"
	ErrTypeAccessDenied       = "ACCESS_DENIED"
)

// FlexString decodes a JSON string or number (Fincra ids are numeric on
// some endpoints and strings on others) and treats null as empty.
type FlexString string

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*f = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = FlexString(s)
		return nil
	}
	*f = FlexString(string(b))
	return nil
}

// String returns the value as a string.
func (f FlexString) String() string { return string(f) }

// FlexFloat decodes a JSON number or numeric string (refund amounts arrive
// as "5914.000000") and treats null or an unparsable string as zero.
type FlexFloat float64

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexFloat) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			*f = 0
			return nil //nolint:nilerr // a non-numeric amount string is treated as unknown
		}
		*f = FlexFloat(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = FlexFloat(v)
	return nil
}

// Float returns the value as float64.
func (f FlexFloat) Float() float64 { return float64(f) }

// Customer is the payer on a collection.
type Customer struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// --- hosted checkout ---

// CheckoutRequest creates a hosted checkout link (POST /checkout/payments).
type CheckoutRequest struct {
	Amount                int64             `json:"amount"`
	Currency              string            `json:"currency"`
	Customer              Customer          `json:"customer"`
	Reference             string            `json:"reference"`
	FeeBearer             string            `json:"feeBearer"`
	RedirectURL           string            `json:"redirectUrl,omitempty"`
	PaymentMethods        []string          `json:"paymentMethods,omitempty"`
	DefaultPaymentMethod  string            `json:"defaultPaymentMethod,omitempty"`
	SettlementDestination string            `json:"settlementDestination,omitempty"`
	Metadata              map[string]string `json:"metadata,omitempty"`
}

// CheckoutLink is the created hosted checkout.
type CheckoutLink struct {
	Link      string `json:"link"`
	Reference string `json:"reference"`
	PayCode   string `json:"payCode"`
}

// CheckoutPayment is a hosted checkout as returned by verify
// (GET /checkout/payments/merchant-reference/{reference}).
type CheckoutPayment struct {
	ID                FlexString     `json:"id"`
	Amount            FlexFloat      `json:"amount"`
	AmountExpected    FlexFloat      `json:"amountExpected"`
	AmountReceived    FlexFloat      `json:"amountReceived"`
	Currency          string         `json:"currency"`
	Reference         string         `json:"reference"` // Fincra pay code (fcr-p-...)
	MerchantReference string         `json:"merchantReference"`
	Status            string         `json:"status"`
	VarianceType      string         `json:"varianceType"`
	Message           string         `json:"message"`
	ActionRequired    string         `json:"actionRequired"`
	FeeBearer         string         `json:"feeBearer"`
	Metadata          map[string]any `json:"metadata"`
	Customer          Customer       `json:"customer"`
	UpdatedAt         string         `json:"updatedAt"`
}

// --- direct charge ---

// ChargeRequest initiates a direct charge (POST /checkout/charges).
type ChargeRequest struct {
	Type                  string   `json:"type"`
	Amount                int64    `json:"amount"`
	Reference             string   `json:"reference"`
	Currency              string   `json:"currency"`
	Customer              Customer `json:"customer"`
	Phone                 string   `json:"phone,omitempty"`
	Operator              string   `json:"operator,omitempty"`
	Description           string   `json:"description,omitempty"`
	SettlementDestination string   `json:"settlementDestination,omitempty"`
	CallbackURL           string   `json:"callbackUrl,omitempty"`
	WebhookURL            string   `json:"webhookUrl,omitempty"`
}

// ChargeAuthorization is a pending charge's authorization requirement.
type ChargeAuthorization struct {
	Mode     string `json:"mode"`
	Redirect string `json:"redirect"`
}

// VirtualAccount is the account a bank transfer charge must be paid into.
type VirtualAccount struct {
	ID            string `json:"id"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	BankCode      string `json:"bankCode"`
	BankName      string `json:"bankName"`
	ExpiresAt     string `json:"expiresAt"`
}

// Charge is a direct charge as returned by initiate, authorize and verify.
type Charge struct {
	ID              FlexString          `json:"id"`
	Authorization   ChargeAuthorization `json:"authorization"`
	AuthModel       string              `json:"auth_model"`
	Amount          FlexFloat           `json:"amount"`
	AmountExpected  FlexFloat           `json:"amountExpected"`
	AmountReceived  FlexFloat           `json:"amountReceived"`
	VarianceType    string              `json:"varianceType"`
	Currency        string              `json:"currency"`
	Fee             FlexFloat           `json:"fee"`
	Message         string              `json:"message"`
	ActionRequired  string              `json:"actionRequired"`
	Status          string              `json:"status"`
	Reference       string              `json:"reference"`
	ChargeReference string              `json:"chargeReference"`
	Type            string              `json:"type"`
	Customer        Customer            `json:"customer"`
	Metadata        map[string]any      `json:"metadata"`
	VirtualAccount  *VirtualAccount     `json:"virtualAccount,omitempty"`
}

// AuthorizeRequest completes a pending charge (POST /checkout/charges/{id}/authorize).
type AuthorizeRequest struct {
	OTP string `json:"otp,omitempty"`
}

// --- payouts ---

// Beneficiary is a payout's recipient.
type Beneficiary struct {
	FirstName         string `json:"firstName,omitempty"`
	LastName          string `json:"lastName,omitempty"`
	AccountHolderName string `json:"accountHolderName"`
	Type              string `json:"type"`
	Country           string `json:"country"`
	AccountNumber     string `json:"accountNumber"`
	MobileMoneyCode   string `json:"mobileMoneyCode,omitempty"`
	Email             string `json:"email,omitempty"`
}

// PayoutRequest creates a payout (POST /disbursements/payouts).
type PayoutRequest struct {
	Business            string      `json:"business"`
	SourceCurrency      string      `json:"sourceCurrency"`
	DestinationCurrency string      `json:"destinationCurrency"`
	Amount              json.Number `json:"amount"`
	Description         string      `json:"description,omitempty"`
	PaymentDestination  string      `json:"paymentDestination"`
	CustomerReference   string      `json:"customerReference"`
	Beneficiary         Beneficiary `json:"beneficiary"`
}

// PayoutRecipient is the recipient echoed on a fetched payout.
type PayoutRecipient struct {
	AccountNumber     string `json:"accountNumber"`
	AccountHolderName string `json:"accountHolderName"`
	Name              string `json:"name"`
}

// Payout is a payout as returned by create and fetch.
type Payout struct {
	ID                  FlexString      `json:"id"`
	AmountSent          FlexFloat       `json:"amountSent"`
	AmountReceived      FlexFloat       `json:"amountReceived"`
	SourceCurrency      string          `json:"sourceCurrency"`
	DestinationCurrency string          `json:"destinationCurrency"`
	Fee                 FlexFloat       `json:"fee"`
	Status              string          `json:"status"`
	Reference           string          `json:"reference"`
	CustomerReference   string          `json:"customerReference"`
	Message             string          `json:"message"`
	Reason              string          `json:"reason"`
	PaymentDestination  string          `json:"paymentDestination"`
	TraceID             string          `json:"traceId"`
	Recipient           PayoutRecipient `json:"recipient"`
}

// --- catalog ---

// Provider is a bank or mobile money operator from GET /core/banks.
type Provider struct {
	ID   FlexString `json:"id"`
	Code string     `json:"code"`
	Name string     `json:"name"`
}

// envelope is Fincra's response wrapper; collection endpoints answer
// {"status": true}, payout and core endpoints {"success": true}.
type envelope[T any] struct {
	Status  *bool  `json:"status"`
	Success *bool  `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}
