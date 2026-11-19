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
	"net/http"
	"strings"
)

// AirtelCredentials holds per-request credentials for Airtel Money API.
type AirtelCredentials struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	CountryCode  string
	Currency     string
	Environment  string
	// DisbursementPin is the wallet PIN disbursements are authorised with;
	// it is sent RSA-encrypted with PublicKey (see EncryptPin).
	DisbursementPin string
	PublicKey       string

	baseURL string
}

// SetBaseURL points the credentials at another API host (tests).
func (c *AirtelCredentials) SetBaseURL(u string) { c.baseURL = u }

// BaseURL returns the Airtel Money API base URL based on environment.
func (c *AirtelCredentials) BaseURL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	if c.Environment == "production" {
		return "https://openapi.airtel.africa"
	}
	return "https://openapiuat.airtel.africa"
}

// TokenResponse represents the OAuth token response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// APIError is an Airtel answer that is not a success: a non-2xx HTTP status,
// or a 2xx whose status.success is false.
type APIError struct {
	Operation  string
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" || e.Message != "" {
		return fmt.Sprintf("%s refused: %s - %s", e.Operation, e.Code, e.Message)
	}
	return fmt.Sprintf("%s failed with status %d: %s", e.Operation, e.StatusCode, e.Message)
}

// IsNotFound reports whether Airtel knows no transaction for the id.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// IsRejected reports whether Airtel refused the request (4xx, or a 2xx with
// success=false): nothing was applied, so the worker may fail the entity.
func IsRejected(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode < http.StatusInternalServerError
}

// CollectionRequest represents a USSD push collection request.
//
// Airtel's USSD push (POST /merchant/v2/payments/) takes no callback URL:
// the collection callback endpoint is registered on the Airtel developer
// portal per application, so none is carried here.
type CollectionRequest struct {
	Reference   string
	PhoneNumber string
	Amount      string
	Currency    string
	CountryCode string
}

// collectionPayload is the JSON payload sent to the Airtel API.
type collectionPayload struct {
	Reference   string             `json:"reference"`
	Subscriber  subscriberPayload  `json:"subscriber"`
	Transaction transactionPayload `json:"transaction"`
}

type subscriberPayload struct {
	Country  string `json:"country"`
	Currency string `json:"currency"`
	Msisdn   string `json:"msisdn"`
}

type transactionPayload struct {
	Amount   string `json:"amount"`
	Country  string `json:"country"`
	Currency string `json:"currency"`
	ID       string `json:"id"`
}

// CollectionResponse represents the API response for a collection request.
type CollectionResponse struct {
	Data struct {
		Transaction struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"transaction"`
	} `json:"data"`
	Status struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		ResultCode string `json:"result_code"`
		Success    bool   `json:"success"`
	} `json:"status"`
}

// DisbursementRequest represents a disbursement request. EncryptedPin is
// the wallet PIN RSA-encrypted with Airtel's public key (EncryptPin); the
// v2 API refuses a disbursement without it.
type DisbursementRequest struct {
	Reference    string
	PhoneNumber  string
	Amount       string
	Currency     string
	CountryCode  string
	EncryptedPin string
}

// disbursementPayload is the JSON payload sent to the Airtel API.
type disbursementPayload struct {
	Payee       payeePayload       `json:"payee"`
	Reference   string             `json:"reference"`
	Pin         string             `json:"pin"`
	Transaction transactionPayload `json:"transaction"`
}

type payeePayload struct {
	Msisdn string `json:"msisdn"`
	Name   string `json:"name,omitempty"`
}

// DisbursementResponse represents the API response for a disbursement.
type DisbursementResponse struct {
	Data struct {
		Transaction struct {
			ID          string `json:"id"`
			Status      string `json:"status"`
			ReferenceID string `json:"reference_id"`
		} `json:"transaction"`
	} `json:"data"`
	Status struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		ResultCode string `json:"result_code"`
		Success    bool   `json:"success"`
	} `json:"status"`
}

// StatusResponse represents the transaction status API response. Amount is
// not documented on every enquiry answer; when present it is checked
// against the requested amount.
type StatusResponse struct {
	Data struct {
		Transaction struct {
			AirtelMoneyID FlexID     `json:"airtel_money_id"`
			ID            string     `json:"id"`
			Message       string     `json:"message"`
			Status        string     `json:"status"`
			Amount        FlexString `json:"amount"`
		} `json:"transaction"`
	} `json:"data"`
	Status struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		ResultCode string `json:"result_code"`
		Success    bool   `json:"success"`
	} `json:"status"`
}

// CollectionCallbackBody represents the callback payload from Airtel for
// collections and disbursements. Airtel documents airtel_money_id as a
// string; FlexID also accepts the {"id": "..."} object form seen in some
// markets. Amount is optional and only checked when present.
type CollectionCallbackBody struct {
	Transaction struct {
		ID            string     `json:"id"`
		Message       string     `json:"message"`
		StatusCode    string     `json:"status_code"`
		AirtelMoneyID FlexID     `json:"airtel_money_id"`
		Amount        FlexString `json:"amount"`
	} `json:"transaction"`
}

// DisbursementCallbackBody is the disbursement callback payload (same shape
// as a collection callback).
type DisbursementCallbackBody = CollectionCallbackBody

// FlexID decodes an id Airtel sends either as a string or as {"id": "..."}.
type FlexID string

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexID) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = FlexID(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err == nil {
		*f = FlexID(n.String())
		return nil
	}
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return fmt.Errorf("airtel_money_id is neither a string nor an object: %w", err)
	}
	*f = FlexID(obj.ID)
	return nil
}

// FlexString decodes a JSON string or number into a string.
type FlexString string

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = FlexString(strings.TrimSpace(s))
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = FlexString(n.String())
	return nil
}

// Status extras shared between the queue workers and the callback handler.
const (
	// ExtraTransactionID is Airtel's transaction id returned when the
	// request was accepted; its presence shows we issued the request.
	ExtraTransactionID = "transaction_id"
	// ExtraRequestedAmount is the amount sent to Airtel.
	ExtraRequestedAmount = "requested_amount"
	// ExtraRequestedCurrency is the currency sent to Airtel.
	ExtraRequestedCurrency = "requested_currency"
	// ExtraCredentialsConnection is the settings key the credentials came
	// from (a key name, never a secret).
	ExtraCredentialsConnection = "credentials_connection"
)
