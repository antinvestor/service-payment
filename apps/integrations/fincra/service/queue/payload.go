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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/pitabwire/frame/v3/security"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	providerName = "fincra"

	nanosPerCent = 10_000_000

	minMSISDNDigits = 9
)

// Failure codes emitted by the workers for problems found before Fincra is
// called. Fincra's own errorType values are passed through as they are.
const (
	failureInvalidAmount     = "INVALID_AMOUNT"
	failureMissingPhone      = "MISSING_PHONE_NUMBER"
	failureMissingEmail      = "MISSING_CUSTOMER_EMAIL"
	failureUnknownOperator   = "UNKNOWN_OPERATOR"
	failureUnknownCountry    = "UNKNOWN_COUNTRY"
	failureUnsupportedMethod = "UNSUPPORTED_METHOD"
	failureCredentials       = "CREDENTIALS_ERROR"
	failureProviderError     = "PROVIDER_ERROR"
	failureMissingCharge     = "MISSING_CHARGE_ID"
)

var failureMessages = map[string]string{
	failureInvalidAmount:             "The amount must be a positive whole number of currency units",
	failureMissingPhone:              "A mobile number is required for mobile money",
	failureMissingEmail:              "An email address is required to collect with Fincra",
	failureUnknownOperator:           "The mobile money operator could not be determined",
	failureUnknownCountry:            "The payee's country could not be determined",
	failureUnsupportedMethod:         "The payment method is not supported by Fincra",
	failureCredentials:               "The payment provider is not configured",
	failureProviderError:             "The payment provider returned an error",
	failureMissingCharge:             "There is no pending charge to authorize",
	client.FailureAmountMismatch:     "Less than the amount due was paid; Fincra returns it to the payer",
	client.ErrTypeInsufficientFunds:  "Insufficient balance to complete the payout",
	client.ErrTypeDuplicateReference: "The payout reference was already used",
	client.ErrTypeValidationFailed:   "The payment details failed validation",
	"EXPIRED":                        "The payment was not completed before it expired",
}

// corridor is a mobile money market: its country and currency.
type corridor struct {
	Country  string
	Currency string
}

// corridorsByPrefix maps an MSISDN country calling code to Fincra's mobile
// money corridors (pay-in and payout currencies listed in the docs).
var corridorsByPrefix = []struct {
	prefix string
	corridor
}{
	{"254", corridor{"KE", "KES"}},
	{"233", corridor{"GH", "GHS"}},
	{"256", corridor{"UG", "UGX"}},
	{"260", corridor{"ZM", "ZMW"}},
	{"255", corridor{"TZ", "TZS"}},
	{"250", corridor{"RW", "RWF"}},
	{"234", corridor{"NG", "NGN"}},
	{"229", corridor{"BJ", "XOF"}},
	{"226", corridor{"BF", "XOF"}},
	{"225", corridor{"CI", "XOF"}},
	{"221", corridor{"SN", "XOF"}},
	{"237", corridor{"CM", "XAF"}},
	{"242", corridor{"CG", "XAF"}},
	{"241", corridor{"GA", "XAF"}},
	{"20", corridor{"EG", "EGP"}},
}

// corridorForPhone returns the corridor of an international MSISDN (digits only).
func corridorForPhone(digits string) (corridor, bool) {
	for _, c := range corridorsByPrefix {
		if strings.HasPrefix(digits, c.prefix) {
			return c.corridor, true
		}
	}
	return corridor{}, false
}

