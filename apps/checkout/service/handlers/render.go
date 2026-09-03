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

// ContactChoice represents a selectable payer contact from the profile
// (email or phone). Free-text contacts are not allowed when profile contacts exist.
type ContactChoice struct {
	ContactID string
	Masked    string
	// Kind is "email" or "phone" — drives which payment methods are offered.
	Kind string
	// Preferred is true for the last successfully used payment contact.
	Preferred bool
}

// MethodChoice represents a payment method option.
type MethodChoice struct {
	Key      string
	Name     string
	Selected bool
	Embed    bool // card form on this page
	Redirect bool
}

// BankTransferDetails are the account details shown on the confirm page for
// push-payment rails (Yellow Card bank receives, Flutterwave bank transfer).
type BankTransferDetails struct {
	BankName      string
	AccountNumber string
	AccountName   string
	Reference     string
	ExpiresAt     string
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
	MaskedEmail   string
	// NeedEmail is true when profile has no email and card requires one (guests only).
	NeedEmail bool
	// NeedName is true when profile has no display name (guests only).
	NeedName bool
	// RequireProfileContacts: hide free-text email/phone; only contact chips pay.
	RequireProfileContacts bool
	// SelectedContactKind is the default selected contact kind (email|phone).
	SelectedContactKind string
	// HasSavedCard enables Link-style "pay with saved card" one-click.
	HasSavedCard bool
	// ShowCardForm when selected method is embedded card.
	ShowCardForm bool
	// CardCryptoURL is GET endpoint that returns AES key for browser encryption.
	CardCryptoURL string
	// AuthorizeURL for PIN/OTP steps.
	AuthorizeURL  string
	Contacts      []ContactChoice
	Methods       []MethodChoice
	CSRF          string
	Status        string
	FailureReason string
	ReturnURL     string
	PollURL       string
	RedirectURL   string
	// NextAction: requires_pin | requires_otp | redirect_url | …
	NextAction string
	// PaymentInstruction for bank transfer notes.
	PaymentInstruction string
	// BankTransfer holds account details the payer must transfer into
	// (portable bank_* status extras from the provider adapter).
	BankTransfer *BankTransferDetails
	// AssetVersion cache-busts /static/* URLs (see web.AssetVersion).
	AssetVersion string
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
		"pay_button":                 "Pay",
		"paying_as":                  "Paying as",
		"change":                     "Change",
		"phone_label":                "Phone number",
		"email_label":                "Email",
		"name_label":                 "Name",
		"card_number":                "Card number",
		"card_expiry":                "Expiry",
		"card_cvv":                   "CVV",
		"card_pin":                   "Card PIN",
		"card_otp":                   "One-time code",
		"card_secure_hint":           "Card details are encrypted in your browser before they leave this page.",
		"use_saved_card":             "Pay with your saved card",
		"continue":                   "Continue",
		"confirm_title":              "Confirm payment",
		"confirm_hint":               "Complete any bank prompts. You stay on this secure page whenever possible.",
		"bank_transfer_title":        "Transfer to this account",
		"bank_name":                  "Bank",
		"bank_account_number":        "Account number",
		"bank_account_name":          "Account name",
		"bank_reference":             "Reference",
		"bank_expires":               "Pay before",
		"retry":                      "Try again",
		"done_title":                 "Payment successful",
		"redirecting":                "Redirecting you back…",
		"failed_title":               "Payment failed",
		"try_again":                  "Try again",
		"gone_title":                 "This link has expired",
		"gone_hint":                  "Checkout links expire after a short time for your security. Start payment again from the app to get a fresh link.",
		"gone_return":                "Back to continue",
		"gone_return_hint":           "Return to the app and choose Continue to payment for a new secure link.",
		"amount_label":               "Amount",
		"too_many_attempts":          "Too many attempts — please try later",
		"cooldown":                   "Please wait a moment before retrying",
		"bad_method":                 "Choose a payment method",
		"amount_required":            "Enter a valid amount",
		"contact_required":           "Select a contact from your profile",
		"contact_not_on_profile":     "That contact is not on your profile",
		"method_not_allowed_contact": "That payment method is not available for this contact. Email pays by card; phone can use mobile money or card.",
		"contact_kind_email":         "Email",
		"contact_kind_phone":         "Phone",
		"select_contact":             "Pay with",
		"no_profile_contacts":        "Add an email or phone to your profile before paying.",
		"redirect_hint":              "Securing your payment…",
	},
	"fr": {
		"pay_button":                 "Payer",
		"paying_as":                  "Payer en tant que",
		"change":                     "Changer",
		"phone_label":                "Numéro de téléphone",
		"email_label":                "E-mail",
		"name_label":                 "Nom",
		"card_number":                "Numéro de carte",
		"card_expiry":                "Expiration",
		"card_cvv":                   "CVV",
		"card_pin":                   "Code PIN",
		"card_otp":                   "Code à usage unique",
		"card_secure_hint":           "Les données de carte sont chiffrées dans votre navigateur avant l’envoi.",
		"use_saved_card":             "Payer avec la carte enregistrée",
		"continue":                   "Continuer",
		"confirm_title":              "Confirmer le paiement",
		"confirm_hint":               "Suivez les instructions de votre banque. Vous restez sur cette page sécurisée autant que possible.",
		"bank_transfer_title":        "Virement vers ce compte",
		"bank_name":                  "Banque",
		"bank_account_number":        "Numéro de compte",
		"bank_account_name":          "Titulaire du compte",
		"bank_reference":             "Référence",
		"bank_expires":               "Payer avant",
		"retry":                      "Réessayer",
		"done_title":                 "Paiement réussi",
		"redirecting":                "Vous êtes redirigé…",
		"failed_title":               "Paiement échoué",
		"try_again":                  "Réessayer",
		"gone_title":                 "Ce lien a expiré",
		"gone_hint":                  "Les liens de paiement expirent rapidement pour votre sécurité. Relancez le paiement depuis l’application pour un nouveau lien.",
		"gone_return":                "Retour pour continuer",
		"gone_return_hint":           "Retournez dans l’application et choisissez Continuer vers le paiement pour un nouveau lien sécurisé.",
		"amount_label":               "Montant",
		"too_many_attempts":          "Trop de tentatives — réessayez plus tard",
		"cooldown":                   "Veuillez patienter avant de réessayer",
		"bad_method":                 "Choisissez un moyen de paiement",
		"amount_required":            "Saisissez un montant valide",
		"contact_required":           "Sélectionnez un contact de votre profil",
		"contact_not_on_profile":     "Ce contact n’est pas sur votre profil",
		"method_not_allowed_contact": "Ce moyen de paiement n’est pas disponible pour ce contact. L’e-mail paie par carte ; le téléphone accepte mobile money ou carte.",
		"contact_kind_email":         "E-mail",
		"contact_kind_phone":         "Téléphone",
		"select_contact":             "Payer avec",
		"no_profile_contacts":        "Ajoutez un e-mail ou un téléphone à votre profil avant de payer.",
		"redirect_hint":              "Sécurisation du paiement…",
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
