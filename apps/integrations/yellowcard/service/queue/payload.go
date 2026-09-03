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

package queue

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	providerName = "yellowcard"

	defaultReason = "other"

	nanosPerUnit = 1e9

	minMSISDNDigits = 7
)

// Failure codes emitted by the workers for problems detected before the
// provider is reached.
const (
	failureCredentials        = "CREDENTIALS"
	failureInvalidCountry     = "INVALID_COUNTRY"
	failureChannelUnavailable = "CHANNEL_UNAVAILABLE"
	failureInvalidNetwork     = "INVALID_NETWORK"
	failureProviderError      = "PROVIDER_ERROR"
	failureInvalidAmount      = "INVALID_AMOUNT"
)

var errUnknownCountry = errors.New("could not determine the destination country")

// corridor is a country and its local currency.
type corridor struct {
	Country  string
	Currency string
}

// phoneCorridors maps E.164 country prefixes to Yellow Card corridors.
// Longer prefixes are matched first so "267" (BW) wins over "26x" and "27" (ZA).
var phoneCorridors = map[string]corridor{
	"267": {"BW", "BWP"},
	"229": {"BJ", "XOF"},
	"237": {"CM", "XAF"},
	"225": {"CI", "XOF"},
	"265": {"MW", "MWK"},
	"250": {"RW", "RWF"},
	"228": {"TG", "XOF"},
	"256": {"UG", "UGX"},
	"260": {"ZM", "ZMW"},
	"234": {"NG", "NGN"},
	"255": {"TZ", "TZS"},
	"241": {"GA", "XAF"},
	"242": {"CG", "XAF"},
	"254": {"KE", "KES"},
	"221": {"SN", "XOF"},
	"223": {"ML", "XOF"},
	"226": {"BF", "XOF"},
	"233": {"GH", "GHS"},
	"27":  {"ZA", "ZAR"},
}

// countryCurrencies is the reverse view used when only a country is known.
var countryCurrencies = func() map[string]string {
	m := map[string]string{}
	for _, c := range phoneCorridors {
		m[c.Country] = c.Currency
	}
	return m
}()

var prefixesLongestFirst = func() []string {
	out := make([]string, 0, len(phoneCorridors))
	for p := range phoneCorridors {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) != len(out[j]) {
			return len(out[i]) > len(out[j])
		}
		return out[i] < out[j]
	})
	return out
}()

// normalizeMSISDN reduces a phone number to E.164 with a leading "+".
// It returns "" when fewer than 7 digits remain.
func normalizeMSISDN(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := strings.TrimLeft(b.String(), "0")
	if len(digits) < minMSISDNDigits {
		return ""
	}
	return "+" + digits
}

// corridorForPhone finds the corridor for an E.164 number.
func corridorForPhone(msisdn string) (corridor, bool) {
	digits := strings.TrimPrefix(msisdn, "+")
	for _, p := range prefixesLongestFirst {
		if strings.HasPrefix(digits, p) {
			return phoneCorridors[p], true
		}
	}
	return corridor{}, false
}

// resolveCorridor determines the country and local currency for a payment.
// Precedence: explicit extras, then the phone prefix, then credential
// defaults. The currency additionally honours the payment's own currency.
func resolveCorridor(
	phone string,
	extra *structpb.Struct,
	paymentCurrency string,
	creds *client.Credentials,
) (corridor, error) {
	out := corridor{
		Country:  strings.ToUpper(extraString(extra, collection.ExtraCountry, collection.ExtraCustomerCountry)),
		Currency: strings.ToUpper(extraString(extra, collection.ExtraCurrency)),
	}
	if out.Currency == "" {
		out.Currency = strings.ToUpper(paymentCurrency)
	}

	if out.Country == "" {
		if c, ok := corridorForPhone(phone); ok {
			out.Country = c.Country
			if out.Currency == "" {
				out.Currency = c.Currency
			}
		}
	}
	if out.Country == "" && creds != nil {
		out.Country = strings.ToUpper(creds.Country)
	}
	if out.Country == "" {
		return corridor{}, errUnknownCountry
	}
	if out.Currency == "" {
		out.Currency = countryCurrencies[out.Country]
	}
	if out.Currency == "" && creds != nil {
		out.Currency = strings.ToUpper(creds.Currency)
	}
	return out, nil
}

// moneyToLocalAmount rounds a protobuf Money to whole currency units, which
// is what Yellow Card's integer localAmount field expects.
func moneyToLocalAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
}) int64 {
	if amount == nil {
		return 0
	}
	total := float64(amount.GetUnits()) + float64(amount.GetNanos())/nanosPerUnit
	return int64(math.Round(total))
}

