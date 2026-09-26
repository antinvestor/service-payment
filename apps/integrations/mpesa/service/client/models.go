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
	"encoding/base64"
	"encoding/json"
	"time"
)

// MpesaCredentials holds the per-request credentials for M-Pesa API calls.
// These can be resolved from queue message headers or fall back to config.
type MpesaCredentials struct {
	ConsumerKey        string
	ConsumerSecret     string
	Shortcode          string
	Passkey            string
	CallbackURL        string
	InitiatorName      string
	SecurityCredential string
	Environment        string // sandbox or production
}

// BaseURL returns the Daraja API base URL based on environment.
func (c *MpesaCredentials) BaseURL() string {
	if c.Environment == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

// TokenResponse represents the OAuth token response from Daraja API.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// STKPushRequest represents the STK Push (Lipa Na M-Pesa) request.
type STKPushRequest struct {
	BusinessShortCode string
	Password          string
	Timestamp         string
	TransactionType   string
	Amount            string
	PartyA            string // Customer phone number
	PartyB            string // Business shortcode
	PhoneNumber       string
	CallBackURL       string
	AccountReference  string
	TransactionDesc   string
}

// stkPushPayload is the JSON payload sent to the Daraja API.
type stkPushPayload struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

// STKPushResponse represents the STK Push response.
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

// B2CRequest represents a B2C payment request.
type B2CRequest struct {
	OriginatorConversationID string
	InitiatorName            string
	SecurityCredential       string
	CommandID                string // BusinessPayment, SalaryPayment, PromotionPayment
	Amount                   string
	PartyA                   string // Business shortcode
	PartyB                   string // Customer phone number
	Remarks                  string
	QueueTimeOutURL          string
	ResultURL                string
	Occasion                 string
}

// b2cPayload is the JSON payload sent to the Daraja API.
type b2cPayload struct {
	OriginatorConversationID string `json:"OriginatorConversationID"`
	InitiatorName            string `json:"InitiatorName"`
	SecurityCredential       string `json:"SecurityCredential"`
	CommandID                string `json:"CommandID"`
	Amount                   string `json:"Amount"`
	PartyA                   string `json:"PartyA"`
	PartyB                   string `json:"PartyB"`
	Remarks                  string `json:"Remarks"`
	QueueTimeOutURL          string `json:"QueueTimeOutURL"`
	ResultURL                string `json:"ResultURL"`
	Occasion                 string `json:"Occasion"`
}

// B2CResponse represents a B2C payment response.
type B2CResponse struct {
	ConversationID           string `json:"ConversationID"`
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ResponseCode             string `json:"ResponseCode"`
	ResponseDescription      string `json:"ResponseDescription"`
}

// STKCallbackBody represents the STK Push callback payload.
type STKCallbackBody struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  *struct {
				Item []CallbackItem `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// CallbackItem represents a key-value item in callback metadata.
type CallbackItem struct {
	Name  string `json:"Name"`
	Value any    `json:"Value"`
}

// C2BValidationRequest represents a C2B validation callback.
type C2BValidationRequest struct {
	TransactionType   string `json:"TransType"`
	TransID           string `json:"TransID"`
	TransTime         string `json:"TransTime"`
	TransAmount       string `json:"TransAmount"`
	BusinessShortCode string `json:"BusinessShortCode"`
	BillRefNumber     string `json:"BillRefNumber"`
	InvoiceNumber     string `json:"InvoiceNumber"`
	OrgAccountBalance string `json:"OrgAccountBalance"`
	ThirdPartyTransID string `json:"ThirdPartyTransID"`
	MSISDN            string `json:"MSISDN"`
	FirstName         string `json:"FirstName"`
	MiddleName        string `json:"MiddleName"`
	LastName          string `json:"LastName"`
}

// B2CCallbackBody represents the B2C result callback payload.
type B2CCallbackBody struct {
	Result struct {
		ResultType               int    `json:"ResultType"`
		ResultCode               int    `json:"ResultCode"`
		ResultDesc               string `json:"ResultDesc"`
		OriginatorConversationID string `json:"OriginatorConversationID"`
		ConversationID           string `json:"ConversationID"`
		TransactionID            string `json:"TransactionID"`
		ResultParameters         *struct {
			ResultParameter []CallbackItem `json:"ResultParameter"`
		} `json:"ResultParameters"`
	} `json:"Result"`
}

// Keys shared between the STK push worker and the STK callback handler.
const (
	// CallbackParamPromptID is the callback URL query parameter carrying the
	// payment-service prompt id the STK push was issued for.
	CallbackParamPromptID = "prompt_id"
	// ExtraCheckoutRequestID binds a prompt status to its Daraja CheckoutRequestID.
	ExtraCheckoutRequestID = "checkout_request_id"
	// ExtraRequestedAmount is the amount sent to Daraja in the STK push.
	ExtraRequestedAmount = "requested_amount"
	// ExtraCredentialsConnection is the settings key the STK push credentials
	// were resolved from (a key name, never a secret).
	ExtraCredentialsConnection = "credentials_connection"
)

// STKPassword builds the Lipa Na M-Pesa password: base64(shortcode+passkey+timestamp).
func STKPassword(shortcode, passkey, timestamp string) string {
	return base64.StdEncoding.EncodeToString([]byte(shortcode + passkey + timestamp))
}

// STKQueryRequest is the Daraja STK Push Query payload.
type STKQueryRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	CheckoutRequestID string `json:"CheckoutRequestID"`
}

// NewSTKQueryRequest builds a query for checkoutRequestID signed with creds at now.
func NewSTKQueryRequest(creds *MpesaCredentials, checkoutRequestID string, now time.Time) *STKQueryRequest {
	timestamp := now.Format("20060102150405")
	return &STKQueryRequest{
		BusinessShortCode: creds.Shortcode,
		Password:          STKPassword(creds.Shortcode, creds.Passkey, timestamp),
		Timestamp:         timestamp,
		CheckoutRequestID: checkoutRequestID,
	}
}

// STKQueryResponse is the Daraja STK Push Query response. ResultCode is "0"
// only when the customer completed the payment.
type STKQueryResponse struct {
	ResponseCode        FlexString `json:"ResponseCode"`
	ResponseDescription string     `json:"ResponseDescription"`
	MerchantRequestID   string     `json:"MerchantRequestID"`
	CheckoutRequestID   string     `json:"CheckoutRequestID"`
	ResultCode          FlexString `json:"ResultCode"`
	ResultDesc          string     `json:"ResultDesc"`
}

// FlexString decodes a JSON string or number into a string (Daraja returns
// result codes as either depending on the endpoint and environment).
type FlexString string

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = FlexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = FlexString(n.String())
	return nil
}
