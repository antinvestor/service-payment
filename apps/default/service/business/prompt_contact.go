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
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"

	"github.com/antinvestor/service-payments/apps/default/service/contacts"
)

// ErrPayerContactUnknown is returned when a prompt names a payer contact the
// payer's profile does not hold.
var ErrPayerContactUnknown = contacts.ErrUnknown

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
	if contactID == "" || strings.TrimSpace(source.GetProfileId()) == "" || contacts.IsPhoneNumber(contactID) {
		return "", nil
	}
	contact, err := contacts.Resolve(ctx, profiles, source)
	if err != nil {
		return "", err
	}
	return contact.Detail, nil
}