// buildParty assembles the KYC party from portable extras. Institutions use
// the partner's own business identity; retail parties carry whatever the
// product knows about the customer (Yellow Card applies reduced KYC when
// only name, country and email are present).
func buildParty(extra *structpb.Struct, phone, country string, creds *client.Credentials) client.Party {
	if creds != nil && strings.EqualFold(creds.CustomerType, client.CustomerTypeInstitution) {
		return client.Party{
			BusinessID:   creds.BusinessID,
			BusinessName: creds.BusinessName,
			Email:        extraString(extra, collection.ExtraCustomerEmail, collection.ExtraEmail),
		}
	}
	return client.Party{
		Name:               extraString(extra, collection.ExtraCustomerName, collection.ExtraDisplayName),
		Email:              extraString(extra, collection.ExtraCustomerEmail, collection.ExtraEmail),
		Country:            country,
		Phone:              phone,
		Address:            extraString(extra, collection.ExtraCustomerAddress),
		DOB:                extraString(extra, collection.ExtraCustomerDOB),
		IDType:             extraString(extra, collection.ExtraCustomerIDType),
		IDNumber:           extraString(extra, collection.ExtraCustomerIDNumber),
		AdditionalIDType:   extraString(extra, collection.ExtraCustomerAdditionalIDType),
		AdditionalIDNumber: extraString(extra, collection.ExtraCustomerAdditionalIDNumber),
	}
}

// customerUID returns a stable identifier for the paying customer so Yellow
// Card can track reduced-KYC lifetime limits per person.
func customerUID(headers map[string]string, extra *structpb.Struct, profileID string) string {
	if v := extraString(extra, collection.ExtraCustomerID); v != "" {
		return v
	}
	if profileID != "" {
		return profileID
	}
	if v := headers["tenant_id"]; v != "" {
		return v
	}
	return "anonymous"
}

// resolveChannelType maps portable extras to a Yellow Card channel type.
// Empty means "choose automatically".
func resolveChannelType(extra *structpb.Struct, creds *client.Credentials) string {
	switch strings.ToLower(extraString(extra, collection.ExtraChannelType)) {
	case client.ChannelTypeMomo:
		return client.ChannelTypeMomo
	case client.ChannelTypeBank:
		return client.ChannelTypeBank
	}
	switch strings.ToLower(extraString(extra, collection.ExtraPaymentMethodType)) {
	case collection.PaymentMethodTypeMobileMoney, "momo", "mobile_money_momo":
		return client.ChannelTypeMomo
	case collection.PaymentMethodTypeBankTransfer, "bank":
		return client.ChannelTypeBank
	}
	if creds != nil {
		switch strings.ToLower(creds.ChannelType) {
		case client.ChannelTypeMomo, client.ChannelTypeBank:
			return strings.ToLower(creds.ChannelType)
		}
	}
	return ""
}

// formatLocalAmount renders "UGX 5,000".
func formatLocalAmount(amount int64, currency string) string {
	s := strconv.FormatInt(amount, 10)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	out := b.String()
	if neg {
		out = "-" + out
	}
	if currency == "" {
		return out
	}
	return currency + " " + out
}

func momoInstruction(amount int64, currency, network string) string {
	via := ""
	if network != "" {
		via = " (" + network + ")"
	}
	return fmt.Sprintf(
		"Approve the %s payment request sent to your phone%s. Enter your mobile money PIN to complete.",
		formatLocalAmount(amount, currency), via,
	)
}

func bankInstruction(info *client.BankInfo, amount int64, currency, reference, expiresAt string) string {
	if info == nil {
		return fmt.Sprintf("Transfer %s to the bank account shown to complete this payment.",
			formatLocalAmount(amount, currency))
	}
	if info.PaymentLink != "" && info.AccountNumber == "" {
		return fmt.Sprintf("Complete the %s payment through your bank using the secure link.",
			formatLocalAmount(amount, currency))
	}
	msg := fmt.Sprintf("Transfer exactly %s to %s account %s", formatLocalAmount(amount, currency), info.Name, info.AccountNumber)
	if info.AccountName != "" {
		msg += " (" + info.AccountName + ")"
	}
	if reference != "" {
		msg += " using reference " + reference
	}
	msg += "."
	if expiresAt != "" {
		msg += " The account details expire at " + expiresAt + "."
	}
	return msg
}

