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

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

// Clues are the quick-repeat hints stored under the "checkout" key of a
// profile's properties payload (Stripe Link-style memory).
//
// Profiles typically have multiple contacts (email + phones). After a successful
// payment we record which contact(s) were preferred so the next checkout can
// preselect them without re-entry.
type Clues struct {
	LastMethod    string `json:"lastMethod"`    // registry key or category
	LastProvider  string `json:"lastProvider"`  // preferred registry key (e.g. mpesa)
	LastContactID string `json:"lastContactId"` // preferred payment contact (usually phone)
	// Preferred email / phone contact ids when profile has several of each.
	LastEmailContactID string `json:"lastEmailContactId,omitempty"`
	LastPhoneContactID string `json:"lastPhoneContactId,omitempty"`
	LastCurrency       string `json:"lastCurrency"`
	LastCountry        string `json:"lastCountry"` // ISO 3166-1 alpha-2 when known
	LastPaidAt         string `json:"lastPaidAt"`
	// Tokenized instrument for one-click card / subscription renewals.
	PaymentMethodID    string `json:"paymentMethodId,omitempty"`
	ProviderCustomerID string `json:"providerCustomerId,omitempty"`
}

// CluesFromProperties extracts checkout clues from profile properties.
func CluesFromProperties(props *structpb.Struct) Clues {
	if props == nil {
		return Clues{}
	}
	checkout := props.GetFields()["checkout"].GetStructValue()
	if checkout == nil {
		return Clues{}
	}
	f := checkout.GetFields()
	c := Clues{
		LastMethod:         f["lastMethod"].GetStringValue(),
		LastProvider:       f["lastProvider"].GetStringValue(),
		LastContactID:      f["lastContactId"].GetStringValue(),
		LastEmailContactID: f["lastEmailContactId"].GetStringValue(),
		LastPhoneContactID: f["lastPhoneContactId"].GetStringValue(),
		LastCurrency:       f["lastCurrency"].GetStringValue(),
		LastCountry:        f["lastCountry"].GetStringValue(),
		LastPaidAt:         f["lastPaidAt"].GetStringValue(),
		PaymentMethodID:    f["paymentMethodId"].GetStringValue(),
		ProviderCustomerID: f["providerCustomerId"].GetStringValue(),
	}
	// Back-compat: older clues only stored lastContactId — treat as phone prefer.
	if c.LastPhoneContactID == "" && c.LastContactID != "" {
		c.LastPhoneContactID = c.LastContactID
	}
	return c
}

// ToProperties renders the clues as a properties patch for profile Update.
func (c Clues) ToProperties() *structpb.Struct {
	// Prefer explicit phone prefer id for lastContactId (method preselect).
	phonePrefer := c.LastPhoneContactID
	if phonePrefer == "" {
		phonePrefer = c.LastContactID
	}
	checkout := map[string]any{
		"lastMethod":    c.LastMethod,
		"lastProvider":  c.LastProvider,
		"lastContactId": phonePrefer,
		"lastCurrency":  c.LastCurrency,
		"lastPaidAt":    c.LastPaidAt,
	}
	if c.LastCountry != "" {
		checkout["lastCountry"] = c.LastCountry
	}
	if c.LastEmailContactID != "" {
		checkout["lastEmailContactId"] = c.LastEmailContactID
	}
	if phonePrefer != "" {
		checkout["lastPhoneContactId"] = phonePrefer
	}
	if c.PaymentMethodID != "" {
		checkout["paymentMethodId"] = c.PaymentMethodID
	}
	if c.ProviderCustomerID != "" {
		checkout["providerCustomerId"] = c.ProviderCustomerID
	}
	props, _ := structpb.NewStruct(map[string]any{"checkout": checkout})
	return props
}

// GuestHints are device-local hints for unauthenticated payers (Link device memory).
type GuestHints struct {
	Phone   string `json:"phone"`
	Method  string `json:"method"`
	Country string `json:"country,omitempty"`
}

const guestHintsVersion = "v1"

// EncodeGuestHints signs hints into a cookie value: v1.<b64(json)>.<hmac-hex>.
func EncodeGuestHints(secret []byte, hints GuestHints) string {
	payload, _ := json.Marshal(hints)
	body := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(body))
	return guestHintsVersion + "." + body + "." + hex.EncodeToString(mac.Sum(nil))
}

// DecodeGuestHints verifies and decodes a cookie value.
func DecodeGuestHints(secret []byte, value string) (GuestHints, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 || parts[0] != guestHintsVersion {
		return GuestHints{}, false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[1]))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return GuestHints{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return GuestHints{}, false
	}
	var hints GuestHints
	if err = json.Unmarshal(payload, &hints); err != nil {
		return GuestHints{}, false
	}
	return hints, true
}
