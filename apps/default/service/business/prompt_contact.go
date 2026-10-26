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
	"context"
	"errors"
	"fmt"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
)

// ErrPayerContactUnknown is returned when a prompt names a payer contact the
// payer's profile does not hold.
var ErrPayerContactUnknown = errors.New("the payer's profile has no such contact")

// payerNumber is the number a prompt dials. A caller that names the payer by
// profile and contact id never handles the number: it is read from the
// profile service here, as the prompt leaves for the provider. A contact
// given as a number already is used as it is. It returns "" when the prompt
// names no contact to resolve.
func payerNumber(
	ctx context.Context,
	profiles profilev1connect.ProfileServiceClient,
	source *commonv1.ContactLink,
) (string, error) {
	contactID := strings.TrimSpace(source.GetContactId())
	profileID := strings.TrimSpace(source.GetProfileId())
	if contactID == "" || profileID == "" || isPhoneNumber(contactID) {
		return "", nil
	}
	if profiles == nil {
		return "", fmt.Errorf("resolve payer contact %s: profile service not configured", contactID)
	}
	resp, err := profiles.GetById(ctx, &profilev1.GetByIdRequest{Id: profileID})
	if err != nil {
		return "", fmt.Errorf("resolve payer contact %s: %w", contactID, err)
	}
	for _, c := range resp.GetData().GetContacts() {
		if c.GetId() == contactID && strings.TrimSpace(c.GetDetail()) != "" {
			return strings.TrimSpace(c.GetDetail()), nil
		}
	}
	return "", ErrPayerContactUnknown
}

// isPhoneNumber reports whether s is a phone number (digits, optionally
// led by +) rather than a contact id.
func isPhoneNumber(s string) bool {
	digits := strings.TrimPrefix(s, "+")
	if digits == "" {
		return false
	}
	for _, r := range digits {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
