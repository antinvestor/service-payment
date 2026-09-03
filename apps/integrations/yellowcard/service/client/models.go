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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Receive and send statuses reported by Yellow Card.
const (
	StatusCreated          = "created"
	StatusPendingApproval  = "pending_approval"
	StatusProcess          = "process"
	StatusProcessing       = "processing"
	StatusPendingLiquidity = "pending_liquidity"
	StatusPending          = "pending"
	StatusComplete         = "complete"
	StatusFailed           = "failed"
	StatusExpired          = "expired"
	StatusCancelled        = "cancelled"
	StatusPendingRefund    = "pending_refund"
	StatusRefundProcessing = "refund_processing"
	StatusRefundFailed     = "refund_failed"
	StatusRefunded         = "refunded"
)

// Channel types, ramp types and account types.
const (
	ChannelTypeMomo = "momo"
	ChannelTypeBank = "bank"

	RampTypeDeposit  = "deposit"
	RampTypeWithdraw = "withdraw"

	ChannelStatusActive = "active"
)

// KYC customer types.
const (
	CustomerTypeRetail      = "retail"
	CustomerTypeInstitution = "institution"
)

// API error codes documented by Yellow Card.
const (
	ErrCodeAuthentication     = "AuthenticationError"
	ErrCodePermission         = "PermissionError"
	ErrCodePaymentValidation  = "PaymentValidationError"
	ErrCodePaymentExpired     = "PaymentExpired"
	ErrCodePaymentInvalidSt   = "PaymentInvalidState"
	ErrCodePaymentNotFound    = "PaymentNotFound"
	ErrCodeCollectionNotFound = "CollectionNotFoundError"
	ErrCodeInternalServer     = "InternalServerError"
)

// Transaction error codes found in the errorCode field of receives and sends.
const (
	TxErrExpired             = "EXPIRED"
	TxErrInvalidRecipient    = "INVALID_RECIPIENT"
	TxErrValidationFailed    = "VALIDATION_FAILED"
	TxErrInvalidNetwork      = "INVALID_NETWORK"
	TxErrInvalidCurrency     = "INVALID_CURRENCY"
	TxErrInsufficientBalance = "INSUFFICIENT_BALANCE"
	TxErrRefused             = "REFUSED"
	TxErrGatewayTimeout      = "GATEWAY_TIMEOUT"
	TxErrProviderError       = "PROVIDER_ERROR"
	TxErrPossibleDuplicate   = "POSSIBLE_DUPLICATE"
	TxErrNameMismatch        = "NAME_MISMATCH"
	TxErrOtherError          = "OTHER_ERROR"
	TxErrFraudCheck          = "FRAUD_CHECK"
)

const (
	sandboxBaseURL    = "https://sandbox.api.yellowcard.io/business"
	productionBaseURL = "https://api.yellowcard.io/business"
)

// Credentials holds the per-request credentials for Yellow Card API calls.
// These can be resolved from queue message headers, the settings service or
// fall back to configuration.
type Credentials struct {
	// APIKey identifies the partner API key.
	APIKey string
	// SecretKey signs requests and, by default, verifies webhooks.
	SecretKey string
	// Environment selects the API host: sandbox or production.
	Environment string
	// BaseURL overrides the environment-derived API host when set (used in tests).
	BaseURL string
	// Country is an optional default ISO 3166-1 alpha-2 country code.
	Country string
	// Currency is an optional default ISO 4217 currency code.
	Currency string
	// Network is an optional default network id, code or name.
	Network string
	// ChannelType optionally forces momo or bank.
	ChannelType string
	// CustomerType is retail (default) or institution.
	CustomerType string
	// BusinessID and BusinessName identify the partner when CustomerType is institution.
	BusinessID   string
	BusinessName string
	// WebhookSecret optionally overrides SecretKey for webhook verification.
	WebhookSecret string
}

// ResolveBaseURL returns the Yellow Card API base URL for these credentials.
func (c *Credentials) ResolveBaseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	if strings.EqualFold(c.Environment, "production") {
		return productionBaseURL
	}
	return sandboxBaseURL
}

// ResolveWebhookSecret returns the secret used to verify webhook signatures.
func (c *Credentials) ResolveWebhookSecret() string {
	if c.WebhookSecret != "" {
		return c.WebhookSecret
	}
	return c.SecretKey
}

// APIError is a structured error response from the Yellow Card API.
type APIError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("yellowcard api error %d %s: %s", e.HTTPStatus, e.Code, e.Message)
}

// IsNotFound reports whether the error is a Yellow Card not-found response.
func (e *APIError) IsNotFound() bool {
	return e.HTTPStatus == 404 || //nolint:mnd // http status
		e.Code == ErrCodePaymentNotFound || e.Code == ErrCodeCollectionNotFound
}

