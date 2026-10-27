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

// Package contacts resolves the contact a payment names, by profile and
// contact id, to its detail and kind. Callers outside the payment service name
// payers and payees by id only; the number or address is read here, as a
// payment leaves for its provider, and never stored.
package contacts

import (
	"context"
	"errors"
	"fmt"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
)

// Route types name the kind of contact a payment reaches, so payment routes
// are chosen by the recipient's contact rather than by the caller.
const (
	RouteTypeMSISDN = "msisdn"
	RouteTypeEmail  = "email"
)

// ErrUnknown is returned when a payment names a contact its profile does not
// hold.
var ErrUnknown = errors.New("the profile has no such contact")

// Contact is a resolved contact: its detail (number or address) and the
// route type that reaches it.
type Contact struct {
	ID     string
	Detail string
	Kind   string
}

// Resolve reads the contact link names. A contact given as a phone number is
// used as it is. It returns a zero Contact when link names no contact.
func Resolve(
	ctx context.Context,
	profiles profilev1connect.ProfileServiceClient,
	link *commonv1.ContactLink,
) (Contact, error) {
	contactID := strings.TrimSpace(link.GetContactId())
	profileID := strings.TrimSpace(link.GetProfileId())
	if contactID == "" {
		return Contact{}, nil
	}
	if IsPhoneNumber(contactID) {
		return Contact{Detail: contactID, Kind: RouteTypeMSISDN}, nil
	}
	if profileID == "" {
		return Contact{}, fmt.Errorf("resolve contact %s: %w", contactID, ErrUnknown)
	}
	if profiles == nil {
		return Contact{}, fmt.Errorf("resolve contact %s: profile service not configured", contactID)
	}
	resp, err := profiles.GetById(ctx, &profilev1.GetByIdRequest{Id: profileID})
	if err != nil {
		return Contact{}, fmt.Errorf("resolve contact %s: %w", contactID, err)
	}
	for _, c := range resp.GetData().GetContacts() {
		if c.GetId() != contactID || strings.TrimSpace(c.GetDetail()) == "" {
			continue
		}
		kind := RouteTypeEmail
		if c.GetType() == profilev1.ContactType_MSISDN {
			kind = RouteTypeMSISDN
		}
		return Contact{ID: c.GetId(), Detail: strings.TrimSpace(c.GetDetail()), Kind: kind}, nil
	}
	return Contact{}, ErrUnknown
}

// ErrNoMobileNumber is returned when a payee named by profile alone has no
// mobile number to be paid on.
var ErrNoMobileNumber = errors.New("the profile has no mobile number")

// ResolvePayee reads the contact a payment out reaches. A payee named by
// profile alone is paid on the profile's mobile number: its first verified
// one, else its first.
func ResolvePayee(
	ctx context.Context,
	profiles profilev1connect.ProfileServiceClient,
	link *commonv1.ContactLink,
) (Contact, error) {
	if strings.TrimSpace(link.GetContactId()) != "" {
		return Resolve(ctx, profiles, link)
	}
	profileID := strings.TrimSpace(link.GetProfileId())
	if profileID == "" {
		return Contact{}, nil
	}
	if profiles == nil {
		return Contact{}, fmt.Errorf("resolve payee %s: profile service not configured", profileID)
	}
	resp, err := profiles.GetById(ctx, &profilev1.GetByIdRequest{Id: profileID})
	if err != nil {
		return Contact{}, fmt.Errorf("resolve payee %s: %w", profileID, err)
	}
	var chosen *profilev1.ContactObject
	for _, c := range resp.GetData().GetContacts() {
		if c.GetType() != profilev1.ContactType_MSISDN || strings.TrimSpace(c.GetDetail()) == "" {
			continue
		}
		if chosen == nil || (c.GetVerified() && !chosen.GetVerified()) {
			chosen = c
		}
	}
	if chosen == nil {
		return Contact{}, ErrNoMobileNumber
	}
	return Contact{ID: chosen.GetId(), Detail: strings.TrimSpace(chosen.GetDetail()), Kind: RouteTypeMSISDN}, nil
}

// IsPhoneNumber reports whether s is a phone number (digits, optionally led
// by +) rather than a contact id.
func IsPhoneNumber(s string) bool {
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
