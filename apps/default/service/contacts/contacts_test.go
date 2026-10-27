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

package contacts_test

import (
	"context"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antinvestor/service-payments/apps/default/service/contacts"
)

type profiles struct {
	profilev1connect.ProfileServiceClient

	contacts []*profilev1.ContactObject
}

func (p *profiles) GetById(context.Context, *profilev1.GetByIdRequest) (*profilev1.GetByIdResponse, error) {
	return &profilev1.GetByIdResponse{Data: &profilev1.ProfileObject{Id: "p1", Contacts: p.contacts}}, nil
}

func link(profileID, contactID string) *commonv1.ContactLink {
	return &commonv1.ContactLink{ProfileId: profileID, ContactId: contactID}
}

func TestResolveReadsTheContactsDetailAndKindByID(t *testing.T) {
	t.Parallel()
	svc := &profiles{contacts: []*profilev1.ContactObject{
		{Id: "c-mail", Type: profilev1.ContactType_EMAIL, Detail: "a@b.c"},
		{Id: "c-phone", Type: profilev1.ContactType_MSISDN, Detail: "+256700000001"},
	}}

	got, err := contacts.Resolve(t.Context(), svc, link("p1", "c-phone"))
	require.NoError(t, err)
	assert.Equal(t, contacts.Contact{ID: "c-phone", Detail: "+256700000001", Kind: contacts.RouteTypeMSISDN}, got)

	got, err = contacts.Resolve(t.Context(), svc, link("p1", "c-mail"))
	require.NoError(t, err)
	assert.Equal(t, contacts.RouteTypeEmail, got.Kind)

	_, err = contacts.Resolve(t.Context(), svc, link("p1", "c-other"))
	require.ErrorIs(t, err, contacts.ErrUnknown, "a contact the profile does not hold")

	got, err = contacts.Resolve(t.Context(), nil, link("", "+256700000009"))
	require.NoError(t, err)
	assert.Equal(t, contacts.Contact{Detail: "+256700000009", Kind: contacts.RouteTypeMSISDN}, got, "a number is used as it is")
}

// A payee named by profile alone is paid on its mobile number, a verified
// one first.
func TestResolvePayeeDefaultsToTheProfilesMobileNumber(t *testing.T) {
	t.Parallel()
	svc := &profiles{contacts: []*profilev1.ContactObject{
		{Id: "c-mail", Type: profilev1.ContactType_EMAIL, Detail: "a@b.c", Verified: true},
		{Id: "c-unverified", Type: profilev1.ContactType_MSISDN, Detail: "+256700000001"},
		{Id: "c-verified", Type: profilev1.ContactType_MSISDN, Detail: "+256700000002", Verified: true},
	}}

	got, err := contacts.ResolvePayee(t.Context(), svc, link("p1", ""))
	require.NoError(t, err)
	assert.Equal(t, contacts.Contact{ID: "c-verified", Detail: "+256700000002", Kind: contacts.RouteTypeMSISDN}, got)

	got, err = contacts.ResolvePayee(t.Context(), svc, link("p1", "c-unverified"))
	require.NoError(t, err)
	assert.Equal(t, "c-unverified", got.ID, "a named contact is paid as named")

	_, err = contacts.ResolvePayee(t.Context(), &profiles{contacts: []*profilev1.ContactObject{
		{Id: "c-mail", Type: profilev1.ContactType_EMAIL, Detail: "a@b.c"},
	}}, link("p1", ""))
	require.ErrorIs(t, err, contacts.ErrNoMobileNumber)
}