// failureMessages gives operators and payers a readable reason for each
// Yellow Card transaction error code.
var failureMessages = map[string]string{
	client.TxErrExpired:             "The payment was not completed before it expired",
	client.TxErrInvalidRecipient:    "The recipient details failed validation",
	client.TxErrValidationFailed:    "The transaction failed validation",
	client.TxErrInvalidNetwork:      "The selected mobile money or bank network is invalid",
	client.TxErrInvalidCurrency:     "The currency is not supported for this country",
	client.TxErrInsufficientBalance: "Insufficient balance to complete the transaction",
	client.TxErrRefused:             "The customer did not approve the payment",
	client.TxErrGatewayTimeout:      "The payment provider timed out",
	client.TxErrProviderError:       "The payment provider returned an error",
	client.TxErrPossibleDuplicate:   "The provider flagged the transaction as a possible duplicate",
	client.TxErrNameMismatch:        "The payer name does not match the customer on record",
	client.TxErrFraudCheck:          "The transaction was declined by risk checks",
	client.TxErrOtherError:          "The transaction failed",
}

// failureExtras builds failure_code / failure_message extras.
func failureExtras(code, message string) map[string]any {
	if code == "" {
		code = client.TxErrOtherError
	}
	if message == "" {
		message = failureMessages[code]
	}
	if message == "" {
		message = failureMessages[client.TxErrOtherError]
	}
	return map[string]any{
		"failure_code":    code,
		"failure_message": message,
	}
}

// providerFailureExtras converts a client error into failure extras.
func providerFailureExtras(err error) map[string]any {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		return failureExtras(apiErr.Code, apiErr.Message)
	}
	return failureExtras(failureProviderError, err.Error())
}

// isTerminalFailure reports whether a Yellow Card status ends the payment unsuccessfully.
func isTerminalFailure(status string) bool {
	switch strings.ToLower(status) {
	case client.StatusFailed, client.StatusExpired, client.StatusCancelled:
		return true
	}
	return false
}

// receiveExtras builds the portable extras for an in-flight receive so the
// checkout page can instruct the payer without knowing the provider.
func receiveExtras(r *client.Receive, channelType, networkName string) map[string]any {
	localAmount := int64(math.Round(r.ConvertedAmount))
	if localAmount == 0 {
		localAmount = int64(math.Round(r.LocalAmount))
	}
	extras := map[string]any{
		"entity_type":             "prompt",
		collection.ExtraProvider:  providerName,
		"receive_id":              r.ID,
		"sequence_id":             r.SequenceID,
		"status":                  r.Status,
		"channel_type":            channelType,
		"channel_id":              r.ChannelID,
		"network_id":              r.Source.NetworkID,
		"country":                 r.Country,
		"currency":                r.Currency,
		"local_amount":            strconv.FormatInt(localAmount, 10),
		"converted_amount":        strconv.FormatFloat(r.ConvertedAmount, 'f', -1, 64),
		"rate":                    strconv.FormatFloat(r.Rate, 'f', -1, 64),
		"service_fee_local":       strconv.FormatFloat(r.ServiceFeeAmountLocal, 'f', -1, 64),
		"partner_fee_local":       strconv.FormatFloat(r.PartnerFeeAmountLocal, 'f', -1, 64),
		"expires_at":              r.ExpiresAt,
		"provider_transaction_id": r.Reference,
	}
	if networkName != "" {
		extras["network"] = networkName
	}

	if channelType == client.ChannelTypeBank {
		reference := r.Reference
		extras[collection.ExtraPaymentInstruction] = bankInstruction(r.BankInfo, localAmount, r.Currency, reference, r.ExpiresAt)
		if r.BankInfo != nil {
			if r.BankInfo.Name != "" {
				extras[collection.ExtraBankName] = r.BankInfo.Name
			}
			if r.BankInfo.AccountNumber != "" {
				extras[collection.ExtraBankAccountNumber] = r.BankInfo.AccountNumber
			}
			if r.BankInfo.AccountName != "" {
				extras[collection.ExtraBankAccountName] = r.BankInfo.AccountName
			}
			if r.BankInfo.PaymentLink != "" {
				extras[collection.ExtraCheckoutURL] = r.BankInfo.PaymentLink
				extras[collection.ExtraAuthRedirectURL] = r.BankInfo.PaymentLink
				extras[collection.ExtraNextAction] = collection.NextActionRedirectURL
				extras[collection.ExtraNextActionType] = collection.NextActionRedirectURL
			}
		}
		if reference != "" {
			extras[collection.ExtraPaymentReference] = reference
		}
		if r.ExpiresAt != "" {
			extras[collection.ExtraPaymentExpiresAt] = r.ExpiresAt
		}
		return extras
	}

	extras[collection.ExtraPaymentInstruction] = momoInstruction(localAmount, r.Currency, networkName)
	return extras
}

// extraString reads the first non-empty string field from a payment's Extra struct.
func extraString(extra *structpb.Struct, keys ...string) string {
	if extra == nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := extra.GetFields()[key]; ok {
			if s := strings.TrimSpace(v.GetStringValue()); s != "" {
				return s
			}
		}
	}
	return ""
}