// IsDuplicate reports whether the error indicates the sequenceId was already
// used, in which case the existing record should be looked up instead.
func (e *APIError) IsDuplicate() bool {
	if e.HTTPStatus != 400 && e.HTTPStatus != 409 { //nolint:mnd // http status
		return false
	}
	msg := strings.ToLower(e.Message)
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "sequenceid") ||
		strings.Contains(msg, "sequence id") ||
		strings.Contains(msg, "already exist")
}

// IsNotFound reports whether err is an API not-found error.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsNotFound()
}

// IsDuplicate reports whether err is an API duplicate-sequence error.
func IsDuplicate(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsDuplicate()
}

// Party carries the KYC metadata of a receive recipient or a send sender.
type Party struct {
	Name               string `json:"name,omitempty"`
	Country            string `json:"country,omitempty"`
	Address            string `json:"address,omitempty"`
	DOB                string `json:"dob,omitempty"`
	Email              string `json:"email,omitempty"`
	IDNumber           string `json:"idNumber,omitempty"`
	IDType             string `json:"idType,omitempty"`
	AdditionalIDType   string `json:"additionalIdType,omitempty"`
	AdditionalIDNumber string `json:"additionalIdNumber,omitempty"`
	Phone              string `json:"phone,omitempty"`
	BusinessID         string `json:"businessId,omitempty"`
	BusinessName       string `json:"businessName,omitempty"`
}

// Source describes where a receive is collected from.
type Source struct {
	AccountType   string `json:"accountType"`
	AccountNumber string `json:"accountNumber,omitempty"`
	NetworkID     string `json:"networkId,omitempty"`
	AccountName   string `json:"accountName,omitempty"`
	AccountBank   string `json:"accountBank,omitempty"`
}

// Destination describes where a send is delivered to.
type Destination struct {
	AccountNumber string `json:"accountNumber,omitempty"`
	AccountType   string `json:"accountType"`
	NetworkID     string `json:"networkId,omitempty"`
	AccountName   string `json:"accountName,omitempty"`
	AccountBank   string `json:"accountBank,omitempty"`
	NetworkName   string `json:"networkName,omitempty"`
	PhoneNumber   string `json:"phoneNumber,omitempty"`
	Country       string `json:"country,omitempty"`
	Branch        string `json:"branch,omitempty"`
	BranchCode    string `json:"branchCode,omitempty"`
}

// BankInfo carries the bank details (or payment link) a customer must pay
// into for bank-channel receives. Unknown keys are retained in Extra.
type BankInfo struct {
	Name          string
	AccountNumber string
	AccountName   string
	PaymentLink   string
	Extra         map[string]any
}

// UnmarshalJSON accepts the documented fields and detects a payment link
// under any of the keys Yellow Card uses for redirect channels.
func (b *BankInfo) UnmarshalJSON(data []byte) error {
	raw := map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	b.Extra = map[string]any{}
	for k, v := range raw {
		s, _ := v.(string)
		switch k {
		case "name", "bankName":
			b.Name = s
		case "accountNumber":
			b.AccountNumber = s
		case "accountName":
			b.AccountName = s
		case "paymentLink", "paymentUrl", "url", "link", "redirectUrl", "checkoutUrl":
			if s != "" {
				b.PaymentLink = s
			}
		default:
			b.Extra[k] = v
		}
	}
	return nil
}

// MarshalJSON emits the canonical fields (used by tests and debugging).
func (b BankInfo) MarshalJSON() ([]byte, error) {
	out := map[string]any{}
	for k, v := range b.Extra {
		out[k] = v
	}
	if b.Name != "" {
		out["name"] = b.Name
	}
	if b.AccountNumber != "" {
		out["accountNumber"] = b.AccountNumber
	}
	if b.AccountName != "" {
		out["accountName"] = b.AccountName
	}
	if b.PaymentLink != "" {
		out["paymentLink"] = b.PaymentLink
	}
	return json.Marshal(out)
}

