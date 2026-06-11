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
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/antinvestor/service-payments/apps/integrations/pawapay/config"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	centsPerUnit = 100
	nanosPerCent = 1e7

	customerMessageMinLen = 4
	customerMessageMaxLen = 22
)

// paymentUUID maps an internal payment ID to the UUIDv4 format pawaPay
// requires for deposit/payout/refund IDs. IDs that are already UUIDs pass
// through unchanged; any other ID is hashed deterministically so queue
// redeliveries reuse the same pawaPay ID and are deduplicated by pawaPay
// (DUPLICATE_IGNORED) instead of double-charging.
func paymentUUID(id string) string {
	if u, err := uuid.Parse(id); err == nil {
		return u.String()
	}

	sum := sha256.Sum256([]byte(id))
	b := sum[:16]
	b[6] = (b[6] & 0x0f) | 0x40 //nolint:mnd // version 4 bits per RFC 4122
	b[8] = (b[8] & 0x3f) | 0x80 //nolint:mnd // RFC 4122 variant bits
	u, _ := uuid.FromBytes(b)
	return u.String()
}

// formatMoneyAmount converts a protobuf Money value into pawaPay's decimal
// string format: no leading zeroes, no trailing decimal zeroes, at most two
// decimal places (the maximum any pawaPay provider supports).
func formatMoneyAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
}) string {
	if amount == nil {
		return "0"
	}

	units := amount.GetUnits()
	cents := (int64(amount.GetNanos()) + nanosPerCent/2) / nanosPerCent // round to hundredths
	if cents >= centsPerUnit {
		units += cents / centsPerUnit
		cents %= centsPerUnit
	}

	if cents <= 0 {
		return strconv.FormatInt(units, 10)
	}

	s := fmt.Sprintf("%d.%02d", units, cents)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// sanitizeCustomerMessage reduces a free-text narration to pawaPay's
// customerMessage constraints: 4-22 alphanumeric characters and spaces.
// It returns an empty string when no valid message can be produced, in
// which case the field is omitted and pawaPay defaults to the company name.
func sanitizeCustomerMessage(msg string) string {
	var b strings.Builder
	for _, r := range msg {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ':
			b.WriteRune(r)
		default:
			// drop unsupported characters
		}
	}

	clean := strings.Join(strings.Fields(b.String()), " ")
	if len(clean) > customerMessageMaxLen {
		clean = strings.TrimSpace(clean[:customerMessageMaxLen])
	}
	if len(clean) < customerMessageMinLen {
		return ""
	}
	return clean
}

// paymentMetadata builds the metadata attached to every pawaPay payment so
// callbacks can be correlated back to the originating record, tenant and
// credential connection. The webhook server treats callbacks as untrusted
// notifications and re-fetches the payment from pawaPay; these values are
// only trusted once they come back on that verified record.
func paymentMetadata(entityType, internalID string, headers map[string]string) map[string]string {
	return map[string]string{
		"entityType":  entityType,
		"paymentId":   internalID,
		"tenantId":    headers["tenant_id"],
		"partitionId": headers["partition_id"],
		"connection":  headers[config.HeaderConnectionCredentials],
	}
}

// resolveProvider determines the mobile money provider for a payment, in
// order of precedence: the payment's extra "provider" field, the
// credential/config default, and finally the pawaPay predict-provider
// endpoint. The prediction also sanitises the phone number into the MSISDN
// format the rest of the API expects.
// It returns the provider code and the (possibly sanitised) MSISDN.
func resolveProvider(
	ctx context.Context,
	pawapayCli client.PawapayClient,
	creds *client.Credentials,
	extraProvider, phoneNumber string,
) (string, string, error) {
	if extraProvider != "" {
		return extraProvider, phoneNumber, nil
	}
	if creds.Provider != "" {
		return creds.Provider, phoneNumber, nil
	}

	prediction, err := pawapayCli.PredictProvider(ctx, creds, phoneNumber)
	if err != nil {
		return "", "", fmt.Errorf("predict provider for phone number: %w", err)
	}
	return prediction.Provider, prediction.PhoneNumber, nil
}

// failureCode extracts the failure code from an optional failure reason.
func failureCode(reason *client.FailureReason) string {
	if reason == nil {
		return ""
	}
	return reason.FailureCode
}

// failureMessage extracts the failure message from an optional failure reason.
func failureMessage(reason *client.FailureReason) string {
	if reason == nil {
		return ""
	}
	return reason.FailureMessage
}

// extraString reads a string field from a payment's Extra struct.
func extraString(extra *structpb.Struct, key string) string {
	if extra == nil {
		return ""
	}
	if v, ok := extra.GetFields()[key]; ok {
		return v.GetStringValue()
	}
	return ""
}