// networkPrefixes maps a country calling code to the mobile network of a
// national number's leading digits. Longest prefix wins. It only names the
// operator; Fincra's own provider list decides the code that is sent.
var networkPrefixes = map[string]map[string]string{
	"254": {
		"70": "MPESA", "71": "MPESA", "72": "MPESA", "740": "MPESA", "741": "MPESA", "742": "MPESA",
		"743": "MPESA", "745": "MPESA", "746": "MPESA", "748": "MPESA", "757": "MPESA", "758": "MPESA",
		"759": "MPESA", "768": "MPESA", "769": "MPESA", "79": "MPESA", "110": "MPESA", "111": "MPESA",
		"112": "MPESA", "113": "MPESA", "114": "MPESA", "115": "MPESA",
		"73": "AIRTEL", "750": "AIRTEL", "751": "AIRTEL", "752": "AIRTEL", "753": "AIRTEL", "754": "AIRTEL",
		"755": "AIRTEL", "756": "AIRTEL", "762": "AIRTEL", "78": "AIRTEL", "100": "AIRTEL", "101": "AIRTEL",
		"102": "AIRTEL",
		"763": "EQUITEL", "764": "EQUITEL", "765": "EQUITEL", "766": "EQUITEL",
	},
	"256": {"77": "MTN", "78": "MTN", "76": "MTN", "39": "MTN", "70": "AIRTEL", "75": "AIRTEL", "74": "AIRTEL", "20": "AIRTEL"},
	"233": {
		"24": "MTN", "54": "MTN", "55": "MTN", "59": "MTN", "25": "MTN", "53": "MTN",
		"20": "VODAFONE", "50": "VODAFONE",
		"26": "AIRTELTIGO", "56": "AIRTELTIGO", "27": "AIRTELTIGO", "57": "AIRTELTIGO",
	},
	"260": {"96": "MTN", "76": "MTN", "97": "AIRTEL", "77": "AIRTEL", "95": "ZAMTEL", "75": "ZAMTEL"},
	"255": {
		"74": "MPESA", "75": "MPESA", "76": "MPESA",
		"68": "AIRTEL", "69": "AIRTEL", "78": "AIRTEL",
		"65": "TIGO", "67": "TIGO", "71": "TIGO",
	},
	"250": {"78": "MTN", "79": "MTN", "72": "AIRTEL", "73": "AIRTEL"},
}

// detectNetwork names the mobile network of an international MSISDN, or "".
func detectNetwork(digits string) string {
	for cc, table := range networkPrefixes {
		if !strings.HasPrefix(digits, cc) {
			continue
		}
		national := digits[len(cc):]
		best, network := 0, ""
		for prefix, name := range table {
			if len(prefix) > best && strings.HasPrefix(national, prefix) {
				best, network = len(prefix), name
			}
		}
		return network
	}
	return ""
}

// normalizeMSISDN returns the digits of a phone number with no leading
// zeros or plus, or "" when it is too short to be an international number.
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
	return digits
}

// money is the protobuf Money shape the workers read.
type money interface {
	GetUnits() int64
	GetNanos() int32
	GetCurrencyCode() string
}

// wholeUnits returns a positive amount with no fractional part. Fincra's
// collection amount is an integer, so a fractional amount is refused rather
// than rounded into a different charge.
func wholeUnits(m money) (int64, bool) {
	if m == nil || m.GetNanos() != 0 || m.GetUnits() <= 0 {
		return 0, false
	}
	return m.GetUnits(), true
}

// payoutAmount is a positive payout amount as an exact decimal with at most
// two places. whole requires a whole number (UGX and Airtel payouts).
func payoutAmount(m money, whole bool) (json.Number, bool) {
	if m == nil || m.GetUnits() < 0 || m.GetNanos() < 0 || (m.GetUnits() == 0 && m.GetNanos() == 0) {
		return "", false
	}
	if m.GetNanos()%nanosPerCent != 0 || (whole && m.GetNanos() != 0) {
		return "", false
	}
	if m.GetNanos() == 0 {
		return json.Number(strconv.FormatInt(m.GetUnits(), 10)), true
	}
	cents := m.GetNanos() / nanosPerCent
	return json.Number(fmt.Sprintf("%d.%02d", m.GetUnits(), cents)), true
}

// requiresWholePayout reports Fincra's documented whole-number payout rules.
func requiresWholePayout(currency, providerCode string) bool {
	return strings.EqualFold(currency, "UGX") || strings.Contains(strings.ToUpper(providerCode), "AIRTEL")
}

// extraString returns the first non-empty string among the extras' keys.
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