// ReceiveRequest submits a receive (collection) request.
type ReceiveRequest struct {
	ChannelID    string `json:"channelId,omitempty"`
	ChannelType  string `json:"channelType,omitempty"`
	SequenceID   string `json:"sequenceId"`
	LocalAmount  int64  `json:"localAmount,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	Country      string `json:"country,omitempty"`
	Currency     string `json:"currency,omitempty"`
	Source       Source `json:"source"`
	Recipient    Party  `json:"recipient"`
	CustomerType string `json:"customerType"`
	CustomerUID  string `json:"customerUID"`
	RedirectURL  string `json:"redirectUrl,omitempty"`
	Reason       string `json:"reason,omitempty"`
	ForceAccept  bool   `json:"forceAccept"`
}

// Receive is a receive record as returned by submit, accept and lookup.
type Receive struct {
	ID                    string    `json:"id"`
	SequenceID            string    `json:"sequenceId"`
	Status                string    `json:"status"`
	ChannelID             string    `json:"channelId"`
	Country               string    `json:"country"`
	Currency              string    `json:"currency"`
	Amount                float64   `json:"amount"`
	LocalAmount           float64   `json:"localAmount"`
	ConvertedAmount       float64   `json:"convertedAmount"`
	Rate                  float64   `json:"rate"`
	ServiceFeeAmountLocal float64   `json:"serviceFeeAmountLocal"`
	ServiceFeeAmountUSD   float64   `json:"serviceFeeAmountUSD"`
	PartnerFeeAmountLocal float64   `json:"partnerFeeAmountLocal"`
	PartnerFeeAmountUSD   float64   `json:"partnerFeeAmountUSD"`
	Source                Source    `json:"source"`
	Recipient             Party     `json:"recipient"`
	BankInfo              *BankInfo `json:"bankInfo,omitempty"`
	Reference             string    `json:"reference"`
	DepositID             string    `json:"depositId"`
	SessionID             string    `json:"sessionId"`
	ErrorCode             string    `json:"errorCode"`
	ExpiresAt             string    `json:"expiresAt"`
	CreatedAt             string    `json:"createdAt"`
	UpdatedAt             string    `json:"updatedAt"`
}

// SendRequest submits a send (disbursement) request.
type SendRequest struct {
	ChannelID    string      `json:"channelId,omitempty"`
	ChannelType  string      `json:"channelType,omitempty"`
	SequenceID   string      `json:"sequenceId"`
	LocalAmount  int64       `json:"localAmount,omitempty"`
	Amount       int64       `json:"amount,omitempty"`
	Country      string      `json:"country,omitempty"`
	Currency     string      `json:"currency,omitempty"`
	Reason       string      `json:"reason"`
	Sender       Party       `json:"sender"`
	Destination  Destination `json:"destination"`
	CustomerType string      `json:"customerType"`
	CustomerUID  string      `json:"customerUID"`
	ForceAccept  bool        `json:"forceAccept"`
}

// Send is a send record as returned by submit, accept and lookup.
type Send struct {
	ID              string      `json:"id"`
	SequenceID      string      `json:"sequenceId"`
	Status          string      `json:"status"`
	ChannelID       string      `json:"channelId"`
	Country         string      `json:"country"`
	Currency        string      `json:"currency"`
	Amount          float64     `json:"amount"`
	LocalAmount     float64     `json:"localAmount"`
	ConvertedAmount float64     `json:"convertedAmount"`
	Rate            float64     `json:"rate"`
	Reason          string      `json:"reason"`
	Sender          Party       `json:"sender"`
	Destination     Destination `json:"destination"`
	Reference       string      `json:"reference"`
	SessionID       string      `json:"sessionId"`
	ErrorCode       string      `json:"errorCode"`
	ExpiresAt       string      `json:"expiresAt"`
	CreatedAt       string      `json:"createdAt"`
	UpdatedAt       string      `json:"updatedAt"`
}

// Channel is a payment rail for one country, currency, channel type and ramp type.
type Channel struct {
	ID                      string  `json:"id"`
	Country                 string  `json:"country"`
	Currency                string  `json:"currency"`
	CountryCurrency         string  `json:"countryCurrency"`
	Status                  string  `json:"status"`
	APIStatus               string  `json:"apiStatus"`
	WidgetStatus            string  `json:"widgetStatus"`
	ChannelType             string  `json:"channelType"`
	RampType                string  `json:"rampType"`
	SettlementType          string  `json:"settlementType"`
	VendorID                string  `json:"vendorId"`
	Min                     float64 `json:"min"`
	Max                     float64 `json:"max"`
	FeeLocal                float64 `json:"feeLocal"`
	FeeUSD                  float64 `json:"feeUSD"`
	EstimatedSettlementTime int     `json:"estimatedSettlementTime"`
}

// Network is a bank or mobile money operator a customer interfaces with.
type Network struct {
	ID                       string   `json:"id"`
	Code                     string   `json:"code"`
	Name                     string   `json:"name"`
	Country                  string   `json:"country"`
	Status                   string   `json:"status"`
	AccountNumberType        string   `json:"accountNumberType"`
	CountryAccountNumberType string   `json:"countryAccountNumberType"`
	ChannelIDs               []string `json:"channelIds"`
}

// Rate is the partner buy/sell rate for one currency against USD.
type Rate struct {
	Code      string  `json:"code"`
	Locale    string  `json:"locale"`
	RateID    string  `json:"rateId"`
	Buy       float64 `json:"buy"`
	Sell      float64 `json:"sell"`
	UpdatedAt string  `json:"updatedAt"`
}

type ratesResponse struct {
	Rates []Rate `json:"rates"`
}
