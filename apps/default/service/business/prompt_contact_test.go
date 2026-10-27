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
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// profileWithContacts answers GetById with one profile's contacts.
type profileWithContacts struct {
	profilev1connect.ProfileServiceClient

	profile *profilev1.ProfileObject
	calls   int
}

func (p *profileWithContacts) GetById(
	_ context.Context,
	req *profilev1.GetByIdRequest,
) (*profilev1.GetByIdResponse, error) {
	p.calls++
	if req.GetId() != p.profile.GetId() {
		return &profilev1.GetByIdResponse{}, nil
	}
	return &profilev1.GetByIdResponse{Data: p.profile}, nil
}

func payerProfile() *profileWithContacts {
	return &profileWithContacts{profile: &profilev1.ProfileObject{
		Id: "profile-1",
		Contacts: []*profilev1.ContactObject{
			{Id: "contact-email", Detail: "payer@example.com"},
			{Id: "contact-msisdn", Detail: "+256701000111"},
		},
	}}
}

func TestPayerNumberIsReadFromTheProfileByContactID(t *testing.T) {
	profiles := payerProfile()
	number, err := payerNumber(context.Background(), profiles,
		&commonv1.ContactLink{ProfileId: "profile-1", ContactId: "contact-msisdn"})
	require.NoError(t, err)
	assert.Equal(t, "+256701000111", number)
}

func TestPayerNumberGivenAsANumberIsUsedAsItIs(t *testing.T) {
	profiles := payerProfile()
	number, err := payerNumber(context.Background(), profiles,
		&commonv1.ContactLink{ProfileId: "profile-1", ContactId: "+256701000111"})
	require.NoError(t, err)
	assert.Empty(t, number, "nothing to resolve")
	assert.Zero(t, profiles.calls)
}

func TestPayerContactNotOnTheProfileIsRefused(t *testing.T) {
	_, err := payerNumber(context.Background(), payerProfile(),
		&commonv1.ContactLink{ProfileId: "profile-1", ContactId: "contact-of-someone-else"})
	require.ErrorIs(t, err, ErrPayerContactUnknown)
}
