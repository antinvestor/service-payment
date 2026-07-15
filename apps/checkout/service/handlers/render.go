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

package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"io"
	"unicode"

	"github.com/antinvestor/service-payments/apps/checkout/service/web"
)

// ContactChoice represents a selectable payer contact.
type ContactChoice struct {
	ContactID string
	Masked    string
}

// MethodChoice represents a payment method option.
type MethodChoice struct {
	Key      string
	Name     string
	Selected bool
}

// PageData holds all values needed to render any checkout page.
type PageData struct {
	Lang          string
	SessionRef    string
	MerchantName  string
	Description   string
	AmountDisplay string
	Variable      bool
	Currency      string
	PayerName     string
	MaskedPhone   string
	Contacts      []ContactChoice
	Methods       []MethodChoice
	CSRF          string
	Status        string
	FailureReason string
	ReturnURL     string
	PollURL       string
	RedirectURL   string
}

// MaskMsisdn masks a phone number for display.
// Strip leading + and whitespace; count only digits.
// Fewer than 9 digits → "••••".
// Otherwise "+{first3} {digit4}•• ••{last3}".
func MaskMsisdn(msisdn string) string {
	// strip non-digit characters to count, but preserve original digits
	var digits []rune
	for _, r := range msisdn {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	if len(digits) < minMsisdnDigits {
		return "••••"
	}
	last3 := string(digits[len(digits)-3:])
	first3 := string(digits[:3])
	digit4 := string(digits[3:4])
	return "+" + first3 + " " + digit4 + "•• ••" + last3
}

// CSRFToken returns a hex-encoded HMAC-SHA256 of "csrf:"+ref using secret.
func CSRFToken(secret []byte, ref string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("csrf:" + ref))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyCSRF returns true when token equals CSRFToken(secret, ref).
func VerifyCSRF(secret []byte, ref, token string) bool {
	expected := CSRFToken(secret, ref)
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return false
	}
	return hmac.Equal(expectedBytes, tokenBytes)
}

// minMsisdnDigits is the minimum number of digits required for masking.
const minMsisdnDigits = 9

// translations holds UI strings keyed by language then key.
//
//nolint:gochecknoglobals // package-level translation table; no mutable state.
var translations = map[string]map[string]string{
	"en": {
		"pay_button":        "Pay",
		"paying_as":         "Paying as",
		"change":            "Change",
		"phone_label":       "Phone number",
		"confirm_title":     "Confirm payment",
		"confirm_hint":      "Check your phone and follow the prompt to complete payment.",
		"retry":             "Try again",
		"done_title":        "Payment successful",
		"redirecting":       "Redirecting you back…",
		"failed_title":      "Payment failed",
		"try_again":         "Try again",
		"gone_title":        "This link has expired",
		"amount_label":      "Amount",
		"too_many_attempts": "Too many attempts — please try later",
		"cooldown":          "Please wait a moment before retrying",
		"bad_method":        "Choose a payment method",
		"amount_required":   "Enter a valid amount",
		"contact_required":  "Enter your phone number",
		"redirect_hint":     "Taking you to secure payment…",
	},
	"fr": {
		"pay_button":        "Payer",
		"paying_as":         "Payer en tant que",
		"change":            "Changer",
		"phone_label":       "Numéro de téléphone",
		"confirm_title":     "Confirmer le paiement",
		"confirm_hint":      "Vérifiez votre téléphone et suivez les instructions pour finaliser le paiement.",
		"retry":             "Réessayer",
		"done_title":        "Paiement réussi",
		"redirecting":       "Vous êtes redirigé…",
		"failed_title":      "Paiement échoué",
		"try_again":         "Réessayer",
		"gone_title":        "Ce lien a expiré",
		"amount_label":      "Montant",
		"too_many_attempts": "Trop de tentatives — réessayez plus tard",
		"cooldown":          "Veuillez patienter avant de réessayer",
		"bad_method":        "Choisissez un moyen de paiement",
		"amount_required":   "Saisissez un montant valide",
		"contact_required":  "Saisissez votre numéro de téléphone",
		"redirect_hint":     "Redirection vers le paiement sécurisé…",
	},
}

// T returns the translation for key in lang (fallback: en, then the key itself).
func T(lang, key string) string {
	if m, found := translations[lang]; found {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, found := translations["en"]; found {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

// Renderer parses and executes checkout page templates.
type Renderer struct {
	tmpl   *template.Template
	secret []byte
}

// NewRenderer creates a Renderer by parsing all embedded templates with the T func.
func NewRenderer(secret []byte) (*Renderer, error) {
	funcMap := template.FuncMap{
		"t": T,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(web.Templates, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: tmpl, secret: secret}, nil
}

// Render executes the template named page+".html" writing to w.
func (r *Renderer) Render(w io.Writer, page string, data PageData) error {
	name := page + ".html"
	// Build a new template set with the funcmap so nested templates work.
	t := r.tmpl.Lookup(name)
	if t == nil {
		return &templateNotFoundError{name: name}
	}
	return t.Execute(w, data)
}

type templateNotFoundError struct{ name string }

func (e *templateNotFoundError) Error() string {
	return "template not found: " + e.name
}