// errMissingEmail is returned when no customer email can be found.
var errMissingEmail = errors.New("no customer email")

// buildCustomer assembles Fincra's required customer from the prompt.
func buildCustomer(extra *structpb.Struct, phone, fallbackEmail string) (client.Customer, error) {
	email := extraString(extra, collection.ExtraCustomerEmail, collection.ExtraEmail)
	if email == "" {
		email = fallbackEmail
	}
	if email == "" {
		return client.Customer{}, errMissingEmail
	}
	name := extraString(extra, collection.ExtraCustomerName, collection.ExtraDisplayName, "name")
	if name == "" {
		name = "Customer"
	}
	return client.Customer{Name: name, Email: email, PhoneNumber: phone}, nil
}

// tenancy reads the tenant and partition the message was published under.
func tenancy(ctx context.Context, headers map[string]string) (string, string) {
	tenantID, partitionID := headers["tenant_id"], headers["partition_id"]
	if claims := security.ClaimsFromContext(ctx); claims != nil {
		if tenantID == "" {
			tenantID = claims.GetTenantID()
		}
		if partitionID == "" {
			partitionID = claims.GetPartitionID()
		}
	}
	return tenantID, partitionID
}

// webhookURL is this service's webhook endpoint carrying the tenancy and
// credential connection the callback must be processed under.
func webhookURL(base, tenantID, partitionID, connection string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return ""
	}
	q := url.Values{}
	if tenantID != "" {
		q.Set("tenant_id", tenantID)
	}
	if partitionID != "" {
		q.Set("partition_id", partitionID)
	}
	if connection != "" {
		q.Set("connection", connection)
	}
	u := base + "/webhook/fincra"
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

// collectionMethod is the requested rail: mobile_money, bank_transfer,
// card, or "" when the prompt does not say.
func collectionMethod(extra *structpb.Struct) string {
	switch strings.ToLower(extraString(extra, collection.ExtraPaymentMethodType, collection.ExtraChannelType)) {
	case collection.PaymentMethodTypeMobileMoney, "momo", "mobile-money":
		return client.ChargeTypeMobileMoney
	case collection.PaymentMethodTypeBankTransfer, "bank", "transfer":
		return client.ChargeTypeBankTransfer
	case collection.PaymentMethodTypeCard:
		return client.ChargeTypeCard
	}
	return ""
}

// parseCodeMap parses a {"KEY":"CODE"} JSON object with upper-cased keys.
func parseCodeMap(raw string) map[string]string {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	var m map[string]string
	if json.Unmarshal([]byte(raw), &m) != nil {
		return out
	}
	for k, v := range m {
		if v = strings.TrimSpace(v); v != "" {
			out[strings.ToUpper(strings.TrimSpace(k))] = v
		}
	}
	return out
}

var (
	breakTags = regexp.MustCompile(`(?i)<\s*(br|/p|/li|/ol|/ul)\s*/?>`)
	anyTag    = regexp.MustCompile(`<[^>]*>`)
	spaces    = regexp.MustCompile(`[ \t]+`)
	newlines  = regexp.MustCompile(`\s*\n\s*`)
)

// plainText turns Fincra's HTML payment instructions into plain text.
func plainText(s string) string {
	s = breakTags.ReplaceAllString(s, "\n")
	s = anyTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = spaces.ReplaceAllString(s, " ")
	s = newlines.ReplaceAllString(s, "\n")
	return strings.TrimSpace(s)
}

// failureExtras is the portable failure detail for a status update.
func failureExtras(code, message string) map[string]any {
	if code == "" {
		code = failureProviderError
	}
	if message == "" {
		message = failureMessages[code]
	}
	if message == "" {
		message = failureMessages[failureProviderError]
	}
	return map[string]any{"failure_code": code, "failure_message": message}
}

// providerFailureExtras describes a failed Fincra call.
func providerFailureExtras(err error) map[string]any {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		return failureExtras(apiErr.Code(), apiErr.Message)
	}
	return failureExtras(failureProviderError, err.Error())
}

// formatAmount renders a float amount without trailing zeros.
func formatAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
