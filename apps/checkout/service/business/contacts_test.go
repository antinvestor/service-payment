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
	"testing"

	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyContactDetail(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want contactKind
	}{
		{"bob@example.com", contactKindEmail},
		{"  alice@stawi.org ", contactKindEmail},
		{"254712345678", contactKindPhone},
		{"+254 712 345 678", contactKindPhone},
		{"+1-415-555-0100", contactKindPhone},
		{"", contactKindUnknown},
		{"not-an-email", contactKindUnknown},
		{"123", contactKindUnknown}, // too short
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, classifyContactDetail(tc.in), tc.in)
	}
}

func TestClassifyProfileContact_ZeroTypePhoneIsPhone(t *testing.T) {
	t.Parallel()
	// EMAIL is protobuf zero — missing type + phone detail must not become email.
	c := &profilev1.ContactObject{
		Id:     "c1",
		Type:   profilev1.ContactType_EMAIL, // 0
		Detail: "254712345678",
	}
	assert.Equal(t, contactKindPhone, classifyProfileContact(c))
}

func TestClassifyProfileContact_TypedEmail(t *testing.T) {
	t.Parallel()
	c := &profilev1.ContactObject{
		Id:     "c2",
		Type:   profilev1.ContactType_EMAIL,
		Detail: "bob@example.com",
	}
	assert.Equal(t, contactKindEmail, classifyProfileContact(c))
}

func TestClassifyProfileContact_MsisdnTypedEmailDetail(t *testing.T) {
	t.Parallel()
	// Mislabelled: type phone but value is email.
	c := &profilev1.ContactObject{
		Id:     "c3",
		Type:   profilev1.ContactType_MSISDN,
		Detail: "bob@example.com",
	}
	assert.Equal(t, contactKindEmail, classifyProfileContact(c))
}

func TestFirstEmailAndPhoneSplit(t *testing.T) {
	t.Parallel()
	contacts := []*profilev1.ContactObject{
		{Id: "p1", Type: profilev1.ContactType_EMAIL, Detail: "254700000001"}, // zero-type phone
		{Id: "e1", Type: profilev1.ContactType_EMAIL, Detail: "bob@example.com"},
		{Id: "p2", Type: profilev1.ContactType_MSISDN, Detail: "+254711000000"},
	}
	assert.Equal(t, "bob@example.com", firstEmailFromContacts(contacts))
	phones := phoneContactsFromProfile(contacts)
	require.Len(t, phones, 2)
	assert.Equal(t, "254700000001", phones[0]["msisdn"])
	assert.Equal(t, "+254711000000", phones[1]["msisdn"])
}

func TestNormalizeCallerContacts(t *testing.T) {
	t.Parallel()
	in := []PayerContactInput{
		{ContactID: "1", Msisdn: "bob@example.com"},
		{ContactID: "2", Msisdn: "254712345678"},
	}
	assert.Equal(t, "bob@example.com", firstEmailFromCallerContacts(in))
	phones := normalizeCallerPhoneContacts(in)
	require.Len(t, phones, 1)
	assert.Equal(t, "254712345678", phones[0]["msisdn"])
}
